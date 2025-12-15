{{- define "animus-datapilot.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "animus-datapilot.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "animus-datapilot.labels" -}}
app.kubernetes.io/name: {{ include "animus-datapilot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}

{{- define "animus-datapilot.secretsName" -}}
{{ include "animus-datapilot.fullname" . }}-secrets
{{- end -}}

{{- define "animus-datapilot.postgres.serviceName" -}}
{{ include "animus-datapilot.fullname" . }}-postgres
{{- end -}}

{{- define "animus-datapilot.minio.serviceName" -}}
{{ include "animus-datapilot.fullname" . }}-minio
{{- end -}}

{{- define "animus-datapilot.databaseUrl" -}}
{{- if .Values.database.url -}}
{{ .Values.database.url -}}
{{- else if .Values.postgres.enabled -}}
{{- printf "postgres://%s:%s@%s:%d/%s?sslmode=disable" .Values.postgres.user .Values.postgres.password (include "animus-datapilot.postgres.serviceName" .) (.Values.postgres.port | int) .Values.postgres.db -}}
{{- else -}}
{{- fail "Either database.url must be set or postgres.enabled must be true" -}}
{{- end -}}
{{- end -}}

{{- define "animus-datapilot.minioEndpoint" -}}
{{- if .Values.minio.enabled -}}
{{- printf "%s:%d" (include "animus-datapilot.minio.serviceName" .) (.Values.minio.apiPort | int) -}}
{{- else -}}
{{- required "minio.endpoint is required when minio.enabled=false" .Values.minio.endpoint -}}
{{- end -}}
{{- end -}}

{{- define "animus-datapilot.uiImage" -}}
{{- $repo := .Values.ui.image.repository | default "" -}}
{{- $tag := .Values.ui.image.tag | default "" -}}
{{- if $repo -}}
{{- printf "%s:%s" $repo (default .Values.image.tag $tag) -}}
{{- else -}}
{{- printf "%s/ui:%s" .Values.image.repository (default "latest" .Values.image.tag) -}}
{{- end -}}
{{- end -}}

{{- define "animus-datapilot.uiPullPolicy" -}}
{{- default .Values.image.pullPolicy .Values.ui.image.pullPolicy -}}
{{- end -}}

{{- define "animus-datapilot.gatewayServiceURL" -}}
{{- printf "http://%s-gateway:%d" (include "animus-datapilot.fullname" .) (.Values.services.gateway.port | int) -}}
{{- end -}}
