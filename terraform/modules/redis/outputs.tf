# Redis Module Outputs

output "redis_instance_id" {
  description = "The ID of the Redis instance"
  value       = google_redis_instance.main.id
}

output "redis_instance_name" {
  description = "The name of the Redis instance"
  value       = google_redis_instance.main.name
}

output "redis_host" {
  description = "The IP address of the Redis instance"
  value       = google_redis_instance.main.host
}

output "redis_port" {
  description = "The port of the Redis instance"
  value       = google_redis_instance.main.port
}

output "redis_connection_string" {
  description = "Redis connection string"
  value       = "redis://${google_redis_instance.main.host}:${google_redis_instance.main.port}"
  sensitive   = true
}

output "redis_memory_size_gb" {
  description = "The memory size of the Redis instance in GB"
  value       = google_redis_instance.main.memory_size_gb
}

output "redis_version" {
  description = "The version of Redis"
  value       = google_redis_instance.main.redis_version
}

output "redis_tier" {
  description = "The service tier of the Redis instance"
  value       = google_redis_instance.main.tier
}

output "redis_region" {
  description = "The region of the Redis instance"
  value       = google_redis_instance.main.region
}

output "redis_backup_instance_id" {
  description = "The ID of the backup Redis instance"
  value       = var.environment == "prod" ? google_redis_instance.backup[0].id : null
}

output "redis_backup_host" {
  description = "The IP address of the backup Redis instance"
  value       = var.environment == "prod" ? google_redis_instance.backup[0].host : null
}

output "redis_current_location_id" {
  description = "The current zone where the Redis endpoint is placed"
  value       = google_redis_instance.main.current_location_id
}

output "redis_persistence_iam_identity" {
  description = "Cloud IAM identity used by import/export operations"
  value       = google_redis_instance.main.persistence_iam_identity
}

output "redis_server_ca_certs" {
  description = "List of server CA certificates for the instance"
  value       = google_redis_instance.main.server_ca_certs
  sensitive   = true
}

output "redis_auth_string" {
  description = "AUTH string for the Redis instance"
  value       = google_redis_instance.main.auth_string
  sensitive   = true
}

output "redis_maintenance_schedule" {
  description = "Upcoming maintenance schedule for the instance"
  value       = google_redis_instance.main.maintenance_schedule
}

output "redis_nodes" {
  description = "Info per node"
  value       = google_redis_instance.main.nodes
}
