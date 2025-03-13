package main

import (
	"gitlab.com/common/linter/analyzers/properorder"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(properorder.New())
}
