#!/bin/bash

# Go Coffee - Service Orchestration Script
# Builds and starts all microservices with health monitoring and dependency management
# Version: 2.0.0
# Usage: ./start-all-services.sh [OPTIONS]
#   -c, --core-only     Start only core production services
#   -d, --dev-mode      Start in development mode with hot reload
#   -m, --monitor       Enable continuous health monitoring
#   -p, --production    Start in production mode with optimizations
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory for relative imports
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared library
source "$SCRIPT_DIR/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🚀 Go Coffee Service Orchestration"

# =============================================================================
# CONFIGURATION
# =============================================================================

START_MODE="all"
DEV_MODE=false
MONITOR_MODE=false
PRODUCTION_MODE=false
STARTUP_TIMEOUT=60
HEALTH_CHECK_INTERVAL=10
MAX_STARTUP_RETRIES=3

# Service port mappings (service:port:health_path)
declare -A SERVICE_PORTS=(
    ["auth-service"]="8091:/health"
    ["payment-service"]="8093:/health"
    ["order-service"]="8094:/health"
    ["kitchen-service"]="8095:/health"
    ["user-gateway"]="8096:/health"
    ["security-gateway"]="8097:/health"
    ["communication-hub"]="8098:/health"
    ["ai-search"]="8099:/health"
    ["ai-service"]="8100:/health"
    ["ai-arbitrage-service"]="8101:/health"
    ["ai-order-service"]="8102:/health"
    ["market-data-service"]="8103:/health"
    ["defi-service"]="8104:/health"
    ["bright-data-hub-service"]="8105:/health"
    ["llm-orchestrator"]="8106:/health"
    ["llm-orchestrator-simple"]="8107:/health"
    ["redis-mcp-server"]="8108:/health"
    ["mcp-ai-integration"]="8109:/health"
    ["task-cli"]="8110:/health"
    ["api-gateway"]="8080:/health"
)

# Service startup order (dependencies first)
STARTUP_ORDER=(
    # Infrastructure services first
    "auth-service"
    "redis-mcp-server"
    "security-gateway"

    # Core business services
    "payment-service"
    "order-service"
    "kitchen-service"
    "user-gateway"
    "communication-hub"

    # AI services
    "ai-search"
    "ai-service"
    "llm-orchestrator"
    "llm-orchestrator-simple"
    "ai-arbitrage-service"
    "ai-order-service"
    "mcp-ai-integration"

    # External integration services
    "market-data-service"
    "defi-service"
    "bright-data-hub-service"

    # Utility services
    "task-cli"

    # Gateway last (depends on all other services)
    "api-gateway"
)

# Track running services
declare -A RUNNING_SERVICES=()
declare -A SERVICE_PIDS=()

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--core-only)
                START_MODE="core"
                shift
                ;;
            -d|--dev-mode)
                DEV_MODE=true
                shift
                ;;
            -m|--monitor)
                MONITOR_MODE=true
                shift
                ;;
            -p|--production)
                PRODUCTION_MODE=true
                shift
                ;;
            -h|--help)
                show_usage "start-all-services.sh" \
                    "Build and start all Go Coffee microservices with health monitoring" \
                    "  ./start-all-services.sh [OPTIONS]

  Options:
    -c, --core-only     Start only core production services
    -d, --dev-mode      Start in development mode with hot reload
    -m, --monitor       Enable continuous health monitoring
    -p, --production    Start in production mode with optimizations
    -h, --help          Show this help message

  Examples:
    ./start-all-services.sh                    # Start all services
    ./start-all-services.sh --core-only        # Start only core services
    ./start-all-services.sh --dev-mode         # Development mode
    ./start-all-services.sh --production -m    # Production with monitoring"
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
# SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

# Get services to start based on mode
get_services_to_start() {
    case $START_MODE in
        "core")
            # Filter startup order to include only core services
            local core_services=()
            for service in "${STARTUP_ORDER[@]}"; do
                for core_service in "${CORE_SERVICES[@]}"; do
                    if [[ "$service" == "$core_service" ]]; then
                        core_services+=("$service")
                        break
                    fi
                done
            done
            echo "${core_services[@]}"
            ;;
        "all")
            echo "${STARTUP_ORDER[@]}"
            ;;
        *)
            print_error "Unknown start mode: $START_MODE"
            exit 1
            ;;
    esac
}

# Enhanced service health check
check_service_health_enhanced() {
    local service_name=$1
    local port_info=${SERVICE_PORTS[$service_name]}
    local port=$(echo "$port_info" | cut -d':' -f1)
    local health_path=$(echo "$port_info" | cut -d':' -f2)
    local health_url="http://localhost:$port$health_path"

    if curl -s --max-time 5 "$health_url" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Wait for service to be healthy with retries
wait_for_service_health() {
    local service_name=$1
    local max_attempts=${2:-30}
    local attempt=1

    print_progress "Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        if check_service_health_enhanced "$service_name"; then
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

# Build service with enhanced error handling
build_service_enhanced() {
    local service_name=$1
    local service_path="cmd/$service_name"

    if [[ ! -d "$service_path" ]]; then
        print_warning "$service_name: Service directory not found at $service_path"
        return 1
    fi

    print_progress "Building $service_name..."

    local build_flags=()
    if [[ "$PRODUCTION_MODE" == "true" ]]; then
        build_flags+=("-ldflags" "-s -w")  # Strip debug info for production
    fi

    if [[ "$DEV_MODE" == "true" ]]; then
        build_flags+=("-race")  # Enable race detection in dev mode
    fi

    # Build with timeout
    if timeout ${STARTUP_TIMEOUT}s go build "${build_flags[@]}" -o "bin/$service_name" "$service_path/main.go" 2>/dev/null; then
        print_status "$service_name built successfully"
        return 0
    else
        print_error "Failed to build $service_name"
        return 1
    fi
}

# Start service with enhanced monitoring
start_service_enhanced() {
    local service_name=$1
    local port_info=${SERVICE_PORTS[$service_name]}
    local port=$(echo "$port_info" | cut -d':' -f1)
    local retries=0

    print_progress "Starting $service_name on port $port..."

    # Check if port is available
    if ! is_port_available $port; then
        print_warning "Port $port is already in use. Attempting to free it..."
        kill_by_port $port
        wait_for_port $port 10
    fi

    # Build the service first
    if ! build_service_enhanced "$service_name"; then
        return 1
    fi

    # Start service with retries
    while [[ $retries -lt $MAX_STARTUP_RETRIES ]]; do
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

        # Start service in background
        local log_file="logs/${service_name}.log"
        mkdir -p logs

        nohup "./bin/$service_name" > "$log_file" 2>&1 &
        local pid=$!

        # Save PID
        SERVICE_PIDS[$service_name]=$pid
        echo $pid > "pids/${service_name}.pid"

        print_debug "$service_name started with PID $pid"

        # Wait for service to be healthy
        if wait_for_service_health "$service_name" 30; then
            RUNNING_SERVICES[$service_name]="RUNNING"
            print_status "$service_name started successfully on port $port"
            return 0
        else
            print_warning "$service_name failed to start (attempt $((retries + 1))/$MAX_STARTUP_RETRIES)"

            # Kill the failed process
            if kill -0 $pid 2>/dev/null; then
                kill $pid 2>/dev/null || true
            fi

            ((retries++))
            sleep 5
        fi
    done

    print_error "$service_name failed to start after $MAX_STARTUP_RETRIES attempts"
    return 1
}

# Stop all services gracefully
stop_all_services() {
    print_header "🛑 Stopping All Services"

    # Stop services in reverse order
    local services_to_stop=($(get_services_to_start))
    local reversed_services=()

    # Reverse the array
    for ((i=${#services_to_stop[@]}-1; i>=0; i--)); do
        reversed_services+=("${services_to_stop[i]}")
    done

    for service_name in "${reversed_services[@]}"; do
        if [[ -n "${SERVICE_PIDS[$service_name]:-}" ]]; then
            local pid=${SERVICE_PIDS[$service_name]}

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
            unset SERVICE_PIDS[$service_name]
            unset RUNNING_SERVICES[$service_name]
        fi
    done

    print_status "All services stopped gracefully"
}

# Show comprehensive service status
show_service_status() {
    print_header "📊 Service Status Dashboard"

    local running_count=0
    local total_count=0

    for service_name in $(get_services_to_start); do
        ((total_count++))
        local port_info=${SERVICE_PORTS[$service_name]}
        local port=$(echo "$port_info" | cut -d':' -f1)

        if check_service_health_enhanced "$service_name"; then
            print_status "$service_name (port $port) - HEALTHY"
            ((running_count++))
        else
            print_error "$service_name (port $port) - UNHEALTHY"
        fi
    done

    echo ""
    print_info "Services Status: $running_count/$total_count healthy"

    if [[ $running_count -eq $total_count ]]; then
        print_success "🎉 All services are healthy!"
    else
        print_warning "⚠️  Some services need attention"
    fi

    print_header "🌐 Service Endpoints"
    print_info "API Gateway: http://localhost:8080"
    print_info "Health Check: http://localhost:8080/health"
    print_info "Service Status: http://localhost:8080/api/v1/status"
    print_info "Documentation: http://localhost:8080/docs"

    if [[ "$DEV_MODE" == "true" ]]; then
        print_info "Development Dashboard: http://localhost:8080/dev"
    fi

    echo ""
}

# Continuous health monitoring
monitor_services() {
    print_header "🔍 Starting Continuous Health Monitoring"
    print_info "Monitoring interval: ${HEALTH_CHECK_INTERVAL}s"
    print_info "Press Ctrl+C to stop monitoring and all services"

    while true; do
        sleep $HEALTH_CHECK_INTERVAL

        local unhealthy_services=()
        for service_name in $(get_services_to_start); do
            if [[ "${RUNNING_SERVICES[$service_name]:-}" == "RUNNING" ]]; then
                if ! check_service_health_enhanced "$service_name"; then
                    unhealthy_services+=("$service_name")
                fi
            fi
        done

        if [[ ${#unhealthy_services[@]} -gt 0 ]]; then
            print_warning "Unhealthy services detected: ${unhealthy_services[*]}"

            # Attempt to restart unhealthy services
            for service_name in "${unhealthy_services[@]}"; do
                print_info "Attempting to restart $service_name..."

                # Stop the service
                if [[ -n "${SERVICE_PIDS[$service_name]:-}" ]]; then
                    local pid=${SERVICE_PIDS[$service_name]}
                    kill -9 $pid 2>/dev/null || true
                fi

                # Restart the service
                if start_service_enhanced "$service_name"; then
                    print_status "$service_name restarted successfully"
                else
                    print_error "Failed to restart $service_name"
                fi
            done
        else
            print_debug "All services are healthy ($(date))"
        fi
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_args "$@"

    # Set up cleanup on exit
    trap stop_all_services EXIT INT TERM

    # Check dependencies
    local deps=("go" "curl" "lsof" "timeout")
    check_dependencies "${deps[@]}" || exit 1

    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run from project root."
        exit 1
    fi

    # Create necessary directories
    mkdir -p bin logs pids

    # Get services to start
    local services_to_start=($(get_services_to_start))
    local total_services=${#services_to_start[@]}

    print_info "Start mode: $START_MODE"
    print_info "Services to start: $total_services"
    print_info "Development mode: $DEV_MODE"
    print_info "Production mode: $PRODUCTION_MODE"
    print_info "Monitor mode: $MONITOR_MODE"

    # Show services list
    print_header "📋 Services to Start"
    for service in "${services_to_start[@]}"; do
        local port_info=${SERVICE_PORTS[$service]}
        local port=$(echo "$port_info" | cut -d':' -f1)
        print_info "  • $service (port $port)"
    done

    # Start services in dependency order
    print_header "🚀 Starting Services"

    local failed_services=()
    local successful_services=()

    for service_name in "${services_to_start[@]}"; do
        if start_service_enhanced "$service_name"; then
            successful_services+=("$service_name")
            print_debug "$service_name added to successful services"
        else
            failed_services+=("$service_name")
            print_error "Failed to start $service_name"

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
    echo -e "${BOLD}Total Services:${NC} $total_services"
    echo -e "${GREEN}Successful:${NC} ${#successful_services[@]}"
    echo -e "${RED}Failed:${NC} ${#failed_services[@]}"
    echo -e "${BLUE}Startup Time:${NC} ${total_time}s"

    if [[ ${#failed_services[@]} -eq 0 ]]; then
        print_success "🎉 All services started successfully!"

        # Show service status
        show_service_status

        # Test API Gateway if it's running
        if [[ " ${successful_services[*]} " =~ " api-gateway " ]]; then
            print_header "🧪 Testing API Gateway"

            if check_service_health_enhanced "api-gateway"; then
                print_status "API Gateway is responding"

                # Test service status endpoint
                print_info "Testing service discovery..."
                if curl -s "http://localhost:8080/api/v1/status" >/dev/null 2>&1; then
                    print_status "Service discovery is working"
                else
                    print_warning "Service discovery endpoint not available"
                fi
            else
                print_warning "API Gateway health check failed"
            fi
        fi

        print_header "🎯 Next Steps"
        print_info "All services are running successfully!"
        print_info "Available endpoints:"
        print_info "  • API Gateway: http://localhost:8080"
        print_info "  • Health Check: http://localhost:8080/health"
        print_info "  • Documentation: http://localhost:8080/docs"

        if [[ "$DEV_MODE" == "true" ]]; then
            print_info "  • Development Dashboard: http://localhost:8080/dev"
        fi

        print_info ""
        print_info "Management commands:"
        print_info "  • Check status: ./scripts/health-check.sh"
        print_info "  • View logs: tail -f logs/SERVICE.log"
        print_info "  • Stop services: Ctrl+C or ./scripts/stop-all-services.sh"

        # Start monitoring if requested
        if [[ "$MONITOR_MODE" == "true" ]]; then
            monitor_services
        else
            print_info ""
            print_info "Services are running. Press Ctrl+C to stop all services."

            # Keep script running with basic health checks
            while true; do
                sleep $HEALTH_CHECK_INTERVAL

                # Basic health check
                local unhealthy_count=0
                for service_name in "${successful_services[@]}"; do
                    if ! check_service_health_enhanced "$service_name"; then
                        ((unhealthy_count++))
                    fi
                done

                if [[ $unhealthy_count -gt 0 ]]; then
                    print_warning "$unhealthy_count services are unhealthy. Use --monitor for auto-restart."
                fi
            done
        fi

    else
        print_warning "⚠️  Some services failed to start (${#failed_services[@]}/$total_services)"

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
