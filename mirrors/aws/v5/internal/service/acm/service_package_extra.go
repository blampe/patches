package acm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
)

func (p *servicePackage) pulumiCustomizeRetries(cfg aws.Config) func(*acm.Options) {
	return func(o *acm.Options) {
		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws.Ternary {
			if tfawserr_sdkv2.ErrMessageContains(err, "LimitExceededException", "the maximum number of") &&
				tfawserr_sdkv2.ErrMessageContains(err, "LimitExceededException", "certificates in the last year") {
				return aws.FalseTernary
			}
			return aws.UnknownTernary // Delegate to configured Retryer.
		}))
	}
}
