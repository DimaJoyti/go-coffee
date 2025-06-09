#!/bin/bash

# Kitchen Service Test Runner
# This script runs comprehensive tests for the kitchen service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DB=15
REDIS_URL="redis://localhost:6379/$TEST_DB"
COVERAGE_THRESHOLD=80

echo -e "${BLUE}🧪 Kitchen Service Test Suite${NC}"
echo "=================================="

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check prerequisites
print_info "Checking prerequisites..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

# Check if Redis is running
if ! redis-cli ping &> /dev/null; then
    echo -e "${RED}❌ Redis is not running. Please start Redis server.${NC}"
    exit 1
fi

print_status 0 "Prerequisites check passed"

# Clean test database
print_info "Cleaning test database..."
redis-cli -n $TEST_DB FLUSHDB > /dev/null
print_status 0 "Test database cleaned"

# Set environment variables for testing
export REDIS_URL=$REDIS_URL
export ENVIRONMENT=test
export LOG_LEVEL=error

# Navigate to project root
cd "$(dirname "$0")/.."

# Run unit tests
print_info "Running unit tests..."
go test -v -race -timeout=30s ./internal/kitchen/domain/... 2>&1 | tee test-results-unit.log
UNIT_EXIT_CODE=${PIPESTATUS[0]}
print_status $UNIT_EXIT_CODE "Unit tests"

# Run application layer tests
print_info "Running application layer tests..."
go test -v -race -timeout=30s ./internal/kitchen/application/... 2>&1 | tee test-results-application.log
APP_EXIT_CODE=${PIPESTATUS[0]}
print_status $APP_EXIT_CODE "Application layer tests"

# Run infrastructure tests (if Redis is available)
print_info "Running infrastructure tests..."
go test -v -race -timeout=60s ./internal/kitchen/infrastructure/... 2>&1 | tee test-results-infrastructure.log
INFRA_EXIT_CODE=${PIPESTATUS[0]}
print_status $INFRA_EXIT_CODE "Infrastructure tests"

# Run integration tests
print_info "Running integration tests..."
go test -v -race -timeout=120s -tags=integration ./internal/kitchen/... 2>&1 | tee test-results-integration.log
INTEGRATION_EXIT_CODE=${PIPESTATUS[0]}
print_status $INTEGRATION_EXIT_CODE "Integration tests"

# Run tests with coverage
print_info "Running tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/kitchen/... 2>&1 | tee test-results-coverage.log
COVERAGE_EXIT_CODE=${PIPESTATUS[0]}

if [ $COVERAGE_EXIT_CODE -eq 0 ]; then
    # Generate coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    # Check coverage threshold
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    COVERAGE_INT=${COVERAGE%.*}
    
    if [ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ]; then
        print_status 0 "Coverage tests (${COVERAGE}% >= ${COVERAGE_THRESHOLD}%)"
    else
        echo -e "${YELLOW}⚠️  Coverage is ${COVERAGE}% (below threshold of ${COVERAGE_THRESHOLD}%)${NC}"
    fi
else
    print_status $COVERAGE_EXIT_CODE "Coverage tests"
fi

# Run benchmark tests
print_info "Running benchmark tests..."
go test -v -bench=. -benchmem ./internal/kitchen/... 2>&1 | tee test-results-benchmark.log
BENCH_EXIT_CODE=${PIPESTATUS[0]}
print_status $BENCH_EXIT_CODE "Benchmark tests"

# Run race condition tests
print_info "Running race condition tests..."
go test -v -race -count=10 ./internal/kitchen/domain/... 2>&1 | tee test-results-race.log
RACE_EXIT_CODE=${PIPESTATUS[0]}
print_status $RACE_EXIT_CODE "Race condition tests"

# Run static analysis
print_info "Running static analysis..."

# Check if golangci-lint is installed
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./internal/kitchen/... 2>&1 | tee test-results-lint.log
    LINT_EXIT_CODE=${PIPESTATUS[0]}
    print_status $LINT_EXIT_CODE "Static analysis (golangci-lint)"
else
    echo -e "${YELLOW}⚠️  golangci-lint not installed, skipping static analysis${NC}"
fi

# Check if go vet passes
go vet ./internal/kitchen/... 2>&1 | tee test-results-vet.log
VET_EXIT_CODE=${PIPESTATUS[0]}
print_status $VET_EXIT_CODE "Go vet analysis"

# Check for potential security issues
if command -v gosec &> /dev/null; then
    gosec ./internal/kitchen/... 2>&1 | tee test-results-security.log
    SECURITY_EXIT_CODE=${PIPESTATUS[0]}
    print_status $SECURITY_EXIT_CODE "Security analysis (gosec)"
else
    echo -e "${YELLOW}⚠️  gosec not installed, skipping security analysis${NC}"
fi

# Test API endpoints (if service is running)
print_info "Testing API endpoints..."

# Start the service in background for API testing
export HTTP_PORT=8888
export GRPC_PORT=9999
export HEALTH_PORT=8889

# Build the service
go build -o kitchen-service-test cmd/kitchen-service/main.go

# Start service in background
./kitchen-service-test &
SERVICE_PID=$!

# Wait for service to start
sleep 5

# Test health endpoint
if curl -f http://localhost:8889/health > /dev/null 2>&1; then
    print_status 0 "Health endpoint test"
else
    print_status 1 "Health endpoint test"
fi

# Test HTTP API endpoints
if curl -f http://localhost:8888/api/v1/kitchen/queue/status > /dev/null 2>&1; then
    print_status 0 "HTTP API test"
else
    print_status 1 "HTTP API test"
fi

# Test gRPC endpoint (if grpcurl is available)
if command -v grpcurl &> /dev/null; then
    if grpcurl -plaintext localhost:9999 grpc.health.v1.Health/Check > /dev/null 2>&1; then
        print_status 0 "gRPC API test"
    else
        print_status 1 "gRPC API test"
    fi
else
    echo -e "${YELLOW}⚠️  grpcurl not installed, skipping gRPC API test${NC}"
fi

# Stop the service
kill $SERVICE_PID 2>/dev/null || true
rm -f kitchen-service-test

# Load testing (if hey is available)
if command -v hey &> /dev/null; then
    print_info "Running load tests..."
    
    # Start service again for load testing
    ./kitchen-service-test &
    SERVICE_PID=$!
    sleep 3
    
    # Run load test
    hey -n 1000 -c 10 -t 30 http://localhost:8888/api/v1/kitchen/health > load-test-results.log 2>&1
    LOAD_EXIT_CODE=$?
    
    # Stop service
    kill $SERVICE_PID 2>/dev/null || true
    
    print_status $LOAD_EXIT_CODE "Load tests"
else
    echo -e "${YELLOW}⚠️  hey not installed, skipping load tests${NC}"
fi

# Clean up test database
redis-cli -n $TEST_DB FLUSHDB > /dev/null

# Generate test report
print_info "Generating test report..."

cat > test-report.md << EOF
# Kitchen Service Test Report

Generated: $(date)

## Test Results

| Test Type | Status | Details |
|-----------|--------|---------|
| Unit Tests | $([ $UNIT_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Domain layer tests |
| Application Tests | $([ $APP_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Application layer tests |
| Infrastructure Tests | $([ $INFRA_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Redis integration tests |
| Integration Tests | $([ $INTEGRATION_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | End-to-end workflow tests |
| Coverage Tests | $([ $COVERAGE_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Code coverage: ${COVERAGE:-N/A}% |
| Benchmark Tests | $([ $BENCH_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Performance benchmarks |
| Race Condition Tests | $([ $RACE_EXIT_CODE -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL") | Concurrency safety tests |

## Files Generated

- \`coverage.html\` - Coverage report
- \`test-results-*.log\` - Detailed test logs
- \`load-test-results.log\` - Load test results

## Coverage Report

$([ -f coverage.out ] && go tool cover -func=coverage.out | tail -1 || echo "Coverage report not available")

EOF

print_status 0 "Test report generated (test-report.md)"

# Summary
echo ""
echo -e "${BLUE}📊 Test Summary${NC}"
echo "================"

TOTAL_TESTS=7
PASSED_TESTS=0

[ $UNIT_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $APP_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $INFRA_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $INTEGRATION_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $COVERAGE_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $BENCH_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))
[ $RACE_EXIT_CODE -eq 0 ] && ((PASSED_TESTS++))

echo "Tests Passed: $PASSED_TESTS/$TOTAL_TESTS"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}🎉 All tests passed! Kitchen service is ready for production.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please check the logs and fix issues.${NC}"
    exit 1
fi
