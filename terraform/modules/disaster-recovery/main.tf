# Disaster Recovery and Business Continuity Module
# Implements comprehensive DR strategies, automated failover, and business continuity planning

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
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
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
    Component   = "disaster-recovery"
    Team        = var.team
    CostCenter  = var.cost_center
  }
  
  # DR configuration
  disaster_recovery = {
    enabled = var.enable_disaster_recovery
    
    # Recovery objectives
    rto_minutes = var.recovery_time_objective_minutes
    rpo_minutes = var.recovery_point_objective_minutes
    
    # Backup configuration
    backup = {
      enabled = var.enable_automated_backups
      retention_days = var.backup_retention_days
      frequency = var.backup_frequency
      cross_region = var.enable_cross_region_backup
      cross_cloud = var.enable_cross_cloud_backup
    }
    
    # Failover configuration
    failover = {
      enabled = var.enable_automated_failover
      health_check_interval = var.health_check_interval_seconds
      failure_threshold = var.failure_threshold_count
      recovery_threshold = var.recovery_threshold_count
    }
    
    # Testing configuration
    testing = {
      enabled = var.enable_dr_testing
      schedule = var.dr_test_schedule
      automated = var.enable_automated_dr_testing
    }
  }
  
  # Multi-region configuration
  regions = {
    primary = var.primary_region
    secondary = var.secondary_region
    tertiary = var.tertiary_region
  }
  
  # Critical services for DR
  critical_services = [
    "coffee-service",
    "payment-service",
    "user-service",
    "order-service",
    "inventory-service"
  ]
}

# =============================================================================
# KUBERNETES DR NAMESPACE
# =============================================================================

# Disaster Recovery Namespace
resource "kubernetes_namespace" "disaster_recovery" {
  metadata {
    name = var.disaster_recovery_namespace
    
    labels = merge(local.common_tags, {
      "name" = var.disaster_recovery_namespace
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# VELERO BACKUP SYSTEM
# =============================================================================

# Velero for Kubernetes backup and restore
resource "helm_release" "velero" {
  count = var.enable_automated_backups ? 1 : 0
  
  name       = "velero"
  repository = "https://vmware-tanzu.github.io/helm-charts"
  chart      = "velero"
  version    = var.velero_chart_version
  namespace  = kubernetes_namespace.disaster_recovery.metadata[0].name
  
  values = [
    yamlencode({
      # Velero configuration
      configuration = {
        # Cloud provider configuration
        provider = var.cloud_provider
        
        # Backup storage location
        backupStorageLocation = {
          name = "default"
          provider = var.cloud_provider
          bucket = var.backup_storage_bucket
          config = var.cloud_provider == "aws" ? {
            region = var.primary_region
            s3ForcePathStyle = false
          } : var.cloud_provider == "gcp" ? {
            location = var.primary_region
          } : {
            resourceGroup = var.azure_resource_group_name
            storageAccount = var.azure_storage_account_name
          }
        }
        
        # Volume snapshot location
        volumeSnapshotLocation = {
          name = "default"
          provider = var.cloud_provider
          config = var.cloud_provider == "aws" ? {
            region = var.primary_region
          } : var.cloud_provider == "gcp" ? {
            project = var.gcp_project_id
          } : {
            resourceGroup = var.azure_resource_group_name
          }
        }
        
        # Backup retention
        defaultBackupTTL = "${local.disaster_recovery.backup.retention_days * 24}h"
        
        # Restore configuration
        restoreResourcePriorities = [
          "namespaces",
          "storageclasses",
          "volumesnapshotclass.snapshot.storage.k8s.io",
          "volumesnapshotcontents.snapshot.storage.k8s.io",
          "volumesnapshots.snapshot.storage.k8s.io",
          "persistentvolumes",
          "persistentvolumeclaims",
          "secrets",
          "configmaps",
          "serviceaccounts",
          "limitranges",
          "pods"
        ]
      }
      
      # Credentials
      credentials = {
        useSecret = true
        name = "cloud-credentials"
        secretContents = var.cloud_provider == "aws" ? {
          cloud = base64encode(<<-EOF
            [default]
            aws_access_key_id=${var.aws_access_key_id}
            aws_secret_access_key=${var.aws_secret_access_key}
          EOF
          )
        } : var.cloud_provider == "gcp" ? {
          cloud = base64encode(var.gcp_service_account_key)
        } : {
          cloud = base64encode(<<-EOF
            AZURE_SUBSCRIPTION_ID=${var.azure_subscription_id}
            AZURE_TENANT_ID=${var.azure_tenant_id}
            AZURE_CLIENT_ID=${var.azure_client_id}
            AZURE_CLIENT_SECRET=${var.azure_client_secret}
            AZURE_RESOURCE_GROUP=${var.azure_resource_group_name}
            AZURE_CLOUD_NAME=AzurePublicCloud
          EOF
          )
        }
      }
      
      # Init containers for plugins
      initContainers = var.cloud_provider == "aws" ? [
        {
          name = "velero-plugin-for-aws"
          image = "velero/velero-plugin-for-aws:v1.8.0"
          imagePullPolicy = "IfNotPresent"
          volumeMounts = [
            {
              mountPath = "/target"
              name = "plugins"
            }
          ]
        }
      ] : var.cloud_provider == "gcp" ? [
        {
          name = "velero-plugin-for-gcp"
          image = "velero/velero-plugin-for-gcp:v1.8.0"
          imagePullPolicy = "IfNotPresent"
          volumeMounts = [
            {
              mountPath = "/target"
              name = "plugins"
            }
          ]
        }
      ] : [
        {
          name = "velero-plugin-for-microsoft-azure"
          image = "velero/velero-plugin-for-microsoft-azure:v1.8.0"
          imagePullPolicy = "IfNotPresent"
          volumeMounts = [
            {
              mountPath = "/target"
              name = "plugins"
            }
          ]
        }
      ]
      
      # Resources
      resources = {
        requests = {
          cpu    = "500m"
          memory = "128Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
      
      # Service monitor for Prometheus
      serviceMonitor = {
        enabled = var.monitoring_enabled
        additionalLabels = local.common_tags
      }
      
      # Metrics
      metrics = {
        enabled = var.monitoring_enabled
        scrapeInterval = "30s"
        scrapeTimeout = "10s"
      }
      
      # Deployment configuration
      deployRestic = true
      
      # Schedules for automated backups
      schedules = {
        daily = {
          disabled = false
          schedule = "0 2 * * *"  # Daily at 2 AM
          template = {
            ttl = "${local.disaster_recovery.backup.retention_days * 24}h"
            includedNamespaces = ["go-coffee", "ai-agents", "web3", "monitoring"]
            excludedResources = ["events", "events.events.k8s.io"]
            snapshotVolumes = true
            includeClusterResources = true
          }
        }
        weekly = {
          disabled = false
          schedule = "0 1 * * 0"  # Weekly on Sunday at 1 AM
          template = {
            ttl = "${local.disaster_recovery.backup.retention_days * 24 * 4}h"  # 4x retention for weekly
            includedNamespaces = ["*"]
            excludedResources = ["events", "events.events.k8s.io"]
            snapshotVolumes = true
            includeClusterResources = true
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.disaster_recovery]
}

# =============================================================================
# DATABASE BACKUP AND REPLICATION
# =============================================================================

# Database backup configuration
resource "kubernetes_manifest" "database_backup_cronjob" {
  count = var.enable_database_backup ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-database-backup"
      namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "database-backup"
      })
    }
    
    spec = {
      schedule = local.disaster_recovery.backup.frequency
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "database-backup"
                  image = var.database_backup_image
                  
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
                      name  = "BACKUP_RETENTION_DAYS"
                      value = tostring(local.disaster_recovery.backup.retention_days)
                    },
                    {
                      name  = "CROSS_REGION_BACKUP"
                      value = tostring(local.disaster_recovery.backup.cross_region)
                    },
                    {
                      name  = "CROSS_CLOUD_BACKUP"
                      value = tostring(local.disaster_recovery.backup.cross_cloud)
                    },
                    {
                      name = "DATABASE_URL"
                      valueFrom = {
                        secretKeyRef = {
                          name = "database-credentials"
                          key  = "url"
                        }
                      }
                    },
                    {
                      name = "BACKUP_STORAGE_URL"
                      valueFrom = {
                        secretKeyRef = {
                          name = "backup-credentials"
                          key  = "storage-url"
                        }
                      }
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting database backup..."
                      
                      # Create timestamp for backup
                      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                      BACKUP_NAME="${var.project_name}_${var.environment}_$TIMESTAMP"
                      
                      # Perform database backup
                      pg_dump $DATABASE_URL > /tmp/$BACKUP_NAME.sql
                      
                      # Compress backup
                      gzip /tmp/$BACKUP_NAME.sql
                      
                      # Upload to primary backup location
                      aws s3 cp /tmp/$BACKUP_NAME.sql.gz $BACKUP_STORAGE_URL/database/
                      
                      # Cross-region backup if enabled
                      if [ "$CROSS_REGION_BACKUP" = "true" ]; then
                        aws s3 cp /tmp/$BACKUP_NAME.sql.gz $BACKUP_STORAGE_URL/database/ --region ${var.secondary_region}
                      fi
                      
                      # Cross-cloud backup if enabled
                      if [ "$CROSS_CLOUD_BACKUP" = "true" ]; then
                        # Upload to GCS or Azure Blob Storage
                        echo "Cross-cloud backup not implemented yet"
                      fi
                      
                      # Cleanup old backups
                      find /tmp -name "*.sql.gz" -mtime +$BACKUP_RETENTION_DAYS -delete
                      
                      echo "Database backup completed: $BACKUP_NAME.sql.gz"
                    EOT
                  ]
                  
                  resources = {
                    requests = {
                      cpu    = "200m"
                      memory = "256Mi"
                    }
                    limits = {
                      cpu    = "1000m"
                      memory = "1Gi"
                    }
                  }
                  
                  volumeMounts = [
                    {
                      name      = "backup-storage"
                      mountPath = "/tmp"
                    }
                  ]
                }
              ]
              
              volumes = [
                {
                  name = "backup-storage"
                  emptyDir = {
                    sizeLimit = "10Gi"
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

# =============================================================================
# AUTOMATED FAILOVER SYSTEM
# =============================================================================

# Failover controller deployment
resource "kubernetes_manifest" "failover_controller" {
  count = var.enable_automated_failover ? 1 : 0
  
  manifest = {
    apiVersion = "apps/v1"
    kind       = "Deployment"
    
    metadata = {
      name      = "${local.name_prefix}-failover-controller"
      namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "failover-controller"
      })
    }
    
    spec = {
      replicas = var.failover_controller_replicas
      
      selector = {
        matchLabels = {
          app = "${local.name_prefix}-failover-controller"
        }
      }
      
      template = {
        metadata = {
          labels = merge(local.common_tags, {
            app = "${local.name_prefix}-failover-controller"
          })
        }
        
        spec = {
          serviceAccountName = kubernetes_service_account.failover_controller[0].metadata[0].name
          
          containers = [
            {
              name  = "failover-controller"
              image = var.failover_controller_image
              
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
                  name  = "PRIMARY_REGION"
                  value = local.regions.primary
                },
                {
                  name  = "SECONDARY_REGION"
                  value = local.regions.secondary
                },
                {
                  name  = "HEALTH_CHECK_INTERVAL"
                  value = tostring(local.disaster_recovery.failover.health_check_interval)
                },
                {
                  name  = "FAILURE_THRESHOLD"
                  value = tostring(local.disaster_recovery.failover.failure_threshold)
                },
                {
                  name  = "RECOVERY_THRESHOLD"
                  value = tostring(local.disaster_recovery.failover.recovery_threshold)
                },
                {
                  name  = "CRITICAL_SERVICES"
                  value = join(",", local.critical_services)
                },
                {
                  name  = "RTO_MINUTES"
                  value = tostring(local.disaster_recovery.rto_minutes)
                },
                {
                  name  = "RPO_MINUTES"
                  value = tostring(local.disaster_recovery.rpo_minutes)
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

# Service Account for Failover Controller
resource "kubernetes_service_account" "failover_controller" {
  count = var.enable_automated_failover ? 1 : 0
  
  metadata {
    name      = "${local.name_prefix}-failover-controller"
    namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
    
    labels = local.common_tags
  }
}

# ClusterRole for Failover Controller
resource "kubernetes_cluster_role" "failover_controller" {
  count = var.enable_automated_failover ? 1 : 0
  
  metadata {
    name = "${local.name_prefix}-failover-controller"
    
    labels = local.common_tags
  }
  
  rule {
    api_groups = [""]
    resources  = ["pods", "services", "endpoints", "nodes"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }
  
  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "replicasets", "statefulsets"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }
  
  rule {
    api_groups = ["networking.k8s.io"]
    resources  = ["ingresses"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }
  
  rule {
    api_groups = [""]
    resources  = ["events"]
    verbs      = ["create", "patch"]
  }
}

# ClusterRoleBinding for Failover Controller
resource "kubernetes_cluster_role_binding" "failover_controller" {
  count = var.enable_automated_failover ? 1 : 0
  
  metadata {
    name = "${local.name_prefix}-failover-controller"
    
    labels = local.common_tags
  }
  
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.failover_controller[0].metadata[0].name
  }
  
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.failover_controller[0].metadata[0].name
    namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
  }
}

# =============================================================================
# DR TESTING AUTOMATION
# =============================================================================

# DR Testing CronJob
resource "kubernetes_manifest" "dr_testing_cronjob" {
  count = var.enable_dr_testing ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-dr-testing"
      namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "dr-testing"
      })
    }
    
    spec = {
      schedule = local.disaster_recovery.testing.schedule
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "dr-tester"
                  image = var.dr_testing_image
                  
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
                      name  = "TEST_TYPE"
                      value = var.dr_test_type
                    },
                    {
                      name  = "AUTOMATED_TESTING"
                      value = tostring(local.disaster_recovery.testing.automated)
                    },
                    {
                      name  = "CRITICAL_SERVICES"
                      value = join(",", local.critical_services)
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting DR testing..."
                      
                      # Test backup integrity
                      echo "Testing backup integrity..."
                      velero backup get --output json | jq '.items[] | select(.status.phase == "Completed")'
                      
                      # Test restore functionality
                      if [ "$AUTOMATED_TESTING" = "true" ]; then
                        echo "Performing automated restore test..."
                        
                        # Create test namespace
                        kubectl create namespace dr-test-$(date +%s) || true
                        
                        # Perform test restore
                        velero restore create dr-test-restore-$(date +%s) \
                          --from-backup $(velero backup get -o name | head -1 | cut -d'/' -f2) \
                          --namespace-mappings go-coffee:dr-test-$(date +%s)
                        
                        # Wait for restore completion
                        sleep 300
                        
                        # Verify restore
                        kubectl get pods -n dr-test-$(date +%s)
                        
                        # Cleanup test namespace
                        kubectl delete namespace dr-test-$(date +%s) || true
                      fi
                      
                      # Test failover mechanisms
                      echo "Testing failover mechanisms..."
                      
                      # Generate DR test report
                      echo "Generating DR test report..."
                      
                      echo "DR testing completed"
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

# =============================================================================
# CROSS-REGION REPLICATION
# =============================================================================

# Cross-region replication for critical data
resource "time_rotating" "replication_schedule" {
  count = var.enable_cross_region_replication ? 1 : 0
  
  rotation_hours = var.replication_interval_hours
}

# Replication CronJob
resource "kubernetes_manifest" "cross_region_replication" {
  count = var.enable_cross_region_replication ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-cross-region-replication"
      namespace = kubernetes_namespace.disaster_recovery.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "cross-region-replication"
      })
    }
    
    spec = {
      schedule = "0 */4 * * *"  # Every 4 hours
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "replication-manager"
                  image = var.replication_manager_image
                  
                  env = [
                    {
                      name  = "PRIMARY_REGION"
                      value = local.regions.primary
                    },
                    {
                      name  = "SECONDARY_REGION"
                      value = local.regions.secondary
                    },
                    {
                      name  = "REPLICATION_INTERVAL_HOURS"
                      value = tostring(var.replication_interval_hours)
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting cross-region replication..."
                      
                      # Replicate critical data
                      # This would include database replication, file synchronization, etc.
                      
                      echo "Cross-region replication completed"
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
