// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/bits"
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

// goarchList is a list of GOARCH values.
type goarchList []string

var allPlatforms goarchList = nil

func TestAugment(t *testing.T) {
	t.Parallel()

	gm := map[string]string{"/root": "main"}
	newCallSrc := func(f string, a Args, s string, l int) Call {
		c := newCall(f, a, s, l)
		// Simulate findRoots().
		if !c.updateLocations(goroot, goroot, gm, gopaths) {
			t.Fatalf("c.updateLocations(%v, %v, %v, %v) failed on %s", goroot, goroot, gm, gopaths, s)
		}
		return c
	}

	// For test case "negative int32".
	negInt := uint64(4294967173)
	negPtr := false
	if bits.UintSize == 64 {
		negInt = 828928688005
		negPtr = true
	}

	type testCase struct {
		name  string
		input string
		// Starting with go1.11, inlining is enabled. The stack trace may (it
		// depends on tool chain version) not contain much information about the
		// arguments and shows as elided. Non-pointer call may show an elided
		// argument, while there was no argument listed before.
		mayBeInlined bool
		// archBlock lists the CPU architectures to skip this test case on.
		//
		// Many test are hard to parse in 32 bits. Eventually we should fix these
		// but I don't have time for this.
		//
		// The list is based on https://github.com/maruel/panicparse/issues/80.
		// I was able to locally reproduce amd64 and 386 but not the rest.
		archBlock goarchList
		want      Stack
	}
	data := []testCase{
		{
			"local function doesn't interfere",
			`func main() {
				f("yo")
			}
			func f(s string) {
				a := func(i int) int {
					return 1 + i
				}
				_ = a(3)
				panic("ooh")
			}`,
			// The function became inlinable in go 1.17.
			true,
			goarchList{"arm", "arm64", "ppc64le", "riscv64"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 2}},
							}}},
							Processed: []string{"string(0x2fffffff, len=2)"},
						},
						"/root/main.go",
						10),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"func",
			`func main() {
				f(func() string { return "ooh" })
			}
			func f(a func() string) {
				panic(a())
			}`,
			true,
			goarchList{"arm"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"func(0x2fffffff)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"func ellipsis",
			`func main() {
				f(func() string { return "ooh" })
			}
			func f(a ...func() string) {
				panic(a[0]())
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}},
							}}},
							Processed: []string{"<unknown>{0x2fffffff, 0x1, 0x1}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"interface{}",
			`func main() {
				f(make([]interface{}, 5, 7))
			}
			func f(a []interface{}) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}},
							}}},
							Processed: []string{"[]interface{}(0x2fffffff len=5 cap=7)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"[]int",
			`func main() {
				f(make([]int, 5, 7))
			}
			func f(a []int) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 5}, {Value: 7}},
							}}},
							Processed: []string{"[]int(0x2fffffff len=5 cap=7)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"[]interface{}",
			`func main() {
				f([]interface{}{"ooh"})
			}
			func f(a []interface{}) {
				panic(a[0].(string))
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 1}, {Value: 1}},
							}}},
							Processed: []string{"[]interface{}(0x2fffffff len=1 cap=1)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"map[int]int",
			`func main() {
				f(map[int]int{1: 2})
			}
			func f(a map[int]int) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"map[int]int(0x2fffffff)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"map[interface{}]interface{}",
			`func main() {
				f(make(map[interface{}]interface{}))
			}
			func f(a map[interface{}]interface{}) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"map[interface{}]interface{}(0x2fffffff)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"chan int",
			`func main() {
				f(make(chan int))
			}
			func f(a chan int) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"chan int(0x2fffffff)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"chan interface{}",
			`func main() {
				f(make(chan interface{}))
			}
			func f(a chan interface{}) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"chan interface{}(0x2fffffff)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"non-pointer method",
			`func main() {
				var s S
				s.f()
			}
			type S struct {}
			func (s S) f() {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.S.f",
						Args{Values: []Arg{{IsAggregate: true, Fields: Args{}}}},
						"/root/main.go",
						8),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"pointer method",
			`func main() {
				var s S
				s.f()
			}
			type S struct {}
			func (s *S) f() {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.(*S).f",
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"*S(0x2fffffff)"},
						},
						"/root/main.go",
						8),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"string",
			`func main() {
			  f("ooh")
			}
			func f(s string) {
				panic(s)
			}`,
			true,
			goarchList{"arm", "arm64", "ppc64le", "riscv64"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 3}},
							}}},
							Processed: []string{"string(0x2fffffff, len=3)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"string and int",
			`func main() {
			  f("ooh", 42)
			}
			func f(s string, i int) {
				panic(s)
			}`,
			true,
			goarchList{"arm", "arm64", "ppc64le", "riscv64"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{
								{IsAggregate: true, Fields: Args{
									Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 3}},
								}},
								{Value: 42},
							},
							Processed: []string{"string(0x2fffffff, len=3)", "42"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"values are elided",
			`func main() {
				f(0, 0, 0, 0, 0, 0, 0, 0, 42, 43, 44, 45, nil)
			}
			func f(s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12 int, s13 interface{}) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{}, {}, {}, {}, {}, {}, {}, {}, {Value: 42}, {Value: 43}},
							Processed: []string{"0", "0", "0", "0", "0", "0", "0", "0", "42", "43"},
							Elided:    true,
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"error",
			`import "errors"
			func main() {
				f(errors.New("ooh"))
			}
			func f(err error) {
				panic(err.Error())
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: pointer, IsPtr: true}},
							}}},
							Processed: []string{"error{0x2fffffff, 0x2fffffff}"},
						},
						"/root/main.go",
						7),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"error unnamed",
			`import "errors"
			func main() {
				f(errors.New("ooh"))
			}
			func f(error) {
				panic("ooh")
			}`,
			true,
			goarchList{"386", "arm", "arm64", "mipsle", "mips64le", "ppc64le", "riscv64", "s390x"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{Value: pointer, IsPtr: true}, {Value: 3}},
							}}},
							Processed: []string{"error{0x2fffffff, 0x3}"},
						},
						"/root/main.go",
						7),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"float32",
			`func main() {
				f(0.5)
			}
			func f(v float32) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"0.5"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"float64",
			`func main() {
				f(0.5)
			}
			func f(v float64) {
				panic("ooh")
			}`,
			true,
			goarchList{"386", "arm", "mipsle"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						// The value is NOT a pointer but floating point encoding is not
						// deterministic.
						Args{
							Values:    []Arg{{Value: pointer, IsPtr: true}},
							Processed: []string{"0.5"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"uint",
			`func main() {
				f(123)
			}
			func f(v uint) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: 123}},
							Processed: []string{"123"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"negative int32",
			`func main() {
				f(-123)
			}
			func f(v int32) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values:    []Arg{{Value: negInt, IsPtr: negPtr}},
							Processed: []string{"-123"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"array",
			`func main() {
				f([3]byte{2, 3, 4})
			}
			func f(v2 [3]byte) {
				panic("ooh")
			}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{
								{IsAggregate: true, Fields: Args{
									Values: []Arg{{Value: 2}, {Value: 3}, {Value: 4}},
								}},
							},
							Processed: []string{"[3]byte{0x2, 0x3, 0x4}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"deeply nested aggregate type",
			`func main() {
				f(a{b{c{d{13}}}})
			}
			func f(v a) {
				panic("ooh")
			}
			type a struct{ b }
			type b struct{ c }
			type c struct{ d }
			type d struct{ i int }`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{IsAggregate: true, Fields: Args{
									Values: []Arg{{IsAggregate: true, Fields: Args{
										Values: []Arg{{IsAggregate: true, Fields: Args{
											Values: []Arg{{Value: 13}},
										}}},
									}}},
								}}},
							}}},
							Processed: []string{"a{0xd}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"deeply nested aggregate type with elision",
			`func main() {
				f(a{b{c{d{e{13}}}}})
			}
			func f(v a) {
				panic("ooh")
			}
			type a struct{ b }
			type b struct{ c }
			type c struct{ d }
			type d struct{ e }
			type e struct{ i int }`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{{IsAggregate: true, Fields: Args{
									Values: []Arg{{IsAggregate: true, Fields: Args{
										Values: []Arg{{IsAggregate: true, Fields: Args{
											Values: []Arg{{IsAggregate: true, Fields: Args{
												Elided: true,
											}}},
										}}},
									}}},
								}}},
							}}},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"argument offsets partially too large",
			`func main() {
				f(a{b{c{d{e{}}}}}, 13, []int{14, 15})
			}
			func f(v a, i int, s []int) {
				panic("ooh")
			}
			type a struct{ b }
			type b struct{ c }
			type c struct{ d }
			type d struct{ e }
			type e struct{ i [27]int }`,
			true,
			goarchList{"386", "arm", "mipsle"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{
								{IsAggregate: true, Fields: Args{
									Values: []Arg{{IsAggregate: true, Fields: Args{
										Values: []Arg{{IsAggregate: true, Fields: Args{
											Values: []Arg{{IsAggregate: true, Fields: Args{
												Values: []Arg{{IsAggregate: true, Fields: Args{
													Elided: true,
												}}},
											}}},
										}}},
									}}},
								}},
								{Value: 13},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: pointer, IsPtr: true},
										{Value: 2},
										{IsOffsetTooLarge: true},
									},
								}},
							},
							Processed: []string{"a{}", "13", "[]int(0x2fffffff len=2 cap=_)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"argument offsets entirely too large",
			`func main() {
				f(a{b{c{d{e{}}}}}, 13, []int{14, 15})
			}
			func f(v a, i int, s []int) {
				panic("ooh")
			}
			type a struct{ b }
			type b struct{ c }
			type c struct{ d }
			type d struct{ e }
			type e struct{ i [30]int }`,
			true,
			goarchList{"386", "arm", "mipsle"},
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.f",
						Args{
							Values: []Arg{
								{IsAggregate: true, Fields: Args{
									Values: []Arg{{IsAggregate: true, Fields: Args{
										Values: []Arg{{IsAggregate: true, Fields: Args{
											Values: []Arg{{IsAggregate: true, Fields: Args{
												Values: []Arg{{IsAggregate: true, Fields: Args{
													Elided: true,
												}}},
											}}},
										}}},
									}}},
								}},
								{IsOffsetTooLarge: true},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{IsOffsetTooLarge: true},
										{IsOffsetTooLarge: true},
										{IsOffsetTooLarge: true},
									},
								}},
							},
							Processed: []string{"a{}", "_", "[]int(_ len=_ cap=_)"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		// The following subtests are adapted from TestTracebackArgs in
		// src/runtime/traceback_test.go in the Go runtime.
		{
			"testTracebackArgs1",
			`func main() {
					testTracebackArgs1(1, 2, 3, 4, 5)
				}
				func testTracebackArgs1(a, b, c, d, e int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs1",
						Args{
							Values: []Arg{
								{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
							},
							Processed: []string{"1", "2", "3", "4", "5"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs2",
			`func main() {
					testTracebackArgs2(false, struct {
						a, b, c int
						x       [2]int
					}{1, 2, 3, [2]int{4, 5}}, [0]int{}, [3]byte{6, 7, 8})
				}
				func testTracebackArgs2(a bool, b struct {
					a, b, c int
					x       [2]int
				}, _ [0]int, d [3]byte) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs2",
						Args{
							Values: []Arg{
								{Value: 0},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: 1},
										{Value: 2},
										{Value: 3},
										{IsAggregate: true, Fields: Args{
											Values: []Arg{
												{Value: 4},
												{Value: 5},
											},
										}},
									},
								}},
								{IsAggregate: true},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: 6},
										{Value: 7},
										{Value: 8},
									},
								}},
							},
							Processed: []string{"false", "<unknown>{0x1, 0x2, 0x3, 0x4, 0x5}", "[0]int{}", "[3]byte{0x6, 0x7, 0x8}"},
						},
						"/root/main.go",
						12),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs3",
			`func main() {
					testTracebackArgs3([3]byte{1, 2, 3}, 4, 5, 6, [3]byte{7, 8, 9})
				}
				func testTracebackArgs3(x [3]byte, a, b, c int, y [3]byte) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs3",
						Args{
							Values: []Arg{
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: 1},
										{Value: 2},
										{Value: 3},
									},
								}},
								{Value: 4},
								{Value: 5},
								{Value: 6},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: 7},
										{Value: 8},
										{Value: 9},
									},
								}},
							},
							Processed: []string{"[3]byte{0x1, 0x2, 0x3}", "4", "5", "6", "[3]byte{0x7, 0x8, 0x9}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs4",
			`func main() {
					testTracebackArgs4(true, [1][1][1][1][1][1][1][1][1][1]int{})
				}
				func testTracebackArgs4(a bool, x [1][1][1][1][1][1][1][1][1][1]int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs4",
						Args{
							Values: []Arg{
								{Value: 1},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{{IsAggregate: true, Fields: Args{
										Values: []Arg{{IsAggregate: true, Fields: Args{
											Values: []Arg{{IsAggregate: true, Fields: Args{
												Values: []Arg{{IsAggregate: true, Fields: Args{
													Elided: true,
												}}},
											}}},
										}}},
									}}},
								}},
							},
							Processed: []string{"true"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs5",
			`func main() {
					z := [0]int{}
					testTracebackArgs5(false, struct {
						x int
						y [0]int
						z [2][0]int
					}{1, z, [2][0]int{}}, z, z, z, z, z, z, z, z, z, z, z, z)
				}
				func testTracebackArgs5(a bool, x struct {
					x int
					y [0]int
					z [2][0]int
				}, _, _, _, _, _, _, _, _, _, _, _, _ [0]int) {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs5",
						Args{
							Values: []Arg{
								{Value: 0},
								{IsAggregate: true, Fields: Args{
									Values: []Arg{
										{Value: 1},
										{IsAggregate: true},
										{IsAggregate: true, Fields: Args{
											Values: []Arg{
												{IsAggregate: true}, {IsAggregate: true},
											},
										}},
									},
								}},
								{IsAggregate: true}, {IsAggregate: true},
								{IsAggregate: true}, {IsAggregate: true},
								{IsAggregate: true},
							},
							Processed: []string{"false", "<unknown>{0x1}"},
							Elided:    true,
						},
						"/root/main.go",
						15),
					newCallSrc("main.main", Args{}, "/root/main.go", 4),
				},
			},
		},
		{
			"testTracebackArgs6a",
			`func main() {
					testTracebackArgs6a(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
				}
				func testTracebackArgs6a(a, b, c, d, e, f, g, h, i, j int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs6a",
						Args{
							Values: []Arg{
								{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
								{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
							},
							Processed: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs6b",
			`func main() {
					testTracebackArgs6b(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
				}
				func testTracebackArgs6b(a, b, c, d, e, f, g, h, i, j, k int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs6b",
						Args{
							Values: []Arg{
								{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
								{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
							},
							Processed: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
							Elided:    true,
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs7a",
			`func main() {
					testTracebackArgs7a([10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
				}
				func testTracebackArgs7a(a [10]int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs7a",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
									{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
								},
							}}},
							Processed: []string{"[10]int{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs7b",
			`func main() {
					testTracebackArgs7b([11]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
				}
				func testTracebackArgs7b(a [11]int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs7b",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
									{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
								},
								Elided: true,
							}}},
							Processed: []string{"[11]int{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, ...}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs7c",
			`func main() {
					testTracebackArgs7c([10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 11)
				}
				func testTracebackArgs7c(a [10]int, b int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs7c",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
									{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
								},
							}}},
							Processed: []string{"[10]int{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa}"},
							Elided:    true,
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs7d",
			`func main() {
					testTracebackArgs7d([11]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 12)
				}
				func testTracebackArgs7d(a [11]int, b int) int {
					panic("ooh")
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs7d",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
									{Value: 6}, {Value: 7}, {Value: 8}, {Value: 9}, {Value: 10},
								},
								Elided: true,
							}}},
							Processed: []string{"[11]int{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, ...}"},
							Elided:    true,
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs8a",
			`func main() {
					testTracebackArgs8a(testArgsType8a{1, 2, 3, 4, 5, 6, 7, 8, [2]int{9, 10}})
				}
				func testTracebackArgs8a(a testArgsType8a) int {
					panic("ooh")
				}
				type testArgsType8a struct {
					a, b, c, d, e, f, g, h int
					i                      [2]int
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs8a",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
									{Value: 5}, {Value: 6}, {Value: 7}, {Value: 8},
									{IsAggregate: true, Fields: Args{
										Values: []Arg{{Value: 9}, {Value: 10}},
									}},
								},
							}}},
							Processed: []string{"testArgsType8a{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs8b",
			`func main() {
					testTracebackArgs8b(testArgsType8b{1, 2, 3, 4, 5, 6, 7, 8, [3]int{9, 10, 11}})
				}
				func testTracebackArgs8b(a testArgsType8b) int {
					panic("ooh")
				}
				type testArgsType8b struct {
					a, b, c, d, e, f, g, h int
					i                      [3]int
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs8b",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
									{Value: 5}, {Value: 6}, {Value: 7}, {Value: 8},
									{IsAggregate: true, Fields: Args{
										Values: []Arg{{Value: 9}, {Value: 10}},
										Elided: true,
									}},
								},
							}}},
							Processed: []string{"testArgsType8b{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs8c",
			`func main() {
					testTracebackArgs8c(testArgsType8c{1, 2, 3, 4, 5, 6, 7, 8, [2]int{9, 10}, 11})
				}
				func testTracebackArgs8c(a testArgsType8c) int {
					panic("ooh")
				}
				type testArgsType8c struct {
					a, b, c, d, e, f, g, h int
					i                      [2]int
					j                      int
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs8c",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
									{Value: 5}, {Value: 6}, {Value: 7}, {Value: 8},
									{IsAggregate: true, Fields: Args{
										Values: []Arg{{Value: 9}, {Value: 10}},
									}},
								},
								Elided: true,
							}}},
							Processed: []string{"testArgsType8c{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, ...}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
		{
			"testTracebackArgs8d",
			`func main() {
					testTracebackArgs8d(testArgsType8d{1, 2, 3, 4, 5, 6, 7, 8, [3]int{9, 10, 11}, 12})
				}
				func testTracebackArgs8d(a testArgsType8d) int {
					panic("ooh")
				}
				type testArgsType8d struct {
					a, b, c, d, e, f, g, h int
					i                      [3]int
					j                      int
				}`,
			true,
			allPlatforms,
			Stack{
				Calls: []Call{
					newCallSrc(
						"main.testTracebackArgs8d",
						Args{
							Values: []Arg{{IsAggregate: true, Fields: Args{
								Values: []Arg{
									{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
									{Value: 5}, {Value: 6}, {Value: 7}, {Value: 8},
									{IsAggregate: true, Fields: Args{
										Values: []Arg{{Value: 9}, {Value: 10}},
										Elided: true,
									}},
								},
								Elided: true,
							}}},
							Processed: []string{"testArgsType8d{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, ...}"},
						},
						"/root/main.go",
						6),
					newCallSrc("main.main", Args{}, "/root/main.go", 3),
				},
			},
		},
	}

	for i, line := range data {
		i := i
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			t.Parallel()
			for _, arch := range line.archBlock {
				if runtime.GOARCH == arch {
					t.Skipf("skipping test on platform %s", arch)
				}
			}
			// Marshal the code a bit to make it nicer. Inject 'package main'.
			lines := append([]string{"package main"}, strings.Split(line.input, "\n")...)
			for j := 2; j < len(lines); j++ {
				// Strip the 3 first tab characters. It's very adhoc but good enough here
				// and makes test failure much more readable.
				if lines[j][:3] != "\t\t\t" {
					t.Fatal("expected line to start with 3 tab characters")
				}
				lines[j] = lines[j][3:]
			}
			input := strings.Join(lines, "\n")

			// Create one temporary directory by subtest.
			root, err := ioutil.TempDir("", "stack")
			if err != nil {
				t.Fatalf("failed to create temporary directory: %v", err)
			}
			defer func() {
				if err2 := os.RemoveAll(root); err2 != nil {
					t.Fatalf("failed to remove temporary directory %q: %v", root, err2)
				}
			}()
			main := filepath.Join(root, "main.go")
			if err := ioutil.WriteFile(main, []byte(input), 0500); err != nil {
				t.Fatalf("failed to write %q: %v", main, err)
			}

			if runtime.GOOS == "windows" {
				root = strings.Replace(root, pathSeparator, "/", -1)
			}
			const prefix = "/root"
			for j, c := range line.want.Calls {
				if strings.HasPrefix(c.RemoteSrcPath, prefix) {
					line.want.Calls[j].RemoteSrcPath = root + c.RemoteSrcPath[len(prefix):]
				}
				if strings.HasPrefix(c.LocalSrcPath, prefix) {
					line.want.Calls[j].LocalSrcPath = root + c.LocalSrcPath[len(prefix):]
				}
				if strings.HasPrefix(c.DirSrc, "root") {
					line.want.Calls[j].DirSrc = path.Base(root) + c.DirSrc[4:]
				}
			}

			// Run the command up to twice.

			// Only disable inlining if necessary.
			disableInline := line.mayBeInlined
			content := getCrash(t, main, disableInline)
			t.Log("First")
			// Warning: this function modifies want.
			testAugmentCommon(t, content, false, line.want)

			// If inlining was disabled, try a second time but zap things out.
			if disableInline {
				for j := range line.want.Calls {
					line.want.Calls[j].Args.Processed = nil
				}
				content := getCrash(t, main, false)
				t.Log("Second")
				testAugmentCommon(t, content, line.mayBeInlined, line.want)
			}
		})
	}
}

func testAugmentCommon(t *testing.T, content []byte, mayBeInlined bool, want Stack) {
	// Analyze it.
	prefix := bytes.Buffer{}
	s, suffix, err := ScanSnapshot(bytes.NewBuffer(content), &prefix, defaultOpts())
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}
	// On go1.4, there's one less empty line.
	if got := prefix.String(); got != "panic: ooh\n\n" && got != "panic: ooh\n" {
		t.Fatalf("Unexpected panic output:\n%#v", got)
	}
	compareString(t, "exit status 2\n", string(suffix))
	if !s.guessPaths() {
		t.Error("expected success")
	}

	if err := s.augment(); err != nil {
		t.Errorf("augment() returned %v", err)
	}
	got := s.Goroutines[0].Signature.Stack
	zapPointers(t, &want, &got)

	// On go1.11 with non-pointer method, it shows elided argument where
	// there used to be none before. It's only for test case "non-pointer
	// method".
	if mayBeInlined {
		for j := range got.Calls {
			if !want.Calls[j].Args.Elided {
				got.Calls[j].Args.Elided = false
			}
			if got.Calls[j].Args.Values == nil {
				want.Calls[j].Args.Values = nil
			}
		}
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Logf("Different (-want +got):\n%s", diff)
		t.Logf("Output:\n%s", content)
		t.FailNow()
	}
}

func TestAugmentErr(t *testing.T) {
	t.Parallel()
	root, err := ioutil.TempDir("", "stack")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer func() {
		if err2 := os.RemoveAll(root); err2 != nil {
			t.Fatalf("failed to remove temporary directory %q: %v", root, err2)
		}
	}()

	tree := map[string]string{
		"bad.go":       "bad content",
		"foo.asm":      "; good but ignored",
		"good.go":      "package main",
		"no_access.go": "package main",
	}
	createTree(t, root, tree)

	type dataLine struct {
		name  string
		src   string
		line  int
		args  Args
		errRe string
	}
	// Note: these tests assumes an OS running in English-US locale. That should
	// eventually be fixed, maybe by using regexes?
	msgRe := ".+"
	if runtime.GOOS != "windows" {
		msgRe = "no such file or directory"
	}
	data := []dataLine{
		{
			name:  "assembly is skipped",
			src:   "foo.asm",
			args:  Args{Values: []Arg{{}}},
			errRe: regexp.QuoteMeta(fmt.Sprintf("cannot load non-go file %q", filepath.Join(root, "foo.asm"))),
		},
		{
			name: "assembly is skipped (no arg)",
			src:  "foo.asm",
		},
		{
			name:  "invalid line number",
			src:   "good.go",
			line:  2,
			args:  Args{Values: []Arg{{}}},
			errRe: "line 2 is over line count of 1",
		},
		{
			name: "invalid line number (no arg)",
			src:  "good.go",
			line: 2,
		},
		{
			name:  "missing file",
			src:   "missing.go",
			args:  Args{Values: []Arg{{}}},
			errRe: regexp.QuoteMeta(fmt.Sprintf("open %s: ", filepath.Join(root, "missing.go"))) + msgRe,
		},
		{
			name: "missing file (no arg)",
			src:  "missing.go",
		},
		{
			name: "invalid go code (no arg)",
			src:  "bad.go",
		},
		{
			name: "no I/O access (no arg)",
			src:  "no_access.go",
		},
	}
	if internaltest.GetGoMinorVersion() > 11 {
		// The format changed between 1.9 and 1.12.
		data = append(data, dataLine{
			name:  "invalid go code",
			src:   "bad.go",
			args:  Args{Values: []Arg{{}}},
			errRe: regexp.QuoteMeta(fmt.Sprintf("failed to parse %s:1:1: expected 'package', found bad", filepath.Join(root, "bad.go"))),
		})
	}
	if runtime.GOOS != "windows" {
		// Chmod has no effect on Windows.
		compareErr(t, nil, os.Chmod(filepath.Join(root, "no_access.go"), 0))
		data = append(data, dataLine{
			name:  "no I/O access",
			src:   "no_access.go",
			args:  Args{Values: []Arg{{}}},
			errRe: regexp.QuoteMeta(fmt.Sprintf("open %s: permission denied", filepath.Join(root, "no_access.go"))),
		})
	}

	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("%d-%s", i, line.name), func(t *testing.T) {
			l := line.line
			if l == 0 {
				l = 1
			}
			s := Snapshot{
				Goroutines: []*Goroutine{
					{Signature: Signature{Stack: Stack{Calls: []Call{
						{LocalSrcPath: filepath.Join(root, line.src), Args: line.args, Line: l},
					}}}},
				}}
			if err := s.augment(); (line.errRe == "") != (err == nil) {
				t.Fatalf("want: %q; got:  %q", line.errRe, err)
			} else if err != nil {
				if m, err2 := regexp.MatchString(line.errRe, err.Error()); err2 != nil {
					t.Fatal(err2)
				} else if !m {
					t.Fatalf("want: %q; got: %q", line.errRe, err)
				}
			}
		})
	}
}

func TestLineToByteOffsets(t *testing.T) {
	src := "\n\n\n"
	want := []int{0, 0, 1, 2, 3}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
	}
	src = "hello"
	want = []int{0, 0}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
	}
	src = "this\nis\na\ntest"
	want = []int{0, 0, 5, 8, 10}
	if diff := cmp.Diff(want, lineToByteOffsets([]byte(src))); diff != "" {
		t.Error(diff)
	}
}

//

const pointer = uint64(0x2fffffff)
const pointerStr = "0x2fffffff"
const is64Bit = uint64(^uintptr(0)) == ^uint64(0)

func overrideEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func getCrash(t *testing.T, main string, disableInline bool) []byte {
	args := []string{"run"}
	if disableInline {
		// Disable both optimization (-N) and inlining (-l).
		args = append(args, "-gcflags", "-N -l")
	}
	cmd := exec.Command("go", append(args, main)...)
	// Use the Go 1.4 compatible format.
	cmd.Env = overrideEnv(os.Environ(), "GOTRACEBACK", "1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Helper()
		t.Fatal("expected error since this is supposed to crash")
	}
	return out
}

// zapPointers zaps out pointers in got.
func zapPointers(t *testing.T, want, got *Stack) {
	for i := range got.Calls {
		if i >= len(want.Calls) {
			// When using GOTRACEBACK=2, it'll include runtime.main() and
			// runtime.goexit(). Ignore these since they could be changed in a future
			// version.
			got.Calls = got.Calls[:len(want.Calls)]
			break
		}
		gotArgs := got.Calls[i].Args
		ptrsToReplace := zapPointersInArgs(t, want.Calls[i].Args, gotArgs)
		for _, ptr := range ptrsToReplace {
			for k := range gotArgs.Processed {
				gotArgs.Processed[k] = strings.Replace(gotArgs.Processed[k], ptr, pointerStr, -1)
			}
		}
	}
}

func zapPointersInArgs(t *testing.T, want, got Args) (ptrs []string) {
	for j := range got.Values {
		if j >= len(want.Values) {
			break
		}
		if want.Values[j].IsAggregate && got.Values[j].IsAggregate {
			ptrs = append(ptrs, zapPointersInArgs(t, want.Values[j].Fields, got.Values[j].Fields)...)
		} else if want.Values[j].IsPtr && got.Values[j].IsPtr {
			// Record the existing pointer value and then replace.
			ptrs = append(ptrs, fmt.Sprintf("0x%x", got.Values[j].Value))
			got.Values[j].Value = want.Values[j].Value
		}
	}
	return ptrs
}
