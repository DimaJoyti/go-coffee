# 🚀 1: System Design Fundamentals

## 📋 Overview

Master core system design concepts using Go Coffee components as practical examples. This establishes the foundation for all advanced topics.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Understand scalability principles and trade-offs
- Know reliability and availability patterns
- Grasp consistency models and their implications
- Master performance metrics and optimization
- Analyze real-world implementations in Go Coffee

---

## 📖 1.1 Scalability Fundamentals

### Core Concepts

#### Horizontal vs Vertical Scaling
- **Vertical Scaling (Scale Up)**: Add more power (CPU, RAM) to existing machines
- **Horizontal Scaling (Scale Out)**: Add more machines to the pool of resources

#### Load Distribution Patterns
- **Round Robin**: Distribute requests evenly across servers
- **Weighted Round Robin**: Distribute based on server capacity
- **Least Connections**: Route to server with fewest active connections
- **IP Hash**: Route based on client IP hash

#### Bottleneck Identification
- **CPU Bound**: High CPU utilization
- **Memory Bound**: High memory usage
- **I/O Bound**: Disk or network limitations
- **Database Bound**: Database query performance

### 🔍 Go Coffee Analysis

#### Study the API Gateway Scaling Pattern

<augment_code_snippet path="api-gateway/server/http_server.go" mode="EXCERPT">
````go
func NewHTTPServer(config *config.Config, producerClient pb.ProducerServiceClient) *HTTPServer {
    server := &HTTPServer{
        config:         config,
        producerClient: producerClient,
    }

    // Create router with middleware chain
    mux := http.NewServeMux()
    
    // Register routes with load balancing support
    mux.HandleFunc("/order", server.handleOrder)
    mux.HandleFunc("/health", server.handleHealth)
    
    // HTTP server with optimized timeouts
    server.Server = &http.Server{
        Addr:         fmt.Sprintf(":%d", config.Server.Port),
        Handler:      server.logMiddleware(server.corsMiddleware(mux)),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    return server
}
````
</augment_code_snippet>

#### Analyze Producer Service Scaling

<augment_code_snippet path="producer/kafka/producer.go" mode="EXCERPT">
````go
type Producer struct {
    writer *kafka.Writer
    config *config.Config
    logger *slog.Logger
}

func NewProducer(config *config.Config, logger *slog.Logger) *Producer {
    writer := &kafka.Writer{
        Addr:         kafka.TCP(config.Kafka.Brokers...),
        Topic:        config.Kafka.Topic,
        Balancer:     &kafka.LeastBytes{}, // Load balancing strategy
        RequiredAcks: kafka.RequireAll,    // Reliability setting
        Async:        true,                // Performance optimization
    }
    
    return &Producer{
        writer: writer,
        config: config,
        logger: logger,
    }
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 1.1: Analyze Current Scalability

#### Step 1: Run Load Tests
```bash
# Install load testing tools
go install github.com/rakyll/hey@latest

# Test API Gateway performance
hey -n 1000 -c 10 http://localhost:8080/health

# Test Producer Service
hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" \
  -d '{"customer_name":"Test","coffee_type":"Latte"}' \
  http://localhost:3000/order
```

#### Step 2: Monitor Resource Usage
```bash
# Monitor Docker containers
docker stats

# Monitor system resources
top
htop

# Check Kafka performance
docker exec -it kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 --describe --all-groups
```

#### Step 3: Identify Bottlenecks
```bash
# Check database connections
docker exec -it postgres psql -U go_coffee_user -d go_coffee \
  -c "SELECT count(*) FROM pg_stat_activity;"

# Check Redis performance
docker exec -it redis redis-cli info stats
```

### 💡 Practice Question 1.1
**"The Go Coffee API Gateway is receiving 1000 requests/second, but response times are increasing. How would you scale it?"**

**Analysis Framework:**
1. **Identify the bottleneck** (CPU, memory, database, network)
2. **Choose scaling strategy** (horizontal vs vertical)
3. **Implement load balancing**
4. **Add monitoring and alerting**

---

## 📖 1.2 Reliability & Availability

### Core Concepts

#### Fault Tolerance Patterns
- **Circuit Breaker**: Prevent cascading failures
- **Retry with Backoff**: Handle transient failures
- **Bulkhead**: Isolate critical resources
- **Timeout**: Prevent hanging requests

#### Redundancy Strategies
- **Active-Active**: Multiple active instances
- **Active-Passive**: Primary with standby
- **N+1 Redundancy**: Extra capacity for failures
- **Geographic Distribution**: Multi-region deployment

#### Disaster Recovery
- **RTO (Recovery Time Objective)**: Maximum downtime
- **RPO (Recovery Point Objective)**: Maximum data loss
- **Backup Strategies**: Full, incremental, differential
- **Failover Procedures**: Automated vs manual

### 🔍 Go Coffee Analysis

#### Study Consumer Service Reliability

<augment_code_snippet path="consumer/worker/worker.go" mode="EXCERPT">
````go
func (w *Worker) processMessage(ctx context.Context, msg kafka.Message) error {
    // Implement retry logic with exponential backoff
    maxRetries := 3
    baseDelay := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := w.handleOrder(ctx, msg.Value)
        if err == nil {
            return nil // Success
        }
        
        // Log the error and retry
        w.logger.Error("Failed to process message", 
            "attempt", attempt+1,
            "error", err,
            "message_key", string(msg.Key))
        
        if attempt < maxRetries-1 {
            delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
            time.Sleep(delay)
        }
    }
    
    // Send to dead letter queue after max retries
    return w.sendToDeadLetterQueue(msg)
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 1.2: Implement Circuit Breaker

#### Step 1: Create Circuit Breaker Package
```bash
mkdir -p pkg/resilience
cd pkg/resilience
```

#### Step 2: Implement Circuit Breaker
```go
// pkg/resilience/circuit_breaker.go
package resilience

import (
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    state       State
    failures    int
    lastFailure time.Time
    mutex       sync.RWMutex
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures: maxFailures,
        timeout:     timeout,
        state:       StateClosed,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if cb.state == StateOpen {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = StateHalfOpen
        } else {
            return errors.New("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
        return err
    }
    
    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = StateClosed
    return nil
}
```

#### Step 3: Integrate with Producer Service
```go
// producer/handler/order_handler.go
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    err := h.circuitBreaker.Execute(func() error {
        return h.producer.PublishOrder(order)
    })
    
    if err != nil {
        if err.Error() == "circuit breaker is open" {
            http.Error(w, "Service temporarily unavailable", 
                http.StatusServiceUnavailable)
            return
        }
        // Handle other errors...
    }
}
```

### 💡 Practice Question 1.2
**"How would you ensure 99.99% uptime for the Go Coffee ordering system?"**

**Solution Framework:**
1. **Eliminate single points of failure**
2. **Implement redundancy at every layer**
3. **Add health checks and monitoring**
4. **Create automated failover procedures**
5. **Design graceful degradation**

---

## 📖 1.3 Consistency Models

### Core Concepts

#### ACID Properties
- **Atomicity**: All or nothing transactions
- **Consistency**: Data integrity constraints
- **Isolation**: Concurrent transaction isolation
- **Durability**: Committed data persists

#### Consistency Levels
- **Strong Consistency**: All nodes see the same data simultaneously
- **Eventual Consistency**: Nodes will eventually converge
- **Weak Consistency**: No guarantees about when data will be consistent
- **Causal Consistency**: Causally related operations are seen in order

#### CAP Theorem
- **Consistency**: All nodes see the same data
- **Availability**: System remains operational
- **Partition Tolerance**: System continues despite network failures

### 🔍 Go Coffee Analysis

#### Study Database Transactions

<augment_code_snippet path="crypto-wallet/internal/database/postgres.go" mode="EXCERPT">
````go
func (db *PostgresDB) CreateOrderWithPayment(ctx context.Context, order *Order, payment *Payment) error {
    tx, err := db.conn.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable, // Strong consistency
    })
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback() // Ensure rollback on error
    
    // Create order
    _, err = tx.ExecContext(ctx, 
        "INSERT INTO orders (id, customer_id, total) VALUES ($1, $2, $3)",
        order.ID, order.CustomerID, order.Total)
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    
    // Process payment
    _, err = tx.ExecContext(ctx,
        "INSERT INTO payments (id, order_id, amount, status) VALUES ($1, $2, $3, $4)",
        payment.ID, order.ID, payment.Amount, "completed")
    if err != nil {
        return fmt.Errorf("failed to process payment: %w", err)
    }
    
    // Commit transaction - ensures atomicity
    return tx.Commit()
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 1.3: Implement Eventual Consistency

#### Step 1: Create Event Store
```go
// pkg/eventstore/event_store.go
package eventstore

type Event struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    Data      []byte    `json:"data"`
    Timestamp time.Time `json:"timestamp"`
    Version   int       `json:"version"`
}

type EventStore interface {
    SaveEvents(streamID string, events []Event, expectedVersion int) error
    GetEvents(streamID string, fromVersion int) ([]Event, error)
}

type RedisEventStore struct {
    client redis.Client
}

func (es *RedisEventStore) SaveEvents(streamID string, events []Event, expectedVersion int) error {
    pipe := es.client.Pipeline()
    
    for _, event := range events {
        eventData, _ := json.Marshal(event)
        pipe.RPush(ctx, "stream:"+streamID, eventData)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}
```

### 💡 Practice Question 1.3
**"How would you handle inventory consistency across multiple Go Coffee locations?"**

**Solution Options:**
1. **Strong Consistency**: Use distributed transactions (2PC)
2. **Eventual Consistency**: Use event sourcing with compensation
3. **Hybrid**: Strong for critical operations, eventual for others

---

## 📖 1.4 Performance Metrics

### Core Concepts

#### Latency vs Throughput
- **Latency**: Time to process a single request
- **Throughput**: Number of requests processed per unit time
- **Trade-off**: Often inversely related

#### Response Time Percentiles
- **P50 (Median)**: 50% of requests complete faster
- **P95**: 95% of requests complete faster
- **P99**: 99% of requests complete faster
- **P99.9**: 99.9% of requests complete faster

#### Performance Metrics
- **QPS**: Queries per second
- **TPS**: Transactions per second
- **Error Rate**: Percentage of failed requests
- **Apdex**: Application Performance Index

### 🔍 Go Coffee Analysis

#### Study Monitoring Implementation

<augment_code_snippet path="pkg/monitoring/metrics.go" mode="EXCERPT">
````go
type Metrics struct {
    requestDuration *prometheus.HistogramVec
    requestCount    *prometheus.CounterVec
    errorCount      *prometheus.CounterVec
}

func NewMetrics() *Metrics {
    return &Metrics{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_request_duration_seconds",
                Help:    "HTTP request duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method", "endpoint", "status_code"},
        ),
        requestCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "endpoint", "status_code"},
        ),
    }
}

func (m *Metrics) RecordRequest(method, endpoint, statusCode string, duration time.Duration) {
    m.requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
    m.requestCount.WithLabelValues(method, endpoint, statusCode).Inc()
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 1.4: Add Performance Monitoring

#### Step 1: Create Performance Middleware
```go
// pkg/middleware/performance.go
func PerformanceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start)
        
        // Record metrics
        metrics.RecordRequest(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.statusCode),
            duration,
        )
        
        // Log slow requests
        if duration > 100*time.Millisecond {
            log.Printf("Slow request: %s %s took %v", 
                r.Method, r.URL.Path, duration)
        }
    })
}
```

#### Step 2: Add to API Gateway
```go
// api-gateway/server/http_server.go
func (s *HTTPServer) setupMiddleware(mux *http.ServeMux) http.Handler {
    return middleware.PerformanceMiddleware(
        middleware.LoggingMiddleware(
            middleware.CORSMiddleware(mux),
        ),
    )
}
```

### 💡 Practice Question 1.4
**"The Go Coffee API has P95 latency of 500ms. How would you optimize it to under 100ms?"**

**Optimization Strategy:**
1. **Profile the application** to find bottlenecks
2. **Optimize database queries** with indexes and query optimization
3. **Add caching layers** (Redis, CDN)
4. **Optimize serialization** (use faster JSON libraries)
5. **Implement connection pooling**
6. **Add horizontal scaling**

---

## 🎯 1 Completion Checklist

### Knowledge Mastery
- [ ] Understand horizontal vs vertical scaling trade-offs
- [ ] Can identify system bottlenecks using metrics
- [ ] Know reliability patterns (circuit breaker, retry, timeout)
- [ ] Understand consistency models and CAP theorem
- [ ] Can calculate and interpret performance metrics

### Practical Skills
- [ ] Can run load tests and analyze results
- [ ] Can implement circuit breaker pattern
- [ ] Can add performance monitoring to services
- [ ] Can identify and fix performance bottlenecks
- [ ] Can design for high availability

### Go Coffee Analysis
- [ ] Understand API Gateway scaling patterns
- [ ] Analyzed Producer/Consumer reliability patterns
- [ ] Studied database transaction consistency
- [ ] Examined monitoring and metrics implementation
- [ ] Identified potential scaling bottlenecks

###  Readiness
- [ ] Can discuss scalability trade-offs confidently
- [ ] Can design reliable systems with proper redundancy
- [ ] Can choose appropriate consistency models
- [ ] Can estimate performance and capacity requirements
- [ ] Can handle follow-up questions about optimization

---

## 🚀 Next Steps

Once you've completed 1, you're ready for:
- **2**: Architecture Patterns & Design
- **Advanced Topics**: Microservices decomposition, event-driven architecture
- **Hands-on Projects**: Design a new microservice for Go Coffee

**Congratulations on mastering system design fundamentals! 🎉**
