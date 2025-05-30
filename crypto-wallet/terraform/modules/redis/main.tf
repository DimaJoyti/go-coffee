resource "google_redis_instance" "redis" {
  name           = "${var.redis_instance_name}-${var.environment}"
  tier           = var.redis_tier
  memory_size_gb = var.redis_memory_size_gb
  region         = var.region
  
  redis_version  = var.redis_version
  display_name   = "Web3 Wallet Redis ${var.environment} (${var.region})"
  
  authorized_network = var.network_id
  
  redis_configs = {
    "maxmemory-policy" = "allkeys-lru"
  }
  
  labels = {
    environment = var.environment
    region      = var.region
  }
}

# Output the Redis host
output "redis_host" {
  value = google_redis_instance.redis.host
}

# Output the Redis port
output "redis_port" {
  value = google_redis_instance.redis.port
}
