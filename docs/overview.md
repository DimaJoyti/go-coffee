# Coffee Order System Overview

The Coffee Order System is a simple application that demonstrates the use of Kafka for message queuing in a microservices architecture. The system allows users to place coffee orders through a REST API, which are then processed asynchronously by a worker service.

## System Components

The system consists of two main components:

1. **Producer Service**: An HTTP server that receives coffee orders from clients and publishes them to a Kafka topic.
2. **Consumer Service**: A worker service that consumes coffee orders from the Kafka topic and processes them.

## Key Features

- **RESTful API**: The producer service exposes a RESTful API for placing coffee orders.
- **Message Queuing**: Kafka is used for reliable message queuing between the producer and consumer services.
- **Middleware**: The producer service includes middleware for logging, request ID generation, CORS support, and error recovery.
- **Configuration Management**: Both services support configuration through environment variables and configuration files.
- **Error Handling**: Comprehensive error handling is implemented throughout the system.
- **Testing**: Unit tests are included for key components.

## Use Cases

The Coffee Order System can be used as:

1. **Learning Tool**: A simple example of how to use Kafka with Go.
2. **Starting Point**: A foundation for building more complex microservices applications.
3. **Reference Implementation**: A reference for implementing middleware, configuration management, and testing in Go applications.

## Technology Stack

- **Go**: The programming language used for both services.
- **Kafka**: The message broker used for communication between services.
- **Sarama**: A Go client library for Apache Kafka.

## Next Steps

- [Architecture](architecture.md): Learn about the system architecture.
- [Installation](installation.md): Install and run the system.
- [API Reference](api-reference.md): Explore the API endpoints.
