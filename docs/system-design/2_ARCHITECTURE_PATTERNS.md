# 🏛️ 2: Architecture Patterns & Design

## 📋 Overview

Master architectural patterns through Go Coffee's real-world implementations. This covers microservices, event-driven architecture, clean architecture, and service communication patterns.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Understand microservices decomposition strategies
- Master event-driven architecture patterns
- Apply clean architecture principles
- Design effective service communication
- Analyze Go Coffee's architectural decisions

---

## 📖 2.1 Microservices Architecture

### Core Concepts

#### Service Decomposition Strategies
- **Domain-Driven Design (DDD)**: Bounded contexts define service boundaries
- **Business Capability**: Services organized around business functions
- **Data Ownership**: Each service owns its data
- **Team Structure**: Conway's Law - services mirror team structure

#### Service Boundaries
- **High Cohesion**: Related functionality grouped together
- **Loose Coupling**: Minimal dependencies between services
- **Single Responsibility**: Each service has one clear purpose
- **Autonomous Teams**: Services can be developed independently

### 🔍 Go Coffee Analysis

#### Study the Microservices Architecture

<augment_code_snippet path="Makefile" mode="EXCERPT">
````makefile
# Service definitions showing microservices structure
CORE_SERVICES := api-gateway producer consumer streams
WEB3_SERVICES := auth-service order-service kitchen-service payment-service
AI_SERVICES := ai-arbitrage-service ai-order-service
INFRASTRUCTURE_SERVICES := security-gateway user-gateway
LLM_SERVICES := llm-orchestrator llm-orchestrator-simple

ALL_SERVICES := $(CORE_SERVICES) $(WEB3_SERVICES) $(AI_SERVICES) $(INFRASTRUCTURE_SERVICES) $(LLM_SERVICES)
````
</augment_code_snippet>

#### Analyze Service Boundaries in Go Coffee

**Core Coffee Domain:**
- **Producer Service**: Order intake and validation
- **Consumer Service**: Order processing and fulfillment
- **Kitchen Service**: Kitchen operations and staff management
- **API Gateway**: Request routing and load balancing

**Web3 Domain:**
- **Auth Service**: Authentication and authorization
- **Payment Service**: Crypto payment processing
- **Wallet Service**: Blockchain wallet management
- **DeFi Service**: Trading and yield farming

**AI Domain:**
- **AI Agents**: 9 specialized automation agents
- **LLM Orchestrator**: AI model coordination
- **AI Search**: Semantic search capabilities

### 🛠️ Hands-on Exercise 2.1: Design New Microservice

#### Step 1: Create Loyalty Service Structure
```bash
# Create new microservice following Go Coffee patterns
mkdir -p internal/loyalty
cd internal/loyalty

# Follow clean architecture structure
mkdir -p domain application infrastructure transport
mkdir -p domain/{entities,repositories,services}
mkdir -p application/{usecases,dto}
mkdir -p infrastructure/{database,cache,external}
mkdir -p transport/{http,grpc}
```

#### Step 2: Define Domain Entities
```go
// internal/loyalty/domain/entities/loyalty_account.go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type LoyaltyAccount struct {
    ID          uuid.UUID `json:"id"`
    UserID      uuid.UUID `json:"user_id"`
    Points      int       `json:"points"`
    Tier        Tier      `json:"tier"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Tier string

const (
    TierBronze   Tier = "bronze"
    TierSilver   Tier = "silver"
    TierGold     Tier = "gold"
    TierPlatinum Tier = "platinum"
)

type PointTransaction struct {
    ID          uuid.UUID       `json:"id"`
    AccountID   uuid.UUID       `json:"account_id"`
    Points      int             `json:"points"`
    Type        TransactionType `json:"type"`
    Description string          `json:"description"`
    OrderID     *uuid.UUID      `json:"order_id,omitempty"`
    CreatedAt   time.Time       `json:"created_at"`
}

type TransactionType string

const (
    TransactionEarn   TransactionType = "earn"
    TransactionRedeem TransactionType = "redeem"
    TransactionExpire TransactionType = "expire"
)
```

#### Step 3: Define Repository Interface
```go
// internal/loyalty/domain/repositories/loyalty_repository.go
package repositories

import (
    "context"
    "github.com/google/uuid"
    "go-coffee/internal/loyalty/domain/entities"
)

type LoyaltyRepository interface {
    CreateAccount(ctx context.Context, account *entities.LoyaltyAccount) error
    GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*entities.LoyaltyAccount, error)
    UpdateAccount(ctx context.Context, account *entities.LoyaltyAccount) error
    
    AddPointTransaction(ctx context.Context, transaction *entities.PointTransaction) error
    GetTransactionHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]*entities.PointTransaction, error)
    
    GetAccountsByTier(ctx context.Context, tier entities.Tier) ([]*entities.LoyaltyAccount, error)
}
```

#### Step 4: Implement Use Cases
```go
// internal/loyalty/application/usecases/loyalty_usecase.go
package usecases

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "go-coffee/internal/loyalty/domain/entities"
    "go-coffee/internal/loyalty/domain/repositories"
)

type LoyaltyUseCase struct {
    repo repositories.LoyaltyRepository
}

func NewLoyaltyUseCase(repo repositories.LoyaltyRepository) *LoyaltyUseCase {
    return &LoyaltyUseCase{repo: repo}
}

func (uc *LoyaltyUseCase) EarnPoints(ctx context.Context, userID uuid.UUID, points int, orderID uuid.UUID) error {
    // Get or create loyalty account
    account, err := uc.repo.GetAccountByUserID(ctx, userID)
    if err != nil {
        // Create new account if not exists
        account = &entities.LoyaltyAccount{
            ID:        uuid.New(),
            UserID:    userID,
            Points:    0,
            Tier:      entities.TierBronze,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }
        if err := uc.repo.CreateAccount(ctx, account); err != nil {
            return fmt.Errorf("failed to create loyalty account: %w", err)
        }
    }
    
    // Add points
    account.Points += points
    account.UpdatedAt = time.Now()
    
    // Update tier based on points
    account.Tier = uc.calculateTier(account.Points)
    
    // Save account
    if err := uc.repo.UpdateAccount(ctx, account); err != nil {
        return fmt.Errorf("failed to update account: %w", err)
    }
    
    // Record transaction
    transaction := &entities.PointTransaction{
        ID:          uuid.New(),
        AccountID:   account.ID,
        Points:      points,
        Type:        entities.TransactionEarn,
        Description: fmt.Sprintf("Earned from order %s", orderID.String()),
        OrderID:     &orderID,
        CreatedAt:   time.Now(),
    }
    
    return uc.repo.AddPointTransaction(ctx, transaction)
}

func (uc *LoyaltyUseCase) calculateTier(points int) entities.Tier {
    switch {
    case points >= 10000:
        return entities.TierPlatinum
    case points >= 5000:
        return entities.TierGold
    case points >= 1000:
        return entities.TierSilver
    default:
        return entities.TierBronze
    }
}
```

### 💡 Practice Question 2.1
**"How would you decompose a monolithic coffee shop management system into microservices?"**

**Solution Framework:**
1. **Identify Business Capabilities**
   - Order Management
   - Inventory Management
   - Customer Management
   - Payment Processing
   - Kitchen Operations

2. **Define Service Boundaries**
   - Each service owns its data
   - Minimal cross-service transactions
   - Clear API contracts

3. **Consider Data Consistency**
   - Eventual consistency between services
   - Saga pattern for distributed transactions
   - Event sourcing for audit trails

---

## 📖 2.2 Event-Driven Architecture

### Core Concepts

#### Event Sourcing
- **Events as Source of Truth**: Store events, not current state
- **Event Store**: Append-only log of events
- **Event Replay**: Rebuild state from events
- **Temporal Queries**: Query state at any point in time

#### CQRS (Command Query Responsibility Segregation)
- **Separate Models**: Different models for reads and writes
- **Command Side**: Handles business operations
- **Query Side**: Optimized for reads
- **Eventual Consistency**: Query side eventually consistent

#### Saga Patterns
- **Choreography**: Services coordinate through events
- **Orchestration**: Central coordinator manages workflow
- **Compensation**: Rollback through compensating actions

### 🔍 Go Coffee Analysis

#### Study Event-Driven Communication

<augment_code_snippet path="producer/kafka/producer.go" mode="EXCERPT">
````go
type Producer struct {
    writer *kafka.Writer
    config *config.Config
    logger *slog.Logger
}

func (p *Producer) PublishOrder(order *models.Order) error {
    // Create event message
    message := kafka.Message{
        Key:   []byte(order.ID),
        Value: orderJSON,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte("order.created")},
            {Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
        },
    }
    
    // Publish to Kafka
    return p.writer.WriteMessages(context.Background(), message)
}
````
</augment_code_snippet>

#### Analyze AI Agent Event Processing

<augment_code_snippet path="ai-agents/internal/messaging/kafka_consumer.go" mode="EXCERPT">
````go
func (c *Consumer) ProcessEvents(ctx context.Context) error {
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            return err
        }
        
        // Extract event type from headers
        eventType := c.getEventType(msg.Headers)
        
        // Route to appropriate handler
        switch eventType {
        case "order.created":
            c.handleOrderCreated(msg.Value)
        case "inventory.low":
            c.handleInventoryLow(msg.Value)
        case "feedback.received":
            c.handleFeedbackReceived(msg.Value)
        }
    }
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 2.2: Implement Event Sourcing

#### Step 1: Create Event Store
```go
// pkg/eventstore/event_store.go
package eventstore

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
)

type Event struct {
    ID          uuid.UUID   `json:"id"`
    StreamID    string      `json:"stream_id"`
    EventType   string      `json:"event_type"`
    EventData   interface{} `json:"event_data"`
    Metadata    map[string]interface{} `json:"metadata"`
    Version     int         `json:"version"`
    Timestamp   time.Time   `json:"timestamp"`
}

type EventStore interface {
    SaveEvents(ctx context.Context, streamID string, events []Event, expectedVersion int) error
    GetEvents(ctx context.Context, streamID string, fromVersion int) ([]Event, error)
    GetAllEvents(ctx context.Context, fromTimestamp time.Time) ([]Event, error)
}

type PostgresEventStore struct {
    db *sql.DB
}

func (es *PostgresEventStore) SaveEvents(ctx context.Context, streamID string, events []Event, expectedVersion int) error {
    tx, err := es.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Check current version
    var currentVersion int
    err = tx.QueryRowContext(ctx, 
        "SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1", 
        streamID).Scan(&currentVersion)
    if err != nil {
        return err
    }
    
    if currentVersion != expectedVersion {
        return fmt.Errorf("concurrency conflict: expected version %d, got %d", 
            expectedVersion, currentVersion)
    }
    
    // Insert events
    for i, event := range events {
        event.Version = expectedVersion + i + 1
        event.Timestamp = time.Now()
        
        eventData, _ := json.Marshal(event.EventData)
        metadata, _ := json.Marshal(event.Metadata)
        
        _, err = tx.ExecContext(ctx, `
            INSERT INTO events (id, stream_id, event_type, event_data, metadata, version, timestamp)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`,
            event.ID, event.StreamID, event.EventType, eventData, metadata, event.Version, event.Timestamp)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

#### Step 2: Implement Order Aggregate with Event Sourcing
```go
// internal/order/domain/order_aggregate.go
package domain

type OrderAggregate struct {
    ID       uuid.UUID
    Version  int
    events   []eventstore.Event
    
    // Current state
    CustomerID   uuid.UUID
    Items        []OrderItem
    Status       OrderStatus
    TotalAmount  decimal.Decimal
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (o *OrderAggregate) CreateOrder(customerID uuid.UUID, items []OrderItem) error {
    if o.ID != uuid.Nil {
        return errors.New("order already exists")
    }
    
    // Validate business rules
    if len(items) == 0 {
        return errors.New("order must have at least one item")
    }
    
    // Calculate total
    total := decimal.Zero
    for _, item := range items {
        total = total.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
    }
    
    // Create event
    event := eventstore.Event{
        ID:        uuid.New(),
        StreamID:  o.ID.String(),
        EventType: "OrderCreated",
        EventData: OrderCreatedEvent{
            OrderID:     o.ID,
            CustomerID:  customerID,
            Items:       items,
            TotalAmount: total,
            CreatedAt:   time.Now(),
        },
    }
    
    // Apply event
    o.applyEvent(event)
    o.events = append(o.events, event)
    
    return nil
}

func (o *OrderAggregate) applyEvent(event eventstore.Event) {
    switch event.EventType {
    case "OrderCreated":
        data := event.EventData.(OrderCreatedEvent)
        o.ID = data.OrderID
        o.CustomerID = data.CustomerID
        o.Items = data.Items
        o.Status = OrderStatusPending
        o.TotalAmount = data.TotalAmount
        o.CreatedAt = data.CreatedAt
        o.UpdatedAt = data.CreatedAt
    case "OrderConfirmed":
        o.Status = OrderStatusConfirmed
        o.UpdatedAt = time.Now()
    case "OrderCompleted":
        o.Status = OrderStatusCompleted
        o.UpdatedAt = time.Now()
    }
    
    o.Version++
}
```

### 💡 Practice Question 2.2
**"Design an event-driven system for real-time order tracking in Go Coffee."**

**Solution Framework:**
1. **Event Types**
   - OrderCreated, OrderConfirmed, OrderInProgress, OrderReady, OrderCompleted
   
2. **Event Flow**
   - Producer → Kafka → Multiple Consumers (Kitchen, Notifications, Analytics)
   
3. **State Management**
   - Event sourcing for order history
   - CQRS for read models
   - Real-time updates via WebSockets

---

## 📖 2.3 Clean Architecture

### Core Concepts

#### Dependency Inversion
- **High-level modules** don't depend on low-level modules
- **Both depend on abstractions** (interfaces)
- **Abstractions don't depend on details**
- **Details depend on abstractions**

#### Layered Architecture
- **Domain Layer**: Business entities and rules
- **Application Layer**: Use cases and business logic
- **Infrastructure Layer**: External concerns (DB, APIs)
- **Transport Layer**: HTTP, gRPC, CLI interfaces

#### Hexagonal Architecture (Ports and Adapters)
- **Core**: Business logic in the center
- **Ports**: Interfaces for external communication
- **Adapters**: Implementations of ports
- **Isolation**: Core isolated from external concerns

### 🔍 Go Coffee Analysis

#### Study Clean Architecture Implementation

<augment_code_snippet path="internal/kitchen/README.md" mode="EXCERPT">
````text
internal/kitchen/
├── domain/           # Business entities and rules
├── application/      # Use cases and business logic
├── infrastructure/   # External concerns (Redis, AI)
├── transport/        # API layer (gRPC, HTTP, WebSocket)
├── integration/      # Cross-service communication
└── config/          # Configuration management
````
</augment_code_snippet>

#### Analyze Domain Layer Structure

<augment_code_snippet path="internal/kitchen/domain/entities/kitchen.go" mode="EXCERPT">
````go
// Domain entities are pure business objects
type Kitchen struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Location    string    `json:"location"`
    Capacity    int       `json:"capacity"`
    Status      Status    `json:"status"`
    Equipment   []Equipment `json:"equipment"`
    Staff       []Staff   `json:"staff"`
}

// Business rules are enforced in domain methods
func (k *Kitchen) AssignOrder(order *Order) error {
    if k.Status != StatusOperational {
        return errors.New("kitchen is not operational")
    }
    
    if k.GetCurrentLoad() >= k.Capacity {
        return errors.New("kitchen is at full capacity")
    }
    
    return nil
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 2.3: Refactor to Clean Architecture

#### Step 1: Create Domain Layer
```go
// internal/loyalty/domain/services/loyalty_service.go
package services

type LoyaltyDomainService struct{}

func (s *LoyaltyDomainService) CalculatePointsForOrder(orderAmount decimal.Decimal, tier entities.Tier) int {
    basePoints := int(orderAmount.IntPart())
    
    // Apply tier multiplier
    multiplier := s.getTierMultiplier(tier)
    return int(float64(basePoints) * multiplier)
}

func (s *LoyaltyDomainService) getTierMultiplier(tier entities.Tier) float64 {
    switch tier {
    case entities.TierPlatinum:
        return 2.0
    case entities.TierGold:
        return 1.5
    case entities.TierSilver:
        return 1.2
    default:
        return 1.0
    }
}

func (s *LoyaltyDomainService) CanRedeem(account *entities.LoyaltyAccount, points int) bool {
    return account.Points >= points && points > 0
}
```

#### Step 2: Create Application Layer
```go
// internal/loyalty/application/usecases/redeem_points_usecase.go
package usecases

type RedeemPointsUseCase struct {
    loyaltyRepo    repositories.LoyaltyRepository
    rewardRepo     repositories.RewardRepository
    domainService  *services.LoyaltyDomainService
    eventPublisher ports.EventPublisher
}

func (uc *RedeemPointsUseCase) Execute(ctx context.Context, req *dto.RedeemPointsRequest) (*dto.RedeemPointsResponse, error) {
    // Get loyalty account
    account, err := uc.loyaltyRepo.GetAccountByUserID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get loyalty account: %w", err)
    }
    
    // Get reward details
    reward, err := uc.rewardRepo.GetByID(ctx, req.RewardID)
    if err != nil {
        return nil, fmt.Errorf("failed to get reward: %w", err)
    }
    
    // Check if redemption is allowed (domain logic)
    if !uc.domainService.CanRedeem(account, reward.PointsCost) {
        return nil, errors.New("insufficient points for redemption")
    }
    
    // Deduct points
    account.Points -= reward.PointsCost
    account.UpdatedAt = time.Now()
    
    // Save changes
    if err := uc.loyaltyRepo.UpdateAccount(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to update account: %w", err)
    }
    
    // Record transaction
    transaction := &entities.PointTransaction{
        ID:          uuid.New(),
        AccountID:   account.ID,
        Points:      -reward.PointsCost,
        Type:        entities.TransactionRedeem,
        Description: fmt.Sprintf("Redeemed: %s", reward.Name),
        CreatedAt:   time.Now(),
    }
    
    if err := uc.loyaltyRepo.AddPointTransaction(ctx, transaction); err != nil {
        return nil, fmt.Errorf("failed to record transaction: %w", err)
    }
    
    // Publish event
    event := events.PointsRedeemedEvent{
        UserID:    req.UserID,
        RewardID:  req.RewardID,
        Points:    reward.PointsCost,
        Timestamp: time.Now(),
    }
    
    uc.eventPublisher.Publish(ctx, "loyalty.points.redeemed", event)
    
    return &dto.RedeemPointsResponse{
        Success:        true,
        RemainingPoints: account.Points,
        RedemptionID:   transaction.ID,
    }, nil
}
```

#### Step 3: Create Infrastructure Layer
```go
// internal/loyalty/infrastructure/database/postgres_loyalty_repository.go
package database

type PostgresLoyaltyRepository struct {
    db *sql.DB
}

func (r *PostgresLoyaltyRepository) CreateAccount(ctx context.Context, account *entities.LoyaltyAccount) error {
    query := `
        INSERT INTO loyalty_accounts (id, user_id, points, tier, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)`
    
    _, err := r.db.ExecContext(ctx, query,
        account.ID, account.UserID, account.Points, account.Tier,
        account.CreatedAt, account.UpdatedAt)
    
    return err
}

func (r *PostgresLoyaltyRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*entities.LoyaltyAccount, error) {
    query := `
        SELECT id, user_id, points, tier, created_at, updated_at
        FROM loyalty_accounts
        WHERE user_id = $1`
    
    var account entities.LoyaltyAccount
    err := r.db.QueryRowContext(ctx, query, userID).Scan(
        &account.ID, &account.UserID, &account.Points, &account.Tier,
        &account.CreatedAt, &account.UpdatedAt)
    
    if err == sql.ErrNoRows {
        return nil, repositories.ErrAccountNotFound
    }
    
    return &account, err
}
```

#### Step 4: Create Transport Layer
```go
// internal/loyalty/transport/http/loyalty_handler.go
package http

type LoyaltyHandler struct {
    redeemUseCase *usecases.RedeemPointsUseCase
    earnUseCase   *usecases.EarnPointsUseCase
}

func (h *LoyaltyHandler) RedeemPoints(w http.ResponseWriter, r *http.Request) {
    var req dto.RedeemPointsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Extract user ID from JWT token
    userID, err := h.extractUserID(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    req.UserID = userID
    
    // Execute use case
    resp, err := h.redeemUseCase.Execute(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

### 💡 Practice Question 2.3
**"How would you refactor a monolithic Go Coffee service to follow clean architecture principles?"**

**Solution Framework:**
1. **Identify Layers**
   - Extract business logic to domain layer
   - Move use cases to application layer
   - Isolate external dependencies in infrastructure
   - Create clean transport layer

2. **Define Interfaces**
   - Repository interfaces in domain
   - Use case interfaces in application
   - External service interfaces (ports)

3. **Dependency Injection**
   - Inject dependencies through constructors
   - Use interfaces, not concrete types
   - Wire dependencies in main function

---

## 🎯 2 Completion Checklist

### Knowledge Mastery
- [ ] Understand microservices decomposition strategies
- [ ] Can design event-driven architectures
- [ ] Know clean architecture principles
- [ ] Understand service communication patterns
- [ ] Can analyze architectural trade-offs

### Practical Skills
- [ ] Can design new microservices following Go Coffee patterns
- [ ] Can implement event sourcing and CQRS
- [ ] Can refactor code to clean architecture
- [ ] Can design service APIs and contracts
- [ ] Can handle cross-service communication

### Go Coffee Analysis
- [ ] Analyzed microservices structure and boundaries
- [ ] Studied event-driven communication patterns
- [ ] Examined clean architecture implementation
- [ ] Understood service decomposition decisions
- [ ] Identified architectural patterns used

###  Readiness
- [ ] Can discuss microservices vs monolith trade-offs
- [ ] Can design event-driven systems
- [ ] Can explain clean architecture benefits
- [ ] Can handle service communication challenges
- [ ] Can design for scalability and maintainability

---

## 🚀 Next Steps

Ready for **3: Data Management & Storage**:
- Database design patterns
- Caching strategies
- Data consistency models
- Transaction management
- Storage optimization

**Excellent progress on mastering architecture patterns! 🎉**
