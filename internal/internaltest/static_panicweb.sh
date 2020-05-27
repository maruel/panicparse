#!/bin/bash
# Copyright 2020 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -eu

eval `go env | grep 'GOROOT\|GOPATH'`
go install github.com/maruel/panicparse/cmd/panicweb
panicweb -port 1212 &
trap "trap - TERM && kill -- -$$" INT TERM EXIT
sleep 1
echo "Sleeping 2 minutes..."
sleep 124
curl -sS 'http://localhost:1212/debug/pprof/goroutine?debug=2' \
	| sed -e "s#\t$GOROOT/#\t/goroot/#g" \
	| sed -e "s#\t$GOPATH/#\t/gopath/#g" > static_panicweb.txt
echo "Copied $(cat static_panicweb.txt | wc -l) lines into static_panicweb.txt."
echo "Add static_panicweb.txt content into static_panicweb.go."
