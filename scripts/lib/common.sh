#!/bin/bash

# Go Coffee - Shared Script Library
# Common functions and utilities for all Go Coffee scripts
# Version: 2.0.0
# Author: Go Coffee Development Team

set -euo pipefail

# =============================================================================
# COLOR DEFINITIONS (Standardized across all scripts)
# =============================================================================

export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export PURPLE='\033[0;35m'
export CYAN='\033[0;36m'
export WHITE='\033[1;37m'
export BOLD='\033[1m'
export NC='\033[0m' # No Color

# =============================================================================
# LOGGING FUNCTIONS (Standardized output)
# =============================================================================

print_header() {
    echo -e "\n${BOLD}${BLUE}$1${NC}"
    echo -e "${BLUE}$(printf '=%.0s' $(seq 1 ${#1}))${NC}"
}

print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        echo -e "${PURPLE}🐛 DEBUG: $1${NC}"
    fi
}

print_progress() {
    echo -e "${CYAN}🔄 $1${NC}"
}

print_success() {
    echo -e "${GREEN}🎉 $1${NC}"
}

# =============================================================================
# SERVICE DEFINITIONS (Complete inventory)
# =============================================================================

# Core production services
export CORE_SERVICES=(
    "ai-search"
    "ai-service" 
    "ai-arbitrage-service"
    "ai-order-service"
    "auth-service"
    "communication-hub"
    "user-gateway"
    "security-gateway"
    "kitchen-service"
    "order-service"
    "payment-service"
    "api-gateway"
    "market-data-service"
    "defi-service"
    "bright-data-hub-service"
    "llm-orchestrator"
    "llm-orchestrator-simple"
    "redis-mcp-server"
    "mcp-ai-integration"
    "task-cli"
)

# Test and demo services
export TEST_SERVICES=(
    "ai-arbitrage-demo"
    "auth-test"
    "config-test"
    "test-server"
    "simple-auth"
    "redis-mcp-demo"
    "llm-orchestrator-minimal"
)

# AI-specific services
export AI_SERVICES=(
    "ai-search"
    "ai-service"
    "ai-arbitrage-service"
    "ai-order-service"
    "llm-orchestrator"
    "llm-orchestrator-simple"
    "llm-orchestrator-minimal"
    "mcp-ai-integration"
)

# Infrastructure services
export INFRASTRUCTURE_SERVICES=(
    "auth-service"
    "communication-hub"
    "user-gateway"
    "security-gateway"
    "api-gateway"
    "redis-mcp-server"
)

# Business logic services
export BUSINESS_SERVICES=(
    "kitchen-service"
    "order-service"
    "payment-service"
    "market-data-service"
    "defi-service"
    "bright-data-hub-service"
)

# All services combined
export ALL_SERVICES=("${CORE_SERVICES[@]}" "${TEST_SERVICES[@]}")

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if port is available
is_port_available() {
    local port=$1
    ! lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1
}

# Get process ID using port
get_pid_by_port() {
    local port=$1
    lsof -ti:$port 2>/dev/null || echo ""
}

# Kill process by port
kill_by_port() {
    local port=$1
    local pid=$(get_pid_by_port $port)
    if [[ -n "$pid" ]]; then
        print_info "Killing process on port $port (PID: $pid)"
        kill -9 $pid 2>/dev/null || true
        return 0
    fi
    return 1
}

# Wait for port to be available
wait_for_port() {
    local port=$1
    local timeout=${2:-30}
    local count=0
    
    print_progress "Waiting for port $port to be available..."
    
    while ! is_port_available $port && [ $count -lt $timeout ]; do
        sleep 1
        count=$((count + 1))
    done
    
    if [ $count -ge $timeout ]; then
        print_error "Timeout waiting for port $port to be available"
        return 1
    fi
    
    print_status "Port $port is now available"
    return 0
}

# Check if service is healthy
check_service_health() {
    local service_name=$1
    local health_url=${2:-"http://localhost:8080/health"}
    local timeout=${3:-10}
    
    print_progress "Checking health of $service_name..."
    
    if curl -s --max-time $timeout "$health_url" >/dev/null 2>&1; then
        print_status "$service_name is healthy"
        return 0
    else
        print_warning "$service_name health check failed"
        return 1
    fi
}

# Check dependencies
check_dependencies() {
    local deps=("$@")
    local missing=()
    
    print_info "Checking dependencies..."
    
    for dep in "${deps[@]}"; do
        if ! command_exists "$dep"; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -gt 0 ]; then
        print_error "Missing dependencies: ${missing[*]}"
        print_info "Please install missing dependencies and try again"
        return 1
    fi
    
    print_status "All dependencies are available"
    return 0
}

# Build service with timeout
build_service_with_timeout() {
    local service_name=$1
    local service_path=$2
    local timeout=${3:-60}
    local output_dir=${4:-"bin"}
    
    print_progress "Building $service_name..."
    
    # Create output directory if it doesn't exist
    mkdir -p "$output_dir"
    
    # Use timeout to prevent hanging builds
    if timeout ${timeout}s go build -o "$output_dir/$service_name" "$service_path" 2>/dev/null; then
        print_status "$service_name built successfully"
        return 0
    else
        print_error "$service_name build failed or timed out"
        return 1
    fi
}

# Test service with coverage
test_service_with_coverage() {
    local service_name=$1
    local test_path=${2:-"./..."}
    local timeout=${3:-120}
    
    print_progress "Testing $service_name..."
    
    if timeout ${timeout}s go test -v -race -timeout=60s -coverprofile=coverage.out "$test_path" 2>/dev/null; then
        print_status "$service_name tests passed"
        return 0
    else
        print_warning "$service_name tests failed or timed out"
        return 1
    fi
}

# Cleanup function for script exit
cleanup_on_exit() {
    local cleanup_func=${1:-""}
    
    if [[ -n "$cleanup_func" ]]; then
        trap "$cleanup_func" EXIT INT TERM
    fi
}

# Show script usage
show_usage() {
    local script_name=$1
    local description=$2
    local usage_text=$3

    echo -e "${BOLD}${BLUE}$script_name${NC}"
    echo -e "${BLUE}$(printf '=%.0s' $(seq 1 ${#script_name}))${NC}"
    echo -e "\n${YELLOW}Description:${NC}"
    echo -e "  $description"
    echo -e "\n${YELLOW}Usage:${NC}"
    echo -e "$usage_text"
    echo ""
}

# =============================================================================
# SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

# Start service in background
start_service_background() {
    local service_name=$1
    local service_path=$2
    local port=${3:-8080}
    local log_file=${4:-"logs/${service_name}.log"}

    print_progress "Starting $service_name on port $port..."

    # Create logs directory
    mkdir -p logs

    # Start service in background
    nohup "$service_path" > "$log_file" 2>&1 &
    local pid=$!

    # Save PID for later cleanup
    echo $pid > "pids/${service_name}.pid"

    # Wait a moment and check if service started
    sleep 2
    if kill -0 $pid 2>/dev/null; then
        print_status "$service_name started successfully (PID: $pid)"
        return 0
    else
        print_error "$service_name failed to start"
        return 1
    fi
}

# Stop service by PID file
stop_service_by_pid() {
    local service_name=$1
    local pid_file="pids/${service_name}.pid"

    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if kill -0 $pid 2>/dev/null; then
            print_progress "Stopping $service_name (PID: $pid)..."
            kill -TERM $pid

            # Wait for graceful shutdown
            local count=0
            while kill -0 $pid 2>/dev/null && [ $count -lt 10 ]; do
                sleep 1
                count=$((count + 1))
            done

            # Force kill if still running
            if kill -0 $pid 2>/dev/null; then
                print_warning "Force killing $service_name..."
                kill -9 $pid
            fi

            print_status "$service_name stopped"
        fi
        rm -f "$pid_file"
    else
        print_warning "No PID file found for $service_name"
    fi
}

# Get service status
get_service_status() {
    local service_name=$1
    local port=${2:-8080}
    local pid_file="pids/${service_name}.pid"

    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if kill -0 $pid 2>/dev/null; then
            if ! is_port_available $port; then
                echo "RUNNING"
            else
                echo "STARTED_NO_PORT"
            fi
        else
            echo "STOPPED"
        fi
    else
        echo "NOT_STARTED"
    fi
}

# =============================================================================
# DOCKER FUNCTIONS
# =============================================================================

# Check if Docker is running
check_docker() {
    if ! command_exists docker; then
        print_error "Docker is not installed"
        return 1
    fi

    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running"
        return 1
    fi

    print_status "Docker is available and running"
    return 0
}

# Build Docker image
build_docker_image() {
    local image_name=$1
    local dockerfile_path=${2:-"Dockerfile"}
    local context_path=${3:-"."}

    print_progress "Building Docker image: $image_name..."

    if docker build -t "$image_name" -f "$dockerfile_path" "$context_path"; then
        print_status "Docker image $image_name built successfully"
        return 0
    else
        print_error "Failed to build Docker image $image_name"
        return 1
    fi
}

# =============================================================================
# KUBERNETES FUNCTIONS
# =============================================================================

# Check if kubectl is available
check_kubectl() {
    if ! command_exists kubectl; then
        print_error "kubectl is not installed"
        return 1
    fi

    if ! kubectl cluster-info >/dev/null 2>&1; then
        print_error "kubectl cannot connect to cluster"
        return 1
    fi

    print_status "kubectl is available and connected"
    return 0
}

# Apply Kubernetes manifest
apply_k8s_manifest() {
    local manifest_path=$1
    local namespace=${2:-"default"}

    print_progress "Applying Kubernetes manifest: $manifest_path..."

    if kubectl apply -f "$manifest_path" -n "$namespace"; then
        print_status "Kubernetes manifest applied successfully"
        return 0
    else
        print_error "Failed to apply Kubernetes manifest"
        return 1
    fi
}

# =============================================================================
# INITIALIZATION
# =============================================================================

# Create necessary directories
init_directories() {
    mkdir -p bin logs pids coverage
    print_debug "Created necessary directories: bin, logs, pids, coverage"
}

# Initialize common script environment
init_common() {
    # Create directories
    init_directories

    # Set up cleanup on exit
    cleanup_on_exit "cleanup_pids"

    print_debug "Common script environment initialized"
}

# Cleanup PID files on exit
cleanup_pids() {
    if [[ -d "pids" ]]; then
        for pid_file in pids/*.pid; do
            if [[ -f "$pid_file" ]]; then
                local service_name=$(basename "$pid_file" .pid)
                stop_service_by_pid "$service_name"
            fi
        done
    fi
}

# Auto-initialize when sourced (commented out to avoid issues)
# if [[ "${BASH_SOURCE[0]:-}" != "${0:-}" ]]; then
#     init_common
# fi
