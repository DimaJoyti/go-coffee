.PHONY: help build test clean dev start-all stop-all payment auth order kitchen gateway bitcoin-test

# Default target
help:
	@echo "☕ Go Coffee - Available Commands:"
	@echo "=================================="
	@echo "  build          - Build all services"
	@echo "  test           - Run all tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Start development environment"
	@echo ""
	@echo "🚀 Service Management:"
	@echo "  start-all      - Start all microservices"
	@echo "  stop-all       - Stop all running services"
	@echo "  status         - Check service status"
	@echo ""
	@echo "🔧 Individual Services:"
	@echo "  payment        - Run payment service"
	@echo "  auth           - Run auth service"
	@echo "  order          - Run order service"
	@echo "  kitchen        - Run kitchen service"
	@echo "  gateway        - Run API gateway"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  bitcoin-test   - Test Bitcoin implementation"
	@echo "  test-payment   - Test payment service"
	@echo "  test-all       - Run comprehensive tests"

# Build all services
build:
	@echo "🔨 Building all services..."
	@mkdir -p bin
	go build -o bin/payment-service ./cmd/payment-service
	go build -o bin/auth-service ./cmd/auth-service
	go build -o bin/order-service ./cmd/order-service
	go build -o bin/kitchen-service ./cmd/kitchen-service
	go build -o bin/api-gateway ./cmd/api-gateway
	@echo "✅ Build complete"

# Run all tests
test:
	@echo "🧪 Running tests..."
	go test ./pkg/... -v
	go test ./internal/... -v
	@echo "✅ Tests complete"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/
	go clean
	@echo "✅ Clean complete"

# Start development environment
dev:
	@echo "🐳 Starting development environment..."
	docker-compose -f deployments/docker/docker-compose.dev.yml up -d
	@echo "✅ Development environment started"

# Start all services
start-all:
	@echo "🚀 Starting all microservices..."
	chmod +x scripts/start-all-services.sh
	./scripts/start-all-services.sh

# Stop all services
stop-all:
	@echo "🛑 Stopping all services..."
	pkill -f "payment-service|auth-service|order-service|kitchen-service|api-gateway" || true
	@echo "✅ All services stopped"

# Check service status
status:
	@echo "📊 Service Status:"
	@echo "=================="
	@curl -s http://localhost:8080/api/v1/gateway/services 2>/dev/null | head -c 300 || echo "API Gateway not responding"
	@echo ""

# Individual service commands
payment:
	@echo "💰 Starting payment service..."
	cd cmd/payment-service && go run .

auth:
	@echo "🔐 Starting auth service..."
	cd cmd/auth-service && go run .

order:
	@echo "📋 Starting order service..."
	cd cmd/order-service && go run .

kitchen:
	@echo "👨‍🍳 Starting kitchen service..."
	cd cmd/kitchen-service && go run .

gateway:
	@echo "🌐 Starting API gateway..."
	cd cmd/api-gateway && go run .

# Testing commands
bitcoin-test:
	@echo "₿ Testing Bitcoin implementation..."
	cd pkg/bitcoin && go test -v

test-payment:
	@echo "💰 Testing payment service..."
	chmod +x scripts/test-payment-service.sh
	./scripts/test-payment-service.sh

test-all:
	@echo "🧪 Running comprehensive tests..."
	make test
	make bitcoin-test
	@echo "✅ All tests complete"
