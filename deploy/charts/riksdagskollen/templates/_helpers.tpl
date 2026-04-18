{{/*
Expand the name of the chart.
*/}}
{{- define "riksdagskollen.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "riksdagskollen.fullname" -}}
{{- printf "%s" (include "riksdagskollen.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "riksdagskollen.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "riksdagskollen.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "riksdagskollen.fullname" . }}-backend
app.kubernetes.io/component: backend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "riksdagskollen.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "riksdagskollen.fullname" . }}-frontend
app.kubernetes.io/component: frontend
{{- end }}
