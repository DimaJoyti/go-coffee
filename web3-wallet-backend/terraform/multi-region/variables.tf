variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "primary_region" {
  description = "The primary region for the project"
  type        = string
  default     = "us-central1"
}

variable "regions" {
  description = "The regions to deploy to"
  type        = list(string)
  default     = ["us-central1", "europe-west1", "asia-east1"]
}

variable "subnet_cidr_ranges" {
  description = "The CIDR ranges for the subnets in each region"
  type        = map(string)
  default     = {
    "us-central1" = "10.0.0.0/20"
    "europe-west1" = "10.0.16.0/20"
    "asia-east1" = "10.0.32.0/20"
  }
}

variable "environment" {
  description = "The environment (e.g., dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "node_count" {
  description = "The number of nodes in each GKE cluster"
  type        = number
  default     = 3
}

variable "node_machine_type" {
  description = "The machine type for the GKE nodes"
  type        = string
  default     = "e2-standard-2"
}

variable "node_disk_size_gb" {
  description = "The disk size for the GKE nodes in GB"
  type        = number
  default     = 100
}

variable "node_disk_type" {
  description = "The disk type for the GKE nodes"
  type        = string
  default     = "pd-standard"
}

variable "node_preemptible" {
  description = "Whether the GKE nodes are preemptible"
  type        = bool
  default     = false
}

variable "redis_version" {
  description = "The Redis version"
  type        = string
  default     = "REDIS_6_X"
}

variable "redis_tier" {
  description = "The Redis tier"
  type        = string
  default     = "STANDARD_HA"
}

variable "redis_memory_size_gb" {
  description = "The Redis memory size in GB"
  type        = number
  default     = 5
}

variable "kafka_version" {
  description = "The Kafka version"
  type        = string
  default     = "3.3.1"
}

variable "kafka_topic_name" {
  description = "The Kafka topic name"
  type        = string
  default     = "supply-events"
}

variable "kafka_processed_topic_name" {
  description = "The Kafka processed topic name"
  type        = string
  default     = "processed-events"
}
