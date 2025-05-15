# Configuration Guide

This guide describes how to configure the Coffee Order System.

## Configuration Methods

The Coffee Order System can be configured using the following methods:

1. **Environment Variables**: Set environment variables to configure the system.
2. **Configuration Files**: Use JSON configuration files to configure the system.
3. **Infrastructure as Code (Terraform)**: Use Terraform to configure and deploy the infrastructure.

Environment variables take precedence over configuration files.

## Producer Service Configuration

### Environment Variables

The Producer service can be configured using the following environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `SERVER_PORT` | HTTP server port | `3000` |
| `KAFKA_BROKERS` | Kafka broker addresses (JSON array) | `["localhost:9092"]` |
| `KAFKA_TOPIC` | Kafka topic for orders | `"coffee_orders"` |
| `KAFKA_RETRY_MAX` | Maximum number of retries for Kafka producer | `5` |
| `KAFKA_REQUIRED_ACKS` | Required acknowledgments for Kafka producer | `"all"` |
| `CONFIG_FILE` | Path to configuration file | `""` |

### Configuration File

The Producer service can also be configured using a JSON configuration file. Here's an example:

```json
{
  "server": {
    "port": 3000
  },
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "coffee_orders",
    "retry_max": 5,
    "required_acks": "all"
  }
}
```

To use a configuration file, set the `CONFIG_FILE` environment variable to the path of the file:

```bash
export CONFIG_FILE=config.json
```

## Consumer Service Configuration

### Environment Variables

The Consumer service can be configured using the following environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `KAFKA_BROKERS` | Kafka broker addresses (JSON array) | `["localhost:9092"]` |
| `KAFKA_TOPIC` | Kafka topic for orders | `"coffee_orders"` |
| `CONFIG_FILE` | Path to configuration file | `""` |

### Configuration File

The Consumer service can also be configured using a JSON configuration file. Here's an example:

```json
{
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "coffee_orders"
  }
}
```

To use a configuration file, set the `CONFIG_FILE` environment variable to the path of the file:

```bash
export CONFIG_FILE=config.json
```

## Kafka Configuration

### Required Acknowledgments

The `KAFKA_REQUIRED_ACKS` environment variable (or `kafka.required_acks` in the configuration file) can have the following values:

- `"none"`: No acknowledgment is required.
- `"local"`: Only the leader must acknowledge the message.
- `"all"`: All replicas must acknowledge the message.

## Example: Running with Custom Configuration

### Using Environment Variables

```bash
# Producer
export SERVER_PORT=8080
export KAFKA_BROKERS='["kafka1:9092","kafka2:9092"]'
export KAFKA_TOPIC="custom_orders"
cd producer
go run main.go

# Consumer
export KAFKA_BROKERS='["kafka1:9092","kafka2:9092"]'
export KAFKA_TOPIC="custom_orders"
cd consumer
go run main.go
```

### Using Configuration Files

```bash
# Producer
export CONFIG_FILE=custom-config.json
cd producer
go run main.go

# Consumer
export CONFIG_FILE=custom-config.json
cd consumer
go run main.go
```

## Infrastructure as Code (Terraform)

The Coffee Order System can be deployed to Google Cloud Platform (GCP) using Terraform. Terraform is used to automate the creation and management of the infrastructure.

### Terraform Configuration

The Terraform configuration for the Coffee Order System includes:

1. **Network Infrastructure**: VPC, subnets, firewall rules, NAT
2. **GKE Cluster**: Kubernetes cluster for deploying applications
3. **Kafka**: Deployment of Kafka using Helm
4. **Monitoring**: Deployment of Prometheus and Grafana using Helm

### Terraform Variables

The Terraform configuration can be customized using the following variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `project_id` | GCP project ID | - |
| `region` | GCP region | `europe-west3` |
| `zone` | GCP zone | `europe-west3-a` |
| `environment` | Deployment environment | `dev` |
| `network_name` | VPC network name | `coffee-network` |
| `subnet_name` | Subnet name | `coffee-subnet` |
| `subnet_cidr` | Subnet CIDR block | `10.0.0.0/24` |
| `gke_cluster_name` | GKE cluster name | `coffee-cluster` |
| `gke_node_count` | Number of nodes in GKE cluster | `3` |
| `gke_machine_type` | Machine type for GKE nodes | `e2-standard-2` |
| `kafka_instance_name` | Kafka instance name | `coffee-kafka` |
| `kafka_topic_name` | Kafka topic name | `coffee_orders` |
| `kafka_processed_topic_name` | Kafka processed topic name | `processed_orders` |

For more information about Terraform configuration, see [Terraform](terraform.md).

## Next Steps

- [API Reference](api-reference.md): Explore the API endpoints.
- [Kafka Integration](kafka-integration.md): Learn about Kafka integration.
- [Troubleshooting](troubleshooting.md): Troubleshoot configuration issues.
- [Terraform](terraform.md): Learn about Terraform configuration.
