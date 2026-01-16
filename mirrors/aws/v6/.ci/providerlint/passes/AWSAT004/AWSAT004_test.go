// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package AWSAT004_test

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/blampe/patches/mirrors/aws/v6/ci/providerlint/passes/AWSAT004"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT004(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT004.Analyzer, "testdata/src/a")
}
