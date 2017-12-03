// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal is for use for panic.
package internal

// Callback calls back a function through an external then internal function.
func Callback(f func()) {
	callback(f)
}

func callback(f func()) {
	f()
}
