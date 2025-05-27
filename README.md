# Coffee Order System with Kafka and Go

A simple coffee ordering system using Kafka for message queuing, implemented in Go.

## New Project Structure

The project has been reorganized to improve code structure and reuse common components:

### Shared Library (pkg)

- `pkg/models`: Shared data models
- `pkg/kafka`: Shared code for Kafka integration
- `pkg/config`: Shared configuration code
- `pkg/logger`: Shared logging code
- `pkg/errors`: Shared error handling code

### Services

Each service has a standardized structure:

- `cmd/`: Entry points for services
- `internal/`: Internal service code
  - `internal/handler`: HTTP/gRPC handlers
  - `internal/service`: Business logic
  - `internal/repository`: Data access
- `config/`: Configuration files

## System Components

The system consists of four main components:

1. **API Gateway**: gRPC gateway that provides a unified API for clients and communicates with the Producer service
2. **Producer**: Service that receives coffee orders and sends them to Kafka (supports both HTTP and gRPC)
3. **Streams Processor**: Service that processes coffee orders using Kafka Streams
4. **Consumer**: Service that consumes processed coffee orders from Kafka and executes them

## Features

- RESTful API for placing coffee orders
- gRPC API for internal service communication
- API Gateway for unified client access
- Kafka integration for message queuing
- Kafka Streams for event processing
- Middleware for HTTP server:
  - Logging middleware
  - Request ID middleware
  - CORS middleware
  - Error recovery middleware
- Configuration management via environment variables and config files
- Monitoring and alerting with Prometheus and Grafana
- Containerization with Docker and orchestration with Kubernetes

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Kafka server running on localhost:9092

### Running the Producer

```bash
cd producer
go mod tidy
go run cmd/producer/main.go
```

The producer will start an HTTP server on port 3000.

### Running the Streams Processor

```bash
cd streams
go mod tidy
go run cmd/streams/main.go
```

The streams processor will start processing messages from the "coffee_orders" Kafka topic and sending them to the "processed_orders" topic.

### Running the Consumer

```bash
cd consumer
go mod tidy
go run cmd/consumer/main.go
```

The consumer will start listening for messages on both the "coffee_orders" and "processed_orders" Kafka topics.

### Running the API Gateway

```bash
cd api-gateway
go mod tidy
go run cmd/api-gateway/main.go
```

The API Gateway will start an HTTP server on port 8080 and connect to the Producer service via gRPC.

### Running All Components

You can use the provided scripts to run all components at once:

```bash
# On Linux/macOS
./run.sh

# On Windows
run.bat
```

## API Endpoints

### Place an Order

```http
POST /order
```

Request body:

```json
{
  "customer_name": "John Doe",
  "coffee_type": "Latte"
}
```

Response:

```json
{
  "success": true,
  "msg": "Order for John Doe placed successfully!"
}
```

### Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

## Configuration

Both the producer and consumer can be configured using environment variables or a config file.

### API Gateway Configuration

Environment variables:

- `SERVER_PORT`: HTTP server port (default: 8080)
- `PRODUCER_GRPC_ADDRESS`: gRPC address of the Producer service (default: "localhost:50051")
- `CONSUMER_GRPC_ADDRESS`: gRPC address of the Consumer service (default: "localhost:50052")
- `CONFIG_FILE`: Path to config file (optional)

### Producer Configuration

Environment variables:

- `SERVER_PORT`: HTTP server port (default: 3000)
- `KAFKA_BROKERS`: Kafka broker addresses (default: ["localhost:9092"])
- `KAFKA_TOPIC`: Kafka topic for orders (default: "coffee_orders")
- `KAFKA_RETRY_MAX`: Maximum number of retries for Kafka producer (default: 5)
- `KAFKA_REQUIRED_ACKS`: Required acknowledgments for Kafka producer (default: "all")
- `CONFIG_FILE`: Path to config file (optional)

### Streams Processor Configuration

Environment variables:

- `KAFKA_BROKERS`: Kafka broker addresses (default: ["localhost:9092"])
- `KAFKA_INPUT_TOPIC`: Kafka topic for input orders (default: "coffee_orders")
- `KAFKA_OUTPUT_TOPIC`: Kafka topic for processed orders (default: "processed_orders")
- `KAFKA_APPLICATION_ID`: Kafka Streams application ID (default: "coffee-streams-app")
- `KAFKA_AUTO_OFFSET_RESET`: Auto offset reset configuration (default: "earliest")
- `KAFKA_PROCESSING_GUARANTEE`: Processing guarantee (default: "at_least_once")
- `CONFIG_FILE`: Path to config file (optional)

### Consumer Configuration

Environment variables:

- `KAFKA_BROKERS`: Kafka broker addresses (default: ["localhost:9092"])
- `KAFKA_TOPIC`: Kafka topic for orders (default: "coffee_orders")
- `KAFKA_PROCESSED_TOPIC`: Kafka topic for processed orders (default: "processed_orders")
- `CONFIG_FILE`: Path to config file (optional)

## Documentation

For more detailed information, please refer to the following documentation:

- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Development Guide](docs/development-guide.md)
- [Installation](docs/installation.md)
- [Kafka Streams](docs/kafka-streams.md)
- [Monitoring](docs/monitoring.md)
- [Docker and Kubernetes](docs/docker-kubernetes.md)
- [FAQ](docs/faq.md)
- [Roadmap](docs/roadmap.md)
