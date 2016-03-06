// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal implements panicparse
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
	"strings"
	"syscall"

	"github.com/maruel/panicparse/Godeps/_workspace/src/github.com/mattn/go-colorable"
	"github.com/maruel/panicparse/Godeps/_workspace/src/github.com/mattn/go-isatty"
	"github.com/maruel/panicparse/Godeps/_workspace/src/github.com/mgutz/ansi"
	"github.com/maruel/panicparse/stack"
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

// process copies stdin to stdout and processes any "panic: " line found.
func process(in io.Reader, out io.Writer, p *stack.Palette, s stack.Similarity, fullPath bool, filterPackage, filterPath string) error {
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	if len(goroutines) == 1 && showBanner() {
		_, _ = io.WriteString(out, "\nTo see all goroutines, visit https://github.com/maruel/panicparse#GOTRACEBACK\n\n")
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, s))
	filtered := make([]stack.Bucket, 0, len(buckets))
	for _, b := range buckets {
		include := (filterPackage == "" && filterPath == "")
		for _, c := range b.Stack.Calls {
			if filterPackage != "" && c.Func.PkgName() == filterPackage {
				include = true
			}
			if filterPath != "" && strings.Contains(c.SourcePath, filterPath) {
				include = true
			}
		}
		if include {
			filtered = append(filtered, b)
		}
	}
	buckets = filtered
	srcLen, pkgLen := stack.CalcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		_, _ = io.WriteString(out, p.BucketHeader(&bucket, fullPath, len(buckets) > 1))
		_, _ = io.WriteString(out, p.StackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	return err
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
	signals := make(chan os.Signal)
	go func() {
		for {
			<-signals
		}
	}()
	signal.Notify(signals, os.Interrupt, syscall.SIGQUIT)
	aggressive := flag.Bool("aggressive", false, "Aggressive deduplication including non pointers")
	fullPath := flag.Bool("full-path", false, "Print full sources path")
	noColor := flag.Bool("no-color", !isatty.IsTerminal(os.Stdout.Fd()) || os.Getenv("TERM") == "dumb", "Disable coloring")
	forceColor := flag.Bool("force-color", false, "Forcibly enable coloring when with stdout is redirected")
	verboseFlag := flag.Bool("v", false, "Enables verbose logging output")
	filterPackage := flag.String("filter-package", "", "only show stacktraces that involve specified package")
	filterPath := flag.String("filter-path", "", "only show stacktraces that involve a file whose path includes this substring")
	flag.Parse()

	log.SetFlags(log.Lmicroseconds)
	if !*verboseFlag {
		log.SetOutput(ioutil.Discard)
	}

	s := stack.AnyPointer
	if *aggressive {
		s = stack.AnyValue
	}

	var out io.Writer
	p := &defaultPalette
	if *noColor && !*forceColor {
		p = &stack.Palette{}
		out = os.Stdout
	} else {
		out = colorable.NewColorableStdout()
	}

	var in *os.File
	switch flag.NArg() {
	case 0:
		in = os.Stdin
	case 1:
		var err error
		name := flag.Arg(0)
		if in, err = os.Open(name); err != nil {
			return fmt.Errorf("did you mean to specify a valid stack dump file name? %s", err)
		}
		defer in.Close()
	default:
		return errors.New("pipe from stdin or specify a single file")
	}
	return process(in, out, p, s, *fullPath, *filterPackage, *filterPath)
}
