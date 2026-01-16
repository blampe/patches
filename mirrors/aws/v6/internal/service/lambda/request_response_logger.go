package lambda

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"io"
	"net/http"
	"strings"
	"time"
	_ "unsafe"
)

const (
	lambdaCreateOperation             = "CreateFunction"
	lambdaUpdateFunctionCodeOperation = "UpdateFunctionCode"
)

// Replaces the upstream logging middleware from https://github.com/hashicorp/aws-sdk-go-base/blob/main/logger.go#L107
// We do not want to log the Lambda Archive that is part of the request body because this leads to bloating memory
type wrappedRequestResponseLogger struct {
	wrapped middleware.DeserializeMiddleware
}

// ID is the middleware identifier.
func (r *wrappedRequestResponseLogger) ID() string {
	return "PULUMI_AWS_RequestResponseLogger"
}

func NewWrappedRequestResponseLogger(wrapped middleware.DeserializeMiddleware) middleware.DeserializeMiddleware {
	return &wrappedRequestResponseLogger{wrapped: wrapped}
}

//go:linkname decomposeHTTPResponse github.com/hashicorp/aws-sdk-go-base/v2.decomposeHTTPResponse
func decomposeHTTPResponse(ctx context.Context, resp *http.Response, elapsed time.Duration) (map[string]any, error)

func (r *wrappedRequestResponseLogger) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler,
) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	if awsmiddleware.GetServiceID(ctx) == "Lambda" {
		if op := awsmiddleware.GetOperationName(ctx); op != lambdaCreateOperation && op != lambdaUpdateFunctionCodeOperation {
			// pass through to the wrapped response logger for all other lambda operations that do not send the code as part of the request body
			return r.wrapped.HandleDeserialize(ctx, in, next)
		}
	}

	// Inlined the logging middleware from https://github.com/hashicorp/aws-sdk-go-base/blob/main/logger.go and patching
	// out the request body logging
	logger := logging.RetrieveLogger(ctx)
	region := awsmiddleware.GetRegion(ctx)

	if signingRegion := awsmiddleware.GetSigningRegion(ctx); signingRegion != region { //nolint:staticcheck // Not retrievable elsewhere
		ctx = logger.SetField(ctx, string(logging.SigningRegionKey), signingRegion)
	}
	if awsmiddleware.GetEndpointSource(ctx) == aws.EndpointSourceCustom {
		ctx = logger.SetField(ctx, string(logging.CustomEndpointKey), true)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unexpected request middleware type %T", in.Request)
	}

	rc := req.Build(ctx)

	originalBody := rc.Body
	// remove the body from the logging output. This is the main change compared to the upstream logging middleware
	redactedBody := strings.NewReader("[Redacted]")
	rc.Body = io.NopCloser(redactedBody)
	rc.ContentLength = redactedBody.Size()

	requestFields, err := logging.DecomposeHTTPRequest(ctx, rc)
	if err != nil {
		return out, metadata, fmt.Errorf("decomposing request: %w", err)
	}
	logger.Debug(ctx, "HTTP Request Sent", requestFields)

	// reconstruct the original request
	req, err = req.SetStream(originalBody)
	if err != nil {
		return out, metadata, err
	}
	in.Request = req

	start := time.Now()
	out, metadata, err = next.HandleDeserialize(ctx, in)
	duration := time.Since(start)

	if err != nil {
		return out, metadata, err
	}

	if res, ok := out.RawResponse.(*smithyhttp.Response); !ok {
		return out, metadata, fmt.Errorf("unknown response type: %T", out.RawResponse)
	} else {
		responseFields, err := decomposeHTTPResponse(ctx, res.Response, duration)
		if err != nil {
			return out, metadata, fmt.Errorf("decomposing response: %w", err)
		}
		logger.Debug(ctx, "HTTP Response Received", responseFields)
	}

	return out, metadata, err
}
