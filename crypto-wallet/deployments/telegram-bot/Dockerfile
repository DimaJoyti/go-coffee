# Multi-stage build for Telegram Bot
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the telegram bot binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-bot cmd/telegram-bot/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/telegram-bot .

# Copy configuration files
COPY --from=builder /app/config ./config

# Create logs directory
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (if using webhook mode)
EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1

# Run the telegram bot
CMD ["./telegram-bot"]
