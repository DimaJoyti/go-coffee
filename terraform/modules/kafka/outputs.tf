output "bootstrap_servers" {
  description = "Bootstrap сервери Kafka"
  value       = "${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092"
}

output "topics" {
  description = "Створені топіки Kafka"
  value       = [var.kafka_topic_name, var.kafka_processed_topic_name]
}

output "kafka_namespace" {
  description = "Простір імен Kubernetes для Kafka"
  value       = kubernetes_namespace.kafka.metadata[0].name
}
