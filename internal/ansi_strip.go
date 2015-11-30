// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import "io"

// NewANSIStripper processes out on the fly stripping out ANSI codes.
func NewANSIStripper(out io.Writer) io.Writer {
	return &ansiStripper{out: out}
}

// Private details.

type ansiState int

const (
	stateOutsideANSI ansiState = iota
	stateEscapeChar1
	stateEscapeChar2
)

const (
	escapeChar1 byte = '\x1b'
	escapeChar2 byte = '['
	colorChar   byte = 'm'
)

type ansiStripper struct {
	out   io.Writer
	state ansiState
}

func (a *ansiStripper) Write(p []byte) (int, error) {
	// TODO(maruel): This code is broken, it won't work if the escape code is
	// split across two Write() calls but it's good enough for now.
	toWriteStart := 0
	toWriteEnd := 0
	for i, ch := range p {
		switch a.state {
		case stateOutsideANSI:
			if ch == escapeChar1 {
				a.state = stateEscapeChar1
			}
		case stateEscapeChar1:
			switch ch {
			case escapeChar1:
				break
			case escapeChar2:
				a.state = stateEscapeChar2
				toWriteEnd = i - 1
			default:
				a.state = stateOutsideANSI
			}
		case stateEscapeChar2:
			if !(('0' <= ch && ch <= '9') || ch == ';') {
				if toWriteEnd > 0 {
					if _, err := a.out.Write(p[toWriteStart:toWriteEnd]); err != nil {
						return i, err
					}
				}
				toWriteStart = i + 1
				a.state = stateOutsideANSI
			}
		}
	}

	var err error
	if a.state == stateOutsideANSI {
		_, err = a.out.Write(p[toWriteStart:])
	}

	return len(p), err
}
