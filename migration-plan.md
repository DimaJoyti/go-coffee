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
- [ ] Convert Gin handlers to standard HTTP
- [ ] Add proper error handling
- [ ] Add comprehensive tests
- [ ] Update Docker configuration

### 2. Auth Service
- [ ] Move to internal/auth/
- [ ] Update import paths
- [ ] Standardize configuration
- [ ] Add gRPC interface

### 3. Order Service
- [ ] Move to internal/order/
- [ ] Update import paths
- [ ] Add event publishing
- [ ] Integrate with payment service

### 4. Kitchen Service
- [ ] Move to internal/kitchen/
- [ ] Update import paths
- [ ] Add real-time updates
- [ ] Integrate with order service

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
