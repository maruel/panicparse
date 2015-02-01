panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.


Usage
-----

### Install

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


Sample output
-------------

TestBufferStressShort-16 deadlocked with a total of 1012 goroutines. There was
respectively 491 and 512 goroutines with the exact same last 2 stack traces. In
this case, this densified a 12239 lines stack trace into a 29 lines highly
readable summary. The third one is actually where the deadlock was caused as it
missed a sync.Cond signal.

    $ go test -v |& panicparse -all

    === RUN TestBufferReaderLaggard-16
    --- PASS: TestBufferReaderLaggard-16 (0.00s)
    === RUN TestBufferWriteClosed-16
    --- PASS: TestBufferWriteClosed-16 (0.00s)
    === RUN TestBufferFlushEmpty-16
    --- PASS: TestBufferFlushEmpty-16 (0.00s)
    === RUN TestBufferFlushBlocking-16
    --- PASS: TestBufferFlushBlocking-16 (0.00s)
    === RUN TestBufferStressShort-16
    fatal error: all goroutines are asleep - deadlock!

    exit status 2
    FAIL    github.com/maruel/circular      0.228s

    1: chan receive
      testing.go:556: testing.RunTests
      testing.go:485: testing.(*M).Run
      _testmain.go:94: main.main
    1: semacquire
      waitgroup.go:132: sync.(*WaitGroup).Wait
      circular_test.go:445: circular.stressTest
      circular_test.go:399: circular.TestBufferStressShort
      testing.go:447: testing.tRunner
      testing.go:555: testing.RunTests
    1: semacquire
      cond.go:62: sync.(*Cond).Wait
      circular.go:115: circular.(*Buffer).Write
      circular_test.go:467: circular.writeOk
      circular_test.go:439: circular.func·028
      circular_test.go:546: circular.func·029
      circular_test.go:547: circular.(*End).Go
    497: semacquire
      mutex.go:66: sync.(*Mutex).Lock
      circular.go:66: circular.(*Buffer).Write
      circular_test.go:467: circular.writeOk
      circular_test.go:439: circular.func·028
      circular_test.go:546: circular.func·029
      circular_test.go:547: circular.(*End).Go
    512: semacquire
      cond.go:62: sync.(*Cond).Wait
      circular.go:264: circular.(*Buffer).WriteTo
      circular_test.go:430: circular.func·027
      circular_test.go:433: circular.stressTest
