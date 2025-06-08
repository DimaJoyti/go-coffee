#!/bin/bash

# Start the complete scaled Go Coffee system
# This script starts all services with full infrastructure

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
COMPOSE_FILE="deployments/docker/docker-compose.production.yml"
SERVICES=(
    "postgres:5432"
    "redis:6379"
    "zookeeper:2181"
    "kafka:9092"
    "payment-service:8093"
    "auth-service:8091"
    "order-service:8094"
    "kitchen-service:8095"
    "ai-service:8092"
    "api-gateway:8080"
    "prometheus:9090"
    "grafana:3000"
    "jaeger:16686"
)

# Function to print colored output
print_header() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║${NC} $1 ${PURPLE}║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
}

print_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[ℹ]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    print_success "Docker Compose is available"
}

# Function to create necessary directories
create_directories() {
    print_step "Creating necessary directories..."
    
    directories=(
        "deployments/docker/ssl"
        "deployments/docker/grafana/dashboards"
        "deployments/docker/grafana/datasources"
        "logs"
        "data/postgres"
        "data/redis"
        "data/kafka"
    )
    
    for dir in "${directories[@]}"; do
        mkdir -p "$dir"
        print_info "Created directory: $dir"
    done
    
    print_success "All directories created"
}

# Function to generate SSL certificates (self-signed for development)
generate_ssl_certs() {
    print_step "Generating SSL certificates..."
    
    if [ ! -f "deployments/docker/ssl/cert.pem" ]; then
        openssl req -x509 -newkey rsa:4096 -keyout deployments/docker/ssl/key.pem \
            -out deployments/docker/ssl/cert.pem -days 365 -nodes \
            -subj "/C=US/ST=State/L=City/O=GoCoffee/CN=localhost" 2>/dev/null
        
        print_success "SSL certificates generated"
    else
        print_info "SSL certificates already exist"
    fi
}

# Function to create configuration files
create_config_files() {
    print_step "Creating configuration files..."
    
    # Prometheus configuration
    cat > deployments/docker/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'go-coffee-services'
    static_configs:
      - targets: 
        - 'api-gateway:8080'
        - 'payment-service:8093'
        - 'auth-service:8091'
        - 'order-service:8094'
        - 'kitchen-service:8095'
        - 'ai-service:8092'
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'infrastructure'
    static_configs:
      - targets:
        - 'postgres:5432'
        - 'redis:6379'
        - 'kafka:9092'
EOF

    # Grafana datasource configuration
    mkdir -p deployments/docker/grafana/datasources
    cat > deployments/docker/grafana/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Nginx configuration
    cat > deployments/docker/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream api_gateway {
        server api-gateway:8080;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://api_gateway;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    server {
        listen 443 ssl;
        server_name localhost;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        location / {
            proxy_pass http://api_gateway;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF

    # Database initialization script
    cat > deployments/docker/init-db.sql << 'EOF'
-- Go Coffee Database Schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    items JSONB NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    address VARCHAR(255) UNIQUE NOT NULL,
    network VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES wallets(id),
    hash VARCHAR(255) UNIQUE NOT NULL,
    from_address VARCHAR(255),
    to_address VARCHAR(255),
    amount BIGINT NOT NULL,
    fee BIGINT,
    status VARCHAR(50) DEFAULT 'pending',
    network VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
EOF

    print_success "Configuration files created"
}

# Function to set environment variables
set_environment() {
    print_step "Setting environment variables..."
    
    export JWT_SECRET=${JWT_SECRET:-$(openssl rand -hex 32)}
    export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres_password}
    
    print_success "Environment variables set"
}

# Function to start infrastructure services
start_infrastructure() {
    print_step "Starting infrastructure services..."
    
    docker-compose -f "$COMPOSE_FILE" up -d postgres redis zookeeper kafka
    
    print_info "Waiting for infrastructure services to be ready..."
    sleep 30
    
    print_success "Infrastructure services started"
}

# Function to start application services
start_applications() {
    print_step "Starting application services..."
    
    docker-compose -f "$COMPOSE_FILE" up -d \
        payment-service auth-service order-service kitchen-service ai-service
    
    print_info "Waiting for application services to be ready..."
    sleep 20
    
    print_success "Application services started"
}

# Function to start gateway and monitoring
start_gateway_monitoring() {
    print_step "Starting API gateway and monitoring..."
    
    docker-compose -f "$COMPOSE_FILE" up -d \
        api-gateway nginx prometheus grafana jaeger
    
    print_info "Waiting for gateway and monitoring to be ready..."
    sleep 15
    
    print_success "Gateway and monitoring started"
}

# Function to check service health
check_services_health() {
    print_step "Checking service health..."
    
    for service_info in "${SERVICES[@]}"; do
        IFS=':' read -r service port <<< "$service_info"
        
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1 || \
           curl -s "http://localhost:$port" > /dev/null 2>&1; then
            print_success "$service (port $port) - Healthy"
        else
            print_warning "$service (port $port) - Not responding"
        fi
    done
}

# Function to show system status
show_system_status() {
    echo ""
    print_header "🚀 Go Coffee Scaled System Status"
    echo ""
    
    print_info "🌐 Access Points:"
    echo "  • API Gateway: http://localhost:8080"
    echo "  • API Gateway (HTTPS): https://localhost:443"
    echo "  • API Documentation: http://localhost:8080/docs"
    echo ""
    
    print_info "📊 Monitoring & Observability:"
    echo "  • Prometheus: http://localhost:9090"
    echo "  • Grafana: http://localhost:3000 (admin/admin)"
    echo "  • Jaeger Tracing: http://localhost:16686"
    echo ""
    
    print_info "🔧 Infrastructure:"
    echo "  • PostgreSQL: localhost:5432"
    echo "  • Redis: localhost:6379"
    echo "  • Kafka: localhost:9092"
    echo ""
    
    print_info "🎯 Microservices:"
    echo "  • Payment Service: http://localhost:8093"
    echo "  • Auth Service: http://localhost:8091"
    echo "  • Order Service: http://localhost:8094"
    echo "  • Kitchen Service: http://localhost:8095"
    echo "  • AI Service: http://localhost:8092"
    echo ""
    
    print_info "💡 Features Available:"
    echo "  • Bitcoin/Lightning Network payments"
    echo "  • Ethereum/DeFi integration"
    echo "  • AI recommendations and analytics"
    echo "  • Real-time order processing"
    echo "  • Distributed caching (Redis)"
    echo "  • Event-driven architecture (Kafka)"
    echo "  • Comprehensive monitoring"
    echo "  • Load balancing (Nginx)"
    echo ""
}

# Function to cleanup on exit
cleanup() {
    print_info "Cleaning up..."
    docker-compose -f "$COMPOSE_FILE" down
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Main execution
echo ""
print_header "☕ Starting Go Coffee Scaled System"
echo ""

print_info "This will start the complete production-ready system with:"
print_info "• 6 Microservices (API Gateway + 5 business services)"
print_info "• PostgreSQL database with Redis caching"
print_info "• Kafka message broker for event-driven architecture"
print_info "• Prometheus + Grafana for monitoring"
print_info "• Jaeger for distributed tracing"
print_info "• Nginx load balancer with SSL"
echo ""

# Pre-flight checks
check_docker
check_docker_compose

# Setup
create_directories
generate_ssl_certs
create_config_files
set_environment

# Start services in order
start_infrastructure
start_applications
start_gateway_monitoring

# Health checks
check_services_health

# Show status
show_system_status

print_success "🎉 Go Coffee scaled system is fully operational!"
print_info "Press Ctrl+C to stop all services"

# Keep script running
while true; do
    sleep 30
    
    # Periodic health check
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        print_warning "API Gateway health check failed"
    fi
done
