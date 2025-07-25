# PostgreSQL Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for PostgreSQL instance"
  type        = string
}

variable "instance_name" {
  description = "Name of the PostgreSQL instance"
  type        = string
}

variable "database_version" {
  description = "PostgreSQL version"
  type        = string
  default     = "POSTGRES_15"
  
  validation {
    condition = contains([
      "POSTGRES_13",
      "POSTGRES_14",
      "POSTGRES_15",
      "POSTGRES_16"
    ], var.database_version)
    error_message = "Database version must be one of: POSTGRES_13, POSTGRES_14, POSTGRES_15, POSTGRES_16."
  }
}

variable "tier" {
  description = "Machine type for PostgreSQL instance"
  type        = string
  default     = "db-custom-2-8192"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "network_name" {
  description = "VPC network name"
  type        = string
  default     = "default"
}

variable "availability_type" {
  description = "Availability type for the instance"
  type        = string
  default     = "ZONAL"
  
  validation {
    condition = contains([
      "ZONAL",
      "REGIONAL"
    ], var.availability_type)
    error_message = "Availability type must be either ZONAL or REGIONAL."
  }
}

variable "disk_type" {
  description = "Disk type for the instance"
  type        = string
  default     = "PD_SSD"
  
  validation {
    condition = contains([
      "PD_SSD",
      "PD_HDD"
    ], var.disk_type)
    error_message = "Disk type must be either PD_SSD or PD_HDD."
  }
}

variable "disk_size" {
  description = "Disk size in GB"
  type        = number
  default     = 100
  
  validation {
    condition     = var.disk_size >= 10 && var.disk_size <= 65536
    error_message = "Disk size must be between 10 and 65536 GB."
  }
}

variable "disk_autoresize_limit" {
  description = "Maximum disk size for autoresize in GB"
  type        = number
  default     = 1000
}

variable "backup_enabled" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

variable "backup_location" {
  description = "Backup location"
  type        = string
  default     = null
}

variable "point_in_time_recovery_enabled" {
  description = "Enable point-in-time recovery"
  type        = bool
  default     = true
}

variable "transaction_log_retention_days" {
  description = "Transaction log retention days"
  type        = number
  default     = 7
  
  validation {
    condition     = var.transaction_log_retention_days >= 1 && var.transaction_log_retention_days <= 7
    error_message = "Transaction log retention days must be between 1 and 7."
  }
}

variable "retained_backups" {
  description = "Number of backups to retain"
  type        = number
  default     = 30
}

variable "authorized_networks" {
  description = "List of authorized networks"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "database_flags" {
  description = "Database flags for PostgreSQL optimization"
  type = list(object({
    name  = string
    value = string
  }))
  default = [
    {
      name  = "shared_preload_libraries"
      value = "pg_stat_statements"
    },
    {
      name  = "log_statement"
      value = "all"
    },
    {
      name  = "log_min_duration_statement"
      value = "1000"
    },
    {
      name  = "max_connections"
      value = "200"
    }
  ]
}

variable "databases" {
  description = "List of databases to create"
  type        = list(string)
  default     = ["go_coffee"]
}

variable "db_username" {
  description = "Database username"
  type        = string
  default     = "go_coffee_user"
}

variable "create_read_replica" {
  description = "Create read replica for production"
  type        = bool
  default     = true
}

variable "read_replica_region" {
  description = "Region for read replica"
  type        = string
  default     = ""
}

variable "read_replica_tier" {
  description = "Machine type for read replica"
  type        = string
  default     = ""
}

variable "enable_monitoring" {
  description = "Enable monitoring and alerting"
  type        = bool
  default     = true
}

variable "notification_channels" {
  description = "List of notification channels for alerts"
  type        = list(string)
  default     = []
}
