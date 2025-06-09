# Migration Plan

## Services to Migrate

### High Priority (Core Business Logic)

- [ ] payment-service (crypto-wallet + crypto-terminal)
- [ ] auth-service (existing)
- [ ] order-service (existing)
- [ ] kitchen-service (existing)

### Medium Priority (AI/ML Features)

- [ ] ai-service (ai-agents + ai-arbitrage)
- [ ] notification-service (new)
- [ ] analytics-service (new)

### Low Priority (Infrastructure)

- [ ] api-gateway (new)

## Migration Steps per Service

### 1. Payment Service

- [x] Move Bitcoin implementation to pkg/bitcoin/
- [x] Create internal/payment/ structure
- [x] Create cmd/payment-service/
- [x] Convert Gin handlers to standard HTTP
- [x] Add proper error handling
- [x] Add comprehensive tests
- [x] Update Docker configuration

### 2. Order Service

- [x] Move to internal/order/
- [x] Create transport/http layer
- [x] Convert Gin handlers to standard HTTP
- [x] Add proper error handling
- [x] Add comprehensive tests
- [x] Add middleware (logging, recovery, CORS, auth)
- [x] Update Docker configuration

### 3. DeFi Service

- [x] Move to internal/defi/
- [x] Create transport/http layer
- [x] Convert Gin handlers to standard HTTP
- [x] Add proper error handling
- [x] Add comprehensive tests
- [x] Add middleware (logging, recovery, CORS, auth, security, metrics)
- [ ] Update Docker configuration

### 5. Auth Service

- [x] Move to internal/auth/ (Clean Architecture Foundation)
  - [x] Basic domain layer (user.go, session.go, token.go, repository.go)
  - [x] Basic application layer (service.go, dto.go, interfaces.go)
  - [x] Basic infrastructure layer (repository, security)
  - [x] Basic transport layer (gRPC)
- [x] Update import paths
- [x] Standardize configuration
- [x] Add gRPC interface
- [ ] Complete Clean Architecture Implementation
  - [ ] Enhanced domain layer with events and business rules
  - [ ] Complete application layer with all use cases
  - [ ] Full infrastructure layer (Redis, security, events)
  - [ ] Modern transport layer (HTTP, gRPC, WebSocket)
  - [ ] Authentication middleware and security
  - [ ] Event-driven architecture integration
- [ ] Replace Gin with clean HTTP handlers
- [ ] Add comprehensive testing
- [ ] Add real-time session management
- [ ] Integrate with other services

### 4. Kitchen Service

- [x] Move to internal/kitchen/ (Clean Architecture)
  - [x] Create domain layer (equipment.go, staff.go, order.go, queue.go, events.go, repository.go)
  - [x] Create application layer (interfaces.go, dto.go, kitchen_service.go, queue_service.go)
  - [x] Implement core business logic with proper domain entities
  - [x] Add comprehensive event system
  - [x] Define repository interfaces
- [x] Create infrastructure layer (Redis implementations)
  - [x] Redis repository implementations (equipment, staff, order, queue, workflow, metrics)
  - [x] Repository manager with transaction support
  - [x] AI optimizer service implementation
  - [x] Queue service with optimization and rebalancing
- [x] Create transport layer (gRPC, WebSocket, HTTP)
  - [x] gRPC server with full API implementation
  - [x] WebSocket server for real-time updates
  - [x] HTTP REST API handlers
  - [x] Authentication and authorization middleware
  - [x] Protocol buffer converters
  - [x] Unified transport server with graceful shutdown
- [x] Update import paths across codebase
  - [x] Updated main.go with clean architecture imports
  - [x] Created configuration management system
  - [x] Added order service integration client
  - [x] Implemented event integration service
  - [x] Updated dependency injection system
- [x] Add real-time updates integration
  - [x] Event-driven architecture with Redis pub/sub
  - [x] Cross-service event integration
  - [x] WebSocket real-time broadcasting
- [x] Integrate with order service
  - [x] Order service client implementation
  - [x] Event-based order synchronization
  - [x] Status update propagation

## Configuration Migration

- [ ] Consolidate environment files
- [ ] Update Docker Compose
- [ ] Create Kubernetes manifests
- [ ] Set up monitoring

## Testing Strategy

- [ ] Unit tests for each service
- [ ] Integration tests
- [ ] E2E tests
- [ ] Performance tests

## Deployment Strategy

- [ ] Blue-green deployment
- [ ] Database migrations
- [ ] Service discovery
- [ ] Load balancing
