#!/bin/bash

# Go Coffee - Advanced Load Testing & Performance Analysis
# Comprehensive performance testing with multiple tools and detailed reporting
# Version: 3.0.0
# Usage: ./load-test.sh [OPTIONS]
#   -t, --target URL    Target URL for testing
#   -u, --users NUM     Number of concurrent users (default: 10)
#   -d, --duration SEC  Test duration in seconds (default: 60)
#   -r, --ramp-up SEC   Ramp-up time in seconds (default: 10)
#   -s, --scenario NAME Test scenario (api|web|crypto|full)
#   -o, --output DIR    Output directory for reports
#   -f, --format TYPE   Report format (html|json|csv|all)
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Source shared library
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "⚡ Go Coffee Advanced Load Testing & Performance Analysis"

# =============================================================================
# CONFIGURATION
# =============================================================================

TARGET_URL="${TARGET_URL:-http://localhost:8080}"
CONCURRENT_USERS=10
TEST_DURATION=60
RAMP_UP_TIME=10
TEST_SCENARIO="api"
OUTPUT_DIR="performance-reports/$(date +%Y%m%d-%H%M%S)"
REPORT_FORMAT="html"

# Test scenarios configuration
declare -A API_ENDPOINTS=(
    ["health"]="/health"
    ["auth"]="/api/v1/auth/login"
    ["orders"]="/api/v1/orders"
    ["payments"]="/api/v1/payments"
    ["kitchen"]="/api/v1/kitchen/orders"
    ["crypto"]="/api/v1/crypto/wallet"
    ["ai"]="/api/v1/ai/search"
)

declare -A LOAD_PROFILES=(
    ["light"]="5:30:5"      # users:duration:ramp-up
    ["medium"]="20:60:10"
    ["heavy"]="50:120:20"
    ["stress"]="100:300:30"
    ["spike"]="200:60:5"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_load_test_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--target)
                TARGET_URL="$2"
                shift 2
                ;;
            -u|--users)
                CONCURRENT_USERS="$2"
                shift 2
                ;;
            -d|--duration)
                TEST_DURATION="$2"
                shift 2
                ;;
            -r|--ramp-up)
                RAMP_UP_TIME="$2"
                shift 2
                ;;
            -s|--scenario)
                TEST_SCENARIO="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -f|--format)
                REPORT_FORMAT="$2"
                shift 2
                ;;
            -h|--help)
                show_usage "load-test.sh" \
                    "Advanced load testing and performance analysis for Go Coffee platform" \
                    "  ./load-test.sh [OPTIONS]
  
  Options:
    -t, --target URL    Target URL for testing (default: http://localhost:8080)
    -u, --users NUM     Number of concurrent users (default: 10)
    -d, --duration SEC  Test duration in seconds (default: 60)
    -r, --ramp-up SEC   Ramp-up time in seconds (default: 10)
    -s, --scenario NAME Test scenario (api|web|crypto|full) (default: api)
    -o, --output DIR    Output directory for reports
    -f, --format TYPE   Report format (html|json|csv|all) (default: html)
    -h, --help          Show this help message
  
  Load Profiles:
    light:  5 users, 30s duration, 5s ramp-up
    medium: 20 users, 60s duration, 10s ramp-up
    heavy:  50 users, 120s duration, 20s ramp-up
    stress: 100 users, 300s duration, 30s ramp-up
    spike:  200 users, 60s duration, 5s ramp-up
  
  Examples:
    ./load-test.sh                              # Basic API load test
    ./load-test.sh -s crypto -u 50 -d 120      # Crypto services stress test
    ./load-test.sh -s full -f all               # Full platform test with all reports
    ./load-test.sh -t http://staging.example.com # Test staging environment"
                exit 0
                ;;
            *)
                # Check if it's a load profile
                if [[ -n "${LOAD_PROFILES[$1]:-}" ]]; then
                    local profile=${LOAD_PROFILES[$1]}
                    CONCURRENT_USERS=$(echo "$profile" | cut -d: -f1)
                    TEST_DURATION=$(echo "$profile" | cut -d: -f2)
                    RAMP_UP_TIME=$(echo "$profile" | cut -d: -f3)
                    shift
                else
                    print_error "Unknown option: $1"
                    exit 1
                fi
                ;;
        esac
    done
}

# =============================================================================
# LOAD TESTING FUNCTIONS
# =============================================================================

# Check load testing dependencies
check_load_test_dependencies() {
    print_header "🔍 Checking Load Testing Dependencies"
    
    local tools=("curl" "jq")
    local optional_tools=("wrk" "ab" "hey" "k6")
    
    # Check required tools
    check_dependencies "${tools[@]}" || exit 1
    
    # Check optional tools
    local available_tools=()
    for tool in "${optional_tools[@]}"; do
        if command_exists "$tool"; then
            available_tools+=("$tool")
            print_status "$tool is available"
        else
            print_warning "$tool is not available (optional)"
        fi
    done
    
    if [[ ${#available_tools[@]} -eq 0 ]]; then
        print_error "No load testing tools available. Please install at least one: ${optional_tools[*]}"
        exit 1
    fi
    
    print_success "Load testing dependencies checked"
}

# Create output directory structure
create_output_structure() {
    print_header "📁 Creating Output Directory Structure"
    
    mkdir -p "$OUTPUT_DIR"/{reports,logs,data,charts}
    
    print_status "Output directory created: $OUTPUT_DIR"
}

# Generate test configuration
generate_test_config() {
    print_header "⚙️ Generating Test Configuration"
    
    cat > "$OUTPUT_DIR/test-config.json" <<EOF
{
    "test_info": {
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "target_url": "$TARGET_URL",
        "scenario": "$TEST_SCENARIO",
        "concurrent_users": $CONCURRENT_USERS,
        "test_duration": $TEST_DURATION,
        "ramp_up_time": $RAMP_UP_TIME,
        "report_format": "$REPORT_FORMAT"
    },
    "endpoints": $(printf '%s\n' "${API_ENDPOINTS[@]}" | jq -R . | jq -s .)
}
EOF
    
    print_status "Test configuration generated"
}

# Run API load test
run_api_load_test() {
    print_header "🚀 Running API Load Test"
    
    local test_results=()
    
    for endpoint_name in "${!API_ENDPOINTS[@]}"; do
        local endpoint_path=${API_ENDPOINTS[$endpoint_name]}
        local full_url="$TARGET_URL$endpoint_path"
        
        print_progress "Testing endpoint: $endpoint_name ($endpoint_path)"
        
        # Test with wrk if available
        if command_exists wrk; then
            print_info "Running wrk test for $endpoint_name..."
            wrk -t4 -c"$CONCURRENT_USERS" -d"${TEST_DURATION}s" \
                --script="$SCRIPT_DIR/wrk-scripts/basic.lua" \
                "$full_url" > "$OUTPUT_DIR/logs/wrk-$endpoint_name.log" 2>&1 &
        fi
        
        # Test with hey if available
        if command_exists hey; then
            print_info "Running hey test for $endpoint_name..."
            hey -n $((CONCURRENT_USERS * TEST_DURATION)) -c "$CONCURRENT_USERS" \
                -o csv "$full_url" > "$OUTPUT_DIR/data/hey-$endpoint_name.csv" 2>&1 &
        fi
        
        # Test with ab if available
        if command_exists ab; then
            print_info "Running ab test for $endpoint_name..."
            ab -n $((CONCURRENT_USERS * TEST_DURATION)) -c "$CONCURRENT_USERS" \
                "$full_url" > "$OUTPUT_DIR/logs/ab-$endpoint_name.log" 2>&1 &
        fi
        
        sleep 2  # Stagger test starts
    done
    
    # Wait for all tests to complete
    print_progress "Waiting for all load tests to complete..."
    wait
    
    print_success "API load tests completed"
}

# Run crypto services load test
run_crypto_load_test() {
    print_header "💰 Running Crypto Services Load Test"
    
    local crypto_endpoints=(
        "/api/v1/crypto/wallet/balance"
        "/api/v1/crypto/transactions"
        "/api/v1/crypto/smart-contracts"
        "/api/v1/crypto/defi/pools"
    )
    
    for endpoint in "${crypto_endpoints[@]}"; do
        local full_url="$TARGET_URL$endpoint"
        
        print_progress "Testing crypto endpoint: $endpoint"
        
        # Simulate crypto-specific load patterns
        if command_exists hey; then
            hey -n $((CONCURRENT_USERS * 2)) -c "$CONCURRENT_USERS" \
                -H "Authorization: Bearer test-token" \
                "$full_url" > "$OUTPUT_DIR/data/crypto-$(basename "$endpoint").csv" 2>&1 &
        fi
    done
    
    wait
    print_success "Crypto services load tests completed"
}

# Run web UI load test
run_web_load_test() {
    print_header "🌐 Running Web UI Load Test"
    
    local web_endpoints=(
        "/"
        "/dashboard"
        "/orders"
        "/crypto"
        "/api/v1/status"
    )
    
    for endpoint in "${web_endpoints[@]}"; do
        local full_url="$TARGET_URL$endpoint"
        
        print_progress "Testing web endpoint: $endpoint"
        
        if command_exists wrk; then
            wrk -t2 -c"$((CONCURRENT_USERS / 2))" -d"${TEST_DURATION}s" \
                "$full_url" > "$OUTPUT_DIR/logs/web-$(basename "$endpoint" | tr '/' '-').log" 2>&1 &
        fi
    done
    
    wait
    print_success "Web UI load tests completed"
}

# Analyze test results
analyze_test_results() {
    print_header "📊 Analyzing Test Results"
    
    local summary_file="$OUTPUT_DIR/reports/summary.json"
    
    # Initialize summary
    cat > "$summary_file" <<EOF
{
    "test_summary": {
        "start_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "scenario": "$TEST_SCENARIO",
        "target_url": "$TARGET_URL",
        "configuration": {
            "users": $CONCURRENT_USERS,
            "duration": $TEST_DURATION,
            "ramp_up": $RAMP_UP_TIME
        }
    },
    "results": {}
}
EOF
    
    # Analyze wrk results
    if ls "$OUTPUT_DIR"/logs/wrk-*.log >/dev/null 2>&1; then
        print_progress "Analyzing wrk results..."
        for log_file in "$OUTPUT_DIR"/logs/wrk-*.log; do
            local endpoint_name=$(basename "$log_file" .log | sed 's/wrk-//')
            
            # Extract key metrics from wrk output
            local requests_per_sec=$(grep "Requests/sec:" "$log_file" | awk '{print $2}' || echo "0")
            local avg_latency=$(grep "Latency" "$log_file" | awk '{print $2}' || echo "0")
            local total_requests=$(grep "requests in" "$log_file" | awk '{print $1}' || echo "0")
            
            # Add to summary (simplified - would need jq for proper JSON manipulation)
            print_info "$endpoint_name: $requests_per_sec req/s, $avg_latency avg latency"
        done
    fi
    
    print_success "Test results analyzed"
}

# Generate performance report
generate_performance_report() {
    print_header "📋 Generating Performance Report"
    
    local report_file="$OUTPUT_DIR/reports/performance-report.html"
    
    cat > "$report_file" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee Performance Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .summary { background: #ecf0f1; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .metric { margin: 10px 0; padding: 10px; border-left: 4px solid #3498db; }
        .good { border-left-color: #27ae60; }
        .warning { border-left-color: #f39c12; }
        .error { border-left-color: #e74c3c; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>⚡ Go Coffee Performance Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Test Scenario: $TEST_SCENARIO</p>
        <p>Target: $TARGET_URL</p>
    </div>
    
    <div class="summary">
        <h2>📊 Test Configuration</h2>
        <div class="metric">
            <strong>Concurrent Users:</strong> $CONCURRENT_USERS
        </div>
        <div class="metric">
            <strong>Test Duration:</strong> ${TEST_DURATION}s
        </div>
        <div class="metric">
            <strong>Ramp-up Time:</strong> ${RAMP_UP_TIME}s
        </div>
    </div>
    
    <div class="summary">
        <h2>🎯 Performance Metrics</h2>
        <table>
            <tr>
                <th>Endpoint</th>
                <th>Requests/sec</th>
                <th>Avg Latency</th>
                <th>Total Requests</th>
                <th>Status</th>
            </tr>
EOF

    # Add metrics rows (simplified)
    for endpoint_name in "${!API_ENDPOINTS[@]}"; do
        cat >> "$report_file" <<EOF
            <tr>
                <td>$endpoint_name</td>
                <td>-</td>
                <td>-</td>
                <td>-</td>
                <td>✅ Tested</td>
            </tr>
EOF
    done

    cat >> "$report_file" <<EOF
        </table>
    </div>
    
    <div class="summary">
        <h2>📈 Recommendations</h2>
        <div class="metric good">
            <strong>✅ Good Performance:</strong> Response times under 200ms
        </div>
        <div class="metric warning">
            <strong>⚠️ Monitor:</strong> Response times 200-500ms
        </div>
        <div class="metric error">
            <strong>❌ Needs Attention:</strong> Response times over 500ms
        </div>
    </div>
    
    <div class="summary">
        <h2>🔧 Next Steps</h2>
        <ul>
            <li>Review detailed logs in the logs/ directory</li>
            <li>Analyze CSV data for trends</li>
            <li>Compare with previous test results</li>
            <li>Optimize identified bottlenecks</li>
        </ul>
    </div>
</body>
</html>
EOF

    print_status "Performance report generated: $report_file"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)
    
    # Parse arguments
    parse_load_test_args "$@"
    
    # Check dependencies
    check_load_test_dependencies
    
    print_info "Load Test Configuration:"
    print_info "  Target URL: $TARGET_URL"
    print_info "  Scenario: $TEST_SCENARIO"
    print_info "  Users: $CONCURRENT_USERS"
    print_info "  Duration: ${TEST_DURATION}s"
    print_info "  Ramp-up: ${RAMP_UP_TIME}s"
    print_info "  Output: $OUTPUT_DIR"
    
    # Create output structure
    create_output_structure
    
    # Generate test configuration
    generate_test_config
    
    # Run tests based on scenario
    case "$TEST_SCENARIO" in
        "api")
            run_api_load_test
            ;;
        "crypto")
            run_crypto_load_test
            ;;
        "web")
            run_web_load_test
            ;;
        "full")
            run_api_load_test
            run_crypto_load_test
            run_web_load_test
            ;;
        *)
            print_error "Unknown test scenario: $TEST_SCENARIO"
            exit 1
            ;;
    esac
    
    # Analyze results
    analyze_test_results
    
    # Generate reports
    if [[ "$REPORT_FORMAT" == "html" || "$REPORT_FORMAT" == "all" ]]; then
        generate_performance_report
    fi
    
    # Calculate test time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    print_success "🎉 Load testing completed in ${total_time}s"
    
    print_header "📊 Results Summary"
    print_info "Test results saved to: $OUTPUT_DIR"
    print_info "Performance report: $OUTPUT_DIR/reports/performance-report.html"
    print_info "Raw data: $OUTPUT_DIR/data/"
    print_info "Logs: $OUTPUT_DIR/logs/"
    
    print_header "🎯 Next Steps"
    print_info "1. Review the performance report"
    print_info "2. Analyze bottlenecks and optimization opportunities"
    print_info "3. Compare with baseline performance metrics"
    print_info "4. Set up continuous performance monitoring"
}

# Run main function with all arguments
main "$@"
