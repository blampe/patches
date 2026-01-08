// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
)

func TestResourceDirectoryReadDoesNotPanicOnMissingVpcSettings(t *testing.T) {
	d := ResourceDirectory().TestResourceData()
	dir := &awstypes.DirectoryDescription{
		VpcSettings: nil,
	}
	diags := resourceDirectoryReadDescription(d, dir)
	if diags.HasError() {
		t.Errorf("Unexpected errors in diags: %v", diags)
		t.Fail()
	}
}
