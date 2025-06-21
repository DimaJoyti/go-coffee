# Staging Environment Variables

variable "project_id" {
  description = "The GCP project ID for staging"
  type        = string
}

variable "region" {
  description = "The GCP region for staging deployment"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone for staging deployment"
  type        = string
  default     = "us-central1-a"
}

variable "alert_email" {
  description = "Email address for staging alerts"
  type        = string
}

variable "domain_name" {
  description = "Staging domain name"
  type        = string
  default     = "staging.gocoffee.dev"
}

variable "ssl_certificate_domains" {
  description = "Domains for SSL certificates in staging"
  type        = list(string)
  default     = [
    "staging.gocoffee.dev",
    "*.staging.gocoffee.dev",
    "api.staging.gocoffee.dev",
    "app.staging.gocoffee.dev"
  ]
}

variable "backup_retention_days" {
  description = "Number of days to retain backups in staging"
  type        = number
  default     = 14
}

variable "monitoring_retention_days" {
  description = "Number of days to retain monitoring data"
  type        = number
  default     = 30
}

variable "enable_audit_logging" {
  description = "Enable audit logging for staging"
  type        = bool
  default     = true
}

variable "enable_vpc_flow_logs" {
  description = "Enable VPC flow logs for staging"
  type        = bool
  default     = true
}

variable "enable_binary_authorization" {
  description = "Enable binary authorization for staging"
  type        = bool
  default     = false
}

variable "enable_pod_security_policy" {
  description = "Enable pod security policy for staging"
  type        = bool
  default     = false
}

variable "maintenance_window" {
  description = "Maintenance window configuration"
  type = object({
    day         = number
    hour        = number
    update_track = string
  })
  default = {
    day         = 6  # Saturday
    hour        = 4  # 4 AM UTC
    update_track = "regular"
  }
}

variable "node_pools" {
  description = "Additional node pools for staging workloads"
  type = map(object({
    machine_type     = string
    min_count       = number
    max_count       = number
    disk_size_gb    = number
    disk_type       = string
    preemptible     = bool
    node_locations  = list(string)
    taints = list(object({
      key    = string
      value  = string
      effect = string
    }))
    labels = map(string)
  }))
  default = {
    "ai-workloads" = {
      machine_type    = "n1-standard-2"
      min_count      = 0
      max_count      = 3
      disk_size_gb   = 100
      disk_type      = "pd-standard"
      preemptible    = true
      node_locations = ["us-central1-a"]
      taints = [
        {
          key    = "workload-type"
          value  = "ai"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        workload-type = "ai"
        environment   = "staging"
      }
    }
  }
}

variable "database_flags" {
  description = "Database flags for staging PostgreSQL"
  type = map(string)
  default = {
    "log_statement"                = "ddl"
    "log_min_duration_statement"   = "2000"
    "log_connections"              = "on"
    "log_disconnections"           = "on"
    "max_connections"             = "100"
    "shared_buffers"              = "128MB"
    "effective_cache_size"        = "512MB"
    "maintenance_work_mem"        = "32MB"
    "checkpoint_completion_target" = "0.9"
    "wal_buffers"                 = "8MB"
    "default_statistics_target"   = "100"
    "random_page_cost"            = "1.1"
    "effective_io_concurrency"    = "200"
  }
}

variable "redis_config" {
  description = "Redis configuration for staging"
  type = map(string)
  default = {
    "maxmemory-policy"     = "allkeys-lru"
    "timeout"              = "300"
    "tcp-keepalive"        = "60"
    "maxclients"           = "5000"
    "save"                 = "900 1 300 10 60 1000"
    "stop-writes-on-bgsave-error" = "yes"
    "rdbcompression"       = "yes"
    "rdbchecksum"          = "yes"
    "appendonly"           = "yes"
    "appendfsync"          = "everysec"
  }
}

variable "network_security" {
  description = "Network security configuration"
  type = object({
    enable_private_google_access = bool
    enable_flow_logs            = bool
    flow_logs_sampling          = number
    enable_firewall_rules       = bool
    allowed_ingress_cidrs       = list(string)
    allowed_egress_cidrs        = list(string)
  })
  default = {
    enable_private_google_access = true
    enable_flow_logs            = true
    flow_logs_sampling          = 0.5
    enable_firewall_rules       = true
    allowed_ingress_cidrs       = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
    allowed_egress_cidrs        = ["0.0.0.0/0"]
  }
}

variable "monitoring_config" {
  description = "Monitoring and alerting configuration"
  type = object({
    enable_uptime_checks     = bool
    enable_log_based_metrics = bool
    notification_channels    = list(string)
    alert_policies = map(object({
      display_name = string
      conditions = list(object({
        display_name = string
        filter       = string
        comparison   = string
        threshold    = number
        duration     = string
      }))
    }))
  })
  default = {
    enable_uptime_checks     = true
    enable_log_based_metrics = true
    notification_channels    = []
    alert_policies = {
      "high-cpu" = {
        display_name = "High CPU Usage"
        conditions = [
          {
            display_name = "CPU usage above 75%"
            filter       = "resource.type=\"gke_container\""
            comparison   = "COMPARISON_GREATER_THAN"
            threshold    = 0.75
            duration     = "300s"
          }
        ]
      }
      "high-memory" = {
        display_name = "High Memory Usage"
        conditions = [
          {
            display_name = "Memory usage above 80%"
            filter       = "resource.type=\"gke_container\""
            comparison   = "COMPARISON_GREATER_THAN"
            threshold    = 0.8
            duration     = "300s"
          }
        ]
      }
    }
  }
}

variable "backup_config" {
  description = "Backup configuration for staging"
  type = object({
    database_backup_schedule    = string
    database_backup_retention   = number
    storage_backup_schedule     = string
    storage_backup_retention    = number
    cross_region_backup_enabled = bool
    backup_encryption_enabled   = bool
  })
  default = {
    database_backup_schedule    = "0 3 * * *"  # Daily at 3 AM
    database_backup_retention   = 14
    storage_backup_schedule     = "0 4 * * 0"  # Weekly on Sunday at 4 AM
    storage_backup_retention    = 4
    cross_region_backup_enabled = false
    backup_encryption_enabled   = true
  }
}

variable "cost_optimization" {
  description = "Cost optimization settings for staging"
  type = object({
    enable_preemptible_nodes     = bool
    enable_cluster_autoscaling   = bool
    enable_vertical_pod_autoscaling = bool
    enable_node_auto_provisioning  = bool
    auto_shutdown_schedule       = string
    auto_startup_schedule        = string
  })
  default = {
    enable_preemptible_nodes     = false
    enable_cluster_autoscaling   = true
    enable_vertical_pod_autoscaling = true
    enable_node_auto_provisioning  = false
    auto_shutdown_schedule       = "0 22 * * 1-5"  # Shutdown at 10 PM on weekdays
    auto_startup_schedule        = "0 8 * * 1-5"   # Startup at 8 AM on weekdays
  }
}

variable "feature_flags" {
  description = "Feature flags for staging environment"
  type = object({
    enable_workload_identity     = bool
    enable_binary_authorization  = bool
    enable_istio                = bool
    enable_knative              = bool
    enable_anthos_service_mesh   = bool
    enable_config_connector      = bool
  })
  default = {
    enable_workload_identity     = true
    enable_binary_authorization  = false
    enable_istio                = true
    enable_knative              = false
    enable_anthos_service_mesh   = false
    enable_config_connector      = false
  }
}
