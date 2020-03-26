// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package htmlstack

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"regexp"
	"runtime"
	"testing"

	"github.com/maruel/panicparse/internal/internaltest"
	"github.com/maruel/panicparse/stack"
)

func TestWrite2Buckets(t *testing.T) {
	buckets := getBuckets()
	buf := bytes.Buffer{}
	if err := Write(&buf, buckets, true); err != nil {
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
	buckets := getBuckets()[:1]
	buf := bytes.Buffer{}
	if err := Write(&buf, buckets, true); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 4000 || l > 10000 {
		t.Fatalf("unexpected length %d", l)
	}
}

func TestGenerate(t *testing.T) {
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
	ver := runtime.Version()
	data := []struct {
		c           stack.Call
		url, branch template.URL
		pkgURL      template.URL
	}{
		// Stdlib.
		{
			stack.Call{
				SrcPath:      "/goroot/src/net/http/server.go",
				LocalSrcPath: "/goroot/src/net/http/server.go",
				Line:         2933,
				Func:         stack.Func{Raw: "net/http.(*Server).Serve"},
				IsStdlib:     true,
				RelSrcPath:   "net/http/server.go",
			},
			template.URL("https://github.com/golang/go/blob/" + ver + "/src/net/http/server.go#L2933"),
			template.URL(ver),
			"https://golang.org/pkg/net/http#Server.Serve",
		},
		// Go mod ref.
		{
			stack.Call{
				SrcPath:      "/home/maruel/go/pkg/mod/github.com/mattn/go-colorable@v0.1.6/noncolorable.go",
				LocalSrcPath: "/home/maruel/go/pkg/mod/github.com/mattn/go-colorable@v0.1.6/noncolorable.go",
				Line:         30,
				Func:         stack.Func{Raw: "github.com/mattn/go-colorable.(*NonColorable).Write"},
				RelSrcPath:   "github.com/mattn/go-colorable@v0.1.6/noncolorable.go",
			},
			"https://github.com/mattn/go-colorable/blob/v0.1.6/noncolorable.go#L30",
			"v0.1.6",
			"https://pkg.go.dev/github.com/mattn/go-colorable@v0.1.6#NonColorable.Write",
		},
		// Go mod auto-ref.
		{
			stack.Call{
				SrcPath:      "/home/user/go/pkg/mod/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go",
				LocalSrcPath: "/home/user/go/pkg/mod/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go",
				Line:         1160,
				Func:         stack.Func{Raw: "golang.org/x/sys/unix.Nanosleep"},
				RelSrcPath:   "golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go",
			},
			"https://github.com/golang/sys/blob/d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go#L1160",
			"v0.0.0-20200223170610-d5e6a3e2c0ae",
			"https://pkg.go.dev/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix#Nanosleep",
		},
		// Vendor.
		{
			stack.Call{
				SrcPath:      "/home/user/go/src/github.com/maruel/panicparse/vendor/golang.org/x/sys/unix/zsyscall_linux_amd64.go",
				LocalSrcPath: "/home/user/go/src/github.com/maruel/panicparse/vendor/golang.org/x/sys/unix/zsyscall_linux_amd64.go",
				Line:         1100,
				Func:         stack.Func{Raw: "github.com/maruel/panicparse/vendor/golang.org/x/sys/unix.Nanosleep"},
				RelSrcPath:   "github.com/maruel/panicparse/vendor/golang.org/x/sys/unix/zsyscall_linux_amd64.go",
			},
			"https://github.com/golang/sys/blob/master/unix/zsyscall_linux_amd64.go#L1100",
			"master",
			"https://godoc.org/golang.org/x/sys/unix#Nanosleep",
		},
		{
			stack.Call{SrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
		},
		{
			stack.Call{LocalSrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
		},
		{
			stack.Call{},
			"",
			"",
			"",
		},
	}
	for _, line := range data {
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
	}
}

func TestSymbol(t *testing.T) {
	data := []struct {
		in       stack.Func
		expected template.URL
	}{
		{
			stack.Func{Raw: "github.com/mattn/go-colorable.(*NonColorable).Write"},
			"NonColorable.Write",
		},
		{
			stack.Func{Raw: "golang.org/x/sys/unix.Nanosleep"},
			"Nanosleep",
		},
		{
			stack.Func{},
			"",
		},
		{
			stack.Func{Raw: "foo/bar"},
			"",
		},
	}
	for _, line := range data {
		if s := symbol(&line.in); s != line.expected {
			t.Fatalf("%q != %q", s, line.expected)
		}
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
		if err := Write(ioutil.Discard, buckets, true); err != nil {
			b.Fatal(err)
		}
	}
}

//

// loadGoroutines should match what is in regen.go.
func loadGoroutines() ([]byte, error) {
	htmlRaw, err := ioutil.ReadFile("goroutines.tpl")
	if err != nil {
		return nil, err
	}
	// Strip out leading whitespace.
	re := regexp.MustCompile("(\\n[ \\t]*)+")
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
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    72,
							Func:    stack.Func{Raw: "main.func·001"},
							Args:    stack.Args{Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
						},
						{
							SrcPath:  "/golang/src/sort/slices.go",
							Line:     72,
							Func:     stack.Func{Raw: "sliceInternal"},
							Args:     stack.Args{Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
							IsStdlib: true,
						},
						{
							SrcPath:  "/golang/src/sort/slices.go",
							Line:     72,
							Func:     stack.Func{Raw: "Slice"},
							Args:     stack.Args{Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
							IsStdlib: true,
						},
						{
							SrcPath: "/gopath/src/foo/bar.go",
							Line:    72,
							Func:    stack.Func{Raw: "DoStuff"},
							Args:    stack.Args{Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
						},
						{
							SrcPath: "/gopath/src/foo/bar.go",
							Line:    72,
							Func:    stack.Func{Raw: "doStuffInternal"},
							Args: stack.Args{
								Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}},
								Elided: true,
							},
						},
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
				Stack: stack.Stack{
					Calls:  []stack.Call{},
					Elided: true,
				},
			},
		},
	}
}
