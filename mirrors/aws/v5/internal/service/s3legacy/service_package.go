package s3legacy

import (
	"context"

	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceBucketLegacy,
			TypeName: "aws_s3_bucket_legacy",
			Name:     "BucketLegacy",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "bucket",
				ResourceType:        "Bucket",
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return "s3legacy"
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}

// import (
// 	"context"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/aws/retry"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// 	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
// 	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
// 	"github.com/blampe/patches/mirrors/aws/v5/names"
// )

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
// func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*s3.Client, error) {
// 	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

// 	return s3.NewFromConfig(cfg, func(o *s3.Options) {
// 		if endpoint := config["endpoint"].(string); endpoint != "" {
// 			o.BaseEndpoint = aws.String(endpoint)
// 		} else if o.Region == names.USEast1RegionID && config["s3_us_east_1_regional_endpoint"].(string) != "regional" {
// 			// Maintain the AWS SDK for Go v1 default of using the global endpoint in us-east-1.
// 			// See https://github.com/blampe/patches/mirrors/aws/v5/issues/33028.
// 			o.Region = names.GlobalRegionID
// 		}
// 		o.UsePathStyle = config["s3_use_path_style"].(bool)

// 		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
// 			if tfawserr.ErrMessageContains(err, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
// 				return aws.TrueTernary
// 			}
// 			return aws.UnknownTernary // Delegate to configured Retryer.
// 		}))
// 	}), nil
// }
