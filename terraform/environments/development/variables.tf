# Development Environment Variables

variable "project_id" {
  description = "The GCP project ID for development"
  type        = string
}

variable "region" {
  description = "The GCP region for development deployment"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone for development deployment"
  type        = string
  default     = "us-central1-a"
}

variable "alert_email" {
  description = "Email address for development alerts (optional)"
  type        = string
  default     = ""
}

variable "domain_name" {
  description = "Development domain name"
  type        = string
  default     = "dev.gocoffee.local"
}

variable "enable_monitoring" {
  description = "Enable monitoring for development"
  type        = bool
  default     = true
}

variable "enable_logging" {
  description = "Enable logging for development"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Number of days to retain logs in development"
  type        = number
  default     = 7
}

variable "backup_retention_days" {
  description = "Number of days to retain backups in development"
  type        = number
  default     = 7
}

variable "enable_preemptible_nodes" {
  description = "Use preemptible nodes for cost savings in development"
  type        = bool
  default     = true
}

variable "node_machine_type" {
  description = "Machine type for GKE nodes in development"
  type        = string
  default     = "e2-standard-2"
}

variable "min_nodes" {
  description = "Minimum number of nodes in development"
  type        = number
  default     = 1
}

variable "max_nodes" {
  description = "Maximum number of nodes in development"
  type        = number
  default     = 5
}

variable "postgres_tier" {
  description = "PostgreSQL tier for development"
  type        = string
  default     = "db-f1-micro"
}

variable "postgres_disk_size" {
  description = "PostgreSQL disk size in GB for development"
  type        = number
  default     = 20
}

variable "redis_memory_size" {
  description = "Redis memory size in GB for development"
  type        = number
  default     = 1
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_network_policy" {
  description = "Enable network policy (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_pod_security_policy" {
  description = "Enable pod security policy (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_workload_identity" {
  description = "Enable workload identity (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_binary_authorization" {
  description = "Enable binary authorization (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_istio" {
  description = "Enable Istio service mesh (disabled for development)"
  type        = bool
  default     = false
}

variable "enable_cluster_autoscaling" {
  description = "Enable cluster autoscaling"
  type        = bool
  default     = true
}

variable "enable_vertical_pod_autoscaling" {
  description = "Enable vertical pod autoscaling (disabled for development)"
  type        = bool
  default     = false
}

variable "database_flags" {
  description = "Database flags for development PostgreSQL"
  type        = map(string)
  default = {
    "log_statement"              = "none"
    "log_min_duration_statement" = "-1"
    "max_connections"           = "50"
    "shared_buffers"            = "32MB"
    "effective_cache_size"      = "128MB"
  }
}

variable "redis_config" {
  description = "Redis configuration for development"
  type        = map(string)
  default = {
    "maxmemory-policy" = "allkeys-lru"
    "timeout"          = "300"
    "maxclients"       = "1000"
  }
}

variable "network_config" {
  description = "Network configuration for development"
  type = object({
    subnet_cidr   = string
    pods_cidr     = string
    services_cidr = string
    master_cidr   = string
  })
  default = {
    subnet_cidr   = "10.10.0.0/24"
    pods_cidr     = "10.11.0.0/16"
    services_cidr = "10.12.0.0/16"
    master_cidr   = "172.16.10.0/28"
  }
}

variable "maintenance_window" {
  description = "Maintenance window configuration for development"
  type = object({
    day         = number
    hour        = number
    update_track = string
  })
  default = {
    day         = 6  # Saturday
    hour        = 10 # 10 AM UTC
    update_track = "rapid"
  }
}

variable "allowed_ip_ranges" {
  description = "IP ranges allowed to access development environment"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Open for development
}

variable "enable_private_nodes" {
  description = "Enable private nodes (disabled for development ease)"
  type        = bool
  default     = false
}

variable "enable_private_endpoint" {
  description = "Enable private endpoint (disabled for development ease)"
  type        = bool
  default     = false
}

variable "master_authorized_networks" {
  description = "Authorized networks for master access"
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = [
    {
      cidr_block   = "0.0.0.0/0"
      display_name = "All networks (development only)"
    }
  ]
}

variable "resource_labels" {
  description = "Additional resource labels for development"
  type        = map(string)
  default = {
    cost-center = "development"
    auto-delete = "true"
    owner       = "dev-team"
  }
}
