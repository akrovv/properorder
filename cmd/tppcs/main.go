package main

import (
	"github.com/akrovv/properorder/analyzers/properorder"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(properorder.New())
}
