#!/bin/bash

# Go Coffee Core Services Startup Script
# This script starts all core coffee services with proper dependency management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.core.yml"
PROJECT_NAME="go-coffee-core"

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

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose > /dev/null 2>&1; then
        print_error "Docker Compose is not installed. Please install Docker Compose and try again."
        exit 1
    fi
    print_success "Docker Compose is available"
}

# Function to build services
build_services() {
    print_header "Building Go Coffee Services"

    print_status "Building producer service..."
    cd producer && go build -o ../bin/producer ./main.go && cd ..
    print_success "Producer service built"

    print_status "Building consumer service..."
    cd consumer && go build -o ../bin/consumer ./main.go && cd ..
    print_success "Consumer service built"

    print_status "Building streams service..."
    cd streams && go build -o ../bin/streams ./main.go && cd ..
    print_success "Streams service built"

    print_status "Building Web3 payment service..."
    go build -o bin/web3-payment-service ./cmd/web3-payment-service/main.go
    print_success "Web3 payment service built"

    print_status "Building AI orchestrator service..."
    go build -o bin/ai-orchestrator ./cmd/ai-orchestrator/main.go
    print_success "AI orchestrator service built"
}

# Function to start infrastructure services
start_infrastructure() {
    print_header "Starting Infrastructure Services"
    
    print_status "Starting Zookeeper..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d zookeeper
    
    print_status "Waiting for Zookeeper to be ready..."
    sleep 10
    
    print_status "Starting Kafka..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d kafka
    
    print_status "Waiting for Kafka to be ready..."
    sleep 20
    
    print_status "Starting PostgreSQL..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d postgres
    
    print_status "Starting Redis..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d redis
    
    print_status "Waiting for databases to be ready..."
    sleep 15
    
    print_success "Infrastructure services started"
}

# Function to create Kafka topics
create_kafka_topics() {
    print_header "Creating Kafka Topics"
    
    # Wait for Kafka to be fully ready
    print_status "Waiting for Kafka to be fully ready..."
    sleep 10
    
    # Create topics
    docker exec kafka kafka-topics --create --topic coffee_orders --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 --if-not-exists
    docker exec kafka kafka-topics --create --topic processed_orders --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 --if-not-exists
    
    print_success "Kafka topics created"
}

# Function to start application services
start_applications() {
    print_header "Starting Application Services"

    print_status "Starting Producer service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d producer

    print_status "Starting Consumer service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d consumer

    print_status "Starting Streams service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d streams

    print_status "Starting Web3 Payment service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d web3-payment

    print_status "Starting AI Orchestrator service..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d ai-orchestrator

    print_success "Application services started"
}

# Function to start monitoring services
start_monitoring() {
    print_header "Starting Monitoring Services"
    
    print_status "Starting Prometheus..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d prometheus
    
    print_status "Starting Grafana..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d grafana
    
    print_success "Monitoring services started"
}

# Function to check service health
check_health() {
    print_header "Checking Service Health"

    sleep 30  # Wait for services to start

    # Check producer health
    if curl -f http://localhost:3000/health > /dev/null 2>&1; then
        print_success "Producer service is healthy"
    else
        print_warning "Producer service health check failed"
    fi

    # Check consumer health
    if curl -f http://localhost:8081/health > /dev/null 2>&1; then
        print_success "Consumer service is healthy"
    else
        print_warning "Consumer service health check failed"
    fi

    # Check streams health
    if curl -f http://localhost:8082/health > /dev/null 2>&1; then
        print_success "Streams service is healthy"
    else
        print_warning "Streams service health check failed"
    fi

    # Check Web3 payment health
    if curl -f http://localhost:8084/health > /dev/null 2>&1; then
        print_success "Web3 Payment service is healthy"
    else
        print_warning "Web3 Payment service health check failed"
    fi

    # Check AI orchestrator health
    if curl -f http://localhost:8095/health > /dev/null 2>&1; then
        print_success "AI Orchestrator service is healthy"
    else
        print_warning "AI Orchestrator service health check failed"
    fi
}

# Function to show service URLs
show_urls() {
    print_header "Service URLs"
    echo -e "${GREEN}Producer API:${NC}           http://localhost:3000"
    echo -e "${GREEN}Consumer Health:${NC}        http://localhost:8081/health"
    echo -e "${GREEN}Streams Health:${NC}         http://localhost:8082/health"
    echo -e "${GREEN}Web3 Payment API:${NC}       http://localhost:8083"
    echo -e "${GREEN}Web3 Payment Health:${NC}    http://localhost:8084/health"
    echo -e "${GREEN}AI Orchestrator API:${NC}    http://localhost:8094"
    echo -e "${GREEN}AI Orchestrator Health:${NC} http://localhost:8095/health"
    echo -e "${GREEN}Prometheus:${NC}             http://localhost:9090"
    echo -e "${GREEN}Grafana:${NC}                http://localhost:3001 (admin/admin)"
    echo ""
    echo -e "${CYAN}Test the system:${NC}"
    echo "# Traditional coffee order:"
    echo "curl -X POST http://localhost:3000/order -H 'Content-Type: application/json' -d '{\"customer_name\":\"John Doe\",\"coffee_type\":\"Latte\"}'"
    echo ""
    echo "# Crypto payment:"
    echo "curl -X POST http://localhost:8083/payment/create -H 'Content-Type: application/json' -d '{\"order_id\":\"order_123\",\"customer_address\":\"0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1\",\"amount\":\"5.0\",\"currency\":\"USDC\",\"chain\":\"ethereum\"}'"
    echo ""
    echo "# AI agent task:"
    echo "curl -X POST http://localhost:8094/tasks/assign -H 'Content-Type: application/json' -d '{\"agent_id\":\"beverage-inventor\",\"action\":\"create_recipe\",\"inputs\":{\"name\":\"AI Latte\"},\"priority\":\"medium\"}'"
    echo ""
    echo "# AI workflow execution:"
    echo "curl -X POST http://localhost:8094/workflows/execute -H 'Content-Type: application/json' -d '{\"workflow_id\":\"coffee-order-processing\"}'"
    echo ""
    echo "# Run comprehensive tests:"
    echo "./scripts/test-core-services.sh"
    echo "./scripts/test-web3-payment.sh"
    echo "./scripts/test-ai-orchestrator.sh"
}

# Function to show logs
show_logs() {
    print_header "Service Logs"
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f
}

# Main execution
main() {
    print_header "Go Coffee Core Services Startup"
    
    # Check prerequisites
    check_docker
    check_docker_compose
    
    # Build services
    build_services
    
    # Start services in order
    start_infrastructure
    create_kafka_topics
    start_applications
    start_monitoring
    
    # Check health
    check_health
    
    # Show information
    show_urls
    
    print_success "All services started successfully!"
    print_status "Use 'docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f' to view logs"
    print_status "Use 'docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down' to stop all services"
}

# Handle command line arguments
case "${1:-}" in
    "logs")
        show_logs
        ;;
    "stop")
        print_status "Stopping all services..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
        print_success "All services stopped"
        ;;
    "restart")
        print_status "Restarting all services..."
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
        main
        ;;
    "health")
        check_health
        ;;
    *)
        main
        ;;
esac
