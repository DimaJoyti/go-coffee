# PostgreSQL Module for Go Coffee
# Provides managed PostgreSQL instances with high availability and monitoring

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }
}

# Generate random password for database
resource "random_password" "db_password" {
  length  = 32
  special = true
}

# Store password in Secret Manager
resource "google_secret_manager_secret" "db_password" {
  secret_id = "${var.instance_name}-password"
  project   = var.project_id

  replication {
    auto {}
  }

  labels = {
    environment = var.environment
    service     = "go-coffee"
    component   = "database"
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db_password.result
}

# PostgreSQL instance
resource "google_sql_database_instance" "main" {
  name             = var.instance_name
  database_version = var.database_version
  region           = var.region
  project          = var.project_id

  # Deletion protection for production
  deletion_protection = var.environment == "prod"

  settings {
    tier              = var.tier
    availability_type = var.availability_type
    disk_type         = var.disk_type
    disk_size         = var.disk_size
    disk_autoresize   = true
    disk_autoresize_limit = var.disk_autoresize_limit

    # Backup configuration
    backup_configuration {
      enabled                        = var.backup_enabled
      start_time                     = "02:00"
      location                       = var.backup_location
      point_in_time_recovery_enabled = var.point_in_time_recovery_enabled
      transaction_log_retention_days = var.transaction_log_retention_days
      
      backup_retention_settings {
        retained_backups = var.retained_backups
        retention_unit   = "COUNT"
      }
    }

    # IP configuration
    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = data.google_compute_network.vpc.id
      enable_private_path_for_google_cloud_services = true
      
      dynamic "authorized_networks" {
        for_each = var.authorized_networks
        content {
          name  = authorized_networks.value.name
          value = authorized_networks.value.value
        }
      }
    }

    # Database flags for optimization
    dynamic "database_flags" {
      for_each = var.database_flags
      content {
        name  = database_flags.value.name
        value = database_flags.value.value
      }
    }

    # Maintenance window
    maintenance_window {
      day          = 7  # Sunday
      hour         = 2  # 2 AM
      update_track = "stable"
    }

    # Insights configuration
    insights_config {
      query_insights_enabled  = true
      query_string_length     = 1024
      record_application_tags = true
      record_client_address   = true
    }

    # User labels
    user_labels = {
      environment = var.environment
      service     = "go-coffee"
      component   = "database"
      managed_by  = "terraform"
    }
  }

  # Lifecycle management
  lifecycle {
    prevent_destroy = true
    ignore_changes = [
      settings[0].disk_size,
    ]
  }
}

# Data source for VPC network
data "google_compute_network" "vpc" {
  name    = var.network_name
  project = var.project_id
}

# Database user
resource "google_sql_user" "main" {
  name     = var.db_username
  instance = google_sql_database_instance.main.name
  password = random_password.db_password.result
  project  = var.project_id
}

# Create databases
resource "google_sql_database" "databases" {
  for_each = toset(var.databases)
  
  name     = each.value
  instance = google_sql_database_instance.main.name
  project  = var.project_id
  
  charset   = "UTF8"
  collation = "en_US.UTF8"
}

# Read replica for production
resource "google_sql_database_instance" "read_replica" {
  count = var.environment == "prod" && var.create_read_replica ? 1 : 0

  name                 = "${var.instance_name}-read-replica"
  master_instance_name = google_sql_database_instance.main.name
  region               = var.read_replica_region != "" ? var.read_replica_region : var.region
  database_version     = var.database_version
  project              = var.project_id

  replica_configuration {
    failover_target = false
  }

  settings {
    tier              = var.read_replica_tier != "" ? var.read_replica_tier : var.tier
    availability_type = "ZONAL"
    disk_type         = var.disk_type
    disk_autoresize   = true

    ip_configuration {
      ipv4_enabled    = false
      private_network = data.google_compute_network.vpc.id
    }

    user_labels = {
      environment = var.environment
      service     = "go-coffee"
      component   = "database-replica"
      managed_by  = "terraform"
    }
  }
}

# Monitoring alert policy for CPU usage
resource "google_monitoring_alert_policy" "database_cpu" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Database CPU Usage - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Database CPU usage above 80%"
    
    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${google_sql_database_instance.main.name}\" AND metric.type=\"cloudsql.googleapis.com/database/cpu/utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.database_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Database CPU usage has exceeded 80% for more than 5 minutes. This may indicate high load or inefficient queries."
  }
}

# Monitoring alert policy for memory usage
resource "google_monitoring_alert_policy" "database_memory" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Database Memory Usage - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Database memory usage above 85%"
    
    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${google_sql_database_instance.main.name}\" AND metric.type=\"cloudsql.googleapis.com/database/memory/utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.85

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.database_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Database memory usage has exceeded 85% for more than 5 minutes. Consider scaling up the instance or optimizing queries."
  }
}

# Monitoring alert policy for disk usage
resource "google_monitoring_alert_policy" "database_disk" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Database Disk Usage - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Database disk usage above 80%"
    
    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${google_sql_database_instance.main.name}\" AND metric.type=\"cloudsql.googleapis.com/database/disk/utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.database_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Database disk usage has exceeded 80% for more than 5 minutes. Consider increasing disk size or cleaning up old data."
  }
}

# Monitoring alert policy for connection count
resource "google_monitoring_alert_policy" "database_connections" {
  count = var.enable_monitoring ? 1 : 0

  display_name = "Database Connection Count - ${var.instance_name}"
  project      = var.project_id
  combiner     = "OR"

  conditions {
    display_name = "Database connection count above 80% of max"
    
    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${google_sql_database_instance.main.name}\" AND metric.type=\"cloudsql.googleapis.com/database/postgresql/num_backends\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 160  # 80% of default max_connections (200)

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_MEAN"
        group_by_fields      = ["resource.labels.database_id"]
      }
    }
  }

  notification_channels = length(var.notification_channels) > 0 ? var.notification_channels : []

  alert_strategy {
    auto_close = "1800s"
  }

  enabled = true

  documentation {
    content = "Database connection count has exceeded 80% of maximum connections for more than 5 minutes. Consider connection pooling or scaling."
  }
}
