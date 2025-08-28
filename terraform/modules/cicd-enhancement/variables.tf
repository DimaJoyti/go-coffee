# Variables for CI/CD Enhancement Module

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
  description = "Team responsible for CI/CD"
  type        = string
  default     = "platform"
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "platform"
}

# =============================================================================
# CI/CD CONFIGURATION
# =============================================================================

variable "cicd_namespace" {
  description = "Kubernetes namespace for CI/CD tools"
  type        = string
  default     = "cicd"
}

variable "enable_cicd_enhancement" {
  description = "Enable CI/CD enhancement features"
  type        = bool
  default     = true
}

variable "enable_argocd" {
  description = "Enable ArgoCD for GitOps"
  type        = bool
  default     = true
}

variable "enable_tekton_pipelines" {
  description = "Enable Tekton Pipelines"
  type        = bool
  default     = true
}

variable "enable_github_actions_runner" {
  description = "Enable GitHub Actions Runner"
  type        = bool
  default     = true
}

variable "monitoring_enabled" {
  description = "Enable monitoring integration"
  type        = bool
  default     = true
}

# =============================================================================
# PIPELINE STAGES
# =============================================================================

variable "enable_build_stage" {
  description = "Enable build stage in pipeline"
  type        = bool
  default     = true
}

variable "enable_test_stage" {
  description = "Enable test stage in pipeline"
  type        = bool
  default     = true
}

variable "enable_security_scan_stage" {
  description = "Enable security scan stage in pipeline"
  type        = bool
  default     = true
}

variable "enable_deploy_stage" {
  description = "Enable deploy stage in pipeline"
  type        = bool
  default     = true
}

variable "enable_integration_test_stage" {
  description = "Enable integration test stage in pipeline"
  type        = bool
  default     = true
}

variable "enable_performance_test_stage" {
  description = "Enable performance test stage in pipeline"
  type        = bool
  default     = false
}

# =============================================================================
# DEPLOYMENT STRATEGIES
# =============================================================================

variable "deployment_strategy" {
  description = "Deployment strategy"
  type        = string
  default     = "rolling"
  validation {
    condition = contains([
      "rolling", "blue_green", "canary", "recreate"
    ], var.deployment_strategy)
    error_message = "Deployment strategy must be one of: rolling, blue_green, canary, recreate."
  }
}

variable "canary_deployment_percentage" {
  description = "Percentage of traffic for canary deployment"
  type        = number
  default     = 10
  validation {
    condition     = var.canary_deployment_percentage > 0 && var.canary_deployment_percentage <= 100
    error_message = "Canary deployment percentage must be between 1 and 100."
  }
}

variable "enable_blue_green_deployment" {
  description = "Enable blue-green deployment"
  type        = bool
  default     = false
}

variable "enable_automated_rollback" {
  description = "Enable automated rollback on deployment failure"
  type        = bool
  default     = true
}

# =============================================================================
# QUALITY GATES
# =============================================================================

variable "code_coverage_threshold" {
  description = "Minimum code coverage percentage"
  type        = number
  default     = 80
  validation {
    condition     = var.code_coverage_threshold >= 0 && var.code_coverage_threshold <= 100
    error_message = "Code coverage threshold must be between 0 and 100."
  }
}

variable "security_scan_threshold" {
  description = "Security scan severity threshold"
  type        = string
  default     = "HIGH"
  validation {
    condition = contains([
      "LOW", "MEDIUM", "HIGH", "CRITICAL"
    ], var.security_scan_threshold)
    error_message = "Security scan threshold must be one of: LOW, MEDIUM, HIGH, CRITICAL."
  }
}

variable "performance_threshold" {
  description = "Performance test threshold (response time in ms)"
  type        = number
  default     = 1000
}

# =============================================================================
# MULTI-CLOUD DEPLOYMENT
# =============================================================================

variable "enable_aws_deployment" {
  description = "Enable deployment to AWS"
  type        = bool
  default     = true
}

variable "enable_gcp_deployment" {
  description = "Enable deployment to GCP"
  type        = bool
  default     = false
}

variable "enable_azure_deployment" {
  description = "Enable deployment to Azure"
  type        = bool
  default     = false
}

# =============================================================================
# ARGOCD CONFIGURATION
# =============================================================================

variable "argocd_chart_version" {
  description = "Version of ArgoCD Helm chart"
  type        = string
  default     = "5.51.6"
}

variable "argocd_version" {
  description = "ArgoCD application version"
  type        = string
  default     = "v2.9.3"
}

variable "argocd_controller_replicas" {
  description = "Number of ArgoCD controller replicas"
  type        = number
  default     = 1
}

variable "argocd_server_replicas" {
  description = "Number of ArgoCD server replicas"
  type        = number
  default     = 2
}

variable "argocd_repo_server_replicas" {
  description = "Number of ArgoCD repo server replicas"
  type        = number
  default     = 2
}

variable "argocd_ingress_enabled" {
  description = "Enable ingress for ArgoCD"
  type        = bool
  default     = true
}

variable "argocd_hostname" {
  description = "Hostname for ArgoCD ingress"
  type        = string
  default     = "argocd.go-coffee.local"
}

variable "enable_argocd_applicationset" {
  description = "Enable ArgoCD ApplicationSet controller"
  type        = bool
  default     = true
}

variable "enable_argocd_notifications" {
  description = "Enable ArgoCD notifications"
  type        = bool
  default     = true
}

variable "enable_auto_sync" {
  description = "Enable automatic sync in ArgoCD"
  type        = bool
  default     = false
}

# =============================================================================
# TEKTON CONFIGURATION
# =============================================================================

variable "tekton_chart_version" {
  description = "Version of Tekton Helm chart"
  type        = string
  default     = "0.50.0"
}

variable "tekton_version" {
  description = "Tekton Pipelines version"
  type        = string
  default     = "v0.53.0"
}

variable "tekton_controller_replicas" {
  description = "Number of Tekton controller replicas"
  type        = number
  default     = 1
}

variable "tekton_webhook_replicas" {
  description = "Number of Tekton webhook replicas"
  type        = number
  default     = 1
}

# =============================================================================
# GITHUB ACTIONS RUNNER CONFIGURATION
# =============================================================================

variable "actions_runner_controller_version" {
  description = "Version of Actions Runner Controller Helm chart"
  type        = string
  default     = "0.23.7"
}

variable "actions_runner_controller_replicas" {
  description = "Number of Actions Runner Controller replicas"
  type        = number
  default     = 1
}

variable "github_runner_replicas" {
  description = "Number of GitHub runner replicas"
  type        = number
  default     = 3
}

variable "enable_github_webhook_server" {
  description = "Enable GitHub webhook server"
  type        = bool
  default     = true
}

variable "github_webhook_ingress_enabled" {
  description = "Enable ingress for GitHub webhook server"
  type        = bool
  default     = true
}

variable "github_webhook_hostname" {
  description = "Hostname for GitHub webhook ingress"
  type        = string
  default     = "github-webhook.go-coffee.local"
}

# =============================================================================
# GIT CONFIGURATION
# =============================================================================

variable "git_repository_url" {
  description = "Git repository URL"
  type        = string
}

variable "git_target_revision" {
  description = "Git target revision (branch/tag/commit)"
  type        = string
  default     = "main"
}

variable "git_username" {
  description = "Git username for authentication"
  type        = string
  default     = ""
}

variable "git_token" {
  description = "Git token for authentication"
  type        = string
  default     = ""
  sensitive   = true
}

variable "github_token" {
  description = "GitHub token for Actions Runner Controller"
  type        = string
  default     = ""
  sensitive   = true
}

variable "k8s_manifests_path" {
  description = "Path to Kubernetes manifests in git repository"
  type        = string
  default     = "k8s"
}

variable "use_helm_charts" {
  description = "Use Helm charts for deployment"
  type        = bool
  default     = true
}

# =============================================================================
# DOCKER REGISTRY CONFIGURATION
# =============================================================================

variable "docker_registry_server" {
  description = "Docker registry server"
  type        = string
  default     = "docker.io"
}

variable "docker_registry_username" {
  description = "Docker registry username"
  type        = string
  default     = ""
}

variable "docker_registry_password" {
  description = "Docker registry password"
  type        = string
  default     = ""
  sensitive   = true
}

# =============================================================================
# KUBERNETES CONFIGURATION
# =============================================================================

variable "ingress_class_name" {
  description = "Ingress class name"
  type        = string
  default     = "nginx"
}

variable "cert_manager_issuer" {
  description = "Cert-manager cluster issuer name"
  type        = string
  default     = "letsencrypt-prod"
}

# =============================================================================
# NOTIFICATION CONFIGURATION
# =============================================================================

variable "slack_token" {
  description = "Slack token for notifications"
  type        = string
  default     = ""
  sensitive   = true
}

variable "slack_channel" {
  description = "Slack channel for notifications"
  type        = string
  default     = "#deployments"
}

# =============================================================================
# ADVANCED CI/CD FEATURES
# =============================================================================

variable "enable_progressive_delivery" {
  description = "Enable progressive delivery with Flagger"
  type        = bool
  default     = false
}

variable "enable_chaos_engineering" {
  description = "Enable chaos engineering in CI/CD"
  type        = bool
  default     = false
}

variable "enable_load_testing" {
  description = "Enable automated load testing"
  type        = bool
  default     = false
}

variable "enable_security_scanning" {
  description = "Enable comprehensive security scanning"
  type        = bool
  default     = true
}

variable "enable_compliance_checks" {
  description = "Enable compliance checks in pipeline"
  type        = bool
  default     = true
}

variable "enable_artifact_signing" {
  description = "Enable artifact signing and verification"
  type        = bool
  default     = false
}

# =============================================================================
# PIPELINE CONFIGURATION
# =============================================================================

variable "pipeline_timeout_minutes" {
  description = "Pipeline timeout in minutes"
  type        = number
  default     = 60
}

variable "parallel_jobs" {
  description = "Number of parallel jobs in pipeline"
  type        = number
  default     = 3
}

variable "enable_pipeline_caching" {
  description = "Enable pipeline caching"
  type        = bool
  default     = true
}

variable "cache_retention_days" {
  description = "Cache retention period in days"
  type        = number
  default     = 7
}
