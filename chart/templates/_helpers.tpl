{{/*
Expand the name of the chart.
*/}}
{{- define "upgrade-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "upgrade-manager.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "upgrade-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "upgrade-manager.labels" -}}
helm.sh/chart: {{ include "upgrade-manager.chart" . }}
{{ include "upgrade-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- range $key, $value := .Values.additionalLabels }}
{{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "upgrade-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "upgrade-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "upgrade-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "upgrade-manager.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the pvct to use
*/}}
{{- define "upgrade-manager.persistentVolumeClaimName" -}}
{{- if .Values.persistentVolumeClaim.create }}
{{- default (include "upgrade-manager.fullname" .) .Values.persistentVolumeClaim.name }}
{{- else }}
{{- default "default" .Values.persistentVolumeClaim.name }}
{{- end }}
{{- end }}
