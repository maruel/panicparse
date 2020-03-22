// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build windows

package main

import "golang.org/x/sys/windows"

func sysHang() {
	// 49.7 days is enough for everyone.
	windows.SleepEx(0xFFFFFFFF, false)
}
