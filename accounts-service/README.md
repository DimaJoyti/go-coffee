# Accounts Service

The Accounts Service is a microservice that manages accounts, vendors, products, and orders for the Coffee Order System. It provides a GraphQL API for interacting with these entities and integrates with the existing Kafka-based system.

## Features

- **Account Management**: Create, read, update, and delete user accounts
- **Vendor Management**: Create, read, update, and delete vendors
- **Product Management**: Create, read, update, and delete products
- **Order Management**: Create, read, update, and delete orders
- **GraphQL API**: Provides a GraphQL API for interacting with the service
- **Kafka Integration**: Publishes events to Kafka when entities change
- **PostgreSQL Database**: Stores all data in a PostgreSQL database

## Architecture

The Accounts Service follows a clean architecture pattern with the following layers:

- **Models**: Define the domain entities
- **Repository**: Handles data access and persistence
- **Service**: Implements business logic
- **GraphQL**: Provides the API layer
- **Kafka**: Handles event publishing and consumption

## Technologies

- **Go**: Programming language
- **GraphQL**: API query language
- **PostgreSQL**: Database
- **Kafka**: Message broker
- **Docker**: Containerization
- **Kubernetes**: Orchestration

## Getting Started

### Prerequisites

- Go 1.22 or later
- PostgreSQL 16 or later
- Kafka
- Docker (optional)
- Kubernetes (optional)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/coffee-order-system.git
cd coffee-order-system/accounts-service
```

2. Install dependencies:

```bash
go mod download
```

3. Set up the database:

```bash
# Create the database
createdb coffee_accounts

# Run migrations
go run cmd/accounts-service/main.go
```

4. Start the service:

```bash
go run cmd/accounts-service/main.go
```

### Docker

You can also run the service using Docker:

```bash
docker-compose up accounts-service
```

## API Documentation

The Accounts Service provides a GraphQL API. See the [API documentation](docs/api.md) for details.

## Configuration

The service can be configured using environment variables or a configuration file. See the [configuration documentation](docs/configuration.md) for details.

## Kafka Integration

The service integrates with Kafka for event-driven communication with other services. See the [Kafka documentation](docs/kafka.md) for details.

## Deployment

The service can be deployed to various environments. See the [deployment documentation](docs/deployment.md) for details.

## Monitoring and Logging

The service includes built-in monitoring and logging. See the [monitoring documentation](docs/monitoring.md) for details.

## Testing

### Unit Tests

Run the unit tests:

```bash
go test ./...
```

### Integration Tests

Run the integration tests:

```bash
go test -tags=integration ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -am 'Add my feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
