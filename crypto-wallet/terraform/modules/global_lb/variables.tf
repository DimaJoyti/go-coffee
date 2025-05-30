variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "name" {
  description = "The name of the load balancer"
  type        = string
}

variable "backend_services" {
  description = "The backend services to use"
  type        = map(object({
    group  = string
    region = string
  }))
}
