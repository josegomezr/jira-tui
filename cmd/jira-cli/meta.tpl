=== Meta ===
Created: {{ .Fields.Created | fromjiratime | date "2006-01-02 15:04:05 -0700" }} [{{ .Fields.Created | fromjiratime | agoshort }}]
Updated: {{ .Fields.Updated | fromjiratime | date "2006-01-02 15:04:05 -0700" }} [{{ .Fields.Updated | fromjiratime | agoshort }}]
{{- with .Fields.Creator }}
Created by: {{ .DisplayName }} <{{ .Name }}>
{{- end }}
{{- with .Fields.Components }}
Components: {{ range . }}{{ .Name }}{{ end }}
{{- end }}
Priority: {{ .Fields.Priority.Name }}, Status: {{ .Fields.Status.Name }}
{{- with .Fields.Assignee }}
Assigned to: {{ .DisplayName }} <{{ .Name }}>
{{- end }}

{{- with .Fields.IssueLinks }}

=== Issue links [{{ len . }}] ===
{{- range $i, $IssueLink := . }}
- {{ $IssueLink.Type.Name }}
{{- with $IssueLink.InwardIssue }} <- {{ .Key }}: {{ .Fields.Summary }} [{{ .Fields.Status.Name }}]
{{- end }}
{{- with $IssueLink.OutwardIssue }} -> {{ .Key }}: {{ .Fields.Summary }} [{{ .Fields.Status.Name }}]
{{- end }}
{{""}}
{{- end }}
{{- end }}
