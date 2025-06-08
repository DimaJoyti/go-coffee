#!/bin/bash

# Start all Go Coffee services
# This script builds and starts all microservices in the correct order

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
SERVICES=(
    "payment-service:8093"
    "auth-service:8091"
    "order-service:8094"
    "kitchen-service:8095"
    "api-gateway:8080"
)

PIDS=()

# Function to print colored output
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[ℹ]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

print_service() {
    echo -e "${PURPLE}[SERVICE]${NC} $1"
}

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    print_info "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" > /dev/null 2>&1; then
            print_status "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within $max_attempts seconds"
    return 1
}

# Function to build service
build_service() {
    local service_name=$1
    
    print_info "Building $service_name..."
    
    cd "cmd/$service_name"
    
    if go build -o "$service_name" .; then
        print_status "$service_name built successfully"
        cd ../..
        return 0
    else
        print_error "Failed to build $service_name"
        cd ../..
        return 1
    fi
}

# Function to start service
start_service() {
    local service_name=$1
    local port=$2
    
    print_service "Starting $service_name on port $port..."
    
    # Check if port is available
    if ! check_port $port; then
        print_warning "Port $port is already in use. Attempting to kill existing process..."
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        sleep 2
    fi
    
    # Build the service
    if ! build_service "$service_name"; then
        return 1
    fi
    
    # Start the service
    cd "cmd/$service_name"
    
    # Set environment variables based on service
    case $service_name in
        "payment-service")
            PAYMENT_SERVICE_PORT=$port ./"$service_name" &
            ;;
        "auth-service")
            AUTH_SERVICE_PORT=$port ./"$service_name" &
            ;;
        "order-service")
            ORDER_SERVICE_PORT=$port ./"$service_name" &
            ;;
        "kitchen-service")
            KITCHEN_SERVICE_PORT=$port ./"$service_name" &
            ;;
        "api-gateway")
            API_GATEWAY_PORT=$port ./"$service_name" &
            ;;
        *)
            ./"$service_name" &
            ;;
    esac
    
    local pid=$!
    PIDS+=($pid)
    
    cd ../..
    
    print_status "$service_name started with PID $pid"
    
    # Wait for service to be ready
    if wait_for_service "http://localhost:$port" "$service_name"; then
        return 0
    else
        print_error "$service_name failed to start properly"
        return 1
    fi
}

# Function to stop all services
stop_all_services() {
    print_info "Stopping all services..."
    
    for pid in "${PIDS[@]}"; do
        if kill -0 $pid 2>/dev/null; then
            print_info "Stopping process $pid..."
            kill $pid 2>/dev/null || true
        fi
    done
    
    # Wait for processes to stop
    sleep 3
    
    # Force kill if necessary
    for pid in "${PIDS[@]}"; do
        if kill -0 $pid 2>/dev/null; then
            print_warning "Force killing process $pid..."
            kill -9 $pid 2>/dev/null || true
        fi
    done
    
    print_status "All services stopped"
}

# Function to show service status
show_status() {
    echo ""
    print_info "Service Status:"
    echo "==============="
    
    for service_info in "${SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            print_status "$service_name (port $port) - Running"
        else
            print_error "$service_name (port $port) - Not responding"
        fi
    done
    
    echo ""
    print_info "API Gateway: http://localhost:8080"
    print_info "Documentation: http://localhost:8080/docs"
    echo ""
}

# Trap to ensure services are stopped on exit
trap stop_all_services EXIT

echo "☕ Starting Go Coffee Microservices"
echo "===================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Please run this script from the project root directory"
    exit 1
fi

print_info "Starting services in dependency order..."
echo ""

# Start services in order
for service_info in "${SERVICES[@]}"; do
    IFS=':' read -r service_name port <<< "$service_info"
    
    if start_service "$service_name" "$port"; then
        echo ""
    else
        print_error "Failed to start $service_name. Stopping all services."
        exit 1
    fi
done

print_status "🎉 All services started successfully!"
show_status

print_info "Testing API Gateway..."
echo ""

# Test API Gateway
if curl -s "http://localhost:8080/health" > /dev/null; then
    print_status "API Gateway is responding"
    
    # Test service status endpoint
    print_info "Checking service status through gateway..."
    curl -s "http://localhost:8080/api/v1/gateway/services" | head -c 200
    echo "..."
    echo ""
else
    print_error "API Gateway is not responding"
fi

print_info "Services are running. Press Ctrl+C to stop all services."
print_info "Visit http://localhost:8080/docs for API documentation"

# Keep script running
while true; do
    sleep 10
    
    # Check if all services are still running
    all_running=true
    for service_info in "${SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if ! curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            print_warning "$service_name (port $port) is not responding"
            all_running=false
        fi
    done
    
    if [ "$all_running" = false ]; then
        print_error "Some services are not responding. Check logs."
    fi
done
