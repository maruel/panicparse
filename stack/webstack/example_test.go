// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package webstack_test

import (
	"log"
	"net"
	"net/http"
	"strings"
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
	// - Enables "augment" by default, can be disabled manually with "?augment=0".
	// - Forces the "maxmem" value to reduce memory pressure in worst case.
	// - Serializes handler to one at a time.
	// - Throttles requests to once per second.
	// - Limit request source IP to localhost and 100.64.x.x/10. (e.g.
	//   Tailscale).

	const delay = time.Second
	mu := sync.Mutex{}
	var last time.Time
	http.HandleFunc("/debug/panicparse", func(w http.ResponseWriter, req *http.Request) {
		// Only allow requests from localhost or in the 100.64.x.x/10 IPv4 range.
		ok := false
		if i := strings.LastIndexByte(req.RemoteAddr, ':'); i != -1 {
			switch ip := req.RemoteAddr[:i]; ip {
			case "localhost", "127.0.0.1", "[::1]", "::1":
				ok = true
			default:
				p := net.ParseIP(ip).To4()
				ok = p != nil && p[0] == 100 && p[1] >= 64 && p[1] < 128
			}
		}
		if !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

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
		// Enable source scanning by default.
		if req.FormValue("augment") == "" {
			req.Form.Set("augment", "1")
		}
		// Reduces maximum memory usage to 32MiB (from 64MiB) for the goroutines
		// snapshot.
		req.Form.Set("maxmem", "33554432")

		webstack.SnapshotHandler(w, req)
		last = time.Now()
	})

	// Access as http://localhost:6060/debug/panicparse
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
