#!/bin/bash

# Go Coffee Security Platform Setup Script
# This script sets up the complete security platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PLATFORM_NAME="Go Coffee Security Platform"
VERSION="1.0.0"

echo -e "${BLUE}🛡️  $PLATFORM_NAME Setup${NC}"
echo -e "${CYAN}Version: $VERSION${NC}"
echo "=================================================="

# Function to print section headers
print_section() {
    echo -e "\n${YELLOW}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to print success messages
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print error messages
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to print info messages
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_section "Checking Prerequisites"
    
    local missing_deps=()
    
    # Check Go
    if command_exists go; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go $GO_VERSION installed"
    else
        missing_deps+=("go")
        print_error "Go is not installed"
    fi
    
    # Check Docker
    if command_exists docker; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        print_success "Docker $DOCKER_VERSION installed"
    else
        missing_deps+=("docker")
        print_error "Docker is not installed"
    fi
    
    # Check Docker Compose
    if command_exists docker-compose; then
        COMPOSE_VERSION=$(docker-compose --version | awk '{print $3}' | sed 's/,//')
        print_success "Docker Compose $COMPOSE_VERSION installed"
    else
        missing_deps+=("docker-compose")
        print_error "Docker Compose is not installed"
    fi
    
    # Check Make
    if command_exists make; then
        print_success "Make is installed"
    else
        missing_deps+=("make")
        print_error "Make is not installed"
    fi
    
    # Check OpenSSL
    if command_exists openssl; then
        OPENSSL_VERSION=$(openssl version | awk '{print $2}')
        print_success "OpenSSL $OPENSSL_VERSION installed"
    else
        missing_deps+=("openssl")
        print_error "OpenSSL is not installed"
    fi
    
    # Check curl
    if command_exists curl; then
        print_success "curl is installed"
    else
        missing_deps+=("curl")
        print_error "curl is not installed"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo ""
        echo "Please install the missing dependencies and run this script again."
        echo ""
        echo "Installation guides:"
        echo "• Go: https://golang.org/doc/install"
        echo "• Docker: https://docs.docker.com/get-docker/"
        echo "• Docker Compose: https://docs.docker.com/compose/install/"
        echo "• Make: Usually available in build-essential package"
        echo "• OpenSSL: Usually pre-installed on most systems"
        echo "• curl: Usually pre-installed on most systems"
        exit 1
    fi
    
    print_success "All prerequisites are satisfied!"
}

# Function to setup project structure
setup_project_structure() {
    print_section "Setting Up Project Structure"
    
    # Create necessary directories
    local dirs=(
        "keys"
        "logs"
        "data"
        "backups"
        "monitoring/prometheus"
        "monitoring/grafana/provisioning/dashboards"
        "monitoring/grafana/provisioning/datasources"
        "monitoring/alertmanager"
        "monitoring/logstash/pipeline"
        "monitoring/logstash/config"
        "nginx"
        "test/mocks/auth-service"
        "test/mocks/order-service"
        "test/mocks/payment-service"
        "test/mocks/user-service"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
        print_success "Created directory: $dir"
    done
}

# Function to generate security keys
generate_security_keys() {
    print_section "Generating Security Keys"
    
    if [ -f "scripts/generate-security-keys.sh" ]; then
        print_info "Running security keys generation script..."
        bash scripts/generate-security-keys.sh
        print_success "Security keys generated successfully!"
    else
        print_error "Security keys generation script not found!"
        exit 1
    fi
}

# Function to setup environment files
setup_environment_files() {
    print_section "Setting Up Environment Files"
    
    # Copy example environment files
    if [ -f ".env.security.example" ]; then
        if [ ! -f ".env" ]; then
            cp .env.security.example .env
            print_success "Created .env file from template"
        else
            print_info ".env file already exists, skipping..."
        fi
    fi
    
    # Update .env with generated keys if available
    if [ -f "keys/.env.generated" ]; then
        print_info "Updating .env with generated security keys..."
        
        # Extract keys from generated file
        source keys/.env.generated
        
        # Update .env file with generated keys
        sed -i.bak \
            -e "s|AES_KEY=.*|AES_KEY=$AES_KEY|" \
            -e "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" \
            -e "s|REDIS_PASSWORD=.*|REDIS_PASSWORD=$REDIS_PASSWORD|" \
            .env
        
        print_success "Environment file updated with generated keys"
    fi
}

# Function to setup monitoring configuration
setup_monitoring_config() {
    print_section "Setting Up Monitoring Configuration"
    
    # Prometheus configuration
    cat > monitoring/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'security-gateway'
    static_configs:
      - targets: ['security-gateway:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:8081']
    metrics_path: '/metrics'
    scrape_interval: 15s

  - job_name: 'order-service'
    static_configs:
      - targets: ['order-service:8082']
    metrics_path: '/metrics'
    scrape_interval: 15s

  - job_name: 'payment-service'
    static_configs:
      - targets: ['payment-service:8083']
    metrics_path: '/metrics'
    scrape_interval: 15s

  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:8084']
    metrics_path: '/metrics'
    scrape_interval: 15s
EOF
    print_success "Prometheus configuration created"
    
    # AlertManager configuration
    cat > monitoring/alertmanager/alertmanager.yml << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'security@go-coffee.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://localhost:5001/'
    send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF
    print_success "AlertManager configuration created"
    
    # Grafana datasource configuration
    cat > monitoring/grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF
    print_success "Grafana datasource configuration created"
}

# Function to build services
build_services() {
    print_section "Building Services"
    
    # Build Security Gateway
    print_info "Building Security Gateway..."
    if make -f Makefile.security-gateway build; then
        print_success "Security Gateway built successfully"
    else
        print_error "Failed to build Security Gateway"
        exit 1
    fi
    
    # Build other services if Makefiles exist
    local services=("auth" "order" "payment" "user")
    for service in "${services[@]}"; do
        if [ -f "Makefile.$service" ]; then
            print_info "Building $service service..."
            if make -f "Makefile.$service" build; then
                print_success "$service service built successfully"
            else
                print_error "Failed to build $service service"
            fi
        fi
    done
}

# Function to setup Docker environment
setup_docker_environment() {
    print_section "Setting Up Docker Environment"
    
    # Build Docker images
    print_info "Building Docker images..."
    if docker-compose -f docker-compose.security-gateway.yml build; then
        print_success "Docker images built successfully"
    else
        print_error "Failed to build Docker images"
        exit 1
    fi
}

# Function to run tests
run_tests() {
    print_section "Running Tests"
    
    # Run Security Gateway tests
    print_info "Running Security Gateway tests..."
    if make -f Makefile.security-gateway test; then
        print_success "Security Gateway tests passed"
    else
        print_error "Security Gateway tests failed"
    fi
    
    # Run security validation
    if [ -f "keys/validate-keys.sh" ]; then
        print_info "Validating security keys..."
        cd keys && ./validate-keys.sh && cd ..
        print_success "Security keys validation passed"
    fi
}

# Function to start services
start_services() {
    print_section "Starting Services"
    
    print_info "Starting all services with Docker Compose..."
    if docker-compose -f docker-compose.security-gateway.yml up -d; then
        print_success "All services started successfully"
        
        # Wait for services to be ready
        print_info "Waiting for services to be ready..."
        sleep 30
        
        # Check service health
        print_info "Checking service health..."
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            print_success "Security Gateway is healthy"
        else
            print_error "Security Gateway health check failed"
        fi
        
    else
        print_error "Failed to start services"
        exit 1
    fi
}

# Function to run security demo
run_security_demo() {
    print_section "Running Security Demo"
    
    if [ -f "test/security-demo.sh" ]; then
        print_info "Running security demonstration..."
        bash test/security-demo.sh
        print_success "Security demo completed"
    else
        print_error "Security demo script not found"
    fi
}

# Function to display final information
display_final_info() {
    print_section "Setup Complete!"
    
    echo -e "${GREEN}🎉 Go Coffee Security Platform is now ready!${NC}"
    echo ""
    echo -e "${CYAN}📊 Access Points:${NC}"
    echo "• Security Gateway:  http://localhost:8080"
    echo "• Grafana Dashboard: http://localhost:3000 (admin/admin)"
    echo "• Prometheus:        http://localhost:9090"
    echo "• Jaeger Tracing:    http://localhost:16686"
    echo "• Kibana:           http://localhost:5601"
    echo ""
    echo -e "${CYAN}🔧 Useful Commands:${NC}"
    echo "• Check service status: docker-compose -f docker-compose.security-gateway.yml ps"
    echo "• View logs:           docker-compose -f docker-compose.security-gateway.yml logs -f"
    echo "• Stop services:       docker-compose -f docker-compose.security-gateway.yml down"
    echo "• Run security demo:   ./test/security-demo.sh"
    echo "• Rotate keys:         cd keys && ./rotate-keys.sh"
    echo ""
    echo -e "${CYAN}📚 Documentation:${NC}"
    echo "• Security Architecture: docs/SECURITY-ARCHITECTURE.md"
    echo "• Security Gateway:      cmd/security-gateway/README.md"
    echo "• Environment Config:    .env.security.example"
    echo ""
    echo -e "${YELLOW}⚠️  Security Reminders:${NC}"
    echo "• Keep your .env file secure and never commit it to version control"
    echo "• Rotate security keys regularly (every 90 days minimum)"
    echo "• Monitor security alerts and logs regularly"
    echo "• Update dependencies and security patches regularly"
    echo ""
    echo -e "${GREEN}🛡️ Your microservices are now protected with enterprise-grade security!${NC}"
}

# Main setup function
main() {
    echo -e "${PURPLE}Starting Go Coffee Security Platform setup...${NC}"
    echo ""
    
    # Run setup steps
    check_prerequisites
    setup_project_structure
    generate_security_keys
    setup_environment_files
    setup_monitoring_config
    build_services
    setup_docker_environment
    run_tests
    start_services
    run_security_demo
    display_final_info
    
    echo ""
    echo -e "${GREEN}🚀 Setup completed successfully!${NC}"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Go Coffee Security Platform Setup Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --keys-only    Generate security keys only"
        echo "  --build-only   Build services only"
        echo "  --start-only   Start services only"
        echo "  --demo-only    Run security demo only"
        echo ""
        exit 0
        ;;
    --keys-only)
        generate_security_keys
        ;;
    --build-only)
        build_services
        ;;
    --start-only)
        start_services
        ;;
    --demo-only)
        run_security_demo
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
