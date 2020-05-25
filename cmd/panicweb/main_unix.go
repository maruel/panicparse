// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build aix dragonfly freebsd linux netbsd openbsd solaris

package main

import "golang.org/x/sys/unix"

func sysHang() {
	unix.Nanosleep(&unix.Timespec{Sec: 366 * 24 * 60 * 60}, &unix.Timespec{})
}
