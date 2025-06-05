# API Gateway

The API Gateway serves as the entry point for the Go Coffee microservices system, providing HTTP REST endpoints that communicate with backend gRPC services.

## Features

- **HTTP REST API** - Clean REST endpoints for coffee orders
- **gRPC Client** - Communicates with Producer and Consumer services
- **Middleware Support** - CORS, logging, and request ID tracking
- **Health Checks** - Built-in health monitoring
- **Configuration** - Environment-based configuration
- **Error Handling** - Comprehensive error handling and logging

## API Endpoints

### Orders
- `POST /order` - Create a new coffee order
- `GET /order/{id}` - Get order details by ID
- `POST /order/{id}/cancel` - Cancel an order
- `GET /orders` - List all orders

### Health
- `GET /health` - Health check endpoint

## Configuration

The service can be configured using environment variables or a `config.json` file:

### Environment Variables
```bash
SERVER_PORT=8080                          # HTTP server port
PRODUCER_GRPC_ADDRESS=localhost:50051     # Producer service gRPC address
CONSUMER_GRPC_ADDRESS=localhost:50052     # Consumer service gRPC address
GRPC_CONNECTION_TIMEOUT=5s                # gRPC connection timeout
GRPC_MAX_RETRIES=3                        # Maximum connection retries
GRPC_RETRY_DELAY=1s                       # Delay between retries
```

### Configuration File
Create a `config.json` file:
```json
{
  "server": {
    "port": 8080
  },
  "grpc": {
    "producer_address": "localhost:50051",
    "consumer_address": "localhost:50052",
    "connection_timeout": "5s",
    "max_retries": 3,
    "retry_delay": "1s"
  }
}
```

## Building and Running

### Prerequisites
- Go 1.22 or later
- Protocol Buffers compiler (protoc) - optional, pre-generated files included

### Quick Start
```bash
# Install dependencies
make deps

# Build the application
make build

# Run the application
make run
```

### Development Workflow
```bash
# Format, vet, build, and test
make dev

# Or run individual commands
make fmt      # Format code
make vet      # Vet code
make build    # Build application
make test     # Run tests
```

### Manual Build
```bash
# Install dependencies
go mod download
go mod tidy

# Build
go build -o bin/api-gateway .

# Run
./bin/api-gateway
```

## Project Structure

```
api-gateway/
├── client/              # gRPC client implementations
│   └── coffee_client.go
├── config/              # Configuration management
│   └── config.go
├── proto/               # Protocol buffer definitions and generated code
│   ├── coffee_service.proto
│   └── coffee/
│       ├── coffee_service.pb.go
│       └── coffee_service_grpc.pb.go
├── server/              # HTTP server implementation
│   └── http_server.go
├── utils/               # Utility functions
│   └── uuid.go
├── main.go              # Application entry point
├── Makefile             # Build automation
└── README.md            # This file
```

## API Usage Examples

### Create Order
```bash
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "coffee_type": "Latte"
  }'
```

### Get Order
```bash
curl http://localhost:8080/order/{order_id}
```

### List Orders
```bash
curl http://localhost:8080/orders
```

### Cancel Order
```bash
curl -X POST http://localhost:8080/order/{order_id}/cancel
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Development

### Adding New Endpoints
1. Add the endpoint to the proto file
2. Regenerate protobuf code: `make proto`
3. Implement the handler in `server/http_server.go`
4. Register the route in `NewHTTPServer`

### Testing
```bash
# Run all tests
make test

# Test build without running
make test-build

# Manual testing
go run test_build.go
```

## Troubleshooting

### Common Issues

1. **Port already in use**
   - Change the `SERVER_PORT` environment variable
   - Kill the process using the port: `lsof -ti:8080 | xargs kill`

2. **gRPC connection failed**
   - Ensure Producer/Consumer services are running
   - Check the gRPC addresses in configuration
   - Verify network connectivity

3. **Build errors**
   - Run `go mod tidy` to clean dependencies
   - Ensure Go version is 1.22 or later
   - Check for syntax errors: `go vet ./...`

### Logs
The service logs all requests with unique request IDs for tracing:
```
[request-id] GET /health 127.0.0.1:12345
[request-id] Completed in 1.234ms
```

## Contributing

1. Follow Go coding standards
2. Add tests for new functionality
3. Update documentation
4. Run `make dev` before committing
