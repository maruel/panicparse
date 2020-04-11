// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"strings"
	"testing"

	"github.com/maruel/panicparse/stack"
)

var testPalette = &Palette{
	EOLReset:           "A",
	RoutineFirst:       "B",
	Routine:            "C",
	CreatedBy:          "D",
	Package:            "E",
	SrcFile:            "F",
	FuncStdLib:         "G",
	FuncStdLibExported: "H",
	FuncMain:           "I",
	FuncOther:          "J",
	FuncOtherExported:  "K",
	Arguments:          "L",
}

func TestCalcLengths(t *testing.T) {
	t.Parallel()
	b := []*stack.Bucket{
		{
			Signature: stack.Signature{
				Stack: stack.Stack{
					Calls: []stack.Call{
						newCallLocal("main.func·001", stack.Args{}, "/home/user/go/src/foo/baz.go", 123),
					},
				},
			},
			IDs:   []int{},
			First: true,
		},
	}
	srcLen, pkgLen := calcLengths(b, fullPath)
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "/home/user/go/src/foo/baz.go:123", fullPath.formatCall(&b[0].Signature.Stack.Calls[0]))
	compareInt(t, len("/home/user/go/src/foo/baz.go:123"), srcLen)
	compareString(t, "main", b[0].Signature.Stack.Calls[0].Func.PkgName())
	compareInt(t, len("main"), pkgLen)

	srcLen, pkgLen = calcLengths(b, basePath)
	compareString(t, "baz.go:123", basePath.formatCall(&b[0].Signature.Stack.Calls[0]))
	compareInt(t, len("baz.go:123"), srcLen)
	compareString(t, "main", b[0].Signature.Stack.Calls[0].Func.PkgName())
	compareInt(t, len("main"), pkgLen)
}

func TestBucketHeader(t *testing.T) {
	t.Parallel()
	b := &stack.Bucket{
		Signature: stack.Signature{
			State: "chan receive",
			CreatedBy: newCallLocal(
				"main.mainImpl", stack.Args{}, "/home/user/go/src/github.com/foo/bar/baz.go", 74),
			SleepMax: 6,
			SleepMin: 2,
		},
		IDs:   []int{1, 2},
		First: true,
	}
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /home/user/go/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, fullPath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /home/user/go/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, fullPath, false))
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, relPath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, relPath, false))
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(b, basePath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(b, basePath, false))

	b = &stack.Bucket{
		Signature: stack.Signature{
			State:    "b0rked",
			SleepMax: 6,
			SleepMin: 6,
			Locked:   true,
		},
		IDs:   []int{},
		First: true,
	}
	compareString(t, "C0: b0rked [6 minutes] [locked]A\n", testPalette.BucketHeader(b, basePath, false))
}

func TestStackLines(t *testing.T) {
	t.Parallel()
	s := &stack.Signature{
		State: "idle",
		Stack: stack.Stack{
			Calls: []stack.Call{
				newCallLocal(
					"runtime.Epollwait",
					stack.Args{
						Values: []stack.Arg{
							{Value: 4},
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
					"/goroot/src/runtime/sys_linux_amd64.s",
					400),
				newCallLocal(
					"runtime.netpoll",
					stack.Args{Values: []stack.Arg{{Value: 0x901b01}, {}}},
					"/goroot/src/runtime/netpoll_epoll.go",
					68),
				newCallLocal(
					"main.Main",
					stack.Args{Values: []stack.Arg{{Value: 0xc208012000}}},
					"/home/user/go/src/main.go",
					1472),
				newCallLocal(
					"foo.OtherExported",
					stack.Args{},
					"/home/user/go/src/foo/bar.go",
					1575),
				newCallLocal(
					"foo.otherPrivate",
					stack.Args{},
					"/home/user/go/src/foo/bar.go",
					10),
			},
			Elided: true,
		},
	}
	// When printing, it prints the remote path, not the transposed local path.
	want := "" +
		"    Eruntime    F/goroot/src/runtime/sys_linux_amd64.s:400 HEpollwaitL(4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    F/goroot/src/runtime/netpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       F/home/user/go/src/main.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        F/home/user/go/src/foo/bar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        F/home/user/go/src/foo/bar.go:10 JotherPrivateL()A\n" +
		"    (...)\n"
	compareString(t, want, testPalette.StackLines(s, 10, 10, fullPath))
	want = "" +
		"    Eruntime    Fsys_linux_amd64.s:400 HEpollwaitL(4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    Fnetpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       Fmain.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        Fbar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        Fbar.go:10  JotherPrivateL()A\n" +
		"    (...)\n"
	compareString(t, want, testPalette.StackLines(s, 10, 10, basePath))
}

//

func newFunc(s string) stack.Func {
	return stack.Func{Raw: s}
}

func newCallLocal(f string, a stack.Args, s string, l int) stack.Call {
	c := stack.Call{Func: newFunc(f), Args: a, SrcPath: s, Line: l}
	const goroot = "/goroot/src/"
	const gopath = "/home/user/go/src/"
	const gopathmod = "/home/user/go/pkg/mod/"
	// Do the equivalent of Call.updateLocations().
	if strings.HasPrefix(s, goroot) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(goroot):]
		c.IsStdlib = true
	} else if strings.HasPrefix(s, gopath) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(gopath):]
	} else if strings.HasPrefix(s, gopathmod) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(gopathmod):]
	}
	return c
}

func compareInt(t *testing.T, want, got int) {
	helper(t)()
	if want != got {
		t.Fatalf("%d != %d", want, got)
	}
}
