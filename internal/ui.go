// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"strings"

	"github.com/maruel/panicparse/v2/stack"
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
	Race         string

	// Call line.
	Package                     string
	SrcFile                     string
	FuncMain                    string
	FuncLocationUnknown         string
	FuncLocationUnknownExported string
	FuncGoMod                   string
	FuncGoModExported           string
	FuncGOPATH                  string
	FuncGOPATHExported          string
	FuncGoPkg                   string
	FuncGoPkgExported           string
	FuncStdLib                  string
	FuncStdLibExported          string
	Arguments                   string
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
		return fmt.Sprintf("%s:%d", c.RemoteSrcPath, c.Line)
	default:
		return fmt.Sprintf("%s:%d", c.SrcName, c.Line)
	}
}

func (pf pathFormat) createdByString(s *stack.Signature) string {
	if len(s.CreatedBy.Calls) == 0 {
		return ""
	}
	return s.CreatedBy.Calls[0].Func.DirName + "." + s.CreatedBy.Calls[0].Func.Name + " @ " + pf.formatCall(&s.CreatedBy.Calls[0])
}

// calcBucketsLengths returns the maximum length of the source lines and
// package names.
func calcBucketsLengths(a *stack.Aggregated, pf pathFormat) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, e := range a.Buckets {
		for _, line := range e.Signature.Stack.Calls {
			if l := len(pf.formatCall(&line)); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.DirName); l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

// calcGoroutinesLengths returns the maximum length of the source lines and
// package names.
func calcGoroutinesLengths(s *stack.Snapshot, pf pathFormat) (int, int) {
	srcLen := 0
	pkgLen := 0
	for _, e := range s.Goroutines {
		for _, line := range e.Signature.Stack.Calls {
			if l := len(pf.formatCall(&line)); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.DirName); l > pkgLen {
				pkgLen = l
			}
		}
	}
	return srcLen, pkgLen
}

// functionColor returns the color to be used for the function name based on
// the type of package the function is in.
func (p *Palette) functionColor(c *stack.Call) string {
	return p.funcColor(c.Location, c.Func.IsPkgMain, c.Func.IsExported)
}

func (p *Palette) funcColor(l stack.Location, main, exported bool) string {
	if main {
		return p.FuncMain
	}
	switch l {
	default:
		fallthrough
	case stack.LocationUnknown:
		if exported {
			return p.FuncLocationUnknownExported
		}
		return p.FuncLocationUnknown
	case stack.GoMod:
		if exported {
			return p.FuncGoModExported
		}
		return p.FuncGoMod
	case stack.GOPATH:
		if exported {
			return p.FuncGOPATHExported
		}
		return p.FuncGOPATH
	case stack.GoPkg:
		if exported {
			return p.FuncGoPkgExported
		}
		return p.FuncGoPkg
	case stack.Stdlib:
		if exported {
			return p.FuncStdLibExported
		}
		return p.FuncStdLib
	}
}

// routineColor returns the color for the header of the goroutines bucket.
func (p *Palette) routineColor(first, multipleBuckets bool) string {
	if first && multipleBuckets {
		return p.RoutineFirst
	}
	return p.Routine
}

// BucketHeader prints the header of a goroutine signature.
func (p *Palette) BucketHeader(b *stack.Bucket, pf pathFormat, multipleBuckets bool) string {
	extra := ""
	if s := b.SleepString(); s != "" {
		extra += " [" + s + "]"
	}
	if b.Locked {
		extra += " [locked]"
	}
	if c := pf.createdByString(&b.Signature); c != "" {
		extra += p.CreatedBy + " [Created by " + c + "]"
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		p.routineColor(b.First, multipleBuckets), len(b.IDs),
		b.State, extra,
		p.EOLReset)
}

// GoroutineHeader prints the header of a goroutine.
func (p *Palette) GoroutineHeader(g *stack.Goroutine, pf pathFormat, multipleGoroutines bool) string {
	extra := ""
	if s := g.SleepString(); s != "" {
		extra += " [" + s + "]"
	}
	if g.Locked {
		extra += " [locked]"
	}
	if c := pf.createdByString(&g.Signature); c != "" {
		extra += p.CreatedBy + " [Created by " + c + "]"
	}
	if g.RaceAddr != 0 {
		r := "read"
		if g.RaceWrite {
			r = "write"
		}
		extra += fmt.Sprintf("%s%s Race %s @ 0x%08x", p.EOLReset, p.Race, r, g.RaceAddr)
	}
	return fmt.Sprintf(
		"%s%d: %s%s%s\n",
		p.routineColor(g.First, multipleGoroutines), g.ID,
		g.State, extra,
		p.EOLReset)
}

// callLine prints one stack line.
func (p *Palette) callLine(line *stack.Call, srcLen, pkgLen int, pf pathFormat) string {
	return fmt.Sprintf(
		"    %s%-*s %s%-*s %s%s%s(%s)%s",
		p.Package, pkgLen, line.Func.DirName,
		p.SrcFile, srcLen, pf.formatCall(line),
		p.functionColor(line), line.Func.Name,
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
