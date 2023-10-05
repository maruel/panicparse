# panicparse

Parses panic stack traces, densifies and deduplicates goroutines with similar
stack traces. Helps debugging crashes and deadlocks in heavily parallelized
process.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/maruel/panicparse/v2/stack)](https://pkg.go.dev/github.com/maruel/panicparse/v2/stack)
[![codecov](https://codecov.io/gh/maruel/panicparse/branch/main/graph/badge.svg?token=izj1cLjUi3)](https://codecov.io/gh/maruel/panicparse)
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes#-pretty-print-panic-messages)


panicparse helps make sense of Go crash dumps:

![Screencast](https://raw.githubusercontent.com/wiki/maruel/panicparse/parse.gif "Screencast")


## Features

   * Race detector support, e.g. it can parse output produced by `go test -race`
   * HTML export.
   * Easy to use as an [HTTP Handler
     middleware](https://pkg.go.dev/github.com/maruel/panicparse/v2/stack#example-package-HttpHandlerMiddleware).
   * High performance parsing.
   * [HTTP web server](https://pkg.go.dev/github.com/maruel/panicparse/v2/stack/webstack#SnapshotHandler)
     that serves a very tight and swell snapshot of your goroutines, much more
     readable than [net/http/pprof](https://pkg.go.dev/net/http/pprof).
   * &gt;50% more compact output than original stack dump yet more readable.
   * Deduplicates redundant goroutine stacks. Useful for large server crashes.
   * Arguments as pointer IDs instead of raw pointer values.
   * Pushes stdlib-only stacks at the bottom to help focus on important code.
   * Parses the source files if available to augment the output.
   * Works on any platform supported by Go, including Windows, macOS, linux.
   * Full go module support.
   * Requires >=go1.17. Use v2.3.1 for older Go versions.


## Installation

    go install github.com/maruel/panicparse/v2/cmd/pp@latest


## Usage

### Piping a stack trace from another process

#### TL;DR

   * Ubuntu (bash v4 or zsh): `|&`
   * macOS, [install bash 4+](README.md#updating-bash-on-macos), then: `|&`
   * Windows _or_ macOS with stock bash v3: `2>&1 |`
   * [Fish](http://fishshell.com/) shell: `&|`


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


**Windows or macOS native bash** [(which is
3.2.57)](http://meta.ath0.com/2012/02/05/apples-great-gpl-purge/): They don't
have this shortcut, so use the long form:

    go test -v 2>&1 | pp


**Fish**: `&|` redirects stderr and stdout. It's an alias for `2>&1 |`
([fish piping](https://fishshell.com/docs/current/language.html#piping)):

    go test -v &| pp


**PowerShell**: [It has broken `2>&1` redirection](https://connect.microsoft.com/PowerShell/feedback/details/765551/in-powershell-v3-you-cant-redirect-stderr-to-stdout-without-generating-error-records). The workaround is to shell out to cmd.exe. :(


### Investigate deadlock

On POSIX, use `Ctrl-\` to send SIGQUIT to your process, `pp` will ignore
the signal and will parse the stack trace.


### Parsing from a file

To dump to a file then parse, pass the file path of a stack trace

    go test 2> stack.txt
    pp stack.txt


## Tips

### Disable inlining

The Go toolchain inlines functions when it can. This causes traces to be less
informative. Optimization also interfere with traces. You can use the following
to help diagnosing issues:

    go install -gcflags '-N -l' path/to/foo
    foo |& pp

or

    go test -gcflags '-N -l' ./... |& pp


Run `go tool compile -help` to get the full list of valid values for -gcflags.


### GOTRACEBACK

By default, [`GOTRACEBACK`](https://golang.org/pkg/runtime/) defaults to
`single`, which means that a panic will only return the current goroutine trace
alone. To get all goroutines trace and not just the crashing one, set the
environment variable:

    export GOTRACEBACK=all

or `set GOTRACEBACK=all` on Windows. Probably worth to put it in your `.bashrc`.


### Updating bash on macOS

Install bash v4+ on macOS via [homebrew](http://brew.sh) or
[macports](https://www.macports.org/). Your future self will appreciate having
done that.


### If you have `/usr/bin/pp` installed

If you try `pp` for the first time and you get:

    Creating tables and indexes...
    Done.

and/or

    /usr/bin/pp5.18: No input files specified

you may be running the _Perl PAR Packager_ instead of panicparse.

You have two choices, either you put `$GOPATH/bin` at the beginning of `$PATH`
or use long name `panicparse` with:

    go install github.com/maruel/panicparse/v2@latest

then using `panicparse` instead of `pp`:

    go test 2> panicparse

Hint: You may also use shell aliases

    alias gp=panicparse
    go test 2> gp

    alias p=panicparse
    go test 2> p


### webstack in action

The
[webstack.SnapshotHandler](https://pkg.go.dev/github.com/maruel/panicparse/v2/stack/webstack#SnapshotHandler)
http.Handler enables glancing at at a snapshot of your process trivially:

![Screencast](https://raw.githubusercontent.com/wiki/maruel/panicparse/panicparse_webstack.gif "Screencast")


## Authors

`panicparse` was created with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel) and
[friends](https://github.com/maruel/panicparse/graphs/contributors).
