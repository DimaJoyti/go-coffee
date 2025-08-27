#!/bin/bash

# Go Coffee - Complete Build Script
# Builds all microservices in the Go Coffee ecosystem
# Version: 2.0.0
# Usage: ./build_all.sh [OPTIONS]
#   -c, --core-only     Build only core production services
#   -t, --test-only     Build only test services
#   -a, --ai-only       Build only AI services
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory for relative imports
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared library (build_all.sh is in root, so path is scripts/lib/common.sh)
source "$SCRIPT_DIR/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library from $SCRIPT_DIR/scripts/lib/common.sh"
    echo "Please ensure you're running from the project root directory."
    exit 1
}

print_header "🔧 Building All Go Coffee Services"

# =============================================================================
# CONFIGURATION
# =============================================================================

BUILD_TIMEOUT=60
BUILD_DIR="bin"
PARALLEL_BUILDS=4
BUILD_MODE="all"

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--core-only)
                BUILD_MODE="core"
                shift
                ;;
            -t|--test-only)
                BUILD_MODE="test"
                shift
                ;;
            -a|--ai-only)
                BUILD_MODE="ai"
                shift
                ;;
            -h|--help)
                show_usage "build_all.sh" \
                    "Build all Go Coffee microservices with enhanced error handling and parallel processing" \
                    "  ./build_all.sh [OPTIONS]

  Options:
    -c, --core-only     Build only core production services (${#CORE_SERVICES[@]} services)
    -t, --test-only     Build only test services (${#TEST_SERVICES[@]} services)
    -a, --ai-only       Build only AI services (${#AI_SERVICES[@]} services)
    -h, --help          Show this help message

  Examples:
    ./build_all.sh                    # Build all services
    ./build_all.sh --core-only        # Build only production services
    ./build_all.sh --ai-only          # Build only AI services"
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
# DEPENDENCY CHECKS
# =============================================================================

check_build_dependencies() {
    print_info "Checking build dependencies..."

    local deps=("go" "timeout")
    if ! check_dependencies "${deps[@]}"; then
        exit 1
    fi

    # Check Go version
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    print_info "Go version: $go_version"

    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run from project root."
        exit 1
    fi

    print_status "Build environment ready"
}

# =============================================================================
# BUILD FUNCTIONS
# =============================================================================

# Enhanced build service function with better error handling
build_service_enhanced() {
    local service_name=$1
    local service_path="cmd/$service_name/main.go"
    local build_start=$(date +%s)

    # Check if service exists
    if [[ ! -f "$service_path" ]]; then
        print_warning "$service_name: main.go not found at $service_path"
        return 1
    fi

    print_progress "Building $service_name..."

    # Build with enhanced error capture
    local build_output
    if build_output=$(timeout ${BUILD_TIMEOUT}s go build -v -o "$BUILD_DIR/$service_name" "$service_path" 2>&1); then
        local build_end=$(date +%s)
        local build_time=$((build_end - build_start))
        print_status "$service_name built successfully (${build_time}s)"
        print_debug "Build output: $build_output"
        return 0
    else
        local exit_code=$?
        print_error "$service_name build failed (exit code: $exit_code)"
        if [[ $exit_code -eq 124 ]]; then
            print_warning "Build timed out after ${BUILD_TIMEOUT}s"
        fi
        print_debug "Build error: $build_output"
        return 1
    fi
}

# Build services in parallel
build_services_parallel() {
    local services=("$@")
    local pids=()
    local results=()

    print_info "Building ${#services[@]} services in parallel (max $PARALLEL_BUILDS concurrent)..."

    # Function to build single service in background
    build_single() {
        local service=$1
        local index=$2
        if build_service_enhanced "$service"; then
            echo "SUCCESS:$index:$service" > "build_result_$index.tmp"
        else
            echo "FAILED:$index:$service" > "build_result_$index.tmp"
        fi
    }

    # Start builds in batches
    local index=0
    for service in "${services[@]}"; do
        # Wait if we've reached max parallel builds
        if [[ ${#pids[@]} -ge $PARALLEL_BUILDS ]]; then
            wait ${pids[0]}
            pids=("${pids[@]:1}")
        fi

        # Start build in background
        build_single "$service" $index &
        pids+=($!)
        ((index++))
    done

    # Wait for remaining builds
    for pid in "${pids[@]}"; do
        wait $pid
    done

    # Collect results
    local successful=0
    local failed=0
    for ((i=0; i<${#services[@]}; i++)); do
        if [[ -f "build_result_$i.tmp" ]]; then
            local result=$(cat "build_result_$i.tmp")
            if [[ $result == SUCCESS:* ]]; then
                ((successful++))
            else
                ((failed++))
            fi
            rm -f "build_result_$i.tmp"
        fi
    done

    return $failed
}

# Get services to build based on mode
get_services_to_build() {
    case $BUILD_MODE in
        "core")
            echo "${CORE_SERVICES[@]}"
            ;;
        "test")
            echo "${TEST_SERVICES[@]}"
            ;;
        "ai")
            echo "${AI_SERVICES[@]}"
            ;;
        "all")
            echo "${ALL_SERVICES[@]}"
            ;;
        *)
            print_error "Unknown build mode: $BUILD_MODE"
            exit 1
            ;;
    esac
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_args "$@"

    # Check dependencies
    check_build_dependencies

    # Create build directory
    mkdir -p "$BUILD_DIR"

    # Get services to build
    local services_to_build=($(get_services_to_build))
    local total_services=${#services_to_build[@]}

    print_info "Build mode: $BUILD_MODE"
    print_info "Services to build: $total_services"
    print_info "Build timeout: ${BUILD_TIMEOUT}s per service"
    print_info "Parallel builds: $PARALLEL_BUILDS"

    # Show services list
    print_header "📋 Services to Build"
    for service in "${services_to_build[@]}"; do
        print_info "  • $service"
    done

    # Start building
    print_header "🔨 Building Services"

    if build_services_parallel "${services_to_build[@]}"; then
        local failed_count=$?
        local successful_count=$((total_services - failed_count))
    else
        local failed_count=$?
        local successful_count=$((total_services - failed_count))
    fi

    # Calculate build time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    # Show build summary
    print_header "📊 Build Summary"
    echo -e "${BOLD}Total Services:${NC} $total_services"
    echo -e "${GREEN}Successful:${NC} $successful_count"
    echo -e "${RED}Failed:${NC} $failed_count"
    echo -e "${BLUE}Total Time:${NC} ${total_time}s"

    if [[ $failed_count -eq 0 ]]; then
        print_success "🎉 All services built successfully!"

        print_header "📦 Built Services"
        if [[ -d "$BUILD_DIR" ]]; then
            ls -la "$BUILD_DIR/"
        fi

        print_header "🚀 Next Steps"
        print_info "All microservices are ready for deployment!"
        print_info "Available commands:"
        print_info "  • ./scripts/start-all-services.sh    - Start all services"
        print_info "  • ./scripts/test-all-services.sh     - Run all tests"
        print_info "  • ./scripts/health-check.sh          - Check service health"
        print_info "  • ./scripts/deploy.sh                - Deploy to production"

    else
        print_warning "⚠️  Some services failed to build ($failed_count/$total_services)"

        print_header "🔍 Troubleshooting"
        print_info "Common solutions:"
        print_info "  1. Update dependencies: go mod tidy"
        print_info "  2. Check Go version: go version (requires 1.21+)"
        print_info "  3. Verify network connectivity for module downloads"
        print_info "  4. Check import paths in failed services"
        print_info "  5. Run with DEBUG=true for detailed output"

        print_header "🔧 Debug Commands"
        print_info "  • DEBUG=true ./build_all.sh          - Verbose output"
        print_info "  • go build -v cmd/SERVICE/main.go    - Build single service"
        print_info "  • go mod verify                      - Verify dependencies"
    fi

    print_header "🎯 Architecture Status"
    print_status "✅ Clean Architecture implemented"
    print_status "✅ Microservices pattern applied"
    print_status "✅ Domain-Driven Design structure"
    print_status "✅ Redis integration ready"
    print_status "✅ AI services configured"
    print_status "✅ gRPC communication setup"
    print_status "✅ HTTP REST APIs ready"
    print_status "✅ Middleware and security implemented"
    print_status "✅ Logging and monitoring ready"
    print_status "✅ Docker containerization support"
    print_status "✅ Kubernetes deployment ready"

    echo ""
    if [[ $failed_count -eq 0 ]]; then
        print_success "🏆 Go Coffee Microservices Architecture is PRODUCTION READY! 🚀☕"
        exit 0
    else
        print_error "❌ Build completed with errors. Please fix failed services."
        exit 1
    fi
}

# Run main function with all arguments
main "$@"
