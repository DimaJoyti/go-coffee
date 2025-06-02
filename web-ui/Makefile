# Go Coffee Epic UI - Makefile

.PHONY: help install build start stop clean logs test dev

# Default target
help:
	@echo "🚀 Go Coffee Epic UI - Available Commands:"
	@echo ""
	@echo "📦 Setup & Installation:"
	@echo "  make install     - Install dependencies for frontend and backend"
	@echo "  make build       - Build frontend and backend"
	@echo ""
	@echo "🏃 Development:"
	@echo "  make dev         - Start development servers (frontend + backend)"
	@echo "  make dev-frontend - Start only frontend development server"
	@echo "  make dev-backend  - Start only backend development server"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make start       - Start all services with Docker Compose"
	@echo "  make stop        - Stop all services"
	@echo "  make restart     - Restart all services"
	@echo "  make logs        - Show logs from all services"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  make clean       - Clean up containers, images, and volumes"
	@echo "  make test        - Run tests"
	@echo ""

# Install dependencies
install:
	@echo "📦 Installing frontend dependencies..."
	cd frontend && npm install
	@echo "📦 Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "✅ Dependencies installed!"

# Build applications
build:
	@echo "🔨 Building frontend..."
	cd frontend && npm run build
	@echo "🔨 Building backend..."
	cd backend && go build -o bin/web-ui-service ./cmd/web-ui-service
	@echo "✅ Build completed!"

# Development servers
dev:
	@echo "🏃 Starting development servers..."
	@echo "Frontend will be available at: http://localhost:3000"
	@echo "Backend will be available at: http://localhost:8090"
	@echo ""
	@echo "Press Ctrl+C to stop both servers"
	@make -j2 dev-frontend dev-backend

dev-frontend:
	@echo "🎨 Starting frontend development server..."
	cd frontend && npm run dev

dev-backend:
	@echo "⚙️  Starting backend development server..."
	cd backend && go run ./cmd/web-ui-service/main.go

# Docker operations
start:
	@echo "🚀 Starting Go Coffee Epic UI with Docker..."
	./start.sh

stop:
	@echo "🛑 Stopping Go Coffee Epic UI..."
	./stop.sh

restart: stop start

logs:
	@echo "📋 Showing logs from all services..."
	docker-compose -f docker-compose.ui.yml logs -f

# Testing
test:
	@echo "🧪 Running frontend tests..."
	cd frontend && npm run test
	@echo "🧪 Running backend tests..."
	cd backend && go test ./...

# Cleanup
clean:
	@echo "🧹 Cleaning up Docker resources..."
	docker-compose -f docker-compose.ui.yml down -v --remove-orphans
	docker system prune -f
	@echo "🧹 Cleaning build artifacts..."
	rm -rf frontend/.next
	rm -rf frontend/node_modules/.cache
	rm -rf backend/bin
	@echo "✅ Cleanup completed!"

# Health check
health:
	@echo "❤️  Checking service health..."
	@curl -s http://localhost:8090/health | jq . || echo "Backend not responding"
	@curl -s http://localhost:3000 > /dev/null && echo "✅ Frontend is running" || echo "❌ Frontend not responding"

# Show status
status:
	@echo "📊 Service Status:"
	@docker-compose -f docker-compose.ui.yml ps

# Quick setup for new developers
setup: install build
	@echo "🎉 Setup completed! You can now run:"
	@echo "  make dev    - for development"
	@echo "  make start  - for Docker deployment"
