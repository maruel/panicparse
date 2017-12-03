// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build race

package main

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func rerunWithFastCrash() {
	if os.Getenv("GORACE") != "log_path=stderr halt_on_error=1" {
		os.Setenv("GORACE", "log_path=stderr halt_on_error=1")
		c := exec.Command(os.Args[0], os.Args[1:]...)
		c.Stderr = os.Stderr
		if err, ok := c.Run().(*exec.ExitError); ok {
			if status, ok := err.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func panicRace() {
	rerunWithFastCrash()
	i := 0
	for j := 0; j < 2; j++ {
		go func() {
			for {
				i++
			}
		}()
	}
	time.Sleep(time.Minute)
}
