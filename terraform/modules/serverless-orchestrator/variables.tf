# Variables for Multi-Cloud Serverless Orchestrator Module

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "go-coffee"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

# =============================================================================
# CLOUD PROVIDER ENABLEMENT
# =============================================================================

variable "enable_aws" {
  description = "Enable AWS serverless components"
  type        = bool
  default     = true
}

variable "enable_gcp" {
  description = "Enable Google Cloud serverless components"
  type        = bool
  default     = true
}

variable "enable_azure" {
  description = "Enable Azure serverless components"
  type        = bool
  default     = false
}

variable "enable_cross_cloud_routing" {
  description = "Enable cross-cloud event routing"
  type        = bool
  default     = true
}

# =============================================================================
# AWS CONFIGURATION
# =============================================================================

variable "aws_region" {
  description = "AWS region for serverless functions"
  type        = string
  default     = "us-east-1"
}

variable "aws_subnet_ids" {
  description = "AWS subnet IDs for Lambda VPC configuration"
  type        = list(string)
  default     = []
}

variable "aws_vpc_id" {
  description = "AWS VPC ID for security groups"
  type        = string
  default     = ""
}

# =============================================================================
# GOOGLE CLOUD CONFIGURATION
# =============================================================================

variable "gcp_project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "gcp_region" {
  description = "Google Cloud region for serverless functions"
  type        = string
  default     = "us-central1"
}

variable "gcp_vpc_connector" {
  description = "VPC connector for Google Cloud Functions"
  type        = string
  default     = ""
}

variable "max_instances" {
  description = "Maximum number of function instances"
  type        = number
  default     = 100
}

# =============================================================================
# AZURE CONFIGURATION
# =============================================================================

variable "azure_location" {
  description = "Azure location for serverless functions"
  type        = string
  default     = "East US"
}

variable "azure_resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = ""
}

# =============================================================================
# SHARED CONFIGURATION
# =============================================================================

variable "kafka_brokers" {
  description = "Kafka broker endpoints"
  type        = string
  default     = "localhost:9092"
}

variable "redis_url" {
  description = "Redis connection URL"
  type        = string
  default     = "redis://localhost:6379"
}

variable "database_url" {
  description = "Database connection URL"
  type        = string
  sensitive   = true
}

variable "log_level" {
  description = "Logging level for functions"
  type        = string
  default     = "info"
  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_level)
    error_message = "Log level must be one of: debug, info, warn, error."
  }
}

variable "log_retention_days" {
  description = "Log retention period in days"
  type        = number
  default     = 30
}

# =============================================================================
# FUNCTION CONFIGURATION
# =============================================================================

variable "function_timeout" {
  description = "Default function timeout in seconds"
  type        = number
  default     = 60
}

variable "function_memory" {
  description = "Default function memory in MB"
  type        = number
  default     = 512
}

variable "dead_letter_queue_enabled" {
  description = "Enable dead letter queue for failed function executions"
  type        = bool
  default     = true
}

# =============================================================================
# MONITORING CONFIGURATION
# =============================================================================

variable "monitoring_enabled" {
  description = "Enable comprehensive monitoring and alerting"
  type        = bool
  default     = true
}

variable "alert_email" {
  description = "Email address for alerts"
  type        = string
  default     = ""
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
  type        = string
  default     = ""
  sensitive   = true
}

# =============================================================================
# SECURITY CONFIGURATION
# =============================================================================

variable "encryption_at_rest" {
  description = "Enable encryption at rest for function storage"
  type        = bool
  default     = true
}

variable "encryption_in_transit" {
  description = "Enable encryption in transit"
  type        = bool
  default     = true
}

variable "allowed_origins" {
  description = "Allowed origins for CORS"
  type        = list(string)
  default     = ["*"]
}

# =============================================================================
# COST OPTIMIZATION
# =============================================================================

variable "cost_optimization_enabled" {
  description = "Enable cost optimization features"
  type        = bool
  default     = true
}

variable "auto_scaling_enabled" {
  description = "Enable auto-scaling based on demand"
  type        = bool
  default     = true
}

variable "reserved_concurrency" {
  description = "Reserved concurrency for critical functions"
  type        = map(number)
  default     = {}
}

# =============================================================================
# DISASTER RECOVERY
# =============================================================================

variable "backup_enabled" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

variable "multi_region_deployment" {
  description = "Enable multi-region deployment for disaster recovery"
  type        = bool
  default     = false
}

variable "failover_regions" {
  description = "List of failover regions for disaster recovery"
  type        = list(string)
  default     = []
}

# =============================================================================
# CUSTOM FUNCTION CONFIGURATIONS
# =============================================================================

variable "custom_functions" {
  description = "Custom function configurations"
  type = map(object({
    runtime     = string
    handler     = string
    timeout     = number
    memory      = number
    triggers    = list(string)
    environment = map(string)
  }))
  default = {}
}

variable "function_layers" {
  description = "Lambda layers for shared dependencies"
  type        = list(string)
  default     = []
}

# =============================================================================
# INTEGRATION CONFIGURATION
# =============================================================================

variable "api_gateway_integration" {
  description = "Enable API Gateway integration"
  type        = bool
  default     = true
}

variable "websocket_support" {
  description = "Enable WebSocket support for real-time communication"
  type        = bool
  default     = true
}

variable "event_sourcing_enabled" {
  description = "Enable event sourcing pattern"
  type        = bool
  default     = true
}

# =============================================================================
# DEVELOPMENT CONFIGURATION
# =============================================================================

variable "local_development" {
  description = "Enable local development features"
  type        = bool
  default     = false
}

variable "debug_mode" {
  description = "Enable debug mode for development"
  type        = bool
  default     = false
}

variable "hot_reload" {
  description = "Enable hot reload for development"
  type        = bool
  default     = false
}
