# Go Coffee LangGraph Integration Makefile

# Variables
BINARY_NAME=langgraph-server
MAIN_PATH=cmd/main.go
BUILD_DIR=build
DOCKER_IMAGE=go-coffee/langgraph-orchestrator
VERSION?=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

.PHONY: all build clean test coverage lint run docker help

# Default target
all: clean lint test build

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Multi-platform build complete"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./pkg/...
	@echo "✅ Tests complete"

# Run tests with coverage
coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./pkg/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	$(GOTEST) -tags=integration -v ./...
	@echo "✅ Integration tests complete"

# Run benchmarks
benchmark:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./pkg/...
	@echo "✅ Benchmarks complete"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping lint"; \
	fi
	@echo "✅ Lint complete"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GOFMT) ./...
	@echo "✅ Format complete"

# Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	$(GOMOD) tidy
	@echo "✅ Dependencies tidied"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	$(GOMOD) download
	@echo "✅ Dependencies downloaded"

# Run the application
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode
dev:
	@echo "🔧 Running in development mode..."
	$(GOCMD) run $(MAIN_PATH)

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)"

# Push Docker image
docker-push: docker
	@echo "📤 Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
	@echo "✅ Docker image pushed"

# Run with Docker
docker-run:
	@echo "🐳 Running with Docker..."
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE):latest

# Install development tools
install-tools:
	@echo "🛠️  Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "📖 Documentation server: http://localhost:6060/pkg/github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc not installed, install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Security scan
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
	@echo "✅ Security scan complete"

# Performance profiling
profile:
	@echo "📈 Starting performance profiling..."
	$(GOCMD) run $(MAIN_PATH) -cpuprofile=cpu.prof -memprofile=mem.prof
	@echo "✅ Profiling data generated: cpu.prof, mem.prof"

# Example workflows
examples:
	@echo "🎯 Running example workflows..."
	@./scripts/run-examples.sh
	@echo "✅ Examples complete"

# Health check
health:
	@echo "🏥 Checking application health..."
	@curl -f http://localhost:8080/health || echo "❌ Application not healthy"

# Load test
load-test:
	@echo "⚡ Running load test..."
	@if command -v hey >/dev/null 2>&1; then \
		hey -n 1000 -c 10 http://localhost:8080/health; \
	else \
		echo "⚠️  hey not installed, install with: go install github.com/rakyll/hey@latest"; \
	fi

# Database migration (placeholder for future use)
migrate:
	@echo "🗄️  Running database migrations..."
	@echo "⚠️  Database migrations not implemented yet"

# Backup (placeholder for future use)
backup:
	@echo "💾 Creating backup..."
	@echo "⚠️  Backup functionality not implemented yet"

# Release
release: clean lint test build-all
	@echo "🚀 Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@tar -czf release/$(BINARY_NAME)-$(VERSION).tar.gz -C release .
	@echo "✅ Release created: release/$(BINARY_NAME)-$(VERSION).tar.gz"

# Help
help:
	@echo "🤖 Go Coffee LangGraph Integration - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         Build the application"
	@echo "  build-all     Build for multiple platforms"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  run           Build and run the application"
	@echo "  dev           Run in development mode"
	@echo "  fmt           Format code"
	@echo "  tidy          Tidy dependencies"
	@echo "  deps          Download dependencies"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test          Run unit tests"
	@echo "  test-integration  Run integration tests"
	@echo "  coverage      Run tests with coverage"
	@echo "  benchmark     Run benchmarks"
	@echo "  lint          Lint code"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker        Build Docker image"
	@echo "  docker-push   Build and push Docker image"
	@echo "  docker-run    Run with Docker"
	@echo ""
	@echo "Utility Commands:"
	@echo "  docs          Generate documentation"
	@echo "  security      Run security scan"
	@echo "  profile       Performance profiling"
	@echo "  examples      Run example workflows"
	@echo "  health        Check application health"
	@echo "  load-test     Run load test"
	@echo ""
	@echo "Release Commands:"
	@echo "  release       Create release package"
	@echo ""
	@echo "Tools:"
	@echo "  install-tools Install development tools"
	@echo "  help          Show this help message"
