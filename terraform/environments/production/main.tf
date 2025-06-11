# Go Coffee Production Environment

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
  environment = "production"
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
  subnet_cidr   = "10.0.0.0/24"
  pods_cidr     = "10.1.0.0/16"
  services_cidr = "10.2.0.0/16"
  master_cidr   = "172.16.0.0/28"

  # GKE Configuration - Production sizing
  min_nodes           = 3
  max_nodes           = 50
  node_machine_type   = "e2-standard-8"
  node_disk_size      = 200
  use_preemptible_nodes = false

  # Node taints for production workloads
  node_taints = [
    {
      key    = "workload-type"
      value  = "production"
      effect = "NO_SCHEDULE"
    }
  ]

  # Database Configuration - Production sizing
  postgres_version           = "POSTGRES_15"
  postgres_tier             = "db-custom-8-16384"
  postgres_availability_type = "REGIONAL"
  postgres_disk_size        = 500

  # Redis Configuration - Production sizing
  redis_tier         = "STANDARD_HA"
  redis_memory_size  = 16
  redis_version      = "REDIS_7_0"
  redis_replica_count = 2

  # Security Configuration
  enable_deletion_protection = true
  enable_network_policy     = true
  enable_pod_security_policy = true

  # Monitoring Configuration
  enable_monitoring = true
  enable_logging   = true
  log_retention_days = 90

  # Backup Configuration
  backup_retention_days         = 30
  enable_point_in_time_recovery = true

  # Cost Optimization
  enable_cluster_autoscaling        = true
  enable_vertical_pod_autoscaling   = true
  enable_node_auto_provisioning     = false

  # Feature Flags
  enable_workload_identity     = true
  enable_binary_authorization  = true
  enable_istio                = true
  enable_knative              = false

  # Labels
  labels = local.common_labels
}

# Additional production-specific resources

# Cloud Armor Security Policy
resource "google_compute_security_policy" "go_coffee_security_policy" {
  name        = "go-coffee-security-policy"
  description = "Security policy for Go Coffee production environment"
  project     = var.project_id

  # Default rule
  rule {
    action   = "allow"
    priority = "2147483647"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    description = "Default allow rule"
  }

  # Rate limiting rule
  rule {
    action   = "rate_based_ban"
    priority = "1000"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    rate_limit_options {
      conform_action = "allow"
      exceed_action  = "deny(429)"
      enforce_on_key = "IP"
      rate_limit_threshold {
        count        = 100
        interval_sec = 60
      }
      ban_duration_sec = 600
    }
    description = "Rate limiting rule"
  }

  # Block known bad IPs
  rule {
    action   = "deny(403)"
    priority = "500"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = var.blocked_ip_ranges
      }
    }
    description = "Block known bad IPs"
  }
}

# Cloud CDN for static assets
resource "google_compute_backend_bucket" "static_assets" {
  name        = "go-coffee-static-assets"
  description = "Backend bucket for static assets"
  bucket_name = google_storage_bucket.static_assets.name
  enable_cdn  = true
  project     = var.project_id

  cdn_policy {
    cache_mode                   = "CACHE_ALL_STATIC"
    default_ttl                 = 3600
    max_ttl                     = 86400
    negative_caching            = true
    serve_while_stale           = 86400
    signed_url_cache_max_age_sec = 7200
  }
}

# Storage bucket for static assets
resource "google_storage_bucket" "static_assets" {
  name          = "${var.project_id}-static-assets"
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
    origin          = ["https://*.gocoffee.dev"]
    method          = ["GET", "HEAD"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Cloud SQL backup bucket
resource "google_storage_bucket" "sql_backups" {
  name          = "${var.project_id}-sql-backups"
  location      = var.region
  force_destroy = false
  project       = var.project_id

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }

  encryption {
    default_kms_key_name = google_kms_crypto_key.backup_key.id
  }
}

# KMS key for encryption
resource "google_kms_key_ring" "go_coffee_keyring" {
  name     = "go-coffee-keyring"
  location = var.region
  project  = var.project_id
}

resource "google_kms_crypto_key" "backup_key" {
  name     = "backup-key"
  key_ring = google_kms_key_ring.go_coffee_keyring.id
  purpose  = "ENCRYPT_DECRYPT"

  rotation_period = "7776000s" # 90 days

  lifecycle {
    prevent_destroy = true
  }
}

# Cloud Monitoring alerting policy
resource "google_monitoring_alert_policy" "high_cpu_usage" {
  display_name = "High CPU Usage"
  combiner     = "OR"
  project      = var.project_id

  conditions {
    display_name = "CPU usage above 80%"
    condition_threshold {
      filter          = "resource.type=\"gke_container\" AND metric.type=\"kubernetes.io/container/cpu/core_usage_time\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

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

# Notification channel
resource "google_monitoring_notification_channel" "email" {
  display_name = "Email Notification"
  type         = "email"
  project      = var.project_id

  labels = {
    email_address = var.alert_email
  }
}

# Cloud Scheduler for maintenance tasks
resource "google_cloud_scheduler_job" "database_backup" {
  name             = "database-backup"
  description      = "Daily database backup"
  schedule         = "0 2 * * *"
  time_zone        = "UTC"
  attempt_deadline = "320s"
  project          = var.project_id
  region           = var.region

  retry_config {
    retry_count = 3
  }

  http_target {
    http_method = "POST"
    uri         = "https://sqladmin.googleapis.com/sql/v1beta4/projects/${var.project_id}/instances/${module.gcp_infrastructure.postgres_instance_name}/export"
    
    headers = {
      "Content-Type" = "application/json"
    }

    oauth_token {
      service_account_email = google_service_account.backup_service_account.email
    }

    body = base64encode(jsonencode({
      exportContext = {
        fileType = "SQL"
        uri      = "gs://${google_storage_bucket.sql_backups.name}/backup-$(date +%Y%m%d-%H%M%S).sql"
      }
    }))
  }
}

# Service account for backup operations
resource "google_service_account" "backup_service_account" {
  account_id   = "backup-service-account"
  display_name = "Backup Service Account"
  project      = var.project_id
}

resource "google_project_iam_member" "backup_service_account_sql_admin" {
  project = var.project_id
  role    = "roles/cloudsql.admin"
  member  = "serviceAccount:${google_service_account.backup_service_account.email}"
}

resource "google_storage_bucket_iam_member" "backup_service_account_storage_admin" {
  bucket = google_storage_bucket.sql_backups.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.backup_service_account.email}"
}
