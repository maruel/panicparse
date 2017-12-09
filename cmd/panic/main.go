// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panic crashes in various ways.
//
// It is a tool to help test pp.
package main

// To install, run:
//   go install github.com/maruel/panicparse/cmd/panic
//   panic -help
//   panic str |& pp
//
// Some panics require the race detector with -race:
//   go install -race github.com/maruel/panicparse/cmd/panic
//   panic race |& pp
//
// To add a new panic stack signature, add it to types type below, keeping the
// list ordered by name. If you need utility functions, add it in the section
// below. That's it!

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/maruel/panicparse/cmd/panic/internal"
)

// Utility functions.

func panicint(i int) {
	panic(i)
}

func panicstr(a string) {
	panic(a)
}

func panicslicestr(a []string) {
	panic(a)
}

func panicArgsElided(a, b, c, d, e, f, g, h, i, j, k int) {
	panic(a)
}

func recurse(i int) {
	if i > 0 {
		recurse(i - 1)
		return
	}
	panic(42)
}

//

// types is all the supported types of panics.
//
// Keep the list sorted.
//
// TODO(maruel): Figure out a way to reliably trigger "(scan)" output:
// - disable automatic GC with runtime.SetGCPercent(-1)
// - a goroutine with a large number of items in the stack
// - large heap to make the scanning process slow enough
// - trigger a manual GC with go runtime.GC()
// - panic in the meantime
// This would still not be deterministic.
//
// TODO(maruel): Figure out a way to reliably trigger sleep output.
var types = map[string]struct {
	desc string
	f    func()
}{
	"args_elided": {
		"too many args in stack line, causing the call arguments to be elided",
		func() {
			panicArgsElided(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
		},
	},

	"chan_receive": {
		"goroutine blocked on <-c",
		func() {
			c := make(chan bool)
			go func() {
				<-c
				<-c
			}()
			c <- true
			panic(42)
		},
	},

	"chan_send": {
		"goroutine blocked on c<-",
		func() {
			c := make(chan bool)
			go func() {
				c <- true
				c <- true
			}()
			<-c
			panic(42)
		},
	},

	"goroutine_1": {
		"panic in one goroutine",
		func() {
			go func() {
				panicint(42)
			}()
			time.Sleep(time.Minute)
		},
	},

	"goroutine_100": {
		"start 100 goroutines before panicking",
		func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					wg.Done()
					time.Sleep(time.Minute)
				}()
			}
			wg.Wait()
			panicint(42)
		},
	},

	"goroutine_dedupe_pointers": {
		"start 100 goroutines with different pointers before panicking",
		func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(b *int) {
					wg.Done()
					time.Sleep(time.Minute)
				}(new(int))
			}
			wg.Wait()
			panicint(42)
		},
	},

	"int": {
		"panic(42)",
		func() {
			panicint(42)
		},
	},

	"locked": {
		"thread locked goroutine via runtime.LockOSThread()",
		func() {
			runtime.LockOSThread()
			panic(42)
		},
	},

	"other": {
		"panics with other package in the call stack, with both exported and unexpected functions",
		func() {
			internal.Callback(func() {
				panic("allo")
			})
		},
	},

	"race": {
		"will cause a crash by -race detector",
		panicRace,
	},

	"stack_cut_off": {
		"too many call lines in traceback, causing higher up calls to missing",
		func() {
			// Observed limit is 99.
			recurse(100)
		},
	},

	"simple": {
		// This is not used for real, here for documentation.
		"skip the map for a shorter stack trace",
		func() {},
	},

	"slice_str": {
		"panic([]string{\"allo\"}) with cap=2",
		func() {
			a := make([]string, 1, 2)
			a[0] = "allo"
			panicslicestr(a)
		},
	},

	"stdlib": {
		"panics with stdlib in the call stack, with both exported and unexpected functions",
		func() {
			a := []string{"a", "b"}
			sort.Slice(a, func(i, j int) bool {
				panic("allo")
			})
		},
	},

	"stdlib_and_other": {
		"panics with both other and stdlib packages in the call stack",
		func() {
			a := []string{"a", "b"}
			sort.Slice(a, func(i, j int) bool {
				internal.Callback(func() {
					panic("allo")
				})
				return false
			})
		},
	},

	"str": {
		"panic(\"allo\")",
		func() {
			panicstr("allo")
		},
	},
}

//

func main() {
	if len(os.Args) == 2 {
		n := os.Args[1]
		if f, ok := types[n]; ok {
			fmt.Printf("GOTRACEBACK=%s\n", os.Getenv("GOTRACEBACK"))
			if n == "simple" {
				// Since the map lookup creates another call stack entry, add a one-off
				// "simple" panic style to test the very minimal case.
				// types["simple"].f is never called.
				panic("simple")
			}
			f.f()
		}
		fmt.Fprintf(os.Stderr, "unknown panic style %q\n", n)
		os.Exit(1)
	}
	usage()
}

func usage() {
	t := `usage: panic <way>

This tool is meant to be used with pp to test different parsing scenarios and
ensure output on different version of the Go toolchain can be successfully
parsed.

Set GOTRACEBACK before running this tool to see how it affects the panic output.

Select the way to panic:
`
	io.WriteString(os.Stderr, t)
	names := make([]string, 0, len(types))
	m := 0
	for n := range types {
		names = append(names, n)
		if i := len(n); i > m {
			m = i
		}
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(os.Stderr, "- %-*s  %s\n", m, n, types[n].desc)
	}
	os.Exit(2)
}
