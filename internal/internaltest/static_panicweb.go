// Copyright 2020 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internaltest

// staticPanicweb is a snapshot created by running:
//
//  bash static_panicweb.sh
//
// Not using go:generate here since it takes 2 minutes to complete.
const staticPanicweb = `goroutine 135 [running]:
runtime/pprof.writeGoroutineStacks(0x91be20, 0xc0003aa0e0, 0x0, 0x0)
	/goroot/src/runtime/pprof/pprof.go:665 +0x9d
runtime/pprof.writeGoroutine(0x91be20, 0xc0003aa0e0, 0x2, 0x40e256, 0xc0003a2b00)
	/goroot/src/runtime/pprof/pprof.go:654 +0x44
runtime/pprof.(*Profile).WriteTo(0xbe85e0, 0x91be20, 0xc0003aa0e0, 0x2, 0xc0003aa0e0, 0xc00021b9b0)
	/goroot/src/runtime/pprof/pprof.go:329 +0x3da
net/http/pprof.handler.ServeHTTP(0xc000614161, 0x9, 0x9241a0, 0xc0003aa0e0, 0xc0001b2700)
	/goroot/src/net/http/pprof/pprof.go:248 +0x33a
net/http/pprof.Index(0x9241a0, 0xc0003aa0e0, 0xc0001b2700)
	/goroot/src/net/http/pprof/pprof.go:271 +0x735
net/http.HandlerFunc.ServeHTTP(0x8a4508, 0x9241a0, 0xc0003aa0e0, 0xc0001b2700)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0003aa0e0, 0xc0001b2700)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0003aa0e0, 0xc0001b2700)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0001a63c0, 0x924de0, 0xc000190780)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 1 [chan receive, 2 minutes]:
main.main()
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:78 +0x7be

goroutine 34 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc000242000, 0xc000038300)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc000242000, 0xc000038300)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc000242000, 0xc000038300)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc000242000, 0xc000038300)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0001a6000, 0x924de0, 0xc000020780)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 5 [IO wait]:
internal/poll.runtime_pollWait(0x7f5224075f48, 0x72, 0x0)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc318, 0x72, 0x0, 0x0, 0x88206d)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Accept(0xc0000cc300, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:384 +0x1d4
net.(*netFD).accept(0xc0000cc300, 0xecd05bb9ff29cc0c, 0x1000000000000, 0xecd05bb9ff29cc0c)
	/goroot/src/net/fd_unix.go:238 +0x42
net.(*TCPListener).accept(0xc00000e440, 0x5e7959ba, 0xc00019ae28, 0x4bca86)
	/goroot/src/net/tcpsock_posix.go:139 +0x32
net.(*TCPListener).Accept(0xc00000e440, 0xc00019ae78, 0x18, 0xc000001680, 0x6caf0c)
	/goroot/src/net/tcpsock.go:261 +0x64
net/http.(*Server).Serve(0xc000194000, 0x923ee0, 0xc00000e440, 0x0, 0x0)
	/goroot/src/net/http/server.go:2901 +0x25d
net/http.Serve(0x923ee0, 0xc00000e440, 0x91bc00, 0xbf6e60, 0x0, 0x0)
	/goroot/src/net/http/server.go:2468 +0x6e
created by main.main
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:50 +0x3b7

goroutine 22 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075ca8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00029c118, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00029c100, 0xc0002b6000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00029c100, 0xc0002b6000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0020, 0xc0002b6000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002a8000, 0xc0002b6000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d3a6318, 0xc000298d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000290480, 0xc00024cf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc0001262a0, 0xc00024cf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000138400, 0xc00024cf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000138400, 0xc00024cf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000138440, 0xc00024cf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc0000a2f30, 0x91bc60, 0xc000138440, 0x91bc60, 0x2, 0xc0002860f0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000138440, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000174090)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 20 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc000226480)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 21 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc000226480)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 9 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075d88, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a098, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a080, 0xc0000caeb1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a080, 0xc0000caeb1, 0x1, 0x1, 0x6f9080, 0xc00010e0c0, 0xc00008a780)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186018, 0xc0000caeb1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0000caea0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 50 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075e68, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc598, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cc580, 0xc00017a000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cc580, 0xc00017a000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000134030, 0xc00017a000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc000226480, 0xc00017a000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d3a8738, 0xc000298518)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000136420, 0xc000304f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc000284020, 0xc000304f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc00028c000, 0xc000304f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc00028c000, 0xc000304f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc00028c040, 0xc000304f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc00009ef30, 0x91bc60, 0xc00028c040, 0x91bc60, 0x0, 0x0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc00028c040, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000280000)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 56 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075a08, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a198, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a180, 0xc0000cb0f1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a180, 0xc0000cb0f1, 0x1, 0x1, 0x100000000000000, 0x1000000000001, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186020, 0xc0000cb0f1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0000cb0e0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 54 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002a8000)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 55 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002a8000)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 10 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0001940e0, 0xc000038500)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0001940e0, 0xc000038500)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0001940e0, 0xc000038500)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0001940e0, 0xc000038500)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0000dedc0, 0x924de0, 0xc000020900)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 35 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075bc8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc698, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cc680, 0xc0000cb091, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cc680, 0xc0000cb091, 0x1, 0x1, 0xc000090768, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0000100b0, 0xc0000cb091, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0000cb080)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 28 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075848, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc718, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cc700, 0xc0001826d1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cc700, 0xc0001826d1, 0x1, 0x1, 0x100000000000f87, 0x1000000000001, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0000100b8, 0xc0001826d1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0001826c0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 26 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0001447e0)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 27 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0001447e0)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 36 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0002b8000, 0xc000038600)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0002b8000, 0xc000038600)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002b8000, 0xc000038600)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002b8000, 0xc000038600)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0001a60a0, 0x924de0, 0xc0000209c0)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 37 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075ae8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000308118, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000308100, 0xc000303000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000308100, 0xc000303000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000134050, 0xc000303000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0001447e0, 0xc000303000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224035888, 0xc000294d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000136960, 0xc0002c2f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00018e0a0, 0xc0002c2f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000190140, 0xc0002c2f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000190140, 0xc0002c2f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000190180, 0xc0002c2f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc0000a3f30, 0x91bc60, 0xc000190180, 0x91bc60, 0x0, 0x0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000190180, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0001ac090)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 59 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075768, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc898, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cc880, 0xc000250000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cc880, 0xc000250000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0000100d0, 0xc000250000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002265a0, 0xc000250000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d23e3a0, 0xc000295518)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc0000b7140, 0xc0001ccf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc000284160, 0xc0001ccf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc00028c1c0, 0xc0001ccf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc00028c1c0, 0xc0001ccf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc00028c200, 0xc0001ccf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc0001c5f30, 0x91bc60, 0xc00028c200, 0x91bc60, 0x2, 0xc000182570)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc00028c200, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0002801b0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 57 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0001b6000)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 58 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0001b6000)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 11 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc000314000, 0xc0001b2100)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc000314000, 0xc0001b2100)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc000314000, 0xc0001b2100)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc000314000, 0xc0001b2100)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0000dee60, 0x924de0, 0xc000190200)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 12 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075928, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a318, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a300, 0xc0002c4000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a300, 0xc0002c4000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0028, 0xc0002c4000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0001b6000, 0xc0002c4000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d23e5c0, 0xc000298d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000290540, 0xc0001bef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00000e5a0, 0xc0001bef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000020a80, 0xc0001bef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000020a80, 0xc0001bef87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000020ac0, 0xc0001bef87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc00019df30, 0x91bc60, 0xc000020ac0, 0x91bc60, 0x2, 0xc0002860f0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000020ac0, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0002301b0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 67 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075688, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a398, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a380, 0xc000182791, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a380, 0xc000182791, 0x1, 0x1, 0x100000000000000, 0x1000000000001, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186040, 0xc000182791, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000182780)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 16 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002265a0)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 66 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002265a0)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 41 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0002420e0, 0xc0001b2200)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0002420e0, 0xc0001b2200)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002420e0, 0xc0001b2200)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002420e0, 0xc0001b2200)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0001a6140, 0x924de0, 0xc0001902c0)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 29 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002a8120)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 30 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002a8120)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 68 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0002b80e0, 0xc00029e200)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0002b80e0, 0xc00029e200)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002b80e0, 0xc00029e200)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002b80e0, 0xc00029e200)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0000defa0, 0x924de0, 0xc00028c280)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 63 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f52240754c8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cc998, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cc980, 0xc000286671, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cc980, 0xc000286671, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0000100d8, 0xc000286671, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000286660)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 69 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f52240755a8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00029c318, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00029c300, 0xc000318000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00029c300, 0xc000318000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000134058, 0xc000318000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002a8120, 0xc000318000, 0x1000, 0x1000, 0xc20dd8, 0x7f52240354d0, 0xc000299518)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000136a20, 0xc0002d0f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00000e6a0, 0xc0002d0f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000020c00, 0xc0002d0f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000020c00, 0xc0002d0f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000020c40, 0xc0002d0f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc000199f30, 0x91bc60, 0xc000020c40, 0x91bc60, 0x2, 0xc000286510)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000020c40, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000230360)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 98 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f52240753e8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000ccb18, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000ccb00, 0xc0002d2000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000ccb00, 0xc0002d2000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0040, 0xc0002d2000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002266c0, 0xc0002d2000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224037b98, 0xc000258d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc0002909c0, 0xc0003b8f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc000284260, 0xc0003b8f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc00028c400, 0xc0003b8f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc00028c400, 0xc0003b8f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc00028c440, 0xc0003b8f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc000390f30, 0x91bc60, 0xc00028c440, 0x91bc60, 0x2, 0xc0000cb680)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc00028c440, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000280360)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 64 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002266c0)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 65 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002266c0)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 82 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0003aa000, 0xc00039e000)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0003aa000, 0xc00039e000)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0003aa000, 0xc00039e000)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0003aa000, 0xc00039e000)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc000388000, 0x924de0, 0xc000396000)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 83 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075308, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382018, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382000, 0xc0003840a1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382000, 0xc0003840a1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000386000, 0xc0003840a1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000384090)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 104 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0002b81c0, 0xc00029e400)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0002b81c0, 0xc00029e400)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002b81c0, 0xc00029e400)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002b81c0, 0xc00029e400)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0002c81e0, 0x924de0, 0xc00028c500)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 102 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002a8240)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 103 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002a8240)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 84 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0002421c0, 0xc0001b2300)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0002421c0, 0xc0001b2300)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002421c0, 0xc0001b2300)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002421c0, 0xc0001b2300)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0003880a0, 0x924de0, 0xc000190380)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 73 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075148, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382118, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382100, 0xc000182821, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382100, 0xc000182821, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000386008, 0xc000182821, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000182810)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 42 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075228, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00029c518, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00029c500, 0xc0002d4000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00029c500, 0xc0002d4000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0058, 0xc0002d4000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002a8240, 0xc0002d4000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224077ab8, 0xc000258d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000290e40, 0xc00026ef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00018e1a0, 0xc00026ef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000190440, 0xc00026ef87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000190440, 0xc00026ef87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000190480, 0xc00026ef87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc0001c7f30, 0x91bc60, 0xc000190480, 0x91bc60, 0x2, 0xc0000cb680)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000190480, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0001ac1b0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 48 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039d48, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382198, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382180, 0xc0003842e1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382180, 0xc0003842e1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000386010, 0xc0003842e1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0003842d0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 46 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0001b6240)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 47 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0001b6240)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 105 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039f08, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00029c598, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00029c580, 0xc000286b51, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00029c580, 0xc000286b51, 0x1, 0x1, 0x100000000000f87, 0x1000000010000, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0060, 0xc000286b51, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000286b40)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 74 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224075068, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a518, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a500, 0xc0001cf000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a500, 0xc0001cf000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186058, 0xc0001cf000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0001b6240, 0xc0001cf000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224037a88, 0xc000295518)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000180b40, 0xc0003bcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00000e7a0, 0xc0003bcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000020d40, 0xc0003bcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000020d40, 0xc0003bcf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000020d80, 0xc0003bcf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc000391f30, 0x91bc60, 0xc000020d80, 0x91bc60, 0x2, 0xc000182570)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000020d80, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000230510)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 80 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039e28, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000ccd98, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000ccd80, 0xc00026d000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000ccd80, 0xc00026d000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000010100, 0xc00026d000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0002267e0, 0xc00026d000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d23e0f8, 0xc000258d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc0000b7b60, 0xc0001dcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00000e880, 0xc0001dcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000020e40, 0xc0001dcf87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000020e40, 0xc0001dcf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000020e80, 0xc0001dcf87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc000392f30, 0x91bc60, 0xc000020e80, 0x91bc60, 0x2, 0xc0000cb9b0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000020e80, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc000230630)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 78 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0002267e0)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 79 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0002267e0)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 85 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL1Handler(0x9241a0, 0xc0001941c0, 0xc00039e100)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:43 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ee8, 0x9241a0, 0xc0001941c0, 0xc00039e100)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0001941c0, 0xc00039e100)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0001941c0, 0xc00039e100)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc000388140, 0x924de0, 0xc000396140)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 134 [syscall, 2 minutes]:
syscall.Syscall(0x23, 0xc000259fb8, 0xc000259fa8, 0x0, 0xc0001aab40, 0xc000259fa8, 0x1)
	/goroot/src/syscall/asm_linux_amd64.s:18 +0x5
golang.org/x/sys/unix.Nanosleep(0xc000259fb8, 0xc000259fa8, 0x0, 0x1)
	/gopath/pkg/mod/golang.org/x/sys@v0.0.0-20200223170610-d5e6a3e2c0ae/unix/zsyscall_linux_amd64.go:1160 +0x5f
main.sysHang(...)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main_unix.go:12
main.main.func1(0xc0001aab40)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:65 +0x71
created by main.main
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:63 +0x548

goroutine 49 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc000226900)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 130 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc000226900)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 86 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL2Handler(0x9241a0, 0xc0002422a0, 0xc000038b00)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:54 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ef0, 0x9241a0, 0xc0002422a0, 0xc000038b00)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002422a0, 0xc000038b00)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002422a0, 0xc000038b00)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0003881e0, 0x924de0, 0xc000020f00)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 116 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039b88, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382218, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382200, 0xc0000cbdb1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382200, 0xc0000cbdb1, 0x1, 0x1, 0xf87, 0x1e00, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000386018, 0xc0000cbdb1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc0000cbda0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 87 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039c68, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000ccf18, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000ccf00, 0xc0001db000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000ccf00, 0xc0001db000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186060, 0xc0001db000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc000226900, 0xc0001db000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224077700, 0xc000254d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000180c00, 0xc000402f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc0003ae060, 0xc000402f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000396200, 0xc000402f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000396200, 0xc000402f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000396240, 0xc000402f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc000276f30, 0x91bc60, 0xc000396240, 0x91bc60, 0x0, 0x0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000396240, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0003a6090)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 119 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL2Handler(0x9241a0, 0xc0002b8380, 0xc00029e500)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:54 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ef0, 0x9241a0, 0xc0002b8380, 0xc00029e500)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002b8380, 0xc00029e500)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002b8380, 0xc00029e500)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0000df180, 0x924de0, 0xc00028c6c0)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 117 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0003c4000)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 118 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0003c4000)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 106 [chan receive, 2 minutes]:
github.com/maruel/panicparse/cmd/panicweb/internal.URL2Handler(0x9241a0, 0xc0002b82a0, 0xc0001b2500)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:54 +0x229
net/http.HandlerFunc.ServeHTTP(0x8a3ef0, 0x9241a0, 0xc0002b82a0, 0xc0001b2500)
	/goroot/src/net/http/server.go:2012 +0x44
net/http.(*ServeMux).ServeHTTP(0xbf6e60, 0x9241a0, 0xc0002b82a0, 0xc0001b2500)
	/goroot/src/net/http/server.go:2387 +0x1a5
net/http.serverHandler.ServeHTTP(0xc000194000, 0x9241a0, 0xc0002b82a0, 0xc0001b2500)
	/goroot/src/net/http/server.go:2807 +0xa3
net/http.(*conn).serve(0xc0002c83c0, 0x924de0, 0xc000190600)
	/goroot/src/net/http/server.go:1895 +0x86c
created by net/http.(*Server).Serve
	/goroot/src/net/http/server.go:2933 +0x35c

goroutine 107 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f52240399c8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00029c698, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00029c680, 0xc000182cd1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00029c680, 0xc000182cd1, 0x1, 0x1, 0xf87, 0x1e00, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc0002a0068, 0xc000182cd1, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000182cc0)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 91 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039aa8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382398, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382380, 0xc000401000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382380, 0xc000401000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000010118, 0xc000401000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0003c4000, 0xc000401000, 0x1000, 0x1000, 0xc20dd8, 0x7f5224035118, 0xc000255d18)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc00027e0c0, 0xc0002e4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc0003ae160, 0xc0002e4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc0003962c0, 0xc0002e4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc0003962c0, 0xc0002e4f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000396300, 0xc0002e4f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc00038ef30, 0x91bc60, 0xc000396300, 0x91bc60, 0x0, 0x0)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000396300, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0003a61b0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 133 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f52240398e8, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc000382518, 0x72, 0x1000, 0x1000, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc000382500, 0xc0001e1000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc000382500, 0xc0001e1000, 0x1000, 0x1000, 0x400, 0x203000, 0x400)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186068, 0xc0001e1000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*persistConn).Read(0xc0003c4120, 0xc0001e1000, 0x1000, 0x1000, 0xc20dd8, 0x7f524d23fe30, 0xc000256518)
	/goroot/src/net/http/transport.go:1825 +0x75
bufio.(*Reader).Read(0xc000180d20, 0xc0002f4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/bufio/bufio.go:226 +0x24f
io.(*LimitedReader).Read(0xc00018e2a0, 0xc0002f4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/io/io.go:451 +0x63
net/http.(*body).readLocked(0xc000190700, 0xc0002f4f87, 0xe79, 0xe79, 0x187, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:847 +0x5f
net/http.(*body).Read(0xc000190700, 0xc0002f4f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transfer.go:839 +0xf2
net/http.(*bodyEOFSignal).Read(0xc000190740, 0xc0002f4f87, 0xe79, 0xe79, 0x0, 0x0, 0x0)
	/goroot/src/net/http/transport.go:2649 +0xde
bytes.(*Buffer).ReadFrom(0xc0002eaf30, 0x91bc60, 0xc000190740, 0x91bc60, 0x2, 0xc000384660)
	/goroot/src/bytes/buffer.go:204 +0xb1
io/ioutil.readAll(0x91bc60, 0xc000190740, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0)
	/goroot/src/io/ioutil/ioutil.go:36 +0xe3
io/ioutil.ReadAll(...)
	/goroot/src/io/ioutil/ioutil.go:45
github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync.func1(0xc0001ac3f0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:26 +0x69
created by github.com/maruel/panicparse/cmd/panicweb/internal.GetAsync
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/internal/internal.go:25 +0x79

goroutine 131 [select, 2 minutes]:
net/http.(*persistConn).readLoop(0xc0003c4120)
	/goroot/src/net/http/transport.go:2099 +0x99e
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1647 +0xc56

goroutine 132 [select, 2 minutes]:
net/http.(*persistConn).writeLoop(0xc0003c4120)
	/goroot/src/net/http/transport.go:2277 +0x11c
created by net/http.(*Transport).dialConn
	/goroot/src/net/http/transport.go:1648 +0xc7b

goroutine 108 [IO wait, 2 minutes]:
internal/poll.runtime_pollWait(0x7f5224039808, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc0000cd018, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc0000cd000, 0xc000286e21, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc0000cd000, 0xc000286e21, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000010120, 0xc000286e21, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000286e10)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0

goroutine 95 [chan receive, 2 minutes, locked to thread]:
main.(*writeHang).Write(0xc000398180, 0xc0003940bf, 0x1, 0x1, 0x1000000010000, 0xc0003a4360, 0x912ec0)
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:92 +0x58
github.com/mattn/go-colorable.(*NonColorable).Write(0xc000398190, 0xc0003940b8, 0x7, 0x7, 0x15, 0xc000024840, 0x11)
	/gopath/pkg/mod/github.com/mattn/go-colorable@v0.1.6/noncolorable.go:30 +0x2ae
created by main.main
	/gopath/src/github.com/maruel/panicparse/cmd/panicweb/main.go:73 +0x68c

goroutine 226 [IO wait]:
internal/poll.runtime_pollWait(0x7f5224039728, 0x72, 0xffffffffffffffff)
	/goroot/src/runtime/netpoll.go:203 +0x55
internal/poll.(*pollDesc).wait(0xc00018a618, 0x72, 0x0, 0x1, 0xffffffffffffffff)
	/goroot/src/internal/poll/fd_poll_runtime.go:87 +0x45
internal/poll.(*pollDesc).waitRead(...)
	/goroot/src/internal/poll/fd_poll_runtime.go:92
internal/poll.(*FD).Read(0xc00018a600, 0xc000182e81, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/internal/poll/fd_unix.go:169 +0x19b
net.(*netFD).Read(0xc00018a600, 0xc000182e81, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/fd_unix.go:202 +0x4f
net.(*conn).Read(0xc000186070, 0xc000182e81, 0x1, 0x1, 0x0, 0x0, 0x0)
	/goroot/src/net/net.go:184 +0x8e
net/http.(*connReader).backgroundRead(0xc000182e70)
	/goroot/src/net/http/server.go:678 +0x58
created by net/http.(*connReader).startBackgroundRead
	/goroot/src/net/http/server.go:674 +0xd0
`
