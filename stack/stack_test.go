// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maruel/ut"
)

func TestParseDump1(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 1 [running]:",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b20)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"reflect.Value.assignTo(0x570860, 0xc20803f3e0, 0x15)",
		"	" + goroot + "/src/reflect/value.go:2125 +0x368",
		"main.main()",
		"	/gopath/src/github.com/maruel/pre-commit-go/main.go:428 +0x27",
		"",
	}
	extra, goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")))
	ut.AssertEqual(t, nil, err)
	ut.AssertEqual(t, "panic: reflect.Set: value of type\n\n", extra)
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: []Call{
					{SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go", Line: 153, Func: Function{"gopkg.in/yaml%2ev2.handleErr"}, Args: "0xc208033b20"},
					{SourcePath: goroot + "/src/reflect/value.go", Line: 2125, Func: Function{"reflect.Value.assignTo"}, Args: "0x570860, 0xc20803f3e0, 0x15"},
					{SourcePath: "/gopath/src/github.com/maruel/pre-commit-go/main.go", Line: 428, Func: Function{"main.main"}, Args: ""},
				},
			},
			ID:    1,
			First: true,
		},
	}
	ut.AssertEqual(t, expected, goroutines)
}

func TestParseDump2(t *testing.T) {
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 1 [running]:",
		"github.com/maruel/panicparse/ParseDump(0x7f23b0aa5a20, 0xc208036000)",
		"	/gopath/src/github.com/maruel/panicparse/stack.go:272 +0xb77",
		"main.mainImpl(0x0, 0x0)",
		"	/gopath/src/github.com/maruel/panicparse/main.go:90 +0x20c",
		"main.main()",
		"	/gopath/src/github.com/maruel/panicparse/main.go:110 +0x27",
		"",
		"goroutine 5 [syscall]:",
		"os/signal.loop()",
		"	" + goroot + "/src/os/signal/signal_unix.go:21 +0x1f",
		"created by os/signal.init·1",
		"	" + goroot + "/src/os/signal/signal_unix.go:27 +0x35",
		"",
		"goroutine 6 [chan receive]:",
		"main.func·001()",
		"	/gopath/src/github.com/maruel/panicparse/main.go:72 +0x49",
		"created by main.mainImpl",
		"	/gopath/src/github.com/maruel/panicparse/main.go:74 +0xeb",
		"",
		"goroutine 7 [chan receive]:",
		"main.func·001()",
		"	/gopath/src/github.com/maruel/panicparse/main.go:72 +0x49",
		"created by main.mainImpl",
		"	/gopath/src/github.com/maruel/panicparse/main.go:74 +0xeb",
		"",
	}
	extra, goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")))
	ut.AssertEqual(t, nil, err)
	ut.AssertEqual(t, "panic: runtime error: index out of range\n\n", extra)
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: []Call{
					{SourcePath: "/gopath/src/github.com/maruel/panicparse/stack.go", Line: 272, Func: Function{Raw: "github.com/maruel/panicparse/ParseDump"}, Args: "0x7f23b0aa5a20, 0xc208036000"},
					{SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go", Line: 90, Func: Function{Raw: "main.mainImpl"}, Args: "0x0, 0x0"},
					{SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go", Line: 110, Func: Function{Raw: "main.main"}, Args: ""},
				},
				CreatedBy: Function{Raw: ""},
			},
			ID:    1,
			First: true,
		},
		{
			Signature: Signature{
				State: "syscall",
				Stack: []Call{
					{SourcePath: goroot + "/src/os/signal/signal_unix.go", Line: 27, Func: Function{Raw: "os/signal.loop"}, Args: ""},
				},
				CreatedBy: Function{Raw: "os/signal.init·1"},
			},
			ID:    5,
			First: false,
		},
		{
			Signature: Signature{
				State: "chan receive",
				Stack: []Call{
					{SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go", Line: 74, Func: Function{Raw: "main.func·001"}, Args: ""},
				},
				CreatedBy: Function{Raw: "main.mainImpl"},
			},
			ID:    6,
			First: false,
		},
		{
			Signature: Signature{
				State: "chan receive",
				Stack: []Call{
					{SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go", Line: 74, Func: Function{Raw: "main.func·001"}, Args: ""},
				},
				CreatedBy: Function{Raw: "main.mainImpl"},
			},
			ID:    7,
			First: false,
		},
	}
	ut.AssertEqual(t, expectedGR, goroutines)
	buckets := SortBuckets(Bucketize(goroutines))
	expectedBuckets := Buckets{
		{expectedGR[0].Signature, []Goroutine{expectedGR[0]}},
		{expectedGR[2].Signature, []Goroutine{expectedGR[2], expectedGR[3]}},
		{expectedGR[1].Signature, []Goroutine{expectedGR[1]}},
	}
	ut.AssertEqual(t, expectedBuckets, buckets)
}

func TestCallPkg1(t *testing.T) {
	c := Call{SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go", Line: 153, Func: Function{"gopkg.in/yaml%2ev2.handleErr"}, Args: "0xc208033b20"}
	ut.AssertEqual(t, "yaml.go", c.SourceName())
	ut.AssertEqual(t, "yaml.v2/yaml.go", c.PkgSource())
	ut.AssertEqual(t, "gopkg.in/yaml.v2.handleErr", c.Func.String())
	ut.AssertEqual(t, "handleErr", c.Func.Name())
	// This is due to directory name not matching the package name.
	ut.AssertEqual(t, "yaml.v2", c.Func.PkgName())
	ut.AssertEqual(t, false, c.Func.IsExported())
	ut.AssertEqual(t, false, c.IsStdlib())
	ut.AssertEqual(t, false, c.IsPkgMain())
}

func TestCallPkg2(t *testing.T) {
	c := Call{SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go", Line: 153, Func: Function{"gopkg.in/yaml%2ev2.(*decoder).unmarshal"}, Args: "0xc208033b20"}
	ut.AssertEqual(t, "yaml.go", c.SourceName())
	ut.AssertEqual(t, "yaml.v2/yaml.go", c.PkgSource())
	ut.AssertEqual(t, "gopkg.in/yaml.v2.(*decoder).unmarshal", c.Func.String())
	ut.AssertEqual(t, "(*decoder).unmarshal", c.Func.Name())
	// This is due to directory name not matching the package name.
	ut.AssertEqual(t, "yaml.v2", c.Func.PkgName())
	ut.AssertEqual(t, false, c.Func.IsExported())
	ut.AssertEqual(t, false, c.IsStdlib())
	ut.AssertEqual(t, false, c.IsPkgMain())
}

func TestCallStdlib(t *testing.T) {
	c := Call{SourcePath: goroot + "/src/reflect/value.go", Line: 2125, Func: Function{"reflect.Value.assignTo"}, Args: "0x570860, 0xc20803f3e0, 0x15"}
	ut.AssertEqual(t, "value.go", c.SourceName())
	ut.AssertEqual(t, "value.go(2125)", c.SourceLine())
	ut.AssertEqual(t, "reflect/value.go", c.PkgSource())
	ut.AssertEqual(t, "reflect.Value.assignTo", c.Func.String())
	ut.AssertEqual(t, "Value.assignTo", c.Func.Name())
	ut.AssertEqual(t, "reflect", c.Func.PkgName())
	ut.AssertEqual(t, false, c.Func.IsExported())
	ut.AssertEqual(t, true, c.IsStdlib())
	ut.AssertEqual(t, false, c.IsPkgMain())
}

func TestCallMain(t *testing.T) {
	c := Call{SourcePath: "/gopath/src/github.com/maruel/pre-commit-go/main.go", Line: 428, Func: Function{"main.main"}, Args: ""}
	ut.AssertEqual(t, "main.go", c.SourceName())
	ut.AssertEqual(t, "main.go(428)", c.SourceLine())
	ut.AssertEqual(t, "pre-commit-go/main.go", c.PkgSource())
	ut.AssertEqual(t, "main.main", c.Func.String())
	ut.AssertEqual(t, "main", c.Func.Name())
	ut.AssertEqual(t, "main", c.Func.PkgName())
	ut.AssertEqual(t, true, c.Func.IsExported())
	ut.AssertEqual(t, false, c.IsStdlib())
	ut.AssertEqual(t, true, c.IsPkgMain())
}

func TestFunction(t *testing.T) {
	f := Function{"main.func·001"}
	ut.AssertEqual(t, "main.func·001", f.String())
	ut.AssertEqual(t, "func·001", f.Name())
	ut.AssertEqual(t, "main", f.PkgName())
	ut.AssertEqual(t, false, f.IsExported())
}
