#!/bin/bash

# Infrastructure Automation Script
# Manages multi-cloud infrastructure with advanced automation features

set -euo pipefail

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
ACTION="${1:-plan}"
ENABLE_AWS="${ENABLE_AWS:-true}"
ENABLE_GCP="${ENABLE_GCP:-true}"
ENABLE_AZURE="${ENABLE_AZURE:-false}"
AUTO_APPROVE="${AUTO_APPROVE:-false}"
DRIFT_CHECK="${DRIFT_CHECK:-true}"
BACKUP_BEFORE_APPLY="${BACKUP_BEFORE_APPLY:-true}"
PARALLEL_EXECUTION="${PARALLEL_EXECUTION:-false}"

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

# Progress indicator
show_progress() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Validate prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local required_tools=("terraform" "jq" "curl" "git")
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
    
    # Check cloud CLI tools
    if [[ "$ENABLE_AWS" == "true" ]] && ! command_exists "aws"; then
        log_error "AWS CLI is required when AWS is enabled"
        exit 1
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]] && ! command_exists "gcloud"; then
        log_error "Google Cloud CLI is required when GCP is enabled"
        exit 1
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]] && ! command_exists "az"; then
        log_error "Azure CLI is required when Azure is enabled"
        exit 1
    fi
    
    # Check Terraform version
    local tf_version
    tf_version=$(terraform version -json | jq -r '.terraform_version')
    local required_version="1.6.0"
    
    if ! printf '%s\n%s\n' "$required_version" "$tf_version" | sort -V -C; then
        log_error "Terraform version $tf_version is too old. Required: $required_version or newer"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Initialize Terraform workspace
init_terraform() {
    local workspace_dir="$1"
    local workspace_name="$2"
    
    log_info "Initializing Terraform workspace: $workspace_name"
    
    cd "$workspace_dir"
    
    # Initialize Terraform
    terraform init -upgrade
    
    # Create or select workspace
    if terraform workspace list | grep -q "$workspace_name"; then
        terraform workspace select "$workspace_name"
        log_info "Selected existing workspace: $workspace_name"
    else
        terraform workspace new "$workspace_name"
        log_info "Created new workspace: $workspace_name"
    fi
    
    cd "$PROJECT_ROOT"
}

# Generate Terraform variables
generate_terraform_vars() {
    local workspace_dir="$1"
    local vars_file="$workspace_dir/terraform.tfvars"
    
    log_info "Generating Terraform variables file: $vars_file"
    
    cat > "$vars_file" << EOF
# Generated Terraform variables for $ENVIRONMENT environment
# Generated at: $(date)

# Project Configuration
project_name = "$PROJECT_NAME"
environment = "$ENVIRONMENT"
owner = "${OWNER:-terraform}"
cost_center = "${COST_CENTER:-platform}"

# Cloud Provider Configuration
enable_aws = $ENABLE_AWS
enable_gcp = $ENABLE_GCP
enable_azure = $ENABLE_AZURE

# Multi-Cloud Configuration
enable_cross_cloud_networking = ${ENABLE_CROSS_CLOUD_NETWORKING:-true}
enable_global_load_balancing = ${ENABLE_GLOBAL_LOAD_BALANCING:-true}
enable_disaster_recovery = ${ENABLE_DISASTER_RECOVERY:-true}

# Security Configuration
encryption_at_rest = true
encryption_in_transit = true
security_scanning_enabled = true
compliance_frameworks = ["SOC2", "PCI-DSS", "GDPR"]

# Monitoring Configuration
monitoring_enabled = true
log_retention_days = 30
prometheus_retention_days = 15

# Backup Configuration
backup_retention_days = 30
backup_rotation_hours = 24

EOF

    # Add AWS-specific configuration
    if [[ "$ENABLE_AWS" == "true" ]]; then
        cat >> "$vars_file" << EOF

# AWS Configuration
aws_primary_region = "${AWS_PRIMARY_REGION:-us-east-1}"
aws_secondary_region = "${AWS_SECONDARY_REGION:-us-west-2}"
aws_tertiary_region = "${AWS_TERTIARY_REGION:-eu-west-1}"
aws_vpc_cidr = "${AWS_VPC_CIDR:-10.0.0.0/16}"
aws_availability_zones_count = ${AWS_AZ_COUNT:-3}

# AWS EKS Configuration
kubernetes_version = "${KUBERNETES_VERSION:-1.28}"
eks_public_access = ${EKS_PUBLIC_ACCESS:-false}
eks_public_access_cidrs = ["${EKS_PUBLIC_ACCESS_CIDRS:-0.0.0.0/0}"]

# AWS RDS Configuration
aws_rds_instance_class = "${AWS_RDS_INSTANCE_CLASS:-db.t3.medium}"
aws_rds_allocated_storage = ${AWS_RDS_ALLOCATED_STORAGE:-100}
aws_rds_engine_version = "${AWS_RDS_ENGINE_VERSION:-15.4}"

EOF
    fi
    
    # Add GCP-specific configuration
    if [[ "$ENABLE_GCP" == "true" ]]; then
        cat >> "$vars_file" << EOF

# GCP Configuration
gcp_project_id = "$GCP_PROJECT_ID"
gcp_primary_region = "${GCP_PRIMARY_REGION:-us-central1}"
gcp_secondary_region = "${GCP_SECONDARY_REGION:-us-east1}"
gcp_tertiary_region = "${GCP_TERTIARY_REGION:-europe-west1}"
gcp_vpc_cidr_range = "${GCP_VPC_CIDR_RANGE:-10.1.0.0/16}"

# GCP GKE Configuration
gcp_gke_node_count = ${GCP_GKE_NODE_COUNT:-3}
gcp_gke_machine_type = "${GCP_GKE_MACHINE_TYPE:-e2-standard-2}"
gcp_gke_disk_size_gb = ${GCP_GKE_DISK_SIZE_GB:-100}

# GCP Cloud SQL Configuration
gcp_sql_tier = "${GCP_SQL_TIER:-db-custom-2-4096}"
gcp_sql_disk_size = ${GCP_SQL_DISK_SIZE:-100}
gcp_sql_database_version = "${GCP_SQL_DATABASE_VERSION:-POSTGRES_15}"

EOF
    fi
    
    # Add Azure-specific configuration
    if [[ "$ENABLE_AZURE" == "true" ]]; then
        cat >> "$vars_file" << EOF

# Azure Configuration
azure_primary_region = "${AZURE_PRIMARY_REGION:-East US}"
azure_secondary_region = "${AZURE_SECONDARY_REGION:-West US 2}"
azure_tertiary_region = "${AZURE_TERTIARY_REGION:-West Europe}"
azure_vnet_address_space = ["${AZURE_VNET_ADDRESS_SPACE:-10.2.0.0/16}"]

# Azure AKS Configuration
azure_aks_node_count = ${AZURE_AKS_NODE_COUNT:-3}
azure_aks_vm_size = "${AZURE_AKS_VM_SIZE:-Standard_D2s_v3}"
azure_aks_os_disk_size_gb = ${AZURE_AKS_OS_DISK_SIZE_GB:-100}

# Azure PostgreSQL Configuration
azure_postgresql_sku_name = "${AZURE_POSTGRESQL_SKU_NAME:-GP_Standard_D2s_v3}"
azure_postgresql_storage_mb = ${AZURE_POSTGRESQL_STORAGE_MB:-102400}
azure_postgresql_version = "${AZURE_POSTGRESQL_VERSION:-15}"

EOF
    fi
    
    log_success "Terraform variables file generated successfully"
}

# Check for infrastructure drift
check_drift() {
    local workspace_dir="$1"
    
    if [[ "$DRIFT_CHECK" != "true" ]]; then
        log_info "Drift check disabled, skipping..."
        return 0
    fi
    
    log_info "Checking for infrastructure drift..."
    
    cd "$workspace_dir"
    
    # Run terraform plan to check for drift
    local plan_output
    plan_output=$(terraform plan -detailed-exitcode -no-color 2>&1) || true
    local plan_exit_code=$?
    
    case $plan_exit_code in
        0)
            log_success "No infrastructure drift detected"
            ;;
        1)
            log_error "Terraform plan failed"
            echo "$plan_output"
            cd "$PROJECT_ROOT"
            exit 1
            ;;
        2)
            log_warning "Infrastructure drift detected!"
            echo "$plan_output"
            
            # Send notification if webhook is configured
            if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
                curl -X POST -H 'Content-type: application/json' \
                    --data "{\"text\":\"🚨 Infrastructure drift detected in $ENVIRONMENT environment for $PROJECT_NAME\"}" \
                    "$SLACK_WEBHOOK_URL" || true
            fi
            
            if [[ "$ACTION" == "plan" ]]; then
                log_info "Run with 'apply' action to fix the drift"
            fi
            ;;
    esac
    
    cd "$PROJECT_ROOT"
    return $plan_exit_code
}

# Create infrastructure backup
create_backup() {
    local workspace_dir="$1"
    
    if [[ "$BACKUP_BEFORE_APPLY" != "true" ]]; then
        log_info "Backup disabled, skipping..."
        return 0
    fi
    
    log_info "Creating infrastructure backup..."
    
    local backup_dir="$PROJECT_ROOT/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    cd "$workspace_dir"
    
    # Export Terraform state
    terraform state pull > "$backup_dir/terraform.tfstate"
    
    # Copy Terraform configuration
    cp -r . "$backup_dir/terraform_config/"
    
    # Export cloud resources (if tools are available)
    if [[ "$ENABLE_AWS" == "true" ]] && command_exists "aws"; then
        log_info "Backing up AWS resources..."
        aws ec2 describe-instances > "$backup_dir/aws_instances.json" 2>/dev/null || true
        aws rds describe-db-instances > "$backup_dir/aws_rds.json" 2>/dev/null || true
        aws eks list-clusters > "$backup_dir/aws_eks.json" 2>/dev/null || true
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]] && command_exists "gcloud"; then
        log_info "Backing up GCP resources..."
        gcloud compute instances list --format=json > "$backup_dir/gcp_instances.json" 2>/dev/null || true
        gcloud sql instances list --format=json > "$backup_dir/gcp_sql.json" 2>/dev/null || true
        gcloud container clusters list --format=json > "$backup_dir/gcp_gke.json" 2>/dev/null || true
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]] && command_exists "az"; then
        log_info "Backing up Azure resources..."
        az vm list > "$backup_dir/azure_vms.json" 2>/dev/null || true
        az postgres server list > "$backup_dir/azure_postgres.json" 2>/dev/null || true
        az aks list > "$backup_dir/azure_aks.json" 2>/dev/null || true
    fi
    
    cd "$PROJECT_ROOT"
    
    log_success "Backup created at: $backup_dir"
    echo "$backup_dir" > "$PROJECT_ROOT/.last_backup"
}

# Execute Terraform action
execute_terraform() {
    local workspace_dir="$1"
    local action="$2"
    
    log_info "Executing Terraform action: $action"
    
    cd "$workspace_dir"
    
    case "$action" in
        "plan")
            terraform plan -var-file=terraform.tfvars
            ;;
        "apply")
            if [[ "$AUTO_APPROVE" == "true" ]]; then
                terraform apply -var-file=terraform.tfvars -auto-approve
            else
                terraform apply -var-file=terraform.tfvars
            fi
            ;;
        "destroy")
            if [[ "$AUTO_APPROVE" == "true" ]]; then
                terraform destroy -var-file=terraform.tfvars -auto-approve
            else
                terraform destroy -var-file=terraform.tfvars
            fi
            ;;
        "validate")
            terraform validate
            ;;
        "fmt")
            terraform fmt -recursive
            ;;
        *)
            log_error "Unknown action: $action"
            cd "$PROJECT_ROOT"
            exit 1
            ;;
    esac
    
    cd "$PROJECT_ROOT"
}

# Process workspace
process_workspace() {
    local workspace_name="$1"
    local workspace_dir="$TERRAFORM_DIR/$workspace_name"
    
    if [[ ! -d "$workspace_dir" ]]; then
        log_warning "Workspace directory not found: $workspace_dir"
        return 0
    fi
    
    print_separator
    log_info "Processing workspace: $workspace_name"
    
    # Initialize Terraform
    init_terraform "$workspace_dir" "$ENVIRONMENT"
    
    # Generate variables
    generate_terraform_vars "$workspace_dir"
    
    # Check for drift (except for destroy action)
    if [[ "$ACTION" != "destroy" ]]; then
        check_drift "$workspace_dir"
    fi
    
    # Create backup before apply/destroy
    if [[ "$ACTION" == "apply" || "$ACTION" == "destroy" ]]; then
        create_backup "$workspace_dir"
    fi
    
    # Execute Terraform action
    execute_terraform "$workspace_dir" "$ACTION"
    
    log_success "Completed workspace: $workspace_name"
}

# Main execution function
main() {
    print_header "🏗️ Infrastructure Automation - $PROJECT_NAME ($ENVIRONMENT)"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Action: $ACTION"
    log_info "  AWS: $ENABLE_AWS"
    log_info "  GCP: $ENABLE_GCP"
    log_info "  Azure: $ENABLE_AZURE"
    log_info "  Auto Approve: $AUTO_APPROVE"
    log_info "  Drift Check: $DRIFT_CHECK"
    log_info "  Backup: $BACKUP_BEFORE_APPLY"
    log_info "  Parallel: $PARALLEL_EXECUTION"
    
    # Check prerequisites
    check_prerequisites
    
    # Define workspaces to process
    local workspaces=()
    
    if [[ "$ENABLE_AWS" == "true" ]]; then
        workspaces+=("modules/aws-infrastructure")
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]]; then
        workspaces+=("modules/gcp-infrastructure")
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]]; then
        workspaces+=("modules/azure-infrastructure")
    fi
    
    # Add orchestrator modules
    workspaces+=("modules/multi-cloud-orchestrator")
    workspaces+=("modules/serverless-orchestrator")
    workspaces+=("modules/gitops-automation")
    
    # Process workspaces
    if [[ "$PARALLEL_EXECUTION" == "true" ]]; then
        log_info "Processing workspaces in parallel..."
        
        local pids=()
        for workspace in "${workspaces[@]}"; do
            process_workspace "$workspace" &
            pids+=($!)
        done
        
        # Wait for all processes to complete
        for pid in "${pids[@]}"; do
            wait $pid
        done
    else
        log_info "Processing workspaces sequentially..."
        
        for workspace in "${workspaces[@]}"; do
            process_workspace "$workspace"
        done
    fi
    
    print_header "✅ Infrastructure Automation Completed"
    
    # Summary
    log_success "All workspaces processed successfully!"
    log_info "Next steps:"
    log_info "  1. Verify the deployed infrastructure"
    log_info "  2. Run integration tests"
    log_info "  3. Monitor the infrastructure health"
    log_info "  4. Set up alerts and notifications"
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [ACTION] [OPTIONS]

ACTIONS:
    plan        Show what Terraform will do (default)
    apply       Apply the Terraform configuration
    destroy     Destroy the Terraform-managed infrastructure
    validate    Validate the Terraform configuration
    fmt         Format Terraform files

OPTIONS:
    --environment ENV       Environment (dev, staging, prod) [default: dev]
    --project NAME          Project name [default: go-coffee]
    --enable-aws            Enable AWS deployment [default: true]
    --enable-gcp            Enable GCP deployment [default: true]
    --enable-azure          Enable Azure deployment [default: false]
    --auto-approve          Auto approve Terraform actions
    --no-drift-check        Disable drift detection
    --no-backup             Disable backup creation
    --parallel              Execute workspaces in parallel
    --debug                 Enable debug output
    --help                  Show this help message

EXAMPLES:
    $0 plan                                    # Plan infrastructure changes
    $0 apply --auto-approve                    # Apply changes automatically
    $0 destroy --environment prod              # Destroy prod infrastructure
    $0 validate --enable-aws --enable-gcp     # Validate multi-cloud config

ENVIRONMENT VARIABLES:
    ENVIRONMENT                 Environment name
    PROJECT_NAME               Project name
    ENABLE_AWS                 Enable AWS (true/false)
    ENABLE_GCP                 Enable GCP (true/false)
    ENABLE_AZURE               Enable Azure (true/false)
    AUTO_APPROVE               Auto approve actions (true/false)
    DRIFT_CHECK                Enable drift detection (true/false)
    BACKUP_BEFORE_APPLY        Create backup before apply (true/false)
    PARALLEL_EXECUTION         Execute in parallel (true/false)
    SLACK_WEBHOOK_URL          Slack webhook for notifications
    DEBUG                      Enable debug output (true/false)

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        plan|apply|destroy|validate|fmt)
            ACTION="$1"
            shift
            ;;
        --environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --project)
            PROJECT_NAME="$2"
            shift 2
            ;;
        --enable-aws)
            ENABLE_AWS="true"
            shift
            ;;
        --disable-aws)
            ENABLE_AWS="false"
            shift
            ;;
        --enable-gcp)
            ENABLE_GCP="true"
            shift
            ;;
        --disable-gcp)
            ENABLE_GCP="false"
            shift
            ;;
        --enable-azure)
            ENABLE_AZURE="true"
            shift
            ;;
        --disable-azure)
            ENABLE_AZURE="false"
            shift
            ;;
        --auto-approve)
            AUTO_APPROVE="true"
            shift
            ;;
        --no-drift-check)
            DRIFT_CHECK="false"
            shift
            ;;
        --no-backup)
            BACKUP_BEFORE_APPLY="false"
            shift
            ;;
        --parallel)
            PARALLEL_EXECUTION="true"
            shift
            ;;
        --debug)
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

# Execute main function
main
