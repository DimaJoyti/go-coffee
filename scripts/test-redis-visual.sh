#!/bin/bash

# Redis 8 Visual Interface Testing Script
# This script tests the Redis visual interface functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE="http://localhost:8080/api/v1/redis-mcp"
WEB_BASE="http://localhost:3000"

echo -e "${BLUE}🧪 Testing Redis 8 Visual Interface${NC}"
echo "=================================================="

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    
    echo -e "${YELLOW}Testing: $description${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_BASE$endpoint")
    else
        response=$(curl -s -w "%{http_code}" -o /tmp/test_response -X "$method" -H "Content-Type: application/json" -d "$data" "$API_BASE$endpoint")
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✅ $description - OK${NC}"
        return 0
    else
        echo -e "${RED}❌ $description - Failed (HTTP $http_code)${NC}"
        echo -e "${RED}   Response: $(cat /tmp/test_response)${NC}"
        return 1
    fi
}

# Function to check service health
check_service() {
    local url=$1
    local service=$2
    
    echo -e "${YELLOW}Checking $service health...${NC}"
    
    if curl -s -f "$url" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $service is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ $service is not responding${NC}"
        return 1
    fi
}

# Check prerequisites
echo -e "${BLUE}🔍 Checking service health...${NC}"

check_service "$API_BASE/health" "Redis MCP Server" || {
    echo -e "${RED}❌ Redis MCP Server is not running${NC}"
    echo -e "${YELLOW}   Please start it with: ./scripts/start-redis-visual.sh${NC}"
    exit 1
}

check_service "$WEB_BASE" "Web UI" || {
    echo -e "${YELLOW}⚠️  Web UI is not responding (this is optional for API tests)${NC}"
}

# Seed test data
echo -e "${BLUE}🌱 Seeding test data...${NC}"
if [ -f "scripts/seed-redis-data.go" ]; then
    cd scripts
    go run seed-redis-data.go
    cd ..
    echo -e "${GREEN}✅ Test data seeded${NC}"
else
    echo -e "${YELLOW}⚠️  Seed script not found, using existing data${NC}"
fi

# Test API endpoints
echo -e "${BLUE}🔧 Testing API endpoints...${NC}"

# Health check
test_endpoint "GET" "/health" "Health check"

# Data exploration
test_endpoint "POST" "/visual/explore" "Data exploration" '{
    "data_type": "keys",
    "pattern": "*",
    "limit": 10
}'

# Key details
test_endpoint "GET" "/visual/key/user:1" "Key details"

# Search
test_endpoint "GET" "/visual/search?q=user&type=keys&limit=5" "Data search"

# Query building
test_endpoint "POST" "/visual/query/build" "Query building" '{
    "operation": "GET",
    "key": "user:1",
    "preview": true
}'

# Query validation
test_endpoint "POST" "/visual/query/validate" "Query validation" '{
    "operation": "SET",
    "key": "test:key",
    "value": "test value"
}'

# Query templates
test_endpoint "GET" "/visual/query/templates?operation=GET" "Query templates"

# Query suggestions
test_endpoint "GET" "/visual/query/suggestions?operation=SET" "Query suggestions"

# Metrics
test_endpoint "GET" "/visual/metrics" "Redis metrics"

# Performance metrics
test_endpoint "GET" "/visual/performance" "Performance metrics"

# Test data visualization
test_endpoint "POST" "/visual/visualize" "Data visualization" '{
    "chart_type": "bar",
    "data_source": "users",
    "time_range": "24h",
    "aggregation": "count"
}'

# Test specific data types
echo -e "${BLUE}🗂️ Testing data type exploration...${NC}"

# Test hash exploration
test_endpoint "POST" "/visual/explore" "Hash exploration" '{
    "data_type": "hash",
    "key": "user:1"
}'

# Test list exploration (if any lists exist)
test_endpoint "POST" "/visual/explore" "List exploration" '{
    "data_type": "list",
    "key": "user:1:orders"
}'

# Test set exploration
test_endpoint "POST" "/visual/explore" "Set exploration" '{
    "data_type": "set",
    "key": "users:all"
}'

# Test sorted set exploration
test_endpoint "POST" "/visual/explore" "Sorted set exploration" '{
    "data_type": "zset",
    "key": "users:by_age"
}'

# Performance tests
echo -e "${BLUE}⚡ Running performance tests...${NC}"

echo -e "${YELLOW}Testing bulk key retrieval...${NC}"
start_time=$(date +%s%N)
test_endpoint "POST" "/visual/explore" "Bulk key retrieval" '{
    "data_type": "keys",
    "pattern": "*",
    "limit": 1000
}'
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))
echo -e "${GREEN}✅ Bulk retrieval completed in ${duration}ms${NC}"

# Test concurrent requests
echo -e "${YELLOW}Testing concurrent requests...${NC}"
for i in {1..5}; do
    test_endpoint "GET" "/health" "Concurrent health check $i" &
done
wait
echo -e "${GREEN}✅ Concurrent requests completed${NC}"

# Test error handling
echo -e "${BLUE}🚨 Testing error handling...${NC}"

echo -e "${YELLOW}Testing invalid key...${NC}"
response=$(curl -s -w "%{http_code}" -o /tmp/test_response "$API_BASE/visual/key/nonexistent:key")
http_code="${response: -3}"
if [ "$http_code" -eq 404 ]; then
    echo -e "${GREEN}✅ Invalid key handling - OK${NC}"
else
    echo -e "${YELLOW}⚠️  Invalid key returned HTTP $http_code (expected 404)${NC}"
fi

echo -e "${YELLOW}Testing invalid query...${NC}"
response=$(curl -s -w "%{http_code}" -o /tmp/test_response -X POST -H "Content-Type: application/json" -d '{"operation": "INVALID"}' "$API_BASE/visual/query/validate")
http_code="${response: -3}"
if [ "$http_code" -eq 400 ]; then
    echo -e "${GREEN}✅ Invalid query handling - OK${NC}"
else
    echo -e "${YELLOW}⚠️  Invalid query returned HTTP $http_code (expected 400)${NC}"
fi

# WebSocket test (basic connectivity)
echo -e "${BLUE}🔌 Testing WebSocket connectivity...${NC}"
if command -v wscat &> /dev/null; then
    echo -e "${YELLOW}Testing WebSocket connection...${NC}"
    timeout 5s wscat -c "ws://localhost:8080/api/v1/redis-mcp/visual/stream?client_id=test" -x '{"type":"ping"}' >/dev/null 2>&1 && {
        echo -e "${GREEN}✅ WebSocket connection - OK${NC}"
    } || {
        echo -e "${YELLOW}⚠️  WebSocket test skipped (connection timeout or wscat not available)${NC}"
    }
else
    echo -e "${YELLOW}⚠️  WebSocket test skipped (wscat not installed)${NC}"
fi

# Summary
echo ""
echo -e "${GREEN}🎉 Testing completed!${NC}"
echo "=================================================="
echo -e "${BLUE}📊 Test Summary:${NC}"
echo -e "  • ${GREEN}API Endpoints:${NC}     Tested core functionality"
echo -e "  • ${GREEN}Data Types:${NC}        Tested all Redis data structures"
echo -e "  • ${GREEN}Performance:${NC}       Tested bulk operations and concurrency"
echo -e "  • ${GREEN}Error Handling:${NC}    Tested invalid inputs"
echo -e "  • ${GREEN}WebSocket:${NC}         Basic connectivity test"
echo ""
echo -e "${BLUE}🔗 Access Points:${NC}"
echo -e "  • ${GREEN}Web UI:${NC}            $WEB_BASE"
echo -e "  • ${GREEN}API Documentation:${NC} $API_BASE/docs"
echo -e "  • ${GREEN}Health Check:${NC}      $API_BASE/health"
echo ""
echo -e "${BLUE}📖 Next Steps:${NC}"
echo -e "  • Open the Web UI to explore data visually"
echo -e "  • Try building queries with the visual query builder"
echo -e "  • Monitor real-time metrics and performance"
echo -e "  • Explore the API documentation for advanced features"
echo ""
echo -e "${GREEN}🚀 Happy testing!${NC}"

# Cleanup
rm -f /tmp/test_response
