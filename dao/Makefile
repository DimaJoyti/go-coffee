# Developer DAO Platform Makefile

# Variables
BINARY_NAME=developer-dao
VERSION=1.0.0
BUILD_DIR=./bin
CONTRACTS_DIR=./contracts
MIGRATIONS_DIR=./migrations
DOCKER_COMPOSE_FILE=docker-compose.yml

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Docker variables
DOCKER=docker
DOCKER_COMPOSE=docker-compose

# Database variables
DB_HOST=localhost
DB_PORT=5432
DB_NAME=developer_dao
DB_USER=postgres
DB_PASSWORD=password
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: all build clean test deps fmt lint vet security docker-build docker-up docker-down migrate-up migrate-down contracts-install contracts-compile contracts-test contracts-deploy help

# Default target
all: clean deps fmt lint vet test build

# Build all services
build: build-dao-governance build-bounty-service build-solution-marketplace build-developer-portal build-metrics-aggregator

build-dao-governance:
	@echo "Building DAO Governance Service..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/dao-governance-service ./cmd/dao-governance-service

build-bounty-service:
	@echo "Building Bounty Service..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/bounty-service ./cmd/bounty-service

build-solution-marketplace:
	@echo "Building Solution Marketplace..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/solution-marketplace ./cmd/solution-marketplace

build-developer-portal:
	@echo "Building Developer Portal..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/developer-portal ./cmd/developer-portal

build-metrics-aggregator:
	@echo "Building Metrics Aggregator..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/metrics-aggregator ./cmd/metrics-aggregator

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf ./coverage.out

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Security check
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up; \
	else \
		echo "migrate not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

migrate-down:
	@echo "Running database migrations down..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down; \
	else \
		echo "migrate not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

migrate-create:
	@echo "Creating new migration: $(name)"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name); \
	else \
		echo "migrate not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

# Smart contracts
contracts-install:
	@echo "Installing smart contract dependencies..."
	cd $(CONTRACTS_DIR) && npm install

contracts-compile:
	@echo "Compiling smart contracts..."
	cd $(CONTRACTS_DIR) && npx hardhat compile

contracts-test:
	@echo "Testing smart contracts..."
	cd $(CONTRACTS_DIR) && npx hardhat test

contracts-deploy:
	@echo "Deploying smart contracts to $(network)..."
	cd $(CONTRACTS_DIR) && npx hardhat run scripts/deploy.js --network $(network)

contracts-verify:
	@echo "Verifying smart contracts..."
	cd $(CONTRACTS_DIR) && npx hardhat verify --network $(network) $(address)

contracts-clean:
	@echo "Cleaning smart contract artifacts..."
	cd $(CONTRACTS_DIR) && npx hardhat clean

# Docker operations
docker-build:
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build

docker-up:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

docker-down:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

docker-logs:
	@echo "Showing Docker logs..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

# Development
dev-setup: deps contracts-install migrate-up
	@echo "Development environment setup complete!"

dev-run: build
	@echo "Starting development services..."
	./$(BUILD_DIR)/dao-governance-service &
	./$(BUILD_DIR)/bounty-service &
	./$(BUILD_DIR)/solution-marketplace &
	./$(BUILD_DIR)/developer-portal &
	./$(BUILD_DIR)/metrics-aggregator &

# Production
prod-build: clean deps fmt lint vet test build
	@echo "Production build complete!"

prod-deploy: prod-build docker-build
	@echo "Production deployment ready!"

# Utilities
generate-proto:
	@echo "Generating protobuf files..."
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto; \
	else \
		echo "protoc not installed. Please install Protocol Buffers compiler"; \
	fi

benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

profile:
	@echo "Running CPU profile..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	$(GOCMD) tool pprof cpu.prof

# Database utilities
db-create:
	@echo "Creating database..."
	createdb $(DB_NAME)

db-drop:
	@echo "Dropping database..."
	dropdb $(DB_NAME)

db-reset: db-drop db-create migrate-up
	@echo "Database reset complete!"

# Help
help:
	@echo "Available targets:"
	@echo "  all              - Clean, deps, fmt, lint, vet, test, build"
	@echo "  build            - Build all services"
	@echo "  clean            - Clean build artifacts"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  deps             - Install dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  vet              - Vet code"
	@echo "  security         - Run security checks"
	@echo "  migrate-up       - Run database migrations up"
	@echo "  migrate-down     - Run database migrations down"
	@echo "  migrate-create   - Create new migration (use: make migrate-create name=migration_name)"
	@echo "  contracts-*      - Smart contract operations"
	@echo "  docker-*         - Docker operations"
	@echo "  dev-setup        - Setup development environment"
	@echo "  dev-run          - Run development services"
	@echo "  prod-build       - Production build"
	@echo "  prod-deploy      - Production deployment"
	@echo "  help             - Show this help message"
