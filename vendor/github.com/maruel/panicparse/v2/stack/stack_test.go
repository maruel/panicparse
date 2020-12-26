// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFuncInit(t *testing.T) {
	t.Parallel()
	data := []struct {
		raw  string
		want Func
	}{
		{
			"github.com/maruel/panicparse/cmd/panic/internal/%c3%b9tf8.(*Strùct).Pànic",
			Func{
				Complete:   "github.com/maruel/panicparse/cmd/panic/internal/ùtf8.(*Strùct).Pànic",
				ImportPath: "github.com/maruel/panicparse/cmd/panic/internal/ùtf8",
				DirName:    "ùtf8",
				Name:       "(*Strùct).Pànic",
				IsExported: true,
			},
		},
		{
			"gopkg.in/yaml%2ev2.handleErr",
			Func{
				Complete:   "gopkg.in/yaml.v2.handleErr",
				ImportPath: "gopkg.in/yaml.v2",
				DirName:    "yaml.v2",
				Name:       "handleErr",
			},
		},
		{
			"github.com/maruel/panicparse/vendor/golang.org/x/sys/unix.Nanosleep",
			Func{
				Complete:   "github.com/maruel/panicparse/vendor/golang.org/x/sys/unix.Nanosleep",
				ImportPath: "github.com/maruel/panicparse/vendor/golang.org/x/sys/unix",
				DirName:    "unix",
				Name:       "Nanosleep",
				IsExported: true,
			},
		},
		{
			"main.func·001",
			Func{
				Complete:   "main.func·001",
				ImportPath: "main",
				DirName:    "main",
				Name:       "func·001",
				IsPkgMain:  true,
			},
		},
		{
			"gc",
			Func{
				Complete: "gc",
				Name:     "gc",
			},
		},
	}
	for _, line := range data {
		got := newFunc(line.raw)
		if diff := cmp.Diff(line.want, got); diff != "" {
			t.Fatalf("Call mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestCallPkg(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		f    string
		s    string
		// Expectations
		DirSrc       string
		SrcName      string
		LocalSrcPath string
		RelSrcPath   string
		ImportPath   string
		Location     Location
	}{
		{
			name:         "Pkg",
			f:            "gopkg.in/yaml%2ev2.handleErr",
			s:            "/gpremote/src/gopkg.in/yaml.v2/yaml.go",
			DirSrc:       pathJoin("yaml.v2", "yaml.go"),
			SrcName:      "yaml.go",
			LocalSrcPath: "/gplocal/src/gopkg.in/yaml.v2/yaml.go",
			RelSrcPath:   "gopkg.in/yaml.v2/yaml.go",
			ImportPath:   "gopkg.in/yaml.v2",
			Location:     GOPATH,
		},
		{
			name:         "PkgMod",
			f:            "gopkg.in/yaml%2ev2.handleErr",
			s:            "/gpremote/pkg/mod/gopkg.in/yaml.v2@v2.3.0/yaml.go",
			DirSrc:       pathJoin("yaml.v2@v2.3.0", "yaml.go"),
			SrcName:      "yaml.go",
			LocalSrcPath: "/gplocal/pkg/mod/gopkg.in/yaml.v2@v2.3.0/yaml.go",
			RelSrcPath:   "gopkg.in/yaml.v2@v2.3.0/yaml.go",
			ImportPath:   "gopkg.in/yaml.v2@v2.3.0",
			Location:     GoPkg,
		},
		{
			name:         "PkgMethod",
			f:            "gopkg.in/yaml%2ev2.(*decoder).unmarshal",
			s:            "/gpremote/src/gopkg.in/yaml.v2/yaml.go",
			DirSrc:       pathJoin("yaml.v2", "yaml.go"),
			SrcName:      "yaml.go",
			LocalSrcPath: "/gplocal/src/gopkg.in/yaml.v2/yaml.go",
			RelSrcPath:   "gopkg.in/yaml.v2/yaml.go",
			ImportPath:   "gopkg.in/yaml.v2",
			Location:     GOPATH,
		},
		{
			name:         "Stdlib",
			f:            "reflect.Value.assignTo",
			s:            "/grremote/src/reflect/value.go",
			DirSrc:       pathJoin("reflect", "value.go"),
			SrcName:      "value.go",
			LocalSrcPath: "/grlocal/src/reflect/value.go",
			RelSrcPath:   "reflect/value.go",
			ImportPath:   "reflect",
			Location:     Stdlib,
		},
		{
			name:         "Main",
			f:            "main.main",
			s:            "/gpremote/src/github.com/maruel/panicparse/cmd/pp/main.go",
			DirSrc:       pathJoin("pp", "main.go"),
			SrcName:      "main.go",
			LocalSrcPath: "/gplocal/src/github.com/maruel/panicparse/cmd/pp/main.go",
			RelSrcPath:   "github.com/maruel/panicparse/cmd/pp/main.go",
			ImportPath:   "github.com/maruel/panicparse/cmd/pp",
			Location:     GOPATH,
		},
		{
			// See testPanicMismatched in context_test.go.
			name:         "Mismatched",
			f:            "github.com/maruel/panicparse/cmd/panic/internal/incorrect.Panic",
			s:            "/gpremote/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
			DirSrc:       pathJoin("incorrect", "correct.go"),
			SrcName:      "correct.go",
			LocalSrcPath: "/gplocal/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
			RelSrcPath:   "github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
			ImportPath:   "github.com/maruel/panicparse/cmd/panic/internal/incorrect",
			Location:     GOPATH,
		},
		{
			// See testPanicUTF8 in context_test.go.
			name:         "UTF8",
			f:            "github.com/maruel/panicparse/cmd/panic/internal/%c3%b9tf8.(*Strùct).Pànic",
			s:            "/gpremote/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
			DirSrc:       pathJoin("ùtf8", "ùtf8.go"),
			SrcName:      "ùtf8.go",
			LocalSrcPath: "/gplocal/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
			RelSrcPath:   "github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
			ImportPath:   "github.com/maruel/panicparse/cmd/panic/internal/ùtf8",
			Location:     GOPATH,
		},
		{
			name:         "C",
			f:            "findrunnable",
			s:            "/grremote/src/runtime/proc.c",
			DirSrc:       pathJoin("runtime", "proc.c"),
			SrcName:      "proc.c",
			LocalSrcPath: "/grlocal/src/runtime/proc.c",
			RelSrcPath:   "runtime/proc.c",
			ImportPath:   "runtime",
			Location:     Stdlib,
		},
		{
			name:         "Gomod",
			f:            "example.com/foo/bar.Func",
			s:            "/gomod/bar/baz.go",
			DirSrc:       "bar/baz.go",
			SrcName:      "baz.go",
			LocalSrcPath: "/gomod/bar/baz.go",
			RelSrcPath:   "bar/baz.go",
			ImportPath:   "example.com/foo/bar",
			Location:     GoMod,
		},
		{
			name:         "GomodMain",
			f:            "main.main",
			s:            "/gomod/cmd/panic/main.go",
			DirSrc:       "panic/main.go",
			SrcName:      "main.go",
			LocalSrcPath: "/gomod/cmd/panic/main.go",
			RelSrcPath:   "cmd/panic/main.go",
			ImportPath:   "example.com/foo/cmd/panic",
			Location:     GoMod,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			c := newCall(line.f, Args{}, line.s, 153)
			compareString(t, line.DirSrc, c.DirSrc)
			compareString(t, line.SrcName, c.SrcName)
			// Equivalent of calling GuessPaths().
			gp := map[string]string{"/gpremote": "/gplocal"}
			gm := map[string]string{"/gomod": "example.com/foo"}
			if !c.updateLocations("/grremote", "/grlocal", gm, gp) {
				t.Error("Unexpected")
			}
			compareString(t, line.ImportPath, c.ImportPath)
			if line.Location != c.Location {
				t.Errorf("want %s, got %s", line.Location, c.Location)
			}
			compareString(t, line.LocalSrcPath, c.LocalSrcPath)
			compareString(t, line.RelSrcPath, c.RelSrcPath)
		})
	}
}

func TestArgs(t *testing.T) {
	t.Parallel()
	a := Args{
		Values: []Arg{
			{Value: 0x4},
			{Value: 0x7fff671c7118},
			{Value: 0xffffffff00000080},
			{},
			{Value: 0xffffffff0028c1be},
			{Name: "foo"},
			{},
			{},
			{},
			{},
		},
		Elided: true,
	}
	compareString(t, "4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, foo, 0, 0, 0, 0, ...", a.String())

	a = Args{Processed: []string{"yo"}}
	compareString(t, "yo", a.String())
}

func TestSignature(t *testing.T) {
	t.Parallel()
	s := getSignature()
	compareString(t, "", s.SleepString())
	s.SleepMax = 10
	compareString(t, "0~10 minutes", s.SleepString())
	s.SleepMin = 10
	compareString(t, "10 minutes", s.SleepString())
}

func TestSignature_Equal(t *testing.T) {
	t.Parallel()
	s1 := getSignature()
	s2 := getSignature()
	if !s1.equal(s2) {
		t.Fatal("equal")
	}
	s2.State = "foo"
	if s1.equal(s2) {
		t.Fatal("inequal")
	}
}

func TestSignature_Similar(t *testing.T) {
	t.Parallel()
	s1 := getSignature()
	s2 := getSignature()
	if !s1.similar(s2, ExactFlags) {
		t.Fatal("equal")
	}
	s2.State = "foo"
	if s1.similar(s2, ExactFlags) {
		t.Fatal("inequal")
	}
}

func TestSignature_Less(t *testing.T) {
	t.Parallel()
	s1 := getSignature()
	s2 := getSignature()
	if s1.less(s2) || s2.less(s1) {
		t.Fatal("less")
	}
	s2.State = "foo"
	if !s1.less(s2) || s2.less(s1) {
		t.Fatal("not less")
	}
	s2 = getSignature()
	s2.Stack.Calls = s2.Stack.Calls[:1]
	if !s1.less(s2) || s2.less(s1) {
		t.Fatal("not less")
	}
}

//

var (
	goroot     string
	gopaths    map[string]string
	gopath     string
	gomods     map[string]string
	isInGOPATH bool
)

func init() {
	goroot = strings.Replace(runtime.GOROOT(), "\\", "/", -1)
	gopaths = map[string]string{}
	for i, p := range getGOPATHs() {
		gopaths[p] = p
		if i == 0 {
			gopath = p
		}
	}

	// Assumes pwd == this directory.
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Our internal functions work with '/' as path separator.
	pwd = strings.Replace(pwd, "\\", "/", -1)
	gomods = map[string]string{}
	gmc := gomodCache{}
	if prefix, path := gmc.isGoModule(splitPath(pwd)); prefix != "" {
		gomods[prefix] = path
	}

	// When inside GOPATH, no version is added. When outside, the version path is
	// added from the reading of module statement in go.mod.
	for _, p := range getGOPATHs() {
		if strings.HasPrefix(pwd, p) {
			isInGOPATH = true
			break
		}
	}
}

func newFunc(s string) Func {
	f := Func{}
	if s != "" {
		if err := f.Init(s); err != nil {
			panic(err)
		}
	}
	return f
}

func newCall(f string, a Args, s string, l int) Call {
	c := Call{Func: newFunc(f), Args: a}
	c.init(s, l)
	return c
}

func newCallLocal(f string, a Args, s string, l int) Call {
	c := newCall(f, a, s, l)
	r := c.updateLocations(goroot, goroot, gomods, gopaths)
	if !r {
		panic("Unexpected")
	}
	if c.LocalSrcPath == "" || c.RelSrcPath == "" {
		panic(fmt.Sprintf("newCallLocal(%q, %q): invariant failed; gomods=%v, GOPATHs=%v", f, s, gomods, gopaths))
	}
	return c
}

func compareErr(t *testing.T, want, got error) {
	helper(t)()
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil {
		t.Errorf("want: %v, got: %v", want, got)
	} else if want.Error() != got.Error() {
		t.Errorf("want: %q, got: %q", want.Error(), got.Error())
	}
}

func compareString(t *testing.T, want, got string) {
	helper(t)()
	if want != got {
		t.Fatalf("%q != %q", want, got)
	}
}

// similarGoroutines compares slice of Goroutine to be similar enough.
//
// Warning: it mutates inputs.
func similarGoroutines(t *testing.T, want, got []*Goroutine) {
	helper(t)()
	zapGoroutines(t, want, got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Goroutine mismatch (-want +got):\n%s", diff)
	}
}

func zapGoroutines(t *testing.T, a, b []*Goroutine) {
	if len(a) != len(b) {
		t.Error("different []*Goroutine length")
		return
	}
	for i := range a {
		// &(*Goroutine).Signature
		zapSignatures(t, &a[i].Signature, &b[i].Signature)
	}
}

// similarSignatures compares Signature to be similar enough.
//
// Warning: it mutates inputs.
func similarSignatures(t *testing.T, want, got *Signature) {
	helper(t)()
	zapSignatures(t, want, got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Signature mismatch (-want +got):\n%s", diff)
	}
}

func zapSignatures(t *testing.T, a, b *Signature) {
	// Signature.Stack.([]Call)
	if len(a.Stack.Calls) != len(b.Stack.Calls) {
		t.Error("different call length")
		return
	}
	if len(a.CreatedBy.Calls) != 0 && len(b.CreatedBy.Calls) != 0 {
		if a.CreatedBy.Calls[0].Line != 0 && b.CreatedBy.Calls[0].Line != 0 {
			a.CreatedBy.Calls[0].Line = 42
			b.CreatedBy.Calls[0].Line = 42
		}
	}
	zapStacks(t, &a.Stack, &b.Stack)
}

func zapStacks(t *testing.T, a, b *Stack) {
	if len(a.Calls) != len(b.Calls) {
		t.Error("different Stack.[]Call length")
		return
	}
	for i := range a.Calls {
		zapCalls(t, &a.Calls[i], &b.Calls[i])
	}
}

func zapCalls(t *testing.T, a, b *Call) {
	if a.Line != 0 && b.Line != 0 {
		a.Line = 42
		b.Line = 42
	}
	zapArgs(t, &a.Args, &b.Args)
}

func zapArgs(t *testing.T, a, b *Args) {
	if len(a.Values) != len(b.Values) {
		t.Error("different Args.Values length")
		return
	}
	for i := range a.Values {
		if a.Values[i].Value != 0 && b.Values[i].Value != 0 {
			a.Values[i].Value = 42
			b.Values[i].Value = 42
		}
		if a.Values[i].Name != "" && b.Values[i].Name != "" {
			a.Values[i].Name = "foo"
			b.Values[i].Name = "foo"
		}
	}
}

func compareGoroutines(t *testing.T, want, got []*Goroutine) {
	helper(t)()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Goroutine mismatch (-want +got):\n%s", diff)
	}
}

func compareStacks(t *testing.T, want, got *Stack) {
	helper(t)()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Stack mismatch (-want +got):\n%s", diff)
	}
}

func getSignature() *Signature {
	return &Signature{
		State: "chan receive",
		Stack: Stack{
			Calls: []Call{
				{
					Func:          newFunc("main.func·001"),
					Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					RemoteSrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:          72,
				},
				{
					Func:          newFunc("sliceInternal"),
					Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					RemoteSrcPath: "/golang/src/sort/slices.go",
					Line:          72,
					Location:      Stdlib,
				},
				{
					Func:          newFunc("Slice"),
					Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					RemoteSrcPath: "/golang/src/sort/slices.go",
					Line:          72,
					Location:      Stdlib,
				},
				{
					Func:          newFunc("DoStuff"),
					Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					RemoteSrcPath: "/gopath/src/foo/bar.go",
					Line:          72,
				},
				{
					Func: newFunc("doStuffInternal"),
					Args: Args{
						Values: []Arg{{Value: 0x11000000}, {Value: 2}},
						Elided: true,
					},
					RemoteSrcPath: "/gopath/src/foo/bar.go",
					Line:          72,
				},
			},
		},
	}
}

// TestMain manages a temporary directory to build on first use ../cmd/panic
// and clean up at the end.
func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}
