# Service Mesh Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
}

variable "cluster_location" {
  description = "GKE cluster location (region or zone)"
  type        = string
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "enable_istio" {
  description = "Enable Istio service mesh"
  type        = bool
  default     = true
}

variable "istio_version" {
  description = "Istio version to install"
  type        = string
  default     = "1.20.0"
}

variable "enable_mtls" {
  description = "Enable mutual TLS"
  type        = bool
  default     = true
}

variable "enable_tracing" {
  description = "Enable distributed tracing"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Enable monitoring and observability"
  type        = bool
  default     = true
}

variable "tls_secret_name" {
  description = "Name of the TLS secret for HTTPS"
  type        = string
  default     = "go-coffee-tls"
}

variable "services" {
  description = "Map of services for destination rules"
  type = map(object({
    port = number
  }))
  default = {
    "api-gateway" = {
      port = 8080
    }
    "order-service" = {
      port = 8081
    }
    "payment-service" = {
      port = 8082
    }
    "kitchen-service" = {
      port = 8083
    }
    "user-gateway" = {
      port = 8084
    }
    "security-gateway" = {
      port = 8085
    }
    "web-ui-backend" = {
      port = 8086
    }
    "ai-search" = {
      port = 8087
    }
    "bright-data-hub" = {
      port = 8088
    }
    "communication-hub" = {
      port = 8089
    }
    "enterprise-service" = {
      port = 8090
    }
  }
}

variable "ingress_gateway_type" {
  description = "Type of ingress gateway service"
  type        = string
  default     = "LoadBalancer"
  
  validation {
    condition = contains([
      "LoadBalancer",
      "NodePort",
      "ClusterIP"
    ], var.ingress_gateway_type)
    error_message = "Ingress gateway type must be one of: LoadBalancer, NodePort, ClusterIP."
  }
}

variable "tracing_sampling_rate" {
  description = "Tracing sampling rate (0.0 to 100.0)"
  type        = number
  default     = 1.0
  
  validation {
    condition     = var.tracing_sampling_rate >= 0.0 && var.tracing_sampling_rate <= 100.0
    error_message = "Tracing sampling rate must be between 0.0 and 100.0."
  }
}

variable "circuit_breaker_config" {
  description = "Circuit breaker configuration"
  type = object({
    consecutive_gateway_errors = number
    interval                   = string
    base_ejection_time         = string
    max_ejection_percent       = number
  })
  default = {
    consecutive_gateway_errors = 5
    interval                   = "30s"
    base_ejection_time         = "30s"
    max_ejection_percent       = 50
  }
}

variable "connection_pool_config" {
  description = "Connection pool configuration"
  type = object({
    max_connections              = number
    http1_max_pending_requests   = number
    http2_max_requests           = number
    max_requests_per_connection  = number
    max_retries                  = number
  })
  default = {
    max_connections             = 100
    http1_max_pending_requests  = 50
    http2_max_requests          = 100
    max_requests_per_connection = 10
    max_retries                 = 3
  }
}

variable "load_balancer_algorithm" {
  description = "Load balancer algorithm"
  type        = string
  default     = "LEAST_CONN"
  
  validation {
    condition = contains([
      "ROUND_ROBIN",
      "LEAST_CONN",
      "RANDOM",
      "PASSTHROUGH"
    ], var.load_balancer_algorithm)
    error_message = "Load balancer algorithm must be one of: ROUND_ROBIN, LEAST_CONN, RANDOM, PASSTHROUGH."
  }
}

variable "enable_access_logs" {
  description = "Enable access logs for Envoy proxies"
  type        = bool
  default     = true
}

variable "access_log_format" {
  description = "Access log format"
  type        = string
  default     = "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\""
}

variable "enable_prometheus_metrics" {
  description = "Enable Prometheus metrics collection"
  type        = bool
  default     = true
}

variable "prometheus_scrape_interval" {
  description = "Prometheus scrape interval"
  type        = string
  default     = "15s"
}

variable "enable_jaeger_tracing" {
  description = "Enable Jaeger tracing"
  type        = bool
  default     = true
}

variable "jaeger_endpoint" {
  description = "Jaeger collector endpoint"
  type        = string
  default     = "http://jaeger-collector.istio-system:14268/api/traces"
}

variable "enable_grafana_dashboards" {
  description = "Enable Grafana dashboards for Istio"
  type        = bool
  default     = true
}

variable "enable_kiali" {
  description = "Enable Kiali service mesh observability"
  type        = bool
  default     = true
}

variable "kiali_version" {
  description = "Kiali version to install"
  type        = string
  default     = "1.75.0"
}
