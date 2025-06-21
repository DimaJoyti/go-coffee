# Go Coffee Development Environment

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
  environment = "development"
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
  subnet_cidr   = "10.10.0.0/24"
  pods_cidr     = "10.11.0.0/16"
  services_cidr = "10.12.0.0/16"
  master_cidr   = "172.16.10.0/28"

  # GKE Configuration - Development sizing
  min_nodes           = 1
  max_nodes           = 5
  node_machine_type   = "e2-standard-2"
  node_disk_size      = 50
  use_preemptible_nodes = true

  # Node taints for development workloads
  node_taints = [
    {
      key    = "workload-type"
      value  = "development"
      effect = "NO_SCHEDULE"
    }
  ]

  # Database Configuration - Development sizing
  postgres_version           = "POSTGRES_15"
  postgres_tier             = "db-f1-micro"
  postgres_availability_type = "ZONAL"
  postgres_disk_size        = 20

  # Redis Configuration - Development sizing
  redis_tier         = "BASIC"
  redis_memory_size  = 1
  redis_version      = "REDIS_7_0"
  redis_replica_count = 0

  # Security Configuration - Relaxed for development
  enable_deletion_protection = false
  enable_network_policy     = false
  enable_pod_security_policy = false

  # Monitoring Configuration
  enable_monitoring = true
  enable_logging   = true
  log_retention_days = 7

  # Backup Configuration - Minimal for development
  backup_retention_days         = 7
  enable_point_in_time_recovery = false

  # Cost Optimization
  enable_cluster_autoscaling        = true
  enable_vertical_pod_autoscaling   = false
  enable_node_auto_provisioning     = false

  # Feature Flags - Minimal for development
  enable_workload_identity     = false
  enable_binary_authorization  = false
  enable_istio                = false
  enable_knative              = false

  # Labels
  labels = local.common_labels
}

# Development-specific resources

# Storage bucket for development assets
resource "google_storage_bucket" "dev_assets" {
  name          = "${var.project_id}-dev-assets"
  location      = "US"
  force_destroy = true
  project       = var.project_id

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 7
    }
    action {
      type = "Delete"
    }
  }

  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD", "POST"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Simple monitoring for development
resource "google_monitoring_alert_policy" "dev_high_cpu_usage" {
  display_name = "Development High CPU Usage"
  combiner     = "OR"
  project      = var.project_id

  conditions {
    display_name = "CPU usage above 90%"
    condition_threshold {
      filter          = "resource.type=\"gke_container\" AND metric.type=\"kubernetes.io/container/cpu/core_usage_time\""
      duration        = "600s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.9

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = var.alert_email != "" ? [google_monitoring_notification_channel.email[0].name] : []

  alert_strategy {
    auto_close = "3600s"
  }
}

# Conditional notification channel
resource "google_monitoring_notification_channel" "email" {
  count        = var.alert_email != "" ? 1 : 0
  display_name = "Development Email Notification"
  type         = "email"
  project      = var.project_id

  labels = {
    email_address = var.alert_email
  }
}

# Development service account
resource "google_service_account" "dev_service_account" {
  account_id   = "dev-service-account"
  display_name = "Development Service Account"
  project      = var.project_id
}

resource "google_project_iam_member" "dev_service_account_editor" {
  project = var.project_id
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.dev_service_account.email}"
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

output "dev_bucket_name" {
  description = "Development assets bucket name"
  value       = google_storage_bucket.dev_assets.name
}

output "dev_service_account_email" {
  description = "Development service account email"
  value       = google_service_account.dev_service_account.email
}
