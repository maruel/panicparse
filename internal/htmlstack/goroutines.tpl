<!DOCTYPE html>

{{- /* Accepts a Args */ -}}
{{- define "RenderArgs" -}}
  <span class="args"><span>
  {{- $l := len .Values -}}
  {{- $last := minus $l 1 -}}
  {{- $elided := .Elided -}}
  {{- range $i, $e := .Values -}}
    {{- $e.String -}}
    {{- $isNotLast := ne $i $last -}}
    {{- if or $elided $isNotLast}}, {{end -}}
  {{- end -}}
  {{- if $elided}}…{{end -}}
  </span></span>
{{- end -}}

{{- /* Accepts a Call */ -}}
{{- define "RenderCall" -}}
  <span class="call"><a href="{{srcURL .}}">{{.SrcName}}:{{.Line}}</a> <span class="{{funcClass .}}">
  <a href="{{pkgURL .}}">{{.Func.PkgName}}.{{.Func.Name}}</a></span>({{template "RenderArgs" .Args}})</span>
  {{- if isDebug -}}
  <br>SrcPath: {{.SrcPath}}
  <br>LocalSrcPath: {{.LocalSrcPath}}
  <br>Func: {{.Func.Raw}}
  <br>IsStdlib: {{.IsStdlib}}
  {{- end -}}
{{- end -}}

{{- /* Accepts a Stack */ -}}
{{- define "RenderCalls" -}}
  <table class="stack">
    {{- range $i, $e := .Calls -}}
      <tr>
        <td>{{$i}}</td>
        <td>
          <a href="{{pkgURL $e}}">{{$e.Func.PkgName}}</a>
        </td>
        <td>
          <a href="{{srcURL $e}}">{{$e.SrcName}}:{{$e.Line}}</a>
        </td>
        <td>
          <span class="{{funcClass $e}}"><a href="{{pkgURL $e}}">{{$e.Func.Name}}</a></span>({{template "RenderArgs" $e.Args}})
        </td>
      </tr>
    {{- end -}}
    {{- if .Elided}}<tr><td>(…)</td><tr>{{end -}}
  </table>
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
  table.stack {
    margin: 0.6em;
  }
  table.stack tr:hover {
    background-color: #DDD;
  }
  table.stack td {
    font-family: monospace;
    padding: 0.2em 0.4em 0.2em;
  }
  .call {
    font-family: monospace;
  }
  @media screen and (max-width: 500px) {
    h1 {
      font-size: 1.3em;
    }
  }
  @media screen and (max-width: 500px) and (orientation: portrait) {
    .args span {
      display: none;
    }
    .args::after {
      content: '…';
    }
  }
  .created {
    white-space: nowrap;
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
</style>
<div id="content">
  {{- range $i, $e := .Buckets -}}
    {{$l := len $e.IDs}}
    <h1>Signature #{{$i}}: <span class="{{routineClass $e}}">{{$l}} routine{{if ne 1 $l}}s{{end}}: <span class="state">{{$e.State}}</span>
    {{- if $e.SleepMax -}}
      {{- if ne $e.SleepMin $e.SleepMax}} <span class="sleep">[{{$e.SleepMin}}~{{$e.SleepMax}} mins]</span>
      {{- else}} <span class="sleep">[{{$e.SleepMax}} mins]</span>
      {{- end -}}
    {{- end -}}
    </h1>
    {{if $e.Locked}} <span class="locked">[locked]</span>
    {{- end -}}
    {{- if $e.CreatedBy.Func.Raw}} <span class="created">Created by: {{template "RenderCall" $e.CreatedBy}}</span>
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
