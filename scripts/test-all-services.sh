#!/bin/bash

# Go Coffee - Comprehensive Test Suite
# Tests all microservices with coverage reporting and parallel execution
# Version: 2.0.0
# Usage: ./test-all-services.sh [OPTIONS]
#   -c, --core-only     Test only core production services
#   -t, --test-only     Test only test services
#   -a, --ai-only       Test only AI services
#   -f, --fast          Skip integration tests, run unit tests only
#   -v, --verbose       Verbose output with detailed test results
#   -p, --parallel      Run tests in parallel (default: sequential)
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory for relative imports
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared library
source "$SCRIPT_DIR/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🧪 Go Coffee Comprehensive Test Suite"

# =============================================================================
# CONFIGURATION
# =============================================================================

TEST_TIMEOUT=300
COVERAGE_THRESHOLD=70
TEST_MODE="all"
FAST_MODE=false
VERBOSE_MODE=false
PARALLEL_MODE=false
COVERAGE_DIR="coverage"
TEST_RESULTS_DIR="test-results"

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--core-only)
                TEST_MODE="core"
                shift
                ;;
            -t|--test-only)
                TEST_MODE="test"
                shift
                ;;
            -a|--ai-only)
                TEST_MODE="ai"
                shift
                ;;
            -f|--fast)
                FAST_MODE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE_MODE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL_MODE=true
                shift
                ;;
            -h|--help)
                show_usage "test-all-services.sh" \
                    "Comprehensive test suite for all Go Coffee microservices with coverage reporting" \
                    "  ./test-all-services.sh [OPTIONS]

  Options:
    -c, --core-only     Test only core production services (${#CORE_SERVICES[@]} services)
    -t, --test-only     Test only test services (${#TEST_SERVICES[@]} services)
    -a, --ai-only       Test only AI services (${#AI_SERVICES[@]} services)
    -f, --fast          Skip integration tests, run unit tests only
    -v, --verbose       Verbose output with detailed test results
    -p, --parallel      Run tests in parallel (faster but less readable output)
    -h, --help          Show this help message

  Examples:
    ./test-all-services.sh                    # Test all services
    ./test-all-services.sh --core-only        # Test only production services
    ./test-all-services.sh --fast --parallel  # Fast parallel testing
    ./test-all-services.sh --ai-only -v       # Verbose AI service testing"
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
# TEST FUNCTIONS
# =============================================================================

# Enhanced test service function with coverage and detailed reporting
test_service_enhanced() {
    local service_name=$1
    local test_start=$(date +%s)
    local service_path="cmd/$service_name"
    local coverage_file="$COVERAGE_DIR/${service_name}.out"
    local result_file="$TEST_RESULTS_DIR/${service_name}.json"

    # Check if service exists
    if [[ ! -d "$service_path" ]]; then
        print_warning "$service_name: Service directory not found"
        return 1
    fi

    print_progress "Testing $service_name..."

    # Create coverage and results directories
    mkdir -p "$COVERAGE_DIR" "$TEST_RESULTS_DIR"

    # Determine test paths based on service structure
    local test_paths="./..."
    if [[ -d "internal/$service_name" ]]; then
        test_paths="./internal/$service_name/... ./cmd/$service_name/..."
    fi

    # Build test command
    local test_cmd="go test"
    local test_args=("-timeout=${TEST_TIMEOUT}s" "-race")

    if [[ "$VERBOSE_MODE" == "true" ]]; then
        test_args+=("-v")
    fi

    if [[ "$FAST_MODE" == "false" ]]; then
        test_args+=("-coverprofile=$coverage_file" "-covermode=atomic")
    fi

    test_args+=("$test_paths")

    # Run tests with timeout and capture output
    local test_output
    local test_result=0

    if test_output=$($test_cmd "${test_args[@]}" 2>&1); then
        local test_end=$(date +%s)
        local test_time=$((test_end - test_start))

        # Extract test statistics
        local tests_run=$(echo "$test_output" | grep -c "^=== RUN" || echo "0")
        local tests_passed=$(echo "$test_output" | grep -c "--- PASS:" || echo "0")
        local tests_failed=$(echo "$test_output" | grep -c "--- FAIL:" || echo "0")

        # Calculate coverage if available
        local coverage="N/A"
        if [[ -f "$coverage_file" ]]; then
            coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' || echo "N/A")
        fi

        # Create result JSON
        cat > "$result_file" <<EOF
{
    "service": "$service_name",
    "status": "PASSED",
    "duration": ${test_time},
    "tests_run": ${tests_run},
    "tests_passed": ${tests_passed},
    "tests_failed": ${tests_failed},
    "coverage": "$coverage",
    "timestamp": "$(date -Iseconds)"
}
EOF

        print_status "$service_name: PASSED (${test_time}s, coverage: $coverage)"
        if [[ "$VERBOSE_MODE" == "true" ]]; then
            print_debug "Tests run: $tests_run, Passed: $tests_passed, Failed: $tests_failed"
        fi
        return 0
    else
        local test_end=$(date +%s)
        local test_time=$((test_end - test_start))

        # Create failure result JSON
        cat > "$result_file" <<EOF
{
    "service": "$service_name",
    "status": "FAILED",
    "duration": ${test_time},
    "error": "Test execution failed",
    "timestamp": "$(date -Iseconds)"
}
EOF

        print_error "$service_name: FAILED (${test_time}s)"
        if [[ "$VERBOSE_MODE" == "true" ]]; then
            print_debug "Test output: $test_output"
        fi
        return 1
    fi
}

# Test services in parallel
test_services_parallel() {
    local services=("$@")
    local pids=()
    local max_parallel=4

    print_info "Testing ${#services[@]} services in parallel (max $max_parallel concurrent)..."

    # Function to test single service in background
    test_single() {
        local service=$1
        local index=$2
        if test_service_enhanced "$service"; then
            echo "SUCCESS:$index:$service" > "test_result_$index.tmp"
        else
            echo "FAILED:$index:$service" > "test_result_$index.tmp"
        fi
    }

    # Start tests in batches
    local index=0
    for service in "${services[@]}"; do
        # Wait if we've reached max parallel tests
        if [[ ${#pids[@]} -ge $max_parallel ]]; then
            wait ${pids[0]}
            pids=("${pids[@]:1}")
        fi

        # Start test in background
        test_single "$service" $index &
        pids+=($!)
        ((index++))
    done

    # Wait for remaining tests
    for pid in "${pids[@]}"; do
        wait $pid
    done

    # Collect results
    local successful=0
    local failed=0
    for ((i=0; i<${#services[@]}; i++)); do
        if [[ -f "test_result_$i.tmp" ]]; then
            local result=$(cat "test_result_$i.tmp")
            if [[ $result == SUCCESS:* ]]; then
                ((successful++))
            else
                ((failed++))
            fi
            rm -f "test_result_$i.tmp"
        fi
    done

    return $failed
}

# Test services sequentially
test_services_sequential() {
    local services=("$@")
    local successful=0
    local failed=0

    print_info "Testing ${#services[@]} services sequentially..."

    for service in "${services[@]}"; do
        if test_service_enhanced "$service"; then
            ((successful++))
        else
            ((failed++))
        fi
    done

    return $failed
}

# Get services to test based on mode
get_services_to_test() {
    case $TEST_MODE in
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
            print_error "Unknown test mode: $TEST_MODE"
            exit 1
            ;;
    esac
}

# =============================================================================
# REPORTING FUNCTIONS
# =============================================================================

# Generate comprehensive test report
generate_test_report() {
    local total_services=$1
    local successful_services=$2
    local failed_services=$3
    local total_time=$4

    local report_file="$TEST_RESULTS_DIR/summary.json"
    local html_report="$TEST_RESULTS_DIR/report.html"

    print_header "📊 Generating Test Report"

    # Create JSON summary
    cat > "$report_file" <<EOF
{
    "summary": {
        "total_services": $total_services,
        "successful": $successful_services,
        "failed": $failed_services,
        "success_rate": $(( successful_services * 100 / total_services )),
        "total_duration": $total_time,
        "timestamp": "$(date -Iseconds)",
        "test_mode": "$TEST_MODE",
        "fast_mode": $FAST_MODE,
        "parallel_mode": $PARALLEL_MODE
    },
    "services": [
EOF

    # Add individual service results
    local first=true
    for result_file in "$TEST_RESULTS_DIR"/*.json; do
        if [[ -f "$result_file" && "$result_file" != "$report_file" ]]; then
            if [[ "$first" == "true" ]]; then
                first=false
            else
                echo "," >> "$report_file"
            fi
            cat "$result_file" >> "$report_file"
        fi
    done

    echo -e "\n    ]\n}" >> "$report_file"

    print_status "Test report generated: $report_file"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_args "$@"

    # Check dependencies
    local deps=("go" "timeout")
    if [[ "$FAST_MODE" == "false" ]]; then
        deps+=("curl")
    fi
    check_dependencies "${deps[@]}" || exit 1

    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run from project root."
        exit 1
    fi

    # Create necessary directories
    mkdir -p "$COVERAGE_DIR" "$TEST_RESULTS_DIR"

    # Get services to test
    local services_to_test=($(get_services_to_test))
    local total_services=${#services_to_test[@]}

    print_info "Test mode: $TEST_MODE"
    print_info "Services to test: $total_services"
    print_info "Test timeout: ${TEST_TIMEOUT}s per service"
    print_info "Fast mode: $FAST_MODE"
    print_info "Parallel mode: $PARALLEL_MODE"
    print_info "Verbose mode: $VERBOSE_MODE"

    # Show services list
    print_header "📋 Services to Test"
    for service in "${services_to_test[@]}"; do
        print_info "  • $service"
    done

    # Start testing
    print_header "🧪 Running Tests"

    local failed_count=0
    if [[ "$PARALLEL_MODE" == "true" ]]; then
        test_services_parallel "${services_to_test[@]}"
        failed_count=$?
    else
        test_services_sequential "${services_to_test[@]}"
        failed_count=$?
    fi

    local successful_count=$((total_services - failed_count))

    # Calculate total time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    # Generate comprehensive report
    generate_test_report $total_services $successful_count $failed_count $total_time

    # Show final summary
    print_header "📊 Final Test Results"
    echo -e "${BOLD}Total Services:${NC} $total_services"
    echo -e "${GREEN}Successful:${NC} $successful_count"
    echo -e "${RED}Failed:${NC} $failed_count"
    echo -e "${BLUE}Success Rate:${NC} $(( successful_count * 100 / total_services ))%"
    echo -e "${BLUE}Total Time:${NC} ${total_time}s"

    if [[ $failed_count -eq 0 ]]; then
        print_success "🎉 ALL SERVICES PASSED! Platform is ready for deployment!"

        print_header "🚀 Next Steps"
        print_info "All tests passed successfully. You can now:"
        print_info "  • Deploy to staging: ./scripts/deploy.sh --staging"
        print_info "  • Run integration tests: ./scripts/test-integration.sh"
        print_info "  • Start all services: ./scripts/start-all-services.sh"
        print_info "  • Check service health: ./scripts/health-check.sh"

        exit 0
    else
        print_warning "⚠️  Some services failed testing ($failed_count/$total_services)"

        print_header "🔍 Troubleshooting"
        print_info "Services with test failures need attention."

        print_header "🔧 Debug Commands"
        print_info "  • View detailed report: cat $TEST_RESULTS_DIR/summary.json"
        print_info "  • Test single service: go test -v ./cmd/SERVICE/..."
        print_info "  • Run with verbose: ./test-all-services.sh -v"
        print_info "  • Check coverage: go tool cover -html=coverage/SERVICE.out"

        exit 1
    fi
}

# Run main function with all arguments
main "$@"
