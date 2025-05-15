# Вихідні значення для мережі
output "network_name" {
  description = "Назва створеної VPC мережі"
  value       = module.network.network_name
}

output "subnet_name" {
  description = "Назва створеної підмережі"
  value       = module.network.subnet_name
}

output "subnet_cidr" {
  description = "CIDR блок створеної підмережі"
  value       = module.network.subnet_cidr
}

# Вихідні значення для GKE
output "gke_cluster_name" {
  description = "Назва створеного кластера GKE"
  value       = module.gke.cluster_name
}

output "gke_endpoint" {
  description = "Endpoint кластера GKE"
  value       = module.gke.endpoint
}

output "gke_kubeconfig" {
  description = "Команда для отримання kubeconfig"
  value       = "gcloud container clusters get-credentials ${module.gke.cluster_name} --region ${var.region} --project ${var.project_id}"
}

# Вихідні значення для Kafka
output "kafka_bootstrap_servers" {
  description = "Bootstrap сервери Kafka"
  value       = module.kafka.bootstrap_servers
}

output "kafka_topics" {
  description = "Створені топіки Kafka"
  value       = module.kafka.topics
}

# Вихідні значення для моніторингу
output "grafana_url" {
  description = "URL для доступу до Grafana"
  value       = module.monitoring.grafana_url
}

output "prometheus_url" {
  description = "URL для доступу до Prometheus"
  value       = module.monitoring.prometheus_url
}
