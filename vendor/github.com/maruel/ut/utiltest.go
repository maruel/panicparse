// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ut (for UtilTest) contains testing utilities to shorten unit tests.
package ut

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/pmezard/go-difflib/difflib"
)

var newLine = []byte{'\n'}

var blacklistedItems = map[string]bool{
	filepath.Join("runtime", "asm_386.s"):   true,
	filepath.Join("runtime", "asm_amd64.s"): true,
	filepath.Join("runtime", "asm_arm.s"):   true,
	filepath.Join("runtime", "proc.c"):      true,
	filepath.Join("testing", "testing.go"):  true,
	filepath.Join("ut", "utiltest.go"):      true,
	"utiltest.go":                           true,
}

// truncatePath only keep the base filename and its immediate containing
// directory.
func truncatePath(file string) string {
	return filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
}

func isBlacklisted(file string) bool {
	_, ok := blacklistedItems[file]
	return ok
}

// Decorate adds a prefix 'file:line: ' to a string, containing the 3 recent
// callers in the stack.
//
// It skips internal functions. It is mostly meant to be used internally.
//
// It is inspired by testing's decorate().
func Decorate(s string) string {
	type item struct {
		file string
		line int
	}
	items := []item{}
	for i := 1; i < 8 && len(items) < 3; i++ {
		_, file, line, ok := runtime.Caller(i) // decorate + log + public function.
		if ok {
			file = truncatePath(file)
			if !isBlacklisted(file) {
				items = append(items, item{file, line})
			}
		}
	}
	for _, i := range items {
		s = fmt.Sprintf("%s:%d: %s", strings.Replace(i.file, "%", "%%", -1), i.line, s)
	}
	return s
}

// AssertEqual verifies that two objects are equals and calls FailNow() to
// immediately cancel the test case.
//
// It must be called from the main goroutine. Other goroutines must call
// ExpectEqual* flavors.
//
// Equality is determined via reflect.DeepEqual().
func AssertEqual(t testing.TB, expected, actual interface{}) {
	AssertEqualf(t, expected, actual, "AssertEqual() failure.\n%# v", formatAsDiff(expected, actual))
}

// AssertEqualIndex verifies that two objects are equals and calls FailNow() to
// immediately cancel the test case.
//
// It must be called from the main goroutine. Other goroutines must call
// ExpectEqual* flavors.
//
// It is meant to be used in loops where a list of intrant->expected is
// processed so the assert failure message contains the index of the failing
// expectation.
//
// Equality is determined via reflect.DeepEqual().
func AssertEqualIndex(t testing.TB, index int, expected, actual interface{}) {
	AssertEqualf(t, expected, actual, "AssertEqualIndex() failure.\nIndex: %d\n%# v", index, formatAsDiff(expected, actual))
}

// AssertEqualf verifies that two objects are equals and calls FailNow() to
// immediately cancel the test case.
//
// It must be called from the main goroutine. Other goroutines must call
// ExpectEqual* flavors.
//
// This functions enables specifying an arbitrary string on failure.
//
// Equality is determined via reflect.DeepEqual().
func AssertEqualf(t testing.TB, expected, actual interface{}, format string, items ...interface{}) {
	// This is cheezy, as there's no way to figure out if the test was properly
	// started by the test framework.
	found := false
	root := ""
	for i := 1; ; i++ {
		if _, file, _, ok := runtime.Caller(i); ok {
			if filepath.Base(file) == "testing.go" {
				found = true
				break
			}
			root = file
		} else {
			break
		}
	}
	if !found {
		t.Logf(Decorate("ut.AssertEqual*() function MUST be called from within main test goroutine, use ut.ExpectEqual*() instead; found %s."), root)
		// TODO(maruel): Warning: this will be enforced soon.
		//t.Fail()
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(Decorate(format), items...)
	}
}

// ExpectEqual verifies that two objects are equals and calls Fail() to mark
// the test case as failed but let it continue.
//
// It is fine to call this function from another goroutine than the main test
// case goroutine.
//
// Equality is determined via reflect.DeepEqual().
func ExpectEqual(t testing.TB, expected, actual interface{}) {
	ExpectEqualf(t, expected, actual, "ExpectEqual() failure.\n%# v", formatAsDiff(expected, actual))
}

// ExpectEqualIndex verifies that two objects are equals and calls Fail() to
// mark the test case as failed but let it continue.
//
// It is fine to call this function from another goroutine than the main test
// case goroutine.
//
// It is meant to be used in loops where a list of intrant->expected is
// processed so the assert failure message contains the index of the failing
// expectation.
//
// Equality is determined via reflect.DeepEqual().
func ExpectEqualIndex(t testing.TB, index int, expected, actual interface{}) {
	ExpectEqualf(t, expected, actual, "ExpectEqualIndex() failure.\nIndex: %d\n%# v", index, formatAsDiff(expected, actual))
}

// ExpectEqualf verifies that two objects are equals and calls Fail() to mark
// the test case as failed but let it continue.
//
// It is fine to call this function from another goroutine than the main test
// case goroutine.
//
// This functions enables specifying an arbitrary string on failure.
//
// Equality is determined via reflect.DeepEqual().
func ExpectEqualf(t testing.TB, expected, actual interface{}, format string, items ...interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		// Errorf() is thread-safe, t.Fatalf() is not.
		t.Errorf(Decorate(format), items...)
	}
}

// testingWriter is used by NewWriter().
type testingWriter struct {
	t testing.TB
	b bytes.Buffer
}

func (t testingWriter) Write(p []byte) (int, error) {
	n, err := t.b.Write(p)
	if err != nil || n != len(p) {
		return n, err
	}
	// Manually scan for lines.
	for {
		b := t.b.Bytes()
		i := bytes.Index(b, newLine)
		if i == -1 {
			break
		}
		t.t.Log(string(b[:i]))
		t.b.Next(i + 1)
	}
	return n, err
}

func (t testingWriter) Close() error {
	remaining := t.b.Bytes()
	if len(remaining) != 0 {
		t.t.Log(string(remaining))
	}
	return nil
}

// NewWriter adapts a testing.TB into a io.WriteCloser that can be used
// with to log.SetOutput().
//
// Don't forget to defer foo.Close().
func NewWriter(t testing.TB) io.WriteCloser {
	return &testingWriter{t: t}
}

// Private stuff.

// format creates a formatter that is both pretty and size limited.
//
// The limit is hardcoded to 2048. If you need more, edit the sources or send a
// pull request.
func format(i interface{}) fmt.Formatter {
	return &formatter{formatter: pretty.Formatter(i), limit: 2048}
}

type formatter struct {
	formatter fmt.Formatter
	limit     int
	size      int
}

func (f *formatter) Format(s fmt.State, c rune) {
	l := &limiter{s, f.limit, f.size}
	f.formatter.Format(l, c)
	f.size = l.size
}

// formatAsDiff returns a formatable object that will print itself as the diff
// between two objects.
func formatAsDiff(expected, actual interface{}) fmt.Formatter {
	return &formatterAsDiff{expected, actual}
}

type formatterAsDiff struct {
	expected, actual interface{}
}

func (f *formatterAsDiff) Format(s fmt.State, c rune) {
	// Format the items and escape those pesky ANSI codes.
	expected := strings.Replace(pretty.Sprintf("%# v", f.expected), "\033", "\\033", -1)
	actual := strings.Replace(pretty.Sprintf("%# v", f.actual), "\033", "\\033", -1)
	if strings.IndexByte(expected, '\n') == -1 && strings.IndexByte(actual, '\n') == -1 {
		fmt.Fprintf(s, "Expected: %s\nActual:   %s", expected, actual)
		return
	}
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  3,
	}
	_ = difflib.WriteUnifiedDiff(s, diff)
}

type limiter struct {
	fmt.State
	limit int
	size  int
}

func (l *limiter) Write(data []byte) (int, error) {
	var err error
	if l.size <= l.limit {
		if chunk := len(data); chunk+l.size > l.limit {
			chunk = l.limit - l.size
			if chunk != 0 {
				_, err = l.State.Write(data[:chunk])
			}
			_, err = l.State.Write([]byte("..."))
			l.size = l.limit + 1
		} else {
			_, err = l.State.Write(data)
		}
	}
	return len(data), err
}
