<!DOCTYPE html>

{{- /* Join a list */ -}}
{{- define "Join" -}}
  {{- if . -}}
    {{- $l := len . -}}
    {{- $last := minus $l 1 -}}
    {{- range $i, $e := . -}}
      {{- $e -}}
      {{- $isNotLast := ne $i $last -}}
      {{- if $isNotLast}}, {{end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- /* Accepts a Args */ -}}
{{- define "RenderArgs" -}}
  <span class="args"><span>
  {{- $elided := .Elided -}}
  {{- if .Processed -}}
    {{- $l := len .Processed -}}
    {{- $last := minus $l 1 -}}
    {{- range $i, $e := .Processed -}}
      {{- $e -}}
      {{- $isNotLast := ne $i $last -}}
      {{- if or $elided $isNotLast}}, {{end -}}
    {{- end -}}
  {{- else -}}
    {{- $l := len .Values -}}
    {{- $last := minus $l 1 -}}
    {{- range $i, $e := .Values -}}
      {{- $e.String -}}
      {{- $isNotLast := ne $i $last -}}
      {{- if or $elided $isNotLast}}, {{end -}}
    {{- end -}}
  {{- end -}}
  {{- if $elided}}…{{end -}}
  </span></span>
{{- end -}}

{{- /* Accepts a Call */ -}}
{{- define "RenderCall" -}}
  <span class="call"><a href="{{srcURL .}}">{{.SrcName}}:{{.Line}}</a> <span class="{{funcClass .}}">
  <a href="{{pkgURL .}}">{{.Func.DirName}}.{{.Func.Name}}</a></span>({{template "RenderArgs" .Args}})</span>
  {{- if isDebug -}}
  <br>SrcPath: {{.SrcPath}}
  <br>LocalSrcPath: {{.LocalSrcPath}}
  <br>Func: {{.Func.Complete}}
  <br>Location: {{.Location}}
  {{- end -}}
{{- end -}}

{{- /* Accepts a Stack */ -}}
{{- define "RenderCalls" -}}
  <table class="stack">
    {{- range $i, $e := .Calls -}}
      <tr>
        <td>{{$i}}</td>
        <td>
          <a href="{{pkgURL $e}}">{{$e.Func.DirName}}</a>
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
  .race {
    font-weight: 700;
    color: #600;
  }
  #content {
    width: 100%;
  }

  {{- /* Highlights based on stack.Location value.
         TODO(maruel): Redo the color selection as part of
         https://github.com/maruel/panicparse/issues/26
      */ -}}
  .FuncMain {
    color: #808000;
  }
  .FuncLocationUnknownExported {
    color: #00B000;
  }
  .FuncLocationUnknown {
    color: #00B000;
  }
  .FuncGoModExported {
    color: #C00000;
  }
  .FuncGoMod {
    color: #800000;
  }
  .FuncGOPATHExported {
    color: #C00000;
  }
  .FuncGOPATH {
    color: #800000;
  }
  .FuncGoPkgExported {
    color: #C00000;
  }
  .FuncGoPkg {
    color: #800000;
  }
  .FuncStdlibExported {
    color: #00B000;
  }
  .FuncStdlib {
    color: #006000;
  }
  {{- /* Highlight on first routine (if any) */ -}}
  .RoutineFirst {
  }
  .Routine {
  }
</style>
<div id="content">
  {{- if .Aggregated -}}
    {{- range $i, $e := .Aggregated.Buckets -}}
      {{$l := len $e.IDs}}
      <h1>Signature #{{$i}}: <span class="{{bucketClass $e}}">{{$l}} routine{{if ne 1 $l}}s{{end}}: <span class="state">{{$e.State}}</span>
      {{- if $e.SleepMax -}}
        {{- if ne $e.SleepMin $e.SleepMax}} <span class="sleep">[{{$e.SleepMin}}~{{$e.SleepMax}} mins]</span>
        {{- else}} <span class="sleep">[{{$e.SleepMax}} mins]</span>
        {{- end -}}
      {{- end -}}
      </h1>
      {{if $e.Locked}} <span class="locked">[locked]</span>
      {{- end -}}
      {{- if $e.CreatedBy.Calls}} <span class="created">Created by: {{template "RenderCall" index $e.CreatedBy.Calls 0}}</span>
      {{- end -}}
      {{template "RenderCalls" $e.Signature.Stack}}
    {{- end -}}
  {{- else -}}
    {{- range $i, $e := .Snapshot.Goroutines -}}
      <h1>Routine {{$e.ID}}: <span class="{{routineClass $e}}">: <span class="state">{{$e.State}}</span>
      {{- if $e.SleepMax -}}
        {{- if ne $e.SleepMin $e.SleepMax}} <span class="sleep">[{{$e.SleepMin}}~{{$e.SleepMax}} mins]</span>
        {{- else}} <span class="sleep">[{{$e.SleepMax}} mins]</span>
        {{- end -}}
      {{- end -}}
      </h1>
      {{if $e.Locked}} <span class="locked">[locked]</span>
      {{- end -}}
      {{if $e.RaceAddr}} <span class="race">Race {{if $e.RaceWrite}}write{{else}}read{{end}} @ {{$e.RaceAddr}}</span><br>
      {{- end -}}
      {{- if $e.CreatedBy.Calls}} <span class="created">Created by: {{template "RenderCall" index $e.CreatedBy.Calls 0}}</span>
      {{- end -}}
      {{template "RenderCalls" $e.Signature.Stack}}
    {{- end -}}
  {{- end -}}
</div>
<p>
<div id="legend">
  Created on {{.Now.String}}:
  <ul>
    <li>{{.Version}}</li>
    {{- if and .Snapshot.LocalGOROOT (ne .Snapshot.RemoteGOROOT .Snapshot.LocalGOROOT) -}}
      <li>GOROOT (remote): {{.Snapshot.RemoteGOROOT}}</li>
      <li>GOROOT (local): {{.Snapshot.LocalGOROOT}}</li>
    {{- else -}}
      <li>GOROOT: {{.Snapshot.RemoteGOROOT}}</li>
    {{- end -}}
    <li>GOPATH: {{template "Join" .Snapshot.LocalGOPATHs}}</li>
    {{- if .Snapshot.LocalGomods -}}
      <li>go modules (local):
        <ul>
        {{- range $path, $import := .Snapshot.LocalGomods -}}
          <li>{{$path}}: {{$import}}</li>
        {{- end -}}
        </ul>
      </li>
    {{- end -}}
    <li>GOMAXPROCS: {{.GOMAXPROCS}}</li>
  </ul>
  {{- .Footer -}}
</div>
