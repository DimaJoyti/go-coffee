# Development Guide

This guide provides information for developers who want to contribute to or extend the Coffee Order System.

## Development Environment Setup

### Prerequisites

- **Go**: Version 1.22 or higher
- **Kafka**: A running Kafka instance
- **Git**: For version control
- **IDE**: Any Go-compatible IDE (VSCode, GoLand, etc.)

### Setting Up the Development Environment

1. Clone the repository:

```bash
git clone https://github.com/yourusername/coffee-order-system.git
cd coffee-order-system
```

2. Install dependencies:

```bash
# Producer dependencies
cd producer
go mod tidy

# Consumer dependencies
cd ../consumer
go mod tidy
```

3. Set up Kafka (if not already running):

```bash
# Using Docker
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e ALLOW_PLAINTEXT_LISTENER=yes \
  bitnami/kafka
```

## Project Structure

### Producer Service

```
producer/
├── config/
│   └── config.go       # Configuration management
├── handler/
│   ├── handler.go      # HTTP handlers
│   └── handler_test.go # Handler tests
├── kafka/
│   └── producer.go     # Kafka producer
├── middleware/
│   ├── middleware.go      # HTTP middleware
│   └── middleware_test.go # Middleware tests
├── config.json         # Configuration file
├── go.mod              # Go module file
├── go.sum              # Go module checksum
└── main.go             # Entry point
```

### Consumer Service

```
consumer/
├── config/
│   └── config.go       # Configuration management
├── kafka/
│   └── consumer.go     # Kafka consumer
├── config.json         # Configuration file
├── go.mod              # Go module file
├── go.sum              # Go module checksum
└── main.go             # Entry point
```

## Adding a New Feature

### Adding a New API Endpoint

1. Add a new handler function in `producer/handler/handler.go`:

```go
// NewEndpoint handles requests to the new endpoint
func (h *Handler) NewEndpoint(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

2. Register the handler in `producer/main.go`:

```go
mux.HandleFunc("/new-endpoint", h.NewEndpoint)
```

3. Add tests for the new handler in `producer/handler/handler_test.go`.

### Adding a New Kafka Topic

1. Update the configuration in `producer/config/config.go` and `consumer/config/config.go`:

```go
type KafkaConfig struct {
    Brokers      []string `json:"brokers"`
    Topic        string   `json:"topic"`
    NewTopic     string   `json:"new_topic"` // Add this line
    RetryMax     int      `json:"retry_max"`
    RequiredAcks string   `json:"required_acks"`
}
```

2. Update the configuration loading in both services.

3. Add code to publish to and consume from the new topic.

## Coding Standards

### Go Formatting

Use `gofmt` or `go fmt` to format your code:

```bash
go fmt ./...
```

### Error Handling

Always check errors and return them to the caller:

```go
result, err := someFunction()
if err != nil {
    return err
}
```

### Logging

Use the standard `log` package for logging:

```go
log.Printf("Something happened: %v", value)
```

### Testing

Write tests for all new functionality:

```go
func TestSomething(t *testing.T) {
    // Test implementation
}
```

Run tests with:

```bash
go test ./...
```

## Pull Request Process

1. Fork the repository.
2. Create a new branch for your feature.
3. Implement your feature.
4. Write tests for your feature.
5. Run all tests to ensure they pass.
6. Submit a pull request.

## Continuous Integration

The project does not currently have a CI/CD pipeline set up. However, you can run the following commands locally to ensure your changes meet the project standards:

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Check for common mistakes
go vet ./...
```

## Next Steps

- [Testing](testing.md): Learn more about testing the system.
- [API Reference](api-reference.md): Explore the API endpoints.
- [Kafka Integration](kafka-integration.md): Learn about Kafka integration.
