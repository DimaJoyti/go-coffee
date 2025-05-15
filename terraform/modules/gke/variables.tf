variable "project_id" {
  description = "ID проекту GCP"
  type        = string
}

variable "region" {
  description = "Регіон GCP для розгортання ресурсів"
  type        = string
}

variable "zone" {
  description = "Зона GCP для розгортання ресурсів"
  type        = string
}

variable "gke_cluster_name" {
  description = "Назва кластера GKE"
  type        = string
}

variable "network_name" {
  description = "Назва VPC мережі"
  type        = string
}

variable "subnet_name" {
  description = "Назва підмережі"
  type        = string
}

variable "gke_node_count" {
  description = "Кількість вузлів у кластері GKE"
  type        = number
}

variable "gke_machine_type" {
  description = "Тип машини для вузлів GKE"
  type        = string
}

variable "gke_min_node_count" {
  description = "Мінімальна кількість вузлів для автоскейлінгу"
  type        = number
}

variable "gke_max_node_count" {
  description = "Максимальна кількість вузлів для автоскейлінгу"
  type        = number
}

variable "environment" {
  description = "Середовище розгортання (dev, staging, prod)"
  type        = string
}
