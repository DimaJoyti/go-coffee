# Модуль для створення мережевої інфраструктури

# Створення VPC мережі
resource "google_compute_network" "vpc_network" {
  name                    = "${var.network_name}-${var.environment}"
  auto_create_subnetworks = false
  description             = "VPC мережа для системи замовлення кави"
}

# Створення підмережі
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.subnet_name}-${var.environment}"
  region        = var.region
  network       = google_compute_network.vpc_network.id
  ip_cidr_range = var.subnet_cidr
  description   = "Підмережа для системи замовлення кави"

  # Увімкнення Private Google Access
  private_ip_google_access = true

  # Увімкнення логування потоків
  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# Створення правила брандмауера для внутрішнього трафіку
resource "google_compute_firewall" "allow_internal" {
  name    = "allow-internal-${var.environment}"
  network = google_compute_network.vpc_network.id

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  source_ranges = [var.subnet_cidr]
  description   = "Дозволяє внутрішній трафік у мережі"
}

# Створення правила брандмауера для доступу до Kubernetes API
resource "google_compute_firewall" "allow_k8s_api" {
  name    = "allow-k8s-api-${var.environment}"
  network = google_compute_network.vpc_network.id

  allow {
    protocol = "tcp"
    ports    = ["443", "8443"]
  }

  source_ranges = ["0.0.0.0/0"]  # В продакшн слід обмежити до конкретних IP-адрес
  description   = "Дозволяє доступ до Kubernetes API"
}

# Створення правила брандмауера для доступу до HTTP/HTTPS
resource "google_compute_firewall" "allow_http_https" {
  name    = "allow-http-https-${var.environment}"
  network = google_compute_network.vpc_network.id

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]  # В продакшн слід обмежити до конкретних IP-адрес
  description   = "Дозволяє HTTP/HTTPS трафік"
}

# Створення Cloud NAT для доступу до інтернету з приватних інстансів
resource "google_compute_router" "router" {
  name    = "router-${var.environment}"
  region  = var.region
  network = google_compute_network.vpc_network.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat-${var.environment}"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}
