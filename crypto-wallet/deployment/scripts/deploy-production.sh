#!/bin/bash

# Cryptocurrency Automation Platform - Production Deployment Script
# This script deploys the complete crypto automation platform to Kubernetes

set -euo pipefail

# Configuration
NAMESPACE="crypto-automation"
DEPLOYMENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PRODUCTION_DIR="${DEPLOYMENT_DIR}/production"
KUBECTL_TIMEOUT="300s"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    # Check if required files exist
    local required_files=(
        "${PRODUCTION_DIR}/crypto-automation-platform.yaml"
        "${PRODUCTION_DIR}/infrastructure.yaml"
        "${PRODUCTION_DIR}/ingress-monitoring.yaml"
    )
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            log_error "Required file not found: $file"
            exit 1
        fi
    done
    
    log_success "Prerequisites check passed"
}

# Create namespace
create_namespace() {
    log_info "Creating namespace: $NAMESPACE"
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Namespace $NAMESPACE already exists"
    else
        kubectl create namespace "$NAMESPACE"
        kubectl label namespace "$NAMESPACE" app=crypto-automation-platform environment=production
        log_success "Namespace $NAMESPACE created"
    fi
}

# Deploy secrets (with user input for sensitive data)
deploy_secrets() {
    log_info "Deploying secrets..."
    
    # Check if secrets already exist
    if kubectl get secret crypto-platform-secrets -n "$NAMESPACE" &> /dev/null; then
        log_warning "Secrets already exist. Skipping secret creation."
        return
    fi
    
    log_info "Please provide the following sensitive information:"
    
    # Database password
    read -s -p "Enter database password: " DB_PASSWORD
    echo
    
    # Redis password
    read -s -p "Enter Redis password: " REDIS_PASSWORD
    echo
    
    # JWT secret key
    read -s -p "Enter JWT secret key: " JWT_SECRET
    echo
    
    # Flashbots private key
    read -s -p "Enter Flashbots private key: " FLASHBOTS_KEY
    echo
    
    # Infura API key
    read -p "Enter Infura API key: " INFURA_KEY
    
    # Alchemy API key
    read -p "Enter Alchemy API key: " ALCHEMY_KEY
    
    # Create secrets
    kubectl create secret generic crypto-platform-secrets \
        --from-literal=database-password="$DB_PASSWORD" \
        --from-literal=redis-password="$REDIS_PASSWORD" \
        --from-literal=jwt-secret-key="$JWT_SECRET" \
        --from-literal=flashbots-private-key="$FLASHBOTS_KEY" \
        --from-literal=infura-api-key="$INFURA_KEY" \
        --from-literal=alchemy-api-key="$ALCHEMY_KEY" \
        -n "$NAMESPACE"
    
    # Create Grafana secrets
    kubectl create secret generic grafana-secrets \
        --from-literal=admin-password="crypto_grafana_admin123" \
        -n "$NAMESPACE"
    
    log_success "Secrets deployed"
}

# Deploy infrastructure (PostgreSQL, Redis, Monitoring)
deploy_infrastructure() {
    log_info "Deploying infrastructure components..."
    
    kubectl apply -f "${PRODUCTION_DIR}/infrastructure.yaml" --timeout="$KUBECTL_TIMEOUT"
    
    # Wait for PostgreSQL to be ready
    log_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    
    # Wait for Redis to be ready
    log_info "Waiting for Redis to be ready..."
    kubectl wait --for=condition=ready pod -l app=redis -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    
    # Wait for Prometheus to be ready
    log_info "Waiting for Prometheus to be ready..."
    kubectl wait --for=condition=ready pod -l app=prometheus -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    
    log_success "Infrastructure components deployed"
}

# Deploy main application
deploy_application() {
    log_info "Deploying crypto automation platform..."
    
    kubectl apply -f "${PRODUCTION_DIR}/crypto-automation-platform.yaml" --timeout="$KUBECTL_TIMEOUT"
    
    # Wait for deployment to be ready
    log_info "Waiting for application to be ready..."
    kubectl wait --for=condition=available deployment/crypto-automation-api -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    
    log_success "Crypto automation platform deployed"
}

# Deploy ingress and monitoring
deploy_ingress_monitoring() {
    log_info "Deploying ingress and monitoring components..."
    
    kubectl apply -f "${PRODUCTION_DIR}/ingress-monitoring.yaml" --timeout="$KUBECTL_TIMEOUT"
    
    # Wait for Grafana to be ready
    log_info "Waiting for Grafana to be ready..."
    kubectl wait --for=condition=ready pod -l app=grafana -n "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
    
    log_success "Ingress and monitoring components deployed"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check all pods are running
    local failed_pods
    failed_pods=$(kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l)
    
    if [[ $failed_pods -gt 0 ]]; then
        log_warning "Some pods are not running:"
        kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running
    else
        log_success "All pods are running"
    fi
    
    # Check services
    log_info "Checking services..."
    kubectl get services -n "$NAMESPACE"
    
    # Check ingress
    log_info "Checking ingress..."
    kubectl get ingress -n "$NAMESPACE"
    
    # Test API health endpoint
    log_info "Testing API health endpoint..."
    local api_pod
    api_pod=$(kubectl get pods -n "$NAMESPACE" -l app=crypto-automation-platform,component=api -o jsonpath='{.items[0].metadata.name}')
    
    if kubectl exec -n "$NAMESPACE" "$api_pod" -- curl -f http://localhost:8080/health &> /dev/null; then
        log_success "API health check passed"
    else
        log_warning "API health check failed"
    fi
    
    log_success "Deployment verification completed"
}

# Display deployment information
display_info() {
    log_info "Deployment Information:"
    echo
    echo "Namespace: $NAMESPACE"
    echo "API Endpoint: https://api.crypto-automation.com"
    echo "Monitoring: https://monitoring.crypto-automation.com"
    echo
    echo "To check deployment status:"
    echo "  kubectl get all -n $NAMESPACE"
    echo
    echo "To view logs:"
    echo "  kubectl logs -f deployment/crypto-automation-api -n $NAMESPACE"
    echo
    echo "To access Grafana:"
    echo "  Username: admin"
    echo "  Password: crypto_grafana_admin123"
    echo
    echo "To port-forward for local access:"
    echo "  kubectl port-forward service/crypto-automation-api-service 8080:80 -n $NAMESPACE"
    echo "  kubectl port-forward service/grafana-service 3000:3000 -n $NAMESPACE"
    echo
}

# Cleanup function
cleanup() {
    log_info "Cleaning up deployment..."
    
    read -p "Are you sure you want to delete the entire deployment? (yes/no): " confirm
    if [[ $confirm == "yes" ]]; then
        kubectl delete namespace "$NAMESPACE" --timeout="$KUBECTL_TIMEOUT"
        log_success "Deployment cleaned up"
    else
        log_info "Cleanup cancelled"
    fi
}

# Main deployment function
main() {
    local action="${1:-deploy}"
    
    case $action in
        "deploy")
            log_info "Starting production deployment of Crypto Automation Platform..."
            check_prerequisites
            create_namespace
            deploy_secrets
            deploy_infrastructure
            deploy_application
            deploy_ingress_monitoring
            verify_deployment
            display_info
            log_success "Deployment completed successfully!"
            ;;
        "cleanup")
            cleanup
            ;;
        "verify")
            verify_deployment
            ;;
        "info")
            display_info
            ;;
        *)
            echo "Usage: $0 [deploy|cleanup|verify|info]"
            echo
            echo "Commands:"
            echo "  deploy   - Deploy the complete platform (default)"
            echo "  cleanup  - Remove the entire deployment"
            echo "  verify   - Verify the current deployment"
            echo "  info     - Display deployment information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
