// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package webstack provides a http.HandlerFunc that serves a snapshot similar
// to net/http/pprof.Index().
//
// Contrary to net/http/pprof, the handler is not automatically registered.
package webstack

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"

	"github.com/maruel/panicparse/v2/stack"
)

// SnapshotHandler implements http.HandlerFunc to returns a panicparse HTML
// format for a snapshot of the current goroutines.
//
// For best results, compile the executable with inlining disabled with
// -gcflags '-l'.
//
// Arguments are passed as form values. If you want to change the default,
// override the form values in a wrapper as shown in the example.
//
// The implementation is designed to be reasonably fast, it currently does a
// small amount of disk I/O only for file presence.
//
// It is a direct replacement for "/debug/pprof/goroutine?debug=2" handler in
// net/http/pprof.
//
// augment: (default: 1) When set to 1, panicparse tries to find the sources on
// disk to improve the display of arguments based on type information. This is
// slower and should be avoided on high utilization server.
//
// maxmem: (default: 67108864) maximum amount of temporary memory to use to
// generate a snapshot. In practice at least the double of this is used.
// Minimum is 1048576.
//
// similarity: (default: "anypointer") Can be one of stack.Similarity value in
// lowercase: "exactflags", "exactlines", "anypointer" or "anyvalue".
func SnapshotHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	maxmem := 64 << 20
	if s := req.FormValue("maxmem"); s != "" {
		var err error
		if maxmem, err = strconv.Atoi(s); err != nil {
			http.Error(w, "invalid maxmem value", http.StatusBadRequest)
			return
		}
	}
	opts := stack.DefaultOpts()
	if s := req.FormValue("augment"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 || v > 1 {
			http.Error(w, "invalid augment value", http.StatusBadRequest)
			return
		}
		if v == 0 {
			opts.AnalyzeSources = false
		}
	}
	c, err := snapshot(maxmem, opts)
	if err != nil {
		http.Error(w, "failed to process the snapshot, try a larger maxmem value", http.StatusInternalServerError)
		return
	}

	var s stack.Similarity
	switch req.FormValue("similarity") {
	case "exactflags":
		s = stack.ExactFlags
	case "exactlines":
		s = stack.ExactLines
	case "anypointer", "":
		s = stack.AnyPointer
	case "anyvalue":
		s = stack.AnyValue
	default:
		http.Error(w, "invalid similarity value", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = c.Aggregate(s).ToHTML(w, "")
}

// snapshot returns a Context based on the snapshot of the stacks of the
// current process.
func snapshot(maxmem int, opts *stack.Opts) (*stack.Snapshot, error) {
	// We don't know how big the buffer needs to be to collect all the
	// goroutines. Start with 1 MB and try a few times, doubling each time. Give
	// up and use a truncated trace if maxmem is not enough.
	buf := make([]byte, 1<<20)
	if maxmem < len(buf) {
		maxmem = len(buf)
	}
	for i := 0; ; i++ {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		if len(buf) >= maxmem {
			break
		}
		l := len(buf) * 2
		if l > maxmem {
			l = maxmem
		}
		buf = make([]byte, l)
	}
	s, _, err := stack.ScanSnapshot(bytes.NewReader(buf), ioutil.Discard, opts)
	// That's expected.
	if err == io.EOF {
		err = nil
	}
	return s, err
}
