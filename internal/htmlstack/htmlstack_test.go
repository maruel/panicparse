// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package htmlstack

import (
	"bytes"
	"testing"

	"github.com/maruel/panicparse/stack"
)

func TestWrite(t *testing.T) {
	buckets := []*stack.Bucket{
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
							Args:    stack.Args{Values: []stack.Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
						},
					},
				},
			},
			IDs:   []int{1, 2},
			First: true,
		},
		{
			IDs: []int{3},
		},
	}
	buf := bytes.Buffer{}
	if err := Write(&buf, buckets, true); err != nil {
		t.Fatal(err)
	}
	// We expect this to be fairly static across Go versions. We want to know if
	// it changes significantly, thus assert the approximate size. This is being
	// tested on travis.
	if l := buf.Len(); l < 9170 || l > 9900 {
		t.Fatalf("unexpected length %d", l)
	}

	// Exercise a condition when there's only one bucket.
	buf.Reset()
	buckets = buckets[:1]
	if err := Write(&buf, buckets, true); err != nil {
		t.Fatal(err)
	}
}
