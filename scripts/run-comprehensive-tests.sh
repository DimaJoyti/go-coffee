#!/bin/bash

# Comprehensive Test Suite for Go Coffee Enterprise Platform
# This script runs all test categories: unit, integration, load, security, and end-to-end tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT="${ENVIRONMENT:-test}"
PARALLEL_TESTS="${PARALLEL_TESTS:-true}"
COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-80}"
LOAD_TEST_DURATION="${LOAD_TEST_DURATION:-300s}"
LOAD_TEST_USERS="${LOAD_TEST_USERS:-100}"
SKIP_SLOW_TESTS="${SKIP_SLOW_TESTS:-false}"
GENERATE_REPORTS="${GENERATE_REPORTS:-true}"

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to print colored output
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

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================${NC}"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Function to update test counters
update_test_results() {
    local result=$1
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    case $result in
        "pass")
            PASSED_TESTS=$((PASSED_TESTS + 1))
            ;;
        "fail")
            FAILED_TESTS=$((FAILED_TESTS + 1))
            ;;
        "skip")
            SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
            ;;
    esac
}

# Function to run unit tests
run_unit_tests() {
    print_header "Running Unit Tests"
    
    print_step "Setting up test environment..."
    export GO_ENV=test
    export DATABASE_URL="postgres://test:test@localhost:5432/go_coffee_test?sslmode=disable"
    export REDIS_URL="redis://localhost:6379/1"
    
    # Create test reports directory
    mkdir -p test-reports/unit
    
    print_step "Running Go unit tests with coverage..."
    if go test -v -race -coverprofile=test-reports/unit/coverage.out -covermode=atomic ./... > test-reports/unit/results.txt 2>&1; then
        print_success "Unit tests passed"
        update_test_results "pass"
        
        # Generate coverage report
        go tool cover -html=test-reports/unit/coverage.out -o test-reports/unit/coverage.html
        
        # Check coverage threshold
        local coverage=$(go tool cover -func=test-reports/unit/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
            print_success "Coverage threshold met: ${coverage}% >= ${COVERAGE_THRESHOLD}%"
        else
            print_warning "Coverage below threshold: ${coverage}% < ${COVERAGE_THRESHOLD}%"
        fi
    else
        print_error "Unit tests failed"
        update_test_results "fail"
        cat test-reports/unit/results.txt
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running Integration Tests"
    
    if [[ "$SKIP_SLOW_TESTS" == "true" ]]; then
        print_warning "Skipping integration tests (SKIP_SLOW_TESTS=true)"
        update_test_results "skip"
        return 0
    fi
    
    print_step "Starting test infrastructure..."
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for services to be ready
    print_step "Waiting for test services to be ready..."
    sleep 30
    
    # Check service health
    local services=("postgres" "redis" "kafka")
    for service in "${services[@]}"; do
        if ! docker-compose -f docker-compose.test.yml ps "$service" | grep -q "Up"; then
            print_error "$service is not running"
            update_test_results "fail"
            return 1
        fi
    done
    
    mkdir -p test-reports/integration
    
    print_step "Running integration tests..."
    if go test -v -tags=integration ./test/integration/... > test-reports/integration/results.txt 2>&1; then
        print_success "Integration tests passed"
        update_test_results "pass"
    else
        print_error "Integration tests failed"
        update_test_results "fail"
        cat test-reports/integration/results.txt
    fi
    
    print_step "Cleaning up test infrastructure..."
    docker-compose -f docker-compose.test.yml down
}

# Function to run end-to-end tests
run_e2e_tests() {
    print_header "Running End-to-End Tests"
    
    if [[ "$SKIP_SLOW_TESTS" == "true" ]]; then
        print_warning "Skipping E2E tests (SKIP_SLOW_TESTS=true)"
        update_test_results "skip"
        return 0
    fi
    
    print_step "Starting full platform for E2E tests..."
    ./scripts/start-core-services.sh
    
    # Wait for all services to be ready
    print_step "Waiting for platform to be ready..."
    sleep 60
    
    mkdir -p test-reports/e2e
    
    # Run E2E test scenarios
    local scenarios=("order-processing" "payment-flow" "ai-workflow" "analytics-pipeline")
    
    for scenario in "${scenarios[@]}"; do
        print_step "Running E2E scenario: $scenario"
        
        if go test -v -tags=e2e ./test/e2e/"$scenario"_test.go > test-reports/e2e/"$scenario".txt 2>&1; then
            print_success "E2E scenario '$scenario' passed"
            update_test_results "pass"
        else
            print_error "E2E scenario '$scenario' failed"
            update_test_results "fail"
            cat test-reports/e2e/"$scenario".txt
        fi
    done
    
    print_step "Stopping platform..."
    ./scripts/stop-core-services.sh
}

# Function to run load tests
run_load_tests() {
    print_header "Running Load Tests"
    
    if [[ "$SKIP_SLOW_TESTS" == "true" ]]; then
        print_warning "Skipping load tests (SKIP_SLOW_TESTS=true)"
        update_test_results "skip"
        return 0
    fi
    
    # Check if k6 is installed
    if ! command -v k6 &> /dev/null; then
        print_warning "k6 not installed, skipping load tests"
        update_test_results "skip"
        return 0
    fi
    
    print_step "Starting platform for load testing..."
    ./scripts/start-core-services.sh
    
    # Wait for services to be ready
    sleep 60
    
    mkdir -p test-reports/load
    
    # Run load tests for different services
    local services=("producer" "web3-payment" "ai-orchestrator" "analytics")
    
    for service in "${services[@]}"; do
        print_step "Running load test for $service service..."
        
        if k6 run \
            --duration="$LOAD_TEST_DURATION" \
            --vus="$LOAD_TEST_USERS" \
            --out json=test-reports/load/"$service"-results.json \
            test/load/"$service"-load-test.js > test-reports/load/"$service".txt 2>&1; then
            print_success "Load test for '$service' completed"
            update_test_results "pass"
        else
            print_error "Load test for '$service' failed"
            update_test_results "fail"
            cat test-reports/load/"$service".txt
        fi
    done
    
    print_step "Stopping platform..."
    ./scripts/stop-core-services.sh
}

# Function to run security tests
run_security_tests() {
    print_header "Running Security Tests"
    
    mkdir -p test-reports/security
    
    # Static security analysis
    print_step "Running static security analysis..."
    if command -v gosec &> /dev/null; then
        if gosec -fmt json -out test-reports/security/gosec-report.json ./...; then
            print_success "Static security analysis completed"
            update_test_results "pass"
        else
            print_warning "Static security analysis found issues"
            update_test_results "fail"
        fi
    else
        print_warning "gosec not installed, skipping static analysis"
        update_test_results "skip"
    fi
    
    # Dependency vulnerability scan
    print_step "Running dependency vulnerability scan..."
    if command -v nancy &> /dev/null; then
        if go list -json -m all | nancy sleuth > test-reports/security/nancy-report.txt 2>&1; then
            print_success "Dependency scan completed"
            update_test_results "pass"
        else
            print_warning "Dependency vulnerabilities found"
            update_test_results "fail"
        fi
    else
        print_warning "nancy not installed, skipping dependency scan"
        update_test_results "skip"
    fi
    
    # Container security scan (if Docker images exist)
    print_step "Running container security scan..."
    if command -v trivy &> /dev/null; then
        local images=("go-coffee/producer" "go-coffee/web3-payment" "go-coffee/ai-orchestrator" "go-coffee/analytics")
        
        for image in "${images[@]}"; do
            if docker images | grep -q "$image"; then
                if trivy image --format json --output test-reports/security/"$(basename "$image")"-trivy.json "$image:latest"; then
                    print_success "Container scan for '$image' completed"
                    update_test_results "pass"
                else
                    print_warning "Container vulnerabilities found in '$image'"
                    update_test_results "fail"
                fi
            fi
        done
    else
        print_warning "trivy not installed, skipping container scan"
        update_test_results "skip"
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_header "Running Performance Tests"
    
    mkdir -p test-reports/performance
    
    print_step "Running Go benchmark tests..."
    if go test -bench=. -benchmem -cpuprofile=test-reports/performance/cpu.prof -memprofile=test-reports/performance/mem.prof ./... > test-reports/performance/benchmark.txt 2>&1; then
        print_success "Performance benchmarks completed"
        update_test_results "pass"
        
        # Generate performance profiles
        go tool pprof -http=:8080 test-reports/performance/cpu.prof &
        PPROF_PID=$!
        sleep 2
        kill $PPROF_PID 2>/dev/null || true
    else
        print_error "Performance benchmarks failed"
        update_test_results "fail"
        cat test-reports/performance/benchmark.txt
    fi
}

# Function to generate test reports
generate_test_reports() {
    print_header "Generating Test Reports"
    
    if [[ "$GENERATE_REPORTS" != "true" ]]; then
        print_status "Report generation disabled"
        return 0
    fi
    
    mkdir -p test-reports/summary
    
    # Generate HTML test report
    cat > test-reports/summary/index.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee Platform Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .summary { background: #ecf0f1; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .pass { color: #27ae60; }
        .fail { color: #e74c3c; }
        .skip { color: #f39c12; }
        .section { margin: 20px 0; padding: 15px; border-left: 4px solid #3498db; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Go Coffee Platform Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Environment: $ENVIRONMENT</p>
    </div>
    
    <div class="summary">
        <h2>Test Summary</h2>
        <p><strong>Total Tests:</strong> $TOTAL_TESTS</p>
        <p class="pass"><strong>Passed:</strong> $PASSED_TESTS</p>
        <p class="fail"><strong>Failed:</strong> $FAILED_TESTS</p>
        <p class="skip"><strong>Skipped:</strong> $SKIPPED_TESTS</p>
        <p><strong>Success Rate:</strong> $(( PASSED_TESTS * 100 / (TOTAL_TESTS == 0 ? 1 : TOTAL_TESTS) ))%</p>
    </div>
    
    <div class="section">
        <h3>Test Categories</h3>
        <ul>
            <li>Unit Tests - Code-level testing with coverage analysis</li>
            <li>Integration Tests - Service-to-service interaction testing</li>
            <li>End-to-End Tests - Complete workflow validation</li>
            <li>Load Tests - Performance and scalability testing</li>
            <li>Security Tests - Vulnerability and security analysis</li>
            <li>Performance Tests - Benchmark and profiling analysis</li>
        </ul>
    </div>
    
    <div class="section">
        <h3>Platform Components Tested</h3>
        <ul>
            <li>Core Kafka Services (Producer, Consumer, Streams)</li>
            <li>Web3 Payment Processing</li>
            <li>AI Agent Ecosystem (9 agents + orchestrator)</li>
            <li>Analytics & Business Intelligence</li>
            <li>Multi-Region Deployment</li>
            <li>Enterprise Security & Compliance</li>
        </ul>
    </div>
</body>
</html>
EOF
    
    # Generate JSON summary
    cat > test-reports/summary/results.json << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "environment": "$ENVIRONMENT",
    "summary": {
        "total": $TOTAL_TESTS,
        "passed": $PASSED_TESTS,
        "failed": $FAILED_TESTS,
        "skipped": $SKIPPED_TESTS,
        "success_rate": $(( PASSED_TESTS * 100 / (TOTAL_TESTS == 0 ? 1 : TOTAL_TESTS) ))
    },
    "categories": {
        "unit": "completed",
        "integration": "completed",
        "e2e": "completed",
        "load": "completed",
        "security": "completed",
        "performance": "completed"
    }
}
EOF
    
    print_success "Test reports generated in test-reports/summary/"
}

# Function to display test summary
display_test_summary() {
    print_header "Test Execution Summary"
    
    echo -e "${CYAN}Test Results:${NC}"
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo -e "${YELLOW}Skipped: $SKIPPED_TESTS${NC}"
    
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        local success_rate=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
        echo -e "Success Rate: ${success_rate}%"
        
        if [[ $success_rate -ge 90 ]]; then
            print_success "Excellent test results! 🎉"
        elif [[ $success_rate -ge 80 ]]; then
            print_success "Good test results! ✅"
        elif [[ $success_rate -ge 70 ]]; then
            print_warning "Acceptable test results, but room for improvement"
        else
            print_error "Poor test results, immediate attention required"
        fi
    fi
    
    if [[ "$GENERATE_REPORTS" == "true" ]]; then
        echo ""
        echo -e "${CYAN}Test Reports:${NC}"
        echo -e "HTML Report: test-reports/summary/index.html"
        echo -e "JSON Summary: test-reports/summary/results.json"
        echo -e "Coverage Report: test-reports/unit/coverage.html"
    fi
}

# Main test execution function
main() {
    print_header "Go Coffee Platform Comprehensive Test Suite"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --skip-slow)
                SKIP_SLOW_TESTS="true"
                shift
                ;;
            --no-reports)
                GENERATE_REPORTS="false"
                shift
                ;;
            --coverage-threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --load-duration)
                LOAD_TEST_DURATION="$2"
                shift 2
                ;;
            --load-users)
                LOAD_TEST_USERS="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --environment ENV          Test environment (default: test)"
                echo "  --skip-slow               Skip slow tests (integration, e2e, load)"
                echo "  --no-reports              Skip report generation"
                echo "  --coverage-threshold N    Coverage threshold percentage (default: 80)"
                echo "  --load-duration DURATION  Load test duration (default: 300s)"
                echo "  --load-users N            Load test concurrent users (default: 100)"
                echo "  --help                    Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Create test reports directory
    mkdir -p test-reports
    
    # Execute all test categories
    run_unit_tests
    run_integration_tests
    run_e2e_tests
    run_load_tests
    run_security_tests
    run_performance_tests
    
    # Generate reports and summary
    generate_test_reports
    display_test_summary
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function with all arguments
main "$@"
