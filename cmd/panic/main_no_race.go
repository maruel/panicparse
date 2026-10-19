// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Disable intentional race scenarios in builds without race instrumentation.

//go:build !race

package main

const raceEnabled = false
