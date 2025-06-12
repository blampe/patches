package s3legacy

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

type tfawserrshim struct{}

var tfawserr *tfawserrshim = &tfawserrshim{}

func (*tfawserrshim) ErrCodeEquals(err error, codes ...string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		for _, code := range codes {
			if awsErr.Code() == code {
				return true
			}
		}
	}
	return false
}

func (*tfawserrshim) ErrMessageContains(err error, code string, message string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == code && strings.Contains(awsErr.Message(), message)
	}
	return false
}

func (*tfawserrshim) ErrStatusCodeEquals(err error, statusCode int) bool {
	var awsErr awserr.RequestFailure
	if errors.As(err, &awsErr) {
		return awsErr.StatusCode() == statusCode
	}
	return false
}
