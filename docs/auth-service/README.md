# 🔐 Auth Service - Secure Authentication Microservice

<div align="center">

![Go Coffee Auth Service](https://img.shields.io/badge/Go%20Coffee-Auth%20Service-blue?style=for-the-badge&logo=go)
![Version](https://img.shields.io/badge/version-1.0.0-green?style=for-the-badge)
![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)

**Enterprise-grade authentication microservice with JWT tokens, refresh tokens, and comprehensive security features**

[🚀 Quick Start](#-quick-start) • [📖 Documentation](#-documentation) • [🔧 API Reference](#-api-reference) • [🛡️ Security](#️-security) • [📊 Monitoring](#-monitoring)

</div>

---

## 🌟 Features

<table>
<tr>
<td width="50%">

### 🔑 **Authentication & Authorization**
- ✅ JWT access tokens (15min TTL)
- ✅ Refresh tokens (7 days TTL)
- ✅ Automatic token refresh
- ✅ Secure token revocation
- ✅ Multi-session management
- ✅ Role-based access control

</td>
<td width="50%">

### 🛡️ **Security Features**
- ✅ bcrypt password hashing (cost 12)
- ✅ Rate limiting (100 req/min)
- ✅ Account lockout protection
- ✅ Security event logging
- ✅ CORS configuration
- ✅ Password policy enforcement

</td>
</tr>
<tr>
<td width="50%">

### 🏗️ **Architecture**
- ✅ Clean Architecture pattern
- ✅ Domain-Driven Design
- ✅ Interface-driven development
- ✅ Dependency injection
- ✅ Repository pattern
- ✅ SOLID principles

</td>
<td width="50%">

### 📊 **Observability**
- ✅ Structured logging (zap)
- ✅ Prometheus metrics
- ✅ Distributed tracing (Jaeger)
- ✅ Health checks
- ✅ Grafana dashboards
- ✅ Performance monitoring

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Prerequisites

<table>
<tr>
<td><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go"></td>
<td><img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis" alt="Redis"></td>
<td><img src="https://img.shields.io/badge/Docker-20+-2496ED?style=flat&logo=docker" alt="Docker"></td>
</tr>
</table>

### 🐳 Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Start all services with Docker Compose
make -f Makefile.auth docker-run

# Check service health
curl http://localhost:8080/health
```

### 🔧 Local Development

```bash
# Install dependencies
make -f Makefile.auth deps

# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Build and run
make -f Makefile.auth build
JWT_SECRET=your-super-secret-key make -f Makefile.auth run-dev
```

### ✅ Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Register a test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

---

## 📖 Documentation

<div align="center">

| 📚 **Guide** | 📝 **Description** | 🔗 **Link** |
|-------------|-------------------|-------------|
| **API Reference** | Complete API documentation with examples | [📖 View](./api-reference.md) |
| **Architecture Guide** | System design and architecture patterns | [🏗️ View](./architecture.md) |
| **Security Guide** | Security features and best practices | [🛡️ View](./security.md) |
| **Deployment Guide** | Production deployment instructions | [🚀 View](./deployment.md) |
| **Development Guide** | Local development setup | [🔧 View](./development.md) |
| **Configuration** | Configuration options and examples | [⚙️ View](./configuration.md) |

</div>

---

## 🔧 API Reference

### 🌐 Base URL
```
HTTP:  http://localhost:8080
gRPC:  localhost:50053
```

### 📋 Endpoints Overview

<details>
<summary><b>🔑 Authentication Endpoints</b></summary>

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/register` | Register new user | ❌ |
| `POST` | `/api/v1/auth/login` | User login | ❌ |
| `POST` | `/api/v1/auth/logout` | User logout | ✅ |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | ❌ |
| `POST` | `/api/v1/auth/validate` | Validate token | ❌ |
| `POST` | `/api/v1/auth/change-password` | Change password | ✅ |
| `GET` | `/api/v1/auth/me` | Get user info | ✅ |

</details>

<details>
<summary><b>📊 System Endpoints</b></summary>

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/api/v1/docs` | API documentation |

</details>

### 🔍 Example Requests

<details>
<summary><b>👤 User Registration</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

**Response:**
```json
{
  "user": {
    "id": "user_123",
    "email": "user@example.com",
    "role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

</details>

<details>
<summary><b>🔐 User Login</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "remember_me": true,
    "device_info": {
      "device_type": "desktop",
      "os": "macOS",
      "browser": "Chrome"
    }
  }'
```

</details>

<details>
<summary><b>🔄 Token Refresh</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

</details>

---

## 🛡️ Security

### 🔒 Security Features

<div align="center">

| 🛡️ **Feature** | ✅ **Status** | 📝 **Description** |
|----------------|---------------|-------------------|
| **Password Hashing** | ✅ Implemented | bcrypt with cost 12 |
| **JWT Security** | ✅ Implemented | HS256 signing, short TTL |
| **Rate Limiting** | ✅ Implemented | 100 req/min per IP |
| **Account Lockout** | ✅ Implemented | 5 failed attempts |
| **Session Management** | ✅ Implemented | Multi-session support |
| **Security Logging** | ✅ Implemented | All events tracked |

</div>

### 🔐 Password Policy

- **Minimum length:** 8 characters
- **Required:** Uppercase, lowercase, numbers, symbols
- **Forbidden:** Common passwords, sequential chars
- **Validation:** Real-time strength checking

### 🎯 JWT Configuration

```yaml
security:
  jwt_secret: "${JWT_SECRET}"
  access_token_ttl: "15m"    # Short-lived
  refresh_token_ttl: "168h"  # 7 days
  bcrypt_cost: 12
  max_login_attempts: 5
  lockout_duration: "30m"
```

---

## 📊 Monitoring

### 🔍 Health Checks

```bash
# Service health
curl http://localhost:8080/health

# Detailed health with dependencies
curl http://localhost:8080/health?detailed=true
```

### 📈 Metrics & Dashboards

<table>
<tr>
<td width="33%">

**🔥 Prometheus**
- Port: `9090`
- Metrics: Request latency, error rates, active sessions
- Alerts: High error rate, service down

</td>
<td width="33%">

**📊 Grafana**
- Port: `3000`
- Login: `admin/admin`
- Dashboards: Auth service overview, security events

</td>
<td width="33%">

**🔍 Jaeger**
- Port: `16686`
- Tracing: Request flows, performance bottlenecks
- Sampling: 10% of requests

</td>
</tr>
</table>

### 📋 Key Metrics

- `auth_requests_total` - Total authentication requests
- `auth_requests_duration` - Request duration histogram
- `auth_active_sessions` - Current active sessions
- `auth_failed_logins` - Failed login attempts
- `auth_token_refreshes` - Token refresh operations

---

## 🔧 Development

### 🛠️ Available Commands

```bash
# Development
make -f Makefile.auth build          # Build binary
make -f Makefile.auth run-dev        # Run in development mode
make -f Makefile.auth test           # Run tests
make -f Makefile.auth lint           # Run linter

# Docker
make -f Makefile.auth docker-build   # Build Docker image
make -f Makefile.auth docker-run     # Run with Docker Compose
make -f Makefile.auth docker-logs    # View logs

# Database
make -f Makefile.auth redis-cli      # Connect to Redis
make -f Makefile.auth redis-monitor  # Monitor Redis commands

# API Testing
make -f Makefile.auth test-api       # Test all endpoints
make -f Makefile.auth test-register  # Test registration
make -f Makefile.auth test-login     # Test login
```

### 🏗️ Project Structure

```
auth-service/
├── 📁 cmd/auth-service/           # Application entry point
├── 📁 internal/auth/              # Core business logic
│   ├── 📁 domain/                 # Domain entities
│   ├── 📁 application/            # Use cases
│   ├── 📁 infrastructure/         # External concerns
│   └── 📁 transport/              # API handlers
├── 📁 docs/                       # Documentation
├── 📁 configs/                    # Configuration files
└── 📁 scripts/                    # Utility scripts
```

---

## 🚀 Deployment

### 🐳 Docker Deployment

```bash
# Production build
docker build -f cmd/auth-service/Dockerfile -t auth-service:latest .

# Run with environment
docker run -d \
  -p 8080:8080 \
  -p 50053:50053 \
  -e JWT_SECRET=your-production-secret \
  -e REDIS_URL=redis://redis:6379 \
  auth-service:latest
```

### ☸️ Kubernetes Deployment

```bash
# Apply manifests
kubectl apply -f k8s/auth-service/

# Check status
kubectl get pods -l app=auth-service
```

### 🌍 Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | ✅ | - | JWT signing secret |
| `REDIS_URL` | ❌ | `redis://localhost:6379` | Redis connection |
| `LOG_LEVEL` | ❌ | `info` | Logging level |
| `ENVIRONMENT` | ❌ | `development` | Runtime environment |

---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details.

### 📋 Development Workflow

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes
4. **Add** tests for new functionality
5. **Run** `make -f Makefile.auth check`
6. **Submit** a pull request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

---

<div align="center">

**Made with ❤️ by the Go Coffee Team**

[🐛 Report Bug](https://github.com/DimaJoyti/go-coffee/issues) • [✨ Request Feature](https://github.com/DimaJoyti/go-coffee/issues) • [💬 Discussions](https://github.com/DimaJoyti/go-coffee/discussions)

</div>
