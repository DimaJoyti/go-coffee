# Fintech Platform Makefile

.PHONY: help build run test clean docker-build docker-run fintech-build fintech-run

# Variables
BINARY_NAME=fintech-platform
DOCKER_IMAGE=fintech-platform
VERSION?=latest
GOOS?=linux
GOARCH?=amd64

# Default target
help:
	@echo "🏦 Fintech Platform - Available Commands"
	@echo ""
	@echo "🔨 Build Commands:"
	@echo "  build              - Build all services"
	@echo "  fintech-build      - Build fintech API service"
	@echo "  build-linux        - Build for Linux (production)"
	@echo ""
	@echo "🚀 Run Commands:"
	@echo "  run                - Run all services with Docker Compose"
	@echo "  fintech-run        - Run fintech API service locally"
	@echo "  run-dev            - Run in development mode"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  test               - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-performance   - Run performance tests"
	@echo "  test-security      - Run security tests"
	@echo "  test-all           - Run all tests"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-build       - Build Docker images"
	@echo "  docker-run         - Run with Docker Compose"
	@echo "  docker-stop        - Stop Docker services"
	@echo "  docker-clean       - Clean Docker resources"
	@echo ""
	@echo "☸️  Kubernetes Commands:"
	@echo "  k8s-deploy         - Deploy to Kubernetes"
	@echo "  k8s-delete         - Delete from Kubernetes"
	@echo "  helm-install       - Install with Helm"
	@echo "  helm-upgrade       - Upgrade with Helm"
	@echo ""
	@echo "🔧 Database Commands:"
	@echo "  migrate-up         - Run database migrations"
	@echo "  migrate-down       - Rollback database migrations"
	@echo "  migrate-create     - Create new migration"
	@echo ""
	@echo "🧹 Cleanup Commands:"
	@echo "  clean              - Clean build artifacts"
	@echo "  clean-all          - Clean everything"

# Build Commands
build: fintech-build

fintech-build:
	@echo "🔨 Building Fintech API..."
	cd web3-wallet-backend && go build -o ../bin/$(BINARY_NAME) ./cmd/fintech/main.go

build-linux:
	@echo "🔨 Building for Linux..."
	cd web3-wallet-backend && GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ../bin/$(BINARY_NAME)-linux ./cmd/fintech/main.go

# Run Commands
run: docker-run

fintech-run:
	@echo "🚀 Starting Fintech API..."
	cd web3-wallet-backend && go run ./cmd/fintech/main.go

run-dev:
	@echo "🚀 Starting development environment..."
	docker-compose -f docker-compose.fintech.yml up -d postgres redis
	@echo "⏳ Waiting for services to be ready..."
	sleep 10
	$(MAKE) fintech-run

# Test Commands
test:
	@echo "🧪 Running unit tests..."
	cd web3-wallet-backend && go test -v -race -coverprofile=coverage.out ./...

test-integration:
	@echo "🧪 Running integration tests..."
	cd web3-wallet-backend && INTEGRATION_TESTS=1 go test -v -tags=integration ./...

test-performance:
	@echo "🧪 Running performance tests..."
	k6 run tests/performance/load-test.js

test-security:
	@echo "🧪 Running security tests..."
	cd web3-wallet-backend && gosec ./...
	cd web3-wallet-backend && go list -json -deps ./... | nancy sleuth

test-all: test test-integration test-performance test-security

# Docker Commands
docker-build:
	@echo "🐳 Building Docker images..."
	docker build -f web3-wallet-backend/Dockerfile.fintech -t $(DOCKER_IMAGE):$(VERSION) .

docker-run:
	@echo "🐳 Starting services with Docker Compose..."
	docker-compose -f docker-compose.fintech.yml up -d

docker-stop:
	@echo "🐳 Stopping Docker services..."
	docker-compose -f docker-compose.fintech.yml down

docker-clean:
	@echo "🐳 Cleaning Docker resources..."
	docker-compose -f docker-compose.fintech.yml down -v --remove-orphans
	docker system prune -f

# Kubernetes Commands
k8s-deploy:
	@echo "☸️  Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secrets.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/fintech-api.yaml
	kubectl apply -f k8s/monitoring.yaml

k8s-delete:
	@echo "☸️  Deleting from Kubernetes..."
	kubectl delete -f k8s/ --ignore-not-found=true

helm-install:
	@echo "☸️  Installing with Helm..."
	helm install fintech-platform ./helm-chart \
		--namespace fintech-platform \
		--create-namespace \
		--values helm-chart/values.yaml

helm-upgrade:
	@echo "☸️  Upgrading with Helm..."
	helm upgrade fintech-platform ./helm-chart \
		--namespace fintech-platform \
		--values helm-chart/values.yaml

# Database Commands
migrate-up:
	@echo "🗄️  Running database migrations..."
	cd web3-wallet-backend && go run ./cmd/migrate/main.go up

migrate-down:
	@echo "🗄️  Rolling back database migrations..."
	cd web3-wallet-backend && go run ./cmd/migrate/main.go down

migrate-create:
	@echo "🗄️  Creating new migration..."
	@read -p "Enter migration name: " name; \
	cd web3-wallet-backend && go run ./cmd/migrate/main.go create $$name

# Development Commands
deps:
	@echo "📦 Installing dependencies..."
	cd web3-wallet-backend && go mod download
	cd web3-wallet-backend && go mod tidy

lint:
	@echo "🔍 Running linter..."
	cd web3-wallet-backend && golangci-lint run

format:
	@echo "🎨 Formatting code..."
	cd web3-wallet-backend && go fmt ./...
	cd web3-wallet-backend && goimports -w .

security-scan:
	@echo "🔒 Running security scan..."
	cd web3-wallet-backend && gosec ./...
	cd web3-wallet-backend && go list -json -deps ./... | nancy sleuth

# Monitoring Commands
logs:
	@echo "📋 Showing application logs..."
	docker-compose -f docker-compose.fintech.yml logs -f fintech-api

logs-db:
	@echo "📋 Showing database logs..."
	docker-compose -f docker-compose.fintech.yml logs -f postgres

logs-redis:
	@echo "📋 Showing Redis logs..."
	docker-compose -f docker-compose.fintech.yml logs -f redis

# Cleanup Commands
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web3-wallet-backend/coverage.out

clean-all: clean docker-clean
	@echo "🧹 Cleaning everything..."
	cd web3-wallet-backend && go clean -cache -modcache -testcache

# Utility Commands
check-deps:
	@echo "🔍 Checking dependencies..."
	cd web3-wallet-backend && go mod verify
	cd web3-wallet-backend && go list -u -m all

update-deps:
	@echo "📦 Updating dependencies..."
	cd web3-wallet-backend && go get -u ./...
	cd web3-wallet-backend && go mod tidy

# Environment setup
setup-dev:
	@echo "🔧 Setting up development environment..."
	cp .env.fintech.example .env
	$(MAKE) deps
	$(MAKE) docker-run
	@echo "⏳ Waiting for services..."
	sleep 30
	$(MAKE) migrate-up
	@echo "✅ Development environment ready!"

# Production deployment
deploy-prod:
	@echo "🚀 Deploying to production..."
	$(MAKE) test-all
	$(MAKE) docker-build
	$(MAKE) k8s-deploy
	@echo "✅ Production deployment complete!"
