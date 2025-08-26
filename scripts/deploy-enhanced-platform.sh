#!/bin/bash

# Enhanced Go Coffee Platform Deployment Script
# This script deploys the complete platform with enterprise-grade infrastructure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT="${ENVIRONMENT:-production}"
NAMESPACE="go-coffee-platform"
MONITORING_NAMESPACE="go-coffee-monitoring"
SECURITY_NAMESPACE="go-coffee-security"
KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
REGISTRY="${REGISTRY:-ghcr.io/dimajoyti/go-coffee}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
DRY_RUN="${DRY_RUN:-false}"
SKIP_TESTS="${SKIP_TESTS:-false}"
FORCE_DEPLOY="${FORCE_DEPLOY:-false}"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================${NC}"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check helm
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed or not in PATH"
        exit 1
    fi
    
    # Check docker
    if ! command -v docker &> /dev/null; then
        print_error "docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster"
        print_status "Please check your kubeconfig: $KUBECONFIG"
        exit 1
    fi
    
    # Check cluster version
    local k8s_version=$(kubectl version --short --client | grep -o 'v[0-9]\+\.[0-9]\+')
    print_status "Kubernetes client version: $k8s_version"
    
    # Check if running in dry-run mode
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "Running in DRY-RUN mode - no changes will be applied"
    fi
    
    print_success "Prerequisites check completed"
}

# Function to build and push images
build_and_push_images() {
    print_header "Building and Pushing Container Images"
    
    local services=("producer" "consumer" "streams" "web3-payment-service" "ai-orchestrator")
    
    for service in "${services[@]}"; do
        print_step "Building $service image..."
        
        if [[ "$DRY_RUN" == "true" ]]; then
            print_status "DRY RUN: Would build and push $REGISTRY/$service:$IMAGE_TAG"
        else
            # Build image
            docker build -t "$REGISTRY/$service:$IMAGE_TAG" -f "Dockerfile.$service" .
            
            # Push image
            docker push "$REGISTRY/$service:$IMAGE_TAG"
            
            print_success "$service image built and pushed"
        fi
    done
}

# Function to create namespaces
create_namespaces() {
    print_header "Creating Namespaces"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would create namespaces"
        kubectl apply --dry-run=client -f k8s/enhanced/namespace.yaml
    else
        kubectl apply -f k8s/enhanced/namespace.yaml
        print_success "Namespaces created"
    fi
}

# Function to deploy security configurations
deploy_security() {
    print_header "Deploying Security Configurations"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy security configurations"
        kubectl apply --dry-run=client -f k8s/enhanced/security.yaml
    else
        # Update secrets with actual values if environment variables are set
        if [[ -n "$DATABASE_PASSWORD" ]]; then
            kubectl create secret generic go-coffee-secrets \
                --from-literal=database-url="postgres://go_coffee_user:$DATABASE_PASSWORD@postgres:5432/go_coffee?sslmode=require" \
                --namespace="$NAMESPACE" \
                --dry-run=client -o yaml | kubectl apply -f -
        fi
        
        kubectl apply -f k8s/enhanced/security.yaml
        print_success "Security configurations deployed"
    fi
}

# Function to deploy OpenTelemetry
deploy_opentelemetry() {
    print_header "Deploying OpenTelemetry Stack"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy OpenTelemetry stack"
        kubectl apply --dry-run=client -f k8s/enhanced/opentelemetry.yaml
    else
        kubectl apply -f k8s/enhanced/opentelemetry.yaml
        
        # Wait for OpenTelemetry collector to be ready
        print_status "Waiting for OpenTelemetry collector to be ready..."
        kubectl wait --for=condition=ready pod -l app=otel-collector -n "$MONITORING_NAMESPACE" --timeout=300s
        
        print_success "OpenTelemetry stack deployed"
    fi
}

# Function to deploy monitoring stack
deploy_monitoring() {
    print_header "Deploying Monitoring Stack"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy monitoring stack"
        kubectl apply --dry-run=client -f k8s/enhanced/monitoring-stack.yaml
    else
        kubectl apply -f k8s/enhanced/monitoring-stack.yaml
        
        # Wait for monitoring services to be ready
        print_status "Waiting for monitoring services to be ready..."
        kubectl wait --for=condition=ready pod -l app=jaeger -n "$MONITORING_NAMESPACE" --timeout=300s
        kubectl wait --for=condition=ready pod -l app=tempo -n "$MONITORING_NAMESPACE" --timeout=300s
        
        print_success "Monitoring stack deployed"
    fi
}

# Function to deploy core services
deploy_core_services() {
    print_header "Deploying Core Services"
    
    # Update image tags in manifests
    local temp_dir=$(mktemp -d)
    cp k8s/enhanced/go-coffee-services.yaml "$temp_dir/"
    
    # Replace image references
    sed -i "s|image: go-coffee/|image: $REGISTRY/|g" "$temp_dir/go-coffee-services.yaml"
    sed -i "s|:latest|:$IMAGE_TAG|g" "$temp_dir/go-coffee-services.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy core services"
        kubectl apply --dry-run=client -f "$temp_dir/go-coffee-services.yaml"
    else
        kubectl apply -f "$temp_dir/go-coffee-services.yaml"
        
        # Wait for services to be ready
        print_status "Waiting for core services to be ready..."
        kubectl wait --for=condition=ready pod -l app=producer-service -n "$NAMESPACE" --timeout=300s
        kubectl wait --for=condition=ready pod -l app=web3-payment-service -n "$NAMESPACE" --timeout=300s
        kubectl wait --for=condition=ready pod -l app=ai-orchestrator-service -n "$NAMESPACE" --timeout=300s
        
        print_success "Core services deployed"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Function to run health checks
run_health_checks() {
    print_header "Running Health Checks"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would run health checks"
        return 0
    fi
    
    # Check service health
    local services=("producer-service" "web3-payment-service" "ai-orchestrator-service")
    
    for service in "${services[@]}"; do
        print_step "Checking health of $service..."
        
        # Get service endpoint
        local service_ip=$(kubectl get service "$service" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}')
        local service_port=$(kubectl get service "$service" -n "$NAMESPACE" -o jsonpath='{.spec.ports[0].port}')
        
        # Run health check from within cluster
        if kubectl run health-check-$service --rm -i --restart=Never --image=curlimages/curl -- \
           curl -f "http://$service_ip:$service_port/health" > /dev/null 2>&1; then
            print_success "$service is healthy"
        else
            print_warning "$service health check failed"
        fi
    done
}

# Function to run smoke tests
run_smoke_tests() {
    print_header "Running Smoke Tests"
    
    if [[ "$SKIP_TESTS" == "true" ]]; then
        print_warning "Skipping smoke tests"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would run smoke tests"
        return 0
    fi
    
    # Port forward services for testing
    print_status "Setting up port forwarding for tests..."
    
    kubectl port-forward service/producer-service 3000:80 -n "$NAMESPACE" &
    local producer_pid=$!
    
    kubectl port-forward service/web3-payment-service 8083:80 -n "$NAMESPACE" &
    local web3_pid=$!
    
    kubectl port-forward service/ai-orchestrator-service 8094:80 -n "$NAMESPACE" &
    local ai_pid=$!
    
    # Wait for port forwarding to be ready
    sleep 10
    
    # Run tests
    local test_failed=false
    
    if [[ -f "scripts/test-core-services.sh" ]]; then
        print_step "Running core services tests..."
        if ! ./scripts/test-core-services.sh; then
            print_error "Core services tests failed"
            test_failed=true
        fi
    fi
    
    if [[ -f "scripts/test-web3-payment.sh" ]]; then
        print_step "Running Web3 payment tests..."
        if ! ./scripts/test-web3-payment.sh; then
            print_error "Web3 payment tests failed"
            test_failed=true
        fi
    fi
    
    if [[ -f "scripts/test-ai-orchestrator.sh" ]]; then
        print_step "Running AI orchestrator tests..."
        if ! ./scripts/test-ai-orchestrator.sh; then
            print_error "AI orchestrator tests failed"
            test_failed=true
        fi
    fi
    
    # Cleanup port forwarding
    kill $producer_pid $web3_pid $ai_pid 2>/dev/null || true
    
    if [[ "$test_failed" == "true" ]] && [[ "$FORCE_DEPLOY" != "true" ]]; then
        print_error "Smoke tests failed. Use FORCE_DEPLOY=true to deploy anyway."
        exit 1
    elif [[ "$test_failed" == "true" ]]; then
        print_warning "Smoke tests failed but deployment forced"
    else
        print_success "All smoke tests passed"
    fi
}

# Function to display deployment summary
display_summary() {
    print_header "Deployment Summary"
    
    echo -e "${CYAN}Environment:${NC} $ENVIRONMENT"
    echo -e "${CYAN}Namespace:${NC} $NAMESPACE"
    echo -e "${CYAN}Image Tag:${NC} $IMAGE_TAG"
    echo -e "${CYAN}Registry:${NC} $REGISTRY"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        echo ""
        echo -e "${CYAN}Service Endpoints:${NC}"
        
        # Get service endpoints
        local producer_ip=$(kubectl get service producer-service -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "N/A")
        local web3_ip=$(kubectl get service web3-payment-service -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "N/A")
        local ai_ip=$(kubectl get service ai-orchestrator-service -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "N/A")
        
        echo -e "${GREEN}Producer Service:${NC}        $producer_ip:80"
        echo -e "${GREEN}Web3 Payment Service:${NC}    $web3_ip:80"
        echo -e "${GREEN}AI Orchestrator Service:${NC} $ai_ip:80"
        
        echo ""
        echo -e "${CYAN}Monitoring:${NC}"
        local jaeger_ip=$(kubectl get service jaeger-query -n "$MONITORING_NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "N/A")
        local otel_ip=$(kubectl get service otel-collector -n "$MONITORING_NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "N/A")
        
        echo -e "${GREEN}Jaeger UI:${NC}               $jaeger_ip:16686"
        echo -e "${GREEN}OpenTelemetry Collector:${NC} $otel_ip:4317"
        
        echo ""
        echo -e "${CYAN}Access Commands:${NC}"
        echo "kubectl port-forward service/producer-service 3000:80 -n $NAMESPACE"
        echo "kubectl port-forward service/web3-payment-service 8083:80 -n $NAMESPACE"
        echo "kubectl port-forward service/ai-orchestrator-service 8094:80 -n $NAMESPACE"
        echo "kubectl port-forward service/jaeger-query 16686:16686 -n $MONITORING_NAMESPACE"
    fi
    
    print_success "Enhanced Go Coffee Platform deployment completed!"
}

# Main deployment function
main() {
    print_header "Enhanced Go Coffee Platform Deployment"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --image-tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            --registry)
                REGISTRY="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            --force-deploy)
                FORCE_DEPLOY="true"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --environment ENV     Deployment environment (default: production)"
                echo "  --image-tag TAG       Container image tag (default: latest)"
                echo "  --registry REGISTRY   Container registry (default: ghcr.io/dimajoyti/go-coffee)"
                echo "  --dry-run            Run in dry-run mode"
                echo "  --skip-tests         Skip smoke tests"
                echo "  --force-deploy       Deploy even if tests fail"
                echo "  --help               Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_prerequisites
    build_and_push_images
    create_namespaces
    deploy_security
    deploy_opentelemetry
    deploy_monitoring
    deploy_core_services
    run_health_checks
    run_smoke_tests
    display_summary
}

# Run main function with all arguments
main "$@"
