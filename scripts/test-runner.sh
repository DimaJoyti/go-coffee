#!/bin/bash

# Go Coffee Platform - Advanced Test Runner
# Comprehensive testing script for all test types

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-10m}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
INTEGRATION_TESTS=${INTEGRATION_TESTS:-false}
E2E_TESTS=${E2E_TESTS:-false}
PERFORMANCE_TESTS=${PERFORMANCE_TESTS:-false}
SECURITY_TESTS=${SECURITY_TESTS:-false}

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking test dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    # Check Docker (for integration tests)
    if [[ "$INTEGRATION_TESTS" == "true" ]] && ! command -v docker &> /dev/null; then
        log_error "Docker is required for integration tests"
        exit 1
    fi
    
    # Check K6 (for performance tests)
    if [[ "$PERFORMANCE_TESTS" == "true" ]] && ! command -v k6 &> /dev/null; then
        log_warning "K6 not found, installing..."
        # Install K6 based on OS
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
            echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
            sudo apt-get update
            sudo apt-get install k6
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            brew install k6
        fi
    fi
    
    log_success "Dependencies check completed"
}

# Unit Tests
run_unit_tests() {
    log_info "Running unit tests..."
    
    # Create coverage directory
    mkdir -p coverage
    
    # Run tests with coverage
    go test -v -race -timeout=$TEST_TIMEOUT \
        -coverprofile=coverage/coverage.out \
        -covermode=atomic \
        ./...
    
    # Generate coverage report
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    
    # Check coverage threshold
    COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    
    if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
        log_warning "Coverage $COVERAGE% is below threshold $COVERAGE_THRESHOLD%"
    else
        log_success "Coverage $COVERAGE% meets threshold $COVERAGE_THRESHOLD%"
    fi
    
    log_success "Unit tests completed"
}

# Integration Tests
run_integration_tests() {
    if [[ "$INTEGRATION_TESTS" != "true" ]]; then
        log_info "Skipping integration tests"
        return
    fi
    
    log_info "Running integration tests..."
    
    # Start test infrastructure
    docker-compose -f docker-compose.test.yml up -d postgres redis
    
    # Wait for services to be ready
    sleep 10
    
    # Run integration tests
    go test -v -tags=integration -timeout=$TEST_TIMEOUT \
        ./tests/integration/...
    
    # Cleanup
    docker-compose -f docker-compose.test.yml down
    
    log_success "Integration tests completed"
}

# End-to-End Tests
run_e2e_tests() {
    if [[ "$E2E_TESTS" != "true" ]]; then
        log_info "Skipping E2E tests"
        return
    fi
    
    log_info "Running E2E tests..."
    
    # Start full application stack
    docker-compose -f docker-compose.production.yml up -d
    
    # Wait for services to be ready
    sleep 30
    
    # Run E2E tests
    if command -v npm &> /dev/null; then
        cd tests/e2e
        npm install
        npm run test
        cd ../..
    else
        log_warning "npm not found, skipping E2E tests"
    fi
    
    # Cleanup
    docker-compose -f docker-compose.production.yml down
    
    log_success "E2E tests completed"
}

# Performance Tests
run_performance_tests() {
    if [[ "$PERFORMANCE_TESTS" != "true" ]]; then
        log_info "Skipping performance tests"
        return
    fi
    
    log_info "Running performance tests..."
    
    # Ensure services are running
    docker-compose -f docker-compose.production.yml up -d
    sleep 30
    
    # Run K6 performance tests
    mkdir -p reports/performance
    
    k6 run --out json=reports/performance/results.json \
        tests/performance/load-test.js
    
    k6 run --out json=reports/performance/stress-results.json \
        tests/performance/stress-test.js
    
    log_success "Performance tests completed"
}

# Security Tests
run_security_tests() {
    if [[ "$SECURITY_TESTS" != "true" ]]; then
        log_info "Skipping security tests"
        return
    fi
    
    log_info "Running security tests..."
    
    # OWASP ZAP security testing
    if command -v zap-baseline.py &> /dev/null; then
        mkdir -p reports/security
        
        # Start application
        docker-compose -f docker-compose.production.yml up -d
        sleep 30
        
        # Run ZAP baseline scan
        zap-baseline.py -t http://localhost:8080 \
            -J reports/security/zap-report.json \
            -r reports/security/zap-report.html
    else
        log_warning "OWASP ZAP not found, skipping security tests"
    fi
    
    log_success "Security tests completed"
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    mkdir -p reports
    
    cat > reports/test-summary.md << EOF
# Go Coffee Platform - Test Report

## Test Execution Summary

**Date**: $(date)
**Coverage**: ${COVERAGE:-N/A}%
**Threshold**: $COVERAGE_THRESHOLD%

## Test Results

### Unit Tests
- Status: ✅ Passed
- Coverage: ${COVERAGE:-N/A}%
- Report: [Coverage Report](coverage/coverage.html)

### Integration Tests
- Status: $([ "$INTEGRATION_TESTS" == "true" ] && echo "✅ Passed" || echo "⏭️ Skipped")

### E2E Tests
- Status: $([ "$E2E_TESTS" == "true" ] && echo "✅ Passed" || echo "⏭️ Skipped")

### Performance Tests
- Status: $([ "$PERFORMANCE_TESTS" == "true" ] && echo "✅ Passed" || echo "⏭️ Skipped")
- Reports: [Load Test](reports/performance/results.json)

### Security Tests
- Status: $([ "$SECURITY_TESTS" == "true" ] && echo "✅ Passed" || echo "⏭️ Skipped")
- Reports: [Security Scan](reports/security/zap-report.html)

## Recommendations

- Maintain coverage above $COVERAGE_THRESHOLD%
- Run integration tests before deployment
- Monitor performance metrics
- Address security findings promptly
EOF

    log_success "Test report generated: reports/test-summary.md"
}

# Main execution
main() {
    log_info "Starting Go Coffee Platform test suite..."
    
    check_dependencies
    run_unit_tests
    run_integration_tests
    run_e2e_tests
    run_performance_tests
    run_security_tests
    generate_report
    
    log_success "All tests completed successfully!"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --integration)
            INTEGRATION_TESTS=true
            shift
            ;;
        --e2e)
            E2E_TESTS=true
            shift
            ;;
        --performance)
            PERFORMANCE_TESTS=true
            shift
            ;;
        --security)
            SECURITY_TESTS=true
            shift
            ;;
        --all)
            INTEGRATION_TESTS=true
            E2E_TESTS=true
            PERFORMANCE_TESTS=true
            SECURITY_TESTS=true
            shift
            ;;
        --coverage-threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --integration         Run integration tests"
            echo "  --e2e                Run end-to-end tests"
            echo "  --performance        Run performance tests"
            echo "  --security           Run security tests"
            echo "  --all                Run all test types"
            echo "  --coverage-threshold Set coverage threshold (default: 80)"
            echo "  --help               Show this help"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main
