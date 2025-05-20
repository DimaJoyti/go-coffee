provider "google" {
  project = var.project_id
  region  = var.primary_region
}

# Create VPC network
resource "google_compute_network" "vpc" {
  name                    = "${var.project_id}-vpc"
  auto_create_subnetworks = false
}

# Create subnets in each region
resource "google_compute_subnetwork" "subnets" {
  for_each      = toset(var.regions)
  name          = "${var.project_id}-subnet-${each.value}"
  ip_cidr_range = var.subnet_cidr_ranges[each.value]
  region        = each.value
  network       = google_compute_network.vpc.id
}

# Create GKE clusters in each region
module "gke_clusters" {
  source   = "../modules/gke"
  for_each = toset(var.regions)

  project_id            = var.project_id
  region                = each.value
  network_name          = google_compute_network.vpc.name
  subnet_name           = google_compute_subnetwork.subnets[each.value].name
  gke_cluster_name      = "${var.project_id}-cluster"
  environment           = var.environment
  node_count            = var.node_count
  node_machine_type     = var.node_machine_type
  node_disk_size_gb     = var.node_disk_size_gb
  node_disk_type        = var.node_disk_type
  node_preemptible      = var.node_preemptible
}

# Create Redis clusters in each region
module "redis_clusters" {
  source   = "../modules/redis"
  for_each = toset(var.regions)

  project_id         = var.project_id
  region             = each.value
  redis_instance_name = "${var.project_id}-redis"
  redis_version      = var.redis_version
  redis_tier         = var.redis_tier
  redis_memory_size_gb = var.redis_memory_size_gb
  environment        = var.environment
  network_id         = google_compute_network.vpc.id
}

# Create Kafka clusters in each region
module "kafka_clusters" {
  source   = "../modules/kafka"
  for_each = toset(var.regions)

  project_id            = var.project_id
  region                = each.value
  kafka_instance_name   = "${var.project_id}-kafka"
  kafka_version         = var.kafka_version
  kafka_topic_name      = var.kafka_topic_name
  kafka_processed_topic_name = var.kafka_processed_topic_name
  environment           = var.environment
  depends_on            = [module.gke_clusters]
}

# Create global load balancer
module "global_lb" {
  source = "../modules/global_lb"

  project_id      = var.project_id
  name            = "${var.project_id}-lb"
  backend_services = {
    for region in var.regions : region => {
      group = module.gke_clusters[region].instance_group
      region = region
    }
  }
}

# Output the load balancer IP
output "load_balancer_ip" {
  value = module.global_lb.ip_address
}

# Output Redis hosts
output "redis_hosts" {
  value = {
    for region in var.regions : region => module.redis_clusters[region].redis_host
  }
}

# Output Redis ports
output "redis_ports" {
  value = {
    for region in var.regions : region => module.redis_clusters[region].redis_port
  }
}

# Output Kafka brokers
output "kafka_brokers" {
  value = {
    for region in var.regions : region => module.kafka_clusters[region].kafka_brokers
  }
}
