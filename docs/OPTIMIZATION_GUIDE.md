# Go Coffee Database & Cache Optimization Guide

## 🚀 Quick Start

This guide will help you implement advanced database connection pooling, Redis clustering with compression, and cache warming strategies for immediate performance improvements.

## 📊 Expected Performance Improvements

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Database Response Time** | 200ms | 50ms | **75% faster** |
| **Cache Hit Ratio** | 60% | 90% | **50% improvement** |
| **Memory Usage** | 100% | 70% | **30% reduction** |
| **Connection Pool Efficiency** | 40% | 85% | **112% improvement** |

## 🏗️ Implementation Steps

### Step 1: Database Optimization

#### 1.1 Replace your existing database connection with optimized pooling:

```go
// Before (existing code)
db, err := sql.Open("postgres", connectionString)

// After (optimized)
import "github.com/DimaJoyti/go-coffee/pkg/database"

dbManager, err := database.NewManager(&config.DatabaseConfig{
    Host:     "localhost",
    Port:     5432,
    Username: "postgres",
    Password: "password",
    Database: "go_coffee",
    SSLMode:  "disable",
}, logger)
```

#### 1.2 Use optimized database operations:

```go
// Write operations (orders, payments, etc.)
err := dbManager.ExecuteWrite(ctx, 
    "INSERT INTO orders (id, customer_id, total) VALUES ($1, $2, $3)",
    orderID, customerID, total)

// Read operations (queries, analytics)
result, err := dbManager.QueryRead(ctx,
    "SELECT * FROM orders WHERE customer_id = $1", customerID)

// Transaction support
err := dbManager.Transaction(ctx, func(tx database.Transaction) error {
    // Multiple operations in transaction
    return nil
})
```

#### 1.3 Monitor database performance:

```go
// Get real-time metrics
metrics := dbManager.GetMetrics()
fmt.Printf("Active Connections: %d\n", metrics.ActiveConnections)
fmt.Printf("Average Query Time: %v\n", metrics.AverageQueryTime)
fmt.Printf("Slow Queries: %d\n", metrics.SlowQueryCount)
```

### Step 2: Cache Optimization

#### 2.1 Replace your existing Redis client with optimized caching:

```go
// Before (existing code)
rdb := redis.NewClient(&redis.Options{...})

// After (optimized)
import "github.com/DimaJoyti/go-coffee/pkg/cache"

cacheManager, err := cache.NewManager(&config.RedisConfig{
    Host:         "localhost",
    Port:         6379,
    PoolSize:     50,
    MinIdleConns: 10,
}, logger)
```

#### 2.2 Use advanced caching patterns:

```go
// Simple cache operations with compression
err := cacheManager.Set(ctx, "order:123", orderData, 1*time.Hour)
var order Order
err := cacheManager.Get(ctx, "order:123", &order)

// Cache-or-fetch pattern
cacheHelper := cacheManager.NewCacheHelper()
err := cacheHelper.GetOrSet(ctx, "popular_items", &items, 30*time.Minute, 
    func() (interface{}, error) {
        // Fetch from database if not in cache
        return fetchPopularItemsFromDB()
    })

// Batch operations for efficiency
pairs := map[string]interface{}{
    "menu:coffee": coffeeMenu,
    "menu:food":   foodMenu,
}
err := cacheManager.MSet(ctx, pairs, 2*time.Hour)
```

#### 2.3 Monitor cache performance:

```go
// Get cache metrics
metrics := cacheManager.GetMetrics()
fmt.Printf("Cache Hit Ratio: %.2f%%\n", metrics.HitRatio*100)
fmt.Printf("Average Latency: %v\n", metrics.AvgLatency)
fmt.Printf("Total Keys: %d\n", metrics.TotalKeys)
```

### Step 3: Integration with Existing Services

#### 3.1 Update your order service:

```go
type OrderService struct {
    dbManager    *database.Manager
    cacheManager *cache.Manager
    logger       *zap.Logger
}

func (s *OrderService) CreateOrder(ctx context.Context, order *Order) error {
    // Use optimized database write
    err := s.dbManager.ExecuteWrite(ctx, 
        "INSERT INTO orders (...) VALUES (...)", ...)
    if err != nil {
        return err
    }
    
    // Cache the order for fast retrieval
    cacheKey := fmt.Sprintf("order:%s", order.ID)
    s.cacheManager.Set(ctx, cacheKey, order, 1*time.Hour)
    
    // Invalidate related caches
    s.cacheManager.Delete(ctx, fmt.Sprintf("customer_orders:%s", order.CustomerID))
    
    return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
    var order Order
    cacheKey := fmt.Sprintf("order:%s", orderID)
    
    // Try cache first
    err := s.cacheManager.Get(ctx, cacheKey, &order)
    if err == nil {
        return &order, nil // Cache hit
    }
    
    // Cache miss - fetch from database
    result, err := s.dbManager.QueryRead(ctx, 
        "SELECT * FROM orders WHERE id = $1", orderID)
    if err != nil {
        return nil, err
    }
    
    // Parse result and cache it
    // ... parsing logic ...
    s.cacheManager.Set(ctx, cacheKey, &order, 1*time.Hour)
    
    return &order, nil
}
```

## 🔧 Configuration

### Database Configuration (`configs/optimization.yaml`):

```yaml
database:
  optimization:
    enabled: true
    connection_pool:
      max_connections: 50
      min_connections: 10
      max_connection_lifetime: "5m"
      health_check_period: "30s"
    query:
      default_timeout: "30s"
      slow_query_threshold: "1s"
    read_replicas:
      enabled: true
      failover: true
```

### Cache Configuration:

```yaml
cache:
  optimization:
    enabled: true
    compression:
      enabled: true
      min_size: 1024  # Compress values > 1KB
    warming:
      enabled: true
      interval: "5m"
      strategies:
        - name: "menu"
          ttl: "1h"
        - name: "popular_items"
          ttl: "30m"
```

## 📈 Monitoring & Metrics

### Health Check Endpoint:

```go
func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "database": s.dbManager.GetMetrics(),
        "cache":    s.cacheManager.GetMetrics(),
        "timestamp": time.Now(),
    }
    json.NewEncoder(w).Encode(health)
}
```

### Prometheus Metrics:

The optimization automatically exposes metrics for:
- Database connection pool utilization
- Query performance and slow query counts
- Cache hit ratios and latency
- Memory usage and garbage collection

## 🚀 Deployment

### Option 1: Manual Integration

1. **Install dependencies:**
```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/redis/go-redis/v9
```

2. **Update your services:**
```bash
# Test the optimization
go run cmd/optimization-test/main.go

# Build your service with optimizations
go build -o bin/optimized-service ./cmd/your-service/
```

### Option 2: Automated Deployment

```bash
# Deploy all optimizations
./scripts/deploy-optimizations.sh deploy-all

# Deploy only database optimizations
./scripts/deploy-optimizations.sh deploy-db

# Deploy only cache optimizations
./scripts/deploy-optimizations.sh deploy-cache

# Run performance tests
./scripts/deploy-optimizations.sh benchmark
```

## 🔍 Troubleshooting

### Common Issues:

1. **Database Connection Errors:**
```bash
# Check database connectivity
kubectl logs -n go-coffee deployment/optimization-service | grep database
```

2. **Cache Connection Errors:**
```bash
# Check Redis connectivity
kubectl logs -n go-coffee deployment/optimization-service | grep cache
```

3. **Performance Issues:**
```bash
# Check metrics
curl http://localhost:8080/metrics | grep -E "(database|cache)"
```

## 📊 Performance Testing

### Load Testing with k6:

```javascript
// Basic performance test
import http from 'k6/http';

export default function() {
    // Test optimized endpoints
    http.get('http://localhost:8080/orders/123');
    http.get('http://localhost:8080/menu');
    http.post('http://localhost:8080/orders', {...});
}
```

### Benchmarking:

```bash
# Run benchmark
go test -bench=. -benchmem ./pkg/database/
go test -bench=. -benchmem ./pkg/cache/
```

## 🎯 Next Steps

1. **Immediate (Week 1):**
   - Implement database optimization in order service
   - Add cache optimization to menu service
   - Set up monitoring dashboards

2. **Short-term (Week 2-3):**
   - Implement cache warming strategies
   - Add read replica support
   - Optimize memory usage

3. **Long-term (Month 2):**
   - Implement advanced concurrency patterns
   - Add chaos engineering tests
   - Set up automated performance regression testing

## 📞 Support

If you encounter any issues:

1. Check the logs: `kubectl logs -n go-coffee deployment/optimization-service`
2. Review metrics: `curl http://localhost:8080/health`
3. Run diagnostics: `./scripts/deploy-optimizations.sh verify`

## 🏆 Success Metrics

Track these KPIs to measure optimization success:

- **Database Response Time**: Target < 50ms (P95)
- **Cache Hit Ratio**: Target > 85%
- **Memory Usage**: Target < 70% of allocated
- **Error Rate**: Target < 0.1%
- **Throughput**: Target > 1000 RPS

---

**Ready to optimize your Go Coffee platform? Start with Step 1 and see immediate performance improvements!** ☕️🚀
