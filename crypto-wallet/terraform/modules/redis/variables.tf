variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "region" {
  description = "The region to deploy to"
  type        = string
}

variable "redis_instance_name" {
  description = "The name of the Redis instance"
  type        = string
}

variable "redis_version" {
  description = "The Redis version"
  type        = string
}

variable "redis_tier" {
  description = "The Redis tier"
  type        = string
}

variable "redis_memory_size_gb" {
  description = "The Redis memory size in GB"
  type        = number
}

variable "environment" {
  description = "The environment (e.g., dev, staging, prod)"
  type        = string
}

variable "network_id" {
  description = "The ID of the network to deploy to"
  type        = string
}
