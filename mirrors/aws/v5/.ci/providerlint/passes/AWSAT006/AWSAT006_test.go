package AWSAT006_test

import (
	"testing"

	"github.com/blampe/patches/mirrors/aws/v5/ci/providerlint/passes/AWSAT006"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT006(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT006.Analyzer, "testdata/src/a")
}
