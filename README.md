panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

[![GoDoc](https://godoc.org/github.com/maruel/panicparse/stack?status.svg)](https://godoc.org/github.com/maruel/panicparse/stack)
[![Build Status](https://travis-ci.org/maruel/panicparse.svg?branch=master)](https://travis-ci.org/maruel/panicparse)
[![Coverage Status](https://img.shields.io/coveralls/maruel/panicparse.svg)](https://coveralls.io/r/maruel/panicparse?branch=master)


Features
--------

   * >50% more compact output than original stack dump yet more readable.
   * Exported symbols are bold, private symbols are darker.
   * Stdlib is green, main is yellow, rest is red.
   * Deduplicate redundant goroutine stacks. Useful for large server crashes.
   * Arguments as pointer IDs instead of raw pointer values.
   * Pushes stdlib-only stacks at the bottom to help focus on important code.
   * Usable as a library!
   * Works on Windows.


Screenshot
----------

Converts this [hard to read stack dump](https://raw.githubusercontent.com/wiki/maruel/panicparse/sample3.txt) into something nicer:

![Screenshot](https://raw.githubusercontent.com/wiki/maruel/panicparse/screenshot3.png "Screenshot")


Installation
------------

    go get github.com/maruel/panicparse


Usage
-----

### Piping a stack trace from another process

Run test and prints a concise stack trace upon deadlock in bash

    go test -v |& panicparse

`|&` tells bash to redirect stderr to stdout, it's an alias for `2>&1 |`.
panic() and Go's native deadlock detector always print to stderr.

On Windows, a better trick can be used so that only stderr is piped to
panicparse, leaving stdout alone:

    go test -v 2>&1 1>con: | panicparse


### Investigate deadlock

On POSIX, use `Ctrl-\` to send SIGQUIT to your process, panicparse will ignore
the signal and will parse the stack trace.


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    panicparse stack.txt
