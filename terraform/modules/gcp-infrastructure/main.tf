# Go Coffee GCP Infrastructure Module
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Local variables
locals {
  project_id = var.project_id
  region     = var.region
  zone       = var.zone
  
  # Common labels
  common_labels = {
    project     = "go-coffee"
    environment = var.environment
    managed_by  = "terraform"
    team        = "platform"
  }
  
  # Network configuration
  network_name    = "${var.project_name}-vpc"
  subnet_name     = "${var.project_name}-subnet"
  cluster_name    = "${var.project_name}-gke"
  
  # Database configuration
  db_instance_name = "${var.project_name}-postgres"
  redis_instance_name = "${var.project_name}-redis"
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "compute.googleapis.com",
    "container.googleapis.com",
    "sqladmin.googleapis.com",
    "redis.googleapis.com",
    "monitoring.googleapis.com",
    "logging.googleapis.com",
    "cloudtrace.googleapis.com",
    "servicenetworking.googleapis.com",
    "secretmanager.googleapis.com",
    "artifactregistry.googleapis.com",
  ])
  
  project = local.project_id
  service = each.value
  
  disable_dependent_services = false
  disable_on_destroy        = false
}

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = local.network_name
  auto_create_subnetworks = false
  project                 = local.project_id
  
  depends_on = [google_project_service.required_apis]
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = local.subnet_name
  ip_cidr_range = var.subnet_cidr
  region        = local.region
  network       = google_compute_network.vpc.id
  project       = local.project_id
  
  # Enable private Google access
  private_ip_google_access = true
  
  # Secondary IP ranges for GKE
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = var.pods_cidr
  }
  
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = var.services_cidr
  }
}

# Cloud NAT for outbound internet access
resource "google_compute_router" "router" {
  name    = "${var.project_name}-router"
  region  = local.region
  network = google_compute_network.vpc.id
  project = local.project_id
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.project_name}-nat"
  router                            = google_compute_router.router.name
  region                            = local.region
  project                           = local.project_id
  nat_ip_allocate_option            = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  
  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = local.cluster_name
  location = local.region
  project  = local.project_id
  
  # Network configuration
  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
  
  # Remove default node pool
  remove_default_node_pool = true
  initial_node_count       = 1
  
  # Networking
  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }
  
  # Security
  enable_shielded_nodes = true
  
  # Workload Identity
  workload_identity_config {
    workload_pool = "${local.project_id}.svc.id.goog"
  }
  
  # Addons
  addons_config {
    http_load_balancing {
      disabled = false
    }
    horizontal_pod_autoscaling {
      disabled = false
    }
    network_policy_config {
      disabled = false
    }
    gcp_filestore_csi_driver_config {
      enabled = true
    }
  }
  
  # Network policy
  network_policy {
    enabled = true
  }
  
  # Monitoring and logging
  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
  
  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
  
  # Master auth
  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }
  
  # Private cluster configuration
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = var.master_cidr
  }
  
  # Maintenance window
  maintenance_policy {
    recurring_window {
      start_time = "2023-01-01T02:00:00Z"
      end_time   = "2023-01-01T06:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=SA"
    }
  }
  
  depends_on = [
    google_project_service.required_apis,
    google_compute_subnetwork.subnet,
  ]
}

# GKE Node Pool
resource "google_container_node_pool" "primary_nodes" {
  name       = "${local.cluster_name}-nodes"
  location   = local.region
  cluster    = google_container_cluster.primary.name
  project    = local.project_id
  
  # Autoscaling
  autoscaling {
    min_node_count = var.min_nodes
    max_node_count = var.max_nodes
  }
  
  # Node configuration
  node_config {
    preemptible  = var.use_preemptible_nodes
    machine_type = var.node_machine_type
    disk_size_gb = var.node_disk_size
    disk_type    = "pd-ssd"
    
    # Service account
    service_account = google_service_account.gke_nodes.email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    
    # Workload Identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    
    # Shielded instance
    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }
    
    # Labels
    labels = merge(local.common_labels, {
      node_pool = "primary"
    })
    
    # Taints for system workloads
    dynamic "taint" {
      for_each = var.node_taints
      content {
        key    = taint.value.key
        value  = taint.value.value
        effect = taint.value.effect
      }
    }
  }
  
  # Node management
  management {
    auto_repair  = true
    auto_upgrade = true
  }
  
  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

# Service Account for GKE nodes
resource "google_service_account" "gke_nodes" {
  account_id   = "${var.project_name}-gke-nodes"
  display_name = "GKE Nodes Service Account"
  project      = local.project_id
}

# IAM bindings for GKE nodes
resource "google_project_iam_member" "gke_nodes" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/monitoring.viewer",
    "roles/stackdriver.resourceMetadata.writer",
  ])
  
  project = local.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

# Cloud SQL PostgreSQL instance
resource "google_sql_database_instance" "postgres" {
  name             = local.db_instance_name
  database_version = var.postgres_version
  region           = local.region
  project          = local.project_id
  
  settings {
    tier              = var.postgres_tier
    availability_type = var.postgres_availability_type
    disk_size         = var.postgres_disk_size
    disk_type         = "PD_SSD"
    disk_autoresize   = true
    
    # Backup configuration
    backup_configuration {
      enabled                        = true
      start_time                     = "02:00"
      location                       = local.region
      point_in_time_recovery_enabled = true
      backup_retention_settings {
        retained_backups = 7
        retention_unit   = "COUNT"
      }
    }
    
    # IP configuration
    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc.id
      require_ssl     = true
    }
    
    # Maintenance window
    maintenance_window {
      day          = 7
      hour         = 2
      update_track = "stable"
    }
    
    # Database flags
    database_flags {
      name  = "log_statement"
      value = "all"
    }
    
    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"
    }
  }
  
  deletion_protection = var.enable_deletion_protection
  
  depends_on = [
    google_project_service.required_apis,
    google_service_networking_connection.private_vpc_connection,
  ]
}

# Private VPC connection for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "${var.project_name}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
  project       = local.project_id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Redis instance
resource "google_redis_instance" "cache" {
  name           = local.redis_instance_name
  tier           = var.redis_tier
  memory_size_gb = var.redis_memory_size
  region         = local.region
  project        = local.project_id
  
  authorized_network = google_compute_network.vpc.id
  
  redis_version     = var.redis_version
  display_name      = "Go Coffee Redis Cache"
  
  # High availability
  replica_count = var.redis_replica_count
  
  depends_on = [google_project_service.required_apis]
}
