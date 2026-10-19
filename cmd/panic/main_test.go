// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Tests panic scenarios and trusted executable restart behavior.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	t.Parallel()
	if !testing.Verbose() {
		stdErr = io.Discard
		defer func() {
			stdErr = os.Stderr
		}()
	}
	for name, l := range types {
		if name == "simple" {
			// It's safe.
			l.f()
			continue
		}
		if name == "goroutine_1" || name == "asleep" {
			// goroutine_1 panics in a separate goroutine, so it's tricky to catch.
			// asleep will just hang because there are other goroutine.
			continue
		}
		if name == "race" {
			if raceEnabled {
				// It's not safe, it'll crash the program in a way that cannot be trapped.
				continue
			}
			// It's safe.
			l.f()
			continue
		}

		l := l
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if err := recover(); err == nil {
					t.Fatal("expected error")
				}
			}()
			l.f()
		})
	}
}

// TestRerunWithFastCrash verifies restart uses the executable rather than argv[0].
func TestRerunWithFastCrash(t *testing.T) {
	if os.Getenv("PANICPARSE_REEXEC_TEST") == "1" {
		if os.Getenv("GORACE") == "log_path=stderr halt_on_error=1" {
			fmt.Fprintln(os.Stderr, "trusted restart")
			return
		}
		os.Args[0] = "missing-panicparse-executable"
		rerunWithFastCrash()
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestRerunWithFastCrash$")
	cmd.Env = append(os.Environ(), "PANICPARSE_REEXEC_TEST=1", "GORACE=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("restart failed: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("trusted restart")) {
		t.Fatalf("restarted process did not run: %s", out)
	}
}
