// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestAggregateNotAggressive(t *testing.T) {
	// 2 goroutines with similar but not exact same signature.
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 6 [chan receive]:",
		"main.func·001(0x11000000, 2)",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"",
		"goroutine 7 [chan receive]:",
		"main.func·001(0x21000000, 2)",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	actual := Aggregate(c.Goroutines, ExactLines)
	expected := []*Bucket{
		{
			Signature: Signature{
				State: "chan receive",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    72,
							Func:    Func{Raw: "main.func·001"},
							Args:    Args{Values: []Arg{{Value: 0x11000000, Name: ""}, {Value: 2}}},
						},
					},
				},
			},
			IDs:   []int{6},
			First: true,
		},
		{
			Signature: Signature{
				State: "chan receive",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    72,
							Func:    Func{Raw: "main.func·001"},
							Args:    Args{Values: []Arg{{Value: 0x21000000, Name: "#1"}, {Value: 2}}},
						},
					},
				},
			},
			IDs: []int{7},
		},
	}
	compareBuckets(t, expected, actual)
}

func TestAggregateExactMatching(t *testing.T) {
	// 2 goroutines with the exact same signature.
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 6 [chan receive]:",
		"main.func·001()",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"created by main.mainImpl",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:74 +0xeb",
		"",
		"goroutine 7 [chan receive]:",
		"main.func·001()",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"created by main.mainImpl",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:74 +0xeb",
		"",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), &bytes.Buffer{}, false)
	if err != nil {
		t.Fatal(err)
	}
	actual := Aggregate(c.Goroutines, ExactLines)
	expected := []*Bucket{
		{
			Signature: Signature{
				State: "chan receive",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    72,
							Func:    Func{Raw: "main.func·001"},
						},
					},
				},
				CreatedBy: Call{
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    74,
					Func:    Func{Raw: "main.mainImpl"},
				},
			},
			IDs:   []int{6, 7},
			First: true,
		},
	}
	compareBuckets(t, expected, actual)
}

func TestAggregateAggressive(t *testing.T) {
	// 3 goroutines with similar signatures.
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 6 [chan receive, 10 minutes]:",
		"main.func·001(0x11000000, 2)",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"",
		"goroutine 7 [chan receive, 50 minutes]:",
		"main.func·001(0x21000000, 2)",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"",
		"goroutine 8 [chan receive, 100 minutes]:",
		"main.func·001(0x21000000, 2)",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:72 +0x49",
		"",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	actual := Aggregate(c.Goroutines, AnyPointer)
	expected := []*Bucket{
		{
			Signature: Signature{
				State:    "chan receive",
				SleepMin: 10,
				SleepMax: 100,
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    72,
							Func:    Func{Raw: "main.func·001"},
							Args:    Args{Values: []Arg{{Value: 0x11000000, Name: "*"}, {Value: 2}}},
						},
					},
				},
			},
			IDs:   []int{6, 7, 8},
			First: true,
		},
	}
	compareBuckets(t, expected, actual)
}

func compareBuckets(t *testing.T, expected, actual []*Bucket) {
	if len(expected) != len(actual) {
		t.Fatalf("Different []Bucket length:\n- %v\n- %v", expected, actual)
	}
	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			t.Fatalf("Different Bucket:\n- %#v\n- %#v", expected[i], actual[i])
		}
	}
}
