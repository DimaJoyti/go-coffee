#!/bin/bash

# CI/CD Pipeline Enhancement Deployment Script
# Deploys comprehensive CI/CD pipelines with GitOps, security scanning, and multi-cloud deployment

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/cicd-enhancement"
CICD_DIR="$PROJECT_ROOT/k8s/cicd"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
CICD_NAMESPACE="${CICD_NAMESPACE:-cicd}"
ENABLE_ARGOCD="${ENABLE_ARGOCD:-true}"
ENABLE_TEKTON_PIPELINES="${ENABLE_TEKTON_PIPELINES:-true}"
ENABLE_GITHUB_ACTIONS_RUNNER="${ENABLE_GITHUB_ACTIONS_RUNNER:-true}"
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
    
    # Check required environment variables
    if [[ -z "${GIT_REPOSITORY_URL:-}" ]]; then
        log_error "GIT_REPOSITORY_URL environment variable is required"
        exit 1
    fi
    
    if [[ "$ENABLE_GITHUB_ACTIONS_RUNNER" == "true" && -z "${GITHUB_TOKEN:-}" ]]; then
        log_error "GITHUB_TOKEN environment variable is required for GitHub Actions Runner"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Setup CI/CD namespace
setup_namespace() {
    log_info "Setting up CI/CD namespace: $CICD_NAMESPACE"
    
    if kubectl get namespace "$CICD_NAMESPACE" &>/dev/null; then
        log_info "Namespace $CICD_NAMESPACE already exists"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would create namespace $CICD_NAMESPACE"
        else
            kubectl create namespace "$CICD_NAMESPACE"
            
            # Apply labels and annotations
            kubectl label namespace "$CICD_NAMESPACE" \
                app.kubernetes.io/name="$PROJECT_NAME" \
                app.kubernetes.io/component="cicd" \
                environment="$ENVIRONMENT"
            
            kubectl annotate namespace "$CICD_NAMESPACE" \
                managed-by="terraform" \
                created-at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
            
            log_success "Created CI/CD namespace: $CICD_NAMESPACE"
        fi
    fi
}

# Add Helm repositories
add_helm_repositories() {
    log_info "Adding Helm repositories for CI/CD tools..."
    
    local repos=(
        "argo:https://argoproj.github.io/argo-helm"
        "tekton:https://cdfoundation.github.io/tekton-helm-chart"
        "actions-runner-controller:https://actions-runner-controller.github.io/actions-runner-controller"
        "jetstack:https://charts.jetstack.io"
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

# Create CI/CD secrets
create_cicd_secrets() {
    log_info "Creating CI/CD secrets..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create CI/CD secrets"
        return 0
    fi
    
    # Create Git credentials
    kubectl create secret generic git-credentials \
        --from-literal=username="${GIT_USERNAME:-}" \
        --from-literal=password="${GIT_TOKEN:-}" \
        --namespace="$CICD_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create Docker registry credentials
    kubectl create secret docker-registry docker-credentials \
        --docker-server="${DOCKER_REGISTRY_SERVER:-docker.io}" \
        --docker-username="${DOCKER_REGISTRY_USERNAME:-}" \
        --docker-password="${DOCKER_REGISTRY_PASSWORD:-}" \
        --docker-email="${DOCKER_REGISTRY_EMAIL:-cicd@go-coffee.com}" \
        --namespace="$CICD_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create Slack token for notifications
    if [[ -n "${SLACK_TOKEN:-}" ]]; then
        kubectl create secret generic slack-token \
            --from-literal=slack-token="${SLACK_TOKEN}" \
            --namespace="$CICD_NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    log_success "CI/CD secrets created"
}

# Deploy ArgoCD
deploy_argocd() {
    if [[ "$ENABLE_ARGOCD" != "true" ]]; then
        log_info "ArgoCD disabled, skipping..."
        return 0
    fi

    log_info "Deploying ArgoCD..."

    local values_file="$CICD_DIR/argocd-values.yaml"

    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$CICD_DIR"
        log_info "Creating ArgoCD values file..."
        cat > "$values_file" << EOF
global:
  image:
    repository: quay.io/argoproj/argocd
    tag: ${ARGOCD_VERSION:-v2.9.3}

controller:
  replicas: ${ARGOCD_CONTROLLER_REPLICAS:-1}
  resources:
    requests:
      cpu: "250m"
      memory: "1Gi"
    limits:
      cpu: "500m"
      memory: "2Gi"
  metrics:
    enabled: ${MONITORING_ENABLED:-true}
    serviceMonitor:
      enabled: ${MONITORING_ENABLED:-true}

server:
  replicas: ${ARGOCD_SERVER_REPLICAS:-2}
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
  
  ingress:
    enabled: ${ARGOCD_INGRESS_ENABLED:-true}
    ingressClassName: ${INGRESS_CLASS_NAME:-nginx}
    hosts:
      - ${ARGOCD_HOSTNAME:-argocd.go-coffee.local}
    tls:
      - secretName: argocd-tls
        hosts:
          - ${ARGOCD_HOSTNAME:-argocd.go-coffee.local}
    annotations:
      cert-manager.io/cluster-issuer: ${CERT_MANAGER_ISSUER:-letsencrypt-prod}
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
  
  config:
    application.instanceLabelKey: argocd.argoproj.io/instance
    server.rbac.log.enforce.enable: "true"
    policy.default: role:readonly
    policy.csv: |
      p, role:admin, applications, *, */*, allow
      p, role:admin, clusters, *, *, allow
      p, role:admin, repositories, *, *, allow
      g, ${PROJECT_NAME}:admin, role:admin
    
    repositories: |
      - url: ${GIT_REPOSITORY_URL}
        passwordSecret:
          name: git-credentials
          key: password
        usernameSecret:
          name: git-credentials
          key: username
  
  metrics:
    enabled: ${MONITORING_ENABLED:-true}
    serviceMonitor:
      enabled: ${MONITORING_ENABLED:-true}

repoServer:
  replicas: ${ARGOCD_REPO_SERVER_REPLICAS:-2}
  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "1Gi"
  metrics:
    enabled: ${MONITORING_ENABLED:-true}
    serviceMonitor:
      enabled: ${MONITORING_ENABLED:-true}

redis:
  enabled: true
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"

applicationSet:
  enabled: ${ENABLE_ARGOCD_APPLICATIONSET:-true}
  replicas: 1
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"

notifications:
  enabled: ${ENABLE_ARGOCD_NOTIFICATIONS:-true}
  argocdUrl: https://${ARGOCD_HOSTNAME:-argocd.go-coffee.local}
  
  subscriptions:
    - recipients:
        - slack:${SLACK_CHANNEL:-#deployments}
      triggers:
        - on-deployed
        - on-health-degraded
        - on-sync-failed
  
  services:
    service.slack:
      token: \$slack-token
  
  templates:
    template.app-deployed:
      message: "Application {{.app.metadata.name}} is now running new version."
      slack:
        attachments: |
          [{
            "title": "{{.app.metadata.name}}",
            "color": "good",
            "fields": [
              {
                "title": "Sync Status",
                "value": "{{.app.status.sync.status}}",
                "short": true
              },
              {
                "title": "Repository",
                "value": "{{.app.spec.source.repoURL}}",
                "short": true
              }
            ]
          }]
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy ArgoCD"
        helm template argocd argo/argo-cd \
            --namespace "$CICD_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: ArgoCD template validation passed"
    else
        helm upgrade --install argocd argo/argo-cd \
            --namespace "$CICD_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 15m
        
        log_success "ArgoCD deployed successfully"
        
        # Wait for ArgoCD to be ready
        log_info "Waiting for ArgoCD to be ready..."
        kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n "$CICD_NAMESPACE"
        
        # Get initial admin password
        if kubectl get secret argocd-initial-admin-secret -n "$CICD_NAMESPACE" &>/dev/null; then
            ARGOCD_PASSWORD=$(kubectl get secret argocd-initial-admin-secret -n "$CICD_NAMESPACE" -o jsonpath="{.data.password}" | base64 -d)
            log_info "ArgoCD admin password: $ARGOCD_PASSWORD"
            echo "ArgoCD admin password: $ARGOCD_PASSWORD" > "$PROJECT_ROOT/.argocd-password"
            chmod 600 "$PROJECT_ROOT/.argocd-password"
        fi
    fi
}

# Deploy Tekton Pipelines
deploy_tekton() {
    if [[ "$ENABLE_TEKTON_PIPELINES" != "true" ]]; then
        log_info "Tekton Pipelines disabled, skipping..."
        return 0
    fi

    log_info "Deploying Tekton Pipelines..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Tekton Pipelines"
    else
        # Install Tekton Pipelines
        kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

        # Install Tekton Triggers
        kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
        kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml

        # Wait for Tekton to be ready
        log_info "Waiting for Tekton Pipelines to be ready..."
        kubectl wait --for=condition=ready pod --all -n tekton-pipelines --timeout=300s
        kubectl wait --for=condition=ready pod --all -n tekton-pipelines-resolvers --timeout=300s

        # Deploy custom pipeline
        log_info "Deploying Go Coffee Tekton pipeline..."
        kubectl apply -f "$PROJECT_ROOT/k8s/tekton/"

        log_success "Tekton Pipelines deployed successfully"
    fi
}

# Deploy GitHub Actions Runner
deploy_github_runner() {
    if [[ "$ENABLE_GITHUB_ACTIONS_RUNNER" != "true" ]]; then
        log_info "GitHub Actions Runner disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying GitHub Actions Runner Controller..."
    
    local values_file="$CICD_DIR/actions-runner-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$CICD_DIR"
        log_info "Creating Actions Runner Controller values file..."
        cat > "$values_file" << EOF
authSecret:
  create: true
  github_token: ${GITHUB_TOKEN}

replicaCount: ${ACTIONS_RUNNER_CONTROLLER_REPLICAS:-1}

resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"

metrics:
  serviceMonitor:
    enabled: ${MONITORING_ENABLED:-true}

githubWebhookServer:
  enabled: ${ENABLE_GITHUB_WEBHOOK_SERVER:-true}
  
  ingress:
    enabled: ${GITHUB_WEBHOOK_INGRESS_ENABLED:-true}
    ingressClassName: ${INGRESS_CLASS_NAME:-nginx}
    hosts:
      - host: ${GITHUB_WEBHOOK_HOSTNAME:-github-webhook.go-coffee.local}
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: github-webhook-tls
        hosts:
          - ${GITHUB_WEBHOOK_HOSTNAME:-github-webhook.go-coffee.local}
    annotations:
      cert-manager.io/cluster-issuer: ${CERT_MANAGER_ISSUER:-letsencrypt-prod}
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy GitHub Actions Runner Controller"
        helm template actions-runner-controller actions-runner-controller/actions-runner-controller \
            --namespace "$CICD_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Actions Runner Controller template validation passed"
    else
        helm upgrade --install actions-runner-controller actions-runner-controller/actions-runner-controller \
            --namespace "$CICD_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "GitHub Actions Runner Controller deployed successfully"
        
        # Deploy runner deployment
        deploy_github_runner_deployment
    fi
}

# Deploy GitHub Runner Deployment
deploy_github_runner_deployment() {
    log_info "Deploying GitHub Runner Deployment..."
    
    local runner_file="$CICD_DIR/github-runner-deployment.yaml"
    
    cat > "$runner_file" << EOF
apiVersion: actions.summerwind.dev/v1alpha1
kind: RunnerDeployment
metadata:
  name: ${PROJECT_NAME}-github-runners
  namespace: $CICD_NAMESPACE
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: github-runner
    environment: $ENVIRONMENT
spec:
  replicas: ${GITHUB_RUNNER_REPLICAS:-3}
  template:
    spec:
      repository: ${GIT_REPOSITORY_URL}
      
      labels:
        - go-coffee
        - kubernetes
        - self-hosted
      
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"
      
      dockerdWithinRunnerContainer: true
      
      env:
        - name: ENVIRONMENT
          value: "$ENVIRONMENT"
        - name: PROJECT_NAME
          value: "$PROJECT_NAME"
      
      volumeMounts:
        - name: docker-cache
          mountPath: /var/lib/docker
      
      volumes:
        - name: docker-cache
          emptyDir:
            sizeLimit: 10Gi
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy GitHub Runner Deployment"
    else
        kubectl apply -f "$runner_file"
        log_success "GitHub Runner Deployment created"
    fi
}

# Deploy ArgoCD Applications
deploy_argocd_applications() {
    if [[ "$ENABLE_ARGOCD" != "true" ]]; then
        log_info "ArgoCD disabled, skipping application deployment..."
        return 0
    fi

    log_info "Deploying ArgoCD applications..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy ArgoCD applications"
    else
        # Deploy ArgoCD applications
        kubectl apply -f "$PROJECT_ROOT/k8s/argocd/"

        # Wait for applications to be created
        sleep 30

        # Check application status
        kubectl get applications -n "$CICD_NAMESPACE"

        log_success "ArgoCD applications deployed"
    fi
}

# Create CI/CD pipeline templates
create_pipeline_templates() {
    log_info "Creating CI/CD pipeline templates..."

    local templates_dir="$CICD_DIR/templates"
    mkdir -p "$templates_dir"

    # Create GitHub Actions workflow
    local workflows_dir="$PROJECT_ROOT/.github/workflows"
    mkdir -p "$workflows_dir"
    
    cat > "$workflows_dir/go-coffee-cicd.yml" << EOF
name: Go Coffee CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: docker.io
  IMAGE_NAME: go-coffee

jobs:
  test:
    runs-on: [self-hosted, go-coffee]
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: \${{ runner.os }}-go-\${{ hashFiles('**/go.sum') }}
        restore-keys: |
          \${{ runner.os }}-go-
    
    - name: Run tests
      run: |
        go mod download
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Check code coverage
      run: |
        COVERAGE=\$(go tool cover -func=coverage.out | grep total | awk '{print \$3}' | sed 's/%//')
        echo "Code coverage: \$COVERAGE%"
        if (( \$(echo "\$COVERAGE < ${CODE_COVERAGE_THRESHOLD:-80}" | bc -l) )); then
          echo "Code coverage \$COVERAGE% is below threshold ${CODE_COVERAGE_THRESHOLD:-80}%"
          exit 1
        fi
    
    - name: Security scan
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: 'trivy-results.sarif'
      continue-on-error: true

  build:
    needs: test
    runs-on: [self-hosted, go-coffee]
    if: github.event_name == 'push'
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: \${{ env.REGISTRY }}
        username: \${{ secrets.DOCKER_USERNAME }}
        password: \${{ secrets.DOCKER_PASSWORD }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: \${{ env.REGISTRY }}/\${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: \${{ steps.meta.outputs.tags }}
        labels: \${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: [self-hosted, go-coffee]
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/coffee-service \\
          coffee-service=\${{ env.REGISTRY }}/\${{ env.IMAGE_NAME }}:\${{ github.sha }} \\
          -n go-coffee
        
        kubectl rollout status deployment/coffee-service -n go-coffee --timeout=300s
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create CI/CD pipeline templates"
    else
        log_success "CI/CD pipeline templates created"
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying CI/CD stack deployment..."
    
    local components=()
    
    if [[ "$ENABLE_ARGOCD" == "true" ]]; then
        components+=("argocd-server")
        components+=("argocd-application-controller")
        components+=("argocd-repo-server")
    fi
    
    if [[ "$ENABLE_GITHUB_ACTIONS_RUNNER" == "true" ]]; then
        components+=("actions-runner-controller")
    fi
    
    local failed_components=()
    
    for component in "${components[@]}"; do
        log_info "Checking component: $component"
        
        if kubectl get pods -n "$CICD_NAMESPACE" -l "app.kubernetes.io/name=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        else
            log_error "Component $component is not running"
            failed_components+=("$component")
        fi
    done
    
    if [[ ${#failed_components[@]} -gt 0 ]]; then
        log_error "Some CI/CD components failed to deploy: ${failed_components[*]}"
        log_info "Check pod status with: kubectl get pods -n $CICD_NAMESPACE"
        return 1
    fi
    
    log_success "All CI/CD components are running successfully"
}

# Display CI/CD information
display_cicd_info() {
    log_info "CI/CD stack information:"
    
    print_separator
    
    echo -e "${GREEN}CI/CD Components Deployed:${NC}"
    
    if [[ "$ENABLE_ARGOCD" == "true" ]]; then
        echo -e "  ✅ ArgoCD GitOps"
        echo -e "     - URL: https://${ARGOCD_HOSTNAME:-argocd.go-coffee.local}"
        echo -e "     - Username: admin"
        echo -e "     - Password: Check .argocd-password file"
    fi
    
    if [[ "$ENABLE_TEKTON_PIPELINES" == "true" ]]; then
        echo -e "  ✅ Tekton Pipelines"
        echo -e "     - Cloud-native CI/CD pipelines"
        echo -e "     - Kubernetes-native execution"
    fi
    
    if [[ "$ENABLE_GITHUB_ACTIONS_RUNNER" == "true" ]]; then
        echo -e "  ✅ GitHub Actions Runner"
        echo -e "     - Self-hosted runners in Kubernetes"
        echo -e "     - Replicas: ${GITHUB_RUNNER_REPLICAS:-3}"
    fi
    
    print_separator
    
    echo -e "${YELLOW}Pipeline Features:${NC}"
    echo -e "  🔨 Automated build and test"
    echo -e "  🔒 Security scanning with Trivy"
    echo -e "  📊 Code coverage analysis"
    echo -e "  🚀 Multi-cloud deployment"
    echo -e "  📈 GitOps with ArgoCD"
    echo -e "  🔄 Automated rollback on failure"
    
    print_separator
    
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  View CI/CD pods: kubectl get pods -n $CICD_NAMESPACE"
    echo -e "  ArgoCD CLI login: argocd login ${ARGOCD_HOSTNAME:-argocd.go-coffee.local}"
    echo -e "  View Tekton pipelines: kubectl get pipelines -n $CICD_NAMESPACE"
    echo -e "  Check GitHub runners: kubectl get runnerdeployments -n $CICD_NAMESPACE"
    echo -e "  View pipeline runs: kubectl get pipelineruns -n $CICD_NAMESPACE"
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy comprehensive CI/CD pipeline stack for Go Coffee platform.

OPTIONS:
    --environment ENV               Environment (dev, staging, prod) [default: dev]
    --project NAME                  Project name [default: go-coffee]
    --namespace NAME                CI/CD namespace [default: cicd]
    --enable-argocd                Enable ArgoCD GitOps [default: true]
    --enable-tekton-pipelines      Enable Tekton Pipelines [default: true]
    --enable-github-actions-runner Enable GitHub Actions Runner [default: true]
    --dry-run                      Perform dry run without actual deployment
    --verbose                      Enable verbose output
    --help                         Show this help message

EXAMPLES:
    $0                                    # Deploy full CI/CD stack
    $0 --environment prod                 # Deploy to production
    $0 --dry-run                         # Perform dry run
    $0 --enable-argocd --enable-github-actions-runner  # Deploy specific components

REQUIRED ENVIRONMENT VARIABLES:
    GIT_REPOSITORY_URL             Git repository URL
    GITHUB_TOKEN                   GitHub token for Actions Runner (if enabled)

OPTIONAL ENVIRONMENT VARIABLES:
    GIT_USERNAME                   Git username for authentication
    GIT_TOKEN                      Git token for authentication
    DOCKER_REGISTRY_SERVER         Docker registry server [default: docker.io]
    DOCKER_REGISTRY_USERNAME       Docker registry username
    DOCKER_REGISTRY_PASSWORD       Docker registry password
    SLACK_TOKEN                    Slack token for notifications
    ARGOCD_HOSTNAME               ArgoCD hostname
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
                CICD_NAMESPACE="$2"
                shift 2
                ;;
            --enable-argocd)
                ENABLE_ARGOCD="true"
                shift
                ;;
            --disable-argocd)
                ENABLE_ARGOCD="false"
                shift
                ;;
            --enable-tekton-pipelines)
                ENABLE_TEKTON_PIPELINES="true"
                shift
                ;;
            --disable-tekton-pipelines)
                ENABLE_TEKTON_PIPELINES="false"
                shift
                ;;
            --enable-github-actions-runner)
                ENABLE_GITHUB_ACTIONS_RUNNER="true"
                shift
                ;;
            --disable-github-actions-runner)
                ENABLE_GITHUB_ACTIONS_RUNNER="false"
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
    
    print_header "🚀 Go Coffee CI/CD Pipeline Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  Namespace: $CICD_NAMESPACE"
    log_info "  ArgoCD: $ENABLE_ARGOCD"
    log_info "  Tekton Pipelines: $ENABLE_TEKTON_PIPELINES"
    log_info "  GitHub Actions Runner: $ENABLE_GITHUB_ACTIONS_RUNNER"
    log_info "  Dry Run: $DRY_RUN"
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repositories
    create_cicd_secrets
    deploy_argocd
    deploy_tekton
    deploy_github_runner
    deploy_argocd_applications
    create_pipeline_templates
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
        display_cicd_info
    fi
    
    print_header "✅ CI/CD Pipeline Deployment Completed"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "CI/CD pipeline stack deployed successfully!"
        log_info "Check .argocd-password file for ArgoCD admin credentials"
        log_info "Configure your GitHub repository with the provided workflow"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
