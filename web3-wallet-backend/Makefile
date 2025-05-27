# Go Coffee Web3 Wallet Backend Makefile

# Variables
BINARY_NAME=web3-wallet-backend
DOCKER_IMAGE=go-coffee/web3-wallet-backend
VERSION?=latest
GO_VERSION=1.21

# Build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# Directories
BUILD_DIR=build
COVERAGE_DIR=coverage
DOCS_DIR=docs

.PHONY: help build test clean docker run deps lint format check coverage integration-test unit-test solana-test

# Default target
help: ## Show this help message
	@echo "Go Coffee Web3 Wallet Backend"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Dependencies
deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	go mod verify

# Build
build: deps ## Build the application
	@echo "🔨 Building application..."
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wallet-service
	go build $(LDFLAGS) -o $(BUILD_DIR)/defi-service ./cmd/defi-service
	go build $(LDFLAGS) -o $(BUILD_DIR)/trading-service ./cmd/trading-service

# Clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	go clean -cache
	go clean -testcache

# Testing
test: unit-test ## Run all tests

unit-test: ## Run unit tests
	@echo "🧪 Running unit tests..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -coverprofile=$(COVERAGE_DIR)/unit.out ./internal/...
	go test -v -race -coverprofile=$(COVERAGE_DIR)/pkg.out ./pkg/...

integration-test: ## Run integration tests (requires network)
	@echo "🔗 Running integration tests..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -tags=integration -coverprofile=$(COVERAGE_DIR)/integration.out ./tests/integration/...

solana-test: ## Run Solana-specific tests
	@echo "☀️ Running Solana tests..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -run=".*Solana.*" -coverprofile=$(COVERAGE_DIR)/solana.out ./...

# Coverage
coverage: unit-test ## Generate coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=$(COVERAGE_DIR)/unit.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1

coverage-integration: integration-test ## Generate integration coverage report
	@echo "📊 Generating integration coverage report..."
	go tool cover -html=$(COVERAGE_DIR)/integration.out -o $(COVERAGE_DIR)/integration-coverage.html
	go tool cover -func=$(COVERAGE_DIR)/integration.out | tail -1

# Code quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run ./...

format: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...
	goimports -w .

check: format lint test ## Run all checks (format, lint, test)

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-run: ## Run Docker container
	@echo "🚀 Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 $(DOCKER_IMAGE):latest

docker-push: docker-build ## Push Docker image
	@echo "📤 Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

# Development
run-wallet: build ## Run wallet service
	@echo "🚀 Starting wallet service..."
	./$(BUILD_DIR)/$(BINARY_NAME)

run-defi: build ## Run DeFi service
	@echo "🚀 Starting DeFi service..."
	./$(BUILD_DIR)/defi-service

run-trading: build ## Run trading service
	@echo "🚀 Starting trading service..."
	./$(BUILD_DIR)/trading-service

# Database
db-migrate: ## Run database migrations
	@echo "🗄️ Running database migrations..."
	migrate -path ./migrations -database "postgres://localhost/go_coffee_dev?sslmode=disable" up

db-rollback: ## Rollback database migrations
	@echo "🗄️ Rolling back database migrations..."
	migrate -path ./migrations -database "postgres://localhost/go_coffee_dev?sslmode=disable" down 1

# Tools
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	mkdir -p $(DOCS_DIR)
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Benchmarks
bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

# Security
security: ## Run security checks
	@echo "🔒 Running security checks..."
	gosec ./...

# Performance
profile: ## Run performance profiling
	@echo "📈 Running performance profiling..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...

# Environment setup
setup-dev: install-tools deps ## Setup development environment
	@echo "🛠️ Setting up development environment..."
	@echo "✅ Development environment ready!"

# CI/CD
ci: check coverage ## Run CI pipeline
	@echo "🔄 Running CI pipeline..."
	@echo "✅ CI pipeline completed!"

# Monitoring
health-check: ## Check service health
	@echo "❤️ Checking service health..."
	curl -f http://localhost:8080/health || exit 1
	curl -f http://localhost:9090/health || exit 1

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@echo "☸️ Deploying to Kubernetes..."
	kubectl apply -f k8s/

k8s-delete: ## Delete from Kubernetes
	@echo "☸️ Deleting from Kubernetes..."
	kubectl delete -f k8s/

# Load testing
load-test: ## Run load tests
	@echo "🚛 Running load tests..."
	k6 run tests/load/wallet-service.js
	k6 run tests/load/defi-service.js

# Quick commands
quick-test: ## Quick test (no coverage)
	@echo "⚡ Running quick tests..."
	go test -short ./...

quick-build: ## Quick build (no deps check)
	@echo "⚡ Quick build..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wallet-service

# Version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Build Time: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
