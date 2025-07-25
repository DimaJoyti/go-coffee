#!/bin/bash

# ☕ Go Coffee - Advanced Monitoring and Observability Deployment Script
# Deploys comprehensive monitoring stack with OpenTelemetry, Prometheus, Grafana, Jaeger, Loki, and Alertmanager

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
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-go-coffee-monitoring}"
ENVIRONMENT="${ENVIRONMENT:-production}"
CLUSTER_NAME="${CLUSTER_NAME:-go-coffee-cluster}"

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
    local tools=("kubectl" "helm")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is not installed or not in PATH"
        fi
    done
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl is not properly configured or cluster is not accessible"
    fi
    
    # Check Helm version
    local helm_version=$(helm version --short | cut -d'+' -f1 | cut -d'v' -f2)
    if [[ $(echo "$helm_version 3.10.0" | tr " " "\n" | sort -V | head -n1) != "3.10.0" ]]; then
        error "Helm version 3.10.0 or higher is required. Current: $helm_version"
    fi
    
    success "All prerequisites met"
}

# Create namespace and RBAC
setup_namespace() {
    log "Setting up monitoring namespace and RBAC..."
    
    # Create namespace
    kubectl create namespace "$MONITORING_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespace
    kubectl label namespace "$MONITORING_NAMESPACE" \
        name="$MONITORING_NAMESPACE" \
        monitoring=enabled \
        --overwrite
    
    # Create service accounts
    local service_accounts=("prometheus" "grafana" "jaeger" "loki" "fluent-bit" "otel-collector" "alertmanager")
    for sa in "${service_accounts[@]}"; do
        kubectl create serviceaccount "$sa" -n "$MONITORING_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    done
    
    # Create cluster roles and bindings
    cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: go-coffee-monitoring
rules:
- apiGroups: [""]
  resources: ["nodes", "nodes/proxy", "services", "endpoints", "pods", "ingresses", "configmaps"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions", "networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: go-coffee-monitoring
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: go-coffee-monitoring
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: $MONITORING_NAMESPACE
- kind: ServiceAccount
  name: otel-collector
  namespace: $MONITORING_NAMESPACE
- kind: ServiceAccount
  name: fluent-bit
  namespace: $MONITORING_NAMESPACE
EOF
    
    success "Namespace and RBAC configured"
}

# Add Helm repositories
add_helm_repos() {
    log "Adding Helm repositories..."
    
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
    helm repo add elastic https://helm.elastic.co
    helm repo add fluent https://fluent.github.io/helm-charts
    helm repo update
    
    success "Helm repositories added"
}

# Deploy OpenTelemetry Collector
deploy_otel_collector() {
    log "Deploying OpenTelemetry Collector..."
    
    kubectl apply -f "$SCRIPT_DIR/opentelemetry/otel-collector.yaml"
    
    # Wait for deployment
    kubectl wait --for=condition=available --timeout=300s deployment/otel-collector -n "$MONITORING_NAMESPACE"
    
    success "OpenTelemetry Collector deployed"
}

# Deploy Prometheus
deploy_prometheus() {
    log "Deploying Prometheus..."
    
    # Apply Prometheus configuration
    kubectl apply -f "$SCRIPT_DIR/prometheus/prometheus-config.yaml"
    
    # Deploy Prometheus using Helm
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace "$MONITORING_NAMESPACE" \
        --values - <<EOF
prometheus:
  prometheusSpec:
    configMaps:
      - prometheus-config
    retention: 30d
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 100Gi
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 8Gi
    serviceMonitorSelectorNilUsesHelmValues: false
    ruleSelectorNilUsesHelmValues: false
    
alertmanager:
  enabled: false  # We'll deploy our custom Alertmanager
  
grafana:
  enabled: false  # We'll deploy Grafana separately

nodeExporter:
  enabled: true

kubeStateMetrics:
  enabled: true

defaultRules:
  create: true
  rules:
    alertmanager: true
    etcd: true
    configReloaders: true
    general: true
    k8s: true
    kubeApiserverAvailability: true
    kubeApiserverBurnrate: true
    kubeApiserverHistogram: true
    kubeApiserverSlos: true
    kubelet: true
    kubeProxy: true
    kubePrometheusGeneral: true
    kubePrometheusNodeRecording: true
    kubernetesApps: true
    kubernetesResources: true
    kubernetesStorage: true
    kubernetesSystem: true
    network: true
    node: true
    nodeExporterAlerting: true
    nodeExporterRecording: true
    prometheus: true
    prometheusOperator: true
EOF
    
    # Wait for Prometheus to be ready
    kubectl wait --for=condition=available --timeout=600s deployment/prometheus-kube-prometheus-prometheus-operator -n "$MONITORING_NAMESPACE"
    
    success "Prometheus deployed"
}

# Deploy Alertmanager
deploy_alertmanager() {
    log "Deploying Alertmanager..."
    
    kubectl apply -f "$SCRIPT_DIR/alertmanager/alertmanager-config.yaml"
    
    # Wait for deployment
    kubectl wait --for=condition=available --timeout=300s deployment/alertmanager -n "$MONITORING_NAMESPACE"
    
    success "Alertmanager deployed"
}

# Deploy Jaeger
deploy_jaeger() {
    log "Deploying Jaeger..."
    
    # Deploy Elasticsearch for Jaeger storage
    helm upgrade --install elasticsearch elastic/elasticsearch \
        --namespace "$MONITORING_NAMESPACE" \
        --values - <<EOF
replicas: 1
minimumMasterNodes: 1
esConfig:
  elasticsearch.yml: |
    cluster.name: "jaeger"
    network.host: "0.0.0.0"
    discovery.type: single-node
    xpack.security.enabled: false
    xpack.monitoring.enabled: false
volumeClaimTemplate:
  accessModes: ["ReadWriteOnce"]
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 50Gi
resources:
  requests:
    cpu: 500m
    memory: 2Gi
  limits:
    cpu: 1000m
    memory: 4Gi
EOF
    
    # Wait for Elasticsearch
    kubectl wait --for=condition=ready --timeout=600s pod -l app=elasticsearch-master -n "$MONITORING_NAMESPACE"
    
    # Deploy Jaeger
    kubectl apply -f "$SCRIPT_DIR/jaeger/jaeger-deployment.yaml"
    
    # Wait for Jaeger components
    kubectl wait --for=condition=available --timeout=300s deployment/jaeger-collector -n "$MONITORING_NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/jaeger-query -n "$MONITORING_NAMESPACE"
    
    success "Jaeger deployed"
}

# Deploy Loki
deploy_loki() {
    log "Deploying Loki..."
    
    kubectl apply -f "$SCRIPT_DIR/loki/loki-config.yaml"
    
    # Wait for Loki
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=loki -n "$MONITORING_NAMESPACE"
    
    success "Loki deployed"
}

# Deploy Grafana
deploy_grafana() {
    log "Deploying Grafana..."
    
    # Create Grafana configuration
    kubectl create configmap grafana-dashboards \
        --from-file="$SCRIPT_DIR/grafana/dashboards/" \
        -n "$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Deploy Grafana
    helm upgrade --install grafana grafana/grafana \
        --namespace "$MONITORING_NAMESPACE" \
        --values - <<EOF
adminPassword: admin123  # Change this in production
persistence:
  enabled: true
  storageClassName: fast-ssd
  size: 10Gi

datasources:
  datasources.yaml:
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus-kube-prometheus-prometheus:9090
      access: proxy
      isDefault: true
    - name: Loki
      type: loki
      url: http://loki:3100
      access: proxy
    - name: Jaeger
      type: jaeger
      url: http://jaeger-query:16686
      access: proxy

dashboardProviders:
  dashboardproviders.yaml:
    apiVersion: 1
    providers:
    - name: 'go-coffee'
      orgId: 1
      folder: 'Go Coffee'
      type: file
      disableDeletion: false
      editable: true
      options:
        path: /var/lib/grafana/dashboards/go-coffee

dashboardsConfigMaps:
  go-coffee: grafana-dashboards

resources:
  requests:
    cpu: 200m
    memory: 512Mi
  limits:
    cpu: 500m
    memory: 1Gi

service:
  type: ClusterIP
  port: 80

ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - grafana.gocoffee.dev
  tls:
    - secretName: grafana-tls
      hosts:
        - grafana.gocoffee.dev
EOF
    
    # Wait for Grafana
    kubectl wait --for=condition=available --timeout=300s deployment/grafana -n "$MONITORING_NAMESPACE"
    
    success "Grafana deployed"
}

# Configure service monitors
configure_service_monitors() {
    log "Configuring service monitors..."
    
    # Create service monitor for Go Coffee services
    cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: go-coffee-services
  namespace: $MONITORING_NAMESPACE
  labels:
    app.kubernetes.io/name: go-coffee
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: go-coffee
  namespaceSelector:
    matchNames:
    - go-coffee
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    honorLabels: true
EOF
    
    success "Service monitors configured"
}

# Verify deployment
verify_deployment() {
    log "Verifying monitoring stack deployment..."
    
    # Check all pods are running
    info "Checking pod status..."
    kubectl get pods -n "$MONITORING_NAMESPACE" -o wide
    
    # Check services
    info "Checking services..."
    kubectl get services -n "$MONITORING_NAMESPACE"
    
    # Check ingresses
    if kubectl get ingress -n "$MONITORING_NAMESPACE" &> /dev/null; then
        info "Checking ingresses..."
        kubectl get ingress -n "$MONITORING_NAMESPACE"
    fi
    
    # Health checks
    info "Performing health checks..."
    
    # Prometheus health check
    if kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n "$MONITORING_NAMESPACE" &
    then
        local pf_pid=$!
        sleep 5
        if curl -f http://localhost:9090/-/healthy &> /dev/null; then
            success "Prometheus health check passed"
        else
            warn "Prometheus health check failed"
        fi
        kill $pf_pid 2>/dev/null || true
    fi
    
    # Grafana health check
    if kubectl port-forward svc/grafana 3000:80 -n "$MONITORING_NAMESPACE" &
    then
        local pf_pid=$!
        sleep 5
        if curl -f http://localhost:3000/api/health &> /dev/null; then
            success "Grafana health check passed"
        else
            warn "Grafana health check failed"
        fi
        kill $pf_pid 2>/dev/null || true
    fi
    
    success "Monitoring stack verification completed"
}

# Display access information
display_access_info() {
    echo -e "${CYAN}"
    cat << EOF

    🎉 Go Coffee Monitoring Stack Deployed Successfully!
    
    📊 Access Points:
    
    Grafana Dashboard:
    - URL: https://grafana.gocoffee.dev (if ingress configured)
    - Local: kubectl port-forward svc/grafana 3000:80 -n $MONITORING_NAMESPACE
    - Username: admin
    - Password: admin123 (change in production!)
    
    Prometheus:
    - Local: kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n $MONITORING_NAMESPACE
    
    Jaeger UI:
    - Local: kubectl port-forward svc/jaeger-query 16686:16686 -n $MONITORING_NAMESPACE
    
    Alertmanager:
    - Local: kubectl port-forward svc/alertmanager 9093:9093 -n $MONITORING_NAMESPACE
    
    📈 Key Features Deployed:
    ✅ OpenTelemetry Collector for unified observability
    ✅ Prometheus for metrics collection and alerting
    ✅ Grafana for visualization and dashboards
    ✅ Jaeger for distributed tracing
    ✅ Loki for log aggregation
    ✅ Alertmanager for alert routing and notifications
    ✅ Fluent Bit for log collection
    ✅ Custom dashboards for Go Coffee business metrics
    ✅ Comprehensive alerting rules
    
    🔧 Next Steps:
    1. Configure DNS for ingress endpoints
    2. Set up SSL certificates
    3. Configure Slack/email notifications in Alertmanager
    4. Customize dashboards for your specific needs
    5. Set up log retention policies
    
    📚 Documentation: https://docs.gocoffee.dev/monitoring
    
EOF
    echo -e "${NC}"
}

# Cleanup function
cleanup() {
    if [[ "${1:-}" == "destroy" ]]; then
        warn "Destroying monitoring stack..."
        
        # Delete Helm releases
        helm list -n "$MONITORING_NAMESPACE" -o json | jq -r '.[].name' | while read release; do
            helm uninstall "$release" -n "$MONITORING_NAMESPACE"
        done
        
        # Delete Kubernetes resources
        kubectl delete namespace "$MONITORING_NAMESPACE" --ignore-not-found=true
        
        # Delete cluster roles
        kubectl delete clusterrole go-coffee-monitoring --ignore-not-found=true
        kubectl delete clusterrolebinding go-coffee-monitoring --ignore-not-found=true
        
        success "Monitoring stack destroyed"
    fi
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    ☕ Go Coffee - Advanced Monitoring & Observability
    =================================================
    
    Deploying comprehensive monitoring stack:
    • OpenTelemetry Collector
    • Prometheus & Alertmanager
    • Grafana with custom dashboards
    • Jaeger for distributed tracing
    • Loki for log aggregation
    • Fluent Bit for log collection
    
EOF
    echo -e "${NC}"
    
    info "Starting monitoring stack deployment..."
    info "Namespace: $MONITORING_NAMESPACE"
    info "Environment: $ENVIRONMENT"
    info "Cluster: $CLUSTER_NAME"
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repos
    deploy_otel_collector
    deploy_prometheus
    deploy_alertmanager
    deploy_jaeger
    deploy_loki
    deploy_grafana
    configure_service_monitors
    verify_deployment
    display_access_info
    
    success "🎉 Advanced monitoring and observability stack deployment completed!"
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
