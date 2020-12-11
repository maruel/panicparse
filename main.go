// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panicparse: analyzes stack dump of Go processes and simplifies it.
//
// It is mostly useful on servers will large number of identical goroutines,
// making the crash dump harder to read than strictly necessary.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/maruel/panicparse/v2/internal"
)

func main() {
	SetupCloseHandler()
	if err := internal.Main(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}
