# PostgreSQL Module Outputs

output "instance_name" {
  description = "The name of the PostgreSQL instance"
  value       = google_sql_database_instance.main.name
}

output "instance_connection_name" {
  description = "The connection name of the PostgreSQL instance"
  value       = google_sql_database_instance.main.connection_name
}

output "private_ip_address" {
  description = "The private IP address of the PostgreSQL instance"
  value       = google_sql_database_instance.main.private_ip_address
}

output "public_ip_address" {
  description = "The public IP address of the PostgreSQL instance"
  value       = google_sql_database_instance.main.public_ip_address
}

output "self_link" {
  description = "The URI of the PostgreSQL instance"
  value       = google_sql_database_instance.main.self_link
}

output "server_ca_cert" {
  description = "The CA certificate information used to connect to the SQL instance via SSL"
  value       = google_sql_database_instance.main.server_ca_cert
  sensitive   = true
}

output "service_account_email_address" {
  description = "The service account email address assigned to the instance"
  value       = google_sql_database_instance.main.service_account_email_address
}

output "database_version" {
  description = "The PostgreSQL version"
  value       = google_sql_database_instance.main.database_version
}

output "first_ip_address" {
  description = "The first IP address of the PostgreSQL instance"
  value       = google_sql_database_instance.main.first_ip_address
}

output "ip_address" {
  description = "The IP addresses assigned to the instance"
  value       = google_sql_database_instance.main.ip_address
}

output "database_names" {
  description = "List of created database names"
  value       = [for db in google_sql_database.databases : db.name]
}

output "database_username" {
  description = "Database username"
  value       = google_sql_user.main.name
}

output "database_password_secret_id" {
  description = "Secret Manager secret ID for database password"
  value       = google_secret_manager_secret.db_password.secret_id
}

output "database_password_secret_version" {
  description = "Secret Manager secret version for database password"
  value       = google_secret_manager_secret_version.db_password.name
  sensitive   = true
}

output "read_replica_connection_name" {
  description = "The connection name of the read replica instance"
  value       = var.environment == "prod" && var.create_read_replica ? google_sql_database_instance.read_replica[0].connection_name : null
}

output "read_replica_private_ip_address" {
  description = "The private IP address of the read replica instance"
  value       = var.environment == "prod" && var.create_read_replica ? google_sql_database_instance.read_replica[0].private_ip_address : null
}

output "connection_string" {
  description = "PostgreSQL connection string"
  value       = "postgresql://${google_sql_user.main.name}:${random_password.db_password.result}@${google_sql_database_instance.main.private_ip_address}:5432/${google_sql_database.databases["go_coffee_main"].name}"
  sensitive   = true
}

output "jdbc_connection_string" {
  description = "JDBC connection string for PostgreSQL"
  value       = "jdbc:postgresql://${google_sql_database_instance.main.private_ip_address}:5432/${google_sql_database.databases["go_coffee_main"].name}"
}

output "instance_settings" {
  description = "The settings of the PostgreSQL instance"
  value = {
    tier              = google_sql_database_instance.main.settings[0].tier
    availability_type = google_sql_database_instance.main.settings[0].availability_type
    disk_type         = google_sql_database_instance.main.settings[0].disk_type
    disk_size         = google_sql_database_instance.main.settings[0].disk_size
  }
}

output "backup_configuration" {
  description = "The backup configuration of the PostgreSQL instance"
  value = {
    enabled                        = google_sql_database_instance.main.settings[0].backup_configuration[0].enabled
    start_time                     = google_sql_database_instance.main.settings[0].backup_configuration[0].start_time
    location                       = google_sql_database_instance.main.settings[0].backup_configuration[0].location
    point_in_time_recovery_enabled = google_sql_database_instance.main.settings[0].backup_configuration[0].point_in_time_recovery_enabled
  }
}
