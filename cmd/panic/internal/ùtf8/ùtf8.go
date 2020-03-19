// Package ùtf8 tests a package and function with non-ASCII names.
package ùtf8

// Strùct is a totally normal structure with a totally normal name.
type Strùct struct {
}

// Pànic panics.
func (s *Strùct) Pànic() {
	panic(42)
}
