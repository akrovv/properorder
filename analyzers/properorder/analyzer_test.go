package properorder_test

import (
	"testing"

	"gitlab.com/common/linter/analyzers/properorder"
	"gitlab.com/common/linter/internal/tests"
)

func TestAll(t *testing.T) {
	tests.AnalyzeCodeInTestdata(t, properorder.New())
}
