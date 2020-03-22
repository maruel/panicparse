// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// panicweb implements a simulation of a web server that panics.
//
// It starts a web server, a few handlers and a few hanging clients, then
// panics.
//
// It loads both panicparse's http handler and pprof's one for comparison.
//
// It is separate from the panic tool because importing "net/http" creates a
// background thread, which breaks the "asleep" panic case in tool panic.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/maruel/panicparse/cmd/panicweb/internal"
	"github.com/maruel/panicparse/stack/webstack"
	"github.com/mattn/go-colorable"
)

func main() {
	allowremote := flag.Bool("allowremote", false, "allows access from non-localhost; implies -wait")
	sleep := flag.Bool("wait", false, "sleep instead of crashing")
	port := flag.Int("port", 0, "specify a port number, defaults to a ephemeral port; implies -wait")
	limit := flag.Bool("limit", false, "throttle, port limit")
	flag.Parse()

	if *port != 0 || *allowremote {
		*sleep = true
	}
	addr := fmt.Sprintf(":%d", *port)
	if !*allowremote {
		addr = "localhost" + addr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on localhost: %v", err)
	}
	http.HandleFunc("/url1", internal.URL1Handler)
	http.HandleFunc("/url2", internal.URL2Handler)
	if *limit {
		// This is similar to ExampleSnapshotHandler_complex in stack/webstack,
		// albeit form values are not altered.
		const delay = time.Second
		mu := sync.Mutex{}
		var last time.Time
		http.HandleFunc("/panicparse", func(w http.ResponseWriter, req *http.Request) {
			// Only allow requests from localhost or in the 100.64.x.x/10 IPv4 range
			// (e.g. Tailscale).
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
			log.Printf("- %s: %t", req.RemoteAddr, ok)
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

			webstack.SnapshotHandler(w, req)
			last = time.Now()
		})
	} else {
		http.HandleFunc("/panicparse", webstack.SnapshotHandler)
	}
	go http.Serve(ln, http.DefaultServeMux)

	// Start many clients.
	url := "http://" + ln.Addr().String() + "/"
	for i := 0; i < 10; i++ {
		internal.GetAsync(url + "url1")
	}
	for i := 0; i < 3; i++ {
		internal.GetAsync(url + "url2")
	}

	// Try to get something hung in package golang.org/x/unix.
	wait := make(chan struct{})
	go func() {
		wait <- struct{}{}
		sysHang()
	}()
	<-wait

	// It's convoluted but colorable is the only go module used by panicparse
	// that is both versioned and can be hacked to call back user code.
	w := writeHang{hung: make(chan struct{}), unblock: make(chan struct{})}
	v := colorable.NewNonColorable(&w)
	go v.Write([]byte("foo bar"))
	<-w.hung

	if *sleep {
		fmt.Printf("Compare:\n- %spanicparse\n- %sdebug/pprof/goroutine?debug=2\n", url, url)
		<-make(chan struct{})
	} else {
		panic("Here's a snapshot of a normal web server.")
	}
}

type writeHang struct {
	hung    chan struct{}
	unblock chan struct{}
}

func (w *writeHang) Write(b []byte) (int, error) {
	runtime.LockOSThread()
	w.hung <- struct{}{}
	<-w.unblock
	return 0, nil
}
