# GitOps Automation Module
# Manages GitOps workflows with ArgoCD, Flux, and infrastructure drift detection

terraform {
  required_version = ">= 1.6.0"
  required_providers {
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
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "~> 16.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common labels
  common_labels = {
    "app.kubernetes.io/name"       = var.project_name
    "app.kubernetes.io/instance"   = var.environment
    "app.kubernetes.io/version"    = var.app_version
    "app.kubernetes.io/component"  = "gitops"
    "app.kubernetes.io/part-of"    = var.project_name
    "app.kubernetes.io/managed-by" = "terraform"
    "environment"                  = var.environment
    "team"                        = var.team
  }
  
  # ArgoCD applications configuration
  argocd_applications = {
    core-services = {
      namespace = "go-coffee"
      path      = "k8s/overlays/${var.environment}/core"
      sync_policy = {
        automated = true
        prune     = true
        self_heal = true
      }
    }
    ai-agents = {
      namespace = "ai-agents"
      path      = "k8s/overlays/${var.environment}/ai-agents"
      sync_policy = {
        automated = true
        prune     = true
        self_heal = true
      }
    }
    web3-services = {
      namespace = "web3"
      path      = "k8s/overlays/${var.environment}/web3"
      sync_policy = {
        automated = false
        prune     = true
        self_heal = false
      }
    }
    infrastructure = {
      namespace = "infrastructure"
      path      = "k8s/overlays/${var.environment}/infrastructure"
      sync_policy = {
        automated = false
        prune     = false
        self_heal = false
      }
    }
    monitoring = {
      namespace = "monitoring"
      path      = "k8s/overlays/${var.environment}/monitoring"
      sync_policy = {
        automated = true
        prune     = true
        self_heal = true
      }
    }
  }
}

# =============================================================================
# NAMESPACE CREATION
# =============================================================================

# ArgoCD Namespace
resource "kubernetes_namespace" "argocd" {
  count = var.enable_argocd ? 1 : 0
  
  metadata {
    name = var.argocd_namespace
    
    labels = merge(local.common_labels, {
      "app.kubernetes.io/component" = "argocd"
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# Flux System Namespace
resource "kubernetes_namespace" "flux_system" {
  count = var.enable_flux ? 1 : 0
  
  metadata {
    name = var.flux_namespace
    
    labels = merge(local.common_labels, {
      "app.kubernetes.io/component" = "flux"
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# ARGOCD INSTALLATION
# =============================================================================

# ArgoCD Helm Release
resource "helm_release" "argocd" {
  count = var.enable_argocd ? 1 : 0
  
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = var.argocd_chart_version
  namespace  = kubernetes_namespace.argocd[0].metadata[0].name
  
  # ArgoCD Configuration
  values = [
    yamlencode({
      global = {
        image = {
          tag = var.argocd_version
        }
      }
      
      # Server Configuration
      server = {
        replicas = var.argocd_server_replicas
        
        config = {
          "application.instanceLabelKey" = "argocd.argoproj.io/instance"
          "server.rbac.log.enforce.enable" = "true"
          "exec.enabled" = "false"
          "admin.enabled" = "true"
          "timeout.reconciliation" = "180s"
          "timeout.hard.reconciliation" = "0s"
          
          # OIDC Configuration
          "oidc.config" = var.argocd_oidc_config
          
          # Repository credentials
          "repositories" = yamlencode([
            {
              url = var.git_repository_url
              passwordSecret = {
                name = "repo-credentials"
                key  = "password"
              }
              usernameSecret = {
                name = "repo-credentials"
                key  = "username"
              }
            }
          ])
        }
        
        # Ingress Configuration
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
        
        # Metrics and Monitoring
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
          }
        }
      }
      
      # Repository Server Configuration
      repoServer = {
        replicas = var.argocd_repo_server_replicas
        
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
          }
        }
      }
      
      # Application Controller Configuration
      controller = {
        replicas = var.argocd_controller_replicas
        
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
          }
        }
      }
      
      # Redis Configuration
      redis = {
        enabled = true
        metrics = {
          enabled = var.monitoring_enabled
          serviceMonitor = {
            enabled = var.monitoring_enabled
          }
        }
      }
      
      # Notifications Configuration
      notifications = {
        enabled = var.argocd_notifications_enabled
        
        notifiers = {
          "service.slack" = {
            token = "$slack-token"
          }
          "service.email.gmail" = {
            username = "$email-username"
            password = "$email-password"
            host = "smtp.gmail.com"
            port = 587
            from = "$email-from"
          }
        }
        
        templates = {
          "template.app-deployed" = {
            email = {
              subject = "New version of an application {{.app.metadata.name}} is up and running."
            }
            message = "{{if eq .serviceType \"slack\"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} is now running new version."
          }
          "template.app-health-degraded" = {
            email = {
              subject = "Application {{.app.metadata.name}} has degraded."
            }
            message = "{{if eq .serviceType \"slack\"}}:exclamation:{{end}} Application {{.app.metadata.name}} has degraded."
          }
          "template.app-sync-failed" = {
            email = {
              subject = "Application {{.app.metadata.name}} sync is failed."
            }
            message = "{{if eq .serviceType \"slack\"}}:exclamation:{{end}} Application {{.app.metadata.name}} sync is failed."
          }
        }
        
        triggers = {
          "trigger.on-deployed" = [
            {
              when = "app.status.operationState.phase in ['Succeeded'] and app.status.health.status == 'Healthy'"
              send = ["app-deployed"]
            }
          ]
          "trigger.on-health-degraded" = [
            {
              when = "app.status.health.status == 'Degraded'"
              send = ["app-health-degraded"]
            }
          ]
          "trigger.on-sync-failed" = [
            {
              when = "app.status.operationState.phase in ['Error', 'Failed']"
              send = ["app-sync-failed"]
            }
          ]
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.argocd]
}

# =============================================================================
# ARGOCD APPLICATIONS
# =============================================================================

# ArgoCD Applications
resource "kubernetes_manifest" "argocd_applications" {
  for_each = var.enable_argocd ? local.argocd_applications : {}
  
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    
    metadata = {
      name      = "${local.name_prefix}-${each.key}"
      namespace = kubernetes_namespace.argocd[0].metadata[0].name
      
      labels = merge(local.common_labels, {
        "app.kubernetes.io/component" = each.key
      })
      
      finalizers = ["resources-finalizer.argocd.argoproj.io"]
    }
    
    spec = {
      project = "default"
      
      source = {
        repoURL        = var.git_repository_url
        targetRevision = var.git_target_revision
        path           = each.value.path
      }
      
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = each.value.namespace
      }
      
      syncPolicy = {
        automated = each.value.sync_policy.automated ? {
          prune    = each.value.sync_policy.prune
          selfHeal = each.value.sync_policy.self_heal
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
# ARGOCD PROJECT CONFIGURATION
# =============================================================================

# ArgoCD Project for Go Coffee
resource "kubernetes_manifest" "argocd_project" {
  count = var.enable_argocd ? 1 : 0
  
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "AppProject"
    
    metadata = {
      name      = local.name_prefix
      namespace = kubernetes_namespace.argocd[0].metadata[0].name
      
      labels = local.common_labels
    }
    
    spec = {
      description = "Go Coffee Platform Project"
      
      sourceRepos = [
        var.git_repository_url,
        "https://charts.helm.sh/stable",
        "https://kubernetes-charts.storage.googleapis.com",
        "https://argoproj.github.io/argo-helm"
      ]
      
      destinations = [
        {
          namespace = "*"
          server    = "https://kubernetes.default.svc"
        }
      ]
      
      clusterResourceWhitelist = [
        {
          group = "*"
          kind  = "*"
        }
      ]
      
      namespaceResourceWhitelist = [
        {
          group = "*"
          kind  = "*"
        }
      ]
      
      roles = [
        {
          name = "admin"
          description = "Admin access to Go Coffee project"
          policies = [
            "p, proj:${local.name_prefix}:admin, applications, *, ${local.name_prefix}/*, allow",
            "p, proj:${local.name_prefix}:admin, repositories, *, *, allow",
            "p, proj:${local.name_prefix}:admin, clusters, *, *, allow"
          ]
          groups = var.argocd_admin_groups
        },
        {
          name = "developer"
          description = "Developer access to Go Coffee project"
          policies = [
            "p, proj:${local.name_prefix}:developer, applications, get, ${local.name_prefix}/*, allow",
            "p, proj:${local.name_prefix}:developer, applications, sync, ${local.name_prefix}/*, allow"
          ]
          groups = var.argocd_developer_groups
        }
      ]
    }
  }
  
  depends_on = [helm_release.argocd]
}

# =============================================================================
# SECRETS MANAGEMENT
# =============================================================================

# Repository credentials secret
resource "kubernetes_secret" "repo_credentials" {
  count = var.enable_argocd ? 1 : 0
  
  metadata {
    name      = "repo-credentials"
    namespace = kubernetes_namespace.argocd[0].metadata[0].name
    
    labels = merge(local.common_labels, {
      "argocd.argoproj.io/secret-type" = "repository"
    })
  }
  
  data = {
    type     = "git"
    url      = var.git_repository_url
    username = var.git_username
    password = var.git_token
  }
  
  type = "Opaque"
}

# Notification secrets
resource "kubernetes_secret" "notification_secrets" {
  count = var.enable_argocd && var.argocd_notifications_enabled ? 1 : 0
  
  metadata {
    name      = "argocd-notifications-secret"
    namespace = kubernetes_namespace.argocd[0].metadata[0].name
    
    labels = local.common_labels
  }
  
  data = {
    slack-token      = var.slack_token
    email-username   = var.email_username
    email-password   = var.email_password
    email-from       = var.email_from
  }
  
  type = "Opaque"
}

# =============================================================================
# INFRASTRUCTURE DRIFT DETECTION
# =============================================================================

# Drift Detection CronJob
resource "kubernetes_manifest" "drift_detection_cronjob" {
  count = var.enable_drift_detection ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-drift-detection"
      namespace = var.drift_detection_namespace
      
      labels = merge(local.common_labels, {
        "app.kubernetes.io/component" = "drift-detection"
      })
    }
    
    spec = {
      schedule = var.drift_detection_schedule
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "drift-detector"
                  image = var.drift_detection_image
                  
                  env = [
                    {
                      name  = "TERRAFORM_CLOUD_TOKEN"
                      valueFrom = {
                        secretKeyRef = {
                          name = "terraform-cloud-credentials"
                          key  = "token"
                        }
                      }
                    },
                    {
                      name  = "WORKSPACE_ID"
                      value = var.terraform_workspace_id
                    },
                    {
                      name  = "SLACK_WEBHOOK_URL"
                      valueFrom = {
                        secretKeyRef = {
                          name = "notification-secrets"
                          key  = "slack-webhook-url"
                        }
                      }
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting infrastructure drift detection..."
                      
                      # Run Terraform plan to detect drift
                      terraform plan -detailed-exitcode -no-color > /tmp/plan.out 2>&1
                      PLAN_EXIT_CODE=$?
                      
                      if [ $PLAN_EXIT_CODE -eq 2 ]; then
                        echo "Infrastructure drift detected!"
                        
                        # Send notification to Slack
                        curl -X POST -H 'Content-type: application/json' \
                          --data "{\"text\":\"🚨 Infrastructure drift detected in ${var.environment} environment. Please review and apply changes.\"}" \
                          $SLACK_WEBHOOK_URL
                        
                        # Create Kubernetes event
                        kubectl create event drift-detected \
                          --type=Warning \
                          --reason=InfrastructureDrift \
                          --message="Infrastructure drift detected in Terraform workspace"
                      else
                        echo "No infrastructure drift detected."
                      fi
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
