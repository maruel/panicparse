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

// Palette defines the color used.
//
// An empty object Palette{} can be used to disable coloring.
type Palette struct {
	EOLReset string

	// Routine header.
	RoutineFirst string
	Routine      string
	CreatedBy    string

	// Call line.
	Package                string
	SourceFile             string
	FunctionStdLib         string
	FunctionStdLibExported string
	FunctionMain           string
	FunctionOther          string
	FunctionOtherExported  string
	Arguments              string
}

// DefaultPalette is the default recommended palette.
var DefaultPalette = Palette{
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

// calcLengths returns the maximum length of the source lines and package names.
func calcLengths(buckets stack.Buckets, fullPath bool) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			l := 0
			if fullPath {
				l = len(line.FullSourceLine())
			} else {
				l = len(line.SourceLine())
			}
			if l > srcLen {
				srcLen = l
			}
			l = len(line.Func.PkgName())
			if l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

// functionColor returns the color to be used for the function name based on
// the type of package the function is in.
func (p *Palette) functionColor(line *stack.Call) string {
	if line.IsStdlib() {
		if line.Func.IsExported() {
			return p.FunctionStdLibExported
		}
		return p.FunctionStdLib
	} else if line.IsPkgMain() {
		return p.FunctionMain
	} else if line.Func.IsExported() {
		return p.FunctionOtherExported
	}
	return p.FunctionOther
}

// routineColor returns the color for the header of the goroutines bucket.
func (p *Palette) routineColor(bucket *stack.Bucket, multipleBuckets bool) string {
	if bucket.First() && multipleBuckets {
		return p.RoutineFirst
	}
	return p.Routine
}

// bucketHeader prints the header of a goroutine signature.
func (p *Palette) bucketHeader(bucket *stack.Bucket, fullPath, multipleBuckets bool) string {
	extra := ""
	if bucket.SleepMax != 0 {
		if bucket.SleepMin != bucket.SleepMax {
			extra += fmt.Sprintf(" [%d~%d minutes]", bucket.SleepMin, bucket.SleepMax)
		} else {
			extra += fmt.Sprintf(" [%d minutes]", bucket.SleepMax)
		}
	}
	if bucket.Locked {
		extra += " [locked]"
	}
	created := bucket.CreatedBy.Func.PkgDotName()
	if created != "" {
		created += " @ "
		if fullPath {
			created += bucket.CreatedBy.FullSourceLine()
		} else {
			created += bucket.CreatedBy.SourceLine()
		}
		extra += p.CreatedBy + " [Created by " + created + "]"
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		p.routineColor(bucket, multipleBuckets), len(bucket.Routines),
		bucket.State, extra,
		p.EOLReset)
}

// callLine prints one stack line.
func (p *Palette) callLine(line *stack.Call, srcLen, pkgLen int, fullPath bool) string {
	src := ""
	if fullPath {
		src = line.FullSourceLine()
	} else {
		src = line.SourceLine()
	}
	return fmt.Sprintf(
		"    %s%-*s %s%-*s %s%s%s(%s)%s",
		p.Package, pkgLen, line.Func.PkgName(),
		p.SourceFile, srcLen, src,
		p.functionColor(line), line.Func.Name(),
		p.Arguments, line.Args,
		p.EOLReset)
}

// stackLines prints one complete stack trace, without the header.
func (p *Palette) stackLines(signature *stack.Signature, srcLen, pkgLen int, fullPath bool) string {
	out := make([]string, len(signature.Stack.Calls))
	for i := range signature.Stack.Calls {
		out[i] = p.callLine(&signature.Stack.Calls[i], srcLen, pkgLen, fullPath)
	}
	if signature.Stack.Elided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}

// Process copies stdin to stdout and processes any "panic: " line found.
func Process(in io.Reader, out io.Writer, p *Palette, s stack.Similarity, fullPath bool) error {
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	if len(goroutines) == 1 && showBanner() {
		_, _ = io.WriteString(out, "\nTo see all goroutines, visit https://github.com/maruel/panicparse#GOTRACEBACK\n\n")
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, s))
	srcLen, pkgLen := calcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		_, _ = io.WriteString(out, p.bucketHeader(&bucket, fullPath, len(buckets) > 1))
		_, _ = io.WriteString(out, p.stackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
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
	p := &DefaultPalette
	if *noColor && !*forceColor {
		p = &Palette{}
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
	return Process(in, out, p, s, *fullPath)
}
