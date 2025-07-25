# 🚪 API Gateway - Complete Documentation

## 📋 Overview

The API Gateway serves as the central entry point for all client requests to the Go Coffee platform. It handles routing, authentication, rate limiting, load balancing, and provides a unified interface for all microservices.

## 🏗️ Architecture

### **Core Components**
- **Router**: Request routing and path matching
- **Authentication Middleware**: JWT validation and user context
- **Rate Limiter**: Request throttling and abuse prevention
- **Load Balancer**: Backend service distribution
- **Circuit Breaker**: Fault tolerance and resilience
- **Metrics Collector**: Performance monitoring and analytics

### **Service Discovery**
- Automatic service registration and discovery
- Health check integration
- Dynamic routing updates
- Failover and redundancy

## 🔗 API Endpoints

### **1. Health and Status**

#### **GET /health**
Returns the health status of the API Gateway and connected services.

**Request:**
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "auth-service": "healthy",
    "order-service": "healthy",
    "payment-service": "healthy",
    "kitchen-service": "healthy"
  },
  "uptime": "72h30m15s"
}
```

#### **GET /metrics**
Returns Prometheus-compatible metrics for monitoring.

**Request:**
```http
GET /metrics
```

**Response:**
```
# HELP api_gateway_requests_total Total number of requests
# TYPE api_gateway_requests_total counter
api_gateway_requests_total{method="GET",status="200"} 1234
api_gateway_requests_total{method="POST",status="201"} 567

# HELP api_gateway_request_duration_seconds Request duration in seconds
# TYPE api_gateway_request_duration_seconds histogram
api_gateway_request_duration_seconds_bucket{le="0.1"} 800
api_gateway_request_duration_seconds_bucket{le="0.5"} 1200
```

### **2. Authentication Endpoints**

#### **POST /api/v1/auth/login**
Authenticates a user and returns JWT tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "remember_me": true
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "customer"
  }
}
```

#### **POST /api/v1/auth/refresh**
Refreshes an expired access token using a refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### **POST /api/v1/auth/logout**
Invalidates the current session and tokens.

**Request:**
```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "message": "Successfully logged out"
}
```

### **3. Order Management**

#### **GET /api/v1/orders**
Retrieves a list of orders for the authenticated user.

**Request:**
```http
GET /api/v1/orders?page=1&limit=10&status=pending
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)
- `status` (optional): Filter by order status
- `from_date` (optional): Filter orders from date (ISO 8601)
- `to_date` (optional): Filter orders to date (ISO 8601)

**Response:**
```json
{
  "orders": [
    {
      "id": "ord_123e4567e89b12d3",
      "status": "pending",
      "total_amount": 15.99,
      "currency": "USD",
      "items": [
        {
          "id": "item_456",
          "name": "Espresso",
          "quantity": 2,
          "price": 4.50,
          "customizations": ["extra shot", "oat milk"]
        }
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "estimated_completion": "2024-01-15T10:45:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

#### **POST /api/v1/orders**
Creates a new order.

**Request:**
```json
{
  "items": [
    {
      "menu_item_id": "menu_123",
      "quantity": 2,
      "customizations": ["extra shot", "oat milk"],
      "special_instructions": "Extra hot please"
    }
  ],
  "delivery_address": {
    "street": "123 Coffee St",
    "city": "Seattle",
    "state": "WA",
    "zip_code": "98101",
    "country": "US"
  },
  "payment_method_id": "pm_456",
  "tip_amount": 2.50,
  "notes": "Please call when ready"
}
```

**Response:**
```json
{
  "order_id": "ord_123e4567e89b12d3",
  "status": "confirmed",
  "total_amount": 18.49,
  "estimated_completion": "2024-01-15T10:45:00Z",
  "tracking_url": "https://app.gocoffee.com/track/ord_123e4567e89b12d3"
}
```

#### **GET /api/v1/orders/{order_id}**
Retrieves details for a specific order.

**Request:**
```http
GET /api/v1/orders/ord_123e4567e89b12d3
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "id": "ord_123e4567e89b12d3",
  "status": "in_progress",
  "total_amount": 18.49,
  "currency": "USD",
  "items": [...],
  "payment": {
    "method": "credit_card",
    "status": "paid",
    "transaction_id": "txn_789"
  },
  "delivery": {
    "type": "pickup",
    "estimated_time": "2024-01-15T10:45:00Z",
    "actual_time": null,
    "location": "Downtown Seattle"
  },
  "timeline": [
    {
      "status": "confirmed",
      "timestamp": "2024-01-15T10:30:00Z",
      "message": "Order confirmed and payment processed"
    },
    {
      "status": "in_progress",
      "timestamp": "2024-01-15T10:35:00Z",
      "message": "Barista started preparing your order"
    }
  ]
}
```

### **4. Menu and Products**

#### **GET /api/v1/menu**
Retrieves the current menu with all available items.

**Request:**
```http
GET /api/v1/menu?category=coffee&available=true
```

**Query Parameters:**
- `category` (optional): Filter by category (coffee, tea, food, etc.)
- `available` (optional): Filter by availability (true/false)
- `location_id` (optional): Filter by location availability

**Response:**
```json
{
  "categories": [
    {
      "id": "cat_coffee",
      "name": "Coffee",
      "description": "Premium coffee beverages",
      "items": [
        {
          "id": "menu_123",
          "name": "Espresso",
          "description": "Rich, bold espresso shot",
          "price": 4.50,
          "currency": "USD",
          "available": true,
          "preparation_time": 180,
          "calories": 5,
          "allergens": [],
          "customizations": [
            {
              "name": "Milk Type",
              "options": ["whole", "2%", "oat", "almond", "soy"],
              "default": "whole",
              "price_modifier": 0.50
            }
          ],
          "nutritional_info": {
            "calories": 5,
            "protein": 0.6,
            "carbs": 0.8,
            "fat": 0.2,
            "caffeine": 63
          }
        }
      ]
    }
  ],
  "last_updated": "2024-01-15T08:00:00Z"
}
```

### **5. Payment Processing**

#### **POST /api/v1/payments/process**
Processes a payment for an order.

**Request:**
```json
{
  "order_id": "ord_123e4567e89b12d3",
  "payment_method": {
    "type": "credit_card",
    "card_token": "tok_visa_1234",
    "billing_address": {
      "street": "123 Main St",
      "city": "Seattle",
      "state": "WA",
      "zip_code": "98101",
      "country": "US"
    }
  },
  "amount": 18.49,
  "currency": "USD",
  "tip_amount": 2.50
}
```

**Response:**
```json
{
  "transaction_id": "txn_789abc123",
  "status": "succeeded",
  "amount_charged": 18.49,
  "processing_fee": 0.59,
  "net_amount": 17.90,
  "receipt_url": "https://receipts.gocoffee.com/txn_789abc123",
  "estimated_settlement": "2024-01-16T10:30:00Z"
}
```

## 🔒 Authentication and Authorization

### **JWT Token Structure**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "role": "customer",
    "permissions": ["order:create", "order:read"],
    "iat": 1642248600,
    "exp": 1642252200,
    "iss": "go-coffee-auth"
  }
}
```

### **Authorization Headers**
All protected endpoints require the Authorization header:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### **Role-Based Access Control**
- **Customer**: Can create and view own orders
- **Barista**: Can view and update order status
- **Manager**: Can view all orders and analytics
- **Admin**: Full system access

## ⚡ Rate Limiting

### **Rate Limit Headers**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642252200
X-RateLimit-Window: 3600
```

### **Rate Limit Tiers**
- **Anonymous**: 100 requests/hour
- **Authenticated**: 1000 requests/hour
- **Premium**: 5000 requests/hour
- **Enterprise**: 10000 requests/hour

## 🚨 Error Handling

### **Error Response Format**
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid or malformed",
    "details": "Missing required field: email",
    "request_id": "req_123abc456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### **HTTP Status Codes**
- **200**: Success
- **201**: Created
- **400**: Bad Request
- **401**: Unauthorized
- **403**: Forbidden
- **404**: Not Found
- **429**: Too Many Requests
- **500**: Internal Server Error
- **502**: Bad Gateway
- **503**: Service Unavailable

## 📊 Monitoring and Metrics

### **Key Metrics**
- **Request Rate**: Requests per second
- **Response Time**: P50, P95, P99 latencies
- **Error Rate**: Percentage of failed requests
- **Throughput**: Successful requests per second
- **Availability**: Service uptime percentage

### **Health Check Endpoints**
- `/health`: Overall service health
- `/health/deep`: Detailed health with dependencies
- `/ready`: Readiness for traffic
- `/live`: Liveness check for orchestrators

## 🔧 Configuration

### **Environment Variables**
```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0
ENV=production

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRY=3600
REFRESH_TOKEN_EXPIRY=604800

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=3600

# Service Discovery
CONSUL_URL=http://consul:8500
SERVICE_NAME=api-gateway

# Monitoring
METRICS_ENABLED=true
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268
```

## 🚀 Deployment

### **Docker Configuration**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api-gateway ./cmd/api-gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]
```

### **Kubernetes Deployment**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: gocoffee/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: api-gateway-secrets
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```
