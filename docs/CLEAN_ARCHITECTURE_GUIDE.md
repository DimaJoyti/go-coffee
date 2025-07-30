# Go Coffee - Clean Architecture Guide

This document outlines the new clean architecture structure for the Go Coffee project, following Go best practices and clean architecture principles.

## 🏗️ New Project Structure

```
go-coffee/
├── cmd/                          # Application entry points
│   ├── api-gateway/             # API Gateway service
│   │   └── main.go
│   ├── auth-service/            # Authentication service
│   │   └── main.go
│   ├── order-service/           # Order management
│   │   └── main.go
│   ├── payment-service/         # Payment processing (Bitcoin/Crypto)
│   │   └── main.go
│   ├── kitchen-service/         # Kitchen operations
│   │   └── main.go
│   ├── ai-service/              # AI/ML services
│   │   └── main.go
│   ├── notification-service/    # Notifications
│   │   └── main.go
│   └── analytics-service/       # Analytics and reporting
│       └── main.go
├── internal/                     # Private application code
│   ├── api-gateway/             # API Gateway implementation
│   ├── auth/                    # Authentication logic
│   ├── order/                   # Order management logic
│   ├── payment/                 # Payment processing logic
│   ├── kitchen/                 # Kitchen operations logic
│   ├── ai/                      # AI/ML logic
│   ├── notification/            # Notification logic
│   └── analytics/               # Analytics logic
├── pkg/                         # Public libraries (reusable)
│   ├── bitcoin/                 # Bitcoin implementation ✅
│   ├── blockchain/              # Blockchain utilities
│   ├── config/                  # Configuration management ✅
│   ├── database/                # Database utilities ✅
│   ├── logger/                  # Logging utilities ✅
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Shared data models ✅
│   ├── security/                # Security utilities
│   ├── kafka/                   # Kafka utilities ✅
│   ├── redis/                   # Redis utilities ✅
│   └── utils/                   # Common utilities
├── api/                         # API definitions
│   ├── proto/                   # gRPC definitions
│   │   ├── auth/
│   │   ├── order/
│   │   ├── payment/
│   │   └── kitchen/
│   ├── openapi/                 # OpenAPI/Swagger specs
│   │   ├── auth.yaml
│   │   ├── order.yaml
│   │   ├── payment.yaml
│   │   └── kitchen.yaml
│   └── graphql/                 # GraphQL schemas
│       └── schema.graphql
├── web/                         # Web UI
│   ├── frontend/                # React/Vue frontend
│   │   ├── src/
│   │   ├── public/
│   │   └── package.json
│   └── static/                  # Static assets
├── deployments/                 # Deployment configurations
│   ├── docker/                  # Docker files
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.dev.yml
│   │   └── Dockerfile.*
│   ├── kubernetes/              # K8s manifests
│   │   ├── namespace.yaml
│   │   ├── services/
│   │   └── ingress/
│   └── terraform/               # Infrastructure as code
│       ├── main.tf
│       └── modules/
├── configs/                     # Configuration files
│   ├── development/
│   │   ├── app.env
│   │   └── database.env
│   ├── production/
│   │   ├── app.env
│   │   └── database.env
│   └── testing/
│       ├── app.env
│       └── database.env
├── scripts/                     # Build and deployment scripts
│   ├── build.sh
│   ├── deploy.sh
│   ├── test.sh
│   └── migrate.sh
├── docs/                        # Documentation
│   ├── api/                     # API documentation
│   ├── architecture/            # Architecture docs
│   └── deployment/              # Deployment guides
├── test/                        # Integration tests
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── tools/                       # Development tools
│   ├── mockgen/
│   └── protoc/
├── go.mod                       # Go modules
├── go.sum
├── Makefile                     # Build automation
└── README.md                    # Project documentation
```

## 🎯 Architecture Principles

### 1. Clean Architecture Layers

```
┌─────────────────────────────────────┐
│           Frameworks & Drivers      │  ← cmd/, web/, external APIs
├─────────────────────────────────────┤
│           Interface Adapters        │  ← internal/*/handlers, gateways
├─────────────────────────────────────┤
│           Application Business      │  ← internal/*/services, use cases
├─────────────────────────────────────┤
│           Enterprise Business       │  ← pkg/models, domain entities
└─────────────────────────────────────┘
```

### 2. Dependency Direction
- Dependencies point inward (toward business logic)
- Inner layers don't know about outer layers
- Use interfaces for dependency inversion

### 3. Service Communication
- **Synchronous**: gRPC for service-to-service
- **Asynchronous**: Kafka for events
- **Caching**: Redis for performance
- **Storage**: PostgreSQL for persistence

## 🔄 Migration Plan

### 1: Core Infrastructure ✅
- [x] Create new directory structure
- [x] Move Bitcoin implementation to `pkg/bitcoin/`
- [x] Update import paths
- [x] Create shared models in `pkg/models/`

### 2: Payment Service (In Progress)
- [ ] Complete payment service implementation
- [ ] Add proper HTTP handlers (standard library)
- [ ] Integrate with existing config system
- [ ] Add comprehensive tests

### 3: Service Consolidation
- [ ] Merge crypto-wallet → payment-service
- [ ] Merge ai-agents + ai-arbitrage → ai-service
- [ ] Update auth-service structure
- [ ] Update order-service structure
- [ ] Update kitchen-service structure

### 4: API Standardization
- [ ] Create gRPC definitions in `api/proto/`
- [ ] Create OpenAPI specs in `api/openapi/`
- [ ] Implement API Gateway
- [ ] Add middleware for logging, auth, rate limiting

### 5: Deployment & DevOps
- [ ] Consolidate Docker configurations
- [ ] Create Kubernetes manifests
- [ ] Set up CI/CD pipelines
- [ ] Add monitoring and observability

## 📋 Service Responsibilities

### Payment Service
- Bitcoin wallet operations
- Transaction creation and signing
- Address validation
- Multisig support
- Message signing/verification

### Auth Service
- User authentication
- JWT token management
- Role-based access control
- Session management

### Order Service
- Order creation and management
- Order status tracking
- Order history
- Integration with payment and kitchen

### Kitchen Service
- Order queue management
- Preparation tracking
- Inventory management
- Staff coordination

### AI Service
- Order recommendations
- Demand forecasting
- Price optimization
- Customer behavior analysis

### Notification Service
- Email notifications
- SMS notifications
- Push notifications
- Event-driven messaging

### Analytics Service
- Business metrics
- Performance monitoring
- Reporting
- Data aggregation

## 🛠️ Development Guidelines

### Code Organization
- Use interfaces for all external dependencies
- Keep business logic in service layer
- Use dependency injection
- Follow Go naming conventions

### Testing Strategy
- Unit tests for business logic
- Integration tests for service interactions
- E2E tests for critical user flows
- Mock external dependencies

### Configuration Management
- Environment-specific configs in `configs/`
- Use environment variables for secrets
- Validate configuration on startup
- Support hot-reloading in development

### Error Handling
- Use structured errors with context
- Log errors with appropriate levels
- Return meaningful error messages
- Implement circuit breakers for external calls

### Security
- Encrypt sensitive data at rest
- Use HTTPS for all communications
- Implement rate limiting
- Validate all inputs
- Use secure defaults

## 🚀 Getting Started

### 1. Development Setup
```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Install dependencies
go mod download

# Set up environment
cp configs/development/app.env.example configs/development/app.env

# Run tests
make test

# Start services
make dev
```

### 2. Service Development
```bash
# Generate service template
make generate-service name=my-service

# Run specific service
make run-service service=payment-service

# Build service
make build-service service=payment-service
```

### 3. API Development
```bash
# Generate gRPC code
make generate-proto

# Generate OpenAPI docs
make generate-docs

# Validate API specs
make validate-api
```

## 📚 Additional Resources

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Microservices Patterns](https://microservices.io/patterns/)

---

This architecture provides a solid foundation for scaling the Go Coffee application while maintaining clean separation of concerns and following Go best practices.
