# ⚙️ Auth Service Configuration Guide

<div align="center">

![Configuration](https://img.shields.io/badge/Configuration-Complete-green?style=for-the-badge)
![Environment](https://img.shields.io/badge/Environment-Variables-blue?style=for-the-badge)
![YAML](https://img.shields.io/badge/YAML-Config-orange?style=for-the-badge)

**Complete configuration guide for all deployment scenarios**

</div>

---

## 🎯 Configuration Overview

The Auth Service supports multiple configuration methods with a clear hierarchy:

1. **Environment Variables** (highest priority)
2. **Configuration Files** (YAML)
3. **Default Values** (lowest priority)

### 📋 Configuration Sources

<table>
<tr>
<td width="33%">

**🌍 Environment Variables**
- Production secrets
- Runtime overrides
- Container deployment
- CI/CD pipelines

</td>
<td width="33%">

**📄 YAML Files**
- Structured configuration
- Environment-specific settings
- Default values
- Complex nested objects

</td>
<td width="33%">

**🔧 Command Line**
- Development overrides
- Testing scenarios
- Debug settings
- One-time configurations

</td>
</tr>
</table>

---

## 🌍 Environment Variables

### 🔐 Security Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | ✅ | - | JWT signing secret (256+ bits) |
| `JWT_ACCESS_TTL` | ❌ | `15m` | Access token TTL |
| `JWT_REFRESH_TTL` | ❌ | `168h` | Refresh token TTL (7 days) |
| `BCRYPT_COST` | ❌ | `12` | bcrypt hashing cost |
| `MAX_LOGIN_ATTEMPTS` | ❌ | `5` | Max failed login attempts |
| `LOCKOUT_DURATION` | ❌ | `30m` | Account lockout duration |

### 🗄️ Database Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_URL` | ❌ | `redis://localhost:6379` | Redis connection URL |
| `REDIS_PASSWORD` | ❌ | - | Redis authentication password |
| `REDIS_DB` | ❌ | `0` | Redis database number |
| `REDIS_POOL_SIZE` | ❌ | `10` | Connection pool size |
| `REDIS_MAX_RETRIES` | ❌ | `3` | Maximum retry attempts |

### 🌐 Server Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_PORT` | ❌ | `8080` | HTTP server port |
| `GRPC_PORT` | ❌ | `50053` | gRPC server port |
| `HOST` | ❌ | `0.0.0.0` | Server bind address |
| `READ_TIMEOUT` | ❌ | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | ❌ | `30s` | HTTP write timeout |
| `SHUTDOWN_TIMEOUT` | ❌ | `30s` | Graceful shutdown timeout |

### 📊 Monitoring Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LOG_LEVEL` | ❌ | `info` | Logging level (debug, info, warn, error) |
| `LOG_FORMAT` | ❌ | `json` | Log format (json, text) |
| `PROMETHEUS_ENABLED` | ❌ | `true` | Enable Prometheus metrics |
| `JAEGER_ENDPOINT` | ❌ | - | Jaeger tracing endpoint |
| `ENVIRONMENT` | ❌ | `development` | Runtime environment |

### 🛡️ Rate Limiting Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RATE_LIMIT_ENABLED` | ❌ | `true` | Enable rate limiting |
| `RATE_LIMIT_RPM` | ❌ | `100` | Requests per minute |
| `RATE_LIMIT_BURST` | ❌ | `20` | Burst size |

---

## 📄 YAML Configuration

### 🔧 Complete Configuration File

```yaml
# cmd/auth-service/config/config.yaml

# Server Configuration
server:
  http_port: 8080
  grpc_port: 50053
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  shutdown_timeout: "30s"
  max_header_bytes: 1048576  # 1MB

# Redis Configuration
redis:
  url: "${REDIS_URL:redis://localhost:6379}"
  password: "${REDIS_PASSWORD:}"
  db: 0
  max_retries: 3
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  pool_timeout: "4s"
  idle_timeout: "5m"
  max_conn_age: "30m"

# Security Configuration
security:
  jwt:
    secret: "${JWT_SECRET}"
    access_token_ttl: "15m"
    refresh_token_ttl: "168h"  # 7 days
    issuer: "auth-service"
    audience: "go-coffee-users"
    algorithm: "HS256"
  
  password:
    bcrypt_cost: 12
    policy:
      min_length: 8
      max_length: 128
      require_uppercase: true
      require_lowercase: true
      require_numbers: true
      require_symbols: true
      forbidden_patterns:
        - "password"
        - "123456"
        - "qwerty"
  
  account:
    max_login_attempts: 5
    lockout_duration: "30m"
    session_timeout: "24h"
    max_sessions_per_user: 10

# Rate Limiting Configuration
rate_limiting:
  enabled: true
  global:
    requests_per_minute: 100
    burst_size: 20
    cleanup_interval: "1m"
  
  endpoints:
    "/api/v1/auth/login":
      requests_per_minute: 10
      burst_size: 3
    "/api/v1/auth/register":
      requests_per_minute: 5
      burst_size: 2
    "/api/v1/auth/refresh":
      requests_per_minute: 20
      burst_size: 5

# CORS Configuration
cors:
  enabled: true
  allowed_origins:
    - "https://app.go-coffee.com"
    - "https://admin.go-coffee.com"
    - "http://localhost:3000"  # Development
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Origin"
    - "Content-Type"
    - "Authorization"
    - "X-Requested-With"
    - "X-Request-ID"
  expose_headers:
    - "Content-Length"
    - "X-Request-ID"
    - "X-Rate-Limit-Remaining"
  allow_credentials: true
  max_age: 86400  # 24 hours

# Logging Configuration
logging:
  level: "${LOG_LEVEL:info}"
  format: "${LOG_FORMAT:json}"
  output: "stdout"
  file:
    enabled: false
    path: "/var/log/auth-service.log"
    max_size: 100  # MB
    max_age: 7     # days
    max_backups: 3
    compress: true
  
  # Structured logging fields
  fields:
    service: "auth-service"
    version: "1.0.0"
  
  # Log sampling (reduce log volume in production)
  sampling:
    enabled: false
    initial: 100
    thereafter: 100

# Monitoring Configuration
monitoring:
  enabled: true
  
  # Health checks
  health:
    enabled: true
    path: "/health"
    detailed_path: "/health/detailed"
  
  # Prometheus metrics
  metrics:
    enabled: true
    port: 9090
    path: "/metrics"
    namespace: "auth_service"
    subsystem: ""
  
  # Distributed tracing
  tracing:
    enabled: true
    service_name: "auth-service"
    jaeger:
      endpoint: "${JAEGER_ENDPOINT:http://localhost:14268/api/traces}"
      sample_rate: 0.1
      max_tag_value_length: 256
  
  # Profiling (development only)
  pprof:
    enabled: false
    port: 6060

# TLS Configuration
tls:
  enabled: false
  cert_file: "/etc/ssl/certs/auth-service.crt"
  key_file: "/etc/ssl/private/auth-service.key"
  min_version: "1.2"
  cipher_suites:
    - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
    - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
    - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"

# Feature Flags
features:
  registration_enabled: true
  password_reset_enabled: false
  multi_factor_auth_enabled: false
  session_analytics_enabled: true
  security_notifications_enabled: true

# Environment
environment: "${ENVIRONMENT:development}"
```

---

## 🌍 Environment-Specific Configurations

### 🔧 Development Configuration

```yaml
# config/development.yaml
server:
  http_port: 8080
  grpc_port: 50053

redis:
  url: "redis://localhost:6379"
  db: 0

security:
  jwt:
    secret: "development-secret-key-not-for-production"
    access_token_ttl: "1h"  # Longer for development
    refresh_token_ttl: "24h"

logging:
  level: "debug"
  format: "text"  # More readable in development

monitoring:
  tracing:
    sample_rate: 1.0  # Trace everything in development
  pprof:
    enabled: true

features:
  registration_enabled: true
  password_reset_enabled: true
```

### 🧪 Testing Configuration

```yaml
# config/testing.yaml
server:
  http_port: 8081
  grpc_port: 50054

redis:
  url: "redis://localhost:6379"
  db: 1  # Different database for tests

security:
  jwt:
    secret: "test-secret-key"
    access_token_ttl: "5m"
    refresh_token_ttl: "10m"
  
  password:
    bcrypt_cost: 4  # Faster for tests

rate_limiting:
  enabled: false  # Disable for tests

logging:
  level: "warn"  # Reduce noise in tests

monitoring:
  enabled: false
```

### 🚀 Production Configuration

```yaml
# config/production.yaml
server:
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"

redis:
  url: "${REDIS_URL}"
  password: "${REDIS_PASSWORD}"
  pool_size: 20
  max_retries: 5

security:
  jwt:
    secret: "${JWT_SECRET}"
    access_token_ttl: "15m"
    refresh_token_ttl: "168h"
  
  password:
    bcrypt_cost: 14  # Higher security

rate_limiting:
  enabled: true
  global:
    requests_per_minute: 1000  # Higher for production

cors:
  allowed_origins:
    - "https://app.go-coffee.com"
    - "https://admin.go-coffee.com"

logging:
  level: "info"
  format: "json"
  sampling:
    enabled: true

monitoring:
  tracing:
    sample_rate: 0.01  # Sample 1% in production

tls:
  enabled: true
  cert_file: "/etc/ssl/certs/auth-service.crt"
  key_file: "/etc/ssl/private/auth-service.key"
```

---

## 🐳 Docker Configuration

### 📦 Docker Environment

```dockerfile
# Dockerfile environment variables
ENV ENVIRONMENT=production
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json

# Runtime configuration
ENV JWT_SECRET=""
ENV REDIS_URL="redis://redis:6379"
ENV REDIS_PASSWORD=""

# Server configuration
ENV HTTP_PORT=8080
ENV GRPC_PORT=50053
ENV HOST=0.0.0.0

# Monitoring
ENV PROMETHEUS_ENABLED=true
ENV JAEGER_ENDPOINT="http://jaeger:14268/api/traces"
```

### 🐳 Docker Compose Environment

```yaml
# docker-compose.yml
version: '3.8'

services:
  auth-service:
    environment:
      # Security
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      
      # Database
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
      
      # Server
      - ENVIRONMENT=production
      - LOG_LEVEL=info
      
      # Monitoring
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - PROMETHEUS_ENABLED=true
    
    # Load configuration from file
    volumes:
      - ./config/production.yaml:/app/config/config.yaml:ro
```

---

## ☸️ Kubernetes Configuration

### 🔧 ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: auth-service
data:
  config.yaml: |
    server:
      http_port: 8080
      grpc_port: 50053
    
    redis:
      url: "redis://redis-service:6379"
      pool_size: 20
    
    security:
      jwt:
        access_token_ttl: "15m"
        refresh_token_ttl: "168h"
      password:
        bcrypt_cost: 12
    
    logging:
      level: "info"
      format: "json"
    
    monitoring:
      enabled: true
      tracing:
        enabled: true
        service_name: "auth-service"
```

### 🔐 Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: auth-service
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  redis-password: <base64-encoded-redis-password>
```

### 🚀 Deployment Environment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  template:
    spec:
      containers:
      - name: auth-service
        env:
        # From ConfigMap
        - name: ENVIRONMENT
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        
        # From Secret
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: jwt-secret
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: redis-password
        
        # Computed values
        - name: REDIS_URL
          value: "redis://:$(REDIS_PASSWORD)@redis-service:6379"
        
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
      
      volumes:
      - name: config
        configMap:
          name: auth-service-config
```

---

## 🔧 Configuration Validation

### ✅ Validation Rules

```go
// Configuration validation example
func ValidateConfig(cfg *Config) error {
    var errors []string
    
    // JWT Secret validation
    if cfg.Security.JWT.Secret == "" {
        errors = append(errors, "JWT_SECRET is required")
    }
    if len(cfg.Security.JWT.Secret) < 32 {
        errors = append(errors, "JWT_SECRET must be at least 32 characters")
    }
    
    // Redis URL validation
    if cfg.Redis.URL == "" {
        errors = append(errors, "REDIS_URL is required")
    }
    
    // Port validation
    if cfg.Server.HTTPPort < 1 || cfg.Server.HTTPPort > 65535 {
        errors = append(errors, "HTTP_PORT must be between 1 and 65535")
    }
    
    // bcrypt cost validation
    if cfg.Security.Password.BcryptCost < 10 || cfg.Security.Password.BcryptCost > 15 {
        errors = append(errors, "BCRYPT_COST must be between 10 and 15")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}
```

### 🧪 Configuration Testing

```bash
# Test configuration loading
go run ./cmd/auth-service --config-test

# Validate environment variables
./scripts/validate-config.sh

# Test with different environments
ENVIRONMENT=development go run ./cmd/auth-service --config-test
ENVIRONMENT=production go run ./cmd/auth-service --config-test
```

---

## 🔍 Troubleshooting

### 🚨 Common Configuration Issues

<table>
<tr>
<td width="50%">

**❌ JWT Secret Issues**
```bash
# Error: JWT_SECRET not set
export JWT_SECRET="your-256-bit-secret"

# Error: JWT_SECRET too short
# Use at least 32 characters
openssl rand -base64 32
```

</td>
<td width="50%">

**❌ Redis Connection Issues**
```bash
# Error: Redis connection failed
# Check Redis URL format
REDIS_URL="redis://username:password@host:port/db"

# Test Redis connection
redis-cli -u $REDIS_URL ping
```

</td>
</tr>
<tr>
<td width="50%">

**❌ Port Binding Issues**
```bash
# Error: Port already in use
# Check what's using the port
lsof -i :8080

# Use different port
HTTP_PORT=8081 ./auth-service
```

</td>
<td width="50%">

**❌ Configuration File Issues**
```bash
# Error: Config file not found
# Check file path
ls -la ./config/config.yaml

# Validate YAML syntax
yamllint ./config/config.yaml
```

</td>
</tr>
</table>

---

<div align="center">

**⚙️ Configuration Documentation**

[🏠 Main README](./README.md) • [📖 API Reference](./api-reference.md) • [🏗️ Architecture](./architecture.md) • [🚀 Deployment](./deployment.md)

</div>
