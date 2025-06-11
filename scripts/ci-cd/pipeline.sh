#!/bin/bash

# Go Coffee - Advanced CI/CD Pipeline
# Comprehensive CI/CD pipeline with testing, security, deployment, and monitoring
# Version: 3.0.0
# Usage: ./pipeline.sh [OPTIONS] STAGE
#   -e, --environment   Target environment (development|staging|production)
#   -b, --branch        Git branch to deploy
#   -t, --tag           Git tag to deploy
#   -s, --skip-tests    Skip test execution
#   -f, --force         Force deployment without confirmations
#   -d, --dry-run       Dry run mode (no actual deployment)
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Source shared library
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🚀 Go Coffee Advanced CI/CD Pipeline"

# =============================================================================
# CONFIGURATION
# =============================================================================

ENVIRONMENT="${ENVIRONMENT:-development}"
GIT_BRANCH="${GIT_BRANCH:-main}"
GIT_TAG=""
SKIP_TESTS=false
FORCE_DEPLOYMENT=false
DRY_RUN=false
PIPELINE_STAGE=""

# Pipeline configuration
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io/dimajoyti}"
KUBERNETES_NAMESPACE="${KUBERNETES_NAMESPACE:-go-coffee}"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"
TEAMS_WEBHOOK="${TEAMS_WEBHOOK:-}"

# Pipeline stages
PIPELINE_STAGES=(
    "validate"
    "build"
    "test"
    "security"
    "package"
    "deploy"
    "verify"
    "notify"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_pipeline_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -b|--branch)
                GIT_BRANCH="$2"
                shift 2
                ;;
            -t|--tag)
                GIT_TAG="$2"
                shift 2
                ;;
            -s|--skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            -f|--force)
                FORCE_DEPLOYMENT=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_usage "pipeline.sh" \
                    "Advanced CI/CD pipeline for Go Coffee platform" \
                    "  ./pipeline.sh [OPTIONS] STAGE
  
  Stages:
    validate    Validate code and environment
    build       Build all services
    test        Run comprehensive tests
    security    Security scanning and analysis
    package     Package and containerize
    deploy      Deploy to target environment
    verify      Post-deployment verification
    notify      Send notifications
    full        Run complete pipeline
  
  Options:
    -e, --environment   Target environment (development|staging|production)
    -b, --branch        Git branch to deploy (default: main)
    -t, --tag           Git tag to deploy (overrides branch)
    -s, --skip-tests    Skip test execution
    -f, --force         Force deployment without confirmations
    -d, --dry-run       Dry run mode (no actual deployment)
    -h, --help          Show this help message
  
  Examples:
    ./pipeline.sh full                          # Complete pipeline
    ./pipeline.sh deploy -e production -t v1.0.0 # Deploy tag to production
    ./pipeline.sh test -s                       # Run tests only
    ./pipeline.sh build -d                      # Dry run build"
                exit 0
                ;;
            validate|build|test|security|package|deploy|verify|notify|full)
                PIPELINE_STAGE="$1"
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$PIPELINE_STAGE" ]]; then
        print_error "No pipeline stage specified"
        exit 1
    fi
}

# =============================================================================
# PIPELINE STAGE FUNCTIONS
# =============================================================================

# Stage 1: Validate
stage_validate() {
    print_header "✅ Stage 1: Validation"
    
    # Check Git repository state
    print_progress "Checking Git repository state..."
    
    if [[ ! -d ".git" ]]; then
        print_error "Not a Git repository"
        exit 1
    fi
    
    # Check for uncommitted changes
    if [[ -n "$(git status --porcelain)" ]] && [[ "$FORCE_DEPLOYMENT" != "true" ]]; then
        print_error "Uncommitted changes detected. Use --force to override."
        exit 1
    fi
    
    # Validate target branch/tag
    if [[ -n "$GIT_TAG" ]]; then
        if ! git tag | grep -q "^$GIT_TAG$"; then
            print_error "Git tag '$GIT_TAG' does not exist"
            exit 1
        fi
        print_status "Using Git tag: $GIT_TAG"
    else
        if ! git branch -r | grep -q "origin/$GIT_BRANCH"; then
            print_error "Git branch '$GIT_BRANCH' does not exist on remote"
            exit 1
        fi
        print_status "Using Git branch: $GIT_BRANCH"
    fi
    
    # Check environment configuration
    print_progress "Validating environment configuration..."
    
    case "$ENVIRONMENT" in
        "development"|"staging"|"production")
            print_status "Environment '$ENVIRONMENT' is valid"
            ;;
        *)
            print_error "Invalid environment: $ENVIRONMENT"
            exit 1
            ;;
    esac
    
    # Check required tools
    local required_tools=("docker" "kubectl" "helm" "git")
    check_dependencies "${required_tools[@]}" || exit 1
    
    print_success "Validation stage completed"
}

# Stage 2: Build
stage_build() {
    print_header "🔨 Stage 2: Build"
    
    # Checkout target branch/tag
    if [[ -n "$GIT_TAG" ]]; then
        print_progress "Checking out tag: $GIT_TAG"
        git checkout "$GIT_TAG"
    else
        print_progress "Checking out branch: $GIT_BRANCH"
        git checkout "$GIT_BRANCH"
        git pull origin "$GIT_BRANCH"
    fi
    
    # Build all services
    print_progress "Building all Go Coffee services..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would build all services"
    else
        if ! "$PROJECT_ROOT/build_all.sh" --core-only; then
            print_error "Build failed"
            exit 1
        fi
    fi
    
    # Generate build metadata
    local build_info_file="build-info.json"
    cat > "$build_info_file" <<EOF
{
    "build_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "git_commit": "$(git rev-parse HEAD)",
    "git_branch": "$GIT_BRANCH",
    "git_tag": "$GIT_TAG",
    "environment": "$ENVIRONMENT",
    "pipeline_version": "3.0.0"
}
EOF
    
    print_success "Build stage completed"
}

# Stage 3: Test
stage_test() {
    print_header "🧪 Stage 3: Testing"
    
    if [[ "$SKIP_TESTS" == "true" ]]; then
        print_warning "Skipping tests as requested"
        return 0
    fi
    
    # Run unit tests
    print_progress "Running unit tests..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would run unit tests"
    else
        if ! "$PROJECT_ROOT/scripts/test-all-services.sh" --core-only --fast; then
            print_error "Unit tests failed"
            exit 1
        fi
    fi
    
    # Run integration tests
    print_progress "Running integration tests..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would run integration tests"
    else
        # Start services for integration testing
        "$PROJECT_ROOT/scripts/start-all-services.sh" --core-only &
        local services_pid=$!
        
        # Wait for services to be ready
        sleep 30
        
        # Run integration tests
        if ! "$PROJECT_ROOT/scripts/test-all-services.sh" --core-only; then
            kill $services_pid 2>/dev/null || true
            print_error "Integration tests failed"
            exit 1
        fi
        
        # Stop services
        kill $services_pid 2>/dev/null || true
    fi
    
    print_success "Testing stage completed"
}

# Stage 4: Security
stage_security() {
    print_header "🔒 Stage 4: Security Scanning"
    
    print_progress "Running security scans..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would run security scans"
    else
        if ! "$PROJECT_ROOT/scripts/security/security-scan.sh" --scope full --level standard --quiet; then
            print_warning "Security scan completed with findings"
            # Don't fail pipeline on security findings in non-production
            if [[ "$ENVIRONMENT" == "production" ]]; then
                print_error "Security scan failed in production environment"
                exit 1
            fi
        fi
    fi
    
    print_success "Security scanning completed"
}

# Stage 5: Package
stage_package() {
    print_header "📦 Stage 5: Packaging"
    
    # Build Docker images
    print_progress "Building Docker images..."
    
    local image_tag
    if [[ -n "$GIT_TAG" ]]; then
        image_tag="$GIT_TAG"
    else
        image_tag="$(git rev-parse --short HEAD)"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would build Docker images with tag: $image_tag"
    else
        # Build and tag images
        local services=("api-gateway" "auth-service" "payment-service" "order-service")
        
        for service in "${services[@]}"; do
            print_progress "Building Docker image for $service..."
            
            if [[ -f "cmd/$service/Dockerfile" ]]; then
                docker build -t "$DOCKER_REGISTRY/go-coffee-$service:$image_tag" \
                    -f "cmd/$service/Dockerfile" .
                
                # Tag as latest for development
                if [[ "$ENVIRONMENT" == "development" ]]; then
                    docker tag "$DOCKER_REGISTRY/go-coffee-$service:$image_tag" \
                        "$DOCKER_REGISTRY/go-coffee-$service:latest"
                fi
            fi
        done
        
        # Push images to registry
        if [[ "$ENVIRONMENT" != "development" ]]; then
            print_progress "Pushing images to registry..."
            
            for service in "${services[@]}"; do
                docker push "$DOCKER_REGISTRY/go-coffee-$service:$image_tag"
                
                if [[ "$ENVIRONMENT" == "development" ]]; then
                    docker push "$DOCKER_REGISTRY/go-coffee-$service:latest"
                fi
            done
        fi
    fi
    
    print_success "Packaging stage completed"
}

# Stage 6: Deploy
stage_deploy() {
    print_header "🚀 Stage 6: Deployment"
    
    # Confirm deployment for production
    if [[ "$ENVIRONMENT" == "production" ]] && [[ "$FORCE_DEPLOYMENT" != "true" ]]; then
        print_warning "Deploying to PRODUCTION environment"
        read -p "Are you sure you want to continue? (yes/no): " -r
        if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            print_info "Deployment cancelled"
            exit 0
        fi
    fi
    
    print_progress "Deploying to $ENVIRONMENT environment..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would deploy to $ENVIRONMENT"
    else
        # Deploy using enhanced deployment script
        if ! "$PROJECT_ROOT/scripts/deploy.sh" --env "$ENVIRONMENT" --backup k8s; then
            print_error "Deployment failed"
            exit 1
        fi
    fi
    
    print_success "Deployment stage completed"
}

# Stage 7: Verify
stage_verify() {
    print_header "✅ Stage 7: Post-Deployment Verification"
    
    print_progress "Running post-deployment verification..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would run post-deployment verification"
    else
        # Wait for deployment to stabilize
        sleep 30
        
        # Run health checks
        if ! "$PROJECT_ROOT/scripts/health-check.sh" --environment "$ENVIRONMENT"; then
            print_error "Post-deployment health checks failed"
            exit 1
        fi
        
        # Run smoke tests
        print_progress "Running smoke tests..."
        
        # Basic API smoke tests
        local api_url
        case "$ENVIRONMENT" in
            "development")
                api_url="http://localhost:8080"
                ;;
            "staging")
                api_url="https://staging-api.go-coffee.com"
                ;;
            "production")
                api_url="https://api.go-coffee.com"
                ;;
        esac
        
        if ! curl -s --max-time 10 "$api_url/health" >/dev/null; then
            print_error "Smoke test failed: API not responding"
            exit 1
        fi
    fi
    
    print_success "Verification stage completed"
}

# Stage 8: Notify
stage_notify() {
    print_header "📢 Stage 8: Notifications"
    
    local deployment_status="SUCCESS"
    local git_info
    
    if [[ -n "$GIT_TAG" ]]; then
        git_info="Tag: $GIT_TAG"
    else
        git_info="Branch: $GIT_BRANCH ($(git rev-parse --short HEAD))"
    fi
    
    local message="🚀 Go Coffee deployment to $ENVIRONMENT completed successfully!
    
📋 Deployment Details:
• Environment: $ENVIRONMENT
• $git_info
• Pipeline Version: 3.0.0
• Timestamp: $(date)

✅ All stages completed successfully"
    
    # Send Slack notification
    if [[ -n "$SLACK_WEBHOOK" ]] && [[ "$DRY_RUN" != "true" ]]; then
        print_progress "Sending Slack notification..."
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$message\"}" \
            "$SLACK_WEBHOOK" >/dev/null 2>&1 || true
    fi
    
    # Send Teams notification
    if [[ -n "$TEAMS_WEBHOOK" ]] && [[ "$DRY_RUN" != "true" ]]; then
        print_progress "Sending Teams notification..."
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$message\"}" \
            "$TEAMS_WEBHOOK" >/dev/null 2>&1 || true
    fi
    
    print_success "Notifications sent"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)
    
    # Parse arguments
    parse_pipeline_args "$@"
    
    print_info "CI/CD Pipeline Configuration:"
    print_info "  Stage: $PIPELINE_STAGE"
    print_info "  Environment: $ENVIRONMENT"
    print_info "  Branch: $GIT_BRANCH"
    print_info "  Tag: ${GIT_TAG:-none}"
    print_info "  Skip Tests: $SKIP_TESTS"
    print_info "  Force: $FORCE_DEPLOYMENT"
    print_info "  Dry Run: $DRY_RUN"
    
    # Execute pipeline stage(s)
    case "$PIPELINE_STAGE" in
        "validate")
            stage_validate
            ;;
        "build")
            stage_build
            ;;
        "test")
            stage_test
            ;;
        "security")
            stage_security
            ;;
        "package")
            stage_package
            ;;
        "deploy")
            stage_deploy
            ;;
        "verify")
            stage_verify
            ;;
        "notify")
            stage_notify
            ;;
        "full")
            stage_validate
            stage_build
            stage_test
            stage_security
            stage_package
            stage_deploy
            stage_verify
            stage_notify
            ;;
        *)
            print_error "Unknown pipeline stage: $PIPELINE_STAGE"
            exit 1
            ;;
    esac
    
    # Calculate pipeline time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    print_success "🎉 Pipeline stage '$PIPELINE_STAGE' completed in ${total_time}s"
    
    if [[ "$PIPELINE_STAGE" == "full" ]]; then
        print_header "🏆 Complete Pipeline Summary"
        print_success "All pipeline stages completed successfully!"
        print_info "Environment: $ENVIRONMENT"
        print_info "Total Time: ${total_time}s"
        print_info "Git Reference: ${GIT_TAG:-$GIT_BRANCH}"
    fi
}

# Run main function with all arguments
main "$@"
