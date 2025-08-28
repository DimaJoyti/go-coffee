# Unified Multi-Cloud Monitoring and Observability Module
# Provides comprehensive monitoring across AWS, GCP, and Azure

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
    grafana = {
      source  = "grafana/grafana"
      version = "~> 2.0"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common labels/tags
  common_labels = {
    "app.kubernetes.io/name"       = var.project_name
    "app.kubernetes.io/instance"   = var.environment
    "app.kubernetes.io/component"  = "monitoring"
    "app.kubernetes.io/part-of"    = var.project_name
    "app.kubernetes.io/managed-by" = "terraform"
    "environment"                  = var.environment
    "team"                        = var.team
  }
  
  # Monitoring stack configuration
  monitoring_stack = {
    prometheus = {
      enabled           = var.enable_prometheus
      retention_days    = var.prometheus_retention_days
      storage_size      = var.prometheus_storage_size
      replicas         = var.prometheus_replicas
    }
    grafana = {
      enabled          = var.enable_grafana
      admin_password   = var.grafana_admin_password
      replicas        = var.grafana_replicas
    }
    alertmanager = {
      enabled         = var.enable_alertmanager
      replicas       = var.alertmanager_replicas
    }
    jaeger = {
      enabled        = var.enable_jaeger
      storage_type   = var.jaeger_storage_type
    }
    loki = {
      enabled        = var.enable_loki
      storage_size   = var.loki_storage_size
    }
  }
  
  # Alert rules configuration
  alert_rules = {
    infrastructure = [
      {
        name = "HighCPUUsage"
        expr = "cpu_usage_percent > 80"
        for  = "5m"
        severity = "warning"
        summary = "High CPU usage detected"
      },
      {
        name = "HighMemoryUsage"
        expr = "memory_usage_percent > 85"
        for  = "5m"
        severity = "warning"
        summary = "High memory usage detected"
      },
      {
        name = "DiskSpaceLow"
        expr = "disk_usage_percent > 90"
        for  = "2m"
        severity = "critical"
        summary = "Disk space critically low"
      }
    ]
    application = [
      {
        name = "HighErrorRate"
        expr = "error_rate > 0.05"
        for  = "2m"
        severity = "critical"
        summary = "High error rate detected"
      },
      {
        name = "HighLatency"
        expr = "response_time_p95 > 1000"
        for  = "5m"
        severity = "warning"
        summary = "High response latency detected"
      },
      {
        name = "ServiceDown"
        expr = "up == 0"
        for  = "1m"
        severity = "critical"
        summary = "Service is down"
      }
    ]
    business = [
      {
        name = "LowOrderVolume"
        expr = "coffee_orders_per_minute < 1"
        for  = "10m"
        severity = "warning"
        summary = "Low coffee order volume"
      },
      {
        name = "PaymentFailures"
        expr = "payment_failure_rate > 0.02"
        for  = "3m"
        severity = "critical"
        summary = "High payment failure rate"
      }
    ]
  }
}

# =============================================================================
# KUBERNETES MONITORING NAMESPACE
# =============================================================================

# Monitoring Namespace
resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = var.monitoring_namespace
    
    labels = merge(local.common_labels, {
      "name" = var.monitoring_namespace
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# PROMETHEUS STACK
# =============================================================================

# Prometheus Operator Helm Release
resource "helm_release" "prometheus_stack" {
  count = var.enable_prometheus ? 1 : 0
  
  name       = "prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = var.prometheus_stack_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name
  
  values = [
    yamlencode({
      # Prometheus Configuration
      prometheus = {
        enabled = local.monitoring_stack.prometheus.enabled
        
        prometheusSpec = {
          replicas = local.monitoring_stack.prometheus.replicas
          retention = "${local.monitoring_stack.prometheus.retention_days}d"
          
          storageSpec = {
            volumeClaimTemplate = {
              spec = {
                storageClassName = var.storage_class_name
                accessModes = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = local.monitoring_stack.prometheus.storage_size
                  }
                }
              }
            }
          }
          
          # Resource limits
          resources = {
            requests = {
              cpu    = var.prometheus_cpu_request
              memory = var.prometheus_memory_request
            }
            limits = {
              cpu    = var.prometheus_cpu_limit
              memory = var.prometheus_memory_limit
            }
          }
          
          # Service monitor selector
          serviceMonitorSelectorNilUsesHelmValues = false
          serviceMonitorSelector = {}
          
          # Pod monitor selector
          podMonitorSelectorNilUsesHelmValues = false
          podMonitorSelector = {}
          
          # Rule selector
          ruleSelectorNilUsesHelmValues = false
          ruleSelector = {}
          
          # External labels
          externalLabels = {
            cluster     = var.cluster_name
            environment = var.environment
            region      = var.region
          }
          
          # Remote write configuration for multi-cloud
          remoteWrite = var.enable_remote_write ? [
            {
              url = var.remote_write_url
              basicAuth = {
                username = {
                  name = "prometheus-remote-write"
                  key  = "username"
                }
                password = {
                  name = "prometheus-remote-write"
                  key  = "password"
                }
              }
            }
          ] : []
        }
        
        # Ingress configuration
        ingress = {
          enabled = var.prometheus_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [var.prometheus_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-prometheus-tls"
              hosts = [var.prometheus_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
            "nginx.ingress.kubernetes.io/auth-type" = "basic"
            "nginx.ingress.kubernetes.io/auth-secret" = "prometheus-basic-auth"
          }
        }
      }
      
      # Grafana Configuration
      grafana = {
        enabled = local.monitoring_stack.grafana.enabled
        
        replicas = local.monitoring_stack.grafana.replicas
        
        adminPassword = local.monitoring_stack.grafana.admin_password
        
        # Persistence
        persistence = {
          enabled = true
          storageClassName = var.storage_class_name
          size = var.grafana_storage_size
        }
        
        # Resource limits
        resources = {
          requests = {
            cpu    = var.grafana_cpu_request
            memory = var.grafana_memory_request
          }
          limits = {
            cpu    = var.grafana_cpu_limit
            memory = var.grafana_memory_limit
          }
        }
        
        # Grafana configuration
        "grafana.ini" = {
          server = {
            root_url = "https://${var.grafana_hostname}"
          }
          security = {
            admin_user     = "admin"
            admin_password = local.monitoring_stack.grafana.admin_password
          }
          auth = {
            disable_login_form = false
          }
          "auth.anonymous" = {
            enabled = false
          }
        }
        
        # Data sources
        datasources = {
          "datasources.yaml" = {
            apiVersion = 1
            datasources = [
              {
                name      = "Prometheus"
                type      = "prometheus"
                url       = "http://prometheus-stack-kube-prom-prometheus:9090"
                access    = "proxy"
                isDefault = true
              },
              {
                name   = "Loki"
                type   = "loki"
                url    = "http://loki:3100"
                access = "proxy"
              },
              {
                name = "Jaeger"
                type = "jaeger"
                url  = "http://jaeger-query:16686"
                access = "proxy"
              }
            ]
          }
        }
        
        # Dashboard providers
        dashboardProviders = {
          "dashboardproviders.yaml" = {
            apiVersion = 1
            providers = [
              {
                name = "default"
                orgId = 1
                folder = ""
                type = "file"
                disableDeletion = false
                editable = true
                options = {
                  path = "/var/lib/grafana/dashboards/default"
                }
              }
            ]
          }
        }
        
        # Pre-configured dashboards
        dashboards = {
          default = {
            "go-coffee-overview" = {
              gnetId = 1860
              revision = 27
              datasource = "Prometheus"
            }
            "kubernetes-cluster-monitoring" = {
              gnetId = 7249
              revision = 1
              datasource = "Prometheus"
            }
            "kafka-overview" = {
              gnetId = 7589
              revision = 5
              datasource = "Prometheus"
            }
          }
        }
        
        # Ingress configuration
        ingress = {
          enabled = var.grafana_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [var.grafana_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-grafana-tls"
              hosts = [var.grafana_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
          }
        }
      }
      
      # AlertManager Configuration
      alertmanager = {
        enabled = local.monitoring_stack.alertmanager.enabled
        
        alertmanagerSpec = {
          replicas = local.monitoring_stack.alertmanager.replicas
          
          storage = {
            volumeClaimTemplate = {
              spec = {
                storageClassName = var.storage_class_name
                accessModes = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = var.alertmanager_storage_size
                  }
                }
              }
            }
          }
          
          # Resource limits
          resources = {
            requests = {
              cpu    = var.alertmanager_cpu_request
              memory = var.alertmanager_memory_request
            }
            limits = {
              cpu    = var.alertmanager_cpu_limit
              memory = var.alertmanager_memory_limit
            }
          }
        }
        
        # AlertManager configuration
        config = {
          global = {
            smtp_smarthost = var.smtp_smarthost
            smtp_from = var.smtp_from
          }
          
          route = {
            group_by = ["alertname", "cluster", "service"]
            group_wait = "10s"
            group_interval = "10s"
            repeat_interval = "1h"
            receiver = "web.hook"
            routes = [
              {
                match = {
                  severity = "critical"
                }
                receiver = "critical-alerts"
                group_wait = "5s"
                repeat_interval = "30m"
              }
            ]
          }
          
          receivers = [
            {
              name = "web.hook"
              webhook_configs = [
                {
                  url = var.webhook_url
                  send_resolved = true
                }
              ]
            },
            {
              name = "critical-alerts"
              slack_configs = [
                {
                  api_url = var.slack_webhook_url
                  channel = "#alerts"
                  title = "Critical Alert - {{ .GroupLabels.alertname }}"
                  text = "{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}"
                  send_resolved = true
                }
              ]
              email_configs = [
                {
                  to = var.alert_email
                  subject = "Critical Alert: {{ .GroupLabels.alertname }}"
                  body = "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
                }
              ]
            }
          ]
        }
        
        # Ingress configuration
        ingress = {
          enabled = var.alertmanager_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [var.alertmanager_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-alertmanager-tls"
              hosts = [var.alertmanager_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
            "nginx.ingress.kubernetes.io/auth-type" = "basic"
            "nginx.ingress.kubernetes.io/auth-secret" = "alertmanager-basic-auth"
          }
        }
      }
      
      # Node Exporter
      nodeExporter = {
        enabled = true
      }
      
      # Kube State Metrics
      kubeStateMetrics = {
        enabled = true
      }
      
      # Default rules
      defaultRules = {
        create = true
        rules = {
          alertmanager = true
          etcd = true
          configReloaders = true
          general = true
          k8s = true
          kubeApiserverAvailability = true
          kubeApiserverBurnrate = true
          kubeApiserverHistogram = true
          kubeApiserverSlos = true
          kubelet = true
          kubeProxy = true
          kubePrometheusGeneral = true
          kubePrometheusNodeRecording = true
          kubernetesApps = true
          kubernetesResources = true
          kubernetesStorage = true
          kubernetesSystem = true
          network = true
          node = true
          nodeExporterAlerting = true
          nodeExporterRecording = true
          prometheus = true
          prometheusOperator = true
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.monitoring]
}

# =============================================================================
# LOKI LOGGING STACK
# =============================================================================

# Loki Helm Release
resource "helm_release" "loki" {
  count = var.enable_loki ? 1 : 0
  
  name       = "loki"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "loki-stack"
  version    = var.loki_stack_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name
  
  values = [
    yamlencode({
      loki = {
        enabled = true
        
        persistence = {
          enabled = true
          storageClassName = var.storage_class_name
          size = local.monitoring_stack.loki.storage_size
        }
        
        config = {
          auth_enabled = false
          
          server = {
            http_listen_port = 3100
          }
          
          ingester = {
            lifecycler = {
              address = "127.0.0.1"
              ring = {
                kvstore = {
                  store = "inmemory"
                }
                replication_factor = 1
              }
            }
            chunk_idle_period = "1h"
            max_chunk_age = "1h"
            chunk_target_size = 1048576
            chunk_retain_period = "30s"
            max_transfer_retries = 0
          }
          
          schema_config = {
            configs = [
              {
                from = "2020-10-24"
                store = "boltdb-shipper"
                object_store = "filesystem"
                schema = "v11"
                index = {
                  prefix = "index_"
                  period = "24h"
                }
              }
            ]
          }
          
          storage_config = {
            boltdb_shipper = {
              active_index_directory = "/loki/boltdb-shipper-active"
              cache_location = "/loki/boltdb-shipper-cache"
              cache_ttl = "24h"
              shared_store = "filesystem"
            }
            filesystem = {
              directory = "/loki/chunks"
            }
          }
          
          limits_config = {
            reject_old_samples = true
            reject_old_samples_max_age = "168h"
          }
          
          chunk_store_config = {
            max_look_back_period = "0s"
          }
          
          table_manager = {
            retention_deletes_enabled = false
            retention_period = "0s"
          }
          
          compactor = {
            working_directory = "/loki/boltdb-shipper-compactor"
            shared_store = "filesystem"
          }
        }
      }
      
      promtail = {
        enabled = true
        
        config = {
          server = {
            http_listen_port = 3101
          }
          
          positions = {
            filename = "/tmp/positions.yaml"
          }
          
          clients = [
            {
              url = "http://loki:3100/loki/api/v1/push"
            }
          ]
          
          scrape_configs = [
            {
              job_name = "kubernetes-pods"
              kubernetes_sd_configs = [
                {
                  role = "pod"
                }
              ]
              relabel_configs = [
                {
                  source_labels = ["__meta_kubernetes_pod_annotation_prometheus_io_scrape"]
                  action = "keep"
                  regex = "true"
                },
                {
                  source_labels = ["__meta_kubernetes_pod_annotation_prometheus_io_path"]
                  action = "replace"
                  target_label = "__metrics_path__"
                  regex = "(.+)"
                }
              ]
            }
          ]
        }
      }
      
      fluent-bit = {
        enabled = false
      }
      
      grafana = {
        enabled = false
      }
      
      prometheus = {
        enabled = false
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.monitoring]
}

# =============================================================================
# JAEGER TRACING STACK
# =============================================================================

# Jaeger Helm Release
resource "helm_release" "jaeger" {
  count = var.enable_jaeger ? 1 : 0
  
  name       = "jaeger"
  repository = "https://jaegertracing.github.io/helm-charts"
  chart      = "jaeger"
  version    = var.jaeger_chart_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name
  
  values = [
    yamlencode({
      provisionDataStore = {
        cassandra = false
        elasticsearch = local.monitoring_stack.jaeger.storage_type == "elasticsearch"
        kafka = false
      }
      
      allInOne = {
        enabled = local.monitoring_stack.jaeger.storage_type == "memory"
        
        ingress = {
          enabled = var.jaeger_ingress_enabled
          ingressClassName = var.ingress_class_name
          hosts = [var.jaeger_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-jaeger-tls"
              hosts = [var.jaeger_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
          }
        }
      }
      
      storage = {
        type = local.monitoring_stack.jaeger.storage_type
      }
      
      agent = {
        enabled = true
      }
      
      collector = {
        enabled = local.monitoring_stack.jaeger.storage_type != "memory"
      }
      
      query = {
        enabled = local.monitoring_stack.jaeger.storage_type != "memory"
        
        ingress = {
          enabled = var.jaeger_ingress_enabled && local.monitoring_stack.jaeger.storage_type != "memory"
          ingressClassName = var.ingress_class_name
          hosts = [var.jaeger_hostname]
          tls = [
            {
              secretName = "${local.name_prefix}-jaeger-tls"
              hosts = [var.jaeger_hostname]
            }
          ]
          annotations = {
            "cert-manager.io/cluster-issuer" = var.cert_manager_issuer
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.monitoring]
}
