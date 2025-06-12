// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"testing"
)

func TestResourceLandingZoneNormalizeManifest(t *testing.T) {
	actual, err := resourceLandingZoneNormalizeManifest(`
	{
	  "governedRegions": [
	    "REGION"
	  ],
	  "organizationStructure": {
	    "security": {
	      "name": "Security"
	    }
	  },
	  "centralizedLogging": {
	    "accountId": "89XXXXXXXX39",
	    "configurations": {
	      "accessLoggingBucket": {
		"retentionDays": "3650"
	      },
	      "kmsKeyArn": "arn:PARTITION:kms:REGION:89XXXXXXXX25:key/10e27ec4-5555-4444-b408-777777777777",
	      "loggingBucket": {
		"retentionDays": "365"
	      }
	    },
	    "enabled": true
	  },
	  "securityRoles": {
	    "accountId": "89XXXXXXXX42"
	  },
	  "accessManagement": {
	    "enabled": true
	  }
	}`)
	if err != nil {
		t.Error(err)
	}
	expected := `{"accessManagement":{"enabled":true},"centralizedLogging":{"accountId":"89XXXXXXXX39","configurations":{"accessLoggingBucket":{"retentionDays":3650},"kmsKeyArn":"arn:PARTITION:kms:REGION:89XXXXXXXX25:key/10e27ec4-5555-4444-b408-777777777777","loggingBucket":{"retentionDays":365}},"enabled":true},"governedRegions":["REGION"],"organizationStructure":{"security":{"name":"Security"}},"securityRoles":{"accountId":"89XXXXXXXX42"}}`
	if expected != actual {
		t.Logf("Expected: %s", expected)
		t.Logf("Actual: %s", actual)
		t.Error("Unexpected result")
	}
}
