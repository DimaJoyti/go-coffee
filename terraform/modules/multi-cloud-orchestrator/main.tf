# Multi-Cloud Orchestrator Module
# Manages infrastructure across AWS, GCP, and Azure with advanced automation

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common tags
  common_tags = {
    Project              = var.project_name
    Environment          = var.environment
    ManagedBy           = "terraform"
    Component           = "multi-cloud-orchestrator"
    Team                = "platform"
    CostCenter          = var.cost_center
    Owner               = var.owner
    CreatedAt           = timestamp()
    TerraformWorkspace  = terraform.workspace
  }
  
  # Cloud provider configurations
  enabled_providers = {
    aws   = var.enable_aws
    gcp   = var.enable_gcp
    azure = var.enable_azure
  }
  
  # Multi-cloud regions mapping
  regions = {
    aws = {
      primary   = var.aws_primary_region
      secondary = var.aws_secondary_region
      tertiary  = var.aws_tertiary_region
    }
    gcp = {
      primary   = var.gcp_primary_region
      secondary = var.gcp_secondary_region
      tertiary  = var.gcp_tertiary_region
    }
    azure = {
      primary   = var.azure_primary_region
      secondary = var.azure_secondary_region
      tertiary  = var.azure_tertiary_region
    }
  }
}

# =============================================================================
# MULTI-CLOUD INFRASTRUCTURE ORCHESTRATION
# =============================================================================

# AWS Infrastructure Module
module "aws_infrastructure" {
  count = var.enable_aws ? 1 : 0
  
  source = "../aws-infrastructure"
  
  project_name    = var.project_name
  environment     = var.environment
  region          = local.regions.aws.primary
  cost_center     = var.cost_center
  
  # VPC Configuration
  vpc_cidr                = var.aws_vpc_cidr
  availability_zones_count = var.aws_availability_zones_count
  
  # EKS Configuration
  kubernetes_version      = var.kubernetes_version
  eks_public_access      = var.eks_public_access
  eks_public_access_cidrs = var.eks_public_access_cidrs
  eks_cluster_log_types  = var.eks_cluster_log_types
  
  # Node Group Configuration
  node_groups = var.aws_node_groups
  
  # Database Configuration
  rds_instance_class     = var.aws_rds_instance_class
  rds_allocated_storage  = var.aws_rds_allocated_storage
  rds_engine_version     = var.aws_rds_engine_version
  
  # Redis Configuration
  redis_node_type        = var.aws_redis_node_type
  redis_num_cache_nodes  = var.aws_redis_num_cache_nodes
  redis_engine_version   = var.aws_redis_engine_version
  
  # Monitoring
  monitoring_enabled     = var.monitoring_enabled
  log_retention_days     = var.log_retention_days
  
  # Security
  encryption_at_rest     = var.encryption_at_rest
  encryption_in_transit  = var.encryption_in_transit
  
  tags = local.common_tags
}

# GCP Infrastructure Module
module "gcp_infrastructure" {
  count = var.enable_gcp ? 1 : 0
  
  source = "../gcp-infrastructure"
  
  project_id      = var.gcp_project_id
  project_name    = var.project_name
  environment     = var.environment
  region          = local.regions.gcp.primary
  cost_center     = var.cost_center
  
  # Network Configuration
  vpc_cidr_range         = var.gcp_vpc_cidr_range
  subnet_cidr_ranges     = var.gcp_subnet_cidr_ranges
  
  # GKE Configuration
  kubernetes_version     = var.kubernetes_version
  gke_node_count        = var.gcp_gke_node_count
  gke_machine_type      = var.gcp_gke_machine_type
  gke_disk_size_gb      = var.gcp_gke_disk_size_gb
  
  # Database Configuration
  sql_tier              = var.gcp_sql_tier
  sql_disk_size         = var.gcp_sql_disk_size
  sql_database_version  = var.gcp_sql_database_version
  
  # Redis Configuration
  redis_memory_size_gb  = var.gcp_redis_memory_size_gb
  redis_version         = var.gcp_redis_version
  
  # Monitoring
  monitoring_enabled    = var.monitoring_enabled
  log_retention_days    = var.log_retention_days
  
  # Security
  encryption_at_rest    = var.encryption_at_rest
  encryption_in_transit = var.encryption_in_transit
  
  labels = local.common_tags
}

# Azure Infrastructure Module
module "azure_infrastructure" {
  count = var.enable_azure ? 1 : 0
  
  source = "../azure-infrastructure"
  
  project_name         = var.project_name
  environment          = var.environment
  location             = local.regions.azure.primary
  cost_center          = var.cost_center
  
  # Network Configuration
  vnet_address_space   = var.azure_vnet_address_space
  subnet_address_prefixes = var.azure_subnet_address_prefixes
  
  # AKS Configuration
  kubernetes_version   = var.kubernetes_version
  aks_node_count      = var.azure_aks_node_count
  aks_vm_size         = var.azure_aks_vm_size
  aks_os_disk_size_gb = var.azure_aks_os_disk_size_gb
  
  # Database Configuration
  postgresql_sku_name  = var.azure_postgresql_sku_name
  postgresql_storage_mb = var.azure_postgresql_storage_mb
  postgresql_version   = var.azure_postgresql_version
  
  # Redis Configuration
  redis_capacity       = var.azure_redis_capacity
  redis_family         = var.azure_redis_family
  redis_sku_name       = var.azure_redis_sku_name
  
  # Monitoring
  monitoring_enabled   = var.monitoring_enabled
  log_retention_days   = var.log_retention_days
  
  # Security
  encryption_at_rest   = var.encryption_at_rest
  encryption_in_transit = var.encryption_in_transit
  
  tags = local.common_tags
}

# =============================================================================
# CROSS-CLOUD NETWORKING
# =============================================================================

# VPN Connections between clouds (when multiple clouds are enabled)
resource "random_password" "vpn_shared_key" {
  count = var.enable_cross_cloud_networking && length([for k, v in local.enabled_providers : k if v]) > 1 ? 1 : 0
  
  length  = 32
  special = true
}

# AWS to GCP VPN Connection
module "aws_gcp_vpn" {
  count = var.enable_aws && var.enable_gcp && var.enable_cross_cloud_networking ? 1 : 0
  
  source = "../cross-cloud-vpn"
  
  project_name = var.project_name
  environment  = var.environment
  
  # AWS Configuration
  aws_vpc_id           = module.aws_infrastructure[0].vpc_id
  aws_route_table_ids  = module.aws_infrastructure[0].private_route_table_ids
  aws_cidr_block       = var.aws_vpc_cidr
  
  # GCP Configuration
  gcp_project_id       = var.gcp_project_id
  gcp_network_name     = module.gcp_infrastructure[0].vpc_network_name
  gcp_region           = local.regions.gcp.primary
  gcp_cidr_range       = var.gcp_vpc_cidr_range
  
  # VPN Configuration
  shared_secret        = random_password.vpn_shared_key[0].result
  
  tags = local.common_tags
}

# =============================================================================
# GLOBAL LOAD BALANCING
# =============================================================================

# Global Load Balancer Configuration
resource "google_compute_global_address" "global_lb_ip" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name         = "${local.name_prefix}-global-lb-ip"
  address_type = "EXTERNAL"
  
  labels = local.common_tags
}

# Global Load Balancer
resource "google_compute_global_forwarding_rule" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name                  = "${local.name_prefix}-global-lb"
  target                = google_compute_target_https_proxy.global_lb[0].id
  port_range           = "443"
  ip_address           = google_compute_global_address.global_lb_ip[0].address
  load_balancing_scheme = "EXTERNAL"
  
  labels = local.common_tags
}

# HTTPS Proxy for Global Load Balancer
resource "google_compute_target_https_proxy" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name             = "${local.name_prefix}-global-lb-proxy"
  url_map          = google_compute_url_map.global_lb[0].id
  ssl_certificates = [google_compute_managed_ssl_certificate.global_lb[0].id]
}

# URL Map for Global Load Balancer
resource "google_compute_url_map" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name            = "${local.name_prefix}-global-lb-url-map"
  default_service = google_compute_backend_service.global_lb[0].id
  
  # Route traffic based on geographic location
  dynamic "host_rule" {
    for_each = var.global_lb_host_rules
    content {
      hosts        = host_rule.value.hosts
      path_matcher = host_rule.value.path_matcher
    }
  }
  
  dynamic "path_matcher" {
    for_each = var.global_lb_path_matchers
    content {
      name            = path_matcher.value.name
      default_service = path_matcher.value.default_service
      
      dynamic "path_rule" {
        for_each = path_matcher.value.path_rules
        content {
          paths   = path_rule.value.paths
          service = path_rule.value.service
        }
      }
    }
  }
}

# Backend Service for Global Load Balancer
resource "google_compute_backend_service" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name                  = "${local.name_prefix}-global-lb-backend"
  protocol              = "HTTP"
  port_name             = "http"
  load_balancing_scheme = "EXTERNAL"
  timeout_sec           = 30
  
  health_checks = [google_compute_health_check.global_lb[0].id]
  
  # Add backends from different regions/clouds
  dynamic "backend" {
    for_each = var.global_lb_backends
    content {
      group           = backend.value.group
      balancing_mode  = backend.value.balancing_mode
      capacity_scaler = backend.value.capacity_scaler
    }
  }
  
  # CDN Configuration
  cdn_policy {
    cache_mode                   = "CACHE_ALL_STATIC"
    default_ttl                 = 3600
    max_ttl                     = 86400
    negative_caching            = true
    serve_while_stale           = 86400
    signed_url_cache_max_age_sec = 7200
  }
}

# Health Check for Global Load Balancer
resource "google_compute_health_check" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name                = "${local.name_prefix}-global-lb-health-check"
  check_interval_sec  = 30
  timeout_sec         = 10
  healthy_threshold   = 2
  unhealthy_threshold = 3
  
  http_health_check {
    port         = 80
    request_path = "/health"
  }
}

# Managed SSL Certificate
resource "google_compute_managed_ssl_certificate" "global_lb" {
  count = var.enable_global_load_balancing && var.enable_gcp ? 1 : 0
  
  name = "${local.name_prefix}-global-lb-ssl-cert"
  
  managed {
    domains = var.ssl_certificate_domains
  }
}

# =============================================================================
# DISASTER RECOVERY CONFIGURATION
# =============================================================================

# Cross-region backup configuration
resource "time_rotating" "backup_rotation" {
  count = var.enable_disaster_recovery ? 1 : 0
  
  rotation_hours = var.backup_rotation_hours
}

# Backup policies for each cloud provider
locals {
  backup_policies = {
    aws = {
      enabled = var.enable_aws && var.enable_disaster_recovery
      retention_days = var.backup_retention_days
      backup_window = var.aws_backup_window
      regions = [local.regions.aws.primary, local.regions.aws.secondary]
    }
    gcp = {
      enabled = var.enable_gcp && var.enable_disaster_recovery
      retention_days = var.backup_retention_days
      backup_window = var.gcp_backup_window
      regions = [local.regions.gcp.primary, local.regions.gcp.secondary]
    }
    azure = {
      enabled = var.enable_azure && var.enable_disaster_recovery
      retention_days = var.backup_retention_days
      backup_window = var.azure_backup_window
      regions = [local.regions.azure.primary, local.regions.azure.secondary]
    }
  }
}

# =============================================================================
# MONITORING AND OBSERVABILITY
# =============================================================================

# Multi-cloud monitoring configuration
module "monitoring_stack" {
  count = var.monitoring_enabled ? 1 : 0
  
  source = "../monitoring"
  
  project_name = var.project_name
  environment  = var.environment
  
  # Cloud provider configurations
  enable_aws   = var.enable_aws
  enable_gcp   = var.enable_gcp
  enable_azure = var.enable_azure
  
  # Monitoring configuration
  prometheus_retention_days = var.prometheus_retention_days
  grafana_admin_password   = var.grafana_admin_password
  alert_manager_config     = var.alert_manager_config
  
  # Notification channels
  slack_webhook_url        = var.slack_webhook_url
  pagerduty_service_key   = var.pagerduty_service_key
  email_notifications     = var.email_notifications
  
  tags = local.common_tags
}

# =============================================================================
# SECURITY AND COMPLIANCE
# =============================================================================

# Security scanning and compliance
module "security_stack" {
  count = var.security_scanning_enabled ? 1 : 0
  
  source = "../security"
  
  project_name = var.project_name
  environment  = var.environment
  
  # Cloud provider configurations
  enable_aws   = var.enable_aws
  enable_gcp   = var.enable_gcp
  enable_azure = var.enable_azure
  
  # Security configuration
  enable_vulnerability_scanning = var.enable_vulnerability_scanning
  enable_compliance_monitoring  = var.enable_compliance_monitoring
  enable_threat_detection      = var.enable_threat_detection
  
  # Compliance frameworks
  compliance_frameworks = var.compliance_frameworks
  
  tags = local.common_tags
}
