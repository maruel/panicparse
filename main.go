// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panicparse: analyzes stack dump of Go processes and simplifies it.
//
// It is mostly useful on servers will large number of identical goroutines,
// making the crash dump harder to read than strictly necesary.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	reRoutineHeader = regexp.MustCompile("^goroutine (\\d+) \\[([^\\]]+)\\]\\:$")
	reFile          = regexp.MustCompile("^\t(.+\\.go)\\:(\\d+) \\+0x[0-9a-f]+$")
	reCreated       = regexp.MustCompile("^created by (.+)$")
	reFunc          = regexp.MustCompile("^(.+)\\((.*)\\)$")

	all = flag.Bool("all", false, "print all output before the stack dump")
)

// Call is an item in the stack trace.
type Call struct {
	Base     string
	Path     string
	Line     int
	FuncName string
}

// Goroutine represents the state of one goroutine.
type Goroutine struct {
	ID    int
	State string
	Stack []Call
}

// Eq ignores the ID.
func (r *Goroutine) Eq(l *Goroutine) bool {
	if r.State != l.State || len(r.Stack) != len(l.Stack) {
		return false
	}
	for i := range r.Stack {
		if r.Stack[i] != l.Stack[i] {
			return false
		}
	}
	return true
}

func (r *Goroutine) PrettyStack() string {
	out := []string{}
	for _, line := range r.Stack {
		out = append(out, fmt.Sprintf("  %s:%d: %s", line.Base, line.Line, line.FuncName))
	}
	return strings.Join(out, "\n")
}

// Bucketize returns the number of similar goroutines.
func Bucketize(goroutines []Goroutine) map[*Goroutine]int {
	out := map[*Goroutine]int{}
	// O(n²). Fix eventually.
	for _, r := range goroutines {
		found := false
		for k := range out {
			if r.Eq(k) {
				out[k] += 1
				found = true
				break
			}
		}
		if !found {
			k := &Goroutine{
				ID:    r.ID,
				State: r.State,
				Stack: r.Stack,
			}
			out[k] = 1
		}
	}
	return out
}

type Bucket struct {
	Goroutine
	Count int
}

type Buckets []Bucket

func (b Buckets) Len() int {
	return len(b)
}

func (b Buckets) Less(i, j int) bool {
	return b[i].Count < b[j].Count
}

func (b Buckets) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

func SortBuckets(buckets map[*Goroutine]int) Buckets {
	out := make(Buckets, 0, len(buckets))
	for r, count := range buckets {
		out = append(out, Bucket{*r, count})
	}
	sort.Sort(out)
	return out
}

// ParseDump processes the output from runtime.Stack().
//
// It supports piping from another command and assumes there is junk before the
// actual stack trace.
func ParseDump(r io.Reader) (string, []Goroutine, error) {
	goroutines := make([]Goroutine, 0, 16)
	var goroutine *Goroutine
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	header := ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			if goroutine == nil {
				header += line + "\n"
			}
			goroutine = nil
			continue
		}

		if goroutine == nil {
			if match := reRoutineHeader.FindStringSubmatch(line); match != nil {
				if id, err := strconv.Atoi(match[1]); err == nil {
					goroutines = append(goroutines, Goroutine{ID: id, State: match[2], Stack: []Call{}})
					goroutine = &goroutines[len(goroutines)-1]
					continue
				}
			}
			header += line + "\n"
			continue
		}

		if match := reFile.FindStringSubmatch(line); match != nil {
			num, err := strconv.Atoi(match[2])
			if err != nil {
				return header, goroutines, fmt.Errorf("failed to parse int on line: \"%s\"", line)
			}
			goroutine.Stack[len(goroutine.Stack)-1].Base = filepath.Base(match[1])
			goroutine.Stack[len(goroutine.Stack)-1].Path = match[1]
			goroutine.Stack[len(goroutine.Stack)-1].Line = num
		} else if match := reCreated.FindStringSubmatch(line); match != nil {
			goroutine.Stack = append(goroutine.Stack, Call{FuncName: filepath.Base(match[1])})
		} else if match := reFunc.FindStringSubmatch(line); match != nil {
			goroutine.Stack = append(goroutine.Stack, Call{FuncName: filepath.Base(match[1])})
		} else {
			header += line + "\n"
			goroutine = nil
		}
	}
	return header, goroutines, scanner.Err()
}

func mainImpl() error {
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

	header, goroutines, err := ParseDump(in)
	if err != nil {
		return err
	}
	if *all {
		fmt.Printf("%s\n", header)
	}
	for _, r := range SortBuckets(Bucketize(goroutines)) {
		fmt.Printf("%d: %s\n%s\n", r.Count, r.State, r.PrettyStack())
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}
