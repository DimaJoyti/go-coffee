#!/bin/bash

# ☕ Go Coffee - Complete Infrastructure Deployment Script
# Deploys the entire Go Coffee ecosystem using Terraform and Kubernetes

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TERRAFORM_DIR="$PROJECT_ROOT/terraform"
HELM_DIR="$PROJECT_ROOT/helm"
K8S_DIR="$PROJECT_ROOT/k8s"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
CLOUD_PROVIDER="${CLOUD_PROVIDER:-gcp}"
PROJECT_ID="${PROJECT_ID:-go-coffee-dev}"
REGION="${REGION:-europe-west3}"
CLUSTER_NAME="${CLUSTER_NAME:-go-coffee-cluster}"
DOMAIN="${DOMAIN:-gocoffee.dev}"
ENABLE_MONITORING="${ENABLE_MONITORING:-true}"
ENABLE_SERVICE_MESH="${ENABLE_SERVICE_MESH:-true}"
ENABLE_GITOPS="${ENABLE_GITOPS:-false}"

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check required tools
    local tools=("terraform" "kubectl" "helm" "gcloud" "docker")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is not installed or not in PATH"
        fi
    done
    
    # Check Terraform version
    local tf_version=$(terraform version -json | jq -r '.terraform_version')
    if [[ $(echo "$tf_version 1.5.0" | tr " " "\n" | sort -V | head -n1) != "1.5.0" ]]; then
        error "Terraform version 1.5.0 or higher is required. Current: $tf_version"
    fi
    
    # Check kubectl version
    if ! kubectl version --client &> /dev/null; then
        error "kubectl is not properly configured"
    fi
    
    # Check Helm version
    local helm_version=$(helm version --short | cut -d'+' -f1 | cut -d'v' -f2)
    if [[ $(echo "$helm_version 3.10.0" | tr " " "\n" | sort -V | head -n1) != "3.10.0" ]]; then
        error "Helm version 3.10.0 or higher is required. Current: $helm_version"
    fi
    
    success "All prerequisites met"
}

# Setup cloud provider authentication
setup_cloud_auth() {
    log "Setting up cloud provider authentication..."
    
    case "$CLOUD_PROVIDER" in
        "gcp")
            if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
                error "No active GCP authentication found. Run 'gcloud auth login'"
            fi
            
            # Set project
            gcloud config set project "$PROJECT_ID"
            
            # Enable required APIs
            local apis=(
                "container.googleapis.com"
                "compute.googleapis.com"
                "iam.googleapis.com"
                "cloudresourcemanager.googleapis.com"
                "sqladmin.googleapis.com"
                "redis.googleapis.com"
                "monitoring.googleapis.com"
                "logging.googleapis.com"
                "secretmanager.googleapis.com"
            )
            
            for api in "${apis[@]}"; do
                info "Enabling $api..."
                gcloud services enable "$api" --project="$PROJECT_ID"
            done
            ;;
        "aws")
            if ! aws sts get-caller-identity &> /dev/null; then
                error "AWS credentials not configured. Run 'aws configure'"
            fi
            ;;
        "azure")
            if ! az account show &> /dev/null; then
                error "Azure credentials not configured. Run 'az login'"
            fi
            ;;
        *)
            error "Unsupported cloud provider: $CLOUD_PROVIDER"
            ;;
    esac
    
    success "Cloud provider authentication configured"
}

# Deploy Terraform infrastructure
deploy_terraform() {
    log "Deploying Terraform infrastructure..."
    
    cd "$TERRAFORM_DIR"
    
    # Initialize Terraform
    info "Initializing Terraform..."
    terraform init -upgrade
    
    # Create terraform.tfvars if it doesn't exist
    if [[ ! -f "terraform.tfvars" ]]; then
        info "Creating terraform.tfvars..."
        cat > terraform.tfvars <<EOF
project_id = "$PROJECT_ID"
region = "$REGION"
environment = "$ENVIRONMENT"
gke_cluster_name = "$CLUSTER_NAME"
enable_monitoring = $ENABLE_MONITORING
enable_service_mesh = $ENABLE_SERVICE_MESH
EOF
    fi
    
    # Plan and apply
    info "Planning Terraform deployment..."
    terraform plan -out=tfplan
    
    info "Applying Terraform configuration..."
    terraform apply tfplan
    
    # Get cluster credentials
    case "$CLOUD_PROVIDER" in
        "gcp")
            info "Getting GKE cluster credentials..."
            gcloud container clusters get-credentials "$CLUSTER_NAME" \
                --region="$REGION" \
                --project="$PROJECT_ID"
            ;;
        "aws")
            info "Getting EKS cluster credentials..."
            aws eks update-kubeconfig \
                --region="$REGION" \
                --name="$CLUSTER_NAME"
            ;;
    esac
    
    success "Terraform infrastructure deployed"
}

# Deploy Kubernetes resources
deploy_kubernetes() {
    log "Deploying Kubernetes resources..."
    
    # Create namespaces
    info "Creating namespaces..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: go-coffee
  labels:
    istio-injection: enabled
    name: go-coffee
---
apiVersion: v1
kind: Namespace
metadata:
  name: go-coffee-monitoring
  labels:
    name: go-coffee-monitoring
---
apiVersion: v1
kind: Namespace
metadata:
  name: go-coffee-system
  labels:
    name: go-coffee-system
EOF
    
    # Deploy base Kubernetes resources
    if [[ -d "$K8S_DIR/base" ]]; then
        info "Deploying base Kubernetes resources..."
        kubectl apply -k "$K8S_DIR/base"
    fi
    
    success "Kubernetes resources deployed"
}

# Deploy Helm charts
deploy_helm() {
    log "Deploying Helm charts..."
    
    # Add Helm repositories
    info "Adding Helm repositories..."
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo update
    
    # Deploy infrastructure dependencies
    if [[ "$ENABLE_MONITORING" == "true" ]]; then
        info "Deploying monitoring stack..."
        helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
            --namespace go-coffee-monitoring \
            --create-namespace \
            --values "$HELM_DIR/monitoring/prometheus-values.yaml" \
            --wait
            
        helm upgrade --install grafana grafana/grafana \
            --namespace go-coffee-monitoring \
            --values "$HELM_DIR/monitoring/grafana-values.yaml" \
            --wait
    fi
    
    if [[ "$ENABLE_SERVICE_MESH" == "true" ]]; then
        info "Deploying Istio service mesh..."
        helm upgrade --install istio-base istio/base \
            --namespace istio-system \
            --create-namespace \
            --wait
            
        helm upgrade --install istiod istio/istiod \
            --namespace istio-system \
            --wait
            
        helm upgrade --install istio-gateway istio/gateway \
            --namespace istio-ingress \
            --create-namespace \
            --wait
    fi
    
    # Deploy Go Coffee platform
    info "Deploying Go Coffee platform..."
    helm upgrade --install go-coffee-platform "$HELM_DIR/go-coffee-platform" \
        --namespace go-coffee \
        --create-namespace \
        --values "$HELM_DIR/go-coffee-platform/values-$ENVIRONMENT.yaml" \
        --set global.environment="$ENVIRONMENT" \
        --set global.domain="$DOMAIN" \
        --wait \
        --timeout=10m
    
    success "Helm charts deployed"
}

# Setup GitOps (optional)
setup_gitops() {
    if [[ "$ENABLE_GITOPS" != "true" ]]; then
        return 0
    fi
    
    log "Setting up GitOps with ArgoCD..."
    
    # Install ArgoCD
    info "Installing ArgoCD..."
    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    
    # Wait for ArgoCD to be ready
    info "Waiting for ArgoCD to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd
    
    # Apply ArgoCD applications
    if [[ -d "$PROJECT_ROOT/gitops/argocd" ]]; then
        info "Applying ArgoCD applications..."
        kubectl apply -f "$PROJECT_ROOT/gitops/argocd/"
    fi
    
    # Get ArgoCD admin password
    local argocd_password=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
    info "ArgoCD admin password: $argocd_password"
    
    success "GitOps setup completed"
}

# Verify deployment
verify_deployment() {
    log "Verifying deployment..."
    
    # Check cluster status
    info "Checking cluster status..."
    kubectl cluster-info
    
    # Check nodes
    info "Checking nodes..."
    kubectl get nodes -o wide
    
    # Check Go Coffee pods
    info "Checking Go Coffee pods..."
    kubectl get pods -n go-coffee -o wide
    
    # Check services
    info "Checking services..."
    kubectl get services -n go-coffee
    
    # Check ingress
    if kubectl get ingress -n go-coffee &> /dev/null; then
        info "Checking ingress..."
        kubectl get ingress -n go-coffee
    fi
    
    # Health checks
    info "Performing health checks..."
    local api_gateway_url="http://$(kubectl get service go-coffee-platform-api-gateway -n go-coffee -o jsonpath='{.status.loadBalancer.ingress[0].ip}'):80"
    
    if curl -f "$api_gateway_url/health" &> /dev/null; then
        success "API Gateway health check passed"
    else
        warn "API Gateway health check failed"
    fi
    
    success "Deployment verification completed"
}

# Cleanup function
cleanup() {
    if [[ "${1:-}" == "destroy" ]]; then
        warn "Destroying infrastructure..."
        
        # Delete Helm releases
        helm list --all-namespaces -o json | jq -r '.[] | select(.name | startswith("go-coffee")) | "\(.name) \(.namespace)"' | while read name namespace; do
            helm uninstall "$name" -n "$namespace"
        done
        
        # Delete Kubernetes resources
        kubectl delete namespace go-coffee go-coffee-monitoring go-coffee-system --ignore-not-found=true
        
        # Destroy Terraform
        cd "$TERRAFORM_DIR"
        terraform destroy -auto-approve
        
        success "Infrastructure destroyed"
    fi
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    ☕ Go Coffee - Complete Infrastructure Deployment
    ================================================
    
    Next-Generation Web3 Coffee Ecosystem
    Multi-Cloud, AI-Powered, DeFi-Integrated
    
EOF
    echo -e "${NC}"
    
    info "Starting deployment with the following configuration:"
    info "Environment: $ENVIRONMENT"
    info "Cloud Provider: $CLOUD_PROVIDER"
    info "Project ID: $PROJECT_ID"
    info "Region: $REGION"
    info "Cluster Name: $CLUSTER_NAME"
    info "Domain: $DOMAIN"
    info "Monitoring: $ENABLE_MONITORING"
    info "Service Mesh: $ENABLE_SERVICE_MESH"
    info "GitOps: $ENABLE_GITOPS"
    
    # Trap cleanup on exit
    trap 'cleanup' EXIT
    
    # Execute deployment steps
    check_prerequisites
    setup_cloud_auth
    deploy_terraform
    deploy_kubernetes
    deploy_helm
    setup_gitops
    verify_deployment
    
    success "🎉 Go Coffee infrastructure deployment completed successfully!"
    
    echo -e "${CYAN}"
    cat << EOF

    🚀 Your Go Coffee platform is now running!
    
    Access Points:
    - API Gateway: http://$api_gateway_url
    - Grafana: http://grafana.$DOMAIN
    - ArgoCD: http://argocd.$DOMAIN (if GitOps enabled)
    
    Next Steps:
    1. Configure DNS to point to your load balancer IP
    2. Set up SSL certificates
    3. Configure monitoring alerts
    4. Deploy your applications
    
    Documentation: https://docs.gocoffee.dev
    
EOF
    echo -e "${NC}"
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "destroy")
        cleanup destroy
        ;;
    "verify")
        verify_deployment
        ;;
    *)
        echo "Usage: $0 [deploy|destroy|verify]"
        exit 1
        ;;
esac
