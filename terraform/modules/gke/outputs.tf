output "cluster_name" {
  description = "Назва створеного кластера GKE"
  value       = google_container_cluster.primary.name
}

output "cluster_id" {
  description = "ID створеного кластера GKE"
  value       = google_container_cluster.primary.id
}

output "endpoint" {
  description = "Endpoint кластера GKE"
  value       = google_container_cluster.primary.endpoint
}

output "ca_certificate" {
  description = "Сертифікат CA кластера GKE"
  value       = google_container_cluster.primary.master_auth.0.cluster_ca_certificate
  sensitive   = true
}
