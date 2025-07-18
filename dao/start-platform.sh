#!/bin/bash

# Developer DAO Platform - Complete Startup Script
set -e

echo "🚀 Developer DAO Platform - Complete Startup"
echo "============================================="

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

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23+ and try again."
        exit 1
    fi
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js 18+ and try again."
        exit 1
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed. Please install npm and try again."
        exit 1
    fi
    
    print_success "All prerequisites are installed"
}

# Setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    # Create .env file if it doesn't exist
    if [ ! -f ".env" ]; then
        print_warning ".env file not found. Creating from .env.example..."
        cp .env.example .env
        print_warning "Please edit .env file with your API keys before running the platform"
    fi
    
    # Create logs directory
    mkdir -p logs
    
    print_success "Environment setup complete"
}

# Build backend services
build_backend() {
    print_status "Building backend services..."
    
    # Use the existing run-local.sh script
    if [ -f "run-local.sh" ]; then
        ./run-local.sh build
    else
        print_error "run-local.sh not found. Please ensure you're in the dao directory."
        exit 1
    fi
    
    print_success "Backend services built"
}

# Start backend services
start_backend() {
    print_status "Starting backend services..."
    
    # Start backend services in background
    ./run-local.sh start > logs/backend-startup.log 2>&1 &
    
    # Wait for services to start
    print_status "Waiting for backend services to start..."
    sleep 10
    
    # Check if services are running
    services_running=0
    for port in 8080 8081 8082 8084; do
        if curl -f -s http://localhost:$port/health > /dev/null 2>&1; then
            services_running=$((services_running + 1))
        fi
    done
    
    if [ $services_running -eq 4 ]; then
        print_success "All backend services are running"
    else
        print_warning "$services_running/4 backend services are running"
        print_warning "Check logs/backend-startup.log for details"
    fi
}

# Setup frontend
setup_frontend() {
    print_status "Setting up frontend applications..."
    
    cd web
    
    # Use the frontend build script
    if [ -f "build-frontend.sh" ]; then
        ./build-frontend.sh install
    else
        print_error "build-frontend.sh not found in web directory"
        exit 1
    fi
    
    cd ..
    print_success "Frontend setup complete"
}

# Start frontend
start_frontend() {
    print_status "Starting frontend applications..."
    
    cd web
    ./build-frontend.sh start > ../logs/frontend-startup.log 2>&1 &
    cd ..
    
    # Wait for frontend to start
    print_status "Waiting for frontend applications to start..."
    sleep 15
    
    # Check if frontend is running
    frontend_running=0
    if curl -f -s http://localhost:3002 > /dev/null 2>&1; then
        frontend_running=$((frontend_running + 1))
    fi
    if curl -f -s http://localhost:3003 > /dev/null 2>&1; then
        frontend_running=$((frontend_running + 1))
    fi
    
    if [ $frontend_running -eq 2 ]; then
        print_success "All frontend applications are running"
    else
        print_warning "$frontend_running/2 frontend applications are running"
        print_warning "Check logs/frontend-startup.log for details"
    fi
}

# Health check
health_check() {
    print_status "Performing comprehensive health check..."
    
    echo ""
    echo "🔍 Backend Services:"
    
    # Check backend services
    services=("8080:Bounty Service" "8081:Marketplace Service" "8082:Metrics Service" "8084:DAO Governance")
    
    for service in "${services[@]}"; do
        port=$(echo $service | cut -d: -f1)
        name=$(echo $service | cut -d: -f2)
        
        if curl -f -s http://localhost:$port/health > /dev/null 2>&1; then
            print_success "$name is healthy (port $port)"
        else
            print_warning "$name health check failed (port $port)"
        fi
    done
    
    echo ""
    echo "🎨 Frontend Applications:"
    
    # Check frontend applications
    if curl -f -s http://localhost:3002 > /dev/null 2>&1; then
        print_success "DAO Portal is accessible (port 3002)"
    else
        print_warning "DAO Portal health check failed (port 3002)"
    fi

    if curl -f -s http://localhost:3003 > /dev/null 2>&1; then
        print_success "Governance UI is accessible (port 3003)"
    else
        print_warning "Governance UI health check failed (port 3003)"
    fi
}

# Show platform URLs
show_platform_urls() {
    echo ""
    echo "🎉 Developer DAO Platform is Running!"
    echo "====================================="
    echo ""
    echo "📱 Frontend Applications:"
    echo "   DAO Portal:          http://localhost:3002"
    echo "   Governance UI:       http://localhost:3003"
    echo ""
    echo "🔧 Backend Services:"
    echo "   Bounty Service:      http://localhost:8080"
    echo "   Marketplace Service: http://localhost:8081"
    echo "   Metrics Service:     http://localhost:8082"
    echo "   DAO Governance:      http://localhost:8084"
    echo ""
    echo "📚 API Documentation:"
    echo "   Swagger UI:          http://localhost:8080/swagger/index.html"
    echo ""
    echo "📊 Logs:"
    echo "   Backend Services:    tail -f logs/bounty-service.log"
    echo "   Frontend Apps:       tail -f logs/frontend-startup.log"
    echo ""
    echo "🛑 To stop the platform:"
    echo "   ./start-platform.sh stop"
    echo ""
}

# Stop platform
stop_platform() {
    print_status "Stopping Developer DAO Platform..."
    
    # Stop backend services
    if [ -f "run-local.sh" ]; then
        ./run-local.sh stop
    fi
    
    # Stop frontend applications
    cd web
    if [ -f "build-frontend.sh" ]; then
        ./build-frontend.sh stop
    fi
    cd ..
    
    print_success "Developer DAO Platform stopped"
}

# Main function
main() {
    case "${1:-}" in
        "start")
            check_prerequisites
            setup_environment
            build_backend
            start_backend
            setup_frontend
            start_frontend
            health_check
            show_platform_urls
            ;;
        "stop")
            stop_platform
            ;;
        "status")
            health_check
            show_platform_urls
            ;;
        "backend")
            check_prerequisites
            setup_environment
            build_backend
            start_backend
            ;;
        "frontend")
            check_prerequisites
            setup_frontend
            start_frontend
            ;;
        "health")
            health_check
            ;;
        *)
            echo "Developer DAO Platform - Complete Startup Script"
            echo ""
            echo "Usage: $0 {start|stop|status|backend|frontend|health}"
            echo ""
            echo "Commands:"
            echo "  start    - Start complete platform (backend + frontend)"
            echo "  stop     - Stop complete platform"
            echo "  status   - Show platform status and URLs"
            echo "  backend  - Start only backend services"
            echo "  frontend - Start only frontend applications"
            echo "  health   - Perform health checks"
            echo ""
            echo "Examples:"
            echo "  $0 start     # Start complete platform"
            echo "  $0 backend   # Start only backend services"
            echo "  $0 frontend  # Start only frontend applications"
            echo "  $0 stop      # Stop everything"
            ;;
    esac
}

# Run main function
main "$@"
