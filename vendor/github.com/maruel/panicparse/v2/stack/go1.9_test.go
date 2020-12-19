// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build go1.9

package stack

import "testing"

func helper(t *testing.T) func() {
	return t.Helper
}
