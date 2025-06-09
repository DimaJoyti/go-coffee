# Kitchen Service - Clean Architecture Implementation

The Kitchen Service is a comprehensive microservice designed to manage coffee shop kitchen operations with real-time optimization, staff management, equipment monitoring, and order processing capabilities.

## 🏗️ Architecture Overview

This service follows **Clean Architecture** principles with clear separation of concerns:

```
internal/kitchen/
├── domain/           # Business entities and rules
├── application/      # Use cases and business logic
├── infrastructure/   # External concerns (Redis, AI)
├── transport/        # API layer (gRPC, HTTP, WebSocket)
├── integration/      # Cross-service communication
└── config/          # Configuration management
```

## 🎯 Core Features

### 📋 Order Management
- **Queue Management**: Intelligent order queuing with priority handling
- **Status Tracking**: Real-time order status updates
- **Processing Optimization**: AI-driven order processing optimization
- **Cross-Service Integration**: Seamless integration with order service

### 👥 Staff Management
- **Availability Tracking**: Real-time staff availability monitoring
- **Skill-Based Assignment**: Intelligent staff assignment based on skills
- **Workload Balancing**: Automatic workload distribution
- **Performance Analytics**: Staff performance tracking and analytics

### 🔧 Equipment Management
- **Status Monitoring**: Real-time equipment status tracking
- **Maintenance Scheduling**: Predictive maintenance scheduling
- **Capacity Management**: Equipment capacity optimization
- **Efficiency Tracking**: Equipment efficiency analytics

### 🤖 AI-Powered Optimization
- **Queue Optimization**: AI-driven queue reordering for efficiency
- **Capacity Prediction**: Predictive capacity planning
- **Workflow Optimization**: Intelligent workflow suggestions
- **Performance Analytics**: Advanced performance insights

### 🔄 Real-Time Updates
- **WebSocket Integration**: Real-time updates for kitchen dashboards
- **Event-Driven Architecture**: Comprehensive event system
- **Cross-Service Events**: Integration with other microservices
- **Live Monitoring**: Real-time kitchen monitoring

## 🚀 API Endpoints

### gRPC API
- **Equipment Management**: CRUD operations for kitchen equipment
- **Staff Management**: Staff scheduling and assignment
- **Order Processing**: Order queue management and processing
- **Analytics**: Performance metrics and reporting

### HTTP REST API
- **Equipment**: `/api/v1/kitchen/equipment`
- **Staff**: `/api/v1/kitchen/staff`
- **Orders**: `/api/v1/kitchen/orders`
- **Queue**: `/api/v1/kitchen/queue`
- **Metrics**: `/api/v1/kitchen/metrics`

### WebSocket API
- **Real-time Updates**: `/ws`
- **Event Streaming**: Live kitchen events
- **Dashboard Integration**: Real-time dashboard updates

## 🔧 Configuration

### Environment Variables

```bash
# Service Configuration
SERVICE_NAME=kitchen-service
SERVICE_VERSION=1.0.0
ENVIRONMENT=development
LOG_LEVEL=info

# Database Configuration
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Transport Configuration
HTTP_PORT=8080
GRPC_PORT=9090
JWT_SECRET=your-secret-key
ENABLE_CORS=true
ENABLE_AUTH=false

# Integration Configuration
ORDER_SERVICE_ENABLED=true
ORDER_SERVICE_ADDRESS=localhost:50051
ORDER_SERVICE_SYNC_ENABLED=true
ORDER_SERVICE_SYNC_INTERVAL=30s

# AI Configuration
AI_ENABLED=true
AI_OPTIMIZATION_ENABLED=true
AI_PREDICTION_ENABLED=true
AI_LEARNING_RATE=0.01

# Monitoring Configuration
METRICS_ENABLED=true
METRICS_PORT=9091
HEALTH_ENABLED=true
HEALTH_PORT=8081
```

## 🏃‍♂️ Running the Service

### Prerequisites
- Go 1.21+
- Redis 6.0+
- Protocol Buffers compiler (for gRPC)

### Development Mode
```bash
# Start Redis
docker run -d -p 6379:6379 redis:latest

# Run the service
cd cmd/kitchen-service
go run main.go
```

### Production Mode
```bash
# Build the service
go build -o kitchen-service cmd/kitchen-service/main.go

# Run with production configuration
ENVIRONMENT=production ./kitchen-service
```

### Docker
```bash
# Build Docker image
docker build -t kitchen-service .

# Run with Docker Compose
docker-compose up kitchen-service
```

## 📊 Monitoring & Observability

### Health Checks
- **HTTP Health**: `GET /health`
- **gRPC Health**: Health check service
- **Redis Health**: Connection monitoring

### Metrics
- **Prometheus Metrics**: `/metrics` endpoint
- **Custom Metrics**: Kitchen-specific metrics
- **Performance Tracking**: Request/response times

### Logging
- **Structured Logging**: JSON format with fields
- **Log Levels**: Debug, Info, Warn, Error
- **Request Tracing**: Request ID tracking

## 🔄 Event System

### Published Events
- `kitchen.order.added_to_queue`
- `kitchen.order.status_changed`
- `kitchen.order.completed`
- `kitchen.order.overdue`
- `kitchen.equipment.status_changed`
- `kitchen.staff.assigned`
- `kitchen.queue.status_changed`

### Consumed Events
- `order.created`
- `order.updated`
- `order.cancelled`
- `order.payment_confirmed`

## 🧪 Testing

### Unit Tests
```bash
# Run unit tests
go test ./internal/kitchen/...

# Run with coverage
go test -cover ./internal/kitchen/...
```

### Integration Tests
```bash
# Run integration tests
go test -tags=integration ./internal/kitchen/...
```

### Load Testing
```bash
# gRPC load testing
ghz --insecure --proto kitchen.proto --call kitchen.KitchenService/GetQueueStatus localhost:9090

# HTTP load testing
ab -n 1000 -c 10 http://localhost:8080/api/v1/kitchen/queue/status
```

## 🔐 Security

### Authentication
- **JWT Tokens**: Bearer token authentication
- **Role-Based Access**: Staff, Manager roles
- **Permission System**: Granular permissions

### Authorization
- **RBAC**: Role-based access control
- **Resource Protection**: API endpoint protection
- **Audit Logging**: Security event logging

## 📈 Performance

### Optimization Features
- **Connection Pooling**: Redis connection pooling
- **Caching**: Intelligent caching strategies
- **Batch Operations**: Bulk data operations
- **Async Processing**: Non-blocking operations

### Scalability
- **Horizontal Scaling**: Multiple service instances
- **Load Balancing**: Request distribution
- **Circuit Breakers**: Fault tolerance
- **Rate Limiting**: Request throttling

## 🔧 Development

### Code Structure
```
domain/
├── equipment.go      # Equipment entity and business rules
├── staff.go         # Staff entity and business rules
├── order.go         # Kitchen order entity
├── queue.go         # Queue management
├── events.go        # Domain events
└── repository.go    # Repository interfaces

application/
├── interfaces.go    # Service interfaces
├── dto.go          # Data transfer objects
├── kitchen_service.go  # Main kitchen service
└── queue_service.go    # Queue management service

infrastructure/
├── repository/     # Redis implementations
└── ai/            # AI optimization services

transport/
├── grpc/          # gRPC server and handlers
├── http/          # HTTP REST handlers
├── websocket/     # WebSocket server
└── middleware/    # Authentication and logging
```

### Adding New Features
1. **Domain Layer**: Add entities and business rules
2. **Application Layer**: Implement use cases
3. **Infrastructure Layer**: Add external integrations
4. **Transport Layer**: Expose APIs
5. **Tests**: Add comprehensive tests

## 🤝 Integration

### Order Service Integration
- **Event-Driven**: Automatic order synchronization
- **Status Updates**: Bi-directional status updates
- **Error Handling**: Robust error handling and retries

### External Services
- **Payment Service**: Payment status integration
- **Notification Service**: Customer notifications
- **Analytics Service**: Performance data export

## 📚 API Documentation

### gRPC Documentation
- **Proto Files**: `proto/kitchen/` directory
- **Generated Docs**: Auto-generated from proto files
- **Examples**: gRPC client examples

### REST API Documentation
- **OpenAPI Spec**: Available at `/swagger`
- **Postman Collection**: API testing collection
- **Examples**: cURL examples for all endpoints

## 🐛 Troubleshooting

### Common Issues
1. **Redis Connection**: Check Redis URL and connectivity
2. **Port Conflicts**: Ensure ports 8080, 9090 are available
3. **Memory Usage**: Monitor Redis memory usage
4. **Event Processing**: Check event queue status

### Debug Mode
```bash
LOG_LEVEL=debug go run main.go
```

### Health Checks
```bash
# Check service health
curl http://localhost:8081/health

# Check Redis connectivity
redis-cli ping
```

## 📄 License

This project is part of the Go Coffee microservices platform.

## 🤝 Contributing

1. Follow Clean Architecture principles
2. Add comprehensive tests
3. Update documentation
4. Follow Go best practices
5. Add proper error handling
