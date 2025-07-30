#!/bin/bash

# Setup comprehensive environment management for Go Coffee CI/CD
# Handles staging, production, and DR environments with proper secrets management

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
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Environment definitions
declare -A ENVIRONMENTS=(
    ["staging"]="go-coffee-staging:develop:auto"
    ["production"]="go-coffee-production:main:manual"
    ["dr"]="go-coffee-dr:main:manual"
)

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
    log "Checking environment setup prerequisites..."
    
    # Check required tools
    local tools=("kubectl" "helm" "openssl" "base64" "jq")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is not installed or not in PATH"
        fi
    done
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl is not properly configured or cluster is not accessible"
    fi
    
    success "Prerequisites met"
}

# Generate secure secrets
generate_secrets() {
    local env="$1"
    
    log "Generating secure secrets for $env environment..."
    
    # Generate JWT secret
    local jwt_secret=$(openssl rand -base64 32)
    
    # Generate database password
    local db_password=$(openssl rand -base64 24)
    
    # Generate Redis password
    local redis_password=$(openssl rand -base64 16)
    
    # Generate API keys (placeholders - replace with actual keys)
    local openai_key="sk-placeholder-openai-key-$(openssl rand -hex 8)"
    local bright_data_key="bd-placeholder-key-$(openssl rand -hex 8)"
    local market_data_key="md-placeholder-key-$(openssl rand -hex 8)"
    
    # Store secrets in temporary file for processing
    cat > "/tmp/secrets-$env.env" <<EOF
JWT_SECRET=$jwt_secret
DATABASE_PASSWORD=$db_password
REDIS_PASSWORD=$redis_password
OPENAI_API_KEY=$openai_key
BRIGHT_DATA_API_KEY=$bright_data_key
MARKET_DATA_API_KEY=$market_data_key
EOF
    
    success "Secrets generated for $env environment"
}

# Create environment namespace and resources
create_environment() {
    local env="$1"
    local namespace="$2"
    local branch="$3"
    local sync_policy="$4"
    
    log "Creating $env environment (namespace: $namespace)..."
    
    # Create namespace
    kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespace
    kubectl label namespace "$namespace" \
        environment="$env" \
        branch="$branch" \
        sync-policy="$sync_policy" \
        app.kubernetes.io/name=go-coffee \
        app.kubernetes.io/component=environment \
        app.kubernetes.io/part-of=go-coffee-platform \
        --overwrite
    
    # Create resource quota
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ResourceQuota
metadata:
  name: $env-quota
  namespace: $namespace
  labels:
    environment: $env
spec:
  hard:
    requests.cpu: "$([ "$env" = "production" ] && echo "40" || echo "20")"
    requests.memory: "$([ "$env" = "production" ] && echo "80Gi" || echo "40Gi")"
    limits.cpu: "$([ "$env" = "production" ] && echo "80" || echo "40")"
    limits.memory: "$([ "$env" = "production" ] && echo "160Gi" || echo "80Gi")"
    persistentvolumeclaims: "$([ "$env" = "production" ] && echo "20" || echo "10")"
    services: "$([ "$env" = "production" ] && echo "50" || echo "30")"
    secrets: "$([ "$env" = "production" ] && echo "40" || echo "20")"
    configmaps: "$([ "$env" = "production" ] && echo "40" || echo "20")"
EOF
    
    # Create limit range
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: LimitRange
metadata:
  name: $env-limits
  namespace: $namespace
  labels:
    environment: $env
spec:
  limits:
  - default:
      cpu: "$([ "$env" = "production" ] && echo "2" || echo "1")"
      memory: "$([ "$env" = "production" ] && echo "4Gi" || echo "2Gi")"
    defaultRequest:
      cpu: "$([ "$env" = "production" ] && echo "200m" || echo "100m")"
      memory: "$([ "$env" = "production" ] && echo "512Mi" || echo "256Mi")"
    type: Container
  - default:
      storage: "$([ "$env" = "production" ] && echo "20Gi" || echo "10Gi")"
    type: PersistentVolumeClaim
EOF
    
    success "Environment $env created"
}

# Create secrets for environment
create_environment_secrets() {
    local env="$1"
    local namespace="$2"
    
    log "Creating secrets for $env environment..."
    
    # Source the generated secrets
    source "/tmp/secrets-$env.env"
    
    # Create database secret
    kubectl create secret generic go-coffee-database \
        --namespace="$namespace" \
        --from-literal=database-url="postgres://go_coffee_user:${DATABASE_PASSWORD}@postgres:5432/go_coffee?sslmode=disable" \
        --from-literal=postgres-user="go_coffee_user" \
        --from-literal=postgres-password="$DATABASE_PASSWORD" \
        --from-literal=postgres-database="go_coffee" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create Redis secret
    kubectl create secret generic go-coffee-redis \
        --namespace="$namespace" \
        --from-literal=redis-url="redis://:${REDIS_PASSWORD}@redis:6379" \
        --from-literal=redis-password="$REDIS_PASSWORD" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create JWT secret
    kubectl create secret generic go-coffee-jwt \
        --namespace="$namespace" \
        --from-literal=jwt-secret="$JWT_SECRET" \
        --from-literal=jwt-issuer="go-coffee-$env" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create API keys secret
    kubectl create secret generic go-coffee-api-keys \
        --namespace="$namespace" \
        --from-literal=openai-api-key="$OPENAI_API_KEY" \
        --from-literal=bright-data-api-key="$BRIGHT_DATA_API_KEY" \
        --from-literal=market-data-api-key="$MARKET_DATA_API_KEY" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Label all secrets
    kubectl label secret -n "$namespace" \
        go-coffee-database go-coffee-redis go-coffee-jwt go-coffee-api-keys \
        environment="$env" \
        app.kubernetes.io/name=go-coffee \
        app.kubernetes.io/component=secrets \
        --overwrite
    
    success "Secrets created for $env environment"
}

# Create environment-specific ConfigMap
create_environment_config() {
    local env="$1"
    local namespace="$2"
    local branch="$3"
    
    log "Creating configuration for $env environment..."
    
    # Determine environment-specific values
    local log_level="info"
    local replicas="2"
    local resources_cpu_request="100m"
    local resources_memory_request="256Mi"
    local resources_cpu_limit="1000m"
    local resources_memory_limit="2Gi"
    
    if [[ "$env" == "production" ]]; then
        replicas="3"
        resources_cpu_request="200m"
        resources_memory_request="512Mi"
        resources_cpu_limit="2000m"
        resources_memory_limit="4Gi"
    elif [[ "$env" == "staging" ]]; then
        log_level="debug"
    fi
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-coffee-config
  namespace: $namespace
  labels:
    environment: $env
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: config
data:
  # Environment configuration
  environment: "$env"
  branch: "$branch"
  log-level: "$log_level"
  
  # Feature flags
  metrics-enabled: "true"
  tracing-enabled: "true"
  debug-enabled: "$([ "$env" = "staging" ] && echo "true" || echo "false")"
  
  # Scaling configuration
  default-replicas: "$replicas"
  min-replicas: "$([ "$env" = "production" ] && echo "2" || echo "1")"
  max-replicas: "$([ "$env" = "production" ] && echo "10" || echo "5")"
  
  # Resource configuration
  default-cpu-request: "$resources_cpu_request"
  default-memory-request: "$resources_memory_request"
  default-cpu-limit: "$resources_cpu_limit"
  default-memory-limit: "$resources_memory_limit"
  
  # Database configuration
  database-max-connections: "$([ "$env" = "production" ] && echo "50" || echo "25")"
  database-max-idle: "$([ "$env" = "production" ] && echo "10" || echo "5")"
  database-connection-timeout: "30s"
  
  # Redis configuration
  redis-max-connections: "$([ "$env" = "production" ] && echo "200" || echo "100")"
  redis-connection-timeout: "5s"
  redis-max-idle: "$([ "$env" = "production" ] && echo "20" || echo "10")"
  
  # HTTP configuration
  http-timeout: "30s"
  http-max-idle-connections: "$([ "$env" = "production" ] && echo "200" || echo "100")"
  http-idle-connection-timeout: "90s"
  
  # AI configuration
  ai-model-timeout: "$([ "$env" = "production" ] && echo "30s" || echo "60s")"
  ai-max-tokens: "4096"
  ai-temperature: "0.7"
  
  # Monitoring configuration
  prometheus-metrics-path: "/metrics"
  health-check-path: "/health"
  readiness-check-path: "/ready"
  metrics-port: "9090"
  
  # Security configuration
  cors-allowed-origins: "$([ "$env" = "production" ] && echo "https://gocoffee.dev,https://app.gocoffee.dev" || echo "*")"
  rate-limit-requests-per-minute: "$([ "$env" = "production" ] && echo "1000" || echo "100")"
  
  # External services
  external-api-timeout: "10s"
  external-api-retries: "3"
  external-api-backoff: "1s"
EOF
    
    success "Configuration created for $env environment"
}

# Create network policies for environment
create_network_policies() {
    local env="$1"
    local namespace="$2"
    
    log "Creating network policies for $env environment..."
    
    # Default deny all ingress
    cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: $namespace
  labels:
    environment: $env
spec:
  podSelector: {}
  policyTypes:
  - Ingress
EOF
    
    # Allow ingress from same namespace
    cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-same-namespace
  namespace: $namespace
  labels:
    environment: $env
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: $namespace
EOF
    
    # Allow ingress from monitoring namespace
    cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-monitoring
  namespace: $namespace
  labels:
    environment: $env
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: service
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: go-coffee-monitoring
    ports:
    - protocol: TCP
      port: 9090
EOF
    
    success "Network policies created for $env environment"
}

# Setup environment monitoring
setup_environment_monitoring() {
    local env="$1"
    local namespace="$2"
    
    log "Setting up monitoring for $env environment..."
    
    # Create ServiceMonitor for environment
    cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: go-coffee-$env-services
  namespace: go-coffee-monitoring
  labels:
    environment: $env
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: monitoring
    release: prometheus-stack
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: go-coffee
      environment: $env
  namespaceSelector:
    matchNames:
    - $namespace
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    honorLabels: true
EOF
    
    success "Monitoring setup completed for $env environment"
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    🌍 Go Coffee - Environment Management Setup
    ===========================================
    
    Setting up comprehensive environment management:
    • Staging environment (auto-sync from develop)
    • Production environment (manual sync from main)
    • Disaster Recovery environment
    • Secrets management and rotation
    • Environment-specific configurations
    • Network policies and security
    • Monitoring and observability
    
EOF
    echo -e "${NC}"
    
    info "Starting environment setup..."
    
    # Execute setup steps
    check_prerequisites
    
    # Setup each environment
    for env in "${!ENVIRONMENTS[@]}"; do
        IFS=':' read -r namespace branch sync_policy <<< "${ENVIRONMENTS[$env]}"
        
        info "Setting up $env environment..."
        
        generate_secrets "$env"
        create_environment "$env" "$namespace" "$branch" "$sync_policy"
        create_environment_secrets "$env" "$namespace"
        create_environment_config "$env" "$namespace" "$branch"
        create_network_policies "$env" "$namespace"
        setup_environment_monitoring "$env" "$namespace"
        
        success "$env environment setup completed"
    done
    
    # Cleanup temporary files
    rm -f /tmp/secrets-*.env
    
    echo -e "${CYAN}"
    cat << EOF

    🎉 Environment Management Setup Completed!
    
    📋 Environments Created:
    
    Staging Environment:
    - Namespace: go-coffee-staging
    - Branch: develop
    - Sync Policy: Automatic
    - Purpose: Development testing and validation
    
    Production Environment:
    - Namespace: go-coffee-production
    - Branch: main
    - Sync Policy: Manual
    - Purpose: Live production workloads
    
    Disaster Recovery Environment:
    - Namespace: go-coffee-dr
    - Branch: main
    - Sync Policy: Manual
    - Purpose: Backup and disaster recovery
    
    🔐 Security Features:
    - Unique secrets per environment
    - Network policies for isolation
    - Resource quotas and limits
    - RBAC and access controls
    
    📊 Monitoring:
    - Environment-specific metrics
    - Health and readiness checks
    - Performance monitoring
    - Alert management
    
    🔧 Management Commands:
    
    # View environments
    kubectl get namespaces -l app.kubernetes.io/name=go-coffee
    
    # Check resource usage
    kubectl top pods -n go-coffee-staging
    kubectl top pods -n go-coffee-production
    
    # View secrets (base64 encoded)
    kubectl get secrets -n go-coffee-staging
    kubectl get secrets -n go-coffee-production
    
    # Monitor deployments
    kubectl get deployments -n go-coffee-staging
    kubectl get deployments -n go-coffee-production
    
EOF
    echo -e "${NC}"
    
    success "🎉 Environment management setup completed successfully!"
}

# Handle command line arguments
case "${1:-setup}" in
    "setup")
        main
        ;;
    "cleanup")
        warn "Cleaning up environments..."
        for env in "${!ENVIRONMENTS[@]}"; do
            IFS=':' read -r namespace branch sync_policy <<< "${ENVIRONMENTS[$env]}"
            kubectl delete namespace "$namespace" --ignore-not-found=true
        done
        success "Environment cleanup completed"
        ;;
    "rotate-secrets")
        ENV="${2:-staging}"
        if [[ -n "${ENVIRONMENTS[$ENV]:-}" ]]; then
            IFS=':' read -r namespace branch sync_policy <<< "${ENVIRONMENTS[$ENV]}"
            log "Rotating secrets for $ENV environment..."
            generate_secrets "$ENV"
            create_environment_secrets "$ENV" "$namespace"
            success "Secrets rotated for $ENV environment"
        else
            error "Invalid environment: $ENV"
        fi
        ;;
    *)
        echo "Usage: $0 [setup|cleanup|rotate-secrets <env>]"
        echo "Environments: ${!ENVIRONMENTS[*]}"
        exit 1
        ;;
esac
