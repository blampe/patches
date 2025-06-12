package AWSAT005_test

import (
	"testing"

	"github.com/blampe/patches/mirrors/aws/v5/ci/providerlint/passes/AWSAT005"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT005(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT005.Analyzer, "testdata/src/a")
}
