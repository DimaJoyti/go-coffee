#!/bin/bash

# Crypto Market Terminal Startup Script
# This script helps you get the crypto terminal up and running quickly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ASCII Art Banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔═══════════════════════════════════════════════════════════════╗
    ║                                                               ║
    ║   ██████╗██████╗ ██╗   ██╗██████╗ ████████╗ ██████╗          ║
    ║  ██╔════╝██╔══██╗╚██╗ ██╔╝██╔══██╗╚══██╔══╝██╔═══██╗         ║
    ║  ██║     ██████╔╝ ╚████╔╝ ██████╔╝   ██║   ██║   ██║         ║
    ║  ██║     ██╔══██╗  ╚██╔╝  ██╔═══╝    ██║   ██║   ██║         ║
    ║  ╚██████╗██║  ██║   ██║   ██║        ██║   ╚██████╔╝         ║
    ║   ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚═╝        ╚═╝    ╚═════╝          ║
    ║                                                               ║
    ║              MARKET TERMINAL v1.0.0                          ║
    ║         Advanced Cryptocurrency Trading Platform              ║
    ║                                                               ║
    ╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."
    
    local missing_deps=()
    
    if ! command_exists docker; then
        missing_deps+=("docker")
    fi
    
    if ! command_exists docker-compose; then
        missing_deps+=("docker-compose")
    fi
    
    if ! command_exists go; then
        missing_deps+=("go")
    fi
    
    if ! command_exists git; then
        missing_deps+=("git")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        echo ""
        echo "Please install the missing dependencies:"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        echo "  - Go: https://golang.org/dl/"
        echo "  - Git: https://git-scm.com/downloads"
        exit 1
    fi
    
    print_status "All prerequisites are installed ✓"
}

# Check if Docker is running
check_docker() {
    print_step "Checking Docker status..."
    
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    print_status "Docker is running ✓"
}

# Setup environment
setup_environment() {
    print_step "Setting up environment..."
    
    # Create .env file if it doesn't exist
    if [ ! -f .env ]; then
        print_status "Creating .env file..."
        cat > .env << EOF
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=crypto_terminal
DATABASE_USER=postgres
DATABASE_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=2

# API Keys (optional - add your own)
COINGECKO_API_KEY=

# Logging
LOG_LEVEL=info

# Server Configuration
SERVER_PORT=8090
EOF
        print_status "Created .env file with default configuration"
    else
        print_status ".env file already exists"
    fi
}

# Install Go dependencies
install_dependencies() {
    print_step "Installing Go dependencies..."
    
    if [ -f go.mod ]; then
        go mod tidy
        print_status "Go dependencies installed ✓"
    else
        print_warning "go.mod not found, skipping Go dependencies"
    fi
}

# Start infrastructure services
start_infrastructure() {
    print_step "Starting infrastructure services (PostgreSQL, Redis)..."
    
    docker-compose up -d postgres redis
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check if services are healthy
    if docker-compose ps postgres | grep -q "healthy\|Up"; then
        print_status "PostgreSQL is ready ✓"
    else
        print_warning "PostgreSQL might not be ready yet"
    fi
    
    if docker-compose ps redis | grep -q "healthy\|Up"; then
        print_status "Redis is ready ✓"
    else
        print_warning "Redis might not be ready yet"
    fi
}

# Build and start the application
start_application() {
    print_step "Building and starting the crypto terminal..."
    
    # Build the application
    if command_exists make; then
        make build
    else
        go build -o build/crypto-terminal ./cmd/terminal
    fi
    
    print_status "Application built successfully ✓"
    
    # Start the application
    print_status "Starting crypto terminal on port 8090..."
    echo ""
    echo -e "${PURPLE}🚀 Crypto Market Terminal is starting...${NC}"
    echo ""
    echo "📊 Dashboard: http://localhost:8090"
    echo "🔍 Health Check: http://localhost:8090/health"
    echo "📡 API Docs: http://localhost:8090/api/v1"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop the application${NC}"
    echo ""
    
    # Run the application
    if [ -f build/crypto-terminal ]; then
        ./build/crypto-terminal
    else
        go run ./cmd/terminal
    fi
}

# Start with Docker Compose (alternative method)
start_with_docker() {
    print_step "Starting with Docker Compose..."
    
    docker-compose up --build
}

# Stop all services
stop_services() {
    print_step "Stopping all services..."
    docker-compose down
    print_status "All services stopped ✓"
}

# Clean up
cleanup() {
    print_step "Cleaning up..."
    docker-compose down -v
    docker system prune -f
    print_status "Cleanup completed ✓"
}

# Show help
show_help() {
    echo "Crypto Market Terminal Startup Script"
    echo ""
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  start       Start the crypto terminal (default)"
    echo "  docker      Start with Docker Compose"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  clean       Clean up containers and volumes"
    echo "  dev         Start in development mode"
    echo "  test        Run tests"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start    # Start the terminal normally"
    echo "  $0 docker   # Start with Docker Compose"
    echo "  $0 dev      # Start in development mode"
    echo ""
}

# Development mode
start_dev() {
    print_step "Starting in development mode..."
    
    # Start infrastructure
    start_infrastructure
    
    # Install dependencies
    install_dependencies
    
    # Start with hot reload
    print_status "Starting with hot reload..."
    if command_exists make; then
        make run-dev
    else
        LOG_LEVEL=debug go run ./cmd/terminal
    fi
}

# Run tests
run_tests() {
    print_step "Running tests..."
    
    if command_exists make; then
        make test
    else
        go test ./...
    fi
}

# Main function
main() {
    print_banner
    
    case "${1:-start}" in
        "start")
            check_prerequisites
            check_docker
            setup_environment
            install_dependencies
            start_infrastructure
            start_application
            ;;
        "docker")
            check_prerequisites
            check_docker
            setup_environment
            start_with_docker
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            main start
            ;;
        "clean")
            cleanup
            ;;
        "dev")
            check_prerequisites
            check_docker
            setup_environment
            start_dev
            ;;
        "test")
            check_prerequisites
            install_dependencies
            run_tests
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n${YELLOW}Shutting down gracefully...${NC}"; exit 0' INT

# Run main function
main "$@"
