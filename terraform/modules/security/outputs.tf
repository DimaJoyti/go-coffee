# Security Module Outputs

output "binary_authorization_enabled" {
  description = "Whether Binary Authorization is enabled"
  value       = var.enable_binary_authorization
}

output "binary_authorization_policy_id" {
  description = "Binary Authorization policy ID"
  value       = var.enable_binary_authorization ? google_binary_authorization_policy.policy[0].id : null
}

output "attestor_name" {
  description = "Binary Authorization attestor name"
  value       = var.enable_binary_authorization ? google_binary_authorization_attestor.attestor[0].name : null
}

output "workload_identity_enabled" {
  description = "Whether Workload Identity is enabled"
  value       = var.enable_workload_identity
}

output "service_accounts" {
  description = "Created Google service accounts for workload identity"
  value = var.enable_workload_identity ? {
    for sa_name, sa in google_service_account.workload_identity : sa_name => {
      email        = sa.email
      unique_id    = sa.unique_id
      display_name = sa.display_name
    }
  } : {}
}

output "kubernetes_service_accounts" {
  description = "Created Kubernetes service accounts"
  value = var.enable_workload_identity ? {
    for sa_name, sa in kubernetes_service_account.workload_identity : sa_name => {
      name      = sa.metadata[0].name
      namespace = sa.metadata[0].namespace
    }
  } : {}
}

output "network_policies_enabled" {
  description = "Whether network policies are enabled"
  value       = var.enable_network_policy
}

output "pod_security_policy_enabled" {
  description = "Whether Pod Security Standards are enabled"
  value       = var.enable_pod_security_policy
}

output "rbac_enabled" {
  description = "Whether RBAC is enabled"
  value       = var.enable_rbac
}

output "security_monitoring_enabled" {
  description = "Whether security monitoring is enabled"
  value       = var.enable_security_monitoring
}

output "security_configuration" {
  description = "Security configuration summary"
  value = {
    binary_authorization    = var.enable_binary_authorization
    workload_identity      = var.enable_workload_identity
    network_policies       = var.enable_network_policy
    pod_security_standards = var.enable_pod_security_policy
    rbac                   = var.enable_rbac
    security_monitoring    = var.enable_security_monitoring
    environment            = var.environment
  }
}

output "compliance_frameworks" {
  description = "Enabled compliance frameworks"
  value       = var.compliance_frameworks
}

output "pod_security_standards" {
  description = "Pod Security Standards configuration"
  value       = var.pod_security_standards
}

output "network_policy_config" {
  description = "Network policy configuration"
  value       = var.network_policy_config
}

output "rbac_config" {
  description = "RBAC configuration"
  value       = var.rbac_config
}

output "security_scanning_config" {
  description = "Security scanning configuration"
  value       = var.security_scanning_config
}

output "workload_identity_bindings" {
  description = "Workload identity bindings between GSA and KSA"
  value = var.enable_workload_identity ? {
    for sa_name, sa_config in var.service_accounts : sa_name => {
      gsa_email     = google_service_account.workload_identity[sa_name].email
      ksa_name      = sa_config.ksa_name
      ksa_namespace = sa_config.namespace
      roles         = sa_config.roles
    }
  } : {}
}

output "security_alerts" {
  description = "Configured security alert policies"
  value = var.enable_security_monitoring ? [
    {
      name         = "Security Violations - Go Coffee"
      display_name = "Binary Authorization violations"
      project      = var.project_id
    }
  ] : []
}

output "container_analysis_note" {
  description = "Container Analysis note for attestation"
  value       = var.enable_binary_authorization ? google_container_analysis_note.note[0].name : null
}

output "security_best_practices" {
  description = "Security best practices implemented"
  value = {
    least_privilege_access = var.enable_rbac && var.enable_workload_identity
    network_segmentation   = var.enable_network_policy
    container_security     = var.enable_binary_authorization
    runtime_security       = var.enable_pod_security_policy
    monitoring_alerting    = var.enable_security_monitoring
    compliance_ready       = length(var.compliance_frameworks) > 0
  }
}

output "security_recommendations" {
  description = "Security recommendations for improvement"
  value = {
    enable_binary_authorization = !var.enable_binary_authorization ? "Consider enabling Binary Authorization for production workloads" : null
    enable_workload_identity   = !var.enable_workload_identity ? "Enable Workload Identity for secure GCP service access" : null
    enable_network_policies    = !var.enable_network_policy ? "Enable Network Policies for network segmentation" : null
    enable_security_monitoring = !var.enable_security_monitoring ? "Enable security monitoring for threat detection" : null
  }
}
