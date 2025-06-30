# Observability & Monitoring Package

This package provides comprehensive observability and monitoring capabilities for the Go Coffee AI Agents using OpenTelemetry, structured logging, and metrics collection.

## Overview

The observability package implements the three pillars of observability:

1. **Distributed Tracing**: Track requests across service boundaries with OpenTelemetry
2. **Metrics Collection**: Collect and expose business and system metrics via Prometheus
3. **Structured Logging**: JSON-formatted logs with trace correlation and context

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Observability   │    │ Telemetry       │    │ Metrics         │
│ Manager         │───▶│ Provider        │    │ Collector       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Structured      │    │ Tracing Helper  │    │ Business Logger │
│ Logger          │    │ (OpenTelemetry) │    │ & Audit Logger  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ JSON Logs       │    │ Jaeger/OTLP     │    │ Prometheus      │
│ with Trace IDs  │    │ Exporters       │    │ Metrics         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. Configuration (`config.go`)

Comprehensive configuration for all observability components:

```yaml
observability:
  service_name: "beverage-inventor-agent"
  service_version: "1.0.0"
  environment: "production"
  
  tracing:
    enabled: true
    sampling_rate: 0.1  # 10% sampling in production
    
  metrics:
    enabled: true
    collect_interval: 30s
    
  logging:
    enabled: true
    level: "info"
    format: "json"
    include_trace: true
    include_span: true
    
  exporters:
    jaeger:
      enabled: true
      endpoint: "http://jaeger:14268/api/traces"
    prometheus:
      enabled: true
      port: 9090
      path: "/metrics"
    otlp:
      enabled: true
      endpoint: "http://otel-collector:4317"
```

### 2. Telemetry Provider (`provider.go`)

Sets up OpenTelemetry components:

```go
// Initialize telemetry
config := observability.GetConfigForEnvironment("beverage-inventor", "production")
err := observability.InitGlobalObservability(config)
if err != nil {
    log.Fatal("Failed to initialize observability:", err)
}

// Get components
tracer := observability.GetGlobalTracer("beverage-service")
meter := observability.GetGlobalMeter("beverage-service")
```

### 3. Metrics Collection (`metrics.go`)

Comprehensive metrics for business and system monitoring:

#### Business Metrics
- `beverages_created_total`: Total beverages created
- `tasks_created_total`: Total tasks created
- `ai_requests_total`: Total AI provider requests
- `notifications_sent_total`: Total notifications sent

#### System Metrics
- `request_duration_seconds`: HTTP request latency
- `database_query_duration_seconds`: Database operation latency
- `kafka_publish_duration_seconds`: Kafka message publishing latency
- `circuit_breaker_trips_total`: Circuit breaker activations

#### Usage Example
```go
metrics := observability.GetGlobalMetrics()

// Record beverage creation
metrics.RecordBeverageCreated(ctx, "Mars Base", true, duration)

// Record AI request
metrics.RecordAIRequest(ctx, "gemini", "generate_description", duration, true)

// Record Kafka message
metrics.RecordKafkaMessage(ctx, "beverage.created", "produce", duration)
```

### 4. Distributed Tracing (`tracing.go`)

OpenTelemetry-based distributed tracing:

```go
tracing := observability.GetGlobalTracing()

// Start HTTP span
ctx, span := tracing.StartHTTPSpan(ctx, "POST", "/api/beverages", userAgent)
defer span.End()

// Start database span
ctx, dbSpan := tracing.StartDatabaseSpan(ctx, "INSERT", "beverages")
defer dbSpan.End()

// Start AI span
ctx, aiSpan := tracing.StartAISpan(ctx, "gemini", "generate_description")
defer aiSpan.End()

// Record success/error
if err != nil {
    tracing.RecordError(span, err, "Operation failed")
} else {
    tracing.RecordSuccess(span, "Operation completed")
}
```

### 5. Structured Logging (`logging.go`)

JSON-formatted logging with trace correlation:

```go
logger := observability.GetGlobalLogger()

// Basic logging
logger.Info("Beverage created", 
    "beverage_id", "123",
    "name", "Cosmic Coffee",
    "theme", "Mars Base")

// Context-aware logging (includes trace IDs)
logger.InfoContext(ctx, "AI request completed",
    "provider", "gemini",
    "duration_ms", 1500,
    "tokens_used", 150)

// Business event logging
businessLogger := observability.GetGlobalObservability().GetBusinessLogger()
businessLogger.LogBeverageCreated(ctx, "123", "Cosmic Coffee", "Mars Base", true, duration)
```

### 6. Observability Manager (`manager.go`)

Coordinates all observability components:

```go
// Initialize observability
config := observability.DefaultTelemetryConfig("beverage-inventor")
manager, err := observability.NewObservabilityManager(config)

// Record business events with full observability
err = manager.RecordBusinessEvent(ctx, "beverage_creation", func(ctx context.Context) error {
    return beverageService.Create(ctx, beverage)
})

// Get health status
health := manager.HealthCheck()

// Shutdown gracefully
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
manager.Shutdown(ctx)
```

## Integration Examples

### Beverage Inventor Agent Integration

```go
// Initialize observability in main.go
func main() {
    config := observability.GetConfigForEnvironment("beverage-inventor", os.Getenv("ENVIRONMENT"))
    if err := observability.InitGlobalObservability(config); err != nil {
        log.Fatal("Failed to initialize observability:", err)
    }
    defer observability.ShutdownGlobalObservability(context.Background())

    // Get observability components
    logger := observability.GetGlobalLogger()
    metrics := observability.GetGlobalMetrics()
    tracing := observability.GetGlobalTracing()

    // Initialize use case with observability
    useCase := usecases.NewBeverageInventorUseCase(
        beverageRepo,
        eventPublisher,
        aiProvider,
        taskManager,
        notificationSvc,
        logger,
    )

    logger.Info("Beverage Inventor Agent started")
}
```

### HTTP Handler Integration

```go
func (h *BeverageHandler) CreateBeverage(w http.ResponseWriter, r *http.Request) {
    // Start tracing
    tracing := observability.GetGlobalTracing()
    ctx, span := tracing.StartHTTPSpan(r.Context(), r.Method, r.URL.String(), r.UserAgent())
    defer span.End()

    start := time.Now()
    
    // Process request
    beverage, err := h.useCase.InventBeverage(ctx, request)
    duration := time.Since(start)

    // Record metrics
    metrics := observability.GetGlobalMetrics()
    metrics.RecordRequest(ctx, r.Method, r.URL.Path, duration, err == nil)

    if err != nil {
        tracing.RecordError(span, err, "Beverage creation failed")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tracing.RecordSuccess(span, "Beverage created successfully")
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(beverage)
}
```

### Use Case Integration

```go
func (uc *BeverageInventorUseCase) InventBeverage(ctx context.Context, req *InventBeverageRequest) (*InventBeverageResponse, error) {
    // Business event tracing
    return observability.TraceOperation(ctx, "beverage_invention", func(ctx context.Context) error {
        logger := observability.GetGlobalLogger()
        businessLogger := observability.GetGlobalObservability().GetBusinessLogger()
        
        logger.InfoContext(ctx, "Starting beverage invention",
            "ingredients", req.Ingredients,
            "theme", req.Theme,
            "use_ai", req.UseAI)

        start := time.Now()
        
        // Generate beverage
        beverage, err := uc.generateBeverage(ctx, req)
        if err != nil {
            return err
        }

        // Log business event
        businessLogger.LogBeverageCreated(ctx, 
            beverage.ID.String(), 
            beverage.Name, 
            beverage.Theme, 
            req.UseAI, 
            time.Since(start))

        return nil
    })
}
```

## Monitoring & Alerting

### Prometheus Metrics

Access metrics at `http://localhost:9090/metrics`:

```
# HELP beverages_created_total Total number of beverages created
# TYPE beverages_created_total counter
beverages_created_total{theme="Mars Base",ai_used="true"} 42

# HELP request_duration_seconds Request duration in seconds
# TYPE request_duration_seconds histogram
request_duration_seconds_bucket{method="POST",endpoint="/api/beverages",le="0.1"} 95
request_duration_seconds_bucket{method="POST",endpoint="/api/beverages",le="0.5"} 99
request_duration_seconds_bucket{method="POST",endpoint="/api/beverages",le="1.0"} 100
```

### Grafana Dashboards

Key metrics to monitor:

1. **Request Rate**: `rate(requests_total[5m])`
2. **Error Rate**: `rate(requests_error_total[5m]) / rate(requests_total[5m])`
3. **Response Time**: `histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m]))`
4. **Business Metrics**: `rate(beverages_created_total[1h])`

### Alerting Rules

```yaml
groups:
  - name: beverage-inventor-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(requests_error_total[5m]) / rate(requests_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          
      - alert: SlowRequests
        expr: histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow requests detected"
```

## Distributed Tracing

### Jaeger Integration

View traces in Jaeger UI at `http://localhost:16686`:

- **Service Map**: Visualize service dependencies
- **Trace Timeline**: See request flow across services
- **Error Analysis**: Identify failing operations
- **Performance Analysis**: Find slow operations

### Trace Correlation

Logs include trace and span IDs for correlation:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "Beverage created successfully",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "beverage_id": "123",
  "name": "Cosmic Coffee"
}
```

## Best Practices

### 1. Trace Naming
- Use descriptive operation names
- Include business context in span names
- Follow consistent naming conventions

### 2. Metric Labels
- Keep cardinality low (< 1000 unique combinations)
- Use meaningful label names
- Avoid user-specific labels

### 3. Log Correlation
- Always include trace and span IDs
- Use structured logging with consistent fields
- Include business context in logs

### 4. Error Handling
- Record errors in spans with proper status codes
- Include error context in logs
- Use appropriate log levels

### 5. Performance
- Use sampling for high-volume traces
- Batch metric exports
- Avoid blocking operations in hot paths

This observability package provides enterprise-grade monitoring and debugging capabilities for the Go Coffee AI agent ecosystem, enabling effective troubleshooting, performance optimization, and business insights.
