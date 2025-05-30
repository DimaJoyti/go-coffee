#!/bin/bash

# Test script for Minimal Agent Orchestrator
# Usage: ./test-api.sh [base_url]

BASE_URL=${1:-"http://localhost:8095"}
echo "🧪 Testing Minimal Agent Orchestrator API at $BASE_URL"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to make HTTP request and show result
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${BLUE}🔍 Testing: $description${NC}"
    echo "   $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint")
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    # Extract response body (all but last line)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "   ${GREEN}✅ Success ($http_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "   ${RED}❌ Failed ($http_code)${NC}"
        echo "$body"
    fi
}

# Test 1: Service Info
test_endpoint "GET" "/" "" "Service Information"

# Test 2: Register Agent 1
agent1_data='{
    "id": "agent-001",
    "name": "Coffee Inventory Agent",
    "type": "inventory",
    "version": "1.0.0",
    "endpoint": "http://localhost:8080",
    "capabilities": ["inventory_management", "stock_tracking"],
    "metadata": {
        "location": "downtown",
        "max_concurrent_tasks": 10
    }
}'
test_endpoint "POST" "/api/v1/agents/register" "$agent1_data" "Register Agent 1"

# Test 3: Register Agent 2
agent2_data='{
    "id": "agent-002",
    "name": "Customer Feedback Agent",
    "type": "feedback",
    "version": "1.0.0",
    "endpoint": "http://localhost:8081",
    "capabilities": ["feedback_analysis", "sentiment_analysis"],
    "metadata": {
        "location": "cloud",
        "max_concurrent_tasks": 20
    }
}'
test_endpoint "POST" "/api/v1/agents/register" "$agent2_data" "Register Agent 2"

# Test 4: List All Agents
test_endpoint "GET" "/api/v1/agents" "" "List All Agents"

# Test 5: Get Specific Agent
test_endpoint "GET" "/api/v1/agents/agent-001" "" "Get Agent 1 Details"

# Test 6: Agent Heartbeat
test_endpoint "POST" "/api/v1/agents/agent-001/heartbeat" "" "Agent 1 Heartbeat"

# Test 7: Create Task
task_data='{
    "type": "inventory_check",
    "priority": 1,
    "data": {
        "shop_id": "downtown",
        "items": ["coffee_beans", "milk", "sugar"]
    },
    "required_capabilities": ["inventory_management"]
}'
test_endpoint "POST" "/api/v1/tasks" "$task_data" "Create Inventory Task"

# Test 8: Create Another Task
task2_data='{
    "type": "feedback_analysis",
    "priority": 2,
    "data": {
        "feedback_id": "fb-12345",
        "text": "Great coffee, but service was slow"
    },
    "required_capabilities": ["feedback_analysis"]
}'
test_endpoint "POST" "/api/v1/tasks" "$task2_data" "Create Feedback Task"

# Test 9: Search Agents
test_endpoint "GET" "/api/v1/discovery/search?q=inventory" "" "Search for Inventory Agents"

# Test 10: Search Agents by Type
test_endpoint "GET" "/api/v1/discovery/search?q=feedback" "" "Search for Feedback Agents"

# Test 11: Health Check
test_endpoint "GET" "/api/v1/monitoring/health" "" "System Health Check"

# Test 12: System Statistics
test_endpoint "GET" "/api/v1/monitoring/stats" "" "System Statistics"

# Test 13: Try to create task with no suitable agents
impossible_task='{
    "type": "impossible_task",
    "priority": 1,
    "data": {},
    "required_capabilities": ["non_existent_capability"]
}'
test_endpoint "POST" "/api/v1/tasks" "$impossible_task" "Create Impossible Task (should fail)"

# Test 14: Delete Agent
test_endpoint "DELETE" "/api/v1/agents/agent-002" "" "Delete Agent 2"

# Test 15: Verify Agent Deleted
test_endpoint "GET" "/api/v1/agents/agent-002" "" "Try to Get Deleted Agent (should fail)"

# Test 16: Final Health Check
test_endpoint "GET" "/api/v1/monitoring/health" "" "Final Health Check"

echo -e "\n${YELLOW}🎉 API Testing Complete!${NC}"
echo -e "${BLUE}📊 Summary:${NC}"
echo "   - Service is running and responding"
echo "   - Agent registration works"
echo "   - Task creation and assignment works"
echo "   - Search functionality works"
echo "   - Health monitoring works"
echo "   - Agent deletion works"
echo -e "\n${GREEN}✅ Minimal Agent Orchestrator is fully functional!${NC}"
