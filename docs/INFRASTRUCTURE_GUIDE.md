# ☕ Go Coffee - Complete Infrastructure Guide

## 🚀 Overview

This guide provides comprehensive instructions for deploying and managing the Go Coffee platform infrastructure using Terraform and Kubernetes. The platform supports multi-cloud deployment with enterprise-grade security, monitoring, and scalability.

## 🏗️ Architecture Overview

### Core Components

- **Multi-Cloud Infrastructure**: GCP (primary), AWS, Azure support
- **Container Orchestration**: Kubernetes with GKE/EKS/AKS
- **Service Mesh**: Istio for traffic management and security
- **Monitoring**: Prometheus, Grafana, Jaeger for observability
- **Databases**: PostgreSQL (Cloud SQL), Redis (Memorystore)
- **Message Queue**: Apache Kafka for event streaming
- **CI/CD**: GitHub Actions with GitOps (ArgoCD)
- **Security**: Workload Identity, Network Policies, mTLS

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer / Ingress                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Istio Gateway                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 API Gateway                                 │
│              (go-coffee-api-gateway)                        │
└─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┘
      │     │     │     │     │     │     │     │     │
      ▼     ▼     ▼     ▼     ▼     ▼     ▼     ▼     ▼
   Order  Pay  Kitchen User  Sec  WebUI  AI   Bright Comm
  Service ment Service Gate  Gate Backend Search Data  Hub
          Svc         way   way          Agent  Hub
```

## 📋 Prerequisites

### Required Tools

- **Terraform** >= 1.5.0
- **kubectl** >= 1.28.0
- **Helm** >= 3.10.0
- **Docker** >= 20.10.0
- **gcloud CLI** (for GCP)
- **aws CLI** (for AWS)
- **az CLI** (for Azure)

### Cloud Provider Setup

#### Google Cloud Platform (GCP)

```bash
# Install gcloud CLI
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Authenticate
gcloud auth login
gcloud auth application-default login

# Set project
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable container.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable redis.googleapis.com
```

#### Amazon Web Services (AWS)

```bash
# Install AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Configure credentials
aws configure
```

#### Microsoft Azure

```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Login
az login
```

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

### 2. Configure Environment

```bash
# Copy example configuration
cp terraform/terraform.tfvars.example terraform/terraform.tfvars

# Edit configuration
vim terraform/terraform.tfvars
```

### 3. Deploy Infrastructure

```bash
# Make deployment script executable
chmod +x scripts/deploy-complete-infrastructure.sh

# Deploy to development environment
export ENVIRONMENT=dev
export CLOUD_PROVIDER=gcp
export PROJECT_ID=your-project-id
export REGION=europe-west3

./scripts/deploy-complete-infrastructure.sh deploy
```

## 📁 Directory Structure

```
go-coffee/
├── terraform/                 # Infrastructure as Code
│   ├── main.tf                # Main Terraform configuration
│   ├── variables.tf           # Variable definitions
│   ├── outputs.tf             # Output values
│   ├── modules/               # Terraform modules
│   │   ├── network/           # VPC and networking
│   │   ├── gke/               # Google Kubernetes Engine
│   │   ├── postgresql/        # Cloud SQL PostgreSQL
│   │   ├── redis/             # Redis/Memorystore
│   │   ├── kafka/             # Apache Kafka
│   │   ├── monitoring/        # Monitoring stack
│   │   ├── service-mesh/      # Istio service mesh
│   │   ├── security/          # Security policies
│   │   └── aws-infrastructure/# AWS infrastructure
├── helm/                      # Helm charts
│   └── go-coffee-platform/    # Main application chart
│       ├── Chart.yaml         # Chart metadata
│       ├── values.yaml        # Default values
│       ├── values-dev.yaml    # Development values
│       ├── values-staging.yaml# Staging values
│       ├── values-prod.yaml   # Production values
│       └── templates/         # Kubernetes manifests
├── k8s/                       # Kubernetes manifests
│   ├── base/                  # Base configurations
│   └── overlays/              # Environment-specific overlays
├── gitops/                    # GitOps configurations
│   └── argocd/                # ArgoCD applications
├── scripts/                   # Deployment scripts
├── monitoring/                # Monitoring configurations
└── docs/                      # Documentation
```

## ⚙️ Configuration

### Terraform Variables

Key variables in `terraform/terraform.tfvars`:

```hcl
# Project Configuration
project_id = "go-coffee-prod"
region     = "europe-west3"
environment = "prod"

# Cluster Configuration
gke_cluster_name = "go-coffee-cluster"
gke_node_count   = 3
gke_machine_type = "e2-standard-4"

# Database Configuration
postgres_instance_name = "go-coffee-postgres"
postgres_tier         = "db-custom-2-8192"

# Redis Configuration
redis_instance_name = "go-coffee-redis"
redis_memory_size_gb = 4

# Monitoring
enable_monitoring = true
enable_service_mesh = true

# Security
enable_workload_identity = true
enable_network_policy = true
```

### Helm Values

Key configurations in `helm/go-coffee-platform/values.yaml`:

```yaml
global:
  environment: prod
  imageRegistry: ghcr.io/dimajoyti/go-coffee
  domain: gocoffee.dev
  
services:
  apiGateway:
    enabled: true
    replicaCount: 3
    image:
      repository: ghcr.io/dimajoyti/go-coffee/api-gateway
      tag: "1.0.0"
    
monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
    
serviceMesh:
  istio:
    enabled: true
    mtls:
      enabled: true
```

## 🔧 Deployment Options

### Environment-Specific Deployments

#### Development
```bash
export ENVIRONMENT=dev
export ENABLE_MONITORING=false
export ENABLE_SERVICE_MESH=false
./scripts/deploy-complete-infrastructure.sh deploy
```

#### Staging
```bash
export ENVIRONMENT=staging
export ENABLE_MONITORING=true
export ENABLE_SERVICE_MESH=true
./scripts/deploy-complete-infrastructure.sh deploy
```

#### Production
```bash
export ENVIRONMENT=prod
export ENABLE_MONITORING=true
export ENABLE_SERVICE_MESH=true
export ENABLE_GITOPS=true
./scripts/deploy-complete-infrastructure.sh deploy
```

### Multi-Cloud Deployment

#### Deploy to AWS
```bash
export CLOUD_PROVIDER=aws
export AWS_REGION=us-west-2
./scripts/deploy-complete-infrastructure.sh deploy
```

#### Deploy to Azure
```bash
export CLOUD_PROVIDER=azure
export AZURE_REGION=westeurope
./scripts/deploy-complete-infrastructure.sh deploy
```

## 📊 Monitoring and Observability

### Prometheus Metrics

The platform exposes comprehensive metrics:

- **Application Metrics**: Request rates, latency, errors
- **Business Metrics**: Orders, revenue, customer satisfaction
- **Infrastructure Metrics**: CPU, memory, disk, network
- **Custom Metrics**: AI agent performance, Web3 transactions

### Grafana Dashboards

Pre-configured dashboards include:

- **Go Coffee Overview**: High-level business metrics
- **Service Performance**: Individual service metrics
- **Infrastructure Health**: Cluster and node metrics
- **AI Agents**: AI service performance
- **Web3 Trading**: DeFi and trading metrics

### Distributed Tracing

Jaeger provides end-to-end request tracing across:

- HTTP requests through the API gateway
- gRPC calls between services
- Database queries
- External API calls
- AI model inference
- Blockchain transactions

## 🔒 Security

### Network Security

- **Network Policies**: Restrict pod-to-pod communication
- **Private Clusters**: Nodes in private subnets
- **Firewall Rules**: Controlled ingress/egress
- **VPC Peering**: Secure multi-region connectivity

### Identity and Access

- **Workload Identity**: Secure GCP service access
- **RBAC**: Role-based access control
- **Service Accounts**: Least privilege principle
- **Pod Security Standards**: Restricted security contexts

### Data Protection

- **Encryption at Rest**: Database and storage encryption
- **Encryption in Transit**: TLS/mTLS everywhere
- **Secret Management**: Kubernetes secrets + external secret stores
- **Backup and Recovery**: Automated database backups

## 🔄 CI/CD and GitOps

### GitHub Actions

Automated workflows for:

- **Infrastructure Deployment**: Terraform apply/destroy
- **Application Deployment**: Helm chart updates
- **Security Scanning**: Vulnerability and compliance checks
- **Testing**: Integration and end-to-end tests

### ArgoCD GitOps

- **Declarative Configuration**: Git as source of truth
- **Automated Sync**: Continuous deployment
- **Rollback Capability**: Easy rollback to previous versions
- **Multi-Environment**: Separate apps per environment

## 🛠️ Maintenance

### Scaling

#### Horizontal Pod Autoscaling
```bash
kubectl get hpa -n go-coffee
```

#### Cluster Autoscaling
```bash
kubectl get nodes
kubectl describe node NODE_NAME
```

### Updates

#### Rolling Updates
```bash
helm upgrade go-coffee-platform ./helm/go-coffee-platform \
  --namespace go-coffee \
  --values values-prod.yaml
```

#### Infrastructure Updates
```bash
cd terraform
terraform plan
terraform apply
```

### Backup and Recovery

#### Database Backups
```bash
# Manual backup
gcloud sql backups create --instance=go-coffee-postgres

# Restore from backup
gcloud sql backups restore BACKUP_ID --restore-instance=go-coffee-postgres
```

## 🚨 Troubleshooting

### Common Issues

#### Pod Startup Issues
```bash
kubectl describe pod POD_NAME -n go-coffee
kubectl logs POD_NAME -n go-coffee --previous
```

#### Service Connectivity
```bash
kubectl get svc -n go-coffee
kubectl get endpoints -n go-coffee
```

#### Istio Issues
```bash
istioctl proxy-status
istioctl proxy-config cluster POD_NAME
```

### Health Checks

```bash
# Check cluster health
kubectl cluster-info

# Check service health
curl http://API_GATEWAY_URL/health

# Check monitoring
kubectl port-forward svc/prometheus-server 9090:80 -n go-coffee-monitoring
```

## 📚 Additional Resources

- [Go Coffee Documentation](https://docs.gocoffee.dev)
- [Terraform Documentation](https://terraform.io/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs)
- [Helm Documentation](https://helm.sh/docs)
- [Istio Documentation](https://istio.io/docs)
- [Prometheus Documentation](https://prometheus.io/docs)

## 🤝 Contributing

Please read [CONTRIBUTING.md](../CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
