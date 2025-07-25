# Service Mesh Module Outputs

output "istio_enabled" {
  description = "Whether Istio is enabled"
  value       = var.enable_istio
}

output "istio_version" {
  description = "Installed Istio version"
  value       = var.enable_istio ? var.istio_version : null
}

output "istio_namespace" {
  description = "Istio system namespace"
  value       = var.enable_istio ? kubernetes_namespace.istio_system[0].metadata[0].name : null
}

output "go_coffee_namespace" {
  description = "Go Coffee application namespace with Istio injection"
  value       = var.enable_istio ? kubernetes_namespace.go_coffee[0].metadata[0].name : null
}

output "istio_gateway_external_ip" {
  description = "External IP of Istio ingress gateway"
  value       = var.enable_istio ? helm_release.istio_gateway[0].status[0].load_balancer[0].ingress[0].ip : null
}

output "mtls_enabled" {
  description = "Whether mutual TLS is enabled"
  value       = var.enable_mtls
}

output "tracing_enabled" {
  description = "Whether distributed tracing is enabled"
  value       = var.enable_tracing
}

output "monitoring_enabled" {
  description = "Whether monitoring is enabled"
  value       = var.enable_monitoring
}

output "service_mesh_config" {
  description = "Service mesh configuration summary"
  value = {
    istio_enabled     = var.enable_istio
    mtls_enabled      = var.enable_mtls
    tracing_enabled   = var.enable_tracing
    monitoring_enabled = var.enable_monitoring
    istio_version     = var.istio_version
    environment       = var.environment
  }
}

output "circuit_breaker_config" {
  description = "Circuit breaker configuration"
  value       = var.circuit_breaker_config
}

output "connection_pool_config" {
  description = "Connection pool configuration"
  value       = var.connection_pool_config
}

output "load_balancer_algorithm" {
  description = "Load balancer algorithm"
  value       = var.load_balancer_algorithm
}

output "gateway_hosts" {
  description = "Gateway hosts configuration"
  value       = var.enable_istio ? ["*"] : []
}

output "virtual_services" {
  description = "List of configured virtual services"
  value = var.enable_istio ? [
    {
      name      = "go-coffee-api-gateway"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
      hosts     = ["*"]
      gateways  = ["go-coffee-gateway"]
    }
  ] : []
}

output "destination_rules" {
  description = "List of configured destination rules"
  value = var.enable_istio ? [
    for service_name, config in var.services : {
      name      = "${service_name}-destination-rule"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
      host      = service_name
      port      = config.port
    }
  ] : []
}

output "authorization_policies" {
  description = "List of configured authorization policies"
  value = var.enable_istio ? [
    {
      name      = "go-coffee-authz"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
  ] : []
}

output "peer_authentication_policies" {
  description = "List of configured peer authentication policies"
  value = var.enable_istio && var.enable_mtls ? [
    {
      name      = "default"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
      mode      = "STRICT"
    }
  ] : []
}

output "tls_configuration" {
  description = "TLS configuration for the gateway"
  value = {
    enabled     = var.enable_istio
    secret_name = var.tls_secret_name
    mode        = "SIMPLE"
  }
}

output "observability_tools" {
  description = "Enabled observability tools"
  value = {
    prometheus_metrics = var.enable_prometheus_metrics
    jaeger_tracing     = var.enable_jaeger_tracing
    grafana_dashboards = var.enable_grafana_dashboards
    kiali_enabled      = var.enable_kiali
    access_logs        = var.enable_access_logs
  }
}

output "helm_releases" {
  description = "Installed Helm releases"
  value = var.enable_istio ? {
    istio_base    = helm_release.istio_base[0].name
    istiod        = helm_release.istiod[0].name
    istio_gateway = helm_release.istio_gateway[0].name
  } : {}
}
