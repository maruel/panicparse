// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/panicparse/internal/internaltest"
)

func TestAggregateNotAggressive(t *testing.T) {
	t.Parallel()
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
	want := []*Bucket{
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
						newCall(
							"main.func·001",
							Args{Values: []Arg{{Value: 0x21000000, Name: "#1"}, {Value: 2}}},
							"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							72),
					},
				},
			},
			IDs: []int{7},
		},
	}
	compareBuckets(t, want, Aggregate(c.Goroutines, ExactLines))
}

func TestAggregateExactMatching(t *testing.T) {
	t.Parallel()
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
	want := []*Bucket{
		{
			Signature: Signature{
				State: "chan receive",
				CreatedBy: newCall(
					"main.mainImpl",
					Args{},
					"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					74),
				Stack: Stack{
					Calls: []Call{
						newCall(
							"main.func·001",
							Args{},
							"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							72),
					},
				},
			},
			IDs:   []int{6, 7},
			First: true,
		},
	}
	compareBuckets(t, want, Aggregate(c.Goroutines, ExactLines))
}

func TestAggregateAggressive(t *testing.T) {
	t.Parallel()
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
	want := []*Bucket{
		{
			Signature: Signature{
				State:    "chan receive",
				SleepMin: 10,
				SleepMax: 100,
				Stack: Stack{
					Calls: []Call{
						newCall(
							"main.func·001",
							Args{Values: []Arg{{Value: 0x11000000, Name: "*"}, {Value: 2}}},
							"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							72),
					},
				},
			},
			IDs:   []int{6, 7, 8},
			First: true,
		},
	}
	compareBuckets(t, want, Aggregate(c.Goroutines, AnyPointer))
}

func BenchmarkAggregate(b *testing.B) {
	b.ReportAllocs()
	c, err := ParseDump(bytes.NewReader(internaltest.StaticPanicwebOutput()), ioutil.Discard, true)
	if err != nil {
		b.Fatal(err)
	}
	if c == nil {
		b.Fatal("missing context")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buckets := Aggregate(c.Goroutines, AnyPointer)
		if len(buckets) < 5 {
			b.Fatal("expected more buckets")
		}
	}
}

func compareBuckets(t *testing.T, want, got []*Bucket) {
	helper(t)()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Bucket mismatch (-want +got):\n%s", diff)
	}
}
