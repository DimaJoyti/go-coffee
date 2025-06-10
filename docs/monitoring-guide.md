# 📊 Go Coffee - Enhanced Monitoring & Metrics Guide

This guide covers the comprehensive monitoring and metrics system implemented in the Go Coffee platform using clean architecture principles.

## 🎯 Overview

The Go Coffee monitoring system provides:
- **Enhanced Health Checks** - Comprehensive infrastructure health monitoring
- **Prometheus Metrics** - Detailed metrics collection and exposure
- **Performance Monitoring** - Request/response performance tracking
- **Infrastructure Health** - Real-time infrastructure component monitoring
- **Request Tracing** - Distributed tracing capabilities
- **Alerting & Dashboards** - Monitoring visualization and alerting

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Gateway  │    │   Monitoring    │    │   Prometheus    │
│                 │───▶│   Middleware    │───▶│    Metrics      │
│  Clean HTTP     │    │                 │    │    Server       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Infrastructure │    │  Health Checker │    │    Grafana      │
│   Container     │───▶│                 │───▶│   Dashboards    │
│                 │    │  Component      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Components

### 1. Health Check System

**Location**: `pkg/infrastructure/monitoring/health.go`

**Features**:
- Concurrent health checks for all infrastructure components
- Configurable timeouts and thresholds
- Detailed health status reporting
- Component-specific health checks

**Health Checks**:
- ✅ Database connectivity and connection pool status
- ✅ Redis connectivity and pool statistics
- ✅ Cache functionality and hit ratios
- ✅ Session manager functionality
- ✅ JWT service token generation/validation
- ✅ Event store, publisher, and subscriber
- ✅ System resources (memory, goroutines, GC)

**Endpoints**:
- `GET /health` - Basic health check
- `GET /health/detailed` - Comprehensive health report with metrics

### 2. Prometheus Metrics

**Location**: `pkg/infrastructure/monitoring/prometheus.go`

**Metrics Categories**:

#### HTTP Metrics
- `go_coffee_http_requests_total` - Total HTTP requests by method, path, status
- `go_coffee_http_request_duration_seconds` - Request duration histogram
- `go_coffee_http_request_size_bytes` - Request size histogram
- `go_coffee_http_response_size_bytes` - Response size histogram

#### Infrastructure Metrics
- `go_coffee_database_connections` - Database connection pool metrics
- `go_coffee_redis_connections` - Redis connection pool metrics
- `go_coffee_cache_hit_ratio` - Cache hit ratio by cache type
- `go_coffee_session_active_sessions` - Active session count

#### System Metrics
- `go_coffee_system_goroutines` - Number of goroutines
- `go_coffee_system_memory_usage_bytes` - Memory usage
- `go_coffee_system_gc_duration_seconds` - Garbage collection duration

#### Business Metrics
- `go_coffee_business_orders_total` - Total orders by status and type
- `go_coffee_business_order_duration_seconds` - Order processing duration
- `go_coffee_business_user_sessions` - User sessions by role
- `go_coffee_business_errors_total` - Error count by type and component

#### Health Metrics
- `go_coffee_health_check_status` - Health check status (1=healthy, 0=unhealthy)
- `go_coffee_health_check_duration_seconds` - Health check duration

**Endpoints**:
- `GET /metrics` - Prometheus metrics endpoint

### 3. Performance Monitoring

**Location**: `pkg/infrastructure/middleware/performance.go`

**Features**:
- Request/response performance tracking
- Slow request detection and logging
- Request tracing with trace/span IDs
- Request profiling capabilities

**Middleware**:
- `PerformanceMiddleware` - Records HTTP metrics and detects slow requests
- `TracingMiddleware` - Adds tracing headers and context
- `ProfilingMiddleware` - Profiles requests based on sampling rate

### 4. Infrastructure Health Monitoring

**Integration**: Embedded in infrastructure container

**Monitoring**:
- Database connection health and statistics
- Redis connection health and pool stats
- Cache functionality and performance
- Session manager operations
- Event system health
- System resource utilization

## 🚀 Usage

### Starting the Monitoring System

The monitoring system is automatically initialized when starting the User Gateway:

```bash
cd cmd/user-gateway
go run main.go
```

**Monitoring Endpoints**:
- Health Check: http://localhost:8080/health
- Detailed Health: http://localhost:8080/health/detailed
- Prometheus Metrics: http://localhost:8080/metrics
- Prometheus Server: http://localhost:9090/metrics

### Configuration

**Environment Variables**:
```bash
# Prometheus configuration
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
PROMETHEUS_METRICS_PATH=/metrics

# Health check configuration
HEALTH_CHECK_INTERVAL=30s
HEALTH_CHECK_TIMEOUT=10s

# Performance monitoring
PERFORMANCE_MONITORING_ENABLED=true
SLOW_REQUEST_THRESHOLD=1s

# Tracing configuration
TRACING_ENABLED=true
TRACING_SAMPLE_RATE=1.0
```

**Configuration File**: `config/monitoring.yaml`

### Viewing Metrics

#### 1. Health Check Response
```bash
curl http://localhost:8080/health/detailed
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": "45ms",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "Database is healthy",
      "duration": "12ms",
      "metadata": {
        "open_connections": 10,
        "in_use": 2,
        "idle": 8
      }
    },
    "redis": {
      "status": "healthy",
      "message": "Redis is healthy",
      "duration": "8ms",
      "metadata": {
        "total_conns": 5,
        "idle_conns": 3,
        "stale_conns": 0
      }
    }
  },
  "summary": {
    "total": 8,
    "healthy": 8,
    "unhealthy": 0,
    "degraded": 0,
    "unknown": 0
  }
}
```

#### 2. Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```

**Sample Metrics**:
```
# HELP go_coffee_http_requests_total Total number of HTTP requests
# TYPE go_coffee_http_requests_total counter
go_coffee_http_requests_total{method="GET",path="/health",status="200"} 15

# HELP go_coffee_http_request_duration_seconds HTTP request duration in seconds
# TYPE go_coffee_http_request_duration_seconds histogram
go_coffee_http_request_duration_seconds_bucket{method="GET",path="/health",le="0.1"} 12
go_coffee_http_request_duration_seconds_bucket{method="GET",path="/health",le="0.3"} 15

# HELP go_coffee_system_goroutines Number of goroutines
# TYPE go_coffee_system_goroutines gauge
go_coffee_system_goroutines 45

# HELP go_coffee_database_connections Number of database connections
# TYPE go_coffee_database_connections gauge
go_coffee_database_connections{state="active"} 2
go_coffee_database_connections{state="idle"} 8
```

## 📈 Monitoring Best Practices

### 1. Health Check Strategy
- **Critical Components**: Database, Redis (required=true)
- **Optional Components**: Cache, Events (required=false)
- **Timeout Configuration**: 5s for external services, 3s for internal
- **Failure Threshold**: 3 consecutive failures before marking unhealthy

### 2. Metrics Collection
- **HTTP Metrics**: Track all requests with method, path, status
- **Performance Metrics**: Monitor P95 response times
- **Error Metrics**: Track error rates by component
- **Business Metrics**: Monitor order processing and user sessions

### 3. Alerting Rules
- **Error Rate**: Alert if > 5% error rate for 5 minutes
- **Response Time**: Alert if P95 > 2s for 3 minutes
- **Infrastructure**: Alert immediately on database/Redis failures
- **Resources**: Alert if memory > 80% or goroutines > 1000

### 4. Dashboard Design
- **Infrastructure Dashboard**: Health, connections, resources
- **Performance Dashboard**: Response times, throughput, errors
- **Business Dashboard**: Orders, users, revenue metrics

## 🔍 Troubleshooting

### Common Issues

#### 1. Health Check Failures
```bash
# Check detailed health status
curl http://localhost:8080/health/detailed

# Check infrastructure logs
docker logs go-coffee-user-gateway
```

#### 2. Metrics Not Appearing
```bash
# Verify Prometheus endpoint
curl http://localhost:8080/metrics

# Check Prometheus server
curl http://localhost:9090/metrics
```

#### 3. Performance Issues
```bash
# Check slow request logs
grep "Slow request detected" /var/log/go-coffee/app.log

# Monitor goroutine count
curl -s http://localhost:8080/metrics | grep goroutines
```

### Debugging Commands

```bash
# Health check with verbose output
curl -v http://localhost:8080/health/detailed

# Metrics with filtering
curl -s http://localhost:8080/metrics | grep "go_coffee_http"

# Performance monitoring
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/orders
```

## 🎯 Next Steps

1. **Grafana Integration**: Set up Grafana dashboards for visualization
2. **Alertmanager**: Configure alerting rules and notifications
3. **Log Aggregation**: Integrate with ELK stack or similar
4. **Distributed Tracing**: Add Jaeger or Zipkin integration
5. **Custom Metrics**: Add business-specific metrics
6. **SLA Monitoring**: Implement SLA tracking and reporting

## 📚 References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go Metrics Best Practices](https://prometheus.io/docs/guides/go-application/)
- [Clean Architecture Monitoring](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
