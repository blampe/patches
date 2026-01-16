// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"errors"

	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/blampe/patches/mirrors/aws/v6/internal/errs"
)

func failedItemsError(apiObject map[string]awstypes.FailedItemDetails) error {
	var es []error

	for k, v := range apiObject {
		es = append(es, errs.APIError(v.FailureCode, k))
	}

	return errors.Join(es...)
}
