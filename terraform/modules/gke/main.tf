# Модуль для створення кластера GKE

# Створення кластера GKE
resource "google_container_cluster" "primary" {
  name     = "${var.gke_cluster_name}-${var.environment}"
  location = var.region

  # Ми створюємо кластер з мінімальною кількістю вузлів
  # і потім використовуємо node pool для фактичних вузлів
  remove_default_node_pool = true
  initial_node_count       = 1

  # Налаштування мережі
  network    = var.network_name
  subnetwork = var.subnet_name

  # Налаштування приватного кластера
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  # Налаштування IP-адрес для master
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"  # В продакшн слід обмежити до конкретних IP-адрес
      display_name = "All"
    }
  }

  # Увімкнення Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Налаштування логування та моніторингу
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  # Налаштування мережевої політики
  network_policy {
    enabled  = true
    provider = "CALICO"
  }

  # Налаштування автоматичного оновлення вузлів
  release_channel {
    channel = "REGULAR"
  }
}

# Створення node pool для кластера
resource "google_container_node_pool" "primary_nodes" {
  name       = "${var.gke_cluster_name}-node-pool-${var.environment}"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = var.gke_node_count

  # Налаштування автоскейлінгу
  autoscaling {
    min_node_count = var.gke_min_node_count
    max_node_count = var.gke_max_node_count
  }

  # Налаштування автоматичного оновлення вузлів
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  # Налаштування вузлів
  node_config {
    preemptible  = var.environment != "prod"  # Використовуємо preemptible для dev/staging
    machine_type = var.gke_machine_type

    # Налаштування дисків
    disk_type    = "pd-standard"
    disk_size_gb = 100

    # Налаштування OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/compute",
    ]

    # Налаштування Workload Identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    # Мітки для вузлів
    labels = {
      environment = var.environment
    }

    # Теги для вузлів
    tags = ["gke-node", var.environment]
  }
}
