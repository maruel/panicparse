// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const crash = `panic: oh no!

goroutine 1 [running]:
panic(0x0, 0x0)
	/home/user/src/golang/src/runtime/panic.go:464 +0x3e6
main.crash2(0x7fe50b49d028, 0xc82000a1e0)
	/home/user/go/src/github.com/maruel/panicparse/cmd/pp/main.go:45 +0x23
main.main()
	/home/user/go/src/github.com/maruel/panicparse/cmd/pp/main.go:50 +0xa6
`

func Example() {
	// Optional: Check for GOTRACEBACK being set, in particular if there is only
	// one goroutine returned.
	in := bytes.NewBufferString(crash)
	c, err := ParseDump(in, os.Stdout, true)
	if err != nil {
		return
	}

	// Find out similar goroutine traces and group them into buckets.
	buckets := SortBuckets(Bucketize(c.Goroutines, AnyValue))

	// Calculate alignment.
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(line.SrcLine()); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.PkgName()); l > pkgLen {
				pkgLen = l
			}
		}
	}

	for _, bucket := range buckets {
		// Print the goroutine header.
		extra := ""
		if s := bucket.SleepString(); s != "" {
			extra += " [" + s + "]"
		}
		if bucket.Locked {
			extra += " [locked]"
		}
		if c := bucket.CreatedByString(false); c != "" {
			extra += " [Created by " + c + "]"
		}
		fmt.Printf("%d: %s%s\n", len(bucket.Routines), bucket.State, extra)

		// Print the stack lines.
		for _, line := range bucket.Stack.Calls {
			fmt.Printf(
				"    %-*s %-*s %s(%s)\n",
				pkgLen, line.Func.PkgName(), srcLen, line.SrcLine(),
				line.Func.Name(), &line.Args)
		}
		if bucket.Stack.Elided {
			io.WriteString(os.Stdout, "    (...)\n")
		}
	}
	// Output:
	// panic: oh no!
	//
	// 1: running
	//          panic.go:464 panic(0, 0)
	//     main main.go:45   crash2(0x7fe50b49d028, 0xc82000a1e0)
	//     main main.go:50   main()
}

func TestParseDump1(t *testing.T) {
	// One call from main, one from stdlib, one from third party.
	// Create a long first line that will be ignored. It is to guard against
	// https://github.com/maruel/panicparse/issues/17.
	long := strings.Repeat("a", bufio.MaxScanTokenSize+1)
	data := []string{
		long,
		"panic: reflect.Set: value of type",
		"",
		"goroutine 1 [running]:",
		"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek()",
		" ??:0 +0x6d",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b20)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"reflect.Value.assignTo(0x570860, 0xc20803f3e0, 0x15)",
		"	/goroot/src/reflect/value.go:2125 +0x368",
		"main.main()",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:428 +0x27",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, true)
	if err != nil {
		t.Fatal(err)
	}
	compareString(t, long+"\npanic: reflect.Set: value of type\n\n", extra.String())
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "??",
							Func:    Func{Raw: "github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek"},
						},
						{
							SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
							Line:    153,
							Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
							Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
						},
						{
							SrcPath: "/goroot/src/reflect/value.go",
							Line:    2125,
							Func:    Func{Raw: "reflect.Value.assignTo"},
							Args:    Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
						},
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    428,
							Func:    Func{Raw: "main.main"},
						},
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	for i := range expected {
		expected[i].updateLocations(c.GOROOT, c.localgoroot, c.GOPATHs)
	}
	compareGoroutines(t, expected, c.Goroutines)
}

func TestParseDumpLongWait(t *testing.T) {
	// One call from main, one from stdlib, one from third party.
	data := []string{
		"panic: bleh",
		"",
		"goroutine 1 [chan send, 100 minutes]:",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b20)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"",
		"goroutine 2 [chan send, locked to thread]:",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b21)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"",
		"goroutine 3 [chan send, 101 minutes, locked to thread]:",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b22)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, true)
	if err != nil {
		t.Fatal(err)
	}
	compareString(t, "panic: bleh\n\n", extra.String())
	expected := []Goroutine{
		{
			Signature: Signature{
				State:    "chan send",
				SleepMin: 100,
				SleepMax: 100,
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
							Line:    153,
							Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
							Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
						},
					},
				},
			},
			ID:    1,
			First: true,
		},
		{
			Signature: Signature{
				State:  "chan send",
				Locked: true,
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
							Line:    153,
							Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
							Args:    Args{Values: []Arg{{Value: 0xc208033b21, Name: "#1"}}},
						},
					},
				},
			},
			ID: 2,
		},
		{
			Signature: Signature{
				State:    "chan send",
				SleepMin: 101,
				SleepMax: 101,
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
							Line:    153,
							Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
							Args:    Args{Values: []Arg{{Value: 0xc208033b22, Name: "#2"}}},
						},
					},
				},
				Locked: true,
			},
			ID: 3,
		},
	}
	for i := range expected {
		expected[i].updateLocations(c.GOROOT, c.localgoroot, c.GOPATHs)
	}
	compareGoroutines(t, expected, c.Goroutines)
}

func TestParseDumpAsm(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 16 [garbage collection]:",
		"runtime.switchtoM()",
		"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "garbage collection",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/goroot/src/runtime/asm_amd64.s",
							Line:    198,
							Func:    Func{Raw: "runtime.switchtoM"},
						},
					},
				},
			},
			ID:    16,
			First: true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
	compareString(t, "panic: reflect.Set: value of type\n\n", extra.String())
}

func TestParseDumpLineErr(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 1 [running]:",
		"github.com/maruel/panicparse/stack/stack.recurseType()",
		"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:12345678901234567890",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	compareErr(t, errors.New("failed to parse int on line: \"/gopath/src/github.com/maruel/panicparse/stack/stack.go:12345678901234567890\""), err)
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{Calls: []Call{{Func: Func{Raw: "github.com/maruel/panicparse/stack/stack.recurseType"}}}},
			},
			ID:    1,
			First: true,
		},
	}
	for i := range expected {
		expected[i].updateLocations(c.GOROOT, c.localgoroot, c.GOPATHs)
	}
	compareGoroutines(t, expected, c.Goroutines)
}

func TestParseDumpValueErr(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 1 [running]:",
		"github.com/maruel/panicparse/stack/stack.recurseType(123456789012345678901)",
		"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:9",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	compareErr(t, errors.New("failed to parse int on line: \"github.com/maruel/panicparse/stack/stack.recurseType(123456789012345678901)\""), err)
	expected := []Goroutine{
		{
			Signature: Signature{State: "running"},
			ID:        1,
			First:     true,
		},
	}
	for i := range expected {
		expected[i].updateLocations(c.GOROOT, c.localgoroot, c.GOPATHs)
	}
	compareGoroutines(t, expected, c.Goroutines)
}

func TestParseDumpOrderErr(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 16 [garbage collection]:",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"runtime.switchtoM()",
		"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	compareErr(t, errors.New("unexpected order on line: \"/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6\""), err)
	expected := []Goroutine{
		{
			Signature: Signature{State: "garbage collection"},
			ID:        16,
			First:     true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
	compareString(t, "panic: reflect.Set: value of type\n\n", extra.String())
}

func TestParseDumpElided(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 16 [garbage collection]:",
		"github.com/maruel/panicparse/stack/stack.recurseType(0x7f4fa9a3ec70, 0xc208062580, 0x7f4fa9a3e818, 0x50a820, 0xc20803a8a0)",
		"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:53 +0x845 fp=0xc20cfc66d8 sp=0xc20cfc6470",
		"...additional frames elided...",
		"created by testing.RunTests",
		"\t/goroot/src/testing/testing.go:555 +0xa8b",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "garbage collection",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    53,
							Func:    Func{Raw: "github.com/maruel/panicparse/stack/stack.recurseType"},
							Args: Args{
								Values: []Arg{
									{Value: 0x7f4fa9a3ec70},
									{Value: 0xc208062580},
									{Value: 0x7f4fa9a3e818},
									{Value: 0x50a820},
									{Value: 0xc20803a8a0},
								},
							},
						},
					},
					Elided: true,
				},
				CreatedBy: Call{
					SrcPath: "/goroot/src/testing/testing.go",
					Line:    555,
					Func:    Func{Raw: "testing.RunTests"},
				},
			},
			ID:    16,
			First: true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
	compareString(t, "panic: reflect.Set: value of type\n\n", extra.String())
}

func TestParseDumpSysCall(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 5 [syscall]:",
		"runtime.notetsleepg(0x918100, 0xffffffffffffffff, 0x1)",
		"\t/goroot/src/runtime/lock_futex.go:201 +0x52 fp=0xc208018f68 sp=0xc208018f40",
		"runtime.signal_recv(0x0)",
		"\t/goroot/src/runtime/sigqueue.go:109 +0x135 fp=0xc208018fa0 sp=0xc208018f68",
		"os/signal.loop()",
		"\t/goroot/src/os/signal/signal_unix.go:21 +0x1f fp=0xc208018fe0 sp=0xc208018fa0",
		"runtime.goexit()",
		"\t/goroot/src/runtime/asm_amd64.s:2232 +0x1 fp=0xc208018fe8 sp=0xc208018fe0",
		"created by os/signal.init·1",
		"\t/goroot/src/os/signal/signal_unix.go:27 +0x35",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "syscall",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/goroot/src/runtime/lock_futex.go",
							Line:    201,
							Func:    Func{Raw: "runtime.notetsleepg"},
							Args: Args{
								Values: []Arg{
									{Value: 0x918100},
									{Value: 0xffffffffffffffff},
									{Value: 0x1},
								},
							},
						},
						{
							SrcPath: "/goroot/src/runtime/sigqueue.go",
							Line:    109,
							Func:    Func{Raw: "runtime.signal_recv"},
							Args: Args{
								Values: []Arg{{}},
							},
						},
						{
							SrcPath: "/goroot/src/os/signal/signal_unix.go",
							Line:    21,
							Func:    Func{Raw: "os/signal.loop"},
						},
						{
							SrcPath: "/goroot/src/runtime/asm_amd64.s",
							Line:    2232,
							Func:    Func{Raw: "runtime.goexit"},
						},
					},
				},
				CreatedBy: Call{
					SrcPath: "/goroot/src/os/signal/signal_unix.go",
					Line:    27,
					Func:    Func{Raw: "os/signal.init·1"},
				},
			},
			ID:    5,
			First: true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
	compareString(t, "panic: reflect.Set: value of type\n\n", extra.String())
}

func TestParseDumpUnavail(t *testing.T) {
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 24 [running]:",
		"\tgoroutine running on other thread; stack unavailable",
		"created by github.com/maruel/panicparse/stack.New",
		"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:131 +0x381",
		"",
	}
	extra := &bytes.Buffer{}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), extra, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{{SrcPath: "<unavailable>"}},
				},
				CreatedBy: Call{
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    131,
					Func:    Func{Raw: "github.com/maruel/panicparse/stack.New"},
				},
			},
			ID:    24,
			First: true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
	compareString(t, "panic: reflect.Set: value of type\n\n", extra.String())
}

func TestParseDumpSameBucket(t *testing.T) {
	// 2 goroutines with the same signature
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
	expectedBuckets := Buckets{{expectedGR[0].Signature, []Goroutine{expectedGR[0], expectedGR[1]}}}
	compareBuckets(t, expectedBuckets, SortBuckets(Bucketize(c.Goroutines, ExactLines)))
}

func TestBucketizeNotAggressive(t *testing.T) {
	// 2 goroutines with the same signature
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
	expectedBuckets := Buckets{
		{expectedGR[0].Signature, []Goroutine{expectedGR[0]}},
		{expectedGR[1].Signature, []Goroutine{expectedGR[1]}},
	}
	compareBuckets(t, expectedBuckets, SortBuckets(Bucketize(c.Goroutines, ExactLines)))
}

func TestBucketizeAggressive(t *testing.T) {
	// 2 goroutines with the same signature
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
	expectedBuckets := Buckets{{signature, []Goroutine{expectedGR[0], expectedGR[1], expectedGR[2]}}}
	compareBuckets(t, expectedBuckets, SortBuckets(Bucketize(c.Goroutines, AnyPointer)))
}

func TestParseDumpNoOffset(t *testing.T) {
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 37 [runnable]:",
		"github.com/maruel/panicparse/stack.func·002()",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:110",
		"created by github.com/maruel/panicparse/stack.New",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:113 +0x43b",
		"",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "runnable",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    110,
							Func:    Func{Raw: "github.com/maruel/panicparse/stack.func·002"},
						},
					},
				},
				CreatedBy: Call{
					SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
					Line:    113,
					Func:    Func{Raw: "github.com/maruel/panicparse/stack.New"},
				},
			},
			ID:    37,
			First: true,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
}

func TestParseDumpJunk(t *testing.T) {
	// For coverage of scanLines.
	data := []string{
		"panic: reflect.Set: value of type",
		"",
		"goroutine 1 [running]:",
		"junk",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	expectedGR := []Goroutine{
		{
			Signature: Signature{State: "running"},
			ID:        1,
			First:     true,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
}

func TestParseCCode(t *testing.T) {
	data := []string{
		"SIGQUIT: quit",
		"PC=0x43f349",
		"",
		"goroutine 0 [idle]:",
		"runtime.epollwait(0x4, 0x7fff671c7118, 0xffffffff00000080, 0x0, 0xffffffff0028c1be, 0x0, 0x0, 0x0, 0x0, 0x0, ...)",
		"        /goroot/src/runtime/sys_linux_amd64.s:400 +0x19",
		"runtime.netpoll(0x901b01, 0x0)",
		"        /goroot/src/runtime/netpoll_epoll.go:68 +0xa3",
		"findrunnable(0xc208012000)",
		"        /goroot/src/runtime/proc.c:1472 +0x485",
		"schedule()",
		"        /goroot/src/runtime/proc.c:1575 +0x151",
		"runtime.park_m(0xc2080017a0)",
		"        /goroot/src/runtime/proc.c:1654 +0x113",
		"runtime.mcall(0x432684)",
		"        /goroot/src/runtime/asm_amd64.s:186 +0x5a",
		"",
	}
	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	expectedGR := []Goroutine{
		{
			Signature: Signature{
				State: "idle",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "/goroot/src/runtime/sys_linux_amd64.s",
							Line:    400,
							Func:    Func{Raw: "runtime.epollwait"},
							Args: Args{
								Values: []Arg{
									{Value: 0x4},
									{Value: 0x7fff671c7118},
									{Value: 0xffffffff00000080},
									{},
									{Value: 0xffffffff0028c1be},
									{},
									{},
									{},
									{},
									{},
								},
								Elided: true,
							},
						},
						{
							SrcPath: "/goroot/src/runtime/netpoll_epoll.go",
							Line:    68,
							Func:    Func{Raw: "runtime.netpoll"},
							Args:    Args{Values: []Arg{{Value: 0x901b01}, {}}},
						},
						{
							SrcPath: "/goroot/src/runtime/proc.c",
							Line:    1472,
							Func:    Func{Raw: "findrunnable"},
							Args:    Args{Values: []Arg{{Value: 0xc208012000}}},
						},
						{
							SrcPath: "/goroot/src/runtime/proc.c",
							Line:    1575,
							Func:    Func{Raw: "schedule"},
						},
						{
							SrcPath: "/goroot/src/runtime/proc.c",
							Line:    1654,
							Func:    Func{Raw: "runtime.park_m"},
							Args:    Args{Values: []Arg{{Value: 0xc2080017a0}}},
						},
						{
							SrcPath: "/goroot/src/runtime/asm_amd64.s",
							Line:    186,
							Func:    Func{Raw: "runtime.mcall"},
							Args:    Args{Values: []Arg{{Value: 0x432684}}},
						},
					},
				},
			},
			ID:    0,
			First: true,
		},
	}
	compareGoroutines(t, expectedGR, c.Goroutines)
}

func TestParseWithCarriageReturn(t *testing.T) {
	data := []string{
		"goroutine 1 [running]:",
		"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek()",
		" ??:0 +0x6d",
		"gopkg.in/yaml%2ev2.handleErr(0xc208033b20)",
		"	/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
		"reflect.Value.assignTo(0x570860, 0xc20803f3e0, 0x15)",
		"	/goroot/src/reflect/value.go:2125 +0x368",
		"main.main()",
		"	/gopath/src/github.com/maruel/panicparse/stack/stack.go:428 +0x27",
		"",
	}

	c, err := ParseDump(bytes.NewBufferString(strings.Join(data, "\r\n")), ioutil.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						{
							SrcPath: "??",
							Func:    Func{Raw: "github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek"},
						},
						{
							SrcPath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
							Line:    153,
							Func:    Func{Raw: "gopkg.in/yaml%2ev2.handleErr"},
							Args:    Args{Values: []Arg{{Value: 0xc208033b20}}},
						},
						{
							SrcPath: "/goroot/src/reflect/value.go",
							Line:    2125,
							Func:    Func{Raw: "reflect.Value.assignTo"},
							Args:    Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
						},
						{
							SrcPath: "/gopath/src/github.com/maruel/panicparse/stack/stack.go",
							Line:    428,
							Func:    Func{Raw: "main.main"},
						},
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	compareGoroutines(t, expected, c.Goroutines)
}

//

func compareErr(t *testing.T, expected, actual error) {
	if expected.Error() != actual.Error() {
		t.Fatalf("%v != %v", expected, actual)
	}
}

func compareGoroutines(t *testing.T, expected, actual []Goroutine) {
	if len(expected) != len(actual) {
		t.Fatalf("Different []Goroutine length:\n- %v\n- %v", expected, actual)
	}
	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			t.Fatalf("Different Goroutine:\n- %v\n- %v", expected[i], actual[i])
		}
	}
}

func compareBuckets(t *testing.T, expected, actual Buckets) {
	if len(expected) != len(actual) {
		t.Fatalf("Different Buckets length:\n- %v\n- %v", expected, actual)
	}
	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			t.Fatalf("Different Bucket:\n- %#v\n- %#v", expected[i], actual[i])
		}
	}
}

func compareString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("%q != %q", expected, actual)
	}
}
