// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maruel/panicparse/v2/stack"
)

var testPalette = &Palette{
	EOLReset:                    "A",
	RoutineFirst:                "B",
	Routine:                     "C",
	CreatedBy:                   "D",
	Package:                     "E",
	SrcFile:                     "F",
	FuncMain:                    "G",
	FuncLocationUnknown:         "H",
	FuncLocationUnknownExported: "I",
	FuncGoMod:                   "J",
	FuncGoModExported:           "K",
	FuncGOPATH:                  "L",
	FuncGOPATHExported:          "M",
	FuncGoPkg:                   "N",
	FuncGoPkgExported:           "O",
	FuncStdLib:                  "P",
	FuncStdLibExported:          "Q",
	Arguments:                   "R",
}

func TestCalcBucketsLengths(t *testing.T) {
	t.Parallel()
	a := stack.Aggregated{
		Buckets: []*stack.Bucket{
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
		},
	}
	srcLen, pkgLen := calcBucketsLengths(&a, fullPath)
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "/home/user/go/src/foo/baz.go:123", fullPath.formatCall(&a.Buckets[0].Signature.Stack.Calls[0]))
	compareInt(t, len("/home/user/go/src/foo/baz.go:123"), srcLen)
	compareString(t, "main", a.Buckets[0].Signature.Stack.Calls[0].Func.ImportPath)
	compareInt(t, len("main"), pkgLen)

	srcLen, pkgLen = calcBucketsLengths(&a, basePath)
	compareString(t, "baz.go:123", basePath.formatCall(&a.Buckets[0].Signature.Stack.Calls[0]))
	compareInt(t, len("baz.go:123"), srcLen)
	compareString(t, "main", a.Buckets[0].Signature.Stack.Calls[0].Func.ImportPath)
	compareInt(t, len("main"), pkgLen)
}

func TestBucketHeader(t *testing.T) {
	t.Parallel()
	b := stack.Bucket{
		Signature: stack.Signature{
			State: "chan receive",
			CreatedBy: stack.Stack{
				Calls: []stack.Call{
					newCallLocal("main.mainImpl", stack.Args{}, "/home/user/go/src/github.com/foo/bar/baz.go", 74),
				},
			},
			SleepMax: 6,
			SleepMin: 2,
		},
		IDs:   []int{1, 2},
		First: true,
	}
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /home/user/go/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(&b, fullPath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /home/user/go/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(&b, fullPath, false))
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(&b, relPath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(&b, relPath, false))
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(&b, basePath, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(&b, basePath, false))

	b = stack.Bucket{
		Signature: stack.Signature{
			State:    "b0rked",
			SleepMax: 6,
			SleepMin: 6,
			Locked:   true,
		},
		IDs:   []int{},
		First: true,
	}
	compareString(t, "C0: b0rked [6 minutes] [locked]A\n", testPalette.BucketHeader(&b, basePath, false))
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
		"    Eruntime    F/goroot/src/runtime/sys_linux_amd64.s:400 QEpollwaitR(4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    F/goroot/src/runtime/netpoll_epoll.go:68 PnetpollR(0x901b01, 0)A\n" +
		"    Emain       F/home/user/go/src/main.go:1472 GMainR(0xc208012000)A\n" +
		"    Efoo        F/home/user/go/src/foo/bar.go:1575 MOtherExportedR()A\n" +
		"    Efoo        F/home/user/go/src/foo/bar.go:10 LotherPrivateR()A\n" +
		"    (...)\n"
	compareString(t, want, testPalette.StackLines(s, 10, 10, fullPath))
	want = "" +
		"    Eruntime    Fsys_linux_amd64.s:400 QEpollwaitR(4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    Fnetpoll_epoll.go:68 PnetpollR(0x901b01, 0)A\n" +
		"    Emain       Fmain.go:1472 GMainR(0xc208012000)A\n" +
		"    Efoo        Fbar.go:1575 MOtherExportedR()A\n" +
		"    Efoo        Fbar.go:10  LotherPrivateR()A\n" +
		"    (...)\n"
	compareString(t, want, testPalette.StackLines(s, 10, 10, basePath))
}

//

func newFunc(s string) stack.Func {
	f := stack.Func{}
	if err := f.Init(s); err != nil {
		panic(err)
	}
	return f
}

func newCallLocal(f string, a stack.Args, s string, l int) stack.Call {
	c := stack.Call{Func: newFunc(f), Args: a, RemoteSrcPath: s, Line: l}
	// Do the equivalent of Call.init().
	c.SrcName = filepath.Base(c.RemoteSrcPath)
	c.DirSrc = path.Join(filepath.Base(c.RemoteSrcPath[:len(c.RemoteSrcPath)-len(c.SrcName)-1]), c.SrcName)
	const goroot = "/goroot/src/"
	const gopath = "/home/user/go/src/"
	const gopathmod = "/home/user/go/pkg/mod/"
	// Do the equivalent of Call.updateLocations().
	if strings.HasPrefix(s, goroot) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(goroot):]
		c.Location = stack.Stdlib
	} else if strings.HasPrefix(s, gopath) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(gopath):]
		c.Location = stack.GOPATH
	} else if strings.HasPrefix(s, gopathmod) {
		c.LocalSrcPath = s
		c.RelSrcPath = s[len(gopathmod):]
		c.Location = stack.GoPkg
	}
	return c
}

func compareInt(t *testing.T, want, got int) {
	helper(t)()
	if want != got {
		t.Fatalf("%d != %d", want, got)
	}
}
