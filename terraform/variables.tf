variable "project_id" {
  description = "ID проекту GCP"
  type        = string
}

variable "region" {
  description = "Регіон GCP для розгортання ресурсів"
  type        = string
  default     = "europe-west3"
}

variable "zone" {
  description = "Зона GCP для розгортання ресурсів"
  type        = string
  default     = "europe-west3-a"
}

variable "environment" {
  description = "Середовище розгортання (dev, staging, prod)"
  type        = string
  default     = "dev"
}

# Змінні для мережі
variable "network_name" {
  description = "Назва VPC мережі"
  type        = string
  default     = "coffee-network"
}

variable "subnet_name" {
  description = "Назва підмережі"
  type        = string
  default     = "coffee-subnet"
}

variable "subnet_cidr" {
  description = "CIDR блок для підмережі"
  type        = string
  default     = "10.0.0.0/24"
}

# Змінні для GKE
variable "gke_cluster_name" {
  description = "Назва кластера GKE"
  type        = string
  default     = "coffee-cluster"
}

variable "gke_node_count" {
  description = "Кількість вузлів у кластері GKE"
  type        = number
  default     = 3
}

variable "gke_machine_type" {
  description = "Тип машини для вузлів GKE"
  type        = string
  default     = "e2-standard-2"
}

variable "gke_min_node_count" {
  description = "Мінімальна кількість вузлів для автоскейлінгу"
  type        = number
  default     = 1
}

variable "gke_max_node_count" {
  description = "Максимальна кількість вузлів для автоскейлінгу"
  type        = number
  default     = 5
}

# Змінні для Kafka
variable "kafka_instance_name" {
  description = "Назва інстансу Kafka"
  type        = string
  default     = "coffee-kafka"
}

variable "kafka_version" {
  description = "Версія Kafka"
  type        = string
  default     = "3.4"
}

variable "kafka_topic_name" {
  description = "Назва топіку Kafka"
  type        = string
  default     = "coffee_orders"
}

variable "kafka_processed_topic_name" {
  description = "Назва топіку для оброблених замовлень"
  type        = string
  default     = "processed_orders"
}

# Змінні для моніторингу
variable "enable_monitoring" {
  description = "Увімкнути моніторинг"
  type        = bool
  default     = true
}

variable "grafana_admin_password" {
  description = "Пароль адміністратора Grafana"
  type        = string
  sensitive   = true
  default     = "admin"  # Змінити в terraform.tfvars
}
