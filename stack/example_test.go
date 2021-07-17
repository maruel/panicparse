// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"

	"github.com/maruel/panicparse/v2/stack"
)

// Runs a crashing program and converts it to a dense text format like pp does.
func Example_text() {
	source := `package main

	import "time"

	func main() {
		c := crashy{}
		go func() {
			c.die(42.)
		}()
		select {}
	}

	type crashy struct{}

	func (c crashy) die(f float64) {
		time.Sleep(time.Millisecond)
		panic(int(f))
	}`

	// Skipped error handling to make the example shorter.
	root, _ := ioutil.TempDir("", "stack")
	defer os.RemoveAll(root)
	p := filepath.Join(root, "main.go")
	ioutil.WriteFile(p, []byte(source), 0600)
	// Disable both optimization (-N) and inlining (-l).
	c := exec.Command("go", "run", "-gcflags", "-N -l", p)
	// This is important, otherwise only the panicking goroutine will be printed.
	c.Env = append(os.Environ(), "GOTRACEBACK=1")
	raw, _ := c.CombinedOutput()
	stream := bytes.NewReader(raw)

	s, suffix, err := stack.ScanSnapshot(stream, os.Stdout, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	// Find out similar goroutine traces and group them into buckets.
	buckets := s.Aggregate(stack.AnyValue).Buckets

	// Calculate alignment.
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(fmt.Sprintf("%s:%d", line.SrcName, line.Line)); l > srcLen {
				srcLen = l
			}
			if l := len(filepath.Base(line.Func.ImportPath)); l > pkgLen {
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

		if len(bucket.CreatedBy.Calls) != 0 {
			extra += fmt.Sprintf(" [Created by %s.%s @ %s:%d]", bucket.CreatedBy.Calls[0].Func.DirName, bucket.CreatedBy.Calls[0].Func.Name, bucket.CreatedBy.Calls[0].SrcName, bucket.CreatedBy.Calls[0].Line)
		}
		fmt.Printf("%d: %s%s\n", len(bucket.IDs), bucket.State, extra)

		// Print the stack lines.
		for _, line := range bucket.Stack.Calls {
			fmt.Printf(
				"    %-*s %-*s %s(%s)\n",
				pkgLen, line.Func.DirName, srcLen,
				fmt.Sprintf("%s:%d", line.SrcName, line.Line),
				line.Func.Name, &line.Args)
		}
		if bucket.Stack.Elided {
			io.WriteString(os.Stdout, "    (...)\n")
		}
	}

	// If there was any remaining data in the pipe, dump it now.
	if len(suffix) != 0 {
		os.Stdout.Write(suffix)
	}
	if err == nil {
		io.Copy(os.Stdout, stream)
	}

	// Output:
	// panic: 42
	//
	// 1: running [Created by main.main @ main.go:7]
	//     main main.go:17 crashy.die(42)
	//     main main.go:8  main.func1()
	// 1: select (no cases)
	//     main main.go:10 main()
	// exit status 2
}

// Process multiple consecutive goroutine snapshots.
func Example_stream() {
	// Stream of stack traces:
	var r io.Reader
	var w io.Writer
	opts := stack.DefaultOpts()
	for {
		s, suffix, err := stack.ScanSnapshot(r, w, opts)
		if s != nil {
			// Process the snapshot...
		}

		if err != nil && err != io.EOF {
			if len(suffix) != 0 {
				w.Write(suffix)
			}
			log.Fatal(err)
		}
		// Prepend the suffix that was read to the rest of the input stream to
		// catch the next snapshot signature:
		r = io.MultiReader(bytes.NewReader(suffix), r)
	}
}

// Converts a stack trace from os.Stdin into HTML on os.Stdout, discarding
// everything else.
func Example_hTML() {
	s, _, err := stack.ScanSnapshot(os.Stdin, ioutil.Discard, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	if s != nil {
		s.Aggregate(stack.AnyValue).ToHTML(os.Stdout, "")
	}
}

// A sample parseStack function expects a stdlib stacktrace from runtime.Stack or debug.Stack and returns
// the parsed stack object.
func Example_simple() {
	parseStack := func(rawStack []byte) stack.Stack {
		s, _, err := stack.ScanSnapshot(bytes.NewReader(rawStack), ioutil.Discard, stack.DefaultOpts())
		if err != nil && err != io.EOF {
			panic(err)
		}

		if len(s.Goroutines) > 1 {
			panic(errors.New("provided stacktrace had more than one goroutine"))
		}

		return s.Goroutines[0].Signature.Stack
	}

	parsedStack := parseStack(debug.Stack())
	fmt.Printf("parsedStack: %#v", parsedStack)
}
