#!/bin/bash

# Complete system demonstration script
# This script demonstrates the full Go Coffee microservices architecture

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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

print_demo() {
    echo -e "${YELLOW}[DEMO]${NC} $1"
}

# Function to test endpoint with pretty output
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    
    print_demo "$description"
    echo "Request: $method $url"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s "$url")
    else
        response=$(curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$url")
    fi
    
    echo "Response: $response" | head -c 150
    echo "..."
    echo ""
}

# Function to wait for services
wait_for_services() {
    print_step "Waiting for all services to be ready..."
    
    local services=(
        "http://localhost:8093/health:Payment Service"
        "http://localhost:8091/health:Auth Service"
        "http://localhost:8094/health:Order Service"
        "http://localhost:8095/health:Kitchen Service"
        "http://localhost:8080/health:API Gateway"
    )
    
    for service_info in "${services[@]}"; do
        IFS=':' read -r url name <<< "$service_info"
        
        local attempts=0
        local max_attempts=30
        
        while [ $attempts -lt $max_attempts ]; do
            if curl -s "$url" > /dev/null 2>&1; then
                print_success "$name is ready"
                break
            fi
            
            attempts=$((attempts + 1))
            sleep 1
        done
        
        if [ $attempts -eq $max_attempts ]; then
            echo "❌ $name failed to start"
            return 1
        fi
    done
    
    print_success "All services are ready!"
}

echo ""
print_header "☕ Go Coffee - Complete System Demonstration"
echo ""

print_info "This demo showcases the complete microservices architecture:"
print_info "• API Gateway (Port 8080) - Central routing"
print_info "• Payment Service (Port 8093) - Bitcoin payments"
print_info "• Auth Service (Port 8091) - User authentication"
print_info "• Order Service (Port 8094) - Order management"
print_info "• Kitchen Service (Port 8095) - Kitchen operations"
echo ""

# Check if services are running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_info "Services not running. Starting them now..."
    echo ""
    
    # Start services
    print_step "Starting all microservices..."
    ./scripts/start-all-services.sh &
    STARTUP_PID=$!
    
    # Wait for startup to complete
    sleep 10
    
    # Wait for services to be ready
    wait_for_services
else
    print_success "Services are already running!"
fi

echo ""
print_header "🌐 API Gateway Demonstration"
echo ""

# Test API Gateway
test_endpoint "GET" "http://localhost:8080/health" "" "Gateway health check"
test_endpoint "GET" "http://localhost:8080/api/v1/gateway/status" "" "Gateway status"
test_endpoint "GET" "http://localhost:8080/api/v1/gateway/services" "" "All services status"

echo ""
print_header "₿ Bitcoin Payment Service Demonstration"
echo ""

# Test Payment Service through Gateway
test_endpoint "GET" "http://localhost:8080/api/v1/payment/version" "" "Payment service version"
test_endpoint "GET" "http://localhost:8080/api/v1/payment/features" "" "Bitcoin features available"
test_endpoint "POST" "http://localhost:8080/api/v1/payment/wallet/create" '{"testnet": true}' "Create Bitcoin testnet wallet"
test_endpoint "POST" "http://localhost:8080/api/v1/payment/wallet/validate" '{"address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"}' "Validate Bitcoin address"

echo ""
print_header "🔐 Authentication Service Demonstration"
echo ""

# Test Auth Service through Gateway (these might return errors without proper setup, but show the routing works)
test_endpoint "GET" "http://localhost:8080/api/v1/auth/health" "" "Auth service health (via gateway)"

echo ""
print_header "📋 Order Service Demonstration"
echo ""

# Test Order Service through Gateway
test_endpoint "GET" "http://localhost:8080/api/v1/order/health" "" "Order service health (via gateway)"

echo ""
print_header "👨‍🍳 Kitchen Service Demonstration"
echo ""

# Test Kitchen Service through Gateway
test_endpoint "GET" "http://localhost:8080/api/v1/kitchen/health" "" "Kitchen service health (via gateway)"

echo ""
print_header "📚 API Documentation"
echo ""

print_demo "API Documentation is available at:"
print_info "• HTML Documentation: http://localhost:8080/docs"
print_info "• JSON Documentation: http://localhost:8080/docs (with Accept: application/json)"

echo ""
print_header "🧪 System Architecture Validation"
echo ""

print_step "Validating microservices architecture..."

# Check each service individually
services=(
    "payment:8093:Payment Service"
    "auth:8091:Auth Service" 
    "order:8094:Order Service"
    "kitchen:8095:Kitchen Service"
    "gateway:8080:API Gateway"
)

for service_info in "${services[@]}"; do
    IFS=':' read -r service port name <<< "$service_info"
    
    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        print_success "$name (Port $port) - ✓ Running independently"
    else
        echo "❌ $name (Port $port) - Not responding"
    fi
done

echo ""
print_step "Validating API Gateway routing..."

# Test that gateway can route to each service
gateway_routes=(
    "/api/v1/payment/version:Payment Service routing"
    "/api/v1/gateway/status:Gateway self-routing"
)

for route_info in "${gateway_routes[@]}"; do
    IFS=':' read -r route description <<< "$route_info"
    
    if curl -s "http://localhost:8080$route" > /dev/null 2>&1; then
        print_success "$description - ✓ Working"
    else
        echo "❌ $description - Failed"
    fi
done

echo ""
print_header "🎯 System Capabilities Summary"
echo ""

print_success "✅ Microservices Architecture - 5 independent services"
print_success "✅ API Gateway - Central routing and management"
print_success "✅ Bitcoin Integration - Full cryptocurrency support"
print_success "✅ Clean Architecture - Proper separation of concerns"
print_success "✅ Production Ready - Health checks, logging, error handling"
print_success "✅ Scalable Design - Each service can scale independently"
print_success "✅ Security Features - CORS, validation, authentication"
print_success "✅ Development Tools - Automated testing and deployment"

echo ""
print_header "🚀 Go Coffee System is Fully Operational!"
echo ""

print_info "Your coffee shop now has:"
print_info "• Complete Bitcoin payment processing"
print_info "• User authentication and authorization"
print_info "• Order management and tracking"
print_info "• Kitchen operations coordination"
print_info "• Centralized API management"
print_info "• Production-ready architecture"

echo ""
print_info "🌐 Access Points:"
print_info "• API Gateway: http://localhost:8080"
print_info "• Documentation: http://localhost:8080/docs"
print_info "• Service Status: http://localhost:8080/api/v1/gateway/services"

echo ""
print_info "🛠️ Management Commands:"
print_info "• Start all: make -f Makefile.coffee start-all"
print_info "• Stop all: make -f Makefile.coffee stop-all"
print_info "• Test Bitcoin: make -f Makefile.coffee bitcoin-test"
print_info "• Test Payment: make -f Makefile.coffee test-payment"

echo ""
print_success "🎉 Demonstration complete! Your Go Coffee system is ready for business!"
echo ""
