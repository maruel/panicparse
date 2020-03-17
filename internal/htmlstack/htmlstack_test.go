// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package htmlstack

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"testing"

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
	if l := buf.Len(); l < 9950 || l > 9960 {
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
	if l := buf.Len(); l < 9820 || l > 9840 {
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
