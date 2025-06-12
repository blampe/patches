package s3legacy

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/blampe/patches/mirrors/aws/v5/internal/tfresource"
)

// FORK: Adding the retryOnAWSCode to the fork for the old AWS S3 Logic
func retryOnAWSCode(code string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := retry.Retry(2*time.Minute, func() *retry.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			if tfawserr.ErrCodeEquals(err, code) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = f()
	}

	return resp, err
}
