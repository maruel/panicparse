// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal implements the handlers for panicweb so they are in a
// separate package than "main".
package internal

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Unblock unblocks one http server handler.
var Unblock = make(chan struct{})

// GetAsync does an HTTP GET to the URL but leaves the actual fetching to a
// goroutine.
func GetAsync(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("get %s: %v", url, err)
	}
	go func() {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("failed to read: %v", err)
		}
		resp.Body.Close()
		log.Fatal("the goal is to not complete this request")
	}()
}

// URL1Handler is a http.HandlerFunc that hangs.
func URL1Handler(w http.ResponseWriter, req *http.Request) {
	// Respond the HTTP header to unblock the http.Get() function.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", "100000")
	w.WriteHeader(200)
	b := [4096]byte{}
	w.Write(b[:])
	<-Unblock
}

// URL2Handler is a http.HandlerFunc that hangs.
func URL2Handler(w http.ResponseWriter, req *http.Request) {
	// Respond the HTTP header to unblock the http.Get() function.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", "100000")
	w.WriteHeader(200)
	b := [4096]byte{}
	w.Write(b[:])
	<-Unblock
}
