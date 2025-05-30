# Multi-Region Deployment Guide

This guide provides step-by-step instructions for deploying the Web3 wallet backend system across multiple regions.

## Prerequisites

- Google Cloud Platform account with billing enabled
- Terraform 1.0+
- kubectl
- Helm 3+
- gcloud CLI
- Docker

## Step 1: Configure GCP Project

1. Create a new GCP project or use an existing one:

```bash
gcloud projects create web3-wallet-backend --name="Web3 Wallet Backend"
gcloud config set project web3-wallet-backend
```

2. Enable required APIs:

```bash
gcloud services enable compute.googleapis.com
gcloud services enable container.googleapis.com
gcloud services enable redis.googleapis.com
gcloud services enable servicenetworking.googleapis.com
```

## Step 2: Configure Terraform Variables

1. Navigate to the Terraform directory:

```bash
cd web3-wallet-backend/terraform/multi-region
```

2. Create a `terraform.tfvars` file:

```bash
cp terraform.tfvars.example terraform.tfvars
```

3. Edit the `terraform.tfvars` file with your configuration:

```hcl
project_id = "web3-wallet-backend"
primary_region = "us-central1"
regions = ["us-central1", "europe-west1", "asia-east1"]
environment = "prod"
node_count = 3
node_machine_type = "e2-standard-2"
node_disk_size_gb = 100
node_disk_type = "pd-standard"
node_preemptible = false
redis_version = "REDIS_6_X"
redis_tier = "STANDARD_HA"
redis_memory_size_gb = 5
kafka_version = "3.3.1"
```

## Step 3: Deploy Infrastructure with Terraform

1. Initialize Terraform:

```bash
terraform init
```

2. Create a Terraform plan:

```bash
terraform plan -out=tfplan
```

3. Apply the Terraform plan:

```bash
terraform apply tfplan
```

4. Save the outputs for later use:

```bash
terraform output > terraform_outputs.txt
```

## Step 4: Configure kubectl

1. Configure kubectl for each region:

```bash
# For the primary region
gcloud container clusters get-credentials $(terraform output -raw primary_cluster_name) --region $(terraform output -raw primary_region)

# Save the kubeconfig
cp ~/.kube/config ~/.kube/config.primary

# For each additional region
for region in $(terraform output -json regions | jq -r '.[]'); do
  if [ "$region" != "$(terraform output -raw primary_region)" ]; then
    gcloud container clusters get-credentials $(terraform output -raw "${region}_cluster_name") --region $region
    cp ~/.kube/config ~/.kube/config.$region
  fi
done
```

2. Create a script to switch between regions:

```bash
cat > switch-region.sh << 'EOF'
#!/bin/bash
if [ -z "$1" ]; then
  echo "Usage: $0 <region>"
  exit 1
fi

if [ -f ~/.kube/config.$1 ]; then
  cp ~/.kube/config.$1 ~/.kube/config
  echo "Switched to $1 region"
else
  echo "Region $1 not found"
  exit 1
fi
EOF

chmod +x switch-region.sh
```

## Step 5: Build and Push Docker Images

1. Build the Docker images:

```bash
# Set up Docker repository
export REGISTRY=gcr.io/web3-wallet-backend

# Build and push Supply Service
docker build -t $REGISTRY/supply-service:latest -f build/supply-service/Dockerfile .
docker push $REGISTRY/supply-service:latest

# Build and push Order Service
docker build -t $REGISTRY/order-service:latest -f build/order-service/Dockerfile .
docker push $REGISTRY/order-service:latest

# Build and push Claiming Service
docker build -t $REGISTRY/claiming-service:latest -f build/claiming-service/Dockerfile .
docker push $REGISTRY/claiming-service:latest
```

## Step 6: Deploy Services to Kubernetes

1. Update Kubernetes manifests with your registry:

```bash
sed -i "s|\${REGISTRY}|$REGISTRY|g" kubernetes/manifests/*.yaml
```

2. Deploy services to each region:

```bash
# For each region
for region in $(terraform output -json regions | jq -r '.[]'); do
  # Switch to region
  ./switch-region.sh $region
  
  # Create namespace
  kubectl create namespace web3-wallet
  
  # Create ConfigMap
  kubectl create configmap web3-wallet-config --namespace web3-wallet \
    --from-literal=REDIS_HOST=$(terraform output -json redis_hosts | jq -r ".$region") \
    --from-literal=REDIS_PORT=$(terraform output -json redis_ports | jq -r ".$region") \
    --from-literal=KAFKA_BROKERS=$(terraform output -json kafka_brokers | jq -r ".$region")
  
  # Deploy services
  kubectl apply -f kubernetes/manifests/
done
```

## Step 7: Configure Global Load Balancer

The global load balancer is automatically configured by Terraform. You can access the services using the load balancer IP:

```bash
export LB_IP=$(terraform output -raw load_balancer_ip)
echo "Load Balancer IP: $LB_IP"
```

## Step 8: Set Up Monitoring

1. Deploy Prometheus and Grafana:

```bash
# For each region
for region in $(terraform output -json regions | jq -r '.[]'); do
  # Switch to region
  ./switch-region.sh $region
  
  # Deploy Prometheus and Grafana
  kubectl apply -f kubernetes/manifests/monitoring/
done
```

2. Access Grafana:

```bash
# Get Grafana password
kubectl get secret --namespace monitoring grafana -o jsonpath="{.data.admin-password}" | base64 --decode

# Port forward Grafana
kubectl port-forward --namespace monitoring svc/grafana 3000:80
```

3. Import dashboards from `kubernetes/manifests/monitoring/dashboards/`.

## Step 9: Test Failover

1. Test automatic failover by simulating a region failure:

```bash
# Switch to primary region
./switch-region.sh $(terraform output -raw primary_region)

# Scale down all deployments to 0
kubectl scale deployment --all --replicas=0 --namespace web3-wallet
```

2. Verify that traffic is routed to the next available region.

## Step 10: Set Up CI/CD Pipeline

1. Create a CI/CD pipeline using Cloud Build:

```bash
# Create a Cloud Build trigger
gcloud builds triggers create github \
  --repo-name=web3-wallet-backend \
  --branch-pattern=main \
  --build-config=cloudbuild.yaml
```

2. Create a `cloudbuild.yaml` file:

```yaml
steps:
  # Build and push Docker images
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/supply-service:$COMMIT_SHA', '-f', 'build/supply-service/Dockerfile', '.']
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/supply-service:$COMMIT_SHA']
  
  # Update Kubernetes deployments
  - name: 'gcr.io/cloud-builders/kubectl'
    args:
      - 'set'
      - 'image'
      - 'deployment/supply-service'
      - 'supply-service=gcr.io/$PROJECT_ID/supply-service:$COMMIT_SHA'
    env:
      - 'CLOUDSDK_COMPUTE_REGION=us-central1'
      - 'CLOUDSDK_CONTAINER_CLUSTER=web3-wallet-backend-cluster-prod'

  # Repeat for other services and regions
```

## Conclusion

You have successfully deployed the Web3 wallet backend system across multiple regions. The system is now highly available, scalable, and resilient to regional failures.
