# Web3 Wallet Backend Documentation

## Overview

This documentation provides comprehensive information about the Web3 Wallet Backend system, a multi-region, high-performance platform designed to handle Web3 wallet operations with specialized services for shopper supply management, order processing, and order claiming.

## Documentation Structure

### 📋 System Design
- **[System Design](system-design.md)** - Comprehensive system architecture and design decisions
- **[Architecture Decisions](architecture-decisions.md)** - Detailed ADRs explaining key architectural choices
- **[Multi-Region Implementation](multi-region-implementation.md)** - Technical implementation details

### 🚀 Deployment & Operations
- **[Multi-Region Deployment Guide](multi-region-deployment-guide.md)** - Step-by-step deployment instructions
- **[Performance Tuning Guide](performance-tuning-guide.md)** - Optimization recommendations
- **[Implementation Summary](../IMPLEMENTATION-SUMMARY.md)** - High-level implementation overview

### 🔌 API Documentation
- **[API Design](api-design.md)** - REST and gRPC API specifications
- **[Coffee Crypto API](coffee-crypto-api.md)** - Coffee purchase with cryptocurrency APIs
- **[Coffee Purchase Flow](coffee-purchase-flow.md)** - Complete end-to-end purchase flow
- **[Protocol Buffers](../api/proto/)** - gRPC service definitions

### 🏗️ Infrastructure
- **[Terraform Modules](../terraform/)** - Infrastructure as Code
- **[Kubernetes Manifests](../kubernetes/)** - Container orchestration
- **[Docker Images](../build/)** - Container definitions

## Quick Start

### Prerequisites

- Google Cloud Platform account
- Terraform 1.0+
- kubectl
- Helm 3+
- Docker

### Basic Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/web3-wallet-backend.git
   cd web3-wallet-backend
   ```

2. **Configure environment**
   ```bash
   cp terraform/multi-region/terraform.tfvars.example terraform/multi-region/terraform.tfvars
   # Edit terraform.tfvars with your configuration
   ```

3. **Deploy infrastructure**
   ```bash
   cd terraform/multi-region
   terraform init
   terraform apply
   ```

4. **Deploy services**
   ```bash
   kubectl apply -f kubernetes/manifests/
   ```

## System Architecture

### High-Level Overview

The system is built as a distributed, multi-region architecture with the following key components:

```text
┌─────────────────────────────────────────────────────────────────┐
│                        Global Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  Global Load Balancer │ CDN │ DNS │ WAF │ DDoS Protection      │
└─────────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │   Region A   │ │   Region B  │ │  Region C  │
        │ (us-central) │ │(europe-west)│ │(asia-east) │
        └──────────────┘ └─────────────┘ └────────────┘
```

### Core Services

1. **Coffee Shop Service** - Manages coffee shop locations and information
2. **Product Catalog Service** - Handles coffee menu items and pricing
3. **Supply Service** - Manages coffee inventory and supply data
4. **Order Service** - Processes coffee orders and customizations
5. **Payment Service** - Handles cryptocurrency payments (BTC, ETH, USDC, USDT)
6. **Claiming Service** - Manages order pickup and delivery coordination
7. **Wallet Service** - Web3 wallet functionality and blockchain integration
8. **Notification Service** - Real-time order status and payment notifications
9. **Price Service** - Real-time cryptocurrency price feeds and conversion

### Technology Stack

- **Container Orchestration**: Kubernetes (GKE)
- **Service Communication**: gRPC
- **Event Streaming**: Apache Kafka
- **Caching**: Redis Cluster
- **Database**: PostgreSQL
- **Infrastructure**: Terraform
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack

## API Documentation

The [API Documentation](api-documentation.md) provides detailed information about the Web3 Wallet Backend API endpoints, request/response formats, and usage examples. It covers all the services provided by the system, including:

- Wallet API
- Transaction API
- Smart Contract API
- Security API

## Developer Guide

The [Developer Guide](developer-guide.md) provides detailed information for developers who want to contribute to or extend the Web3 Wallet Backend system. It covers:

- Architecture Overview
- Development Environment Setup
- Project Structure
- Core Components
- Adding a New Feature
- Testing
- Deployment
- Best Practices

## Architecture

The [Architecture](architecture.md) document provides a detailed overview of the system architecture, including:

- System Components
- Communication Patterns
- Data Flow
- Security Model
- Scalability Considerations

## Deployment Guide

The [Deployment Guide](deployment-guide.md) provides detailed instructions for deploying the Web3 Wallet Backend system in various environments, including:

- Development Environment
- Testing Environment
- Production Environment
- Docker Deployment
- Kubernetes Deployment
- Cloud Provider Deployment (AWS, GCP, Azure)

## Security

The [Security](security.md) document provides detailed information about the security aspects of the Web3 Wallet Backend system, including:

- Authentication and Authorization
- Key Management
- Encryption
- Secure Communication
- Audit Logging
- Compliance Considerations

## API Examples

### Coffee Shop Discovery

```bash
# Find nearby coffee shops
curl "https://api.cryptocoffee.com/api/v1/coffee-shops?lat=40.7128&lng=-74.0060&radius=5000" \
  -H "Authorization: Bearer $TOKEN"

# Get shop menu
curl https://api.cryptocoffee.com/api/v1/coffee-shops/{shop_id}/menu \
  -H "Authorization: Bearer $TOKEN"
```

### Coffee Order with Crypto Payment

```bash
# Create coffee order
curl -X POST https://api.cryptocoffee.com/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shop_id": "uuid",
    "order_type": "pickup",
    "pickup_time": "2024-01-01T15:30:00Z",
    "payment_currency": "ETH",
    "items": [
      {
        "product_id": "uuid",
        "quantity": 2,
        "customizations": {
          "size": "Large",
          "milk": "Oat Milk"
        }
      }
    ]
  }'

# Check payment status
curl https://api.cryptocoffee.com/api/v1/payments/{payment_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Order Pickup

```bash
# Generate pickup code
curl -X POST https://api.cryptocoffee.com/api/v1/orders/{order_id}/claim \
  -H "Authorization: Bearer $TOKEN"

# Verify pickup code (shop endpoint)
curl -X POST https://api.cryptocoffee.com/api/v1/claims/verify \
  -H "Authorization: Bearer $SHOP_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"claim_code": "ABC123", "shop_id": "uuid"}'
```

### Cryptocurrency Prices

```bash
# Get current crypto prices
curl "https://api.cryptocoffee.com/api/v1/crypto/prices?currencies=BTC,ETH,USDC,USDT" \
  -H "Authorization: Bearer $TOKEN"
```

## Development

### Local Development

1. **Install dependencies**
   ```bash
   go mod tidy
   ```

2. **Start local services**
   ```bash
   docker-compose up -d
   ```

3. **Run services**
   ```bash
   go run cmd/coffee-shop-service/main.go
   go run cmd/order-service/main.go
   go run cmd/payment-service/main.go
   go run cmd/claiming-service/main.go
   ```

### Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run load tests
k6 run tests/load/coffee-order-flow.js
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines

- Follow Go best practices
- Write comprehensive tests
- Update documentation
- Use conventional commits

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
