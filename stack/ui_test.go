// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"testing"

	"github.com/maruel/ut"
)

var p = &Palette{
	EOLReset:               "A",
	RoutineFirst:           "B",
	Routine:                "C",
	CreatedBy:              "D",
	Package:                "E",
	SourceFile:             "F",
	FunctionStdLib:         "G",
	FunctionStdLibExported: "H",
	FunctionMain:           "I",
	FunctionOther:          "J",
	FunctionOtherExported:  "K",
	Arguments:              "L",
}

func TestCalcLengths(t *testing.T) {
	defer reset()
	goroot = "/goroot"
	dest := getGOPATHs()[0]
	gopaths = map[string]string{"/gopath": dest}
	b := Buckets{
		{
			Signature{Stack: Stack{Calls: []Call{{SourcePath: "/gopath/foo/baz.go", Line: 123, Func: Function{"foo.func·001"}}}}},
			nil,
		},
	}
	srcLen, pkgLen := CalcLengths(b, true)
	// When printing, it prints the remote path, not the transposed local path.
	ut.AssertEqual(t, "/gopath/foo/baz.go:123", b[0].Signature.Stack.Calls[0].FullSourceLine())
	ut.AssertEqual(t, len("/gopath/foo/baz.go:123"), srcLen)
	ut.AssertEqual(t, len("foo"), pkgLen)

	srcLen, pkgLen = CalcLengths(b, false)
	ut.AssertEqual(t, "baz.go:123", b[0].Signature.Stack.Calls[0].SourceLine())
	ut.AssertEqual(t, len("baz.go:123"), srcLen)
	ut.AssertEqual(t, len("foo"), pkgLen)
}

func TestBucketHeader(t *testing.T) {
	defer reset()
	goroot = "/goroot"
	dest := getGOPATHs()[0]
	gopaths = map[string]string{"/gopath": dest}
	b := &Bucket{
		Signature{
			State: "chan receive",
			CreatedBy: Call{
				SourcePath: "/gopath/src/github.com/foo/bar/baz.go",
				Line:       74,
				Func:       Function{"main.mainImpl"},
			},
			SleepMax: 6,
			SleepMin: 2,
		},
		[]Goroutine{
			{
				First: true,
			},
			{},
		},
	}
	// When printing, it prints the remote path, not the transposed local path.
	ut.AssertEqual(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /gopath/src/github.com/foo/bar/baz.go:74]A\n", p.BucketHeader(b, true, true))
	ut.AssertEqual(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /gopath/src/github.com/foo/bar/baz.go:74]A\n", p.BucketHeader(b, true, false))
	ut.AssertEqual(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", p.BucketHeader(b, false, true))
	ut.AssertEqual(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", p.BucketHeader(b, false, false))

	b = &Bucket{
		Signature{
			State:    "b0rked",
			SleepMax: 6,
			SleepMin: 6,
			Locked:   true,
		},
		nil,
	}
	ut.AssertEqual(t, "C0: b0rked [6 minutes] [locked]A\n", p.BucketHeader(b, false, false))
}

func TestStackLines(t *testing.T) {
	defer reset()
	goroot = "/goroot"
	dest := getGOPATHs()[0]
	gopaths = map[string]string{"/gopath": dest}
	s := &Signature{
		State: "idle",
		Stack: Stack{
			Calls: []Call{
				{
					SourcePath: "/goroot/src/runtime/sys_linux_amd64.s",
					Line:       400,
					Func:       Function{"runtime.Epollwait"},
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
					SourcePath: "/goroot/src/runtime/netpoll_epoll.go",
					Line:       68,
					Func:       Function{"runtime.netpoll"},
					Args:       Args{Values: []Arg{{Value: 0x901b01}, {}}},
				},
				{
					SourcePath: "/gopath/src/main.go",
					Line:       1472,
					Func:       Function{"main.Main"},
					Args:       Args{Values: []Arg{{Value: 0xc208012000}}},
				},
				{
					SourcePath: "/gopath/src/foo/bar.go",
					Line:       1575,
					Func:       Function{"foo.OtherExported"},
				},
				{
					SourcePath: "/gopath/src/foo/bar.go",
					Line:       10,
					Func:       Function{"foo.otherPrivate"},
				},
			},
			Elided: true,
		},
	}
	// When printing, it prints the remote path, not the transposed local path.
	expected := "" +
		"    Eruntime    F/goroot/src/runtime/sys_linux_amd64.s:400 HEpollwaitL(0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    F/goroot/src/runtime/netpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       F/gopath/src/main.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        F/gopath/src/foo/bar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        F/gopath/src/foo/bar.go:10 JotherPrivateL()A\n" +
		"    (...)\n"
	ut.AssertEqual(t, expected, p.StackLines(s, 10, 10, true))
	expected = "" +
		"    Eruntime    Fsys_linux_amd64.s:400 HEpollwaitL(0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    Fnetpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       Fmain.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        Fbar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        Fbar.go:10  JotherPrivateL()A\n" +
		"    (...)\n"
	ut.AssertEqual(t, expected, p.StackLines(s, 10, 10, false))
}
