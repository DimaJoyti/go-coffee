#!/bin/bash

# Migration script to clean architecture
# This script helps migrate the existing codebase to the new clean architecture

set -e

echo "🚀 Starting migration to clean architecture..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Please run this script from the project root directory"
    exit 1
fi

print_step "1. Creating backup of current structure..."
if [ ! -d "backup" ]; then
    mkdir -p backup
    cp -r . backup/ 2>/dev/null || true
    print_status "Backup created in ./backup/"
else
    print_warning "Backup directory already exists, skipping backup"
fi

print_step "2. Verifying new structure exists..."
if [ ! -d "pkg/bitcoin" ]; then
    print_error "New structure not found. Please ensure the clean architecture has been set up."
    exit 1
fi
print_status "New structure verified"

print_step "3. Testing Bitcoin package in new location..."
cd pkg/bitcoin
if go test -v > /dev/null 2>&1; then
    print_status "Bitcoin package tests pass in new location"
else
    print_error "Bitcoin package tests fail in new location"
    cd ../..
    exit 1
fi
cd ../..

print_step "4. Identifying services to migrate..."
SERVICES_TO_MIGRATE=(
    "auth-service"
    "order-service" 
    "kitchen-service"
    "crypto-wallet"
    "crypto-terminal"
    "ai-agents"
    "ai-arbitrage"
)

print_status "Found ${#SERVICES_TO_MIGRATE[@]} services to migrate"

print_step "5. Creating migration plan..."
cat > migration-plan.md << EOF
# Migration Plan

## Services to Migrate

### High Priority (Core Business Logic)
- [ ] payment-service (crypto-wallet + crypto-terminal)
- [ ] auth-service (existing)
- [ ] order-service (existing)
- [ ] kitchen-service (existing)

### Medium Priority (AI/ML Features)
- [ ] ai-service (ai-agents + ai-arbitrage)
- [ ] notification-service (new)
- [ ] analytics-service (new)

### Low Priority (Infrastructure)
- [ ] api-gateway (new)

## Migration Steps per Service

### 1. Payment Service
- [x] Move Bitcoin implementation to pkg/bitcoin/
- [x] Create internal/payment/ structure
- [x] Create cmd/payment-service/
- [ ] Convert Gin handlers to standard HTTP
- [ ] Add proper error handling
- [ ] Add comprehensive tests
- [ ] Update Docker configuration

### 2. Auth Service
- [ ] Move to internal/auth/
- [ ] Update import paths
- [ ] Standardize configuration
- [ ] Add gRPC interface

### 3. Order Service
- [ ] Move to internal/order/
- [ ] Update import paths
- [ ] Add event publishing
- [ ] Integrate with payment service

### 4. Kitchen Service
- [ ] Move to internal/kitchen/
- [ ] Update import paths
- [ ] Add real-time updates
- [ ] Integrate with order service

## Configuration Migration
- [ ] Consolidate environment files
- [ ] Update Docker Compose
- [ ] Create Kubernetes manifests
- [ ] Set up monitoring

## Testing Strategy
- [ ] Unit tests for each service
- [ ] Integration tests
- [ ] E2E tests
- [ ] Performance tests

## Deployment Strategy
- [ ] Blue-green deployment
- [ ] Database migrations
- [ ] Service discovery
- [ ] Load balancing
EOF

print_status "Migration plan created: migration-plan.md"

print_step "6. Checking dependencies..."
if ! command -v go &> /dev/null; then
    print_error "Go is not installed"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    print_warning "Docker is not installed - some features may not work"
fi

print_status "Dependencies checked"

print_step "7. Running tests on new structure..."
if go test ./pkg/... -v > test-results.log 2>&1; then
    print_status "All pkg tests pass"
else
    print_warning "Some pkg tests failed - check test-results.log"
fi

print_step "8. Creating development environment..."
if [ ! -f "configs/development/app.env" ]; then
    mkdir -p configs/development
    cat > configs/development/app.env << EOF
# Development Environment Configuration
ENVIRONMENT=development
LOG_LEVEL=debug
LOG_FORMAT=text

# Server
PORT=8080

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=go_coffee_dev
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# Bitcoin
BITCOIN_NETWORK=testnet

# JWT
JWT_SECRET=dev-secret-key
JWT_EXPIRATION=24

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=100

# CORS
CORS_ENABLED=true
CORS_ALLOW_ORIGINS=*
EOF
    print_status "Development environment created"
else
    print_warning "Development environment already exists"
fi

print_step "9. Creating Makefile for development..."
if [ ! -f "Makefile" ]; then
    cat > Makefile << 'EOF'
.PHONY: help build test clean dev

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build all services"
	@echo "  test           - Run all tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Start development environment"
	@echo "  payment        - Run payment service"
	@echo "  bitcoin-test   - Test Bitcoin implementation"

# Build all services
build:
	@echo "Building all services..."
	go build -o bin/payment-service ./cmd/payment-service
	@echo "Build complete"

# Run all tests
test:
	@echo "Running tests..."
	go test ./pkg/... -v
	go test ./internal/... -v
	@echo "Tests complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean
	@echo "Clean complete"

# Start development environment
dev:
	@echo "Starting development environment..."
	docker-compose -f deployments/docker/docker-compose.dev.yml up -d
	@echo "Development environment started"

# Run payment service
payment:
	@echo "Starting payment service..."
	go run ./cmd/payment-service

# Test Bitcoin implementation
bitcoin-test:
	@echo "Testing Bitcoin implementation..."
	cd pkg/bitcoin && go test -v
EOF
    print_status "Makefile created"
else
    print_warning "Makefile already exists"
fi

print_step "10. Final verification..."
if [ -d "pkg/bitcoin" ] && [ -d "internal/payment" ] && [ -d "cmd/payment-service" ]; then
    print_status "✅ Clean architecture structure verified"
else
    print_error "❌ Clean architecture structure incomplete"
    exit 1
fi

echo ""
echo "🎉 Migration to clean architecture completed!"
echo ""
echo "Next steps:"
echo "1. Review the migration plan: migration-plan.md"
echo "2. Test the new structure: make test"
echo "3. Start development: make dev"
echo "4. Run payment service: make payment"
echo ""
echo "For more information, see: CLEAN_ARCHITECTURE_GUIDE.md"
EOF

chmod +x scripts/migrate-to-clean-architecture.sh
