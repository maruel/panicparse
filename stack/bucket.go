// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"sort"
)

// Similarity is the level at which two call lines arguments must match to be
// considered similar enough to coalesce them.
type Similarity int

const (
	// ExactFlags requires same bits (e.g. Locked).
	ExactFlags Similarity = iota
	// ExactLines requests the exact same arguments on the call line.
	ExactLines
	// AnyPointer considers different pointers a similar call line.
	AnyPointer
	// AnyValue accepts any value as similar call line.
	AnyValue
)

// Bucketize merges similar goroutines into buckets.
//
// The buckets are ordering in order of relevancy.
func Bucketize(goroutines []Goroutine, similar Similarity) []Bucket {
	b := map[*Signature][]Goroutine{}
	// O(n²). Fix eventually.
	for _, routine := range goroutines {
		found := false
		for key := range b {
			// When a match is found, this effectively drops the other goroutine ID.
			if key.similar(&routine.Signature, similar) {
				found = true
				if !key.equal(&routine.Signature) {
					// Almost but not quite equal. There's different pointers passed
					// around but the same values. Zap out the different values.
					newKey := key.merge(&routine.Signature)
					b[newKey] = append(b[key], routine)
					delete(b, key)
				} else {
					b[key] = append(b[key], routine)
				}
				break
			}
		}
		if !found {
			key := &Signature{}
			*key = routine.Signature
			b[key] = []Goroutine{routine}
		}
	}
	return sortBuckets(b)
}

// Bucket is a stack trace signature and the list of goroutines that fits this
// signature.
type Bucket struct {
	Signature
	Routines []Goroutine
}

// First returns true if it contains the first goroutine, e.g. the ones that
// likely generated the panic() call, if any.
func (b *Bucket) First() bool {
	for _, r := range b.Routines {
		if r.First {
			return true
		}
	}
	return false
}

// Less does reverse sort.
func (b *Bucket) Less(r *Bucket) bool {
	if b.First() {
		return true
	}
	if r.First() {
		return false
	}
	return b.Signature.Less(&r.Signature)
}

//

// buckets is a list of Bucket sorted by repeation count.
type buckets []Bucket

func (b buckets) Len() int {
	return len(b)
}

func (b buckets) Less(i, j int) bool {
	return b[i].Less(&b[j])
}

func (b buckets) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// sortBuckets creates a list of Bucket from each goroutine stack trace count.
func sortBuckets(b map[*Signature][]Goroutine) []Bucket {
	out := make(buckets, 0, len(b))
	for signature, count := range b {
		out = append(out, Bucket{*signature, count})
	}
	sort.Sort(out)
	return out
}
