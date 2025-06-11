# Go Coffee GCP Infrastructure Outputs

# Network Outputs
output "network_name" {
  description = "Name of the VPC network"
  value       = google_compute_network.vpc.name
}

output "network_id" {
  description = "ID of the VPC network"
  value       = google_compute_network.vpc.id
}

output "subnet_name" {
  description = "Name of the subnet"
  value       = google_compute_subnetwork.subnet.name
}

output "subnet_id" {
  description = "ID of the subnet"
  value       = google_compute_subnetwork.subnet.id
}

output "subnet_cidr" {
  description = "CIDR block of the subnet"
  value       = google_compute_subnetwork.subnet.ip_cidr_range
}

# GKE Cluster Outputs
output "cluster_name" {
  description = "Name of the GKE cluster"
  value       = google_container_cluster.primary.name
}

output "cluster_id" {
  description = "ID of the GKE cluster"
  value       = google_container_cluster.primary.id
}

output "cluster_endpoint" {
  description = "Endpoint of the GKE cluster"
  value       = google_container_cluster.primary.endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "CA certificate of the GKE cluster"
  value       = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "cluster_location" {
  description = "Location of the GKE cluster"
  value       = google_container_cluster.primary.location
}

output "cluster_node_version" {
  description = "Node version of the GKE cluster"
  value       = google_container_cluster.primary.node_version
}

output "cluster_master_version" {
  description = "Master version of the GKE cluster"
  value       = google_container_cluster.primary.master_version
}

# Node Pool Outputs
output "node_pool_name" {
  description = "Name of the primary node pool"
  value       = google_container_node_pool.primary_nodes.name
}

output "node_pool_instance_group_urls" {
  description = "Instance group URLs of the node pool"
  value       = google_container_node_pool.primary_nodes.instance_group_urls
}

# Service Account Outputs
output "gke_service_account_email" {
  description = "Email of the GKE nodes service account"
  value       = google_service_account.gke_nodes.email
}

output "gke_service_account_id" {
  description = "ID of the GKE nodes service account"
  value       = google_service_account.gke_nodes.id
}

# Database Outputs
output "postgres_instance_name" {
  description = "Name of the PostgreSQL instance"
  value       = google_sql_database_instance.postgres.name
}

output "postgres_connection_name" {
  description = "Connection name of the PostgreSQL instance"
  value       = google_sql_database_instance.postgres.connection_name
}

output "postgres_private_ip" {
  description = "Private IP address of the PostgreSQL instance"
  value       = google_sql_database_instance.postgres.private_ip_address
  sensitive   = true
}

output "postgres_public_ip" {
  description = "Public IP address of the PostgreSQL instance"
  value       = google_sql_database_instance.postgres.public_ip_address
  sensitive   = true
}

# Redis Outputs
output "redis_instance_id" {
  description = "ID of the Redis instance"
  value       = google_redis_instance.cache.id
}

output "redis_host" {
  description = "Host of the Redis instance"
  value       = google_redis_instance.cache.host
  sensitive   = true
}

output "redis_port" {
  description = "Port of the Redis instance"
  value       = google_redis_instance.cache.port
}

output "redis_auth_string" {
  description = "Auth string of the Redis instance"
  value       = google_redis_instance.cache.auth_string
  sensitive   = true
}

# Regional Configuration Outputs
output "region" {
  description = "GCP region"
  value       = local.region
}

output "zone" {
  description = "GCP zone"
  value       = local.zone
}

output "project_id" {
  description = "GCP project ID"
  value       = local.project_id
}

# Kubernetes Configuration
output "kubeconfig_raw" {
  description = "Raw kubeconfig for the GKE cluster"
  value = templatefile("${path.module}/kubeconfig-template.yaml", {
    cluster_name           = google_container_cluster.primary.name
    cluster_endpoint       = google_container_cluster.primary.endpoint
    cluster_ca_certificate = google_container_cluster.primary.master_auth[0].cluster_ca_certificate
    project_id            = local.project_id
    region                = local.region
  })
  sensitive = true
}

# Connection Information
output "connection_info" {
  description = "Connection information for services"
  value = {
    cluster = {
      name     = google_container_cluster.primary.name
      endpoint = google_container_cluster.primary.endpoint
      location = google_container_cluster.primary.location
    }
    database = {
      host             = google_sql_database_instance.postgres.private_ip_address
      connection_name  = google_sql_database_instance.postgres.connection_name
      port            = 5432
    }
    redis = {
      host = google_redis_instance.cache.host
      port = google_redis_instance.cache.port
    }
    network = {
      vpc_name    = google_compute_network.vpc.name
      subnet_name = google_compute_subnetwork.subnet.name
      subnet_cidr = google_compute_subnetwork.subnet.ip_cidr_range
    }
  }
  sensitive = true
}

# Resource URLs for CLI
output "resource_urls" {
  description = "URLs for accessing resources"
  value = {
    gke_cluster = "https://console.cloud.google.com/kubernetes/clusters/details/${local.region}/${google_container_cluster.primary.name}/details?project=${local.project_id}"
    sql_instance = "https://console.cloud.google.com/sql/instances/${google_sql_database_instance.postgres.name}/overview?project=${local.project_id}"
    redis_instance = "https://console.cloud.google.com/memorystore/redis/locations/${local.region}/instances/${google_redis_instance.cache.name}/details/overview?project=${local.project_id}"
    vpc_network = "https://console.cloud.google.com/networking/networks/details/${google_compute_network.vpc.name}?project=${local.project_id}"
  }
}

# Cost Estimation
output "estimated_monthly_cost" {
  description = "Estimated monthly cost breakdown"
  value = {
    gke_cluster = "~$73/month (1 node e2-standard-4)"
    postgres_db = "~$45/month (db-custom-2-4096)"
    redis_cache = "~$25/month (4GB STANDARD_HA)"
    networking  = "~$5/month (NAT Gateway)"
    total_estimated = "~$148/month"
    note = "Costs are estimates and may vary based on actual usage"
  }
}
