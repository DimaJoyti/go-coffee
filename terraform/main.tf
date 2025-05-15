# Основний файл Terraform для системи замовлення кави

# Створення мережевої інфраструктури
module "network" {
  source = "./modules/network"

  project_id   = var.project_id
  region       = var.region
  network_name = var.network_name
  subnet_name  = var.subnet_name
  subnet_cidr  = var.subnet_cidr
  environment  = var.environment
}

# Створення кластера GKE
module "gke" {
  source = "./modules/gke"

  project_id        = var.project_id
  region            = var.region
  zone              = var.zone
  gke_cluster_name  = var.gke_cluster_name
  network_name      = module.network.network_name
  subnet_name       = module.network.subnet_name
  gke_node_count    = var.gke_node_count
  gke_machine_type  = var.gke_machine_type
  gke_min_node_count = var.gke_min_node_count
  gke_max_node_count = var.gke_max_node_count
  environment       = var.environment
}

# Створення Kafka (використовуючи Helm chart)
module "kafka" {
  source = "./modules/kafka"

  depends_on = [module.gke]

  project_id            = var.project_id
  region                = var.region
  kafka_instance_name   = var.kafka_instance_name
  kafka_version         = var.kafka_version
  kafka_topic_name      = var.kafka_topic_name
  kafka_processed_topic_name = var.kafka_processed_topic_name
  environment           = var.environment
}

# Налаштування моніторингу (Prometheus + Grafana)
module "monitoring" {
  source = "./modules/monitoring"

  depends_on = [module.gke]

  project_id            = var.project_id
  region                = var.region
  enable_monitoring     = var.enable_monitoring
  grafana_admin_password = var.grafana_admin_password
  environment           = var.environment
}
