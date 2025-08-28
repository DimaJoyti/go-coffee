#!/bin/bash

# Multi-Cloud Serverless Stack Deployment Script
# Deploys serverless functions across AWS, GCP, and Azure

set -euo pipefail

# Make script executable
chmod +x "$0"

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform/modules/serverless-orchestrator"
FUNCTIONS_DIR="$PROJECT_ROOT/serverless/functions"

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="${PROJECT_NAME:-go-coffee}"
ENABLE_AWS="${ENABLE_AWS:-true}"
ENABLE_GCP="${ENABLE_GCP:-true}"
ENABLE_AZURE="${ENABLE_AZURE:-false}"
DRY_RUN="${DRY_RUN:-false}"
VERBOSE="${VERBOSE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check required tools
    local required_tools=("terraform" "go" "zip" "jq")
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool is required but not installed"
            exit 1
        fi
    done
    
    # Check cloud CLI tools based on enabled providers
    if [[ "$ENABLE_AWS" == "true" ]]; then
        if ! command -v aws &> /dev/null; then
            log_error "AWS CLI is required when AWS is enabled"
            exit 1
        fi
        
        # Check AWS credentials
        if ! aws sts get-caller-identity &> /dev/null; then
            log_error "AWS credentials not configured"
            exit 1
        fi
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]]; then
        if ! command -v gcloud &> /dev/null; then
            log_error "Google Cloud CLI is required when GCP is enabled"
            exit 1
        fi
        
        # Check GCP authentication
        if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 &> /dev/null; then
            log_error "GCP authentication not configured"
            exit 1
        fi
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]]; then
        if ! command -v az &> /dev/null; then
            log_error "Azure CLI is required when Azure is enabled"
            exit 1
        fi
        
        # Check Azure authentication
        if ! az account show &> /dev/null; then
            log_error "Azure authentication not configured"
            exit 1
        fi
    fi
    
    log_success "Prerequisites check completed"
}

build_functions() {
    log_info "Building serverless functions..."
    
    local functions=("coffee-order-processor" "ai-agent-coordinator" "defi-arbitrage-scanner" "inventory-optimizer" "notification-dispatcher")
    
    for func in "${functions[@]}"; do
        local func_dir="$FUNCTIONS_DIR/$func"
        
        if [[ ! -d "$func_dir" ]]; then
            log_warning "Function directory not found: $func_dir"
            continue
        fi
        
        log_info "Building function: $func"
        
        # Create deployment directory
        local deploy_dir="$func_dir/deployment"
        mkdir -p "$deploy_dir"
        
        # Build Go binary
        cd "$func_dir"
        
        if [[ "$ENABLE_AWS" == "true" ]]; then
            log_info "Building for AWS Lambda..."
            GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$deploy_dir/main" main.go
            
            # Create deployment package
            cd "$deploy_dir"
            zip -r "../deployment.zip" main
            cd ..
        fi
        
        if [[ "$ENABLE_GCP" == "true" ]]; then
            log_info "Building for Google Cloud Functions..."
            # For Cloud Functions, we need the source code
            cp -r . "$deploy_dir/gcp/"
            cd "$deploy_dir/gcp"
            zip -r "../gcp-deployment.zip" .
            cd ../..
        fi
        
        if [[ "$ENABLE_AZURE" == "true" ]]; then
            log_info "Building for Azure Functions..."
            GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$deploy_dir/main" main.go
            
            # Create Azure Functions structure
            local azure_dir="$deploy_dir/azure"
            mkdir -p "$azure_dir"
            cp "$deploy_dir/main" "$azure_dir/"
            
            # Create host.json for Azure Functions
            cat > "$azure_dir/host.json" << EOF
{
    "version": "2.0",
    "extensionBundle": {
        "id": "Microsoft.Azure.Functions.ExtensionBundle",
        "version": "[2.*, 3.0.0)"
    },
    "functionTimeout": "00:05:00"
}
EOF
            
            cd "$azure_dir"
            zip -r "../azure-deployment.zip" .
            cd ../../..
        fi
        
        log_success "Built function: $func"
    done
    
    cd "$PROJECT_ROOT"
    log_success "All functions built successfully"
}

deploy_infrastructure() {
    log_info "Deploying serverless infrastructure..."
    
    cd "$TERRAFORM_DIR"
    
    # Initialize Terraform
    log_info "Initializing Terraform..."
    terraform init
    
    # Create terraform.tfvars
    cat > terraform.tfvars << EOF
project_name = "$PROJECT_NAME"
environment = "$ENVIRONMENT"
enable_aws = $ENABLE_AWS
enable_gcp = $ENABLE_GCP
enable_azure = $ENABLE_AZURE
enable_cross_cloud_routing = true

# Database configuration
database_url = "$DATABASE_URL"
kafka_brokers = "$KAFKA_BROKERS"
redis_url = "$REDIS_URL"

# Monitoring
monitoring_enabled = true
log_level = "info"
log_retention_days = 30

# Security
encryption_at_rest = true
encryption_in_transit = true

# Cost optimization
cost_optimization_enabled = true
auto_scaling_enabled = true
EOF
    
    # Add cloud-specific configurations
    if [[ "$ENABLE_AWS" == "true" ]]; then
        cat >> terraform.tfvars << EOF

# AWS Configuration
aws_region = "${AWS_REGION:-us-east-1}"
aws_vpc_id = "$AWS_VPC_ID"
aws_subnet_ids = ["$AWS_SUBNET_IDS"]
EOF
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]]; then
        cat >> terraform.tfvars << EOF

# GCP Configuration
gcp_project_id = "$GCP_PROJECT_ID"
gcp_region = "${GCP_REGION:-us-central1}"
gcp_vpc_connector = "$GCP_VPC_CONNECTOR"
EOF
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]]; then
        cat >> terraform.tfvars << EOF

# Azure Configuration
azure_location = "${AZURE_LOCATION:-East US}"
azure_resource_group_name = "$AZURE_RESOURCE_GROUP_NAME"
EOF
    fi
    
    # Plan deployment
    log_info "Planning Terraform deployment..."
    if [[ "$DRY_RUN" == "true" ]]; then
        terraform plan -var-file=terraform.tfvars
        log_info "DRY RUN: Terraform plan completed"
        return 0
    fi
    
    # Apply deployment
    log_info "Applying Terraform deployment..."
    terraform apply -var-file=terraform.tfvars -auto-approve
    
    log_success "Infrastructure deployed successfully"
    cd "$PROJECT_ROOT"
}

verify_deployment() {
    log_info "Verifying serverless deployment..."
    
    local verification_failed=false
    
    if [[ "$ENABLE_AWS" == "true" ]]; then
        log_info "Verifying AWS Lambda functions..."
        
        local aws_functions=("coffee-order-processor" "ai-agent-coordinator" "defi-arbitrage-scanner")
        
        for func in "${aws_functions[@]}"; do
            local func_name="$PROJECT_NAME-$ENVIRONMENT-$func"
            
            if aws lambda get-function --function-name "$func_name" &> /dev/null; then
                log_success "AWS Lambda function verified: $func_name"
            else
                log_error "AWS Lambda function not found: $func_name"
                verification_failed=true
            fi
        done
    fi
    
    if [[ "$ENABLE_GCP" == "true" ]]; then
        log_info "Verifying Google Cloud Functions..."
        
        local gcp_functions=("coffee-order-processor" "ai-agent-coordinator" "defi-arbitrage-scanner")
        
        for func in "${gcp_functions[@]}"; do
            local func_name="$PROJECT_NAME-$ENVIRONMENT-$func"
            
            if gcloud functions describe "$func_name" --region="${GCP_REGION:-us-central1}" &> /dev/null; then
                log_success "Google Cloud Function verified: $func_name"
            else
                log_error "Google Cloud Function not found: $func_name"
                verification_failed=true
            fi
        done
    fi
    
    if [[ "$ENABLE_AZURE" == "true" ]]; then
        log_info "Verifying Azure Functions..."
        
        local azure_functions=("coffee-order-processor" "ai-agent-coordinator" "defi-arbitrage-scanner")
        
        for func in "${azure_functions[@]}"; do
            local func_name="$PROJECT_NAME-$ENVIRONMENT-$func"
            
            if az functionapp show --name "$func_name" --resource-group "$AZURE_RESOURCE_GROUP_NAME" &> /dev/null; then
                log_success "Azure Function verified: $func_name"
            else
                log_error "Azure Function not found: $func_name"
                verification_failed=true
            fi
        done
    fi
    
    if [[ "$verification_failed" == "true" ]]; then
        log_error "Deployment verification failed"
        exit 1
    fi
    
    log_success "Deployment verification completed successfully"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    
    # Clean up deployment artifacts
    find "$FUNCTIONS_DIR" -name "deployment" -type d -exec rm -rf {} + 2>/dev/null || true
    find "$FUNCTIONS_DIR" -name "*.zip" -type f -delete 2>/dev/null || true
    
    log_success "Cleanup completed"
}

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy multi-cloud serverless stack for Go Coffee platform.

OPTIONS:
    -e, --environment ENV    Environment (dev, staging, prod) [default: dev]
    -p, --project NAME       Project name [default: go-coffee]
    --enable-aws             Enable AWS deployment [default: true]
    --enable-gcp             Enable GCP deployment [default: true]
    --enable-azure           Enable Azure deployment [default: false]
    --dry-run               Perform dry run without actual deployment
    -v, --verbose           Enable verbose output
    -h, --help              Show this help message

EXAMPLES:
    $0                                    # Deploy to dev environment
    $0 -e prod --enable-aws --enable-gcp # Deploy to prod with AWS and GCP
    $0 --dry-run                         # Perform dry run
    $0 --help                            # Show help

ENVIRONMENT VARIABLES:
    DATABASE_URL             Database connection URL
    KAFKA_BROKERS           Kafka broker endpoints
    REDIS_URL               Redis connection URL
    AWS_REGION              AWS region
    AWS_VPC_ID              AWS VPC ID
    AWS_SUBNET_IDS          AWS subnet IDs (comma-separated)
    GCP_PROJECT_ID          Google Cloud project ID
    GCP_REGION              Google Cloud region
    GCP_VPC_CONNECTOR       GCP VPC connector
    AZURE_LOCATION          Azure location
    AZURE_RESOURCE_GROUP_NAME Azure resource group name

EOF
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -p|--project)
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
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            -v|--verbose)
                VERBOSE="true"
                set -x
                shift
                ;;
            -h|--help)
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
    
    # Validate environment
    if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
        log_error "Invalid environment: $ENVIRONMENT. Must be dev, staging, or prod."
        exit 1
    fi
    
    print_header "🚀 Go Coffee Multi-Cloud Serverless Deployment"
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Project: $PROJECT_NAME"
    log_info "  AWS: $ENABLE_AWS"
    log_info "  GCP: $ENABLE_GCP"
    log_info "  Azure: $ENABLE_AZURE"
    log_info "  Dry Run: $DRY_RUN"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    build_functions
    deploy_infrastructure
    
    if [[ "$DRY_RUN" != "true" ]]; then
        verify_deployment
    fi
    
    print_header "✅ Deployment Completed Successfully"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Serverless stack deployed successfully!"
        log_info "Next steps:"
        log_info "  1. Test the deployed functions"
        log_info "  2. Monitor function performance"
        log_info "  3. Set up alerts and notifications"
        log_info "  4. Configure auto-scaling policies"
    else
        log_info "Dry run completed. No resources were deployed."
    fi
}

# Execute main function
main "$@"
