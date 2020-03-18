// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !go1.9

package stack

import "testing"

func helper(t *testing.T) func() {
	// testing.T.Helper() was added in Go 1.9.
	return func() {}
}
