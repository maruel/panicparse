// Package correct is in directory incorrect. If the call stack is
// incorrect.Panic(), you know the parsing failed.
package correct

// Panic panics.
func Panic() {
	panic(42)
}
