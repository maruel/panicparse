// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"path/filepath"
	"testing"
)

func TestCallPkg1(t *testing.T) {
	c := Call{
		SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:       153,
		Func:       Function{"gopkg.in/yaml%2ev2.handleErr"},
		Args:       Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
	compareString(t, "yaml.go", c.SourceName())
	compareString(t, filepath.Join("yaml.v2", "yaml.go"), c.PkgSource())
	compareString(t, "gopkg.in/yaml.v2.handleErr", c.Func.String())
	compareString(t, "handleErr", c.Func.Name())
	// This is due to directory name not matching the package name.
	compareString(t, "yaml.v2", c.Func.PkgName())
	compareBool(t, false, c.Func.IsExported())
	compareBool(t, false, c.IsStdlib)
	compareBool(t, false, c.IsPkgMain())
}

func TestCallPkg2(t *testing.T) {
	c := Call{
		SourcePath: "/gopath/src/gopkg.in/yaml.v2/yaml.go",
		Line:       153,
		Func:       Function{"gopkg.in/yaml%2ev2.(*decoder).unmarshal"},
		Args:       Args{Values: []Arg{{Value: 0xc208033b20}}},
	}
	compareString(t, "yaml.go", c.SourceName())
	compareString(t, filepath.Join("yaml.v2", "yaml.go"), c.PkgSource())
	// TODO(maruel): Using '/' for this function is inconsistent on Windows
	// w.r.t. other functions.
	compareString(t, "gopkg.in/yaml.v2.(*decoder).unmarshal", c.Func.String())
	compareString(t, "(*decoder).unmarshal", c.Func.Name())
	// This is due to directory name not matching the package name.
	compareString(t, "yaml.v2", c.Func.PkgName())
	compareBool(t, false, c.Func.IsExported())
	compareBool(t, false, c.IsStdlib)
	compareBool(t, false, c.IsPkgMain())
}

func TestCallStdlib(t *testing.T) {
	c := Call{
		SourcePath: "/goroot/src/reflect/value.go",
		Line:       2125,
		Func:       Function{"reflect.Value.assignTo"},
		Args:       Args{Values: []Arg{{Value: 0x570860}, {Value: 0xc20803f3e0}, {Value: 0x15}}},
	}
	c.updateLocations("/goroot", "/goroot", nil)
	compareString(t, "value.go", c.SourceName())
	compareString(t, "value.go:2125", c.SourceLine())
	compareString(t, filepath.Join("reflect", "value.go"), c.PkgSource())
	compareString(t, "reflect.Value.assignTo", c.Func.String())
	compareString(t, "Value.assignTo", c.Func.Name())
	compareString(t, "reflect", c.Func.PkgName())
	compareBool(t, false, c.Func.IsExported())
	compareBool(t, true, c.IsStdlib)
	compareBool(t, false, c.IsPkgMain())
}

func TestCallMain(t *testing.T) {
	c := Call{
		SourcePath: "/gopath/src/github.com/maruel/panicparse/cmd/pp/main.go",
		Line:       428,
		Func:       Function{"main.main"},
	}
	compareString(t, "main.go", c.SourceName())
	compareString(t, "main.go:428", c.SourceLine())
	compareString(t, filepath.Join("pp", "main.go"), c.PkgSource())
	compareString(t, "main.main", c.Func.String())
	compareString(t, "main", c.Func.Name())
	compareString(t, "main", c.Func.PkgName())
	compareBool(t, true, c.Func.IsExported())
	compareBool(t, false, c.IsStdlib)
	compareBool(t, true, c.IsPkgMain())
}

func TestCallC(t *testing.T) {
	c := Call{
		SourcePath: "/goroot/src/runtime/proc.c",
		Line:       1472,
		Func:       Function{"findrunnable"},
		Args:       Args{Values: []Arg{{Value: 0xc208012000}}},
	}
	c.updateLocations("/goroot", "/goroot", nil)
	compareString(t, "proc.c", c.SourceName())
	compareString(t, "proc.c:1472", c.SourceLine())
	compareString(t, filepath.Join("runtime", "proc.c"), c.PkgSource())
	compareString(t, "findrunnable", c.Func.String())
	compareString(t, "findrunnable", c.Func.Name())
	compareString(t, "", c.Func.PkgName())
	compareBool(t, false, c.Func.IsExported())
	compareBool(t, true, c.IsStdlib)
	compareBool(t, false, c.IsPkgMain())
}

func TestArgs(t *testing.T) {
	a := Args{
		Values: []Arg{
			{Value: 0x4},
			{Value: 0x7fff671c7118},
			{Value: 0xffffffff00000080},
			{},
			{Value: 0xffffffff0028c1be},
			{},
			{},
			{},
			{},
			{},
		},
		Elided: true,
	}
	compareString(t, "0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...", a.String())
}

func TestFunctionAnonymous(t *testing.T) {
	f := Function{"main.func·001"}
	compareString(t, "main.func·001", f.String())
	compareString(t, "main.func·001", f.PkgDotName())
	compareString(t, "func·001", f.Name())
	compareString(t, "main", f.PkgName())
	compareBool(t, false, f.IsExported())
}

func TestFunctionGC(t *testing.T) {
	f := Function{"gc"}
	compareString(t, "gc", f.String())
	compareString(t, "gc", f.PkgDotName())
	compareString(t, "gc", f.Name())
	compareString(t, "", f.PkgName())
	compareBool(t, false, f.IsExported())
}

//

func compareBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Fatalf("%t != %t", expected, actual)
	}
}
