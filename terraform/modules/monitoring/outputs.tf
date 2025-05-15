output "prometheus_url" {
  description = "URL для доступу до Prometheus"
  value       = var.enable_monitoring ? "http://prometheus-server.${kubernetes_namespace.monitoring.metadata[0].name}.svc.cluster.local:80" : ""
}

output "grafana_url" {
  description = "URL для доступу до Grafana"
  value       = var.enable_monitoring ? "http://grafana.${kubernetes_namespace.monitoring.metadata[0].name}.svc.cluster.local:80" : ""
}

output "monitoring_namespace" {
  description = "Простір імен Kubernetes для моніторингу"
  value       = kubernetes_namespace.monitoring.metadata[0].name
}
