{{- define "name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "fullname" -}}
{{- .Chart.Name -}}
{{- end -}}


{{- define "domain" -}}
{{- if hasKey .Values "domainSuffix" -}}
{{- printf "%s.%s" .Values.ingress.subdomain .Values.domainSuffix -}}
{{- else -}}
{{- printf "%s.%s.%s" .Values.ingress.subdomain .Release.Namespace .Values.domain -}}
{{- end -}}
{{- end -}}
