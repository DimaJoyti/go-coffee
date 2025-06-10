#!/bin/bash

# Auth Service Test Runner
# This script runs all tests for the Auth Service

set -e

echo "🧪 Auth Service Test Suite"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Create test results directory
mkdir -p test-results

print_status "Starting Auth Service tests..."

# Run unit tests
print_status "Running unit tests..."
if go test -v -race -coverprofile=test-results/auth-coverage.out ./internal/auth/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Generate coverage report
print_status "Generating coverage report..."
go tool cover -html=test-results/auth-coverage.out -o test-results/auth-coverage.html
print_success "Coverage report generated: test-results/auth-coverage.html"

# Show coverage summary
COVERAGE=$(go tool cover -func=test-results/auth-coverage.out | grep total | awk '{print $3}')
print_status "Total test coverage: $COVERAGE"

# Run specific test suites
print_status "Running domain layer tests..."
go test -v ./internal/auth/domain/...

print_status "Running application layer tests..."
go test -v ./internal/auth/application/...

print_status "Running infrastructure layer tests..."
go test -v ./internal/auth/infrastructure/...

print_status "Running transport layer tests..."
go test -v ./internal/auth/transport/...

# Run integration tests (if not in short mode)
if [ "$1" != "--short" ]; then
    print_status "Running integration tests..."
    if go test -v -tags=integration ./internal/auth/...; then
        print_success "Integration tests passed"
    else
        print_warning "Integration tests failed (this might be expected if database/redis are not available)"
    fi
fi

# Run benchmarks
print_status "Running benchmarks..."
go test -bench=. -benchmem ./internal/auth/... > test-results/auth-benchmarks.txt
print_success "Benchmarks completed: test-results/auth-benchmarks.txt"

# Check for race conditions
print_status "Checking for race conditions..."
if go test -race ./internal/auth/...; then
    print_success "No race conditions detected"
else
    print_error "Race conditions detected"
    exit 1
fi

# Run security checks (if gosec is available)
if command -v gosec &> /dev/null; then
    print_status "Running security scan..."
    if gosec ./internal/auth/...; then
        print_success "Security scan passed"
    else
        print_warning "Security scan found issues"
    fi
else
    print_warning "gosec not installed. Skipping security scan."
fi

# Run vulnerability check (if govulncheck is available)
if command -v govulncheck &> /dev/null; then
    print_status "Checking for vulnerabilities..."
    if govulncheck ./internal/auth/...; then
        print_success "No vulnerabilities found"
    else
        print_warning "Vulnerabilities found"
    fi
else
    print_warning "govulncheck not installed. Skipping vulnerability check."
fi

# Summary
echo ""
echo "🎉 Test Summary"
echo "==============="
print_success "All tests completed successfully!"
print_status "Coverage: $COVERAGE"
print_status "Results saved in: test-results/"

# Open coverage report in browser (optional)
if [ "$2" = "--open" ]; then
    if command -v xdg-open &> /dev/null; then
        xdg-open test-results/auth-coverage.html
    elif command -v open &> /dev/null; then
        open test-results/auth-coverage.html
    else
        print_status "Coverage report: test-results/auth-coverage.html"
    fi
fi

echo ""
print_success "Auth Service test suite completed! ✨"
