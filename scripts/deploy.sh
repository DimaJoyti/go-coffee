#!/bin/bash

# Go Coffee Platform Deployment Script
# This script handles deployment to different environments

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io/dimajoyti}"
DOCKER_TAG="${DOCKER_TAG:-latest}"
ENVIRONMENT="${ENVIRONMENT:-development}"
NAMESPACE="${NAMESPACE:-go-coffee}"

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

# Help function
show_help() {
    cat << EOF
Go Coffee Platform Deployment Script

Usage: $0 [OPTIONS] COMMAND

Commands:
    docker          Build and run with Docker Compose
    k8s             Deploy to Kubernetes
    monitoring      Deploy monitoring stack
    cleanup         Clean up resources
    help            Show this help message

Options:
    -e, --env       Environment (development|staging|production) [default: development]
    -n, --namespace Kubernetes namespace [default: go-coffee]
    -t, --tag       Docker image tag [default: latest]
    -r, --registry  Docker registry [default: ghcr.io/dimajoyti]
    -h, --help      Show this help message

Examples:
    $0 docker                           # Run with Docker Compose (development)
    $0 -e production k8s               # Deploy to Kubernetes (production)
    $0 -e staging monitoring           # Deploy monitoring stack (staging)
    $0 cleanup                         # Clean up all resources

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -t|--tag)
                DOCKER_TAG="$2"
                shift 2
                ;;
            -r|--registry)
                DOCKER_REGISTRY="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            docker|k8s|monitoring|cleanup|help)
                COMMAND="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    if [[ -z "${COMMAND:-}" ]]; then
        log_error "No command specified"
        show_help
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if required tools are installed
    local tools=("docker" "docker-compose")
    
    if [[ "$COMMAND" == "k8s" || "$COMMAND" == "monitoring" ]]; then
        tools+=("kubectl" "helm")
    fi

    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool is not installed or not in PATH"
            exit 1
        fi
    done

    # Check if Docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Load environment configuration
load_env_config() {
    log_info "Loading environment configuration for: $ENVIRONMENT"

    local env_file="$PROJECT_ROOT/.env.$ENVIRONMENT"
    if [[ -f "$env_file" ]]; then
        set -a
        source "$env_file"
        set +a
        log_success "Loaded environment configuration from $env_file"
    else
        log_warning "Environment file $env_file not found, using defaults"
    fi

    # Export variables for Docker Compose and Kubernetes
    export ENVIRONMENT
    export DOCKER_REGISTRY
    export DOCKER_TAG
    export NAMESPACE
}

# Docker Compose deployment
deploy_docker() {
    log_info "Deploying with Docker Compose..."

    cd "$PROJECT_ROOT/docker"

    # Determine compose files based on environment
    local compose_files="-f docker-compose.yml"
    
    case "$ENVIRONMENT" in
        development)
            compose_files="$compose_files -f docker-compose.override.yml"
            ;;
        production)
            compose_files="$compose_files -f docker-compose.prod.yml"
            ;;
        staging)
            compose_files="$compose_files -f docker-compose.staging.yml"
            ;;
    esac

    # Build and start services
    log_info "Building and starting services..."
    docker-compose $compose_files build --parallel
    docker-compose $compose_files up -d

    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 30

    # Check service health
    check_docker_health

    log_success "Docker deployment completed successfully"
}

# Check Docker service health
check_docker_health() {
    log_info "Checking service health..."

    local services=("postgres" "redis" "auth-service")
    local healthy_count=0

    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "healthy\|Up"; then
            log_success "$service is healthy"
            ((healthy_count++))
        else
            log_warning "$service is not healthy"
        fi
    done

    if [[ $healthy_count -eq ${#services[@]} ]]; then
        log_success "All core services are healthy"
    else
        log_warning "Some services are not healthy, check logs with: docker-compose logs"
    fi
}

# Kubernetes deployment
deploy_k8s() {
    log_info "Deploying to Kubernetes..."

    cd "$PROJECT_ROOT"

    # Check if kubectl is configured
    if ! kubectl cluster-info &> /dev/null; then
        log_error "kubectl is not configured or cluster is not accessible"
        exit 1
    fi

    # Create namespace if it doesn't exist
    log_info "Creating namespace: $NAMESPACE"
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # Apply Kubernetes manifests
    log_info "Applying Kubernetes manifests..."
    
    # Apply in order: namespace, secrets, configmaps, services, deployments
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/secrets.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/postgres.yaml
    kubectl apply -f k8s/redis.yaml
    
    # Wait for infrastructure to be ready
    log_info "Waiting for infrastructure to be ready..."
    kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n "$NAMESPACE" --timeout=300s

    # Apply application services
    kubectl apply -f k8s/go-coffee-services.yaml

    # Wait for application services
    log_info "Waiting for application services to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=go-coffee -n "$NAMESPACE" --timeout=600s

    # Check deployment status
    check_k8s_health

    log_success "Kubernetes deployment completed successfully"
}

# Check Kubernetes deployment health
check_k8s_health() {
    log_info "Checking Kubernetes deployment health..."

    # Check pod status
    log_info "Pod status:"
    kubectl get pods -n "$NAMESPACE"

    # Check service status
    log_info "Service status:"
    kubectl get services -n "$NAMESPACE"

    # Check if all pods are running
    local not_running=$(kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l)
    
    if [[ $not_running -eq 0 ]]; then
        log_success "All pods are running"
    else
        log_warning "$not_running pods are not running"
        kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running
    fi
}

# Deploy monitoring stack
deploy_monitoring() {
    log_info "Deploying monitoring stack..."

    cd "$PROJECT_ROOT"

    if [[ "$COMMAND" == "k8s" ]]; then
        # Deploy monitoring to Kubernetes
        kubectl apply -f k8s/monitoring.yaml
        
        # Wait for monitoring services
        log_info "Waiting for monitoring services to be ready..."
        kubectl wait --for=condition=ready pod -l app=prometheus -n go-coffee-monitoring --timeout=300s
        kubectl wait --for=condition=ready pod -l app=grafana -n go-coffee-monitoring --timeout=300s
        
    else
        # Deploy monitoring with Docker Compose
        cd docker
        docker-compose -f docker-compose.yml up -d prometheus grafana jaeger alertmanager
    fi

    log_success "Monitoring stack deployed successfully"
    
    # Show monitoring URLs
    show_monitoring_urls
}

# Show monitoring URLs
show_monitoring_urls() {
    log_info "Monitoring URLs:"
    echo "  Prometheus: http://localhost:9090"
    echo "  Grafana: http://localhost:3000 (admin/admin)"
    echo "  Jaeger: http://localhost:16686"
    echo "  AlertManager: http://localhost:9093"
}

# Cleanup resources
cleanup() {
    log_info "Cleaning up resources..."

    # Docker cleanup
    if command -v docker-compose &> /dev/null; then
        cd "$PROJECT_ROOT/docker"
        docker-compose down -v --remove-orphans
        docker system prune -f
    fi

    # Kubernetes cleanup
    if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null; then
        kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
        kubectl delete namespace go-coffee-monitoring --ignore-not-found=true
    fi

    log_success "Cleanup completed"
}

# Main execution
main() {
    parse_args "$@"
    check_prerequisites
    load_env_config

    case "$COMMAND" in
        docker)
            deploy_docker
            ;;
        k8s)
            deploy_k8s
            ;;
        monitoring)
            deploy_monitoring
            ;;
        cleanup)
            cleanup
            ;;
        help)
            show_help
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
