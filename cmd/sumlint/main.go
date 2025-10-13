package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/gomoni/sumlint/internal/sumlint"
)

func main() {
	unitchecker.Main(sumlint.Analyzer)
}
