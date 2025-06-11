#!/bin/bash

# Test script for LLM Orchestrator Simple
set -e

BASE_URL="http://localhost:8080"
ORCHESTRATOR_PID=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to start orchestrator
start_orchestrator() {
    print_status "Starting LLM Orchestrator..."
    ./bin/llm-orchestrator-simple --config=config/llm-orchestrator-simple.yaml --port=8080 --log-level=info &
    ORCHESTRATOR_PID=$!
    sleep 3
    
    if kill -0 $ORCHESTRATOR_PID 2>/dev/null; then
        print_success "LLM Orchestrator started with PID: $ORCHESTRATOR_PID"
    else
        print_error "Failed to start LLM Orchestrator"
        exit 1
    fi
}

# Function to stop orchestrator
stop_orchestrator() {
    if [ ! -z "$ORCHESTRATOR_PID" ]; then
        print_status "Stopping LLM Orchestrator..."
        kill $ORCHESTRATOR_PID 2>/dev/null || true
        wait $ORCHESTRATOR_PID 2>/dev/null || true
        print_success "LLM Orchestrator stopped"
    fi
}

# Function to test health endpoint
test_health() {
    print_status "Testing health endpoint..."
    response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "$BASE_URL/health")
    
    if [ "$response" = "200" ]; then
        print_success "Health check passed"
        cat /tmp/health_response.json | jq '.' 2>/dev/null || cat /tmp/health_response.json
    else
        print_error "Health check failed with status: $response"
        return 1
    fi
}

# Function to test metrics endpoint
test_metrics() {
    print_status "Testing metrics endpoint..."
    response=$(curl -s -w "%{http_code}" -o /tmp/metrics_response.json "$BASE_URL/metrics")
    
    if [ "$response" = "200" ]; then
        print_success "Metrics endpoint working"
        cat /tmp/metrics_response.json | jq '.' 2>/dev/null || cat /tmp/metrics_response.json
    else
        print_error "Metrics endpoint failed with status: $response"
        return 1
    fi
}

# Function to test workload creation
test_create_workload() {
    print_status "Testing workload creation..."
    
    workload_data='{
        "name": "test-llama2",
        "modelName": "llama2",
        "modelType": "text-generation",
        "resources": {
            "cpu": "2000m",
            "memory": "8Gi",
            "gpu": 1
        },
        "labels": {
            "environment": "test",
            "team": "ai-research"
        }
    }'
    
    response=$(curl -s -w "%{http_code}" -o /tmp/create_response.json \
        -H "Content-Type: application/json" \
        -d "$workload_data" \
        "$BASE_URL/workloads")
    
    if [ "$response" = "201" ]; then
        print_success "Workload created successfully"
        cat /tmp/create_response.json | jq '.' 2>/dev/null || cat /tmp/create_response.json
        
        # Extract workload ID for further tests
        WORKLOAD_ID=$(cat /tmp/create_response.json | jq -r '.id' 2>/dev/null || echo "")
        if [ ! -z "$WORKLOAD_ID" ] && [ "$WORKLOAD_ID" != "null" ]; then
            print_success "Workload ID: $WORKLOAD_ID"
        fi
    else
        print_error "Workload creation failed with status: $response"
        return 1
    fi
}

# Function to test workload listing
test_list_workloads() {
    print_status "Testing workload listing..."
    response=$(curl -s -w "%{http_code}" -o /tmp/list_response.json "$BASE_URL/workloads")
    
    if [ "$response" = "200" ]; then
        print_success "Workload listing successful"
        cat /tmp/list_response.json | jq '.' 2>/dev/null || cat /tmp/list_response.json
    else
        print_error "Workload listing failed with status: $response"
        return 1
    fi
}

# Function to test workload retrieval
test_get_workload() {
    if [ -z "$WORKLOAD_ID" ]; then
        print_warning "No workload ID available, skipping get test"
        return 0
    fi
    
    print_status "Testing workload retrieval for ID: $WORKLOAD_ID"
    response=$(curl -s -w "%{http_code}" -o /tmp/get_response.json "$BASE_URL/workloads/$WORKLOAD_ID")
    
    if [ "$response" = "200" ]; then
        print_success "Workload retrieval successful"
        cat /tmp/get_response.json | jq '.' 2>/dev/null || cat /tmp/get_response.json
    else
        print_error "Workload retrieval failed with status: $response"
        return 1
    fi
}

# Function to test scheduling
test_schedule_workload() {
    if [ -z "$WORKLOAD_ID" ]; then
        print_warning "No workload ID available, skipping schedule test"
        return 0
    fi
    
    print_status "Testing workload scheduling for ID: $WORKLOAD_ID"
    
    schedule_data="{\"workloadId\": \"$WORKLOAD_ID\"}"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/schedule_response.json \
        -H "Content-Type: application/json" \
        -d "$schedule_data" \
        "$BASE_URL/schedule")
    
    if [ "$response" = "200" ]; then
        print_success "Workload scheduling successful"
        cat /tmp/schedule_response.json | jq '.' 2>/dev/null || cat /tmp/schedule_response.json
    else
        print_error "Workload scheduling failed with status: $response"
        return 1
    fi
}

# Function to test status endpoint
test_status() {
    print_status "Testing status endpoint..."
    response=$(curl -s -w "%{http_code}" -o /tmp/status_response.json "$BASE_URL/status")
    
    if [ "$response" = "200" ]; then
        print_success "Status endpoint working"
        cat /tmp/status_response.json | jq '.' 2>/dev/null || cat /tmp/status_response.json
    else
        print_error "Status endpoint failed with status: $response"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    stop_orchestrator
    rm -f /tmp/*_response.json
    print_success "Cleanup completed"
}

# Main test execution
main() {
    print_status "Starting LLM Orchestrator API Tests"
    
    # Set trap for cleanup
    trap cleanup EXIT
    
    # Check if binary exists
    if [ ! -f "./bin/llm-orchestrator-simple" ]; then
        print_error "LLM Orchestrator binary not found. Please build it first:"
        print_error "go build -o bin/llm-orchestrator-simple ./cmd/llm-orchestrator-simple"
        exit 1
    fi
    
    # Check if config exists
    if [ ! -f "config/llm-orchestrator-simple.yaml" ]; then
        print_error "Configuration file not found: config/llm-orchestrator-simple.yaml"
        exit 1
    fi
    
    # Start orchestrator
    start_orchestrator
    
    # Wait for startup
    sleep 2
    
    # Run tests
    test_health || exit 1
    test_metrics || exit 1
    test_status || exit 1
    test_list_workloads || exit 1
    test_create_workload || exit 1
    test_get_workload || exit 1
    test_schedule_workload || exit 1
    
    # Wait a bit for metrics to update
    sleep 5
    
    # Test metrics again to see updated values
    print_status "Testing updated metrics..."
    test_metrics || exit 1
    
    print_success "All tests passed! 🎉"
}

# Run main function
main "$@"
