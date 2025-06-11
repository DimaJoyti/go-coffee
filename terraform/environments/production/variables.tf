# Production Environment Variables

variable "project_id" {
  description = "The GCP project ID for production"
  type        = string
}

variable "region" {
  description = "The GCP region for production deployment"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone for production deployment"
  type        = string
  default     = "us-central1-a"
}

variable "alert_email" {
  description = "Email address for production alerts"
  type        = string
}

variable "blocked_ip_ranges" {
  description = "List of IP ranges to block in production"
  type        = list(string)
  default     = []
}

variable "domain_name" {
  description = "Production domain name"
  type        = string
  default     = "gocoffee.dev"
}

variable "ssl_certificate_domains" {
  description = "Domains for SSL certificates"
  type        = list(string)
  default     = [
    "gocoffee.dev",
    "*.gocoffee.dev",
    "api.gocoffee.dev",
    "app.gocoffee.dev"
  ]
}

variable "backup_retention_days" {
  description = "Number of days to retain backups in production"
  type        = number
  default     = 90
}

variable "monitoring_retention_days" {
  description = "Number of days to retain monitoring data"
  type        = number
  default     = 180
}

variable "enable_audit_logging" {
  description = "Enable audit logging for production"
  type        = bool
  default     = true
}

variable "enable_vpc_flow_logs" {
  description = "Enable VPC flow logs for production"
  type        = bool
  default     = true
}

variable "enable_binary_authorization" {
  description = "Enable binary authorization for production"
  type        = bool
  default     = true
}

variable "enable_pod_security_policy" {
  description = "Enable pod security policy for production"
  type        = bool
  default     = true
}

variable "maintenance_window" {
  description = "Maintenance window configuration"
  type = object({
    day         = number
    hour        = number
    update_track = string
  })
  default = {
    day         = 7  # Sunday
    hour        = 2  # 2 AM UTC
    update_track = "stable"
  }
}

variable "node_pools" {
  description = "Additional node pools for production workloads"
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
      machine_type    = "n1-standard-4"
      min_count      = 0
      max_count      = 10
      disk_size_gb   = 200
      disk_type      = "pd-ssd"
      preemptible    = false
      node_locations = ["us-central1-a", "us-central1-b"]
      taints = [
        {
          key    = "workload-type"
          value  = "ai"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        workload-type = "ai"
        gpu-enabled   = "true"
      }
    }
    "web3-workloads" = {
      machine_type    = "e2-standard-4"
      min_count      = 1
      max_count      = 5
      disk_size_gb   = 100
      disk_type      = "pd-standard"
      preemptible    = false
      node_locations = ["us-central1-a", "us-central1-c"]
      taints = [
        {
          key    = "workload-type"
          value  = "web3"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        workload-type = "web3"
        blockchain    = "enabled"
      }
    }
  }
}

variable "database_flags" {
  description = "Database flags for production PostgreSQL"
  type = map(string)
  default = {
    "log_statement"                = "all"
    "log_min_duration_statement"   = "1000"
    "log_connections"              = "on"
    "log_disconnections"           = "on"
    "log_lock_waits"              = "on"
    "log_temp_files"              = "0"
    "track_activity_query_size"    = "2048"
    "track_io_timing"             = "on"
    "shared_preload_libraries"     = "pg_stat_statements"
    "max_connections"             = "200"
    "shared_buffers"              = "256MB"
    "effective_cache_size"        = "1GB"
    "maintenance_work_mem"        = "64MB"
    "checkpoint_completion_target" = "0.9"
    "wal_buffers"                 = "16MB"
    "default_statistics_target"   = "100"
    "random_page_cost"            = "1.1"
    "effective_io_concurrency"    = "200"
  }
}

variable "redis_config" {
  description = "Redis configuration for production"
  type = map(string)
  default = {
    "maxmemory-policy"     = "allkeys-lru"
    "timeout"              = "300"
    "tcp-keepalive"        = "60"
    "maxclients"           = "10000"
    "save"                 = "900 1 300 10 60 10000"
    "stop-writes-on-bgsave-error" = "yes"
    "rdbcompression"       = "yes"
    "rdbchecksum"          = "yes"
    "appendonly"           = "yes"
    "appendfsync"          = "everysec"
    "no-appendfsync-on-rewrite" = "no"
    "auto-aof-rewrite-percentage" = "100"
    "auto-aof-rewrite-min-size" = "64mb"
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
            display_name = "CPU usage above 80%"
            filter       = "resource.type=\"gke_container\""
            comparison   = "COMPARISON_GREATER_THAN"
            threshold    = 0.8
            duration     = "300s"
          }
        ]
      }
      "high-memory" = {
        display_name = "High Memory Usage"
        conditions = [
          {
            display_name = "Memory usage above 85%"
            filter       = "resource.type=\"gke_container\""
            comparison   = "COMPARISON_GREATER_THAN"
            threshold    = 0.85
            duration     = "300s"
          }
        ]
      }
      "pod-crash-loop" = {
        display_name = "Pod Crash Loop"
        conditions = [
          {
            display_name = "Pod restart count high"
            filter       = "resource.type=\"k8s_container\""
            comparison   = "COMPARISON_GREATER_THAN"
            threshold    = 5
            duration     = "300s"
          }
        ]
      }
    }
  }
}

variable "backup_config" {
  description = "Backup configuration for production"
  type = object({
    database_backup_schedule    = string
    database_backup_retention   = number
    storage_backup_schedule     = string
    storage_backup_retention    = number
    cross_region_backup_enabled = bool
    backup_encryption_enabled   = bool
  })
  default = {
    database_backup_schedule    = "0 2 * * *"  # Daily at 2 AM
    database_backup_retention   = 30
    storage_backup_schedule     = "0 3 * * 0"  # Weekly on Sunday at 3 AM
    storage_backup_retention    = 12
    cross_region_backup_enabled = true
    backup_encryption_enabled   = true
  }
}

variable "disaster_recovery" {
  description = "Disaster recovery configuration"
  type = object({
    enable_multi_region        = bool
    secondary_region          = string
    rpo_hours                 = number
    rto_hours                 = number
    enable_cross_region_backup = bool
    enable_geo_redundancy     = bool
  })
  default = {
    enable_multi_region        = true
    secondary_region          = "us-east1"
    rpo_hours                 = 4
    rto_hours                 = 2
    enable_cross_region_backup = true
    enable_geo_redundancy     = true
  }
}

variable "compliance" {
  description = "Compliance and governance configuration"
  type = object({
    enable_audit_logs          = bool
    enable_access_transparency = bool
    enable_vpc_sc             = bool
    data_classification       = string
    retention_policy_days     = number
  })
  default = {
    enable_audit_logs          = true
    enable_access_transparency = true
    enable_vpc_sc             = false
    data_classification       = "confidential"
    retention_policy_days     = 2555  # 7 years
  }
}
