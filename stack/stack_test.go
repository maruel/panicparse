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
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCallPkg(t *testing.T) {
	c := Call{
		SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:    153,
		Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
		Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallPkgMethod(t *testing.T) {
	c := Call{
		SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:    153,
		Func:    Func{Raw: "gopkg.in/yaml%2ev2.(*decoder).unmarshal"},
		Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallPkgRemote(t *testing.T) {
	c := Call{
		SrcPath: "/remote/src/gopkg.in/yaml.v2/yaml.go",
		Line:    153,
		Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
		Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/remote": "/local"})
	compareString(t, "gopkg.in/yaml.v2", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/local/src/gopkg.in/yaml.v2/yaml.go", c.LocalSrcPath)
	compareString(t, "gopkg.in/yaml.v2/yaml.go", c.RelSrcPath)
}

func TestCallStdlib(t *testing.T) {
	c := Call{
		SrcPath: "/goroot/src/reflect/value.go",
		Line:    2125,
		Func:    Func{Raw: "reflect.Value.assignTo"},
		Args:    Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "reflect", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/goroot/src/reflect/value.go", c.LocalSrcPath)
	compareString(t, "reflect/value.go", c.RelSrcPath)
}

func TestCallStdlibRemote(t *testing.T) {
	c := Call{
		SrcPath: "/remote/src/reflect/value.go",
		Line:    2125,
		Func:    Func{Raw: "reflect.Value.assignTo"},
		Args:    Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
	}
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
	c.updateLocations("/remote", "/local", map[string]string{"/gopath": "/gopath"})
	compareString(t, "reflect", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/local/src/reflect/value.go", c.LocalSrcPath)
	compareString(t, "reflect/value.go", c.RelSrcPath)
}

func TestCallMain(t *testing.T) {
	c := Call{
		SrcPath: "/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go",
		Line:    428,
		Func:    Func{Raw: "main.main"},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/pp", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/pp/main.go", c.RelSrcPath)
}

func TestCallMismatched(t *testing.T) {
	// See testPanicMismatched in context_test.go.
	c := Call{
		SrcPath:      "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
		LocalSrcPath: "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go",
		Line:         7,
		Func:         Func{Raw: "github.com/maruel/panicparse/cmd/panic/internal/incorrect.Panic"},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/incorrect/correct.go", c.RelSrcPath)
}

func TestCallUTF8(t *testing.T) {
	// See testPanicUTF8 in context_test.go.
	c := Call{
		SrcPath:      "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
		LocalSrcPath: "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go",
		Line:         10,
		Func:         Func{Raw: "github.com/maruel/panicparse/cmd/panic/internal/%c3%b9tf8.(*Strùct).Pànic"},
		Args:         Args{Values: []Arg{{Value: 0xc0000b2e48}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8", c.ImportPath())
	compareBool(t, false, c.IsStdlib)
	compareString(t, "/gopath/src/github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go", c.LocalSrcPath)
	compareString(t, "github.com/maruel/panicparse/cmd/panic/internal/ùtf8/ùtf8.go", c.RelSrcPath)
}

func TestCallC(t *testing.T) {
	c := Call{
		SrcPath: "/goroot/src/runtime/proc.c",
		Line:    1472,
		Func:    Func{Raw: "findrunnable"},
		Args:    Args{Values: []Arg{{Value: 0xc208012000}}},
	}
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
	c.updateLocations("/goroot", "/goroot", map[string]string{"/gopath": "/gopath"})
	compareString(t, "runtime", c.ImportPath())
	compareBool(t, true, c.IsStdlib)
	compareString(t, "/goroot/src/runtime/proc.c", c.LocalSrcPath)
	compareString(t, "runtime/proc.c", c.RelSrcPath)
}

func TestArgs(t *testing.T) {
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
	compareString(t, "0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, foo, 0, 0, 0, 0, ...", a.String())

	a = Args{Processed: []string{"yo"}}
	compareString(t, "yo", a.String())
}

func TestFuncAnonymous(t *testing.T) {
	f := Func{Raw: "main.func·001"}
	compareString(t, "main.func·001", f.String())
	compareString(t, "main.func·001", f.PkgDotName())
	compareString(t, "func·001", f.Name())
	compareString(t, "main", f.PkgName())
	compareBool(t, false, f.IsExported())
}

func TestFuncGC(t *testing.T) {
	f := Func{Raw: "gc"}
	compareString(t, "gc", f.String())
	compareString(t, "gc", f.PkgDotName())
	compareString(t, "gc", f.Name())
	compareString(t, "", f.PkgName())
	compareBool(t, false, f.IsExported())
}

func TestSignature(t *testing.T) {
	s := getSignature()
	compareString(t, "", s.SleepString())
	s.SleepMax = 10
	compareString(t, "0~10 minutes", s.SleepString())
	s.SleepMin = 10
	compareString(t, "10 minutes", s.SleepString())
	compareString(t, "", s.CreatedByString(true))
	s.CreatedBy = Call{
		SrcPath: "/gopath/src/foo/bar.go",
		Line:    72,
		Func:    Func{Raw: "DoStuff"},
		Args:    Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
	}
	compareString(t, "DoStuff @ bar.go:72", s.CreatedByString(false))
	compareString(t, "DoStuff @ /gopath/src/foo/bar.go:72", s.CreatedByString(true))
}

func TestSignature_Equal(t *testing.T) {
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

func compareBool(t *testing.T, expected, actual bool) {
	helper(t)()
	if expected != actual {
		t.Fatalf("%t != %t", expected, actual)
	}
}

func compareErr(t *testing.T, expected, actual error) {
	helper(t)()
	if actual == nil || expected.Error() != actual.Error() {
		t.Fatalf("%v != %v", expected, actual)
	}
}

func compareString(t *testing.T, expected, actual string) {
	helper(t)()
	if expected != actual {
		t.Fatalf("%q != %q", expected, actual)
	}
}

// similarGoroutines compares goroutines to be similar enough.
//
// Warning: it mutates inputs.
func similarGoroutines(t *testing.T, expected, actual []*Goroutine) {
	helper(t)()
	zapGoroutines(t, expected, actual)
	if diff := cmp.Diff(expected, actual); diff != "" {
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

func zapSignatures(t *testing.T, a, b *Signature) {
	// Signature.Stack.([]Call)
	if len(a.Stack.Calls) != len(b.Stack.Calls) {
		t.Error("different call length")
		return
	}
	zapStacks(t, &a.Stack, &b.Stack)
}

func zapStacks(t *testing.T, a, b *Stack) {
	if len(a.Calls) != len(b.Calls) {
		t.Error("different Stack.[]Call length")
		return
	}
	for i := range a.Calls {
		if a.Calls[i].Line != 0 && b.Calls[i].Line != 0 {
			a.Calls[i].Line = 42
			b.Calls[i].Line = 42
		}
		zapArgs(t, &a.Calls[i].Args, &b.Calls[i].Args)
	}
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
	}
}

func compareGoroutines(t *testing.T, expected, actual []*Goroutine) {
	helper(t)()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("Goroutine mismatch (-want +got):\n%s", diff)
	}
}

func compareSignatures(t *testing.T, expected, actual *Signature) {
	helper(t)()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("Signature mismatch (-want +got):\n%s", diff)
	}
}

func getSignature() *Signature {
	return &Signature{
		State: "chan receive",
		Stack: Stack{
			Calls: []Call{
				{
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    72,
					Func:    Func{Raw: "main.func·001"},
					Args:    Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
				},
				{
					SrcPath:  "/golang/src/sort/slices.go",
					Line:     72,
					Func:     Func{Raw: "sliceInternal"},
					Args:     Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
					IsStdlib: true,
				},
				{
					SrcPath:  "/golang/src/sort/slices.go",
					Line:     72,
					Func:     Func{Raw: "Slice"},
					Args:     Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
					IsStdlib: true,
				},
				{
					SrcPath: "/gopath/src/foo/bar.go",
					Line:    72,
					Func:    Func{Raw: "DoStuff"},
					Args:    Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
				},
				{
					SrcPath: "/gopath/src/foo/bar.go",
					Line:    72,
					Func:    Func{Raw: "doStuffInternal"},
					Args: Args{
						Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}},
						Elided: true,
					},
				},
			},
		},
	}
}

var (
	// tmpBuildDir is initialized by testMain().
	tmpBuildDir string

	// panicPath is the path to github.com/maruel/panicparse/cmd/panic compiled.
	// Use getPanic() instead.
	panicPath     string
	panicPathOnce sync.Once

	// panicRacePath is the path to github.com/maruel/panicparse/cmd/panic
	// compiled with -race.
	// Use getPanicRace() instead.
	panicRacePath     string
	panicRacePathOnce sync.Once

	// panicwebPath is the path to github.com/maruel/panicparse/cmd/panicweb
	// compiled.
	// Use getPanicweb() instead.
	panicwebPath     string
	panicwebPathOnce sync.Once
)

func getPanic(t *testing.T) string {
	panicPathOnce.Do(func() {
		if panicPath = build("panic", false); panicPath == "" {
			t.Fatal("building panic failed")
		}
	})
	return panicPath
}

func getPanicRace(t *testing.T) string {
	panicRacePathOnce.Do(func() {
		if panicRacePath = build("panic", true); panicRacePath == "" {
			t.Fatal("building panic with race detector failed")
		}
	})
	return panicRacePath
}

func getPanicweb(t *testing.T) string {
	panicwebPathOnce.Do(func() {
		if panicwebPath = build("panicweb", false); panicwebPath == "" {
			t.Fatal("building panicweb failed")
		}
	})
	return panicwebPath
}

// TestMain manages a temporary directory to build on first use ../cmd/panic
// and clean up at the end.
func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

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

func build(s string, race bool) string {
	out := filepath.Join(tmpBuildDir, s)
	if race {
		out += "_race"
	}
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	log.Printf("building %s", out)
	// Disable inlining otherwise the inlining varies between local execution and
	// remote execution. This can be observed as Elided being true without any
	// argument.
	args := []string{"build", "-gcflags", "-l", "-o", out}
	if race {
		args = append(args, "-race")
	}
	c := exec.Command("go", append(args, "../cmd/"+s)...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return ""
	}
	return out
}
