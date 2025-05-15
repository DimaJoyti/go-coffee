variable "project_id" {
  description = "ID проекту GCP"
  type        = string
}

variable "region" {
  description = "Регіон GCP для розгортання ресурсів"
  type        = string
}

variable "enable_monitoring" {
  description = "Увімкнути моніторинг"
  type        = bool
  default     = true
}

variable "grafana_admin_password" {
  description = "Пароль адміністратора Grafana"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Середовище розгортання (dev, staging, prod)"
  type        = string
}
