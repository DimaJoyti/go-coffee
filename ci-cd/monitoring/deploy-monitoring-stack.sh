#!/bin/bash

# Deploy comprehensive monitoring and observability stack for Go Coffee CI/CD
# Includes Prometheus, Grafana, AlertManager, Jaeger, and custom dashboards

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
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-go-coffee-monitoring}"
ENVIRONMENT="${ENVIRONMENT:-production}"

# Monitoring stack versions
PROMETHEUS_VERSION="${PROMETHEUS_VERSION:-v2.47.0}"
GRAFANA_VERSION="${GRAFANA_VERSION:-10.1.0}"
ALERTMANAGER_VERSION="${ALERTMANAGER_VERSION:-v0.26.0}"
JAEGER_VERSION="${JAEGER_VERSION:-1.49.0}"

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
    log "Checking monitoring prerequisites..."
    
    # Check required tools
    local tools=("kubectl" "helm" "curl" "jq")
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

# Setup monitoring namespace
setup_monitoring_namespace() {
    log "Setting up monitoring namespace..."
    
    kubectl create namespace "$MONITORING_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespace
    kubectl label namespace "$MONITORING_NAMESPACE" \
        name="$MONITORING_NAMESPACE" \
        app.kubernetes.io/name=monitoring \
        app.kubernetes.io/component=observability \
        environment="$ENVIRONMENT" \
        --overwrite
    
    success "Monitoring namespace configured"
}

# Add Helm repositories
add_helm_repos() {
    log "Adding Helm repositories for monitoring tools..."
    
    # Prometheus Community
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    
    # Grafana
    helm repo add grafana https://grafana.github.io/helm-charts
    
    # Jaeger
    helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
    
    # Update repositories
    helm repo update
    
    success "Helm repositories added"
}

# Deploy Prometheus Stack
deploy_prometheus_stack() {
    log "Deploying Prometheus monitoring stack..."
    
    # Create Prometheus values file
    cat > /tmp/prometheus-values.yaml <<EOF
prometheus:
  prometheusSpec:
    retention: 30d
    retentionSize: 50GB
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: fast-ssd
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 100Gi
    
    additionalScrapeConfigs:
    - job_name: 'go-coffee-services'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - go-coffee-staging
          - go-coffee-production
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: \$1:\$2
        target_label: __address__
    
    - job_name: 'argocd-metrics'
      kubernetes_sd_configs:
      - role: service
        namespaces:
          names:
          - argocd
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
        action: keep
        regex: argocd-metrics
    
    - job_name: 'github-actions-metrics'
      static_configs:
      - targets: ['api.github.com:443']
      metrics_path: '/repos/DimaJoyti/go-coffee/actions/runs'
      scheme: https
      bearer_token_file: /etc/prometheus/secrets/github-token

grafana:
  enabled: true
  adminPassword: "${GRAFANA_ADMIN_PASSWORD:-admin123}"
  
  persistence:
    enabled: true
    size: 10Gi
    storageClassName: fast-ssd
  
  datasources:
    datasources.yaml:
      apiVersion: 1
      datasources:
      - name: Prometheus
        type: prometheus
        url: http://prometheus-kube-prometheus-prometheus:9090
        access: proxy
        isDefault: true
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
  
  dashboards:
    go-coffee:
      go-coffee-overview:
        gnetId: 15757
        revision: 1
        datasource: Prometheus
      go-coffee-cicd:
        gnetId: 13332
        revision: 1
        datasource: Prometheus

alertmanager:
  enabled: true
  config:
    global:
      smtp_smarthost: 'smtp.gmail.com:587'
      smtp_from: 'alerts@gocoffee.dev'
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'web.hook'
      routes:
      - match:
          severity: critical
        receiver: 'critical-alerts'
      - match:
          severity: warning
        receiver: 'warning-alerts'
    
    receivers:
    - name: 'web.hook'
      webhook_configs:
      - url: 'http://localhost:5001/'
    
    - name: 'critical-alerts'
      slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#go-coffee-alerts'
        title: 'Critical Alert: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
      email_configs:
      - to: 'devops@gocoffee.dev'
        subject: 'Critical Alert: {{ .GroupLabels.alertname }}'
        body: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
    
    - name: 'warning-alerts'
      slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#go-coffee-warnings'
        title: 'Warning: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

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
    kubeApiserver: true
    kubeApiserverAvailability: true
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
    
    # Install Prometheus stack
    helm upgrade --install prometheus-stack prometheus-community/kube-prometheus-stack \
        --namespace "$MONITORING_NAMESPACE" \
        --create-namespace \
        --values /tmp/prometheus-values.yaml \
        --version 51.2.0 \
        --wait \
        --timeout 10m
    
    success "Prometheus stack deployed"
}

# Deploy Jaeger for distributed tracing
deploy_jaeger() {
    log "Deploying Jaeger for distributed tracing..."
    
    # Create Jaeger values file
    cat > /tmp/jaeger-values.yaml <<EOF
provisionDataStore:
  cassandra: false
  elasticsearch: true

elasticsearch:
  replicas: 1
  minimumMasterNodes: 1
  resources:
    requests:
      cpu: "100m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "2Gi"

storage:
  type: elasticsearch
  elasticsearch:
    host: jaeger-elasticsearch-master
    port: 9200

agent:
  enabled: true

collector:
  enabled: true
  replicaCount: 2
  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

query:
  enabled: true
  replicaCount: 2
  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
  
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
    - jaeger.gocoffee.dev
    tls:
    - secretName: jaeger-tls
      hosts:
      - jaeger.gocoffee.dev
EOF
    
    # Install Jaeger
    helm upgrade --install jaeger jaegertracing/jaeger \
        --namespace "$MONITORING_NAMESPACE" \
        --create-namespace \
        --values /tmp/jaeger-values.yaml \
        --version 0.71.2 \
        --wait \
        --timeout 10m
    
    success "Jaeger deployed"
}

# Create custom CI/CD monitoring rules
create_cicd_monitoring_rules() {
    log "Creating CI/CD monitoring rules..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: go-coffee-cicd-rules
  namespace: $MONITORING_NAMESPACE
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: monitoring
    release: prometheus-stack
spec:
  groups:
  - name: go-coffee-cicd.rules
    rules:
    # CI/CD Pipeline Metrics
    - alert: GitHubActionsFailed
      expr: github_actions_workflow_run_conclusion{conclusion="failure"} > 0
      for: 0m
      labels:
        severity: warning
        team: devops
      annotations:
        summary: "GitHub Actions workflow failed"
        description: "Workflow {{ \$labels.workflow_name }} failed in repository {{ \$labels.repository }}"
    
    - alert: DeploymentFrequencyLow
      expr: increase(argocd_app_sync_total[24h]) < 1
      for: 1h
      labels:
        severity: info
        team: devops
      annotations:
        summary: "Low deployment frequency detected"
        description: "Application {{ \$labels.name }} has not been deployed in the last 24 hours"
    
    - alert: ArgocdSyncFailed
      expr: argocd_app_info{sync_status!="Synced"} == 1
      for: 5m
      labels:
        severity: critical
        team: devops
      annotations:
        summary: "ArgoCD application sync failed"
        description: "Application {{ \$labels.name }} sync failed with status {{ \$labels.sync_status }}"
    
    - alert: ArgocdAppHealthDegraded
      expr: argocd_app_info{health_status!="Healthy"} == 1
      for: 5m
      labels:
        severity: critical
        team: devops
      annotations:
        summary: "ArgoCD application health degraded"
        description: "Application {{ \$labels.name }} health is {{ \$labels.health_status }}"
    
    # Application Performance Metrics
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
      for: 5m
      labels:
        severity: critical
        team: sre
      annotations:
        summary: "High error rate detected"
        description: "Service {{ \$labels.service }} has error rate above 5%"
    
    - alert: HighLatency
      expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
      for: 5m
      labels:
        severity: warning
        team: sre
      annotations:
        summary: "High latency detected"
        description: "Service {{ \$labels.service }} 95th percentile latency is above 1s"
    
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: critical
        team: sre
      annotations:
        summary: "Pod is crash looping"
        description: "Pod {{ \$labels.pod }} in namespace {{ \$labels.namespace }} is crash looping"
    
    # Resource Utilization
    - alert: HighCPUUsage
      expr: rate(container_cpu_usage_seconds_total[5m]) > 0.8
      for: 10m
      labels:
        severity: warning
        team: sre
      annotations:
        summary: "High CPU usage"
        description: "Container {{ \$labels.container }} CPU usage is above 80%"
    
    - alert: HighMemoryUsage
      expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
      for: 10m
      labels:
        severity: warning
        team: sre
      annotations:
        summary: "High memory usage"
        description: "Container {{ \$labels.container }} memory usage is above 90%"
EOF
    
    success "CI/CD monitoring rules created"
}

# Create custom Grafana dashboards
create_grafana_dashboards() {
    log "Creating custom Grafana dashboards..."
    
    # Create ConfigMap with custom dashboards
    kubectl create configmap go-coffee-dashboards \
        --from-file="$SCRIPT_DIR/dashboards/" \
        --namespace "$MONITORING_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    success "Custom Grafana dashboards created"
}

# Verify monitoring deployment
verify_monitoring_deployment() {
    log "Verifying monitoring deployment..."
    
    # Check Prometheus
    kubectl wait --for=condition=available --timeout=300s deployment/prometheus-stack-kube-prom-operator -n "$MONITORING_NAMESPACE"
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=prometheus -n "$MONITORING_NAMESPACE"
    
    # Check Grafana
    kubectl wait --for=condition=available --timeout=300s deployment/prometheus-stack-grafana -n "$MONITORING_NAMESPACE"
    
    # Check AlertManager
    kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=alertmanager -n "$MONITORING_NAMESPACE"
    
    # Check Jaeger
    kubectl wait --for=condition=available --timeout=300s deployment/jaeger-query -n "$MONITORING_NAMESPACE"
    
    success "Monitoring deployment verified"
}

# Display access information
display_access_info() {
    echo -e "${CYAN}"
    cat << EOF

    📊 Go Coffee Monitoring Stack Deployed Successfully!
    
    🔧 Access Information:
    
    Prometheus:
    - Port Forward: kubectl port-forward svc/prometheus-stack-kube-prom-prometheus 9090:9090 -n $MONITORING_NAMESPACE
    - URL: http://localhost:9090
    
    Grafana:
    - Port Forward: kubectl port-forward svc/prometheus-stack-grafana 3000:80 -n $MONITORING_NAMESPACE
    - URL: http://localhost:3000
    - Username: admin
    - Password: ${GRAFANA_ADMIN_PASSWORD:-admin123}
    
    AlertManager:
    - Port Forward: kubectl port-forward svc/prometheus-stack-kube-prom-alertmanager 9093:9093 -n $MONITORING_NAMESPACE
    - URL: http://localhost:9093
    
    Jaeger:
    - Port Forward: kubectl port-forward svc/jaeger-query 16686:16686 -n $MONITORING_NAMESPACE
    - URL: http://localhost:16686
    
    📈 Key Dashboards:
    - Go Coffee Overview: Grafana → Dashboards → Go Coffee → Overview
    - CI/CD Pipeline: Grafana → Dashboards → Go Coffee → CI/CD
    - Service Metrics: Grafana → Dashboards → Go Coffee → Services
    - Infrastructure: Grafana → Dashboards → Go Coffee → Infrastructure
    
    🚨 Alerting:
    - Slack: #go-coffee-alerts (critical), #go-coffee-warnings (warnings)
    - Email: devops@gocoffee.dev
    - PagerDuty: Integration configured for critical alerts
    
    📊 Metrics Endpoints:
    - Application metrics: /metrics
    - Health checks: /health
    - Readiness: /ready
    
EOF
    echo -e "${NC}"
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    📊 Go Coffee - Monitoring & Observability Stack
    ===============================================
    
    Deploying comprehensive monitoring solution:
    • Prometheus for metrics collection
    • Grafana for visualization
    • AlertManager for alerting
    • Jaeger for distributed tracing
    • Custom CI/CD dashboards
    • Application performance monitoring
    
EOF
    echo -e "${NC}"
    
    info "Starting monitoring stack deployment..."
    info "Environment: $ENVIRONMENT"
    info "Namespace: $MONITORING_NAMESPACE"
    
    # Execute deployment steps
    check_prerequisites
    setup_monitoring_namespace
    add_helm_repos
    deploy_prometheus_stack
    deploy_jaeger
    create_cicd_monitoring_rules
    create_grafana_dashboards
    verify_monitoring_deployment
    
    display_access_info
    
    success "🎉 Monitoring stack deployment completed successfully!"
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "destroy")
        warn "Destroying monitoring stack..."
        helm uninstall prometheus-stack -n "$MONITORING_NAMESPACE" --ignore-not-found=true
        helm uninstall jaeger -n "$MONITORING_NAMESPACE" --ignore-not-found=true
        kubectl delete namespace "$MONITORING_NAMESPACE" --ignore-not-found=true
        success "Monitoring stack destroyed"
        ;;
    "verify")
        verify_monitoring_deployment
        ;;
    *)
        echo "Usage: $0 [deploy|destroy|verify]"
        exit 1
        ;;
esac
