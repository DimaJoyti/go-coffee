# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the DeFi service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o defi-service ./cmd/defi-service

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/defi-service .

# Copy configuration files
COPY --from=builder /app/config ./config

# Create logs directory
RUN mkdir -p logs

# Expose ports
EXPOSE 8085 50055

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

# Run the service
CMD ["./defi-service"]
