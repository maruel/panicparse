// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package webstack

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestSnapshotHandler(t *testing.T) {
	data := []string{
		"/debug",
		"/debug?augment=1",
		"/debug?maxmem=1",
		"/debug?maxmem=2097152",
		"/debug?similarity=exactflags",
		"/debug?similarity=exactlines",
		"/debug?similarity=anypointer",
		"/debug?similarity=anyvalue",
	}
	for _, url := range data {
		url := url
		t.Run(url, func(t *testing.T) {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			SnapshotHandler(w, req)
			if w.Code != 200 {
				t.Fatalf("%s: %d\n%s", url, w.Code, w.Body.String())
			}
		})
	}
}

func TestSnapshotHandler_Err(t *testing.T) {
	t.Parallel()
	data := []string{
		"/debug?augment=2",
		"/debug?maxmem=abc",
		"/debug?similarity=alike",
	}
	for _, url := range data {
		url := url
		t.Run(url, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			SnapshotHandler(w, req)
			if w.Code != 400 {
				t.Fatalf("%s: %d\n%s", url, w.Code, w.Body.String())
			}
		})
	}
}

func TestSnapshotHandler_Method_POST(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/debug", nil)
	w := httptest.NewRecorder()
	SnapshotHandler(w, req)
	if w.Code != 405 {
		t.Fatalf("%d\n%s", w.Code, w.Body.String())
	}
}

func TestSnapshotHandler_LargeMemory(t *testing.T) {
	// Try to create a stack frame over 1MiB in size when serialized to string.
	// This is tricky since this is dependent on many factors out of our control.
	// Do this by starting a lot of callbacks with a lot of arguments.
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := 0
	alive := make(chan struct{})
	// Assuming >400 bytes per goroutine, 2500 parallel goroutines is enough to
	// use more than 1MiB of call stack. We must not put it too high or it'll
	// crash on Travis.
	const parallel = 2500
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			alive <- struct{}{}
			dummy(ctx, &a, &a, &a, &a, &a, &a, &a, &a, &a)
		}()
	}
	for i := 0; i < parallel; i++ {
		<-alive
	}

	// Normal.
	req := httptest.NewRequest("GET", "/debug", nil)
	w := httptest.NewRecorder()
	SnapshotHandler(w, req)
	if w.Code != 200 {
		t.Fatalf("%d\n%s", w.Code, w.Body.String())
	}

	// Cut off. That's 1<<20 + 1
	req = httptest.NewRequest("GET", "/debug?maxmem=1048577", nil)
	w = httptest.NewRecorder()
	SnapshotHandler(w, req)
	// It can result in a 500 because the cut off is arbitrary, making parsing to
	// fail.
	if w.Code != 200 && w.Code != 500 {
		t.Fatalf("%d\n%s", w.Code, w.Body.String())
	}

	cancel()
	wg.Wait()
}

func BenchmarkSnapshotHandle(b *testing.B) {
	// TODO(maruel): We should hook runtime.Stack() to make it a deterministic
	// output with internaltest.StaticPanicwebOutput().
	b.ReportAllocs()
	req := httptest.NewRequest("GET", "/", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		SnapshotHandler(w, req)
		if w.Code != 200 {
			b.Fatalf("%d\n%s", w.Code, w.Body.String())
		}
	}
}

func dummy(ctx context.Context, a1, a2, a3, a4, a5, a6, a7, a8, a9 *int) {
	<-ctx.Done()
}
