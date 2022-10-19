// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/maruel/panicparse/v2/internal/internaltest"
)

func TestAggregated_ToHTML_2Buckets(t *testing.T) {
	t.Parallel()
	buf := bytes.Buffer{}
	if err := getBuckets().ToHTML(&buf, ""); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 4000 || l > 10000 {
		t.Fatalf("unexpected length %d", l)
	}
}

func TestAggregated_ToHTML_1Bucket(t *testing.T) {
	t.Parallel()
	// Exercise a condition when there's only one bucket.
	buf := bytes.Buffer{}
	a := getBuckets()
	a.Buckets = a.Buckets[:1]
	if err := a.ToHTML(&buf, ""); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 4000 || l > 10000 {
		t.Fatalf("unexpected length %d", l)
	}
	if strings.Contains(buf.String(), "foo-bar") {
		t.Fatal("unexpected")
	}
}

func TestAggregated_ToHTML_1Bucket_Footer(t *testing.T) {
	t.Parallel()
	buf := bytes.Buffer{}
	a := getBuckets()
	a.Buckets = a.Buckets[:1]
	if err := a.ToHTML(&buf, "foo-bar"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "foo-bar") {
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
	const prefix = "devel +"
	if strings.HasPrefix(ver, prefix) {
		ver = ver[len(prefix) : len(prefix)+10]
	}
	ver = url.QueryEscape(ver)
	data := []struct {
		name        string
		c           Call
		url, branch template.URL
		pkgURL      template.URL
		loc         Location
	}{
		{
			"stdlib",
			newCallLocal(
				"net/http.(*Server).Serve",
				Args{},
				goroot+"/src/net/http/server.go",
				2933),
			template.URL("https://github.com/golang/go/blob/" + ver + "/src/net/http/server.go#L2933"),
			template.URL(ver),
			"https://golang.org/pkg/net/http#Server.Serve",
			Stdlib,
		},
		{
			"gomodref",
			newCallLocal(
				"github.com/mattn/go-colorable.(*NonColorable).Write",
				Args{},
				gopath+"/pkg/mod/github.com/mattn/go-colorable@v0.1.6/noncolorable.go",
				30),
			"https://github.com/mattn/go-colorable/blob/v0.1.6/noncolorable.go#L30",
			"v0.1.6",
			"https://pkg.go.dev/github.com/mattn/go-colorable@v0.1.6#NonColorable.Write",
			GoPkg,
		},
		/* TODO(maruel): Fix this.
		{
			"gomodref_with_dot",
			newCallLocal(
				"gopkg.in/fsnotify%2ev1.NewWatcher",
				Args{},
				gopath+"/pkg/mod/gopkg.in/fsnotify.v1@v1.4.7/inotify.go",
				59),
			"file:////home/user/go/pkg/mod/gopkg.in/fsnotify.v1@v1.4.7/inotify.go",
			"v1.4.7",
			"https://pkg.go.dev/gopkg.in/fsnotify.v1@v1.4.7#NewWatcher",
			GoPkg,
		},
		*/
		{
			"gomod_commit_ref",
			newCallLocal(
				"golang.org/x/sys/unix.Nanosleep",
				Args{},
				gopath+"/pkg/mod/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go",
				1160),
			"https://github.com/golang/sys/blob/d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go#L1160",
			"v0.0.0-20200223170610-d5e6a3e2c0ae",
			"https://pkg.go.dev/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix#Nanosleep",
			GoPkg,
		},
		{
			"vendor",
			newCallLocal(
				"github.com/maruel/panicparse/vendor/golang.org/x/sys/unix.Nanosleep",
				Args{},
				gopath+"/src/github.com/maruel/panicparse/vendor/golang.org/x/sys/unix/zsyscall_linux_amd64.go",
				1100),
			"https://github.com/golang/sys/blob/master/unix/zsyscall_linux_amd64.go#L1100",
			"master",
			"https://godoc.org/golang.org/x/sys/unix#Nanosleep",
			GOPATH,
		},
		{
			"windows",
			Call{RemoteSrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
			LocationUnknown,
		},
		{
			"windows_local",
			Call{LocalSrcPath: "c:/random.go"},
			"file:///c:/random.go",
			"",
			"",
			LocationUnknown,
		},
		{
			"empty",
			Call{},
			"",
			"",
			"",
			LocationUnknown,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			url, branch := getSrcBranchURL(&line.c)
			if url != line.url {
				t.Errorf("%q != %q", url, line.url)
			}
			if branch != line.branch {
				t.Errorf("%q != %q", branch, line.branch)
			}
			if url := srcURL(&line.c); url != line.url {
				t.Errorf("%q != %q", url, line.url)
			}
			if url := pkgURL(&line.c); url != line.pkgURL {
				t.Errorf("%q != %q", url, line.pkgURL)
			}
			if line.c.Location != line.loc {
				t.Errorf("%s != %s", line.loc, line.c.Location)
			}
		})
	}
}

func TestSymbol(t *testing.T) {
	t.Parallel()
	data := []struct {
		in   Func
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
			Func{},
			"",
		},
		{
			newFunc("main.baz"),
			"baz",
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

func TestSnapshot_ToHTML(t *testing.T) {
	t.Parallel()
	data := internaltest.PanicOutputs()["race"]
	if data == nil {
		t.Skip("-race is unsupported on this platform")
	}
	s, _, err := ScanSnapshot(bytes.NewReader(data), io.Discard, DefaultOpts())
	if err != nil {
		t.Fatal(err)
	}
	if s.Goroutines == nil {
		t.Fatal("missing context")
	}
	if s.Goroutines[0].RaceAddr == 0 {
		t.Fatal("expected a race")
	}
	if !s.IsRace() {
		t.Fatal("expected a race")
	}
	if err := s.ToHTML(io.Discard, ""); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkAggregated_ToHTML(b *testing.B) {
	b.ReportAllocs()
	s, _, err := ScanSnapshot(bytes.NewReader(internaltest.StaticPanicwebOutput()), io.Discard, DefaultOpts())
	if err != io.EOF {
		b.Fatal(err)
	}
	if s == nil {
		b.Fatal("missing context")
	}
	a := s.Aggregate(AnyPointer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a.ToHTML(io.Discard, ""); err != nil {
			b.Fatal(err)
		}
	}
}

//

// loadGoroutines should match what is in regen.go.
func loadGoroutines() ([]byte, error) {
	htmlRaw, err := os.ReadFile("goroutines.tpl")
	if err != nil {
		return nil, err
	}
	// Strip out leading whitespace.
	re := regexp.MustCompile("(\n[ \t]*)+")
	htmlRaw = re.ReplaceAll(htmlRaw, []byte("\n"))
	return htmlRaw, nil
}

// getBuckets returns a slice for testing.
func getBuckets() *Aggregated {
	return &Aggregated{
		Snapshot: &Snapshot{
			LocalGOROOT:   runtime.GOROOT(),
			LocalGOPATHs:  []string{"/gopath"},
			RemoteGOROOT:  "/golang",
			RemoteGOPATHs: map[string]string{"/gopath": "/gopath"},
			LocalGomods:   map[string]string{"/tmp": "example.com/foo"},
		},
		Buckets: []*Bucket{
			{
				Signature: Signature{
					State: "chan receive",
					Stack: Stack{
						Calls: []Call{
							newCall(
								"main.func·001",
								Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
								"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
								72),
							{
								Func:          newFunc("sliceInternal"),
								Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
								RemoteSrcPath: "/golang/src/sort/slices.go",
								Line:          72,
								Location:      Stdlib,
							},
							{
								Func:          newFunc("Slice"),
								Args:          Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
								RemoteSrcPath: "/golang/src/sort/slices.go",
								Line:          72,
								Location:      Stdlib,
							},
							newCall(
								"DoStuff",
								Args{Values: []Arg{{Value: 0x11000000}, {Value: 2}}},
								"/gopath/src/foo/bar.go",
								72),
							newCall(
								"doStuffInternal",
								Args{
									Values: []Arg{{Value: 0x11000000}, {Value: 2}},
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
				Signature: Signature{
					State: "running",
					Stack: Stack{Elided: true},
				},
			},
		},
	}
}
