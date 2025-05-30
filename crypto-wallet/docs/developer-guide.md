# Web3 Wallet Backend Developer Guide

This guide provides detailed information for developers who want to contribute to or extend the Web3 Wallet Backend system.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Structure](#project-structure)
4. [Core Components](#core-components)
5. [Adding a New Feature](#adding-a-new-feature)
6. [Testing](#testing)
7. [Deployment](#deployment)
8. [Best Practices](#best-practices)

## Architecture Overview

The Web3 Wallet Backend is built using a microservices architecture with the following components:

1. **API Gateway**: Entry point for all client requests, handles authentication, rate limiting, and request routing.
2. **Wallet Service**: Manages wallet operations such as creation, import/export, and balance checking.
3. **Transaction Service**: Handles blockchain transactions, gas estimation, and transaction status tracking.
4. **Smart Contract Service**: Manages smart contract interactions, deployment, and event monitoring.
5. **Security Service**: Provides cryptographic operations, key management, and JWT token handling.

Each service is designed to be independently deployable and scalable. Services communicate with each other using gRPC for internal communication, while the API Gateway exposes a RESTful API for external clients.

## Development Environment Setup

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 16 or higher
- Redis 7 or higher
- Docker and Docker Compose (optional, for containerized development)
- Kubernetes (optional, for containerized deployment)
- Protobuf compiler and Go plugins (for gRPC development)

### Setting Up the Development Environment

1. Clone the repository:

```bash
git clone https://github.com/yourusername/web3-wallet-backend.git
cd web3-wallet-backend
```

2. Install Go dependencies:

```bash
go mod tidy
```

3. Install Protobuf compiler and Go plugins:

```bash
# Install protoc
# For macOS
brew install protobuf

# For Ubuntu/Debian
apt-get install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4. Generate gRPC code:

```bash
cd api
./generate.sh  # or generate.bat on Windows
cd ..
```

5. Set up the database:

```bash
# Create the database
createdb web3_wallet

# Run migrations
go run db/migrate.go -up -config config/config.yaml
```

6. Start the services:

```bash
# Using the provided script
./run.sh  # or run.bat on Windows

# Or start each service individually
go run cmd/api-gateway/main.go
go run cmd/wallet-service/main.go
go run cmd/transaction-service/main.go
go run cmd/smart-contract-service/main.go
go run cmd/security-service/main.go
```

7. Using Docker Compose (optional):

```bash
docker-compose up -d
```

## Project Structure

```
web3-wallet-backend/
├── api/                    # API definitions
│   ├── proto/              # Protocol buffer definitions
│   ├── generate.sh         # Script to generate gRPC code
│   └── generate.bat        # Script to generate gRPC code (Windows)
├── build/                  # Dockerfiles for each service
├── cmd/                    # Entry points for services
│   ├── api-gateway/        # API Gateway service
│   ├── wallet-service/     # Wallet service
│   ├── transaction-service/ # Transaction service
│   ├── smart-contract-service/ # Smart Contract service
│   └── security-service/   # Security service
├── config/                 # Configuration files
├── db/                     # Database migrations
│   ├── migrations/         # Migration files
│   └── migrate.go          # Migration tool
├── docs/                   # Documentation
├── internal/               # Internal packages
│   ├── wallet/             # Wallet implementation
│   ├── transaction/        # Transaction implementation
│   ├── smartcontract/      # Smart Contract implementation
│   ├── security/           # Security implementation
│   └── common/             # Shared code
├── kubernetes/             # Kubernetes manifests
│   ├── manifests/          # Kubernetes resource definitions
│   └── helm/               # Helm charts (optional)
├── pkg/                    # Shared packages
│   ├── blockchain/         # Blockchain clients
│   ├── crypto/             # Cryptographic utilities
│   ├── logger/             # Logging utilities
│   ├── config/             # Configuration utilities
│   └── models/             # Data models
├── docker-compose.yml      # Docker Compose configuration
├── go.mod                  # Go module file
├── go.sum                  # Go module checksum
├── README.md               # Project README
├── run.sh                  # Script to run all services (Unix)
└── run.bat                 # Script to run all services (Windows)
```

## Core Components

### Configuration

The system uses a YAML-based configuration system with environment variable overrides. The main configuration file is located at `config/config.yaml`.

```go
// Load configuration
cfg, err := config.LoadConfig("config/config.yaml")
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### Logging

The system uses a structured logging system based on zap. Logs can be output in JSON format for production or console format for development.

```go
// Initialize logger
logger := logger.NewLogger(cfg.Logging)
logger.Info("Starting service")
```

### Database Access

The system uses sqlx for database access, with repository interfaces for each service.

```go
// Create database connection
db, err := sqlx.Connect("postgres", dbURL)
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Create repository
repo := wallet.NewPostgresRepository(db, logger, cfg.Keystore.Path)
```

### Blockchain Clients

The system includes clients for various blockchains (Ethereum, Binance Smart Chain, Polygon) based on go-ethereum.

```go
// Create Ethereum client
ethClient, err := blockchain.NewEthereumClient(cfg.Blockchain.Ethereum, logger)
if err != nil {
    log.Fatalf("Failed to create Ethereum client: %v", err)
}
```

### Cryptographic Utilities

The system includes utilities for key management, encryption, and signing.

```go
// Create key manager
keyManager := crypto.NewKeyManager(cfg.Keystore.Path)

// Generate key pair
privateKey, publicKey, address, err := keyManager.GenerateKeyPair()
```

### Service Layer

Each service has a service layer that implements the business logic.

```go
// Create wallet service
walletService := wallet.NewService(
    repo,
    ethClient,
    bscClient,
    polygonClient,
    keyManager,
    logger,
    cfg.Keystore.Path,
)
```

### gRPC Handlers

Each service has gRPC handlers that implement the gRPC service interfaces.

```go
// Create gRPC handler
handler := wallet.NewGRPCHandler(walletService, logger)

// Register handler with gRPC server
pb.RegisterWalletServiceServer(grpcServer, handler)
```

## Adding a New Feature

To add a new feature to the system, follow these steps:

1. **Define the API**: Add the new endpoint to the appropriate proto file in `api/proto/`.

2. **Generate gRPC Code**: Run the `generate.sh` or `generate.bat` script to generate the Go code for the new endpoint.

3. **Implement the Model**: Add the new data model to the appropriate file in `pkg/models/`.

4. **Implement the Repository**: Add the new repository methods to the appropriate repository interface and implementation.

5. **Implement the Service**: Add the new service methods to the appropriate service.

6. **Implement the Handler**: Add the new handler methods to the appropriate gRPC handler.

7. **Update the API Gateway**: Add the new endpoint to the API Gateway if it's a public API.

8. **Write Tests**: Add tests for the new feature.

9. **Update Documentation**: Update the API documentation and developer guide.

## Testing

The system includes unit tests, integration tests, and end-to-end tests.

### Running Unit Tests

```bash
go test ./...
```

### Running Integration Tests

```bash
go test -tags=integration ./...
```

### Running End-to-End Tests

```bash
go test -tags=e2e ./...
```

## Deployment

The system can be deployed using Docker and Kubernetes.

### Docker Deployment

```bash
# Build the Docker images
docker-compose build

# Start the services
docker-compose up -d
```

### Kubernetes Deployment

```bash
# Apply the Kubernetes manifests
kubectl apply -f kubernetes/manifests/

# Or using Helm
helm install web3-wallet ./kubernetes/helm/web3-wallet
```

## Best Practices

### Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for code style.
- Use `gofmt` or `goimports` to format your code.
- Use meaningful variable and function names.
- Add comments to explain complex logic.

### Error Handling

- Always check errors and return them to the caller.
- Use custom error types for domain-specific errors.
- Log errors with context.

### Logging

- Use structured logging with fields.
- Log at the appropriate level (debug, info, warn, error).
- Include relevant context in log messages.

### Security

- Never store private keys in plaintext.
- Use secure random number generators for cryptographic operations.
- Validate all user input.
- Use prepared statements for database queries.
- Use HTTPS for all external communication.

### Performance

- Use connection pooling for database connections.
- Use caching for frequently accessed data.
- Use asynchronous processing for long-running tasks.
- Monitor and optimize database queries.

### Monitoring

- Implement health checks for all services.
- Expose metrics for monitoring.
- Set up alerts for critical errors.
- Monitor resource usage (CPU, memory, disk, network).
