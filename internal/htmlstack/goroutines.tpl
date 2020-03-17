<!DOCTYPE html>

{{- define "RenderCall" -}}
{{.SrcLine}} <span class="{{funcClass .}}">{{.Func.Name}}</span>({{.Args}})
{{- end -}}

<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>PanicParse</title>
<link rel="shortcut icon" type="image/png" href="data:image/png;base64,{{.Favicon}}"/>
<style>
	body {
		background: black;
		color: lightgray;
	}
	body, pre {
		font-family: Menlo, monospace;
		font-weight: bold;
	}
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
</style>
<div id="legend">Generated on {{.Now.String}}.
{{if .NeedsEnv}}
<br>To see all goroutines, visit <a
href=https://github.com/maruel/panicparse#gotraceback>github.com/maruel/panicparse</a>.<br>
{{end}}
</div>
<div id="content">
{{range .Buckets}}
	<h1>{{if .First}}Panicking {{end}}Routine</h1>
	<span class="{{routineClass .}}">{{len .IDs}}: <span class="state">{{.State}}</span>
	{{if .SleepMax -}}
	  {{- if ne .SleepMin .SleepMax}} <span class="sleep">[{{.SleepMin}}~{{.SleepMax}} minutes]</span>
		{{- else}} <span class="sleep">[{{.SleepMax}} minutes]</span>
		{{- end -}}
	{{- end}}
	{{if .Locked}} <span class="locked">[locked]</span>
	{{- end -}}
	{{- if .CreatedBy.SrcPath}} <span class="created">[Created by {{template "RenderCall" .CreatedBy}}]</span>
	{{- end -}}
	<h2>Stack</h2>
	{{range .Signature.Stack.Calls}}
	- {{template "RenderCall" .}}<br>
	{{- end}}
	{{if .Stack.Elided}}(...)<br>{{end}}
{{end}}
</div>
