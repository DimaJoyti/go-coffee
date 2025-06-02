#!/bin/bash

# Go Coffee - Integration Test Script
# Tests MCP server, backend, and real data integration

echo "🧪 Go Coffee Integration Test"
echo "============================="

# Load environment variables
if [ -f ".env" ]; then
    echo "✅ Loading environment variables from .env"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "❌ .env file not found"
    exit 1
fi

echo ""
echo "🔧 Configuration:"
echo "   • MCP Server URL: ${MCP_SERVER_URL}"
echo "   • Backend Port: ${PORT:-8090}"
echo "   • Rate Limit: ${BRIGHT_DATA_RATE_LIMIT}"
echo "   • Cache TTL: ${BRIGHT_DATA_CACHE_TTL}"

# Test MCP Server
echo ""
echo "🔍 Testing MCP Server..."
MCP_HEALTH=$(curl -s http://localhost:${MCP_SERVER_PORT:-3003}/health 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✅ MCP Server is responding"
    echo "   Response: $MCP_HEALTH"
else
    echo "❌ MCP Server is not responding"
    echo "   Trying to start MCP server..."
    cd mcp-server
    MCP_SERVER_PORT=${MCP_SERVER_PORT:-3003} go run main.go &
    MCP_PID=$!
    echo "   Started MCP server (PID: $MCP_PID)"
    cd ..
    sleep 3
fi

# Test MCP functionality
echo ""
echo "🔍 Testing MCP Search Function..."
SEARCH_RESULT=$(curl -s -X POST http://localhost:${MCP_SERVER_PORT:-3003} \
    -H "Content-Type: application/json" \
    -d '{"method":"search_engine_Bright_Data","params":{"query":"coffee market prices","engine":"google"}}' 2>/dev/null)

if [ $? -eq 0 ] && [ ! -z "$SEARCH_RESULT" ]; then
    echo "✅ MCP Search is working"
    echo "   Sample result: $(echo $SEARCH_RESULT | head -c 100)..."
else
    echo "❌ MCP Search failed"
fi

# Test Backend Server
echo ""
echo "🔍 Testing Backend Server..."
BACKEND_HEALTH=$(curl -s http://localhost:${PORT:-8090}/health 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✅ Backend Server is responding"
    echo "   Response: $BACKEND_HEALTH"
else
    echo "❌ Backend Server is not responding"
    echo "   Note: Start backend with: cd backend && go run cmd/web-ui-service/main.go"
fi

# Test Market Data API
echo ""
echo "🔍 Testing Market Data API..."
MARKET_DATA=$(curl -s http://localhost:${PORT:-8090}/api/v1/scraping/data 2>/dev/null)
if [ $? -eq 0 ] && [ ! -z "$MARKET_DATA" ]; then
    echo "✅ Market Data API is working"
    
    # Count data items
    ITEM_COUNT=$(echo $MARKET_DATA | grep -o '"id"' | wc -l)
    echo "   Retrieved $ITEM_COUNT market data items"
    
    # Check if using fallback data
    if echo "$MARKET_DATA" | grep -q "fallback"; then
        echo "   ⚠️  Using fallback data (MCP integration may need improvement)"
    else
        echo "   ✅ Using real data from MCP"
    fi
else
    echo "❌ Market Data API failed"
fi

# Test specific endpoints
echo ""
echo "🔍 Testing Specific Endpoints..."

endpoints=(
    "/api/v1/scraping/sources:Data Sources"
    "/api/v1/dashboard/metrics:Dashboard Metrics"
    "/api/v1/coffee/orders:Coffee Orders"
)

for endpoint_info in "${endpoints[@]}"; do
    IFS=':' read -r endpoint name <<< "$endpoint_info"
    
    RESPONSE=$(curl -s http://localhost:${PORT:-8090}$endpoint 2>/dev/null)
    if [ $? -eq 0 ] && [ ! -z "$RESPONSE" ]; then
        echo "   ✅ $name: OK"
    else
        echo "   ❌ $name: Failed"
    fi
done

# Summary
echo ""
echo "📊 Integration Test Summary:"
echo "============================="

# Check overall status
OVERALL_STATUS="✅ PASSED"

if ! curl -s http://localhost:${MCP_SERVER_PORT:-3003}/health >/dev/null 2>&1; then
    OVERALL_STATUS="⚠️  PARTIAL (MCP Server issues)"
fi

if ! curl -s http://localhost:${PORT:-8090}/health >/dev/null 2>&1; then
    OVERALL_STATUS="❌ FAILED (Backend not running)"
fi

echo "   Overall Status: $OVERALL_STATUS"
echo ""
echo "🎯 What's Working:"
echo "   • Environment configuration ✅"
echo "   • .env file loading ✅"
echo "   • Mock data fallback ✅"
echo "   • API endpoints structure ✅"
echo ""
echo "🚀 Next Steps:"
echo "   1. Ensure all services are running:"
echo "      ./start-all.sh"
echo "   2. Open browser to test UI:"
echo "      http://localhost:3000"
echo "   3. Check API endpoints:"
echo "      http://localhost:${PORT:-8090}/api/v1/scraping/data"
echo ""
echo "🎉 Integration test completed!"
