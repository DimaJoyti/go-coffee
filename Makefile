# Go Coffee - Simplified Build System
# ===================================

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Project configuration
PROJECT_NAME := go-coffee
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build configuration
BUILD_DIR := bin
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Service definitions
CORE_SERVICES := api-gateway producer consumer streams
WEB3_SERVICES := auth-service order-service kitchen-service payment-service
AI_SERVICES := ai-arbitrage-service ai-order-service
INFRASTRUCTURE_SERVICES := security-gateway user-gateway

ALL_SERVICES := $(CORE_SERVICES) $(WEB3_SERVICES) $(AI_SERVICES) $(INFRASTRUCTURE_SERVICES)

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Go Coffee - Build System$(NC)"
	@echo "========================="
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# =============================================================================
# Build Targets
# =============================================================================

.PHONY: build
build: clean deps build-core build-web3 ## Build all services

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(CYAN)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean -cache
	@echo "$(GREEN)✅ Clean complete$(NC)"

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "$(CYAN)📦 Managing dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

.PHONY: build-core
build-core: ## Build core coffee services
	@echo "$(CYAN)☕ Building Core Coffee Services...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for service in $(CORE_SERVICES); do \
		echo "$(BLUE)Building $$service...$(NC)"; \
		if [ -d "$$service" ]; then \
			cd $$service && go build $(LDFLAGS) -o ../$(BUILD_DIR)/$$service ./main.go && echo "$(GREEN)  ✅ $$service built$(NC)" && cd ..; \
		elif [ -d "cmd/$$service" ]; then \
			go build $(LDFLAGS) -o $(BUILD_DIR)/$$service ./cmd/$$service && echo "$(GREEN)  ✅ $$service built$(NC)"; \
		else \
			echo "$(YELLOW)  ⚠️  $$service directory not found$(NC)"; \
		fi; \
	done

.PHONY: build-web3
build-web3: ## Build Web3 services
	@echo "$(CYAN)🌐 Building Web3 Services...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for service in $(WEB3_SERVICES); do \
		echo "$(BLUE)Building $$service...$(NC)"; \
		if [ -d "cmd/$$service" ]; then \
			go build $(LDFLAGS) -o $(BUILD_DIR)/$$service ./cmd/$$service && echo "$(GREEN)  ✅ $$service built$(NC)"; \
		else \
			echo "$(YELLOW)  ⚠️  $$service directory not found$(NC)"; \
		fi; \
	done

.PHONY: test
test: ## Run tests
	@echo "$(CYAN)🧪 Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✅ Tests complete$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(CYAN)🔍 Running linter...$(NC)"
	@golangci-lint run
	@echo "$(GREEN)✅ Linting complete$(NC)"

.PHONY: format
format: ## Format code
	@echo "$(CYAN)🎨 Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .
	@echo "$(GREEN)✅ Formatting complete$(NC)"

# =============================================================================
# Docker Targets
# =============================================================================

.PHONY: docker-build
docker-build: ## Build Docker images
	@echo "$(CYAN)🐳 Building Docker images...$(NC)"
	@docker-compose build
	@echo "$(GREEN)✅ Docker build complete$(NC)"

.PHONY: docker-up
docker-up: ## Start services with Docker Compose
	@echo "$(CYAN)🚀 Starting services...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✅ Services started$(NC)"

.PHONY: docker-down
docker-down: ## Stop services
	@echo "$(CYAN)⏹️ Stopping services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✅ Services stopped$(NC)"

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	@docker-compose logs -f

# =============================================================================
# Development Targets
# =============================================================================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(CYAN)🔧 Setting up development environment...$(NC)"
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✅ Development environment ready$(NC)"

.PHONY: run-auth
run-auth: ## Run auth service
	@echo "$(CYAN)🔐 Starting auth service...$(NC)"
	@go run cmd/auth-service/main.go

.PHONY: run-simple-auth
run-simple-auth: ## Run simple auth service
	@echo "$(CYAN)🔐 Starting simple auth service...$(NC)"
	@go run cmd/auth-service/simple_main.go

# =============================================================================
# Utility Targets
# =============================================================================

.PHONY: check
check: lint test ## Run all checks

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(CYAN)🔧 Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✅ Tools installed$(NC)"
