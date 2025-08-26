#!/bin/bash

# Go Coffee Core Services Test Script
# This script tests the core coffee ordering flow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PRODUCER_URL="http://localhost:3000"
CONSUMER_HEALTH_URL="http://localhost:8081/health"
STREAMS_HEALTH_URL="http://localhost:8082/health"

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

# Function to test service health
test_health() {
    local service_name=$1
    local url=$2
    
    print_status "Testing $service_name health..."
    
    if curl -f -s "$url" > /dev/null; then
        print_success "$service_name is healthy"
        return 0
    else
        print_error "$service_name health check failed"
        return 1
    fi
}

# Function to test order placement
test_order_placement() {
    print_status "Testing order placement..."
    
    local response=$(curl -s -X POST "$PRODUCER_URL/order" \
        -H "Content-Type: application/json" \
        -d '{
            "customer_name": "Test Customer",
            "coffee_type": "Latte"
        }')
    
    if echo "$response" | grep -q "success.*true"; then
        print_success "Order placed successfully"
        echo "Response: $response"
        return 0
    else
        print_error "Order placement failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test order retrieval
test_order_retrieval() {
    print_status "Testing order retrieval..."
    
    local response=$(curl -s "$PRODUCER_URL/orders")
    
    if echo "$response" | grep -q "Test Customer"; then
        print_success "Order retrieval successful"
        echo "Found orders: $response"
        return 0
    else
        print_warning "No orders found or retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test multiple orders
test_multiple_orders() {
    print_status "Testing multiple order placement..."
    
    local orders=(
        '{"customer_name": "Alice Johnson", "coffee_type": "Espresso"}'
        '{"customer_name": "Bob Smith", "coffee_type": "Cappuccino"}'
        '{"customer_name": "Carol Davis", "coffee_type": "Americano"}'
        '{"customer_name": "David Wilson", "coffee_type": "Mocha"}'
        '{"customer_name": "Eva Brown", "coffee_type": "Macchiato"}'
    )
    
    local success_count=0
    
    for order in "${orders[@]}"; do
        local response=$(curl -s -X POST "$PRODUCER_URL/order" \
            -H "Content-Type: application/json" \
            -d "$order")
        
        if echo "$response" | grep -q "success.*true"; then
            ((success_count++))
            print_success "Order placed: $(echo "$order" | jq -r '.customer_name') - $(echo "$order" | jq -r '.coffee_type')"
        else
            print_error "Failed to place order: $order"
        fi
        
        # Small delay between orders
        sleep 0.5
    done
    
    print_status "Successfully placed $success_count out of ${#orders[@]} orders"
}

# Function to test system load
test_system_load() {
    print_status "Testing system under load..."
    
    local total_orders=20
    local success_count=0
    
    for i in $(seq 1 $total_orders); do
        local customer_name="LoadTest_Customer_$i"
        local coffee_types=("Latte" "Cappuccino" "Espresso" "Americano" "Mocha")
        local coffee_type=${coffee_types[$((i % ${#coffee_types[@]}))]}
        
        local response=$(curl -s -X POST "$PRODUCER_URL/order" \
            -H "Content-Type: application/json" \
            -d "{\"customer_name\": \"$customer_name\", \"coffee_type\": \"$coffee_type\"}")
        
        if echo "$response" | grep -q "success.*true"; then
            ((success_count++))
        fi
        
        # Progress indicator
        if ((i % 5 == 0)); then
            print_status "Processed $i/$total_orders orders..."
        fi
    done
    
    print_success "Load test completed: $success_count/$total_orders orders successful"
}

# Function to check metrics
test_metrics() {
    print_status "Testing metrics endpoints..."
    
    # Test producer metrics
    if curl -f -s "$PRODUCER_URL/metrics" | grep -q "go_"; then
        print_success "Producer metrics available"
    else
        print_warning "Producer metrics not available"
    fi
    
    # Test consumer metrics
    if curl -f -s "http://localhost:8081/metrics" | grep -q "go_"; then
        print_success "Consumer metrics available"
    else
        print_warning "Consumer metrics not available"
    fi
    
    # Test streams metrics
    if curl -f -s "http://localhost:8082/metrics" | grep -q "go_"; then
        print_success "Streams metrics available"
    else
        print_warning "Streams metrics not available"
    fi
}

# Function to run comprehensive tests
run_comprehensive_tests() {
    print_header "Go Coffee Core Services Test Suite"
    
    local failed_tests=0
    
    # Test 1: Health checks
    print_header "Test 1: Health Checks"
    test_health "Producer" "$PRODUCER_URL/health" || ((failed_tests++))
    test_health "Consumer" "$CONSUMER_HEALTH_URL" || ((failed_tests++))
    test_health "Streams" "$STREAMS_HEALTH_URL" || ((failed_tests++))
    
    # Test 2: Basic functionality
    print_header "Test 2: Basic Order Flow"
    test_order_placement || ((failed_tests++))
    sleep 2  # Allow time for processing
    test_order_retrieval || ((failed_tests++))
    
    # Test 3: Multiple orders
    print_header "Test 3: Multiple Orders"
    test_multiple_orders || ((failed_tests++))
    
    # Test 4: System load
    print_header "Test 4: Load Testing"
    test_system_load || ((failed_tests++))
    
    # Test 5: Metrics
    print_header "Test 5: Metrics Endpoints"
    test_metrics || ((failed_tests++))
    
    # Summary
    print_header "Test Summary"
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests passed! ✅"
        print_status "The Go Coffee core services are working correctly."
    else
        print_error "$failed_tests test(s) failed ❌"
        print_status "Please check the logs for more details."
    fi
    
    return $failed_tests
}

# Function to show system status
show_system_status() {
    print_header "System Status"
    
    echo -e "${CYAN}Service Health:${NC}"
    curl -s "$PRODUCER_URL/health" | jq '.' 2>/dev/null || echo "Producer: Not responding"
    curl -s "$CONSUMER_HEALTH_URL" | jq '.' 2>/dev/null || echo "Consumer: Not responding"
    curl -s "$STREAMS_HEALTH_URL" | jq '.' 2>/dev/null || echo "Streams: Not responding"
    
    echo -e "\n${CYAN}Recent Orders:${NC}"
    curl -s "$PRODUCER_URL/orders" | jq '.' 2>/dev/null || echo "No orders found"
}

# Main execution
main() {
    case "${1:-}" in
        "health")
            print_header "Health Check"
            test_health "Producer" "$PRODUCER_URL/health"
            test_health "Consumer" "$CONSUMER_HEALTH_URL"
            test_health "Streams" "$STREAMS_HEALTH_URL"
            ;;
        "order")
            test_order_placement
            ;;
        "load")
            test_system_load
            ;;
        "metrics")
            test_metrics
            ;;
        "status")
            show_system_status
            ;;
        "all"|"")
            run_comprehensive_tests
            ;;
        *)
            echo "Usage: $0 [health|order|load|metrics|status|all]"
            echo "  health  - Check service health"
            echo "  order   - Test single order placement"
            echo "  load    - Run load test"
            echo "  metrics - Check metrics endpoints"
            echo "  status  - Show system status"
            echo "  all     - Run all tests (default)"
            ;;
    esac
}

# Check if jq is available for JSON parsing
if ! command -v jq > /dev/null 2>&1; then
    print_warning "jq is not installed. JSON output will be raw."
fi

# Run main function with arguments
main "$@"
