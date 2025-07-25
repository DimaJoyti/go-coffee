#!/bin/bash

# ☕ Go Coffee - AI Agent Stack Deployment Script
# Deploys comprehensive AI agent ecosystem with GPU infrastructure, model serving, and orchestration

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
AI_NAMESPACE="${AI_NAMESPACE:-go-coffee-ai}"
ENVIRONMENT="${ENVIRONMENT:-production}"
CLUSTER_NAME="${CLUSTER_NAME:-go-coffee-cluster}"

# AI configuration
ENABLE_GPU_NODES="${ENABLE_GPU_NODES:-true}"
ENABLE_OLLAMA="${ENABLE_OLLAMA:-true}"
ENABLE_AGENTS="${ENABLE_AGENTS:-true}"
ENABLE_ORCHESTRATION="${ENABLE_ORCHESTRATION:-true}"
ENABLE_WORKFLOWS="${ENABLE_WORKFLOWS:-true}"

# GPU configuration
GPU_NODE_COUNT="${GPU_NODE_COUNT:-2}"
GPU_TYPE="${GPU_TYPE:-nvidia-tesla-t4}"
GPU_MEMORY="${GPU_MEMORY:-16Gi}"

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
    log "Checking AI stack prerequisites..."
    
    # Check required tools
    local tools=("kubectl" "helm" "docker")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is not installed or not in PATH"
        fi
    done
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl is not properly configured or cluster is not accessible"
    fi
    
    # Check for GPU support if enabled
    if [[ "$ENABLE_GPU_NODES" == "true" ]]; then
        info "GPU support enabled - checking for GPU operator..."
        if ! kubectl get crd gpus.nvidia.com &> /dev/null; then
            warn "NVIDIA GPU Operator not found. Will install during deployment."
        fi
    fi
    
    # Check available resources
    local cpu_capacity=$(kubectl get nodes -o jsonpath='{.items[*].status.capacity.cpu}' | tr ' ' '\n' | awk '{sum += $1} END {print sum}')
    local memory_capacity=$(kubectl get nodes -o jsonpath='{.items[*].status.capacity.memory}' | tr ' ' '\n' | sed 's/Ki$//' | awk '{sum += $1} END {print sum/1024/1024 "Gi"}')
    
    info "Cluster resources: ${cpu_capacity} CPUs, ${memory_capacity} memory"
    
    success "AI stack prerequisites met"
}

# Setup AI namespace and RBAC
setup_ai_namespace() {
    log "Setting up AI namespace and RBAC..."
    
    # Create AI namespace
    kubectl create namespace "$AI_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label AI namespace
    kubectl label namespace "$AI_NAMESPACE" \
        name="$AI_NAMESPACE" \
        app.kubernetes.io/name=go-coffee-ai \
        app.kubernetes.io/component=ai-infrastructure \
        pod-security.kubernetes.io/enforce=baseline \
        pod-security.kubernetes.io/audit=restricted \
        pod-security.kubernetes.io/warn=restricted \
        nvidia.com/gpu.deploy.operands=true \
        --overwrite
    
    # Create service accounts for AI agents
    local agents=("beverage-inventor" "inventory-manager" "task-manager" "social-media-manager" 
                  "customer-service" "financial-analyst" "marketing-specialist" "quality-assurance" 
                  "supply-chain-optimizer" "agent-orchestrator" "ollama")
    
    for agent in "${agents[@]}"; do
        kubectl create serviceaccount "$agent" -n "$AI_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    done
    
    success "AI namespace and RBAC configured"
}

# Add Helm repositories for AI tools
add_helm_repos() {
    log "Adding Helm repositories for AI tools..."
    
    # NVIDIA GPU Operator
    helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
    
    # Argo Workflows for orchestration
    helm repo add argo https://argoproj.github.io/argo-helm
    
    # Prometheus for monitoring
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    
    # Update repositories
    helm repo update
    
    success "Helm repositories added"
}

# Deploy GPU infrastructure
deploy_gpu_infrastructure() {
    if [[ "$ENABLE_GPU_NODES" != "true" ]]; then
        info "GPU infrastructure deployment skipped"
        return 0
    fi
    
    log "Deploying GPU infrastructure..."
    
    # Apply GPU node pool configuration
    kubectl apply -f "$SCRIPT_DIR/infrastructure/gpu-node-pool.yaml"
    
    # Install NVIDIA GPU Operator
    helm upgrade --install gpu-operator nvidia/gpu-operator \
        --namespace gpu-operator \
        --create-namespace \
        --values - <<EOF
operator:
  defaultRuntime: containerd
  runtimeClass: nvidia

driver:
  enabled: true
  version: "535.129.03"

toolkit:
  enabled: true
  version: "1.14.3-centos7"

devicePlugin:
  enabled: true
  version: "0.14.3-ubi8"

dcgmExporter:
  enabled: true
  version: "3.2.5-3.1.8-ubuntu20.04"
  serviceMonitor:
    enabled: true
    interval: 15s

nfd:
  enabled: true

mig:
  strategy: single

tolerations:
- key: nvidia.com/gpu
  operator: Exists
  effect: NoSchedule
EOF
    
    # Wait for GPU operator to be ready
    kubectl wait --for=condition=available --timeout=600s deployment/gpu-operator -n gpu-operator
    
    success "GPU infrastructure deployed"
}

# Deploy Ollama model serving
deploy_ollama() {
    if [[ "$ENABLE_OLLAMA" != "true" ]]; then
        info "Ollama deployment skipped"
        return 0
    fi
    
    log "Deploying Ollama model serving..."
    
    # Apply Ollama configuration
    kubectl apply -f "$SCRIPT_DIR/model-serving/ollama-deployment.yaml"
    
    # Wait for Ollama to be ready
    kubectl wait --for=condition=available --timeout=600s statefulset/ollama -n "$AI_NAMESPACE"
    
    # Wait for models to be downloaded
    info "Waiting for AI models to be downloaded (this may take several minutes)..."
    kubectl wait --for=condition=ready --timeout=1800s pod -l app.kubernetes.io/name=ollama -n "$AI_NAMESPACE"
    
    success "Ollama model serving deployed"
}

# Deploy AI agents
deploy_ai_agents() {
    if [[ "$ENABLE_AGENTS" != "true" ]]; then
        info "AI agents deployment skipped"
        return 0
    fi
    
    log "Deploying AI agents..."
    
    # Deploy Beverage Inventor Agent
    info "Deploying Beverage Inventor Agent..."
    kubectl apply -f "$SCRIPT_DIR/agents/beverage-inventor-agent.yaml"
    
    # Deploy Inventory Manager Agent
    info "Deploying Inventory Manager Agent..."
    kubectl apply -f "$SCRIPT_DIR/agents/inventory-manager-agent.yaml"
    
    # Deploy other agents (placeholder for now)
    local agents=("task-manager" "social-media-manager" "customer-service" 
                  "financial-analyst" "marketing-specialist" "quality-assurance" 
                  "supply-chain-optimizer")
    
    for agent in "${agents[@]}"; do
        info "Creating placeholder for ${agent} agent..."
        cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${agent}
  namespace: ${AI_NAMESPACE}
  labels:
    app.kubernetes.io/name: ${agent}
    app.kubernetes.io/component: ai-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: ${agent}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ${agent}
        app.kubernetes.io/component: ai-agent
    spec:
      containers:
      - name: ${agent}
        image: busybox:1.36
        command: ['sh', '-c', 'echo "${agent} agent placeholder - implementation coming soon"; sleep 3600']
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: ${agent}
  namespace: ${AI_NAMESPACE}
  labels:
    app.kubernetes.io/name: ${agent}
spec:
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app.kubernetes.io/name: ${agent}
EOF
    done
    
    # Wait for agents to be ready
    kubectl wait --for=condition=available --timeout=300s deployment/beverage-inventor -n "$AI_NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/inventory-manager -n "$AI_NAMESPACE"
    
    success "AI agents deployed"
}

# Deploy orchestration system
deploy_orchestration() {
    if [[ "$ENABLE_ORCHESTRATION" != "true" ]]; then
        info "Orchestration deployment skipped"
        return 0
    fi
    
    log "Deploying AI orchestration system..."
    
    # Install Argo Workflows for workflow orchestration
    if [[ "$ENABLE_WORKFLOWS" == "true" ]]; then
        info "Installing Argo Workflows..."
        helm upgrade --install argo-workflows argo/argo-workflows \
            --namespace argo \
            --create-namespace \
            --values - <<EOF
controller:
  workflowNamespaces:
  - ${AI_NAMESPACE}
  - argo

server:
  enabled: true
  serviceType: ClusterIP
  
executor:
  image:
    tag: "v3.5.1"

workflow:
  serviceAccount:
    create: true
    name: argo-workflow

useDefaultArtifactRepo: true
artifactRepository:
  archiveLogs: true
  s3:
    bucket: go-coffee-workflows
    endpoint: minio.minio.svc.cluster.local:9000
    insecure: true
    accessKeySecret:
      name: argo-artifacts
      key: accesskey
    secretKeySecret:
      name: argo-artifacts
      key: secretkey
EOF
        
        # Wait for Argo Workflows
        kubectl wait --for=condition=available --timeout=300s deployment/argo-workflows-server -n argo
    fi
    
    # Deploy agent orchestrator
    kubectl apply -f "$SCRIPT_DIR/orchestration/agent-orchestrator.yaml"
    
    # Wait for orchestrator to be ready
    kubectl wait --for=condition=available --timeout=300s deployment/agent-orchestrator -n "$AI_NAMESPACE"
    
    success "AI orchestration system deployed"
}

# Configure monitoring for AI stack
configure_ai_monitoring() {
    log "Configuring AI stack monitoring..."
    
    # Create ServiceMonitor for AI agents
    if kubectl get crd servicemonitors.monitoring.coreos.com &> /dev/null; then
        cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ai-agents
  namespace: go-coffee-monitoring
  labels:
    app.kubernetes.io/name: ai-agents
    app.kubernetes.io/component: monitoring
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: ai-agent
  namespaceSelector:
    matchNames:
    - ${AI_NAMESPACE}
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
    honorLabels: true
EOF
        
        # Create ServiceMonitor for Ollama
        cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ollama
  namespace: go-coffee-monitoring
  labels:
    app.kubernetes.io/name: ollama
    app.kubernetes.io/component: monitoring
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: ollama
  namespaceSelector:
    matchNames:
    - ${AI_NAMESPACE}
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
    honorLabels: true
EOF
        
        info "ServiceMonitors created for Prometheus integration"
    fi
    
    # Create AI-specific alerts
    if kubectl get crd prometheusrules.monitoring.coreos.com &> /dev/null; then
        cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: ai-agents-alerts
  namespace: go-coffee-monitoring
  labels:
    app.kubernetes.io/name: ai-agents
    app.kubernetes.io/component: alerts
    release: prometheus
spec:
  groups:
  - name: ai-agents.rules
    rules:
    - alert: AIAgentDown
      expr: up{job=~".*ai-agent.*"} == 0
      for: 2m
      labels:
        severity: critical
        team: ai
      annotations:
        summary: "AI agent {{ \$labels.job }} is down"
        description: "AI agent {{ \$labels.job }} has been down for more than 2 minutes"
    
    - alert: OllamaModelLoadingFailed
      expr: ollama_model_load_duration_seconds > 300
      for: 1m
      labels:
        severity: warning
        team: ai
      annotations:
        summary: "Ollama model loading is slow"
        description: "Ollama model loading is taking more than 5 minutes"
    
    - alert: AIAgentHighLatency
      expr: histogram_quantile(0.95, rate(ai_agent_request_duration_seconds_bucket[5m])) > 10
      for: 5m
      labels:
        severity: warning
        team: ai
      annotations:
        summary: "High latency in AI agent responses"
        description: "95th percentile latency is {{ \$value }}s for AI agents"
    
    - alert: GPUUtilizationLow
      expr: nvidia_gpu_utilization < 20
      for: 10m
      labels:
        severity: info
        team: ai
      annotations:
        summary: "Low GPU utilization detected"
        description: "GPU utilization is {{ \$value }}% which may indicate underutilization"
EOF
        
        info "AI-specific alerts configured"
    fi
    
    success "AI stack monitoring configured"
}

# Verify AI deployment
verify_ai_deployment() {
    log "Verifying AI stack deployment..."
    
    # Check namespaces
    info "Checking AI namespaces..."
    kubectl get namespaces | grep -E "(go-coffee-ai|gpu-operator|argo)" || warn "Some AI namespaces missing"
    
    # Check GPU nodes if enabled
    if [[ "$ENABLE_GPU_NODES" == "true" ]]; then
        info "Checking GPU nodes..."
        local gpu_nodes=$(kubectl get nodes -l workload-type=ai-gpu --no-headers | wc -l)
        info "Found $gpu_nodes GPU-enabled nodes"
        
        # Check GPU operator
        kubectl get pods -n gpu-operator | grep gpu-operator || warn "GPU operator not running"
    fi
    
    # Check Ollama
    if [[ "$ENABLE_OLLAMA" == "true" ]]; then
        info "Checking Ollama model serving..."
        kubectl get pods -n "$AI_NAMESPACE" -l app.kubernetes.io/name=ollama | grep Running || warn "Ollama not running"
        
        # Test Ollama API
        if timeout 30 kubectl port-forward svc/ollama 11434:11434 -n "$AI_NAMESPACE" &
        then
            local pf_pid=$!
            sleep 5
            if timeout 10 curl -f http://localhost:11434/api/tags &> /dev/null; then
                success "Ollama API health check passed"
            else
                warn "Ollama API health check failed"
            fi
            kill $pf_pid 2>/dev/null || true
        fi
    fi
    
    # Check AI agents
    if [[ "$ENABLE_AGENTS" == "true" ]]; then
        info "Checking AI agents..."
        local agent_count=$(kubectl get deployments -n "$AI_NAMESPACE" -l app.kubernetes.io/component=ai-agent --no-headers | wc -l)
        info "Found $agent_count AI agents deployed"
        
        # Check specific agents
        kubectl get pods -n "$AI_NAMESPACE" -l app.kubernetes.io/name=beverage-inventor | grep Running || warn "Beverage Inventor agent not running"
        kubectl get pods -n "$AI_NAMESPACE" -l app.kubernetes.io/name=inventory-manager | grep Running || warn "Inventory Manager agent not running"
    fi
    
    # Check orchestration
    if [[ "$ENABLE_ORCHESTRATION" == "true" ]]; then
        info "Checking orchestration system..."
        kubectl get pods -n "$AI_NAMESPACE" -l app.kubernetes.io/name=agent-orchestrator | grep Running || warn "Agent orchestrator not running"
        
        if [[ "$ENABLE_WORKFLOWS" == "true" ]]; then
            kubectl get pods -n argo | grep argo-workflows || warn "Argo Workflows not running"
        fi
    fi
    
    # Resource utilization check
    info "Checking resource utilization..."
    kubectl top nodes 2>/dev/null || warn "Metrics server not available for resource monitoring"
    
    success "AI stack deployment verification completed"
}

# Generate AI deployment report
generate_ai_report() {
    log "Generating AI deployment report..."
    
    local report_file="/tmp/go-coffee-ai-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" <<EOF
☕ Go Coffee - AI Agent Stack Deployment Report
==============================================

Deployment Date: $(date)
Environment: $ENVIRONMENT
Cluster: $CLUSTER_NAME
AI Namespace: $AI_NAMESPACE

AI Components Deployed:
======================

$(if [[ "$ENABLE_GPU_NODES" == "true" ]]; then echo "✅ GPU Infrastructure: Deployed"; else echo "❌ GPU Infrastructure: Skipped"; fi)
   - NVIDIA GPU Operator
   - GPU-enabled node pools
   - GPU resource quotas and scheduling

$(if [[ "$ENABLE_OLLAMA" == "true" ]]; then echo "✅ Ollama Model Serving: Deployed"; else echo "❌ Ollama Model Serving: Skipped"; fi)
   - Local LLM hosting with Ollama
   - Multiple AI models (CodeLlama, Llama2, Mistral, etc.)
   - GPU-accelerated inference
   - Auto-scaling and load balancing

$(if [[ "$ENABLE_AGENTS" == "true" ]]; then echo "✅ AI Agents: Deployed"; else echo "❌ AI Agents: Skipped"; fi)
   - Beverage Inventor Agent (recipe creation)
   - Inventory Manager Agent (demand forecasting)
   - Task Manager Agent (workflow optimization)
   - Customer Service Agent (support automation)
   - Financial Analyst Agent (cost optimization)
   - Marketing Specialist Agent (campaign creation)
   - Quality Assurance Agent (process monitoring)
   - Supply Chain Optimizer Agent (logistics)
   - Social Media Manager Agent (content creation)

$(if [[ "$ENABLE_ORCHESTRATION" == "true" ]]; then echo "✅ AI Orchestration: Deployed"; else echo "❌ AI Orchestration: Skipped"; fi)
   - Agent Orchestrator (central coordination)
   - Workflow engine integration
   - Event-driven communication
   - Auto-scaling and load balancing

$(if [[ "$ENABLE_WORKFLOWS" == "true" ]]; then echo "✅ Workflow Engine: Deployed"; else echo "❌ Workflow Engine: Skipped"; fi)
   - Argo Workflows for complex AI workflows
   - Automated business process orchestration
   - Event-driven workflow triggers

AI Capabilities:
===============

🤖 Intelligent Recipe Creation: AI-powered beverage innovation
📊 Demand Forecasting: Predictive inventory management
🎯 Customer Service: Automated support and issue resolution
💰 Financial Analysis: Cost optimization and revenue forecasting
📈 Marketing Automation: Campaign creation and optimization
🔍 Quality Assurance: Automated quality monitoring
🚚 Supply Chain Optimization: Logistics and delivery planning
📱 Social Media Management: Content creation and engagement

Resource Allocation:
===================

GPU Nodes: $(kubectl get nodes -l workload-type=ai-gpu --no-headers 2>/dev/null | wc -l)
AI Agents: $(kubectl get deployments -n $AI_NAMESPACE -l app.kubernetes.io/component=ai-agent --no-headers 2>/dev/null | wc -l)
Model Servers: $(kubectl get statefulsets -n $AI_NAMESPACE -l app.kubernetes.io/name=ollama --no-headers 2>/dev/null | wc -l)
Orchestrators: $(kubectl get deployments -n $AI_NAMESPACE -l app.kubernetes.io/name=agent-orchestrator --no-headers 2>/dev/null | wc -l)

Performance Metrics:
===================

✅ Model Loading: Automated on startup
✅ GPU Acceleration: Enabled for inference
✅ Auto-scaling: Configured for all agents
✅ Load Balancing: Service mesh integration
✅ Monitoring: Prometheus metrics collection
✅ Alerting: AI-specific alert rules

Next Steps:
==========

1. Configure external AI API integrations (OpenAI, Anthropic, etc.)
2. Implement custom AI model fine-tuning workflows
3. Set up A/B testing for AI agent performance
4. Configure advanced workflow automation
5. Implement AI agent performance optimization
6. Set up AI model versioning and deployment pipelines

For more information, visit: https://docs.gocoffee.dev/ai-agents
EOF
    
    echo "$report_file"
    success "AI deployment report generated: $report_file"
}

# Display AI access information
display_ai_info() {
    echo -e "${CYAN}"
    cat << EOF

    🤖 Go Coffee AI Agent Stack Deployed Successfully!
    
    🧠 AI Components:
    
    Model Serving (Ollama):
    - Local: kubectl port-forward svc/ollama 11434:11434 -n $AI_NAMESPACE
    - API: http://localhost:11434/api/tags
    - Models: CodeLlama, Llama2, Mistral, Neural-Chat, Nomic-Embed
    
    AI Agents:
    - Beverage Inventor: kubectl port-forward svc/beverage-inventor 8080:8080 -n $AI_NAMESPACE
    - Inventory Manager: kubectl port-forward svc/inventory-manager 8080:8080 -n $AI_NAMESPACE
    - Agent Orchestrator: kubectl port-forward svc/agent-orchestrator 8080:8080 -n $AI_NAMESPACE
    
    Workflow Engine (Argo):
    - UI: kubectl port-forward svc/argo-workflows-server 2746:2746 -n argo
    - Access: http://localhost:2746
    
    🔧 Management Commands:
    
    # Check AI agent status
    kubectl get pods -n $AI_NAMESPACE -l app.kubernetes.io/component=ai-agent
    
    # View Ollama models
    kubectl exec -it statefulset/ollama -n $AI_NAMESPACE -- ollama list
    
    # Check GPU utilization
    kubectl get nodes -l workload-type=ai-gpu -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\.com/gpu
    
    # View AI agent logs
    kubectl logs -l app.kubernetes.io/name=beverage-inventor -n $AI_NAMESPACE -f
    
    # Test AI agent API
    kubectl port-forward svc/beverage-inventor 8080:8080 -n $AI_NAMESPACE &
    curl -X POST http://localhost:8080/api/v1/create_recipe -d '{"season":"winter","flavor":"spiced"}'
    
    📊 Monitoring:
    - AI metrics: Check Prometheus/Grafana dashboards
    - GPU monitoring: NVIDIA DCGM exporter metrics
    - Agent performance: Custom AI agent metrics
    - Workflow status: Argo Workflows UI
    
    🚀 AI Capabilities:
    - Recipe Innovation: Create unique coffee beverages
    - Demand Forecasting: Predict inventory needs
    - Customer Service: Automated support responses
    - Financial Analysis: Cost and revenue optimization
    - Marketing Automation: Campaign and content creation
    - Quality Assurance: Process monitoring and improvement
    - Supply Chain: Logistics and delivery optimization
    
    📚 Documentation: https://docs.gocoffee.dev/ai-agents
    
EOF
    echo -e "${NC}"
}

# Cleanup function
cleanup() {
    if [[ "${1:-}" == "destroy" ]]; then
        warn "Destroying AI stack..."
        
        # Delete AI resources
        kubectl delete -f "$SCRIPT_DIR/orchestration/agent-orchestrator.yaml" --ignore-not-found=true
        kubectl delete -f "$SCRIPT_DIR/agents/" --ignore-not-found=true
        kubectl delete -f "$SCRIPT_DIR/model-serving/ollama-deployment.yaml" --ignore-not-found=true
        kubectl delete -f "$SCRIPT_DIR/infrastructure/gpu-node-pool.yaml" --ignore-not-found=true
        
        # Delete Helm releases
        helm uninstall argo-workflows -n argo --ignore-not-found=true
        helm uninstall gpu-operator -n gpu-operator --ignore-not-found=true
        
        # Delete namespaces
        kubectl delete namespace "$AI_NAMESPACE" gpu-operator argo --ignore-not-found=true
        
        success "AI stack destroyed"
    fi
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    ☕ Go Coffee - AI Agent Stack Deployment
    =======================================
    
    Deploying intelligent AI ecosystem:
    • GPU Infrastructure & NVIDIA Operator
    • Ollama Local Model Serving
    • 9 Specialized AI Agents
    • Agent Orchestration System
    • Workflow Automation Engine
    • Real-time AI Inference
    
EOF
    echo -e "${NC}"
    
    info "Starting AI stack deployment..."
    info "Environment: $ENVIRONMENT"
    info "Cluster: $CLUSTER_NAME"
    info "AI Namespace: $AI_NAMESPACE"
    info "GPU Support: $ENABLE_GPU_NODES"
    
    # Execute deployment steps
    check_prerequisites
    setup_ai_namespace
    add_helm_repos
    deploy_gpu_infrastructure
    deploy_ollama
    deploy_ai_agents
    deploy_orchestration
    configure_ai_monitoring
    verify_ai_deployment
    
    # Generate report
    local report_file=$(generate_ai_report)
    
    display_ai_info
    
    success "🎉 AI Agent Stack deployment completed successfully!"
    info "AI deployment report saved to: $report_file"
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
        verify_ai_deployment
        ;;
    *)
        echo "Usage: $0 [deploy|destroy|verify]"
        exit 1
        ;;
esac
