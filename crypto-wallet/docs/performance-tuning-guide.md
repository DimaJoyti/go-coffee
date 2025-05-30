# Performance Tuning Guide

This guide provides recommendations for tuning the performance of the Web3 wallet backend system.

## Redis Performance Tuning

### Connection Pooling

Optimize Redis connection pooling:

```go
redis.Config{
    PoolSize:            100,  // Adjust based on load
    MinIdleConns:        10,   // Keep some connections ready
    IdleTimeout:         10 * time.Minute,
    IdleCheckFrequency:  1 * time.Minute,
}
```

Recommendations:
- Set `PoolSize` to match your expected concurrent requests
- Keep `MinIdleConns` at 10-20% of `PoolSize`
- Monitor connection usage with Redis INFO command

### Pipelining

Use Redis pipelining for batch operations:

```go
pipe := redisClient.Pipeline()
pipe.Get(ctx, "key1")
pipe.Get(ctx, "key2")
pipe.Get(ctx, "key3")
results, err := pipe.Exec(ctx)
```

Recommendations:
- Use pipelining when performing multiple operations on the same keys
- Limit pipeline size to 100-1000 commands per batch
- Monitor pipeline performance with Redis SLOWLOG

### Cache TTL Optimization

Optimize cache TTL based on data volatility:

```go
// Frequently changing data
s.cache.Set(ctx, "volatile-key", data, 5 * time.Minute)

// Relatively stable data
s.cache.Set(ctx, "stable-key", data, 1 * time.Hour)

// Reference data
s.cache.Set(ctx, "reference-key", data, 24 * time.Hour)
```

Recommendations:
- Set shorter TTLs for frequently changing data
- Set longer TTLs for stable data
- Monitor cache hit rates and adjust TTLs accordingly

### Redis Cluster Configuration

Optimize Redis cluster for multi-region deployment:

```go
redis.ClusterOptions{
    RouteByLatency:     true,  // Route to lowest latency node
    RouteRandomly:      false, // Don't use random routing
    MaxRedirects:       3,     // Limit redirects
    ReadOnly:           true,  // Enable read-only queries on replicas
}
```

Recommendations:
- Enable `RouteByLatency` for multi-region deployments
- Use read replicas for read-heavy workloads
- Monitor cluster node health and rebalance as needed

## Kafka Performance Tuning

### Producer Configuration

Optimize Kafka producer for throughput:

```go
kafka.ConfigMap{
    "compression.type":    "snappy",  // Use compression
    "batch.size":          16384,     // Increase batch size
    "linger.ms":           5,         // Wait for more messages
    "acks":                "1",       // Acknowledge from leader only
    "retries":             3,         // Limit retries
    "retry.backoff.ms":    100,       // Backoff between retries
}
```

Recommendations:
- Use `snappy` compression for good balance of CPU and compression ratio
- Increase `batch.size` for higher throughput
- Set `linger.ms` to 5-10ms to allow batching
- Use `acks=1` for better performance when full durability isn't required

### Consumer Configuration

Optimize Kafka consumer for throughput:

```go
kafka.ConfigMap{
    "fetch.min.bytes":      1024,    // Minimum data to fetch
    "fetch.max.bytes":      52428800, // Maximum data to fetch (50MB)
    "max.poll.records":     500,     // Maximum records per poll
    "max.partition.fetch.bytes": 1048576, // 1MB per partition
    "auto.offset.reset":    "earliest", // Start from earliest offset
}
```

Recommendations:
- Increase `fetch.min.bytes` to reduce network round trips
- Set `max.poll.records` based on message processing time
- Monitor consumer lag and adjust parameters accordingly

### Partitioning Strategy

Optimize Kafka topic partitioning:

```bash
# Create topic with optimal partitions
kafka-topics.sh --create --topic order-events \
  --partitions 24 \
  --replication-factor 3
```

Recommendations:
- Set partition count to 2-3x the number of consumer instances
- Use a good partitioning key (e.g., user ID, order ID)
- Monitor partition balance and adjust as needed

### Consumer Group Configuration

Optimize consumer groups for parallel processing:

```go
// Ensure consumers are balanced across partitions
kafka.ConfigMap{
    "group.id":             "order-processor",
    "partition.assignment.strategy": "cooperative-sticky",
    "session.timeout.ms":   30000,
    "heartbeat.interval.ms": 3000,
}
```

Recommendations:
- Use `cooperative-sticky` assignment strategy for rebalancing
- Set `session.timeout.ms` to 3x `heartbeat.interval.ms`
- Monitor consumer group rebalancing events

## Service Performance Tuning

### Database Connection Pooling

Optimize database connection pooling:

```go
db.SetMaxOpenConns(100)      // Maximum open connections
db.SetMaxIdleConns(25)       // Keep some connections ready
db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections
```

Recommendations:
- Set `MaxOpenConns` based on database capacity and service count
- Keep `MaxIdleConns` at 25-30% of `MaxOpenConns`
- Monitor connection usage and adjust accordingly

### Timeouts and Retries

Configure proper timeouts and retries:

```go
// Context timeout for database operations
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// Retry with exponential backoff
backoff := time.Second
for retries := 3; retries > 0; retries-- {
    err := operation()
    if err == nil {
        break
    }
    time.Sleep(backoff)
    backoff *= 2 // Exponential backoff
}
```

Recommendations:
- Set timeouts based on expected operation time
- Use exponential backoff for retries
- Limit retry count to prevent cascading failures

### Circuit Breakers

Implement circuit breakers for external service calls:

```go
// Using github.com/sony/gobreaker
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "database-circuit",
    MaxRequests: 5,
    Interval:    10 * time.Second,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 10 && failureRatio >= 0.5
    },
})

// Use the circuit breaker
result, err := cb.Execute(func() (interface{}, error) {
    return service.Call()
})
```

Recommendations:
- Implement circuit breakers for all external service calls
- Configure circuit breakers to trip after 50% failure rate
- Set appropriate reset timeouts based on service recovery time

### Kubernetes Resource Limits

Configure appropriate Kubernetes resource limits:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

Recommendations:
- Set CPU requests based on average usage
- Set memory requests based on application needs
- Monitor resource usage and adjust accordingly
- Use Horizontal Pod Autoscaler for automatic scaling

## Monitoring and Profiling

### Prometheus Metrics

Implement key performance metrics:

```go
// Request duration histogram
requestDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
    },
    []string{"method", "path", "status"},
)

// Cache hit ratio gauge
cacheHitRatio := prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "cache_hit_ratio",
        Help: "Cache hit ratio",
    },
)
```

Recommendations:
- Track request durations with histograms
- Monitor cache hit ratios
- Track database query times
- Monitor Kafka producer/consumer lag

### Profiling

Enable runtime profiling:

```go
import _ "net/http/pprof"

// In your main function
go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

Recommendations:
- Use pprof for CPU and memory profiling
- Periodically capture profiles during load tests
- Analyze profiles to identify bottlenecks

## Load Testing

### Load Test Strategy

1. **Baseline Testing**: Establish performance baseline
2. **Component Testing**: Test individual components
3. **Integration Testing**: Test component interactions
4. **Scalability Testing**: Test with increasing load
5. **Endurance Testing**: Test over extended periods
6. **Chaos Testing**: Test with simulated failures

Recommendations:
- Use tools like k6, Locust, or JMeter
- Test with realistic data and scenarios
- Monitor all system components during tests
- Analyze results and adjust configuration accordingly

## Conclusion

Performance tuning is an iterative process. Monitor your system, identify bottlenecks, make adjustments, and repeat. Focus on the components that have the biggest impact on overall performance.
