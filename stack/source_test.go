// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAugment(t *testing.T) {
	t.Parallel()

	gm := map[string]string{"/root": "main"}
	newCallSrc := func(f string, a Args, s string, l int) Call {
		c := newCall(f, a, s, l)
		// Simulate findRoots().
		if !c.updateLocations(goroot, goroot, gm, gopaths) {
			t.Fatalf("c.updateLocations(%v, %v, %v, %v) failed on %s", goroot, goroot, gm, gopaths, s)
		}
		return c
	}

	data := []struct {
		name  string
		input string
		// Starting with go1.11, inlining is enabled. The stack trace may (it
		// depends on tool chain version) not contain much information about the
		// arguments and shows as elided. Non-pointer call may show an elided
		// argument, while there was no argument listed before.
		mayBeInlined bool
		want         Stack
	}{
		{
			"Local function doesn't interfere",
			`func main() {
				f("yo")
			}
			func f(s string) {
				a := func(i int) int {
					return 1 + i
				}
				_ = a(3)
				panic("ooh")
			}`,
			false,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 2}}},
						"/root/main.go",
						10),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"func",
			`func main() {
				f(func() string { return "ooh" })
			}
			func f(a func() string) {
				panic(a())
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"func ellipsis",
			`func main() {
				f(func() string { return "ooh" })
			}
			func f(a ...func() string) {
				panic(a[0]())
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"interface{}",
			`func main() {
				f(make([]interface{}, 5, 7))
			}
			func f(a []interface{}) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"[]int",
			`func main() {
				f(make([]int, 5, 7))
			}
			func f(a []int) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"[]interface{}",
			`func main() {
				f([]interface{}{"ooh"})
			}
			func f(a []interface{}) {
				panic(a[0].(string))
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"map[int]int",
			`func main() {
				f(map[int]int{1: 2})
			}
			func f(a map[int]int) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"map[interface{}]interface{}",
			`func main() {
				f(make(map[interface{}]interface{}))
			}
			func f(a map[interface{}]interface{}) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"chan int",
			`func main() {
				f(make(chan int))
			}
			func f(a chan int) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"chan interface{}",
			`func main() {
				f(make(chan interface{}))
			}
			func f(a chan interface{}) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"non-pointer method",
			`func main() {
				var s S
				s.f()
			}
			type S struct {}
			func (s S) f() {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.S.f",
						Args{},
						"/root/main.go",
						8),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"pointer method",
			`func main() {
				var s S
				s.f()
			}
			type S struct {}
			func (s *S) f() {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.(*S).f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						8),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"string",
			`func main() {
			  f("ooh")
			}
			func f(s string) {
				panic(s)
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 3}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"string and int",
			`func main() {
			  f("ooh", 42)
			}
			func f(s string, i int) {
				panic(s)
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 3}, {Value: 42}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"values are elided",
			`func main() {
				f(0, 0, 0, 0, 0, 0, 0, 0, 42, 43, 44, 45, nil)
			}
			func f(s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12 int, s13 interface{}) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{
								{}, {}, {}, {}, {}, {}, {}, {}, {Value: 42}, {Value: 43},
							},
							Elided: true,
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"error",
			`import "errors"
			func main() {
				f(errors.New("ooh"))
			}
			func f(err error) {
				panic(err.Error())
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: pointer, IsPtr: true}},
							Processed: []string{"0x2fffffff", "0x2fffffff"},
						},
						"/root/main.go",
						7),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"error unnamed",
			`import "errors"
			func main() {
				f(errors.New("ooh"))
			}
			func f(error) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: pointer, IsPtr: true}},
							Processed: []string{"0x2fffffff", "0x2fffffff"},
						},
						"/root/main.go",
						7),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"float32",
			`func main() {
				f(0.5)
			}
			func f(v float32) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"float64",
			`func main() {
				f(0.5)
			}
			func f(v float64) {
				panic("ooh")
			}`,
			true,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{Values: []Arg{{Value: pointer, IsPtr: true}}},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
	}

	for i, line := range data {
		i := i
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			// Marshal the code a bit to make it nicer. Inject 'package main'.
			lines := append([]string{"package main"}, strings.Split(line.input, "\n")...)
			for j := 2; j < len(lines); j++ {
				// Strip the 3 first tab characters. It's very adhoc but good enough here
				// and makes test failure much more readable.
				if lines[j][:3] != "\t\t\t" {
					t.Fatal("expected line to start with 3 tab characters")
				}
				lines[j] = lines[j][3:]
			}
			input := strings.Join(lines, "\n")

			// Create one temporary directory by subtest.
			root, err := ioutil.TempDir("", "stack")
			if err != nil {
				t.Fatalf("failed to create temporary directory: %v", err)
			}
			defer func() {
				if err2 := os.RemoveAll(root); err2 != nil {
					t.Fatalf("failed to remove temporary directory %q: %v", root, err2)
				}
			}()
			main := filepath.Join(root, "main.go")
			if err := ioutil.WriteFile(main, []byte(input), 0500); err != nil {
				t.Fatalf("failed to write %q: %v", main, err)
			}

			if runtime.GOOS == "windows" {
				root = strings.Replace(root, pathSeparator, "/", -1)
			}
			const prefix = "/root"
			for j, c := range line.want.Calls {
				if strings.HasPrefix(c.RemoteSrcPath, prefix) {
					line.want.Calls[j].RemoteSrcPath = root + c.RemoteSrcPath[len(prefix):]
				}
				if strings.HasPrefix(c.LocalSrcPath, prefix) {
					line.want.Calls[j].LocalSrcPath = root + c.LocalSrcPath[len(prefix):]
				}
				if strings.HasPrefix(c.DirSrc, "root") {
					line.want.Calls[j].DirSrc = path.Base(root) + c.DirSrc[4:]
				}
			}

			// Run the command up to twice.

			// Only disable inlining if necessary.
			disableInline := hasInlining && line.mayBeInlined
			content := getCrash(t, main, disableInline)
			t.Log("First")
			// Warning: this function modifies want.
			testAugmentCommon(t, content, false, line.want)

			// If inlining was disabled, try a second time but zap things out.
			if disableInline {
				for j := range line.want.Calls {
					line.want.Calls[j].Args.Processed = nil
				}
				content := getCrash(t, main, false)
				t.Log("Second")
				testAugmentCommon(t, content, line.mayBeInlined, line.want)
			}
		})
	}
}

func testAugmentCommon(t *testing.T, content []byte, mayBeInlined bool, want Stack) {
	// Analyze it.
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(bytes.NewBuffer(content), &prefix, DefaultOpts())
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}
	// On go1.4, there's one less empty line.
	if got := prefix.String(); got != "panic: ooh\n\n" && got != "panic: ooh\n" {
		t.Fatalf("Unexpected panic output:\n%#v", got)
	}
	compareString(t, "exit status 2\n", string(suffix))
	if !s.GuessPaths() {
		t.Error("expected success")
	}

	if err := Augment(s.Goroutines); err != nil {
		t.Errorf("Augment() returned %v", err)
	}
	got := s.Goroutines[0].Signature.Stack
	zapPointers(t, &want, &got)

	// On go1.11 with non-pointer method, it shows elided argument where
	// there used to be none before. It's only for test case "non-pointer
	// method".
	if mayBeInlined {
		for j := range got.Calls {
			if !want.Calls[j].Args.Elided {
				got.Calls[j].Args.Elided = false
			}
			if got.Calls[j].Args.Values == nil {
				want.Calls[j].Args.Values = nil
			}
		}
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Logf("Different (-want +got):\n%s", diff)
		t.Logf("Output:\n%s", content)
		t.FailNow()
	}
}

func TestAugmentDummy(t *testing.T) {
	t.Parallel()
	g := []*Goroutine{
		{
			Signature: Signature{
				Stack: Stack{
					Calls: []Call{{RemoteSrcPath: "missing.go"}},
				},
			},
		},
	}
	// There's no error because there's no Call with LocalSrcPath set.
	if err := Augment(g); err != nil {
		t.Error(err)
	}
	g[0].Stack.Calls[0].LocalSrcPath = "missing.go"
	if err := Augment(g); err == nil {
		t.Error("expected error")
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()
	c := &cache{
		files:  map[string][]byte{"bad.go": []byte("bad content")},
		parsed: map[string]*parsedFile{},
	}
	c.load("foo.asm")
	c.load("bad.go")
	c.load("doesnt_exist.go")
	if l := len(c.parsed); l != 3 {
		t.Fatalf("want 3, got %d", l)
	}
	if c.parsed["foo.asm"] != nil {
		t.Fatalf("foo.asm is not present; should not have been loaded")
	}
	if c.parsed["bad.go"] != nil {
		t.Fatalf("bad.go is not valid code; should not have been loaded")
	}
	if c.parsed["doesnt_exist.go"] != nil {
		t.Fatalf("doesnt_exist.go is not present; should not have been loaded")
	}
	if c.getFuncAST(&Call{RemoteSrcPath: "other"}) != nil {
		t.Fatalf("there's no 'other'")
	}
}

//

const pointer = uint64(0x2fffffff)
const pointerStr = "0x2fffffff"

func overrideEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func getCrash(t *testing.T, main string, disableInline bool) []byte {
	args := []string{"run"}
	if disableInline {
		args = append(args, "-gcflags", "-l")
	}
	cmd := exec.Command("go", append(args, main)...)
	// Use the Go 1.4 compatible format.
	cmd.Env = overrideEnv(os.Environ(), "GOTRACEBACK", "1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error since this is supposed to crash")
	}
	return out
}

// zapPointers zaps out pointers.
func zapPointers(t *testing.T, want, got *Stack) {
	helper(t)()
	for i := range got.Calls {
		if i >= len(want.Calls) {
			// When using GOTRACEBACK=2, it'll include runtime.main() and
			// runtime.goexit(). Ignore these since they could be changed in a future
			// version.
			got.Calls = got.Calls[:len(want.Calls)]
			break
		}
		for j := range got.Calls[i].Args.Values {
			if j >= len(want.Calls[i].Args.Values) {
				break
			}
			if want.Calls[i].Args.Values[j].Value == pointer {
				// Replace the pointer value.
				if got.Calls[i].Args.Values[j].Value == 0 {
					t.Fatalf("Call %d, value %d, expected pointer, got 0", i, j)
				}
				old := fmt.Sprintf("0x%x", got.Calls[i].Args.Values[j].Value)
				got.Calls[i].Args.Values[j].Value = pointer
				for k := range got.Calls[i].Args.Processed {
					got.Calls[i].Args.Processed[k] = strings.Replace(got.Calls[i].Args.Processed[k], old, pointerStr, -1)
				}
			}
		}
	}
}
