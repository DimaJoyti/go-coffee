# API Design Documentation

## Overview

The Web3 Wallet Backend exposes RESTful APIs and gRPC services for a comprehensive coffee purchasing platform with cryptocurrency payments. The system supports Bitcoin, Ethereum, and major altcoins for seamless coffee ordering, payment processing, and order management. This document outlines the API design, endpoints, and data models for the complete coffee commerce ecosystem.

## API Architecture

### API Gateway Pattern

```text
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                             │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │Authentication│ │Rate Limiting│ │   Logging   │ │Load Balancer│ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
┌───────▼──────┐        ┌──────▼──────┐        ┌──────▼──────┐
│Supply Service│        │Order Service│        │Claim Service│
│   (gRPC)     │        │   (gRPC)    │        │   (gRPC)    │
└──────────────┘        └─────────────┘        └─────────────┘
```

## REST API Endpoints

### Supply Management API

#### Create Supply
```http
POST /api/v1/supplies
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "1.5"
}

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "1.5",
  "status": "pending",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Supply
```http
GET /api/v1/supplies/{id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "1.5",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### List Supplies
```http
GET /api/v1/supplies?user_id=uuid&currency=ETH&status=active&page=1&limit=20
Authorization: Bearer <token>

Response:
{
  "supplies": [...],
  "total": 100,
  "page": 1,
  "limit": 20,
  "has_next": true
}
```

#### Update Supply
```http
PUT /api/v1/supplies/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "active",
  "amount": "2.0"
}

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "2.0",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z"
}
```

#### Delete Supply
```http
DELETE /api/v1/supplies/{id}
Authorization: Bearer <token>

Response: 204 No Content
```

### Order Management API

#### Create Order
```http
POST /api/v1/orders
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "uuid",
  "currency": "ETH",
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2
    }
  ]
}

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "3.0",
  "status": "pending",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "items": [
    {
      "id": "uuid",
      "order_id": "uuid",
      "product_id": "uuid",
      "product_name": "Product Name",
      "price": "1.5",
      "quantity": 2
    }
  ]
}
```

#### Get Order
```http
GET /api/v1/orders/{id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "3.0",
  "status": "confirmed",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z",
  "items": [...]
}
```

#### List Orders
```http
GET /api/v1/orders?user_id=uuid&status=confirmed&page=1&limit=20
Authorization: Bearer <token>

Response:
{
  "orders": [...],
  "total": 50,
  "page": 1,
  "limit": 20,
  "has_next": true
}
```

#### Update Order
```http
PUT /api/v1/orders/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "confirmed"
}

Response:
{
  "id": "uuid",
  "user_id": "uuid",
  "currency": "ETH",
  "amount": "3.0",
  "status": "confirmed",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z",
  "items": [...]
}
```

### Claiming API

#### Claim Order
```http
POST /api/v1/claims
Content-Type: application/json
Authorization: Bearer <token>

{
  "order_id": "uuid",
  "user_id": "uuid"
}

Response:
{
  "id": "uuid",
  "order_id": "uuid",
  "user_id": "uuid",
  "status": "claimed",
  "claimed_at": "2024-01-01T00:00:00Z"
}
```

#### Get Claim
```http
GET /api/v1/claims/{id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "order_id": "uuid",
  "user_id": "uuid",
  "status": "processed",
  "claimed_at": "2024-01-01T00:00:00Z",
  "processed_at": "2024-01-01T00:01:00Z"
}
```

#### List Claims
```http
GET /api/v1/claims?user_id=uuid&status=processed&page=1&limit=20
Authorization: Bearer <token>

Response:
{
  "claims": [...],
  "total": 25,
  "page": 1,
  "limit": 20,
  "has_next": false
}
```

## gRPC Service Definitions

### Supply Service

```protobuf
syntax = "proto3";

package supply;

service SupplyService {
  rpc GetSupply(GetSupplyRequest) returns (GetSupplyResponse);
  rpc CreateSupply(CreateSupplyRequest) returns (CreateSupplyResponse);
  rpc UpdateSupply(UpdateSupplyRequest) returns (UpdateSupplyResponse);
  rpc DeleteSupply(DeleteSupplyRequest) returns (DeleteSupplyResponse);
  rpc ListSupplies(ListSuppliesRequest) returns (ListSuppliesResponse);
}

message Supply {
  string id = 1;
  string user_id = 2;
  string currency = 3;
  double amount = 4;
  string status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}
```

### Order Service

```protobuf
syntax = "proto3";

package order;

service OrderService {
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc UpdateOrder(UpdateOrderRequest) returns (UpdateOrderResponse);
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message Order {
  string id = 1;
  string user_id = 2;
  string currency = 3;
  double amount = 4;
  string status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  repeated OrderItem items = 8;
}

message OrderItem {
  string id = 1;
  string order_id = 2;
  string product_id = 3;
  string product_name = 4;
  double price = 5;
  int32 quantity = 6;
}
```

### Claiming Service

```protobuf
syntax = "proto3";

package claiming;

service ClaimingService {
  rpc ClaimOrder(ClaimOrderRequest) returns (ClaimOrderResponse);
  rpc GetClaim(GetClaimRequest) returns (GetClaimResponse);
  rpc ListClaims(ListClaimsRequest) returns (ListClaimsResponse);
}

message Claim {
  string id = 1;
  string order_id = 2;
  string user_id = 3;
  string status = 4;
  google.protobuf.Timestamp claimed_at = 5;
  google.protobuf.Timestamp processed_at = 6;
}
```

## Error Handling

### HTTP Status Codes

- `200 OK` - Successful GET, PUT requests
- `201 Created` - Successful POST requests
- `204 No Content` - Successful DELETE requests
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., already claimed)
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid",
    "details": [
      {
        "field": "amount",
        "message": "Amount must be positive"
      }
    ],
    "request_id": "uuid"
  }
}
```

## Authentication & Authorization

### JWT Token Structure

```json
{
  "sub": "user_id",
  "iat": 1640995200,
  "exp": 1641081600,
  "aud": "web3-wallet-backend",
  "iss": "web3-wallet-auth",
  "roles": ["user", "trader"],
  "permissions": ["supply:read", "order:write", "claim:execute"]
}
```

### Permission Model

- `supply:read` - Read supply data
- `supply:write` - Create/update supplies
- `supply:delete` - Delete supplies
- `order:read` - Read order data
- `order:write` - Create/update orders
- `order:delete` - Delete orders
- `claim:read` - Read claim data
- `claim:execute` - Claim orders

## Rate Limiting

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Window: 3600
```

### Rate Limit Tiers

- **Basic**: 100 requests/hour
- **Premium**: 1,000 requests/hour
- **Enterprise**: 10,000 requests/hour

## Versioning

### API Versioning Strategy

- URL versioning: `/api/v1/`, `/api/v2/`
- Backward compatibility for at least 2 versions
- Deprecation notices 6 months before removal
- Version-specific documentation

## WebSocket API

### Real-time Updates

```javascript
// Connect to WebSocket
const ws = new WebSocket('wss://api.web3wallet.com/ws');

// Subscribe to events
ws.send(JSON.stringify({
  type: 'subscribe',
  channels: ['supplies', 'orders', 'claims'],
  user_id: 'uuid'
}));

// Receive events
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
};
```

### Event Types

- `supply.created`
- `supply.updated`
- `supply.deleted`
- `order.created`
- `order.updated`
- `order.deleted`
- `order.claimed`
- `claim.processed`

## SDK Examples

### JavaScript SDK

```javascript
import { Web3WalletClient } from '@web3wallet/sdk';

const client = new Web3WalletClient({
  apiKey: 'your-api-key',
  baseURL: 'https://api.web3wallet.com'
});

// Create supply
const supply = await client.supplies.create({
  user_id: 'uuid',
  currency: 'ETH',
  amount: '1.5'
});

// Claim order
const claim = await client.claims.claimOrder({
  order_id: 'uuid',
  user_id: 'uuid'
});
```

### Python SDK

```python
from web3wallet import Web3WalletClient

client = Web3WalletClient(
    api_key='your-api-key',
    base_url='https://api.web3wallet.com'
)

# Create order
order = client.orders.create(
    user_id='uuid',
    currency='ETH',
    items=[{'product_id': 'uuid', 'quantity': 2}]
)

# List claims
claims = client.claims.list(user_id='uuid', status='processed')
```

## Testing

### API Testing Strategy

1. **Unit Tests**: Test individual endpoints
2. **Integration Tests**: Test service interactions
3. **Load Tests**: Test performance under load
4. **Contract Tests**: Test API contracts
5. **End-to-End Tests**: Test complete workflows

### Test Data

- Use deterministic test data
- Implement data factories for test objects
- Clean up test data after each test
- Use separate test databases

## Documentation

### Interactive API Documentation

- Swagger/OpenAPI specification
- Interactive API explorer
- Code examples in multiple languages
- Postman collections

### SDK Documentation

- Getting started guides
- API reference
- Code examples
- Best practices
