// Copyright 2018 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// TODO(maruel): This is a global state, affected by ParseDump(). This will
// be refactored in v2.
var (
	// goroot is the GOROOT as detected in the traceback, not the on the host.
	//
	// It can be empty if no root was determined, for example the traceback
	// contains only non-stdlib source references.
	goroot string
	// gopaths is the GOPATH as detected in the traceback, with the value being
	// the corresponding path mapped to the host.
	//
	// It can be empty if only stdlib code is in the traceback or if no local
	// sources were matched up. In the general case there is only one.
	gopaths map[string]string
	// Corresponding local values on the host.
	localgoroot  = runtime.GOROOT()
	localgopaths = getGOPATHs()
)

// ParseDump processes the output from runtime.Stack().
//
// It supports piping from another command and assumes there is junk before the
// actual stack trace. The junk is streamed to out.
func ParseDump(r io.Reader, out io.Writer) ([]Goroutine, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)
	s := scanningState{}
	for scanner.Scan() {
		line, err := s.scan(scanner.Text())
		if line != "" {
			_, _ = io.WriteString(out, line)
		}
		if err != nil {
			return s.goroutines, err
		}
	}
	nameArguments(s.goroutines)
	// Mutate global state.
	// TODO(maruel): Make this part of the context instead of a global.
	if goroot == "" {
		findRoots(s.goroutines)
	}
	return s.goroutines, scanner.Err()
}

// NoRebase disables GOROOT and GOPATH guessing in ParseDump().
//
// BUG: This function will be removed in v2, as ParseDump() will accept a flag
// explicitly.
func NoRebase() {
	goroot = runtime.GOROOT()
	gopaths = map[string]string{}
	for _, p := range getGOPATHs() {
		gopaths[p] = p
	}
}

// Private stuff.

// scanLines is similar to bufio.ScanLines except that it:
//     - doesn't drop '\n'
//     - doesn't strip '\r'
//     - returns when the data is bufio.MaxScanTokenSize bytes
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	if len(data) >= bufio.MaxScanTokenSize {
		// Returns the line even if it is not at EOF nor has a '\n', otherwise the
		// scanner will return bufio.ErrTooLong which is definitely not what we
		// want.
		return len(data), data, nil
	}
	return 0, nil, nil
}

// scanningState is the state of the scan to detect and process a stack trace.
//
// TODO(maruel): Use a formal state machine. Patterns follows:
// - reRoutineHeader
//   Either:
//     - reUnavail
//     - reFunc + reFile in a loop
//     - reElided
//   Optionally ends with:
//     - reCreated + reFile
type scanningState struct {
	goroutines []Goroutine
	goroutine  *Goroutine

	created   bool
	firstLine bool // firstLine is the first line after the reRoutineHeader header line.
}

func (s *scanningState) scan(line string) (string, error) {
	if line == "\n" || line == "\r\n" {
		if s.goroutine != nil {
			// goroutines are separated by an empty line.
			s.goroutine = nil
			return "", nil
		}
	} else if line[len(line)-1] == '\n' {
		if s.goroutine == nil {
			if match := reRoutineHeader.FindStringSubmatch(line); match != nil {
				if id, err := strconv.Atoi(match[1]); err == nil {
					// See runtime/traceback.go.
					// "<state>, \d+ minutes, locked to thread"
					items := strings.Split(match[2], ", ")
					sleep := 0
					locked := false
					for i := 1; i < len(items); i++ {
						if items[i] == lockedToThread {
							locked = true
							continue
						}
						// Look for duration, if any.
						if match2 := reMinutes.FindStringSubmatch(items[i]); match2 != nil {
							sleep, _ = strconv.Atoi(match2[1])
						}
					}
					s.goroutines = append(s.goroutines, Goroutine{
						Signature: Signature{
							State:    items[0],
							SleepMin: sleep,
							SleepMax: sleep,
							Locked:   locked,
						},
						ID:    id,
						First: len(s.goroutines) == 0,
					})
					s.goroutine = &s.goroutines[len(s.goroutines)-1]
					s.firstLine = true
					return "", nil
				}
			}
		} else {
			if s.firstLine {
				s.firstLine = false
				if match := reUnavail.FindStringSubmatch(line); match != nil {
					// Generate a fake stack entry.
					s.goroutine.Stack.Calls = []Call{{SourcePath: "<unavailable>"}}
					return "", nil
				}
			}

			if match := reFile.FindStringSubmatch(line); match != nil {
				// Triggers after a reFunc or a reCreated.
				num, err := strconv.Atoi(match[2])
				if err != nil {
					return "", fmt.Errorf("failed to parse int on line: %q", strings.TrimSpace(line))
				}
				if s.created {
					s.created = false
					s.goroutine.CreatedBy.SourcePath = match[1]
					s.goroutine.CreatedBy.Line = num
				} else {
					i := len(s.goroutine.Stack.Calls) - 1
					if i < 0 {
						return "", fmt.Errorf("unexpected order on line: %q", strings.TrimSpace(line))
					}
					s.goroutine.Stack.Calls[i].SourcePath = match[1]
					s.goroutine.Stack.Calls[i].Line = num
				}
				return "", nil
			}

			if match := reCreated.FindStringSubmatch(line); match != nil {
				s.created = true
				s.goroutine.CreatedBy.Func.Raw = match[1]
				return "", nil
			}

			if match := reFunc.FindStringSubmatch(line); match != nil {
				args := Args{}
				for _, a := range strings.Split(match[2], ", ") {
					if a == "..." {
						args.Elided = true
						continue
					}
					if a == "" {
						// Remaining values were dropped.
						break
					}
					v, err := strconv.ParseUint(a, 0, 64)
					if err != nil {
						return "", fmt.Errorf("failed to parse int on line: %q", strings.TrimSpace(line))
					}
					args.Values = append(args.Values, Arg{Value: v})
				}
				s.goroutine.Stack.Calls = append(s.goroutine.Stack.Calls, Call{Func: Function{match[1]}, Args: args})
				return "", nil
			}

			if match := reElided.FindStringSubmatch(line); match != nil {
				s.goroutine.Stack.Elided = true
				return "", nil
			}
		}
	}
	s.goroutine = nil
	return line, nil
}

// hasPathPrefix returns true if any of s is the prefix of p.
func hasPathPrefix(p string, s map[string]string) bool {
	for prefix := range s {
		if strings.HasPrefix(p, prefix+"/") {
			return true
		}
	}
	return false
}

// getFiles returns all the source files deduped and ordered.
func getFiles(goroutines []Goroutine) []string {
	files := map[string]struct{}{}
	for _, g := range goroutines {
		for _, c := range g.Stack.Calls {
			files[c.SourcePath] = struct{}{}
		}
	}
	out := make([]string, 0, len(files))
	for f := range files {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// splitPath splits a path into its components.
//
// The first item has its initial path separator kept.
func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	var out []string
	s := ""
	for _, c := range p {
		if c != '/' || (len(out) == 0 && strings.Count(s, "/") == len(s)) {
			s += string(c)
		} else if s != "" {
			out = append(out, s)
			s = ""
		}
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

// isFile returns true if the path is a valid file.
func isFile(p string) bool {
	// TODO(maruel): Is it faster to open the file or to stat it? Worth a perf
	// test on Windows.
	i, err := os.Stat(p)
	return err == nil && !i.IsDir()
}

// rootedIn returns a root if the file split in parts is rooted in root.
func rootedIn(root string, parts []string) string {
	//log.Printf("rootIn(%s, %v)", root, parts)
	for i := 1; i < len(parts); i++ {
		suffix := filepath.Join(parts[i:]...)
		if isFile(filepath.Join(root, suffix)) {
			return filepath.Join(parts[:i]...)
		}
	}
	return ""
}

// findRoots sets global variables goroot and gopath.
//
// TODO(maruel): In v2, it will be a property of the new struct that will
// contain the goroutines.
func findRoots(goroutines []Goroutine) {
	gopaths = map[string]string{}
	for _, f := range getFiles(goroutines) {
		// TODO(maruel): Could a stack dump have mixed cases? I think it's
		// possible, need to confirm and handle.
		//log.Printf("  Analyzing %s", f)
		if goroot != "" && strings.HasPrefix(f, goroot+"/") {
			continue
		}
		if gopaths != nil && hasPathPrefix(f, gopaths) {
			continue
		}
		parts := splitPath(f)
		if goroot == "" {
			if r := rootedIn(localgoroot, parts); r != "" {
				goroot = r
				log.Printf("Found GOROOT=%s", goroot)
				continue
			}
		}
		found := false
		for _, l := range localgopaths {
			if r := rootedIn(l, parts); r != "" {
				log.Printf("Found GOPATH=%s", r)
				gopaths[r] = l
				found = true
				break
			}
		}
		if !found {
			// If the source is not found, just too bad.
			//log.Printf("Failed to find locally: %s / %s", f, goroot)
		}
	}
}

func getGOPATHs() []string {
	var out []string
	for _, v := range filepath.SplitList(os.Getenv("GOPATH")) {
		// Disallow non-absolute paths?
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		homeDir := ""
		u, err := user.Current()
		if err != nil {
			homeDir = os.Getenv("HOME")
			if homeDir == "" {
				panic(fmt.Sprintf("Could not get current user or $HOME: %s\n", err.Error()))
			}
		} else {
			homeDir = u.HomeDir
		}
		out = []string{homeDir + "go"}
	}
	return out
}
