#!/bin/bash

# Developer DAO Platform - Local Development Runner
set -e

echo "🚀 Developer DAO Platform - Local Development"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23+ and try again."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
}

# Build all services
build_services() {
    print_status "Building Go services..."
    
    mkdir -p bin
    
    # Build bounty service
    print_status "Building bounty service..."
    go build -o bin/bounty-service ./cmd/bounty-service
    
    # Build marketplace service
    print_status "Building marketplace service..."
    go build -o bin/marketplace-service ./cmd/marketplace-service
    
    # Build metrics service
    print_status "Building metrics service..."
    go build -o bin/metrics-service ./cmd/metrics-service
    
    # Build DAO governance service
    print_status "Building DAO governance service..."
    go build -o bin/dao-governance-service ./cmd/dao-governance-service
    
    print_success "All services built successfully"
}

# Start services
start_services() {
    print_status "Starting Developer DAO Platform services..."
    
    # Create logs directory
    mkdir -p logs
    
    # Start bounty service
    print_status "Starting bounty service on port 8080..."
    ./bin/bounty-service > logs/bounty-service.log 2>&1 &
    BOUNTY_PID=$!
    echo $BOUNTY_PID > logs/bounty-service.pid
    
    # Wait a moment for service to start
    sleep 2
    
    # Start marketplace service
    print_status "Starting marketplace service on port 8081..."
    ./bin/marketplace-service > logs/marketplace-service.log 2>&1 &
    MARKETPLACE_PID=$!
    echo $MARKETPLACE_PID > logs/marketplace-service.pid
    
    # Wait a moment for service to start
    sleep 2
    
    # Start metrics service
    print_status "Starting metrics service on port 8082..."
    ./bin/metrics-service > logs/metrics-service.log 2>&1 &
    METRICS_PID=$!
    echo $METRICS_PID > logs/metrics-service.pid
    
    # Wait a moment for service to start
    sleep 2
    
    # Start DAO governance service
    print_status "Starting DAO governance service on port 8084..."
    ./bin/dao-governance-service > logs/dao-governance-service.log 2>&1 &
    DAO_PID=$!
    echo $DAO_PID > logs/dao-governance-service.pid
    
    # Wait for services to fully start
    sleep 5
    
    print_success "All services started"
}

# Health check
health_check() {
    print_status "Performing health checks..."
    
    # Check services
    services=("8080:bounty-service" "8081:marketplace-service" "8082:metrics-service" "8084:dao-governance-service")
    
    for service in "${services[@]}"; do
        port=$(echo $service | cut -d: -f1)
        name=$(echo $service | cut -d: -f2)
        
        if curl -f -s http://localhost:$port/health > /dev/null 2>&1; then
            print_success "$name is healthy (port $port)"
        else
            print_warning "$name health check failed (port $port)"
            print_warning "Check logs/$(echo $name | tr '-' '_').log for details"
        fi
    done
}

# Show status
show_status() {
    echo ""
    echo "🎉 Developer DAO Platform Status"
    echo "================================"
    echo ""
    echo "📱 Services:"
    echo "   Bounty Service:      http://localhost:8080"
    echo "   Marketplace Service: http://localhost:8081"
    echo "   Metrics Service:     http://localhost:8082"
    echo "   DAO Governance:      http://localhost:8084"
    echo ""
    echo "📚 API Documentation:"
    echo "   Swagger UI:         http://localhost:8080/swagger/index.html"
    echo ""
    echo "📊 Logs:"
    echo "   Bounty Service:     tail -f logs/bounty-service.log"
    echo "   Marketplace:        tail -f logs/marketplace-service.log"
    echo "   Metrics:           tail -f logs/metrics-service.log"
    echo "   DAO Governance:    tail -f logs/dao-governance-service.log"
    echo ""
    echo "🛑 To stop services:"
    echo "   ./run-local.sh stop"
    echo ""
}

# Stop services
stop_services() {
    print_status "Stopping Developer DAO Platform services..."
    
    # Stop services using PID files
    for service in bounty-service marketplace-service metrics-service dao-governance-service; do
        if [ -f "logs/${service}.pid" ]; then
            PID=$(cat logs/${service}.pid)
            if kill -0 $PID 2>/dev/null; then
                print_status "Stopping $service (PID: $PID)..."
                kill $PID
                rm logs/${service}.pid
            else
                print_warning "$service was not running"
            fi
        fi
    done
    
    print_success "All services stopped"
}

# Show logs
show_logs() {
    if [ -z "$2" ]; then
        print_status "Available logs:"
        echo "  bounty-service"
        echo "  marketplace-service"
        echo "  metrics-service"
        echo "  dao-governance-service"
        echo ""
        echo "Usage: ./run-local.sh logs <service-name>"
        return
    fi
    
    SERVICE="$2"
    LOG_FILE="logs/${SERVICE}.log"
    
    if [ -f "$LOG_FILE" ]; then
        print_status "Showing logs for $SERVICE (press Ctrl+C to exit):"
        tail -f "$LOG_FILE"
    else
        print_error "Log file not found: $LOG_FILE"
    fi
}

# Main function
main() {
    case "${1:-}" in
        "build")
            check_go
            build_services
            ;;
        "start")
            check_go
            build_services
            start_services
            health_check
            show_status
            ;;
        "stop")
            stop_services
            ;;
        "status")
            health_check
            show_status
            ;;
        "logs")
            show_logs "$@"
            ;;
        "health")
            health_check
            ;;
        "clean")
            stop_services
            print_status "Cleaning up..."
            rm -rf bin logs
            print_success "Cleanup completed"
            ;;
        *)
            echo "Developer DAO Platform - Local Development Runner"
            echo ""
            echo "Usage: $0 {build|start|stop|status|logs|health|clean}"
            echo ""
            echo "Commands:"
            echo "  build   - Build all services"
            echo "  start   - Build and start all services"
            echo "  stop    - Stop all running services"
            echo "  status  - Show service status and URLs"
            echo "  logs    - Show logs for a specific service"
            echo "  health  - Perform health checks"
            echo "  clean   - Stop services and clean up files"
            echo ""
            echo "Examples:"
            echo "  $0 start                    # Start all services"
            echo "  $0 logs bounty-service      # Show bounty service logs"
            echo "  $0 stop                     # Stop all services"
            ;;
    esac
}

# Run main function
main "$@"
