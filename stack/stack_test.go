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
	// One call from main, one from stdlib, one from third party.
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
	extra := &bytes.Buffer{}
	goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra)
	ut.AssertEqual(t, nil, err)
	ut.AssertEqual(t, "panic: reflect.Set: value of type\n\n", extra.String())
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: []Call{
					{
						SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
						Line:       153,
						Func:       Function{"gopkg.in/yaml%2ev2.handleErr"},
						Args:       Args{Values: []Arg{{Value: 0xc208033b20}}},
					},
					{
						SourcePath: goroot + "/src/reflect/value.go",
						Line:       2125,
						Func:       Function{"reflect.Value.assignTo"},
						Args:       Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
					},
					{
						SourcePath: "/gopath/src/github.com/maruel/pre-commit-go/main.go",
						Line:       428,
						Func:       Function{"main.main"},
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	ut.AssertEqual(t, expected, goroutines)
}

func TestParseDumpSameBucket(t *testing.T) {
	// 2 goroutines with the same signature
	data := []string{
		"panic: runtime error: index out of range",
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
	goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), &bytes.Buffer{})
	ut.AssertEqual(t, nil, err)
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "chan receive",
				Stack: []Call{
					{
						SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go",
						Line:       72,
						Func:       Function{"main.func·001"},
					},
				},
				CreatedBy: Call{
					SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go",
					Line:       74,
					Func:       Function{"main.mainImpl"},
				},
			},
			ID:    6,
			First: true,
		},
		{
			Signature: Signature{
				State: "chan receive",
				Stack: []Call{
					{
						SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go",
						Line:       72,
						Func:       Function{"main.func·001"},
					},
				},
				CreatedBy: Call{
					SourcePath: "/gopath/src/github.com/maruel/panicparse/main.go",
					Line:       74,
					Func:       Function{"main.mainImpl"},
				},
			},
			ID:    7,
			First: false,
		},
	}
	ut.AssertEqual(t, expectedGR, goroutines)
	expectedBuckets := Buckets{{expectedGR[0].Signature, []Goroutine{expectedGR[0], expectedGR[1]}}}
	ut.AssertEqual(t, expectedBuckets, SortBuckets(Bucketize(goroutines)))
}

func TestParseDumpNoOffset(t *testing.T) {
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 37 [runnable]:",
		"github.com/luci/luci-go/client/archiver.func·002()",
		"	/gopath/src/github.com/luci/luci-go/client/archiver/archiver.go:110",
		"created by github.com/luci/luci-go/client/archiver.New",
		"	/gopath/src/github.com/luci/luci-go/client/archiver/archiver.go:113 +0x43b",
	}
	goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), &bytes.Buffer{})
	ut.AssertEqual(t, nil, err)
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "runnable",
				Stack: []Call{
					{
						SourcePath: "/gopath/src/github.com/luci/luci-go/client/archiver/archiver.go",
						Line:       110,
						Func:       Function{"github.com/luci/luci-go/client/archiver.func·002"},
					},
				},
				CreatedBy: Call{
					SourcePath: "/gopath/src/github.com/luci/luci-go/client/archiver/archiver.go",
					Line:       113,
					Func:       Function{"github.com/luci/luci-go/client/archiver.New"},
				},
			},
			ID:    37,
			First: true,
		},
	}
	ut.AssertEqual(t, expectedGR, goroutines)
}

func TestParseCCode(t *testing.T) {
	data := []string{
		"SIGQUIT: quit",
		"PC=0x43f349",
		"",
		"goroutine 0 [idle]:",
		"runtime.epollwait(0x4, 0x7fff671c7118, 0xffffffff00000080, 0x0, 0xffffffff0028c1be, 0x0, 0x0, 0x0, 0x0, 0x0, ...)",
		"        " + goroot + "/src/runtime/sys_linux_amd64.s:400 +0x19",
		"runtime.netpoll(0x901b01, 0x0)",
		"        " + goroot + "/src/runtime/netpoll_epoll.go:68 +0xa3",
		"findrunnable(0xc208012000)",
		"        " + goroot + "/src/runtime/proc.c:1472 +0x485",
		"schedule()",
		"        " + goroot + "/src/runtime/proc.c:1575 +0x151",
		"runtime.park_m(0xc2080017a0)",
		"        " + goroot + "/src/runtime/proc.c:1654 +0x113",
		"runtime.mcall(0x432684)",
		"        " + goroot + "/src/runtime/asm_amd64.s:186 +0x5a",
	}
	goroutines, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), &bytes.Buffer{})
	ut.AssertEqual(t, nil, err)
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "idle",
				Stack: []Call{
					{
						SourcePath: goroot + "/src/runtime/sys_linux_amd64.s",
						Line:       400,
						Func:       Function{"runtime.epollwait"},
						Args: Args{
							Values: []Arg{
								{Value: 0x4},
								{Value: 0x7fff671c7118},
								{Value: 0xffffffff00000080},
								{},
								{Value: 0xffffffff0028c1be},
								{},
								{},
								{},
								{},
								{},
							},
							Elided: true,
						},
					},
					{
						SourcePath: goroot + "/src/runtime/netpoll_epoll.go",
						Line:       68,
						Func:       Function{"runtime.netpoll"},
						Args:       Args{Values: []Arg{{Value: 0x901b01}, {}}},
					},
					{
						SourcePath: goroot + "/src/runtime/proc.c",
						Line:       1472,
						Func:       Function{"findrunnable"},
						Args:       Args{Values: []Arg{{Value: 0xc208012000}}},
					},
					{
						SourcePath: goroot + "/src/runtime/proc.c",
						Line:       1575,
						Func:       Function{"schedule"},
					},
					{
						SourcePath: goroot + "/src/runtime/proc.c",
						Line:       1654,
						Func:       Function{"runtime.park_m"},
						Args:       Args{Values: []Arg{{Value: 0xc2080017a0}}},
					},
					{
						SourcePath: goroot + "/src/runtime/asm_amd64.s",
						Line:       186,
						Func:       Function{"runtime.mcall"},
						Args:       Args{Values: []Arg{{Value: 0x432684}}},
					},
				},
			},
			ID:    0,
			First: true,
		},
	}
	ut.AssertEqual(t, expectedGR, goroutines)
}

func TestCallPkg1(t *testing.T) {
	c := Call{
		SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:       153,
		Func:       Function{"gopkg.in/yaml%2ev2.handleErr"},
		Args:       Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
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
	c := Call{
		SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:       153,
		Func:       Function{"gopkg.in/yaml%2ev2.(*decoder).unmarshal"},
		Args:       Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
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
	c := Call{
		SourcePath: goroot + "/src/reflect/value.go",
		Line:       2125,
		Func:       Function{"reflect.Value.assignTo"},
		Args:       Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
	}
	ut.AssertEqual(t, "value.go", c.SourceName())
	ut.AssertEqual(t, "value.go:2125", c.SourceLine())
	ut.AssertEqual(t, "reflect/value.go", c.PkgSource())
	ut.AssertEqual(t, "reflect.Value.assignTo", c.Func.String())
	ut.AssertEqual(t, "Value.assignTo", c.Func.Name())
	ut.AssertEqual(t, "reflect", c.Func.PkgName())
	ut.AssertEqual(t, false, c.Func.IsExported())
	ut.AssertEqual(t, true, c.IsStdlib())
	ut.AssertEqual(t, false, c.IsPkgMain())
}

func TestCallMain(t *testing.T) {
	c := Call{
		SourcePath: "/gopath/src/github.com/maruel/pre-commit-go/main.go",
		Line:       428,
		Func:       Function{"main.main"},
	}
	ut.AssertEqual(t, "main.go", c.SourceName())
	ut.AssertEqual(t, "main.go:428", c.SourceLine())
	ut.AssertEqual(t, "pre-commit-go/main.go", c.PkgSource())
	ut.AssertEqual(t, "main.main", c.Func.String())
	ut.AssertEqual(t, "main", c.Func.Name())
	ut.AssertEqual(t, "main", c.Func.PkgName())
	ut.AssertEqual(t, true, c.Func.IsExported())
	ut.AssertEqual(t, false, c.IsStdlib())
	ut.AssertEqual(t, true, c.IsPkgMain())
}

func TestCallC(t *testing.T) {
	c := Call{
		SourcePath: goroot + "/src/runtime/proc.c",
		Line:       1472,
		Func:       Function{"findrunnable"},
		Args:       Args{Values: []Arg{{Value: 0xc208012000}}},
	}
	ut.AssertEqual(t, "proc.c", c.SourceName())
	ut.AssertEqual(t, "proc.c:1472", c.SourceLine())
	ut.AssertEqual(t, "runtime/proc.c", c.PkgSource())
	ut.AssertEqual(t, "findrunnable", c.Func.String())
	ut.AssertEqual(t, "findrunnable", c.Func.Name())
	ut.AssertEqual(t, "", c.Func.PkgName())
	ut.AssertEqual(t, false, c.Func.IsExported())
	ut.AssertEqual(t, true, c.IsStdlib())
	ut.AssertEqual(t, false, c.IsPkgMain())
}

func TestArgs(t *testing.T) {
	a := Args{
		Values: []Arg{
			{Value: 0x4},
			{Value: 0x7fff671c7118},
			{Value: 0xffffffff00000080},
			{},
			{Value: 0xffffffff0028c1be},
			{},
			{},
			{},
			{},
			{},
		},
		Elided: true,
	}
	ut.AssertEqual(t, "0x4, 0x7fff671c7118, 0xffffffff00000080, 0x0, 0xffffffff0028c1be, 0x0, 0x0, 0x0, 0x0, 0x0, ...", a.String())
}

func TestFunctionAnonymous(t *testing.T) {
	f := Function{"main.func·001"}
	ut.AssertEqual(t, "main.func·001", f.String())
	ut.AssertEqual(t, "func·001", f.Name())
	ut.AssertEqual(t, "main", f.PkgName())
	ut.AssertEqual(t, false, f.IsExported())
}
