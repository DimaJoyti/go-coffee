#!/bin/bash

# Go Coffee Platform - Cloudflare Environment Configuration
# This script sets up environment-specific configurations

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Cloudflare Account Configuration
export CLOUDFLARE_ACCOUNT_ID="6244f6d02d9c7684386c1c849bdeaf56"

# KV Namespace IDs (created via MCP tools)
export KV_CACHE_ID="f12294ae42924b729024f030e9b5611c"
export KV_SESSIONS_ID="be1b38800d8d4f00820a04d1e5866552"
export KV_ORDERS_ID="6a65944658824e6b8bbc9f9e24e10317"

# R2 Bucket Names (created via MCP tools)
export R2_ASSETS_BUCKET="go-coffee-assets"
export R2_IMAGES_BUCKET="go-coffee-images"
export R2_BACKUPS_BUCKET="go-coffee-backups"

# Environment-specific configuration
setup_environment() {
    local env="${1:-production}"
    
    log_info "Setting up environment: $env"
    
    case "$env" in
        "production")
            setup_production
            ;;
        "staging")
            setup_staging
            ;;
        "development")
            setup_development
            ;;
        *)
            echo "Unknown environment: $env"
            echo "Available environments: production, staging, development"
            exit 1
            ;;
    esac
    
    log_success "Environment $env configured"
}

setup_production() {
    export ENVIRONMENT="production"
    export LOG_LEVEL="warn"
    
    # API URLs
    export NEXT_PUBLIC_API_URL="https://api.go-coffee.com"
    export NEXT_PUBLIC_WS_URL="wss://events.go-coffee.com"
    export NEXT_PUBLIC_CDN_URL="https://cdn.go-coffee.com"
    
    # Domains
    export FRONTEND_DOMAIN="go-coffee.com"
    export API_DOMAIN="api.go-coffee.com"
    export EVENTS_DOMAIN="events.go-coffee.com"
    export AI_DOMAIN="ai.go-coffee.com"
    
    # Feature flags
    export ORDER_PROCESSING_ENABLED="true"
    export PAYMENT_PROCESSING_ENABLED="true"
    export INVENTORY_TRACKING_ENABLED="true"
    export NOTIFICATION_ENABLED="true"
    export AI_COORDINATION_ENABLED="true"
    export ENABLE_AWS_ROUTING="true"
    export ENABLE_GCP_ROUTING="true"
    export ENABLE_AZURE_ROUTING="true"
    
    # Rate limiting
    export RATE_LIMIT_REQUESTS="1000"
    export RATE_LIMIT_WINDOW="60"
    
    # Security
    export CORS_ORIGINS="https://go-coffee.com,https://app.go-coffee.com,https://admin.go-coffee.com"
    
    log_info "Production environment configured"
}

setup_staging() {
    export ENVIRONMENT="staging"
    export LOG_LEVEL="info"
    
    # API URLs
    export NEXT_PUBLIC_API_URL="https://api-staging.go-coffee.com"
    export NEXT_PUBLIC_WS_URL="wss://events-staging.go-coffee.com"
    export NEXT_PUBLIC_CDN_URL="https://cdn-staging.go-coffee.com"
    
    # Domains
    export FRONTEND_DOMAIN="staging.go-coffee.com"
    export API_DOMAIN="api-staging.go-coffee.com"
    export EVENTS_DOMAIN="events-staging.go-coffee.com"
    export AI_DOMAIN="ai-staging.go-coffee.com"
    
    # Feature flags
    export ORDER_PROCESSING_ENABLED="true"
    export PAYMENT_PROCESSING_ENABLED="true"
    export INVENTORY_TRACKING_ENABLED="true"
    export NOTIFICATION_ENABLED="false"
    export AI_COORDINATION_ENABLED="true"
    export ENABLE_AWS_ROUTING="true"
    export ENABLE_GCP_ROUTING="false"
    export ENABLE_AZURE_ROUTING="false"
    
    # Rate limiting (more permissive for testing)
    export RATE_LIMIT_REQUESTS="2000"
    export RATE_LIMIT_WINDOW="60"
    
    # Security
    export CORS_ORIGINS="https://staging.go-coffee.com,https://app-staging.go-coffee.com"
    
    log_info "Staging environment configured"
}

setup_development() {
    export ENVIRONMENT="development"
    export LOG_LEVEL="debug"
    
    # API URLs
    export NEXT_PUBLIC_API_URL="https://api-dev.go-coffee.com"
    export NEXT_PUBLIC_WS_URL="wss://events-dev.go-coffee.com"
    export NEXT_PUBLIC_CDN_URL="https://cdn-dev.go-coffee.com"
    
    # Domains
    export FRONTEND_DOMAIN="dev.go-coffee.com"
    export API_DOMAIN="api-dev.go-coffee.com"
    export EVENTS_DOMAIN="events-dev.go-coffee.com"
    export AI_DOMAIN="ai-dev.go-coffee.com"
    
    # Feature flags (limited for development)
    export ORDER_PROCESSING_ENABLED="true"
    export PAYMENT_PROCESSING_ENABLED="false"
    export INVENTORY_TRACKING_ENABLED="false"
    export NOTIFICATION_ENABLED="false"
    export AI_COORDINATION_ENABLED="true"
    export ENABLE_AWS_ROUTING="false"
    export ENABLE_GCP_ROUTING="false"
    export ENABLE_AZURE_ROUTING="false"
    
    # Rate limiting (very permissive for development)
    export RATE_LIMIT_REQUESTS="10000"
    export RATE_LIMIT_WINDOW="60"
    
    # Security (permissive for development)
    export CORS_ORIGINS="*"
    
    log_info "Development environment configured"
}

# Secrets configuration (to be set manually)
setup_secrets_template() {
    log_info "Setting up secrets template..."
    
    cat > .env.secrets.template << EOF
# Go Coffee Platform - Secrets Template
# Copy this file to .env.secrets and fill in the actual values

# Stripe Configuration
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Database Configuration
DATABASE_URL=postgresql://user:password@host:port/database

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Twilio Configuration (SMS notifications)
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=your-twilio-auth-token

# SendGrid Configuration (Email notifications)
SENDGRID_API_KEY=SG.your-sendgrid-api-key

# Google Analytics
NEXT_PUBLIC_GOOGLE_ANALYTICS_ID=G-...

# Sentry (Error tracking)
NEXT_PUBLIC_SENTRY_DSN=https://...

# Additional API Keys
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
EOF
    
    log_success "Secrets template created at .env.secrets.template"
    log_info "Please copy to .env.secrets and fill in actual values"
}

# Load secrets from file
load_secrets() {
    local secrets_file="${1:-.env.secrets}"
    
    if [[ -f "$secrets_file" ]]; then
        log_info "Loading secrets from $secrets_file"
        set -a  # automatically export all variables
        source "$secrets_file"
        set +a
        log_success "Secrets loaded"
    else
        log_info "Secrets file $secrets_file not found"
        log_info "Run: setup_secrets_template to create template"
    fi
}

# Export configuration for use in other scripts
export_config() {
    log_info "Exporting configuration..."
    
    # Create environment file
    cat > .env.cloudflare << EOF
# Go Coffee Platform - Cloudflare Configuration
# Generated on $(date)

CLOUDFLARE_ACCOUNT_ID=$CLOUDFLARE_ACCOUNT_ID
ENVIRONMENT=$ENVIRONMENT
LOG_LEVEL=$LOG_LEVEL

# KV Namespaces
KV_CACHE_ID=$KV_CACHE_ID
KV_SESSIONS_ID=$KV_SESSIONS_ID
KV_ORDERS_ID=$KV_ORDERS_ID

# R2 Buckets
R2_ASSETS_BUCKET=$R2_ASSETS_BUCKET
R2_IMAGES_BUCKET=$R2_IMAGES_BUCKET
R2_BACKUPS_BUCKET=$R2_BACKUPS_BUCKET

# API URLs
NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL
NEXT_PUBLIC_CDN_URL=$NEXT_PUBLIC_CDN_URL

# Domains
FRONTEND_DOMAIN=$FRONTEND_DOMAIN
API_DOMAIN=$API_DOMAIN
EVENTS_DOMAIN=$EVENTS_DOMAIN
AI_DOMAIN=$AI_DOMAIN

# Feature Flags
ORDER_PROCESSING_ENABLED=$ORDER_PROCESSING_ENABLED
PAYMENT_PROCESSING_ENABLED=$PAYMENT_PROCESSING_ENABLED
INVENTORY_TRACKING_ENABLED=$INVENTORY_TRACKING_ENABLED
NOTIFICATION_ENABLED=$NOTIFICATION_ENABLED
AI_COORDINATION_ENABLED=$AI_COORDINATION_ENABLED
ENABLE_AWS_ROUTING=$ENABLE_AWS_ROUTING
ENABLE_GCP_ROUTING=$ENABLE_GCP_ROUTING
ENABLE_AZURE_ROUTING=$ENABLE_AZURE_ROUTING

# Rate Limiting
RATE_LIMIT_REQUESTS=$RATE_LIMIT_REQUESTS
RATE_LIMIT_WINDOW=$RATE_LIMIT_WINDOW

# Security
CORS_ORIGINS=$CORS_ORIGINS
EOF
    
    log_success "Configuration exported to .env.cloudflare"
}

# Main function
main() {
    local command="${1:-setup}"
    local environment="${2:-production}"
    
    case "$command" in
        "setup")
            setup_environment "$environment"
            export_config
            ;;
        "secrets")
            setup_secrets_template
            ;;
        "load")
            load_secrets "${2:-.env.secrets}"
            ;;
        "export")
            export_config
            ;;
        *)
            echo "Usage: $0 [setup|secrets|load|export] [environment]"
            echo ""
            echo "Commands:"
            echo "  setup <env>    Set up environment configuration"
            echo "  secrets        Create secrets template"
            echo "  load <file>    Load secrets from file"
            echo "  export         Export current configuration"
            echo ""
            echo "Environments: production, staging, development"
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
