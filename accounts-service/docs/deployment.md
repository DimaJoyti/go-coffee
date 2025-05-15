# Deployment Guide

This guide explains how to deploy the Accounts Service to a production environment.

## Prerequisites

- Docker
- Kubernetes cluster (GKE, EKS, AKS, or local)
- kubectl
- Terraform (for infrastructure provisioning)
- Google Cloud SDK (for GCP deployment)

## Deployment Options

The Accounts Service can be deployed in several ways:

1. **Docker Compose**: For local development and testing
2. **Kubernetes**: For production deployment
3. **Google Cloud Platform (GCP)**: Using GKE and Cloud SQL

## Docker Compose Deployment

For local development and testing, you can use Docker Compose:

```bash
docker-compose up accounts-service
```

## Kubernetes Deployment

For production deployment, you can use Kubernetes:

### 1. Build and Push Docker Image

```bash
# Build the Docker image
docker build -t yourusername/accounts-service:latest accounts-service/

# Push the Docker image to a registry
docker push yourusername/accounts-service:latest
```

### 2. Deploy to Kubernetes

```bash
# Apply Kubernetes manifests
kubectl apply -f kubernetes/accounts-service/
```

### 3. Verify Deployment

```bash
# Check deployment status
kubectl get deployments accounts-service

# Check pods
kubectl get pods -l app=accounts-service

# Check service
kubectl get services accounts-service
```

## Google Cloud Platform (GCP) Deployment

For deployment to GCP, you can use Terraform and GKE:

### 1. Set Up Infrastructure with Terraform

```bash
# Initialize Terraform
cd terraform/gcp
terraform init

# Create terraform.tfvars file
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

# Plan the deployment
terraform plan -out=tfplan

# Apply the deployment
terraform apply tfplan
```

### 2. Configure kubectl to use GKE

```bash
gcloud container clusters get-credentials coffee-order-system-gke --region us-central1
```

### 3. Deploy to GKE

```bash
# Apply Kubernetes manifests
kubectl apply -f kubernetes/accounts-service/
```

## CI/CD Pipeline

The Accounts Service uses GitHub Actions for CI/CD. The pipeline is defined in `.github/workflows/accounts-service.yml` and includes the following stages:

1. **Test**: Run unit and integration tests
2. **Build**: Build and push Docker image to Google Container Registry
3. **Deploy**: Deploy to GKE using Kustomize

### Setting Up GitHub Actions

To use the CI/CD pipeline, you need to set up the following secrets in your GitHub repository:

- `GCP_PROJECT_ID`: Your Google Cloud Platform project ID
- `GCP_SA_KEY`: Your Google Cloud Platform service account key (JSON)

## Monitoring and Logging

The Accounts Service includes built-in monitoring and logging:

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics`. You can use Prometheus and Grafana to monitor the service.

### Structured Logging

The service uses structured logging with zap. Logs are output in JSON format by default, making them easy to parse and analyze with tools like Elasticsearch, Logstash, and Kibana (ELK stack).

## Scaling

The service can be scaled horizontally using Kubernetes Horizontal Pod Autoscaler (HPA). The HPA is configured to scale based on CPU and memory usage.

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Check database credentials
   - Verify network connectivity
   - Check database logs

2. **Kafka Connection Issues**
   - Check Kafka broker addresses
   - Verify network connectivity
   - Check Kafka logs

3. **Pod Crashes**
   - Check pod logs: `kubectl logs <pod-name>`
   - Check pod events: `kubectl describe pod <pod-name>`

### Debugging

1. **Enable Debug Logging**
   - Set `LOG_LEVEL` to `debug`
   - Set `LOG_DEVELOPMENT` to `true`

2. **Check Health Endpoint**
   - `curl http://<service-address>/health`

3. **Check Metrics Endpoint**
   - `curl http://<service-address>/metrics`
