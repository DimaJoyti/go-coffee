resource "google_container_cluster" "primary" {
  name     = "${var.gke_cluster_name}-${var.environment}"
  location = var.region

  # We create a cluster with minimal nodes
  # and then use node pool for actual nodes
  remove_default_node_pool = true
  initial_node_count       = 1

  # Network configuration
  network    = var.network_name
  subnetwork = var.subnet_name

  # Private cluster configuration
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  # Master authorized networks configuration
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"  # In production, restrict to specific IPs
      display_name = "All"
    }
  }

  # Workload identity configuration
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
}

resource "google_container_node_pool" "primary_nodes" {
  name       = "${var.gke_cluster_name}-node-pool"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = var.node_count

  node_config {
    preemptible  = var.node_preemptible
    machine_type = var.node_machine_type
    disk_size_gb = var.node_disk_size_gb
    disk_type    = var.node_disk_type

    # Google recommends custom service accounts with minimal permissions
    service_account = google_service_account.gke_sa.email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    # Workload identity configuration
    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    labels = {
      environment = var.environment
    }

    tags = ["gke-node", "${var.gke_cluster_name}-${var.environment}"]
  }
}

resource "google_service_account" "gke_sa" {
  account_id   = "${var.gke_cluster_name}-sa-${var.environment}"
  display_name = "GKE Service Account for ${var.gke_cluster_name} ${var.environment}"
}

resource "google_project_iam_member" "gke_sa_roles" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/monitoring.viewer",
    "roles/storage.objectViewer",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.gke_sa.email}"
}

# Output the instance group
output "instance_group" {
  value = google_container_node_pool.primary_nodes.instance_group_urls[0]
}
