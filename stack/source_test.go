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

	"github.com/maruel/ut"
)

func TestAugment(t *testing.T) {
	data := []struct {
		name     string
		input    string
		expected Stack
	}{
		{
			"Local function doesn't interfere",
			`package main
			func f(s string) {
				a := func(i int) int {
					return 1 + i
				}
				_ = a(3)
				panic("ooh")
			}
			func main() {
				f("yo")
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 7, Func: Function{"main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer, Name: ""}, {Value: 0x2}},
						},
					},
					{SourcePath: "main.go", Line: 10, Func: Function{"main.main"}},
				},
			},
		},
		{
			"func",
			`package main
			func f(a func() string) {
				panic(a())
			}
			func main() {
				f(func() string { return "ooh" })
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{Values: []Arg{{Value: pointer}}},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"func elipsis",
			`package main
			func f(a ...func() string) {
				panic(a[0]())
			}
			func main() {
				f(func() string { return "ooh" })
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: 0x1}, {Value: 0x1}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"interface{}",
			`package main
			func f(a []interface{}) {
				panic("ooh")
			}
			func main() {
				f(make([]interface{}, 5, 7))
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: 0x5}, {Value: 0x7}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"[]int",
			`package main
			func f(a []int) {
				panic("ooh")
			}
			func main() {
				f(make([]int, 5, 7))
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: 5}, {Value: 7}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"[]interface{}",
			`package main
			func f(a []interface{}) {
				panic(a[0].(string))
			}
			func main() {
				f([]interface{}{"ooh"})
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: 1}, {Value: 1}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"non-pointer method",
			`package main
			type S struct {
			}
			func (s S) f() {
				panic("ooh")
			}
			func main() {
				var s S
				s.f()
			}`,
			Stack{
				Calls: []Call{
					{SourcePath: "main.go", Line: 5, Func: Function{Raw: "main.S.f"}},
					{SourcePath: "main.go", Line: 9, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"pointer method",
			`package main
			type S struct {
			}
			func (s *S) f() {
				panic("ooh")
			}
			func main() {
				var s S
				s.f()
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 5, Func: Function{Raw: "main.(*S).f"},
						Args: Args{Values: []Arg{{Value: pointer}}},
					},
					{SourcePath: "main.go", Line: 9, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"string",
			`package main
			func f(s string) {
				panic(s)
			}
			func main() {
			  f("ooh")
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{Values: []Arg{{Value: pointer}, {Value: 0x3}}},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"string and int",
			`package main
			func f(s string, i int) {
				panic(s)
			}
			func main() {
			  f("ooh", 42)
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{Values: []Arg{{Value: pointer}, {Value: 0x3}, {Value: 42}}},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"values are elided",
			`package main
			func f(s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12 int, s13 interface{}) {
				panic("ooh")
			}
			func main() {
				f(0, 0, 0, 0, 0, 0, 0, 0, 42, 43, 44, 45, nil)
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{}, {}, {}, {}, {}, {}, {}, {}, {Value: 42}, {Value: 43}},
							Elided: true,
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"error",
			`package main
			import "errors"
			func f(err error) {
				panic(err.Error())
			}
			func main() {
				f(errors.New("ooh"))
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 4, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: pointer}},
						},
					},
					{SourcePath: "main.go", Line: 7, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"error unnamed",
			`package main
			import "errors"
			func f(error) {
				panic("ooh")
			}
			func main() {
				f(errors.New("ooh"))
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 4, Func: Function{Raw: "main.f"},
						Args: Args{
							Values: []Arg{{Value: pointer}, {Value: pointer}},
						},
					},
					{SourcePath: "main.go", Line: 7, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"float32",
			`package main
			func f(v float32) {
				panic("ooh")
			}
			func main() {
				f(0.5)
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							// The value is NOT a pointer but floating point encoding is not
							// deterministic.
							Values: []Arg{{Value: pointer}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
		{
			"float64",
			`package main
			func f(v float64) {
				panic("ooh")
			}
			func main() {
				f(0.5)
			}`,
			Stack{
				Calls: []Call{
					{
						SourcePath: "main.go", Line: 3, Func: Function{Raw: "main.f"},
						Args: Args{
							// The value is NOT a pointer but floating point encoding is not
							// deterministic.
							Values: []Arg{{Value: pointer}},
						},
					},
					{SourcePath: "main.go", Line: 6, Func: Function{Raw: "main.main"}},
				},
			},
		},
	}

	for i, line := range data {
		extra := bytes.Buffer{}
		_, content := getCrash(t, line.input)
		goroutines, err := ParseDump(bytes.NewBuffer(content), &extra)
		if err != nil {
			t.Fatalf("failed to parse input for test %s: %v", line.name, err)
		}
		// On go1.4, there's one less space.
		actual := extra.String()
		if actual != "panic: ooh\n\nexit status 2\n" && actual != "panic: ooh\nexit status 2\n" {
			t.Fatalf("Unexpected panic output:\n%#v", actual)
		}
		s := goroutines[0].Signature.Stack
		t.Logf("Test: %v", line.name)
		zapPointers(t, line.name, &line.expected, &s)
		zapPaths(&s)
		ut.AssertEqualIndex(t, i, line.expected, s)
	}
}

func TestAugmentDummy(t *testing.T) {
	goroutines := []Goroutine{
		{
			Signature: Signature{
				Stack: Stack{
					Calls: []Call{{SourcePath: "missing.go"}},
				},
			},
		},
	}
	Augment(goroutines)
}

func TestLoad(t *testing.T) {
	c := &cache{
		files:  map[string][]byte{"bad.go": []byte("bad content")},
		parsed: map[string]*parsedFile{},
	}
	c.load("foo.asm")
	c.load("bad.go")
	c.load("doesnt_exist.go")
	if l := len(c.parsed); l != 3 {
		t.Fatalf("expected 3, got %d", l)
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
	if c.getFuncAST(&Call{SourcePath: "other"}) != nil {
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

func getCrash(t *testing.T, content string) (string, []byte) {
	name, err := ioutil.TempDir("", "panicparse")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(name); err != nil {
			t.Fatalf("failed to remove temporary directory %q: %v", name, err)
		}
	}()
	main := filepath.Join(name, "main.go")
	if err := ioutil.WriteFile(main, []byte(content), 0500); err != nil {
		t.Fatalf("failed to write %q: %v", main, err)
	}
	cmd := exec.Command("go", "run", main)
	// Use the Go 1.4 compatible format.
	cmd.Env = overrideEnv(os.Environ(), "GOTRACEBACK", "1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error since this is supposed to crash")
	}
	return main, out
}

// zapPointers zaps out pointers.
func zapPointers(t *testing.T, name string, expected, s *Stack) {
	for i := range s.Calls {
		if i >= len(expected.Calls) {
			// When using GOTRACEBACK=2, it'll include runtime.main() and
			// runtime.goexit(). Ignore these since they could be changed in a future
			// version.
			s.Calls = s.Calls[:len(expected.Calls)]
			break
		}
		for j := range s.Calls[i].Args.Values {
			if j >= len(expected.Calls[i].Args.Values) {
				break
			}
			if expected.Calls[i].Args.Values[j].Value == pointer {
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
		s.Calls[j].SourcePath = filepath.Base(s.Calls[j].SourcePath)
	}
}
