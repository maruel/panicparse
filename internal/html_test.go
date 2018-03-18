// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/maruel/panicparse/stack"
)

func TestWriteToHTML(t *testing.T) {
	f, err := ioutil.TempFile("", "panicparse")
	if err != nil {
		t.Fatal(err)
	}
	n := f.Name()
	f.Close()
	defer func() {
		if err := os.Remove(n); err != nil {
			t.Fatal(err)
		}
	}()
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
	if err := writeToHTML(n, buckets, true); err != nil {
		t.Fatal(err)
	}
}
