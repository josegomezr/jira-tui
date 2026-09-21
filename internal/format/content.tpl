{{- if .IncludeTitle }}
{{ trim .Issue.Fields.Summary }}

{{- end }}
=== Description ===

{{ jiratomd .Issue.Fields.Description | trim }}

{{- with .Issue.Fields.Attachments }}

=== Attachments [{{ len . }}] ===
{{ range $i, $Attachment := . }}
[{{ printf "%2d" (add 1 $i) }}] {{ $Attachment.Author.DisplayName }} <{{ $Attachment.Author.Name }}> @ {{ $Attachment.Created | toDate "2006-01-02T15:04:05.000-0700" | date "2006-01-02 15:04:05 -0700" }} [{{ $Attachment.Created | toDate "2006-01-02T15:04:05.000-0700" | agoshort }}]
     Filename: {{ $Attachment.Filename }}
     Content : {{ $Attachment.Content }}
{{- end }}
{{- end }}

{{- with .Issue.Fields.Comments }}

=== Comments [{{ len .Comments }}] ===
{{ range $i, $Comment := .Comments }}
{{ printf "[+ %2d]" (add 1 $i) }} {{ $Comment.Author.DisplayName }} <{{ $Comment.Author.Name }}> @ {{ $Comment.Created | toDate "2006-01-02T15:04:05.000-0700" | date "2006-01-02 15:04:05 -0700" }} [{{ $Comment.Created | toDate "2006-01-02T15:04:05.000-0700" | agoshort }}]
{{ jiratomd $Comment.Body | trim }}
{{ printf "[/ %2d]" (add 1 $i) }}
{{""}}
{{- end }}
{{- end }}
