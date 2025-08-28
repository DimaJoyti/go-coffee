# Variables for Cost Optimization Module

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
  description = "Team responsible for cost optimization"
  type        = string
  default     = "platform"
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "platform"
}

# =============================================================================
# COST OPTIMIZATION CONFIGURATION
# =============================================================================

variable "cost_optimization_namespace" {
  description = "Kubernetes namespace for cost optimization tools"
  type        = string
  default     = "cost-optimization"
}

variable "enable_cost_optimization" {
  description = "Enable cost optimization features"
  type        = bool
  default     = true
}

variable "enable_rightsizing" {
  description = "Enable resource rightsizing recommendations"
  type        = bool
  default     = true
}

variable "enable_intelligent_scheduling" {
  description = "Enable intelligent workload scheduling"
  type        = bool
  default     = true
}

variable "enable_advanced_autoscaling" {
  description = "Enable advanced autoscaling features"
  type        = bool
  default     = true
}

variable "enable_cost_analyzer" {
  description = "Enable cost analyzer (KubeCost)"
  type        = bool
  default     = true
}

variable "enable_cluster_autoscaling" {
  description = "Enable cluster autoscaling"
  type        = bool
  default     = true
}

variable "monitoring_enabled" {
  description = "Enable monitoring integration"
  type        = bool
  default     = true
}

# =============================================================================
# RIGHTSIZING CONFIGURATION
# =============================================================================

variable "cpu_threshold_upper" {
  description = "Upper CPU utilization threshold for rightsizing (percentage)"
  type        = number
  default     = 80
  validation {
    condition     = var.cpu_threshold_upper > 0 && var.cpu_threshold_upper <= 100
    error_message = "CPU threshold must be between 0 and 100."
  }
}

variable "cpu_threshold_lower" {
  description = "Lower CPU utilization threshold for rightsizing (percentage)"
  type        = number
  default     = 20
  validation {
    condition     = var.cpu_threshold_lower > 0 && var.cpu_threshold_lower <= 100
    error_message = "CPU threshold must be between 0 and 100."
  }
}

variable "memory_threshold_upper" {
  description = "Upper memory utilization threshold for rightsizing (percentage)"
  type        = number
  default     = 85
  validation {
    condition     = var.memory_threshold_upper > 0 && var.memory_threshold_upper <= 100
    error_message = "Memory threshold must be between 0 and 100."
  }
}

variable "memory_threshold_lower" {
  description = "Lower memory utilization threshold for rightsizing (percentage)"
  type        = number
  default     = 25
  validation {
    condition     = var.memory_threshold_lower > 0 && var.memory_threshold_lower <= 100
    error_message = "Memory threshold must be between 0 and 100."
  }
}

variable "evaluation_period_days" {
  description = "Number of days to evaluate for rightsizing recommendations"
  type        = number
  default     = 7
  validation {
    condition     = var.evaluation_period_days > 0 && var.evaluation_period_days <= 30
    error_message = "Evaluation period must be between 1 and 30 days."
  }
}

variable "rightsizing_schedule" {
  description = "Cron schedule for rightsizing analysis"
  type        = string
  default     = "0 6 * * 1" # Weekly on Monday at 6 AM
}

variable "rightsizing_analyzer_image" {
  description = "Container image for rightsizing analyzer"
  type        = string
  default     = "go-coffee/rightsizing-analyzer:latest"
}

# =============================================================================
# AUTOSCALING CONFIGURATION
# =============================================================================

variable "scale_down_delay" {
  description = "Delay before scaling down after scale up"
  type        = string
  default     = "10m"
}

variable "scale_up_delay" {
  description = "Delay before scaling up"
  type        = string
  default     = "0s"
}

variable "max_scale_factor" {
  description = "Maximum scale factor for autoscaling"
  type        = number
  default     = 10
}

variable "cluster_autoscaler_version" {
  description = "Version of cluster autoscaler Helm chart"
  type        = string
  default     = "9.29.0"
}

variable "vpa_chart_version" {
  description = "Version of VPA Helm chart"
  type        = string
  default     = "4.4.2"
}

# =============================================================================
# SCHEDULING CONFIGURATION
# =============================================================================

variable "cost_weight" {
  description = "Weight for cost in scheduling decisions (0-100)"
  type        = number
  default     = 40
  validation {
    condition     = var.cost_weight >= 0 && var.cost_weight <= 100
    error_message = "Cost weight must be between 0 and 100."
  }
}

variable "performance_weight" {
  description = "Weight for performance in scheduling decisions (0-100)"
  type        = number
  default     = 40
  validation {
    condition     = var.performance_weight >= 0 && var.performance_weight <= 100
    error_message = "Performance weight must be between 0 and 100."
  }
}

variable "availability_weight" {
  description = "Weight for availability in scheduling decisions (0-100)"
  type        = number
  default     = 20
  validation {
    condition     = var.availability_weight >= 0 && var.availability_weight <= 100
    error_message = "Availability weight must be between 0 and 100."
  }
}

variable "scheduler_replicas" {
  description = "Number of scheduler replicas"
  type        = number
  default     = 2
}

variable "cost_optimizer_scheduler_image" {
  description = "Container image for cost optimizer scheduler"
  type        = string
  default     = "go-coffee/cost-optimizer-scheduler:latest"
}

# =============================================================================
# KUBECOST CONFIGURATION
# =============================================================================

variable "kubecost_version" {
  description = "Version of KubeCost Helm chart"
  type        = string
  default     = "1.106.2"
}

variable "kubecost_ingress_enabled" {
  description = "Enable ingress for KubeCost"
  type        = bool
  default     = true
}

variable "kubecost_hostname" {
  description = "Hostname for KubeCost ingress"
  type        = string
  default     = "kubecost.go-coffee.local"
}

variable "kubecost_storage_size" {
  description = "Storage size for KubeCost data"
  type        = string
  default     = "32Gi"
}

variable "cloud_provider_api_key" {
  description = "API key for cloud provider cost data"
  type        = string
  default     = ""
  sensitive   = true
}

# =============================================================================
# CLUSTER CONFIGURATION
# =============================================================================

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
}

variable "cloud_provider" {
  description = "Cloud provider (aws, gcp, azure)"
  type        = string
  validation {
    condition     = contains(["aws", "gcp", "azure"], var.cloud_provider)
    error_message = "Cloud provider must be one of: aws, gcp, azure."
  }
}

variable "max_nodes_total" {
  description = "Maximum total number of nodes in the cluster"
  type        = number
  default     = 100
}

variable "min_cores_total" {
  description = "Minimum total CPU cores in the cluster"
  type        = number
  default     = 0
}

variable "max_cores_total" {
  description = "Maximum total CPU cores in the cluster"
  type        = number
  default     = 1000
}

variable "min_memory_total" {
  description = "Minimum total memory in the cluster (GiB)"
  type        = number
  default     = 0
}

variable "max_memory_total" {
  description = "Maximum total memory in the cluster (GiB)"
  type        = number
  default     = 4000
}

# =============================================================================
# AWS CONFIGURATION
# =============================================================================

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_oidc_provider_arn" {
  description = "ARN of the AWS OIDC provider for EKS"
  type        = string
  default     = ""
}

variable "aws_oidc_provider_url" {
  description = "URL of the AWS OIDC provider for EKS"
  type        = string
  default     = ""
}

variable "enable_aws_spot_instances" {
  description = "Enable AWS Spot instances for cost optimization"
  type        = bool
  default     = true
}

variable "enable_aws_reserved_instances" {
  description = "Enable AWS Reserved instances recommendations"
  type        = bool
  default     = true
}

variable "enable_aws_savings_plans" {
  description = "Enable AWS Savings Plans recommendations"
  type        = bool
  default     = true
}

variable "enable_aws_cost_anomaly_detection" {
  description = "Enable AWS Cost Anomaly Detection"
  type        = bool
  default     = true
}

variable "cost_anomaly_threshold" {
  description = "Threshold for cost anomaly alerts (USD)"
  type        = number
  default     = 100
}

variable "cost_alert_email" {
  description = "Email address for cost alerts"
  type        = string
  default     = "platform@go-coffee.com"
}

# =============================================================================
# GCP CONFIGURATION
# =============================================================================

variable "gcp_project_id" {
  description = "Google Cloud project ID"
  type        = string
  default     = ""
}

variable "gcp_region" {
  description = "Google Cloud region"
  type        = string
  default     = "us-central1"
}

variable "enable_gcp_preemptible_instances" {
  description = "Enable GCP Preemptible instances for cost optimization"
  type        = bool
  default     = true
}

variable "enable_gcp_committed_use_discounts" {
  description = "Enable GCP Committed Use Discounts recommendations"
  type        = bool
  default     = true
}

variable "enable_gcp_sustained_use_discounts" {
  description = "Enable GCP Sustained Use Discounts"
  type        = bool
  default     = true
}

# =============================================================================
# AZURE CONFIGURATION
# =============================================================================

variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  default     = ""
}

variable "azure_location" {
  description = "Azure location"
  type        = string
  default     = "East US"
}

variable "enable_azure_spot_instances" {
  description = "Enable Azure Spot instances for cost optimization"
  type        = bool
  default     = true
}

variable "enable_azure_reserved_instances" {
  description = "Enable Azure Reserved instances recommendations"
  type        = bool
  default     = true
}

variable "enable_azure_hybrid_benefit" {
  description = "Enable Azure Hybrid Benefit recommendations"
  type        = bool
  default     = true
}

# =============================================================================
# MONITORING CONFIGURATION
# =============================================================================

variable "prometheus_url" {
  description = "Prometheus server URL"
  type        = string
  default     = "http://prometheus-stack-kube-prom-prometheus:9090"
}

variable "prometheus_fqdn" {
  description = "Prometheus FQDN for external access"
  type        = string
  default     = "prometheus.go-coffee.local"
}

variable "grafana_fqdn" {
  description = "Grafana FQDN for external access"
  type        = string
  default     = "grafana.go-coffee.local"
}

# =============================================================================
# KUBERNETES CONFIGURATION
# =============================================================================

variable "storage_class_name" {
  description = "Storage class name for persistent volumes"
  type        = string
  default     = "gp2"
}

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
# BUDGET AND COST CONTROLS
# =============================================================================

variable "monthly_budget_limit" {
  description = "Monthly budget limit in USD"
  type        = number
  default     = 10000
}

variable "budget_alert_thresholds" {
  description = "Budget alert thresholds (percentages)"
  type        = list(number)
  default     = [50, 75, 90, 100]
}

variable "cost_allocation_tags" {
  description = "Tags for cost allocation"
  type        = map(string)
  default = {
    "CostCenter" = "platform"
    "Team"       = "engineering"
    "Project"    = "go-coffee"
  }
}

variable "enable_cost_controls" {
  description = "Enable automated cost controls"
  type        = bool
  default     = true
}

variable "max_daily_spend" {
  description = "Maximum daily spend threshold in USD"
  type        = number
  default     = 500
}

# =============================================================================
# OPTIMIZATION POLICIES
# =============================================================================

variable "optimization_policies" {
  description = "Cost optimization policies"
  type = map(object({
    enabled = bool
    threshold = number
    action = string
  }))
  default = {
    "idle_resources" = {
      enabled = true
      threshold = 5  # 5% utilization
      action = "recommend_termination"
    }
    "oversized_resources" = {
      enabled = true
      threshold = 20  # 20% utilization
      action = "recommend_downsize"
    }
    "unused_volumes" = {
      enabled = true
      threshold = 7  # 7 days unused
      action = "recommend_deletion"
    }
    "unattached_volumes" = {
      enabled = true
      threshold = 1  # 1 day unattached
      action = "recommend_deletion"
    }
  }
}

variable "auto_apply_recommendations" {
  description = "Automatically apply cost optimization recommendations"
  type        = bool
  default     = false
}

variable "recommendation_approval_required" {
  description = "Require approval for cost optimization recommendations"
  type        = bool
  default     = true
}

# =============================================================================
# NOTIFICATION CONFIGURATION
# =============================================================================

variable "slack_webhook_url" {
  description = "Slack webhook URL for cost alerts"
  type        = string
  default     = ""
  sensitive   = true
}

variable "email_notifications" {
  description = "Email addresses for cost notifications"
  type        = list(string)
  default     = ["platform@go-coffee.com"]
}

variable "webhook_url" {
  description = "Generic webhook URL for cost alerts"
  type        = string
  default     = ""
  sensitive   = true
}
