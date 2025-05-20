variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "region" {
  description = "The region to deploy to"
  type        = string
}

variable "network_name" {
  description = "The name of the network to deploy to"
  type        = string
}

variable "subnet_name" {
  description = "The name of the subnet to deploy to"
  type        = string
}

variable "gke_cluster_name" {
  description = "The name of the GKE cluster"
  type        = string
}

variable "environment" {
  description = "The environment (e.g., dev, staging, prod)"
  type        = string
}

variable "node_count" {
  description = "The number of nodes in the GKE cluster"
  type        = number
}

variable "node_machine_type" {
  description = "The machine type for the GKE nodes"
  type        = string
}

variable "node_disk_size_gb" {
  description = "The disk size for the GKE nodes in GB"
  type        = number
}

variable "node_disk_type" {
  description = "The disk type for the GKE nodes"
  type        = string
}

variable "node_preemptible" {
  description = "Whether the GKE nodes are preemptible"
  type        = bool
}
