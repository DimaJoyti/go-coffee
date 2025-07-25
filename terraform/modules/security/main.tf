# Security Module for Go Coffee
# Provides comprehensive security controls including RBAC, network policies, and compliance

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

data "google_client_config" "default" {}

# Enable Binary Authorization for production
resource "google_binary_authorization_policy" "policy" {
  count = var.enable_binary_authorization ? 1 : 0

  project = var.project_id

  admission_whitelist_patterns {
    name_pattern = "gcr.io/${var.project_id}/*"
  }

  admission_whitelist_patterns {
    name_pattern = "ghcr.io/dimajoyti/go-coffee/*"
  }
  
  # Allow Google system images
  admission_whitelist_patterns {
    name_pattern = "gcr.io/gke-release/*"
  }
  
  admission_whitelist_patterns {
    name_pattern = "k8s.gcr.io/*"
  }

  default_admission_rule {
    evaluation_mode  = "REQUIRE_ATTESTATION"
    enforcement_mode = "ENFORCED_BLOCK_AND_AUDIT_LOG"

    require_attestations_by = [
      google_binary_authorization_attestor.attestor[0].name
    ]
  }

  cluster_admission_rules {
    cluster                = data.google_container_cluster.cluster.id
    evaluation_mode        = "REQUIRE_ATTESTATION"
    enforcement_mode       = "ENFORCED_BLOCK_AND_AUDIT_LOG"
    
    require_attestations_by = [
      google_binary_authorization_attestor.attestor[0].name
    ]
  }
}

# Create attestor for Binary Authorization
resource "google_binary_authorization_attestor" "attestor" {
  count = var.enable_binary_authorization ? 1 : 0

  name    = "go-coffee-attestor"
  project = var.project_id

  attestation_authority_note {
    note_reference = google_container_analysis_note.note[0].name
    public_keys {
      ascii_armored_pgp_public_key = var.pgp_public_key
    }
  }
}

# Create Container Analysis note
resource "google_container_analysis_note" "note" {
  count = var.enable_binary_authorization ? 1 : 0

  name    = "go-coffee-attestor-note"
  project = var.project_id

  attestation_authority {
    hint {
      human_readable_name = "Go Coffee Attestor"
    }
  }
}

# Create service accounts for workload identity
resource "google_service_account" "workload_identity" {
  for_each = var.enable_workload_identity ? var.service_accounts : {}

  account_id   = each.key
  display_name = each.value.display_name
  description  = each.value.description
  project      = var.project_id
}

# Bind service accounts to Kubernetes service accounts
resource "google_service_account_iam_binding" "workload_identity" {
  for_each = var.enable_workload_identity ? var.service_accounts : {}

  service_account_id = google_service_account.workload_identity[each.key].name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[${each.value.namespace}/${each.value.ksa_name}]"
  ]
}

# Grant necessary IAM roles to service accounts
resource "google_project_iam_member" "service_account_roles" {
  for_each = var.enable_workload_identity ? local.sa_role_bindings : {}

  project = var.project_id
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.workload_identity[each.value.sa_name].email}"
}

# Local values for service account role bindings
locals {
  sa_role_bindings = var.enable_workload_identity ? {
    for pair in flatten([
      for sa_name, sa_config in var.service_accounts : [
        for role in sa_config.roles : {
          key     = "${sa_name}-${replace(role, "/", "-")}"
          sa_name = sa_name
          role    = role
        }
      ]
    ]) : pair.key => pair
  } : {}
}

# Create Kubernetes service accounts
resource "kubernetes_service_account" "workload_identity" {
  for_each = var.enable_workload_identity ? var.service_accounts : {}

  metadata {
    name      = each.value.ksa_name
    namespace = each.value.namespace
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.workload_identity[each.key].email
    }
    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
      "environment"                  = var.environment
    }
  }

  automount_service_account_token = true
}

# Network Policies for Go Coffee namespace
resource "kubernetes_network_policy" "go_coffee_default_deny" {
  count = var.enable_network_policy ? 1 : 0

  metadata {
    name      = "default-deny-all"
    namespace = "go-coffee"
  }

  spec {
    pod_selector {}
    policy_types = ["Ingress", "Egress"]
  }
}

resource "kubernetes_network_policy" "go_coffee_allow_internal" {
  count = var.enable_network_policy ? 1 : 0

  metadata {
    name      = "allow-internal-communication"
    namespace = "go-coffee"
  }

  spec {
    pod_selector {
      match_labels = {
        "app.kubernetes.io/name" = "go-coffee"
      }
    }

    policy_types = ["Ingress", "Egress"]

    ingress {
      from {
        namespace_selector {
          match_labels = {
            name = "go-coffee"
          }
        }
      }
      from {
        namespace_selector {
          match_labels = {
            name = "istio-system"
          }
        }
      }
    }

    egress {
      # Allow communication within go-coffee namespace
      to {
        namespace_selector {
          match_labels = {
            name = "go-coffee"
          }
        }
      }
    }

    egress {
      # Allow communication to istio-system namespace
      to {
        namespace_selector {
          match_labels = {
            name = "istio-system"
          }
        }
      }
    }

    egress {
      # Allow DNS queries
      to {}
      ports {
        protocol = "UDP"
        port     = "53"
      }
    }

    egress {
      # Allow HTTPS to external services (Google APIs, etc.)
      to {}
      ports {
        protocol = "TCP"
        port     = "443"
      }
    }

    egress {
      # Allow HTTP for health checks and internal communication
      to {}
      ports {
        protocol = "TCP"
        port     = "80"
      }
    }
  }
}

# Pod Security Standards
resource "kubernetes_manifest" "pod_security_policy" {
  count = var.enable_pod_security_policy ? 1 : 0

  manifest = {
    apiVersion = "v1"
    kind       = "Namespace"
    metadata = {
      name = "go-coffee"
      labels = {
        "pod-security.kubernetes.io/enforce" = "restricted"
        "pod-security.kubernetes.io/audit"   = "restricted"
        "pod-security.kubernetes.io/warn"    = "restricted"
      }
    }
  }
}

# RBAC for Go Coffee services
resource "kubernetes_role" "go_coffee_role" {
  count = var.enable_rbac ? 1 : 0

  metadata {
    namespace = "go-coffee"
    name      = "go-coffee-role"
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "services", "endpoints", "configmaps", "secrets"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "replicasets"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["networking.k8s.io"]
    resources  = ["networkpolicies"]
    verbs      = ["get", "list", "watch"]
  }
}

resource "kubernetes_role_binding" "go_coffee_role_binding" {
  count = var.enable_rbac ? 1 : 0

  metadata {
    name      = "go-coffee-role-binding"
    namespace = "go-coffee"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.go_coffee_role[0].metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "default"
    namespace = "go-coffee"
  }
}

# Security monitoring and alerting
resource "google_monitoring_alert_policy" "security_violations" {
  count = var.enable_security_monitoring ? 1 : 0

  display_name = "Security Violations - Go Coffee"
  project      = var.project_id
  
  # Required combiner for alert policy conditions
  combiner = "OR"

  conditions {
    display_name = "Binary Authorization violations"

    condition_threshold {
      filter          = "resource.type=\"k8s_cluster\" AND protoPayload.serviceName=\"binaryauthorization.googleapis.com\" AND protoPayload.methodName=\"binauthz.Verify\" AND protoPayload.response.decision!=\"ALLOW\""
      duration        = "60s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["resource.label.cluster_name", "resource.label.location"]
      }

      trigger {
        count = 1
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "1800s"
  }
  
  documentation {
    content   = "Alert triggered when Binary Authorization violations are detected in the Go Coffee cluster. This indicates attempts to deploy unauthorized container images."
    mime_type = "text/markdown"
  }
}

# Additional security monitoring alerts
resource "google_monitoring_alert_policy" "pod_security_violations" {
  count = var.enable_security_monitoring ? 1 : 0

  display_name = "Pod Security Policy Violations - Go Coffee"
  project      = var.project_id
  
  combiner = "OR"

  conditions {
    display_name = "Pod Security Standard violations"
    
    condition_threshold {
      filter          = "resource.type=\"k8s_cluster\" AND protoPayload.serviceName=\"k8s.io\" AND protoPayload.methodName=\"admission.k8s.io/v1.AdmissionReview\" AND protoPayload.response.allowed=false AND protoPayload.request.namespace=\"go-coffee\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["resource.label.cluster_name", "resource.label.namespace_name"]
      }

      trigger {
        count = 1
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "3600s"
  }
  
  documentation {
    content   = "Alert triggered when Pod Security Standard violations are detected. This indicates attempts to create pods that violate security policies."
    mime_type = "text/markdown"
  }
}

# Network policy violations
resource "google_monitoring_alert_policy" "network_policy_violations" {
  count = var.enable_security_monitoring ? 1 : 0

  display_name = "Network Policy Violations - Go Coffee"
  project      = var.project_id
  
  combiner = "OR"

  conditions {
    display_name = "Blocked network connections"
    
    condition_threshold {
      filter          = "resource.type=\"k8s_cluster\" AND jsonPayload.kind=\"Event\" AND jsonPayload.reason=\"NetworkPolicyViolation\" AND resource.labels.namespace_name=\"go-coffee\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 10

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["resource.label.cluster_name", "resource.label.namespace_name"]
      }

      trigger {
        count = 1
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "1800s"
  }
  
  documentation {
    content   = "Alert triggered when network traffic is being blocked by network policies. High rates may indicate misconfigurations or potential security threats."
    mime_type = "text/markdown"
  }
}

# Workload Identity violations
resource "google_monitoring_alert_policy" "workload_identity_violations" {
  count = var.enable_security_monitoring ? 1 : 0

  display_name = "Workload Identity Violations - Go Coffee"
  project      = var.project_id
  
  combiner = "OR"

  conditions {
    display_name = "Workload Identity authentication failures"
    
    condition_threshold {
      filter          = "resource.type=\"gce_instance\" AND protoPayload.serviceName=\"iamcredentials.googleapis.com\" AND protoPayload.methodName=\"GenerateAccessToken\" AND protoPayload.authorizationInfo.granted=false AND protoPayload.authenticationInfo.principalEmail=~\".*go-coffee.*\""
      duration        = "180s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 5

      aggregations {
        alignment_period     = "180s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
        group_by_fields      = ["protoPayload.authenticationInfo.principalEmail"]
      }

      trigger {
        count = 1
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "1800s"
  }
  
  documentation {
    content   = "Alert triggered when Workload Identity authentication failures are detected. This may indicate misconfigurations or unauthorized access attempts."
    mime_type = "text/markdown"
  }
}
