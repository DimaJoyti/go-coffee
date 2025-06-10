# 🔐 Auth Service - Clean Architecture Implementation

A comprehensive authentication and authorization service built with **Clean Architecture** principles, featuring modern security practices, real-time capabilities, and event-driven design.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    TRANSPORT LAYER                          │
├─────────────────┬─────────────────┬─────────────────────────┤
│   HTTP/REST     │     gRPC        │      WebSocket          │
│ Clean Handlers  │   Enhanced      │   Real-time Hub         │
│ CQRS-based      │   Foundation    │   Event Notifications   │
└─────────────────┴─────────────────┴─────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                  MIDDLEWARE LAYER                           │
├─────────────────┬─────────────────┬─────────────────────────┤
│ Authentication  │   Authorization │     Security            │
│ JWT Validation  │   Role-based    │   Rate Limiting         │
│ Session Mgmt    │   Permissions   │   Headers & CORS        │
└─────────────────┴─────────────────┴─────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                 APPLICATION LAYER                           │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Command Bus   │    Query Bus    │    Event Handlers       │
│ 20+ Commands    │  15+ Queries    │   Domain Events         │
│ CQRS Pattern    │  Pagination     │   Cross-cutting         │
└─────────────────┴─────────────────┴─────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                   DOMAIN LAYER                              │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Aggregates    │ Business Rules  │      Events             │
│ User Aggregate  │ Password Rules  │  40+ Event Types        │
│ Session Entity  │ Security Rules  │  Event Sourcing         │
│ Token Entity    │ Validation      │  AggregateRoot          │
└─────────────────┴─────────────────┴─────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│               INFRASTRUCTURE LAYER                          │
├─────────────────┬─────────────────┬─────────────────────────┤
│  Event Store    │  Session Cache  │    Security Services    │
│ Redis-based     │ Token Mapping   │   Rate Limiter          │
│ Concurrency     │ User Sessions   │   Password Service      │
│ Event Pub/Sub   │ Auto Cleanup    │   JWT Service           │
└─────────────────┴─────────────────┴─────────────────────────┘
```

## 🚀 Features

### 🔐 Authentication & Authorization
- **JWT-based authentication** with access and refresh tokens
- **Role-based access control** (User, Moderator, Admin)
- **Permission-based authorization** for fine-grained control
- **Multi-factor authentication** (TOTP, SMS, Email)
- **Session management** with Redis-backed storage
- **Device fingerprinting** and trusted device management

### 🛡️ Security Features
- **Advanced rate limiting** (sliding window + token bucket algorithms)
- **Account lockout policies** with configurable thresholds
- **Password complexity validation** with business rules
- **Security event tracking** and audit logging
- **Risk score management** for adaptive security
- **CSRF, XSS, and injection protection**

### ⚡ Real-time Features
- **WebSocket connections** for live updates
- **Real-time auth event notifications**
- **Session activity monitoring**
- **Security alerts** and notifications
- **User presence tracking**

### 📊 Event-Driven Architecture
- **40+ domain events** for comprehensive tracking
- **Event sourcing capabilities** with Redis store
- **Event publishing** with pub/sub pattern
- **Cross-service event integration**
- **Optimistic concurrency control**

## 📁 Project Structure

```
internal/auth/
├── domain/                    # Domain Layer
│   ├── user.go               # User aggregate root
│   ├── session.go            # Session entity
│   ├── token.go              # Token entity
│   ├── events.go             # 40+ domain events
│   ├── rules.go              # Business rules engine
│   ├── aggregate.go          # User aggregate with logic
│   └── repository.go         # Repository interfaces
├── application/               # Application Layer
│   ├── commands/             # Command definitions (CQRS)
│   ├── queries/              # Query definitions (CQRS)
│   ├── handlers/             # Command/Query handlers
│   ├── bus/                  # Command/Query buses
│   ├── service.go            # Application services
│   ├── dto.go                # Data transfer objects
│   └── interfaces.go         # Service interfaces
├── infrastructure/           # Infrastructure Layer
│   ├── events/               # Event store & publisher
│   ├── cache/                # Redis session cache
│   ├── security/             # Security services
│   ├── middleware/           # Security middleware
│   ├── repository/           # Data persistence
│   └── container/            # Dependency injection
└── transport/                # Transport Layer
    ├── http/                 # REST API handlers
    ├── websocket/            # Real-time WebSocket
    └── grpc/                 # gRPC services
```

## 🔧 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Redis 6+

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Set up environment variables:**
```bash
export DATABASE_URL="postgres://user:pass@localhost/go_coffee_auth"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="your-super-secret-jwt-key"
```

3. **Run database migrations:**
```bash
go run cmd/migrate/main.go up
```

4. **Start the service:**
```bash
go run cmd/auth/main.go
```

### Testing

Run the comprehensive test suite:
```bash
go run cmd/auth-test/main.go
```

Or run unit tests:
```bash
go test ./internal/auth/...
```

## 📚 API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "role": "user"
}
```

#### Login User
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### Validate Token
```http
POST /api/v1/auth/validate
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Get User Info
```http
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

### WebSocket Events

Connect to WebSocket endpoint:
```
ws://localhost:8080/ws/auth
```

Subscribe to events:
```json
{
  "type": "subscribe",
  "data": {
    "channel": "user_events"
  }
}
```

## 🔒 Security Configuration

### JWT Configuration
```yaml
jwt:
  secret: "your-super-secret-key"
  access_token_ttl: "15m"
  refresh_token_ttl: "7d"
  issuer: "go-coffee-auth"
```

### Rate Limiting
```yaml
rate_limiting:
  login_attempts: 5
  window: "15m"
  global_requests: 1000
  per_user_requests: 100
```

### Password Policy
```yaml
password_policy:
  min_length: 8
  require_uppercase: true
  require_lowercase: true
  require_digits: true
  require_special: true
  max_age_days: 90
```

## 🧪 Testing

### Unit Tests
```bash
# Run all unit tests
go test ./internal/auth/domain/...
go test ./internal/auth/application/...
go test ./internal/auth/infrastructure/...

# Run with coverage
go test -cover ./internal/auth/...
```

### Integration Tests
```bash
# Run integration tests
go test ./internal/auth/integration_test.go

# Run with test runner
go run cmd/auth-test/main.go
```

### Load Testing
```bash
# Install k6
brew install k6

# Run load tests
k6 run scripts/load-test.js
```

## 📊 Monitoring & Observability

### Metrics
- Request/response times
- Authentication success/failure rates
- Active sessions count
- Rate limiting violations
- Security events frequency

### Health Checks
```http
GET /health
```

### Logs
Structured JSON logging with configurable levels:
- Authentication events
- Security violations
- Performance metrics
- Error tracking

## 🚀 Deployment

### Docker
```bash
# Build image
docker build -t go-coffee-auth .

# Run container
docker run -p 8080:8080 go-coffee-auth
```

### Kubernetes
```bash
# Apply manifests
kubectl apply -f k8s/
```

### Environment Variables
```bash
# Required
DATABASE_URL=postgres://user:pass@localhost/db
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key

# Optional
LOG_LEVEL=info
PORT=8080
ENVIRONMENT=production
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Clean Architecture by Robert C. Martin
- Domain-Driven Design by Eric Evans
- Event Sourcing patterns
- CQRS implementation patterns
