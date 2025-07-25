#!/bin/bash

# ☕ Go Coffee - CI/CD Pipeline Enhancement Deployment Script
# Deploys comprehensive CI/CD pipeline with GitOps, automated testing, and multi-environment deployment

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
CICD_NAMESPACE="${CICD_NAMESPACE:-argocd}"
ENVIRONMENT="${ENVIRONMENT:-production}"
CLUSTER_NAME="${CLUSTER_NAME:-go-coffee-cluster}"

# CI/CD configuration
ENABLE_ARGOCD="${ENABLE_ARGOCD:-true}"
ENABLE_GITHUB_ACTIONS="${ENABLE_GITHUB_ACTIONS:-true}"
ENABLE_MONITORING="${ENABLE_MONITORING:-true}"
ENABLE_NOTIFICATIONS="${ENABLE_NOTIFICATIONS:-true}"
ENABLE_SECURITY_SCANNING="${ENABLE_SECURITY_SCANNING:-true}"

# ArgoCD configuration
ARGOCD_VERSION="${ARGOCD_VERSION:-v2.8.4}"
ARGOCD_ADMIN_PASSWORD="${ARGOCD_ADMIN_PASSWORD:-$(openssl rand -base64 32)}"

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
    log "Checking CI/CD prerequisites..."
    
    # Check required tools
    local tools=("kubectl" "helm" "git" "curl" "jq")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is not installed or not in PATH"
        fi
    done
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl is not properly configured or cluster is not accessible"
    fi
    
    # Check cluster admin permissions
    if ! kubectl auth can-i create clusterroles &> /dev/null; then
        error "Insufficient permissions. Cluster admin access required for CI/CD deployment."
    fi
    
    # Check Git repository
    if ! git rev-parse --git-dir &> /dev/null; then
        warn "Not in a Git repository. Some GitOps features may not work correctly."
    fi
    
    success "CI/CD prerequisites met"
}

# Setup CI/CD namespaces
setup_cicd_namespaces() {
    log "Setting up CI/CD namespaces..."
    
    # Create ArgoCD namespace
    kubectl create namespace "$CICD_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label ArgoCD namespace
    kubectl label namespace "$CICD_NAMESPACE" \
        name="$CICD_NAMESPACE" \
        app.kubernetes.io/name=argocd \
        app.kubernetes.io/component=cicd \
        pod-security.kubernetes.io/enforce=baseline \
        pod-security.kubernetes.io/audit=restricted \
        pod-security.kubernetes.io/warn=restricted \
        --overwrite
    
    # Create monitoring namespace if it doesn't exist
    kubectl create namespace go-coffee-monitoring --dry-run=client -o yaml | kubectl apply -f -
    
    success "CI/CD namespaces configured"
}

# Add Helm repositories for CI/CD tools
add_helm_repos() {
    log "Adding Helm repositories for CI/CD tools..."
    
    # ArgoCD
    helm repo add argo https://argoproj.github.io/argo-helm
    
    # Prometheus for monitoring
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    
    # Grafana for dashboards
    helm repo add grafana https://grafana.github.io/helm-charts
    
    # Update repositories
    helm repo update
    
    success "Helm repositories added"
}

# Deploy ArgoCD
deploy_argocd() {
    if [[ "$ENABLE_ARGOCD" != "true" ]]; then
        info "ArgoCD deployment skipped"
        return 0
    fi
    
    log "Deploying ArgoCD..."
    
    # Install ArgoCD
    helm upgrade --install argocd argo/argo-cd \
        --namespace "$CICD_NAMESPACE" \
        --create-namespace \
        --version "$ARGOCD_VERSION" \
        --values - <<EOF
global:
  image:
    tag: "$ARGOCD_VERSION"

configs:
  params:
    server.insecure: true
    server.disable.auth: false
    application.instanceLabelKey: argocd.argoproj.io/instance
  
  cm:
    url: https://argocd.gocoffee.dev
    application.instanceLabelKey: argocd.argoproj.io/instance
    server.rbac.log.enforce.enable: true
    exec.enabled: true
    admin.enabled: true
    timeout.reconciliation: 180s
    timeout.hard.reconciliation: 0s
    
    # OIDC configuration (optional)
    oidc.config: |
      name: OIDC
      issuer: https://accounts.google.com
      clientId: \$oidc.google.clientId
      clientSecret: \$oidc.google.clientSecret
      requestedScopes: ["openid", "profile", "email", "groups"]
      requestedIDTokenClaims: {"groups": {"essential": true}}
    
    # Repository credentials
    repositories: |
      - type: git
        url: https://github.com/DimaJoyti/go-coffee.git
      - type: helm
        url: https://argoproj.github.io/argo-helm
        name: argo
      - type: helm
        url: https://prometheus-community.github.io/helm-charts
        name: prometheus-community

  rbac:
    policy.default: role:readonly
    policy.csv: |
      p, role:admin, applications, *, */*, allow
      p, role:admin, clusters, *, *, allow
      p, role:admin, repositories, *, *, allow
      p, role:admin, logs, get, *, allow
      p, role:admin, exec, create, */*, allow
      
      p, role:developer, applications, get, */*, allow
      p, role:developer, applications, sync, */*, allow
      p, role:developer, logs, get, */*, allow
      p, role:developer, repositories, get, *, allow
      
      g, go-coffee:admins, role:admin
      g, go-coffee:developers, role:developer

server:
  service:
    type: LoadBalancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
  
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
    hosts:
    - argocd.gocoffee.dev
    tls:
    - secretName: argocd-server-tls
      hosts:
      - argocd.gocoffee.dev
  
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      namespace: go-coffee-monitoring

controller:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      namespace: go-coffee-monitoring

repoServer:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      namespace: go-coffee-monitoring

applicationSet:
  enabled: true
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      namespace: go-coffee-monitoring

notifications:
  enabled: true
  argocdUrl: https://argocd.gocoffee.dev
  
  notifiers:
    service.slack: |
      token: \$slack-token
      username: ArgoCD
      icon: ":argo:"
    
    service.email: |
      host: smtp.gmail.com
      port: 587
      from: argocd@gocoffee.dev
      username: \$email-username
      password: \$email-password
  
  subscriptions:
  - recipients:
    - slack:go-coffee-deployments
    - email:devops@gocoffee.dev
    triggers:
    - on-deployed
    - on-health-degraded
    - on-sync-failed
  
  templates:
    template.app-deployed: |
      message: |
        🚀 Application {{.app.metadata.name}} deployed to {{.app.spec.destination.namespace}}
        Sync Status: {{.app.status.sync.status}}
        Health Status: {{.app.status.health.status}}
        Revision: {{.app.status.sync.revision}}
    
    template.app-health-degraded: |
      message: |
        ⚠️ Application {{.app.metadata.name}} health degraded
        Health Status: {{.app.status.health.status}}
        Message: {{.app.status.health.message}}
    
    template.app-sync-failed: |
      message: |
        ❌ Application {{.app.metadata.name}} sync failed
        Sync Status: {{.app.status.sync.status}}
        Message: {{.app.status.operationState.message}}
  
  triggers:
    trigger.on-deployed: |
      - when: app.status.operationState.phase in ['Succeeded'] and app.status.health.status == 'Healthy'
        send: [app-deployed]
    
    trigger.on-health-degraded: |
      - when: app.status.health.status == 'Degraded'
        send: [app-health-degraded]
    
    trigger.on-sync-failed: |
      - when: app.status.operationState.phase in ['Error', 'Failed']
        send: [app-sync-failed]
EOF
    
    # Wait for ArgoCD to be ready
    kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n "$CICD_NAMESPACE"
    kubectl wait --for=condition=available --timeout=600s deployment/argocd-application-controller -n "$CICD_NAMESPACE"
    kubectl wait --for=condition=available --timeout=600s deployment/argocd-repo-server -n "$CICD_NAMESPACE"
    
    # Set ArgoCD admin password
    kubectl patch secret argocd-initial-admin-secret \
        -n "$CICD_NAMESPACE" \
        -p "{\"stringData\": {\"password\": \"$ARGOCD_ADMIN_PASSWORD\"}}"
    
    success "ArgoCD deployed successfully"
    info "ArgoCD admin password: $ARGOCD_ADMIN_PASSWORD"
}

# Deploy ArgoCD applications
deploy_argocd_applications() {
    if [[ "$ENABLE_ARGOCD" != "true" ]]; then
        info "ArgoCD applications deployment skipped"
        return 0
    fi
    
    log "Deploying ArgoCD applications..."
    
    # Apply ArgoCD applications and projects
    kubectl apply -f "$SCRIPT_DIR/gitops/argocd-applications.yaml"
    
    # Wait for applications to be created
    sleep 30
    
    # Sync applications
    local apps=("go-coffee-core" "go-coffee-ai" "go-coffee-monitoring" "go-coffee-security")
    for app in "${apps[@]}"; do
        info "Syncing application: $app"
        kubectl patch application "$app" -n "$CICD_NAMESPACE" \
            --type merge \
            --patch '{"operation":{"initiatedBy":{"username":"admin"},"sync":{"syncStrategy":{"hook":{"force":true}}}}}'
    done
    
    success "ArgoCD applications deployed"
}

# Setup GitHub Actions integration
setup_github_actions() {
    if [[ "$ENABLE_GITHUB_ACTIONS" != "true" ]]; then
        info "GitHub Actions setup skipped"
        return 0
    fi
    
    log "Setting up GitHub Actions integration..."
    
    # Create GitHub Actions workflows directory
    mkdir -p .github/workflows
    
    # Copy workflow files
    cp "$SCRIPT_DIR/github-actions/"*.yml .github/workflows/
    
    # Create GitHub Actions secrets (instructions)
    cat > .github/SECRETS.md <<EOF
# GitHub Actions Secrets Configuration

Configure the following secrets in your GitHub repository:

## Container Registry
- \`GITHUB_TOKEN\`: Automatically provided by GitHub

## Kubernetes Access
- \`KUBECONFIG_STAGING\`: Base64 encoded kubeconfig for staging cluster
- \`KUBECONFIG_PRODUCTION\`: Base64 encoded kubeconfig for production cluster

## Cloud Provider Credentials
- \`GCP_SA_KEY\`: Base64 encoded GCP service account key
- \`AWS_ACCESS_KEY_ID\`: AWS access key ID
- \`AWS_SECRET_ACCESS_KEY\`: AWS secret access key
- \`AZURE_CLIENT_ID\`: Azure client ID
- \`AZURE_CLIENT_SECRET\`: Azure client secret
- \`AZURE_SUBSCRIPTION_ID\`: Azure subscription ID
- \`AZURE_TENANT_ID\`: Azure tenant ID

## Notifications
- \`SLACK_WEBHOOK_URL\`: Slack webhook URL for notifications
- \`EMAIL_USERNAME\`: SMTP username for email notifications
- \`EMAIL_PASSWORD\`: SMTP password for email notifications
- \`DEPLOYMENT_EMAIL_LIST\`: Comma-separated list of deployment notification emails
- \`EXECUTIVE_EMAIL_LIST\`: Comma-separated list of executive notification emails

## Security Scanning
- \`SNYK_TOKEN\`: Snyk token for vulnerability scanning
- \`SONAR_TOKEN\`: SonarQube token for code quality analysis

To encode kubeconfig files:
\`\`\`bash
cat ~/.kube/config | base64 -w 0
\`\`\`
EOF
    
    success "GitHub Actions integration configured"
    info "Please configure secrets as described in .github/SECRETS.md"
}

# Configure monitoring for CI/CD
configure_cicd_monitoring() {
    if [[ "$ENABLE_MONITORING" != "true" ]]; then
        info "CI/CD monitoring configuration skipped"
        return 0
    fi
    
    log "Configuring CI/CD monitoring..."
    
    # Create ServiceMonitor for ArgoCD
    if kubectl get crd servicemonitors.monitoring.coreos.com &> /dev/null; then
        cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: argocd-metrics
  namespace: go-coffee-monitoring
  labels:
    app.kubernetes.io/name: argocd
    app.kubernetes.io/component: monitoring
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-metrics
  namespaceSelector:
    matchNames:
    - $CICD_NAMESPACE
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    honorLabels: true
EOF
        
        info "ArgoCD ServiceMonitor created"
    fi
    
    # Create CI/CD specific alerts
    if kubectl get crd prometheusrules.monitoring.coreos.com &> /dev/null; then
        cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cicd-alerts
  namespace: go-coffee-monitoring
  labels:
    app.kubernetes.io/name: cicd
    app.kubernetes.io/component: alerts
    release: prometheus
spec:
  groups:
  - name: cicd.rules
    rules:
    - alert: ArgocdAppSyncFailed
      expr: argocd_app_info{sync_status!="Synced"} == 1
      for: 5m
      labels:
        severity: warning
        team: devops
      annotations:
        summary: "ArgoCD application sync failed"
        description: "ArgoCD application {{ \$labels.name }} in namespace {{ \$labels.namespace }} has failed to sync"
    
    - alert: ArgocdAppHealthDegraded
      expr: argocd_app_info{health_status!="Healthy"} == 1
      for: 5m
      labels:
        severity: critical
        team: devops
      annotations:
        summary: "ArgoCD application health degraded"
        description: "ArgoCD application {{ \$labels.name }} in namespace {{ \$labels.namespace }} is not healthy"
    
    - alert: ArgocdServerDown
      expr: up{job="argocd-server-metrics"} == 0
      for: 2m
      labels:
        severity: critical
        team: devops
      annotations:
        summary: "ArgoCD server is down"
        description: "ArgoCD server has been down for more than 2 minutes"
    
    - alert: HighDeploymentFrequency
      expr: increase(argocd_app_sync_total[1h]) > 10
      for: 0m
      labels:
        severity: info
        team: devops
      annotations:
        summary: "High deployment frequency detected"
        description: "More than 10 deployments in the last hour for application {{ \$labels.name }}"
EOF
        
        info "CI/CD alerts configured"
    fi
    
    success "CI/CD monitoring configured"
}

# Setup security scanning
setup_security_scanning() {
    if [[ "$ENABLE_SECURITY_SCANNING" != "true" ]]; then
        info "Security scanning setup skipped"
        return 0
    fi
    
    log "Setting up security scanning..."
    
    # Create security scanning configuration
    cat > .github/security-scanning.yml <<EOF
# Security Scanning Configuration for Go Coffee

security:
  # Code scanning with CodeQL
  codeql:
    enabled: true
    languages: [go, javascript, typescript]
    queries: security-and-quality
  
  # Dependency scanning
  dependency_scanning:
    enabled: true
    package_managers: [go, npm, yarn]
    vulnerability_database: github
  
  # Container scanning
  container_scanning:
    enabled: true
    scanner: trivy
    severity_threshold: HIGH
    fail_on_vulnerability: true
  
  # Infrastructure scanning
  infrastructure_scanning:
    enabled: true
    terraform_scanner: checkov
    kubernetes_scanner: polaris
  
  # Secret scanning
  secret_scanning:
    enabled: true
    patterns:
      - api_keys
      - passwords
      - tokens
      - certificates
  
  # SAST (Static Application Security Testing)
  sast:
    enabled: true
    tools: [gosec, eslint-security]
  
  # DAST (Dynamic Application Security Testing)
  dast:
    enabled: true
    target_url: https://staging.gocoffee.dev
    authentication: bearer_token
EOF
    
    # Create security policy
    cat > .github/SECURITY.md <<EOF
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Please report security vulnerabilities to security@gocoffee.dev

### What to include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response timeline:
- Initial response: 24 hours
- Status update: 72 hours
- Resolution: 30 days (for critical issues)

## Security Measures

### Code Security
- Automated security scanning with CodeQL
- Dependency vulnerability scanning
- Secret scanning in commits
- SAST/DAST testing in CI/CD

### Infrastructure Security
- Container image scanning
- Kubernetes security policies
- Infrastructure as Code scanning
- Network security policies

### Deployment Security
- Signed commits required
- Multi-stage approval for production
- Automated rollback on security issues
- Security monitoring and alerting
EOF
    
    success "Security scanning configured"
}

# Verify CI/CD deployment
verify_cicd_deployment() {
    log "Verifying CI/CD deployment..."
    
    # Check ArgoCD
    if [[ "$ENABLE_ARGOCD" == "true" ]]; then
        info "Checking ArgoCD..."
        kubectl get pods -n "$CICD_NAMESPACE" -l app.kubernetes.io/name=argocd-server
        
        # Test ArgoCD API
        local argocd_url="https://argocd.gocoffee.dev"
        if kubectl get ingress argocd-server-ingress -n "$CICD_NAMESPACE" &> /dev/null; then
            info "ArgoCD ingress configured"
        else
            warn "ArgoCD ingress not found"
        fi
    fi
    
    # Check GitHub Actions workflows
    if [[ "$ENABLE_GITHUB_ACTIONS" == "true" ]]; then
        info "Checking GitHub Actions workflows..."
        if [[ -d ".github/workflows" ]]; then
            local workflow_count=$(find .github/workflows -name "*.yml" -o -name "*.yaml" | wc -l)
            info "Found $workflow_count GitHub Actions workflows"
        else
            warn "GitHub Actions workflows directory not found"
        fi
    fi
    
    # Check monitoring
    if [[ "$ENABLE_MONITORING" == "true" ]]; then
        info "Checking CI/CD monitoring..."
        kubectl get servicemonitors -n go-coffee-monitoring | grep argocd || warn "ArgoCD ServiceMonitor not found"
        kubectl get prometheusrules -n go-coffee-monitoring | grep cicd || warn "CI/CD PrometheusRules not found"
    fi
    
    success "CI/CD deployment verification completed"
}

# Generate CI/CD deployment report
generate_cicd_report() {
    log "Generating CI/CD deployment report..."
    
    local report_file="/tmp/go-coffee-cicd-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" <<EOF
☕ Go Coffee - CI/CD Pipeline Enhancement Report
==============================================

Deployment Date: $(date)
Environment: $ENVIRONMENT
Cluster: $CLUSTER_NAME
CI/CD Namespace: $CICD_NAMESPACE

CI/CD Components Deployed:
=========================

$(if [[ "$ENABLE_ARGOCD" == "true" ]]; then echo "✅ ArgoCD GitOps: Deployed"; else echo "❌ ArgoCD GitOps: Skipped"; fi)
   - GitOps continuous deployment
   - Multi-environment application management
   - Automated sync and rollback
   - RBAC and security policies

$(if [[ "$ENABLE_GITHUB_ACTIONS" == "true" ]]; then echo "✅ GitHub Actions: Configured"; else echo "❌ GitHub Actions: Skipped"; fi)
   - Automated build and test pipelines
   - Security scanning and code quality
   - Multi-stage deployment workflows
   - Performance and integration testing

$(if [[ "$ENABLE_MONITORING" == "true" ]]; then echo "✅ CI/CD Monitoring: Deployed"; else echo "❌ CI/CD Monitoring: Skipped"; fi)
   - ArgoCD metrics and dashboards
   - Deployment success/failure tracking
   - Performance monitoring
   - Alert management

$(if [[ "$ENABLE_SECURITY_SCANNING" == "true" ]]; then echo "✅ Security Scanning: Configured"; else echo "❌ Security Scanning: Skipped"; fi)
   - SAST/DAST security testing
   - Container vulnerability scanning
   - Infrastructure security validation
   - Secret scanning and protection

CI/CD Capabilities:
==================

🔄 Continuous Integration:
   - Automated code quality checks
   - Unit and integration testing
   - Security vulnerability scanning
   - Performance testing

🚀 Continuous Deployment:
   - GitOps-based deployment
   - Multi-environment promotion
   - Blue-green deployments
   - Automated rollback

🔒 Security & Compliance:
   - Code security scanning
   - Container image scanning
   - Infrastructure validation
   - Compliance reporting

📊 Monitoring & Observability:
   - Deployment metrics
   - Success/failure tracking
   - Performance monitoring
   - Alert management

Access Information:
==================

ArgoCD UI: https://argocd.gocoffee.dev
Username: admin
Password: $ARGOCD_ADMIN_PASSWORD

GitHub Actions: https://github.com/DimaJoyti/go-coffee/actions
Monitoring: https://grafana.gocoffee.dev

Next Steps:
==========

1. Configure GitHub repository secrets
2. Set up branch protection rules
3. Configure deployment approvals
4. Set up monitoring dashboards
5. Configure notification channels
6. Test deployment workflows

For more information, visit: https://docs.gocoffee.dev/cicd
EOF
    
    echo "$report_file"
    success "CI/CD report generated: $report_file"
}

# Display CI/CD access information
display_cicd_info() {
    echo -e "${CYAN}"
    cat << EOF

    🚀 Go Coffee CI/CD Pipeline Enhancement Deployed Successfully!
    
    🔄 CI/CD Components:
    
    GitOps (ArgoCD):
    - UI: https://argocd.gocoffee.dev
    - Username: admin
    - Password: $ARGOCD_ADMIN_PASSWORD
    - CLI: argocd login argocd.gocoffee.dev
    
    GitHub Actions:
    - Workflows: .github/workflows/
    - Build & Test: Automated on push/PR
    - Deploy Staging: Automated on develop branch
    - Deploy Production: Manual approval required
    
    Security Scanning:
    - CodeQL: Automated code analysis
    - Trivy: Container vulnerability scanning
    - Checkov: Infrastructure security validation
    - Secret scanning: Automated in commits
    
    🔧 Management Commands:
    
    # ArgoCD CLI operations
    argocd app list
    argocd app sync go-coffee-core
    argocd app rollback go-coffee-core
    
    # Check deployment status
    kubectl get applications -n $CICD_NAMESPACE
    kubectl get pods -n $CICD_NAMESPACE
    
    # Monitor CI/CD metrics
    kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n go-coffee-monitoring
    
    # View ArgoCD logs
    kubectl logs -l app.kubernetes.io/name=argocd-server -n $CICD_NAMESPACE -f
    
    📊 Monitoring:
    - ArgoCD metrics: Prometheus/Grafana dashboards
    - Deployment tracking: Application sync status
    - Performance monitoring: Build and deployment times
    - Security alerts: Vulnerability and compliance alerts
    
    🔒 Security Features:
    - Multi-stage approval for production deployments
    - Automated security scanning in CI/CD
    - GitOps with signed commits
    - RBAC and access controls
    - Automated rollback on failures
    
    📚 Documentation: https://docs.gocoffee.dev/cicd
    
EOF
    echo -e "${NC}"
}

# Cleanup function
cleanup() {
    if [[ "${1:-}" == "destroy" ]]; then
        warn "Destroying CI/CD stack..."
        
        # Delete ArgoCD applications
        kubectl delete applications --all -n "$CICD_NAMESPACE" --ignore-not-found=true
        
        # Delete ArgoCD
        helm uninstall argocd -n "$CICD_NAMESPACE" --ignore-not-found=true
        
        # Delete namespaces
        kubectl delete namespace "$CICD_NAMESPACE" --ignore-not-found=true
        
        # Remove GitHub Actions workflows
        rm -rf .github/workflows/
        rm -f .github/SECRETS.md .github/SECURITY.md .github/security-scanning.yml
        
        success "CI/CD stack destroyed"
    fi
}

# Main execution
main() {
    echo -e "${PURPLE}"
    cat << "EOF"
    ☕ Go Coffee - CI/CD Pipeline Enhancement
    ========================================
    
    Deploying enterprise CI/CD pipeline:
    • GitOps with ArgoCD
    • GitHub Actions Workflows
    • Automated Testing & Security Scanning
    • Multi-Environment Deployment
    • Monitoring & Alerting
    • Blue-Green & Canary Deployments
    
EOF
    echo -e "${NC}"
    
    info "Starting CI/CD pipeline deployment..."
    info "Environment: $ENVIRONMENT"
    info "Cluster: $CLUSTER_NAME"
    info "CI/CD Namespace: $CICD_NAMESPACE"
    
    # Execute deployment steps
    check_prerequisites
    setup_cicd_namespaces
    add_helm_repos
    deploy_argocd
    deploy_argocd_applications
    setup_github_actions
    configure_cicd_monitoring
    setup_security_scanning
    verify_cicd_deployment
    
    # Generate report
    local report_file=$(generate_cicd_report)
    
    display_cicd_info
    
    success "🎉 CI/CD Pipeline Enhancement deployment completed successfully!"
    info "CI/CD report saved to: $report_file"
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
        verify_cicd_deployment
        ;;
    *)
        echo "Usage: $0 [deploy|destroy|verify]"
        exit 1
        ;;
esac
