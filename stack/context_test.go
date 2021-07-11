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
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/panicparse/v2/internal/internaltest"
)

func TestScanSnapshotErr(t *testing.T) {
	t.Parallel()
	data := []*Opts{
		nil,
		{LocalGOROOT: "\\"},
		{LocalGOPATHs: []string{"\\"}},
	}
	for _, opts := range data {
		if _, _, err := ScanSnapshot(&bytes.Buffer{}, ioutil.Discard, opts); err == nil {
			t.Fatal("expected error")
		}
	}
}

func TestScanSnapshotSynthetic(t *testing.T) {
	t.Parallel()
	data := []struct {
		name   string
		in     []string
		prefix string
		suffix string
		err    error
		want   []*Goroutine
	}{
		{
			name: "Nothing",
			err:  io.EOF,
		},
		{
			name:   "NothingEmpty",
			in:     make([]string, 111),
			prefix: strings.Repeat("\n", 110),
			err:    io.EOF,
		},
		{
			name:   "NothingLong",
			in:     []string{strings.Repeat("a", bufio.MaxScanTokenSize+10)},
			prefix: strings.Repeat("a", bufio.MaxScanTokenSize+10),
			err:    io.EOF,
		},

		// One call from main, one from stdlib, one from third party.
		// Create a long first line that will be ignored. It is to guard against
		// https://github.com/maruel/panicparse/issues/17.
		{
			name: "long,main,stdlib,third",
			in: []string{
				strings.Repeat("a", bufio.MaxScanTokenSize+1),
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek()",
				" ??:0 +0x6d",
				"gopkg.in/yaml%2ev2.handleErr(0x433b20)",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"reflect.Value.assignTo(0x570860, 0xc20803f3e0, 0x15)",
				"\t/goroot/src/reflect/value.go:2125 +0x368",
				"main.main()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:428 +0x27",
				"",
			},
			prefix: strings.Repeat("a", bufio.MaxScanTokenSize+1) + "\npanic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek",
									Args{}, "??", 0),
								newCall(
									"gopkg.in/yaml%2ev2.handleErr",
									Args{Values: []Arg{{Value: 0x433b20, IsPtr: true}}},
									"/gopath/src/gopkg.in/yaml.v2/yaml.go",
									153),
								newCall(
									"reflect.Value.assignTo",
									Args{Values: []Arg{{Value: 0x570860, IsPtr: true}, {Value: 0xc20803f3e0, IsPtr: true}, {Value: 0x15}}},
									"/goroot/src/reflect/value.go",
									2125),
								newCall(
									"main.main",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									428),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "LongWait",
			in: []string{
				"panic: bleh",
				"",
				"goroutine 1 [chan send, 100 minutes]:",
				"gopkg.in/yaml%2ev2.handleErr(0x433b20)",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"",
				"goroutine 2 [chan send, locked to thread]:",
				"gopkg.in/yaml%2ev2.handleErr(0x8033b21)",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"",
				"goroutine 3 [chan send, 101 minutes, locked to thread]:",
				"gopkg.in/yaml%2ev2.handleErr(0x8033b22)",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"",
			},
			prefix: "panic: bleh\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State:    "chan send",
						SleepMin: 100,
						SleepMax: 100,
						Stack: Stack{
							Calls: []Call{
								newCall(
									"gopkg.in/yaml%2ev2.handleErr",
									Args{Values: []Arg{{Value: 0x433b20, IsPtr: true}}},
									"/gopath/src/gopkg.in/yaml.v2/yaml.go",
									153),
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
								newCall(
									"gopkg.in/yaml%2ev2.handleErr",
									Args{Values: []Arg{{Value: 0x8033b21, Name: "#1", IsPtr: true}}},
									"/gopath/src/gopkg.in/yaml.v2/yaml.go",
									153),
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
								newCall(
									"gopkg.in/yaml%2ev2.handleErr",
									Args{Values: []Arg{{Value: 0x8033b22, Name: "#2", IsPtr: true}}},
									"/gopath/src/gopkg.in/yaml.v2/yaml.go",
									153),
							},
						},
						Locked: true,
					},
					ID: 3,
				},
			},
		},

		{
			name: "Assembly",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 16 [garbage collection]:",
				"runtime.switchtoM()",
				"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "garbage collection",
						Stack: Stack{
							Calls: []Call{
								newCall(
									"runtime.switchtoM",
									Args{},
									"/goroot/src/runtime/asm_amd64.s",
									198),
							},
						},
					},
					ID:    16,
					First: true,
				},
			},
		},

		{
			name: "Assembly1.3",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 16 [garbage collection]:",
				"runtime.switchtoM()",
				"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0 pc=0x5007be",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "garbage collection",
						Stack: Stack{
							Calls: []Call{
								newCall(
									"runtime.switchtoM",
									Args{},
									"/goroot/src/runtime/asm_amd64.s",
									198),
							},
						},
					},
					ID:    16,
					First: true,
				},
			},
		},

		{
			name: "LineErr",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack/stack.recurseType()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:12345678901234567890",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:12345678901234567890\n",
			err:    errors.New("failed to parse int on line: \"/gopath/src/github.com/maruel/panicparse/stack/stack.go:12345678901234567890\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall("github.com/maruel/panicparse/stack/stack.recurseType", Args{}, "", 0),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "CreatedErr",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack/stack.recurseType()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:1",
				"created by testing.RunTests",
				"\t/goroot/src/testing/testing.go:123456789012345678901 +0xa8b",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "\t/goroot/src/testing/testing.go:123456789012345678901 +0xa8b\n",
			err:    errors.New("failed to parse int on line: \"/goroot/src/testing/testing.go:123456789012345678901 +0xa8b\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State:     "running",
						CreatedBy: Stack{Calls: []Call{newCall("testing.RunTests", Args{}, "", 0)}},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack/stack.recurseType",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									1),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "ValueErr",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack/stack.recurseType(123456789012345678901)",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:9",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "github.com/maruel/panicparse/stack/stack.recurseType(123456789012345678901)\n" +
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:9\n",
			err: errors.New("failed to parse int on line: \"github.com/maruel/panicparse/stack/stack.recurseType(123456789012345678901)\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall("github.com/maruel/panicparse/stack/stack.recurseType", Args{}, "", 0),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "InconsistentIndent",
			in: []string{
				"  goroutine 1 [running]:",
				"  github.com/maruel/panicparse/stack/stack.recurseType()",
				" \t/gopath/src/github.com/maruel/panicparse/stack/stack.go:1",
				"",
			},
			suffix: " \t/gopath/src/github.com/maruel/panicparse/stack/stack.go:1\n",
			err:    errors.New(`inconsistent indentation: " \t/gopath/src/github.com/maruel/panicparse/stack/stack.go:1", expected "  "`),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall("github.com/maruel/panicparse/stack/stack.recurseType", Args{}, "", 0),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "OrderErr",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 16 [garbage collection]:",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"runtime.switchtoM()",
				"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6\n" +
				"runtime.switchtoM()\n" +
				"\t/goroot/src/runtime/asm_amd64.s:198 fp=0xc20cfb80d8 sp=0xc20cfb80d0\n",
			err: errors.New("expected a function after a goroutine header, got: \"/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6\""),
			want: []*Goroutine{
				{
					Signature: Signature{State: "garbage collection"},
					ID:        16,
					First:     true,
				},
			},
		},

		{
			name: "Elided",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 16 [garbage collection]:",
				"github.com/maruel/panicparse/stack/stack.recurseType(0x7f4fa9a3ec70, 0xc208062580, 0x7f4fa9a3e818, 0x50a820, 0xc20803a8a0)",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:53 +0x845 fp=0xc20cfc66d8 sp=0xc20cfc6470",
				"...additional frames elided...",
				"created by testing.RunTests",
				"\t/goroot/src/testing/testing.go:555 +0xa8b",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "garbage collection",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"testing.RunTests",
									Args{},
									"/goroot/src/testing/testing.go",
									555),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack/stack.recurseType",
									Args{
										Values: []Arg{
											{Value: 0x7f4fa9a3ec70, IsPtr: true},
											{Value: 0xc208062580, IsPtr: true},
											{Value: 0x7f4fa9a3e818, IsPtr: true},
											{Value: 0x50a820, IsPtr: true},
											{Value: 0xc20803a8a0, IsPtr: true},
										},
									},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									53),
							},
							Elided: true,
						},
					},
					ID:    16,
					First: true,
				},
			},
		},

		{
			name: "Syscall",
			in: []string{
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
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "syscall",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"os/signal.init·1",
									Args{},
									"/goroot/src/os/signal/signal_unix.go",
									27),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"runtime.notetsleepg",
									Args{
										Values: []Arg{
											{Value: 0x918100, IsPtr: true},
											{Value: 0xffffffffffffffff},
											{Value: 0x1},
										},
									},
									"/goroot/src/runtime/lock_futex.go",
									201),
								newCall(
									"runtime.signal_recv",
									Args{Values: []Arg{{}}},
									"/goroot/src/runtime/sigqueue.go",
									109),
								newCall(
									"os/signal.loop",
									Args{},
									"/goroot/src/os/signal/signal_unix.go",
									21),
								newCall(
									"runtime.goexit",
									Args{},
									"/goroot/src/runtime/asm_amd64.s",
									2232),
							},
						},
					},
					ID:    5,
					First: true,
				},
			},
		},

		{
			name: "UnavailCreated",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 24 [running]:",
				"\tgoroutine running on other thread; stack unavailable",
				"created by github.com/maruel/panicparse/stack.New",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:131 +0x381",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.New",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									131),
							},
						},
						Stack: Stack{
							Calls: []Call{newCall("", Args{}, "<unavailable>", 0)},
						},
					},
					ID:    24,
					First: true,
				},
			},
		},

		{
			name: "Unavail",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 24 [running]:",
				"\tgoroutine running on other thread; stack unavailable",
				"",
				"",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{newCall("", Args{}, "<unavailable>", 0)},
						},
					},
					ID:    24,
					First: true,
				},
			},
		},

		{
			name: "UnavailError",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 24 [running]:",
				"\tgoroutine running on other thread; stack unavailable",
				"junk",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "junk",
			err:    errors.New("expected empty line after unavailable stack, got: \"junk\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{newCall("", Args{}, "<unavailable>", 0)},
						},
					},
					ID:    24,
					First: true,
				},
			},
		},

		{
			name: "NoOffset",
			in: []string{
				"panic: runtime error: index out of range",
				"",
				"goroutine 37 [runnable]:",
				"github.com/maruel/panicparse/stack.func·002()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:110",
				"created by github.com/maruel/panicparse/stack.New",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:113 +0x43b",
				"",
			},
			prefix: "panic: runtime error: index out of range\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "runnable",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.New",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									113),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.func·002",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									110),
							},
						},
					},
					ID:    37,
					First: true,
				},
			},
		},

		// For coverage of scanLines.
		{
			name: "HeaderError",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"junk",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "junk",
			err:    errors.New("expected a function after a goroutine header, got: \"junk\""),
			want: []*Goroutine{
				{
					Signature: Signature{State: "running"},
					ID:        1,
					First:     true,
				},
			},
		},

		// For coverage of scanLines.
		{
			name: "FileError",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack.func·002()",
				"junk",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "junk",
			err:    errors.New("expected a file after a function, got: \"junk\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall("github.com/maruel/panicparse/stack.func·002", Args{}, "", 0),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		// For coverage of scanLines.
		{
			name: "Created",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack.func·002()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:110",
				"created by github.com/maruel/panicparse/stack.New",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:131 +0x381",
				"exit status 2",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "exit status 2",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.New",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									131),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.func·002",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									110),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		// For coverage of scanLines.
		{
			name: "CreatedError",
			in: []string{
				"panic: reflect.Set: value of type",
				"",
				"goroutine 1 [running]:",
				"github.com/maruel/panicparse/stack.func·002()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:110",
				"created by github.com/maruel/panicparse/stack.New",
				"junk",
			},
			prefix: "panic: reflect.Set: value of type\n\n",
			suffix: "junk",
			err:    errors.New("expected a file after a created line, got: \"junk\""),
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall("github.com/maruel/panicparse/stack.New", Args{}, "", 0),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/maruel/panicparse/stack.func·002",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									110),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		{
			name: "CCode",
			in: []string{
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
			},
			prefix: "SIGQUIT: quit\nPC=0x43f349\n\n",
			err:    io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "idle",
						Stack: Stack{
							Calls: []Call{
								newCall(
									"runtime.epollwait",
									Args{
										Values: []Arg{
											{Value: 0x4},
											{Value: 0x7fff671c7118, IsPtr: true},
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
									"/goroot/src/runtime/sys_linux_amd64.s",
									400),
								newCall(
									"runtime.netpoll",
									Args{Values: []Arg{{Value: 0x901b01, IsPtr: true}, {}}},
									"/goroot/src/runtime/netpoll_epoll.go",
									68),
								newCall(
									"findrunnable",
									Args{Values: []Arg{{Value: 0xc208012000, IsPtr: true}}},
									"/goroot/src/runtime/proc.c",
									1472),
								newCall("schedule", Args{}, "/goroot/src/runtime/proc.c", 1575),
								newCall(
									"runtime.park_m",
									Args{Values: []Arg{{Value: 0xc2080017a0, IsPtr: true}}},
									"/goroot/src/runtime/proc.c",
									1654),
								newCall(
									"runtime.mcall",
									Args{Values: []Arg{{Value: 0x432684, IsPtr: true}}},
									"/goroot/src/runtime/asm_amd64.s",
									186),
							},
						},
					},
					ID:    0,
					First: true,
				},
			},
		},

		{
			name: "WithCarriageReturn",
			in: []string{
				"goroutine 1 [running]:",
				"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek()",
				" ??:0 +0x6d",
				"gopkg.in/yaml%2ev2.handleErr(0x433b20)",
				"\t/gopath/src/gopkg.in/yaml.v2/yaml.go:153 +0xc6",
				"reflect.Value.assignTo(0x570860, 0xc20803f3e0, 0x15)",
				"\t/goroot/src/reflect/value.go:2125 +0x368",
				"main.main()",
				"\t/gopath/src/github.com/maruel/panicparse/stack/stack.go:428 +0x27",
				"",
			},
			err: io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						Stack: Stack{
							Calls: []Call{
								newCall(
									"github.com/cockroachdb/cockroach/storage/engine._Cfunc_DBIterSeek",
									Args{},
									"??",
									0),
								newCall(
									"gopkg.in/yaml%2ev2.handleErr",
									Args{Values: []Arg{{Value: 0x433b20, IsPtr: true}}},
									"/gopath/src/gopkg.in/yaml.v2/yaml.go",
									153),
								newCall(
									"reflect.Value.assignTo",
									Args{Values: []Arg{{Value: 0x570860, IsPtr: true}, {Value: 0xc20803f3e0, IsPtr: true}, {Value: 0x15}}},
									"/goroot/src/reflect/value.go",
									2125),
								newCall(
									"main.main",
									Args{},
									"/gopath/src/github.com/maruel/panicparse/stack/stack.go",
									428),
							},
						},
					},
					ID:    1,
					First: true,
				},
			},
		},

		// goconvey is culprit of this.
		{
			name: "Indented",
			in: []string{
				"Failures:",
				"",
				"  * /home/maruel/go/src/foo/bar_test.go",
				"  Line 209:",
				"  Expected: '(*errors.errorString){s:\"context canceled\"}'",
				"  Actual:   'nil'",
				"  (Should resemble)!",
				"  goroutine 8 [running]:",
				"  foo/bar.TestArchiveFail.func1.2()",
				"        /home/maruel/go/foo/bar_test.go:209 +0x469",
				"  foo/bar.TestArchiveFail(0x3382000)",
				"        /home/maruel/go/src/foo/bar_test.go:155 +0xf1",
				"  testing.tRunner(0x3382000, 0x1615bf8)",
				"        /home/maruel/golang/go/src/testing/testing.go:865 +0xc0",
				"  created by testing.(*T).Run",
				"        /home/maruel/golang/go/src/testing/testing.go:916 +0x35a",
				"",
				"",
			},
			prefix: strings.Join([]string{
				"Failures:",
				"",
				"  * /home/maruel/go/src/foo/bar_test.go",
				"  Line 209:",
				"  Expected: '(*errors.errorString){s:\"context canceled\"}'",
				"  Actual:   'nil'",
				"  (Should resemble)!",
				"",
			}, "\n"),
			err: io.EOF,
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"testing.(*T).Run",
									Args{},
									"/home/maruel/golang/go/src/testing/testing.go",
									916),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"foo/bar.TestArchiveFail.func1.2",
									Args{},
									"/home/maruel/go/foo/bar_test.go",
									209),
								newCall(
									"foo/bar.TestArchiveFail",
									Args{Values: []Arg{{Value: 0x3382000, Name: "#1", IsPtr: true}}},
									"/home/maruel/go/src/foo/bar_test.go",
									155),
								newCall(
									"testing.tRunner",
									Args{Values: []Arg{{Value: 0x3382000, Name: "#1", IsPtr: true}, {Value: 0x1615bf8, IsPtr: true}}},
									"/home/maruel/golang/go/src/testing/testing.go",
									865),
							},
						},
					},
					ID:    8,
					First: true,
				},
			},
		},

		{
			name:   "Race",
			in:     []string{string(internaltest.StaticPanicRaceOutput())},
			prefix: "\nGOTRACEBACK=all\n",
			want: []*Goroutine{
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"main.panicRace",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									153,
								),
								newCall(
									"main.main",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									54,
								),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"main.panicDoRaceRead",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									137,
								),
								newCall(
									"main.panicRace.func2",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									154),
							},
						},
					},
					ID:       8,
					First:    true,
					RaceAddr: 0xc000014100,
				},
				{
					Signature: Signature{
						State: "running",
						CreatedBy: Stack{
							Calls: []Call{
								newCall(
									"main.panicRace",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									150,
								),
								newCall(
									"main.main",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									54,
								),
							},
						},
						Stack: Stack{
							Calls: []Call{
								newCall(
									"main.panicDoRaceWrite",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									132),
								newCall(
									"main.panicRace.func1",
									Args{},
									"/go/src/github.com/maruel/panicparse/cmd/panic/main.go",
									151),
							},
						},
					},
					ID:        7,
					RaceWrite: true,
					RaceAddr:  0xc000014100,
				},
			},
		},

		{
			name: "RaceHdr1Err",
			in: []string{
				string(raceHeaderFooter),
			},
			prefix: string(raceHeaderFooter),
			err:    io.EOF,
		},

		{
			name: "RaceHdr2Err",
			in: []string{
				string(raceHeaderFooter),
				"",
			},
			// TODO(maruel): This is incorrect.
			prefix: "",
			err:    io.EOF,
		},

		{
			name: "RaceHdr3Err",
			in: []string{
				string(raceHeaderFooter),
				string(raceHeader),
			},
			// TODO(maruel): This is incorrect.
			prefix: "",
			err:    io.EOF,
		},

		{
			name: "RaceHdr4Err",
			in: []string{
				string(raceHeaderFooter),
				string(raceHeader),
				"",
			},
			// TODO(maruel): This is incorrect.
			prefix: "",
			err:    io.EOF,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			prefix := bytes.Buffer{}
			r := bytes.NewBufferString(strings.Join(line.in, "\n"))
			s, suffix, err := ScanSnapshot(r, &prefix, defaultOpts())
			compareErr(t, line.err, err)
			if line.want == nil {
				if s != nil {
					t.Fatalf("unexpected %v", s)
				}
			} else {
				if s == nil {
					t.Fatalf("expected snapshot")
				}
				compareGoroutines(t, line.want, s.Goroutines)
			}
			compareString(t, line.prefix, prefix.String())
			rest, err := ioutil.ReadAll(r)
			compareErr(t, nil, err)
			compareString(t, line.suffix, string(suffix)+string(rest))
		})
	}
}

func TestScanSnapshotSyntheticTwoSnapshots(t *testing.T) {
	t.Parallel()
	in := bytes.Buffer{}
	in.WriteString("Ya\n")
	in.Write(internaltest.PanicOutputs()["simple"])
	in.WriteString("Ye\n")
	in.Write(internaltest.PanicOutputs()["int"])
	in.WriteString("Yo\n")
	panicParseDir := getPanicParseDir(t)
	ppDir := pathJoin(panicParseDir, "cmd", "panic")

	// First stack:
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(&in, &prefix, defaultOpts())
	compareErr(t, nil, err)
	if !s.guessPaths() {
		t.Error("expected success")
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.main",
							Args{},
							pathJoin(ppDir, "main.go"),
							70,
						),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	compareGoroutines(t, want, s.Goroutines)
	compareString(t, "Ya\nGOTRACEBACK=all\npanic: simple\n\n", prefix.String())

	prefix.Reset()
	r := io.MultiReader(bytes.NewReader(suffix), &in)
	s, suffix, err = ScanSnapshot(r, &prefix, defaultOpts())
	compareErr(t, nil, err)
	if !s.guessPaths() {
		t.Error("expected success")
	}
	want = []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicint",
							Args{Values: []Arg{{Value: 42}}},
							pathJoin(ppDir, "main.go"),
							89,
						),
						newCallLocal(
							"main.glob..func9",
							Args{},
							pathJoin(ppDir, "main.go"),
							310,
						),
						newCallLocal(
							"main.main",
							Args{},
							pathJoin(ppDir, "main.go"),
							72,
						),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	compareGoroutines(t, want, s.Goroutines)
	compareString(t, "Ye\nGOTRACEBACK=all\npanic: 42\n\n", prefix.String())
	compareString(t, "Yo\n", string(suffix))
}

func TestSplitPath(t *testing.T) {
	t.Parallel()
	if p := splitPath(""); p != nil {
		t.Fatalf("expected nil, got: %v", p)
	}
}

func TestGetGOPATHs(t *testing.T) {
	// This test cannot run in parallel.
	old := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", old)
	os.Setenv("GOPATH", "")
	if p := getGOPATHs(); len(p) != 1 {
		// It's the home directory + /go.
		t.Fatalf("expected only one path: %v", p)
	}

	root, err := ioutil.TempDir("", "stack")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	}()
	os.Setenv("GOPATH", filepath.Join(root, "a")+string(filepath.ListSeparator)+filepath.Join(root, "b")+string(filepath.Separator))
	if p := getGOPATHs(); len(p) != 2 {
		t.Fatalf("expected two paths: %v", p)
	}
}

// TestGomoduleComplex is an integration test that creates a non-trivial tree
// of go modules using the "replace" statement.
func TestGomoduleComplex(t *testing.T) {
	// This test cannot run in parallel.
	if internaltest.GetGoMinorVersion() < 11 {
		t.Skip("requires go module support")
	}
	old := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", old)
	root, err := ioutil.TempDir("", "stack")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	}()

	os.Setenv("GOPATH", filepath.Join(root, "go"))
	tree := map[string]string{
		"pkg1/go.mod": "module example.com/pkg1\n" +
			"require (\n" +
			"\texample.com/pkg2 v0.0.1\n" +
			"\texample.com/pkg3 v0.0.1\n" +
			")\n" +
			"replace example.com/pkg2 => ../pkg2\n" +
			// This is kind of a hack to force testing with a package inside GOPATH,
			// since this won't normally work by default.
			"replace example.com/pkg3 => ../go/src/example.com/pkg3\n",
		"pkg1/cmd/main.go": "package main\n" +
			"import \"example.com/pkg1/internal\"\n" +
			"func main() {\n" +
			"\tinternal.CallCallDie()\n" +
			"}\n",
		"pkg1/internal/int.go": "package internal\n" +
			"import \"example.com/pkg2\"\n" +
			"func CallCallDie() {\n" +
			"\tpkg2.CallDie()\n" +
			"}\n",

		"pkg2/go.mod": "module example.com/pkg2\n" +
			"require (\n" +
			"\texample.com/pkg3 v0.0.1\n" +
			")\n" +
			// This is kind of a hack to force testing with a package inside GOPATH,
			// since this won't normally work by default.
			"replace example.com/pkg3 => ../go/src/example.com/pkg3\n",
		"pkg2/src2.go": "package pkg2\n" +
			"import \"example.com/pkg3\"\n" +
			"func CallDie() { pkg3.Die() }\n",

		"go/src/example.com/pkg3/go.mod": "module example.com/pkg3\n",
		"go/src/example.com/pkg3/src3.go": "package pkg3\n" +
			"func Die() { panic(42) }\n",
	}
	createTree(t, root, tree)

	exe := filepath.Join(root, "yo")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if err = internaltest.Compile("./cmd", exe, filepath.Join(root, "pkg1"), true, false); err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(exe).CombinedOutput()
	if err == nil {
		t.Error("expected failure")
	}
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(bytes.NewReader(out), &prefix, defaultOpts())
	compareErr(t, io.EOF, err)
	if !s.guessPaths() {
		t.Error("expected success")
	}
	if s == nil {
		t.Fatal("expected snapshot")
	}
	if s.IsRace() {
		t.Fatal("unexpected race")
	}
	compareString(t, "panic: 42\n\n", prefix.String())
	compareString(t, "", string(suffix))
	wantGOROOT := ""
	compareString(t, wantGOROOT, s.RemoteGOROOT)
	compareString(t, runtime.GOROOT(), strings.Replace(s.LocalGOROOT, "/", pathSeparator, -1))

	rootRemote := root
	if runtime.GOOS == "windows" {
		// On Windows, we must make the path to be POSIX style.
		rootRemote = strings.Replace(root, pathSeparator, "/", -1)
	}
	rootLocal := rootRemote
	if runtime.GOOS == "darwin" {
		// On MacOS, the path is a symlink and it will be somehow evaluated when we
		// get the traces back. This must NOT be run on Windows otherwise the path
		// will be converted to 8.3 format.
		if rootRemote, err = filepath.EvalSymlinks(rootLocal); err != nil {
			t.Fatal(err)
		}
	}

	// This part is a bit tricky. The symlink is evaluated on the left since,
	// since it's what is the "remote" path, but it is not on the right since,
	// which is the "local" path. This difference only exists on MacOS.
	wantGOPATHs := map[string]string{
		pathJoin(rootRemote, "go"): pathJoin(rootLocal, "go"),
	}
	if diff := cmp.Diff(s.RemoteGOPATHs, wantGOPATHs); diff != "" {
		t.Fatalf("+want/-got: %s", diff)
	}

	// Local go module search is on the path with symlink evaluated on MacOS.
	// This is kind of confusing because it is the "remote" path.
	wantGomods := map[string]string{
		pathJoin(rootRemote, "pkg1"): "example.com/pkg1",
		pathJoin(rootRemote, "pkg2"): "example.com/pkg2",
	}
	if diff := cmp.Diff(s.LocalGomods, wantGomods); diff != "" {
		t.Fatalf("+want/-got: %s", diff)
	}

	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						{
							Func:          newFunc("example.com/pkg3.Die"),
							Args:          Args{Elided: true},
							RemoteSrcPath: pathJoin(rootRemote, "go", "src", "example.com", "pkg3", "src3.go"),
							Line:          2,
							SrcName:       "src3.go",
							DirSrc:        "pkg3/src3.go",
							LocalSrcPath:  pathJoin(rootLocal, "go", "src", "example.com", "pkg3", "src3.go"),
							RelSrcPath:    "example.com/pkg3/src3.go",
							ImportPath:    "example.com/pkg3",
							Location:      GOPATH,
						},
						{
							Func:          newFunc("example.com/pkg2.CallDie"),
							Args:          Args{Elided: true},
							RemoteSrcPath: pathJoin(rootRemote, "pkg2", "src2.go"),
							Line:          3,
							SrcName:       "src2.go",
							DirSrc:        "pkg2/src2.go",
							// Since this was found locally as a go module using the remote
							// path, this is correct, even if confusing.
							LocalSrcPath: pathJoin(rootRemote, "pkg2", "src2.go"),
							RelSrcPath:   "src2.go",
							ImportPath:   "example.com/pkg2",
							Location:     GoMod,
						},
						{
							Func:          newFunc("example.com/pkg1/internal.CallCallDie"),
							RemoteSrcPath: pathJoin(rootRemote, "pkg1", "internal", "int.go"),
							Line:          2,
							SrcName:       "int.go",
							DirSrc:        "internal/int.go",
							// Since this was found locally as a go module using the remote
							// path, this is correct, even if confusing.
							LocalSrcPath: pathJoin(rootRemote, "pkg1", "internal", "int.go"),
							RelSrcPath:   "internal/int.go",
							ImportPath:   "example.com/pkg1/internal",
							Location:     GoMod,
						},
						{
							Func:          newFunc("main.main"),
							RemoteSrcPath: pathJoin(rootRemote, "pkg1", "cmd", "main.go"),
							Line:          4,
							SrcName:       "main.go",
							DirSrc:        "cmd/main.go",
							// Since this was found locally as a go module using the remote
							// path, this is correct, even if confusing.
							LocalSrcPath: pathJoin(rootRemote, "pkg1", "cmd", "main.go"),
							RelSrcPath:   "cmd/main.go",
							ImportPath:   "example.com/pkg1/cmd",
							Location:     GoMod,
						},
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)
}

func TestGoRun(t *testing.T) {
	t.Parallel()
	root, err := ioutil.TempDir("", "stack")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	}()

	p := filepath.Join(root, "main.go")
	content := "package main\nfunc main() { panic(42) }\n"
	if err = ioutil.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "run", p)
	out, err := c.CombinedOutput()
	if err == nil {
		t.Fatal("expected failure")
	}
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(bytes.NewReader(out), &prefix, defaultOpts())
	compareErr(t, nil, err)
	compareString(t, "panic: 42\n\n", prefix.String())
	compareString(t, "exit status 2\n", string(suffix))
	if s == nil {
		t.Fatal("expected snapshot")
	}
	if runtime.GOOS == "windows" {
		// On Windows, we must make the path to be POSIX style.
		p = strings.Replace(p, pathSeparator, "/", -1)
	}

	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						{
							Func: Func{
								Complete:   "main.main",
								ImportPath: "main",
								DirName:    "main",
								Name:       "main",
								IsExported: true,
								IsPkgMain:  true,
							},
							RemoteSrcPath: p,
							Line:          2,
							SrcName:       "main.go",
							DirSrc:        path.Base(path.Dir(p)) + "/main.go",
							ImportPath:    "main",
							Location:      LocationUnknown,
						},
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)

	if !s.guessPaths() {
		t.Error("expected success")
	}
	want[0].Stack.Calls[0].LocalSrcPath = p
	want[0].Stack.Calls[0].RelSrcPath = "main.go"
	// This is not technically true, when using go run there's no need for a
	// go.mod file, but I don't think it's worth handling specifically.
	want[0].Stack.Calls[0].Location = GoMod
	similarGoroutines(t, want, s.Goroutines)
}

// TestPanic runs github.com/maruel/panicparse/v2/cmd/panic with every
// supported panic modes.
func TestPanic(t *testing.T) {
	t.Parallel()
	cmds := internaltest.PanicOutputs()
	want := map[string]int{
		"chan_receive":              2,
		"chan_send":                 2,
		"goroutine_1":               2,
		"goroutine_dedupe_pointers": 101,
		"goroutine_100":             101,
	}

	panicParseDir := getPanicParseDir(t)
	ppDir := pathJoin(panicParseDir, "cmd", "panic")

	// Test runtime code. For those not in "custom", just assert that they
	// succeed.
	custom := map[string]func(*testing.T, *Snapshot, *bytes.Buffer, string){
		"args_elided": testPanicArgsElided,
		"mismatched":  testPanicMismatched,
		"race":        testPanicRace,
		"str":         testPanicStr,
		"utf8":        testPanicUTF8,
	}
	// Make sure all custom handlers are showing up in cmds.
	for n := range custom {
		if _, ok := cmds[n]; !ok {
			t.Fatalf("untested mode: %q in:\n%v", n, cmds)
		}
	}

	for cmd, data := range cmds {
		cmd := cmd
		data := data
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			prefix := bytes.Buffer{}
			s, suffix, err := ScanSnapshot(bytes.NewReader(data), &prefix, defaultOpts())
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if s == nil {
				t.Fatal("context is nil")
			}
			if !s.guessPaths() {
				t.Fatal("expected GuessPaths to work")
			}
			if f := custom[cmd]; f != nil {
				f(t, s, &prefix, ppDir)
				return
			}
			e := want[cmd]
			if e == 0 {
				e = 1
			}
			if got := len(s.Goroutines); got != e {
				t.Fatalf("unexpected Goroutines; want %d, got %d", e, got)
			}
			compareString(t, "", string(suffix))
		})
	}
}

func testPanicArgsElided(t *testing.T, s *Snapshot, b *bytes.Buffer, ppDir string) {
	if s.RemoteGOROOT != "" {
		t.Fatalf("RemoteGOROOT is %q", s.RemoteGOROOT)
	}
	if b.String() != "GOTRACEBACK=all\npanic: 1\n\n" {
		t.Fatalf("output: %q", b.String())
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicArgsElided",
							Args{
								Values: []Arg{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}, {Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10}},
								Elided: true,
							},
							pathJoin(ppDir, "main.go"),
							58),
						newCallLocal("main.glob..func1", Args{}, pathJoin(ppDir, "main.go"), 134),
						newCallLocal("main.main", Args{}, pathJoin(ppDir, "main.go"), 340),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)
}

func testPanicMismatched(t *testing.T, s *Snapshot, b *bytes.Buffer, ppDir string) {
	if s.RemoteGOROOT != "" {
		t.Fatalf("RemoteGOROOT is %q", s.RemoteGOROOT)
	}
	if b.String() != "GOTRACEBACK=all\npanic: 42\n\n" {
		t.Fatalf("output: %q", b.String())
	}
	ver := "/v2"
	if !internaltest.IsUsingModules() {
		ver = ""
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							// This is important to note here that the Go runtime prints out
							// the package path, and not the package name.
							//
							// Here the package name is "correct". There is no way to deduce
							// this from the stack trace.
							"github.com/maruel/panicparse"+ver+"/cmd/panic/internal/incorrect.Panic",
							Args{},
							pathJoin(ppDir, "internal", "incorrect", "correct.go"),
							7),
						newCallLocal("main.glob..func20", Args{}, pathJoin(ppDir, "main.go"), 314),
						newCallLocal("main.main", Args{}, pathJoin(ppDir, "main.go"), 340),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)
}

func testPanicRace(t *testing.T, s *Snapshot, b *bytes.Buffer, ppDir string) {
	if s.RemoteGOROOT != "" {
		t.Fatalf("RemoteGOROOT is %q", s.RemoteGOROOT)
	}
	if b.String() != "GOTRACEBACK=all\n" {
		t.Fatalf("output: %q", b.String())
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				CreatedBy: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicRace",
							Args{},
							pathJoin(ppDir, "main.go"),
							151,
						),
						newCallLocal(
							"main.main",
							Args{},
							pathJoin(ppDir, "main.go"),
							72,
						),
					},
				},
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicDoRaceRead",
							Args{},
							pathJoin(ppDir, "main.go"),
							150),
						newCallLocal(
							"main.panicRace.func2",
							Args{},
							pathJoin(ppDir, "main.go"),
							135),
					},
				},
			},
			RaceAddr: pointer,
		},
		{
			Signature: Signature{
				State: "running",
				CreatedBy: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicRace",
							Args{},
							pathJoin(ppDir, "main.go"),
							151,
						),
						newCallLocal(
							"main.main",
							Args{},
							pathJoin(ppDir, "main.go"),
							72,
						),
					},
				},
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicDoRaceWrite",
							Args{},
							pathJoin(ppDir, "main.go"),
							145),
						newCallLocal(
							"main.panicRace.func1",
							Args{},
							pathJoin(ppDir, "main.go"),
							132),
					},
				},
			},
			RaceWrite: true,
			RaceAddr:  pointer,
		},
	}
	// IDs are not deterministic, so zap them too but take them for the race
	// detector first.
	for i, g := range s.Goroutines {
		g.ID = i + 1
		if g.RaceAddr > 4*1024*1024 {
			g.RaceAddr = pointer
		}
	}
	// Sometimes the read is detected first.
	if s.Goroutines[0].RaceWrite {
		want[0], want[1] = want[1], want[0]
	}
	// These fields are order-dependent, so set them last.
	want[0].ID = 1
	want[1].ID = 2
	want[0].First = true
	want[1].First = false
	similarGoroutines(t, want, s.Goroutines)
}

func testPanicStr(t *testing.T, s *Snapshot, b *bytes.Buffer, ppDir string) {
	if s.RemoteGOROOT != "" {
		t.Fatalf("RemoteGOROOT is %q", s.RemoteGOROOT)
	}
	if b.String() != "GOTRACEBACK=all\npanic: allo\n\n" {
		t.Fatalf("output: %q", b.String())
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.panicstr",
							Args{Values: []Arg{{Value: 0x123456, IsPtr: true}, {Value: 4}}},
							pathJoin(ppDir, "main.go"),
							50),
						newCallLocal("main.glob..func19", Args{}, pathJoin(ppDir, "main.go"), 307),
						newCallLocal("main.main", Args{}, pathJoin(ppDir, "main.go"), 340),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)
}

func testPanicUTF8(t *testing.T, s *Snapshot, b *bytes.Buffer, ppDir string) {
	if s.RemoteGOROOT != "" {
		t.Fatalf("RemoteGOROOT is %q", s.RemoteGOROOT)
	}
	if b.String() != "GOTRACEBACK=all\npanic: 42\n\n" {
		t.Fatalf("output: %q", b.String())
	}
	ver := "/v2"
	if !internaltest.IsUsingModules() {
		ver = ""
	}
	want := []*Goroutine{
		{
			Signature: Signature{
				State: "running",
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							// This is important to note here the inconsistency in the Go
							// runtime stack generator. The path is escaped, but symbols are
							// not.
							"github.com/maruel/panicparse"+ver+"/cmd/panic/internal/utf8.(*Strùct).Pànic",
							Args{Values: []Arg{{Value: 0xc0000b2e48, IsPtr: true}}},
							// See TestCallUTF8 in stack_test.go for exercising the methods on
							// Call in this situation.
							pathJoin(ppDir, "internal", "utf8", "ùtf8.go"),
							10),
						newCallLocal("main.glob..func21", Args{}, pathJoin(ppDir, "main.go"), 322),
						newCallLocal("main.main", Args{}, pathJoin(ppDir, "main.go"), 340),
					},
				},
			},
			ID:    1,
			First: true,
		},
	}
	similarGoroutines(t, want, s.Goroutines)
}

// TestPanicweb implements the parsing of panicweb output.
//
// panicweb is a separate binary from the rest of panic because importing the
// "net" package causes a background thread to be started, which breaks "panic
// asleep".
func TestPanicweb(t *testing.T) {
	t.Parallel()
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(bytes.NewReader(internaltest.PanicwebOutput()), &prefix, defaultOpts())
	if err != io.EOF {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("snapshot is nil")
	}
	compareString(t, "panic: Here's a snapshot of a normal web server.\n\n", prefix.String())
	compareString(t, "", string(suffix))
	if s.RemoteGOROOT != "" {
		t.Fatalf("unexpected RemoteGOROOT: %q", s.RemoteGOROOT)
	}
	if !s.guessPaths() {
		t.Error("expected success")
	}
	if s.RemoteGOROOT != strings.Replace(runtime.GOROOT(), "\\", "/", -1) {
		t.Fatalf("RemoteGOROOT mismatch; want:%q got:%q", runtime.GOROOT(), s.RemoteGOROOT)
	}
	if got := len(s.Goroutines); got < 30 {
		t.Fatalf("unexpected Goroutines; want at least 30, got %d", got)
	}
	// The goal here is not to find the exact match since it'll change across
	// OSes and Go versions, but to find some of the expected signatures.
	pwebDir := pathJoin(getPanicParseDir(t), "cmd", "panicweb")
	// Reduce the goroutines and categorize the signatures.
	var types []panicwebSignatureType
	for _, b := range s.Aggregate(AnyPointer).Buckets {
		types = append(types, identifyPanicwebSignature(t, b, pwebDir))
	}
	// Count the expected types.
	if v := pstCount(types, pstUnknown); v != 0 {
		t.Fatalf("found %d unknown signatures", v)
	}
	if v := pstCount(types, pstMain); v != 1 {
		t.Fatalf("found %d pstMain signatures", v)
	}
	if v := pstCount(types, pstURL1handler); v != 1 {
		t.Fatalf("found %d URL1Handler signatures", v)
	}
	if v := pstCount(types, pstURL2handler); v != 1 {
		t.Fatalf("found %d URL2Handler signatures", v)
	}
	if v := pstCount(types, pstClient); v == 0 {
		t.Fatalf("found %d client signatures", v)
	}
	if v := pstCount(types, pstServe); v != 1 {
		t.Fatalf("found %d serve signatures", v)
	}
	if v := pstCount(types, pstColorable); v != 1 {
		t.Fatalf("found %d colorable signatures", v)
	}
	if v := pstCount(types, pstStdlib); v < 3 {
		t.Fatalf("found %d stdlib signatures", v)
	}
}

func TestIsGomodule(t *testing.T) {
	t.Parallel()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Our internal functions work with '/' as path separator.
	parts := splitPath(strings.Replace(pwd, "\\", "/", -1))
	gmc := gomodCache{}
	root, importPath := gmc.isGoModule(parts)
	if want := strings.Join(parts[:len(parts)-1], "/"); want != root {
		t.Errorf("want: %q, got: %q", want, root)
	}
	if want := "github.com/maruel/panicparse/v2"; want != importPath {
		t.Errorf("want: %q, got: %q", want, importPath)
	}
	got := reModule.FindStringSubmatch("foo\r\nmodule bar\r\nbaz")
	if diff := cmp.Diff([]string{"module bar\r", "bar"}, got); diff != "" {
		t.Fatalf("-want, +got:\n%s", diff)
	}
}

func TestAtou(t *testing.T) {
	t.Parallel()
	if i, b := atou([]byte("a")); i != 0 || b {
		t.Error("oops")
	}
}

func TestTrimLeftSpace(t *testing.T) {
	t.Parallel()
	if trimLeftSpace(nil) != nil {
		t.Error("oops")
	}
}

func BenchmarkScanSnapshot_Guess(b *testing.B) {
	b.ReportAllocs()
	data := internaltest.StaticPanicwebOutput()
	opts := defaultOpts()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _, err := ScanSnapshot(bytes.NewReader(data), ioutil.Discard, opts)
		if err != io.EOF {
			b.Fatal(err)
		}
		if s == nil {
			b.Fatal("missing context")
		}
	}
}

func BenchmarkScanSnapshot_NoGuess(b *testing.B) {
	b.ReportAllocs()
	data := internaltest.StaticPanicwebOutput()
	opts := defaultOpts()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _, err := ScanSnapshot(bytes.NewReader(data), ioutil.Discard, opts)
		if err != io.EOF {
			b.Fatal(err)
		}
		if s == nil {
			b.Fatal("missing context")
		}
	}
}

func BenchmarkScanSnapshot_Passthru(b *testing.B) {
	b.ReportAllocs()
	buf := make([]byte, b.N)
	for i := range buf {
		buf[i] = 'i'
		if i%16 == 0 {
			buf[i] = '\n'
		}
	}
	prefix := bytes.Buffer{}
	prefix.Grow(len(buf))
	r := bytes.NewReader(buf)
	opts := defaultOpts()
	b.ResetTimer()
	s, suffix, err := ScanSnapshot(r, &prefix, opts)
	if err != io.EOF {
		b.Fatal(err)
	}
	if s != nil {
		b.Fatalf("unexpected %v", s)
	}
	b.StopTimer()
	if !bytes.Equal(prefix.Bytes(), buf) {
		b.Fatal("unexpected prefix")
	}
	if len(suffix) != 0 {
		b.Fatal("unexpected suffix")
	}
}

//

type panicwebSignatureType int

const (
	pstUnknown panicwebSignatureType = iota
	pstMain
	pstURL1handler
	pstURL2handler
	pstClient
	pstServe
	pstColorable
	pstStdlib
)

func pstCount(s []panicwebSignatureType, t panicwebSignatureType) int {
	i := 0
	for _, v := range s {
		if v == t {
			i++
		}
	}
	return i
}

// identifyPanicwebSignature tries to assign one of the predefined signature to
// the bucket provided.
//
// One challenge is that the path will be different depending if this test is
// run within GOPATH or outside.
func identifyPanicwebSignature(t *testing.T, b *Bucket, pwebDir string) panicwebSignatureType {
	ver := ""
	if !isInGOPATH {
		ver = "/v2"
	}

	// The first bucket (the one calling panic()) is deterministic.
	if b.First {
		if len(b.IDs) != 1 {
			t.Fatal("first bucket is not correct")
			return pstUnknown
		}
		crash := Signature{
			State: "running",
			Stack: Stack{
				Calls: []Call{
					newCallLocal("main.main", Args{}, pathJoin(pwebDir, "main.go"), 80),
				},
			},
		}
		similarSignatures(t, &crash, &b.Signature)
		return pstMain
	}

	// We should find exactly 10 sleeping routines in the URL1Handler handler
	// signature and 3 in URL2Handler.
	if s := b.Stack.Calls[0].Func.Name; s == "URL1Handler" || s == "URL2Handler" {
		if b.State != "chan receive" {
			t.Fatalf("suspicious: %#v", b)
			return pstUnknown
		}
		if b.Stack.Calls[0].ImportPath != "github.com/maruel/panicparse"+ver+"/cmd/panicweb/internal" {
			t.Fatalf("suspicious: %q\n%#v", b.Stack.Calls[0].ImportPath, b)
			return pstUnknown
		}
		if b.Stack.Calls[0].SrcName != "internal.go" {
			t.Fatalf("suspicious: %#v", b)
			return pstUnknown
		}
		if b.CreatedBy.Calls[0].SrcName != "server.go" {
			t.Fatalf("suspicious: %#v", b)
			return pstUnknown
		}
		if b.CreatedBy.Calls[0].ImportPath != "net/http" {
			t.Fatalf("suspicious: %#v", b)
			return pstUnknown
		}
		if b.CreatedBy.Calls[0].Func.Name != "(*Server).Serve" {
			t.Fatalf("suspicious: %#v", b)
			return pstUnknown
		}
		if s == "URL1Handler" {
			return pstURL1handler
		}
		return pstURL2handler
	}

	// Find the client goroutine signatures. For the client, it is likely that
	// they haven't all bucketed perfectly.
	if b.CreatedBy.Calls[0].ImportPath == "github.com/maruel/panicparse"+ver+"/cmd/panicweb/internal" && b.CreatedBy.Calls[0].Func.Name == "GetAsync" {
		// TODO(maruel): More checks.
		return pstClient
	}

	// Now find the two goroutine started by main.
	if b.CreatedBy.Calls[0].ImportPath == "github.com/maruel/panicparse"+ver+"/cmd/panicweb" && b.CreatedBy.Calls[0].Func.ImportPath == "main" && b.CreatedBy.Calls[0].Func.Name == "main" {
		if b.State == "IO wait" {
			return pstServe
		}
		if b.State == "chan receive" {
			localgopath := getGOPATHs()[0]
			// If not using Go modules, the path is different as the version in
			// GOPATH is used instead.
			// Warning: This is brittle and will fail whenever go-colorable is
			// updated.
			v := "@v0.1.7"
			prefix := "pkg/mod"
			if !internaltest.IsUsingModules() {
				v = ""
				prefix = "src"
			}
			pColorable := prefix + "/github.com/mattn/go-colorable" + v + "/noncolorable.go"
			want := Signature{
				State: "chan receive",
				CreatedBy: Stack{
					Calls: []Call{
						newCallLocal("main.main", Args{}, pathJoin(pwebDir, "main.go"), 73),
					},
				},
				Stack: Stack{
					Calls: []Call{
						newCallLocal(
							"main.(*writeHang).Write",
							Args{Values: []Arg{{}, {}, {}, {}, {}, {}, {}}},
							pathJoin(pwebDir, "main.go"),
							92),
						newCallLocal(
							"github.com/mattn/go-colorable.(*NonColorable).Write",
							Args{Values: []Arg{{}, {}, {}, {}, {}, {}, {}}},
							pathJoin(localgopath, pColorable),
							30),
					},
				},
				Locked: true,
			}
			// The arguments content is variable, so just count the number of
			// arguments and give up on the rest.
			for i := range b.Signature.Stack.Calls {
				for j := range b.Signature.Stack.Calls[i].Args.Values {
					b.Signature.Stack.Calls[i].Args.Values[j].Value = 0
					b.Signature.Stack.Calls[i].Args.Values[j].Name = ""
					b.Signature.Stack.Calls[i].Args.Values[j].IsPtr = false
				}
			}
			// Warning: This is brittle and will fail whenever go-colorable is
			// updated. See above.
			similarSignatures(t, &want, &b.Signature)
			return pstColorable
		}
		// That's the unix.Nanosleep() or windows.SleepEx() call.
		if b.State == "syscall" {
			created := Stack{
				Calls: []Call{
					newCallLocal("main.main", Args{}, pathJoin(pwebDir, "main.go"), 63),
				},
			}
			zapStacks(t, &created, &b.CreatedBy)
			compareStacks(t, &created, &b.CreatedBy)
			if l := len(b.IDs); l != 1 {
				t.Fatalf("expected 1 goroutine for the signature, got %d", l)
			}
			if l := len(b.Stack.Calls); l != 4 {
				t.Fatalf("expected %d calls, got %d", 4, l)
			}
			if runtime.GOOS == "windows" {
				if s := b.Stack.Calls[0].RelSrcPath; s != "runtime/syscall_windows.go" {
					t.Fatalf("expected %q file, got %q", "runtime/syscall_windows.go", s)
				}
			} else {
				// The first item shall be an assembly file independent of the OS.
				if s := b.Stack.Calls[0].RelSrcPath; !strings.HasSuffix(s, ".s") {
					t.Fatalf("expected assembly file, got %q", s)
				}
			}
			// Process the golang.org/x/sys call specifically.
			path := "golang.org/x/sys/unix"
			fn := "Nanosleep"
			mainOS := "main_unix.go"
			if runtime.GOOS == "windows" {
				path = "golang.org/x/sys/windows"
				fn = "SleepEx"
				mainOS = "main_windows.go"
			}
			usingModules := internaltest.IsUsingModules()
			if b.Stack.Calls[1].Func.ImportPath != path || b.Stack.Calls[1].Func.Name != fn {
				t.Fatalf("expected %q & %q, got %#v", path, fn, b.Stack.Calls[1].Func)
			}
			prefix := "golang.org/x/sys@v0.0.0-"
			if !usingModules {
				// Assert that there's no version by including the trailing /.
				prefix = "golang.org/x/sys/"
			}
			if !strings.HasPrefix(b.Stack.Calls[1].RelSrcPath, prefix) {
				t.Fatalf("expected %q, got %q", prefix, b.Stack.Calls[1].RelSrcPath)
			}
			if usingModules {
				// Assert that it's using @v0-0-0.<date>-<commit> format.
				ver := strings.SplitN(b.Stack.Calls[1].RelSrcPath[len(prefix):], "/", 2)[0]
				re := regexp.MustCompile(`^\d{14}-[a-f0-9]{12}$`)
				if !re.MatchString(ver) {
					t.Fatalf("unexpected version string %q", ver)
				}
			}
			rest := []Call{
				newCallLocal("main.sysHang", Args{}, pathJoin(pwebDir, mainOS), 12),
				newCallLocal(
					"main.main.func3",
					Args{Values: []Arg{{Value: 0xc000140720, Name: "#135", IsPtr: true}}},
					pathJoin(pwebDir, "main.go"),
					65),
			}
			got := b.Stack.Calls[2:]
			for i := range rest {
				zapCalls(t, &got[i], &rest[i])
			}
			if diff := cmp.Diff(rest, got); diff != "" {
				t.Fatalf("rest of stack mismatch (-want +got):\n%s", diff)
			}
			return pstStdlib
		}
		t.Fatalf("suspicious: %# v", b)
		return pstUnknown
	}

	// The rest should all be created with internal threads.
	if b.CreatedBy.Calls[0].Location == Stdlib {
		return pstStdlib
	}

	// On older Go version, there's often an assembly stack in asm_amd64.s.
	if b.CreatedBy.Calls[0].Func.Complete == "" {
		if len(b.Stack.Calls) == 1 && b.Stack.Calls[0].Func.Complete == "runtime.goexit" {
			return pstStdlib
		}
	}
	t.Logf("CreatedBy import: %s", b.CreatedBy.Calls[0].ImportPath)
	t.Logf("CreatedBy:\n%#v", b.CreatedBy)
	t.Fatalf("unexpected thread started by non-stdlib:\n%#v", b.Stack.Calls)
	return pstUnknown
}

//

func defaultOpts() *Opts {
	o := DefaultOpts()
	o.GuessPaths = false
	o.AnalyzeSources = false
	return o
}

// getPanicParseDir returns the path to the root directory of panicparse
// package, using "/" as path separator.
func getPanicParseDir(t *testing.T) string {
	// We assume that the working directory is the directory containing this
	// source. In Go test framework, this normally holds true. If this ever
	// becomes false, let's fix this.
	thisDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// "/" is used even on Windows in the stack trace, return in this format to
	// simply our life.
	return strings.Replace(filepath.Dir(thisDir), "\\", "/", -1)
}

func createTree(t *testing.T, root string, tree map[string]string) {
	for path, content := range tree {
		p := filepath.Join(root, strings.Replace(path, "/", pathSeparator, -1))
		b := filepath.Dir(p)
		if err := os.MkdirAll(b, 0700); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(p, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
