# ☕ Go Coffee - Enterprise-Grade Infrastructure
# Multi-Cloud, Multi-Region, AI-Powered Coffee Ecosystem

terraform {
  required_version = ">= 1.5"
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
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  # Remote state configuration
  backend "gcs" {
    bucket = var.terraform_state_bucket
    prefix = "go-coffee/terraform.tfstate"
  }
}

# Local values for common configurations
locals {
  common_labels = {
    project     = "go-coffee"
    environment = var.environment
    managed_by  = "terraform"
    version     = var.app_version
  }

  # Multi-region configuration
  regions = var.multi_region_enabled ? var.regions : [var.region]

  # Service configuration
  services = {
    api_gateway         = { port = 8080, replicas = var.environment == "prod" ? 3 : 2 }
    order_service       = { port = 8081, replicas = var.environment == "prod" ? 3 : 2 }
    payment_service     = { port = 8082, replicas = var.environment == "prod" ? 3 : 2 }
    kitchen_service     = { port = 8083, replicas = var.environment == "prod" ? 2 : 1 }
    user_gateway        = { port = 8084, replicas = var.environment == "prod" ? 2 : 1 }
    security_gateway    = { port = 8085, replicas = var.environment == "prod" ? 3 : 2 }
    web_ui_backend      = { port = 8086, replicas = var.environment == "prod" ? 2 : 1 }
    ai_search           = { port = 8087, replicas = var.environment == "prod" ? 2 : 1 }
    bright_data_hub     = { port = 8088, replicas = var.environment == "prod" ? 2 : 1 }
    communication_hub   = { port = 8089, replicas = var.environment == "prod" ? 2 : 1 }
    enterprise_service  = { port = 8090, replicas = var.environment == "prod" ? 3 : 2 }
  }
}

# Random suffix for unique resource names
resource "random_id" "suffix" {
  byte_length = 4
}

# Enhanced Network Infrastructure
module "network" {
  source = "./modules/network"

  project_id   = var.project_id
  region       = var.region
  network_name = "${var.network_name}-${var.environment}"
  subnet_name  = "${var.subnet_name}-${var.environment}"
  subnet_cidr  = var.subnet_cidr
  environment  = var.environment
}

# Enhanced GKE Cluster with Multi-Zone Support
module "gke" {
  source = "./modules/gke"

  project_id        = var.project_id
  region            = var.region
  zone              = var.zone
  gke_cluster_name  = "${var.gke_cluster_name}-${var.environment}"
  network_name      = module.network.network_name
  subnet_name       = module.network.subnet_name
  gke_node_count    = var.gke_node_count
  gke_machine_type  = var.gke_machine_type
  gke_min_node_count = var.gke_min_node_count
  gke_max_node_count = var.gke_max_node_count
  environment       = var.environment

  depends_on = [module.network]
}

# Kafka Infrastructure with High Availability
module "kafka" {
  source = "./modules/kafka"

  project_id                 = var.project_id
  region                     = var.region
  kafka_instance_name        = "${var.kafka_instance_name}-${var.environment}"
  kafka_version              = var.kafka_version
  kafka_topic_name           = var.kafka_topic_name
  kafka_processed_topic_name = var.kafka_processed_topic_name
  environment                = var.environment

  depends_on = [module.gke]
}

# Comprehensive Monitoring Stack
module "monitoring" {
  source = "./modules/monitoring"

  project_id             = var.project_id
  region                 = var.region
  enable_monitoring      = var.enable_monitoring
  grafana_admin_password = var.grafana_admin_password
  environment            = var.environment

  depends_on = [module.gke]
}

# Redis Cluster for Caching and Session Management
module "redis" {
  source = "./modules/redis"

  project_id      = var.project_id
  region          = var.region
  instance_name   = "${var.redis_instance_name}-${var.environment}"
  memory_size_gb  = var.redis_memory_size_gb
  redis_version   = var.redis_version
  environment     = var.environment

  # High availability for production
  tier                    = var.environment == "prod" ? "STANDARD_HA" : "BASIC"
  replica_count          = var.environment == "prod" ? 1 : 0
  read_replicas_mode     = var.environment == "prod" ? "READ_REPLICAS_ENABLED" : "READ_REPLICAS_DISABLED"

  depends_on = [module.network]
}

# PostgreSQL Database for Persistent Storage
module "postgresql" {
  source = "./modules/postgresql"

  project_id        = var.project_id
  region            = var.region
  instance_name     = "${var.postgres_instance_name}-${var.environment}"
  database_version  = var.postgres_version
  tier              = var.postgres_tier
  environment       = var.environment

  # Database configuration
  databases = [
    "go_coffee_main",
    "go_coffee_orders",
    "go_coffee_payments",
    "go_coffee_analytics",
    "go_coffee_ai_agents"
  ]

  # High availability for production
  availability_type = var.environment == "prod" ? "REGIONAL" : "ZONAL"
  backup_enabled    = true

  depends_on = [module.network]
}

# Service Mesh with Istio
module "service_mesh" {
  source = "./modules/service-mesh"

  project_id      = var.project_id
  cluster_name    = module.gke.cluster_name
  cluster_location = var.region
  environment     = var.environment

  # Istio configuration
  enable_istio           = var.enable_service_mesh
  enable_mtls           = var.environment == "prod"
  enable_tracing        = true
  enable_monitoring     = true

  depends_on = [module.gke]
}

# Security and Compliance
module "security" {
  source = "./modules/security"

  project_id       = var.project_id
  region           = var.region
  cluster_name     = module.gke.cluster_name
  cluster_location = var.region
  environment      = var.environment

  # Security configuration
  enable_binary_authorization = var.environment == "prod"
  enable_pod_security_policy  = true
  enable_network_policy       = true
  enable_workload_identity    = true

  depends_on = [module.gke]
}
