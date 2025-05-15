variable "project_id" {
  description = "ID проекту GCP"
  type        = string
}

variable "region" {
  description = "Регіон GCP для розгортання ресурсів"
  type        = string
}

variable "kafka_instance_name" {
  description = "Назва інстансу Kafka"
  type        = string
}

variable "kafka_version" {
  description = "Версія Kafka"
  type        = string
}

variable "kafka_topic_name" {
  description = "Назва топіку Kafka"
  type        = string
}

variable "kafka_processed_topic_name" {
  description = "Назва топіку для оброблених замовлень"
  type        = string
}

variable "environment" {
  description = "Середовище розгортання (dev, staging, prod)"
  type        = string
}
