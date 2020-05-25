// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package htmlstack

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/maruel/panicparse/internal/internaltest"
	"github.com/maruel/panicparse/stack"
)

func TestWrite2Buckets(t *testing.T) {
	buf := bytes.Buffer{}
	if err := Write(&buf, getBuckets(), false, false); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 4000 || l > 10000 {
		t.Fatalf("unexpected length %d", l)
	}
}

func TestWrite1Bucket(t *testing.T) {
	// Exercise a condition when there's only one bucket.
	buf := bytes.Buffer{}
	if err := Write(&buf, getBuckets()[:1], false, false); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 4000 || l > 10000 {
		t.Fatalf("unexpected length %d", l)
	}
}

const needEnvStr = `To see all goroutines`

const liveStr = `document.addEventListener("DOMContentLoaded", ready);`

func TestWrite(t *testing.T) {
	buf := bytes.Buffer{}
	if err := Write(&buf, getBuckets()[:1], false, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), needEnvStr) {
		t.Fatal("unexpected")
	}
	if strings.Contains(buf.String(), liveStr) {
		t.Fatal("unexpected")
	}
}

func TestWriteNeedEnv(t *testing.T) {
	buf := bytes.Buffer{}
	if err := Write(&buf, getBuckets()[:1], true, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), needEnvStr) {
		t.Fatal("expected")
	}
	if strings.Contains(buf.String(), liveStr) {
		t.Fatal("unexpected")
	}
}

func TestWriteLive(t *testing.T) {
	buf := bytes.Buffer{}
	if err := Write(&buf, getBuckets()[:1], false, true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), needEnvStr) {
		t.Fatal("unexpected")
	}
	if !strings.Contains(buf.String(), liveStr) {
		t.Fatal("expected")
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	// Confirms that nobody forgot to regenate data.go.
	htmlRaw, err := loadGoroutines()
	if err != nil {
		t.Fatal(err)
	}
	if string(htmlRaw) != indexHTML {
		t.Fatal("please run go generate")
	}
}

// TestGetSrcBranchURL also tests pkgURL and srcURL and symbol.
func TestGetSrcBranchURL(t *testing.T) {
	t.Parallel()
	ver := runtime.Version()
	data := []struct {
		name        string
		c           stack.Call
		url, branch template.URL
		pkgURL      template.URL
	}{
		{
			"stdlib",
			newCallLocal(
				"net/http.(*Server).Serve",
				stack.Args{},
				"/goroot/src/net/http/server.go",
				2933),
			template.URL("https://github.com/golang/go/blob/" + ver + "/src/net/http/server.go#L2933"),
			template.URL(ver),
			"https://golang.org/pkg/net/http#Server.Serve",
		},
		{
			"gomodref",
			newCallLocal(
				"github.com/mattn/go-colorable.(*NonColorable).Write",
				stack.Args{},
				"/home/user/go/pkg/mod/github.com/mattn/go-colorable@v0.1.6/noncolorable.go",
				30),
			"https://github.com/mattn/go-colorable/blob/v0.1.6/noncolorable.go#L30",
			"v0.1.6",
			"https://pkg.go.dev/github.com/mattn/go-colorable@v0.1.6#NonColorable.Write",
		},
		{
			"gomodref_with_dot",
			newCallLocal(
				"gopkg.in/fsnotify%2ev1.NewWatcher",
				stack.Args{},
				"/home/user/go/pkg/mod/gopkg.in/fsnotify.v1@v1.4.7/inotify.go",
				59),
			"file:////home/user/go/pkg/mod/gopkg.in/fsnotify.v1@v1.4.7/inotify.go",
			"v1.4.7",
			"https://pkg.go.dev/gopkg.in/fsnotify.v1@v1.4.7#NewWatcher",
		},
		{
			"gomod_commit_ref",
			newCallLocal(
				"golang.org/x/sys/unix.Nanosleep",
				stack.Args{},
				"/home/user/go/pkg/mod/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go",
				1160),
			"https://github.com/golang/sys/blob/d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go#L1160",
			"v0.0.0-20200223170610-d5e6a3e2c0ae",
			"https://pkg.go.dev/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix#Nanosleep",
		},
		{
			"vendor",
			newCallLocal(
				"github.com/maruel/panicparse/vendor/golang.org/x/sys/unix.Nanosleep",
				stack.Args{},
				"/home/user/go/src/github.com/maruel/panicparse/vendor/golang.org/x/sys/unix/zsyscall_linux_amd64.go",
				1100),
			"https://github.com/golang/sys/blob/master/unix/zsyscall_linux_amd64.go#L1100",
			"master",
			"https://godoc.org/golang.org/x/sys/unix#Nanosleep",
		},
		{
			"windows",
			stack.Call{SrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
		},
		{
			"windows_local",
			stack.Call{LocalSrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
		},
		{
			"empty",
			stack.Call{},
			"",
			"",
			"",
		},
	}
	for _, line := range data {
		line := line
		t.Run(line.name, func(t *testing.T) {
			t.Parallel()
			url, branch := getSrcBranchURL(&line.c)
			if url != line.url {
				t.Fatalf("%q != %q", url, line.url)
			}
			if branch != line.branch {
				t.Fatalf("%q != %q", branch, line.branch)
			}
			if url := srcURL(&line.c); url != line.url {
				t.Fatalf("%q != %q", url, line.url)
			}
			if url := pkgURL(&line.c); url != line.pkgURL {
				t.Fatalf("%q != %q", url, line.pkgURL)
			}
		})
	}
}

func TestSymbol(t *testing.T) {
	t.Parallel()
	data := []struct {
		in   stack.Func
		want template.URL
	}{
		{
			newFunc("github.com/mattn/go-colorable.(*NonColorable).Write"),
			"NonColorable.Write",
		},
		{
			newFunc("golang.org/x/sys/unix.Nanosleep"),
			"Nanosleep",
		},
		{
			stack.Func{},
			"",
		},
		{
			newFunc("main.baz"),
			"",
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if s := symbol(&line.in); s != line.want {
				t.Fatalf("%q != %q", s, line.want)
			}
		})
	}
}

func BenchmarkWrite(b *testing.B) {
	b.ReportAllocs()
	c, err := stack.ParseDump(bytes.NewReader(internaltest.StaticPanicwebOutput()), ioutil.Discard, true)
	if err != nil {
		b.Fatal(err)
	}
	if c == nil {
		b.Fatal("missing context")
	}
	buckets := stack.Aggregate(c.Goroutines, stack.AnyPointer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Write(ioutil.Discard, buckets, false, false); err != nil {
			b.Fatal(err)
		}
	}
}

//

func newFunc(s string) stack.Func {
	return stack.Func{Raw: s}
}

func newCall(f string, a stack.Args, s string, l int) stack.Call {
	return stack.Call{Func: newFunc(f), Args: a, SrcPath: s, Line: l}
}

func newCallLocal(f string, a stack.Args, s string, l int) stack.Call {
	c := newCall(f, a, s, l)
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

// loadGoroutines should match what is in regen.go.
func loadGoroutines() ([]byte, error) {
	htmlRaw, err := ioutil.ReadFile("goroutines.tpl")
	if err != nil {
		return nil, err
	}
	// Strip out leading whitespace.
	re := regexp.MustCompile("(\n[ \t]*)+")
	htmlRaw = re.ReplaceAll(htmlRaw, []byte("\n"))
	return htmlRaw, nil
}

// getBuckets returns a slice for testing.
func getBuckets() []*stack.Bucket {
	return []*stack.Bucket{
		{
			Signature: stack.Signature{
				State: "chan receive",
				Stack: stack.Stack{
					Calls: []stack.Call{
						newCall(
							"main.func·001",
							stack.Args{Values: []stack.Arg{{Value: 0x11000000}, {Value: 2}}},
							"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							72),
						{
							Func:     newFunc("sliceInternal"),
							Args:     stack.Args{Values: []stack.Arg{{Value: 0x11000000}, {Value: 2}}},
							SrcPath:  "/golang/src/sort/slices.go",
							Line:     72,
							IsStdlib: true,
						},
						{
							Func:     newFunc("Slice"),
							Args:     stack.Args{Values: []stack.Arg{{Value: 0x11000000}, {Value: 2}}},
							SrcPath:  "/golang/src/sort/slices.go",
							Line:     72,
							IsStdlib: true,
						},
						newCall(
							"DoStuff",
							stack.Args{Values: []stack.Arg{{Value: 0x11000000}, {Value: 2}}},
							"/gopath/src/foo/bar.go",
							72),
						newCall(
							"doStuffInternal",
							stack.Args{
								Values: []stack.Arg{{Value: 0x11000000}, {Value: 2}},
								Elided: true,
							},
							"/gopath/src/foo/bar.go",
							72),
					},
				},
			},
			IDs:   []int{1, 2},
			First: true,
		},
		{
			IDs: []int{3},
			Signature: stack.Signature{
				State: "running",
				Stack: stack.Stack{Elided: true},
			},
		},
	}
}
