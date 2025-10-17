package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	sumlint   string
	oneoflint string
)

func TestMain(m *testing.M) {
	var err error
	sumlint, err = executable("sumlint")
	if err != nil {
		log.Fatal(err)
	}
	oneoflint, err = executable("oneoflint")
	if err != nil {
		log.Fatal(err)
	}

	ret := m.Run()
	os.Exit(ret)
}

/*
📦[michal@3key tests]$ go vet -vettool="$(go env GOPATH)/bin/sumlint" .
# github.com/gomoni/sumlint/test
./sum.go:23:2: missing default case on SumFoo: code cannot handle nil interface
./sum.go:29:2: non-exhaustive type switch on SumFoo: missing cases for: github.com/gomoni/sumlint/test.B
*/

func TestLint(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(file)             // dir of main_test.go
	testDir := filepath.Join(root, "test") // or relative you need
	t.Logf("sumlint: %s", sumlint)
	t.Logf("testDir: %s", testDir)
	err := os.Chdir(testDir)
	if err != nil {
		t.Fatalf("can't cd test: %s", err)
	}

	var testCases = []struct {
		scenario string
		given    string
		then     []string
	}{
		{
			scenario: "sumlint",
			given:    sumlint,
			then: []string{`# github.com/gomoni/sumlint/test`,
				`./sum.go:23:2: missing default case on SumFoo: code cannot handle nil interface`,
				`./sum.go:29:2: non-exhaustive type switch on SumFoo: missing cases for: github.com/gomoni/sumlint/test.B`,
			},
		},
		{
			scenario: "oneoflint",
			given:    oneoflint,
			then: []string{`# github.com/gomoni/sumlint/test`,
				`./sum.go:23:2: missing default case on SumFoo: code cannot handle nil interface`,
				`./sum.go:29:2: non-exhaustive type switch on SumFoo: missing cases for: github.com/gomoni/sumlint/test.B`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := t.Context()
			cmd := exec.CommandContext(ctx, "go", "vet", `-vettool=`+tc.given, ".")
			var out, stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			_ = cmd.Run()

			if out.String() != "" {
				t.Errorf("go vet stdout should be empty: got %s", out.String())
			}

			lines := strings.Split(stderr.String(), "\n")
			for idx := range tc.then {
				if tc.then[idx] != lines[idx] {
					t.Errorf("line %d does not match: expected %q: got %q", idx, tc.then[idx], lines[idx])
				}
			}
		})
	}
}

func executable(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	mode := fi.Mode()
	if mode.IsDir() {
		return "", errors.New("is a directory")
	}

	var apath string
	apath, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Any of owner/group/other execute bits
	if mode&0o111 != 0 {
		return apath, nil
	} else if filepath.Ext(path) == "exe" {
		return apath, nil
	}

	return "", fmt.Errorf("wrong %s: not executable", path)
}
