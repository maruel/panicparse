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

func TestCallPkg(t *testing.T) {
	t.Parallel()
	c := newCall(
		"gopkg.in/yaml%2ev2.handleErr",
		Args{Values: []Arg{{Value: 0xc208033b20}}},
		"/gopath/src/gopkg.in/yaml.v2/yaml.go",
		153)
	// Call methods.
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go:153", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("yaml.v2", "yaml.go"), c.PkgSrc())
	compareString(t, "yaml.go:153", c.SrcLine())
	compareString(t, "yaml.go", c.SrcName())

	// Func methods.
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "handleErr", c.Func.Name())
	compareString(t, "yaml.v2.handleErr", c.Func.PkgDotName())
	compareString(t, "yaml.v2", c.Func.PkgName())
	compareString(t, "gopkg.in/yaml.v2.handleErr", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallPkgMethod(t *testing.T) {
	t.Parallel()
	c := newCall(
		"gopkg.in/yaml%2ev2.(*decoder).unmarshal",
		Args{Values: []Arg{{Value: 0xc208033b20}}},
		"/gopath/src/gopkg.in/yaml.v2/yaml.go",
		153)
	// Call methods.
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go:153", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("yaml.v2", "yaml.go"), c.PkgSrc())
	compareString(t, "yaml.go:153", c.SrcLine())
	compareString(t, "yaml.go", c.SrcName())

	// Func methods.
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "(*decoder).unmarshal", c.Func.Name())
	compareString(t, "yaml.v2.(*decoder).unmarshal", c.Func.PkgDotName())
	compareString(t, "yaml.v2", c.Func.PkgName())
	compareString(t, "gopkg.in/yaml.v2.(*decoder).unmarshal", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallPkgRemote(t *testing.T) {
	t.Parallel()
	c := newCall(
		"gopkg.in/yaml%2ev2.handleErr",
		Args{Values: []Arg{{Value: 0xc208033b20}}},
		"/remote/src/gopkg.in/yaml.v2/yaml.go",
		153)
	// Call methods.
	compareString(t, "/remote/src/gopkg.in/yaml.v2/yaml.go:153", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("yaml.v2", "yaml.go"), c.PkgSrc())
	compareString(t, "yaml.go:153", c.SrcLine())
	compareString(t, "yaml.go", c.SrcName())

	// Func methods.
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "handleErr", c.Func.Name())
	compareString(t, "yaml.v2.handleErr", c.Func.PkgDotName())
	compareString(t, "yaml.v2", c.Func.PkgName())
	compareString(t, "gopkg.in/yaml.v2.handleErr", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/remote": "/local"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/local/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallStdlib(t *testing.T) {
	t.Parallel()
	c := newCall(
		"reflect.Value.assignTo",
		Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
		"/goroot/src/reflect/value.go",
		2125)
	// Call methods.
	compareString(t, "/goroot/src/reflect/value.go:2125", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("reflect", "value.go"), c.PkgSrc())
	compareString(t, "value.go:2125", c.SrcLine())
	compareString(t, "value.go", c.SrcName())

	// Func methods.
	compareString(t, "", c.ImportPath())
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "Value.assignTo", c.Func.Name())
	compareString(t, "reflect.Value.assignTo", c.Func.PkgDotName())
	compareString(t, "reflect", c.Func.PkgName())
	compareString(t, "reflect.Value.assignTo", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "reflect", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/goroot/src/reflect/value.go", c.LocalSrcPath)
	compareString(t, "reflect/value.go", c.RelSrcPath)
}

func TestCallStdlibRemote(t *testing.T) {
	t.Parallel()
	c := newCall(
		"reflect.Value.assignTo",
		Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
		"/remote/src/reflect/value.go",
		2125)
	// Call methods.
	compareString(t, "/remote/src/reflect/value.go:2125", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("reflect", "value.go"), c.PkgSrc())
	compareString(t, "value.go:2125", c.SrcLine())
	compareString(t, "value.go", c.SrcName())

	// Func methods.
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "Value.assignTo", c.Func.Name())
	compareString(t, "reflect.Value.assignTo", c.Func.PkgDotName())
	compareString(t, "reflect", c.Func.PkgName())
	compareString(t, "reflect.Value.assignTo", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/remote", "/local", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "reflect", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/local/src/reflect/value.go", c.LocalSrcPath)
	compareString(t, "reflect/value.go", c.RelSrcPath)
}

func TestCallMain(t *testing.T) {
	t.Parallel()
	c := newCall(
		"main.main",
		Args{},
		"/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go",
		428)
	// Call methods.
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go:428", c.FullSrcLine())
	compareBool(t, true, c.IsPkgMain())
	compareString(t, pathJoin("pp", "main.go"), c.PkgSrc())
	compareString(t, "main.go:428", c.SrcLine())
	compareString(t, "main.go", c.SrcName())

	// Func methods.
	compareString(t, "", c.ImportPath())
	compareBool(t, true, c.Func.IsExported())
	compareString(t, "main", c.Func.Name())
	compareString(t, "main.main", c.Func.PkgDotName())
	compareString(t, "main", c.Func.PkgName())
	compareString(t, "main.main", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/pp", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/pp/main.go", c.RelSrcPath)
}

func TestCallMismatched(t *testing.T) {
	t.Parallel()
	// See testPanicMismatched in context_test.go.
	c := newCall(
		"github.com/maruel/panicparse/cmd/panic/internal/incorrect.Panic",
		Args{},
		"/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
		7)
	// Call methods.
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go:7", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("incorrect", "correct.go"), c.PkgSrc())
	compareString(t, "correct.go:7", c.SrcLine())
	compareString(t, "correct.go", c.SrcName())

	// Func methods.
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect", c.ImportPath())
	compareBool(t, true, c.Func.IsExported())
	compareString(t, "Panic", c.Func.Name())
	compareString(t, "incorrect.Panic", c.Func.PkgDotName())
	compareString(t, "incorrect", c.Func.PkgName())
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect.Panic", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go", c.RelSrcPath)
}

func TestCallUTF8(t *testing.T) {
	t.Parallel()
	// See testPanicUTF8 in context_test.go.
	c := newCall(
		"github.com/maruel/panicparse/cmd/panic/internal/%c3%b9tf8.(*Strùct).Pànic",
		Args{Values: []Arg{{Value: 0xc0000b2e48}}},
		"/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
		10)
	// Call methods.
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go:10", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("ùtf8", "ùtf8.go"), c.PkgSrc())
	compareString(t, "ùtf8.go:10", c.SrcLine())
	compareString(t, "ùtf8.go", c.SrcName())

	// Func methods.
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8", c.ImportPath())
	compareBool(t, true, c.Func.IsExported())
	compareString(t, "(*Strùct).Pànic", c.Func.Name())
	compareString(t, "ùtf8.(*Strùct).Pànic", c.Func.PkgDotName())
	compareString(t, "ùtf8", c.Func.PkgName())
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8.(*Strùct).Pànic", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go", c.RelSrcPath)
}

func TestCallC(t *testing.T) {
	t.Parallel()
	c := newCall(
		"findrunnable",
		Args{Values: []Arg{{Value: 0xc208012000}}},
		"/goroot/src/runtime/proc.c",
		1472)
	// Call methods.
	compareString(t, "/goroot/src/runtime/proc.c:1472", c.FullSrcLine())
	compareBool(t, false, c.IsPkgMain())
	compareString(t, pathJoin("runtime", "proc.c"), c.PkgSrc())
	compareString(t, "proc.c:1472", c.SrcLine())
	compareString(t, "proc.c", c.SrcName())

	// Func methods.
	compareString(t, "", c.ImportPath())
	compareBool(t, false, c.Func.IsExported())
	compareString(t, "findrunnable", c.Func.Name())
	compareString(t, "", c.Func.PkgName())
	compareString(t, "findrunnable", c.Func.String())

	// ParseDump(guesspaths=true).
	c.updateLocations("/goroot", "/goroot", "/gomod", "example.com/foo", map[string]string{"/gopath": "/gopath"})
	compareString(t, "runtime", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/goroot/src/runtime/proc.c", c.LocalSrcPath)
	compareString(t, "runtime/proc.c", c.RelSrcPath)
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

func TestFuncAnonymous(t *testing.T) {
	t.Parallel()
	f := Func{Raw: "main.func·001"}
	compareString(t, "main.func·001", f.String())
	compareString(t, "main.func·001", f.PkgDotName())
	compareString(t, "func·001", f.Name())
	compareString(t, "main", f.PkgName())
	compareBool(t, false, f.IsExported())
}

func TestFuncGC(t *testing.T) {
	t.Parallel()
	f := Func{Raw: "gc"}
	compareString(t, "gc", f.String())
	compareString(t, "gc", f.PkgDotName())
	compareString(t, "gc", f.Name())
	compareString(t, "", f.PkgName())
	compareBool(t, false, f.IsExported())
}

func TestSignature(t *testing.T) {
	t.Parallel()
	s := getSignature()
	compareString(t, "", s.SleepString())
	s.SleepMax = 10
	compareString(t, "0~10 minutes", s.SleepString())
	s.SleepMin = 10
	compareString(t, "10 minutes", s.SleepString())
	compareString(t, "", s.CreatedByString(true))
	s.CreatedBy = newCall(
		"DoStuff",
		Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
		"/gopath/src/foo/bar.go",
		72)
	compareString(t, "DoStuff @ bar.go:72", s.CreatedByString(false))
	compareString(t, "DoStuff @ /gopath/src/foo/bar.go:72", s.CreatedByString(true))
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
	if s1.less(s2) {
		t.Fatal("less")
	}
	s2.State = "foo"
	if !s1.less(s2) {
		t.Fatal("not less")
	}
}

//

var (
	goroot     string
	gopaths    map[string]string
	gomod      string
	goimport   string
	isInGOPATH bool
)

func init() {
	goroot = runtime.GOROOT()
	gopaths = map[string]string{}
	for _, p := range getGOPATHs() {
		gopaths[p] = p
	}

	// Assumes pwd == this directory.
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Our internal functions work with '/' as path separator.
	pwd = strings.Replace(pwd, "\\", "/", -1)
	gomod, goimport = isGoModule(splitPath(pwd))

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
	return Func{Raw: s}
}

func newCall(f string, a Args, s string, l int) Call {
	return Call{Func: newFunc(f), Args: a, SrcPath: s, Line: l}
}

func newCallLocal(f string, a Args, s string, l int) Call {
	c := newCall(f, a, s, l)
	c.updateLocations(goroot, goroot, gomod, goimport, gopaths)
	if c.LocalSrcPath == "" || c.RelSrcPath == "" {
		panic(fmt.Sprintf("newCallLocal(%q, %q): invariant failed; gomod=%q, goimport=%q, GOPATHs=%v", f, s, gomod, goimport, gopaths))
	}
	return c
}

func compareBool(t *testing.T, want, got bool) {
	helper(t)()
	if want != got {
		t.Fatalf("%t != %t", want, got)
	}
}

func compareErr(t *testing.T, want, got error) {
	helper(t)()
	if got == nil || want.Error() != got.Error() {
		t.Fatalf("%v != %v", want, got)
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
	if a.CreatedBy.Line != 0 && b.CreatedBy.Line != 0 {
		a.CreatedBy.Line = 42
		b.CreatedBy.Line = 42
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

func getSignature() *Signature {
	return &Signature{
		State: "chan receive",
		Stack: Stack{
			Calls: []Call{
				{
					Func:    newFunc("main.func·001"),
					Args:    Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    72,
				},
				{
					Func:     newFunc("sliceInternal"),
					Args:     Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					SrcPath:  "/golang/src/sort/slices.go",
					Line:     72,
					IsStdlib: true,
				},
				{
					Func:     newFunc("Slice"),
					Args:     Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					SrcPath:  "/golang/src/sort/slices.go",
					Line:     72,
					IsStdlib: true,
				},
				{
					Func:    newFunc("DoStuff"),
					Args:    Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
					SrcPath: "/gopath/src/foo/bar.go",
					Line:    72,
				},
				{
					Func: newFunc("doStuffInternal"),
					Args: Args{
						Values: []Arg{{Value: 0x11000000}, {Value: 2}},
						Elided: true,
					},
					SrcPath: "/gopath/src/foo/bar.go",
					Line:    72,
				},
			},
		},
	}
}

func compareCalls(t *testing.T, want, got *Call) {
	helper(t)()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Call mismatch (-want +got):\n%s", diff)
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
