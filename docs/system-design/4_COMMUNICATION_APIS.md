# 🔗 4: Communication & APIs

## 📋 Overview

Master communication patterns and API design through Go Coffee's comprehensive communication architecture. This covers REST APIs, gRPC, message queues, WebSockets, and API gateway patterns.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Design scalable REST APIs with proper versioning
- Implement high-performance gRPC communication
- Master message queue patterns and event streaming
- Build real-time communication systems
- Analyze Go Coffee's communication architecture

---

## 📖 4.1 REST API Design & Best Practices

### Core Concepts

#### RESTful Principles
- **Resource-Based URLs**: Nouns, not verbs in endpoints
- **HTTP Methods**: GET, POST, PUT, DELETE for CRUD operations
- **Stateless**: Each request contains all necessary information
- **Cacheable**: Responses should be cacheable when appropriate
- **Uniform Interface**: Consistent API design patterns

#### API Versioning Strategies
- **URL Versioning**: `/api/v1/orders`, `/api/v2/orders`
- **Header Versioning**: `Accept: application/vnd.api+json;version=1`
- **Query Parameter**: `/api/orders?version=1`
- **Content Negotiation**: Different response formats

### 🔍 Go Coffee Analysis

#### Study REST API Implementation

<augment_code_snippet path="api-gateway/server/http_server.go" mode="EXCERPT">
````go
func (s *HTTPServer) setupRoutes() {
    mux := http.NewServeMux()
    
    // RESTful order endpoints
    mux.HandleFunc("GET /api/v1/orders", s.handleListOrders)
    mux.HandleFunc("POST /api/v1/orders", s.handleCreateOrder)
    mux.HandleFunc("GET /api/v1/orders/{id}", s.handleGetOrder)
    mux.HandleFunc("PUT /api/v1/orders/{id}", s.handleUpdateOrder)
    mux.HandleFunc("DELETE /api/v1/orders/{id}", s.handleCancelOrder)
    
    // Health and monitoring endpoints
    mux.HandleFunc("GET /health", s.handleHealth)
    mux.HandleFunc("GET /metrics", s.handleMetrics)
    
    // Apply middleware chain
    handler := s.corsMiddleware(
        s.authMiddleware(
            s.loggingMiddleware(
                s.rateLimitMiddleware(mux),
            ),
        ),
    )
    
    s.Server.Handler = handler
}
````
</augment_code_snippet>

#### Analyze API Response Patterns

<augment_code_snippet path="crypto-wallet/api/handlers/coffee_handler.go" mode="EXCERPT">
````go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type Meta struct {
    Timestamp   time.Time `json:"timestamp"`
    RequestID   string    `json:"request_id"`
    Version     string    `json:"version"`
    Pagination  *Pagination `json:"pagination,omitempty"`
}

func (h *CoffeeHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", 
            "Invalid request body", err.Error())
        return
    }
    
    // Validate request
    if err := h.validator.Validate(req); err != nil {
        h.respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", 
            "Request validation failed", err.Error())
        return
    }
    
    // Process order
    order, err := h.orderService.CreateOrder(r.Context(), &req)
    if err != nil {
        h.respondWithError(w, http.StatusInternalServerError, "ORDER_CREATION_FAILED", 
            "Failed to create order", err.Error())
        return
    }
    
    // Success response
    h.respondWithSuccess(w, http.StatusCreated, order, &Meta{
        Timestamp: time.Now(),
        RequestID: middleware.GetRequestID(r.Context()),
        Version:   "v1",
    })
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 4.1: Design Comprehensive REST API

#### Step 1: Create API Specification
```yaml
# api/openapi/coffee-api-v1.yaml
openapi: 3.0.3
info:
  title: Go Coffee API
  description: Comprehensive coffee ordering and management API
  version: 1.0.0
  contact:
    name: Go Coffee Team
    email: api@gocoffee.io

servers:
  - url: https://api.gocoffee.io/v1
    description: Production server
  - url: https://staging-api.gocoffee.io/v1
    description: Staging server

paths:
  /shops:
    get:
      summary: List coffee shops
      parameters:
        - name: lat
          in: query
          schema:
            type: number
            format: double
        - name: lng
          in: query
          schema:
            type: number
            format: double
        - name: radius
          in: query
          schema:
            type: integer
            default: 5000
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
      responses:
        '200':
          description: List of coffee shops
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ShopListResponse'

  /shops/{shopId}/menu:
    get:
      summary: Get shop menu
      parameters:
        - name: shopId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Shop menu
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MenuResponse'

  /orders:
    post:
      summary: Create new order
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Order created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrderResponse'

components:
  schemas:
    Shop:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        address:
          type: string
        location:
          $ref: '#/components/schemas/Location'
        rating:
          type: number
          format: double
          minimum: 0
          maximum: 5
        isOpen:
          type: boolean
        estimatedWaitTime:
          type: integer
          description: Wait time in minutes

    Location:
      type: object
      properties:
        latitude:
          type: number
          format: double
        longitude:
          type: number
          format: double

    MenuItem:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        description:
          type: string
        price:
          type: number
          format: double
        category:
          type: string
        available:
          type: boolean
        customizations:
          type: array
          items:
            $ref: '#/components/schemas/Customization'

    CreateOrderRequest:
      type: object
      required:
        - shopId
        - items
        - paymentMethod
      properties:
        shopId:
          type: string
          format: uuid
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
        paymentMethod:
          type: string
          enum: [card, crypto, cash]
        specialInstructions:
          type: string
          maxLength: 500
```

#### Step 2: Implement Advanced API Handler
```go
// internal/api/handlers/order_handler.go
package handlers

type OrderHandler struct {
    orderService    services.OrderService
    paymentService  services.PaymentService
    validator       *validator.Validate
    logger          *slog.Logger
    metrics         *metrics.Metrics
}

func NewOrderHandler(
    orderService services.OrderService,
    paymentService services.PaymentService,
    validator *validator.Validate,
    logger *slog.Logger,
    metrics *metrics.Metrics,
) *OrderHandler {
    return &OrderHandler{
        orderService:   orderService,
        paymentService: paymentService,
        validator:      validator,
        logger:         logger,
        metrics:        metrics,
    }
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    requestID := middleware.GetRequestID(ctx)
    
    // Start timing for metrics
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        h.metrics.RecordAPICall("POST", "/orders", duration)
    }()
    
    // Parse and validate request
    var req dto.CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Failed to decode request", "request_id", requestID, "error", err)
        h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON", 
            "Invalid JSON in request body")
        return
    }
    
    // Validate request structure
    if err := h.validator.Struct(req); err != nil {
        h.logger.Error("Request validation failed", "request_id", requestID, "error", err)
        h.respondWithValidationError(w, err)
        return
    }
    
    // Extract user ID from JWT token
    userID, err := middleware.GetUserID(ctx)
    if err != nil {
        h.logger.Error("Failed to extract user ID", "request_id", requestID, "error", err)
        h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", 
            "Invalid or missing authentication token")
        return
    }
    
    // Create order
    order, err := h.orderService.CreateOrder(ctx, userID, &req)
    if err != nil {
        h.logger.Error("Failed to create order", "request_id", requestID, "user_id", userID, "error", err)
        
        // Handle different error types
        switch {
        case errors.Is(err, services.ErrShopNotFound):
            h.respondWithError(w, http.StatusNotFound, "SHOP_NOT_FOUND", 
                "The specified coffee shop was not found")
        case errors.Is(err, services.ErrItemNotAvailable):
            h.respondWithError(w, http.StatusBadRequest, "ITEM_NOT_AVAILABLE", 
                "One or more items are not available")
        case errors.Is(err, services.ErrInsufficientInventory):
            h.respondWithError(w, http.StatusBadRequest, "INSUFFICIENT_INVENTORY", 
                "Insufficient inventory for requested items")
        default:
            h.respondWithError(w, http.StatusInternalServerError, "ORDER_CREATION_FAILED", 
                "Failed to create order")
        }
        return
    }
    
    // Log successful order creation
    h.logger.Info("Order created successfully", 
        "request_id", requestID, 
        "user_id", userID, 
        "order_id", order.ID,
        "shop_id", order.ShopID,
        "total_amount", order.TotalAmount)
    
    // Return success response
    response := dto.OrderResponse{
        ID:              order.ID,
        ShopID:          order.ShopID,
        Items:           order.Items,
        TotalAmount:     order.TotalAmount,
        Status:          order.Status,
        EstimatedReady:  order.EstimatedReadyTime,
        CreatedAt:       order.CreatedAt,
    }
    
    h.respondWithSuccess(w, http.StatusCreated, response, &dto.Meta{
        RequestID: requestID,
        Timestamp: time.Now(),
        Version:   "v1",
    })
}

func (h *OrderHandler) respondWithError(w http.ResponseWriter, statusCode int, code, message string) {
    response := dto.APIResponse{
        Success: false,
        Error: &dto.APIError{
            Code:    code,
            Message: message,
        },
        Meta: &dto.Meta{
            Timestamp: time.Now(),
            Version:   "v1",
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) respondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}, meta *dto.Meta) {
    response := dto.APIResponse{
        Success: true,
        Data:    data,
        Meta:    meta,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

#### Step 3: Implement API Versioning Strategy
```go
// internal/api/versioning/version_handler.go
package versioning

type VersionHandler struct {
    v1Handler handlers.OrderHandlerV1
    v2Handler handlers.OrderHandlerV2
}

func (vh *VersionHandler) RouteByVersion(w http.ResponseWriter, r *http.Request) {
    version := vh.extractVersion(r)
    
    switch version {
    case "v1":
        vh.v1Handler.CreateOrder(w, r)
    case "v2":
        vh.v2Handler.CreateOrder(w, r)
    default:
        vh.respondWithError(w, http.StatusBadRequest, "UNSUPPORTED_VERSION", 
            "API version not supported")
    }
}

func (vh *VersionHandler) extractVersion(r *http.Request) string {
    // Try URL path first: /api/v1/orders
    if matches := regexp.MustCompile(`/api/(v\d+)/`).FindStringSubmatch(r.URL.Path); len(matches) > 1 {
        return matches[1]
    }
    
    // Try Accept header: Accept: application/vnd.gocoffee.v1+json
    if accept := r.Header.Get("Accept"); accept != "" {
        if matches := regexp.MustCompile(`vnd\.gocoffee\.(v\d+)`).FindStringSubmatch(accept); len(matches) > 1 {
            return matches[1]
        }
    }
    
    // Try custom header: X-API-Version: v1
    if version := r.Header.Get("X-API-Version"); version != "" {
        return version
    }
    
    // Default to v1
    return "v1"
}
```

### 💡 Practice Question 4.1
**"Design a REST API for Go Coffee that supports both mobile apps and web dashboards with different data requirements."**

**Solution Framework:**
1. **API Design Principles**
   - Resource-based URLs
   - Consistent response formats
   - Proper HTTP status codes
   - Comprehensive error handling

2. **Versioning Strategy**
   - URL-based versioning for simplicity
   - Backward compatibility guarantees
   - Deprecation timeline communication

3. **Response Optimization**
   - Field selection: `?fields=id,name,price`
   - Pagination: `?page=1&limit=20`
   - Filtering: `?category=coffee&available=true`

---

## 📖 4.2 gRPC High-Performance Communication

### Core Concepts

#### gRPC Advantages
- **Performance**: Binary protocol, HTTP/2 multiplexing
- **Type Safety**: Protocol Buffers schema validation
- **Streaming**: Bidirectional streaming support
- **Language Agnostic**: Multi-language support
- **Code Generation**: Automatic client/server code

#### gRPC Patterns
- **Unary RPC**: Simple request-response
- **Server Streaming**: Server sends multiple responses
- **Client Streaming**: Client sends multiple requests
- **Bidirectional Streaming**: Both sides stream

### 🔍 Go Coffee Analysis

#### Study gRPC Implementation

<augment_code_snippet path="proto/order/order.proto" mode="EXCERPT">
````protobuf
syntax = "proto3";

package order.v1;

option go_package = "github.com/DimaJoyti/go-coffee/proto/order/v1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

service OrderService {
  // Unary RPC for creating orders
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  
  // Unary RPC for getting order details
  rpc GetOrder(GetOrderRequest) returns (Order);
  
  // Server streaming for real-time order updates
  rpc StreamOrderUpdates(StreamOrderUpdatesRequest) returns (stream OrderUpdate);
  
  // Client streaming for batch order creation
  rpc CreateBatchOrders(stream CreateOrderRequest) returns (BatchOrderResponse);
  
  // Bidirectional streaming for real-time kitchen communication
  rpc KitchenCommunication(stream KitchenMessage) returns (stream KitchenMessage);
}

message CreateOrderRequest {
  string shop_id = 1;
  string customer_id = 2;
  repeated OrderItem items = 3;
  PaymentMethod payment_method = 4;
  string special_instructions = 5;
}

message Order {
  string id = 1;
  string shop_id = 2;
  string customer_id = 3;
  repeated OrderItem items = 4;
  double total_amount = 5;
  OrderStatus status = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp estimated_ready_time = 8;
}

message OrderItem {
  string product_id = 1;
  string name = 2;
  int32 quantity = 3;
  double price = 4;
  repeated Customization customizations = 5;
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_PREPARING = 3;
  ORDER_STATUS_READY = 4;
  ORDER_STATUS_COMPLETED = 5;
  ORDER_STATUS_CANCELLED = 6;
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 4.2: Implement gRPC Service

#### Step 1: Implement gRPC Server
```go
// internal/grpc/server/order_server.go
package server

type OrderServer struct {
    pb.UnimplementedOrderServiceServer
    orderService services.OrderService
    logger       *slog.Logger
    metrics      *metrics.GRPCMetrics
}

func NewOrderServer(
    orderService services.OrderService,
    logger *slog.Logger,
    metrics *metrics.GRPCMetrics,
) *OrderServer {
    return &OrderServer{
        orderService: orderService,
        logger:       logger,
        metrics:      metrics,
    }
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // Start timing for metrics
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordRPCCall("CreateOrder", duration)
    }()
    
    // Validate request
    if err := s.validateCreateOrderRequest(req); err != nil {
        s.logger.Error("Invalid create order request", "error", err)
        return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
    }
    
    // Convert protobuf to domain model
    orderReq := &domain.CreateOrderRequest{
        ShopID:              req.ShopId,
        CustomerID:          req.CustomerId,
        Items:               s.convertOrderItems(req.Items),
        PaymentMethod:       s.convertPaymentMethod(req.PaymentMethod),
        SpecialInstructions: req.SpecialInstructions,
    }
    
    // Create order
    order, err := s.orderService.CreateOrder(ctx, orderReq)
    if err != nil {
        s.logger.Error("Failed to create order", "error", err)
        
        // Convert domain errors to gRPC status codes
        switch {
        case errors.Is(err, domain.ErrShopNotFound):
            return nil, status.Errorf(codes.NotFound, "shop not found")
        case errors.Is(err, domain.ErrInvalidItems):
            return nil, status.Errorf(codes.InvalidArgument, "invalid items")
        case errors.Is(err, domain.ErrInsufficientInventory):
            return nil, status.Errorf(codes.FailedPrecondition, "insufficient inventory")
        default:
            return nil, status.Errorf(codes.Internal, "internal server error")
        }
    }
    
    // Convert domain model to protobuf
    response := &pb.CreateOrderResponse{
        Order: s.convertOrderToProto(order),
    }
    
    s.logger.Info("Order created successfully", "order_id", order.ID)
    return response, nil
}

func (s *OrderServer) StreamOrderUpdates(req *pb.StreamOrderUpdatesRequest, stream pb.OrderService_StreamOrderUpdatesServer) error {
    ctx := stream.Context()
    
    // Create subscription for order updates
    updates, err := s.orderService.SubscribeToOrderUpdates(ctx, req.OrderId)
    if err != nil {
        return status.Errorf(codes.Internal, "failed to subscribe to updates: %v", err)
    }
    defer updates.Close()
    
    // Stream updates to client
    for {
        select {
        case update := <-updates.Updates():
            orderUpdate := &pb.OrderUpdate{
                OrderId:   update.OrderID,
                Status:    s.convertOrderStatus(update.Status),
                Message:   update.Message,
                Timestamp: timestamppb.New(update.Timestamp),
            }
            
            if err := stream.Send(orderUpdate); err != nil {
                s.logger.Error("Failed to send order update", "error", err)
                return status.Errorf(codes.Internal, "failed to send update")
            }
            
        case <-ctx.Done():
            s.logger.Info("Client disconnected from order updates stream")
            return nil
        }
    }
}

func (s *OrderServer) KitchenCommunication(stream pb.OrderService_KitchenCommunicationServer) error {
    ctx := stream.Context()
    
    // Handle bidirectional streaming
    go func() {
        for {
            msg, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                s.logger.Error("Error receiving kitchen message", "error", err)
                return
            }
            
            // Process kitchen message
            response, err := s.processKitchenMessage(ctx, msg)
            if err != nil {
                s.logger.Error("Error processing kitchen message", "error", err)
                continue
            }
            
            // Send response
            if err := stream.Send(response); err != nil {
                s.logger.Error("Error sending kitchen response", "error", err)
                return
            }
        }
    }()
    
    // Keep connection alive
    <-ctx.Done()
    return nil
}
```

#### Step 2: Implement gRPC Client with Connection Pooling
```go
// internal/grpc/client/order_client.go
package client

type OrderClient struct {
    conn   *grpc.ClientConn
    client pb.OrderServiceClient
    pool   *ConnectionPool
    logger *slog.Logger
}

type ConnectionPool struct {
    connections []*grpc.ClientConn
    current     int
    mutex       sync.RWMutex
}

func NewOrderClient(target string, poolSize int) (*OrderClient, error) {
    pool, err := NewConnectionPool(target, poolSize)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }
    
    conn := pool.GetConnection()
    client := pb.NewOrderServiceClient(conn)
    
    return &OrderClient{
        conn:   conn,
        client: client,
        pool:   pool,
        logger: slog.Default(),
    }, nil
}

func NewConnectionPool(target string, size int) (*ConnectionPool, error) {
    connections := make([]*grpc.ClientConn, size)
    
    for i := 0; i < size; i++ {
        conn, err := grpc.Dial(target,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:                10 * time.Second,
                Timeout:             time.Second,
                PermitWithoutStream: true,
            }),
            grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
                grpc_retry.WithMax(3),
                grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100*time.Millisecond)),
            )),
        )
        if err != nil {
            return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
        }
        connections[i] = conn
    }
    
    return &ConnectionPool{
        connections: connections,
        current:     0,
    }, nil
}

func (cp *ConnectionPool) GetConnection() *grpc.ClientConn {
    cp.mutex.Lock()
    defer cp.mutex.Unlock()
    
    conn := cp.connections[cp.current]
    cp.current = (cp.current + 1) % len(cp.connections)
    return conn
}

func (oc *OrderClient) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // Add timeout to context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Add metadata
    ctx = metadata.AppendToOutgoingContext(ctx,
        "request-id", uuid.New().String(),
        "client-version", "v1.0.0",
    )
    
    return oc.client.CreateOrder(ctx, req)
}
```

### 💡 Practice Question 4.2
**"When would you choose gRPC over REST for Go Coffee's internal service communication?"**

**Solution Framework:**
1. **Use gRPC When:**
   - High-performance internal communication needed
   - Type safety is critical
   - Streaming capabilities required
   - Polyglot environment (multiple languages)

2. **Use REST When:**
   - Public API for external clients
   - Browser-based applications
   - Simple request-response patterns
   - Human-readable debugging needed

---

## 📖 4.3 Message Queues & Event Streaming

### Core Concepts

#### Message Queue Patterns
- **Point-to-Point**: One producer, one consumer
- **Publish-Subscribe**: One producer, multiple consumers
- **Request-Reply**: Async request with correlation ID
- **Dead Letter Queue**: Handle failed messages

#### Kafka Concepts
- **Topics**: Categories of messages
- **Partitions**: Parallel processing units
- **Consumer Groups**: Load balancing consumers
- **Offsets**: Message position tracking

### 🔍 Go Coffee Analysis

#### Study Kafka Implementation

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
        Async:        false,               // Synchronous for reliability
        BatchSize:    100,                 // Batch for performance
        BatchTimeout: 10 * time.Millisecond,
    }
    
    return &Producer{
        writer: writer,
        config: config,
        logger: logger,
    }
}

func (p *Producer) PublishOrder(order *models.Order) error {
    orderJSON, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to marshal order: %w", err)
    }
    
    message := kafka.Message{
        Key:   []byte(order.ID),
        Value: orderJSON,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte("order.created")},
            {Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
            {Key: "version", Value: []byte("v1")},
        },
    }
    
    return p.writer.WriteMessages(context.Background(), message)
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 4.3: Advanced Message Queue Patterns

#### Step 1: Implement Event-Driven Order Processing
```go
// internal/events/order_events.go
package events

type OrderEventType string

const (
    OrderCreated   OrderEventType = "order.created"
    OrderConfirmed OrderEventType = "order.confirmed"
    OrderPreparing OrderEventType = "order.preparing"
    OrderReady     OrderEventType = "order.ready"
    OrderCompleted OrderEventType = "order.completed"
    OrderCancelled OrderEventType = "order.cancelled"
)

type OrderEvent struct {
    ID        string         `json:"id"`
    Type      OrderEventType `json:"type"`
    OrderID   string         `json:"order_id"`
    ShopID    string         `json:"shop_id"`
    Data      interface{}    `json:"data"`
    Timestamp time.Time      `json:"timestamp"`
    Version   string         `json:"version"`
}

type EventPublisher interface {
    PublishOrderEvent(ctx context.Context, event *OrderEvent) error
    PublishBatch(ctx context.Context, events []*OrderEvent) error
}

type KafkaEventPublisher struct {
    writer *kafka.Writer
    logger *slog.Logger
}

func (kep *KafkaEventPublisher) PublishOrderEvent(ctx context.Context, event *OrderEvent) error {
    eventJSON, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    message := kafka.Message{
        Key:   []byte(event.OrderID),
        Value: eventJSON,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(string(event.Type))},
            {Key: "event-id", Value: []byte(event.ID)},
            {Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
            {Key: "version", Value: []byte(event.Version)},
        },
    }
    
    return kep.writer.WriteMessages(ctx, message)
}

func (kep *KafkaEventPublisher) PublishBatch(ctx context.Context, events []*OrderEvent) error {
    messages := make([]kafka.Message, len(events))
    
    for i, event := range events {
        eventJSON, err := json.Marshal(event)
        if err != nil {
            return fmt.Errorf("failed to marshal event %d: %w", i, err)
        }
        
        messages[i] = kafka.Message{
            Key:   []byte(event.OrderID),
            Value: eventJSON,
            Headers: []kafka.Header{
                {Key: "event-type", Value: []byte(string(event.Type))},
                {Key: "event-id", Value: []byte(event.ID)},
                {Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
            },
        }
    }
    
    return kep.writer.WriteMessages(ctx, messages...)
}
```

#### Step 2: Implement Event Consumer with Error Handling
```go
// internal/events/order_consumer.go
package events

type EventHandler interface {
    HandleEvent(ctx context.Context, event *OrderEvent) error
    GetEventTypes() []OrderEventType
}

type OrderEventConsumer struct {
    reader   *kafka.Reader
    handlers map[OrderEventType][]EventHandler
    dlq      DeadLetterQueue
    logger   *slog.Logger
    metrics  *metrics.EventMetrics
}

func NewOrderEventConsumer(brokers []string, topic, groupID string) *OrderEventConsumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:     brokers,
        Topic:       topic,
        GroupID:     groupID,
        StartOffset: kafka.LastOffset,
        MinBytes:    10e3, // 10KB
        MaxBytes:    10e6, // 10MB
        MaxWait:     1 * time.Second,
    })
    
    return &OrderEventConsumer{
        reader:   reader,
        handlers: make(map[OrderEventType][]EventHandler),
        logger:   slog.Default(),
    }
}

func (oec *OrderEventConsumer) RegisterHandler(eventType OrderEventType, handler EventHandler) {
    oec.handlers[eventType] = append(oec.handlers[eventType], handler)
}

func (oec *OrderEventConsumer) Start(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := oec.processMessage(ctx); err != nil {
                oec.logger.Error("Failed to process message", "error", err)
                // Continue processing other messages
            }
        }
    }
}

func (oec *OrderEventConsumer) processMessage(ctx context.Context) error {
    msg, err := oec.reader.ReadMessage(ctx)
    if err != nil {
        return fmt.Errorf("failed to read message: %w", err)
    }
    
    // Parse event
    var event OrderEvent
    if err := json.Unmarshal(msg.Value, &event); err != nil {
        oec.logger.Error("Failed to unmarshal event", "error", err)
        return oec.dlq.SendToDeadLetter(msg, err)
    }
    
    // Get handlers for event type
    handlers, exists := oec.handlers[event.Type]
    if !exists {
        oec.logger.Warn("No handlers for event type", "event_type", event.Type)
        return nil
    }
    
    // Process with each handler
    for _, handler := range handlers {
        if err := oec.processWithRetry(ctx, handler, &event); err != nil {
            oec.logger.Error("Handler failed after retries", 
                "handler", fmt.Sprintf("%T", handler),
                "event_id", event.ID,
                "error", err)
            
            // Send to dead letter queue
            return oec.dlq.SendToDeadLetter(msg, err)
        }
    }
    
    return nil
}

func (oec *OrderEventConsumer) processWithRetry(ctx context.Context, handler EventHandler, event *OrderEvent) error {
    maxRetries := 3
    baseDelay := 100 * time.Millisecond
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := handler.HandleEvent(ctx, event)
        if err == nil {
            return nil
        }
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }
        
        if attempt < maxRetries-1 {
            delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
            time.Sleep(delay)
        }
    }
    
    return fmt.Errorf("handler failed after %d attempts", maxRetries)
}
```

### 💡 Practice Question 4.3
**"Design a message queue architecture for Go Coffee that ensures order processing reliability and handles peak traffic."**

**Solution Framework:**
1. **Topic Design**
   - Separate topics by domain (orders, payments, inventory)
   - Partition by shop_id for parallel processing
   - Use appropriate replication factor

2. **Consumer Strategy**
   - Consumer groups for load balancing
   - Dead letter queues for failed messages
   - Retry mechanisms with exponential backoff

3. **Performance Optimization**
   - Batch processing for throughput
   - Async processing for non-critical operations
   - Monitoring and alerting for queue health

---

## 🎯 4 Completion Checklist

### Knowledge Mastery
- [ ] Understand REST API design principles and best practices
- [ ] Can implement high-performance gRPC services
- [ ] Know message queue patterns and event streaming
- [ ] Understand real-time communication patterns
- [ ] Can design API gateway architectures

### Practical Skills
- [ ] Can design comprehensive REST APIs with proper versioning
- [ ] Can implement gRPC services with streaming capabilities
- [ ] Can build event-driven systems with Kafka
- [ ] Can handle API authentication and authorization
- [ ] Can implement real-time communication with WebSockets

### Go Coffee Analysis
- [ ] Analyzed REST API implementation patterns
- [ ] Studied gRPC service communication
- [ ] Examined Kafka event streaming architecture
- [ ] Understood API gateway routing and middleware
- [ ] Identified communication optimization techniques

###  Readiness
- [ ] Can discuss API design trade-offs and best practices
- [ ] Can explain when to use different communication patterns
- [ ] Can design scalable message queue architectures
- [ ] Can handle API versioning and backward compatibility
- [ ] Can optimize communication performance and reliability

---

## 🚀 Next Steps

Ready for **5: Scalability & Performance**:
- Load balancing strategies
- Horizontal and vertical scaling
- Performance optimization techniques
- Auto-scaling and capacity planning
- CDN and edge computing

**Excellent progress on mastering communication patterns! 🎉**
