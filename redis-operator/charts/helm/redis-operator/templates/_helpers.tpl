{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "redis-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "redis-operator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "redis-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "redis-operator.labels" -}}
helm.sh/chart: {{ include "redis-operator.chart" . }}
{{ include "redis-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "redis-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "redis-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "redis-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "redis-operator.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}


{{/*
[NoSQL Operator Core] Secret template
Arguments:
Dictionary with:
1. "secret" section includes next elements:
    .secretName (required)
    .password (required)
    .username (optional)
Usage example:
{{template "nosql.core.secret" (dict "secret" .Values.redis "values" .Values)}}
*/}}
{{- define "nosql.core.secret" -}}
{{ $_ := set . "userEnv" "" }}
{{ $_ := set . "passEnv" "" }}
{{include "nosql.core.secret.fromEnv" $_ }}
{{- end -}}

{{/*
[NoSQL Operator Core] Secret template with env override
Arguments:
Dictionary with:
1. "secret" section includes next elements:
    .secretName (required)
    .password (required)
    .username (optional)
Usage example:
{{template "nosql.core.secret.fromEnv" (dict "secret" .Values.redis "userEnv" .Values.USERNAME "passEnv" .Values.PASSWORD "values" .Values)}}
*/}}
{{- define "nosql.core.secret.fromEnv" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .secret.secretName }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation
  labels:
    {{ include "redis.defaultLabels" .values | nindent 4 }}
stringData:
  password: {{ include "fromEnv" (dict "envName" .passEnv "default" .secret.password) | quote }}
  {{- if .secret.username }}
  username: {{ include "fromEnv" (dict "envName" .userEnv "default" .secret.username) | quote }}
  {{- end }}
type: Opaque
{{- end -}}

{{/*
[NoSQL Operator Core] Internal secret template
{{template "nosql.core.secret.internal" (dict "secret" .Values.redis "values" .Values)}}
*/}}
{{- define "nosql.core.secret.internal" -}}
{{include "nosql.core.secret" .}}
{{- end -}}

{{/*
[NoSQL Operator Core] External secret template
{{template "nosql.core.secret.external" (dict "secret" .Values.redis "values" .Values)}}
*/}}
{{- define "nosql.core.secret.external" -}}
{{include "nosql.core.secret" .}}
{{- end -}}

{{/*
[Cassandra Operator Core] Docker image
Dictionary with:
1. "deployName" - deploy-param from description.yaml
2. "SERVICE_NAME" - name of service with git group and git repo
3. "vals" - .Values
4.  "default" - default docker image
{{template "find_image" (dict "deployName" "cassandraOperator" "SERVICE_NAME" "cassandra-operator" "vals" .Values "default" .Values.operator.dockerImage) }}

*/}}
{{- define "find_image" -}}
  {{- $image := .default -}}

  {{ printf "%s" $image }}
{{- end -}}

{{- define "dbaasAdapter.certDnsNames" -}}
{{- $dnsNames := list "localhost" (printf "%s.%s" "dbaas-redis-adapter" .Release.Namespace) (printf "%s.%s.svc" "dbaas-redis-adapter" .Release.Namespace) -}}
{{- $dnsNames = concat $dnsNames .Values.dbaas.tls.subjectAlternativeName.additionalDnsNames -}}
{{- $dnsNames | toYaml -}}
{{- end -}}

{{- define "common.certIpAddresses" -}}
{{- $ipAddresses := list "127.0.0.1" -}}
{{- $ipAddresses = concat $ipAddresses .Values.dbaas.tls.subjectAlternativeName.additionalIpAddresses -}}
{{- $ipAddresses | toYaml -}}
{{- end -}}

{{/*
TLS Static Metric secret template
Arguments:
Dictionary with:
* "namespace" is a namespace of application
* "application" is name of application
* "service" is a name of service
* "enabledSsl" is ssl enabled for service
* "secret" is a name of tls secret for service
* "certProvider" is a type of tls certificates provider
* "certificate" is a name of CertManger's Certificate resource for service
Usage example:
{{template "global.tlsStaticMetric" (dict "namespace" .Release.Namespace "application" .Chart.Name "service" .global.name "enabledSsl" (include "global.sslEnabled" .) "secret" (include "global.sslSecretName" .) "certProvider" (include "services.certProvider" .) "certificate" (printf "%s-tls-certificate" (include "global.name")) }}
*/}}
{{- define "global.tlsStaticMetric" -}}
- expr: {{ ternary "1" "0" .enabledSsl }}
  labels:
    namespace: "{{ .namespace }}"
    application: "{{ .application }}"
    service: "{{ .service }}"
    {{ if .enabledSsl }}
    secret: "{{ .secret }}"
    {{ if eq .certProvider "cert-manager" }}
    certificate: "{{ .certificate }}"
    {{ end }}
    {{ end }}
  record: service:tls_status:info
{{- end -}}

{{/*
[Redis Operator Core] from env or from values
Dictionary with:
1. "envName" - name of env var to get value from
2.  "default" - default value from values.yaml
{{template "fromEnv" (dict "envName" .Values.API_DBAAS_ADDRESS "default" .Values.dbaas.aggregator.dbaasAggregatorRegistrationAddress) }}
*/}}
{{- define "fromEnv" -}}
  {{- $envValue := .envName -}}
{{- if and (ne ($envValue | toString) "<nil>") (ne ($envValue | toString) "") -}}
    {{- .envName -}}
  {{- else -}}
    {{- .default -}}
  {{- end -}}
{{- end -}}


{{- define "getResourcesForProfile" -}}
  {{- $flavor := default "" .dotVar | toString }}
{{- if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
  {{- $flavor = .envVar | toString -}}
{{- end -}}
  {{- if eq $flavor "small" -}}
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 250m
        memory: 256Mi
  {{- else if eq $flavor "medium" -}}
    resources:
      requests:
        cpu: 1
        memory: 2048Mi
      limits:
        cpu: 2
        memory: 4096Mi
  {{- else if eq $flavor "large" -}}
    resources:
      requests:
        cpu: 2
        memory: 4096Mi
      limits:
        cpu: 4
        memory: 8192Mi
  {{- else if $flavor -}}
  {{- fail "value for .Values.global.profile is not one of  `small`, `medium`, `large`" }}
  {{- else -}}
    resources:
      requests:
        cpu: {{ default "125m" .values.redis.resources.requests.cpu }}
        memory: {{ default "251Mi" .values.redis.resources.requests.memory }}
      limits:
        cpu: {{ default "250m" .values.redis.resources.limits.cpu }}
        memory: {{ default "256Mi" .values.redis.resources.limits.memory }}
  {{- end -}}
{{- end -}}


{{/*
Dictionary with:
Uses value from values.yaml if defined, otherwise value from environment variable if defined, else - default
1. "dotVar" - parameter defined with dots like dbaas.install
2. "enVar" - parameter defined as environment variable like DBAAS_ENABLED
3.  "default" - default value
{{template "fromValuesThenEnvElseDefault" (dict "dotVar" .Values.dbaas.install "envVar" .Values.DBAAS_ENABLED "default" true ) }}
*/}}
{{- define "fromValuesThenEnvElseDefault" -}}
  {{- if and (ne (.dotVar | toString) "<nil>") (ne (.dotVar | toString) "") -}}
    {{- .dotVar -}}
  {{- else if and (ne (.envVar | toString) "<nil>") (ne (.envVar | toString) "") -}}
    {{- .envVar -}}
  {{- else -}}
    {{- .default -}}
  {{- end -}}
{{- end -}}

{{/* Gateway name for HTTPRoute parentRefs. */}}
{{- define "redis.gatewayApiName" -}}
  {{- if and (ne (.Values.GATEWAY_SYSTEM_NAME | toString) "<nil>") (ne (.Values.GATEWAY_SYSTEM_NAME | toString) "") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.GATEWAY_SYSTEM_NAME | toString -}}
  {{- else -}}
    {{- .Values.dbaas.gatewayApi.gatewayName -}}
  {{- end -}}
{{- end -}}

{{/* Gateway namespace for HTTPRoute parentRefs. */}}
{{- define "redis.gatewayApiNamespace" -}}
  {{- if and (ne (.Values.GATEWAY_SYSTEM_NAMESPACE | toString) "<nil>") (ne (.Values.GATEWAY_SYSTEM_NAMESPACE | toString) "") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.GATEWAY_SYSTEM_NAMESPACE | toString -}}
  {{- else -}}
    {{- .Values.dbaas.gatewayApi.gatewayNamespace -}}
  {{- end -}}
{{- end -}}

{{/*
Common redis resources labels
*/}}
{{- define "redis.defaultLabels" -}}
app.kubernetes.io/version: {{ default "" .Values.ARTIFACT_DESCRIPTOR_VERSION | trunc 63 | trimAll "-_." }}
app.kubernetes.io/part-of: {{ default "dbaas-redis" .Values.PART_OF }}
app.kubernetes.io/managed-by: {{ default "operator" .Values.MANAGED_BY }}
deployment.netcracker.com/sessionId: "{{ default "default-session-id" .Values.DEPLOYMENT_SESSION_ID }}"
{{- end -}}

{{- define "redis.monitoredImages" -}}
  ""
{{- end -}}
