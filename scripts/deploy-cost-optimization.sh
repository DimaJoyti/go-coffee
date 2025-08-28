#!/bin/bash

# Cost Optimization Stack Deployment Script
# Deploys comprehensive cost optimization and resource management tools

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/cost-optimization"
COST_OPTIMIZATION_DIR="$PROJECT_ROOT/k8s/cost-optimization"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
COST_OPTIMIZATION_NAMESPACE="${COST_OPTIMIZATION_NAMESPACE:-cost-optimization}"
ENABLE_RIGHTSIZING="${ENABLE_RIGHTSIZING:-true}"
ENABLE_CLUSTER_AUTOSCALING="${ENABLE_CLUSTER_AUTOSCALING:-true}"
ENABLE_COST_ANALYZER="${ENABLE_COST_ANALYZER:-true}"
ENABLE_INTELLIGENT_SCHEDULING="${ENABLE_INTELLIGENT_SCHEDULING:-true}"
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
    
    # Check metrics server
    if ! kubectl get apiservice v1beta1.metrics.k8s.io &>/dev/null; then
        log_warning "Metrics server not found - some cost optimization features may not work"
    fi
    
    log_success "Prerequisites check completed"
}

# Setup cost optimization namespace
setup_namespace() {
    log_info "Setting up cost optimization namespace: $COST_OPTIMIZATION_NAMESPACE"
    
    if kubectl get namespace "$COST_OPTIMIZATION_NAMESPACE" &>/dev/null; then
        log_info "Namespace $COST_OPTIMIZATION_NAMESPACE already exists"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would create namespace $COST_OPTIMIZATION_NAMESPACE"
        else
            kubectl create namespace "$COST_OPTIMIZATION_NAMESPACE"
            
            # Apply labels and annotations
            kubectl label namespace "$COST_OPTIMIZATION_NAMESPACE" \
                app.kubernetes.io/name="$PROJECT_NAME" \
                app.kubernetes.io/component="cost-optimization" \
                environment="$ENVIRONMENT"
            
            kubectl annotate namespace "$COST_OPTIMIZATION_NAMESPACE" \
                managed-by="terraform" \
                created-at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
            
            log_success "Created cost optimization namespace: $COST_OPTIMIZATION_NAMESPACE"
        fi
    fi
}

# Add Helm repositories
add_helm_repositories() {
    log_info "Adding Helm repositories for cost optimization tools..."
    
    local repos=(
        "fairwinds-stable:https://charts.fairwinds.com/stable"
        "autoscaler:https://kubernetes.github.io/autoscaler"
        "kubecost:https://kubecost.github.io/cost-analyzer/"
        "metrics-server:https://kubernetes-sigs.github.io/metrics-server/"
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

# Deploy metrics server if not present
deploy_metrics_server() {
    log_info "Checking metrics server deployment..."
    
    if kubectl get deployment metrics-server -n kube-system &>/dev/null; then
        log_info "Metrics server already deployed"
        return 0
    fi
    
    log_info "Deploying metrics server..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy metrics server"
    else
        helm upgrade --install metrics-server metrics-server/metrics-server \
            --namespace kube-system \
            --set args[0]="--cert-dir=/tmp" \
            --set args[1]="--secure-port=4443" \
            --set args[2]="--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname" \
            --set args[3]="--kubelet-use-node-status-port" \
            --set args[4]="--metric-resolution=15s" \
            --wait \
            --timeout 5m
        
        log_success "Metrics server deployed successfully"
    fi
}

# Deploy VPA for rightsizing
deploy_vpa() {
    if [[ "$ENABLE_RIGHTSIZING" != "true" ]]; then
        log_info "VPA rightsizing disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Vertical Pod Autoscaler (VPA)..."
    
    local values_file="$COST_OPTIMIZATION_DIR/vpa-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$COST_OPTIMIZATION_DIR"
        log_info "Creating VPA values file..."
        cat > "$values_file" << EOF
recommender:
  enabled: true
  resources:
    requests:
      cpu: "100m"
      memory: "500Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
  extraArgs:
    - --v=4
    - --pod-recommendation-min-cpu-millicores=25
    - --pod-recommendation-min-memory-mb=100
    - --recommendation-margin-fraction=0.15
    - --target-cpu-percentile=0.9
    - --target-memory-percentile=0.9

updater:
  enabled: true
  resources:
    requests:
      cpu: "100m"
      memory: "500Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
  extraArgs:
    - --v=4
    - --min-replicas=2
    - --eviction-tolerance=0.5

admissionController:
  enabled: true
  resources:
    requests:
      cpu: "50m"
      memory: "200Mi"
    limits:
      cpu: "200m"
      memory: "500Mi"
  generateCertificate: true
  extraArgs:
    - --v=4
    - --webhook-timeout-seconds=30

metrics:
  serviceMonitor:
    enabled: ${MONITORING_ENABLED:-true}
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy VPA"
        helm template vpa fairwinds-stable/vpa \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: VPA template validation passed"
    else
        helm upgrade --install vpa fairwinds-stable/vpa \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "VPA deployed successfully"
    fi
}

# Deploy cluster autoscaler
deploy_cluster_autoscaler() {
    if [[ "$ENABLE_CLUSTER_AUTOSCALING" != "true" ]]; then
        log_info "Cluster autoscaling disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Cluster Autoscaler..."
    
    local values_file="$COST_OPTIMIZATION_DIR/cluster-autoscaler-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$COST_OPTIMIZATION_DIR"
        log_info "Creating Cluster Autoscaler values file..."
        cat > "$values_file" << EOF
autoDiscovery:
  clusterName: ${CLUSTER_NAME:-go-coffee-cluster}
  enabled: true

awsRegion: ${AWS_REGION:-us-east-1}

extraArgs:
  v: 4
  stderrthreshold: info
  cloud-provider: ${CLOUD_PROVIDER:-aws}
  skip-nodes-with-local-storage: false
  expander: least-waste
  balance-similar-node-groups: true
  skip-nodes-with-system-pods: false
  scale-down-enabled: true
  scale-down-delay-after-add: 10m
  scale-down-unneeded-time: 10m
  scale-down-utilization-threshold: 0.5
  max-node-provision-time: 15m
  scan-interval: 10s

resources:
  requests:
    cpu: "100m"
    memory: "300Mi"
  limits:
    cpu: "100m"
    memory: "300Mi"

rbac:
  create: true
  serviceAccount:
    create: true
    name: cluster-autoscaler

nodeSelector:
  node-type: cost-optimized

tolerations:
  - key: "node-type"
    operator: "Equal"
    value: "cost-optimized"
    effect: "NoSchedule"
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Cluster Autoscaler"
        helm template cluster-autoscaler autoscaler/cluster-autoscaler \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Cluster Autoscaler template validation passed"
    else
        helm upgrade --install cluster-autoscaler autoscaler/cluster-autoscaler \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Cluster Autoscaler deployed successfully"
    fi
}

# Deploy KubeCost
deploy_kubecost() {
    if [[ "$ENABLE_COST_ANALYZER" != "true" ]]; then
        log_info "Cost analyzer disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying KubeCost cost analyzer..."
    
    local values_file="$COST_OPTIMIZATION_DIR/kubecost-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$COST_OPTIMIZATION_DIR"
        log_info "Creating KubeCost values file..."
        cat > "$values_file" << EOF
global:
  prometheus:
    enabled: false
    fqdn: ${PROMETHEUS_FQDN:-prometheus.go-coffee.local}
  grafana:
    enabled: false
    fqdn: ${GRAFANA_FQDN:-grafana.go-coffee.local}

kubecostFrontend:
  image: "gcr.io/kubecost1/frontend"
  resources:
    requests:
      cpu: "10m"
      memory: "55Mi"
    limits:
      cpu: "100m"
      memory: "256Mi"

kubecostModel:
  image: "gcr.io/kubecost1/cost-model"
  resources:
    requests:
      cpu: "200m"
      memory: "55Mi"
    limits:
      cpu: "800m"
      memory: "256Mi"
  extraEnv:
    - name: CLUSTER_ID
      value: "${CLUSTER_NAME:-go-coffee-cluster}"
    - name: AWS_CLUSTER_ID
      value: "${CLUSTER_NAME:-go-coffee-cluster}"

ingress:
  enabled: ${KUBECOST_INGRESS_ENABLED:-true}
  className: ${INGRESS_CLASS_NAME:-nginx}
  hosts:
    - ${KUBECOST_HOSTNAME:-kubecost.go-coffee.local}
  tls:
    - secretName: kubecost-tls
      hosts:
        - ${KUBECOST_HOSTNAME:-kubecost.go-coffee.local}
  annotations:
    cert-manager.io/cluster-issuer: ${CERT_MANAGER_ISSUER:-letsencrypt-prod}

persistentVolume:
  enabled: true
  size: ${KUBECOST_STORAGE_SIZE:-32Gi}
  storageClass: ${STORAGE_CLASS_NAME:-gp2}

serviceMonitor:
  enabled: ${MONITORING_ENABLED:-true}

costOptimization:
  enabled: true
  recommendations:
    rightSizing:
      enabled: true
      cpuThreshold: ${CPU_THRESHOLD_UPPER:-80}
      memoryThreshold: ${MEMORY_THRESHOLD_UPPER:-85}
    clusterSizing:
      enabled: true
      evaluationPeriod: "${EVALUATION_PERIOD_DAYS:-7}d"
    abandonedResources:
      enabled: true
      threshold: "7d"
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy KubeCost"
        helm template kubecost kubecost/cost-analyzer \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: KubeCost template validation passed"
    else
        helm upgrade --install kubecost kubecost/cost-analyzer \
            --namespace "$COST_OPTIMIZATION_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 15m
        
        log_success "KubeCost deployed successfully"
    fi
}

# Deploy cost optimization policies
deploy_cost_policies() {
    log_info "Deploying cost optimization policies..."
    
    local policies_dir="$COST_OPTIMIZATION_DIR/policies"
    mkdir -p "$policies_dir"
    
    # Create resource quota policy
    cat > "$policies_dir/resource-quotas.yaml" << EOF
apiVersion: v1
kind: ResourceQuota
metadata:
  name: ${PROJECT_NAME}-resource-quota
  namespace: ${COST_OPTIMIZATION_NAMESPACE}
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    persistentvolumeclaims: "10"
    services: "10"
    secrets: "10"
    configmaps: "10"
---
apiVersion: v1
kind: LimitRange
metadata:
  name: ${PROJECT_NAME}-limit-range
  namespace: ${COST_OPTIMIZATION_NAMESPACE}
spec:
  limits:
  - default:
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:
      cpu: "100m"
      memory: "128Mi"
    type: Container
  - max:
      cpu: "2000m"
      memory: "4Gi"
    min:
      cpu: "50m"
      memory: "64Mi"
    type: Container
EOF
    
    # Create VPA policies for Go Coffee services
    cat > "$policies_dir/vpa-policies.yaml" << EOF
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: coffee-service-vpa
  namespace: go-coffee
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: coffee-service
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: coffee-service
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 1000m
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
---
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: ai-agent-vpa
  namespace: ai-agents
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: ai-agent-coordinator
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: ai-agent-coordinator
      minAllowed:
        cpu: 200m
        memory: 256Mi
      maxAllowed:
        cpu: 2000m
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy cost optimization policies"
    else
        kubectl apply -f "$policies_dir/" --recursive
        log_success "Cost optimization policies deployed"
    fi
}

# Create cost optimization dashboard
create_cost_dashboard() {
    log_info "Creating cost optimization dashboard..."
    
    local dashboard_dir="$COST_OPTIMIZATION_DIR/dashboards"
    mkdir -p "$dashboard_dir"
    
    # Create cost optimization dashboard ConfigMap
    cat > "$dashboard_dir/cost-optimization-dashboard.yaml" << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cost-optimization-dashboard
  namespace: ${COST_OPTIMIZATION_NAMESPACE}
  labels:
    grafana_dashboard: "1"
data:
  cost-optimization.json: |
    {
      "dashboard": {
        "title": "Go Coffee Cost Optimization",
        "panels": [
          {
            "title": "Monthly Cost Trend",
            "type": "graph",
            "targets": [
              {
                "expr": "kubecost_cluster_cost_per_month",
                "legendFormat": "Monthly Cost"
              }
            ]
          },
          {
            "title": "Cost by Namespace",
            "type": "piechart",
            "targets": [
              {
                "expr": "sum by (namespace) (kubecost_namespace_cost_per_hour)",
                "legendFormat": "{{namespace}}"
              }
            ]
          },
          {
            "title": "Resource Utilization",
            "type": "graph",
            "targets": [
              {
                "expr": "avg(rate(container_cpu_usage_seconds_total[5m])) * 100",
                "legendFormat": "CPU Utilization %"
              },
              {
                "expr": "avg(container_memory_usage_bytes / container_spec_memory_limit_bytes) * 100",
                "legendFormat": "Memory Utilization %"
              }
            ]
          }
        ]
      }
    }
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create cost optimization dashboard"
    else
        kubectl apply -f "$dashboard_dir/cost-optimization-dashboard.yaml"
        log_success "Cost optimization dashboard created"
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying cost optimization stack deployment..."
    
    local components=()
    
    if [[ "$ENABLE_RIGHTSIZING" == "true" ]]; then
        components+=("vpa-recommender")
        components+=("vpa-updater")
        components+=("vpa-admission-controller")
    fi
    
    if [[ "$ENABLE_CLUSTER_AUTOSCALING" == "true" ]]; then
        components+=("cluster-autoscaler")
    fi
    
    if [[ "$ENABLE_COST_ANALYZER" == "true" ]]; then
        components+=("kubecost-cost-analyzer")
    fi
    
    local failed_components=()
    
    for component in "${components[@]}"; do
        log_info "Checking component: $component"
        
        if kubectl get pods -n "$COST_OPTIMIZATION_NAMESPACE" -l "app.kubernetes.io/name=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        else
            log_error "Component $component is not running"
            failed_components+=("$component")
        fi
    done
    
    if [[ ${#failed_components[@]} -gt 0 ]]; then
        log_error "Some cost optimization components failed to deploy: ${failed_components[*]}"
        log_info "Check pod status with: kubectl get pods -n $COST_OPTIMIZATION_NAMESPACE"
        return 1
    fi
    
    log_success "All cost optimization components are running successfully"
}

# Display cost optimization information
display_cost_info() {
    log_info "Cost optimization stack information:"
    
    print_separator
    
    echo -e "${GREEN}Cost Optimization Tools Deployed:${NC}"
    
    if [[ "$ENABLE_RIGHTSIZING" == "true" ]]; then
        echo -e "  ✅ Vertical Pod Autoscaler (VPA)"
        echo -e "     - Automatic resource rightsizing"
        echo -e "     - CPU/Memory recommendations"
    fi
    
    if [[ "$ENABLE_CLUSTER_AUTOSCALING" == "true" ]]; then
        echo -e "  ✅ Cluster Autoscaler"
        echo -e "     - Automatic node scaling"
        echo -e "     - Cost-optimized node selection"
    fi
    
    if [[ "$ENABLE_COST_ANALYZER" == "true" ]]; then
        echo -e "  ✅ KubeCost Cost Analyzer"
        echo -e "     - Cost visibility and analytics"
        echo -e "     - Resource optimization recommendations"
        echo -e "     - URL: http://localhost:9090 (port-forward)"
        echo -e "     - Command: kubectl port-forward -n $COST_OPTIMIZATION_NAMESPACE svc/kubecost-cost-analyzer 9090:9090"
    fi
    
    print_separator
    
    echo -e "${YELLOW}Cost Optimization Features:${NC}"
    echo -e "  💰 Automated resource rightsizing"
    echo -e "  📊 Real-time cost monitoring"
    echo -e "  🎯 Intelligent workload placement"
    echo -e "  📈 Cost trend analysis"
    echo -e "  ⚡ Auto-scaling optimization"
    echo -e "  🔍 Abandoned resource detection"
    
    print_separator
    
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  View cost optimization pods: kubectl get pods -n $COST_OPTIMIZATION_NAMESPACE"
    echo -e "  View VPA recommendations: kubectl get vpa -A"
    echo -e "  View resource quotas: kubectl get resourcequota -A"
    echo -e "  Check cluster autoscaler logs: kubectl logs -n $COST_OPTIMIZATION_NAMESPACE -l app.kubernetes.io/name=cluster-autoscaler"
    echo -e "  Access KubeCost: kubectl port-forward -n $COST_OPTIMIZATION_NAMESPACE svc/kubecost-cost-analyzer 9090:9090"
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

Deploy comprehensive cost optimization stack for Go Coffee platform.

OPTIONS:
    --environment ENV               Environment (dev, staging, prod) [default: dev]
    --project NAME                  Project name [default: go-coffee]
    --namespace NAME                Cost optimization namespace [default: cost-optimization]
    --enable-rightsizing           Enable VPA rightsizing [default: true]
    --enable-cluster-autoscaling   Enable cluster autoscaling [default: true]
    --enable-cost-analyzer         Enable KubeCost analyzer [default: true]
    --enable-intelligent-scheduling Enable intelligent scheduling [default: true]
    --dry-run                      Perform dry run without actual deployment
    --verbose                      Enable verbose output
    --help                         Show this help message

EXAMPLES:
    $0                                    # Deploy full cost optimization stack
    $0 --environment prod                 # Deploy to production
    $0 --dry-run                         # Perform dry run
    $0 --enable-rightsizing --enable-cost-analyzer  # Deploy specific components

ENVIRONMENT VARIABLES:
    CLUSTER_NAME                   Kubernetes cluster name
    CLOUD_PROVIDER                 Cloud provider (aws, gcp, azure)
    AWS_REGION                     AWS region
    PROMETHEUS_FQDN               Prometheus FQDN
    GRAFANA_FQDN                  Grafana FQDN
    KUBECOST_HOSTNAME             KubeCost hostname
    MONITORING_ENABLED            Enable monitoring integration [default: true]

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
                COST_OPTIMIZATION_NAMESPACE="$2"
                shift 2
                ;;
            --enable-rightsizing)
                ENABLE_RIGHTSIZING="true"
                shift
                ;;
            --disable-rightsizing)
                ENABLE_RIGHTSIZING="false"
                shift
                ;;
            --enable-cluster-autoscaling)
                ENABLE_CLUSTER_AUTOSCALING="true"
                shift
                ;;
            --disable-cluster-autoscaling)
                ENABLE_CLUSTER_AUTOSCALING="false"
                shift
                ;;
            --enable-cost-analyzer)
                ENABLE_COST_ANALYZER="true"
                shift
                ;;
            --disable-cost-analyzer)
                ENABLE_COST_ANALYZER="false"
                shift
                ;;
            --enable-intelligent-scheduling)
                ENABLE_INTELLIGENT_SCHEDULING="true"
                shift
                ;;
            --disable-intelligent-scheduling)
                ENABLE_INTELLIGENT_SCHEDULING="false"
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
    
    print_header "💰 Go Coffee Cost Optimization Stack Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  Namespace: $COST_OPTIMIZATION_NAMESPACE"
    log_info "  Rightsizing: $ENABLE_RIGHTSIZING"
    log_info "  Cluster Autoscaling: $ENABLE_CLUSTER_AUTOSCALING"
    log_info "  Cost Analyzer: $ENABLE_COST_ANALYZER"
    log_info "  Intelligent Scheduling: $ENABLE_INTELLIGENT_SCHEDULING"
    log_info "  Dry Run: $DRY_RUN"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repositories
    deploy_metrics_server
    deploy_vpa
    deploy_cluster_autoscaler
    deploy_kubecost
    deploy_cost_policies
    create_cost_dashboard
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
        display_cost_info
    fi
    
    print_header "✅ Cost Optimization Stack Deployment Completed"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Cost optimization stack deployed successfully!"
        log_info "Monitor cost savings and optimization recommendations in KubeCost dashboard"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
