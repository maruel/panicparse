panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

Also usable as a library: [![GoDoc](https://godoc.org/github.com/maruel/panicparse/stack?status.svg)](https://godoc.org/github.com/maruel/panicparse/stack)

Screenshots
-----------

### Deep stack

![Screenshot](https://raw.githubusercontent.com/wiki/maruel/panicparse/screenshot.png "Screenshot")

### Multiple goroutines

![Screenshot 2](https://raw.githubusercontent.com/wiki/maruel/panicparse/screenshot2.png "Screenshot 2")


Usage
-----

    go get github.com/maruel/panicparse


### Piping a stack trace from another process

Run test and prints a concise stack trace upon deadlock in bash

    go test -v |& panicparse -all

`|&` tells bash to redirect stderr to stdout, it's an alias for `2>&1 |`.
panic() and Go's native deadlock detector always print to stderr.  Using `-all`
tells panicparse to print the output that was printed before and after the stack
trace, generally useful when piping `go test -v` in.

On Windows, a better trick can be used

    go test -v 2>&1 1>con: | panicparse

`-all` is not needed because stdout is not piped in.


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    panicparse stack.txt
