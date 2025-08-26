#!/bin/bash

# Go Coffee LangGraph Integration - Coffee Creation Example
# This script demonstrates how to use the LangGraph integration to create coffee recipes

set -e

# Configuration
BASE_URL="http://localhost:8080"
GRAPH_ID="coffee_creation"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if server is running
check_server() {
    log_info "Checking if LangGraph server is running..."
    
    if curl -s -f "$BASE_URL/health" > /dev/null; then
        log_success "Server is running"
    else
        log_error "Server is not running. Please start it with: make run"
        exit 1
    fi
}

# Execute workflow and return execution ID
execute_workflow() {
    local request_data="$1"
    local description="$2"
    
    log_info "Executing workflow: $description"
    
    local response=$(curl -s -X POST "$BASE_URL/api/v1/workflows/execute" \
        -H "Content-Type: application/json" \
        -d "$request_data")
    
    local execution_id=$(echo "$response" | jq -r '.execution_id')
    local status=$(echo "$response" | jq -r '.status')
    local error=$(echo "$response" | jq -r '.error // empty')
    
    if [ "$error" != "" ]; then
        log_error "Workflow execution failed: $error"
        return 1
    fi
    
    if [ "$status" = "completed" ]; then
        log_success "Workflow completed successfully"
        echo "$response" | jq '.result'
    else
        log_warning "Workflow status: $status"
        echo "$execution_id"
    fi
}

# Wait for execution to complete
wait_for_completion() {
    local execution_id="$1"
    local max_wait=60
    local wait_time=0
    
    log_info "Waiting for execution $execution_id to complete..."
    
    while [ $wait_time -lt $max_wait ]; do
        local response=$(curl -s "$BASE_URL/api/v1/executions/$execution_id")
        local status=$(echo "$response" | jq -r '.status')
        
        case "$status" in
            "completed")
                log_success "Execution completed successfully"
                return 0
                ;;
            "failed")
                log_error "Execution failed"
                echo "$response" | jq '.error // empty'
                return 1
                ;;
            "cancelled")
                log_warning "Execution was cancelled"
                return 1
                ;;
            *)
                log_info "Status: $status (waiting...)"
                sleep 2
                wait_time=$((wait_time + 2))
                ;;
        esac
    done
    
    log_error "Execution timed out after ${max_wait}s"
    return 1
}

# Example 1: Winter Spiced Coffee
example_winter_spiced() {
    log_info "🌨️  Example 1: Winter Spiced Coffee"
    
    local request='{
        "graph_id": "'$GRAPH_ID'",
        "input_data": {
            "season": "winter",
            "flavor_profile": "spicy",
            "temperature": "hot",
            "occasion": "morning",
            "caffeine_level": "high",
            "generate_variations": true,
            "check_inventory": true
        },
        "priority": "medium"
    }'
    
    execute_workflow "$request" "Winter Spiced Coffee Creation"
}

# Example 2: Summer Refreshing Drink
example_summer_refreshing() {
    log_info "☀️  Example 2: Summer Refreshing Drink"
    
    local request='{
        "graph_id": "'$GRAPH_ID'",
        "input_data": {
            "season": "summer",
            "flavor_profile": "fruity",
            "temperature": "cold",
            "occasion": "afternoon",
            "caffeine_level": "medium",
            "dietary_restrictions": ["dairy_free"],
            "generate_variations": false
        },
        "priority": "low"
    }'
    
    execute_workflow "$request" "Summer Refreshing Drink Creation"
}

# Example 3: Fall Comfort Blend
example_fall_comfort() {
    log_info "🍂 Example 3: Fall Comfort Blend"
    
    local request='{
        "graph_id": "'$GRAPH_ID'",
        "input_data": {
            "season": "fall",
            "flavor_profile": "nutty",
            "temperature": "hot",
            "occasion": "evening",
            "caffeine_level": "low",
            "customer_preferences": {
                "sweetness_level": "medium",
                "milk_type": "oat"
            },
            "generate_variations": true
        },
        "priority": "high"
    }'
    
    execute_workflow "$request" "Fall Comfort Blend Creation"
}

# Example 4: Spring Floral Delight
example_spring_floral() {
    log_info "🌸 Example 4: Spring Floral Delight"
    
    local request='{
        "graph_id": "'$GRAPH_ID'",
        "input_data": {
            "season": "spring",
            "flavor_profile": "floral",
            "temperature": "hot",
            "occasion": "special",
            "caffeine_level": "medium",
            "dietary_restrictions": ["vegan", "gluten_free"],
            "generate_variations": true,
            "check_inventory": false
        },
        "priority": "urgent"
    }'
    
    execute_workflow "$request" "Spring Floral Delight Creation"
}

# Show server statistics
show_stats() {
    log_info "📊 Server Statistics"
    
    local stats=$(curl -s "$BASE_URL/api/v1/stats")
    echo "$stats" | jq '.'
}

# List active executions
list_executions() {
    log_info "📋 Active Executions"
    
    local executions=$(curl -s "$BASE_URL/api/v1/executions")
    echo "$executions" | jq '.'
}

# Main execution
main() {
    echo "🤖 Go Coffee LangGraph Integration - Coffee Creation Examples"
    echo "============================================================"
    echo ""
    
    # Check dependencies
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed. Please install jq first."
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed. Please install curl first."
        exit 1
    fi
    
    # Check server
    check_server
    echo ""
    
    # Run examples
    case "${1:-all}" in
        "winter")
            example_winter_spiced
            ;;
        "summer")
            example_summer_refreshing
            ;;
        "fall")
            example_fall_comfort
            ;;
        "spring")
            example_spring_floral
            ;;
        "stats")
            show_stats
            ;;
        "executions")
            list_executions
            ;;
        "all")
            example_winter_spiced
            echo ""
            sleep 1
            
            example_summer_refreshing
            echo ""
            sleep 1
            
            example_fall_comfort
            echo ""
            sleep 1
            
            example_spring_floral
            echo ""
            sleep 1
            
            show_stats
            ;;
        *)
            echo "Usage: $0 [winter|summer|fall|spring|stats|executions|all]"
            echo ""
            echo "Examples:"
            echo "  $0 winter     - Create winter spiced coffee"
            echo "  $0 summer     - Create summer refreshing drink"
            echo "  $0 fall       - Create fall comfort blend"
            echo "  $0 spring     - Create spring floral delight"
            echo "  $0 stats      - Show server statistics"
            echo "  $0 executions - List active executions"
            echo "  $0 all        - Run all examples (default)"
            exit 1
            ;;
    esac
    
    echo ""
    log_success "Examples completed successfully!"
    echo ""
    echo "💡 Tips:"
    echo "  - Check server stats: curl $BASE_URL/api/v1/stats"
    echo "  - View health status: curl $BASE_URL/health"
    echo "  - Monitor executions: curl $BASE_URL/api/v1/executions"
}

# Run main function with all arguments
main "$@"
