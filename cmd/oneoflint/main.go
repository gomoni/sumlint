package main

import (
	"github.com/gomoni/sumlint/internal/lint"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(lint.Oneof)
}
