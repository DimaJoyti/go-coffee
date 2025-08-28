# Variables for Disaster Recovery Module

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

variable "team" {
  description = "Team responsible for disaster recovery"
  type        = string
  default     = "platform"
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "platform"
}

# =============================================================================
# DISASTER RECOVERY CONFIGURATION
# =============================================================================

variable "disaster_recovery_namespace" {
  description = "Kubernetes namespace for disaster recovery tools"
  type        = string
  default     = "disaster-recovery"
}

variable "enable_disaster_recovery" {
  description = "Enable disaster recovery features"
  type        = bool
  default     = true
}

variable "enable_automated_backups" {
  description = "Enable automated backup system"
  type        = bool
  default     = true
}

variable "enable_automated_failover" {
  description = "Enable automated failover system"
  type        = bool
  default     = true
}

variable "enable_dr_testing" {
  description = "Enable disaster recovery testing"
  type        = bool
  default     = true
}

variable "enable_automated_dr_testing" {
  description = "Enable automated disaster recovery testing"
  type        = bool
  default     = false
}

variable "monitoring_enabled" {
  description = "Enable monitoring integration"
  type        = bool
  default     = true
}

# =============================================================================
# RECOVERY OBJECTIVES
# =============================================================================

variable "recovery_time_objective_minutes" {
  description = "Recovery Time Objective (RTO) in minutes"
  type        = number
  default     = 60
  validation {
    condition     = var.recovery_time_objective_minutes > 0
    error_message = "RTO must be greater than 0 minutes."
  }
}

variable "recovery_point_objective_minutes" {
  description = "Recovery Point Objective (RPO) in minutes"
  type        = number
  default     = 15
  validation {
    condition     = var.recovery_point_objective_minutes > 0
    error_message = "RPO must be greater than 0 minutes."
  }
}

# =============================================================================
# BACKUP CONFIGURATION
# =============================================================================

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
  validation {
    condition     = var.backup_retention_days > 0 && var.backup_retention_days <= 365
    error_message = "Backup retention must be between 1 and 365 days."
  }
}

variable "backup_frequency" {
  description = "Backup frequency (cron schedule)"
  type        = string
  default     = "0 2 * * *" # Daily at 2 AM
}

variable "enable_cross_region_backup" {
  description = "Enable cross-region backup replication"
  type        = bool
  default     = true
}

variable "enable_cross_cloud_backup" {
  description = "Enable cross-cloud backup replication"
  type        = bool
  default     = false
}

variable "backup_storage_bucket" {
  description = "Storage bucket for backups"
  type        = string
}

variable "enable_database_backup" {
  description = "Enable database backup"
  type        = bool
  default     = true
}

variable "database_backup_image" {
  description = "Container image for database backup"
  type        = string
  default     = "go-coffee/database-backup:latest"
}

# =============================================================================
# VELERO CONFIGURATION
# =============================================================================

variable "velero_chart_version" {
  description = "Version of Velero Helm chart"
  type        = string
  default     = "5.1.4"
}

# =============================================================================
# FAILOVER CONFIGURATION
# =============================================================================

variable "health_check_interval_seconds" {
  description = "Health check interval in seconds"
  type        = number
  default     = 30
  validation {
    condition     = var.health_check_interval_seconds > 0
    error_message = "Health check interval must be greater than 0 seconds."
  }
}

variable "failure_threshold_count" {
  description = "Number of consecutive failures before triggering failover"
  type        = number
  default     = 3
  validation {
    condition     = var.failure_threshold_count > 0
    error_message = "Failure threshold must be greater than 0."
  }
}

variable "recovery_threshold_count" {
  description = "Number of consecutive successes before triggering failback"
  type        = number
  default     = 5
  validation {
    condition     = var.recovery_threshold_count > 0
    error_message = "Recovery threshold must be greater than 0."
  }
}

variable "failover_controller_replicas" {
  description = "Number of failover controller replicas"
  type        = number
  default     = 2
}

variable "failover_controller_image" {
  description = "Container image for failover controller"
  type        = string
  default     = "go-coffee/failover-controller:latest"
}

# =============================================================================
# REGION CONFIGURATION
# =============================================================================

variable "primary_region" {
  description = "Primary region for deployment"
  type        = string
}

variable "secondary_region" {
  description = "Secondary region for disaster recovery"
  type        = string
}

variable "tertiary_region" {
  description = "Tertiary region for additional redundancy"
  type        = string
  default     = ""
}

variable "enable_cross_region_replication" {
  description = "Enable cross-region data replication"
  type        = bool
  default     = true
}

variable "replication_interval_hours" {
  description = "Replication interval in hours"
  type        = number
  default     = 4
  validation {
    condition     = var.replication_interval_hours > 0 && var.replication_interval_hours <= 24
    error_message = "Replication interval must be between 1 and 24 hours."
  }
}

variable "replication_manager_image" {
  description = "Container image for replication manager"
  type        = string
  default     = "go-coffee/replication-manager:latest"
}

# =============================================================================
# CLOUD PROVIDER CONFIGURATION
# =============================================================================

variable "cloud_provider" {
  description = "Primary cloud provider (aws, gcp, azure)"
  type        = string
  validation {
    condition     = contains(["aws", "gcp", "azure"], var.cloud_provider)
    error_message = "Cloud provider must be one of: aws, gcp, azure."
  }
}

# AWS Configuration
variable "aws_access_key_id" {
  description = "AWS access key ID for backups"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS secret access key for backups"
  type        = string
  default     = ""
  sensitive   = true
}

# GCP Configuration
variable "gcp_project_id" {
  description = "Google Cloud project ID"
  type        = string
  default     = ""
}

variable "gcp_service_account_key" {
  description = "Google Cloud service account key (JSON)"
  type        = string
  default     = ""
  sensitive   = true
}

# Azure Configuration
variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  default     = ""
}

variable "azure_tenant_id" {
  description = "Azure tenant ID"
  type        = string
  default     = ""
}

variable "azure_client_id" {
  description = "Azure client ID"
  type        = string
  default     = ""
}

variable "azure_client_secret" {
  description = "Azure client secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "azure_resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = ""
}

variable "azure_storage_account_name" {
  description = "Azure storage account name for backups"
  type        = string
  default     = ""
}

# =============================================================================
# DR TESTING CONFIGURATION
# =============================================================================

variable "dr_test_schedule" {
  description = "DR testing schedule (cron format)"
  type        = string
  default     = "0 6 1 * *" # Monthly on the 1st at 6 AM
}

variable "dr_test_type" {
  description = "Type of DR test to perform"
  type        = string
  default     = "backup_restore"
  validation {
    condition = contains([
      "backup_restore", "failover", "full_dr", "partial_dr"
    ], var.dr_test_type)
    error_message = "DR test type must be one of: backup_restore, failover, full_dr, partial_dr."
  }
}

variable "dr_testing_image" {
  description = "Container image for DR testing"
  type        = string
  default     = "go-coffee/dr-tester:latest"
}

# =============================================================================
# BUSINESS CONTINUITY CONFIGURATION
# =============================================================================

variable "business_continuity_plan" {
  description = "Business continuity plan configuration"
  type = object({
    critical_services = list(string)
    communication_plan = object({
      primary_contact = string
      escalation_contacts = list(string)
      notification_channels = list(string)
    })
    recovery_procedures = map(object({
      priority = number
      estimated_time_minutes = number
      dependencies = list(string)
      steps = list(string)
    }))
  })
  default = {
    critical_services = [
      "coffee-service",
      "payment-service",
      "user-service",
      "order-service",
      "inventory-service"
    ]
    communication_plan = {
      primary_contact = "platform@go-coffee.com"
      escalation_contacts = [
        "cto@go-coffee.com",
        "ops-manager@go-coffee.com"
      ]
      notification_channels = [
        "slack",
        "email",
        "sms"
      ]
    }
    recovery_procedures = {
      "database_recovery" = {
        priority = 1
        estimated_time_minutes = 30
        dependencies = ["backup_storage"]
        steps = [
          "Identify latest backup",
          "Restore database from backup",
          "Verify data integrity",
          "Update connection strings"
        ]
      }
      "application_recovery" = {
        priority = 2
        estimated_time_minutes = 15
        dependencies = ["database_recovery"]
        steps = [
          "Deploy applications to DR region",
          "Update DNS records",
          "Verify application health",
          "Enable traffic routing"
        ]
      }
    }
  }
}

variable "incident_response_team" {
  description = "Incident response team configuration"
  type = list(object({
    name = string
    role = string
    email = string
    phone = string
    primary = bool
  }))
  default = [
    {
      name = "Platform Team Lead"
      role = "incident_commander"
      email = "platform-lead@go-coffee.com"
      phone = "+1-555-0101"
      primary = true
    },
    {
      name = "DevOps Engineer"
      role = "technical_lead"
      email = "devops@go-coffee.com"
      phone = "+1-555-0102"
      primary = false
    },
    {
      name = "Database Administrator"
      role = "database_specialist"
      email = "dba@go-coffee.com"
      phone = "+1-555-0103"
      primary = false
    }
  ]
}

# =============================================================================
# NOTIFICATION CONFIGURATION
# =============================================================================

variable "slack_webhook_url" {
  description = "Slack webhook URL for DR notifications"
  type        = string
  default     = ""
  sensitive   = true
}

variable "email_notifications" {
  description = "Email addresses for DR notifications"
  type        = list(string)
  default     = ["platform@go-coffee.com"]
}

variable "sms_notifications" {
  description = "Phone numbers for SMS notifications"
  type        = list(string)
  default     = []
}

variable "webhook_url" {
  description = "Generic webhook URL for DR alerts"
  type        = string
  default     = ""
  sensitive   = true
}

# =============================================================================
# COMPLIANCE AND AUDIT
# =============================================================================

variable "compliance_requirements" {
  description = "Compliance requirements for DR"
  type = object({
    frameworks = list(string)
    audit_frequency_days = number
    documentation_required = bool
    testing_frequency_days = number
  })
  default = {
    frameworks = ["SOC2", "PCI-DSS"]
    audit_frequency_days = 90
    documentation_required = true
    testing_frequency_days = 30
  }
}

variable "audit_log_retention_days" {
  description = "Number of days to retain DR audit logs"
  type        = number
  default     = 2555 # 7 years for compliance
}

# =============================================================================
# ADVANCED DR FEATURES
# =============================================================================

variable "enable_chaos_engineering" {
  description = "Enable chaos engineering for DR testing"
  type        = bool
  default     = false
}

variable "enable_automated_rollback" {
  description = "Enable automated rollback on failed deployments"
  type        = bool
  default     = true
}

variable "enable_canary_deployments" {
  description = "Enable canary deployments for safer releases"
  type        = bool
  default     = true
}

variable "enable_blue_green_deployments" {
  description = "Enable blue-green deployments"
  type        = bool
  default     = false
}

variable "data_replication_strategy" {
  description = "Data replication strategy"
  type        = string
  default     = "async"
  validation {
    condition = contains([
      "sync", "async", "semi_sync"
    ], var.data_replication_strategy)
    error_message = "Data replication strategy must be one of: sync, async, semi_sync."
  }
}

# =============================================================================
# COST OPTIMIZATION FOR DR
# =============================================================================

variable "dr_cost_optimization" {
  description = "Cost optimization settings for DR infrastructure"
  type = object({
    use_spot_instances = bool
    auto_shutdown_non_critical = bool
    storage_class_optimization = bool
    cross_region_transfer_optimization = bool
  })
  default = {
    use_spot_instances = true
    auto_shutdown_non_critical = true
    storage_class_optimization = true
    cross_region_transfer_optimization = true
  }
}

variable "dr_budget_limit_monthly" {
  description = "Monthly budget limit for DR infrastructure (USD)"
  type        = number
  default     = 5000
}
