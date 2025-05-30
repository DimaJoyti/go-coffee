variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "region" {
  description = "The region to deploy to"
  type        = string
}

variable "kafka_instance_name" {
  description = "The name of the Kafka instance"
  type        = string
}

variable "kafka_version" {
  description = "The Kafka version"
  type        = string
}

variable "kafka_topic_name" {
  description = "The Kafka topic name"
  type        = string
}

variable "kafka_processed_topic_name" {
  description = "The Kafka processed topic name"
  type        = string
}

variable "environment" {
  description = "The environment (e.g., dev, staging, prod)"
  type        = string
}
