// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate go run regen.go

package htmlstack

import (
	"fmt"
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
		"pkgURL":    pkgURL,
		"srcURL":    srcURL,
		"symbol":    symbol,
		// Needs to be a function and not a variable, otherwise it is not
		// accessible inside inner templates.
		"isDebug": isDebug,
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
		"Version":    runtime.Version(),
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

// pkgURL returns a link to the godoc for the call.
func pkgURL(c *stack.Call) template.URL {
	url := "https://pkg.go.dev/"
	if c.IsStdlib {
		url = "https://golang.org/pkg/"
	}
	if c.Func.IsExported() {
		return template.URL(fmt.Sprintf("%s%s#%s", url, c.ImportPath(), symbol(&c.Func)))
	}
	return template.URL(url + c.ImportPath())
}

// srcURL returns an URL to the sources.
//
// TODO(maruel): Support custom local godoc server as it serves files too.
func srcURL(c *stack.Call) template.URL {
	if c.IsStdlib {
		return template.URL(fmt.Sprintf("https://github.com/golang/go/blob/%s/src/%s#L%d", runtime.Version(), c.RelSrcPath, c.Line))
	}

	// One-off support for github. This will cover a fair share of the URLs, but
	// it'd be nice to support others too. Please submit a PR (including a unit
	// test that I was too lazy to add yet).
	if c.RelSrcPath != "" {
		if parts := strings.SplitN(c.RelSrcPath, "/", 4); len(parts) == 4 && parts[0] == "github.com" {
			// Default to branch master for non-versionned dependencies. It's not
			// optimal but it's better than nothing?
			branch := "master"
			if i := strings.IndexByte(parts[2], '@'); i != -1 {
				// We got a versionned go module.
				branch = parts[2][i+1:]
				parts[2] = parts[2][:i]
			}
			return template.URL(fmt.Sprintf("https://%s/%s/%s/blob/%s/%s#L%d", parts[0], parts[1], parts[2], branch, parts[3], c.Line))
		}

		// TODO(maruel): Add support for golang.org/x/sys.
	}

	if c.LocalSrcPath != "" {
		return template.URL("file:///" + c.LocalSrcPath)
	}
	return template.URL("file:///" + c.SrcPath)
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
