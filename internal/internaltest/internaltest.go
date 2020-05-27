// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internaltest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// IsUsingModules is best guess to know if go module are enabled.
//
// Panics if an internal error occurs.
//
// It reads the current value of GO111MODULES.
func IsUsingModules() bool {
	// Calculate the default. We assume developer builds are recent (go1.14 and
	// later).
	ver := getGoMinorVersion()
	if ver > 0 && ver < 11 {
		// go1.9.7+ and go1.10.3+ were fixed to tolerate semantic versioning import
		// but they do not support the environment variable.
		return false
	}
	def := (ver == 0 || ver >= 14)
	s := os.Getenv("GO111MODULE")
	return (def && (s == "auto" || s == "")) || s == "on"
}

//

var (
	panicwebOnce     sync.Once
	panicwebOutput   []byte
	panicOutputsOnce sync.Once
	panicOutputs     map[string][]byte
)

// getGoMinorVersion returns the Go1 minor version.
//
// Returns 0 for a developper build, panics if can't parse the version.
//
// Ignores the revision (go1.<minor>.<revision>).
func getGoMinorVersion() int {
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
	}
	m, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %q: %v", ver, err))
	}
	return m
}

// build creates a temporary file and returns the path to it.
func build(tool string, race bool) string {
	p := filepath.Join(os.TempDir(), tool)
	if race {
		p += "_race"
	}
	// Starting with go1.11, ioutil.TempFile() supports specifying a suffix. This
	// is necessary to set the ".exe" suffix on Windows. Until we drop support
	// for go1.10 and earlier, do the equivalent ourselves in an lousy way.
	p += fmt.Sprintf("_%d", os.Getpid())
	if runtime.GOOS == "windows" {
		p += ".exe"
	}
	// Disable inlining otherwise the inlining varies between local execution and
	// remote execution. This can be observed as Elided being true without any
	// argument.
	args := []string{"build", "-gcflags", "-l", "-o", p}
	if race {
		args = append(args, "-race")
	}
	path := "github.com/maruel/panicparse/cmd/"
	c := exec.Command("go", append(args, path+tool)...)
	b := bytes.Buffer{}
	c.Stdout = os.Stdout
	if race {
		c.Stderr = &b
	} else {
		c.Stderr = os.Stderr
	}
	if err := c.Run(); err != nil {
		if race {
			if strings.HasPrefix(b.String(), "go test: -race is only supported on ") {
				// Race detector is not supported. Calling code with handle.
				return ""
			}
			os.Stderr.Write(b.Bytes())
			return ""
		}
		return ""
	}
	return p
}

// execRun runs a command and returns the combined output.
//
// It ignores the exit code, since it's meant to run panic, which crashes by
// design.
func execRun(cmd ...string) []byte {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Env = append(os.Environ(), "GOTRACEBACK=all")
	out, _ := c.CombinedOutput()
	return out
}
