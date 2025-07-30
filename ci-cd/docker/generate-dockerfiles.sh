#!/bin/bash

# Generate standardized Dockerfiles for all Go Coffee services
# This script creates optimized, secure Dockerfiles for each service

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
TEMPLATE_FILE="$SCRIPT_DIR/Dockerfile.template"

# Service definitions with their specific configurations
declare -A SERVICES=(
    # Core Business Services
    ["api-gateway"]="8080:gateway"
    ["auth-service"]="8091:auth"
    ["order-service"]="8094:order"
    ["payment-service"]="8093:payment"
    ["kitchen-service"]="8095:kitchen"
    ["user-gateway"]="8096:user"
    ["security-gateway"]="8097:security"
    ["communication-hub"]="8098:communication"
    
    # AI & ML Services
    ["ai-service"]="8100:ai"
    ["ai-search"]="8099:search"
    ["ai-arbitrage-service"]="8101:arbitrage"
    ["ai-order-service"]="8102:ai-order"
    ["llm-orchestrator"]="8106:llm"
    ["llm-orchestrator-simple"]="8107:llm-simple"
    ["mcp-ai-integration"]="8109:mcp-ai"
    
    # Infrastructure Services
    ["market-data-service"]="8103:market"
    ["defi-service"]="8104:defi"
    ["bright-data-hub-service"]="8105:bright-data"
    ["redis-mcp-server"]="8108:redis-mcp"
    ["web-ui-backend"]="3000:web-ui"
)

# Special service configurations
declare -A SPECIAL_CONFIGS=(
    ["web-ui-backend"]="node"
    ["ai-service"]="gpu"
    ["llm-orchestrator"]="gpu"
)

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

# Generate Dockerfile for a Go service
generate_go_dockerfile() {
    local service_name="$1"
    local port="$2"
    local service_type="$3"
    local output_dir="$4"
    
    local dockerfile_content
    dockerfile_content=$(cat "$TEMPLATE_FILE")
    
    # Replace placeholders
    dockerfile_content="${dockerfile_content//\$\{SERVICE_NAME\}/$service_name}"
    dockerfile_content="${dockerfile_content//EXPOSE 8080/EXPOSE $port}"
    
    # Add service-specific configurations
    case "$service_type" in
        "gpu")
            dockerfile_content="${dockerfile_content/FROM alpine:3.19/FROM nvidia\/cuda:12.2-runtime-alpine3.19}"
            dockerfile_content="${dockerfile_content/# Install runtime dependencies/# Install runtime dependencies including CUDA}"
            ;;
        "gateway"|"security")
            # Add nginx for reverse proxy capabilities
            dockerfile_content="${dockerfile_content/ca-certificates/ca-certificates nginx}"
            ;;
    esac
    
    # Write Dockerfile
    echo "$dockerfile_content" > "$output_dir/Dockerfile"
    
    log "Generated Dockerfile for $service_name (port: $port, type: $service_type)"
}

# Generate Dockerfile for Node.js service (web-ui-backend)
generate_node_dockerfile() {
    local service_name="$1"
    local port="$2"
    local output_dir="$3"
    
    cat > "$output_dir/Dockerfile" <<EOF
# Multi-stage build for Node.js service
FROM node:18-alpine AS builder

# Install build dependencies
RUN apk add --no-cache python3 make g++

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Security scanning stage
FROM aquasec/trivy:latest AS security-scan
COPY --from=builder /app /scan
RUN trivy fs --exit-code 1 --no-progress --severity HIGH,CRITICAL /scan

# Final runtime stage
FROM node:18-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy built application
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package*.json ./

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:$port/health || exit 1

# Expose port
EXPOSE $port

# Labels
LABEL maintainer="Go Coffee Team <devops@gocoffee.dev>"
LABEL description="Go Coffee $service_name service"

# Run the service
CMD ["npm", "start"]
EOF
    
    log "Generated Node.js Dockerfile for $service_name (port: $port)"
}

# Generate .dockerignore file
generate_dockerignore() {
    local output_dir="$1"
    
    cat > "$output_dir/.dockerignore" <<EOF
# Git
.git
.gitignore
.gitattributes

# Documentation
*.md
docs/
README*
CHANGELOG*
LICENSE*

# CI/CD
.github/
ci-cd/
.gitlab-ci.yml
Jenkinsfile

# Development
.vscode/
.idea/
*.swp
*.swo
*~

# Testing
*_test.go
test/
tests/
coverage.out
coverage.html

# Build artifacts
bin/
build/
dist/
target/

# Dependencies (will be downloaded in container)
vendor/
node_modules/

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
.tmp/

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment files
.env
.env.local
.env.*.local

# IDE
*.sublime-project
*.sublime-workspace

# Backup files
*.bak
*.backup
*.old

# Database files
*.db
*.sqlite
*.sqlite3

# Certificates
*.pem
*.key
*.crt
*.p12

# Large files
*.tar
*.tar.gz
*.zip
*.rar
EOF
    
    log "Generated .dockerignore file"
}

# Generate docker-compose.yml for development
generate_docker_compose() {
    local output_dir="$1"
    
    cat > "$output_dir/docker-compose.yml" <<EOF
version: '3.8'

services:
  # Infrastructure Services
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: go_coffee
      POSTGRES_USER: go_coffee_user
      POSTGRES_PASSWORD: go_coffee_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U go_coffee_user -d go_coffee"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Core Services
EOF

    # Add each service to docker-compose
    for service in "${!SERVICES[@]}"; do
        IFS=':' read -r port service_type <<< "${SERVICES[$service]}"
        
        cat >> "$output_dir/docker-compose.yml" <<EOF
  $service:
    build:
      context: .
      dockerfile: cmd/$service/Dockerfile
      args:
        SERVICE_NAME: $service
        VERSION: \${VERSION:-dev}
        BUILD_DATE: \${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}
        VCS_REF: \${VCS_REF:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
    ports:
      - "$port:$port"
    environment:
      - DATABASE_URL=postgres://go_coffee_user:go_coffee_password@postgres:5432/go_coffee?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
      - PORT=$port
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:$port/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

EOF
    done

    cat >> "$output_dir/docker-compose.yml" <<EOF

volumes:
  postgres_data:
  redis_data:

networks:
  default:
    name: go-coffee-network
EOF
    
    log "Generated docker-compose.yml for development"
}

# Main execution
main() {
    log "Starting Dockerfile generation for Go Coffee services..."
    
    # Check if template exists
    if [[ ! -f "$TEMPLATE_FILE" ]]; then
        error "Template file not found: $TEMPLATE_FILE"
    fi
    
    # Create output directories and generate Dockerfiles
    for service in "${!SERVICES[@]}"; do
        IFS=':' read -r port service_type <<< "${SERVICES[$service]}"
        
        # Determine output directory
        local output_dir
        if [[ -d "$PROJECT_ROOT/cmd/$service" ]]; then
            output_dir="$PROJECT_ROOT/cmd/$service"
        elif [[ -d "$PROJECT_ROOT/services/$service" ]]; then
            output_dir="$PROJECT_ROOT/services/$service"
        else
            output_dir="$PROJECT_ROOT/deployments/$service"
            mkdir -p "$output_dir"
        fi
        
        info "Processing $service -> $output_dir"
        
        # Generate appropriate Dockerfile
        if [[ "${SPECIAL_CONFIGS[$service]:-}" == "node" ]]; then
            generate_node_dockerfile "$service" "$port" "$output_dir"
        else
            generate_go_dockerfile "$service" "$port" "$service_type" "$output_dir"
        fi
        
        # Generate .dockerignore if it doesn't exist
        if [[ ! -f "$output_dir/.dockerignore" ]]; then
            generate_dockerignore "$output_dir"
        fi
    done
    
    # Generate docker-compose.yml for development
    generate_docker_compose "$PROJECT_ROOT"
    
    # Generate build script
    cat > "$PROJECT_ROOT/build-all-images.sh" <<EOF
#!/bin/bash
# Build all Go Coffee service images

set -e

VERSION=\${VERSION:-\$(git rev-parse --short HEAD)}
BUILD_DATE=\$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VCS_REF=\$(git rev-parse HEAD)

echo "Building all Go Coffee service images..."
echo "Version: \$VERSION"
echo "Build Date: \$BUILD_DATE"
echo "VCS Ref: \$VCS_REF"

EOF

    for service in "${!SERVICES[@]}"; do
        IFS=':' read -r port service_type <<< "${SERVICES[$service]}"
        
        cat >> "$PROJECT_ROOT/build-all-images.sh" <<EOF
echo "Building $service..."
docker build \\
    --build-arg SERVICE_NAME=$service \\
    --build-arg VERSION=\$VERSION \\
    --build-arg BUILD_DATE=\$BUILD_DATE \\
    --build-arg VCS_REF=\$VCS_REF \\
    -f cmd/$service/Dockerfile \\
    -t ghcr.io/dimajoyti/go-coffee/$service:\$VERSION \\
    -t ghcr.io/dimajoyti/go-coffee/$service:latest \\
    .

EOF
    done
    
    cat >> "$PROJECT_ROOT/build-all-images.sh" <<EOF

echo "All images built successfully!"
echo "To push images:"
echo "  docker push ghcr.io/dimajoyti/go-coffee/SERVICE_NAME:\$VERSION"
EOF
    
    chmod +x "$PROJECT_ROOT/build-all-images.sh"
    
    log "Generated build-all-images.sh script"
    log "Dockerfile generation completed successfully!"
    log "Generated Dockerfiles for ${#SERVICES[@]} services"
    
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Review generated Dockerfiles in cmd/*/Dockerfile"
    echo "2. Test builds with: ./build-all-images.sh"
    echo "3. Test locally with: docker-compose up"
    echo "4. Commit changes and push to trigger CI/CD"
}

# Handle command line arguments
case "${1:-generate}" in
    "generate")
        main
        ;;
    "clean")
        log "Cleaning generated Dockerfiles..."
        find "$PROJECT_ROOT" -name "Dockerfile" -path "*/cmd/*" -delete
        find "$PROJECT_ROOT" -name ".dockerignore" -path "*/cmd/*" -delete
        rm -f "$PROJECT_ROOT/docker-compose.yml"
        rm -f "$PROJECT_ROOT/build-all-images.sh"
        log "Cleanup completed"
        ;;
    *)
        echo "Usage: $0 [generate|clean]"
        exit 1
        ;;
esac
