FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o security-service ./cmd/security-service

# Use a minimal alpine image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/security-service .

# Copy configuration
COPY config/config.yaml /app/config/

# Expose port
EXPOSE 50054

# Set the entrypoint
ENTRYPOINT ["/app/security-service"]
