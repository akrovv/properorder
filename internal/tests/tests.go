package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// AnalyzeCodeInTestdata runs analyzer on the testdata directory.
func AnalyzeCodeInTestdata(t *testing.T, analyzer *analysis.Analyzer) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var printer errsPrinter
	analysistest.Run(&printer, filepath.Join(wd, "testdata/src"), analyzer, "./...")
	if printer.PrintedLines > 0 {
		t.Fail()
	}
}

type errsPrinter struct {
	PrintedLines int
}

func (p *errsPrinter) Errorf(format string, args ...interface{}) {
	fmt.Println(args...)
	p.PrintedLines++
}
