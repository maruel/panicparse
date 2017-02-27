// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ut

import (
	"fmt"
	"testing"
)

// WARNING: Any code change to this file will trigger a test failure of
// TestDecorateMax or TestDecorate. Make sure to update the expectation
// accordingly. Sorry for the inconvenience.

// TODO(maruel): It's wrong to hard code the containing path name.
const file = "ut/decorate_test.go"

func a() string {
	return b()
}

func b() string {
	return c()
}

func c() string {
	return d()
}

func d() string {
	return Decorate("Foo")
}

func TestDecorateMax(t *testing.T) {
	t.Parallel()
	// This test is line number dependent. a() is not listed, only b(), c() and
	// d().
	base := 24
	expected := fmt.Sprintf("%s:%d: %s:%d: %s:%d: Foo", file, base, file, base+4, file, base+8)
	AssertEqual(t, expected, a())
}

func TestDecorate(t *testing.T) {
	t.Parallel()
	// This test is line number dependent.
	a := Decorate("Foo")
	expected := fmt.Sprintf("%s:47: Foo", file)
	AssertEqual(t, expected, a)
}
