FROM golang:1.22-alpine AS builder

# Install necessary dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o accounts-service ./cmd/accounts-service

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Create a non-privileged user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/accounts-service .

# Copy configuration files
COPY config.json .

# Copy migrations
COPY db/migrations ./db/migrations

# Set ownership to the non-privileged user
RUN chown -R appuser:appgroup /app

# Switch to the non-privileged user
USER appuser

# Expose the port
EXPOSE 4000

# Set the entrypoint
ENTRYPOINT ["/app/accounts-service"]
