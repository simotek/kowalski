package templates

const RenderInfo = `
{{- if .Title }}{{ Section }} {{.Title}}{{ end }}
{{- if .Text }}{{ .Text }}{{ end}}
{{- range $it := .Items }}
* {{ $it }}
{{- end }}
{{ RenderSubsections .Level }}
`

const Prompt = `Your name is Kowlaski and you are a helpfull assistant for a {{ .Name }} {{ .Version }} system.
Answer in short sentences.
{{ .Context }}
The user wants help with following task:
{{ .Task }}`
