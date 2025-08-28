# CI/CD Pipeline Enhancement Module
# Implements advanced CI/CD pipelines with multi-cloud deployment, security scanning, and automated testing

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
    github = {
      source  = "integrations/github"
      version = "~> 5.0"
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
    Component   = "cicd"
    Team        = var.team
    CostCenter  = var.cost_center
  }
  
  # CI/CD configuration
  cicd_config = {
    enabled = var.enable_cicd_enhancement
    
    # Pipeline stages
    stages = {
      build = var.enable_build_stage
      test = var.enable_test_stage
      security_scan = var.enable_security_scan_stage
      deploy = var.enable_deploy_stage
      integration_test = var.enable_integration_test_stage
      performance_test = var.enable_performance_test_stage
    }
    
    # Deployment strategies
    deployment = {
      strategy = var.deployment_strategy
      canary_percentage = var.canary_deployment_percentage
      blue_green_enabled = var.enable_blue_green_deployment
      rollback_enabled = var.enable_automated_rollback
    }
    
    # Quality gates
    quality_gates = {
      code_coverage_threshold = var.code_coverage_threshold
      security_scan_threshold = var.security_scan_threshold
      performance_threshold = var.performance_threshold
    }
  }
  
  # Multi-cloud deployment targets
  deployment_targets = {
    aws = var.enable_aws_deployment
    gcp = var.enable_gcp_deployment
    azure = var.enable_azure_deployment
  }
}

# =============================================================================
# KUBERNETES CI/CD NAMESPACE
# =============================================================================

# CI/CD Namespace
resource "kubernetes_namespace" "cicd" {
  metadata {
    name = var.cicd_namespace
    
    labels = merge(local.common_tags, {
      "name" = var.cicd_namespace
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# ARGO CD GITOPS
# =============================================================================

# ArgoCD for GitOps deployment
resource "helm_release" "argocd" {
  count = var.enable_argocd ? 1 : 0
  
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = var.argocd_chart_version
  namespace  = kubernetes_namespace.cicd.metadata[0].name
  
  values = [
    yamlencode({
      # Global configuration
      global = {
        image = {
          repository = "quay.io/argoproj/argocd"
          tag = var.argocd_version
        }
      }
      
      # Controller configuration
      controller = {
        replicas = var.argocd_controller_replicas
        
        resources = {
          requests = {
            cpu    = "250m"
            memory = "1Gi"
          }
          limits = {
            cpu    = "500m"
            memory = "2Gi"
          }
        }
        
        # Metrics
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
            additionalLabels = local.common_tags
          }
        }
      }
      
      # Server configuration
      server = {
        replicas = var.argocd_server_replicas
        
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
        
        # Ingress configuration
        ingress = {
          enabled = var.argocd_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [var.argocd_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-argocd-tls"
              hosts = [var.argocd_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
            "nginx.ingress.kubernetes.io/ssl-redirect" = "true"
            "nginx.ingress.kubernetes.io/backend-protocol" = "GRPC"
          }
        }
        
        # Configuration
        config = {
          "application.instanceLabelKey" = "argocd.argoproj.io/instance"
          "server.rbac.log.enforce.enable" = "true"
          "policy.default" = "role:readonly"
          "policy.csv" = <<-EOF
            p, role:admin, applications, *, */*, allow
            p, role:admin, clusters, *, *, allow
            p, role:admin, repositories, *, *, allow
            g, ${var.project_name}:admin, role:admin
          EOF
          
          # Repository credentials
          "repositories" = yamlencode([
            {
              url = var.git_repository_url
              passwordSecret = {
                name = "git-credentials"
                key = "password"
              }
              usernameSecret = {
                name = "git-credentials"
                key = "username"
              }
            }
          ])
        }
        
        # Metrics
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
            additionalLabels = local.common_tags
          }
        }
      }
      
      # Repository server configuration
      repoServer = {
        replicas = var.argocd_repo_server_replicas
        
        resources = {
          requests = {
            cpu    = "100m"
            memory = "256Mi"
          }
          limits = {
            cpu    = "500m"
            memory = "1Gi"
          }
        }
        
        # Metrics
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
            additionalLabels = local.common_tags
          }
        }
      }
      
      # Redis configuration
      redis = {
        enabled = true
        
        resources = {
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "256Mi"
          }
        }
      }
      
      # ApplicationSet controller
      applicationSet = {
        enabled = var.enable_argocd_applicationset
        
        replicas = 1
        
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
      
      # Notifications
      notifications = {
        enabled = var.enable_argocd_notifications
        
        argocdUrl = "https://${var.argocd_hostname}"
        
        subscriptions = [
          {
            recipients = ["slack:${var.slack_channel}"]
            triggers = ["on-deployed", "on-health-degraded", "on-sync-failed"]
          }
        ]
        
        services = {
          service.slack = {
            token = "$slack-token"
          }
        }
        
        templates = {
          "template.app-deployed" = {
            message = "Application {{.app.metadata.name}} is now running new version."
            slack = {
              attachments = "[{\"title\": \"{{.app.metadata.name}}\", \"color\": \"good\", \"fields\": [{\"title\": \"Sync Status\", \"value\": \"{{.app.status.sync.status}}\", \"short\": true}, {\"title\": \"Repository\", \"value\": \"{{.app.spec.source.repoURL}}\", \"short\": true}]}]"
            }
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.cicd]
}

# =============================================================================
# TEKTON PIPELINES
# =============================================================================

# Tekton Pipelines for CI/CD
resource "helm_release" "tekton_pipelines" {
  count = var.enable_tekton_pipelines ? 1 : 0
  
  name       = "tekton-pipelines"
  repository = "https://cdfoundation.github.io/tekton-helm-chart"
  chart      = "tekton-pipeline"
  version    = var.tekton_chart_version
  namespace  = kubernetes_namespace.cicd.metadata[0].name
  
  values = [
    yamlencode({
      # Pipeline configuration
      pipeline = {
        images = {
          entrypoint = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/entrypoint:${var.tekton_version}"
          nop = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/nop:${var.tekton_version}"
          gitInit = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:${var.tekton_version}"
          workingDirInit = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/workingdirinit:${var.tekton_version}"
        }
        
        # Resource requirements
        resources = {
          requests = {
            cpu = "100m"
            memory = "100Mi"
          }
          limits = {
            cpu = "500m"
            memory = "500Mi"
          }
        }
      }
      
      # Controller configuration
      controller = {
        replicas = var.tekton_controller_replicas
        
        resources = {
          requests = {
            cpu = "100m"
            memory = "100Mi"
          }
          limits = {
            cpu = "1000m"
            memory = "1000Mi"
          }
        }
        
        # Metrics
        metrics = {
          enabled = var.monitoring_enabled
        }
      }
      
      # Webhook configuration
      webhook = {
        replicas = var.tekton_webhook_replicas
        
        resources = {
          requests = {
            cpu = "100m"
            memory = "100Mi"
          }
          limits = {
            cpu = "500m"
            memory = "500Mi"
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.cicd]
}

# =============================================================================
# GITHUB ACTIONS RUNNER
# =============================================================================

# GitHub Actions Runner Controller
resource "helm_release" "actions_runner_controller" {
  count = var.enable_github_actions_runner ? 1 : 0
  
  name       = "actions-runner-controller"
  repository = "https://actions-runner-controller.github.io/actions-runner-controller"
  chart      = "actions-runner-controller"
  version    = var.actions_runner_controller_version
  namespace  = kubernetes_namespace.cicd.metadata[0].name
  
  values = [
    yamlencode({
      # Authentication
      authSecret = {
        create = true
        github_token = var.github_token
      }
      
      # Controller configuration
      replicaCount = var.actions_runner_controller_replicas
      
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
      
      # Metrics
      metrics = {
        serviceMonitor = {
          enabled = var.monitoring_enabled
          additionalLabels = local.common_tags
        }
      }
      
      # Webhook server
      githubWebhookServer = {
        enabled = var.enable_github_webhook_server
        
        ingress = {
          enabled = var.github_webhook_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [
            {
              host = var.github_webhook_hostname
              paths = [
                {
                  path = "/"
                  pathType = "Prefix"
                }
              ]
            }
          ]
          tls = [
            {
              secretName = "${local.name_prefix}-github-webhook-tls"
              hosts = [var.github_webhook_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.cicd]
}

# GitHub Actions Runner Deployment
resource "kubernetes_manifest" "github_runner_deployment" {
  count = var.enable_github_actions_runner ? 1 : 0
  
  manifest = {
    apiVersion = "actions.summerwind.dev/v1alpha1"
    kind       = "RunnerDeployment"
    
    metadata = {
      name      = "${local.name_prefix}-github-runners"
      namespace = kubernetes_namespace.cicd.metadata[0].name
      
      labels = local.common_tags
    }
    
    spec = {
      replicas = var.github_runner_replicas
      
      template = {
        spec = {
          repository = var.git_repository_url
          
          # Runner configuration
          labels = [
            "go-coffee",
            "kubernetes",
            "self-hosted"
          ]
          
          # Resources
          resources = {
            requests = {
              cpu    = "500m"
              memory = "1Gi"
            }
            limits = {
              cpu    = "2000m"
              memory = "4Gi"
            }
          }
          
          # Docker in Docker
          dockerdWithinRunnerContainer = true
          
          # Environment variables
          env = [
            {
              name  = "ENVIRONMENT"
              value = var.environment
            },
            {
              name  = "PROJECT_NAME"
              value = var.project_name
            }
          ]
          
          # Volume mounts for caching
          volumeMounts = [
            {
              name      = "docker-cache"
              mountPath = "/var/lib/docker"
            }
          ]
          
          volumes = [
            {
              name = "docker-cache"
              emptyDir = {
                sizeLimit = "10Gi"
              }
            }
          ]
        }
      }
    }
  }
  
  depends_on = [helm_release.actions_runner_controller]
}

# =============================================================================
# CI/CD PIPELINE TEMPLATES
# =============================================================================

# Tekton Pipeline for Go Coffee services
resource "kubernetes_manifest" "go_coffee_pipeline" {
  count = var.enable_tekton_pipelines ? 1 : 0
  
  manifest = {
    apiVersion = "tekton.dev/v1beta1"
    kind       = "Pipeline"
    
    metadata = {
      name      = "${local.name_prefix}-pipeline"
      namespace = kubernetes_namespace.cicd.metadata[0].name
      
      labels = local.common_tags
    }
    
    spec = {
      params = [
        {
          name = "git-url"
          type = "string"
          description = "Git repository URL"
        },
        {
          name = "git-revision"
          type = "string"
          description = "Git revision"
          default = "main"
        },
        {
          name = "image-name"
          type = "string"
          description = "Container image name"
        },
        {
          name = "deployment-name"
          type = "string"
          description = "Kubernetes deployment name"
        }
      ]
      
      workspaces = [
        {
          name = "shared-data"
          description = "Shared workspace for pipeline tasks"
        },
        {
          name = "docker-credentials"
          description = "Docker registry credentials"
        }
      ]
      
      tasks = [
        # Git clone task
        {
          name = "fetch-source"
          taskRef = {
            name = "git-clone"
            kind = "ClusterTask"
          }
          workspaces = [
            {
              name = "output"
              workspace = "shared-data"
            }
          ]
          params = [
            {
              name = "url"
              value = "$(params.git-url)"
            },
            {
              name = "revision"
              value = "$(params.git-revision)"
            }
          ]
        },
        
        # Build and test
        {
          name = "build-and-test"
          runAfter = ["fetch-source"]
          taskSpec = {
            workspaces = [
              {
                name = "source"
                description = "Source code workspace"
              }
            ]
            steps = [
              {
                name = "build"
                image = "golang:1.21"
                workingDir = "$(workspaces.source.path)"
                script = <<-EOF
                  #!/bin/bash
                  set -e
                  
                  echo "Building Go application..."
                  go mod download
                  go build -v ./...
                  
                  echo "Running tests..."
                  go test -v -race -coverprofile=coverage.out ./...
                  
                  echo "Checking code coverage..."
                  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
                  echo "Code coverage: $COVERAGE%"
                  
                  if (( $(echo "$COVERAGE < ${local.cicd_config.quality_gates.code_coverage_threshold}" | bc -l) )); then
                    echo "Code coverage $COVERAGE% is below threshold ${local.cicd_config.quality_gates.code_coverage_threshold}%"
                    exit 1
                  fi
                EOF
              }
            ]
          }
          workspaces = [
            {
              name = "source"
              workspace = "shared-data"
            }
          ]
        },
        
        # Security scan
        {
          name = "security-scan"
          runAfter = ["build-and-test"]
          taskSpec = {
            workspaces = [
              {
                name = "source"
                description = "Source code workspace"
              }
            ]
            steps = [
              {
                name = "trivy-scan"
                image = "aquasec/trivy:latest"
                workingDir = "$(workspaces.source.path)"
                script = <<-EOF
                  #!/bin/bash
                  set -e
                  
                  echo "Running Trivy security scan..."
                  trivy fs --exit-code 1 --severity HIGH,CRITICAL .
                EOF
              }
            ]
          }
          workspaces = [
            {
              name = "source"
              workspace = "shared-data"
            }
          ]
        },
        
        # Build container image
        {
          name = "build-image"
          runAfter = ["security-scan"]
          taskRef = {
            name = "buildah"
            kind = "ClusterTask"
          }
          workspaces = [
            {
              name = "source"
              workspace = "shared-data"
            },
            {
              name = "dockerconfig"
              workspace = "docker-credentials"
            }
          ]
          params = [
            {
              name = "IMAGE"
              value = "$(params.image-name):$(params.git-revision)"
            },
            {
              name = "DOCKERFILE"
              value = "./Dockerfile"
            }
          ]
        },
        
        # Deploy to Kubernetes
        {
          name = "deploy"
          runAfter = ["build-image"]
          taskSpec = {
            params = [
              {
                name = "deployment-name"
                type = "string"
              },
              {
                name = "image-name"
                type = "string"
              },
              {
                name = "git-revision"
                type = "string"
              }
            ]
            steps = [
              {
                name = "deploy"
                image = "bitnami/kubectl:latest"
                script = <<-EOF
                  #!/bin/bash
                  set -e
                  
                  echo "Deploying to Kubernetes..."
                  kubectl set image deployment/$(params.deployment-name) \
                    $(params.deployment-name)=$(params.image-name):$(params.git-revision) \
                    -n go-coffee
                  
                  echo "Waiting for rollout to complete..."
                  kubectl rollout status deployment/$(params.deployment-name) -n go-coffee --timeout=300s
                  
                  echo "Deployment completed successfully"
                EOF
              }
            ]
          }
          params = [
            {
              name = "deployment-name"
              value = "$(params.deployment-name)"
            },
            {
              name = "image-name"
              value = "$(params.image-name)"
            },
            {
              name = "git-revision"
              value = "$(params.git-revision)"
            }
          ]
        }
      ]
    }
  }
}

# =============================================================================
# ARGOCD APPLICATIONS
# =============================================================================

# ArgoCD Application for Go Coffee services
resource "kubernetes_manifest" "argocd_application" {
  count = var.enable_argocd ? 1 : 0

  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"

    metadata = {
      name      = "${local.name_prefix}-app"
      namespace = kubernetes_namespace.cicd.metadata[0].name

      labels = local.common_tags

      finalizers = ["resources-finalizer.argocd.argoproj.io"]
    }

    spec = {
      project = "default"

      source = {
        repoURL        = var.git_repository_url
        targetRevision = var.git_target_revision
        path           = var.k8s_manifests_path

        # Helm configuration if using Helm charts
        helm = var.use_helm_charts ? {
          valueFiles = ["values-${var.environment}.yaml"]

          parameters = [
            {
              name  = "image.tag"
              value = var.git_target_revision
            },
            {
              name  = "environment"
              value = var.environment
            }
          ]
        } : null
      }

      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = "go-coffee"
      }

      syncPolicy = {
        automated = var.enable_auto_sync ? {
          prune    = true
          selfHeal = true
        } : null

        syncOptions = [
          "CreateNamespace=true",
          "PrunePropagationPolicy=foreground",
          "PruneLast=true"
        ]

        retry = {
          limit = 5
          backoff = {
            duration    = "5s"
            factor      = 2
            maxDuration = "3m"
          }
        }
      }

      # Health checks
      ignoreDifferences = [
        {
          group = "apps"
          kind  = "Deployment"
          jsonPointers = ["/spec/replicas"]
        }
      ]
    }
  }

  depends_on = [helm_release.argocd]
}

# =============================================================================
# SECRETS MANAGEMENT
# =============================================================================

# Git credentials secret
resource "kubernetes_secret" "git_credentials" {
  metadata {
    name      = "git-credentials"
    namespace = kubernetes_namespace.cicd.metadata[0].name

    labels = local.common_tags
  }

  data = {
    username = base64encode(var.git_username)
    password = base64encode(var.git_token)
  }

  type = "Opaque"
}

# Docker registry credentials
resource "kubernetes_secret" "docker_credentials" {
  metadata {
    name      = "docker-credentials"
    namespace = kubernetes_namespace.cicd.metadata[0].name

    labels = local.common_tags
  }

  data = {
    ".dockerconfigjson" = base64encode(jsonencode({
      auths = {
        (var.docker_registry_server) = {
          username = var.docker_registry_username
          password = var.docker_registry_password
          auth     = base64encode("${var.docker_registry_username}:${var.docker_registry_password}")
        }
      }
    }))
  }

  type = "kubernetes.io/dockerconfigjson"
}

# Slack notification secret
resource "kubernetes_secret" "slack_token" {
  count = var.enable_argocd_notifications ? 1 : 0

  metadata {
    name      = "slack-token"
    namespace = kubernetes_namespace.cicd.metadata[0].name

    labels = local.common_tags
  }

  data = {
    slack-token = base64encode(var.slack_token)
  }

  type = "Opaque"
}

# =============================================================================
# RBAC CONFIGURATION
# =============================================================================

# Service Account for CI/CD operations
resource "kubernetes_service_account" "cicd_service_account" {
  metadata {
    name      = "${local.name_prefix}-cicd"
    namespace = kubernetes_namespace.cicd.metadata[0].name

    labels = local.common_tags
  }
}

# ClusterRole for CI/CD operations
resource "kubernetes_cluster_role" "cicd_cluster_role" {
  metadata {
    name = "${local.name_prefix}-cicd"

    labels = local.common_tags
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "daemonsets", "replicasets", "statefulsets"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["networking.k8s.io"]
    resources  = ["networkpolicies", "ingresses"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources  = ["roles", "rolebindings"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["batch"]
    resources  = ["jobs", "cronjobs"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["autoscaling"]
    resources  = ["horizontalpodautoscalers"]
    verbs      = ["*"]
  }
}

# ClusterRoleBinding for CI/CD operations
resource "kubernetes_cluster_role_binding" "cicd_cluster_role_binding" {
  metadata {
    name = "${local.name_prefix}-cicd"

    labels = local.common_tags
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.cicd_cluster_role.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.cicd_service_account.metadata[0].name
    namespace = kubernetes_namespace.cicd.metadata[0].name
  }
}
