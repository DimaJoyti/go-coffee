#!/bin/bash

# Build script for AI Arbitrage Service
set -e

echo "🚀 Building AI Arbitrage Service..."

# Change to project root
cd "$(dirname "$0")/.."

# Ensure bin directory exists
mkdir -p bin

# Build the AI Arbitrage service
echo "📦 Building ai-arbitrage-service..."
go build -o bin/ai-arbitrage-service cmd/ai-arbitrage-service/main.go

# Build the Market Data service
echo "📦 Building market-data-service..."
go build -o bin/market-data-service cmd/market-data-service/main.go

echo "✅ Build completed successfully!"

# Show built binaries
echo "📋 Built binaries:"
ls -la bin/

echo ""
echo "🎯 To run the services:"
echo "  AI Arbitrage Service: ./bin/ai-arbitrage-service"
echo "  Market Data Service:  ./bin/market-data-service"
echo ""
echo "🔧 Environment variables:"
echo "  GRPC_PORT=50054 (AI Arbitrage)"
echo "  GRPC_PORT=50055 (Market Data)"
echo "  REDIS_URL=redis://localhost:6379"
echo "  GEMINI_API_KEY=your-gemini-key"
echo "  OLLAMA_BASE_URL=http://localhost:11434"
