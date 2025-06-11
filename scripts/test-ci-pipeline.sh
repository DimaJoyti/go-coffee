#!/bin/bash

# Test CI/CD Pipeline Locally
# This script simulates the CI/CD pipeline steps locally

set -e

echo "🚀 Testing Go Coffee CI/CD Pipeline Locally"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Step 1: Check Go version
echo "Step 1: Checking Go version..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    print_status "Go is installed: $GO_VERSION"
else
    print_error "Go is not installed"
    exit 1
fi

# Step 2: Download dependencies
echo "Step 2: Downloading dependencies..."
if go mod download; then
    print_status "Dependencies downloaded successfully"
else
    print_warning "Some dependencies failed to download"
fi

# Step 3: Run linting (if golangci-lint is available)
echo "Step 3: Running linting..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --timeout=5m; then
        print_status "Linting passed"
    else
        print_warning "Linting found issues"
    fi
else
    print_warning "golangci-lint not installed, skipping linting"
fi

# Step 4: Run security scan (if gosec is available)
echo "Step 4: Running security scan..."
if command -v gosec &> /dev/null; then
    if gosec ./...; then
        print_status "Security scan passed"
    else
        print_warning "Security scan found issues"
    fi
else
    print_warning "gosec not installed, skipping security scan"
fi

# Step 5: Run unit tests
echo "Step 5: Running unit tests..."
if go test -v -race -coverprofile=coverage.out -covermode=atomic ./...; then
    print_status "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Step 6: Test individual services
echo "Step 6: Testing individual services..."
for dir in accounts-service producer consumer streams; do
    if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
        echo "Testing $dir..."
        cd "$dir"
        if go mod tidy && go test -v ./...; then
            print_status "$dir tests passed"
        else
            print_warning "$dir tests had issues"
        fi
        cd ..
    fi
done

# Step 7: Test Docker builds (if Docker is available)
echo "Step 7: Testing Docker builds..."
if command -v docker &> /dev/null; then
    # Test building a few key services
    for service in producer consumer streams; do
        if [ -f "$service/Dockerfile" ]; then
            echo "Building Docker image for $service..."
            if docker build -t go-coffee-$service:test $service/; then
                print_status "$service Docker build successful"
            else
                print_warning "$service Docker build failed"
            fi
        fi
    done
else
    print_warning "Docker not available, skipping Docker builds"
fi

echo ""
echo "🎉 CI/CD Pipeline Test Complete!"
echo "================================="
print_status "Local CI/CD pipeline test finished"
echo "You can now push your changes to trigger the GitHub Actions workflow."
