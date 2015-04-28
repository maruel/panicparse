panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

Also usable as a library: [![GoDoc](https://godoc.org/github.com/maruel/panicparse/stack?status.svg)](https://godoc.org/github.com/maruel/panicparse/stack)


Screenshot
----------

Converts this [hard to read stack dump](https://raw.githubusercontent.com/wiki/maruel/panicparse/sample3.txt) into something nicer:

![Screenshot](https://raw.githubusercontent.com/wiki/maruel/panicparse/screenshot3.png "Screenshot")


Usage
-----

    go get github.com/maruel/panicparse


### Piping a stack trace from another process

Run test and prints a concise stack trace upon deadlock in bash

    go test -v |& panicparse

`|&` tells bash to redirect stderr to stdout, it's an alias for `2>&1 |`.
panic() and Go's native deadlock detector always print to stderr.

On Windows, a better trick can be used so that only stderr is piped to
panicparse, leaving stdout alone:

    go test -v 2>&1 1>con: | panicparse


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    panicparse stack.txt
