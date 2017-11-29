// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Panic crashes in various ways.
//
// It is a tool to help test panicparse.
package main

// To install, run:
//   go install github.com/maruel/panicparse/internal/panic
//   panic -help
//
// You can also run directly:
//   go run ./internal/panic/main.go str |& pp
//
// To add a new panic stack signature, add it to types type below, keeping the
// list ordered by name. If you need utility functions, add it in the section
// below. That's it!

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"
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

//

// types is all the supported types of panics.
//
// Keep the list sorted.
var types = map[string]struct {
	desc string
	f    func()
}{
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

	"int": {
		"panic(42)",
		func() {
			panicint(42)
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

	"str": {
		"panic(\"allo\")",
		func() {
			panicstr("allo")
		},
	},
}

//

func main() {
	fmt.Printf("GOTRACEBACK=%s\n", os.Getenv("GOTRACEBACK"))
	if len(os.Args) == 2 {
		n := os.Args[1]
		if n == "simple" {
			// Since the map lookup creates another call stack entry, add a one-off
			// "simple" to test the very minimal case.
			panic("simple")
		}
		if f, ok := types[n]; ok {
			f.f()
		}
	}
	usage()
}

func usage() {
	t := `usage: panic <way>

This tool is meant to be used with panicparse to test different parsing
scenarios and ensure output on different version of the Go toolchain can be
successfully parsed.

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
