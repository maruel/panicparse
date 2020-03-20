// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate go run regen.go

package htmlstack

import (
	"html/template"
	"io"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/maruel/panicparse/stack"
)

// Write writes buckets as HTML to the writer.
func Write(w io.Writer, buckets []*stack.Bucket, needsEnv bool) error {
	m := template.FuncMap{
		"funcClass": funcClass,
		"minus":     minus,
		"symbol":    symbol,
		// Needs to be a function and not a variable, otherwise it is not
		// accessible inside inner templates.
		"isDebug":    isDebug,
		"getVersion": runtime.Version,
	}
	if len(buckets) > 1 {
		m["routineClass"] = routineClass
	} else {
		m["routineClass"] = func(bucket *stack.Bucket) template.HTML { return "Routine" }
	}
	t, err := template.New("t").Funcs(m).Parse(indexHTML)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Buckets":    buckets,
		"Favicon":    favicon,
		"GOMAXPROCS": runtime.GOMAXPROCS(0),
		"GOPATH":     os.Getenv("GOPATH"),
		"GOROOT":     runtime.GOROOT(),
		"NeedsEnv":   needsEnv,
		"Now":        time.Now().Truncate(time.Second),
	}
	if isDebug() {
	}
	return t.Execute(w, data)
}

//

var reMethodSymbol = regexp.MustCompile(`^\(\*?([^)]+)\)(\..+)$`)

func isDebug() bool {
	// Set to true to log more details in the web page.
	return false
}

func funcClass(line *stack.Call) template.HTML {
	if line.IsStdlib {
		if line.Func.IsExported() {
			return "FuncStdLibExported"
		}
		return "FuncStdLib"
	} else if line.IsPkgMain() {
		return "FuncMain"
	} else if line.Func.IsExported() {
		return "FuncOtherExported"
	}
	return "FuncOther"
}

func minus(i, j int) int {
	return i - j
}

// symbol is the hashtag to use to refer to the symbol in pkg.go.dev
func symbol(f *stack.Func) string {
	i := strings.LastIndexByte(f.Raw, '/')
	if i == -1 {
		return ""
	}
	j := strings.IndexByte(f.Raw[i:], '.')
	if j == -1 {
		return ""
	}
	s, _ := url.QueryUnescape(f.Raw[i+j+1:])
	if reMethodSymbol.MatchString(s) {
		// Transform the method form.
		s = reMethodSymbol.ReplaceAllString(s, "$1$2")
	}
	return s
}

func routineClass(bucket *stack.Bucket) template.HTML {
	if bucket.First {
		return "RoutineFirst"
	}
	return "Routine"
}
