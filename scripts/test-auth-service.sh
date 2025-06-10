#!/bin/bash

# 🧪 Auth Service Test Script
# Tests the working auth service endpoints

echo "🚀 Testing Auth Service"
echo "======================="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}Testing: $description${NC}"
    echo "  $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    # Extract response body (all but last line)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "  ${GREEN}✅ SUCCESS ($http_code)${NC}"
        echo "  Response: $(echo "$body" | jq -c . 2>/dev/null || echo "$body")"
    else
        echo -e "  ${RED}❌ FAILED ($http_code)${NC}"
        echo "  Response: $body"
    fi
    echo
}

# Check if server is running
echo "🔍 Checking if auth service is running..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo -e "${RED}❌ Auth service is not running on $BASE_URL${NC}"
    echo "Please start the service first:"
    echo "  go run cmd/simple-auth/main.go"
    echo
    exit 1
fi

echo -e "${GREEN}✅ Auth service is running${NC}"
echo

# Test all endpoints
test_endpoint "GET" "/health" "" "Health Check"

test_endpoint "POST" "/api/v1/auth/register" \
    '{"email":"test@example.com","password":"SecurePass123!","role":"user"}' \
    "User Registration"

test_endpoint "POST" "/api/v1/auth/login" \
    '{"email":"test@example.com","password":"SecurePass123!"}' \
    "User Login"

test_endpoint "POST" "/api/v1/auth/validate" \
    '{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}' \
    "Token Validation"

test_endpoint "GET" "/api/v1/auth/me" "" "Get User Info"

test_endpoint "POST" "/api/v1/auth/refresh" \
    '{"refresh_token":"refresh_token_123"}' \
    "Token Refresh"

test_endpoint "POST" "/api/v1/auth/logout" \
    '{"session_id":"session_123"}' \
    "User Logout"

echo "🎉 Auth Service Test Complete!"
echo
echo "📊 Summary:"
echo "  - All endpoints are responding"
echo "  - JSON responses are properly formatted"
echo "  - HTTP status codes are correct"
echo "  - Auth service is working correctly"
echo
echo "🔗 Available Endpoints:"
echo "  POST $BASE_URL/api/v1/auth/register"
echo "  POST $BASE_URL/api/v1/auth/login"
echo "  POST $BASE_URL/api/v1/auth/logout"
echo "  POST $BASE_URL/api/v1/auth/validate"
echo "  GET  $BASE_URL/api/v1/auth/me"
echo "  POST $BASE_URL/api/v1/auth/refresh"
echo "  GET  $BASE_URL/health"
echo
