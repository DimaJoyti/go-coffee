# Go Coffee Staging Environment

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }
}

# Provider configuration
provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Local variables
locals {
  environment = "staging"
  common_labels = {
    environment = local.environment
    project     = "go-coffee"
    managed_by  = "terraform"
    team        = "platform"
  }
}

# Main infrastructure module
module "gcp_infrastructure" {
  source = "../../modules/gcp-infrastructure"

  # Project Configuration
  project_id   = var.project_id
  project_name = "go-coffee"
  environment  = local.environment

  # Regional Configuration
  region = var.region
  zone   = var.zone

  # Network Configuration
  subnet_cidr   = "10.20.0.0/24"
  pods_cidr     = "10.21.0.0/16"
  services_cidr = "10.22.0.0/16"
  master_cidr   = "172.16.20.0/28"

  # GKE Configuration - Staging sizing
  min_nodes           = 2
  max_nodes           = 10
  node_machine_type   = "e2-standard-4"
  node_disk_size      = 100
  use_preemptible_nodes = false

  # Node taints for staging workloads
  node_taints = [
    {
      key    = "workload-type"
      value  = "staging"
      effect = "NO_SCHEDULE"
    }
  ]

  # Database Configuration - Staging sizing
  postgres_version           = "POSTGRES_15"
  postgres_tier             = "db-custom-2-4096"
  postgres_availability_type = "ZONAL"
  postgres_disk_size        = 100

  # Redis Configuration - Staging sizing
  redis_tier         = "STANDARD_HA"
  redis_memory_size  = 4
  redis_version      = "REDIS_7_0"
  redis_replica_count = 1

  # Security Configuration - Moderate for staging
  enable_deletion_protection = true
  enable_network_policy     = true
  enable_pod_security_policy = false

  # Monitoring Configuration
  enable_monitoring = true
  enable_logging   = true
  log_retention_days = 30

  # Backup Configuration
  backup_retention_days         = 14
  enable_point_in_time_recovery = true

  # Cost Optimization
  enable_cluster_autoscaling        = true
  enable_vertical_pod_autoscaling   = true
  enable_node_auto_provisioning     = false

  # Feature Flags - Moderate for staging
  enable_workload_identity     = true
  enable_binary_authorization  = false
  enable_istio                = true
  enable_knative              = false

  # Labels
  labels = local.common_labels
}

# Staging-specific resources

# Storage bucket for staging assets
resource "google_storage_bucket" "staging_assets" {
  name          = "${var.project_id}-staging-assets"
  location      = "US"
  force_destroy = false
  project       = var.project_id

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }

  cors {
    origin          = ["https://*.staging.gocoffee.dev"]
    method          = ["GET", "HEAD"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Staging backup bucket
resource "google_storage_bucket" "staging_backups" {
  name          = "${var.project_id}-staging-backups"
  location      = var.region
  force_destroy = false
  project       = var.project_id

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 14
    }
    action {
      type = "Delete"
    }
  }
}

# Monitoring and alerting for staging
resource "google_monitoring_alert_policy" "staging_high_cpu_usage" {
  display_name = "Staging High CPU Usage"
  combiner     = "OR"
  project      = var.project_id

  conditions {
    display_name = "CPU usage above 75%"
    condition_threshold {
      filter          = "resource.type=\"gke_container\" AND metric.type=\"kubernetes.io/container/cpu/core_usage_time\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.75

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = [google_monitoring_notification_channel.email.name]

  alert_strategy {
    auto_close = "1800s"
  }
}

resource "google_monitoring_alert_policy" "staging_high_memory_usage" {
  display_name = "Staging High Memory Usage"
  combiner     = "OR"
  project      = var.project_id

  conditions {
    display_name = "Memory usage above 80%"
    condition_threshold {
      filter          = "resource.type=\"gke_container\" AND metric.type=\"kubernetes.io/container/memory/working_set_bytes\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = [google_monitoring_notification_channel.email.name]

  alert_strategy {
    auto_close = "1800s"
  }
}

# Notification channel
resource "google_monitoring_notification_channel" "email" {
  display_name = "Staging Email Notification"
  type         = "email"
  project      = var.project_id

  labels = {
    email_address = var.alert_email
  }
}

# Staging service account
resource "google_service_account" "staging_service_account" {
  account_id   = "staging-service-account"
  display_name = "Staging Service Account"
  project      = var.project_id
}

resource "google_project_iam_member" "staging_service_account_viewer" {
  project = var.project_id
  role    = "roles/viewer"
  member  = "serviceAccount:${google_service_account.staging_service_account.email}"
}

resource "google_project_iam_member" "staging_service_account_storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.staging_service_account.email}"
}

# Cloud Scheduler for staging maintenance
resource "google_cloud_scheduler_job" "staging_database_backup" {
  name             = "staging-database-backup"
  description      = "Daily staging database backup"
  schedule         = "0 3 * * *"
  time_zone        = "UTC"
  attempt_deadline = "320s"
  project          = var.project_id
  region           = var.region

  retry_config {
    retry_count = 2
  }

  http_target {
    http_method = "POST"
    uri         = "https://sqladmin.googleapis.com/sql/v1beta4/projects/${var.project_id}/instances/${module.gcp_infrastructure.postgres_instance_name}/export"
    
    headers = {
      "Content-Type" = "application/json"
    }

    oauth_token {
      service_account_email = google_service_account.staging_service_account.email
    }

    body = base64encode(jsonencode({
      exportContext = {
        fileType = "SQL"
        uri      = "gs://${google_storage_bucket.staging_backups.name}/backup-$(date +%Y%m%d-%H%M%S).sql"
      }
    }))
  }
}

# Output important values
output "cluster_name" {
  description = "GKE cluster name"
  value       = module.gcp_infrastructure.cluster_name
}

output "cluster_endpoint" {
  description = "GKE cluster endpoint"
  value       = module.gcp_infrastructure.cluster_endpoint
  sensitive   = true
}

output "postgres_connection_name" {
  description = "PostgreSQL connection name"
  value       = module.gcp_infrastructure.postgres_connection_name
}

output "redis_host" {
  description = "Redis host"
  value       = module.gcp_infrastructure.redis_host
}

output "staging_bucket_name" {
  description = "Staging assets bucket name"
  value       = google_storage_bucket.staging_assets.name
}

output "staging_backup_bucket_name" {
  description = "Staging backup bucket name"
  value       = google_storage_bucket.staging_backups.name
}

output "staging_service_account_email" {
  description = "Staging service account email"
  value       = google_service_account.staging_service_account.email
}
