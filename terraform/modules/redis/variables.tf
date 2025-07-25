# Redis Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for Redis instance"
  type        = string
}

variable "instance_name" {
  description = "Name of the Redis instance"
  type        = string
}

variable "memory_size_gb" {
  description = "Memory size in GB for Redis instance"
  type        = number
  default     = 4
  
  validation {
    condition     = var.memory_size_gb >= 1 && var.memory_size_gb <= 300
    error_message = "Memory size must be between 1 and 300 GB."
  }
}

variable "redis_version" {
  description = "Redis version"
  type        = string
  default     = "REDIS_7_0"
  
  validation {
    condition = contains([
      "REDIS_6_X",
      "REDIS_7_0",
      "REDIS_7_2"
    ], var.redis_version)
    error_message = "Redis version must be one of: REDIS_6_X, REDIS_7_0, REDIS_7_2."
  }
}

variable "tier" {
  description = "Redis service tier"
  type        = string
  default     = "BASIC"
  
  validation {
    condition = contains([
      "BASIC",
      "STANDARD_HA"
    ], var.tier)
    error_message = "Tier must be either BASIC or STANDARD_HA."
  }
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

variable "reserved_ip_range" {
  description = "Reserved IP range for Redis instance"
  type        = string
  default     = null
}

variable "replica_count" {
  description = "Number of replica nodes (only for STANDARD_HA tier)"
  type        = number
  default     = 0
  
  validation {
    condition     = var.replica_count >= 0 && var.replica_count <= 5
    error_message = "Replica count must be between 0 and 5."
  }
}

variable "read_replicas_mode" {
  description = "Read replicas mode"
  type        = string
  default     = "READ_REPLICAS_DISABLED"
  
  validation {
    condition = contains([
      "READ_REPLICAS_DISABLED",
      "READ_REPLICAS_ENABLED"
    ], var.read_replicas_mode)
    error_message = "Read replicas mode must be either READ_REPLICAS_DISABLED or READ_REPLICAS_ENABLED."
  }
}

variable "secondary_ip_range" {
  description = "Secondary IP range for read replicas"
  type        = string
  default     = null
}

variable "backup_region" {
  description = "Region for backup Redis instance"
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

variable "max_connections_threshold" {
  description = "Maximum connections threshold for alerting"
  type        = number
  default     = 1000
}
