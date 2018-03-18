// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestBucketizeNotAggressive(t *testing.T) {
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
	expectedGR := []Goroutine{
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
			ID:    6,
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
			ID: 7,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
	expectedBuckets := []Bucket{
		{expectedGR[0].Signature, []Goroutine{expectedGR[0]}},
		{expectedGR[1].Signature, []Goroutine{expectedGR[1]}},
	}
	compareBuckets(t, expectedBuckets, Bucketize(c.Goroutines, ExactLines))
}

func TestBucketizeExactMatching(t *testing.T) {
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
	expectedGR := []Goroutine{
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
			ID:    6,
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
						},
					},
				},
				CreatedBy: Call{
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    74,
					Func:    Func{Raw: "main.mainImpl"},
				},
			},
			ID: 7,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
	expectedBuckets := []Bucket{{expectedGR[0].Signature, []Goroutine{expectedGR[0], expectedGR[1]}}}
	compareBuckets(t, expectedBuckets, Bucketize(c.Goroutines, ExactLines))
}

func TestBucketizeAggressive(t *testing.T) {
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
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State:    "chan receive",
				SleepMin: 10,
				SleepMax: 10,
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
			ID:    6,
			First: true,
		},
		{
			Signature: Signature{
				State:    "chan receive",
				SleepMin: 50,
				SleepMax: 50,
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
			ID: 7,
		},
		{
			Signature: Signature{
				State:    "chan receive",
				SleepMin: 100,
				SleepMax: 100,
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
			ID: 8,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
	signature := Signature{
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
	}
	expectedBuckets := []Bucket{{signature, []Goroutine{expectedGR[0], expectedGR[1], expectedGR[2]}}}
	compareBuckets(t, expectedBuckets, Bucketize(c.Goroutines, AnyPointer))
}
