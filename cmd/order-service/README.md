# Order Service

A high-performance, scalable order management service built with Go and standard HTTP handlers.

## Features

- **Standard HTTP API** - No framework dependencies, pure Go HTTP handlers
- **Clean Architecture** - Separation of concerns with transport, application, and domain layers
- **Comprehensive Testing** - Unit tests with high coverage
- **Docker Support** - Production-ready containerization
- **Monitoring** - Prometheus metrics, Grafana dashboards, Jaeger tracing
- **Health Checks** - Built-in health monitoring
- **Graceful Shutdown** - Proper resource cleanup
- **Security** - Rate limiting, CORS, authentication middleware

## Architecture

```
cmd/order-service/
├── main.go                 # Application entry point
├── Dockerfile             # Container configuration
└── README.md              # This file

internal/order/
├── transport/http/         # HTTP transport layer
│   ├── handlers.go        # HTTP handlers
│   ├── middleware.go      # HTTP middleware
│   ├── handlers_test.go   # Handler tests
│   └── middleware_test.go # Middleware tests
├── application/           # Business logic layer
│   ├── order_service.go   # Order use cases
│   ├── payment_service.go # Payment use cases
│   └── dto.go            # Data transfer objects
├── domain/               # Domain models
│   ├── order.go         # Order aggregate
│   ├── payment.go       # Payment aggregate
│   └── events.go        # Domain events
└── infrastructure/      # External dependencies
    └── repository/      # Data persistence
```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Orders
- `POST /api/v1/orders` - Create new order
- `GET /api/v1/orders/{id}` - Get order by ID
- `POST /api/v1/orders/{id}/confirm` - Confirm order
- `PUT /api/v1/orders/{id}/status` - Update order status
- `DELETE /api/v1/orders/{id}` - Cancel order

### Payments
- `POST /api/v1/payments` - Create payment
- `POST /api/v1/payments/{id}/process` - Process payment
- `POST /api/v1/payments/{id}/refund` - Refund payment

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Redis (for caching)
- PostgreSQL (for persistence)

### Local Development

1. **Clone and build:**
```bash
git clone <repository>
cd go-coffee
go build ./cmd/order-service
```

2. **Run locally:**
```bash
# Set environment variables
export REDIS_URL=redis://localhost:6379
export HTTP_PORT=8081
export GRPC_PORT=50051

# Start the service
./order-service
```

3. **Test the service:**
```bash
curl http://localhost:8081/health
```

### Docker Deployment

1. **Build Docker image:**
```bash
docker build -f cmd/order-service/Dockerfile -t order-service:latest .
```

2. **Run with Docker Compose:**
```bash
docker-compose -f docker-compose.order.yml up -d
```

3. **Check status:**
```bash
docker-compose -f docker-compose.order.yml ps
```

### Using Makefile

```bash
# Build the service
make -f Makefile.order build

# Run tests
make -f Makefile.order test

# Start development environment
make -f Makefile.order dev

# Build Docker image
make -f Makefile.order docker-build

# Start with Docker
make -f Makefile.order docker-run

# Check health
make -f Makefile.order health

# View logs
make -f Makefile.order logs
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8081` | HTTP server port |
| `GRPC_PORT` | `50051` | gRPC server port |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `DATABASE_URL` | - | PostgreSQL connection URL |
| `LOG_LEVEL` | `info` | Logging level |
| `SERVICE_NAME` | `order-service` | Service name for logging |

## Testing

### Unit Tests
```bash
go test ./internal/order/transport/http/...
go test ./internal/order/application/...
go test ./internal/order/domain/...
```

### Integration Tests
```bash
go test -tags=integration ./internal/order/...
```

### API Testing
```bash
# Create order
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "items": [
      {
        "product_id": "coffee-001",
        "name": "Espresso",
        "quantity": 1,
        "unit_price": 500
      }
    ]
  }'

# Get order
curl http://localhost:8081/api/v1/orders/{order-id}?customer_id=customer-123
```

## Monitoring

### Metrics
- Prometheus metrics available at `/metrics`
- Grafana dashboard at `http://localhost:3000`
- Default credentials: `admin/admin`

### Tracing
- Jaeger UI at `http://localhost:16686`

### Logs
- Structured JSON logging
- Log levels: debug, info, warn, error
- Request/response logging with middleware

## Production Deployment

### Docker Compose
```bash
# Production deployment
docker-compose -f docker-compose.order.yml up -d

# Scale services
docker-compose -f docker-compose.order.yml up -d --scale order-service=3
```

### Kubernetes
```bash
# Apply manifests
kubectl apply -f k8s/order-service/

# Check status
kubectl get pods -l app=order-service
```

### Health Checks
- HTTP health endpoint: `/health`
- Docker health check included
- Kubernetes readiness/liveness probes supported

## Security

- **Rate Limiting** - API rate limiting with Redis
- **CORS** - Cross-origin resource sharing
- **Authentication** - JWT token validation (configurable)
- **Input Validation** - Request validation and sanitization
- **Security Headers** - Standard security headers

## Performance

- **Caching** - Redis caching for frequently accessed data
- **Connection Pooling** - Database connection pooling
- **Graceful Shutdown** - Proper resource cleanup
- **Metrics** - Performance monitoring with Prometheus

## Troubleshooting

### Common Issues

1. **Service not starting:**
   - Check Redis connection
   - Verify environment variables
   - Check port availability

2. **Database connection issues:**
   - Verify DATABASE_URL
   - Check PostgreSQL status
   - Review connection pool settings

3. **High memory usage:**
   - Check Redis memory usage
   - Review connection pool sizes
   - Monitor goroutine leaks

### Debug Mode
```bash
export LOG_LEVEL=debug
./order-service
```

### Logs
```bash
# Docker logs
docker-compose -f docker-compose.order.yml logs -f order-service

# Local logs
tail -f ./logs/order-service.log
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

MIT License - see LICENSE file for details.
