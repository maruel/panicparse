// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !race

package main

import (
	"io"
	"os"
)

func panicRace() {
	help := "'panic race' can only be used when built with the race detector.\n" +
		"To build, use:\n" +
		"  go install -race github.com/maruel/panicparse/cmd/panic\n"
	io.WriteString(os.Stderr, help)
	os.Exit(1)
}
