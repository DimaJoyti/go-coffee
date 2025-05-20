# Web3 Wallet Backend Deployment Guide

This guide provides detailed instructions for deploying the Web3 Wallet Backend system in various environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Development Environment](#development-environment)
3. [Testing Environment](#testing-environment)
4. [Production Environment](#production-environment)
5. [Docker Deployment](#docker-deployment)
6. [Kubernetes Deployment](#kubernetes-deployment)
7. [Cloud Provider Deployment](#cloud-provider-deployment)
8. [Configuration](#configuration)
9. [Monitoring](#monitoring)
10. [Backup and Recovery](#backup-and-recovery)
11. [Troubleshooting](#troubleshooting)

## Prerequisites

Before deploying the Web3 Wallet Backend system, ensure that you have the following prerequisites:

- Go 1.22 or higher
- PostgreSQL 16 or higher
- Redis 7 or higher
- Docker and Docker Compose (for containerized deployment)
- Kubernetes (for Kubernetes deployment)
- Access to blockchain nodes (Ethereum, Binance Smart Chain, Polygon)

## Development Environment

### Local Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/web3-wallet-backend.git
cd web3-wallet-backend
```

2. Install Go dependencies:

```bash
go mod tidy
```

3. Generate gRPC code:

```bash
cd api
./generate.sh  # or generate.bat on Windows
cd ..
```

4. Set up the database:

```bash
# Create the database
createdb web3_wallet

# Run migrations
go run db/migrate.go -up -config config/config.yaml
```

5. Start the services:

```bash
# Using the provided script
./run.sh  # or run.bat on Windows

# Or start each service individually
go run cmd/api-gateway/main.go
go run cmd/wallet-service/main.go
go run cmd/transaction-service/main.go
go run cmd/smart-contract-service/main.go
go run cmd/security-service/main.go
```

### Docker Compose Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/web3-wallet-backend.git
cd web3-wallet-backend
```

2. Start the services using Docker Compose:

```bash
docker-compose up -d
```

3. Run database migrations:

```bash
docker-compose exec api-gateway go run db/migrate.go -up -config config/config.yaml
```

## Testing Environment

The testing environment is similar to the development environment but with additional monitoring and testing tools.

### Docker Compose Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/web3-wallet-backend.git
cd web3-wallet-backend
```

2. Start the services using Docker Compose:

```bash
docker-compose -f docker-compose.yml -f docker-compose.testing.yml up -d
```

3. Run database migrations:

```bash
docker-compose exec api-gateway go run db/migrate.go -up -config config/config.yaml
```

4. Run tests:

```bash
docker-compose exec api-gateway go test ./...
```

## Production Environment

The production environment requires additional security measures and high availability configurations.

### Prerequisites

- Production-grade PostgreSQL database
- Production-grade Redis cluster
- Load balancer
- SSL certificates
- Monitoring and alerting system
- Backup system

### Deployment Steps

1. Clone the repository:

```bash
git clone https://github.com/yourusername/web3-wallet-backend.git
cd web3-wallet-backend
```

2. Configure the system for production:

```bash
# Create production configuration
cp config/config.yaml config/config.production.yaml
# Edit the production configuration
vi config/config.production.yaml
```

3. Build the Docker images:

```bash
docker-compose build
```

4. Push the Docker images to a container registry:

```bash
docker-compose push
```

5. Deploy the system using Docker Compose or Kubernetes (see below).

## Docker Deployment

### Docker Compose

1. Create a Docker Compose file for production:

```bash
cp docker-compose.yml docker-compose.production.yml
# Edit the production Docker Compose file
vi docker-compose.production.yml
```

2. Deploy the system:

```bash
docker-compose -f docker-compose.production.yml up -d
```

3. Run database migrations:

```bash
docker-compose -f docker-compose.production.yml exec api-gateway go run db/migrate.go -up -config config/config.production.yaml
```

### Docker Swarm

1. Initialize a Docker Swarm:

```bash
docker swarm init
```

2. Deploy the system:

```bash
docker stack deploy -c docker-compose.production.yml web3-wallet
```

3. Run database migrations:

```bash
docker exec $(docker ps -q -f name=web3-wallet_api-gateway) go run db/migrate.go -up -config config/config.production.yaml
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster
- kubectl configured to access the cluster
- Helm (optional)

### Deployment Steps

1. Create a namespace:

```bash
kubectl apply -f kubernetes/manifests/01-namespace.yaml
```

2. Create ConfigMap and Secret:

```bash
# Edit the ConfigMap and Secret with your production values
vi kubernetes/manifests/02-configmap.yaml
vi kubernetes/manifests/03-secret.yaml

# Apply the ConfigMap and Secret
kubectl apply -f kubernetes/manifests/02-configmap.yaml
kubectl apply -f kubernetes/manifests/03-secret.yaml
```

3. Deploy the database:

```bash
kubectl apply -f kubernetes/manifests/04-postgres.yaml
```

4. Deploy Redis:

```bash
kubectl apply -f kubernetes/manifests/05-redis.yaml
```

5. Deploy the services:

```bash
kubectl apply -f kubernetes/manifests/06-api-gateway.yaml
kubectl apply -f kubernetes/manifests/07-wallet-service.yaml
kubectl apply -f kubernetes/manifests/08-transaction-service.yaml
kubectl apply -f kubernetes/manifests/09-smart-contract-service.yaml
kubectl apply -f kubernetes/manifests/10-security-service.yaml
```

6. Run database migrations:

```bash
kubectl exec -it $(kubectl get pods -n web3-wallet -l app=api-gateway -o jsonpath='{.items[0].metadata.name}') -n web3-wallet -- go run db/migrate.go -up -config config/config.yaml
```

### Helm Deployment (Optional)

1. Install the Helm chart:

```bash
helm install web3-wallet ./kubernetes/helm/web3-wallet
```

2. Run database migrations:

```bash
kubectl exec -it $(kubectl get pods -n web3-wallet -l app=api-gateway -o jsonpath='{.items[0].metadata.name}') -n web3-wallet -- go run db/migrate.go -up -config config/config.yaml
```

## Cloud Provider Deployment

### AWS (Amazon Web Services)

1. Set up an EKS (Elastic Kubernetes Service) cluster:

```bash
eksctl create cluster --name web3-wallet --region us-west-2 --nodegroup-name standard-workers --node-type t3.medium --nodes 3 --nodes-min 1 --nodes-max 4 --managed
```

2. Deploy the system using Kubernetes (see above).

### GCP (Google Cloud Platform)

1. Set up a GKE (Google Kubernetes Engine) cluster:

```bash
gcloud container clusters create web3-wallet --zone us-central1-a --num-nodes 3
```

2. Deploy the system using Kubernetes (see above).

### Azure (Microsoft Azure)

1. Set up an AKS (Azure Kubernetes Service) cluster:

```bash
az aks create --resource-group myResourceGroup --name web3-wallet --node-count 3 --enable-addons monitoring --generate-ssh-keys
```

2. Deploy the system using Kubernetes (see above).

## Configuration

The Web3 Wallet Backend system is configured using a YAML configuration file. The configuration file is located at `config/config.yaml` by default.

### Configuration Options

- `server`: HTTP server configuration
- `database`: Database configuration
- `redis`: Redis configuration
- `blockchain`: Blockchain configuration
- `security`: Security configuration
- `logging`: Logging configuration
- `monitoring`: Monitoring configuration
- `notification`: Notification configuration

### Environment Variables

The configuration can be overridden using environment variables. The environment variables follow the pattern `CONFIG_SECTION_KEY`. For example, to override the database host, you can set the environment variable `CONFIG_DATABASE_HOST`.

## Monitoring

The Web3 Wallet Backend system includes monitoring capabilities using Prometheus and Grafana.

### Prometheus

Prometheus is used to collect metrics from the services. Each service exposes metrics at the `/metrics` endpoint.

### Grafana

Grafana is used to visualize the metrics collected by Prometheus. The system includes pre-configured dashboards for monitoring the services.

## Backup and Recovery

### Database Backup

The PostgreSQL database should be backed up regularly. You can use the following command to create a backup:

```bash
pg_dump -U postgres -d web3_wallet > backup.sql
```

### Database Recovery

To restore the database from a backup, you can use the following command:

```bash
psql -U postgres -d web3_wallet < backup.sql
```

## Troubleshooting

### Common Issues

- **Service not starting**: Check the logs for error messages.
- **Database connection error**: Check the database configuration and ensure that the database is running.
- **Redis connection error**: Check the Redis configuration and ensure that Redis is running.
- **Blockchain connection error**: Check the blockchain configuration and ensure that the blockchain nodes are accessible.

### Logs

The logs are written to the standard output by default. You can view the logs using the following commands:

```bash
# Docker Compose
docker-compose logs -f

# Kubernetes
kubectl logs -f -n web3-wallet $(kubectl get pods -n web3-wallet -l app=api-gateway -o jsonpath='{.items[0].metadata.name}')
```
