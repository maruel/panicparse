// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panicparse: analyzes stack dump of Go processes and simplifies it.
//
// It is mostly useful on servers will large number of identical goroutines,
// making the crash dump harder to read than strictly necesary.
//
// Colors:
//  - Magenta: first goroutine to be listed.
//  - Yellow: main package.
//  - Green: standard library.
//  - Red: other packages.
//
// Bright colors are used for exported symbols.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/maruel/panicparse/stack"
	"github.com/mgutz/ansi"
)

// BUG: Support Windows. https://github.com/shiena/ansicolor seems like a good
// candidate.

var (
	all = flag.Bool("all", false, "print all output before the stack dump")
)

func PrettyStack(r *stack.Signature) string {
	out := []string{}
	srcLen := 0
	pkgLen := 0
	for _, line := range r.Stack {
		l := len(line.SourceLine())
		if l > srcLen {
			srcLen = l
		}
		l = len(line.Func.PkgName())
		if l > pkgLen {
			pkgLen = l
		}
	}
	for _, line := range r.Stack {
		c := ansi.Red
		if line.IsStdlib() {
			if line.Func.IsExported() {
				c = ansi.LightGreen
			} else {
				c = ansi.Green
			}
		} else if line.IsPkgMain() {
			c = ansi.LightYellow
		} else if line.Func.IsExported() {
			c = ansi.LightRed
		}
		s := fmt.Sprintf(
			"  %s%-*s%s %-*s %s%s%s(%s)",
			ansi.LightWhite, pkgLen, line.Func.PkgName(), ansi.Reset,
			srcLen, line.SourceLine(),
			c, line.Func.Name(), ansi.Reset, line.Args)
		out = append(out, s)
	}
	return strings.Join(out, "\n")
}

func mainImpl() error {
	c := make(chan os.Signal)
	go func() {
		for {
			<-c
		}
	}()
	signal.Notify(c, os.Interrupt)

	flag.Parse()
	var in *os.File
	switch name := flag.Arg(0); {
	case name == "":
		in = os.Stdin
	default:
		var err error
		if in, err = os.Open(name); err != nil {
			return err
		}
		defer in.Close()
	}

	header, goroutines, err := stack.ParseDump(in)
	if err != nil {
		return err
	}
	if *all {
		fmt.Printf("%s\n", header)
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines))
	for _, bucket := range buckets {
		extra := ""
		created := bucket.CreatedBy.String()
		if created != "" {
			extra += " [Created by " + created + "]"
		}
		c := ansi.White
		if bucket.First() && len(buckets) > 1 {
			c = ansi.LightMagenta
		}

		fmt.Printf("%s%d: %s%s%s\n", c, len(bucket.Routines), bucket.State, extra, ansi.Reset)
		fmt.Printf("%s\n", PrettyStack(&bucket.Signature))
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}
