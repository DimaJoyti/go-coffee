#!/bin/bash

# Go Coffee Web UI - Enhanced Service Management
# Starts MCP server, backend, and frontend with health monitoring and dependency management
# Version: 2.0.0
# Usage: ./start-all.sh [OPTIONS]
#   -d, --dev-mode      Start in development mode with hot reload
#   -p, --production    Start in production mode with optimizations
#   -b, --build-only    Build without starting services
#   -m, --monitor       Enable continuous health monitoring
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
    # Fallback functions
    print_status() { echo "✅ $1"; }
    print_error() { echo "❌ $1"; }
    print_warning() { echo "⚠️  $1"; }
    print_info() { echo "ℹ️  $1"; }
    print_progress() { echo "🔄 $1"; }
    print_success() { echo "🎉 $1"; }
    print_header() { echo ""; echo "🚀 $1"; echo "======================================"; }
    SHARED_LIB_AVAILABLE=false
fi

print_header "Go Coffee Web UI with Bright Data MCP Integration"

# =============================================================================
# CONFIGURATION
# =============================================================================

DEV_MODE=false
PRODUCTION_MODE=false
BUILD_ONLY=false
MONITOR_MODE=false

# Service configuration
MCP_SERVER_PORT="${MCP_SERVER_PORT:-3001}"
BACKEND_PORT="${PORT:-8090}"
FRONTEND_PORT="3000"

# Service PIDs
declare -A WEB_SERVICE_PIDS=()

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_webui_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dev-mode)
                DEV_MODE=true
                shift
                ;;
            -p|--production)
                PRODUCTION_MODE=true
                shift
                ;;
            -b|--build-only)
                BUILD_ONLY=true
                shift
                ;;
            -m|--monitor)
                MONITOR_MODE=true
                shift
                ;;
            -h|--help)
                show_webui_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_webui_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_webui_help() {
    if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
        show_usage "start-all.sh" \
            "Enhanced web UI service management with health monitoring" \
            "  ./start-all.sh [OPTIONS]

  Options:
    -d, --dev-mode      Start in development mode with hot reload
    -p, --production    Start in production mode with optimizations
    -b, --build-only    Build without starting services
    -m, --monitor       Enable continuous health monitoring
    -h, --help          Show this help message

  Examples:
    ./start-all.sh                      # Start all services normally
    ./start-all.sh --dev-mode           # Development with hot reload
    ./start-all.sh --production         # Production mode
    ./start-all.sh --build-only         # Build only, don't start"
    else
        echo "Go Coffee Web UI Service Management v2.0.0"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  -d, --dev-mode      Start in development mode with hot reload"
        echo "  -p, --production    Start in production mode with optimizations"
        echo "  -b, --build-only    Build without starting services"
        echo "  -m, --monitor       Enable continuous health monitoring"
        echo "  -h, --help          Show this help message"
        echo ""
    fi
}

# Load environment variables
load_webui_environment() {
    print_info "Loading environment configuration..."

    if [[ -f ".env" ]]; then
        print_status "Loading environment variables from .env"
        set -a
        source .env
        set +a
    else
        print_warning ".env file not found, using defaults"
    fi

    # Set mode-specific environment variables
    if [[ "$DEV_MODE" == "true" ]]; then
        export NODE_ENV=development
        export LOG_LEVEL=debug
        export GIN_MODE=debug
    elif [[ "$PRODUCTION_MODE" == "true" ]]; then
        export NODE_ENV=production
        export LOG_LEVEL=warn
        export GIN_MODE=release
    fi

    print_info "Environment: ${NODE_ENV:-development}"
    print_info "MCP Server Port: $MCP_SERVER_PORT"
    print_info "Backend Port: $BACKEND_PORT"
    print_info "Frontend Port: $FRONTEND_PORT"
}

# =============================================================================
# SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

# Enhanced port checking
check_webui_port() {
    local port=$1
    local service_name=$2

    if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
        if is_port_available $port; then
            print_status "$service_name port $port is available"
            return 0
        else
            print_warning "$service_name port $port is already in use"
            # Try to free the port
            kill_by_port $port
            if wait_for_port $port 10; then
                print_status "$service_name port $port is now available"
                return 0
            else
                print_error "Failed to free port $port for $service_name"
                return 1
            fi
        fi
    else
        # Fallback port check
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_error "$service_name port $port is already in use"
            return 1
        else
            print_status "$service_name port $port is available"
            return 0
        fi
    fi
}

# Check all required ports
check_all_webui_ports() {
    print_header "🔍 Checking Required Ports"

    check_webui_port $MCP_SERVER_PORT "MCP Server" || exit 1
    check_webui_port $BACKEND_PORT "Backend" || exit 1
    check_webui_port $FRONTEND_PORT "Frontend" || exit 1

    print_success "All required ports are available"
}

# Start MCP Server with enhanced monitoring
start_mcp_server() {
    print_header "🔧 Starting MCP Server"

    if [[ ! -d "mcp-server" ]]; then
        print_error "MCP server directory not found"
        return 1
    fi

    cd mcp-server

    # Set environment variables
    export MCP_SERVER_PORT
    export LOG_LEVEL="${LOG_LEVEL:-info}"

    print_progress "Starting MCP Server on port $MCP_SERVER_PORT..."

    # Start in background
    if [[ "$DEV_MODE" == "true" ]]; then
        go run main.go &
    else
        # Build first for production
        go build -o mcp-server main.go
        ./mcp-server &
    fi

    local pid=$!
    WEB_SERVICE_PIDS["mcp-server"]=$pid
    echo $pid > ../.mcp.pid

    cd ..

    print_status "MCP Server started (PID: $pid)"

    # Wait for service to be ready
    print_progress "Waiting for MCP server to be ready..."
    local attempts=0
    while [[ $attempts -lt 30 ]]; do
        if curl -s "http://localhost:$MCP_SERVER_PORT/health" >/dev/null 2>&1; then
            print_success "MCP Server is ready and healthy"
            return 0
        fi
        sleep 1
        ((attempts++))
    done

    print_error "MCP Server failed to start or become healthy"
    return 1
}

# Start Backend Server with enhanced monitoring
start_backend_server() {
    print_header "🔧 Starting Backend Server"

    if [[ ! -d "backend" ]]; then
        print_error "Backend directory not found"
        return 1
    fi

    cd backend

    # Set environment variables
    export PORT=$BACKEND_PORT
    export MCP_SERVER_URL="http://localhost:$MCP_SERVER_PORT"
    export LOG_LEVEL="${LOG_LEVEL:-info}"

    print_progress "Starting Backend Server on port $BACKEND_PORT..."

    # Start in background
    if [[ "$DEV_MODE" == "true" ]]; then
        go run cmd/web-ui-service/main.go &
    else
        # Build first for production
        go build -o backend-server cmd/web-ui-service/main.go
        ./backend-server &
    fi

    local pid=$!
    WEB_SERVICE_PIDS["backend"]=$pid
    echo $pid > ../.backend.pid

    cd ..

    print_status "Backend Server started (PID: $pid)"

    # Wait for service to be ready
    print_progress "Waiting for backend server to be ready..."
    local attempts=0
    while [[ $attempts -lt 30 ]]; do
        if curl -s "http://localhost:$BACKEND_PORT/health" >/dev/null 2>&1; then
            print_success "Backend Server is ready and healthy"
            return 0
        fi
        sleep 1
        ((attempts++))
    done

    print_error "Backend Server failed to start or become healthy"
    return 1
}

# Start Frontend Server with enhanced monitoring
start_frontend_server() {
    print_header "🔧 Starting Frontend Server"

    if [[ ! -d "frontend" ]]; then
        print_error "Frontend directory not found"
        return 1
    fi

    cd frontend

    # Check if node_modules exists
    if [[ ! -d "node_modules" ]]; then
        print_progress "Installing frontend dependencies..."
        npm install
    fi

    # Set environment variables
    export NEXT_PUBLIC_API_URL="http://localhost:$BACKEND_PORT"
    export NEXT_PUBLIC_MCP_URL="http://localhost:$MCP_SERVER_PORT"

    print_progress "Starting Frontend Server on port $FRONTEND_PORT..."

    # Start in background
    if [[ "$DEV_MODE" == "true" ]]; then
        npm run dev &
    elif [[ "$PRODUCTION_MODE" == "true" ]]; then
        # Build and start for production
        npm run build
        npm run start &
    else
        npm run dev &
    fi

    local pid=$!
    WEB_SERVICE_PIDS["frontend"]=$pid
    echo $pid > ../.frontend.pid

    cd ..

    print_status "Frontend Server started (PID: $pid)"

    # Wait for service to be ready (frontend takes longer)
    print_progress "Waiting for frontend server to be ready..."
    local attempts=0
    while [[ $attempts -lt 60 ]]; do
        if curl -s "http://localhost:$FRONTEND_PORT" >/dev/null 2>&1; then
            print_success "Frontend Server is ready and accessible"
            return 0
        fi
        sleep 2
        ((attempts++))
    done

    print_warning "Frontend Server may still be starting (check manually)"
    return 0
}

# Stop all services gracefully
stop_all_webui_services() {
    print_header "🛑 Stopping All Web UI Services"

    # Stop services in reverse order
    for service in "frontend" "backend" "mcp-server"; do
        if [[ -n "${WEB_SERVICE_PIDS[$service]:-}" ]]; then
            local pid=${WEB_SERVICE_PIDS[$service]}

            if kill -0 $pid 2>/dev/null; then
                print_progress "Stopping $service (PID: $pid)..."

                # Graceful shutdown
                kill -TERM $pid 2>/dev/null || true

                # Wait for graceful shutdown
                local count=0
                while kill -0 $pid 2>/dev/null && [ $count -lt 10 ]; do
                    sleep 1
                    count=$((count + 1))
                done

                # Force kill if still running
                if kill -0 $pid 2>/dev/null; then
                    print_warning "Force killing $service..."
                    kill -9 $pid 2>/dev/null || true
                fi

                print_status "$service stopped"
            fi
        fi
    done

    # Clean up PID files
    rm -f .*.pid

    print_success "All web UI services stopped gracefully"
}

# Show service status
show_webui_status() {
    print_header "📊 Web UI Service Status"

    local services=("mcp-server:$MCP_SERVER_PORT" "backend:$BACKEND_PORT" "frontend:$FRONTEND_PORT")
    local healthy_count=0

    for service_info in "${services[@]}"; do
        local service_name=$(echo "$service_info" | cut -d: -f1)
        local port=$(echo "$service_info" | cut -d: -f2)

        if curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1 || \
           curl -s --max-time 5 "http://localhost:$port" >/dev/null 2>&1; then
            print_status "$service_name (port $port) - HEALTHY"
            ((healthy_count++))
        else
            print_error "$service_name (port $port) - UNHEALTHY"
        fi
    done

    print_info "Service Health: $healthy_count/3 services healthy"

    print_header "🌐 Service Endpoints"
    print_info "MCP Server:    http://localhost:$MCP_SERVER_PORT/health"
    print_info "Backend API:   http://localhost:$BACKEND_PORT/health"
    print_info "Frontend UI:   http://localhost:$FRONTEND_PORT"
    print_info "Market Data:   http://localhost:$BACKEND_PORT/api/v1/scraping/data"
    print_info "API Docs:      http://localhost:$BACKEND_PORT/docs"
}

# Monitor services continuously
monitor_webui_services() {
    print_header "🔍 Starting Web UI Service Monitoring"
    print_info "Monitoring interval: 30s"
    print_info "Press Ctrl+C to stop monitoring and all services"

    while true; do
        sleep 30

        local unhealthy_services=()
        local services=("mcp-server:$MCP_SERVER_PORT" "backend:$BACKEND_PORT" "frontend:$FRONTEND_PORT")

        for service_info in "${services[@]}"; do
            local service_name=$(echo "$service_info" | cut -d: -f1)
            local port=$(echo "$service_info" | cut -d: -f2)

            if ! curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1 && \
               ! curl -s --max-time 5 "http://localhost:$port" >/dev/null 2>&1; then
                unhealthy_services+=("$service_name")
            fi
        done

        if [[ ${#unhealthy_services[@]} -gt 0 ]]; then
            print_warning "Unhealthy services detected: ${unhealthy_services[*]}"
        else
            print_info "All web UI services are healthy ($(date))"
        fi
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_webui_args "$@"

    # Set up cleanup on exit
    trap stop_all_webui_services EXIT INT TERM

    # Load environment
    load_webui_environment

    print_info "Configuration:"
    print_info "  Development mode: $DEV_MODE"
    print_info "  Production mode: $PRODUCTION_MODE"
    print_info "  Build only: $BUILD_ONLY"
    print_info "  Monitor mode: $MONITOR_MODE"

    # Check dependencies
    local deps=("go" "node" "npm" "curl")
    if [[ "$SHARED_LIB_AVAILABLE" == "true" ]]; then
        check_dependencies "${deps[@]}" || exit 1
    else
        for dep in "${deps[@]}"; do
            if ! command -v "$dep" >/dev/null 2>&1; then
                print_error "$dep is not installed"
                exit 1
            fi
        done
    fi

    # Check ports
    check_all_webui_ports

    # Start services
    print_header "🚀 Starting Web UI Services"

    local failed_services=()

    # Start MCP Server
    if start_mcp_server; then
        print_success "MCP Server started successfully"
    else
        failed_services+=("mcp-server")
    fi

    # Start Backend Server
    if start_backend_server; then
        print_success "Backend Server started successfully"
    else
        failed_services+=("backend")
    fi

    # Exit if build-only mode
    if [[ "$BUILD_ONLY" == "true" ]]; then
        print_success "Build completed successfully!"
        exit 0
    fi

    # Start Frontend Server
    if start_frontend_server; then
        print_success "Frontend Server started successfully"
    else
        failed_services+=("frontend")
    fi

    # Calculate startup time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    # Show startup summary
    print_header "📊 Startup Summary"
    echo -e "${BOLD}Total Services:${NC} 3"
    echo -e "${GREEN}Successful:${NC} $((3 - ${#failed_services[@]}))"
    echo -e "${RED}Failed:${NC} ${#failed_services[@]}"
    echo -e "${BLUE}Startup Time:${NC} ${total_time}s"

    if [[ ${#failed_services[@]} -eq 0 ]]; then
        print_success "🎉 All web UI services started successfully!"

        # Show service status
        show_webui_status

        print_header "🎯 Next Steps"
        print_info "All web UI services are running!"
        print_info "Open your browser and visit: http://localhost:$FRONTEND_PORT"

        # Start monitoring if requested
        if [[ "$MONITOR_MODE" == "true" ]]; then
            monitor_webui_services
        else
            print_info ""
            print_info "Web UI services are running. Press Ctrl+C to stop all services."

            # Keep script running
            while true; do
                sleep 30

                # Basic health check
                local unhealthy_count=0
                local services=("mcp-server:$MCP_SERVER_PORT" "backend:$BACKEND_PORT" "frontend:$FRONTEND_PORT")

                for service_info in "${services[@]}"; do
                    local service_name=$(echo "$service_info" | cut -d: -f1)
                    local port=$(echo "$service_info" | cut -d: -f2)

                    if ! curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1 && \
                       ! curl -s --max-time 5 "http://localhost:$port" >/dev/null 2>&1; then
                        ((unhealthy_count++))
                    fi
                done

                if [[ $unhealthy_count -gt 0 ]]; then
                    print_warning "$unhealthy_count web UI services are unhealthy. Use --monitor for detailed monitoring."
                fi
            done
        fi

    else
        print_warning "⚠️  Some web UI services failed to start"

        print_header "❌ Failed Services"
        for service in "${failed_services[@]}"; do
            print_error "  • $service"
        done

        print_header "🔍 Troubleshooting"
        print_info "Check service logs and dependencies"
        print_info "Ensure all required ports are available"
        print_info "Verify Go and Node.js are properly installed"

        exit 1
    fi
}

# Run main function with all arguments
main "$@"
