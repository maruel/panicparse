// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestMain(t *testing.T) {
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
			// It's not safe, it'll crash the program in a way that cannot be trapped.
			// TODO(maruel): Run it conditionally when not under race detector.
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
