output "network_id" {
  description = "ID створеної VPC мережі"
  value       = google_compute_network.vpc_network.id
}

output "network_name" {
  description = "Назва створеної VPC мережі"
  value       = google_compute_network.vpc_network.name
}

output "subnet_id" {
  description = "ID створеної підмережі"
  value       = google_compute_subnetwork.subnet.id
}

output "subnet_name" {
  description = "Назва створеної підмережі"
  value       = google_compute_subnetwork.subnet.name
}

output "subnet_cidr" {
  description = "CIDR блок створеної підмережі"
  value       = google_compute_subnetwork.subnet.ip_cidr_range
}
