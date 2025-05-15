# Coffee Order System Architecture

This document describes the architecture of the Coffee Order System, including its components, interactions, and design patterns.

## System Architecture

The Coffee Order System follows a microservices architecture with two main services:

1. **Producer Service**: Handles HTTP requests and publishes messages to Kafka.
2. **Consumer Service**: Consumes messages from Kafka and processes them.

These services communicate asynchronously through Kafka, which acts as a message broker.

![Architecture Diagram](images/architecture.png)

## Producer Service

The Producer Service is an HTTP server that exposes a REST API for placing coffee orders. It is structured as follows:

### Components

- **Main**: The entry point of the application, responsible for initializing components and starting the HTTP server.
- **Config**: Manages application configuration from environment variables and configuration files.
- **Handler**: Contains HTTP handlers for processing requests.
- **Kafka**: Provides an abstraction for interacting with Kafka.
- **Middleware**: Contains HTTP middleware for logging, request ID generation, CORS support, and error recovery.

### Request Flow

1. An HTTP request is received by the server.
2. The request passes through a chain of middleware:
   - **Recover Middleware**: Catches panics and returns a 500 error.
   - **Logging Middleware**: Logs request details.
   - **Request ID Middleware**: Assigns a unique ID to the request.
   - **CORS Middleware**: Adds CORS headers to the response.
3. The request is handled by the appropriate handler.
4. The handler processes the request and publishes a message to Kafka.
5. The handler returns a response to the client.

## Consumer Service

The Consumer Service is a worker service that consumes messages from Kafka and processes them. It is structured as follows:

### Components

- **Main**: The entry point of the application, responsible for initializing components and starting the consumer.
- **Config**: Manages application configuration from environment variables and configuration files.
- **Kafka**: Provides an abstraction for interacting with Kafka.

### Processing Flow

1. The consumer connects to Kafka and subscribes to a topic.
2. When a message is received, it is processed by the consumer.
3. The consumer logs the processing of the message.

## Kafka Integration

Kafka is used as a message broker between the Producer and Consumer services. It provides:

- **Decoupling**: The Producer and Consumer services are decoupled, allowing them to operate independently.
- **Reliability**: Messages are persisted in Kafka, ensuring they are not lost if the Consumer service is down.
- **Scalability**: Multiple Consumer instances can be deployed to process messages in parallel.

## Design Patterns

The Coffee Order System uses several design patterns:

- **Dependency Injection**: Components are injected into other components, making the code more testable and maintainable.
- **Middleware Chain**: HTTP middleware is chained together to process requests.
- **Repository Pattern**: The Kafka package provides an abstraction for interacting with Kafka.
- **Configuration Management**: The Config package provides a centralized way to manage application configuration.

## Next Steps

- [Installation](installation.md): Install and run the system.
- [Configuration](configuration.md): Configure the system.
- [API Reference](api-reference.md): Explore the API endpoints.
