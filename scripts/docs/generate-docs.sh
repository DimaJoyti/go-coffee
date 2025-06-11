#!/bin/bash

# Go Coffee - Automated Documentation Generation
# Generates comprehensive documentation for APIs, services, and architecture
# Version: 3.0.0
# Usage: ./generate-docs.sh [OPTIONS]
#   -t, --type TYPE     Documentation type (api|arch|deploy|all)
#   -f, --format FMT    Output format (html|markdown|pdf|all)
#   -o, --output DIR    Output directory
#   -s, --serve         Start documentation server
#   -p, --port PORT     Server port (default: 8000)
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

print_header "📚 Go Coffee Automated Documentation Generation"

# =============================================================================
# CONFIGURATION
# =============================================================================

DOC_TYPE="all"
OUTPUT_FORMAT="html"
OUTPUT_DIR="docs/generated"
SERVE_DOCS=false
SERVER_PORT=8000

# Documentation tools
declare -A DOC_TOOLS=(
    ["swagger"]="API documentation generator"
    ["godoc"]="Go documentation generator"
    ["mermaid"]="Diagram generator"
    ["pandoc"]="Document converter"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_docs_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type)
                DOC_TYPE="$2"
                shift 2
                ;;
            -f|--format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -s|--serve)
                SERVE_DOCS=true
                shift
                ;;
            -p|--port)
                SERVER_PORT="$2"
                shift 2
                ;;
            -h|--help)
                show_usage "generate-docs.sh" \
                    "Automated documentation generation for Go Coffee platform" \
                    "  ./generate-docs.sh [OPTIONS]
  
  Options:
    -t, --type TYPE     Documentation type (api|arch|deploy|all) (default: all)
    -f, --format FMT    Output format (html|markdown|pdf|all) (default: html)
    -o, --output DIR    Output directory (default: docs/generated)
    -s, --serve         Start documentation server after generation
    -p, --port PORT     Server port (default: 8000)
    -h, --help          Show this help message
  
  Documentation Types:
    api:    API documentation (OpenAPI/Swagger)
    arch:   Architecture documentation
    deploy: Deployment documentation
    all:    All documentation types
  
  Examples:
    ./generate-docs.sh                      # Generate all docs in HTML
    ./generate-docs.sh -t api -f markdown  # API docs in Markdown
    ./generate-docs.sh -s -p 9000          # Generate and serve on port 9000
    ./generate-docs.sh -f all               # Generate in all formats"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# DOCUMENTATION GENERATION FUNCTIONS
# =============================================================================

# Check documentation dependencies
check_docs_dependencies() {
    print_header "🔍 Checking Documentation Dependencies"
    
    local required_tools=("go" "curl" "jq")
    local optional_tools=("swagger" "pandoc" "mermaid-cli" "python3")
    
    # Check required tools
    check_dependencies "${required_tools[@]}" || exit 1
    
    # Check optional tools
    for tool in "${optional_tools[@]}"; do
        if command_exists "$tool"; then
            print_status "$tool is available"
        else
            print_warning "$tool is not available (optional)"
            install_doc_tool "$tool"
        fi
    done
    
    print_success "Documentation dependencies checked"
}

# Install documentation tools
install_doc_tool() {
    local tool=$1
    
    case "$tool" in
        "swagger")
            if command_exists go; then
                go install github.com/swaggo/swag/cmd/swag@latest
                print_status "swagger installed"
            fi
            ;;
        "mermaid-cli")
            if command_exists npm; then
                npm install -g @mermaid-js/mermaid-cli
                print_status "mermaid-cli installed"
            fi
            ;;
    esac
}

# Create documentation structure
create_docs_structure() {
    print_header "📁 Creating Documentation Structure"
    
    local dirs=(
        "$OUTPUT_DIR/api"
        "$OUTPUT_DIR/architecture"
        "$OUTPUT_DIR/deployment"
        "$OUTPUT_DIR/assets/images"
        "$OUTPUT_DIR/assets/diagrams"
        "$OUTPUT_DIR/assets/css"
        "$OUTPUT_DIR/assets/js"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
    done
    
    print_status "Documentation structure created: $OUTPUT_DIR"
}

# Generate API documentation
generate_api_docs() {
    print_header "📖 Generating API Documentation"
    
    # Generate OpenAPI/Swagger documentation
    print_progress "Generating OpenAPI specifications..."
    
    # Find all services with main.go files
    local services=()
    for service_dir in cmd/*/; do
        if [[ -f "$service_dir/main.go" ]]; then
            local service_name=$(basename "$service_dir")
            services+=("$service_name")
        fi
    done
    
    # Generate Swagger docs for each service
    for service in "${services[@]}"; do
        print_progress "Generating API docs for $service..."
        
        # Create service-specific API doc
        cat > "$OUTPUT_DIR/api/${service}-api.md" <<EOF
# $service API Documentation

## Overview
This document describes the REST API for the $service service in the Go Coffee platform.

## Base URL
\`\`\`
http://localhost:8080/api/v1/$service
\`\`\`

## Authentication
All API endpoints require authentication via JWT token in the Authorization header:
\`\`\`
Authorization: Bearer <jwt_token>
\`\`\`

## Endpoints

### Health Check
\`\`\`http
GET /health
\`\`\`

Returns the health status of the service.

**Response:**
\`\`\`json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "service": "$service",
  "version": "1.0.0"
}
\`\`\`

### Service-Specific Endpoints
EOF

        # Add service-specific endpoints based on service type
        case "$service" in
            "auth-service")
                cat >> "$OUTPUT_DIR/api/${service}-api.md" <<EOF

#### Login
\`\`\`http
POST /auth/login
\`\`\`

#### Register
\`\`\`http
POST /auth/register
\`\`\`

#### Refresh Token
\`\`\`http
POST /auth/refresh
\`\`\`
EOF
                ;;
            "order-service")
                cat >> "$OUTPUT_DIR/api/${service}-api.md" <<EOF

#### Create Order
\`\`\`http
POST /orders
\`\`\`

#### Get Orders
\`\`\`http
GET /orders
\`\`\`

#### Get Order by ID
\`\`\`http
GET /orders/{id}
\`\`\`

#### Update Order
\`\`\`http
PUT /orders/{id}
\`\`\`
EOF
                ;;
            "payment-service")
                cat >> "$OUTPUT_DIR/api/${service}-api.md" <<EOF

#### Process Payment
\`\`\`http
POST /payments
\`\`\`

#### Get Payment Status
\`\`\`http
GET /payments/{id}
\`\`\`

#### Refund Payment
\`\`\`http
POST /payments/{id}/refund
\`\`\`
EOF
                ;;
        esac
    done
    
    # Generate combined API documentation
    cat > "$OUTPUT_DIR/api/README.md" <<EOF
# Go Coffee Platform API Documentation

## Overview
The Go Coffee platform provides a comprehensive set of REST APIs for managing coffee orders, payments, kitchen operations, and more.

## Services
EOF

    for service in "${services[@]}"; do
        echo "- [$service](./${service}-api.md)" >> "$OUTPUT_DIR/api/README.md"
    done
    
    print_success "API documentation generated"
}

# Generate architecture documentation
generate_architecture_docs() {
    print_header "🏗️ Generating Architecture Documentation"
    
    # Generate system architecture overview
    cat > "$OUTPUT_DIR/architecture/overview.md" <<EOF
# Go Coffee Platform Architecture

## System Overview
The Go Coffee platform is built using a microservices architecture with the following key components:

## Core Services
- **API Gateway**: Central entry point for all client requests
- **Auth Service**: Authentication and authorization
- **Order Service**: Order management and processing
- **Payment Service**: Payment processing and billing
- **Kitchen Service**: Kitchen operations and order fulfillment
- **User Gateway**: User management and profiles

## AI Services
- **AI Search**: Intelligent search and recommendations
- **AI Service**: Core AI processing capabilities
- **AI Arbitrage Service**: Trading and arbitrage algorithms
- **LLM Orchestrator**: Large language model coordination

## Infrastructure
- **Redis**: Caching and session storage
- **PostgreSQL**: Primary database
- **Message Queue**: Asynchronous communication
- **Monitoring**: Prometheus, Grafana, Jaeger

## Architecture Patterns
- **Clean Architecture**: Separation of concerns and dependency inversion
- **Domain-Driven Design**: Business logic organization
- **Event-Driven Architecture**: Asynchronous communication
- **CQRS**: Command Query Responsibility Segregation

## Security
- **JWT Authentication**: Stateless authentication
- **API Rate Limiting**: Protection against abuse
- **Input Validation**: Data sanitization and validation
- **Encryption**: Data encryption at rest and in transit

## Scalability
- **Horizontal Scaling**: Service replication
- **Load Balancing**: Traffic distribution
- **Caching**: Performance optimization
- **Database Sharding**: Data distribution
EOF

    # Generate deployment architecture
    cat > "$OUTPUT_DIR/architecture/deployment.md" <<EOF
# Deployment Architecture

## Container Strategy
All services are containerized using Docker with multi-stage builds for optimization.

## Orchestration
- **Development**: Docker Compose
- **Production**: Kubernetes

## Networking
- **Service Mesh**: Istio for service-to-service communication
- **Ingress**: NGINX Ingress Controller
- **Load Balancer**: Cloud provider load balancers

## Storage
- **Persistent Volumes**: For database storage
- **Object Storage**: For file uploads and static assets
- **Backup Strategy**: Automated backups and disaster recovery

## Monitoring and Observability
- **Metrics**: Prometheus and Grafana
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Tracing**: Jaeger for distributed tracing
- **Alerting**: AlertManager for notifications
EOF

    print_success "Architecture documentation generated"
}

# Generate deployment documentation
generate_deployment_docs() {
    print_header "🚀 Generating Deployment Documentation"
    
    cat > "$OUTPUT_DIR/deployment/README.md" <<EOF
# Go Coffee Deployment Guide

## Prerequisites
- Docker and Docker Compose
- Kubernetes cluster (for production)
- kubectl configured
- Helm 3.x

## Quick Start

### Development Environment
\`\`\`bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Build all services
./build_all.sh

# Start all services
./scripts/start-all-services.sh --dev-mode
\`\`\`

### Production Deployment
\`\`\`bash
# Deploy to Kubernetes
./scripts/deploy.sh --env production k8s

# Verify deployment
./scripts/health-check.sh --environment production
\`\`\`

## Environment Configuration

### Development
- Local Docker containers
- In-memory databases
- Debug logging enabled
- Hot reload for development

### Staging
- Kubernetes deployment
- Persistent databases
- Production-like configuration
- Performance testing

### Production
- High availability setup
- Persistent storage
- Monitoring and alerting
- Backup and disaster recovery

## Monitoring Setup
\`\`\`bash
# Setup observability stack
./scripts/monitoring/setup-observability.sh -c -a

# Access monitoring dashboards
# Grafana: http://localhost:3000
# Prometheus: http://localhost:9090
# Jaeger: http://localhost:16686
\`\`\`

## Security Considerations
- Regular security scans
- Dependency updates
- Secret management
- Network policies
- RBAC configuration

## Troubleshooting
Common issues and solutions for deployment problems.
EOF

    print_success "Deployment documentation generated"
}

# Generate documentation index
generate_docs_index() {
    print_header "📋 Generating Documentation Index"
    
    cat > "$OUTPUT_DIR/index.html" <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Coffee Platform Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 30px; }
        .section { margin: 20px 0; padding: 20px; border-left: 4px solid #3498db; background: #f8f9fa; }
        .nav { display: flex; gap: 20px; margin: 20px 0; }
        .nav a { padding: 10px 20px; background: #3498db; color: white; text-decoration: none; border-radius: 5px; }
        .nav a:hover { background: #2980b9; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0; }
        .card { padding: 20px; border: 1px solid #ddd; border-radius: 5px; background: white; }
        .card h3 { margin-top: 0; color: #2c3e50; }
    </style>
</head>
<body>
    <div class="header">
        <h1>☕ Go Coffee Platform Documentation</h1>
        <p>Comprehensive documentation for the Go Coffee microservices platform</p>
        <p>Generated on: $(date)</p>
    </div>
    
    <div class="nav">
        <a href="api/">📖 API Documentation</a>
        <a href="architecture/">🏗️ Architecture</a>
        <a href="deployment/">🚀 Deployment</a>
    </div>
    
    <div class="grid">
        <div class="card">
            <h3>📖 API Documentation</h3>
            <p>Complete REST API documentation for all microservices including endpoints, request/response formats, and authentication.</p>
            <a href="api/">View API Docs →</a>
        </div>
        
        <div class="card">
            <h3>🏗️ Architecture</h3>
            <p>System architecture overview, design patterns, and technical decisions behind the Go Coffee platform.</p>
            <a href="architecture/">View Architecture →</a>
        </div>
        
        <div class="card">
            <h3>🚀 Deployment</h3>
            <p>Deployment guides, environment setup, and operational procedures for running Go Coffee in production.</p>
            <a href="deployment/">View Deployment Guide →</a>
        </div>
    </div>
    
    <div class="section">
        <h2>🎯 Quick Links</h2>
        <ul>
            <li><a href="https://github.com/DimaJoyti/go-coffee">GitHub Repository</a></li>
            <li><a href="api/README.md">API Overview</a></li>
            <li><a href="architecture/overview.md">Architecture Overview</a></li>
            <li><a href="deployment/README.md">Deployment Guide</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>🛠️ Development Tools</h2>
        <p>Essential scripts and tools for Go Coffee development:</p>
        <ul>
            <li><strong>Build:</strong> <code>./build_all.sh</code></li>
            <li><strong>Test:</strong> <code>./scripts/test-all-services.sh</code></li>
            <li><strong>Deploy:</strong> <code>./scripts/deploy.sh</code></li>
            <li><strong>Monitor:</strong> <code>./scripts/health-check.sh</code></li>
        </ul>
    </div>
</body>
</html>
EOF

    print_success "Documentation index generated"
}

# Start documentation server
start_docs_server() {
    print_header "🌐 Starting Documentation Server"
    
    cd "$OUTPUT_DIR"
    
    if command_exists python3; then
        print_info "Starting Python HTTP server on port $SERVER_PORT..."
        print_info "Documentation available at: http://localhost:$SERVER_PORT"
        print_info "Press Ctrl+C to stop the server"
        
        python3 -m http.server "$SERVER_PORT"
    elif command_exists python; then
        print_info "Starting Python HTTP server on port $SERVER_PORT..."
        python -m SimpleHTTPServer "$SERVER_PORT"
    else
        print_warning "Python not available. Please serve the documentation manually from: $OUTPUT_DIR"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)
    
    # Parse arguments
    parse_docs_args "$@"
    
    # Check dependencies
    check_docs_dependencies
    
    print_info "Documentation Generation Configuration:"
    print_info "  Type: $DOC_TYPE"
    print_info "  Format: $OUTPUT_FORMAT"
    print_info "  Output: $OUTPUT_DIR"
    print_info "  Serve: $SERVE_DOCS"
    print_info "  Port: $SERVER_PORT"
    
    # Create documentation structure
    create_docs_structure
    
    # Generate documentation based on type
    case "$DOC_TYPE" in
        "api")
            generate_api_docs
            ;;
        "arch")
            generate_architecture_docs
            ;;
        "deploy")
            generate_deployment_docs
            ;;
        "all")
            generate_api_docs
            generate_architecture_docs
            generate_deployment_docs
            ;;
        *)
            print_error "Unknown documentation type: $DOC_TYPE"
            exit 1
            ;;
    esac
    
    # Generate documentation index
    generate_docs_index
    
    # Calculate generation time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    print_success "🎉 Documentation generation completed in ${total_time}s"
    
    print_header "📚 Documentation Summary"
    print_info "Documentation generated in: $OUTPUT_DIR"
    print_info "Main index: $OUTPUT_DIR/index.html"
    
    # Start server if requested
    if [[ "$SERVE_DOCS" == "true" ]]; then
        start_docs_server
    else
        print_info "To serve documentation locally, run:"
        print_info "  cd $OUTPUT_DIR && python3 -m http.server $SERVER_PORT"
    fi
}

# Run main function with all arguments
main "$@"
