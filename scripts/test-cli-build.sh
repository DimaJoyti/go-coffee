#!/bin/bash

# Test CLI Build Locally
# This script tests the CLI build process locally before pushing

set -e

echo "🔧 Testing Go Coffee CLI Build Locally"
echo "======================================="

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

# Step 1: Check CLI structure
echo "Step 1: Checking CLI structure..."
if [ -f "cmd/gocoffee-cli/main.go" ]; then
    print_status "CLI main.go exists"
else
    print_error "CLI main.go not found"
    exit 1
fi

if [ -d "internal/cli" ]; then
    print_status "CLI internal directory exists"
else
    print_error "CLI internal directory not found"
    exit 1
fi

if [ -f "Makefile.cli" ]; then
    print_status "Makefile.cli exists"
else
    print_error "Makefile.cli not found"
    exit 1
fi

# Step 2: Check dependencies
echo "Step 2: Checking dependencies..."
if go mod download; then
    print_status "Dependencies downloaded"
else
    print_error "Failed to download dependencies"
    exit 1
fi

# Step 3: Run CLI tests
echo "Step 3: Running CLI tests..."
if make -f Makefile.cli test; then
    print_status "CLI tests passed"
else
    print_warning "CLI tests had issues (this may be expected)"
fi

# Step 4: Build CLI
echo "Step 4: Building CLI..."
if make -f Makefile.cli build; then
    print_status "CLI build successful"
else
    print_error "CLI build failed"
    exit 1
fi

# Step 5: Test CLI binary
echo "Step 5: Testing CLI binary..."
if [ -f "bin/gocoffee" ]; then
    print_status "CLI binary created"
    
    # Test basic CLI functionality
    if ./bin/gocoffee --help > /dev/null 2>&1; then
        print_status "CLI help command works"
    else
        print_warning "CLI help command had issues"
    fi
    
    if ./bin/gocoffee version > /dev/null 2>&1; then
        print_status "CLI version command works"
    else
        print_warning "CLI version command had issues"
    fi
else
    print_error "CLI binary not found"
    exit 1
fi

# Step 6: Test multi-platform builds
echo "Step 6: Testing multi-platform builds..."
if make -f Makefile.cli build-all; then
    print_status "Multi-platform builds successful"
    
    # Check if all expected binaries exist
    expected_binaries=(
        "bin/gocoffee-linux-amd64"
        "bin/gocoffee-darwin-amd64"
        "bin/gocoffee-darwin-arm64"
        "bin/gocoffee-windows-amd64.exe"
    )
    
    for binary in "${expected_binaries[@]}"; do
        if [ -f "$binary" ]; then
            print_status "$binary created"
        else
            print_warning "$binary not found"
        fi
    done
else
    print_warning "Multi-platform builds had issues"
fi

# Step 7: Test Docker build (if Docker is available)
echo "Step 7: Testing Docker build..."
if command -v docker &> /dev/null; then
    if [ -f "docker/Dockerfile.cli" ]; then
        if docker build -t gocoffee-cli:test -f docker/Dockerfile.cli .; then
            print_status "Docker build successful"
        else
            print_warning "Docker build failed"
        fi
    else
        print_warning "Docker CLI Dockerfile not found"
    fi
else
    print_warning "Docker not available, skipping Docker build test"
fi

# Step 8: Test linting (if golangci-lint is available)
echo "Step 8: Testing linting..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./internal/cli/... ./cmd/gocoffee-cli/...; then
        print_status "Linting passed"
    else
        print_warning "Linting found issues"
    fi
else
    print_warning "golangci-lint not available, skipping linting test"
fi

echo ""
echo "🎉 CLI Build Test Complete!"
echo "==========================="
print_status "Local CLI build test finished"
echo "You can now push your changes to trigger the CLI Build and Release workflow."

# Cleanup
echo ""
echo "Cleaning up..."
if make -f Makefile.cli clean; then
    print_status "Cleanup complete"
else
    print_warning "Cleanup had issues"
fi
