package properorder_test

import (
	"testing"

	"github.com/akrovv/properorder/analyzers/properorder"
	"github.com/akrovv/properorder/internal/tests"
)

func TestAll(t *testing.T) {
	tests.AnalyzeCodeInTestdata(t, properorder.New())
}
