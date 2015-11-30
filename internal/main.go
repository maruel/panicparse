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

var (
	boldDefault string
	boldBlack   string
	boldRed     string
	boldGreen   string
	boldYellow  string
	boldBlue    string
	boldMagenta string
	boldCyan    string
	boldWhite   string
)

func init() {
	boldDefault = ansi.ColorCode("default+b")
	boldBlack = ansi.ColorCode("black+b")
	boldRed = ansi.ColorCode("red+b")
	boldGreen = ansi.ColorCode("green+b")
	boldYellow = ansi.ColorCode("yellow+b")
	boldBlue = ansi.ColorCode("blue+b")
	boldMagenta = ansi.ColorCode("magenta+b")
	boldCyan = ansi.ColorCode("cyan+b")
	boldWhite = ansi.ColorCode("white+b")
}

// CalcLengths returns the maximum length of the source lines and package names.
func CalcLengths(buckets stack.Buckets, fullPath bool) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack {
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

// PkgColor returns the color to be used for the package name.
func PkgColor(line *stack.Call) string {
	if line.IsStdlib() {
		if line.Func.IsExported() {
			return boldGreen
		}
		return ansi.Green
	} else if line.IsPkgMain() {
		return boldYellow
	} else if line.Func.IsExported() {
		return boldRed
	}
	return ansi.Red
}

// BucketColor returns the color for the header of the goroutines bucket.
func BucketColor(bucket *stack.Bucket, multipleBuckets bool) string {
	if bucket.First() && multipleBuckets {
		return boldMagenta
	}
	return ""
}

// BucketHeader prints the header of a goroutine signature.
func BucketHeader(bucket *stack.Bucket, fullPath, multipleBuckets bool) string {
	extra := ""
	if bucket.Sleep != 0 {
		extra += fmt.Sprintf(" [%d minutes]", bucket.Sleep)
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
		extra += ansi.LightBlack + " [Created by " + created + "]"
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		BucketColor(bucket, multipleBuckets),
		len(bucket.Routines), bucket.State, extra, ansi.Reset)
}

// StackLine prints one complete stack trace, without the header.
func StackLine(line *stack.Call, srcLen, pkgLen int, fullPath bool) string {
	src := ""
	if fullPath {
		src = line.FullSourceLine()
	} else {
		src = line.SourceLine()
	}
	return fmt.Sprintf(
		"    %s%-*s%s %-*s %s%s%s(%s)",
		boldDefault, pkgLen, line.Func.PkgName(), ansi.Reset,
		srcLen, src,
		PkgColor(line), line.Func.Name(), ansi.Reset, line.Args)
}

// StackLines prints one complete stack trace, without the header.
func StackLines(signature *stack.Signature, srcLen, pkgLen int, fullPath bool) string {
	out := make([]string, len(signature.Stack))
	for i := range signature.Stack {
		out[i] = StackLine(&signature.Stack[i], srcLen, pkgLen, fullPath)
	}
	if signature.StackElided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}

// Process copies stdin to stdout and processes any "panic: " line found.
func Process(in io.Reader, out io.Writer, aggressive, fullPath bool) error {
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	s := stack.AnyPointer
	if aggressive {
		s = stack.AnyValue
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, s))
	srcLen, pkgLen := CalcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		_, _ = io.WriteString(out, BucketHeader(&bucket, fullPath, len(buckets) > 1))
		_, _ = io.WriteString(out, StackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	return err
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

	var out io.Writer
	if *noColor && !*forceColor {
		out = NewANSIStripper(os.Stdout)
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
	return Process(in, out, *aggressive, *fullPath)
}
