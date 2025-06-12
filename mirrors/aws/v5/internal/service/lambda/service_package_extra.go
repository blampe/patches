package lambda

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go/middleware"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
)

// Customize lambda retries.
//
// References:
//
// https://github.com/blampe/patches/mirrors/aws/v5/blob/main/docs/retries-and-waiters.md
// https://github.com/pulumi/pulumi-aws/issues/3196
func (p *servicePackage) pulumiCustomizeLambdaRetries(cfg aws.Config) func(*lambda.Options) {
	retry := retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
		if tfawserr_sdkv2.ErrMessageContains(
			err,
			"KMSAccessDeniedException",
			"Lambda was unable to decrypt the environment variables because KMS access was denied.",
		) {
			// Do not retry this condition at all.
			return aws_sdkv2.FalseTernary
		}
		return aws_sdkv2.UnknownTernary // Delegate
	})

	return func(o *lambda.Options) {
		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry)

		// Switch out the terraform http logging middleware with a custom logging middleware that does not log the
		// lambda code. Logging the lambda code leads to memory bloating because it allocates a lot of copies of the
		// body
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			loggingMiddleware, err := stack.Deserialize.Remove("TF_AWS_RequestResponseLogger")
			if err != nil {
				return err
			}

			err = stack.Deserialize.Add(NewWrappedRequestResponseLogger(loggingMiddleware), middleware.After)
			return err
		})
	}
}
