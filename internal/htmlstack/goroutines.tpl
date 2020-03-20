<!DOCTYPE html>

{{- /* Accepts a Call */ -}}
{{- /*
  TODO(maruel): Use custom local godoc server.
  TODO(maruel): Find a way to link to remote source in a generic way via
  pkg.go.dev.
*/ -}}
{{- define "SrcHostURL" -}}
  {{- if .IsStdlib -}}https://golang.org/{{.RelSrcPath}}#L{{.Line}}{{- else -}}file:///{{.SrcPath}}{{- end -}}
{{- end -}}

{{- /* Accepts a Call */ -}}
{{- define "PkgHostURL" -}}
  {{- if .IsStdlib -}}https://golang.org/pkg/{{- else -}}https://pkg.go.dev/{{- end -}}
{{- end -}}

{{- /* Accepts a Args */ -}}
{{- define "RenderArgs" -}}
  {{- $l := len .Values -}}
  {{- $last := minus $l 1 -}}
  {{- $elided := .Elided -}}
  {{- range $i, $e := .Values -}}
    {{- if ne $e.Name "" -}}
      {{- $e.Name -}}
    {{- else -}}
      {{- printf "0x%08x" $e.Value -}}
    {{- end -}}
    {{- $isNotLast := ne $i $last -}}
    {{- if or $elided $isNotLast -}}, {{end -}}
  {{- end -}}
  {{- if $elided}}...{{end}}
{{- end -}}

{{- /* Accepts a Call */ -}}
{{- define "RenderCall" -}}
  {{- /* TODO(maruel): Add link when possible or full path */ -}}
  {{- /* TODO(maruel): Align horizontally SrcList when used in the stack. */ -}}
  {{- /*
  <span class="call"><a href="{{template "SrcHostURL" .}}">{{.SrcName}}:{{.Line}}</a> <span class="{{funcClass .}}">
  <a href="{{template "PkgHostURL" .}}{{.Func.PkgName}}{{if .Func.IsExported}}#{{symbol .Func}}{{end}}">{{.Func.PkgName}}.{{.Func.Name}}</a></span>({{template "RenderArgs" .Args}})</span>
  */ -}}
  <span class="call">{{.SrcName}}:{{.Line}} <span class="{{funcClass .}}">{{.Func.PkgName}}.{{.Func.Name}}</span>({{template "RenderArgs" .Args}})</span>
  {{- if isDebug -}}
  <br>SrcPath: {{.SrcPath}}
  <br>LocalSrcPath: {{.LocalSrcPath}}
  <br>Func: {{.Func.Raw}}
  <br>IsStdlib: {{.IsStdlib}}
  {{- end -}}
{{- end -}}

{{- /* Accepts a Stack */ -}}
{{- define "RenderCalls" -}}
  <ol>
    {{- range .Calls -}}
      <li>{{template "RenderCall" .}}</li>
    {{- end -}}
    {{- if .Elided}}<li>(...)</li>{{end -}}
  </ol>
{{- end -}}

<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>PanicParse</title>
<link rel="shortcut icon" type="image/gif" href="data:image/gif;base64,{{.Favicon}}"/>
<style>
  {{- /* Minimal CSS reset */ -}}
  * {
    font-family: inherit;
    font-size: 1em;
    margin: 0;
    padding: 0;
  }
  html {
    box-sizing: border-box;
    font-size: 62.5%;
  }
  *, *:before, *:after {
    box-sizing: inherit;
  }
  h1 {
    font-size: 1.5em;
    margin-bottom: 0.2em;
    margin-top: 0.5em;
  }
  h2 {
    font-size: 1.2em;
    margin-bottom: 0.2em;
    margin-top: 0.3em;
  }
  body {
    font-size: 1.6em;
    margin: 2px;
  }
  li {
    margin-left: 2.5em;
  }
  a {
    color: inherit;
    text-decoration: inherit;
  }
  ol, ul {
    margin-bottom: 0.5em;
    margin-top: 0.5em;
  }
  p {
    margin-bottom: 2em;
  }

  {{- /* Highlights */ -}}
  .FuncStdLibExported {
    color: #00B000;
  }
  .FuncStdLib {
    color: #006000;
  }
  .FuncMain {
    color: #808000;
  }
  .FuncOtherExported {
    color: #C00000;
  }
  .FuncOther {
    color: #800000;
  }
  .RoutineFirst {
  }
  .Routine {
  }
  .call {
    font-family: monospace;
  }
</style>
<div id="content">
  {{- range $i, $e := .Buckets -}}
    <h1>Routine {{if .First}}(Panicking){{else}}#{{$i}}{{end}}</h1>
    <span class="{{routineClass $e}}">{{len $e.IDs}}: <span class="state">{{$e.State}}</span>
    {{- if $e.SleepMax -}}
      {{- if ne $e.SleepMin $e.SleepMax}} <span class="sleep">[{{$e.SleepMin}}~{{$e.SleepMax}} minutes]</span>
      {{- else}} <span class="sleep">[{{$e.SleepMax}} minutes]</span>
      {{- end -}}
    {{- end -}}
    {{if $e.Locked}} <span class="locked">[locked]</span>
    {{- end -}}
		{{- /* TODO(maruel): Add link when possible or full path */ -}}
    {{- if $e.CreatedBy.SrcPath}} <span class="created">[Created by {{template "RenderCall" $e.CreatedBy}}]</span>
    {{- end -}}
    {{template "RenderCalls" $e.Signature.Stack}}
  {{- end -}}
</div>
<p>
<div id="legend">
  Created on {{.Now.String}}:
  <ul>
    <li>{{.Version}}</li>
    <li>GOROOT: {{.GOROOT}}</li>
    <li>GOPATH: {{.GOPATH}}</li>
    <li>GOMAXPROCS: {{.GOMAXPROCS}}</li>
    {{- if .NeedsEnv -}}
      <li>To see all goroutines, visit <a
      href=https://github.com/maruel/panicparse#gotraceback>github.com/maruel/panicparse</a></li>
    {{- end -}}
  </ul>
</div>
