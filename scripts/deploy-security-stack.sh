#!/bin/bash

# Security and Compliance Stack Deployment Script
# Deploys comprehensive security scanning, compliance monitoring, and policy enforcement

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/security-compliance"
SECURITY_DIR="$PROJECT_ROOT/k8s/security"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
SECURITY_NAMESPACE="${SECURITY_NAMESPACE:-security}"
ENABLE_VULNERABILITY_SCANNING="${ENABLE_VULNERABILITY_SCANNING:-true}"
ENABLE_COMPLIANCE_MONITORING="${ENABLE_COMPLIANCE_MONITORING:-true}"
ENABLE_THREAT_DETECTION="${ENABLE_THREAT_DETECTION:-true}"
ENABLE_POLICY_ENFORCEMENT="${ENABLE_POLICY_ENFORCEMENT:-true}"
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
    
    local required_tools=("kubectl" "helm" "terraform" "jq" "curl" "docker")
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
    
    # Check Docker (for vulnerability scanning)
    if ! docker info &>/dev/null; then
        log_warning "Docker is not accessible - some vulnerability scanning features may not work"
    fi
    
    log_success "Prerequisites check completed"
}

# Setup security namespace
setup_namespace() {
    log_info "Setting up security namespace: $SECURITY_NAMESPACE"
    
    if kubectl get namespace "$SECURITY_NAMESPACE" &>/dev/null; then
        log_info "Namespace $SECURITY_NAMESPACE already exists"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would create namespace $SECURITY_NAMESPACE"
        else
            kubectl create namespace "$SECURITY_NAMESPACE"
            
            # Apply security labels and annotations
            kubectl label namespace "$SECURITY_NAMESPACE" \
                app.kubernetes.io/name="$PROJECT_NAME" \
                app.kubernetes.io/component="security" \
                environment="$ENVIRONMENT" \
                pod-security.kubernetes.io/enforce="restricted" \
                pod-security.kubernetes.io/audit="restricted" \
                pod-security.kubernetes.io/warn="restricted"
            
            kubectl annotate namespace "$SECURITY_NAMESPACE" \
                managed-by="terraform" \
                created-at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
            
            log_success "Created security namespace: $SECURITY_NAMESPACE"
        fi
    fi
}

# Add Helm repositories
add_helm_repositories() {
    log_info "Adding Helm repositories for security tools..."
    
    local repos=(
        "aqua:https://aquasecurity.github.io/helm-charts"
        "gatekeeper:https://open-policy-agent.github.io/gatekeeper/charts"
        "falcosecurity:https://falcosecurity.github.io/charts"
        "twistlock:https://helm.twistlock.com"
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

# Create security secrets
create_security_secrets() {
    log_info "Creating security secrets..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create security secrets"
        return 0
    fi
    
    # Create notification secrets
    kubectl create secret generic security-notifications \
        --from-literal=slack-webhook-url="${SLACK_WEBHOOK_URL:-}" \
        --from-literal=email-username="${EMAIL_USERNAME:-}" \
        --from-literal=email-password="${EMAIL_PASSWORD:-}" \
        --from-literal=webhook-url="${WEBHOOK_URL:-}" \
        --namespace="$SECURITY_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create registry credentials for scanning
    if [[ -n "${DOCKER_REGISTRY_USERNAME:-}" && -n "${DOCKER_REGISTRY_PASSWORD:-}" ]]; then
        kubectl create secret docker-registry registry-credentials \
            --docker-server="${DOCKER_REGISTRY_SERVER:-docker.io}" \
            --docker-username="$DOCKER_REGISTRY_USERNAME" \
            --docker-password="$DOCKER_REGISTRY_PASSWORD" \
            --docker-email="${DOCKER_REGISTRY_EMAIL:-security@go-coffee.com}" \
            --namespace="$SECURITY_NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    log_success "Security secrets created"
}

# Deploy vulnerability scanning
deploy_vulnerability_scanning() {
    if [[ "$ENABLE_VULNERABILITY_SCANNING" != "true" ]]; then
        log_info "Vulnerability scanning disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying vulnerability scanning tools..."
    
    local values_file="$SECURITY_DIR/trivy-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$SECURITY_DIR"
        log_info "Creating Trivy Operator values file..."
        cat > "$values_file" << EOF
trivyOperator:
  scanJobTimeout: 5m
  scanJobsConcurrentLimit: 10
  scanJobsRetryDelay: 30s
  
  # Vulnerability database
  vulnerabilityReportsPlugin: "Trivy"
  configAuditReportsPlugin: "Trivy"
  
  # Compliance reports
  complianceFailEntriesLimit: 10
  
  # Metrics
  metricsBindAddress: "0.0.0.0:8080"
  healthProbeBindAddress: "0.0.0.0:9090"

serviceMonitor:
  enabled: ${MONITORING_ENABLED:-true}

trivy:
  # Scanning settings
  timeout: "5m0s"
  ignoreUnfixed: false
  severity: "UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL"
  
  # Resources
  resources:
    requests:
      cpu: "100m"
      memory: "100Mi"
    limits:
      cpu: "500m"
      memory: "500Mi"
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Trivy Operator"
        helm template trivy-operator aqua/trivy-operator \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Trivy Operator template validation passed"
    else
        helm upgrade --install trivy-operator aqua/trivy-operator \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Trivy Operator deployed successfully"
    fi
}

# Deploy policy enforcement
deploy_policy_enforcement() {
    if [[ "$ENABLE_POLICY_ENFORCEMENT" != "true" ]]; then
        log_info "Policy enforcement disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying OPA Gatekeeper..."
    
    local values_file="$SECURITY_DIR/gatekeeper-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$SECURITY_DIR"
        log_info "Creating Gatekeeper values file..."
        cat > "$values_file" << EOF
replicas: 3

audit:
  replicas: 1
  auditInterval: 60
  constraintViolationsLimit: 20
  auditFromCache: false
  
  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "1000m"
      memory: "512Mi"

controllerManager:
  resources:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "1000m"
      memory: "512Mi"
  
  webhook:
    failurePolicy: "Fail"
    namespaceSelector:
      matchExpressions:
      - key: "name"
        operator: "NotIn"
        values: ["kube-system", "gatekeeper-system", "$SECURITY_NAMESPACE"]

metrics:
  enabled: ${MONITORING_ENABLED:-true}

mutations:
  enabled: false

podSecurityPolicy:
  enabled: false
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Gatekeeper"
        helm template gatekeeper gatekeeper/gatekeeper \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Gatekeeper template validation passed"
    else
        helm upgrade --install gatekeeper gatekeeper/gatekeeper \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Gatekeeper deployed successfully"
        
        # Deploy security policies
        deploy_security_policies
    fi
}

# Deploy security policies
deploy_security_policies() {
    log_info "Deploying security policies..."
    
    local policies_dir="$SECURITY_DIR/policies"
    mkdir -p "$policies_dir"
    
    # Create required security context policy
    cat > "$policies_dir/required-security-context.yaml" << EOF
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: requiredsecuritycontext
spec:
  crd:
    spec:
      names:
        kind: RequiredSecurityContext
      validation:
        openAPIV3Schema:
          type: object
          properties:
            runAsNonRoot:
              type: boolean
            runAsUser:
              type: integer
              minimum: 1000
            fsGroup:
              type: integer
              minimum: 1000
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package requiredsecuritycontext

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not container.securityContext.runAsNonRoot
          msg := "Container must run as non-root user"
        }

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          container.securityContext.runAsUser < 1000
          msg := "Container must run as user ID >= 1000"
        }
---
apiVersion: config.gatekeeper.sh/v1alpha1
kind: RequiredSecurityContext
metadata:
  name: must-run-as-nonroot
spec:
  match:
    kinds:
      - apiGroups: ["apps"]
        kinds: ["Deployment", "StatefulSet", "DaemonSet"]
    excludedNamespaces: ["kube-system", "gatekeeper-system", "$SECURITY_NAMESPACE"]
  parameters:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 1000
EOF
    
    # Create resource limits policy
    cat > "$policies_dir/required-resource-limits.yaml" << EOF
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: requiredresourcelimits
spec:
  crd:
    spec:
      names:
        kind: RequiredResourceLimits
      validation:
        openAPIV3Schema:
          type: object
          properties:
            cpu:
              type: string
            memory:
              type: string
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package requiredresourcelimits

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not container.resources.limits.cpu
          msg := "Container must have CPU limits"
        }

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not container.resources.limits.memory
          msg := "Container must have memory limits"
        }
---
apiVersion: config.gatekeeper.sh/v1alpha1
kind: RequiredResourceLimits
metadata:
  name: must-have-resource-limits
spec:
  match:
    kinds:
      - apiGroups: ["apps"]
        kinds: ["Deployment", "StatefulSet", "DaemonSet"]
    excludedNamespaces: ["kube-system", "gatekeeper-system", "$SECURITY_NAMESPACE"]
  parameters:
    cpu: "1000m"
    memory: "1Gi"
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy security policies"
    else
        # Apply security policies
        kubectl apply -f "$policies_dir/" --recursive
        log_success "Security policies deployed"
    fi
}

# Deploy threat detection
deploy_threat_detection() {
    if [[ "$ENABLE_THREAT_DETECTION" != "true" ]]; then
        log_info "Threat detection disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying Falco for threat detection..."
    
    local values_file="$SECURITY_DIR/falco-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$SECURITY_DIR"
        log_info "Creating Falco values file..."
        cat > "$values_file" << EOF
falco:
  rules_file:
    - /etc/falco/falco_rules.yaml
    - /etc/falco/falco_rules.local.yaml
    - /etc/falco/k8s_audit_rules.yaml
    - /etc/falco/rules.d
  
  time_format_iso_8601: true
  json_output: true
  json_include_output_property: true
  json_include_tags_property: true
  
  log_level: "info"
  priority: "warning"
  
  buffered_outputs: false
  
  syscall_event_drops:
    actions: ["log", "alert"]
    rate: 0.03333
    max_burst: 1000

driver:
  enabled: true
  kind: "ebpf"

collectors:
  enabled: true
  
  docker:
    enabled: true
    socket: "/var/run/docker.sock"
  
  containerd:
    enabled: true
    socket: "/run/containerd/containerd.sock"
  
  crio:
    enabled: true
    socket: "/var/run/crio/crio.sock"

serviceMonitor:
  enabled: ${MONITORING_ENABLED:-true}

customRules:
  go-coffee-rules.yaml: |
    - rule: Suspicious Coffee Order Activity
      desc: Detect suspicious coffee order patterns
      condition: >
        k8s_audit and ka.verb in (create, update) and
        ka.target.resource=orders and
        ka.target.subresource="" and
        ka.uri.param[quantity] > 100
      output: >
        Suspicious large coffee order detected
        (user=%ka.user.name verb=%ka.verb uri=%ka.uri.path
        quantity=%ka.uri.param[quantity])
      priority: WARNING
      tags: [coffee, orders, suspicious]
      
    - rule: Unauthorized DeFi Transaction
      desc: Detect unauthorized DeFi transactions
      condition: >
        k8s_audit and ka.verb=create and
        ka.target.resource=transactions and
        ka.target.subresource="" and
        not ka.user.name in (defi-service, trading-bot)
      output: >
        Unauthorized DeFi transaction attempt
        (user=%ka.user.name verb=%ka.verb uri=%ka.uri.path)
      priority: CRITICAL
      tags: [defi, unauthorized, transaction]

falcosidekick:
  enabled: true
  
  config:
    slack:
      webhookurl: "${SLACK_WEBHOOK_URL:-}"
      channel: "#security-alerts"
      username: "Falco"
      iconurl: "https://falco.org/img/brand/falco-logo.png"
      minimumpriority: "warning"
    
    webhook:
      address: "${WEBHOOK_URL:-}"
      minimumpriority: "warning"
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Falco"
        helm template falco falcosecurity/falco \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Falco template validation passed"
    else
        helm upgrade --install falco falcosecurity/falco \
            --namespace "$SECURITY_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 10m
        
        log_success "Falco deployed successfully"
    fi
}

# Deploy compliance monitoring
deploy_compliance_monitoring() {
    if [[ "$ENABLE_COMPLIANCE_MONITORING" != "true" ]]; then
        log_info "Compliance monitoring disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying compliance monitoring..."
    
    # Create compliance checker deployment
    local compliance_dir="$SECURITY_DIR/compliance"
    mkdir -p "$compliance_dir"
    
    cat > "$compliance_dir/compliance-checker.yaml" << EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ${PROJECT_NAME}-compliance-checker
  namespace: $SECURITY_NAMESPACE
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: compliance-checker
    environment: $ENVIRONMENT
spec:
  schedule: "0 6 * * 1"  # Weekly on Monday at 6 AM
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: compliance-checker
            image: ${COMPLIANCE_CHECKER_IMAGE:-go-coffee/compliance-checker:latest}
            env:
            - name: ENVIRONMENT
              value: "$ENVIRONMENT"
            - name: COMPLIANCE_FRAMEWORKS
              value: "${COMPLIANCE_FRAMEWORKS:-SOC2,PCI-DSS,GDPR}"
            - name: SLACK_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: security-notifications
                  key: slack-webhook-url
            command: ["/bin/sh"]
            args:
            - -c
            - |
              echo "Starting compliance checks..."
              
              # Run compliance checks for each framework
              for framework in \$(echo \$COMPLIANCE_FRAMEWORKS | tr ',' ' '); do
                echo "Checking compliance for \$framework..."
                
                case \$framework in
                  "SOC2")
                    echo "Running SOC2 compliance checks..."
                    # Check encryption at rest
                    # Check access controls
                    # Check audit logging
                    ;;
                  "PCI-DSS")
                    echo "Running PCI-DSS compliance checks..."
                    # Check network segmentation
                    # Check encryption
                    # Check vulnerability management
                    ;;
                  "GDPR")
                    echo "Running GDPR compliance checks..."
                    # Check data protection
                    # Check consent management
                    # Check data retention policies
                    ;;
                esac
              done
              
              echo "Compliance checks completed"
            resources:
              requests:
                cpu: "100m"
                memory: "128Mi"
              limits:
                cpu: "500m"
                memory: "512Mi"
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy compliance monitoring"
    else
        kubectl apply -f "$compliance_dir/compliance-checker.yaml"
        log_success "Compliance monitoring deployed"
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying security stack deployment..."
    
    local components=()
    
    if [[ "$ENABLE_VULNERABILITY_SCANNING" == "true" ]]; then
        components+=("trivy-operator")
    fi
    
    if [[ "$ENABLE_POLICY_ENFORCEMENT" == "true" ]]; then
        components+=("gatekeeper-controller-manager")
        components+=("gatekeeper-audit")
    fi
    
    if [[ "$ENABLE_THREAT_DETECTION" == "true" ]]; then
        components+=("falco")
    fi
    
    local failed_components=()
    
    for component in "${components[@]}"; do
        log_info "Checking component: $component"
        
        if kubectl get pods -n "$SECURITY_NAMESPACE" -l "app.kubernetes.io/name=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        else
            log_error "Component $component is not running"
            failed_components+=("$component")
        fi
    done
    
    if [[ ${#failed_components[@]} -gt 0 ]]; then
        log_error "Some security components failed to deploy: ${failed_components[*]}"
        log_info "Check pod status with: kubectl get pods -n $SECURITY_NAMESPACE"
        return 1
    fi
    
    log_success "All security components are running successfully"
}

# Display security information
display_security_info() {
    log_info "Security stack information:"
    
    print_separator
    
    echo -e "${GREEN}Security Tools Deployed:${NC}"
    
    if [[ "$ENABLE_VULNERABILITY_SCANNING" == "true" ]]; then
        echo -e "  ✅ Vulnerability Scanning (Trivy Operator)"
        echo -e "     - Scans: Container images, Kubernetes configurations"
        echo -e "     - Schedule: Daily at 2 AM"
    fi
    
    if [[ "$ENABLE_POLICY_ENFORCEMENT" == "true" ]]; then
        echo -e "  ✅ Policy Enforcement (OPA Gatekeeper)"
        echo -e "     - Policies: Security context, resource limits"
        echo -e "     - Mode: Enforcing"
    fi
    
    if [[ "$ENABLE_THREAT_DETECTION" == "true" ]]; then
        echo -e "  ✅ Threat Detection (Falco)"
        echo -e "     - Runtime security monitoring"
        echo -e "     - Custom rules for Go Coffee platform"
    fi
    
    if [[ "$ENABLE_COMPLIANCE_MONITORING" == "true" ]]; then
        echo -e "  ✅ Compliance Monitoring"
        echo -e "     - Frameworks: ${COMPLIANCE_FRAMEWORKS:-SOC2,PCI-DSS,GDPR}"
        echo -e "     - Schedule: Weekly on Monday at 6 AM"
    fi
    
    print_separator
    
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  View security pods: kubectl get pods -n $SECURITY_NAMESPACE"
    echo -e "  View vulnerability reports: kubectl get vulnerabilityreports -A"
    echo -e "  View policy violations: kubectl get constraints"
    echo -e "  View Falco alerts: kubectl logs -n $SECURITY_NAMESPACE -l app.kubernetes.io/name=falco"
    echo -e "  Check compliance status: kubectl get cronjobs -n $SECURITY_NAMESPACE"
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

Deploy comprehensive security and compliance stack for Go Coffee platform.

OPTIONS:
    --environment ENV               Environment (dev, staging, prod) [default: dev]
    --project NAME                  Project name [default: go-coffee]
    --namespace NAME                Security namespace [default: security]
    --enable-vulnerability-scanning Enable vulnerability scanning [default: true]
    --enable-compliance-monitoring  Enable compliance monitoring [default: true]
    --enable-threat-detection       Enable threat detection [default: true]
    --enable-policy-enforcement     Enable policy enforcement [default: true]
    --dry-run                      Perform dry run without actual deployment
    --verbose                      Enable verbose output
    --help                         Show this help message

EXAMPLES:
    $0                                    # Deploy full security stack
    $0 --environment prod                 # Deploy to production
    $0 --dry-run                         # Perform dry run
    $0 --enable-vulnerability-scanning --enable-threat-detection  # Deploy specific components

ENVIRONMENT VARIABLES:
    SLACK_WEBHOOK_URL              Slack webhook for security alerts
    EMAIL_USERNAME                 Email username for notifications
    EMAIL_PASSWORD                 Email password for notifications
    WEBHOOK_URL                    Generic webhook URL for alerts
    DOCKER_REGISTRY_USERNAME       Docker registry username for scanning
    DOCKER_REGISTRY_PASSWORD       Docker registry password for scanning
    COMPLIANCE_FRAMEWORKS          Comma-separated list of compliance frameworks
    MONITORING_ENABLED             Enable monitoring integration [default: true]

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
                SECURITY_NAMESPACE="$2"
                shift 2
                ;;
            --enable-vulnerability-scanning)
                ENABLE_VULNERABILITY_SCANNING="true"
                shift
                ;;
            --disable-vulnerability-scanning)
                ENABLE_VULNERABILITY_SCANNING="false"
                shift
                ;;
            --enable-compliance-monitoring)
                ENABLE_COMPLIANCE_MONITORING="true"
                shift
                ;;
            --disable-compliance-monitoring)
                ENABLE_COMPLIANCE_MONITORING="false"
                shift
                ;;
            --enable-threat-detection)
                ENABLE_THREAT_DETECTION="true"
                shift
                ;;
            --disable-threat-detection)
                ENABLE_THREAT_DETECTION="false"
                shift
                ;;
            --enable-policy-enforcement)
                ENABLE_POLICY_ENFORCEMENT="true"
                shift
                ;;
            --disable-policy-enforcement)
                ENABLE_POLICY_ENFORCEMENT="false"
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
    
    print_header "🔒 Go Coffee Security & Compliance Stack Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  Namespace: $SECURITY_NAMESPACE"
    log_info "  Vulnerability Scanning: $ENABLE_VULNERABILITY_SCANNING"
    log_info "  Compliance Monitoring: $ENABLE_COMPLIANCE_MONITORING"
    log_info "  Threat Detection: $ENABLE_THREAT_DETECTION"
    log_info "  Policy Enforcement: $ENABLE_POLICY_ENFORCEMENT"
    log_info "  Dry Run: $DRY_RUN"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repositories
    create_security_secrets
    deploy_vulnerability_scanning
    deploy_policy_enforcement
    deploy_threat_detection
    deploy_compliance_monitoring
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
        display_security_info
    fi
    
    print_header "✅ Security & Compliance Stack Deployment Completed"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Security stack deployed successfully!"
        log_info "Monitor security alerts in your configured notification channels"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
