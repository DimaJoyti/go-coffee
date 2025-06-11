#!/bin/bash

# Go Coffee Crypto Wallet - Enhanced Service Management
# Manages all crypto wallet microservices with health monitoring and dependency management
# Version: 2.0.0
# Usage: ./run.sh [OPTIONS]
#   -b, --build-only    Build services without starting
#   -s, --start-only    Start services without building (assumes already built)
#   -d, --dev-mode      Start in development mode with enhanced logging
#   -p, --production    Start in production mode with optimizations
#   -t, --test-mode     Start with test configurations
#   -m, --monitor       Enable continuous health monitoring
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Source shared library from main project
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please ensure main project scripts are available."
    exit 1
}

print_header "💰 Go Coffee Crypto Wallet Management"

# =============================================================================
# CONFIGURATION
# =============================================================================

BUILD_ONLY=false
START_ONLY=false
DEV_MODE=false
PRODUCTION_MODE=false
TEST_MODE=false
MONITOR_MODE=false
STARTUP_TIMEOUT=60

# Crypto wallet services with their ports
declare -A CRYPTO_SERVICES=(
    ["api-gateway"]="8080"
    ["wallet-service"]="8081"
    ["transaction-service"]="8082"
    ["smart-contract-service"]="8083"
    ["security-service"]="8084"
    ["defi-service"]="8085"
    ["fintech-api"]="8086"
    ["telegram-bot"]="8087"
)

# Service startup order (dependencies first)
CRYPTO_STARTUP_ORDER=(
    "security-service"
    "wallet-service"
    "transaction-service"
    "smart-contract-service"
    "defi-service"
    "fintech-api"
    "telegram-bot"
    "api-gateway"
)

# Track running services
declare -A CRYPTO_RUNNING_SERVICES=()
declare -A CRYPTO_SERVICE_PIDS=()

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_crypto_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -b|--build-only)
                BUILD_ONLY=true
                shift
                ;;
            -s|--start-only)
                START_ONLY=true
                shift
                ;;
            -d|--dev-mode)
                DEV_MODE=true
                shift
                ;;
            -p|--production)
                PRODUCTION_MODE=true
                shift
                ;;
            -t|--test-mode)
                TEST_MODE=true
                shift
                ;;
            -m|--monitor)
                MONITOR_MODE=true
                shift
                ;;
            -h|--help)
                show_usage "run.sh" \
                    "Enhanced crypto wallet service management with health monitoring" \
                    "  ./run.sh [OPTIONS]

  Options:
    -b, --build-only    Build services without starting
    -s, --start-only    Start services without building (assumes already built)
    -d, --dev-mode      Start in development mode with enhanced logging
    -p, --production    Start in production mode with optimizations
    -t, --test-mode     Start with test configurations
    -m, --monitor       Enable continuous health monitoring
    -h, --help          Show this help message

  Examples:
    ./run.sh                        # Build and start all services
    ./run.sh --dev-mode --monitor   # Development with monitoring
    ./run.sh --build-only           # Build only, don't start
    ./run.sh --production           # Production mode"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                print_info "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# CRYPTO SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

# Build crypto service with enhanced error handling
build_crypto_service() {
    local service_name=$1
    local service_path="cmd/$service_name"

    if [[ ! -d "$service_path" ]]; then
        print_warning "$service_name: Service directory not found at $service_path"
        return 1
    fi

    print_progress "Building $service_name..."

    cd "$service_path"

    local build_flags=()
    if [[ "$PRODUCTION_MODE" == "true" ]]; then
        build_flags+=("-ldflags" "-s -w")  # Strip debug info for production
    fi

    if [[ "$DEV_MODE" == "true" ]]; then
        build_flags+=("-race")  # Enable race detection in dev mode
    fi

    # Build with timeout
    if timeout ${STARTUP_TIMEOUT}s go build "${build_flags[@]}" -o "$service_name" . 2>/dev/null; then
        print_status "$service_name built successfully"
        cd - >/dev/null
        return 0
    else
        print_error "Failed to build $service_name"
        cd - >/dev/null
        return 1
    fi
}

# Start crypto service with enhanced monitoring
start_crypto_service() {
    local service_name=$1
    local port=${CRYPTO_SERVICES[$service_name]}
    local service_path="cmd/$service_name"

    print_progress "Starting $service_name on port $port..."

    # Check if port is available
    if ! is_port_available $port; then
        print_warning "Port $port is already in use. Attempting to free it..."
        kill_by_port $port
        wait_for_port $port 10
    fi

    # Set environment variables
    export SERVICE_NAME="$service_name"
    export SERVICE_PORT="$port"
    export LOG_LEVEL="info"

    if [[ "$DEV_MODE" == "true" ]]; then
        export LOG_LEVEL="debug"
        export DEV_MODE="true"
    fi

    if [[ "$PRODUCTION_MODE" == "true" ]]; then
        export PRODUCTION_MODE="true"
        export LOG_LEVEL="warn"
    fi

    if [[ "$TEST_MODE" == "true" ]]; then
        export TEST_MODE="true"
        export LOG_LEVEL="debug"
    fi

    # Start service in background
    local log_file="logs/${service_name}.log"
    mkdir -p logs

    cd "$service_path"
    nohup "./$service_name" > "../../$log_file" 2>&1 &
    local pid=$!
    cd - >/dev/null

    # Save PID
    CRYPTO_SERVICE_PIDS[$service_name]=$pid
    echo $pid > "pids/${service_name}.pid"

    print_debug "$service_name started with PID $pid"

    # Wait for service to be healthy
    if wait_for_service_health_crypto "$service_name" 30; then
        CRYPTO_RUNNING_SERVICES[$service_name]="RUNNING"
        print_status "$service_name started successfully on port $port"
        return 0
    else
        print_error "$service_name failed to start"

        # Kill the failed process
        if kill -0 $pid 2>/dev/null; then
            kill $pid 2>/dev/null || true
        fi

        return 1
    fi
}

# Wait for crypto service to be healthy
wait_for_service_health_crypto() {
    local service_name=$1
    local max_attempts=${2:-30}
    local port=${CRYPTO_SERVICES[$service_name]}
    local attempt=1

    print_progress "Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        # Check if service responds to health check
        if curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
            print_status "$service_name is healthy!"
            return 0
        fi

        if [[ $((attempt % 5)) -eq 0 ]]; then
            print_debug "Health check attempt $attempt/$max_attempts for $service_name"
        fi

        sleep 2
        attempt=$((attempt + 1))
    done

    print_error "$service_name failed to become healthy within $((max_attempts * 2)) seconds"
    return 1
}

# Stop all crypto services gracefully
stop_crypto_services() {
    print_header "🛑 Stopping Crypto Wallet Services"

    # Stop services in reverse order
    local reversed_services=()
    for ((i=${#CRYPTO_STARTUP_ORDER[@]}-1; i>=0; i--)); do
        reversed_services+=("${CRYPTO_STARTUP_ORDER[i]}")
    done

    for service_name in "${reversed_services[@]}"; do
        if [[ -n "${CRYPTO_SERVICE_PIDS[$service_name]:-}" ]]; then
            local pid=${CRYPTO_SERVICE_PIDS[$service_name]}

            if kill -0 $pid 2>/dev/null; then
                print_progress "Stopping $service_name (PID: $pid)..."

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
                    print_warning "Force killing $service_name..."
                    kill -9 $pid 2>/dev/null || true
                fi

                print_status "$service_name stopped"
            fi

            # Clean up PID file
            rm -f "pids/${service_name}.pid"
            unset CRYPTO_SERVICE_PIDS[$service_name]
            unset CRYPTO_RUNNING_SERVICES[$service_name]
        fi
    done

    print_status "All crypto wallet services stopped gracefully"
}

# Show crypto service status
show_crypto_status() {
    print_header "📊 Crypto Wallet Service Status"

    local running_count=0
    local total_count=0

    for service_name in "${CRYPTO_STARTUP_ORDER[@]}"; do
        ((total_count++))
        local port=${CRYPTO_SERVICES[$service_name]}

        if curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
            print_status "$service_name (port $port) - HEALTHY"
            ((running_count++))
        else
            print_error "$service_name (port $port) - UNHEALTHY"
        fi
    done

    echo ""
    print_info "Crypto Services Status: $running_count/$total_count healthy"

    if [[ $running_count -eq $total_count ]]; then
        print_success "🎉 All crypto wallet services are healthy!"
    else
        print_warning "⚠️  Some crypto services need attention"
    fi

    print_header "🌐 Crypto Service Endpoints"
    print_info "API Gateway: http://localhost:8080"
    print_info "Wallet Service: http://localhost:8081"
    print_info "Transaction Service: http://localhost:8082"
    print_info "Smart Contract Service: http://localhost:8083"
    print_info "Security Service: http://localhost:8084"
    print_info "DeFi Service: http://localhost:8085"
    print_info "Fintech API: http://localhost:8086"
    print_info "Telegram Bot: http://localhost:8087"

    echo ""
}

# Monitor crypto services continuously
monitor_crypto_services() {
    print_header "🔍 Starting Crypto Service Monitoring"
    print_info "Monitoring interval: 30s"
    print_info "Press Ctrl+C to stop monitoring and all services"

    while true; do
        sleep 30

        local unhealthy_services=()
        for service_name in "${CRYPTO_STARTUP_ORDER[@]}"; do
            if [[ "${CRYPTO_RUNNING_SERVICES[$service_name]:-}" == "RUNNING" ]]; then
                local port=${CRYPTO_SERVICES[$service_name]}
                if ! curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
                    unhealthy_services+=("$service_name")
                fi
            fi
        done

        if [[ ${#unhealthy_services[@]} -gt 0 ]]; then
            print_warning "Unhealthy crypto services detected: ${unhealthy_services[*]}"

            # Attempt to restart unhealthy services
            for service_name in "${unhealthy_services[@]}"; do
                print_info "Attempting to restart $service_name..."

                # Stop the service
                if [[ -n "${CRYPTO_SERVICE_PIDS[$service_name]:-}" ]]; then
                    local pid=${CRYPTO_SERVICE_PIDS[$service_name]}
                    kill -9 $pid 2>/dev/null || true
                fi

                # Restart the service
                if start_crypto_service "$service_name"; then
                    print_status "$service_name restarted successfully"
                else
                    print_error "Failed to restart $service_name"
                fi
            done
        else
            print_debug "All crypto services are healthy ($(date))"
        fi
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_crypto_args "$@"

    # Set up cleanup on exit
    trap stop_crypto_services EXIT INT TERM

    # Check dependencies
    local deps=("go" "curl" "lsof" "timeout")
    check_dependencies "${deps[@]}" || exit 1

    # Check if we're in the crypto-wallet directory
    if [[ ! -f "go.mod" ]] || ! grep -q "crypto-wallet" go.mod 2>/dev/null; then
        print_error "Please run this script from the crypto-wallet directory."
        exit 1
    fi

    # Create necessary directories
    mkdir -p bin logs pids

    print_info "Build only: $BUILD_ONLY"
    print_info "Start only: $START_ONLY"
    print_info "Development mode: $DEV_MODE"
    print_info "Production mode: $PRODUCTION_MODE"
    print_info "Test mode: $TEST_MODE"
    print_info "Monitor mode: $MONITOR_MODE"

    # Show services list
    print_header "📋 Crypto Wallet Services"
    for service in "${CRYPTO_STARTUP_ORDER[@]}"; do
        local port=${CRYPTO_SERVICES[$service]}
        print_info "  • $service (port $port)"
    done

    # Build services if not start-only mode
    if [[ "$START_ONLY" != "true" ]]; then
        print_header "🔨 Building Crypto Services"

        local failed_builds=()
        local successful_builds=()

        for service_name in "${CRYPTO_STARTUP_ORDER[@]}"; do
            if build_crypto_service "$service_name"; then
                successful_builds+=("$service_name")
            else
                failed_builds+=("$service_name")
            fi
        done

        print_header "📊 Build Summary"
        echo -e "${BOLD}Total Services:${NC} ${#CRYPTO_STARTUP_ORDER[@]}"
        echo -e "${GREEN}Successful:${NC} ${#successful_builds[@]}"
        echo -e "${RED}Failed:${NC} ${#failed_builds[@]}"

        if [[ ${#failed_builds[@]} -gt 0 ]]; then
            print_warning "⚠️  Some services failed to build:"
            for service in "${failed_builds[@]}"; do
                print_error "  • $service"
            done

            if [[ "$BUILD_ONLY" == "true" ]]; then
                exit 1
            fi
        else
            print_success "🎉 All crypto services built successfully!"
        fi
    fi

    # Exit if build-only mode
    if [[ "$BUILD_ONLY" == "true" ]]; then
        print_success "Build completed successfully!"
        exit 0
    fi

    # Start services
    print_header "🚀 Starting Crypto Services"

    local failed_services=()
    local successful_services=()

    for service_name in "${CRYPTO_STARTUP_ORDER[@]}"; do
        if start_crypto_service "$service_name"; then
            successful_services+=("$service_name")
        else
            failed_services+=("$service_name")

            # In production mode, stop on first failure
            if [[ "$PRODUCTION_MODE" == "true" ]]; then
                print_error "Production mode: Stopping due to service failure"
                exit 1
            fi
        fi

        # Small delay between service starts
        sleep 2
    done

    # Calculate startup time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    # Show startup summary
    print_header "📊 Startup Summary"
    echo -e "${BOLD}Total Services:${NC} ${#CRYPTO_STARTUP_ORDER[@]}"
    echo -e "${GREEN}Successful:${NC} ${#successful_services[@]}"
    echo -e "${RED}Failed:${NC} ${#failed_services[@]}"
    echo -e "${BLUE}Startup Time:${NC} ${total_time}s"

    if [[ ${#failed_services[@]} -eq 0 ]]; then
        print_success "🎉 All crypto wallet services started successfully!"

        # Show service status
        show_crypto_status

        print_header "🎯 Next Steps"
        print_info "All crypto wallet services are running!"
        print_info "Available endpoints:"
        print_info "  • Crypto API Gateway: http://localhost:8080"
        print_info "  • Wallet Management: http://localhost:8081"
        print_info "  • Transaction Processing: http://localhost:8082"
        print_info "  • Smart Contracts: http://localhost:8083"
        print_info "  • Security Services: http://localhost:8084"
        print_info "  • DeFi Integration: http://localhost:8085"
        print_info "  • Fintech API: http://localhost:8086"
        print_info "  • Telegram Bot: http://localhost:8087"

        print_info ""
        print_info "Management commands:"
        print_info "  • Check logs: tail -f logs/SERVICE.log"
        print_info "  • Stop services: Ctrl+C"

        # Start monitoring if requested
        if [[ "$MONITOR_MODE" == "true" ]]; then
            monitor_crypto_services
        else
            print_info ""
            print_info "Crypto wallet services are running. Press Ctrl+C to stop all services."

            # Keep script running with basic health checks
            while true; do
                sleep 30

                # Basic health check
                local unhealthy_count=0
                for service_name in "${successful_services[@]}"; do
                    local port=${CRYPTO_SERVICES[$service_name]}
                    if ! curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
                        ((unhealthy_count++))
                    fi
                done

                if [[ $unhealthy_count -gt 0 ]]; then
                    print_warning "$unhealthy_count crypto services are unhealthy. Use --monitor for auto-restart."
                fi
            done
        fi

    else
        print_warning "⚠️  Some crypto services failed to start (${#failed_services[@]}/${#CRYPTO_STARTUP_ORDER[@]})"

        print_header "❌ Failed Services"
        for service in "${failed_services[@]}"; do
            print_error "  • $service"
        done

        print_header "🔍 Troubleshooting"
        print_info "Check service logs for details:"
        for service in "${failed_services[@]}"; do
            print_info "  • tail -f logs/${service}.log"
        done

        print_info ""
        print_info "Common solutions:"
        print_info "  • Check port availability: netstat -tulpn | grep PORT"
        print_info "  • Verify dependencies: go mod tidy"
        print_info "  • Check service configuration"
        print_info "  • Review service health endpoints"

        exit 1
    fi
}

# Run main function with all arguments
main "$@"
