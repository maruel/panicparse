// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"strings"

	"github.com/maruel/panicparse/stack"
)

// Palette defines the color used.
//
// An empty object Palette{} can be used to disable coloring.
type Palette struct {
	EOLReset string

	// Routine header.
	RoutineFirst string // The first routine printed.
	Routine      string // Following routines.
	CreatedBy    string

	// Call line.
	Package            string
	SrcFile            string
	FuncStdLib         string
	FuncStdLibExported string
	FuncMain           string
	FuncOther          string
	FuncOtherExported  string
	Arguments          string
}

// pathFormat determines how much to show.
type pathFormat int

const (
	fullPath pathFormat = iota
	relPath
	basePath
)

func (pf pathFormat) formatCall(c *stack.Call) string {
	switch pf {
	case relPath:
		if c.RelSrcPath != "" {
			return fmt.Sprintf("%s:%d", c.RelSrcPath, c.Line)
		}
		fallthrough
	case fullPath:
		if c.LocalSrcPath != "" {
			return fmt.Sprintf("%s:%d", c.LocalSrcPath, c.Line)
		}
		return fmt.Sprintf("%s:%d", c.SrcPath, c.Line)
	default:
		return fmt.Sprintf("%s:%d", c.SrcName(), c.Line)
	}
}

func (pf pathFormat) createdByString(s *stack.Signature) string {
	created := s.CreatedBy.Func.PkgDotName()
	if created == "" {
		return ""
	}
	return created + " @ " + pf.formatCall(&s.CreatedBy)
}

// calcLengths returns the maximum length of the source lines and package names.
func calcLengths(buckets []*stack.Bucket, pf pathFormat) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(pf.formatCall(&line)); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.PkgName()); l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

// functionColor returns the color to be used for the function name based on
// the type of package the function is in.
func (p *Palette) functionColor(line *stack.Call) string {
	if line.IsStdlib {
		if line.Func.IsExported() {
			return p.FuncStdLibExported
		}
		return p.FuncStdLib
	} else if line.IsPkgMain() {
		return p.FuncMain
	} else if line.Func.IsExported() {
		return p.FuncOtherExported
	}
	return p.FuncOther
}

// routineColor returns the color for the header of the goroutines bucket.
func (p *Palette) routineColor(bucket *stack.Bucket, multipleBuckets bool) string {
	if bucket.First && multipleBuckets {
		return p.RoutineFirst
	}
	return p.Routine
}

// BucketHeader prints the header of a goroutine signature.
func (p *Palette) BucketHeader(bucket *stack.Bucket, pf pathFormat, multipleBuckets bool) string {
	extra := ""
	if s := bucket.SleepString(); s != "" {
		extra += " [" + s + "]"
	}
	if bucket.Locked {
		extra += " [locked]"
	}
	if c := pf.createdByString(&bucket.Signature); c != "" {
		extra += p.CreatedBy + " [Created by " + c + "]"
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		p.routineColor(bucket, multipleBuckets), len(bucket.IDs),
		bucket.State, extra,
		p.EOLReset)
}

// callLine prints one stack line.
func (p *Palette) callLine(line *stack.Call, srcLen, pkgLen int, pf pathFormat) string {
	return fmt.Sprintf(
		"    %s%-*s %s%-*s %s%s%s(%s)%s",
		p.Package, pkgLen, line.Func.PkgName(),
		p.SrcFile, srcLen, pf.formatCall(line),
		p.functionColor(line), line.Func.Name(),
		p.Arguments, &line.Args,
		p.EOLReset)
}

// StackLines prints one complete stack trace, without the header.
func (p *Palette) StackLines(signature *stack.Signature, srcLen, pkgLen int, pf pathFormat) string {
	out := make([]string, len(signature.Stack.Calls))
	for i := range signature.Stack.Calls {
		out[i] = p.callLine(&signature.Stack.Calls[i], srcLen, pkgLen, pf)
	}
	if signature.Stack.Elided {
		out = append(out, "    (...)")
	}
	return strings.Join(out, "\n") + "\n"
}
