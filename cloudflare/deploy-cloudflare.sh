#!/bin/bash

# Go Coffee Platform - Cloudflare Deployment Script
# This script deploys the complete Go Coffee platform to Cloudflare

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLOUDFLARE_DIR="$PROJECT_ROOT/cloudflare"
ENVIRONMENT="${ENVIRONMENT:-production}"
DRY_RUN="${DRY_RUN:-false}"

# Cloudflare resource IDs (created via MCP tools)
KV_CACHE_ID="f12294ae42924b729024f030e9b5611c"
KV_SESSIONS_ID="be1b38800d8d4f00820a04d1e5866552"
KV_ORDERS_ID="6a65944658824e6b8bbc9f9e24e10317"
R2_ASSETS_BUCKET="go-coffee-assets"
R2_IMAGES_BUCKET="go-coffee-images"
R2_BACKUPS_BUCKET="go-coffee-backups"

# Logging functions
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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if wrangler is installed
    if ! command -v wrangler &> /dev/null; then
        log_error "Wrangler CLI is not installed. Please install it with: npm install -g wrangler"
        exit 1
    fi
    
    # Check if user is logged in to Cloudflare
    if ! wrangler whoami &> /dev/null; then
        log_error "Not logged in to Cloudflare. Please run: wrangler login"
        exit 1
    fi
    
    # Check if Node.js is installed
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed. Please install Node.js 18 or later."
        exit 1
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    log_success "All prerequisites met"
}

# Deploy Workers
deploy_workers() {
    log_info "Deploying Cloudflare Workers..."
    
    local workers=(
        "ai-agent-coordinator"
        "cross-cloud-event-router"
        "coffee-order-processor"
    )
    
    for worker in "${workers[@]}"; do
        log_info "Deploying $worker..."
        
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "DRY RUN: Would deploy $worker to $ENVIRONMENT"
        else
            cd "$CLOUDFLARE_DIR/workers/$worker"
            
            # Deploy to specified environment
            if wrangler deploy --env "$ENVIRONMENT"; then
                log_success "Successfully deployed $worker"
            else
                log_error "Failed to deploy $worker"
                return 1
            fi
        fi
    done
    
    log_success "All Workers deployed successfully"
}

# Deploy Pages (Frontend)
deploy_pages() {
    log_info "Deploying Cloudflare Pages..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would deploy frontend to Pages"
        return 0
    fi
    
    # Build frontend
    log_info "Building frontend..."
    cd "$PROJECT_ROOT/web-ui/frontend"
    
    if [[ ! -f "package.json" ]]; then
        log_error "Frontend package.json not found. Please ensure the frontend is properly set up."
        return 1
    fi
    
    # Install dependencies
    npm ci
    
    # Build for production
    npm run build
    
    # Deploy to Pages
    log_info "Deploying to Cloudflare Pages..."
    cd "$CLOUDFLARE_DIR/pages/go-coffee-frontend"

    if wrangler pages deploy "$PROJECT_ROOT/web-ui/frontend/.next" --project-name "go-coffee-frontend" --env "$ENVIRONMENT"; then
        log_success "Successfully deployed frontend to Pages"
    else
        log_error "Failed to deploy frontend to Pages"
        return 1
    fi
}

# Set up secrets
setup_secrets() {
    log_info "Setting up secrets..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would set up secrets"
        return 0
    fi
    
    # List of secrets to set up
    local secrets=(
        "STRIPE_SECRET_KEY"
        "DATABASE_URL"
        "JWT_SECRET"
        "TWILIO_AUTH_TOKEN"
        "SENDGRID_API_KEY"
    )
    
    for secret in "${secrets[@]}"; do
        if [[ -n "${!secret:-}" ]]; then
            log_info "Setting secret: $secret"
            
            # Set secret for each worker
            for worker in "go-coffee-ai-coordinator" "go-coffee-event-router" "go-coffee-order-processor"; do
                echo "${!secret}" | wrangler secret put "$secret" --name "$worker" --env "$ENVIRONMENT"
            done
        else
            log_warning "Secret $secret not found in environment variables"
        fi
    done
    
    log_success "Secrets configured"
}

# Create queues
create_queues() {
    log_info "Creating Cloudflare Queues..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would create queues"
        return 0
    fi
    
    local queues=(
        "ai-coordination-tasks"
        "cross-cloud-events"
        "coffee-orders"
        "payment-processing"
    )
    
    for queue in "${queues[@]}"; do
        log_info "Creating queue: $queue"
        
        if wrangler queues create "$queue"; then
            log_success "Created queue: $queue"
        else
            log_warning "Queue $queue may already exist"
        fi
        
        # Create dead letter queue
        if wrangler queues create "$queue-dlq"; then
            log_success "Created DLQ: $queue-dlq"
        else
            log_warning "DLQ $queue-dlq may already exist"
        fi
    done
    
    log_success "Queues created"
}

# Configure custom domains
configure_domains() {
    log_info "Configuring custom domains..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would configure domains"
        return 0
    fi
    
    # Note: Domain configuration typically requires manual setup in Cloudflare dashboard
    # or additional API calls not covered by wrangler CLI
    
    log_info "Custom domains need to be configured manually in Cloudflare dashboard:"
    log_info "1. go-coffee.com -> Pages project"
    log_info "2. api.go-coffee.com -> go-coffee-order-processor worker"
    log_info "3. events.go-coffee.com -> go-coffee-event-router worker"
    log_info "4. ai.go-coffee.com -> go-coffee-ai-coordinator worker"
    
    log_success "Domain configuration instructions provided"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would verify deployment"
        return 0
    fi
    
    # Test endpoints
    local endpoints=(
        "https://go-coffee.com"
        "https://api.go-coffee.com/health"
        "https://events.go-coffee.com/health"
        "https://ai.go-coffee.com/health"
    )
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        
        if curl -f -s "$endpoint" > /dev/null; then
            log_success "✓ $endpoint is responding"
        else
            log_warning "✗ $endpoint is not responding (may need DNS propagation)"
        fi
    done
    
    log_success "Deployment verification completed"
}

# Main deployment function
main() {
    log_info "Starting Cloudflare deployment for Go Coffee Platform"
    log_info "Environment: $ENVIRONMENT"
    log_info "Dry Run: $DRY_RUN"
    
    # Execute deployment steps
    check_prerequisites
    create_queues
    deploy_workers
    deploy_pages
    setup_secrets
    configure_domains
    verify_deployment
    
    log_success "🎉 Go Coffee Platform deployed successfully to Cloudflare!"
    log_info "Next steps:"
    log_info "1. Configure custom domains in Cloudflare dashboard"
    log_info "2. Set up DNS records for your domain"
    log_info "3. Configure SSL certificates"
    log_info "4. Test all endpoints and functionality"
    log_info "5. Set up monitoring and alerts"
}

# Handle script arguments
case "${1:-}" in
    --dry-run)
        DRY_RUN=true
        main
        ;;
    --environment)
        ENVIRONMENT="${2:-production}"
        main
        ;;
    --help)
        echo "Usage: $0 [--dry-run] [--environment <env>] [--help]"
        echo ""
        echo "Options:"
        echo "  --dry-run              Run in dry-run mode (no actual deployment)"
        echo "  --environment <env>    Deploy to specific environment (default: production)"
        echo "  --help                 Show this help message"
        echo ""
        echo "Environments: development, staging, production"
        echo ""
        echo "Examples:"
        echo "  $0                           # Deploy to production"
        echo "  $0 --environment staging     # Deploy to staging"
        echo "  $0 --dry-run                 # Dry run for production"
        ;;
    *)
        main
        ;;
esac
