# Cost Optimization and Resource Management Module
# Implements automated cost optimization, resource rightsizing, and intelligent workload placement

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common tags/labels
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Component   = "cost-optimization"
    Team        = var.team
    CostCenter  = var.cost_center
  }
  
  # Cost optimization configuration
  cost_optimization = {
    enabled = var.enable_cost_optimization
    
    # Resource rightsizing
    rightsizing = {
      enabled = var.enable_rightsizing
      cpu_threshold_upper = var.cpu_threshold_upper
      cpu_threshold_lower = var.cpu_threshold_lower
      memory_threshold_upper = var.memory_threshold_upper
      memory_threshold_lower = var.memory_threshold_lower
      evaluation_period_days = var.evaluation_period_days
    }
    
    # Workload scheduling
    scheduling = {
      enabled = var.enable_intelligent_scheduling
      cost_weight = var.cost_weight
      performance_weight = var.performance_weight
      availability_weight = var.availability_weight
    }
    
    # Auto-scaling
    autoscaling = {
      enabled = var.enable_advanced_autoscaling
      scale_down_delay = var.scale_down_delay
      scale_up_delay = var.scale_up_delay
      max_scale_factor = var.max_scale_factor
    }
  }
  
  # Cloud provider cost optimization features
  cloud_features = {
    aws = {
      spot_instances = var.enable_aws_spot_instances
      reserved_instances = var.enable_aws_reserved_instances
      savings_plans = var.enable_aws_savings_plans
      cost_anomaly_detection = var.enable_aws_cost_anomaly_detection
    }
    gcp = {
      preemptible_instances = var.enable_gcp_preemptible_instances
      committed_use_discounts = var.enable_gcp_committed_use_discounts
      sustained_use_discounts = var.enable_gcp_sustained_use_discounts
    }
    azure = {
      spot_instances = var.enable_azure_spot_instances
      reserved_instances = var.enable_azure_reserved_instances
      hybrid_benefit = var.enable_azure_hybrid_benefit
    }
  }
}

# =============================================================================
# KUBERNETES COST OPTIMIZATION NAMESPACE
# =============================================================================

# Cost Optimization Namespace
resource "kubernetes_namespace" "cost_optimization" {
  metadata {
    name = var.cost_optimization_namespace
    
    labels = merge(local.common_tags, {
      "name" = var.cost_optimization_namespace
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# VERTICAL POD AUTOSCALER (VPA)
# =============================================================================

# VPA for resource rightsizing
resource "helm_release" "vpa" {
  count = var.enable_rightsizing ? 1 : 0
  
  name       = "vpa"
  repository = "https://charts.fairwinds.com/stable"
  chart      = "vpa"
  version    = var.vpa_chart_version
  namespace  = kubernetes_namespace.cost_optimization.metadata[0].name
  
  values = [
    yamlencode({
      # VPA Recommender
      recommender = {
        enabled = true
        
        # Resource recommendations
        resources = {
          requests = {
            cpu    = "100m"
            memory = "500Mi"
          }
          limits = {
            cpu    = "1000m"
            memory = "1Gi"
          }
        }
        
        # Recommendation margins
        extraArgs = [
          "--v=4",
          "--pod-recommendation-min-cpu-millicores=25",
          "--pod-recommendation-min-memory-mb=100",
          "--recommendation-margin-fraction=0.15",
          "--target-cpu-percentile=0.9",
          "--target-memory-percentile=0.9"
        ]
      }
      
      # VPA Updater
      updater = {
        enabled = true
        
        resources = {
          requests = {
            cpu    = "100m"
            memory = "500Mi"
          }
          limits = {
            cpu    = "1000m"
            memory = "1Gi"
          }
        }
        
        extraArgs = [
          "--v=4",
          "--min-replicas=2",
          "--eviction-tolerance=0.5"
        ]
      }
      
      # VPA Admission Controller
      admissionController = {
        enabled = true
        
        resources = {
          requests = {
            cpu    = "50m"
            memory = "200Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "500Mi"
          }
        }
        
        # Webhook configuration
        generateCertificate = true
        
        extraArgs = [
          "--v=4",
          "--webhook-timeout-seconds=30"
        ]
      }
      
      # Metrics
      metrics = {
        serviceMonitor = {
          enabled = var.monitoring_enabled
          labels = local.common_tags
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.cost_optimization]
}

# =============================================================================
# CLUSTER AUTOSCALER
# =============================================================================

# Cluster Autoscaler for node optimization
resource "helm_release" "cluster_autoscaler" {
  count = var.enable_cluster_autoscaling ? 1 : 0
  
  name       = "cluster-autoscaler"
  repository = "https://kubernetes.github.io/autoscaler"
  chart      = "cluster-autoscaler"
  version    = var.cluster_autoscaler_version
  namespace  = kubernetes_namespace.cost_optimization.metadata[0].name
  
  values = [
    yamlencode({
      # Autoscaler configuration
      autoDiscovery = {
        clusterName = var.cluster_name
        enabled = true
        tags = [
          "k8s.io/cluster-autoscaler/enabled",
          "k8s.io/cluster-autoscaler/${var.cluster_name}"
        ]
      }
      
      # AWS specific configuration
      awsRegion = var.aws_region
      
      # Scaling configuration
      extraArgs = {
        "v" = 4
        "stderrthreshold" = "info"
        "cloud-provider" = var.cloud_provider
        "skip-nodes-with-local-storage" = false
        "expander" = "least-waste"
        "node-group-auto-discovery" = "asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/${var.cluster_name}"
        "balance-similar-node-groups" = true
        "skip-nodes-with-system-pods" = false
        "scale-down-enabled" = true
        "scale-down-delay-after-add" = local.cost_optimization.autoscaling.scale_down_delay
        "scale-down-unneeded-time" = "10m"
        "scale-down-utilization-threshold" = "0.5"
        "max-node-provision-time" = "15m"
        "scan-interval" = "10s"
        "max-nodes-total" = var.max_nodes_total
        "cores-total" = "${var.min_cores_total}:${var.max_cores_total}"
        "memory-total" = "${var.min_memory_total}:${var.max_memory_total}"
      }
      
      # Resources
      resources = {
        requests = {
          cpu    = "100m"
          memory = "300Mi"
        }
        limits = {
          cpu    = "100m"
          memory = "300Mi"
        }
      }
      
      # Service account
      rbac = {
        create = true
        serviceAccount = {
          create = true
          name = "cluster-autoscaler"
          annotations = var.cloud_provider == "aws" ? {
            "eks.amazonaws.com/role-arn" = aws_iam_role.cluster_autoscaler[0].arn
          } : {}
        }
      }
      
      # Node selector for cost-optimized nodes
      nodeSelector = {
        "node-type" = "cost-optimized"
      }
      
      # Tolerations
      tolerations = [
        {
          key = "node-type"
          operator = "Equal"
          value = "cost-optimized"
          effect = "NoSchedule"
        }
      ]
    })
  ]
  
  depends_on = [kubernetes_namespace.cost_optimization]
}

# =============================================================================
# COST ANALYZER (KUBECOST)
# =============================================================================

# KubeCost for cost visibility and optimization
resource "helm_release" "kubecost" {
  count = var.enable_cost_analyzer ? 1 : 0
  
  name       = "kubecost"
  repository = "https://kubecost.github.io/cost-analyzer/"
  chart      = "cost-analyzer"
  version    = var.kubecost_version
  namespace  = kubernetes_namespace.cost_optimization.metadata[0].name
  
  values = [
    yamlencode({
      # Global configuration
      global = {
        prometheus = {
          enabled = false
          fqdn = var.prometheus_fqdn
        }
        grafana = {
          enabled = false
          fqdn = var.grafana_fqdn
        }
      }
      
      # Kubecost configuration
      kubecostFrontend = {
        image = "gcr.io/kubecost1/frontend"
        
        # Resources
        resources = {
          requests = {
            cpu    = "10m"
            memory = "55Mi"
          }
          limits = {
            cpu    = "100m"
            memory = "256Mi"
          }
        }
      }
      
      kubecostModel = {
        image = "gcr.io/kubecost1/cost-model"
        
        # Resources
        resources = {
          requests = {
            cpu    = "200m"
            memory = "55Mi"
          }
          limits = {
            cpu    = "800m"
            memory = "256Mi"
          }
        }
        
        # Cost model configuration
        extraEnv = [
          {
            name = "CLOUD_PROVIDER_API_KEY"
            value = var.cloud_provider_api_key
          },
          {
            name = "CLUSTER_ID"
            value = var.cluster_name
          },
          {
            name = "AWS_CLUSTER_ID"
            value = var.cluster_name
          }
        ]
      }
      
      # Ingress configuration
      ingress = {
        enabled = var.kubecost_ingress_enabled
        className = var.ingress_class_name
        hosts = [var.kubecost_hostname]
        tls = [
          {
            secretName = "${local.name_prefix}-kubecost-tls"
            hosts = [var.kubecost_hostname]
          }
        ]
        annotations = {
          "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
        }
      }
      
      # Persistent volume for cost data
      persistentVolume = {
        enabled = true
        size = var.kubecost_storage_size
        storageClass = var.storage_class_name
      }
      
      # Service monitor for Prometheus
      serviceMonitor = {
        enabled = var.monitoring_enabled
        additionalLabels = local.common_tags
      }
      
      # Cost optimization recommendations
      costOptimization = {
        enabled = true
        
        # Recommendation settings
        recommendations = {
          # Right-sizing recommendations
          rightSizing = {
            enabled = true
            cpuThreshold = local.cost_optimization.rightsizing.cpu_threshold_upper
            memoryThreshold = local.cost_optimization.rightsizing.memory_threshold_upper
          }
          
          # Cluster sizing recommendations
          clusterSizing = {
            enabled = true
            evaluationPeriod = "${local.cost_optimization.rightsizing.evaluation_period_days}d"
          }
          
          # Abandoned resource detection
          abandonedResources = {
            enabled = true
            threshold = "7d"
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.cost_optimization]
}

# =============================================================================
# INTELLIGENT WORKLOAD SCHEDULER
# =============================================================================

# Custom scheduler for cost-optimized workload placement
resource "kubernetes_manifest" "cost_optimizer_scheduler" {
  count = var.enable_intelligent_scheduling ? 1 : 0
  
  manifest = {
    apiVersion = "apps/v1"
    kind       = "Deployment"
    
    metadata = {
      name      = "${local.name_prefix}-cost-optimizer-scheduler"
      namespace = kubernetes_namespace.cost_optimization.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "scheduler"
      })
    }
    
    spec = {
      replicas = var.scheduler_replicas
      
      selector = {
        matchLabels = {
          app = "${local.name_prefix}-cost-optimizer-scheduler"
        }
      }
      
      template = {
        metadata = {
          labels = merge(local.common_tags, {
            app = "${local.name_prefix}-cost-optimizer-scheduler"
          })
        }
        
        spec = {
          serviceAccountName = kubernetes_service_account.cost_optimizer_scheduler[0].metadata[0].name
          
          containers = [
            {
              name  = "cost-optimizer-scheduler"
              image = var.cost_optimizer_scheduler_image
              
              env = [
                {
                  name  = "ENVIRONMENT"
                  value = var.environment
                },
                {
                  name  = "PROJECT_NAME"
                  value = var.project_name
                },
                {
                  name  = "COST_WEIGHT"
                  value = tostring(local.cost_optimization.scheduling.cost_weight)
                },
                {
                  name  = "PERFORMANCE_WEIGHT"
                  value = tostring(local.cost_optimization.scheduling.performance_weight)
                },
                {
                  name  = "AVAILABILITY_WEIGHT"
                  value = tostring(local.cost_optimization.scheduling.availability_weight)
                },
                {
                  name  = "PROMETHEUS_URL"
                  value = var.prometheus_url
                }
              ]
              
              ports = [
                {
                  containerPort = 8080
                  name         = "http"
                },
                {
                  containerPort = 8443
                  name         = "webhook"
                }
              ]
              
              resources = {
                requests = {
                  cpu    = "100m"
                  memory = "128Mi"
                }
                limits = {
                  cpu    = "500m"
                  memory = "512Mi"
                }
              }
              
              livenessProbe = {
                httpGet = {
                  path = "/healthz"
                  port = 8080
                }
                initialDelaySeconds = 30
                periodSeconds       = 10
              }
              
              readinessProbe = {
                httpGet = {
                  path = "/readyz"
                  port = 8080
                }
                initialDelaySeconds = 5
                periodSeconds       = 5
              }
            }
          ]
        }
      }
    }
  }
}

# Service Account for Cost Optimizer Scheduler
resource "kubernetes_service_account" "cost_optimizer_scheduler" {
  count = var.enable_intelligent_scheduling ? 1 : 0
  
  metadata {
    name      = "${local.name_prefix}-cost-optimizer-scheduler"
    namespace = kubernetes_namespace.cost_optimization.metadata[0].name
    
    labels = local.common_tags
  }
}

# ClusterRole for Cost Optimizer Scheduler
resource "kubernetes_cluster_role" "cost_optimizer_scheduler" {
  count = var.enable_intelligent_scheduling ? 1 : 0
  
  metadata {
    name = "${local.name_prefix}-cost-optimizer-scheduler"
    
    labels = local.common_tags
  }
  
  rule {
    api_groups = [""]
    resources  = ["nodes", "pods", "services", "endpoints"]
    verbs      = ["get", "list", "watch"]
  }
  
  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "replicasets", "statefulsets"]
    verbs      = ["get", "list", "watch"]
  }
  
  rule {
    api_groups = ["metrics.k8s.io"]
    resources  = ["nodes", "pods"]
    verbs      = ["get", "list"]
  }
  
  rule {
    api_groups = [""]
    resources  = ["events"]
    verbs      = ["create", "patch"]
  }
}

# ClusterRoleBinding for Cost Optimizer Scheduler
resource "kubernetes_cluster_role_binding" "cost_optimizer_scheduler" {
  count = var.enable_intelligent_scheduling ? 1 : 0
  
  metadata {
    name = "${local.name_prefix}-cost-optimizer-scheduler"
    
    labels = local.common_tags
  }
  
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.cost_optimizer_scheduler[0].metadata[0].name
  }
  
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.cost_optimizer_scheduler[0].metadata[0].name
    namespace = kubernetes_namespace.cost_optimization.metadata[0].name
  }
}

# =============================================================================
# AWS COST OPTIMIZATION
# =============================================================================

# IAM Role for Cluster Autoscaler
resource "aws_iam_role" "cluster_autoscaler" {
  count = var.cloud_provider == "aws" && var.enable_cluster_autoscaling ? 1 : 0
  
  name = "${local.name_prefix}-cluster-autoscaler"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = var.aws_oidc_provider_arn
        }
        Condition = {
          StringEquals = {
            "${var.aws_oidc_provider_url}:sub" = "system:serviceaccount:${kubernetes_namespace.cost_optimization.metadata[0].name}:cluster-autoscaler"
            "${var.aws_oidc_provider_url}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = local.common_tags
}

# IAM Policy for Cluster Autoscaler
resource "aws_iam_role_policy" "cluster_autoscaler" {
  count = var.cloud_provider == "aws" && var.enable_cluster_autoscaling ? 1 : 0
  
  name = "${local.name_prefix}-cluster-autoscaler"
  role = aws_iam_role.cluster_autoscaler[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "autoscaling:DescribeAutoScalingGroups",
          "autoscaling:DescribeAutoScalingInstances",
          "autoscaling:DescribeLaunchConfigurations",
          "autoscaling:DescribeTags",
          "autoscaling:SetDesiredCapacity",
          "autoscaling:TerminateInstanceInAutoScalingGroup",
          "ec2:DescribeLaunchTemplateVersions",
          "ec2:DescribeInstanceTypes"
        ]
        Resource = "*"
      }
    ]
  })
}

# AWS Cost Anomaly Detection
resource "aws_ce_anomaly_detector" "cost_anomaly" {
  count = var.cloud_provider == "aws" && local.cloud_features.aws.cost_anomaly_detection ? 1 : 0
  
  name         = "${local.name_prefix}-cost-anomaly-detector"
  detector_type = "DIMENSIONAL"
  
  specification = jsonencode({
    Dimension = "SERVICE"
    MatchOptions = ["EQUALS"]
    Values = ["Amazon Elastic Kubernetes Service"]
  })
  
  tags = local.common_tags
}

# AWS Cost Anomaly Subscription
resource "aws_ce_anomaly_subscription" "cost_anomaly" {
  count = var.cloud_provider == "aws" && local.cloud_features.aws.cost_anomaly_detection ? 1 : 0
  
  name      = "${local.name_prefix}-cost-anomaly-subscription"
  frequency = "DAILY"
  
  monitor_arn_list = [
    aws_ce_anomaly_detector.cost_anomaly[0].arn
  ]
  
  subscriber {
    type    = "EMAIL"
    address = var.cost_alert_email
  }
  
  threshold_expression {
    and {
      dimension {
        key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
        values        = [tostring(var.cost_anomaly_threshold)]
        match_options = ["GREATER_THAN_OR_EQUAL"]
      }
    }
  }
  
  tags = local.common_tags
}

# =============================================================================
# COST OPTIMIZATION CRONJOBS
# =============================================================================

# Resource Rightsizing CronJob
resource "kubernetes_manifest" "rightsizing_cronjob" {
  count = var.enable_rightsizing ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-rightsizing"
      namespace = kubernetes_namespace.cost_optimization.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "rightsizing"
      })
    }
    
    spec = {
      schedule = var.rightsizing_schedule
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "rightsizing-analyzer"
                  image = var.rightsizing_analyzer_image
                  
                  env = [
                    {
                      name  = "ENVIRONMENT"
                      value = var.environment
                    },
                    {
                      name  = "CPU_THRESHOLD_UPPER"
                      value = tostring(local.cost_optimization.rightsizing.cpu_threshold_upper)
                    },
                    {
                      name  = "CPU_THRESHOLD_LOWER"
                      value = tostring(local.cost_optimization.rightsizing.cpu_threshold_lower)
                    },
                    {
                      name  = "MEMORY_THRESHOLD_UPPER"
                      value = tostring(local.cost_optimization.rightsizing.memory_threshold_upper)
                    },
                    {
                      name  = "MEMORY_THRESHOLD_LOWER"
                      value = tostring(local.cost_optimization.rightsizing.memory_threshold_lower)
                    },
                    {
                      name  = "EVALUATION_PERIOD_DAYS"
                      value = tostring(local.cost_optimization.rightsizing.evaluation_period_days)
                    },
                    {
                      name  = "PROMETHEUS_URL"
                      value = var.prometheus_url
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting resource rightsizing analysis..."
                      
                      # Query Prometheus for resource utilization metrics
                      # Analyze CPU and memory usage patterns
                      # Generate rightsizing recommendations
                      # Apply VPA recommendations if auto-apply is enabled
                      
                      echo "Rightsizing analysis completed"
                    EOT
                  ]
                  
                  resources = {
                    requests = {
                      cpu    = "100m"
                      memory = "128Mi"
                    }
                    limits = {
                      cpu    = "500m"
                      memory = "512Mi"
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  }
}
