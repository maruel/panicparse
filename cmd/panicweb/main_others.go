// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:build !aix && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !windows
// +build !aix,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package main

func sysHang() {
}
