package main

import (
	"github.com/takoeight0821/exhaustinterface"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(exhaustinterface.Analyzer) }
