# Coffee Purchase with Cryptocurrency API

## Overview

This document outlines the API endpoints for purchasing coffee using cryptocurrency payments. The system supports Bitcoin, Ethereum, and major altcoins for seamless coffee ordering and payment processing.

## Coffee Shop Discovery API

### List Coffee Shops
```http
GET /api/v1/coffee-shops?lat=40.7128&lng=-74.0060&radius=5000
Authorization: Bearer <token>

Response:
{
  "shops": [
    {
      "id": "uuid",
      "name": "Crypto Coffee Co.",
      "address": "123 Blockchain St, New York, NY",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "distance": 250,
      "phone": "+1-555-0123",
      "rating": 4.8,
      "opening_hours": {
        "monday": "07:00-19:00",
        "tuesday": "07:00-19:00",
        "wednesday": "07:00-19:00",
        "thursday": "07:00-19:00",
        "friday": "07:00-20:00",
        "saturday": "08:00-20:00",
        "sunday": "08:00-18:00"
      },
      "accepts_crypto": true,
      "supported_currencies": ["BTC", "ETH", "USDC", "USDT"],
      "status": "open"
    }
  ],
  "total": 15,
  "page": 1,
  "limit": 20
}
```

### Get Coffee Shop Details
```http
GET /api/v1/coffee-shops/{shop_id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "name": "Crypto Coffee Co.",
  "description": "Premium coffee with crypto payments",
  "address": "123 Blockchain St, New York, NY",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "phone": "+1-555-0123",
  "email": "hello@cryptocoffee.com",
  "website": "https://cryptocoffee.com",
  "rating": 4.8,
  "review_count": 342,
  "opening_hours": {...},
  "accepts_crypto": true,
  "supported_currencies": ["BTC", "ETH", "USDC", "USDT"],
  "payment_confirmations_required": {
    "BTC": 1,
    "ETH": 12,
    "USDC": 12,
    "USDT": 12
  },
  "status": "open",
  "features": ["wifi", "outdoor_seating", "drive_through"],
  "images": ["url1", "url2", "url3"]
}
```

## Product Catalog API

### Get Shop Menu
```http
GET /api/v1/coffee-shops/{shop_id}/menu
Authorization: Bearer <token>

Response:
{
  "categories": [
    {
      "name": "Coffee",
      "products": [
        {
          "id": "uuid",
          "name": "Espresso",
          "description": "Rich and bold espresso shot",
          "base_price": 3.50,
          "crypto_prices": {
            "BTC": "0.00008750",
            "ETH": "0.00140000",
            "USDC": "3.50",
            "USDT": "3.50"
          },
          "image_url": "https://cdn.example.com/espresso.jpg",
          "category": "coffee",
          "available": true,
          "preparation_time": 2,
          "customizations": [
            {
              "name": "Size",
              "options": [
                {"name": "Small", "price_modifier": 0.00},
                {"name": "Medium", "price_modifier": 0.50},
                {"name": "Large", "price_modifier": 1.00}
              ]
            },
            {
              "name": "Milk",
              "options": [
                {"name": "Regular", "price_modifier": 0.00},
                {"name": "Oat Milk", "price_modifier": 0.60},
                {"name": "Almond Milk", "price_modifier": 0.50}
              ]
            }
          ],
          "nutritional_info": {
            "calories": 5,
            "caffeine_mg": 63
          }
        }
      ]
    }
  ]
}
```

## Order Creation API

### Create Coffee Order
```http
POST /api/v1/orders
Content-Type: application/json
Authorization: Bearer <token>

{
  "shop_id": "uuid",
  "order_type": "pickup",
  "pickup_time": "2024-01-01T15:30:00Z",
  "payment_currency": "ETH",
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2,
      "customizations": {
        "size": "Large",
        "milk": "Oat Milk"
      }
    },
    {
      "product_id": "uuid2",
      "quantity": 1,
      "customizations": {
        "size": "Medium"
      }
    }
  ],
  "special_instructions": "Extra hot, please"
}

Response:
{
  "id": "uuid",
  "shop_id": "uuid",
  "user_id": "uuid",
  "order_number": "CC-2024-001234",
  "total_amount": 8.50,
  "currency": "USD",
  "crypto_amount": "0.00340000",
  "crypto_currency": "ETH",
  "exchange_rate": "2500.00",
  "order_status": "pending_payment",
  "payment_status": "pending",
  "order_type": "pickup",
  "pickup_time": "2024-01-01T15:30:00Z",
  "estimated_ready_time": "2024-01-01T15:35:00Z",
  "special_instructions": "Extra hot, please",
  "items": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product_name": "Espresso",
      "base_price": 3.50,
      "final_price": 4.60,
      "quantity": 2,
      "customizations": {
        "size": "Large (+$1.00)",
        "milk": "Oat Milk (+$0.60)"
      }
    }
  ],
  "payment_details": {
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96590e4265",
    "payment_expires_at": "2024-01-01T14:45:00Z",
    "qr_code": "data:image/png;base64,..."
  },
  "created_at": "2024-01-01T14:30:00Z"
}
```

## Cryptocurrency Payment API

### Get Payment Status
```http
GET /api/v1/payments/{payment_id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "order_id": "uuid",
  "crypto_currency": "ETH",
  "crypto_amount": "0.00340000",
  "fiat_amount": 8.50,
  "fiat_currency": "USD",
  "exchange_rate": "2500.00",
  "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96590e4265",
  "transaction_hash": "0x1234567890abcdef...",
  "block_number": 18500000,
  "confirmations": 15,
  "required_confirmations": 12,
  "status": "confirmed",
  "expires_at": "2024-01-01T14:45:00Z",
  "confirmed_at": "2024-01-01T14:33:00Z",
  "created_at": "2024-01-01T14:30:00Z"
}
```

### Verify Payment
```http
POST /api/v1/payments/{payment_id}/verify
Content-Type: application/json
Authorization: Bearer <token>

{
  "transaction_hash": "0x1234567890abcdef..."
}

Response:
{
  "verified": true,
  "confirmations": 15,
  "status": "confirmed",
  "order_status": "confirmed"
}
```

## Order Management API

### Get Order Status
```http
GET /api/v1/orders/{order_id}
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "order_number": "CC-2024-001234",
  "shop": {
    "id": "uuid",
    "name": "Crypto Coffee Co.",
    "address": "123 Blockchain St, New York, NY"
  },
  "total_amount": 8.50,
  "currency": "USD",
  "crypto_amount": "0.00340000",
  "crypto_currency": "ETH",
  "order_status": "preparing",
  "payment_status": "confirmed",
  "order_type": "pickup",
  "pickup_time": "2024-01-01T15:30:00Z",
  "estimated_ready_time": "2024-01-01T15:35:00Z",
  "actual_ready_time": null,
  "items": [...],
  "payment": {
    "transaction_hash": "0x1234567890abcdef...",
    "confirmations": 15,
    "confirmed_at": "2024-01-01T14:33:00Z"
  },
  "timeline": [
    {
      "status": "pending_payment",
      "timestamp": "2024-01-01T14:30:00Z"
    },
    {
      "status": "payment_confirmed",
      "timestamp": "2024-01-01T14:33:00Z"
    },
    {
      "status": "confirmed",
      "timestamp": "2024-01-01T14:33:30Z"
    },
    {
      "status": "preparing",
      "timestamp": "2024-01-01T14:35:00Z"
    }
  ],
  "created_at": "2024-01-01T14:30:00Z"
}
```

## Order Claiming API

### Claim Order (Generate Pickup Code)
```http
POST /api/v1/orders/{order_id}/claim
Authorization: Bearer <token>

Response:
{
  "claim_id": "uuid",
  "order_id": "uuid",
  "claim_code": "ABC123",
  "qr_code": "data:image/png;base64,....."
  "status": "claimed",
  "claimed_at": "2024-01-01T15:30:00Z",
  "expires_at": "2024-01-01T16:30:00Z"
}
```

### Verify Pickup Code
```http
POST /api/v1/claims/verify
Content-Type: application/json
Authorization: Bearer <token>

{
  "claim_code": "ABC123",
  "shop_id": "uuid"
}

Response:
{
  "valid": true,
  "claim": {
    "id": "uuid",
    "order_id": "uuid",
    "order_number": "CC-2024-001234",
    "customer_name": "John Doe",
    "items": [
      {
        "product_name": "Large Espresso with Oat Milk",
        "quantity": 2
      }
    ],
    "total_amount": 8.50,
    "payment_status": "confirmed",
    "order_status": "ready"
  }
}
```

## Real-time Updates via WebSocket

### Order Status Updates
```javascript
// Connect to WebSocket
const ws = new WebSocket('wss://api.cryptocoffee.com/ws');

// Subscribe to order updates
ws.send(JSON.stringify({
  type: 'subscribe',
  channels: ['order_updates'],
  order_id: 'uuid'
}));

// Receive real-time updates
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  /*
  {
    "type": "order_status_update",
    "order_id": "uuid",
    "status": "ready",
    "estimated_ready_time": "2024-01-01T15:35:00Z",
    "message": "Your order is ready for pickup!"
  }
  */
};
```

## Cryptocurrency Price API

### Get Current Crypto Prices
```http
GET /api/v1/crypto/prices?currencies=BTC,ETH,USDC,USDT
Authorization: Bearer <token>

Response:
{
  "prices": {
    "BTC": {
      "usd": 40000.00,
      "last_updated": "2024-01-01T14:30:00Z"
    },
    "ETH": {
      "usd": 2500.00,
      "last_updated": "2024-01-01T14:30:00Z"
    },
    "USDC": {
      "usd": 1.00,
      "last_updated": "2024-01-01T14:30:00Z"
    },
    "USDT": {
      "usd": 1.00,
      "last_updated": "2024-01-01T14:30:00Z"
    }
  }
}
```

## Error Handling

### Payment Errors
```json
{
  "error": {
    "code": "PAYMENT_EXPIRED",
    "message": "Payment window has expired",
    "details": {
      "expires_at": "2024-01-01T14:45:00Z",
      "current_time": "2024-01-01T14:46:00Z"
    },
    "request_id": "uuid"
  }
}
```

### Order Errors
```json
{
  "error": {
    "code": "INSUFFICIENT_SUPPLY",
    "message": "Not enough coffee beans in stock",
    "details": {
      "product_id": "uuid",
      "requested_quantity": 5,
      "available_quantity": 2
    },
    "request_id": "uuid"
  }
}
```

## Rate Limiting

- **Order Creation**: 10 orders per hour per user
- **Payment Verification**: 100 requests per hour per user
- **Price Updates**: 1000 requests per hour per user
- **General API**: 1000 requests per hour per user

## Security Features

1. **Payment Security**
   - Time-limited payment windows (15 minutes)
   - Unique wallet addresses per transaction
   - Multi-signature wallet support
   - Transaction verification on blockchain

2. **Order Security**
   - Unique claim codes for pickup
   - QR code verification
   - Time-limited claim windows
   - Shop-specific validation

3. **API Security**
   - JWT authentication
   - Rate limiting
   - Request signing for sensitive operations
   - IP whitelisting for shop endpoints
