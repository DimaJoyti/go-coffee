# Crypto Terminal Makefile

# Variables
APP_NAME := crypto-terminal
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Directories
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

# Docker variables
DOCKER_IMAGE := crypto-terminal
DOCKER_TAG := $(VERSION)
DOCKER_REGISTRY := localhost:5000

.PHONY: all build clean test coverage lint fmt vet deps docker-build docker-run docker-push help

# Default target
all: clean fmt vet test build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/terminal

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/terminal
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/terminal
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/terminal
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/terminal
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/terminal

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -v -bench=. -benchmem ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration ./test/...

# Database migration targets
migrate-up:
	@echo "Running database migrations up..."
	@which migrate > /dev/null || (echo "Please install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/crypto_terminal?sslmode=disable" up

migrate-down:
	@echo "Running database migrations down..."
	@which migrate > /dev/null || (echo "Please install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/crypto_terminal?sslmode=disable" down

# Lint the code
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Vet the code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run ./cmd/terminal

# Run with development configuration
run-dev:
	@echo "Running $(APP_NAME) in development mode..."
	LOG_LEVEL=debug $(GOCMD) run ./cmd/terminal

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8090:8090 --name $(APP_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run-dev:
	@echo "Running Docker container with development setup..."
	docker-compose up --build

docker-push:
	@echo "Pushing Docker image..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)

# Database targets
db-up:
	@echo "Starting database..."
	docker-compose up -d postgres redis

db-down:
	@echo "Stopping database..."
	docker-compose down

db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis

# Development targets
dev-setup:
	@echo "Setting up development environment..."
	$(MAKE) deps
	$(MAKE) db-up
	@echo "Waiting for database to be ready..."
	@sleep 10
	@echo "Development environment ready!"

dev-run:
	@echo "Running development environment..."
	$(MAKE) dev-setup
	$(MAKE) run-dev

# Production targets
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME) ./cmd/terminal

prod-deploy:
	@echo "Deploying to production..."
	$(MAKE) prod-build
	$(MAKE) docker-build
	$(MAKE) docker-push

# Monitoring targets
monitor-up:
	@echo "Starting monitoring stack..."
	docker-compose up -d prometheus grafana

monitor-down:
	@echo "Stopping monitoring stack..."
	docker-compose stop prometheus grafana

# Security targets
security-scan:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Generate targets
generate:
	@echo "Running go generate..."
	$(GOCMD) generate ./...

# Install tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  bench         - Run benchmarks"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  run           - Run the application"
	@echo "  run-dev       - Run in development mode"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-run-dev- Run with docker-compose"
	@echo "  docker-push   - Push Docker image"
	@echo "  db-up         - Start database"
	@echo "  db-down       - Stop database"
	@echo "  db-reset      - Reset database"
	@echo "  dev-setup     - Setup development environment"
	@echo "  dev-run       - Run development environment"
	@echo "  prod-build    - Build for production"
	@echo "  prod-deploy   - Deploy to production"
	@echo "  monitor-up    - Start monitoring stack"
	@echo "  monitor-down  - Stop monitoring stack"
	@echo "  security-scan - Run security scan"
	@echo "  generate      - Run go generate"
	@echo "  install-tools - Install development tools"
	@echo "  help          - Show this help"
