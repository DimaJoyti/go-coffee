# HTTP API Layer

This package provides a comprehensive HTTP API layer for the Go Coffee AI agents, featuring RESTful endpoints, middleware, observability, and robust error handling with Go 1.22's enhanced ServeMux.

## Overview

The HTTP API layer implements:

1. **RESTful API Design**: Clean, consistent REST endpoints for all resources
2. **Advanced Middleware**: Request ID, CORS, logging, tracing, metrics, recovery
3. **Comprehensive Observability**: Full tracing, metrics, and structured logging
4. **Robust Error Handling**: Typed errors with proper HTTP status codes
5. **Resource Management**: Beverages, tasks, AI operations, and system monitoring
6. **API Documentation**: Built-in OpenAPI specification and interactive docs

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ HTTP Server     │    │ Middleware      │    │ Route Handlers  │
│ • Go 1.22 Mux   │───▶│ Chain           │───▶│ • Health        │
│ • Graceful      │    │ • Request ID    │    │ • Beverages     │
│   Shutdown      │    │ • CORS          │    │ • Tasks         │
└─────────────────┘    │ • Logging       │    │ • AI            │
         │              │ • Tracing       │    │ • Metrics       │
         ▼              │ • Metrics       │    └─────────────────┘
┌─────────────────┐    │ • Recovery      │              │
│ Observability   │    └─────────────────┘              ▼
│ • Tracing       │              │              ┌─────────────────┐
│ • Metrics       │              ▼              │ Response Utils  │
│ • Logging       │    ┌─────────────────┐      │ • JSON          │
└─────────────────┘    │ Error Handling  │      │ • Pagination    │
                       │ • Typed Errors  │      │ • Validation    │
                       │ • HTTP Status   │      │ • Filtering     │
                       │ • Recovery      │      └─────────────────┘
                       └─────────────────┘
```

## Components

### 1. HTTP Server (`server.go`)

Modern HTTP server with Go 1.22 ServeMux:

```go
// Initialize HTTP server
server := api.NewServer(httpConfig, logger, metrics, tracing)
if err := server.Start(ctx); err != nil {
    log.Fatal("Failed to start HTTP server:", err)
}

// Graceful shutdown
defer func() {
    if err := server.Stop(ctx); err != nil {
        log.Error("Failed to stop HTTP server:", err)
    }
}()
```

### 2. Middleware Chain (`middleware.go`)

Comprehensive middleware stack:

```go
// Middleware chain (applied in reverse order)
h := handler
h = mc.Recovery(h)      // Panic recovery
h = mc.Metrics(h)       // Metrics collection
h = mc.Tracing(h)       // Distributed tracing
h = mc.Logging(h)       // Request/response logging
h = mc.CORS(h)          // Cross-origin support
h = mc.RequestID(h)     // Unique request IDs

// Optional middleware
h = mc.Authentication(h) // JWT authentication
h = mc.RateLimit(h)     // Rate limiting
h = mc.Security(h)      // Security headers
h = mc.Compression(h)   // Response compression
```

### 3. Route Handlers

#### Health Handler (`handlers/health.go`)
```go
// Comprehensive health checks
GET /health          // Overall system health
GET /health/ready    // Readiness check (K8s)
GET /health/live     // Liveness check (K8s)

// Health check response
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "24h15m30s",
  "checks": {
    "database": {
      "status": "healthy",
      "duration": "5ms",
      "details": {"connections": 15}
    },
    "kafka": {
      "status": "healthy", 
      "duration": "12ms"
    },
    "ai_providers": {
      "status": "warning",
      "message": "1 of 2 providers healthy"
    }
  },
  "summary": {
    "total": 3,
    "healthy": 2,
    "warning": 1,
    "error": 0
  }
}
```

#### Beverage Handler (`handlers/beverage.go`)
```go
// Beverage management endpoints
GET    /api/v1/beverages           // List beverages
POST   /api/v1/beverages           // Create beverage
GET    /api/v1/beverages/{id}      // Get beverage
PUT    /api/v1/beverages/{id}      // Update beverage
DELETE /api/v1/beverages/{id}      // Delete beverage
POST   /api/v1/beverages/generate  // AI-generate beverages
GET    /api/v1/beverages/search    // Search beverages
GET    /api/v1/beverages/stats     // Beverage statistics

// Create beverage request
{
  "name": "Mars Colony Coffee",
  "description": "Robust coffee for harsh Martian environment",
  "theme": "Mars Base",
  "ingredients": [
    {
      "name": "Coffee Beans",
      "quantity": "200",
      "unit": "g",
      "type": "base"
    }
  ],
  "instructions": ["Grind beans", "Brew with hot water"],
  "prep_time_minutes": 15,
  "servings": 2,
  "difficulty": "medium",
  "tags": ["coffee", "mars"]
}

// AI beverage generation
POST /api/v1/beverages/generate
{
  "theme": "Mars Base",
  "ingredients": ["coffee", "protein powder"],
  "dietary_restrictions": ["vegan"],
  "complexity": "medium",
  "count": 3
}
```

#### Task Handler (`handlers/task.go`)
```go
// Task management endpoints
GET    /api/v1/tasks              // List tasks
POST   /api/v1/tasks              // Create task
GET    /api/v1/tasks/{id}         // Get task
PUT    /api/v1/tasks/{id}         // Update task
DELETE /api/v1/tasks/{id}         // Delete task
POST   /api/v1/tasks/generate     // AI-generate tasks
PUT    /api/v1/tasks/{id}/status  // Update task status
GET    /api/v1/tasks/stats        // Task statistics

// Create task request
{
  "title": "Setup Mars Colony Beverage Production",
  "description": "Establish automated beverage production system",
  "priority": "high",
  "status": "open",
  "estimated_time_hours": 40,
  "tags": ["mars", "automation"],
  "due_date": "2024-02-01T00:00:00Z"
}

// AI task generation
POST /api/v1/tasks/generate
{
  "context": "Mars colony beverage production",
  "goal": "Establish sustainable beverage supply chain",
  "priority": "high",
  "skills": ["food science", "logistics"],
  "count": 5
}
```

#### AI Handler (`handlers/ai.go`)
```go
// AI operation endpoints
POST /api/v1/ai/text        // Generate text
POST /api/v1/ai/chat        // Chat completion
POST /api/v1/ai/embedding   // Generate embeddings
GET  /api/v1/ai/providers   // List AI providers
GET  /api/v1/ai/models      // List AI models
GET  /api/v1/ai/usage       // Usage statistics
GET  /api/v1/ai/health      // Provider health

// Text generation request
POST /api/v1/ai/text
{
  "model": "gpt-4",
  "prompt": "Create a coffee recipe for Mars colonists",
  "max_tokens": 500,
  "temperature": 0.8
}

// Chat completion request
POST /api/v1/ai/chat
{
  "model": "gemini-pro",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert beverage scientist"
    },
    {
      "role": "user", 
      "content": "How to make coffee on Mars?"
    }
  ],
  "max_tokens": 1000
}
```

#### Metrics Handler (`handlers/metrics.go`)
```go
// Monitoring endpoints
GET /metrics              // Prometheus metrics
GET /api/v1/metrics       // JSON metrics
GET /api/v1/stats         // Application statistics

// JSON metrics response
{
  "http": {
    "requests_total": 1890,
    "requests_success": 1801,
    "average_duration": 0.245
  },
  "ai": {
    "requests_total": 390,
    "tokens_used": 69134,
    "total_cost": 12.45
  },
  "database": {
    "connections_active": 15,
    "queries_total": 5678
  }
}
```

### 4. Response Utilities (`utils.go`)

Standardized response handling:

```go
// Standard API response format
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Meta      *Meta       `json:"meta,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    RequestID string      `json:"request_id,omitempty"`
}

// Success response
api.WriteJSONResponse(w, http.StatusOK, data)

// Error response
api.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid input")

// Paginated response
api.WritePaginatedResponse(w, http.StatusOK, items, pagination, total)

// Parse request parameters
pagination := api.ParsePaginationParams(r)  // page, per_page
sort := api.ParseSortParams(r, "created_at") // sort_field, sort_order
filter := api.ParseFilterParams(r)          // search, filter_*
```

## API Endpoints

### Health & Monitoring
```
GET  /health                    # System health check
GET  /health/ready              # Readiness probe
GET  /health/live               # Liveness probe
GET  /metrics                   # Prometheus metrics
GET  /api/v1/metrics            # JSON metrics
GET  /api/v1/stats              # Application statistics
```

### Beverages
```
GET    /api/v1/beverages           # List beverages
POST   /api/v1/beverages           # Create beverage
GET    /api/v1/beverages/{id}      # Get beverage
PUT    /api/v1/beverages/{id}      # Update beverage
DELETE /api/v1/beverages/{id}      # Delete beverage
POST   /api/v1/beverages/generate  # AI-generate beverages
GET    /api/v1/beverages/search    # Search beverages
GET    /api/v1/beverages/stats     # Beverage statistics
```

### Tasks
```
GET    /api/v1/tasks              # List tasks
POST   /api/v1/tasks              # Create task
GET    /api/v1/tasks/{id}         # Get task
PUT    /api/v1/tasks/{id}         # Update task
DELETE /api/v1/tasks/{id}         # Delete task
POST   /api/v1/tasks/generate     # AI-generate tasks
PUT    /api/v1/tasks/{id}/status  # Update task status
GET    /api/v1/tasks/stats        # Task statistics
```

### AI Operations
```
POST /api/v1/ai/text        # Generate text
POST /api/v1/ai/chat        # Chat completion
POST /api/v1/ai/embedding   # Generate embeddings
GET  /api/v1/ai/providers   # List AI providers
GET  /api/v1/ai/models      # List AI models
GET  /api/v1/ai/usage       # Usage statistics
GET  /api/v1/ai/health      # Provider health
```

### Documentation
```
GET /api/v1/docs            # API documentation
GET /api/v1/openapi.json    # OpenAPI specification
```

## Configuration

### HTTP Server Configuration
```yaml
http:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
  
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Content-Type", "Authorization"]
    
  rate_limiting:
    enabled: true
    requests_per_second: 100
    burst: 200
    
  security:
    enabled: true
    hsts_max_age: 31536000
    content_type_nosniff: true
    frame_options: "DENY"
```

### Environment Variables
```bash
# HTTP Server
GOCOFFEE_HTTP_PORT=8080
GOCOFFEE_HTTP_READ_TIMEOUT=30s
GOCOFFEE_HTTP_WRITE_TIMEOUT=30s

# CORS
GOCOFFEE_HTTP_CORS_ENABLED=true
GOCOFFEE_HTTP_CORS_ALLOWED_ORIGINS="*"

# Rate Limiting
GOCOFFEE_HTTP_RATE_LIMITING_ENABLED=true
GOCOFFEE_HTTP_RATE_LIMITING_RPS=100

# Security
GOCOFFEE_HTTP_SECURITY_ENABLED=true
```

## Usage Examples

### Starting the Server
```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "go-coffee-ai-agents/internal/api"
    "go-coffee-ai-agents/internal/config"
    "go-coffee-ai-agents/internal/observability"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Initialize observability
    logger := observability.NewStructuredLogger()
    metrics := observability.NewMetricsCollector()
    tracing := observability.NewTracingHelper()
    
    // Initialize HTTP server
    server := api.NewServer(cfg.HTTP, logger, metrics, tracing)
    
    // Start server
    ctx := context.Background()
    if err := server.Start(ctx); err != nil {
        log.Fatal("Failed to start server:", err)
    }
    
    // Wait for shutdown signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    // Graceful shutdown
    if err := server.Stop(ctx); err != nil {
        log.Error("Failed to stop server:", err)
    }
}
```

### Making API Requests
```bash
# Health check
curl http://localhost:8080/health

# List beverages with pagination
curl "http://localhost:8080/api/v1/beverages?page=1&per_page=10&sort_field=created_at&sort_order=desc"

# Generate beverages with AI
curl -X POST http://localhost:8080/api/v1/beverages/generate \
  -H "Content-Type: application/json" \
  -d '{
    "theme": "Mars Base",
    "ingredients": ["coffee", "protein powder"],
    "complexity": "medium",
    "count": 3
  }'

# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Setup Mars Colony Beverage Production",
    "description": "Establish automated beverage production system",
    "priority": "high",
    "status": "open"
  }'

# Generate text with AI
curl -X POST http://localhost:8080/api/v1/ai/text \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "prompt": "Create a coffee recipe for Mars colonists",
    "max_tokens": 500
  }'

# Get system metrics
curl http://localhost:8080/api/v1/metrics

# Get application statistics
curl "http://localhost:8080/api/v1/stats?range=24h"
```

## Error Handling

### Standard Error Response
```json
{
  "success": false,
  "error": {
    "code": "invalid_request",
    "message": "Invalid request body",
    "details": "Field 'name' is required"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### Common Error Codes
```go
// HTTP 400 Bad Request
"invalid_request"      // Invalid request format
"validation_error"     // Input validation failed
"missing_parameter"    // Required parameter missing

// HTTP 401 Unauthorized
"unauthorized"         // Authentication required
"invalid_token"        // Invalid or expired token

// HTTP 403 Forbidden
"forbidden"           // Access denied
"insufficient_permissions" // Lacks required permissions

// HTTP 404 Not Found
"not_found"           // Resource not found
"endpoint_not_found"  // API endpoint not found

// HTTP 409 Conflict
"conflict"            // Resource conflict
"duplicate_resource"  // Resource already exists

// HTTP 429 Too Many Requests
"rate_limit_exceeded" // Rate limit exceeded
"quota_exceeded"      // Usage quota exceeded

// HTTP 500 Internal Server Error
"internal_server_error" // Unexpected server error
"database_error"       // Database operation failed
"ai_service_error"     // AI service unavailable

// HTTP 503 Service Unavailable
"service_unavailable"  // Service temporarily unavailable
"maintenance_mode"     // System under maintenance
```

## Observability

### Request Tracing
Every HTTP request is automatically traced:

```json
{
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "operation": "HTTP GET /api/v1/beverages",
  "duration_ms": 125,
  "status": "success",
  "attributes": {
    "http.method": "GET",
    "http.url": "/api/v1/beverages",
    "http.status_code": 200,
    "http.user_agent": "curl/7.68.0"
  }
}
```

### Request Logging
Structured request/response logging:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "HTTP request completed",
  "request_id": "req_123456789",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "method": "GET",
  "path": "/api/v1/beverages",
  "status_code": 200,
  "duration_ms": 125,
  "response_size": 2048,
  "user_agent": "curl/7.68.0",
  "remote_addr": "192.168.1.100"
}
```

### Metrics Collection
Automatic HTTP metrics:

```
# Request counts
http_requests_total{method="GET",status="200"} 1234
http_requests_total{method="POST",status="201"} 567

# Request duration
http_request_duration_seconds_bucket{le="0.1"} 100
http_request_duration_seconds_bucket{le="0.5"} 450

# Active connections
http_connections_active 15
```

## Best Practices

### 1. API Design
- Use RESTful resource naming conventions
- Implement proper HTTP status codes
- Provide consistent error responses
- Support pagination for list endpoints
- Include resource URLs in responses

### 2. Request Handling
- Validate all input parameters
- Sanitize user input to prevent injection
- Use appropriate timeouts for operations
- Implement idempotency for safe operations
- Support content negotiation

### 3. Error Management
- Return meaningful error messages
- Use appropriate HTTP status codes
- Include error codes for programmatic handling
- Log errors with sufficient context
- Implement proper error recovery

### 4. Performance
- Use efficient JSON serialization
- Implement response compression
- Cache frequently accessed data
- Monitor response times and optimize
- Use connection pooling for databases

### 5. Security
- Validate and sanitize all inputs
- Implement proper authentication/authorization
- Use HTTPS in production
- Set appropriate security headers
- Implement rate limiting and DDoS protection

This HTTP API layer provides a robust, observable, and scalable foundation for the Go Coffee AI agent ecosystem, enabling efficient resource management, AI operations, and system monitoring with enterprise-grade reliability and performance!
