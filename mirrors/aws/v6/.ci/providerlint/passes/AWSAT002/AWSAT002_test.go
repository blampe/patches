// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package AWSAT002_test

import (
	"testing"

	"github.com/blampe/patches/mirrors/aws/v6/ci/providerlint/passes/AWSAT002"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT002(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT002.Analyzer, "testdata/src/a")
}
