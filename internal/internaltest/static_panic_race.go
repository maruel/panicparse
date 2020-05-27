// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internaltest

// staticPanicRace is a snapshot created with:
//
//  go install -race github.com/maruel/panicparse/cmd/panic
//  panic race |& sed "s#$HOME##g"
//
// when installed within $GOPATH.
const staticPanicRace = `
GOTRACEBACK=all
==================
WARNING: DATA RACE
Read at 0x00c000014100 by goroutine 8:
  main.panicDoRaceRead()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:137 +0x3a
  main.panicRace.func2()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:154 +0x38

Previous write at 0x00c000014100 by goroutine 7:
  main.panicDoRaceWrite()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:132 +0x41
  main.panicRace.func1()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:151 +0x38

Goroutine 8 (running) created at:
  main.panicRace()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:153 +0xa1
  main.main()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:54 +0x6c8

Goroutine 7 (running) created at:
  main.panicRace()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:150 +0x7f
  main.main()
      /go/src/github.com/maruel/panicparse/cmd/panic/main.go:54 +0x6c8
==================
`
