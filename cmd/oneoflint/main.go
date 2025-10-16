package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/gomoni/sumlint/internal/lint"
)

func main() {
	unitchecker.Main(lint.Oneof)
}
