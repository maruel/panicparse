// Copyright 2017 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate go run regen.go

package htmlstack

import (
	"html/template"
	"io"
	"time"

	"github.com/maruel/panicparse/stack"
)

// Write writes buckets as HTML to the writer.
func Write(w io.Writer, buckets []*stack.Bucket, needsEnv bool) error {
	m := template.FuncMap{
		"funcClass": funcClass,
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
		"Buckets":  buckets,
		"Favicon":  favicon,
		"NeedsEnv": needsEnv,
		"Now":      time.Now().Truncate(time.Second),
	}
	return t.Execute(w, data)
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

func routineClass(bucket *stack.Bucket) template.HTML {
	if bucket.First {
		return "RoutineFirst"
	}
	return "Routine"
}
