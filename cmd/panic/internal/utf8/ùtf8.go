// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ùtf8 tests a package and function with non-ASCII names.
package ùtf8

// Strùct is a totally normal structure with a totally normal name.
type Strùct struct {
}

// Pànic panics.
func (s *Strùct) Pànic() {
	panic(42)
}
