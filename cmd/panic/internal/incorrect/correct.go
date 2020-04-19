// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package correct is in directory incorrect. If the call stack is
// incorrect.Panic(), you know the parsing failed.
package correct

// Panic panics.
func Panic() {
	panic(42)
}
