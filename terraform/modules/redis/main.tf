# Redis Module for Go Coffee
# Provides managed Redis instances with high availability and monitoring

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Redis instance
resource "google_redis_instance" "main" {
  name           = var.instance_name
  tier           = var.tier
  memory_size_gb = var.memory_size_gb
  region         = var.region
  project        = var.project_id

  # Network configuration
  authorized_network = data.google_compute_network.vpc.id
  connect_mode       = "PRIVATE_SERVICE_ACCESS"

  # Redis configuration
  redis_version     = var.redis_version
  display_name      = "${var.instance_name} (${var.environment})"
  reserved_ip_range = var.reserved_ip_range

  # High availability configuration
  replica_count          = var.replica_count
  read_replicas_mode     = var.read_replicas_mode
  secondary_ip_range     = var.secondary_ip_range

  # Maintenance configuration
  maintenance_policy {
    weekly_maintenance_window {
      day = "SUNDAY"
      start_time {
        hours   = 2
        minutes = 0
        seconds = 0
        nanos   = 0
      }
    }
  }

  # Redis configuration parameters
  redis_configs = {
    maxmemory-policy = "allkeys-lru"
    notify-keyspace-events = "Ex"
    timeout = "300"
  }

  # Labels
  labels = {
    environment = var.environment
    service     = "go-coffee"
    component   = "cache"
    managed_by  = "terraform"
  }

  # Lifecycle management
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      redis_configs["maxmemory"],
    ]
  }
}

# Data source for VPC network
data "google_compute_network" "vpc" {
  name    = var.network_name
  project = var.project_id
}

# Redis backup configuration (for production)
resource "google_redis_instance" "backup" {
  count = var.environment == "prod" ? 1 : 0

  name           = "${var.instance_name}-backup"
  tier           = "BASIC"
  memory_size_gb = var.memory_size_gb
  region         = var.backup_region != "" ? var.backup_region : var.region
  project        = var.project_id

  authorized_network = data.google_compute_network.vpc.id
  connect_mode       = "PRIVATE_SERVICE_ACCESS"

  redis_version = var.redis_version
  display_name  = "${var.instance_name} Backup (${var.environment})"

  labels = {
    environment = var.environment
    service     = "go-coffee"
    component   = "cache-backup"
    managed_by  = "terraform"
  }
}

# Monitoring alert policy for Redis memory usage
resource "google_monitoring_alert_policy" "redis_memory" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Redis Memory Usage - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Redis memory usage above 80%"
    
    condition_threshold {
      filter          = "resource.type=\"redis_instance\" AND resource.labels.instance_id=\"${google_redis_instance.main.id}\" AND metric.type=\"redis.googleapis.com/stats/memory/usage_ratio\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.instance_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Redis memory usage has exceeded 80% for more than 5 minutes. Consider scaling up the instance or optimizing memory usage."
  }
}

# Monitoring alert policy for Redis connections
resource "google_monitoring_alert_policy" "redis_connections" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Redis Connection Count - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Redis connection count above threshold"
    
    condition_threshold {
      filter          = "resource.type=\"redis_instance\" AND resource.labels.instance_id=\"${google_redis_instance.main.id}\" AND metric.type=\"redis.googleapis.com/stats/connections/total\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = var.max_connections_threshold

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.instance_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Redis connection count has exceeded the configured threshold for more than 5 minutes. Consider implementing connection pooling or scaling."
  }
}

# Monitoring alert policy for Redis operations per second
resource "google_monitoring_alert_policy" "redis_ops" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Redis Operations Rate - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Redis operations per second above threshold"
    
    condition_threshold {
      filter          = "resource.type=\"redis_instance\" AND resource.labels.instance_id=\"${google_redis_instance.main.id}\" AND metric.type=\"redis.googleapis.com/stats/operations/total\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 10000  # Adjust based on your expected load

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["resource.labels.instance_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Redis operations per second has exceeded 10,000 ops/sec for more than 5 minutes. Monitor for performance degradation."
  }
}

# Monitoring alert policy for Redis CPU usage
resource "google_monitoring_alert_policy" "redis_cpu" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Redis CPU Usage - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Redis CPU usage above 80%"
    
    condition_threshold {
      filter          = "resource.type=\"redis_instance\" AND resource.labels.instance_id=\"${google_redis_instance.main.id}\" AND metric.type=\"redis.googleapis.com/stats/cpu_usage\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.instance_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Redis CPU usage has exceeded 80% for more than 5 minutes. Consider optimizing workload or scaling up the instance."
  }
}
