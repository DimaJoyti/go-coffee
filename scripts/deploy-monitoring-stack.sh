#!/bin/bash

# Multi-Cloud Monitoring Stack Deployment Script
# Deploys comprehensive monitoring and observability across multiple cloud providers

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/unified-monitoring"
MONITORING_DIR="$PROJECT_ROOT/k8s/monitoring"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
ENABLE_PROMETHEUS="${ENABLE_PROMETHEUS:-true}"
ENABLE_GRAFANA="${ENABLE_GRAFANA:-true}"
ENABLE_LOKI="${ENABLE_LOKI:-true}"
ENABLE_JAEGER="${ENABLE_JAEGER:-true}"
ENABLE_ALERTMANAGER="${ENABLE_ALERTMANAGER:-true}"
DRY_RUN="${DRY_RUN:-false}"
VERBOSE="${VERBOSE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        echo -e "${PURPLE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    fi
}

print_header() {
    echo -e "\n${CYAN}================================${NC}"
    echo -e "${CYAN} $1${NC}"
    echo -e "${CYAN}================================${NC}\n"
}

print_separator() {
    echo -e "${CYAN}--------------------------------${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local required_tools=("kubectl" "helm" "terraform" "jq" "curl")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command_exists "$tool"; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and try again."
        exit 1
    fi
    
    # Check Kubernetes connection
    if ! kubectl cluster-info &>/dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        log_info "Please ensure kubectl is configured and cluster is accessible"
        exit 1
    fi
    
    # Check Helm
    if ! helm version &>/dev/null; then
        log_error "Helm is not properly configured"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Setup monitoring namespace
setup_namespace() {
    log_info "Setting up monitoring namespace: $MONITORING_NAMESPACE"
    
    if kubectl get namespace "$MONITORING_NAMESPACE" &>/dev/null; then
        log_info "Namespace $MONITORING_NAMESPACE already exists"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would create namespace $MONITORING_NAMESPACE"
        else
            kubectl create namespace "$MONITORING_NAMESPACE"
            kubectl label namespace "$MONITORING_NAMESPACE" \
                app.kubernetes.io/name="$PROJECT_NAME" \
                app.kubernetes.io/component="monitoring" \
                environment="$ENVIRONMENT"
            log_success "Created namespace: $MONITORING_NAMESPACE"
        fi
    fi
}

# Add Helm repositories
add_helm_repositories() {
    log_info "Adding Helm repositories..."
    
    local repos=(
        "prometheus-community:https://prometheus-community.github.io/helm-charts"
        "grafana:https://grafana.github.io/helm-charts"
        "jaegertracing:https://jaegertracing.github.io/helm-charts"
        "elastic:https://helm.elastic.co"
    )
    
    for repo in "${repos[@]}"; do
        local name="${repo%%:*}"
        local url="${repo##*:}"
        
        log_info "Adding Helm repository: $name"
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would add Helm repo $name ($url)"
        else
            helm repo add "$name" "$url" || log_warning "Failed to add repo $name"
        fi
    done
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_info "Updating Helm repositories..."
        helm repo update
    fi
    
    log_success "Helm repositories configured"
}

# Create monitoring secrets
create_secrets() {
    log_info "Creating monitoring secrets..."
    
    # Grafana admin password
    local grafana_password="${GRAFANA_ADMIN_PASSWORD:-$(openssl rand -base64 32)}"
    
    # Basic auth for Prometheus/AlertManager
    local basic_auth_user="${BASIC_AUTH_USER:-admin}"
    local basic_auth_password="${BASIC_AUTH_PASSWORD:-$(openssl rand -base64 32)}"
    local basic_auth_hash=$(echo -n "$basic_auth_password" | openssl passwd -apr1 -stdin)
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create monitoring secrets"
        return 0
    fi
    
    # Create Grafana admin secret
    kubectl create secret generic grafana-admin-secret \
        --from-literal=admin-password="$grafana_password" \
        --namespace="$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create basic auth secrets
    kubectl create secret generic prometheus-basic-auth \
        --from-literal=auth="$basic_auth_user:$basic_auth_hash" \
        --namespace="$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic alertmanager-basic-auth \
        --from-literal=auth="$basic_auth_user:$basic_auth_hash" \
        --namespace="$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create notification secrets
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        kubectl create secret generic notification-secrets \
            --from-literal=slack-webhook-url="$SLACK_WEBHOOK_URL" \
            --from-literal=email-username="${EMAIL_USERNAME:-}" \
            --from-literal=email-password="${EMAIL_PASSWORD:-}" \
            --namespace="$MONITORING_NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    # Store credentials for user reference
    cat > "$PROJECT_ROOT/.monitoring-credentials" << EOF
# Monitoring Stack Credentials
# Generated: $(date)

GRAFANA_ADMIN_PASSWORD="$grafana_password"
BASIC_AUTH_USER="$basic_auth_user"
BASIC_AUTH_PASSWORD="$basic_auth_password"

# Access URLs (update with your actual hostnames)
GRAFANA_URL="https://${GRAFANA_HOSTNAME:-grafana.local}"
PROMETHEUS_URL="https://${PROMETHEUS_HOSTNAME:-prometheus.local}"
ALERTMANAGER_URL="https://${ALERTMANAGER_HOSTNAME:-alertmanager.local}"
JAEGER_URL="https://${JAEGER_HOSTNAME:-jaeger.local}"
EOF
    
    chmod 600 "$PROJECT_ROOT/.monitoring-credentials"
    
    log_success "Monitoring secrets created"
    log_info "Credentials saved to: $PROJECT_ROOT/.monitoring-credentials"
}

# Deploy Prometheus stack
deploy_prometheus_stack() {
    if [[ "$ENABLE_PROMETHEUS" != "true" ]]; then
        log_info "Prometheus stack disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Prometheus stack..."
    
    local values_file="$MONITORING_DIR/prometheus-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        log_info "Creating Prometheus values file..."
        cat > "$values_file" << EOF
prometheus:
  prometheusSpec:
    retention: 15d
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: ${STORAGE_CLASS_NAME:-standard}
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: ${PROMETHEUS_STORAGE_SIZE:-50Gi}

grafana:
  adminPassword: ${GRAFANA_ADMIN_PASSWORD:-admin}
  persistence:
    enabled: true
    storageClassName: ${STORAGE_CLASS_NAME:-standard}
    size: ${GRAFANA_STORAGE_SIZE:-10Gi}

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: ${STORAGE_CLASS_NAME:-standard}
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: ${ALERTMANAGER_STORAGE_SIZE:-10Gi}
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Prometheus stack"
        helm template prometheus-stack prometheus-community/kube-prometheus-stack \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Prometheus stack template validation passed"
    else
        helm upgrade --install prometheus-stack prometheus-community/kube-prometheus-stack \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Prometheus stack deployed successfully"
    fi
}

# Deploy Loki stack
deploy_loki_stack() {
    if [[ "$ENABLE_LOKI" != "true" ]]; then
        log_info "Loki stack disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Loki stack..."
    
    local values_file="$MONITORING_DIR/loki-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        log_info "Creating Loki values file..."
        cat > "$values_file" << EOF
loki:
  persistence:
    enabled: true
    storageClassName: ${STORAGE_CLASS_NAME:-standard}
    size: ${LOKI_STORAGE_SIZE:-20Gi}

promtail:
  enabled: true

grafana:
  enabled: false
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Loki stack"
        helm template loki grafana/loki-stack \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Loki stack template validation passed"
    else
        helm upgrade --install loki grafana/loki-stack \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Loki stack deployed successfully"
    fi
}

# Deploy Jaeger
deploy_jaeger() {
    if [[ "$ENABLE_JAEGER" != "true" ]]; then
        log_info "Jaeger disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Jaeger..."
    
    local values_file="$MONITORING_DIR/jaeger-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        log_info "Creating Jaeger values file..."
        cat > "$values_file" << EOF
allInOne:
  enabled: true
  
storage:
  type: memory

agent:
  enabled: true
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Jaeger"
        helm template jaeger jaegertracing/jaeger \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Jaeger template validation passed"
    else
        helm upgrade --install jaeger jaegertracing/jaeger \
            --namespace "$MONITORING_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Jaeger deployed successfully"
    fi
}

# Deploy custom dashboards
deploy_custom_dashboards() {
    log_info "Deploying custom dashboards..."
    
    local dashboards_dir="$MONITORING_DIR/custom-dashboards"
    
    if [[ ! -d "$dashboards_dir" ]]; then
        log_warning "Custom dashboards directory not found: $dashboards_dir"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy custom dashboards"
        return 0
    fi
    
    # Create ConfigMap for custom dashboards
    kubectl create configmap custom-dashboards \
        --from-file="$dashboards_dir" \
        --namespace="$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Custom dashboards deployed"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying monitoring stack deployment..."
    
    local components=()
    
    if [[ "$ENABLE_PROMETHEUS" == "true" ]]; then
        components+=("prometheus-stack-kube-prom-prometheus")
        components+=("prometheus-stack-grafana")
    fi
    
    if [[ "$ENABLE_LOKI" == "true" ]]; then
        components+=("loki")
    fi
    
    if [[ "$ENABLE_JAEGER" == "true" ]]; then
        components+=("jaeger")
    fi
    
    local failed_components=()
    
    for component in "${components[@]}"; do
        log_info "Checking component: $component"
        
        if kubectl get pods -n "$MONITORING_NAMESPACE" -l "app.kubernetes.io/name=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        else
            log_error "Component $component is not running"
            failed_components+=("$component")
        fi
    done
    
    if [[ ${#failed_components[@]} -gt 0 ]]; then
        log_error "Some components failed to deploy: ${failed_components[*]}"
        log_info "Check pod status with: kubectl get pods -n $MONITORING_NAMESPACE"
        return 1
    fi
    
    log_success "All monitoring components are running successfully"
}

# Display access information
display_access_info() {
    log_info "Monitoring stack access information:"
    
    print_separator
    
    if [[ "$ENABLE_GRAFANA" == "true" ]]; then
        echo -e "${GREEN}Grafana Dashboard:${NC}"
        echo -e "  URL: http://localhost:3000 (port-forward)"
        echo -e "  Command: kubectl port-forward -n $MONITORING_NAMESPACE svc/prometheus-stack-grafana 3000:80"
        echo -e "  Username: admin"
        echo -e "  Password: Check .monitoring-credentials file"
        echo
    fi
    
    if [[ "$ENABLE_PROMETHEUS" == "true" ]]; then
        echo -e "${GREEN}Prometheus:${NC}"
        echo -e "  URL: http://localhost:9090 (port-forward)"
        echo -e "  Command: kubectl port-forward -n $MONITORING_NAMESPACE svc/prometheus-stack-kube-prom-prometheus 9090:9090"
        echo
    fi
    
    if [[ "$ENABLE_ALERTMANAGER" == "true" ]]; then
        echo -e "${GREEN}AlertManager:${NC}"
        echo -e "  URL: http://localhost:9093 (port-forward)"
        echo -e "  Command: kubectl port-forward -n $MONITORING_NAMESPACE svc/prometheus-stack-kube-prom-alertmanager 9093:9093"
        echo
    fi
    
    if [[ "$ENABLE_JAEGER" == "true" ]]; then
        echo -e "${GREEN}Jaeger UI:${NC}"
        echo -e "  URL: http://localhost:16686 (port-forward)"
        echo -e "  Command: kubectl port-forward -n $MONITORING_NAMESPACE svc/jaeger-query 16686:16686"
        echo
    fi
    
    print_separator
    
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  View all monitoring pods: kubectl get pods -n $MONITORING_NAMESPACE"
    echo -e "  View monitoring services: kubectl get svc -n $MONITORING_NAMESPACE"
    echo -e "  View monitoring ingresses: kubectl get ingress -n $MONITORING_NAMESPACE"
    echo -e "  Check logs: kubectl logs -n $MONITORING_NAMESPACE -l app.kubernetes.io/name=<component>"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Add any cleanup logic here
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy comprehensive monitoring stack for Go Coffee platform.

OPTIONS:
    --environment ENV           Environment (dev, staging, prod) [default: dev]
    --project NAME              Project name [default: go-coffee]
    --namespace NAME            Monitoring namespace [default: monitoring]
    --enable-prometheus         Enable Prometheus stack [default: true]
    --enable-grafana           Enable Grafana [default: true]
    --enable-loki              Enable Loki logging [default: true]
    --enable-jaeger            Enable Jaeger tracing [default: true]
    --enable-alertmanager      Enable AlertManager [default: true]
    --dry-run                  Perform dry run without actual deployment
    --verbose                  Enable verbose output
    --help                     Show this help message

EXAMPLES:
    $0                                    # Deploy full monitoring stack
    $0 --environment prod                 # Deploy to production
    $0 --dry-run                         # Perform dry run
    $0 --enable-prometheus --enable-grafana  # Deploy only Prometheus and Grafana

ENVIRONMENT VARIABLES:
    GRAFANA_ADMIN_PASSWORD      Grafana admin password
    SLACK_WEBHOOK_URL          Slack webhook for alerts
    EMAIL_USERNAME             Email username for notifications
    EMAIL_PASSWORD             Email password for notifications
    STORAGE_CLASS_NAME         Kubernetes storage class
    PROMETHEUS_STORAGE_SIZE    Prometheus storage size [default: 50Gi]
    GRAFANA_STORAGE_SIZE       Grafana storage size [default: 10Gi]
    LOKI_STORAGE_SIZE          Loki storage size [default: 20Gi]

EOF
}

# Main execution function
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --project)
                PROJECT_NAME="$2"
                shift 2
                ;;
            --namespace)
                MONITORING_NAMESPACE="$2"
                shift 2
                ;;
            --enable-prometheus)
                ENABLE_PROMETHEUS="true"
                shift
                ;;
            --disable-prometheus)
                ENABLE_PROMETHEUS="false"
                shift
                ;;
            --enable-grafana)
                ENABLE_GRAFANA="true"
                shift
                ;;
            --disable-grafana)
                ENABLE_GRAFANA="false"
                shift
                ;;
            --enable-loki)
                ENABLE_LOKI="true"
                shift
                ;;
            --disable-loki)
                ENABLE_LOKI="false"
                shift
                ;;
            --enable-jaeger)
                ENABLE_JAEGER="true"
                shift
                ;;
            --disable-jaeger)
                ENABLE_JAEGER="false"
                shift
                ;;
            --enable-alertmanager)
                ENABLE_ALERTMANAGER="true"
                shift
                ;;
            --disable-alertmanager)
                ENABLE_ALERTMANAGER="false"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --verbose)
                VERBOSE="true"
                DEBUG="true"
                set -x
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_header "📊 Go Coffee Monitoring Stack Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  Namespace: $MONITORING_NAMESPACE"
    log_info "  Prometheus: $ENABLE_PROMETHEUS"
    log_info "  Grafana: $ENABLE_GRAFANA"
    log_info "  Loki: $ENABLE_LOKI"
    log_info "  Jaeger: $ENABLE_JAEGER"
    log_info "  AlertManager: $ENABLE_ALERTMANAGER"
    log_info "  Dry Run: $DRY_RUN"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repositories
    create_secrets
    deploy_prometheus_stack
    deploy_loki_stack
    deploy_jaeger
    deploy_custom_dashboards
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
        display_access_info
    fi
    
    print_header "✅ Monitoring Stack Deployment Completed"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Monitoring stack deployed successfully!"
        log_info "Check .monitoring-credentials file for access credentials"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
