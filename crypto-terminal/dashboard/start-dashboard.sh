#!/bin/bash

# Coffee Trading Dashboard Startup Script
echo "☕ Starting Coffee Trading Dashboard..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ package.json not found. Please run this script from the nextjs-dashboard directory."
    exit 1
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
    if [ $? -ne 0 ]; then
        echo "❌ Failed to install dependencies."
        exit 1
    fi
fi

# Check if .env.local exists
if [ ! -f ".env.local" ]; then
    echo "⚠️  .env.local not found. Creating default configuration..."
    cat > .env.local << EOF
# Next.js Trading Dashboard Environment Configuration
NEXT_PUBLIC_API_URL=http://localhost:8090
NEXT_PUBLIC_WS_URL=ws://localhost:8090
NEXT_PUBLIC_APP_NAME="Coffee Trading Dashboard"
NEXT_PUBLIC_APP_VERSION="1.0.0"
NODE_ENV=development
NEXT_PUBLIC_ENVIRONMENT=development
NEXT_PUBLIC_ENABLE_DEVTOOLS=true
NEXT_PUBLIC_ENABLE_WEBSOCKET=true
NEXT_PUBLIC_ENABLE_NOTIFICATIONS=true
NEXT_PUBLIC_DEFAULT_PORTFOLIO_ID=""
NEXT_PUBLIC_REFRESH_INTERVAL=30000
NEXT_PUBLIC_WEBSOCKET_RECONNECT_ATTEMPTS=5
NEXT_PUBLIC_WEBSOCKET_RECONNECT_DELAY=1000
EOF
    echo "✅ Created .env.local with default settings"
fi

# Check if Go backend is running
echo "🔍 Checking if Go Coffee Trading backend is running..."
if curl -s http://localhost:8090/health > /dev/null; then
    echo "✅ Go backend is running on port 8090"
else
    echo "⚠️  Go backend is not running on port 8090"
    echo "   Please start the Go Coffee Trading backend first:"
    echo "   cd ../cmd/coffee-trading && go run main.go"
    echo ""
    echo "   Continuing anyway... Dashboard will show connection errors until backend is started."
fi

# Start the development server
echo "🚀 Starting Next.js development server on port 3001..."
echo "📱 Dashboard will be available at: http://localhost:3001"
echo "🔄 WebSocket will connect to: ws://localhost:8090"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

npm run dev
