# Go Coffee GCP Infrastructure Variables

# Project Configuration
variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "project_name" {
  description = "The project name used for resource naming"
  type        = string
  default     = "go-coffee"
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"
  
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be one of: development, staging, production."
  }
}

# Regional Configuration
variable "region" {
  description = "The GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone"
  type        = string
  default     = "us-central1-a"
}

# Network Configuration
variable "subnet_cidr" {
  description = "CIDR block for the subnet"
  type        = string
  default     = "10.0.0.0/24"
}

variable "pods_cidr" {
  description = "CIDR block for GKE pods"
  type        = string
  default     = "10.1.0.0/16"
}

variable "services_cidr" {
  description = "CIDR block for GKE services"
  type        = string
  default     = "10.2.0.0/16"
}

variable "master_cidr" {
  description = "CIDR block for GKE master nodes"
  type        = string
  default     = "172.16.0.0/28"
}

# GKE Configuration
variable "min_nodes" {
  description = "Minimum number of nodes in the node pool"
  type        = number
  default     = 1
}

variable "max_nodes" {
  description = "Maximum number of nodes in the node pool"
  type        = number
  default     = 10
}

variable "node_machine_type" {
  description = "Machine type for GKE nodes"
  type        = string
  default     = "e2-standard-4"
}

variable "node_disk_size" {
  description = "Disk size for GKE nodes in GB"
  type        = number
  default     = 100
}

variable "use_preemptible_nodes" {
  description = "Use preemptible nodes for cost savings"
  type        = bool
  default     = false
}

variable "node_taints" {
  description = "List of taints to apply to nodes"
  type = list(object({
    key    = string
    value  = string
    effect = string
  }))
  default = []
}

# Database Configuration
variable "postgres_version" {
  description = "PostgreSQL version"
  type        = string
  default     = "POSTGRES_15"
}

variable "postgres_tier" {
  description = "PostgreSQL instance tier"
  type        = string
  default     = "db-custom-2-4096"
}

variable "postgres_availability_type" {
  description = "PostgreSQL availability type"
  type        = string
  default     = "REGIONAL"
  
  validation {
    condition     = contains(["ZONAL", "REGIONAL"], var.postgres_availability_type)
    error_message = "PostgreSQL availability type must be ZONAL or REGIONAL."
  }
}

variable "postgres_disk_size" {
  description = "PostgreSQL disk size in GB"
  type        = number
  default     = 100
}

# Redis Configuration
variable "redis_tier" {
  description = "Redis tier"
  type        = string
  default     = "STANDARD_HA"
  
  validation {
    condition     = contains(["BASIC", "STANDARD_HA"], var.redis_tier)
    error_message = "Redis tier must be BASIC or STANDARD_HA."
  }
}

variable "redis_memory_size" {
  description = "Redis memory size in GB"
  type        = number
  default     = 4
}

variable "redis_version" {
  description = "Redis version"
  type        = string
  default     = "REDIS_7_0"
}

variable "redis_replica_count" {
  description = "Number of Redis replicas"
  type        = number
  default     = 1
}

# Security Configuration
variable "enable_deletion_protection" {
  description = "Enable deletion protection for critical resources"
  type        = bool
  default     = true
}

variable "enable_network_policy" {
  description = "Enable Kubernetes network policy"
  type        = bool
  default     = true
}

variable "enable_pod_security_policy" {
  description = "Enable Pod Security Policy"
  type        = bool
  default     = true
}

# Monitoring Configuration
variable "enable_monitoring" {
  description = "Enable Google Cloud Monitoring"
  type        = bool
  default     = true
}

variable "enable_logging" {
  description = "Enable Google Cloud Logging"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Log retention period in days"
  type        = number
  default     = 30
}

# Backup Configuration
variable "backup_retention_days" {
  description = "Backup retention period in days"
  type        = number
  default     = 7
}

variable "enable_point_in_time_recovery" {
  description = "Enable point-in-time recovery for databases"
  type        = bool
  default     = true
}

# Cost Optimization
variable "enable_cluster_autoscaling" {
  description = "Enable cluster autoscaling"
  type        = bool
  default     = true
}

variable "enable_vertical_pod_autoscaling" {
  description = "Enable vertical pod autoscaling"
  type        = bool
  default     = true
}

variable "enable_node_auto_provisioning" {
  description = "Enable node auto-provisioning"
  type        = bool
  default     = false
}

# Additional Configuration
variable "labels" {
  description = "Additional labels to apply to resources"
  type        = map(string)
  default     = {}
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = list(string)
  default     = []
}

# Feature Flags
variable "enable_workload_identity" {
  description = "Enable Workload Identity"
  type        = bool
  default     = true
}

variable "enable_binary_authorization" {
  description = "Enable Binary Authorization"
  type        = bool
  default     = false
}

variable "enable_istio" {
  description = "Enable Istio service mesh"
  type        = bool
  default     = false
}

variable "enable_knative" {
  description = "Enable Knative serverless platform"
  type        = bool
  default     = false
}
