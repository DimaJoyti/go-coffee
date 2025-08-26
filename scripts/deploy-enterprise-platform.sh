#!/bin/bash

# Enterprise Go Coffee Platform Deployment Script
# This script deploys the complete enterprise platform with multi-region support

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT="${ENVIRONMENT:-production}"
DEPLOYMENT_TYPE="${DEPLOYMENT_TYPE:-single-region}"
PRIMARY_REGION="${PRIMARY_REGION:-us-east-1}"
SECONDARY_REGION="${SECONDARY_REGION:-us-west-2}"
REGISTRY="${REGISTRY:-ghcr.io/dimajoyti/go-coffee}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ENABLE_DISASTER_RECOVERY="${ENABLE_DISASTER_RECOVERY:-true}"
ENABLE_ANALYTICS="${ENABLE_ANALYTICS:-true}"
ENABLE_MULTI_TENANT="${ENABLE_MULTI_TENANT:-true}"
DRY_RUN="${DRY_RUN:-false}"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================${NC}"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Function to check prerequisites
check_enterprise_prerequisites() {
    print_header "Checking Enterprise Prerequisites"
    
    # Check required tools
    local required_tools=("kubectl" "helm" "docker" "gcloud" "terraform")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            print_error "$tool is not installed or not in PATH"
            exit 1
        fi
    done
    
    # Check cluster connectivity for all regions
    if [[ "$DEPLOYMENT_TYPE" == "multi-region" ]]; then
        print_status "Checking multi-region cluster connectivity..."
        
        # Check primary region
        if ! kubectl cluster-info --context="$PRIMARY_REGION" &> /dev/null; then
            print_error "Cannot connect to primary region cluster: $PRIMARY_REGION"
            exit 1
        fi
        
        # Check secondary region
        if ! kubectl cluster-info --context="$SECONDARY_REGION" &> /dev/null; then
            print_error "Cannot connect to secondary region cluster: $SECONDARY_REGION"
            exit 1
        fi
        
        print_success "Multi-region cluster connectivity verified"
    else
        if ! kubectl cluster-info &> /dev/null; then
            print_error "Cannot connect to Kubernetes cluster"
            exit 1
        fi
    fi
    
    # Check Google Cloud authentication
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 &> /dev/null; then
        print_error "Google Cloud authentication required"
        print_status "Run: gcloud auth login"
        exit 1
    fi
    
    print_success "Enterprise prerequisites check completed"
}

# Function to deploy infrastructure with Terraform
deploy_infrastructure() {
    print_header "Deploying Infrastructure with Terraform"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy infrastructure"
        return 0
    fi
    
    cd terraform/enterprise
    
    # Initialize Terraform
    terraform init
    
    # Plan deployment
    terraform plan \
        -var="environment=$ENVIRONMENT" \
        -var="deployment_type=$DEPLOYMENT_TYPE" \
        -var="primary_region=$PRIMARY_REGION" \
        -var="secondary_region=$SECONDARY_REGION" \
        -var="enable_disaster_recovery=$ENABLE_DISASTER_RECOVERY" \
        -var="enable_analytics=$ENABLE_ANALYTICS" \
        -var="enable_multi_tenant=$ENABLE_MULTI_TENANT" \
        -out=tfplan
    
    # Apply infrastructure
    terraform apply tfplan
    
    cd ../..
    print_success "Infrastructure deployed successfully"
}

# Function to deploy global load balancer
deploy_global_load_balancer() {
    print_header "Deploying Global Load Balancer"
    
    if [[ "$DEPLOYMENT_TYPE" != "multi-region" ]]; then
        print_status "Skipping global load balancer for single-region deployment"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy global load balancer"
        return 0
    fi
    
    # Deploy global load balancer configuration
    kubectl apply -f k8s/multi-region/global-load-balancer.yaml
    
    # Wait for global IP allocation
    print_status "Waiting for global IP allocation..."
    kubectl wait --for=condition=ready computeglobaladdress/go-coffee-global-ip --timeout=300s
    
    # Get the allocated IP
    GLOBAL_IP=$(kubectl get computeglobaladdress go-coffee-global-ip -o jsonpath='{.status.address}')
    print_success "Global load balancer deployed with IP: $GLOBAL_IP"
}

# Function to deploy disaster recovery
deploy_disaster_recovery() {
    print_header "Deploying Disaster Recovery"
    
    if [[ "$ENABLE_DISASTER_RECOVERY" != "true" ]]; then
        print_status "Disaster recovery disabled, skipping..."
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy disaster recovery"
        return 0
    fi
    
    # Deploy disaster recovery configuration
    kubectl apply -f k8s/multi-region/disaster-recovery.yaml
    
    # Setup cross-region database replication
    print_status "Setting up cross-region database replication..."
    kubectl apply -f k8s/multi-region/database-replication.yaml
    
    # Setup Kafka cross-region mirroring
    print_status "Setting up Kafka cross-region mirroring..."
    kubectl apply -f k8s/multi-region/kafka-mirroring.yaml
    
    print_success "Disaster recovery deployed successfully"
}

# Function to deploy analytics service
deploy_analytics_service() {
    print_header "Deploying Analytics & Business Intelligence Service"
    
    if [[ "$ENABLE_ANALYTICS" != "true" ]]; then
        print_status "Analytics service disabled, skipping..."
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would deploy analytics service"
        return 0
    fi
    
    # Build analytics service image
    print_step "Building analytics service image..."
    docker build -t "$REGISTRY/analytics-service:$IMAGE_TAG" -f Dockerfile.analytics-service .
    docker push "$REGISTRY/analytics-service:$IMAGE_TAG"
    
    # Deploy analytics service
    envsubst < k8s/enhanced/analytics-service.yaml | kubectl apply -f -
    
    # Wait for analytics service to be ready
    print_status "Waiting for analytics service to be ready..."
    kubectl wait --for=condition=ready pod -l app=analytics-service --timeout=300s
    
    print_success "Analytics service deployed successfully"
}

# Function to setup multi-tenant configuration
setup_multi_tenant() {
    print_header "Setting up Multi-Tenant Configuration"
    
    if [[ "$ENABLE_MULTI_TENANT" != "true" ]]; then
        print_status "Multi-tenancy disabled, skipping..."
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would setup multi-tenant configuration"
        return 0
    fi
    
    # Create tenant namespaces
    local tenants=("tenant-demo" "tenant-enterprise" "tenant-startup")
    
    for tenant in "${tenants[@]}"; do
        print_step "Creating namespace for $tenant..."
        
        kubectl create namespace "$tenant" --dry-run=client -o yaml | kubectl apply -f -
        
        # Apply tenant-specific configurations
        kubectl label namespace "$tenant" tenant="$tenant"
        kubectl annotate namespace "$tenant" tenant.go-coffee.com/tier="standard"
        
        # Create tenant-specific secrets
        kubectl create secret generic "$tenant-secrets" \
            --from-literal=database-url="postgres://user:pass@postgres:5432/${tenant}_db" \
            --namespace="$tenant" \
            --dry-run=client -o yaml | kubectl apply -f -
    done
    
    print_success "Multi-tenant configuration completed"
}

# Function to run enterprise tests
run_enterprise_tests() {
    print_header "Running Enterprise Test Suite"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would run enterprise tests"
        return 0
    fi
    
    # Run comprehensive test suite
    local test_categories=("performance" "security" "disaster-recovery" "multi-region" "analytics")
    
    for category in "${test_categories[@]}"; do
        print_step "Running $category tests..."
        
        case $category in
            "performance")
                ./scripts/test-performance.sh
                ;;
            "security")
                ./scripts/test-security.sh
                ;;
            "disaster-recovery")
                ./scripts/test-disaster-recovery.sh
                ;;
            "multi-region")
                if [[ "$DEPLOYMENT_TYPE" == "multi-region" ]]; then
                    ./scripts/test-multi-region.sh
                fi
                ;;
            "analytics")
                if [[ "$ENABLE_ANALYTICS" == "true" ]]; then
                    ./scripts/test-analytics.sh
                fi
                ;;
        esac
    done
    
    print_success "Enterprise test suite completed"
}

# Function to setup monitoring and alerting
setup_enterprise_monitoring() {
    print_header "Setting up Enterprise Monitoring & Alerting"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "DRY RUN: Would setup enterprise monitoring"
        return 0
    fi
    
    # Deploy enhanced monitoring stack
    helm upgrade --install monitoring-stack \
        ./helm/monitoring-stack \
        --namespace go-coffee-monitoring \
        --create-namespace \
        --values helm/monitoring-stack/values-enterprise.yaml \
        --set global.environment="$ENVIRONMENT" \
        --set global.multiRegion="$([[ $DEPLOYMENT_TYPE == "multi-region" ]] && echo true || echo false)"
    
    # Setup alerting rules
    kubectl apply -f k8s/enhanced/alerting-rules.yaml
    
    # Configure notification channels
    kubectl apply -f k8s/enhanced/notification-channels.yaml
    
    print_success "Enterprise monitoring setup completed"
}

# Function to display enterprise deployment summary
display_enterprise_summary() {
    print_header "Enterprise Deployment Summary"
    
    echo -e "${CYAN}Deployment Configuration:${NC}"
    echo -e "Environment: $ENVIRONMENT"
    echo -e "Deployment Type: $DEPLOYMENT_TYPE"
    echo -e "Primary Region: $PRIMARY_REGION"
    if [[ "$DEPLOYMENT_TYPE" == "multi-region" ]]; then
        echo -e "Secondary Region: $SECONDARY_REGION"
    fi
    echo -e "Disaster Recovery: $ENABLE_DISASTER_RECOVERY"
    echo -e "Analytics Service: $ENABLE_ANALYTICS"
    echo -e "Multi-Tenancy: $ENABLE_MULTI_TENANT"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        echo ""
        echo -e "${CYAN}Service Endpoints:${NC}"
        
        # Get service endpoints
        if [[ "$DEPLOYMENT_TYPE" == "multi-region" ]]; then
            echo -e "${GREEN}Global Load Balancer:${NC} https://go-coffee.com"
            echo -e "${GREEN}API Endpoint:${NC}        https://api.go-coffee.com"
            echo -e "${GREEN}Analytics Dashboard:${NC} https://analytics.go-coffee.com"
        else
            local producer_ip=$(kubectl get service producer-service -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
            local analytics_ip=$(kubectl get service analytics-service -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
            
            echo -e "${GREEN}Producer Service:${NC}     http://$producer_ip"
            echo -e "${GREEN}Analytics Service:${NC}    http://$analytics_ip"
        fi
        
        echo ""
        echo -e "${CYAN}Monitoring & Observability:${NC}"
        echo -e "${GREEN}Grafana:${NC}              http://grafana.go-coffee.com"
        echo -e "${GREEN}Jaeger:${NC}               http://jaeger.go-coffee.com"
        echo -e "${GREEN}Prometheus:${NC}           http://prometheus.go-coffee.com"
        
        if [[ "$ENABLE_ANALYTICS" == "true" ]]; then
            echo ""
            echo -e "${CYAN}Analytics & BI:${NC}"
            echo -e "${GREEN}Business Dashboard:${NC}   https://analytics.go-coffee.com/dashboard"
            echo -e "${GREEN}Real-time Metrics:${NC}    https://analytics.go-coffee.com/realtime"
            echo -e "${GREEN}Reports:${NC}              https://analytics.go-coffee.com/reports"
        fi
        
        if [[ "$ENABLE_MULTI_TENANT" == "true" ]]; then
            echo ""
            echo -e "${CYAN}Multi-Tenant Access:${NC}"
            echo -e "${GREEN}Demo Tenant:${NC}          https://demo.go-coffee.com"
            echo -e "${GREEN}Enterprise Tenant:${NC}    https://enterprise.go-coffee.com"
            echo -e "${GREEN}Startup Tenant:${NC}       https://startup.go-coffee.com"
        fi
    fi
    
    print_success "Enterprise Go Coffee Platform deployment completed!"
    
    if [[ "$DEPLOYMENT_TYPE" == "multi-region" ]]; then
        print_status "Your platform is now running across multiple regions with:"
        echo "  • Global load balancing and traffic routing"
        echo "  • Cross-region disaster recovery"
        echo "  • Real-time data replication"
        echo "  • Automatic failover capabilities"
    fi
    
    if [[ "$ENABLE_ANALYTICS" == "true" ]]; then
        print_status "Advanced analytics and business intelligence features:"
        echo "  • Real-time dashboards and KPIs"
        echo "  • Predictive analytics and forecasting"
        echo "  • Customer segmentation and LTV analysis"
        echo "  • Automated report generation"
    fi
}

# Main deployment function
main() {
    print_header "Enterprise Go Coffee Platform Deployment"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --deployment-type)
                DEPLOYMENT_TYPE="$2"
                shift 2
                ;;
            --primary-region)
                PRIMARY_REGION="$2"
                shift 2
                ;;
            --secondary-region)
                SECONDARY_REGION="$2"
                shift 2
                ;;
            --disable-disaster-recovery)
                ENABLE_DISASTER_RECOVERY="false"
                shift
                ;;
            --disable-analytics)
                ENABLE_ANALYTICS="false"
                shift
                ;;
            --disable-multi-tenant)
                ENABLE_MULTI_TENANT="false"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --environment ENV              Deployment environment (default: production)"
                echo "  --deployment-type TYPE         single-region or multi-region (default: single-region)"
                echo "  --primary-region REGION        Primary region (default: us-east-1)"
                echo "  --secondary-region REGION      Secondary region (default: us-west-2)"
                echo "  --disable-disaster-recovery    Disable disaster recovery features"
                echo "  --disable-analytics            Disable analytics service"
                echo "  --disable-multi-tenant         Disable multi-tenancy"
                echo "  --dry-run                      Run in dry-run mode"
                echo "  --help                         Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_enterprise_prerequisites
    deploy_infrastructure
    deploy_global_load_balancer
    deploy_disaster_recovery
    deploy_analytics_service
    setup_multi_tenant
    setup_enterprise_monitoring
    run_enterprise_tests
    display_enterprise_summary
}

# Run main function with all arguments
main "$@"
