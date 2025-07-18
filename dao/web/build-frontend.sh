#!/bin/bash

# Frontend Build Script for Developer DAO Platform
set -e

echo "🎨 Building Developer DAO Frontend Applications"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Node.js is installed
check_node() {
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js 18+ and try again."
        exit 1
    fi
    
    NODE_VERSION=$(node --version | sed 's/v//')
    print_success "Node.js $NODE_VERSION is installed"
}

# Check if npm is installed
check_npm() {
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed. Please install npm and try again."
        exit 1
    fi
    
    NPM_VERSION=$(npm --version)
    print_success "npm $NPM_VERSION is installed"
}

# Install dependencies for shared package
install_shared_deps() {
    print_status "Installing shared package dependencies..."
    cd shared
    
    if [ ! -f "package.json" ]; then
        print_error "shared/package.json not found"
        exit 1
    fi
    
    npm install
    print_success "Shared package dependencies installed"
    cd ..
}

# Build shared package
build_shared() {
    print_status "Building shared package..."
    cd shared
    
    npm run build
    print_success "Shared package built successfully"
    cd ..
}

# Install dependencies for dao-portal
install_dao_portal_deps() {
    print_status "Installing dao-portal dependencies..."
    cd dao-portal
    
    if [ ! -f "package.json" ]; then
        print_error "dao-portal/package.json not found"
        exit 1
    fi
    
    npm install
    print_success "DAO Portal dependencies installed"
    cd ..
}

# Install dependencies for governance-ui
install_governance_ui_deps() {
    print_status "Installing governance-ui dependencies..."
    cd governance-ui
    
    if [ ! -f "package.json" ]; then
        print_error "governance-ui/package.json not found"
        exit 1
    fi
    
    npm install
    print_success "Governance UI dependencies installed"
    cd ..
}

# Start dao-portal in development mode
start_dao_portal() {
    print_status "Starting DAO Portal in development mode..."
    cd dao-portal
    
    # Create .env.local if it doesn't exist
    if [ ! -f ".env.local" ]; then
        print_status "Creating .env.local for DAO Portal..."
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_MARKETPLACE_API_URL=http://localhost:8081
NEXT_PUBLIC_METRICS_API_URL=http://localhost:8082
NEXT_PUBLIC_AI_API_URL=http://localhost:8083
NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID=your_wallet_connect_project_id
EOF
    fi
    
    print_success "Starting DAO Portal on http://localhost:3002"
    npm run dev -- --port 3002 &
    DAO_PORTAL_PID=$!
    echo $DAO_PORTAL_PID > ../logs/dao-portal.pid
    cd ..
}

# Start governance-ui in development mode
start_governance_ui() {
    print_status "Starting Governance UI in development mode..."
    cd governance-ui
    
    # Create .env.local if it doesn't exist
    if [ ! -f ".env.local" ]; then
        print_status "Creating .env.local for Governance UI..."
        cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8084
NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID=your_wallet_connect_project_id
EOF
    fi
    
    print_success "Starting Governance UI on http://localhost:3003"
    npm run dev -- --port 3003 &
    GOVERNANCE_UI_PID=$!
    echo $GOVERNANCE_UI_PID > ../logs/governance-ui.pid
    cd ..
}

# Stop frontend applications
stop_frontend() {
    print_status "Stopping frontend applications..."
    
    # Create logs directory if it doesn't exist
    mkdir -p logs
    
    # Stop DAO Portal
    if [ -f "logs/dao-portal.pid" ]; then
        PID=$(cat logs/dao-portal.pid)
        if kill -0 $PID 2>/dev/null; then
            print_status "Stopping DAO Portal (PID: $PID)..."
            kill $PID
            rm logs/dao-portal.pid
        else
            print_warning "DAO Portal was not running"
        fi
    fi
    
    # Stop Governance UI
    if [ -f "logs/governance-ui.pid" ]; then
        PID=$(cat logs/governance-ui.pid)
        if kill -0 $PID 2>/dev/null; then
            print_status "Stopping Governance UI (PID: $PID)..."
            kill $PID
            rm logs/governance-ui.pid
        else
            print_warning "Governance UI was not running"
        fi
    fi
    
    print_success "Frontend applications stopped"
}

# Show status
show_status() {
    echo ""
    echo "🎉 Developer DAO Frontend Status"
    echo "================================"
    echo ""
    echo "📱 Frontend Applications:"
    echo "   DAO Portal:         http://localhost:3002"
    echo "   Governance UI:      http://localhost:3003"
    echo ""
    echo "📊 Logs:"
    echo "   DAO Portal:         tail -f logs/dao-portal.log"
    echo "   Governance UI:      tail -f logs/governance-ui.log"
    echo ""
    echo "🛑 To stop frontend:"
    echo "   ./build-frontend.sh stop"
    echo ""
}

# Main function
main() {
    case "${1:-}" in
        "install")
            check_node
            check_npm
            install_shared_deps
            build_shared
            install_dao_portal_deps
            install_governance_ui_deps
            ;;
        "build")
            check_node
            check_npm
            build_shared
            ;;
        "start")
            check_node
            check_npm
            mkdir -p logs
            start_dao_portal
            sleep 3
            start_governance_ui
            sleep 3
            show_status
            ;;
        "stop")
            stop_frontend
            ;;
        "status")
            show_status
            ;;
        "dev")
            check_node
            check_npm
            install_shared_deps
            build_shared
            install_dao_portal_deps
            install_governance_ui_deps
            mkdir -p logs
            start_dao_portal
            sleep 3
            start_governance_ui
            sleep 3
            show_status
            ;;
        *)
            echo "Developer DAO Frontend Build Script"
            echo ""
            echo "Usage: $0 {install|build|start|stop|status|dev}"
            echo ""
            echo "Commands:"
            echo "  install - Install all dependencies"
            echo "  build   - Build shared package"
            echo "  start   - Start frontend applications"
            echo "  stop    - Stop frontend applications"
            echo "  status  - Show application status"
            echo "  dev     - Install, build, and start (full development setup)"
            echo ""
            echo "Examples:"
            echo "  $0 dev      # Full development setup"
            echo "  $0 start    # Start frontend applications"
            echo "  $0 stop     # Stop frontend applications"
            ;;
    esac
}

# Run main function
main "$@"
