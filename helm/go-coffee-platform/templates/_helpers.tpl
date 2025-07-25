{{/*
Expand the name of the chart.
*/}}
{{- define "go-coffee-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "go-coffee-platform.fullname" -}}
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
{{- define "go-coffee-platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "go-coffee-platform.labels" -}}
helm.sh/chart: {{ include "go-coffee-platform.chart" . }}
{{ include "go-coffee-platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: go-coffee
{{- end }}

{{/*
Selector labels
*/}}
{{- define "go-coffee-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "go-coffee-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service labels for a specific component
*/}}
{{- define "go-coffee-platform.serviceLabels" -}}
{{ include "go-coffee-platform.labels" . }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "go-coffee-platform.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "go-coffee-platform.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create a service name for a specific component
*/}}
{{- define "go-coffee-platform.serviceName" -}}
{{- printf "%s-%s" (include "go-coffee-platform.fullname" .) .component }}
{{- end }}

{{/*
Create a deployment name for a specific component
*/}}
{{- define "go-coffee-platform.deploymentName" -}}
{{- printf "%s-%s" (include "go-coffee-platform.fullname" .) .component }}
{{- end }}

{{/*
Create image name for a specific component
*/}}
{{- define "go-coffee-platform.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- $repository := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Create image pull policy
*/}}
{{- define "go-coffee-platform.imagePullPolicy" -}}
{{- .Values.image.pullPolicy | default "IfNotPresent" }}
{{- end }}

{{/*
Create environment variables for database connection
*/}}
{{- define "go-coffee-platform.databaseEnv" -}}
- name: DB_HOST
  value: {{ .Values.postgresql.host | quote }}
- name: DB_PORT
  value: {{ .Values.postgresql.port | quote }}
- name: DB_NAME
  value: {{ .Values.postgresql.database | quote }}
- name: DB_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "go-coffee-platform.fullname" . }}-db-secret
      key: username
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "go-coffee-platform.fullname" . }}-db-secret
      key: password
{{- end }}

{{/*
Create environment variables for Redis connection
*/}}
{{- define "go-coffee-platform.redisEnv" -}}
- name: REDIS_HOST
  value: {{ .Values.redis.host | quote }}
- name: REDIS_PORT
  value: {{ .Values.redis.port | quote }}
{{- if .Values.redis.password }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "go-coffee-platform.fullname" . }}-redis-secret
      key: password
{{- end }}
{{- end }}

{{/*
Create environment variables for Kafka connection
*/}}
{{- define "go-coffee-platform.kafkaEnv" -}}
- name: KAFKA_BROKERS
  value: {{ .Values.kafka.brokers | quote }}
- name: KAFKA_TOPIC_ORDERS
  value: {{ .Values.kafka.topics.orders | quote }}
- name: KAFKA_TOPIC_PROCESSED_ORDERS
  value: {{ .Values.kafka.topics.processedOrders | quote }}
- name: KAFKA_TOPIC_AI_EVENTS
  value: {{ .Values.kafka.topics.aiEvents | quote }}
{{- end }}

{{/*
Create common environment variables
*/}}
{{- define "go-coffee-platform.commonEnv" -}}
- name: ENVIRONMENT
  value: {{ .Values.global.environment | quote }}
- name: LOG_LEVEL
  value: {{ .Values.global.logLevel | quote }}
- name: METRICS_ENABLED
  value: {{ .Values.global.metrics.enabled | quote }}
- name: TRACING_ENABLED
  value: {{ .Values.global.tracing.enabled | quote }}
{{- if .Values.global.tracing.enabled }}
- name: JAEGER_ENDPOINT
  value: {{ .Values.global.tracing.jaegerEndpoint | quote }}
{{- end }}
{{- end }}

{{/*
Create resource limits and requests
*/}}
{{- define "go-coffee-platform.resources" -}}
{{- if .resources }}
resources:
  {{- if .resources.limits }}
  limits:
    {{- if .resources.limits.cpu }}
    cpu: {{ .resources.limits.cpu }}
    {{- end }}
    {{- if .resources.limits.memory }}
    memory: {{ .resources.limits.memory }}
    {{- end }}
  {{- end }}
  {{- if .resources.requests }}
  requests:
    {{- if .resources.requests.cpu }}
    cpu: {{ .resources.requests.cpu }}
    {{- end }}
    {{- if .resources.requests.memory }}
    memory: {{ .resources.requests.memory }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Create security context
*/}}
{{- define "go-coffee-platform.securityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
    - ALL
{{- end }}

{{/*
Create pod security context
*/}}
{{- define "go-coffee-platform.podSecurityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
{{- end }}

{{/*
Create liveness probe
*/}}
{{- define "go-coffee-platform.livenessProbe" -}}
livenessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
{{- end }}

{{/*
Create readiness probe
*/}}
{{- define "go-coffee-platform.readinessProbe" -}}
readinessProbe:
  httpGet:
    path: /ready
    port: http
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
{{- end }}

{{/*
Create startup probe
*/}}
{{- define "go-coffee-platform.startupProbe" -}}
startupProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 30
{{- end }}

{{/*
Create Istio sidecar injection annotation
*/}}
{{- define "go-coffee-platform.istioAnnotations" -}}
{{- if .Values.serviceMesh.istio.enabled }}
sidecar.istio.io/inject: "true"
{{- if .Values.serviceMesh.istio.proxy }}
sidecar.istio.io/proxyCPU: {{ .Values.serviceMesh.istio.proxy.cpu | quote }}
sidecar.istio.io/proxyMemory: {{ .Values.serviceMesh.istio.proxy.memory | quote }}
{{- end }}
{{- end }}
{{- end }}
