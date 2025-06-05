# API Gateway Makefile

.PHONY: build run test clean proto deps help

# Default target
all: deps proto build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@mkdir -p proto/coffee
	@if command -v protoc >/dev/null 2>&1; then \
		protoc -I=proto \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go-grpc_out=. \
			--go-grpc_opt=paths=source_relative \
			proto/coffee_service.proto; \
		echo "Protobuf code generated successfully"; \
	else \
		echo "protoc not found, using pre-generated files"; \
	fi

# Build the application
build:
	@echo "Building API Gateway..."
	go build -o bin/api-gateway .
	@echo "Build completed successfully"

# Run the application
run:
	@echo "Starting API Gateway..."
	go run .

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Test build without running
test-build:
	@echo "Testing build..."
	go run test_build.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run linter (if golangci-lint is installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Development workflow
dev: fmt vet build test-build

# Help
help:
	@echo "Available targets:"
	@echo "  deps       - Install dependencies"
	@echo "  proto      - Generate protobuf code"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  test       - Run tests"
	@echo "  test-build - Test build without running"
	@echo "  clean      - Clean build artifacts"
	@echo "  fmt        - Format code"
	@echo "  vet        - Vet code"
	@echo "  lint       - Run linter"
	@echo "  dev        - Development workflow (fmt, vet, build, test-build)"
	@echo "  help       - Show this help"
