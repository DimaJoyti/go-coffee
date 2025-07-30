# 💾 3: Data Management & Storage

## 📋 Overview

Master data storage solutions through Go Coffee's comprehensive data architecture. This covers database design, caching strategies, data consistency, transactions, and storage optimization.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Design scalable database schemas
- Implement effective caching strategies
- Understand data consistency models
- Master transaction management
- Optimize storage performance
- Analyze Go Coffee's data architecture

---

## 📖 3.1 Database Design & Schema Modeling

### Core Concepts

#### Relational vs NoSQL Decision Matrix
- **ACID Requirements**: Strong consistency needs → SQL
- **Scale Requirements**: Massive horizontal scale → NoSQL
- **Query Complexity**: Complex joins and analytics → SQL
- **Schema Flexibility**: Evolving data models → NoSQL
- **Consistency vs Availability**: CP vs AP in CAP theorem

#### Schema Design Patterns
- **Normalization**: Reduce data redundancy, ensure consistency
- **Denormalization**: Optimize for read performance
- **Partitioning**: Horizontal scaling through data distribution
- **Indexing**: Query performance optimization

### 🔍 Go Coffee Analysis

#### Study Database Architecture

<augment_code_snippet path="crypto-wallet/db/migrations/001_initial_schema.sql" mode="EXCERPT">
````sql
-- Coffee shop management tables
CREATE TABLE coffee_shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    phone VARCHAR(20),
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product catalog with pricing
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES coffee_shops(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    base_price DECIMAL(10, 2) NOT NULL,
    available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders with crypto payment support
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES coffee_shops(id),
    customer_id UUID,
    total_amount_usd DECIMAL(10, 2) NOT NULL,
    payment_currency VARCHAR(10),
    payment_amount DECIMAL(20, 8),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexing strategy for performance
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_shop_id ON orders(shop_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_products_shop_category ON products(shop_id, category);
````
</augment_code_snippet>

#### Analyze Multi-Database Strategy

**PostgreSQL (Primary OLTP)**:
- Transactional data (orders, payments, users)
- Strong consistency requirements
- Complex relational queries
- ACID compliance

**Redis (Caching & Sessions)**:
- Session management
- Real-time data caching
- Pub/sub for real-time updates
- High-performance key-value operations

**Blockchain (Immutable Ledger)**:
- Payment transactions
- Smart contract state
- Audit trails
- Decentralized consensus

### 🛠️ Hands-on Exercise 3.1: Design Optimized Schema

#### Step 1: Create Performance-Optimized Order Schema
```sql
-- Enhanced order schema with partitioning and indexing
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    order_number BIGSERIAL,
    
    -- Order details
    items JSONB NOT NULL,
    total_amount_usd DECIMAL(10, 2) NOT NULL,
    tax_amount DECIMAL(10, 2) DEFAULT 0,
    tip_amount DECIMAL(10, 2) DEFAULT 0,
    
    -- Payment information
    payment_method VARCHAR(50) NOT NULL,
    payment_currency VARCHAR(10),
    payment_amount DECIMAL(20, 8),
    payment_tx_hash VARCHAR(66),
    
    -- Status and timestamps
    status order_status DEFAULT 'pending',
    estimated_ready_time TIMESTAMP,
    actual_ready_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_amounts CHECK (total_amount_usd > 0),
    CONSTRAINT valid_payment CHECK (
        (payment_method = 'crypto' AND payment_currency IS NOT NULL) OR
        (payment_method != 'crypto' AND payment_currency IS NULL)
    )
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for better performance
CREATE TABLE orders_2024_01 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE orders_2024_02 PARTITION OF orders
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Optimized indexes
CREATE INDEX idx_orders_customer_status ON orders(customer_id, status) 
    WHERE status IN ('pending', 'confirmed', 'in_progress');
CREATE INDEX idx_orders_shop_created ON orders(shop_id, created_at DESC);
CREATE INDEX idx_orders_payment_hash ON orders(payment_tx_hash) 
    WHERE payment_tx_hash IS NOT NULL;

-- GIN index for JSONB items
CREATE INDEX idx_orders_items_gin ON orders USING GIN(items);
```

#### Step 2: Implement Repository with Optimized Queries
```go
// internal/order/infrastructure/database/postgres_order_repository.go
package database

type PostgresOrderRepository struct {
    db *sql.DB
}

func (r *PostgresOrderRepository) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit int, offset int) ([]*entities.Order, error) {
    query := `
        SELECT id, shop_id, customer_id, items, total_amount_usd, 
               payment_method, payment_currency, status, created_at
        FROM orders 
        WHERE customer_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3`
    
    rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to query orders: %w", err)
    }
    defer rows.Close()
    
    var orders []*entities.Order
    for rows.Next() {
        order := &entities.Order{}
        var itemsJSON []byte
        
        err := rows.Scan(
            &order.ID, &order.ShopID, &order.CustomerID,
            &itemsJSON, &order.TotalAmountUSD,
            &order.PaymentMethod, &order.PaymentCurrency,
            &order.Status, &order.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        
        // Parse JSONB items
        if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
            return nil, fmt.Errorf("failed to parse items: %w", err)
        }
        
        orders = append(orders, order)
    }
    
    return orders, nil
}

func (r *PostgresOrderRepository) GetOrdersByShopAndDateRange(ctx context.Context, shopID uuid.UUID, startDate, endDate time.Time) ([]*entities.Order, error) {
    // Optimized query using partition pruning
    query := `
        SELECT id, customer_id, items, total_amount_usd, status, created_at
        FROM orders 
        WHERE shop_id = $1 
          AND created_at >= $2 
          AND created_at < $3
        ORDER BY created_at DESC`
    
    rows, err := r.db.QueryContext(ctx, query, shopID, startDate, endDate)
    if err != nil {
        return nil, fmt.Errorf("failed to query orders by date range: %w", err)
    }
    defer rows.Close()
    
    // Process results...
    return orders, nil
}
```

### 💡 Practice Question 3.1
**"Design a database schema for Go Coffee that can handle 1M orders per day with complex analytics queries."**

**Solution Framework:**
1. **Partitioning Strategy**: Partition by date for time-series queries
2. **Indexing Strategy**: Composite indexes for common query patterns
3. **Read Replicas**: Separate analytics workload from OLTP
4. **Data Archiving**: Move old data to cheaper storage
5. **Materialized Views**: Pre-compute common aggregations

---

## 📖 3.2 Caching Strategies & Implementation

### Core Concepts

#### Cache Patterns
- **Cache-Aside (Lazy Loading)**: Load data on cache miss
- **Write-Through**: Write to cache and database simultaneously
- **Write-Behind (Write-Back)**: Write to cache first, database later
- **Refresh-Ahead**: Proactively refresh before expiration

#### Cache Levels
- **Browser Cache**: Client-side caching
- **CDN**: Geographic content distribution
- **Load Balancer**: SSL termination, routing cache
- **Application Cache**: In-memory application data
- **Database Cache**: Query result caching

#### Cache Invalidation
- **TTL (Time To Live)**: Automatic expiration
- **Event-Based**: Invalidate on data changes
- **Manual**: Explicit cache clearing
- **Write-Through**: Immediate consistency

### 🔍 Go Coffee Analysis

#### Study Redis Caching Implementation

<augment_code_snippet path="pkg/redis/client.go" mode="EXCERPT">
````go
type RedisClient struct {
    client *redis.Client
    logger *slog.Logger
}

func (r *RedisClient) GetCoffeeMenu(ctx context.Context, shopID string) (*models.Menu, error) {
    cacheKey := fmt.Sprintf("menu:shop:%s", shopID)
    
    // Try cache first (Cache-Aside pattern)
    cached, err := r.client.Get(ctx, cacheKey).Result()
    if err == nil {
        var menu models.Menu
        if err := json.Unmarshal([]byte(cached), &menu); err == nil {
            r.logger.Debug("Cache hit for menu", "shop_id", shopID)
            return &menu, nil
        }
    }
    
    // Cache miss - would load from database
    r.logger.Debug("Cache miss for menu", "shop_id", shopID)
    return nil, errors.New("cache miss")
}

func (r *RedisClient) SetCoffeeMenu(ctx context.Context, shopID string, menu *models.Menu, ttl time.Duration) error {
    cacheKey := fmt.Sprintf("menu:shop:%s", shopID)
    
    menuJSON, err := json.Marshal(menu)
    if err != nil {
        return fmt.Errorf("failed to marshal menu: %w", err)
    }
    
    return r.client.Set(ctx, cacheKey, menuJSON, ttl).Err()
}
````
</augment_code_snippet>

#### Analyze Multi-Level Caching Strategy

**Level 1: Application Cache (In-Memory)**
```go
// pkg/cache/memory_cache.go
type MemoryCache struct {
    cache map[string]cacheItem
    mutex sync.RWMutex
    ttl   time.Duration
}

type cacheItem struct {
    data      interface{}
    expiresAt time.Time
}

func (mc *MemoryCache) Get(key string) (interface{}, bool) {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()
    
    item, exists := mc.cache[key]
    if !exists || time.Now().After(item.expiresAt) {
        return nil, false
    }
    
    return item.data, true
}
```

**Level 2: Redis Distributed Cache**
```go
// pkg/cache/redis_cache.go
type RedisCache struct {
    client *redis.Client
}

func (rc *RedisCache) GetWithFallback(ctx context.Context, key string, fallback func() (interface{}, error)) (interface{}, error) {
    // Try Redis first
    cached, err := rc.client.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }
    
    // Fallback to data source
    data, err := fallback()
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    dataJSON, _ := json.Marshal(data)
    rc.client.Set(ctx, key, dataJSON, time.Hour)
    
    return data, nil
}
```

### 🛠️ Hands-on Exercise 3.2: Implement Advanced Caching

#### Step 1: Create Cache-Aside Pattern with Write-Through
```go
// internal/menu/infrastructure/cache/menu_cache.go
package cache

type MenuCacheService struct {
    redisClient *redis.Client
    memCache    *cache.MemoryCache
    repository  repositories.MenuRepository
    logger      *slog.Logger
}

func (mcs *MenuCacheService) GetMenu(ctx context.Context, shopID uuid.UUID) (*entities.Menu, error) {
    cacheKey := fmt.Sprintf("menu:%s", shopID.String())
    
    // Level 1: Check memory cache
    if data, found := mcs.memCache.Get(cacheKey); found {
        mcs.logger.Debug("Memory cache hit", "shop_id", shopID)
        return data.(*entities.Menu), nil
    }
    
    // Level 2: Check Redis cache
    cached, err := mcs.redisClient.Get(ctx, cacheKey).Result()
    if err == nil {
        var menu entities.Menu
        if err := json.Unmarshal([]byte(cached), &menu); err == nil {
            mcs.logger.Debug("Redis cache hit", "shop_id", shopID)
            
            // Populate memory cache
            mcs.memCache.Set(cacheKey, &menu, 5*time.Minute)
            return &menu, nil
        }
    }
    
    // Level 3: Load from database
    mcs.logger.Debug("Cache miss - loading from database", "shop_id", shopID)
    menu, err := mcs.repository.GetByShopID(ctx, shopID)
    if err != nil {
        return nil, fmt.Errorf("failed to load menu from database: %w", err)
    }
    
    // Populate both cache levels
    go mcs.populateCache(ctx, cacheKey, menu)
    
    return menu, nil
}

func (mcs *MenuCacheService) UpdateMenu(ctx context.Context, menu *entities.Menu) error {
    // Write-through pattern: update database and cache
    if err := mcs.repository.Update(ctx, menu); err != nil {
        return fmt.Errorf("failed to update menu in database: %w", err)
    }
    
    cacheKey := fmt.Sprintf("menu:%s", menu.ShopID.String())
    
    // Update both cache levels
    mcs.memCache.Set(cacheKey, menu, 5*time.Minute)
    
    menuJSON, _ := json.Marshal(menu)
    mcs.redisClient.Set(ctx, cacheKey, menuJSON, time.Hour)
    
    // Invalidate related caches
    mcs.invalidateRelatedCaches(ctx, menu.ShopID)
    
    return nil
}

func (mcs *MenuCacheService) invalidateRelatedCaches(ctx context.Context, shopID uuid.UUID) {
    // Invalidate shop-related caches
    patterns := []string{
        fmt.Sprintf("menu:%s", shopID.String()),
        fmt.Sprintf("shop:popular_items:%s", shopID.String()),
        fmt.Sprintf("shop:categories:%s", shopID.String()),
    }
    
    for _, pattern := range patterns {
        mcs.memCache.Delete(pattern)
        mcs.redisClient.Del(ctx, pattern)
    }
}
```

#### Step 2: Implement Cache Warming Strategy
```go
// internal/menu/application/services/cache_warming_service.go
package services

type CacheWarmingService struct {
    menuCache   *cache.MenuCacheService
    shopRepo    repositories.ShopRepository
    analytics   analytics.Service
    logger      *slog.Logger
}

func (cws *CacheWarmingService) WarmPopularMenus(ctx context.Context) error {
    // Get popular shops from analytics
    popularShops, err := cws.analytics.GetPopularShops(ctx, 50)
    if err != nil {
        return fmt.Errorf("failed to get popular shops: %w", err)
    }
    
    // Warm cache for popular shops
    for _, shopID := range popularShops {
        go func(id uuid.UUID) {
            if _, err := cws.menuCache.GetMenu(ctx, id); err != nil {
                cws.logger.Error("Failed to warm cache for shop", "shop_id", id, "error", err)
            }
        }(shopID)
    }
    
    return nil
}

func (cws *CacheWarmingService) ScheduledWarmup() {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            if err := cws.WarmPopularMenus(ctx); err != nil {
                cws.logger.Error("Cache warming failed", "error", err)
            }
        }
    }
}
```

### 💡 Practice Question 3.2
**"Design a caching strategy for Go Coffee that can handle 10K requests/second with 99% cache hit rate."**

**Solution Framework:**
1. **Multi-Level Caching**: Memory → Redis → Database
2. **Cache Warming**: Proactively load popular data
3. **Smart Invalidation**: Event-driven cache updates
4. **Cache Partitioning**: Distribute load across cache nodes
5. **Monitoring**: Track hit rates and performance metrics

---

## 📖 3.3 Data Consistency & Transactions

### Core Concepts

#### ACID Properties
- **Atomicity**: All operations succeed or all fail
- **Consistency**: Data integrity constraints maintained
- **Isolation**: Concurrent transactions don't interfere
- **Durability**: Committed changes persist

#### Isolation Levels
- **Read Uncommitted**: Dirty reads possible
- **Read Committed**: No dirty reads
- **Repeatable Read**: No phantom reads within transaction
- **Serializable**: Full isolation, no anomalies

#### Distributed Transactions
- **Two-Commit (2PC)**: Coordinator ensures atomicity
- **Saga Pattern**: Compensating actions for rollback
- **Event Sourcing**: Append-only event log
- **Eventual Consistency**: Systems converge over time

### 🔍 Go Coffee Analysis

#### Study Transaction Management

<augment_code_snippet path="crypto-wallet/internal/database/postgres.go" mode="EXCERPT">
````go
func (db *PostgresDB) CreateOrderWithPayment(ctx context.Context, order *Order, payment *Payment) error {
    tx, err := db.conn.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable, // Strongest isolation
    })
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback() // Ensure rollback on error
    
    // Create order
    _, err = tx.ExecContext(ctx, 
        "INSERT INTO orders (id, customer_id, total, status) VALUES ($1, $2, $3, $4)",
        order.ID, order.CustomerID, order.Total, "pending")
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
    
    // Update inventory
    for _, item := range order.Items {
        _, err = tx.ExecContext(ctx,
            "UPDATE inventory SET quantity = quantity - $1 WHERE product_id = $2",
            item.Quantity, item.ProductID)
        if err != nil {
            return fmt.Errorf("failed to update inventory: %w", err)
        }
    }
    
    // Commit transaction - ensures atomicity
    return tx.Commit()
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 3.3: Implement Saga Pattern

#### Step 1: Create Saga Orchestrator
```go
// internal/order/application/sagas/order_saga.go
package sagas

type OrderSaga struct {
    orderRepo     repositories.OrderRepository
    paymentRepo   repositories.PaymentRepository
    inventoryRepo repositories.InventoryRepository
    eventBus      events.EventBus
    logger        *slog.Logger
}

type SagaStep struct {
    Name        string
    Execute     func(ctx context.Context, data interface{}) error
    Compensate  func(ctx context.Context, data interface{}) error
}

func (os *OrderSaga) ProcessOrder(ctx context.Context, orderData *OrderData) error {
    sagaID := uuid.New()
    
    steps := []SagaStep{
        {
            Name:       "CreateOrder",
            Execute:    os.createOrder,
            Compensate: os.cancelOrder,
        },
        {
            Name:       "ReserveInventory",
            Execute:    os.reserveInventory,
            Compensate: os.releaseInventory,
        },
        {
            Name:       "ProcessPayment",
            Execute:    os.processPayment,
            Compensate: os.refundPayment,
        },
        {
            Name:       "ConfirmOrder",
            Execute:    os.confirmOrder,
            Compensate: os.cancelOrder,
        },
    }
    
    // Execute saga steps
    executedSteps := []SagaStep{}
    
    for _, step := range steps {
        os.logger.Info("Executing saga step", "saga_id", sagaID, "step", step.Name)
        
        if err := step.Execute(ctx, orderData); err != nil {
            os.logger.Error("Saga step failed", "saga_id", sagaID, "step", step.Name, "error", err)
            
            // Compensate in reverse order
            os.compensate(ctx, executedSteps, orderData)
            return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
        }
        
        executedSteps = append(executedSteps, step)
    }
    
    os.logger.Info("Saga completed successfully", "saga_id", sagaID)
    return nil
}

func (os *OrderSaga) compensate(ctx context.Context, steps []SagaStep, data *OrderData) {
    // Execute compensation in reverse order
    for i := len(steps) - 1; i >= 0; i-- {
        step := steps[i]
        os.logger.Info("Compensating saga step", "step", step.Name)
        
        if err := step.Compensate(ctx, data); err != nil {
            os.logger.Error("Compensation failed", "step", step.Name, "error", err)
            // In production, this would trigger alerts and manual intervention
        }
    }
}

func (os *OrderSaga) createOrder(ctx context.Context, data interface{}) error {
    orderData := data.(*OrderData)
    
    order := &entities.Order{
        ID:         orderData.OrderID,
        CustomerID: orderData.CustomerID,
        Items:      orderData.Items,
        Status:     entities.OrderStatusPending,
        CreatedAt:  time.Now(),
    }
    
    return os.orderRepo.Create(ctx, order)
}

func (os *OrderSaga) reserveInventory(ctx context.Context, data interface{}) error {
    orderData := data.(*OrderData)
    
    for _, item := range orderData.Items {
        if err := os.inventoryRepo.Reserve(ctx, item.ProductID, item.Quantity); err != nil {
            return fmt.Errorf("failed to reserve inventory for product %s: %w", item.ProductID, err)
        }
    }
    
    return nil
}

func (os *OrderSaga) processPayment(ctx context.Context, data interface{}) error {
    orderData := data.(*OrderData)
    
    payment := &entities.Payment{
        ID:       uuid.New(),
        OrderID:  orderData.OrderID,
        Amount:   orderData.TotalAmount,
        Currency: orderData.PaymentCurrency,
        Status:   entities.PaymentStatusPending,
    }
    
    return os.paymentRepo.ProcessPayment(ctx, payment)
}
```

### 💡 Practice Question 3.3
**"How would you ensure data consistency across Go Coffee's microservices when processing an order?"**

**Solution Framework:**
1. **Saga Pattern**: Choreography or orchestration for distributed transactions
2. **Event Sourcing**: Append-only events for audit and consistency
3. **Eventual Consistency**: Accept temporary inconsistency for availability
4. **Compensation**: Rollback through compensating actions
5. **Monitoring**: Track saga execution and failure rates

---

## 🎯 3 Completion Checklist

### Knowledge Mastery
- [ ] Understand database design patterns and trade-offs
- [ ] Can implement multi-level caching strategies
- [ ] Know data consistency models and transaction patterns
- [ ] Understand storage optimization techniques
- [ ] Can design for data scalability and performance

### Practical Skills
- [ ] Can design optimized database schemas
- [ ] Can implement cache-aside and write-through patterns
- [ ] Can handle distributed transactions with sagas
- [ ] Can optimize queries and indexing strategies
- [ ] Can design data partitioning and sharding

### Go Coffee Analysis
- [ ] Analyzed multi-database architecture decisions
- [ ] Studied Redis caching implementation patterns
- [ ] Examined transaction management strategies
- [ ] Understood data consistency approaches
- [ ] Identified storage optimization techniques

###  Readiness
- [ ] Can discuss SQL vs NoSQL trade-offs
- [ ] Can design caching strategies for scale
- [ ] Can handle data consistency challenges
- [ ] Can optimize database performance
- [ ] Can design for data reliability and durability

---

## 🚀 Next Steps

Ready for **4: Communication & APIs**:
- REST API design principles
- gRPC for high-performance communication
- Message queue patterns
- Real-time communication
- API gateway patterns

**Excellent progress on mastering data management! 🎉**
