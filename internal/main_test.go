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
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/panicparse/internal/internaltest"
	"github.com/maruel/panicparse/stack"
)

func TestProcess(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	if err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, nil); err != nil {
		t.Fatal(err)
	}
	want := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, want, out.String())
}

func TestProcessFullPath(t *testing.T) {
	t.Parallel()
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
	want := fmt.Sprintf("GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain F%s:52 ImainL()A\n", p)
	compareString(t, want, out.String())
}

func TestProcessNoColor(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	if err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, nil); err != nil {
		t.Fatal(err)
	}
	want := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, want, out.String())
}

func TestProcessMatch(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", nil, regexp.MustCompile(`notpresent`))
	if err != nil {
		t.Fatal(err)
	}
	want := "GOTRACEBACK=all\npanic: simple\n\n"
	compareString(t, want, out.String())
}

func TestProcessFilter(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	err := process(getReader(t), out, testPalette, stack.AnyPointer, basePath, false, true, "", regexp.MustCompile(`notpresent`), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:52 ImainL()A\n"
	compareString(t, want, out.String())
}

func TestMainFn(t *testing.T) {
	t.Parallel()
	// It doesn't do anything since stdin is closed.
	if err := Main(); err != nil {
		t.Fatal(err)
	}
}

//

func compareString(t *testing.T, want, got string) {
	helper(t)()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func getReader(t *testing.T) io.Reader {
	return bytes.NewReader(internaltest.PanicOutputs()["simple"])
}

// TestMain manages a temporary directory to build on first use ../cmd/panic
// and clean up at the end.
func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	// Set the environment variable so the stack doesn't include the info header.
	os.Setenv("GOTRACEBACK", "all")
	os.Exit(m.Run())
}
