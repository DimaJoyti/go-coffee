#!/bin/bash

# Go Coffee Platform - Enhanced Deployment Script
# Advanced deployment system with rollback, health monitoring, and multi-environment support
# Version: 2.0.0
# Usage: ./deploy.sh [OPTIONS] COMMAND
#   -e, --env ENV       Environment (development|staging|production)
#   -n, --namespace NS  Kubernetes namespace
#   -t, --tag TAG       Docker image tag
#   -r, --registry REG  Docker registry
#   -b, --backup        Create backup before deployment
#   -w, --wait          Wait for deployment completion
#   -h, --help          Show this help message

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Source shared library
source "$SCRIPT_DIR/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🚀 Go Coffee Platform Deployment System"

# =============================================================================
# CONFIGURATION
# =============================================================================

DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io/dimajoyti}"
DOCKER_TAG="${DOCKER_TAG:-latest}"
ENVIRONMENT="${ENVIRONMENT:-development}"
NAMESPACE="${NAMESPACE:-go-coffee}"
CREATE_BACKUP=false
WAIT_FOR_COMPLETION=false
DEPLOYMENT_TIMEOUT=600
ROLLBACK_ENABLED=true

# Deployment services in dependency order
DEPLOYMENT_SERVICES=(
    "postgres"
    "redis"
    "auth-service"
    "payment-service"
    "order-service"
    "kitchen-service"
    "user-gateway"
    "communication-hub"
    "ai-search"
    "ai-service"
    "api-gateway"
)

# Legacy function compatibility
log_info() { print_info "$1"; }
log_success() { print_status "$1"; }
log_warning() { print_warning "$1"; }
log_error() { print_error "$1"; }

# Enhanced help function
show_deploy_help() {
    show_usage "deploy.sh" \
        "Advanced deployment system for Go Coffee platform with rollback and monitoring" \
        "  ./deploy.sh [OPTIONS] COMMAND

  Commands:
    docker          Build and run with Docker Compose
    k8s             Deploy to Kubernetes cluster
    monitoring      Deploy monitoring stack (Prometheus, Grafana, Jaeger)
    rollback        Rollback to previous deployment
    status          Check deployment status
    logs            View deployment logs
    cleanup         Clean up resources
    help            Show this help message

  Options:
    -e, --env ENV       Environment (development|staging|production) [default: development]
    -n, --namespace NS  Kubernetes namespace [default: go-coffee]
    -t, --tag TAG       Docker image tag [default: latest]
    -r, --registry REG  Docker registry [default: ghcr.io/dimajoyti]
    -b, --backup        Create backup before deployment
    -w, --wait          Wait for deployment completion
    -h, --help          Show this help message

  Examples:
    ./deploy.sh docker                           # Docker Compose (development)
    ./deploy.sh -e production -b k8s            # K8s production with backup
    ./deploy.sh -e staging monitoring           # Deploy monitoring (staging)
    ./deploy.sh rollback                        # Rollback last deployment
    ./deploy.sh status                          # Check deployment status
    ./deploy.sh cleanup                         # Clean up all resources"
}

# Enhanced argument parsing
parse_deploy_args() {
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
            -b|--backup)
                CREATE_BACKUP=true
                shift
                ;;
            -w|--wait)
                WAIT_FOR_COMPLETION=true
                shift
                ;;
            -h|--help)
                show_deploy_help
                exit 0
                ;;
            docker|k8s|monitoring|rollback|status|logs|cleanup|help)
                COMMAND="$1"
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_deploy_help
                exit 1
                ;;
        esac
    done

    if [[ -z "${COMMAND:-}" ]]; then
        print_error "No command specified"
        show_deploy_help
        exit 1
    fi
}

# Enhanced prerequisites check
check_deployment_prerequisites() {
    print_header "🔍 Checking Deployment Prerequisites"

    # Check if required tools are installed
    local tools=("docker" "docker-compose")

    if [[ "$COMMAND" == "k8s" || "$COMMAND" == "monitoring" || "$COMMAND" == "rollback" ]]; then
        tools+=("kubectl")
    fi

    if [[ "$COMMAND" == "monitoring" ]]; then
        tools+=("helm")
    fi

    check_dependencies "${tools[@]}" || exit 1

    # Check if Docker is running
    if ! docker info &> /dev/null; then
        print_error "Docker is not running"
        exit 1
    fi

    # Check Kubernetes connectivity if needed
    if [[ "$COMMAND" == "k8s" || "$COMMAND" == "monitoring" || "$COMMAND" == "rollback" ]]; then
        if ! kubectl cluster-info &> /dev/null; then
            print_error "kubectl is not configured or cluster is not accessible"
            exit 1
        fi
        print_status "Kubernetes cluster is accessible"
    fi

    print_status "All prerequisites are available"
}

# Create deployment backup
create_deployment_backup() {
    if [[ "$CREATE_BACKUP" != "true" ]]; then
        return 0
    fi

    print_header "💾 Creating Deployment Backup"

    local backup_dir="backups/$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"

    case "$COMMAND" in
        "k8s")
            # Backup Kubernetes resources
            print_progress "Backing up Kubernetes resources..."
            kubectl get all -n "$NAMESPACE" -o yaml > "$backup_dir/k8s-resources.yaml"
            kubectl get configmaps -n "$NAMESPACE" -o yaml > "$backup_dir/configmaps.yaml"
            kubectl get secrets -n "$NAMESPACE" -o yaml > "$backup_dir/secrets.yaml"
            ;;
        "docker")
            # Backup Docker Compose state
            print_progress "Backing up Docker Compose state..."
            cd "$PROJECT_ROOT/docker"
            docker-compose config > "$backup_dir/docker-compose-resolved.yaml"
            ;;
    esac

    # Save deployment metadata
    cat > "$backup_dir/metadata.json" <<EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "environment": "$ENVIRONMENT",
    "docker_tag": "$DOCKER_TAG",
    "namespace": "$NAMESPACE",
    "command": "$COMMAND"
}
EOF

    print_status "Backup created: $backup_dir"
    echo "$backup_dir" > ".last-backup"
}

# Enhanced Docker deployment with health monitoring
deploy_docker_enhanced() {
    print_header "🐳 Enhanced Docker Deployment"

    cd "$PROJECT_ROOT/docker"

    # Determine compose files based on environment
    local compose_files="-f docker-compose.yml"

    case "$ENVIRONMENT" in
        development)
            if [[ -f "docker-compose.override.yml" ]]; then
                compose_files="$compose_files -f docker-compose.override.yml"
            fi
            ;;
        production)
            if [[ -f "docker-compose.prod.yml" ]]; then
                compose_files="$compose_files -f docker-compose.prod.yml"
            fi
            ;;
        staging)
            if [[ -f "docker-compose.staging.yml" ]]; then
                compose_files="$compose_files -f docker-compose.staging.yml"
            fi
            ;;
    esac

    print_info "Using compose files: $compose_files"
    print_info "Environment: $ENVIRONMENT"
    print_info "Docker tag: $DOCKER_TAG"

    # Build services
    print_progress "Building Docker images..."
    if docker-compose $compose_files build --parallel; then
        print_status "Docker images built successfully"
    else
        print_error "Failed to build Docker images"
        exit 1
    fi

    # Start infrastructure services first
    print_progress "Starting infrastructure services..."
    docker-compose $compose_files up -d postgres redis

    # Wait for infrastructure to be ready
    print_progress "Waiting for infrastructure services..."
    sleep 15

    # Start application services
    print_progress "Starting application services..."
    for service in "${DEPLOYMENT_SERVICES[@]}"; do
        if [[ "$service" != "postgres" && "$service" != "redis" ]]; then
            print_progress "Starting $service..."
            docker-compose $compose_files up -d "$service"
            sleep 5
        fi
    done

    # Wait for all services if requested
    if [[ "$WAIT_FOR_COMPLETION" == "true" ]]; then
        print_progress "Waiting for all services to be healthy..."
        sleep 30
        check_docker_health_enhanced
    fi

    print_success "Docker deployment completed successfully"
    show_docker_endpoints
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

# Enhanced Docker health check
check_docker_health_enhanced() {
    print_header "🏥 Checking Service Health"

    local healthy_count=0
    local total_count=0

    for service in "${DEPLOYMENT_SERVICES[@]}"; do
        ((total_count++))
        print_progress "Checking $service..."

        if docker-compose ps "$service" | grep -q "healthy\|Up"; then
            # Additional health check via HTTP if possible
            case "$service" in
                "auth-service"|"payment-service"|"order-service"|"kitchen-service"|"api-gateway")
                    local port=$(docker-compose port "$service" 8080 2>/dev/null | cut -d: -f2)
                    if [[ -n "$port" ]] && curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
                        print_status "$service is healthy (HTTP check passed)"
                        ((healthy_count++))
                    else
                        print_warning "$service is running but HTTP health check failed"
                    fi
                    ;;
                *)
                    print_status "$service is healthy"
                    ((healthy_count++))
                    ;;
            esac
        else
            print_error "$service is not healthy"
        fi
    done

    print_info "Health Summary: $healthy_count/$total_count services healthy"

    if [[ $healthy_count -eq $total_count ]]; then
        print_success "All services are healthy!"
    else
        print_warning "Some services need attention. Check logs with: docker-compose logs SERVICE"
    fi
}

# Show Docker endpoints
show_docker_endpoints() {
    print_header "🌐 Service Endpoints"
    print_info "API Gateway: http://localhost:8080"
    print_info "Auth Service: http://localhost:8091"
    print_info "Payment Service: http://localhost:8093"
    print_info "Order Service: http://localhost:8094"
    print_info "Kitchen Service: http://localhost:8095"
    print_info "Health Checks: http://localhost:PORT/health"

    if [[ "$ENVIRONMENT" == "development" ]]; then
        print_info "Development Dashboard: http://localhost:8080/dev"
    fi
}

# Deployment status check
check_deployment_status() {
    print_header "📊 Deployment Status"

    case "$ENVIRONMENT" in
        "k8s")
            print_info "Kubernetes deployment status:"
            kubectl get pods -n "$NAMESPACE"
            kubectl get services -n "$NAMESPACE"
            ;;
        *)
            print_info "Docker deployment status:"
            cd "$PROJECT_ROOT/docker"
            docker-compose ps
            ;;
    esac
}

# View deployment logs
view_deployment_logs() {
    print_header "📋 Deployment Logs"

    case "$ENVIRONMENT" in
        "k8s")
            print_info "Recent Kubernetes logs:"
            kubectl logs -n "$NAMESPACE" --tail=100 -l app.kubernetes.io/name=go-coffee
            ;;
        *)
            print_info "Recent Docker logs:"
            cd "$PROJECT_ROOT/docker"
            docker-compose logs --tail=100
            ;;
    esac
}

# Rollback deployment
rollback_deployment() {
    print_header "🔄 Rolling Back Deployment"

    if [[ ! -f ".last-backup" ]]; then
        print_error "No backup found for rollback"
        exit 1
    fi

    local backup_dir=$(cat ".last-backup")
    if [[ ! -d "$backup_dir" ]]; then
        print_error "Backup directory not found: $backup_dir"
        exit 1
    fi

    print_info "Rolling back to: $backup_dir"

    case "$COMMAND" in
        "k8s")
            print_progress "Restoring Kubernetes resources..."
            kubectl apply -f "$backup_dir/k8s-resources.yaml"
            kubectl apply -f "$backup_dir/configmaps.yaml"
            kubectl apply -f "$backup_dir/secrets.yaml"
            ;;
        "docker")
            print_progress "Restoring Docker Compose state..."
            cd "$PROJECT_ROOT/docker"
            docker-compose down
            docker-compose -f "$backup_dir/docker-compose-resolved.yaml" up -d
            ;;
    esac

    print_success "Rollback completed successfully"
}

# Enhanced main execution
main() {
    local start_time=$(date +%s)

    # Parse arguments
    parse_deploy_args "$@"

    # Check prerequisites
    check_deployment_prerequisites

    # Load environment configuration
    load_env_config

    print_info "Deployment Configuration:"
    print_info "  Command: $COMMAND"
    print_info "  Environment: $ENVIRONMENT"
    print_info "  Namespace: $NAMESPACE"
    print_info "  Docker Tag: $DOCKER_TAG"
    print_info "  Registry: $DOCKER_REGISTRY"
    print_info "  Backup: $CREATE_BACKUP"
    print_info "  Wait: $WAIT_FOR_COMPLETION"

    # Create backup if requested
    create_deployment_backup

    case "$COMMAND" in
        docker)
            deploy_docker_enhanced
            ;;
        k8s)
            deploy_k8s
            ;;
        monitoring)
            deploy_monitoring
            ;;
        rollback)
            rollback_deployment
            ;;
        status)
            check_deployment_status
            ;;
        logs)
            view_deployment_logs
            ;;
        cleanup)
            cleanup
            ;;
        help)
            show_deploy_help
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            show_deploy_help
            exit 1
            ;;
    esac

    # Calculate deployment time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    print_success "Deployment operation completed in ${total_time}s"
}

# Run main function with all arguments
main "$@"
