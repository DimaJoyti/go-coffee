#!/bin/bash

# Go Coffee - Start All Services Script
# This script starts MCP server, backend, and frontend

set -e

echo "🚀 Starting Go Coffee with Bright Data MCP Integration"
echo "======================================================"

# Load environment variables
if [ -f ".env" ]; then
    echo "✅ Loading environment variables from .env"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  .env file not found, using defaults"
fi

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "❌ Port $port is already in use"
        return 1
    else
        echo "✅ Port $port is available"
        return 0
    fi
}

# Check required ports
echo ""
echo "🔍 Checking required ports..."
check_port ${MCP_SERVER_PORT:-3001} || exit 1
check_port ${PORT:-8090} || exit 1
check_port 3000 || exit 1

# Start MCP Server
echo ""
echo "🔧 Starting MCP Server on port ${MCP_SERVER_PORT:-3001}..."
cd mcp-server
go run main.go &
MCP_PID=$!
echo "✅ MCP Server started (PID: $MCP_PID)"
cd ..

# Wait for MCP server to start
echo "⏳ Waiting for MCP server to be ready..."
sleep 3

# Test MCP server
if curl -s http://localhost:${MCP_SERVER_PORT:-3001}/health > /dev/null; then
    echo "✅ MCP Server is ready"
else
    echo "❌ MCP Server failed to start"
    kill $MCP_PID 2>/dev/null || true
    exit 1
fi

# Start Backend
echo ""
echo "🔧 Starting Backend Server on port ${PORT:-8090}..."
cd backend
go run cmd/web-ui-service/main.go &
BACKEND_PID=$!
echo "✅ Backend Server started (PID: $BACKEND_PID)"
cd ..

# Wait for backend to start
echo "⏳ Waiting for backend server to be ready..."
sleep 5

# Test backend server
if curl -s http://localhost:${PORT:-8090}/health > /dev/null; then
    echo "✅ Backend Server is ready"
else
    echo "❌ Backend Server failed to start"
    kill $MCP_PID $BACKEND_PID 2>/dev/null || true
    exit 1
fi

# Start Frontend
echo ""
echo "🔧 Starting Frontend Server on port 3000..."
cd frontend
npm run dev &
FRONTEND_PID=$!
echo "✅ Frontend Server started (PID: $FRONTEND_PID)"
cd ..

# Wait for frontend to start
echo "⏳ Waiting for frontend server to be ready..."
sleep 10

echo ""
echo "🎉 All services started successfully!"
echo ""
echo "📊 Service URLs:"
echo "   • MCP Server:    http://localhost:${MCP_SERVER_PORT:-3001}/health"
echo "   • Backend API:   http://localhost:${PORT:-8090}/health"
echo "   • Frontend UI:   http://localhost:3000"
echo "   • Market Data:   http://localhost:${PORT:-8090}/api/v1/scraping/data"
echo ""
echo "🔧 Process IDs:"
echo "   • MCP Server: $MCP_PID"
echo "   • Backend:    $BACKEND_PID"
echo "   • Frontend:   $FRONTEND_PID"
echo ""
echo "🛑 To stop all services, run: ./stop-all.sh"
echo ""
echo "📝 Logs:"
echo "   • Press Ctrl+C to stop all services"
echo "   • Or run: kill $MCP_PID $BACKEND_PID $FRONTEND_PID"

# Save PIDs for stop script
echo "$MCP_PID" > .mcp.pid
echo "$BACKEND_PID" > .backend.pid
echo "$FRONTEND_PID" > .frontend.pid

# Wait for user interrupt
trap 'echo ""; echo "🛑 Stopping all services..."; kill $MCP_PID $BACKEND_PID $FRONTEND_PID 2>/dev/null || true; rm -f .*.pid; echo "✅ All services stopped"; exit 0' INT

echo "✨ All services are running. Press Ctrl+C to stop."
wait
