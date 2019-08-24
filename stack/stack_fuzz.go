// +build gofuzz

package stack

import (
	"bytes"
	"io/ioutil"
)

// Fuzz tests parsing and aggregation
func Fuzz(fuzz []byte) int {
	ctx, err := ParseDump(bytes.NewReader(fuzz), ioutil.Discard, true)
	if err != nil {
		return 0
	}
	if ctx == nil {
		return 0
	}
	Aggregate(ctx.Goroutines, AnyValue)
	return 1
}
