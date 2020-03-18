<!DOCTYPE html>

{{- define "RenderCall" -}}
{{- /* TODO(maruel): Add link when possible or full path */ -}}
{{- /* TODO(maruel): Align horizontally SrcList when used in the stack. */ -}}
{{- /* TODO(maruel): Process Args properly. */ -}}
<span class="call">{{.SrcName}}:{{.Line}} <span class="{{funcClass .}}">{{.Func.Name}}</span>({{.Args}})</span>
{{- end -}}

<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>PanicParse</title>
<link rel="shortcut icon" type="image/png" href="data:image/png;base64,{{.Favicon}}"/>
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
  }
  li {
    margin-left: 2.5em;
  }

  {{- /* Highlights */ -}}
  .FuncStdLibExported {
    color: #7CFC00;
  }
  .FuncStdLib {
    color: #008000;
  }
  .FuncMain {
    color: #C0C000;
  }
  .FuncOtherExported {
    color: #FF0000;
  }
  .FuncOther {
    color: #A00000;
  }
  .RoutineFirst {
  }
  .Routine {
  }
  .call {
    font-family: monospace;
  }
</style>
<div id="legend">Generated on {{.Now.String}}.
  {{- if .NeedsEnv -}}
    <br>To see all goroutines, visit <a
    href=https://github.com/maruel/panicparse#gotraceback>github.com/maruel/panicparse</a>.<br>
  {{- end -}}
  {{- /* TODO(maruel): Add more details, Go version, variables, things that can
  be retrieved quickly. */ -}}
</div>
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
    <h2>Stack</h2>
    <ol>
      {{range $e.Signature.Stack.Calls}}
        <li>{{template "RenderCall" .}}</li>
      {{- end -}}
      {{- if $e.Stack.Elided}}<li>(...)</li>{{end -}}
    </ol>
  {{- end -}}
</div>
