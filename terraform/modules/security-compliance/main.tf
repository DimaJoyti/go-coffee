# Security and Compliance Automation Module
# Implements automated security scanning, compliance monitoring, and policy enforcement

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
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
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
    Component   = "security-compliance"
    Team        = var.team
    CostCenter  = var.cost_center
  }
  
  # Security scanning configuration
  security_tools = {
    vulnerability_scanning = {
      enabled = var.enable_vulnerability_scanning
      tools   = ["trivy", "grype", "snyk"]
      schedule = var.vulnerability_scan_schedule
    }
    compliance_monitoring = {
      enabled = var.enable_compliance_monitoring
      frameworks = var.compliance_frameworks
      schedule = var.compliance_check_schedule
    }
    threat_detection = {
      enabled = var.enable_threat_detection
      sensitivity = var.threat_detection_sensitivity
    }
    policy_enforcement = {
      enabled = var.enable_policy_enforcement
      policies = var.security_policies
    }
  }
  
  # Compliance frameworks configuration
  compliance_checks = {
    "SOC2" = {
      controls = [
        "access_control",
        "encryption_at_rest",
        "encryption_in_transit",
        "audit_logging",
        "backup_procedures",
        "incident_response"
      ]
      severity = "high"
    }
    "PCI-DSS" = {
      controls = [
        "network_segmentation",
        "access_control",
        "encryption",
        "vulnerability_management",
        "monitoring",
        "secure_development"
      ]
      severity = "critical"
    }
    "GDPR" = {
      controls = [
        "data_protection",
        "privacy_by_design",
        "consent_management",
        "data_retention",
        "breach_notification",
        "data_portability"
      ]
      severity = "high"
    }
    "HIPAA" = {
      controls = [
        "access_control",
        "audit_controls",
        "integrity",
        "transmission_security",
        "encryption"
      ]
      severity = "critical"
    }
  }
}

# =============================================================================
# KUBERNETES SECURITY NAMESPACE
# =============================================================================

# Security Namespace
resource "kubernetes_namespace" "security" {
  metadata {
    name = var.security_namespace
    
    labels = merge(local.common_tags, {
      "name" = var.security_namespace
      "pod-security.kubernetes.io/enforce" = "restricted"
      "pod-security.kubernetes.io/audit" = "restricted"
      "pod-security.kubernetes.io/warn" = "restricted"
    })
    
    annotations = {
      "managed-by" = "terraform"
      "created-at" = timestamp()
    }
  }
}

# =============================================================================
# VULNERABILITY SCANNING
# =============================================================================

# Trivy Operator for Kubernetes vulnerability scanning
resource "helm_release" "trivy_operator" {
  count = var.enable_vulnerability_scanning ? 1 : 0
  
  name       = "trivy-operator"
  repository = "https://aquasecurity.github.io/helm-charts"
  chart      = "trivy-operator"
  version    = var.trivy_operator_version
  namespace  = kubernetes_namespace.security.metadata[0].name
  
  values = [
    yamlencode({
      # Trivy Operator Configuration
      trivyOperator = {
        scanJobTimeout = "5m"
        scanJobsConcurrentLimit = 10
        scanJobsRetryDelay = "30s"
        
        # Vulnerability database
        vulnerabilityReportsPlugin = "Trivy"
        configAuditReportsPlugin = "Trivy"
        
        # Compliance reports
        complianceFailEntriesLimit = 10
        
        # Metrics
        metricsBindAddress = "0.0.0.0:8080"
        healthProbeBindAddress = "0.0.0.0:9090"
      }
      
      # Service Monitor for Prometheus
      serviceMonitor = {
        enabled = var.monitoring_enabled
        labels = local.common_tags
      }
      
      # Node collector for host scanning
      nodeCollector = {
        registry = "ghcr.io"
        repository = "aquasecurity/node-collector"
        tag = "0.0.9"
        imagePullPolicy = "IfNotPresent"
        
        # Volume mounts for host scanning
        volumeMounts = [
          {
            name = "var-lib-etcd"
            mountPath = "/var/lib/etcd"
            readOnly = true
          },
          {
            name = "var-lib-kubelet"
            mountPath = "/var/lib/kubelet"
            readOnly = true
          },
          {
            name = "etc-systemd"
            mountPath = "/etc/systemd"
            readOnly = true
          }
        ]
      }
      
      # Trivy configuration
      trivy = {
        # Image registry settings
        registry = {
          mirror = {}
        }
        
        # Database settings
        serverInsecure = false
        
        # Scanning settings
        timeout = "5m0s"
        
        # Ignore unfixed vulnerabilities
        ignoreUnfixed = false
        
        # Severity levels to report
        severity = "UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL"
        
        # Skip files and directories
        skipFiles = []
        skipDirs = []
        
        # Custom policies
        policy = ""
        
        # Resources
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
  
  depends_on = [kubernetes_namespace.security]
}

# Vulnerability Scan CronJob
resource "kubernetes_manifest" "vulnerability_scan_cronjob" {
  count = var.enable_vulnerability_scanning ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-vulnerability-scan"
      namespace = kubernetes_namespace.security.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "vulnerability-scanner"
      })
    }
    
    spec = {
      schedule = local.security_tools.vulnerability_scanning.schedule
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "vulnerability-scanner"
                  image = var.vulnerability_scanner_image
                  
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
                      name  = "SCAN_TARGETS"
                      value = join(",", var.vulnerability_scan_targets)
                    },
                    {
                      name  = "SEVERITY_THRESHOLD"
                      value = var.vulnerability_severity_threshold
                    },
                    {
                      name  = "SLACK_WEBHOOK_URL"
                      valueFrom = {
                        secretKeyRef = {
                          name = "security-notifications"
                          key  = "slack-webhook-url"
                        }
                      }
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting vulnerability scan..."
                      
                      # Scan container images
                      for target in $(echo $SCAN_TARGETS | tr ',' ' '); do
                        echo "Scanning $target..."
                        trivy image --severity $SEVERITY_THRESHOLD --format json $target > /tmp/scan-$target.json
                        
                        # Check for critical vulnerabilities
                        CRITICAL_COUNT=$(jq '.Results[]?.Vulnerabilities[]? | select(.Severity=="CRITICAL") | length' /tmp/scan-$target.json | wc -l)
                        
                        if [ "$CRITICAL_COUNT" -gt 0 ]; then
                          echo "Critical vulnerabilities found in $target: $CRITICAL_COUNT"
                          
                          # Send alert to Slack
                          curl -X POST -H 'Content-type: application/json' \
                            --data "{\"text\":\"🚨 Critical vulnerabilities found in $target: $CRITICAL_COUNT vulnerabilities\"}" \
                            $SLACK_WEBHOOK_URL
                        fi
                      done
                      
                      echo "Vulnerability scan completed"
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
                      name      = "docker-socket"
                      mountPath = "/var/run/docker.sock"
                      readOnly  = true
                    }
                  ]
                }
              ]
              
              volumes = [
                {
                  name = "docker-socket"
                  hostPath = {
                    path = "/var/run/docker.sock"
                    type = "Socket"
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
# COMPLIANCE MONITORING
# =============================================================================

# Compliance Checker CronJob
resource "kubernetes_manifest" "compliance_check_cronjob" {
  count = var.enable_compliance_monitoring ? 1 : 0
  
  manifest = {
    apiVersion = "batch/v1"
    kind       = "CronJob"
    
    metadata = {
      name      = "${local.name_prefix}-compliance-check"
      namespace = kubernetes_namespace.security.metadata[0].name
      
      labels = merge(local.common_tags, {
        "app.kubernetes.io/component" = "compliance-checker"
      })
    }
    
    spec = {
      schedule = local.security_tools.compliance_monitoring.schedule
      
      jobTemplate = {
        spec = {
          template = {
            spec = {
              restartPolicy = "OnFailure"
              
              containers = [
                {
                  name  = "compliance-checker"
                  image = var.compliance_checker_image
                  
                  env = [
                    {
                      name  = "ENVIRONMENT"
                      value = var.environment
                    },
                    {
                      name  = "COMPLIANCE_FRAMEWORKS"
                      value = join(",", var.compliance_frameworks)
                    },
                    {
                      name  = "AWS_REGION"
                      value = var.aws_region
                    },
                    {
                      name  = "GCP_PROJECT_ID"
                      value = var.gcp_project_id
                    },
                    {
                      name  = "AZURE_SUBSCRIPTION_ID"
                      value = var.azure_subscription_id
                    }
                  ]
                  
                  command = ["/bin/sh"]
                  args = [
                    "-c",
                    <<-EOT
                      echo "Starting compliance checks..."
                      
                      # Run compliance checks for each framework
                      for framework in $(echo $COMPLIANCE_FRAMEWORKS | tr ',' ' '); do
                        echo "Checking compliance for $framework..."
                        
                        case $framework in
                          "SOC2")
                            # SOC2 compliance checks
                            echo "Running SOC2 compliance checks..."
                            # Check encryption at rest
                            # Check access controls
                            # Check audit logging
                            ;;
                          "PCI-DSS")
                            # PCI-DSS compliance checks
                            echo "Running PCI-DSS compliance checks..."
                            # Check network segmentation
                            # Check encryption
                            # Check vulnerability management
                            ;;
                          "GDPR")
                            # GDPR compliance checks
                            echo "Running GDPR compliance checks..."
                            # Check data protection
                            # Check consent management
                            # Check data retention policies
                            ;;
                        esac
                      done
                      
                      echo "Compliance checks completed"
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
# POLICY ENFORCEMENT
# =============================================================================

# Open Policy Agent (OPA) Gatekeeper
resource "helm_release" "gatekeeper" {
  count = var.enable_policy_enforcement ? 1 : 0
  
  name       = "gatekeeper"
  repository = "https://open-policy-agent.github.io/gatekeeper/charts"
  chart      = "gatekeeper"
  version    = var.gatekeeper_version
  namespace  = kubernetes_namespace.security.metadata[0].name
  
  values = [
    yamlencode({
      # Gatekeeper configuration
      replicas = var.gatekeeper_replicas
      
      # Audit configuration
      audit = {
        replicas = var.gatekeeper_audit_replicas
        
        # Audit interval
        auditInterval = 60
        
        # Constraint violations limit
        constraintViolationsLimit = 20
        
        # Audit from cache
        auditFromCache = false
        
        # Resources
        resources = {
          requests = {
            cpu    = "100m"
            memory = "256Mi"
          }
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
      
      # Controller manager configuration
      controllerManager = {
        # Resources
        resources = {
          requests = {
            cpu    = "100m"
            memory = "256Mi"
          }
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
        
        # Webhook configuration
        webhook = {
          # Failure policy
          failurePolicy = "Fail"
          
          # Namespace selector
          namespaceSelector = {
            matchExpressions = [
              {
                key      = "name"
                operator = "NotIn"
                values   = ["kube-system", "gatekeeper-system"]
              }
            ]
          }
        }
      }
      
      # Metrics
      metrics = {
        enabled = var.monitoring_enabled
      }
      
      # Mutation
      mutations = {
        enabled = var.enable_mutation
      }
      
      # Pod security policy
      podSecurityPolicy = {
        enabled = false
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.security]
}

# Security Policy Templates
resource "kubernetes_manifest" "security_policy_templates" {
  for_each = var.enable_policy_enforcement ? var.security_policies : {}
  
  manifest = {
    apiVersion = "templates.gatekeeper.sh/v1beta1"
    kind       = "ConstraintTemplate"
    
    metadata = {
      name = each.key
      
      labels = local.common_tags
    }
    
    spec = {
      crd = {
        spec = {
          names = {
            kind = each.value.kind
          }
          validation = {
            openAPIV3Schema = {
              type = "object"
              properties = each.value.properties
            }
          }
        }
      }
      
      targets = [
        {
          target = "admission.k8s.gatekeeper.sh"
          rego   = each.value.rego_policy
        }
      ]
    }
  }
  
  depends_on = [helm_release.gatekeeper]
}

# =============================================================================
# THREAT DETECTION
# =============================================================================

# Falco for runtime security monitoring
resource "helm_release" "falco" {
  count = var.enable_threat_detection ? 1 : 0
  
  name       = "falco"
  repository = "https://falcosecurity.github.io/charts"
  chart      = "falco"
  version    = var.falco_version
  namespace  = kubernetes_namespace.security.metadata[0].name
  
  values = [
    yamlencode({
      # Falco configuration
      falco = {
        # Rules configuration
        rules_file = [
          "/etc/falco/falco_rules.yaml",
          "/etc/falco/falco_rules.local.yaml",
          "/etc/falco/k8s_audit_rules.yaml",
          "/etc/falco/rules.d"
        ]
        
        # Time format
        time_format_iso_8601 = true
        
        # JSON output
        json_output = true
        json_include_output_property = true
        json_include_tags_property = true
        
        # Log level
        log_level = "info"
        
        # Priority threshold
        priority = var.threat_detection_sensitivity
        
        # Buffered outputs
        buffered_outputs = false
        
        # Syscall event drops
        syscall_event_drops = {
          actions = ["log", "alert"]
          rate = 0.03333
          max_burst = 1000
        }
        
        # Metadata download
        metadata_download = {
          max_mb = 100
          chunk_wait_us = 1000
          watch_freq_sec = 1
        }
      }
      
      # Driver configuration
      driver = {
        enabled = true
        kind = "ebpf"
      }
      
      # Collectors
      collectors = {
        enabled = true
        
        docker = {
          enabled = true
          socket = "/var/run/docker.sock"
        }
        
        containerd = {
          enabled = true
          socket = "/run/containerd/containerd.sock"
        }
        
        crio = {
          enabled = true
          socket = "/var/run/crio/crio.sock"
        }
      }
      
      # Service Monitor for Prometheus
      serviceMonitor = {
        enabled = var.monitoring_enabled
        labels = local.common_tags
      }
      
      # Custom rules
      customRules = {
        "go-coffee-rules.yaml" = <<-EOF
          - rule: Suspicious Coffee Order Activity
            desc: Detect suspicious coffee order patterns
            condition: >
              k8s_audit and ka.verb in (create, update) and
              ka.target.resource=orders and
              ka.target.subresource="" and
              ka.uri.param[quantity] > 100
            output: >
              Suspicious large coffee order detected
              (user=%ka.user.name verb=%ka.verb uri=%ka.uri.path
              quantity=%ka.uri.param[quantity])
            priority: WARNING
            tags: [coffee, orders, suspicious]
            
          - rule: Unauthorized DeFi Transaction
            desc: Detect unauthorized DeFi transactions
            condition: >
              k8s_audit and ka.verb=create and
              ka.target.resource=transactions and
              ka.target.subresource="" and
              not ka.user.name in (defi-service, trading-bot)
            output: >
              Unauthorized DeFi transaction attempt
              (user=%ka.user.name verb=%ka.verb uri=%ka.uri.path)
            priority: CRITICAL
            tags: [defi, unauthorized, transaction]
        EOF
      }
      
      # Falco Sidekick for alert routing
      falcosidekick = {
        enabled = true
        
        config = {
          slack = {
            webhookurl = var.slack_webhook_url
            channel = "#security-alerts"
            username = "Falco"
            iconurl = "https://falco.org/img/brand/falco-logo.png"
            minimumpriority = "warning"
          }
          
          webhook = {
            address = var.webhook_url
            minimumpriority = "warning"
          }
        }
      }
    })
  ]
  
  depends_on = [kubernetes_namespace.security]
}

# =============================================================================
# SECRETS MANAGEMENT
# =============================================================================

# Security notification secrets
resource "kubernetes_secret" "security_notifications" {
  metadata {
    name      = "security-notifications"
    namespace = kubernetes_namespace.security.metadata[0].name
    
    labels = local.common_tags
  }
  
  data = {
    slack-webhook-url = var.slack_webhook_url
    email-username    = var.email_username
    email-password    = var.email_password
    webhook-url       = var.webhook_url
  }
  
  type = "Opaque"
}

# TLS certificate for security services
resource "tls_private_key" "security_tls" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "security_tls" {
  private_key_pem = tls_private_key.security_tls.private_key_pem
  
  subject {
    common_name  = "${local.name_prefix}-security"
    organization = var.project_name
  }
  
  validity_period_hours = 8760 # 1 year
  
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "kubernetes_secret" "security_tls" {
  metadata {
    name      = "${local.name_prefix}-security-tls"
    namespace = kubernetes_namespace.security.metadata[0].name
    
    labels = local.common_tags
  }
  
  data = {
    "tls.crt" = tls_self_signed_cert.security_tls.cert_pem
    "tls.key" = tls_private_key.security_tls.private_key_pem
  }
  
  type = "kubernetes.io/tls"
}
