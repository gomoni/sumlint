package sumlint_test

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/gomoni/sumlint/internal/sumlint"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	tests := []string{
		"exhaustive",
		"nonexhaustive",
		"dflt",
	}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			testAnalyzer(t, name)
		})
	}
}

func testAnalyzer(t *testing.T, name string) {
	t.Helper()
	testdata := analysistest.TestData()

	var got []string
	t2 := errorfunc(func(s string) { got = append(got, s) })
	results := analysistest.Run(t2, testdata, sumlint.Analyzer, name)

	// Dump exported facts for debugging.
	for _, r := range results {
		for _, d := range r.Diagnostics {
			t.Logf("DIAG: %+v %s", d.Pos, d.Message)
		}
		if len(r.Facts) == 0 {
			t.Log("  (no facts)")
			continue
		}
		for obj, facts := range r.Facts {
			on := objectString(obj)
			if len(facts) == 0 {
				t.Logf("  %s: (no fact values slice entries)", on)
				continue
			}
			for _, f := range facts {
				t.Logf("  %s: %T %#v", on, f, f)
			}
		}
	}

	if len(got) != 0 {
		t.Fatalf("expected no messages, got: %+v", got)
	}
}

func objectString(obj types.Object) string {
	if obj == nil {
		return "<nil>"
	}
	if pkg := obj.Pkg(); pkg != nil {
		return pkg.Path() + "." + obj.Name()
	}
	return obj.Name()
}

type errorfunc func(string)

func (f errorfunc) Errorf(format string, args ...any) {
	f(fmt.Sprintf(format, args...))
}
