#!/bin/bash

# Go Coffee Epic UI - Start Script

echo "🚀 Starting Go Coffee Epic UI..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install docker-compose first."
    exit 1
fi

# Create necessary directories
mkdir -p frontend/node_modules
mkdir -p backend/bin

echo "📦 Building and starting services..."

# Start services with docker-compose
docker-compose -f docker-compose.ui.yml up --build -d

echo "⏳ Waiting for services to start..."
sleep 10

# Check if services are running
if docker-compose -f docker-compose.ui.yml ps | grep -q "Up"; then
    echo "✅ Services started successfully!"
    echo ""
    echo "🌐 Frontend: http://localhost:3000"
    echo "🔗 Backend API: http://localhost:8090"
    echo "❤️  Health Check: http://localhost:8090/health"
    echo "🔌 WebSocket: ws://localhost:8090/ws/realtime"
    echo ""
    echo "📊 Dashboard sections available:"
    echo "   • Dashboard Overview"
    echo "   • Coffee Orders & Inventory"
    echo "   • DeFi Portfolio & Trading"
    echo "   • AI Agents Monitoring"
    echo "   • Market Data & Analytics (Bright Data)"
    echo "   • Reports & Analytics"
    echo ""
    echo "🛑 To stop: ./stop.sh or docker-compose -f docker-compose.ui.yml down"
else
    echo "❌ Failed to start services. Check logs with:"
    echo "   docker-compose -f docker-compose.ui.yml logs"
fi
