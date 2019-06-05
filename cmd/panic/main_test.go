// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	if !testing.Verbose() {
		stdErr = ioutil.Discard
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

		t.Run(name, func(t *testing.T) {
			defer func() {
				if err := recover(); err == nil {
					t.Fatal("expected error")
				}
			}()
			l.f()
		})
	}
}
