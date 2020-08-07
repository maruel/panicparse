// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"errors"
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
	"github.com/maruel/panicparse/v2/internal/internaltest"
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 2}},
							Processed: []string{"string(0x2fffffff, len=2)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"func(0x2fffffff)"},
						},
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
						Args{
							Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}},
							// TODO(maruel): Handle.
							Processed: []string{"<unknown>(0x2fffffff)", "<unknown>(0x1)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}},
							Processed: []string{"[]interface{}(0x2fffffff len=5 cap=7)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}},
							Processed: []string{"[]int(0x2fffffff len=5 cap=7)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}},
							Processed: []string{"[]interface{}(0x2fffffff len=1 cap=1)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"map[int]int(0x2fffffff)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"map[interface{}]interface{}(0x2fffffff)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"chan int(0x2fffffff)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"chan interface{}(0x2fffffff)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"*S(0x2fffffff)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 3}},
							Processed: []string{"string(0x2fffffff, len=3)"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}, {Value: 3}, {Value: 42}},
							Processed: []string{"string(0x2fffffff, len=3)", "42"},
						},
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
							Values:    []Arg{{}, {}, {}, {}, {}, {}, {}, {}, {Value: 42}, {Value: 43}},
							Processed: []string{"0", "0", "0", "0", "0", "0", "0", "0", "42", "43"},
							Elided:    true,
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
							Processed: []string{"error(0x2fffffff)"},
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
							Processed: []string{"error(0x2fffffff)"},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"0.5"},
						},
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
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"0.5"},
						},
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

func TestLoadErr(t *testing.T) {
	t.Parallel()
	root, err := ioutil.TempDir("", "stack")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer func() {
		if err2 := os.RemoveAll(root); err2 != nil {
			t.Fatalf("failed to remove temporary directory %q: %v", root, err2)
		}
	}()

	tree := map[string]string{
		"bad.go":       "bad content",
		"foo.asm":      "; good but ignored",
		"good.go":      "package main",
		"no_access.go": "package main",
	}
	createTree(t, root, tree)

	type dataLine struct {
		src  string
		line int
		err  error
	}
	// Note: these tests assumes an OS running in English-US locale. That should
	// eventually be fixed, maybe by using regexes?
	data := []dataLine{
		{"foo.asm", 1, fmt.Errorf("cannot load non-go file %q", filepath.Join(root, "foo.asm"))},
		{"good.go", 10, errors.New("line 10 is over line count of 1")},
	}
	if internaltest.GetGoMinorVersion() > 11 {
		// The format changed between 1.9 and 1.12.
		data = append(data, dataLine{"bad.go", 1, fmt.Errorf("failed to parse %s:1:1: expected 'package', found bad", filepath.Join(root, "bad.go"))})
	}
	msg := "The system cannot find the file specified."
	if runtime.GOOS != "windows" {
		msg = "no such file or directory"
		// Chmod has no effect on Windows.
		compareErr(t, nil, os.Chmod(filepath.Join(root, "no_access.go"), 0))
		data = append(data, dataLine{"no_access.go", 1, fmt.Errorf("open %s: permission denied", filepath.Join(root, "no_access.go"))})
	}
	data = append(data, dataLine{"missing.go", 1, fmt.Errorf("open %s: %s", filepath.Join(root, "missing.go"), msg)})

	for _, line := range data {
		g := []*Goroutine{
			{Signature: Signature{Stack: Stack{Calls: []Call{{LocalSrcPath: filepath.Join(root, line.src), Line: line.line}}}}},
		}
		compareErr(t, line.err, Augment(g))
	}
}

func TestLineToByteOffsets(t *testing.T) {
	src := "\n\n\n"
	want := []int{0, 0, 1, 2, 3}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
	}
	src = "hello"
	want = []int{0, 0}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
	}
	src = "this\nis\na\ntest"
	want = []int{0, 0, 5, 8, 10}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
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
