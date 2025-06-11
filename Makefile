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
LLM_SERVICES := llm-orchestrator llm-orchestrator-simple

ALL_SERVICES := $(CORE_SERVICES) $(WEB3_SERVICES) $(AI_SERVICES) $(INFRASTRUCTURE_SERVICES) $(LLM_SERVICES)

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
build: clean deps build-core build-web3 build-llm ## Build all services

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

.PHONY: build-llm
build-llm: ## Build LLM services
	@echo "$(CYAN)🤖 Building LLM Services...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for service in $(LLM_SERVICES); do \
		echo "$(BLUE)Building $$service...$(NC)"; \
		if [ -d "cmd/$$service" ]; then \
			go build $(LDFLAGS) -o $(BUILD_DIR)/$$service ./cmd/$$service && echo "$(GREEN)  ✅ $$service built$(NC)"; \
		else \
			echo "$(YELLOW)  ⚠️  $$service directory not found$(NC)"; \
		fi; \
	done

.PHONY: build-llm-orchestrator
build-llm-orchestrator: ## Build LLM Orchestrator specifically
	@echo "$(CYAN)🤖 Building LLM Orchestrator...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/llm-orchestrator ./cmd/llm-orchestrator
	@echo "$(GREEN)✅ LLM Orchestrator built$(NC)"

.PHONY: build-llm-orchestrator-simple
build-llm-orchestrator-simple: ## Build Simple LLM Orchestrator
	@echo "$(CYAN)🤖 Building Simple LLM Orchestrator...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/llm-orchestrator-simple ./cmd/llm-orchestrator-simple
	@echo "$(GREEN)✅ Simple LLM Orchestrator built$(NC)"

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

# =============================================================================
# LLM Orchestrator Targets
# =============================================================================

.PHONY: test-llm-orchestrator
test-llm-orchestrator: ## Run LLM orchestrator tests
	@echo "$(CYAN)🧪 Running LLM orchestrator tests...$(NC)"
	@go test -v -race ./internal/llm-orchestrator/...
	@echo "$(GREEN)✅ LLM orchestrator tests complete$(NC)"

.PHONY: docker-build-llm-orchestrator
docker-build-llm-orchestrator: ## Build LLM Orchestrator Docker image
	@echo "$(CYAN)🐳 Building LLM Orchestrator Docker image...$(NC)"
	@docker build -f docker/Dockerfile.llm-orchestrator -t ghcr.io/dimajoyti/go-coffee-llm-orchestrator:$(VERSION) .
	@echo "$(GREEN)✅ LLM Orchestrator Docker image built$(NC)"

.PHONY: deploy-llm-orchestrator
deploy-llm-orchestrator: ## Deploy LLM Orchestrator to Kubernetes
	@echo "$(CYAN)🚀 Deploying LLM Orchestrator...$(NC)"
	@kubectl apply -f k8s/llm-orchestrator/namespace.yaml
	@kubectl apply -f k8s/llm-orchestrator/crd.yaml
	@kubectl apply -f k8s/llm-orchestrator/rbac.yaml
	@kubectl apply -f k8s/llm-orchestrator/deployment.yaml
	@echo "$(GREEN)✅ LLM Orchestrator deployed$(NC)"

.PHONY: undeploy-llm-orchestrator
undeploy-llm-orchestrator: ## Remove LLM Orchestrator from Kubernetes
	@echo "$(CYAN)🗑️ Removing LLM Orchestrator...$(NC)"
	@kubectl delete -f k8s/llm-orchestrator/deployment.yaml --ignore-not-found=true
	@kubectl delete -f k8s/llm-orchestrator/rbac.yaml --ignore-not-found=true
	@kubectl delete -f k8s/llm-orchestrator/crd.yaml --ignore-not-found=true
	@kubectl delete -f k8s/llm-orchestrator/namespace.yaml --ignore-not-found=true
	@echo "$(GREEN)✅ LLM Orchestrator removed$(NC)"

.PHONY: logs-llm-orchestrator
logs-llm-orchestrator: ## View LLM Orchestrator logs
	@echo "$(CYAN)📋 Viewing LLM Orchestrator logs...$(NC)"
	@kubectl logs -f -n llm-orchestrator deployment/llm-orchestrator

.PHONY: status-llm-orchestrator
status-llm-orchestrator: ## Check LLM Orchestrator status
	@echo "$(CYAN)📊 Checking LLM Orchestrator status...$(NC)"
	@kubectl get pods -n llm-orchestrator
	@kubectl get llmworkloads -A

.PHONY: run-llm-orchestrator
run-llm-orchestrator: build-llm-orchestrator ## Run LLM Orchestrator locally
	@echo "$(CYAN)🤖 Starting LLM Orchestrator locally...$(NC)"
	@./bin/llm-orchestrator --config=config/llm-orchestrator.yaml --zap-log-level=info

# Simple LLM Orchestrator Targets
.PHONY: build-simple-llm
build-simple-llm: build-llm-orchestrator-simple ## Build Simple LLM Orchestrator

.PHONY: run-simple-llm
run-simple-llm: build-llm-orchestrator-simple ## Run Simple LLM Orchestrator locally
	@echo "$(CYAN)🤖 Starting Simple LLM Orchestrator locally...$(NC)"
	@./bin/llm-orchestrator-simple --config=config/llm-orchestrator-simple.yaml --port=8080 --log-level=info

.PHONY: docker-build-simple-llm
docker-build-simple-llm: ## Build Simple LLM Orchestrator Docker image
	@echo "$(CYAN)🐳 Building Simple LLM Orchestrator Docker image...$(NC)"
	@docker build -f docker/Dockerfile.llm-orchestrator-simple -t ghcr.io/dimajoyti/go-coffee-llm-orchestrator-simple:$(VERSION) .
	@echo "$(GREEN)✅ Simple LLM Orchestrator Docker image built$(NC)"

.PHONY: test-simple-llm
test-simple-llm: ## Test Simple LLM Orchestrator
	@echo "$(CYAN)🧪 Testing Simple LLM Orchestrator...$(NC)"
	@go test -v ./cmd/llm-orchestrator-simple/...
	@echo "$(GREEN)✅ Simple LLM Orchestrator tests complete$(NC)"
