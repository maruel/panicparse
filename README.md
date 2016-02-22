panicparse
==========

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

[![GoDoc](https://godoc.org/github.com/maruel/panicparse/stack?status.svg)](https://godoc.org/github.com/maruel/panicparse/stack)
[![Build Status](https://travis-ci.org/maruel/panicparse.svg?branch=master)](https://travis-ci.org/maruel/panicparse)
[![Chat](https://badges.gitter.im/maruel/panicparse.svg)](https://gitter.im/maruel/panicparse)


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
     * Warning: please pin the version (e.g. vendor it) as breaking changes are
       planned.
   * Works on Windows.


Installation
------------

    go get github.com/maruel/panicparse/cmd/pp


Usage
-----

### Piping a stack trace from another process

#### TL;DR

   * Ubuntu (bash v4 or zsh): `|&`
   * OSX, [install bash 4+](README.md#updating-bash-on-osx), then: `|&`
   * Windows _or_ OSX with stock bash v3: `2>&1 |`
   * [Fish](http://fishshell.com/) shell: `^|`


#### Longer version

`pp` streams its stdin to stdout as long as it doesn't detect any panic.
`panic()` and Go's native deadlock detector [print to
stderr](https://golang.org/src/runtime/panic1.go) via the native [`print()`
function](https://golang.org/pkg/builtin/#print).


**Bash v4** or **zsh**: `|&` tells the shell to redirect stderr to stdout,
it's an alias for `2>&1 |` ([bash
v4](https://www.gnu.org/software/bash/manual/bash.html#Pipelines),
[zsh](http://zsh.sourceforge.net/Doc/Release/Shell-Grammar.html#Simple-Commands-_0026-Pipelines)):

    go test -v |&pp


**Windows or OSX native bash** [(which is
3.2.57)](http://meta.ath0.com/2012/02/05/apples-great-gpl-purge/): They don't
have this shortcut, so use the long form:

    go test -v 2>&1 | pp


**Fish**: It uses [^ for stderr
redirection](http://fishshell.com/docs/current/tutorial.html#tut_pipes_and_redirections)
so the shortcut is `^|`:

    go test -v ^|pp


**PowerShell**: [It has broken `2>&1` redirection](https://connect.microsoft.com/PowerShell/feedback/details/765551/in-powershell-v3-you-cant-redirect-stderr-to-stdout-without-generating-error-records). The workaround is to shell out to cmd.exe. :(


### Investigate deadlock

On POSIX, use `Ctrl-\` to send SIGQUIT to your process, `pp` will ignore
the signal and will parse the stack trace.


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    pp stack.txt


Other screencast
----------------

![Screencast](https://raw.githubusercontent.com/wiki/maruel/panicparse/deadlock.gif "Screencast")

([Source](https://raw.githubusercontent.com/wiki/maruel/panicparse/deadlock.go))


Tips
----

### GOTRACEBACK

Starting with Go 1.6, [`GOTRACEBACK`](https://golang.org/pkg/runtime/) defaults
to `single` instead of `all` / `1` that was used in 1.5 and before. To get all
goroutines trace and not just the crashing one, set the environment variable:

    export GOTRACEBACK=all

or `set GOTRACEBACK=all` on Windows. Probably worth to put it in your `.bashrc`.


### Updating bash on OSX

Install bash v4+ on OSX via [homebrew](http://brew.sh) or
[macports](https://www.macports.org/). Your future self will appreciate having
done that.


### If you have `/usr/bin/pp` installed

You may have the Perl PAR Packager installed. Use long name `panicparse` then;

    go get github.com/maruel/panicparse
    
### You can also use it as a library to log your errors like this
```go
defaultPalette = stack.Palette{
	EOLReset:               resetFG,
	RoutineFirst:           ansi.ColorCode("magenta+b"),
	CreatedBy:              ansi.LightBlack,
	Package:                ansi.ColorCode("default+b"),
	SourceFile:             resetFG,
	FunctionStdLib:         ansi.Green,
	FunctionStdLibExported: ansi.ColorCode("green+b"),
	FunctionMain:           ansi.ColorCode("yellow+b"),
	FunctionOther:          ansi.Red,
	FunctionOtherExported:  ansi.ColorCode("red+b"),
	Arguments:              resetFG,
}

if exception := recover(); exception != nil {
	inputBuffer := &bytes.Buffer{}
	outputBuffer := &bytes.Buffer{}
	trace := make([]byte, 1<<16)
	n := runtime.Stack(trace, true)
	inputBuffer.WriteString(fmt.Sprintf("panic recover %v stack trace %d bytes\n", exception, n))
	inputBuffer.Write(trace)
	goroutines, _ := stack.ParseDump(inputBuffer, outputBuffer)
	fullPath := false
	buckets := stack.SortBuckets(stack.Bucketize(goroutines, stack.AnyPointer))
	srcLen, pkgLen := stack.CalcLengths(buckets, fullPath)
	for _, bucket := range buckets {
		_, _ = io.WriteString(outputBuffer, defaultPalette.BucketHeader(&bucket, fullPath, len(buckets) > 1))
		_, _ = io.WriteString(outputBuffer, defaultPalette.StackLines(&bucket.Signature, srcLen, pkgLen, fullPath))
	}
	println(outputBuffer.String())
}
```
