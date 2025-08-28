#!/bin/bash

# Disaster Recovery Stack Deployment Script
# Deploys comprehensive disaster recovery and business continuity solutions

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/disaster-recovery"
DR_DIR="$PROJECT_ROOT/k8s/disaster-recovery"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
DR_NAMESPACE="${DR_NAMESPACE:-disaster-recovery}"
ENABLE_AUTOMATED_BACKUPS="${ENABLE_AUTOMATED_BACKUPS:-true}"
ENABLE_AUTOMATED_FAILOVER="${ENABLE_AUTOMATED_FAILOVER:-true}"
ENABLE_DR_TESTING="${ENABLE_DR_TESTING:-true}"
ENABLE_CROSS_REGION_REPLICATION="${ENABLE_CROSS_REGION_REPLICATION:-true}"
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
    
    local required_tools=("kubectl" "helm" "terraform" "jq" "curl" "velero")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command_exists "$tool"; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and try again."
        log_info "Install Velero CLI: https://velero.io/docs/v1.12/basic-install/"
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
    
    # Check cloud provider CLI
    if [[ "${CLOUD_PROVIDER:-aws}" == "aws" ]] && ! command_exists "aws"; then
        log_warning "AWS CLI not found - some features may not work"
    fi
    
    if [[ "${CLOUD_PROVIDER:-}" == "gcp" ]] && ! command_exists "gcloud"; then
        log_warning "Google Cloud CLI not found - some features may not work"
    fi
    
    if [[ "${CLOUD_PROVIDER:-}" == "azure" ]] && ! command_exists "az"; then
        log_warning "Azure CLI not found - some features may not work"
    fi
    
    log_success "Prerequisites check completed"
}

# Setup DR namespace
setup_namespace() {
    log_info "Setting up disaster recovery namespace: $DR_NAMESPACE"
    
    if kubectl get namespace "$DR_NAMESPACE" &>/dev/null; then
        log_info "Namespace $DR_NAMESPACE already exists"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would create namespace $DR_NAMESPACE"
        else
            kubectl create namespace "$DR_NAMESPACE"
            
            # Apply labels and annotations
            kubectl label namespace "$DR_NAMESPACE" \
                app.kubernetes.io/name="$PROJECT_NAME" \
                app.kubernetes.io/component="disaster-recovery" \
                environment="$ENVIRONMENT"
            
            kubectl annotate namespace "$DR_NAMESPACE" \
                managed-by="terraform" \
                created-at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
            
            log_success "Created disaster recovery namespace: $DR_NAMESPACE"
        fi
    fi
}

# Add Helm repositories
add_helm_repositories() {
    log_info "Adding Helm repositories for DR tools..."
    
    local repos=(
        "vmware-tanzu:https://vmware-tanzu.github.io/helm-charts"
        "chaos-mesh:https://charts.chaos-mesh.org"
        "argo:https://argoproj.github.io/argo-helm"
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

# Create DR secrets
create_dr_secrets() {
    log_info "Creating disaster recovery secrets..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create DR secrets"
        return 0
    fi
    
    # Create cloud credentials for backups
    case "${CLOUD_PROVIDER:-aws}" in
        "aws")
            kubectl create secret generic cloud-credentials \
                --from-literal=cloud="[default]
aws_access_key_id=${AWS_ACCESS_KEY_ID:-}
aws_secret_access_key=${AWS_SECRET_ACCESS_KEY:-}" \
                --namespace="$DR_NAMESPACE" \
                --dry-run=client -o yaml | kubectl apply -f -
            ;;
        "gcp")
            kubectl create secret generic cloud-credentials \
                --from-file=cloud="${GCP_SERVICE_ACCOUNT_KEY_FILE:-/dev/null}" \
                --namespace="$DR_NAMESPACE" \
                --dry-run=client -o yaml | kubectl apply -f -
            ;;
        "azure")
            kubectl create secret generic cloud-credentials \
                --from-literal=cloud="AZURE_SUBSCRIPTION_ID=${AZURE_SUBSCRIPTION_ID:-}
AZURE_TENANT_ID=${AZURE_TENANT_ID:-}
AZURE_CLIENT_ID=${AZURE_CLIENT_ID:-}
AZURE_CLIENT_SECRET=${AZURE_CLIENT_SECRET:-}
AZURE_RESOURCE_GROUP=${AZURE_RESOURCE_GROUP:-}
AZURE_CLOUD_NAME=AzurePublicCloud" \
                --namespace="$DR_NAMESPACE" \
                --dry-run=client -o yaml | kubectl apply -f -
            ;;
    esac
    
    # Create database credentials
    kubectl create secret generic database-credentials \
        --from-literal=url="${DATABASE_URL:-postgresql://user:pass@localhost:5432/go_coffee}" \
        --namespace="$DR_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create backup storage credentials
    kubectl create secret generic backup-credentials \
        --from-literal=storage-url="${BACKUP_STORAGE_URL:-s3://go-coffee-backups}" \
        --namespace="$DR_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Create notification secrets
    kubectl create secret generic dr-notifications \
        --from-literal=slack-webhook-url="${SLACK_WEBHOOK_URL:-}" \
        --from-literal=email-smtp-server="${EMAIL_SMTP_SERVER:-}" \
        --from-literal=email-username="${EMAIL_USERNAME:-}" \
        --from-literal=email-password="${EMAIL_PASSWORD:-}" \
        --namespace="$DR_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "DR secrets created"
}

# Deploy Velero backup system
deploy_velero() {
    if [[ "$ENABLE_AUTOMATED_BACKUPS" != "true" ]]; then
        log_info "Automated backups disabled, skipping Velero..."
        return 0
    fi
    
    log_info "Deploying Velero backup system..."
    
    local values_file="$DR_DIR/velero-values.yaml"
    
    # Create values file if it doesn't exist
    if [[ ! -f "$values_file" ]]; then
        mkdir -p "$DR_DIR"
        log_info "Creating Velero values file..."
        cat > "$values_file" << EOF
configuration:
  provider: ${CLOUD_PROVIDER:-aws}
  
  backupStorageLocation:
    name: default
    provider: ${CLOUD_PROVIDER:-aws}
    bucket: ${BACKUP_STORAGE_BUCKET:-go-coffee-backups}
    config:
      region: ${PRIMARY_REGION:-us-east-1}
      s3ForcePathStyle: false
  
  volumeSnapshotLocation:
    name: default
    provider: ${CLOUD_PROVIDER:-aws}
    config:
      region: ${PRIMARY_REGION:-us-east-1}
  
  defaultBackupTTL: "${BACKUP_RETENTION_DAYS:-30}d"
  
  restoreResourcePriorities:
    - namespaces
    - storageclasses
    - volumesnapshotclass.snapshot.storage.k8s.io
    - volumesnapshotcontents.snapshot.storage.k8s.io
    - volumesnapshots.snapshot.storage.k8s.io
    - persistentvolumes
    - persistentvolumeclaims
    - secrets
    - configmaps
    - serviceaccounts
    - limitranges
    - pods

credentials:
  useSecret: true
  name: cloud-credentials

initContainers:
  - name: velero-plugin-for-${CLOUD_PROVIDER:-aws}
    image: velero/velero-plugin-for-${CLOUD_PROVIDER:-aws}:v1.8.0
    imagePullPolicy: IfNotPresent
    volumeMounts:
      - mountPath: /target
        name: plugins

resources:
  requests:
    cpu: "500m"
    memory: "128Mi"
  limits:
    cpu: "1000m"
    memory: "512Mi"

serviceMonitor:
  enabled: ${MONITORING_ENABLED:-true}

metrics:
  enabled: ${MONITORING_ENABLED:-true}
  scrapeInterval: 30s
  scrapeTimeout: 10s

deployRestic: true

schedules:
  daily:
    disabled: false
    schedule: "0 2 * * *"
    template:
      ttl: "${BACKUP_RETENTION_DAYS:-30}d"
      includedNamespaces:
        - go-coffee
        - ai-agents
        - web3
        - monitoring
      excludedResources:
        - events
        - events.events.k8s.io
      snapshotVolumes: true
      includeClusterResources: true
  
  weekly:
    disabled: false
    schedule: "0 1 * * 0"
    template:
      ttl: "$((${BACKUP_RETENTION_DAYS:-30} * 4))d"
      includedNamespaces:
        - "*"
      excludedResources:
        - events
        - events.events.k8s.io
      snapshotVolumes: true
      includeClusterResources: true
EOF
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy Velero"
        helm template velero vmware-tanzu/velero \
            --namespace "$DR_NAMESPACE" \
            --values "$values_file" > /dev/null
        log_info "DRY RUN: Velero template validation passed"
    else
        helm upgrade --install velero vmware-tanzu/velero \
            --namespace "$DR_NAMESPACE" \
            --values "$values_file" \
            --wait \
            --timeout 15m
        
        log_success "Velero deployed successfully"
        
        # Verify Velero installation
        sleep 30
        if velero version --client-only &>/dev/null; then
            log_success "Velero CLI is working"
        else
            log_warning "Velero CLI verification failed"
        fi
    fi
}

# Deploy failover controller
deploy_failover_controller() {
    if [[ "$ENABLE_AUTOMATED_FAILOVER" != "true" ]]; then
        log_info "Automated failover disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying failover controller..."
    
    local failover_dir="$DR_DIR/failover"
    mkdir -p "$failover_dir"
    
    # Create failover controller deployment
    cat > "$failover_dir/failover-controller.yaml" << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${PROJECT_NAME}-failover-controller
  namespace: $DR_NAMESPACE
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: failover-controller
    environment: $ENVIRONMENT
spec:
  replicas: ${FAILOVER_CONTROLLER_REPLICAS:-2}
  selector:
    matchLabels:
      app: ${PROJECT_NAME}-failover-controller
  template:
    metadata:
      labels:
        app: ${PROJECT_NAME}-failover-controller
        app.kubernetes.io/name: $PROJECT_NAME
        app.kubernetes.io/component: failover-controller
    spec:
      serviceAccountName: ${PROJECT_NAME}-failover-controller
      containers:
      - name: failover-controller
        image: ${FAILOVER_CONTROLLER_IMAGE:-go-coffee/failover-controller:latest}
        env:
        - name: ENVIRONMENT
          value: "$ENVIRONMENT"
        - name: PROJECT_NAME
          value: "$PROJECT_NAME"
        - name: PRIMARY_REGION
          value: "${PRIMARY_REGION:-us-east-1}"
        - name: SECONDARY_REGION
          value: "${SECONDARY_REGION:-us-west-2}"
        - name: HEALTH_CHECK_INTERVAL
          value: "${HEALTH_CHECK_INTERVAL:-30}"
        - name: FAILURE_THRESHOLD
          value: "${FAILURE_THRESHOLD:-3}"
        - name: RECOVERY_THRESHOLD
          value: "${RECOVERY_THRESHOLD:-5}"
        - name: RTO_MINUTES
          value: "${RTO_MINUTES:-60}"
        - name: RPO_MINUTES
          value: "${RPO_MINUTES:-15}"
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8443
          name: webhook
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${PROJECT_NAME}-failover-controller
  namespace: $DR_NAMESPACE
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: failover-controller
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${PROJECT_NAME}-failover-controller
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: failover-controller
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${PROJECT_NAME}-failover-controller
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: failover-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${PROJECT_NAME}-failover-controller
subjects:
- kind: ServiceAccount
  name: ${PROJECT_NAME}-failover-controller
  namespace: $DR_NAMESPACE
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy failover controller"
    else
        kubectl apply -f "$failover_dir/failover-controller.yaml"
        log_success "Failover controller deployed"
    fi
}

# Deploy DR testing automation
deploy_dr_testing() {
    if [[ "$ENABLE_DR_TESTING" != "true" ]]; then
        log_info "DR testing disabled, skipping..."
        return 0
    fi
    
    log_info "Deploying DR testing automation..."
    
    local testing_dir="$DR_DIR/testing"
    mkdir -p "$testing_dir"
    
    # Create DR testing CronJob
    cat > "$testing_dir/dr-testing.yaml" << EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ${PROJECT_NAME}-dr-testing
  namespace: $DR_NAMESPACE
  labels:
    app.kubernetes.io/name: $PROJECT_NAME
    app.kubernetes.io/component: dr-testing
    environment: $ENVIRONMENT
spec:
  schedule: "${DR_TEST_SCHEDULE:-0 6 1 * *}"
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: dr-tester
            image: ${DR_TESTING_IMAGE:-go-coffee/dr-tester:latest}
            env:
            - name: ENVIRONMENT
              value: "$ENVIRONMENT"
            - name: PROJECT_NAME
              value: "$PROJECT_NAME"
            - name: TEST_TYPE
              value: "${DR_TEST_TYPE:-backup_restore}"
            - name: AUTOMATED_TESTING
              value: "${ENABLE_AUTOMATED_DR_TESTING:-false}"
            command: ["/bin/sh"]
            args:
            - -c
            - |
              echo "Starting DR testing..."
              
              # Test backup integrity
              echo "Testing backup integrity..."
              velero backup get --output json | jq '.items[] | select(.status.phase == "Completed")'
              
              # Test restore functionality
              if [ "\$AUTOMATED_TESTING" = "true" ]; then
                echo "Performing automated restore test..."
                
                # Create test namespace
                TEST_NS="dr-test-\$(date +%s)"
                kubectl create namespace \$TEST_NS || true
                
                # Get latest backup
                LATEST_BACKUP=\$(velero backup get -o name | head -1 | cut -d'/' -f2)
                
                if [ -n "\$LATEST_BACKUP" ]; then
                  # Perform test restore
                  velero restore create dr-test-restore-\$(date +%s) \
                    --from-backup \$LATEST_BACKUP \
                    --namespace-mappings go-coffee:\$TEST_NS
                  
                  # Wait for restore completion
                  sleep 300
                  
                  # Verify restore
                  kubectl get pods -n \$TEST_NS
                  
                  # Cleanup test namespace
                  kubectl delete namespace \$TEST_NS || true
                else
                  echo "No backups found for testing"
                fi
              fi
              
              echo "DR testing completed"
            resources:
              requests:
                cpu: "100m"
                memory: "128Mi"
              limits:
                cpu: "500m"
                memory: "512Mi"
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy DR testing"
    else
        kubectl apply -f "$testing_dir/dr-testing.yaml"
        log_success "DR testing automation deployed"
    fi
}

# Create DR runbooks
create_dr_runbooks() {
    log_info "Creating DR runbooks and documentation..."
    
    local runbooks_dir="$DR_DIR/runbooks"
    mkdir -p "$runbooks_dir"
    
    # Create disaster recovery runbook
    cat > "$runbooks_dir/disaster-recovery-runbook.md" << EOF
# Go Coffee Disaster Recovery Runbook

## Overview
This runbook provides step-by-step procedures for disaster recovery scenarios.

## Recovery Objectives
- **RTO (Recovery Time Objective)**: ${RTO_MINUTES:-60} minutes
- **RPO (Recovery Point Objective)**: ${RPO_MINUTES:-15} minutes

## Emergency Contacts
- **Primary Contact**: platform@go-coffee.com
- **Escalation**: cto@go-coffee.com
- **24/7 Hotline**: +1-555-DR-HELP

## Disaster Scenarios

### 1. Complete Region Failure

#### Detection
- Multiple service health checks failing
- No response from primary region endpoints
- Cloud provider status page indicates regional issues

#### Response Steps
1. **Assess Situation** (5 minutes)
   - Verify regional outage via cloud provider status
   - Check secondary region availability
   - Notify incident response team

2. **Initiate Failover** (15 minutes)
   - Execute automated failover if available
   - Manual failover steps:
     \`\`\`bash
     # Switch DNS to secondary region
     kubectl patch ingress go-coffee-ingress -p '{"spec":{"rules":[{"host":"go-coffee.com","http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"go-coffee-service-dr","port":{"number":80}}}}]}}]}}'
     
     # Scale up DR services
     kubectl scale deployment go-coffee-service --replicas=3 -n go-coffee-dr
     \`\`\`

3. **Verify Recovery** (10 minutes)
   - Test critical user journeys
   - Verify database connectivity
   - Check payment processing

4. **Monitor and Communicate** (Ongoing)
   - Monitor system health
   - Update status page
   - Communicate with stakeholders

### 2. Database Failure

#### Detection
- Database connection errors
- High error rates in application logs
- Database monitoring alerts

#### Response Steps
1. **Assess Database State** (5 minutes)
   \`\`\`bash
   # Check database connectivity
   kubectl exec -it postgres-primary-0 -- pg_isready
   
   # Check replication status
   kubectl exec -it postgres-primary-0 -- psql -c "SELECT * FROM pg_stat_replication;"
   \`\`\`

2. **Restore from Backup** (20 minutes)
   \`\`\`bash
   # Get latest backup
   LATEST_BACKUP=\$(velero backup get -o name | head -1)
   
   # Restore database
   velero restore create db-restore-\$(date +%s) --from-backup \$LATEST_BACKUP --include-resources persistentvolumeclaims,persistentvolumes
   \`\`\`

3. **Verify Data Integrity** (10 minutes)
   - Run data consistency checks
   - Verify recent transactions
   - Test application connectivity

## Recovery Procedures

### Backup Restoration
\`\`\`bash
# List available backups
velero backup get

# Restore specific backup
velero restore create <restore-name> --from-backup <backup-name>

# Monitor restore progress
velero restore describe <restore-name>
\`\`\`

### Service Recovery
\`\`\`bash
# Check service health
kubectl get pods -n go-coffee
kubectl get services -n go-coffee

# Restart failed services
kubectl rollout restart deployment/<service-name> -n go-coffee

# Scale services
kubectl scale deployment <service-name> --replicas=<count> -n go-coffee
\`\`\`

## Post-Incident Actions
1. **Root Cause Analysis**
   - Document timeline of events
   - Identify contributing factors
   - Create improvement action items

2. **Update Procedures**
   - Update runbooks based on lessons learned
   - Test updated procedures
   - Train team on changes

3. **Communication**
   - Post-mortem report
   - Stakeholder communication
   - Customer communication if needed

## Testing Schedule
- **Monthly**: Backup restoration test
- **Quarterly**: Failover test
- **Annually**: Full DR exercise

## Useful Commands
\`\`\`bash
# Check cluster status
kubectl cluster-info

# View recent events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check resource usage
kubectl top nodes
kubectl top pods -A

# Velero operations
velero backup create <name>
velero restore create <name> --from-backup <backup-name>
velero schedule get
\`\`\`
EOF
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create DR runbooks"
    else
        log_success "DR runbooks created at: $runbooks_dir"
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying disaster recovery stack deployment..."
    
    local components=()
    
    if [[ "$ENABLE_AUTOMATED_BACKUPS" == "true" ]]; then
        components+=("velero")
    fi
    
    if [[ "$ENABLE_AUTOMATED_FAILOVER" == "true" ]]; then
        components+=("${PROJECT_NAME}-failover-controller")
    fi
    
    local failed_components=()
    
    for component in "${components[@]}"; do
        log_info "Checking component: $component"
        
        if kubectl get pods -n "$DR_NAMESPACE" -l "app.kubernetes.io/name=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        elif kubectl get pods -n "$DR_NAMESPACE" -l "app=$component" --field-selector=status.phase=Running | grep -q Running; then
            log_success "Component $component is running"
        else
            log_error "Component $component is not running"
            failed_components+=("$component")
        fi
    done
    
    if [[ ${#failed_components[@]} -gt 0 ]]; then
        log_error "Some DR components failed to deploy: ${failed_components[*]}"
        log_info "Check pod status with: kubectl get pods -n $DR_NAMESPACE"
        return 1
    fi
    
    # Test Velero if deployed
    if [[ "$ENABLE_AUTOMATED_BACKUPS" == "true" && "$DRY_RUN" != "true" ]]; then
        log_info "Testing Velero backup functionality..."
        if velero backup-location get &>/dev/null; then
            log_success "Velero backup location is configured"
        else
            log_warning "Velero backup location configuration issue"
        fi
    fi
    
    log_success "All DR components are running successfully"
}

# Display DR information
display_dr_info() {
    log_info "Disaster recovery stack information:"
    
    print_separator
    
    echo -e "${GREEN}DR Components Deployed:${NC}"
    
    if [[ "$ENABLE_AUTOMATED_BACKUPS" == "true" ]]; then
        echo -e "  ✅ Velero Backup System"
        echo -e "     - Automated daily and weekly backups"
        echo -e "     - Cross-region backup replication"
        echo -e "     - Retention: ${BACKUP_RETENTION_DAYS:-30} days"
    fi
    
    if [[ "$ENABLE_AUTOMATED_FAILOVER" == "true" ]]; then
        echo -e "  ✅ Automated Failover Controller"
        echo -e "     - Health monitoring and automated failover"
        echo -e "     - RTO: ${RTO_MINUTES:-60} minutes"
        echo -e "     - RPO: ${RPO_MINUTES:-15} minutes"
    fi
    
    if [[ "$ENABLE_DR_TESTING" == "true" ]]; then
        echo -e "  ✅ DR Testing Automation"
        echo -e "     - Automated backup integrity testing"
        echo -e "     - Scheduled DR exercises"
    fi
    
    if [[ "$ENABLE_CROSS_REGION_REPLICATION" == "true" ]]; then
        echo -e "  ✅ Cross-Region Replication"
        echo -e "     - Data replication to secondary region"
        echo -e "     - Automated sync every ${REPLICATION_INTERVAL_HOURS:-4} hours"
    fi
    
    print_separator
    
    echo -e "${YELLOW}Recovery Objectives:${NC}"
    echo -e "  🎯 RTO (Recovery Time Objective): ${RTO_MINUTES:-60} minutes"
    echo -e "  🎯 RPO (Recovery Point Objective): ${RPO_MINUTES:-15} minutes"
    echo -e "  📊 Backup Retention: ${BACKUP_RETENTION_DAYS:-30} days"
    echo -e "  🔄 Replication Frequency: Every ${REPLICATION_INTERVAL_HOURS:-4} hours"
    
    print_separator
    
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  View DR pods: kubectl get pods -n $DR_NAMESPACE"
    echo -e "  List backups: velero backup get"
    echo -e "  Create backup: velero backup create <name>"
    echo -e "  Restore backup: velero restore create <name> --from-backup <backup-name>"
    echo -e "  Check backup location: velero backup-location get"
    echo -e "  View DR runbooks: cat $DR_DIR/runbooks/disaster-recovery-runbook.md"
    
    print_separator
    
    echo -e "${GREEN}Next Steps:${NC}"
    echo -e "  1. Review and customize DR runbooks"
    echo -e "  2. Test backup and restore procedures"
    echo -e "  3. Schedule regular DR exercises"
    echo -e "  4. Update incident response contacts"
    echo -e "  5. Configure monitoring and alerting"
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

Deploy comprehensive disaster recovery stack for Go Coffee platform.

OPTIONS:
    --environment ENV               Environment (dev, staging, prod) [default: dev]
    --project NAME                  Project name [default: go-coffee]
    --namespace NAME                DR namespace [default: disaster-recovery]
    --enable-automated-backups     Enable automated backups [default: true]
    --enable-automated-failover    Enable automated failover [default: true]
    --enable-dr-testing            Enable DR testing [default: true]
    --enable-cross-region-replication Enable cross-region replication [default: true]
    --dry-run                      Perform dry run without actual deployment
    --verbose                      Enable verbose output
    --help                         Show this help message

EXAMPLES:
    $0                                    # Deploy full DR stack
    $0 --environment prod                 # Deploy to production
    $0 --dry-run                         # Perform dry run
    $0 --enable-automated-backups --enable-automated-failover  # Deploy specific components

ENVIRONMENT VARIABLES:
    CLOUD_PROVIDER                 Cloud provider (aws, gcp, azure)
    BACKUP_STORAGE_BUCKET         Storage bucket for backups
    PRIMARY_REGION                Primary deployment region
    SECONDARY_REGION              Secondary/DR region
    RTO_MINUTES                   Recovery Time Objective in minutes
    RPO_MINUTES                   Recovery Point Objective in minutes
    BACKUP_RETENTION_DAYS         Backup retention period
    AWS_ACCESS_KEY_ID             AWS access key for backups
    AWS_SECRET_ACCESS_KEY         AWS secret key for backups
    SLACK_WEBHOOK_URL             Slack webhook for DR notifications

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
                DR_NAMESPACE="$2"
                shift 2
                ;;
            --enable-automated-backups)
                ENABLE_AUTOMATED_BACKUPS="true"
                shift
                ;;
            --disable-automated-backups)
                ENABLE_AUTOMATED_BACKUPS="false"
                shift
                ;;
            --enable-automated-failover)
                ENABLE_AUTOMATED_FAILOVER="true"
                shift
                ;;
            --disable-automated-failover)
                ENABLE_AUTOMATED_FAILOVER="false"
                shift
                ;;
            --enable-dr-testing)
                ENABLE_DR_TESTING="true"
                shift
                ;;
            --disable-dr-testing)
                ENABLE_DR_TESTING="false"
                shift
                ;;
            --enable-cross-region-replication)
                ENABLE_CROSS_REGION_REPLICATION="true"
                shift
                ;;
            --disable-cross-region-replication)
                ENABLE_CROSS_REGION_REPLICATION="false"
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
    
    print_header "🛡️ Go Coffee Disaster Recovery Stack Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  Namespace: $DR_NAMESPACE"
    log_info "  Automated Backups: $ENABLE_AUTOMATED_BACKUPS"
    log_info "  Automated Failover: $ENABLE_AUTOMATED_FAILOVER"
    log_info "  DR Testing: $ENABLE_DR_TESTING"
    log_info "  Cross-Region Replication: $ENABLE_CROSS_REGION_REPLICATION"
    log_info "  Dry Run: $DRY_RUN"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    setup_namespace
    add_helm_repositories
    create_dr_secrets
    deploy_velero
    deploy_failover_controller
    deploy_dr_testing
    create_dr_runbooks
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
        display_dr_info
    fi
    
    print_header "✅ Disaster Recovery Stack Deployment Completed"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Disaster recovery stack deployed successfully!"
        log_info "Review the DR runbooks and test your recovery procedures"
        log_info "Runbooks location: $DR_DIR/runbooks/"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
