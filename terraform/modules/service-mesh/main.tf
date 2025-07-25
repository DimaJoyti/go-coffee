# Service Mesh Module for Go Coffee
# Provides Istio service mesh with security, observability, and traffic management

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
}

# Get GKE cluster data
data "google_container_cluster" "cluster" {
  name     = var.cluster_name
  location = var.cluster_location
  project  = var.project_id
}

# Configure Kubernetes provider
provider "kubernetes" {
  host                   = "https://${data.google_container_cluster.cluster.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth.0.cluster_ca_certificate)
}

# Configure Helm provider
provider "helm" {
  kubernetes {
    host                   = "https://${data.google_container_cluster.cluster.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth.0.cluster_ca_certificate)
  }
}

data "google_client_config" "default" {}

# Create Istio system namespace
resource "kubernetes_namespace" "istio_system" {
  count = var.enable_istio ? 1 : 0

  metadata {
    name = "istio-system"
    labels = {
      "istio-injection" = "disabled"
      "environment"     = var.environment
      "managed-by"      = "terraform"
    }
  }
}

# Install Istio base
resource "helm_release" "istio_base" {
  count = var.enable_istio ? 1 : 0

  name       = "istio-base"
  repository = "https://istio-release.storage.googleapis.com/charts"
  chart      = "base"
  version    = var.istio_version
  namespace  = kubernetes_namespace.istio_system[0].metadata[0].name

  set {
    name  = "global.meshID"
    value = "mesh1"
  }

  set {
    name  = "global.multiCluster.clusterName"
    value = var.cluster_name
  }

  set {
    name  = "global.network"
    value = "network1"
  }

  depends_on = [kubernetes_namespace.istio_system]
}

# Install Istiod (control plane)
resource "helm_release" "istiod" {
  count = var.enable_istio ? 1 : 0

  name       = "istiod"
  repository = "https://istio-release.storage.googleapis.com/charts"
  chart      = "istiod"
  version    = var.istio_version
  namespace  = kubernetes_namespace.istio_system[0].metadata[0].name

  set {
    name  = "global.meshID"
    value = "mesh1"
  }

  set {
    name  = "global.multiCluster.clusterName"
    value = var.cluster_name
  }

  set {
    name  = "global.network"
    value = "network1"
  }

  # Enable tracing
  set {
    name  = "pilot.traceSampling"
    value = var.enable_tracing ? "100.0" : "1.0"
  }

  # Security settings
  set {
    name  = "global.mtls.auto"
    value = var.enable_mtls
  }

  depends_on = [helm_release.istio_base]
}

# Install Istio Gateway
resource "helm_release" "istio_gateway" {
  count = var.enable_istio ? 1 : 0

  name       = "istio-gateway"
  repository = "https://istio-release.storage.googleapis.com/charts"
  chart      = "gateway"
  version    = var.istio_version
  namespace  = "istio-ingress"

  create_namespace = true

  set {
    name  = "service.type"
    value = "LoadBalancer"
  }

  set {
    name  = "service.annotations.cloud\\.google\\.com/load-balancer-type"
    value = "External"
  }

  depends_on = [helm_release.istiod]
}

# Create Go Coffee namespace with Istio injection
resource "kubernetes_namespace" "go_coffee" {
  count = var.enable_istio ? 1 : 0

  metadata {
    name = "go-coffee"
    labels = {
      "istio-injection" = "enabled"
      "environment"     = var.environment
      "managed-by"      = "terraform"
    }
  }

  depends_on = [helm_release.istiod]
}

# Peer Authentication for mTLS
resource "kubernetes_manifest" "peer_authentication" {
  count = var.enable_istio && var.enable_mtls ? 1 : 0

  manifest = {
    apiVersion = "security.istio.io/v1beta1"
    kind       = "PeerAuthentication"
    metadata = {
      name      = "default"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
    spec = {
      mtls = {
        mode = "STRICT"
      }
    }
  }

  depends_on = [kubernetes_namespace.go_coffee]
}

# Authorization Policy for Go Coffee services
resource "kubernetes_manifest" "authorization_policy" {
  count = var.enable_istio ? 1 : 0

  manifest = {
    apiVersion = "security.istio.io/v1beta1"
    kind       = "AuthorizationPolicy"
    metadata = {
      name      = "go-coffee-authz"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
    spec = {
      rules = [
        {
          from = [
            {
              source = {
                principals = ["cluster.local/ns/go-coffee/sa/default"]
              }
            }
          ]
          to = [
            {
              operation = {
                methods = ["GET", "POST", "PUT", "DELETE"]
              }
            }
          ]
        }
      ]
    }
  }

  depends_on = [kubernetes_namespace.go_coffee]
}

# Virtual Service for Go Coffee API Gateway
resource "kubernetes_manifest" "api_gateway_virtual_service" {
  count = var.enable_istio ? 1 : 0

  manifest = {
    apiVersion = "networking.istio.io/v1beta1"
    kind       = "VirtualService"
    metadata = {
      name      = "go-coffee-api-gateway"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
    spec = {
      hosts = ["*"]
      gateways = ["go-coffee-gateway"]
      http = [
        {
          match = [
            {
              uri = {
                prefix = "/api/v1/"
              }
            }
          ]
          route = [
            {
              destination = {
                host = "api-gateway"
                port = {
                  number = 8080
                }
              }
            }
          ]
          timeout = "30s"
          retries = {
            attempts      = 3
            perTryTimeout = "10s"
          }
        }
      ]
    }
  }

  depends_on = [kubernetes_namespace.go_coffee]
}

# Gateway for external traffic
resource "kubernetes_manifest" "go_coffee_gateway" {
  count = var.enable_istio ? 1 : 0

  manifest = {
    apiVersion = "networking.istio.io/v1beta1"
    kind       = "Gateway"
    metadata = {
      name      = "go-coffee-gateway"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
    spec = {
      selector = {
        istio = "gateway"
      }
      servers = [
        {
          port = {
            number   = 80
            name     = "http"
            protocol = "HTTP"
          }
          hosts = ["*"]
        },
        {
          port = {
            number   = 443
            name     = "https"
            protocol = "HTTPS"
          }
          hosts = ["*"]
          tls = {
            mode = "SIMPLE"
            credentialName = var.tls_secret_name
          }
        }
      ]
    }
  }

  depends_on = [kubernetes_namespace.go_coffee]
}

# Destination Rules for circuit breaking and load balancing
resource "kubernetes_manifest" "destination_rules" {
  for_each = var.enable_istio ? var.services : {}

  manifest = {
    apiVersion = "networking.istio.io/v1beta1"
    kind       = "DestinationRule"
    metadata = {
      name      = "${each.key}-destination-rule"
      namespace = kubernetes_namespace.go_coffee[0].metadata[0].name
    }
    spec = {
      host = each.key
      trafficPolicy = {
        connectionPool = {
          tcp = {
            maxConnections = 100
          }
          http = {
            http1MaxPendingRequests  = 50
            http2MaxRequests         = 100
            maxRequestsPerConnection = 10
            maxRetries               = 3
            consecutiveGatewayErrors = 5
            interval                 = "30s"
            baseEjectionTime         = "30s"
          }
        }
        circuitBreaker = {
          consecutiveGatewayErrors = 5
          interval                 = "30s"
          baseEjectionTime         = "30s"
          maxEjectionPercent       = 50
        }
        loadBalancer = {
          simple = "LEAST_CONN"
        }
      }
    }
  }

  depends_on = [kubernetes_namespace.go_coffee]
}
