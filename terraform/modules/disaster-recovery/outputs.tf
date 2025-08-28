# Outputs for Disaster Recovery Module

# =============================================================================
# NAMESPACE OUTPUTS
# =============================================================================

output "disaster_recovery_namespace" {
  description = "Name of the disaster recovery namespace"
  value       = kubernetes_namespace.disaster_recovery.metadata[0].name
}

# =============================================================================
# VELERO OUTPUTS
# =============================================================================

output "velero_backup_location" {
  description = "Velero backup storage location"
  value       = var.enable_automated_backups ? var.backup_storage_bucket : null
}

output "velero_schedule_names" {
  description = "Names of Velero backup schedules"
  value = var.enable_automated_backups ? [
    "daily",
    "weekly"
  ] : []
}

# =============================================================================
# FAILOVER CONTROLLER OUTPUTS
# =============================================================================

output "failover_controller_service_account" {
  description = "Service account name for failover controller"
  value       = var.enable_automated_failover ? kubernetes_service_account.failover_controller[0].metadata[0].name : null
}

output "failover_controller_cluster_role" {
  description = "Cluster role name for failover controller"
  value       = var.enable_automated_failover ? kubernetes_cluster_role.failover_controller[0].metadata[0].name : null
}

# =============================================================================
# RECOVERY OBJECTIVES OUTPUTS
# =============================================================================

output "recovery_time_objective_minutes" {
  description = "Recovery Time Objective (RTO) in minutes"
  value       = var.recovery_time_objective_minutes
}

output "recovery_point_objective_minutes" {
  description = "Recovery Point Objective (RPO) in minutes"
  value       = var.recovery_point_objective_minutes
}

# =============================================================================
# BACKUP CONFIGURATION OUTPUTS
# =============================================================================

output "backup_retention_days" {
  description = "Backup retention period in days"
  value       = var.backup_retention_days
}

output "backup_frequency" {
  description = "Backup frequency (cron schedule)"
  value       = var.backup_frequency
}

output "cross_region_backup_enabled" {
  description = "Whether cross-region backup is enabled"
  value       = var.enable_cross_region_backup
}

output "cross_cloud_backup_enabled" {
  description = "Whether cross-cloud backup is enabled"
  value       = var.enable_cross_cloud_backup
}

# =============================================================================
# REGION CONFIGURATION OUTPUTS
# =============================================================================

output "primary_region" {
  description = "Primary region for deployment"
  value       = var.primary_region
}

output "secondary_region" {
  description = "Secondary region for disaster recovery"
  value       = var.secondary_region
}

output "tertiary_region" {
  description = "Tertiary region for additional redundancy"
  value       = var.tertiary_region != "" ? var.tertiary_region : null
}

# =============================================================================
# FAILOVER CONFIGURATION OUTPUTS
# =============================================================================

output "health_check_interval_seconds" {
  description = "Health check interval in seconds"
  value       = var.health_check_interval_seconds
}

output "failure_threshold_count" {
  description = "Number of consecutive failures before triggering failover"
  value       = var.failure_threshold_count
}

output "recovery_threshold_count" {
  description = "Number of consecutive successes before triggering failback"
  value       = var.recovery_threshold_count
}

# =============================================================================
# DR TESTING OUTPUTS
# =============================================================================

output "dr_test_schedule" {
  description = "DR testing schedule (cron format)"
  value       = var.dr_test_schedule
}

output "automated_dr_testing_enabled" {
  description = "Whether automated DR testing is enabled"
  value       = var.enable_automated_dr_testing
}

# =============================================================================
# BUSINESS CONTINUITY OUTPUTS
# =============================================================================

output "business_continuity_plan" {
  description = "Business continuity plan configuration"
  value       = var.business_continuity_plan
  sensitive   = false
}

output "incident_response_team" {
  description = "Incident response team configuration"
  value       = var.incident_response_team
  sensitive   = true
}

# =============================================================================
# COMPLIANCE OUTPUTS
# =============================================================================

output "compliance_frameworks" {
  description = "Compliance frameworks for DR"
  value       = var.compliance_requirements.frameworks
}

output "audit_frequency_days" {
  description = "Audit frequency in days"
  value       = var.compliance_requirements.audit_frequency_days
}

output "audit_log_retention_days" {
  description = "Audit log retention period in days"
  value       = var.audit_log_retention_days
}

# =============================================================================
# COST OPTIMIZATION OUTPUTS
# =============================================================================

output "dr_cost_optimization_settings" {
  description = "Cost optimization settings for DR infrastructure"
  value       = var.dr_cost_optimization
}

output "dr_budget_limit_monthly" {
  description = "Monthly budget limit for DR infrastructure (USD)"
  value       = var.dr_budget_limit_monthly
}

# =============================================================================
# NOTIFICATION OUTPUTS
# =============================================================================

output "notification_channels" {
  description = "Configured notification channels"
  value = {
    slack_enabled = var.slack_webhook_url != ""
    email_enabled = length(var.email_notifications) > 0
    sms_enabled   = length(var.sms_notifications) > 0
    webhook_enabled = var.webhook_url != ""
  }
}

# =============================================================================
# FEATURE FLAGS OUTPUTS
# =============================================================================

output "enabled_features" {
  description = "Enabled disaster recovery features"
  value = {
    disaster_recovery     = var.enable_disaster_recovery
    automated_backups     = var.enable_automated_backups
    automated_failover    = var.enable_automated_failover
    dr_testing           = var.enable_dr_testing
    automated_dr_testing = var.enable_automated_dr_testing
    cross_region_replication = var.enable_cross_region_replication
    database_backup      = var.enable_database_backup
    chaos_engineering    = var.enable_chaos_engineering
    automated_rollback   = var.enable_automated_rollback
    canary_deployments   = var.enable_canary_deployments
    blue_green_deployments = var.enable_blue_green_deployments
  }
}

# =============================================================================
# REPLICATION OUTPUTS
# =============================================================================

output "replication_interval_hours" {
  description = "Data replication interval in hours"
  value       = var.replication_interval_hours
}

output "data_replication_strategy" {
  description = "Data replication strategy"
  value       = var.data_replication_strategy
}

# =============================================================================
# CLOUD PROVIDER OUTPUTS
# =============================================================================

output "cloud_provider" {
  description = "Primary cloud provider for DR"
  value       = var.cloud_provider
}

output "backup_storage_configuration" {
  description = "Backup storage configuration by cloud provider"
  value = {
    aws = var.cloud_provider == "aws" ? {
      bucket = var.backup_storage_bucket
      region = var.primary_region
    } : null
    
    gcp = var.cloud_provider == "gcp" ? {
      bucket = var.backup_storage_bucket
      project = var.gcp_project_id
      location = var.primary_region
    } : null
    
    azure = var.cloud_provider == "azure" ? {
      storage_account = var.azure_storage_account_name
      resource_group = var.azure_resource_group_name
      location = var.azure_location
    } : null
  }
}

# =============================================================================
# MONITORING OUTPUTS
# =============================================================================

output "monitoring_enabled" {
  description = "Whether monitoring integration is enabled"
  value       = var.monitoring_enabled
}

# =============================================================================
# SUMMARY OUTPUT
# =============================================================================

output "disaster_recovery_summary" {
  description = "Summary of disaster recovery configuration"
  value = {
    # Core configuration
    environment = var.environment
    project_name = var.project_name
    namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
    
    # Recovery objectives
    rto_minutes = var.recovery_time_objective_minutes
    rpo_minutes = var.recovery_point_objective_minutes
    
    # Regions
    primary_region = var.primary_region
    secondary_region = var.secondary_region
    tertiary_region = var.tertiary_region != "" ? var.tertiary_region : null
    
    # Backup configuration
    backup_enabled = var.enable_automated_backups
    backup_retention_days = var.backup_retention_days
    backup_frequency = var.backup_frequency
    cross_region_backup = var.enable_cross_region_backup
    cross_cloud_backup = var.enable_cross_cloud_backup
    
    # Failover configuration
    automated_failover = var.enable_automated_failover
    health_check_interval = var.health_check_interval_seconds
    failure_threshold = var.failure_threshold_count
    
    # Testing configuration
    dr_testing_enabled = var.enable_dr_testing
    automated_testing = var.enable_automated_dr_testing
    test_schedule = var.dr_test_schedule
    
    # Compliance
    compliance_frameworks = var.compliance_requirements.frameworks
    audit_frequency_days = var.compliance_requirements.audit_frequency_days
    
    # Cost optimization
    monthly_budget_limit = var.dr_budget_limit_monthly
    cost_optimization_enabled = var.dr_cost_optimization.use_spot_instances
  }
}
