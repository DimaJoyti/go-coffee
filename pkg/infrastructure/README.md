# Infrastructure Layer

This package provides a comprehensive infrastructure layer for the Go Coffee platform, implementing Clean Architecture principles with enterprise-grade components for Redis, security, events, database management, and caching.

## 🏗️ Architecture Overview

```
pkg/infrastructure/
├── config/           # Configuration management
├── redis/           # Redis client and connection management
├── cache/           # Caching layer with Redis backend
├── database/        # Database connection and transaction management
├── security/        # JWT, encryption, and security services
├── events/          # Event store, publisher, and subscriber
├── container.go     # Dependency injection container
└── README.md        # This file
```

## 🚀 Features

### ✅ Redis Infrastructure
- **Connection Management**: Standalone, Cluster, and Sentinel support
- **Connection Pooling**: Configurable pool sizes and timeouts
- **Key Namespacing**: Automatic key prefixing for multi-tenancy
- **TLS Support**: Secure connections with certificate validation
- **Health Monitoring**: Connection health checks and statistics

### ✅ Caching Layer
- **Multi-Backend Support**: Redis-based caching with fallback options
- **Serialization**: JSON, MessagePack, and Gob support
- **Batch Operations**: Efficient multi-key operations
- **TTL Management**: Flexible expiration policies
- **Pattern Operations**: Wildcard key matching and deletion

### ✅ Database Infrastructure
- **Connection Pooling**: Advanced connection pool management
- **Transaction Support**: Safe transaction handling with rollback
- **Query Logging**: Configurable query logging and slow query detection
- **Health Monitoring**: Connection statistics and health checks
- **Multiple Databases**: Support for multiple database connections

### ✅ Security Services
- **JWT Management**: Token generation, validation, and refresh
- **Encryption**: AES-256-GCM encryption for sensitive data
- **Password Hashing**: bcrypt with configurable cost
- **Key Derivation**: scrypt-based key derivation
- **Security Policies**: Configurable password and security policies

### ✅ Event Infrastructure
- **Event Store**: Persistent event storage with Redis backend
- **Event Publisher**: Asynchronous event publishing with workers
- **Event Subscriber**: Pattern-based event subscription
- **Event Sourcing**: Complete event history and replay capabilities
- **Retry Logic**: Configurable retry policies with exponential backoff

### ✅ Container Management
- **Dependency Injection**: Centralized dependency management
- **Lifecycle Management**: Proper initialization and shutdown
- **Health Checks**: Comprehensive health monitoring
- **Configuration Validation**: Runtime configuration validation

## 📦 Quick Start

### 1. Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/DimaJoyti/go-coffee/pkg/infrastructure"
    "github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
    "github.com/DimaJoyti/go-coffee/pkg/logger"
)

func main() {
    // Load configuration
    cfg := config.DefaultInfrastructureConfig()
    
    // Create logger
    logger := logger.New("infrastructure")
    
    // Create container
    container := infrastructure.NewContainer(cfg, logger)
    
    // Initialize infrastructure
    ctx := context.Background()
    if err := container.Initialize(ctx); err != nil {
        log.Fatal("Failed to initialize infrastructure:", err)
    }
    defer container.Shutdown(ctx)
    
    // Use infrastructure components
    cache := container.GetCache()
    db := container.GetDatabase()
    jwtService := container.GetJWTService()
    
    // Your application logic here...
}
```

### 2. Redis Usage

```go
// Get Redis client
redisClient := container.GetRedis()

// Basic operations
err := redisClient.Set(ctx, "key", "value", time.Hour)
value, err := redisClient.Get(ctx, "key")

// Hash operations
err = redisClient.HSet(ctx, "user:123", "name", "John", "email", "john@example.com")
userData, err := redisClient.HGetAll(ctx, "user:123")

// Pub/Sub
err = redisClient.Publish(ctx, "notifications", "Hello World")
pubsub := redisClient.Subscribe(ctx, "notifications")
```

### 3. Cache Usage

```go
// Get cache service
cache := container.GetCache()

// Store and retrieve data
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

user := &User{ID: "123", Name: "John"}
err := cache.Set(ctx, "user:123", user, time.Hour)

var retrievedUser User
err = cache.Get(ctx, "user:123", &retrievedUser)

// Batch operations
users := map[string]interface{}{
    "user:1": &User{ID: "1", Name: "Alice"},
    "user:2": &User{ID: "2", Name: "Bob"},
}
err = cache.SetMulti(ctx, users, time.Hour)
```

### 4. Database Usage

```go
// Get database connection
db := container.GetDatabase()

// Simple query
var count int
err := db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)

// Using sqlx features
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

var users []User
err = db.Select(ctx, &users, "SELECT id, name, email FROM users WHERE active = $1", true)

// Transactions
tx, err := db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
if err != nil {
    return err
}

err = tx.Commit()
```

### 5. JWT Usage

```go
// Get JWT service
jwtService := container.GetJWTService()

// Generate tokens
tokenPair, err := jwtService.GenerateTokenPair(ctx, "user123", "john@example.com", "user", nil)

// Validate token
claims, err := jwtService.ValidateAccessToken(ctx, tokenPair.AccessToken)

// Refresh token
newTokenPair, err := jwtService.RefreshAccessToken(ctx, tokenPair.RefreshToken)
```

### 6. Event Usage

```go
// Get event services
eventStore := container.GetEventStore()
eventPublisher := container.GetEventPublisher()
eventSubscriber := container.GetEventSubscriber()

// Create and store event
event := &events.Event{
    ID:            "event-123",
    Type:          "user.created",
    AggregateID:   "user-123",
    AggregateType: "user",
    Data: map[string]interface{}{
        "name":  "John",
        "email": "john@example.com",
    },
}

err := eventStore.SaveEvent(ctx, event)

// Publish event
err = eventPublisher.Publish(ctx, event)

// Subscribe to events
handler := &MyEventHandler{}
err = eventSubscriber.Subscribe(ctx, []string{"user.created"}, handler)
```

## ⚙️ Configuration

### Environment Variables

```bash
# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=go_coffee
DB_USER=postgres
DB_PASSWORD=postgres

# Security
JWT_SECRET_KEY=your-secret-key
AES_ENCRYPTION_KEY=your-32-byte-encryption-key

# Environment
ENVIRONMENT=development
```

### Configuration File

See `config/infrastructure.yaml` for a complete configuration example with all available options.

## 🔧 Advanced Features

### Custom Event Handlers

```go
type UserEventHandler struct {
    logger *logger.Logger
}

func (h *UserEventHandler) Handle(ctx context.Context, event *events.Event) error {
    switch event.Type {
    case "user.created":
        return h.handleUserCreated(ctx, event)
    case "user.updated":
        return h.handleUserUpdated(ctx, event)
    default:
        return nil
    }
}

func (h *UserEventHandler) CanHandle(eventType string) bool {
    return strings.HasPrefix(eventType, "user.")
}

func (h *UserEventHandler) GetHandlerName() string {
    return "UserEventHandler"
}
```

### Health Monitoring

```go
// Check infrastructure health
status, err := container.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
}

log.Printf("Overall status: %s", status.Overall)
log.Printf("Database status: %v", status.Database)
log.Printf("Redis status: %v", status.Redis)
```

## 🧪 Testing

### Unit Tests

```go
func TestRedisClient(t *testing.T) {
    cfg := &config.RedisConfig{
        Host: "localhost",
        Port: 6379,
        DB:   15, // Use test database
    }
    
    logger := logger.New("test")
    client, err := redis.NewClient(cfg, logger)
    require.NoError(t, err)
    defer client.Close()
    
    ctx := context.Background()
    err = client.Set(ctx, "test:key", "test:value", time.Minute)
    require.NoError(t, err)
    
    value, err := client.Get(ctx, "test:key")
    require.NoError(t, err)
    assert.Equal(t, "test:value", value)
}
```

### Integration Tests

```go
func TestInfrastructureContainer(t *testing.T) {
    cfg := config.DefaultInfrastructureConfig()
    cfg.Redis.DB = 15 // Use test database
    
    logger := logger.New("test")
    container := infrastructure.NewContainer(cfg, logger)
    
    ctx := context.Background()
    err := container.Initialize(ctx)
    require.NoError(t, err)
    defer container.Shutdown(ctx)
    
    // Test components
    assert.NotNil(t, container.GetRedis())
    assert.NotNil(t, container.GetCache())
    assert.NotNil(t, container.GetJWTService())
}
```

## 📊 Monitoring and Observability

### Metrics

The infrastructure layer exposes Prometheus metrics for monitoring:

- Connection pool statistics
- Query performance metrics
- Cache hit/miss ratios
- Event processing metrics
- Error rates and latencies

### Health Checks

Built-in health checks for all components:

- Database connectivity
- Redis connectivity
- Cache functionality
- Event infrastructure status

### Logging

Structured logging with configurable levels:

- Query logging with execution times
- Security event logging
- Performance monitoring
- Error tracking

## 🔒 Security Considerations

### Production Deployment

1. **Use strong secrets**: Generate cryptographically secure keys
2. **Enable TLS**: Use encrypted connections for all external services
3. **Network security**: Restrict access to infrastructure components
4. **Monitoring**: Enable comprehensive logging and monitoring
5. **Backup**: Implement regular backup strategies

### Security Best Practices

1. **Rotate secrets**: Regularly rotate JWT and encryption keys
2. **Audit logs**: Monitor security events and access patterns
3. **Rate limiting**: Implement rate limiting to prevent abuse
4. **Input validation**: Validate all inputs and configurations
5. **Least privilege**: Use minimal required permissions

## 🤝 Contributing

1. Follow Clean Architecture principles
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Follow Go best practices and conventions
5. Ensure backward compatibility when possible

## 📝 License

This infrastructure layer is part of the Go Coffee platform and follows the same licensing terms.
