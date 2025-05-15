# Terraform for Coffee Order System

This document describes how to use Terraform to deploy the infrastructure for the Coffee Order System in Google Cloud Platform (GCP).

## Overview

Terraform is used to automate the creation and management of infrastructure in GCP. The Terraform configuration for the Coffee Order System includes:

1. **Network Infrastructure**: VPC, subnets, firewall rules, NAT
2. **GKE Cluster**: Kubernetes cluster for deploying applications
3. **Kafka**: Deployment of Kafka using Helm
4. **Monitoring**: Deployment of Prometheus and Grafana using Helm

## Directory Structure

```
terraform/
├── main.tf                  # Main configuration file
├── variables.tf             # Variable declarations
├── outputs.tf               # Output declarations
├── provider.tf              # GCP provider configuration
├── terraform.tfvars.example # Example variable values
├── modules/
│   ├── network/             # Module for network infrastructure
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── gke/                 # Module for GKE
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── kafka/               # Module for Kafka
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── monitoring/          # Module for monitoring
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
```

## Prerequisites

Before using Terraform to deploy the infrastructure, make sure you have:

1. **Terraform Installed**: Version 1.0.0 or higher
2. **GCP Account**: With permissions to create resources
3. **gcloud CLI Configured**: For authentication to GCP
4. **GCP Project Created**: For deploying resources

## Setup

1. Clone the repository:

```bash
git clone https://github.com/your-username/coffee-order-system.git
cd coffee-order-system
```

2. Create a `terraform.tfvars` file based on `terraform.tfvars.example`:

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

3. Edit `terraform.tfvars` according to your needs:

```hcl
# Basic GCP settings
project_id = "your-gcp-project-id"
region     = "europe-west3"
zone       = "europe-west3-a"
environment = "dev"

# Network settings
network_name = "coffee-network"
subnet_name  = "coffee-subnet"
subnet_cidr  = "10.0.0.0/24"

# GKE settings
gke_cluster_name = "coffee-cluster"
gke_node_count   = 3
gke_machine_type = "e2-standard-2"
gke_min_node_count = 1
gke_max_node_count = 5

# Kafka settings
kafka_instance_name = "coffee-kafka"
kafka_version       = "3.4"
kafka_topic_name    = "coffee_orders"
kafka_processed_topic_name = "processed_orders"

# Monitoring settings
enable_monitoring     = true
grafana_admin_password = "change-me-in-production"
```

## Usage

### Initialize Terraform

```bash
terraform init
```

### Check Deployment Plan

```bash
terraform plan
```

### Deploy Infrastructure

```bash
terraform apply
```

### Destroy Infrastructure

```bash
terraform destroy
```

## Modules

### Network Module

The `network` module creates the network infrastructure for the Coffee Order System:

- VPC network
- Subnet
- Firewall rules for internal traffic
- Firewall rules for Kubernetes API access
- Firewall rules for HTTP/HTTPS access
- Cloud NAT for internet access from private instances

### GKE Module

The `gke` module creates a GKE cluster for deploying applications:

- GKE cluster with private nodes
- Node pool with autoscaling
- Workload Identity configuration
- Logging and monitoring configuration
- Network policy configuration
- Automatic node upgrade configuration

### Kafka Module

The `kafka` module deploys Kafka using Helm:

- Namespace for Kafka
- Kafka deployment using Helm
- Creation of topics for orders and processed orders

### Monitoring Module

The `monitoring` module deploys Prometheus and Grafana using Helm:

- Namespace for monitoring
- Prometheus deployment using Helm
- Grafana deployment using Helm
- Datasource configuration for Prometheus in Grafana

## Output Values

After successful infrastructure deployment, Terraform will output the following values:

- **network_name**: Name of the created VPC network
- **subnet_name**: Name of the created subnet
- **subnet_cidr**: CIDR block of the created subnet
- **gke_cluster_name**: Name of the created GKE cluster
- **gke_endpoint**: Endpoint of the GKE cluster
- **gke_kubeconfig**: Command to get kubeconfig
- **kafka_bootstrap_servers**: Kafka bootstrap servers
- **kafka_topics**: Created Kafka topics
- **grafana_url**: URL to access Grafana
- **prometheus_url**: URL to access Prometheus

## CI/CD Integration

To integrate Terraform with CI/CD, add the following steps to your workflow:

```yaml
- name: Setup Terraform
  uses: hashicorp/setup-terraform@v1
  with:
    terraform_version: 1.0.0

- name: Terraform Init
  run: terraform init
  working-directory: ./terraform

- name: Terraform Validate
  run: terraform validate
  working-directory: ./terraform

- name: Terraform Plan
  run: terraform plan -var-file=terraform.tfvars
  working-directory: ./terraform

- name: Terraform Apply
  if: github.ref == 'refs/heads/main'
  run: terraform apply -auto-approve -var-file=terraform.tfvars
  working-directory: ./terraform
```

## Best Practices

1. **Use a Backend for Terraform State**: Configure a backend for Terraform state to avoid concurrency issues and ensure state security.
2. **Use Variables for Sensitive Data**: Don't store sensitive data (passwords, keys) in code. Use variables and store them in a secure location.
3. **Use Modules**: Split the configuration into modules for better organization and reuse.
4. **Use Tags and Labels**: Add tags and labels to resources for better organization and cost tracking.
5. **Use Versioning**: Specify provider and module versions to ensure stability.

## Troubleshooting

### Error: "Error creating Network"

**Possible Causes**:
1. Insufficient permissions to create a network
2. A network with the same name already exists

**Solutions**:
1. Make sure you have permissions to create a network
2. Change the network name or import the existing network

### Error: "Error creating GKE cluster"

**Possible Causes**:
1. Insufficient permissions to create a GKE cluster
2. Insufficient quotas to create a GKE cluster
3. Incorrect network settings

**Solutions**:
1. Make sure you have permissions to create a GKE cluster
2. Increase quotas for your project
3. Check network settings

## Next Steps

- [Installation](installation.md): Install and run the system
- [Configuration](configuration.md): Configure the system
- [Docker and Kubernetes](docker-kubernetes.md): Use Docker and Kubernetes
