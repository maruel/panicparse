// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maruel/ut"
)

func TestProcess(t *testing.T) {
	// 2 goroutines with the same signature
	data := []string{
		"panic: runtime error: index out of range",
		"",
		"goroutine 11 [running]:",
		"github.com/luci/luci-go/client/archiver.(*archiver).PushFile(0xc208032410, 0xc20968a3c0, 0x5b, 0xc20988c280, 0x7d, 0x0, 0x0)",
		"        /gopath/src/github.com/luci/luci-go/client/archiver/archiver.go:325 +0x2c4",
		"github.com/luci/luci-go/client/isolate.archive(0x7fbdab7a5218, 0xc208032410, 0xc20803b0b0, 0x22, 0xc208046370, 0xc20804666a, 0x17, 0x0, 0x0, 0x0, ...)",
		"        /gopath/src/github.com/luci/luci-go/client/isolate/isolate.go:148 +0x12d2",
		"github.com/luci/luci-go/client/isolate.Archive(0x7fbdab7a5218, 0xc208032410, 0xc20803b0b0, 0x22, 0xc208046370, 0x0, 0x0)",
		"        /gopath/src/github.com/luci/luci-go/client/isolate/isolate.go:102 +0xc9",
		"main.func·004(0x7fffc3b8f13a, 0x2c)",
		"        /gopath/src/github.com/luci/luci-go/client/cmd/isolate/batch_archive.go:166 +0x7cd",
		"created by main.(*batchArchiveRun).main",
		"        /gopath/src/github.com/luci/luci-go/client/cmd/isolate/batch_archive.go:167 +0x42c",
		"",
	}
	out := &bytes.Buffer{}
	err := Process(bytes.NewBufferString(strings.Join(data, "\n")), out)
	ut.AssertEqual(t, nil, err)
	expected := []string{
		"panic: runtime error: index out of range",
		"",
		"\x1b[37m1: running\x1b[90m [Created by main.(*batchArchiveRun).main @ batch_archive.go:167]\x1b[0m",
		"    \x1b[97marchiver\x1b[0m archiver.go:325      \x1b[91m(*archiver).PushFile\x1b[0m(#1, 0xc20968a3c0, 0x5b, 0xc20988c280, 0x7d, 0x0, 0x0)",
		"    \x1b[97misolate \x1b[0m isolate.go:148       \x1b[31marchive\x1b[0m(#4, #1, #2, 0x22, #3, 0xc20804666a, 0x17, 0x0, 0x0, 0x0, ...)",
		"    \x1b[97misolate \x1b[0m isolate.go:102       \x1b[91mArchive\x1b[0m(#4, #1, #2, 0x22, #3, 0x0, 0x0)",
		"    \x1b[97mmain    \x1b[0m batch_archive.go:166 \x1b[93mfunc·004\x1b[0m(0x7fffc3b8f13a, 0x2c)",
		"",
	}
	ut.AssertEqual(t, strings.Join(expected, "\n"), out.String())
}
