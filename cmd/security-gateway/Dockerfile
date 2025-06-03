# Multi-stage build for Security Gateway

# Build stage
FROM golang:1.21-alpine AS builder

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
    -o security-gateway \
    ./cmd/security-gateway

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/security-gateway .

# Copy configuration files
COPY --from=builder /app/cmd/security-gateway/config ./config

# Create directories for logs and data
RUN mkdir -p /var/log /var/data && \
    chown -R appuser:appgroup /app /var/log /var/data

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV LOG_LEVEL=info
ENV ENVIRONMENT=production

# Run the application
CMD ["./security-gateway"]

# Labels for metadata
LABEL maintainer="Go Coffee Team <team@go-coffee.com>"
LABEL version="1.0.0"
LABEL description="Security Gateway for Go Coffee microservices"
LABEL org.opencontainers.image.title="Security Gateway"
LABEL org.opencontainers.image.description="Enterprise-grade security gateway with WAF, rate limiting, and threat detection"
LABEL org.opencontainers.image.vendor="Go Coffee"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/DimaJoyti/go-coffee"
