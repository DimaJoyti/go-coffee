#!/bin/bash

# Go Coffee Crypto Terminal - Enhanced Startup Script
# Advanced cryptocurrency trading platform with Bright Data integration
# Version: 2.0.0
# Usage: ./start.sh [OPTIONS]
#   -m, --mode MODE     Startup mode (start|docker|dev|test|production)
#   -e, --env ENV       Environment (development|staging|production)
#   -w, --watch         Enable hot reload in development
#   -b, --build-only    Build without starting
#   -c, --clean         Clean build before starting
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Source shared library from main project if available
if [[ -f "$PROJECT_ROOT/scripts/lib/common.sh" ]]; then
    source "$PROJECT_ROOT/scripts/lib/common.sh"
    SHARED_LIB_AVAILABLE=true
else
    # Fallback color definitions
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    PURPLE='\033[0;35m'
    CYAN='\033[0;36m'
    NC='\033[0m' # No Color
    SHARED_LIB_AVAILABLE=false
fi

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

# Configuration
STARTUP_MODE="start"
ENVIRONMENT="development"
WATCH_MODE=false
BUILD_ONLY=false
CLEAN_BUILD=false

# Print colored output (use shared library if available, otherwise fallback)
if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
    # Use shared library functions
    print_terminal_status() { print_status "$1"; }
    print_terminal_warning() { print_warning "$1"; }
    print_terminal_error() { print_error "$1"; }
    print_terminal_step() { print_info "$1"; }
else
    # Fallback functions
    print_terminal_status() {
        echo -e "${GREEN}[INFO]${NC} $1"
    }

    print_terminal_warning() {
        echo -e "${YELLOW}[WARN]${NC} $1"
    }

    print_terminal_error() {
        echo -e "${RED}[ERROR]${NC} $1"
    }

    print_terminal_step() {
        echo -e "${BLUE}[STEP]${NC} $1"
    }
fi

# Legacy function names for compatibility
print_status() { print_terminal_status "$1"; }
print_warning() { print_terminal_warning "$1"; }
print_error() { print_terminal_error "$1"; }
print_step() { print_terminal_step "$1"; }

# Check if command exists (use shared library if available)
if [[ "$SHARED_LIB_AVAILABLE" != "true" ]]; then
    command_exists() {
        command -v "$1" >/dev/null 2>&1
    }
fi

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_terminal_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -m|--mode)
                STARTUP_MODE="$2"
                shift 2
                ;;
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -w|--watch)
                WATCH_MODE=true
                shift
                ;;
            -b|--build-only)
                BUILD_ONLY=true
                shift
                ;;
            -c|--clean)
                CLEAN_BUILD=true
                shift
                ;;
            -h|--help)
                show_terminal_help
                exit 0
                ;;
            *)
                # Handle legacy positional arguments
                if [[ -z "$STARTUP_MODE" || "$STARTUP_MODE" == "start" ]]; then
                    STARTUP_MODE="$1"
                else
                    print_terminal_error "Unknown option: $1"
                    show_terminal_help
                    exit 1
                fi
                shift
                ;;
        esac
    done
}

# Show enhanced help
show_terminal_help() {
    if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
        show_usage "start.sh" \
            "Enhanced crypto terminal startup script with Bright Data integration" \
            "  ./start.sh [OPTIONS]

  Options:
    -m, --mode MODE     Startup mode (start|docker|dev|test|production)
    -e, --env ENV       Environment (development|staging|production)
    -w, --watch         Enable hot reload in development
    -b, --build-only    Build without starting
    -c, --clean         Clean build before starting
    -h, --help          Show this help message

  Examples:
    ./start.sh                          # Start normally
    ./start.sh --mode dev --watch       # Development with hot reload
    ./start.sh --mode production        # Production mode
    ./start.sh --build-only --clean     # Clean build only"
    else
        echo "Crypto Market Terminal Startup Script v2.0.0"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  -m, --mode MODE     Startup mode (start|docker|dev|test|production)"
        echo "  -e, --env ENV       Environment (development|staging|production)"
        echo "  -w, --watch         Enable hot reload in development"
        echo "  -b, --build-only    Build without starting"
        echo "  -c, --clean         Clean build before starting"
        echo "  -h, --help          Show this help message"
        echo ""
        echo "Legacy Options:"
        echo "  start       Start the crypto terminal (default)"
        echo "  docker      Start with Docker Compose"
        echo "  stop        Stop all services"
        echo "  restart     Restart all services"
        echo "  clean       Clean up containers and volumes"
        echo "  dev         Start in development mode"
        echo "  test        Run tests"
        echo ""
    fi
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

# Enhanced main function with new argument parsing
main() {
    # Parse arguments first
    parse_terminal_args "$@"

    print_banner

    if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
        print_header "💰 Go Coffee Crypto Terminal v2.0.0"
        print_info "Mode: $STARTUP_MODE"
        print_info "Environment: $ENVIRONMENT"
        print_info "Watch mode: $WATCH_MODE"
        print_info "Build only: $BUILD_ONLY"
        print_info "Clean build: $CLEAN_BUILD"
    else
        echo -e "${BLUE}Starting Crypto Terminal...${NC}"
        echo "Mode: $STARTUP_MODE | Environment: $ENVIRONMENT"
    fi

    case "$STARTUP_MODE" in
        "start")
            check_prerequisites
            check_docker
            setup_environment
            install_dependencies
            if [[ "$CLEAN_BUILD" == "true" ]]; then
                cleanup_build
            fi
            start_infrastructure
            if [[ "$BUILD_ONLY" == "true" ]]; then
                build_application
            else
                start_application
            fi
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
            STARTUP_MODE="start"
            main
            ;;
        "clean")
            cleanup
            ;;
        "dev")
            check_prerequisites
            check_docker
            setup_environment
            if [[ "$WATCH_MODE" == "true" ]]; then
                start_dev_with_watch
            else
                start_dev
            fi
            ;;
        "test")
            check_prerequisites
            install_dependencies
            run_tests
            ;;
        "production")
            ENVIRONMENT="production"
            check_prerequisites
            check_docker
            setup_environment
            install_dependencies
            start_infrastructure
            start_application_production
            ;;
        "help"|"-h"|"--help")
            show_terminal_help
            ;;
        *)
            print_error "Unknown mode: $STARTUP_MODE"
            show_terminal_help
            exit 1
            ;;
    esac
}

# Additional functions for enhanced features
cleanup_build() {
    print_step "Cleaning build artifacts..."
    rm -rf build/
    if command_exists make; then
        make clean 2>/dev/null || true
    fi
    print_status "Build artifacts cleaned ✓"
}

build_application() {
    print_step "Building crypto terminal..."

    mkdir -p build

    local build_flags=()
    if [[ "$ENVIRONMENT" == "production" ]]; then
        build_flags+=("-ldflags" "-s -w")
    fi

    if go build "${build_flags[@]}" -o build/crypto-terminal ./cmd/terminal; then
        print_status "Application built successfully ✓"
    else
        print_error "Build failed"
        exit 1
    fi
}

start_dev_with_watch() {
    print_step "Starting in development mode with hot reload..."

    start_infrastructure
    install_dependencies

    if command_exists make; then
        make run-dev-watch 2>/dev/null || start_dev
    else
        print_warning "Hot reload not available, starting normal dev mode"
        start_dev
    fi
}

start_application_production() {
    print_step "Starting in production mode..."

    # Build with production optimizations
    build_application

    # Set production environment variables
    export ENVIRONMENT=production
    export LOG_LEVEL=warn
    export GIN_MODE=release

    print_status "Starting crypto terminal in production mode..."
    echo ""
    echo -e "${PURPLE}🚀 Crypto Market Terminal (Production)${NC}"
    echo ""
    echo "📊 Dashboard: http://localhost:8090"
    echo "🔍 Health Check: http://localhost:8090/health"
    echo "📡 API: http://localhost:8090/api/v1"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop the application${NC}"
    echo ""

    if [[ -f build/crypto-terminal ]]; then
        ./build/crypto-terminal
    else
        print_error "Production binary not found. Run with --build-only first."
        exit 1
    fi
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n${YELLOW}Shutting down gracefully...${NC}"; exit 0' INT

# Run main function
main "$@"
