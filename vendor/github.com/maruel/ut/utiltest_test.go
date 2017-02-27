// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ut

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func ExampleAssertEqual() {
	// For a func TestXXX(t *testing.T)
	t := &testing.T{}
	AssertEqual(t, "10", strconv.Itoa(10))
}

func ExampleAssertEqualIndex() {
	// For a func TestXXX(t *testing.T)
	t := &testing.T{}

	data := []struct {
		in       int
		expected string
	}{
		{9, "9"},
		{11, "11"},
	}
	for i, item := range data {
		// Call a function to test.
		actual := strconv.Itoa(item.in)
		// Then do an assert as a one-liner.
		AssertEqualIndex(t, i, item.expected, actual)
	}
}

func ExampleExpectEqual() {
	// For a func TestXXX(t *testing.T)
	t := &testing.T{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// ExpectEqual* flavors are safe to call in other goroutines.
		ExpectEqual(t, "10", strconv.Itoa(10))
	}()
	wg.Wait()
}

func ExampleNewWriter() {
	// For a func TestXXX(t *testing.T)
	t := &testing.T{}

	out := NewWriter(t)
	defer out.Close()

	logger := log.New(out, "Foo:", 0)

	// These will be included in the test output only if the test case fails.
	logger.Printf("Q: What is the answer to life the universe and everything?")
	logger.Printf("A: %d", 42)
}

// AssertEqual*

func TestAssertEqual(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	AssertEqual(t, &j, i)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestAssertEqualFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	var err interface{} = 1
	defer func() {
		if err != nil {
			t.Fatalf("unexpected %s", err)
		}
		// Abuse the testing framework so this test is not marked as failed. It is
		// not really skipped but that's the only way to set
		// testing.T.finished = true.
		t.SkipNow()
	}()
	defer func() {
		err = recover()
	}()
	AssertEqual(t2, true, false)
	// This line is never executed.
	t.Fail()
}

func TestAssertEqualIndex(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	AssertEqualIndex(t, 24, &j, i)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestAssertEqualIndexFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	var err interface{} = 1
	defer func() {
		if err != nil {
			t.Fatalf("unexpected %s", err)
		}
		// Abuse the testing framework so this test is not marked as failed. It is
		// not really skipped but that's the only way to set
		// testing.T.finished = true.
		t.SkipNow()
	}()
	defer func() {
		err = recover()
	}()
	AssertEqualIndex(t2, 24, true, false)
	// This line is never executed.
	t.Fail()
}

func TestAssertEqualf(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	AssertEqualf(t, &j, i, "foo %s %d", "bar", 2)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestAssertEqualfFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	var err interface{} = 1
	defer func() {
		if err != nil {
			t.Fatalf("unexpected %s", err)
		}
		// Abuse the testing framework so this test is not marked as failed. It is
		// not really skipped but that's the only way to set
		// testing.T.finished = true.
		t.SkipNow()
	}()
	defer func() {
		err = recover()
	}()
	AssertEqualf(t2, true, false, "foo %s %d", "bar", 2)
	// This line is never executed.
	t.Fail()

}

// ExpectEqual*

func TestExpectEqual(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	ExpectEqual(t, &j, i)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestExpectEqualFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	completed := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Ensure ExpectEqual* can be run from a goroutine.
		defer wg.Done()
		ExpectEqual(t2, true, false)
		completed = true
	}()
	wg.Wait()
	if !completed {
		t.Fatal("didn't complete")
	}
}

func TestExpectEqualIndex(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	ExpectEqualIndex(t, 24, &j, i)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestExpectEqualIndexFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	completed := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Ensure ExpectEqual* can be run from a goroutine.
		defer wg.Done()
		ExpectEqualIndex(t2, 24, true, false)
		completed = true
	}()
	wg.Wait()
	if !completed {
		t.Fatal("didn't complete")
	}
}

func TestExpectEqualf(t *testing.T) {
	t.Parallel()
	j := true
	var i interface{} = &j
	ExpectEqualf(t, &j, i, "foo %s %d", "bar", 2)
	if t.Failed() {
		t.Fatal("Expected success")
	}
}

func TestExpectEqualfFail(t *testing.T) {
	t.Parallel()
	// Abuse the testing framework to mark a fake test as failed.
	t2 := &testing.T{}
	completed := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Ensure ExpectEqual* can be run from a goroutine.
		defer wg.Done()
		ExpectEqualf(t2, true, false, "foo %s %d", "bar", 2)
		completed = true
	}()
	wg.Wait()
	if !completed {
		t.Fatal("didn't complete")
	}

}

// Other

type stubTB struct {
	*testing.T
	out []string
}

func (s *stubTB) Log(args ...interface{}) {
	if len(args) != 1 {
		s.FailNow()
	}
	str, ok := args[0].(string)
	if !ok {
		panic("Unexpected Log() call with something else than string")
	}
	s.out = append(s.out, str)
}

func TestNewWriter(t *testing.T) {
	t.Parallel()
	tStub := &stubTB{T: t}
	out := NewWriter(tStub)
	logger := log.New(out, "Foo:", 0)
	logger.Printf("Q: What is the answer to life the universe and everything?")
	logger.Printf("A: %d", 42)
	ExpectEqual(t, nil, out.Close())
	expected := []string{
		"Foo:Q: What is the answer to life the universe and everything?",
		"Foo:A: 42",
	}
	AssertEqual(t, expected, tStub.out)
}

func TestTruncatePath(t *testing.T) {
	t.Parallel()
	data := []struct{ in, expected string }{
		{"foo", "foo"},
		{filepath.Join("foo", "bar"), filepath.Join("foo", "bar")},
		{filepath.Join("foo", "bar", "baz"), filepath.Join("bar", "baz")},
	}
	for i, line := range data {
		AssertEqualIndex(t, i, line.expected, truncatePath(line.in))
	}
}

func TestFormatter(t *testing.T) {
	t.Parallel()
	large := strings.Repeat("0123456789abcedf", 2048/16+1)
	expected := strings.Repeat("0123456789abcedf", 2048/16) + "..."
	AssertEqual(t, expected, fmt.Sprintf("%s", format(large)))
}

func TestDiffEscapeANSI(t *testing.T) {
	t.Parallel()
	actual := fmt.Sprintf("% #v", formatAsDiff("\033[31mHi", "\033[32mHi"))
	expected := "Expected: \\033[31mHi\nActual:   \\033[32mHi"
	AssertEqual(t, expected, actual)
}
