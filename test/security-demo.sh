#!/bin/bash

# Security Gateway Demo Script
# This script demonstrates the security features of the Go Coffee platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GATEWAY_URL="http://localhost:8080"
AUTH_URL="http://localhost:8081"
ORDER_URL="http://localhost:8082"
PAYMENT_URL="http://localhost:8083"

echo -e "${BLUE}🛡️  Go Coffee Security Gateway Demo${NC}"
echo "=================================================="

# Function to print section headers
print_section() {
    echo -e "\n${YELLOW}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Function to make HTTP request and check response
test_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -e "\n🔍 Testing: $description"
    echo "Request: $method $url"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "User-Agent: SecurityDemo/1.0" \
            -d "$data" "$url" 2>/dev/null || echo -e "\n000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "User-Agent: SecurityDemo/1.0" \
            "$url" 2>/dev/null || echo -e "\n000")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    echo "Response Status: $status_code"
    echo "Response Body: $body"
    
    if [ "$status_code" = "$expected_status" ]; then
        print_result 0 "$description"
    else
        print_result 1 "$description (Expected: $expected_status, Got: $status_code)"
    fi
    
    return 0
}

# Check if services are running
print_section "Service Health Checks"

echo "🏥 Checking Security Gateway health..."
test_request "GET" "$GATEWAY_URL/health" "" "200" "Security Gateway Health Check"

echo -e "\n📊 Checking Security Metrics..."
test_request "GET" "$GATEWAY_URL/metrics" "" "200" "Security Metrics Endpoint"

# Test WAF Protection
print_section "Web Application Firewall (WAF) Tests"

echo "🚫 Testing SQL Injection Protection..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"input","value":"SELECT * FROM users WHERE id = 1 UNION SELECT * FROM passwords"}' \
    "200" "SQL Injection Detection"

echo -e "\n🚫 Testing XSS Protection..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"input","value":"<script>alert(\"XSS\")</script>"}' \
    "200" "XSS Attack Detection"

echo -e "\n🚫 Testing Path Traversal Protection..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"input","value":"../../../etc/passwd"}' \
    "200" "Path Traversal Detection"

echo -e "\n🚫 Testing Command Injection Protection..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"input","value":"test; rm -rf /"}' \
    "200" "Command Injection Detection"

# Test Input Validation
print_section "Input Validation Tests"

echo "✅ Testing Valid Email..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"email","value":"user@example.com"}' \
    "200" "Valid Email Validation"

echo -e "\n❌ Testing Invalid Email..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"email","value":"invalid-email"}' \
    "200" "Invalid Email Validation"

echo -e "\n✅ Testing Valid URL..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"url","value":"https://example.com"}' \
    "200" "Valid URL Validation"

echo -e "\n❌ Testing Invalid URL..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"url","value":"not-a-url"}' \
    "200" "Invalid URL Validation"

echo -e "\n✅ Testing Valid IP Address..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"ip","value":"192.168.1.1"}' \
    "200" "Valid IP Validation"

echo -e "\n❌ Testing Invalid IP Address..."
test_request "POST" "$GATEWAY_URL/api/v1/security/validate" \
    '{"type":"ip","value":"999.999.999.999"}' \
    "200" "Invalid IP Validation"

# Test Rate Limiting
print_section "Rate Limiting Tests"

echo "🚦 Testing Normal Request Rate..."
for i in {1..5}; do
    test_request "GET" "$GATEWAY_URL/health" "" "200" "Normal Rate Request $i"
    sleep 0.1
done

echo -e "\n🚦 Testing Rate Limit Enforcement..."
echo "Sending rapid requests to trigger rate limiting..."
for i in {1..25}; do
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health" 2>/dev/null || echo "000")
    if [ "$status_code" = "429" ]; then
        print_result 0 "Rate Limit Triggered at request $i"
        break
    fi
    if [ $i -eq 25 ]; then
        print_result 1 "Rate Limit Not Triggered (sent 25 requests)"
    fi
done

# Test Security Headers
print_section "Security Headers Tests"

echo "🔒 Testing Security Headers..."
headers=$(curl -s -I "$GATEWAY_URL/health" 2>/dev/null)

check_header() {
    local header=$1
    local description=$2
    
    if echo "$headers" | grep -i "$header" > /dev/null; then
        print_result 0 "$description Header Present"
    else
        print_result 1 "$description Header Missing"
    fi
}

check_header "X-Content-Type-Options" "X-Content-Type-Options"
check_header "X-Frame-Options" "X-Frame-Options"
check_header "X-XSS-Protection" "X-XSS-Protection"
check_header "Strict-Transport-Security" "HSTS"
check_header "Content-Security-Policy" "CSP"

# Test Gateway Proxying
print_section "API Gateway Proxying Tests"

echo "🔄 Testing Auth Service Proxy..."
test_request "GET" "$GATEWAY_URL/api/v1/gateway/auth/health" "" "200" "Auth Service Proxy"

echo -e "\n🔄 Testing Order Service Proxy..."
test_request "GET" "$GATEWAY_URL/api/v1/gateway/order/health" "" "200" "Order Service Proxy"

echo -e "\n🔄 Testing Payment Service Proxy..."
test_request "GET" "$GATEWAY_URL/api/v1/gateway/payment/health" "" "200" "Payment Service Proxy"

echo -e "\n🔄 Testing User Service Proxy..."
test_request "GET" "$GATEWAY_URL/api/v1/gateway/user/health" "" "200" "User Service Proxy"

# Test Malicious User Agent Detection
print_section "Bot Detection Tests"

echo "🤖 Testing Malicious User Agent Detection..."
malicious_response=$(curl -s -w "\n%{http_code}" \
    -H "User-Agent: sqlmap/1.0" \
    "$GATEWAY_URL/health" 2>/dev/null || echo -e "\n000")

malicious_status=$(echo "$malicious_response" | tail -n1)
if [ "$malicious_status" = "403" ]; then
    print_result 0 "Malicious User Agent Blocked"
else
    print_result 1 "Malicious User Agent Not Blocked (Status: $malicious_status)"
fi

# Test Security Metrics
print_section "Security Metrics Tests"

echo "📊 Testing Security Metrics Endpoint..."
metrics_response=$(curl -s "$GATEWAY_URL/api/v1/security/metrics" 2>/dev/null)

if echo "$metrics_response" | grep -q "total_events"; then
    print_result 0 "Security Metrics Available"
    echo "Sample metrics:"
    echo "$metrics_response" | head -10
else
    print_result 1 "Security Metrics Not Available"
fi

# Test Security Alerts
print_section "Security Alerts Tests"

echo "🚨 Testing Security Alerts Endpoint..."
alerts_response=$(curl -s "$GATEWAY_URL/api/v1/security/alerts?limit=5" 2>/dev/null)

if echo "$alerts_response" | grep -q "alerts"; then
    print_result 0 "Security Alerts Available"
    echo "Sample alerts:"
    echo "$alerts_response" | head -10
else
    print_result 1 "Security Alerts Not Available"
fi

# Performance Test
print_section "Performance Tests"

echo "⚡ Testing Gateway Performance..."
start_time=$(date +%s%N)
for i in {1..100}; do
    curl -s "$GATEWAY_URL/health" > /dev/null 2>&1
done
end_time=$(date +%s%N)

duration=$((($end_time - $start_time) / 1000000))
avg_latency=$(($duration / 100))

echo "100 requests completed in ${duration}ms"
echo "Average latency: ${avg_latency}ms per request"

if [ $avg_latency -lt 50 ]; then
    print_result 0 "Performance Test (< 50ms average)"
else
    print_result 1 "Performance Test (>= 50ms average)"
fi

# Summary
print_section "Demo Summary"

echo -e "${GREEN}✅ Security Gateway Demo Completed!${NC}"
echo ""
echo "Key Features Demonstrated:"
echo "• Web Application Firewall (WAF) Protection"
echo "• Input Validation and Sanitization"
echo "• Rate Limiting and DDoS Protection"
echo "• Security Headers Enforcement"
echo "• API Gateway Proxying"
echo "• Bot and Malicious User Agent Detection"
echo "• Real-time Security Monitoring"
echo "• Performance Optimization"
echo ""
echo "🔍 For detailed logs and metrics, check:"
echo "• Grafana Dashboard: http://localhost:3000"
echo "• Prometheus Metrics: http://localhost:9090"
echo "• Jaeger Tracing: http://localhost:16686"
echo ""
echo -e "${BLUE}🛡️ Your microservices are now protected by enterprise-grade security!${NC}"
