panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

[![GoDoc](https://godoc.org/github.com/maruel/panicparse/stack?status.svg)](https://godoc.org/github.com/maruel/panicparse/stack)
[![Build Status](https://travis-ci.org/maruel/panicparse.svg?branch=master)](https://travis-ci.org/maruel/panicparse)


![Screencast](https://raw.githubusercontent.com/wiki/maruel/panicparse/simple.gif "Screencast")

([Source](https://raw.githubusercontent.com/wiki/maruel/panicparse/simple.go))

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


Installation
------------

    go get github.com/maruel/panicparse/cmd/pp


Usage
-----

### Piping a stack trace from another process

#### TL;DR

   * Ubuntu: `|&`
   * OSX, [install bash 4+](README.md#updating-bash-on-osx) then: `|&`
   * Windows _or_ OSX with bash v3: `2>&1 |`


#### Longer version

Run test and prints a concise stack trace upon deadlock in bash v4:

    go test -v |&pp

`|&` tells bash to redirect stderr to stdout,
[it's an alias for `2>&1 |`](https://www.gnu.org/software/bash/manual/bash.html#Pipelines).
panic() and Go's native deadlock detector always print to stderr.

`pp` streams its stdin to stdout as long as it doesn't detect any panic.

On Windows or [OSX native bash (which is
3.2.57)](http://meta.ath0.com/2012/02/05/apples-great-gpl-purge/), use:

    go test -v 2>&1 | pp


### Investigate deadlock

On POSIX, use `Ctrl-\` to send SIGQUIT to your process, `pp` will ignore
the signal and will parse the stack trace.


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    pp stack.txt


### If you have `/usr/bin/pp` installed

You may have the Perl PAR Packager installed. Use long name `panicparse` then;

    go get github.com/maruel/panicparse


## Other screencast

![Screencast](https://raw.githubusercontent.com/wiki/maruel/panicparse/deadlock.gif "Screencast")

([Source](https://raw.githubusercontent.com/wiki/maruel/panicparse/deadlock.go))


## Updating bash on OSX

You can install bash v4+ on OSX via [homebrew](http://brew.sh) or
[macports](https://www.macports.org/).
