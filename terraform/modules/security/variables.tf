# Security Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
  default     = ""
}

variable "cluster_location" {
  description = "GKE cluster location"
  type        = string
  default     = ""
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "enable_binary_authorization" {
  description = "Enable Binary Authorization for container image security"
  type        = bool
  default     = false
}

variable "enable_pod_security_policy" {
  description = "Enable Pod Security Standards"
  type        = bool
  default     = true
}

variable "enable_network_policy" {
  description = "Enable Kubernetes Network Policies"
  type        = bool
  default     = true
}

variable "enable_workload_identity" {
  description = "Enable Workload Identity for secure access to GCP services"
  type        = bool
  default     = true
}

variable "enable_rbac" {
  description = "Enable Role-Based Access Control"
  type        = bool
  default     = true
}

variable "enable_security_monitoring" {
  description = "Enable security monitoring and alerting"
  type        = bool
  default     = true
}

variable "pgp_public_key" {
  description = "PGP public key for Binary Authorization attestor"
  type        = string
  default     = ""
}

variable "service_accounts" {
  description = "Service accounts for workload identity"
  type = map(object({
    display_name = string
    description  = string
    namespace    = string
    ksa_name     = string
    roles        = list(string)
  }))
  default = {
    "go-coffee-api-gateway" = {
      display_name = "Go Coffee API Gateway"
      description  = "Service account for API Gateway service"
      namespace    = "go-coffee"
      ksa_name     = "api-gateway"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent"
      ]
    }
    "go-coffee-order-service" = {
      display_name = "Go Coffee Order Service"
      description  = "Service account for Order service"
      namespace    = "go-coffee"
      ksa_name     = "order-service"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent",
        "roles/cloudsql.client"
      ]
    }
    "go-coffee-payment-service" = {
      display_name = "Go Coffee Payment Service"
      description  = "Service account for Payment service"
      namespace    = "go-coffee"
      ksa_name     = "payment-service"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent",
        "roles/cloudsql.client",
        "roles/secretmanager.secretAccessor"
      ]
    }
    "go-coffee-kitchen-service" = {
      display_name = "Go Coffee Kitchen Service"
      description  = "Service account for Kitchen service"
      namespace    = "go-coffee"
      ksa_name     = "kitchen-service"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent",
        "roles/redis.editor"
      ]
    }
    "go-coffee-ai-services" = {
      display_name = "Go Coffee AI Services"
      description  = "Service account for AI-related services"
      namespace    = "go-coffee"
      ksa_name     = "ai-services"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent",
        "roles/aiplatform.user",
        "roles/secretmanager.secretAccessor"
      ]
    }
    "go-coffee-web3-services" = {
      display_name = "Go Coffee Web3 Services"
      description  = "Service account for Web3 and DeFi services"
      namespace    = "go-coffee"
      ksa_name     = "web3-services"
      roles = [
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
        "roles/cloudtrace.agent",
        "roles/secretmanager.secretAccessor",
        "roles/storage.objectViewer"
      ]
    }
  }
}

variable "notification_channels" {
  description = "List of notification channels for security alerts"
  type        = list(string)
  default     = []
}

variable "network_policy_config" {
  description = "Network policy configuration"
  type = object({
    default_deny_all           = bool
    allow_internal_communication = bool
    allow_dns                  = bool
    allow_external_https       = bool
  })
  default = {
    default_deny_all           = true
    allow_internal_communication = true
    allow_dns                  = true
    allow_external_https       = true
  }
}

variable "pod_security_standards" {
  description = "Pod Security Standards configuration"
  type = object({
    enforce = string
    audit   = string
    warn    = string
  })
  default = {
    enforce = "restricted"
    audit   = "restricted"
    warn    = "restricted"
  }
  
  validation {
    condition = alltrue([
      contains(["privileged", "baseline", "restricted"], var.pod_security_standards.enforce),
      contains(["privileged", "baseline", "restricted"], var.pod_security_standards.audit),
      contains(["privileged", "baseline", "restricted"], var.pod_security_standards.warn)
    ])
    error_message = "Pod security standards must be one of: privileged, baseline, restricted."
  }
}

variable "rbac_config" {
  description = "RBAC configuration"
  type = object({
    enable_cluster_admin = bool
    enable_namespace_admin = bool
    enable_read_only     = bool
  })
  default = {
    enable_cluster_admin   = false
    enable_namespace_admin = true
    enable_read_only       = true
  }
}

variable "security_scanning_config" {
  description = "Security scanning configuration"
  type = object({
    enable_vulnerability_scanning = bool
    enable_binary_authorization   = bool
    enable_admission_controller   = bool
  })
  default = {
    enable_vulnerability_scanning = true
    enable_binary_authorization   = false
    enable_admission_controller   = true
  }
}

variable "compliance_frameworks" {
  description = "Compliance frameworks to enable"
  type        = list(string)
  default     = ["CIS", "PCI-DSS", "SOC2"]
  
  validation {
    condition = alltrue([
      for framework in var.compliance_frameworks : contains([
        "CIS", "PCI-DSS", "SOC2", "HIPAA", "ISO27001"
      ], framework)
    ])
    error_message = "Compliance frameworks must be from: CIS, PCI-DSS, SOC2, HIPAA, ISO27001."
  }
}
