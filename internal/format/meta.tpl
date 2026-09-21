=== Meta ===
{{- with .Fields.Creator }}

Reporter: {{ .DisplayName }} <{{ .Name }}>
{{- end }}

Created at: {{ .Fields.Created | fromjiratime | date "2006-01-02 15:04:05 -0700" }} [{{ .Fields.Created | fromjiratime | agoshort }}]

Last updated: {{ .Fields.Updated | fromjiratime | date "2006-01-02 15:04:05 -0700" }} [{{ .Fields.Updated | fromjiratime | agoshort }}]
{{- with .Fields.Components }}

Components: {{ range $idx, $cat := . }}{{ if (eq $idx 0) }}{{else}}, {{end}}{{ $cat.Name }}{{ end }}
{{- end }}

Priority: {{ .Fields.Priority.Name }}, Status: {{ .Fields.Status.Name }}
{{- with .Fields.Assignee }}

Assigned to: {{ .DisplayName }} <{{ .Name }}>
{{- end }}

{{- with .Fields.IssueLinks }}

=== Issue links [{{ len . }}] ===
{{ range $i, $IssueLink := . }}
- {{ $IssueLink.Type.Name }}
{{- with $IssueLink.InwardIssue }} <- {{ .Key }}: {{ .Fields.Summary }} [{{ .Fields.Status.Name }}]
{{- end }}
{{- with $IssueLink.OutwardIssue }} -> {{ .Key }}: {{ .Fields.Summary }} [{{ .Fields.Status.Name }}]
{{- end }}
{{""}}
{{- end }}
{{- end }}
