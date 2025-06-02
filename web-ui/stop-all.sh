#!/bin/bash

# Go Coffee - Stop All Services Script

echo "🛑 Stopping Go Coffee Services"
echo "==============================="

# Function to stop process by PID file
stop_service() {
    local service_name=$1
    local pid_file=$2
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            echo "🛑 Stopping $service_name (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                echo "⚠️  Force killing $service_name..."
                kill -9 "$pid" 2>/dev/null || true
            fi
            echo "✅ $service_name stopped"
        else
            echo "⚠️  $service_name was not running"
        fi
        rm -f "$pid_file"
    else
        echo "⚠️  No PID file found for $service_name"
    fi
}

# Stop services
stop_service "Frontend" ".frontend.pid"
stop_service "Backend" ".backend.pid"
stop_service "MCP Server" ".mcp.pid"

# Kill any remaining processes on our ports
echo ""
echo "🔍 Checking for remaining processes..."

# Check and kill processes on specific ports
for port in 3000 3001 8090; do
    pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pid" ]; then
        echo "🛑 Killing process on port $port (PID: $pid)"
        kill -9 $pid 2>/dev/null || true
    fi
done

# Clean up any remaining Go processes related to our project
pkill -f "go run.*web-ui-service" 2>/dev/null || true
pkill -f "go run.*mcp-server" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true

echo ""
echo "✅ All Go Coffee services have been stopped"
echo "🧹 Cleanup completed"
