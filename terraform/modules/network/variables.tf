variable "project_id" {
  description = "ID проекту GCP"
  type        = string
}

variable "region" {
  description = "Регіон GCP для розгортання ресурсів"
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

variable "subnet_cidr" {
  description = "CIDR блок для підмережі"
  type        = string
}

variable "environment" {
  description = "Середовище розгортання (dev, staging, prod)"
  type        = string
}
