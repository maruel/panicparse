// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:build !go1.17
// +build !go1.17

package stack

// See https://github.com/maruel/panicparse/issues/61 for explanation.
const combinedAggregateArgs = false
