# Variables for Security and Compliance Module

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
  description = "Team responsible for the security infrastructure"
  type        = string
  default     = "security"
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "security"
}

# =============================================================================
# SECURITY CONFIGURATION
# =============================================================================

variable "security_namespace" {
  description = "Kubernetes namespace for security tools"
  type        = string
  default     = "security"
}

variable "enable_vulnerability_scanning" {
  description = "Enable vulnerability scanning"
  type        = bool
  default     = true
}

variable "enable_compliance_monitoring" {
  description = "Enable compliance monitoring"
  type        = bool
  default     = true
}

variable "enable_threat_detection" {
  description = "Enable threat detection"
  type        = bool
  default     = true
}

variable "enable_policy_enforcement" {
  description = "Enable policy enforcement with OPA Gatekeeper"
  type        = bool
  default     = true
}

variable "enable_mutation" {
  description = "Enable mutation in Gatekeeper"
  type        = bool
  default     = false
}

variable "monitoring_enabled" {
  description = "Enable monitoring integration"
  type        = bool
  default     = true
}

# =============================================================================
# VULNERABILITY SCANNING CONFIGURATION
# =============================================================================

variable "trivy_operator_version" {
  description = "Version of Trivy Operator Helm chart"
  type        = string
  default     = "0.18.4"
}

variable "vulnerability_scanner_image" {
  description = "Container image for vulnerability scanner"
  type        = string
  default     = "aquasec/trivy:latest"
}

variable "vulnerability_scan_schedule" {
  description = "Cron schedule for vulnerability scans"
  type        = string
  default     = "0 2 * * *" # Daily at 2 AM
}

variable "vulnerability_scan_targets" {
  description = "List of container images to scan for vulnerabilities"
  type        = list(string)
  default = [
    "go-coffee/coffee-service:latest",
    "go-coffee/ai-agent:latest",
    "go-coffee/defi-service:latest",
    "go-coffee/web3-gateway:latest",
    "go-coffee/notification-service:latest"
  ]
}

variable "vulnerability_severity_threshold" {
  description = "Minimum severity level to report"
  type        = string
  default     = "HIGH,CRITICAL"
  validation {
    condition = contains([
      "UNKNOWN", "LOW", "MEDIUM", "HIGH", "CRITICAL",
      "LOW,MEDIUM,HIGH,CRITICAL", "MEDIUM,HIGH,CRITICAL", "HIGH,CRITICAL"
    ], var.vulnerability_severity_threshold)
    error_message = "Severity threshold must be a valid Trivy severity level."
  }
}

# =============================================================================
# COMPLIANCE MONITORING CONFIGURATION
# =============================================================================

variable "compliance_frameworks" {
  description = "List of compliance frameworks to monitor"
  type        = list(string)
  default     = ["SOC2", "PCI-DSS", "GDPR"]
  validation {
    condition = alltrue([
      for framework in var.compliance_frameworks :
      contains(["SOC2", "PCI-DSS", "GDPR", "HIPAA", "ISO27001"], framework)
    ])
    error_message = "Compliance frameworks must be from: SOC2, PCI-DSS, GDPR, HIPAA, ISO27001."
  }
}

variable "compliance_checker_image" {
  description = "Container image for compliance checker"
  type        = string
  default     = "go-coffee/compliance-checker:latest"
}

variable "compliance_check_schedule" {
  description = "Cron schedule for compliance checks"
  type        = string
  default     = "0 6 * * 1" # Weekly on Monday at 6 AM
}

# =============================================================================
# POLICY ENFORCEMENT CONFIGURATION
# =============================================================================

variable "gatekeeper_version" {
  description = "Version of Gatekeeper Helm chart"
  type        = string
  default     = "3.14.0"
}

variable "gatekeeper_replicas" {
  description = "Number of Gatekeeper controller replicas"
  type        = number
  default     = 3
}

variable "gatekeeper_audit_replicas" {
  description = "Number of Gatekeeper audit replicas"
  type        = number
  default     = 1
}

variable "security_policies" {
  description = "Security policies for OPA Gatekeeper"
  type = map(object({
    kind       = string
    properties = map(any)
    rego_policy = string
  }))
  default = {
    "requiredsecuritycontext" = {
      kind = "RequiredSecurityContext"
      properties = {
        runAsNonRoot = {
          type = "boolean"
        }
        runAsUser = {
          type = "integer"
          minimum = 1000
        }
        fsGroup = {
          type = "integer"
          minimum = 1000
        }
      }
      rego_policy = <<-EOF
        package requiredsecuritycontext

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not container.securityContext.runAsNonRoot
          msg := "Container must run as non-root user"
        }

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          container.securityContext.runAsUser < 1000
          msg := "Container must run as user ID >= 1000"
        }
      EOF
    }
  }
}

# =============================================================================
# THREAT DETECTION CONFIGURATION
# =============================================================================

variable "falco_version" {
  description = "Version of Falco Helm chart"
  type        = string
  default     = "3.8.4"
}

variable "threat_detection_sensitivity" {
  description = "Threat detection sensitivity level"
  type        = string
  default     = "warning"
  validation {
    condition = contains([
      "emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"
    ], var.threat_detection_sensitivity)
    error_message = "Threat detection sensitivity must be a valid syslog priority level."
  }
}

# =============================================================================
# CLOUD PROVIDER CONFIGURATION
# =============================================================================

variable "aws_region" {
  description = "AWS region for security services"
  type        = string
  default     = "us-east-1"
}

variable "gcp_project_id" {
  description = "Google Cloud project ID"
  type        = string
  default     = ""
}

variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  default     = ""
}

# =============================================================================
# NOTIFICATION CONFIGURATION
# =============================================================================

variable "slack_webhook_url" {
  description = "Slack webhook URL for security alerts"
  type        = string
  default     = ""
  sensitive   = true
}

variable "email_username" {
  description = "Email username for notifications"
  type        = string
  default     = ""
}

variable "email_password" {
  description = "Email password for notifications"
  type        = string
  default     = ""
  sensitive   = true
}

variable "webhook_url" {
  description = "Generic webhook URL for alerts"
  type        = string
  default     = ""
  sensitive   = true
}

# =============================================================================
# ADVANCED SECURITY CONFIGURATION
# =============================================================================

variable "enable_network_policies" {
  description = "Enable Kubernetes network policies"
  type        = bool
  default     = true
}

variable "enable_pod_security_standards" {
  description = "Enable Pod Security Standards"
  type        = bool
  default     = true
}

variable "enable_admission_controllers" {
  description = "Enable additional admission controllers"
  type        = bool
  default     = true
}

variable "enable_rbac_analysis" {
  description = "Enable RBAC analysis and recommendations"
  type        = bool
  default     = true
}

variable "enable_secret_scanning" {
  description = "Enable secret scanning in repositories and containers"
  type        = bool
  default     = true
}

variable "enable_image_signing" {
  description = "Enable container image signing verification"
  type        = bool
  default     = false
}

variable "enable_runtime_protection" {
  description = "Enable runtime protection and anomaly detection"
  type        = bool
  default     = true
}

# =============================================================================
# COMPLIANCE SPECIFIC CONFIGURATION
# =============================================================================

variable "data_classification_levels" {
  description = "Data classification levels for compliance"
  type        = list(string)
  default     = ["public", "internal", "confidential", "restricted"]
}

variable "retention_policies" {
  description = "Data retention policies by classification"
  type = map(object({
    retention_days = number
    backup_required = bool
    encryption_required = bool
  }))
  default = {
    "public" = {
      retention_days = 365
      backup_required = false
      encryption_required = false
    }
    "internal" = {
      retention_days = 1095  # 3 years
      backup_required = true
      encryption_required = true
    }
    "confidential" = {
      retention_days = 2555  # 7 years
      backup_required = true
      encryption_required = true
    }
    "restricted" = {
      retention_days = 2555  # 7 years
      backup_required = true
      encryption_required = true
    }
  }
}

variable "audit_log_retention_days" {
  description = "Number of days to retain audit logs"
  type        = number
  default     = 2555  # 7 years for compliance
}

variable "security_incident_response_team" {
  description = "Security incident response team contacts"
  type = list(object({
    name  = string
    email = string
    role  = string
  }))
  default = [
    {
      name  = "Security Team"
      email = "security@go-coffee.com"
      role  = "primary"
    }
  ]
}

# =============================================================================
# ENCRYPTION CONFIGURATION
# =============================================================================

variable "encryption_key_rotation_days" {
  description = "Number of days between encryption key rotations"
  type        = number
  default     = 90
}

variable "enable_envelope_encryption" {
  description = "Enable envelope encryption for sensitive data"
  type        = bool
  default     = true
}

variable "kms_key_aliases" {
  description = "KMS key aliases for different data types"
  type = map(string)
  default = {
    "database"     = "go-coffee-db-key"
    "application"  = "go-coffee-app-key"
    "backup"       = "go-coffee-backup-key"
    "logs"         = "go-coffee-logs-key"
  }
}

# =============================================================================
# SECURITY SCANNING CONFIGURATION
# =============================================================================

variable "code_scanning_tools" {
  description = "List of code scanning tools to enable"
  type        = list(string)
  default     = ["sonarqube", "semgrep", "bandit", "gosec"]
}

variable "dependency_scanning_tools" {
  description = "List of dependency scanning tools to enable"
  type        = list(string)
  default     = ["snyk", "safety", "audit"]
}

variable "infrastructure_scanning_tools" {
  description = "List of infrastructure scanning tools to enable"
  type        = list(string)
  default     = ["checkov", "tfsec", "terrascan"]
}

variable "container_scanning_registries" {
  description = "Container registries to scan"
  type        = list(string)
  default = [
    "docker.io",
    "gcr.io",
    "public.ecr.aws"
  ]
}

# =============================================================================
# INCIDENT RESPONSE CONFIGURATION
# =============================================================================

variable "incident_response_playbooks" {
  description = "Incident response playbooks configuration"
  type = map(object({
    severity_level = string
    escalation_time_minutes = number
    notification_channels = list(string)
    automated_actions = list(string)
  }))
  default = {
    "security_breach" = {
      severity_level = "critical"
      escalation_time_minutes = 15
      notification_channels = ["slack", "email", "pagerduty"]
      automated_actions = ["isolate_affected_systems", "collect_forensic_data"]
    }
    "compliance_violation" = {
      severity_level = "high"
      escalation_time_minutes = 30
      notification_channels = ["slack", "email"]
      automated_actions = ["generate_compliance_report", "notify_compliance_team"]
    }
    "vulnerability_detected" = {
      severity_level = "medium"
      escalation_time_minutes = 60
      notification_channels = ["slack"]
      automated_actions = ["create_remediation_ticket", "schedule_patching"]
    }
  }
}

variable "security_metrics_retention_days" {
  description = "Number of days to retain security metrics"
  type        = number
  default     = 365
}
