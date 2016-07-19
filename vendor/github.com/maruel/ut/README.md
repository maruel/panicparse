ut (utiltest)
=============

Collection of small functions to shorten Go test cases.

Requires Go 1.2 due to the use of `testing.TB`. If needed, replace with
`*testing.T` at the cost of not being usable in benchmarks.

[![GoDoc](https://godoc.org/github.com/maruel/ut?status.svg)](https://godoc.org/github.com/maruel/ut)
[![Build Status](https://travis-ci.org/maruel/ut.svg?branch=master)](https://travis-ci.org/maruel/ut)
[![Coverage Status](https://img.shields.io/coveralls/maruel/ut.svg)](https://coveralls.io/r/maruel/ut?branch=master)


Examples
--------

	package foo

	import (
		"github.com/maruel/ut"
		"log"
		"strconv"
		"testing"
	)

	func TestItoa(t *testing.T) {
		ut.AssertEqual(t, "42", strconv.Itoa(42))
	}

	func TestItoaDataListDriven(t *testing.T) {
		data := []struct {
			in       int
			expected string
		}{
			{9, "9"},
			{11, "11"},
		}
		for i, item := range data {
			ut.AssertEqualIndex(t, i, item.expected, strconv.Itoa(item.in))
		}
	}

	func TestWithLog(t *testing.T) {
		out := ut.NewWriter(t)
		defer out.Close()

		logger := log.New(out, "Foo:", 0)

		// These will be included in the test output only if the test case fails.
		logger.Printf("Q: What is the answer to life the universe and everything?")
		logger.Printf("A: %d", 42)
	}
