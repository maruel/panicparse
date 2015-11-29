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
			return ansi.LightGreen
		} else {
			return ansi.Green
		}
	} else if line.IsPkgMain() {
		return ansi.LightYellow
	} else if line.Func.IsExported() {
		return ansi.LightRed
	}
	return ansi.Red
}

// PrintStackHeader prints the header of a stack.
func PrintStackHeader(out io.Writer, bucket *stack.Bucket, fullPath, multipleBuckets bool) {
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
	c := ansi.White
	if bucket.First() && multipleBuckets {
		c = ansi.LightMagenta
	}
	fmt.Fprintf(out, "%s%d: %s%s%s\n", c, len(bucket.Routines), bucket.State, extra, ansi.Reset)
}

// PrettyStack prints one complete stack trace, without the header.
func PrettyStack(r *stack.Signature, srcLen, pkgLen int, fullPath bool) string {
	out := []string{}
	for _, line := range r.Stack {
		src := ""
		if fullPath {
			src = line.FullSourceLine()
		} else {
			src = line.SourceLine()
		}
		s := fmt.Sprintf(
			"    %s%-*s%s %-*s %s%s%s(%s)",
			ansi.LightWhite, pkgLen, line.Func.PkgName(), ansi.Reset,
			srcLen, src,
			PkgColor(&line), line.Func.Name(), ansi.Reset, line.Args)
		out = append(out, s)
	}
	if r.StackElided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n")
}

// Process copies stdin to stdout and processes any "panic: " line found.
func Process(in io.Reader, out io.Writer, fullPath bool) error {
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, true))
	srcLen, pkgLen := CalcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		PrintStackHeader(out, &bucket, fullPath, len(buckets) > 1)
		fmt.Fprintf(out, "%s\n", PrettyStack(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	return err
}

func Main() error {
	signals := make(chan os.Signal)
	go func() {
		for {
			<-signals
		}
	}()
	signal.Notify(signals, os.Interrupt, syscall.SIGQUIT)
	// TODO(maruel): Both github.com/mattn/go-colorable and
	// github.com/shiena/ansicolor failed at properly printing colors on my
	// Windows box. Figure this out eventually. In the meantime, default to no
	// color on Windows.
	noColor := flag.Bool("no-color", !isatty.IsTerminal(os.Stdout.Fd()) || os.Getenv("TERM") == "dumb", "Disable coloring")
	forceColor := flag.Bool("force-color", false, "Forcibly enable coloring when with stdout is redirected")
	fullPath := flag.Bool("full-path", false, "Print full sources path")
	verboseFlag := flag.Bool("v", false, "Enables verbose logging output")
	flag.Parse()

	log.SetFlags(log.Lmicroseconds)
	if !*verboseFlag {
		log.SetOutput(ioutil.Discard)
	}

	var out io.Writer
	if *noColor && !*forceColor {
		out = NewAnsiStripper(os.Stdout)
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
	return Process(in, out, *fullPath)
}
