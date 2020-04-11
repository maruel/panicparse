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
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAugment(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		input string
		// Starting with go1.11, the stack trace do not contain much information
		// about the arguments and shows as elided.
		workaroundGo111Elided bool
		// Starting with go1.11, non-pointer call shows an elided argument, while
		// there was no argument listed before.
		workaroundGo111Extra bool
		want                 Stack
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 2}}},
						"main.go",
						10),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 1}, {Value: 1}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 5}, {Value: 7}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 5}, {Value: 7}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 1}, {Value: 1}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			true,
			Stack{
				Calls: []Call{
					newCall(
						"main.S.f",
						Args{},
						"main.go",
						8),
					newCall("main.main", Args{}, "main.go", 4),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.(*S).f",
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						8),
					newCall("main.main", Args{}, "main.go", 4),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 3}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: 3}, {Value: 42}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{
							Values: []Arg{
								{}, {}, {}, {}, {}, {}, {}, {}, {Value: 42}, {Value: 43},
							},
							Elided: true,
						},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: pointer}}},
						"main.go",
						7),
					newCall("main.main", Args{}, "main.go", 4),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						Args{Values: []Arg{{Value: pointer}, {Value: pointer}}},
						"main.go",
						7),
					newCall("main.main", Args{}, "main.go", 4),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
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
			false,
			Stack{
				Calls: []Call{
					newCall(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{Values: []Arg{{Value: pointer}}},
						"main.go",
						6),
					newCall("main.main", Args{}, "main.go", 3),
				},
			},
		},
	}

	for _, line := range data {
		line := line
		t.Run(line.name, func(t *testing.T) {
			t.Parallel()
			// Marshal the code a bit to make it nicer. Inject 'package main'.
			lines := append([]string{"package main"}, strings.Split(line.input, "\n")...)
			for i := 2; i < len(lines); i++ {
				// Strip the 3 first tab characters. It's very adhoc but good enough here
				// and makes test failure much more readable.
				if lines[i][:3] != "\t\t\t" {
					t.Fatal("expected line to start with 3 tab characters")
				}
				lines[i] = lines[i][3:]
			}
			input := strings.Join(lines, "\n")

			// Run the command.
			_, content, clean := getCrash(t, input)

			// Analyze it.
			extra := bytes.Buffer{}
			c, err := ParseDump(bytes.NewBuffer(content), &extra, false)
			if err != nil {
				clean()
				t.Fatalf("failed to parse input for test %s: %v", line.name, err)
			}
			// On go1.4, there's one less space.
			if got := extra.String(); got != "panic: ooh\n\nexit status 2\n" && got != "panic: ooh\nexit status 2\n" {
				clean()
				t.Fatalf("Unexpected panic output:\n%#v", got)
			}

			// On go1.11 with non-pointer method, it shows elided argument where there
			// used to be none before. It's only for test case "non-pointer method".
			if line.workaroundGo111Extra && zapArguments() {
				line.want.Calls[0].Args.Elided = true
			}

			s := c.Goroutines[0].Signature.Stack
			zapPointers(t, line.name, line.workaroundGo111Elided, &line.want, &s)
			zapPaths(&s)
			clean()
			if diff := cmp.Diff(line.want, s); diff != "" {
				t.Logf("Different (-want +got):\n%s", diff)
				t.Logf("Source code:\n%s", input)
				t.Logf("Output:\n%s", content)
				t.FailNow()
			}
		})
	}
}

func TestAugmentDummy(t *testing.T) {
	t.Parallel()
	goroutines := []*Goroutine{
		{
			Signature: Signature{
				Stack: Stack{
					Calls: []Call{{SrcPath: "missing.go"}},
				},
			},
		},
	}
	Augment(goroutines)
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
	if c.getFuncAST(&Call{SrcPath: "other"}) != nil {
		t.Fatalf("there's no 'other'")
	}
}

//

const pointer = uint64(0xfffffffff)
const pointerStr = "0xfffffffff"

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

func getCrash(t *testing.T, content string) (string, []byte, func()) {
	helper(t)()
	//p := getGOPATHs()
	//name, err := ioutil.TempDir(filepath.Join(p[0], "src"), "panicparse")
	name, err := ioutil.TempDir("", "panicparse")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	clean := func() {
		if err := os.RemoveAll(name); err != nil {
			t.Fatalf("failed to remove temporary directory %q: %v", name, err)
		}
	}
	main := filepath.Join(name, "main.go")
	if err := ioutil.WriteFile(main, []byte(content), 0500); err != nil {
		clean()
		t.Fatalf("failed to write %q: %v", main, err)
	}
	cmd := exec.Command("go", "run", main)
	// Use the Go 1.4 compatible format.
	cmd.Env = overrideEnv(os.Environ(), "GOTRACEBACK", "1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		clean()
		t.Fatal("expected error since this is supposed to crash")
	}
	return main, out, clean
}

// zapPointers zaps out pointers.
func zapPointers(t *testing.T, name string, workaroundGo111Elided bool, want, s *Stack) {
	helper(t)()
	for i := range s.Calls {
		if i >= len(want.Calls) {
			// When using GOTRACEBACK=2, it'll include runtime.main() and
			// runtime.goexit(). Ignore these since they could be changed in a future
			// version.
			s.Calls = s.Calls[:len(want.Calls)]
			break
		}
		if workaroundGo111Elided && zapArguments() {
			// See https://github.com/maruel/panicparse/issues/42 for explanation.
			if len(want.Calls[i].Args.Values) != 0 {
				want.Calls[i].Args.Elided = true
			}
			want.Calls[i].Args.Values = nil
			continue
		}
		for j := range s.Calls[i].Args.Values {
			if j >= len(want.Calls[i].Args.Values) {
				break
			}
			if want.Calls[i].Args.Values[j].Value == pointer {
				// Replace the pointer value.
				if s.Calls[i].Args.Values[j].Value == 0 {
					t.Fatalf("%s: Call %d, value %d, expected pointer, got 0", name, i, j)
				}
				old := fmt.Sprintf("0x%x", s.Calls[i].Args.Values[j].Value)
				s.Calls[i].Args.Values[j].Value = pointer
				for k := range s.Calls[i].Args.Processed {
					s.Calls[i].Args.Processed[k] = strings.Replace(s.Calls[i].Args.Processed[k], old, pointerStr, -1)
				}
			}
		}
	}
}

// zapPaths removes the directory part and only keep the base file name.
func zapPaths(s *Stack) {
	for j := range s.Calls {
		s.Calls[j].SrcPath = filepath.Base(s.Calls[j].SrcPath)
		s.Calls[j].LocalSrcPath = ""
	}
}
