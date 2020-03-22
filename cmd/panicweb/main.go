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

	"github.com/maruel/panicparse/cmd/panicweb/internal"
	"github.com/maruel/panicparse/stack/webstack"
	"github.com/mattn/go-colorable"
)

func main() {
	allowremote := flag.Bool("allowremote", false, "allows access from non-localhost")
	sleep := flag.Bool("wait", false, "sleep instead of crashing")
	port := flag.Int("port", 0, "specify a port number, defaults to a ephemeral port")
	flag.Parse()

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
	http.HandleFunc("/panicparse", webstack.SnapshotHandler)
	go http.Serve(ln, http.DefaultServeMux)

	// Start many clients.
	url := "http://" + ln.Addr().String() + "/"
	for i := 0; i < 10; i++ {
		internal.GetAsync(url + "url1")
	}
	for i := 0; i < 3; i++ {
		internal.GetAsync(url + "url2")
	}

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
	w.hung <- struct{}{}
	<-w.unblock
	return 0, nil
}
