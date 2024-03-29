// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internaltest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// PanicwebOutput returns the output of panicweb with inlining disabled.
//
// The function panics if any internal error occurs.
func PanicwebOutput() []byte {
	panicwebOnce.Do(func() {
		p := build("panicweb", false)
		if p == "" {
			panic("building panicweb failed")
		}
		defer func() {
			if err := os.Remove(p); err != nil {
				panic(err)
			}
		}()
		panicwebOutput = execRun(p)
	})
	out := make([]byte, len(panicwebOutput))
	copy(out, panicwebOutput)
	return out
}

// PanicOutputs returns a map of the output of every subcommands.
//
// panic is built with inlining disabled.
//
// The subcommand "race" is built with the race detector. Others are built
// without. In particular "asleep" doesn't work with the race detector.
//
// The function panics if any internal error occurs.
func PanicOutputs() map[string][]byte {
	panicOutputsOnce.Do(func() {
		// Extracts the subcommands, then run each of them individually.
		pplain := build("panic", false)
		if pplain == "" {
			// The odd of this failing is close to nil.
			panic("building panic failed")
		}
		defer func() {
			if err := os.Remove(pplain); err != nil {
				panic(err)
			}
		}()

		prace := build("panic", true)
		if prace == "" {
			// Race detector is not supported on this platform.
		} else {
			defer func() {
				if err := os.Remove(prace); err != nil {
					panic(err)
				}
			}()
		}

		// Collect the subcommands.
		cmds := strings.Split(strings.TrimSpace(string(execRun(pplain, "dump_commands"))), "\n")
		if len(cmds) == 0 {
			panic("no command retrieved")
		}

		// Collect the output of each subcommand.
		panicOutputs = map[string][]byte{}
		for _, cmd := range cmds {
			cmd = strings.TrimSpace(cmd)
			p := pplain
			if cmd == "race" {
				if prace == "" {
					// Race detector is not supported.
					continue
				}
				p = prace
			}
			if panicOutputs[cmd] = execRun(p, cmd); len(panicOutputs[cmd]) == 0 {
				panic(fmt.Sprintf("no output for %s", cmd))
			}
		}
	})
	out := make(map[string][]byte, len(panicOutputs))
	for k, v := range panicOutputs {
		w := make([]byte, len(v))
		copy(w, v)
		out[k] = w
	}
	return out
}

// StaticPanicwebOutput returns a constant version of panicweb output for use
// in benchmarks.
func StaticPanicwebOutput() []byte {
	return []byte(staticPanicweb)
}

// StaticPanicRaceOutput returns a constant version of 'panic race' output.
func StaticPanicRaceOutput() []byte {
	return []byte(staticPanicRace)
}

// IsUsingModules returns if go modules are enabled.
//
// It reads the current value of GO111MODULES.
func IsUsingModules() bool {
	s := os.Getenv("GO111MODULE")
	return s != "off"
}

//

var (
	panicwebOnce     sync.Once
	panicwebOutput   []byte
	panicOutputsOnce sync.Once
	panicOutputs     map[string][]byte
)

// GetGoMinorVersion returns the Go1 minor version.
//
// Returns 0 for a developer build, panics if can't parse the version.
//
// Ignores the revision (go1.<minor>.<revision>).
func GetGoMinorVersion() int {
	ver := runtime.Version()
	if strings.HasPrefix(ver, "devel ") {
		return 0
	}
	if !strings.HasPrefix(ver, "go1.") {
		// This will break on go2. Please submit a PR to fix this once Go2 is
		// released.
		panic(fmt.Sprintf("unexpected go version %q", ver))
	}
	v := ver[4:]
	if i := strings.IndexByte(v, '.'); i != -1 {
		v = v[:i]
	} else if i := strings.Index(v, "beta"); i != -1 {
		v = v[:i]
	} else if i := strings.Index(v, "rc"); i != -1 {
		v = v[:i]
	}

	m, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %q: %v", ver, err))
	}
	return m
}

// build builds to a temporary file and returns the path to it.
func build(tool string, race bool) string {
	s := "panicparse*"
	if runtime.GOOS == "windows" {
		s = "panicparse*.exe"
	}
	f, err := os.CreateTemp("", s)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		return ""
	}
	p := f.Name()
	_ = f.Close()
	path := "github.com/maruel/panicparse/cmd/"
	if IsUsingModules() {
		path = "github.com/maruel/panicparse/v2/cmd/"
	}
	if err := Compile(path+tool, p, "", true, race); err != nil {
		_ = os.Remove(p)
		_, _ = os.Stderr.WriteString(err.Error())
		return ""
	}
	return p
}

var errNoRace = errors.New("platform does not support -race")

// Compile compiles sources into an executable.
func Compile(in, exe, cwd string, disableInlining, race bool) error {
	// Disable optimization (-N) and inlining (-l) otherwise the inlining varies
	// between local execution and remote execution. This can be observed as
	// Elided being true without any argument.
	args := []string{"build", "-o", exe}
	if disableInlining {
		args = append(args, "-gcflags", "-N -l")
	}
	if race {
		args = append(args, "-race")
	}
	/* #nosec G204 */
	c := exec.Command("go", append(args, in)...)
	c.Dir = cwd
	if out, err := c.CombinedOutput(); err != nil {
		if race {
			s := string(out)
			const e1 = "go test: -race is only supported on "
			const e2 = "go build: -race is only supported on "
			if strings.HasPrefix(s, e1) || strings.HasPrefix(s, e2) {
				return errNoRace
			}
		}
		return fmt.Errorf("compile failure: %w\n%s", err, out)
	}
	return nil
}

// execRun runs a command and returns the combined output.
//
// It ignores the exit code, since it's meant to run panic, which crashes by
// design.
func execRun(cmd ...string) []byte {
	/* #nosec G204 */
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Env = append(os.Environ(), "GOTRACEBACK=all")
	out, _ := c.CombinedOutput()
	return out
}
