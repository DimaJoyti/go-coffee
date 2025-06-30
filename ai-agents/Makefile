# Makefile for Go Coffee AI Agents

# Variables
GO_MODULE := go-coffee-ai-agents
PROTO_DIR := proto
OUTPUT_DIR := proto/generated
SCRIPTS_DIR := scripts

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Binary names
BEVERAGE_INVENTOR_BINARY := bin/beverage-inventor-agent
INVENTORY_MANAGER_BINARY := bin/inventory-manager-agent
TASK_MANAGER_BINARY := bin/task-manager-agent
SOCIAL_MEDIA_BINARY := bin/social-media-agent

# Default target
.PHONY: all
all: clean deps proto build test

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Clean, install deps, generate proto, build, and test"
	@echo "  deps             - Install Go dependencies"
	@echo "  proto            - Generate Go code from Protocol Buffer definitions"
	@echo "  build            - Build all agent binaries"
	@echo "  build-beverage   - Build beverage inventor agent"
	@echo "  build-inventory  - Build inventory manager agent"
	@echo "  build-task       - Build task manager agent"
	@echo "  build-social     - Build social media agent"
	@echo "  test             - Run all tests"
	@echo "  test-verbose     - Run tests with verbose output"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format Go code"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-proto      - Clean generated Protocol Buffer files"
	@echo "  install-tools    - Install required development tools"
	@echo "  docker-build     - Build Docker images for all agents"
	@echo "  run-beverage     - Run beverage inventor agent"
	@echo "  run-inventory    - Run inventory manager agent"

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Generate Protocol Buffer code
.PHONY: proto
proto:
	@echo "🔨 Generating Protocol Buffer code..."
	@chmod +x $(SCRIPTS_DIR)/generate-proto.sh
	@./$(SCRIPTS_DIR)/generate-proto.sh

# Clean generated proto files
.PHONY: clean-proto
clean-proto:
	@echo "🧹 Cleaning generated Protocol Buffer files..."
	@rm -rf $(OUTPUT_DIR)

# Build all agents
.PHONY: build
build: build-beverage build-inventory build-task build-social

# Build beverage inventor agent
.PHONY: build-beverage
build-beverage:
	@echo "🔨 Building beverage inventor agent..."
	@mkdir -p bin
	$(GOBUILD) -o $(BEVERAGE_INVENTOR_BINARY) ./beverage-inventor-agent/cmd

# Build inventory manager agent
.PHONY: build-inventory
build-inventory:
	@echo "🔨 Building inventory manager agent..."
	@mkdir -p bin
	$(GOBUILD) -o $(INVENTORY_MANAGER_BINARY) ./inventory-manager-agent/cmd

# Build task manager agent
.PHONY: build-task
build-task:
	@echo "🔨 Building task manager agent..."
	@mkdir -p bin
	$(GOBUILD) -o $(TASK_MANAGER_BINARY) ./task-manager-agent/cmd

# Build social media agent
.PHONY: build-social
build-social:
	@echo "🔨 Building social media agent..."
	@mkdir -p bin
	$(GOBUILD) -o $(SOCIAL_MEDIA_BINARY) ./social-media-agent/cmd

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	$(GOTEST) ./...

# Run tests with verbose output
.PHONY: test-verbose
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -cover ./...
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# Lint code
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting Go code..."
	$(GOFMT) ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "🛠️  Installing development tools..."
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker images..."
	@docker build -t go-coffee/beverage-inventor-agent -f beverage-inventor-agent/Dockerfile .
	@docker build -t go-coffee/inventory-manager-agent -f inventory-manager-agent/Dockerfile .
	@docker build -t go-coffee/task-manager-agent -f task-manager-agent/Dockerfile .
	@docker build -t go-coffee/social-media-agent -f social-media-agent/Dockerfile .

# Run agents locally
.PHONY: run-beverage
run-beverage: build-beverage
	@echo "🚀 Running beverage inventor agent..."
	@cd beverage-inventor-agent && ../$(BEVERAGE_INVENTOR_BINARY)

.PHONY: run-inventory
run-inventory: build-inventory
	@echo "🚀 Running inventory manager agent..."
	@cd inventory-manager-agent && ../$(INVENTORY_MANAGER_BINARY)

.PHONY: run-task
run-task: build-task
	@echo "🚀 Running task manager agent..."
	@cd task-manager-agent && ../$(TASK_MANAGER_BINARY)

.PHONY: run-social
run-social: build-social
	@echo "🚀 Running social media agent..."
	@cd social-media-agent && ../$(SOCIAL_MEDIA_BINARY)

# Development workflow targets
.PHONY: dev-setup
dev-setup: install-tools deps proto
	@echo "🎉 Development environment setup complete!"

.PHONY: dev-build
dev-build: fmt lint test build
	@echo "🎉 Development build complete!"

.PHONY: ci
ci: deps proto fmt lint test-coverage build
	@echo "🎉 CI pipeline complete!"

# Watch and rebuild (requires entr)
.PHONY: watch-beverage
watch-beverage:
	@echo "👀 Watching beverage inventor agent for changes..."
	@find beverage-inventor-agent -name "*.go" | entr -r make run-beverage

.PHONY: watch-test
watch-test:
	@echo "👀 Watching for changes and running tests..."
	@find . -name "*.go" | entr -r make test

# Database migrations (if needed)
.PHONY: migrate-up
migrate-up:
	@echo "⬆️  Running database migrations..."
	@# Add migration commands here

.PHONY: migrate-down
migrate-down:
	@echo "⬇️  Rolling back database migrations..."
	@# Add rollback commands here

# Kubernetes deployment
.PHONY: k8s-deploy
k8s-deploy:
	@echo "☸️  Deploying to Kubernetes..."
	@kubectl apply -f k8s/

.PHONY: k8s-delete
k8s-delete:
	@echo "☸️  Deleting from Kubernetes..."
	@kubectl delete -f k8s/

# Performance benchmarks
.PHONY: bench
bench:
	@echo "🏃 Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Security scan
.PHONY: security
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
.PHONY: docs
docs:
	@echo "📚 Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "📖 Documentation server: http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi
