# 📚 Go Coffee Platform - API Documentation

## 🎯 Overview

The Go Coffee platform provides comprehensive REST APIs for all services including core operations, Web3 payments, AI orchestration, and advanced analytics. All APIs follow RESTful principles and return JSON responses.

## 🔗 Base URLs

### **Production**
- **Core API**: `https://api.go-coffee.com`
- **Web3 Payments**: `https://api.go-coffee.com/web3`
- **AI Orchestrator**: `https://api.go-coffee.com/ai`
- **Analytics**: `https://analytics.go-coffee.com`

### **Local Development**
- **Core API**: `http://localhost:3000`
- **Web3 Payments**: `http://localhost:8083`
- **AI Orchestrator**: `http://localhost:8094`
- **Analytics**: `http://localhost:8096`

## 🔐 Authentication

### **API Key Authentication**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     https://api.go-coffee.com/orders
```

### **JWT Authentication**
```bash
# Login to get JWT token
curl -X POST https://api.go-coffee.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Use JWT token
curl -H "Authorization: Bearer JWT_TOKEN" \
     https://api.go-coffee.com/orders
```

## ☕ Core API Endpoints

### **Orders Management**

#### **Create Order**
```http
POST /orders
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "customer_id": "cust_123",
  "items": [
    {
      "product_id": "prod_latte",
      "quantity": 2,
      "customizations": {
        "size": "large",
        "milk": "oat",
        "shots": 2
      }
    }
  ],
  "location_id": "loc_downtown",
  "payment_method": "card"
}
```

**Response:**
```json
{
  "order_id": "order_456",
  "status": "pending",
  "total_amount": 12.50,
  "estimated_time": "8 minutes",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### **Get Order Status**
```http
GET /orders/{order_id}
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "order_id": "order_456",
  "status": "preparing",
  "items": [...],
  "total_amount": 12.50,
  "estimated_completion": "2024-01-15T10:38:00Z",
  "progress": {
    "current_step": "brewing",
    "completion_percentage": 60
  }
}
```

#### **Update Order Status**
```http
PUT /orders/{order_id}/status
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "status": "ready",
  "notes": "Order ready for pickup"
}
```

### **Products Management**

#### **List Products**
```http
GET /products?category=coffee&available=true
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "products": [
    {
      "product_id": "prod_latte",
      "name": "Latte",
      "description": "Espresso with steamed milk",
      "category": "coffee",
      "price": 5.50,
      "available": true,
      "customizations": {
        "sizes": ["small", "medium", "large"],
        "milk_options": ["whole", "skim", "oat", "almond"],
        "extra_shots": true
      }
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

#### **Create Product**
```http
POST /products
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "name": "Seasonal Pumpkin Latte",
  "description": "Limited time autumn special",
  "category": "seasonal",
  "price": 6.50,
  "ingredients": ["espresso", "steamed_milk", "pumpkin_syrup", "cinnamon"],
  "available": true
}
```

### **Customer Management**

#### **Get Customer Profile**
```http
GET /customers/{customer_id}
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "customer_id": "cust_123",
  "email": "john@example.com",
  "name": "John Doe",
  "preferences": {
    "favorite_drink": "latte",
    "milk_preference": "oat",
    "loyalty_points": 150
  },
  "order_history": {
    "total_orders": 45,
    "total_spent": 247.50,
    "average_order": 5.50
  }
}
```

## 💰 Web3 Payment API

### **Create Crypto Payment**
```http
POST /web3/payment/create
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "order_id": "order_456",
  "customer_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
  "amount": "5.50",
  "currency": "USDC",
  "chain": "ethereum"
}
```

**Response:**
```json
{
  "payment_id": "pay_789",
  "payment_address": "0x1234567890abcdef...",
  "amount": "5.50",
  "currency": "USDC",
  "chain": "ethereum",
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "expires_at": "2024-01-15T10:45:00Z",
  "status": "pending"
}
```

### **Check Payment Status**
```http
GET /web3/payment/{payment_id}/status
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "payment_id": "pay_789",
  "status": "confirmed",
  "transaction_hash": "0xabcdef1234567890...",
  "confirmations": 12,
  "confirmed_at": "2024-01-15T10:42:30Z"
}
```

### **Supported Chains & Currencies**
```http
GET /web3/supported
```

**Response:**
```json
{
  "chains": [
    {
      "name": "ethereum",
      "display_name": "Ethereum",
      "currencies": ["ETH", "USDC", "USDT"]
    },
    {
      "name": "bsc",
      "display_name": "Binance Smart Chain",
      "currencies": ["BNB", "USDC", "USDT"]
    },
    {
      "name": "polygon",
      "display_name": "Polygon",
      "currencies": ["MATIC", "USDC", "USDT"]
    },
    {
      "name": "solana",
      "display_name": "Solana",
      "currencies": ["SOL", "USDC"]
    }
  ]
}
```

## 🤖 AI Orchestrator API

### **Assign Task to Agent**
```http
POST /ai/tasks/assign
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "agent_id": "beverage-inventor",
  "action": "create_recipe",
  "inputs": {
    "name": "Winter Spice Latte",
    "type": "seasonal",
    "flavor_profile": "warm_spices"
  },
  "priority": "medium"
}
```

**Response:**
```json
{
  "task_id": "task_abc123",
  "status": "assigned",
  "agent_id": "beverage-inventor",
  "estimated_completion": "2024-01-15T10:35:00Z"
}
```

### **Execute Workflow**
```http
POST /ai/workflows/execute
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "workflow_id": "coffee-order-processing",
  "inputs": {
    "order_id": "order_456",
    "priority": "high"
  }
}
```

**Response:**
```json
{
  "execution_id": "exec_def456",
  "workflow_id": "coffee-order-processing",
  "status": "running",
  "steps": [
    {
      "step_id": "analyze-order",
      "agent_id": "beverage-inventor",
      "status": "completed"
    },
    {
      "step_id": "check-inventory",
      "agent_id": "inventory-manager",
      "status": "running"
    }
  ]
}
```

### **List Available Agents**
```http
GET /ai/agents
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "agents": [
    {
      "id": "beverage-inventor",
      "name": "Beverage Inventor",
      "status": "active",
      "capabilities": [
        "create_recipe",
        "analyze_trends",
        "optimize_menu"
      ],
      "current_tasks": 3,
      "max_tasks": 10
    },
    {
      "id": "inventory-manager",
      "name": "Inventory Manager",
      "status": "active",
      "capabilities": [
        "check_availability",
        "forecast_demand",
        "manage_suppliers"
      ],
      "current_tasks": 1,
      "max_tasks": 10
    }
  ]
}
```

## 📊 Analytics API

### **Get Business Dashboard**
```http
GET /analytics/dashboard?period=7d
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "period": "7d",
  "kpis": {
    "total_revenue": 15420.50,
    "total_orders": 1247,
    "average_order_value": 12.37,
    "customer_satisfaction": 4.7
  },
  "trends": {
    "revenue_growth": 8.5,
    "order_growth": 12.3,
    "customer_growth": 5.2
  },
  "top_products": [
    {
      "product_id": "prod_latte",
      "name": "Latte",
      "orders": 342,
      "revenue": 1881.00
    }
  ]
}
```

### **Get Predictive Analytics**
```http
GET /analytics/predictions?horizon=30d
Authorization: Bearer TOKEN
```

**Response:**
```json
{
  "horizon": "30d",
  "revenue_forecast": {
    "predicted": 65000.00,
    "confidence": 0.87,
    "lower_bound": 58000.00,
    "upper_bound": 72000.00
  },
  "demand_forecast": [
    {
      "product": "Latte",
      "predicted_demand": 1450,
      "confidence": 0.91
    }
  ],
  "recommendations": [
    {
      "type": "inventory",
      "message": "Increase coffee bean order by 15% for next month",
      "confidence": 0.89
    }
  ]
}
```

### **Generate Report**
```http
POST /analytics/reports/generate
Content-Type: application/json
Authorization: Bearer TOKEN

{
  "type": "sales",
  "period": "monthly",
  "format": "pdf",
  "email": "manager@coffee-shop.com"
}
```

**Response:**
```json
{
  "report_id": "report_xyz789",
  "status": "generating",
  "estimated_completion": "2024-01-15T10:35:00Z",
  "download_url": null
}
```

## 📈 Real-time APIs

### **Server-Sent Events (SSE)**
```http
GET /realtime/events
Authorization: Bearer TOKEN
Accept: text/event-stream
```

**Response Stream:**
```
data: {"type":"order_created","order_id":"order_789","timestamp":"2024-01-15T10:30:00Z"}

data: {"type":"payment_confirmed","payment_id":"pay_456","amount":12.50}

data: {"type":"inventory_low","product_id":"prod_beans","quantity":5}
```

### **WebSocket Connection**
```javascript
const ws = new WebSocket('wss://api.go-coffee.com/ws?token=YOUR_TOKEN');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Real-time update:', data);
};
```

## 🔍 Health & Monitoring

### **Health Check**
```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "2.0.0",
  "services": {
    "database": "ok",
    "redis": "ok",
    "kafka": "ok"
  }
}
```

### **Metrics**
```http
GET /metrics
Accept: text/plain
```

**Response:**
```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1234
http_requests_total{method="POST",status="201"} 567

# HELP order_processing_duration_seconds Order processing time
# TYPE order_processing_duration_seconds histogram
order_processing_duration_seconds_bucket{le="1"} 100
order_processing_duration_seconds_bucket{le="5"} 450
```

## 🚨 Error Handling

### **Standard Error Response**
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid",
    "details": {
      "field": "customer_id",
      "reason": "required field missing"
    },
    "request_id": "req_123456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### **Common Error Codes**
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `429` - Too Many Requests
- `500` - Internal Server Error
- `503` - Service Unavailable

## 📝 Rate Limiting

### **Rate Limit Headers**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

### **Rate Limits by Endpoint**
- **Orders**: 100 requests/minute
- **Payments**: 50 requests/minute
- **Analytics**: 200 requests/minute
- **AI Tasks**: 20 requests/minute

## 🔧 SDKs & Libraries

### **JavaScript/TypeScript**
```bash
npm install @go-coffee/api-client
```

```javascript
import { GoCoffeeClient } from '@go-coffee/api-client';

const client = new GoCoffeeClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.go-coffee.com'
});

const order = await client.orders.create({
  customer_id: 'cust_123',
  items: [{ product_id: 'prod_latte', quantity: 1 }]
});
```

### **Python**
```bash
pip install go-coffee-api
```

```python
from go_coffee import GoCoffeeClient

client = GoCoffeeClient(api_key='your-api-key')
order = client.orders.create(
    customer_id='cust_123',
    items=[{'product_id': 'prod_latte', 'quantity': 1}]
)
```

### **Go**
```bash
go get github.com/DimaJoyti/go-coffee-client
```

```go
import "github.com/DimaJoyti/go-coffee-client"

client := gocoffee.NewClient("your-api-key")
order, err := client.Orders.Create(ctx, &gocoffee.CreateOrderRequest{
    CustomerID: "cust_123",
    Items: []gocoffee.OrderItem{
        {ProductID: "prod_latte", Quantity: 1},
    },
})
```

## 📞 Support

### **API Support**
- **Documentation**: https://docs.go-coffee.com
- **Status Page**: https://status.go-coffee.com
- **Support Email**: api-support@go-coffee.com

### **Rate Limit Increases**
For higher rate limits, contact: enterprise@go-coffee.com

---

**🎉 This comprehensive API documentation covers all endpoints and features of the Go Coffee platform. Use these APIs to build amazing coffee experiences!**
