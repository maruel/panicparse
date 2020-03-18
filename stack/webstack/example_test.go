// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package webstack_test

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/maruel/panicparse/stack/webstack"
)

func ExampleSnapshotHandler() {
	http.HandleFunc("/debug/panicparse", webstack.SnapshotHandler)

	// Access as http://localhost:6060/debug/panicparse
	log.Println(http.ListenAndServe("localhost:6060", nil))
}

func ExampleSnapshotHandler_complex() {
	// This example does a few things:
	// - Disables the use of argument "augment" to limit CPU/disk overhead.
	// - Lowers the "maxmem" default value to reduce memory pressure in worst
	//   case.
	// - Serializes handler to one at a time.
	// - Throttles requests to once per second.

	const delay = time.Second
	mu := sync.Mutex{}
	var last time.Time
	http.HandleFunc("/debug/panicparse", func(w http.ResponseWriter, req *http.Request) {
		// Serialize the handler.
		mu.Lock()
		defer mu.Unlock()

		// Throttle requests.
		if time.Since(last) < delay {
			http.Error(w, "retry later", http.StatusTooManyRequests)
			return
		}

		// Must be called before touching req.Form.
		req.ParseForm()
		// Disable source scanning.
		req.Form.Set("augment", "0")
		// Default to use up to 32MiB for the goroutines snapshot.
		if req.FormValue("maxmem") == "" {
			req.Form.Set("maxmem", "33554432")
		}

		webstack.SnapshotHandler(w, req)
		last = time.Now()
	})

	// Access as http://localhost:6060/debug/panicparse
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
