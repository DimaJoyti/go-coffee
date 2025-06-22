# Multi-stage build for Kitchen Service
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o kitchen-service \
    cmd/kitchen-service/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/kitchen-service .

# Copy configuration files (if any)
COPY --from=builder /app/configs ./configs

# Create logs directory
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090 8081 9091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Set default environment variables
ENV SERVICE_NAME=kitchen-service \
    SERVICE_VERSION=1.0.0 \
    ENVIRONMENT=production \
    LOG_LEVEL=info \
    HTTP_PORT=8080 \
    GRPC_PORT=9090 \
    HEALTH_PORT=8081 \
    METRICS_PORT=9091

# Run the application
CMD ["./kitchen-service"]
