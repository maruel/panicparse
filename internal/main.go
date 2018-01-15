// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal implements panicparse
//
// It is mostly useful on servers will large number of identical goroutines,
// making the crash dump harder to read than strictly necessary.
//
// Colors:
//  - Magenta: first goroutine to be listed.
//  - Yellow: main package.
//  - Green: standard library.
//  - Red: other packages.
//
// Bright colors are used for exported symbols.
package internal

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/maruel/panicparse/stack"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
)

// resetFG is similar to ansi.Reset except that it doesn't reset the
// background color, only the foreground color and the style.
//
// That much for the "ansi" abstraction layer...
const resetFG = ansi.DefaultFG + "\033[m"

// defaultPalette is the default recommended palette.
var defaultPalette = stack.Palette{
	EOLReset:               resetFG,
	RoutineFirst:           ansi.ColorCode("magenta+b"),
	CreatedBy:              ansi.LightBlack,
	Package:                ansi.ColorCode("default+b"),
	SourceFile:             resetFG,
	FunctionStdLib:         ansi.Green,
	FunctionStdLibExported: ansi.ColorCode("green+b"),
	FunctionMain:           ansi.ColorCode("yellow+b"),
	FunctionOther:          ansi.Red,
	FunctionOtherExported:  ansi.ColorCode("red+b"),
	Arguments:              resetFG,
}

func writeToConsole(out io.Writer, p *stack.Palette, buckets stack.Buckets, fullPath, needsEnv bool) error {
	if needsEnv {
		_, _ = io.WriteString(out, "\nTo see all goroutines, visit https://github.com/maruel/panicparse#gotraceback\n\n")
	}
	srcLen, pkgLen := stack.CalcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		_, _ = io.WriteString(out, p.BucketHeader(&bucket, fullPath, len(buckets) > 1))
		_, _ = io.WriteString(out, p.StackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	return nil
}

// process copies stdin to stdout and processes any "panic: " line found.
//
// If html is used, a stack trace is written to this file instead.
func process(in io.Reader, out io.Writer, p *stack.Palette, s stack.Similarity, fullPath, parse, rebase bool, html string) error {
	if !rebase {
		stack.NoRebase()
	}
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	needsEnv := len(goroutines) == 1 && showBanner()
	if parse {
		stack.Augment(goroutines)
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, s))
	if html == "" {
		return writeToConsole(out, p, buckets, fullPath, needsEnv)
	}
	return writeToHTML(html, buckets, needsEnv)
}

func showBanner() bool {
	if !showGOTRACEBACKBanner {
		return false
	}
	gtb := os.Getenv("GOTRACEBACK")
	return gtb == "" || gtb == "single"
}

// Main is implemented here so both 'pp' and 'panicparse' executables can be
// compiled. This is to work around the Perl Package manager 'pp' that is
// preinstalled on some OSes.
func Main() error {
	aggressive := flag.Bool("aggressive", false, "Aggressive deduplication including non pointers")
	parse := flag.Bool("parse", true, "Parses source files to deduct types; use -parse=false to work around bugs in source parser")
	rebase := flag.Bool("rebase", true, "Guess GOROOT and GOPATH")
	verboseFlag := flag.Bool("v", false, "Enables verbose logging output")
	// Console only.
	fullPath := flag.Bool("full-path", false, "Print full sources path")
	noColor := flag.Bool("no-color", !isatty.IsTerminal(os.Stdout.Fd()) || os.Getenv("TERM") == "dumb", "Disable coloring")
	forceColor := flag.Bool("force-color", false, "Forcibly enable coloring when with stdout is redirected")
	// HTML only.
	html := flag.String("html", "", "Output an HTML file")
	flag.Parse()

	log.SetFlags(log.Lmicroseconds)
	if !*verboseFlag {
		log.SetOutput(ioutil.Discard)
	}

	s := stack.AnyPointer
	if *aggressive {
		s = stack.AnyValue
	}

	var out io.Writer = os.Stdout
	p := &defaultPalette
	if *html == "" {
		if *noColor && !*forceColor {
			p = &stack.Palette{}
		} else {
			out = colorable.NewColorableStdout()
		}
	}

	var in *os.File
	switch flag.NArg() {
	case 0:
		in = os.Stdin
		// Explicitly silence SIGQUIT, as it is useful to gather the stack dump
		// from the piped command..
		signals := make(chan os.Signal)
		go func() {
			for {
				<-signals
			}
		}()
		signal.Notify(signals, os.Interrupt, syscall.SIGQUIT)

	case 1:
		// Do not handle SIGQUIT when passed a file to process.
		var err error
		name := flag.Arg(0)
		if in, err = os.Open(name); err != nil {
			return fmt.Errorf("did you mean to specify a valid stack dump file name? %s", err)
		}
		defer in.Close()

	default:
		return errors.New("pipe from stdin or specify a single file")
	}

	return process(in, out, p, s, *fullPath, *parse, *rebase, *html)
}
