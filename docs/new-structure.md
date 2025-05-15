# New Project Structure

This document describes the new structure of the Coffee Order System project.

## Overview

The project has been reorganized to improve code structure and reuse common components. The new structure follows the principles of Clean Architecture and separation of concerns.

## Shared Library (pkg)

The `pkg` directory contains shared code that can be used by all services:

- `pkg/models`: Shared data models
  - `order.go`: Order and ProcessedOrder models
- `pkg/kafka`: Shared code for Kafka integration
  - `producer.go`: Kafka producer interface and implementation
  - `consumer.go`: Kafka consumer interface and implementation
- `pkg/config`: Shared configuration code
  - `config.go`: Configuration loading and environment variable handling
- `pkg/logger`: Shared logging code
  - `logger.go`: Logger interface and implementation
- `pkg/errors`: Shared error handling code
  - `errors.go`: Custom error types and error handling utilities

## Service Structure

Each service (producer, consumer, streams, api-gateway) follows a standardized structure:

- `cmd/`: Entry points for services
  - `cmd/[service]/main.go`: Main entry point for the service
- `internal/`: Internal service code (not exported)
  - `internal/handler`: HTTP/gRPC handlers
  - `internal/service`: Business logic
  - `internal/repository`: Data access
- `config/`: Configuration files
  - `config.json`: Default configuration

### Producer Service

The Producer service has been restructured as follows:

```
producer/
├── cmd/
│   └── producer/
│       └── main.go       # Entry point
├── internal/
│   ├── handler/
│   │   ├── http_handler.go   # HTTP handlers
│   │   ├── grpc_handler.go   # gRPC handlers
│   │   └── middleware.go     # HTTP middleware
│   ├── service/
│   │   └── order_service.go  # Business logic
│   └── repository/
│       └── order_repository.go # Data access
├── config/
│   └── config.json       # Configuration file
├── go.mod               # Go module file
└── go.sum               # Go module checksum
```

### Consumer Service

The Consumer service has been restructured as follows:

```
consumer/
├── cmd/
│   └── consumer/
│       └── main.go       # Entry point
├── internal/
│   ├── handler/
│   │   └── message_handler.go # Kafka message handler
│   ├── service/
│   │   └── order_service.go   # Business logic
│   └── repository/
│       └── order_repository.go # Data access
├── config/
│   └── config.json       # Configuration file
├── go.mod               # Go module file
└── go.sum               # Go module checksum
```

### Streams Processor Service

The Streams Processor service has been restructured as follows:

```
streams/
├── cmd/
│   └── streams/
│       └── main.go       # Entry point
├── internal/
│   ├── processor/
│   │   └── stream_processor.go # Stream processing logic
│   └── service/
│       └── order_service.go    # Business logic
├── config/
│   └── config.json       # Configuration file
├── go.mod               # Go module file
└── go.sum               # Go module checksum
```

### API Gateway Service

The API Gateway service has been restructured as follows:

```
api-gateway/
├── cmd/
│   └── api-gateway/
│       └── main.go       # Entry point
├── internal/
│   ├── handler/
│   │   ├── http_handler.go   # HTTP handlers
│   │   └── middleware.go     # HTTP middleware
│   ├── service/
│   │   └── gateway_service.go # Business logic
│   └── client/
│       └── producer_client.go # gRPC client for Producer
├── config/
│   └── config.json       # Configuration file
├── go.mod               # Go module file
└── go.sum               # Go module checksum
```

## Benefits of the New Structure

The new structure provides several benefits:

1. **Code Reuse**: Common code is shared across services, reducing duplication.
2. **Separation of Concerns**: Each component has a clear responsibility.
3. **Testability**: The structure makes it easier to write unit tests.
4. **Maintainability**: The standardized structure makes it easier to understand and maintain the code.
5. **Scalability**: The structure makes it easier to add new features and services.

## Migration Guide

To migrate from the old structure to the new structure:

1. Create the new directory structure
2. Move the code to the appropriate directories
3. Update imports and dependencies
4. Update the build and run scripts

## Next Steps

The following steps are recommended to further improve the project:

1. Add unit tests for all components
2. Add integration tests for service interactions
3. Add documentation for all public APIs
4. Add CI/CD pipeline for automated testing and deployment
