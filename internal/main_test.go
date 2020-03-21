// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/panicparse/stack"
)

func TestProcess(t *testing.T) {
	out := &bytes.Buffer{}
	if err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, nil); err != nil {
		t.Fatal(err)
	}
	expected := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, expected, out.String())
}

func TestProcessFullPath(t *testing.T) {
	out := &bytes.Buffer{}
	if err := process(getReader(t), out, testPalette, stack.AnyValue, fullPath, false, true, "", nil, nil); err != nil {
		t.Fatal(err)
	}
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// "/" is used even on Windows.
	p := strings.Replace(filepath.Join(filepath.Dir(d), "cmd", "panic", "main.go"), "\\", "/", -1)
	expected := fmt.Sprintf("GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain F%s:52 ImainL()A\n", p)
	compareString(t, expected, out.String())
}

func TestProcessNoColor(t *testing.T) {
	out := &bytes.Buffer{}
	if err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, nil); err != nil {
		t.Fatal(err)
	}
	expected := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, expected, out.String())
}

func TestProcessMatch(t *testing.T) {
	out := &bytes.Buffer{}
	err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, regexp.MustCompile(`notpresent`))
	if err != nil {
		t.Fatal(err)
	}
	expected := "GOTRACEBACK=all\npanic: simple\n\n"
	compareString(t, expected, out.String())
}

func TestProcessFilter(t *testing.T) {
	out := &bytes.Buffer{}
	err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", regexp.MustCompile(`notpresent`), nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, expected, out.String())
}

func TestMainFn(t *testing.T) {
	// It doesn't do anything since stdin is closed.
	if err := Main(); err != nil {
		t.Fatal(err)
	}
}

//

var (
	// tmpBuildDir is initialized by testMain().
	tmpBuildDir string

	// panicPath is the path to github.com/maruel/panicparse/cmd/panic compiled.
	// Use getPanic() instead.
	panicPath     string
	panicPathOnce sync.Once

	data     []byte
	dataOnce sync.Once
)

func compareString(t *testing.T, expected, actual string) {
	helper(t)()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func compareLines(t *testing.T, expected, actual []string) {
	helper(t)()
	for i := 0; i < len(actual) && i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Fatalf("Different lines #%d:\n- %q\n- %q", i, expected[i], actual[i])
		}
	}
	if len(expected) != len(actual) {
		t.Fatalf("different length %d != %d", len(expected), len(actual))
	}
}

func getPanic(t *testing.T) string {
	panicPathOnce.Do(func() {
		if panicPath = build(); panicPath == "" {
			t.Fatal("building panic failed")
		}
	})
	return panicPath
}

func getReader(t *testing.T) io.Reader {
	dataOnce.Do(func() {
		data = execRun(getPanic(t), "simple")
	})
	return bytes.NewReader(data)
}

// execRun runs a command and returns the combined output.
//
// It ignores the exit code, since it's meant to run panic, which crashes by
// design.
func execRun(cmd ...string) []byte {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Env = append(os.Environ(), "GOTRACEBACK=all")
	out, _ := c.CombinedOutput()
	return out
}

func build() string {
	out := filepath.Join(tmpBuildDir, "panic")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	log.Printf("building %s", out)
	// Disable inlining otherwise the inlining varies between local execution and
	// remote execution. This can be observed as Elided being true without any
	// argument.
	args := []string{"build", "-gcflags", "-l", "-o", out}
	c := exec.Command("go", append(args, "../cmd/panic")...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return ""
	}
	return out
}

// TestMain manages a temporary directory to build on first use ../cmd/panic
// and clean up at the end.
func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	os.Setenv("GOTRACEBACK", "all")
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (exit int) {
	var err error
	tmpBuildDir, err = ioutil.TempDir("", "stack")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temporary directory: %v", err)
		return 1
	}
	defer func() {
		log.Printf("deleting %s", tmpBuildDir)
		if err := os.RemoveAll(tmpBuildDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to deletetemporary directory: %v", err)
			if exit == 0 {
				exit = 1
			}
		}
	}()
	return m.Run()
}
