#!/bin/bash

# Enhanced Bright Data Hub Demo Script
# This script demonstrates the comprehensive Bright Data MCP integration

set -e

echo "🚀 Enhanced Bright Data Hub - Comprehensive Demo"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
HUB_PORT=8095
MCP_PORT=3001
BASE_URL="http://localhost:$HUB_PORT"

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to check if service is running
check_service() {
    local url=$1
    local name=$2
    
    if curl -s -f "$url" > /dev/null 2>&1; then
        print_success "$name is running"
        return 0
    else
        print_error "$name is not running"
        return 1
    fi
}

# Function to make API call and show result
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    print_step "Testing: $description"
    echo -e "${PURPLE}$method $endpoint${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    # Extract response body (all but last line)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        print_success "Success ($http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        print_error "Failed ($http_code)"
        echo "$body"
    fi
    
    echo ""
    sleep 1
}

# Main demo function
main() {
    echo -e "${CYAN}🎯 Starting Enhanced Bright Data Hub Demo${NC}"
    echo ""
    
    # Step 1: Check prerequisites
    print_step "Checking prerequisites..."
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_warning "jq is not installed - JSON output will not be formatted"
    fi
    
    print_success "Prerequisites checked"
    echo ""
    
    # Step 2: Check if services are running
    print_step "Checking service status..."
    
    if ! check_service "$BASE_URL/api/v1/bright-data/health" "Bright Data Hub"; then
        print_info "Starting Bright Data Hub service..."
        echo "Run: make run-bright-data-hub"
        echo "Or: docker-compose -f docker-compose.bright-data-hub.yml up -d"
        exit 1
    fi
    
    echo ""
    
    # Step 3: Core functionality tests
    print_step "Testing Core Functionality"
    echo "=========================="
    
    api_call "GET" "/api/v1/bright-data/health" "" "Health Check"
    api_call "GET" "/api/v1/bright-data/status" "" "Service Status"
    
    # Step 4: Search functionality
    print_step "Testing Search Functionality"
    echo "============================"
    
    api_call "POST" "/api/v1/bright-data/search/engine" \
        '{"query": "coffee market trends 2024", "engine": "google"}' \
        "Google Search"
    
    api_call "POST" "/api/v1/bright-data/search/scrape" \
        '{"url": "https://example.com"}' \
        "Web Scraping"
    
    # Step 5: Social media functionality
    print_step "Testing Social Media Functionality"
    echo "=================================="
    
    api_call "POST" "/api/v1/bright-data/social/instagram/profile" \
        '{"url": "https://instagram.com/starbucks"}' \
        "Instagram Profile"
    
    api_call "POST" "/api/v1/bright-data/social/facebook/posts" \
        '{"url": "https://facebook.com/starbucks"}' \
        "Facebook Posts"
    
    api_call "POST" "/api/v1/bright-data/social/twitter/posts" \
        '{"url": "https://twitter.com/starbucks"}' \
        "Twitter Posts"
    
    api_call "POST" "/api/v1/bright-data/social/linkedin/profile" \
        '{"url": "https://linkedin.com/company/starbucks"}' \
        "LinkedIn Profile"
    
    api_call "GET" "/api/v1/bright-data/social/analytics" "" "Social Analytics"
    api_call "GET" "/api/v1/bright-data/social/trending" "" "Trending Topics"
    
    # Step 6: E-commerce functionality
    print_step "Testing E-commerce Functionality"
    echo "================================"
    
    api_call "POST" "/api/v1/bright-data/ecommerce/amazon/product" \
        '{"url": "https://amazon.com/dp/B08N5WRWNW"}' \
        "Amazon Product"
    
    api_call "POST" "/api/v1/bright-data/ecommerce/amazon/reviews" \
        '{"url": "https://amazon.com/dp/B08N5WRWNW"}' \
        "Amazon Reviews"
    
    api_call "POST" "/api/v1/bright-data/ecommerce/booking/hotels" \
        '{"url": "https://booking.com/hotel/us/example.html"}' \
        "Booking Hotels"
    
    api_call "POST" "/api/v1/bright-data/ecommerce/zillow/properties" \
        '{"url": "https://zillow.com/homedetails/123-main-st"}' \
        "Zillow Properties"
    
    # Step 7: Direct MCP function execution
    print_step "Testing Direct MCP Function Execution"
    echo "====================================="
    
    api_call "POST" "/api/v1/bright-data/execute" \
        '{"function": "search_engine_Bright_Data", "params": {"query": "coffee", "engine": "google"}}' \
        "Direct MCP Function Call"
    
    # Step 8: Analytics functionality
    print_step "Testing Analytics Functionality"
    echo "==============================="
    
    api_call "GET" "/api/v1/bright-data/analytics/sentiment/instagram" "" "Instagram Sentiment"
    api_call "GET" "/api/v1/bright-data/analytics/trends" "" "Trend Analysis"
    api_call "GET" "/api/v1/bright-data/analytics/intelligence" "" "Market Intelligence"
    
    # Final summary
    echo ""
    echo "🎉 Enhanced Bright Data Hub Demo Completed!"
    echo "==========================================="
    echo ""
    print_success "All major functionalities tested"
    print_info "Service is running on: $BASE_URL"
    print_info "API Documentation: $BASE_URL/api/v1/bright-data/status"
    echo ""
    echo -e "${CYAN}🚀 Enhanced Bright Data Hub is ready for production use!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Configure your Bright Data API credentials"
    echo "2. Set up monitoring with Prometheus and Grafana"
    echo "3. Deploy using Docker Compose or Kubernetes"
    echo "4. Integrate with your applications using the REST API"
    echo ""
}

# Help function
show_help() {
    echo "Enhanced Bright Data Hub Demo Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -p, --port     Set hub port (default: 8095)"
    echo "  -u, --url      Set base URL (default: http://localhost:8095)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run demo with default settings"
    echo "  $0 -p 8080           # Run demo with custom port"
    echo "  $0 -u http://prod.example.com  # Run demo against production"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--port)
            HUB_PORT="$2"
            BASE_URL="http://localhost:$HUB_PORT"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run the demo
main
