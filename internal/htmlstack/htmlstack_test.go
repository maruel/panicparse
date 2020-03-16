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
	if l := buf.Len(); l < 4980 || l > 4990 {
		t.Fatalf("unexpected length %d", l)
	}
}
