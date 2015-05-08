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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosiner/gohper/termcolor"
	"github.com/maruel/panicparse/stack"
)

func CalcLengths(buckets stack.Buckets) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack {
			l := len(line.SourceLine())
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

func PrettyStack(r *stack.Signature, srcLen, pkgLen int) string {
	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	for _, line := range r.Stack {
		c := termcolor.Red
		if line.IsStdlib() {
			if line.Func.IsExported() {
				c = termcolor.LightGreen
			} else {
				c = termcolor.Green
			}
		} else if line.IsPkgMain() {
			c = termcolor.LightYellow
		} else if line.Func.IsExported() {
			c = termcolor.LightRed
		}

		termcolor.LightWhite.Fprintf(buf, "    %-*s", pkgLen, line.Func.PkgName())
		fmt.Fprintf(buf, " %-*s ", srcLen, line.SourceLine())
		c.RenderTo(buf, line.Func.Name())
		fmt.Fprintf(buf, "(%s)\n", line.Args)
	}
	return buf.String()
}

func Process(in io.Reader, out io.Writer) error {
	goroutines, err := stack.ParseDump(in, out)
	if err != nil {
		return err
	}
	buckets := stack.SortBuckets(stack.Bucketize(goroutines))
	srcLen, pkgLen := CalcLengths(buckets)
	for _, bucket := range buckets {
		extra := ""
		created := bucket.CreatedBy.Func.PkgDotName()
		if created != "" {
			if srcName := bucket.CreatedBy.SourceLine(); srcName != "" {
				created += " @ " + srcName
			}
			extra += " [Created by " + created + "]"
		}
		c := termcolor.White
		if bucket.First() && len(buckets) > 1 {
			c = termcolor.LightMagenta
		}
		c.Fprintf(out, "%d: %s", len(bucket.Routines), bucket.State)
		termcolor.LightBlack.Fprintln(out, extra)
		fmt.Fprintf(out, "%s\n", PrettyStack(&bucket.Signature, srcLen, pkgLen))
	}
	return err
}

func mainImpl() error {
	signals := make(chan os.Signal)
	go func() {
		for {
			<-signals
		}
	}()
	signal.Notify(signals, os.Interrupt, syscall.SIGQUIT)

	var in *os.File
	if len(os.Args) == 1 {
		in = os.Stdin
	} else if len(os.Args) == 2 {
		var err error
		name := os.Args[1]
		if in, err = os.Open(name); err != nil {
			return fmt.Errorf("did you mean to specify a valid stack dump file name? %s", err)
		}
		defer in.Close()
	} else {
		return errors.New("pipe from stdin or specify a single file")
	}
	return Process(in, termcolor.Stdout)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}
