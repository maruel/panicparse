// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/panicparse/internal/internaltest"
	"github.com/maruel/panicparse/v2/stack"
)

func TestProcess(t *testing.T) {
	t.Parallel()
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data := []struct {
		name    string
		palette *Palette
		simil   stack.Similarity
		path    pathFormat
		filter  *regexp.Regexp
		match   *regexp.Regexp
		want    string
	}{
		{
			name:    "BasePath",
			palette: testPalette,
			simil:   stack.AnyPointer,
			path:    basePath,
			want:    "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:70 GmainR()A\n",
		},
		{
			name:    "FullPath",
			palette: testPalette,
			simil:   stack.AnyValue,
			path:    fullPath,
			// "/" is used even on Windows.
			want: fmt.Sprintf("GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain F%s:70 GmainR()A\n", strings.Replace(filepath.Join(filepath.Dir(d), "cmd", "panic", "main.go"), "\\", "/", -1)),
		},
		{
			name:    "NoColor",
			palette: &Palette{},
			simil:   stack.AnyValue,
			path:    basePath,
			want:    "GOTRACEBACK=all\npanic: simple\n\n1: running\n    main main.go:70 main()\n",
		},
		{
			name:    "Match",
			palette: testPalette,
			simil:   stack.AnyValue,
			path:    basePath,
			match:   regexp.MustCompile(`notpresent`),
			want:    "GOTRACEBACK=all\npanic: simple\n\n",
		},
		{
			name:    "Filter",
			palette: testPalette,
			simil:   stack.AnyValue,
			path:    basePath,
			filter:  regexp.MustCompile(`notpresent`),
			want:    "GOTRACEBACK=all\npanic: simple\n\nC1: runningA\n    Emain Fmain.go:70 GmainR()A\n",
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			out := bytes.Buffer{}
			r := bytes.NewReader(internaltest.PanicOutputs()["simple"])
			if err := process(r, &out, line.palette, line.simil, line.path, false, true, "", line.filter, line.match); err != nil {
				t.Fatal(err)
			}
			compareString(t, line.want, out.String())
		})
	}
}

func TestProcessTwoSnapshots(t *testing.T) {
	t.Parallel()
	out := bytes.Buffer{}
	in := bytes.Buffer{}
	in.WriteString("Ya\n")
	in.Write(internaltest.PanicOutputs()["simple"])
	in.WriteString("Ye\n")
	in.Write(internaltest.PanicOutputs()["int"])
	in.WriteString("Yo\n")
	err := process(&in, &out, &Palette{}, stack.AnyPointer, basePath, false, true, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := ("Ya\n" +
		"GOTRACEBACK=all\n" +
		"panic: simple\n\n" +
		"1: running\n" +
		"    main main.go:70 main()\n" +
		"Ye\n" +
		"GOTRACEBACK=all\n" +
		"panic: 42\n\n" +
		"1: running\n" +
		"    main main.go:89  panicint(0x2a)\n" +
		"    main main.go:287 glob..func7()\n" +
		"    main main.go:72  main()\n" +
		"Yo\n")
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

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	// Set the environment variable so the stack doesn't include the info header.
	os.Setenv("GOTRACEBACK", "all")
	os.Exit(m.Run())
}
