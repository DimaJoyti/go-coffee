#!/bin/bash

echo "🔧 Quick Fix Test"
echo "================="

# Test Go compilation
echo "📝 Testing Go compilation..."
cd web-ui
if go build simple-bright-data-test.go; then
    echo "✅ simple-bright-data-test.go compiles successfully"
    rm -f simple-bright-data-test
else
    echo "❌ simple-bright-data-test.go compilation failed"
fi

# Test MCP server compilation
echo ""
echo "📝 Testing MCP server compilation..."
cd mcp-server
if go build main.go; then
    echo "✅ MCP server compiles successfully"
    rm -f main
else
    echo "❌ MCP server compilation failed"
fi
cd ..

# Test backend compilation
echo ""
echo "📝 Testing backend compilation..."
cd backend
if go build cmd/web-ui-service/main.go; then
    echo "✅ Backend compiles successfully"
    rm -f main
else
    echo "❌ Backend compilation failed"
fi
cd ..

echo ""
echo "🎉 All compilation tests completed!"
echo ""
echo "🚀 To start all services:"
echo "   ./start-all.sh"
echo ""
echo "🧪 To run integration test:"
echo "   ./test-integration.sh"
