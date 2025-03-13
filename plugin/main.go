package main

import (
	"errors"
	"github.com/akrovv/properorder/analyzers/properorder"
	"golang.org/x/tools/go/analysis"
)

var linters = map[string]*analysis.Analyzer{
	"properorder": properorder.New(),
}

// golangci-linter plugin (.so lib that exports New function)
func New(conf any) ([]*analysis.Analyzer, error) {
	var analyzers []*analysis.Analyzer
	if conf == nil {
		for _, linter := range linters {
			analyzers = append(analyzers, linter)
		}
	} else {
		confMap, ok := conf.(map[string]interface{})
		if !ok {
			return analyzers, errors.New("conf must be a map[string]interface{}")
		}
		for k, _ := range linters {
			value, ok := confMap[k]
			if !ok {
				analyzers = append(analyzers, linters[k])
				continue
			}
			if confValue, _ := value.(bool); confValue {
				analyzers = append(analyzers, linters[k])
			}
		}
	}

	return analyzers, nil
}
