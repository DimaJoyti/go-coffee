# 📚 DeFi Algorithmic Trading API Documentation

Complete API reference for the Web3 DeFi Algorithmic Trading Platform.

## 🔗 Base URL

```
Production: https://api.defi-trading.com/v1
Staging: https://staging-api.defi-trading.com/v1
Development: http://localhost:8080/api/v1
```

## 🔐 Authentication

All API endpoints require authentication via JWT tokens.

### Get Access Token

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

### Use Token in Requests

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

## 🔄 Arbitrage Trading API

### Detect Arbitrage Opportunities

```http
GET /trading/arbitrage/opportunities
Authorization: Bearer YOUR_JWT_TOKEN
```

**Query Parameters:**
- `min_profit_margin` (float): Minimum profit margin (default: 0.005)
- `max_gas_cost` (float): Maximum gas cost in USD (default: 50)
- `chains` (string): Comma-separated chain names (default: "ethereum,bsc,polygon")

**Response:**
```json
{
  "opportunities": [
    {
      "id": "arb-001",
      "token": {
        "address": "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
        "symbol": "USDC",
        "name": "USD Coin",
        "decimals": 6,
        "chain": "ethereum"
      },
      "source_exchange": {
        "id": "uniswap-v3",
        "name": "Uniswap V3",
        "type": "dex"
      },
      "target_exchange": {
        "id": "1inch",
        "name": "1inch",
        "type": "dex"
      },
      "source_price": "1.000",
      "target_price": "1.015",
      "profit_margin": "0.015",
      "volume": "10000",
      "net_profit": "150",
      "gas_cost": "25",
      "confidence": "0.85",
      "risk": "medium",
      "expires_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 10
}
```

### Execute Arbitrage Trade

```http
POST /trading/arbitrage/execute
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "opportunity_id": "arb-001",
  "amount": "5000",
  "max_slippage": "0.01",
  "wallet_id": "wallet-123",
  "passphrase": "your-wallet-passphrase"
}
```

**Response:**
```json
{
  "transaction_hash": "0x1234567890abcdef...",
  "status": "pending",
  "estimated_profit": "75.50",
  "gas_used": "21000",
  "gas_price": "20",
  "execution_time": "2024-01-15T10:25:30Z"
}
```

## 🌾 Yield Farming API

### Get Yield Opportunities

```http
GET /trading/yield/opportunities
Authorization: Bearer YOUR_JWT_TOKEN
```

**Query Parameters:**
- `min_apy` (float): Minimum APY (default: 0.05)
- `max_risk` (string): Maximum risk level ("low", "medium", "high")
- `protocols` (string): Comma-separated protocols ("uniswap", "aave", "compound")
- `limit` (int): Number of results (default: 10, max: 100)

**Response:**
```json
{
  "opportunities": [
    {
      "id": "yield-001",
      "protocol": "uniswap",
      "chain": "ethereum",
      "pool": {
        "address": "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640",
        "token0": {
          "symbol": "USDC",
          "name": "USD Coin"
        },
        "token1": {
          "symbol": "ETH",
          "name": "Ethereum"
        },
        "fee": "0.003",
        "tvl": "2000000"
      },
      "strategy": "liquidity_provision",
      "apy": "0.125",
      "apr": "0.118",
      "tvl": "2000000",
      "min_deposit": "100",
      "max_deposit": "50000",
      "risk": "medium",
      "impermanent_loss": "0.05",
      "active": true
    }
  ],
  "total": 1
}
```

### Stake in Yield Farm

```http
POST /trading/yield/stake
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "opportunity_id": "yield-001",
  "amount": "1000",
  "auto_compound": true,
  "wallet_id": "wallet-123",
  "passphrase": "your-wallet-passphrase"
}
```

**Response:**
```json
{
  "transaction_hash": "0xabcdef1234567890...",
  "status": "pending",
  "lp_tokens": "22.36",
  "estimated_apy": "0.125",
  "staked_at": "2024-01-15T10:30:00Z"
}
```

## 🤖 Trading Bots API

### Create Trading Bot

```http
POST /trading/bots
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "name": "Conservative Arbitrage Bot",
  "strategy": "arbitrage",
  "config": {
    "max_position_size": "10000",
    "min_profit_margin": "0.005",
    "max_slippage": "0.01",
    "risk_tolerance": "medium",
    "auto_compound": true,
    "max_daily_trades": 50,
    "stop_loss_percent": "0.02",
    "take_profit_percent": "0.05"
  }
}
```

**Response:**
```json
{
  "id": "bot-001",
  "name": "Conservative Arbitrage Bot",
  "strategy": "arbitrage",
  "status": "created",
  "config": {
    "max_position_size": "10000",
    "min_profit_margin": "0.005",
    "risk_tolerance": "medium"
  },
  "created_at": "2024-01-15T10:00:00Z"
}
```

### Start Trading Bot

```http
POST /trading/bots/{bot_id}/start
Authorization: Bearer YOUR_JWT_TOKEN
```

**Response:**
```json
{
  "id": "bot-001",
  "status": "active",
  "started_at": "2024-01-15T10:30:00Z"
}
```

### Get Bot Performance

```http
GET /trading/bots/{bot_id}/performance
Authorization: Bearer YOUR_JWT_TOKEN
```

**Response:**
```json
{
  "bot_id": "bot-001",
  "performance": {
    "total_trades": 100,
    "winning_trades": 70,
    "losing_trades": 30,
    "win_rate": "0.70",
    "total_profit": "2000",
    "total_loss": "500",
    "net_profit": "1500",
    "roi": "0.15",
    "sharpe": "1.5",
    "max_drawdown": "200",
    "avg_trade_profit": "15"
  },
  "recent_trades": [
    {
      "id": "trade-001",
      "type": "arbitrage",
      "profit": "25.50",
      "executed_at": "2024-01-15T09:45:00Z"
    }
  ]
}
```

## 🔗 DeFi Integration API

### Get Token Price

```http
GET /defi/tokens/{token_address}/price
Authorization: Bearer YOUR_JWT_TOKEN
```

**Query Parameters:**
- `chain` (string): Blockchain name (default: "ethereum")

**Response:**
```json
{
  "token": {
    "address": "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
    "symbol": "USDC",
    "name": "USD Coin",
    "decimals": 6,
    "chain": "ethereum"
  },
  "price": "1.000",
  "price_change_24h": "0.001",
  "volume_24h": "1000000",
  "market_cap": "50000000000",
  "last_updated": "2024-01-15T10:30:00Z"
}
```

### Get Swap Quote

```http
POST /defi/swap/quote
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "token_in": "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
  "token_out": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
  "amount_in": "1000",
  "chain": "ethereum"
}
```

**Response:**
```json
{
  "quote": {
    "id": "quote-123",
    "token_in": "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
    "token_out": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
    "amount_in": "1000",
    "amount_out": "0.5",
    "price": "2000",
    "price_impact": "0.01",
    "protocol": "uniswap",
    "gas_estimate": "150000",
    "expires_at": "2024-01-15T10:35:00Z"
  }
}
```

### Execute Swap

```http
POST /defi/swap/execute
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "quote_id": "quote-123",
  "wallet_id": "wallet-123",
  "passphrase": "your-wallet-passphrase",
  "max_slippage": "0.01"
}
```

**Response:**
```json
{
  "transaction_hash": "0x9876543210fedcba...",
  "status": "pending",
  "amount_in": "1000",
  "amount_out": "0.495",
  "gas_used": "145000",
  "executed_at": "2024-01-15T10:30:00Z"
}
```

## 📊 Analytics API

### Get On-Chain Metrics

```http
GET /analytics/onchain/metrics
Authorization: Bearer YOUR_JWT_TOKEN
```

**Query Parameters:**
- `token_address` (string): Token contract address
- `chain` (string): Blockchain name
- `timeframe` (string): "1h", "24h", "7d", "30d"

**Response:**
```json
{
  "token": {
    "address": "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
    "symbol": "USDC",
    "chain": "ethereum"
  },
  "metrics": {
    "price": "1.000",
    "volume_24h": "1000000",
    "liquidity": "5000000",
    "market_cap": "50000000000",
    "holders": 100000,
    "transactions_24h": 50000,
    "volatility": "0.02"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get Market Signals

```http
GET /analytics/market/signals
Authorization: Bearer YOUR_JWT_TOKEN
```

**Response:**
```json
{
  "signals": [
    {
      "id": "signal-001",
      "type": "bullish_divergence",
      "token": {
        "symbol": "ETH",
        "address": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
      },
      "direction": "bullish",
      "strength": "0.75",
      "confidence": "0.85",
      "description": "Strong bullish divergence detected",
      "created_at": "2024-01-15T10:25:00Z",
      "expires_at": "2024-01-15T11:25:00Z"
    }
  ]
}
```

## ❌ Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "Insufficient balance for trade execution",
    "details": {
      "required": "1000",
      "available": "500",
      "currency": "USDC"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-123456"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `INVALID_TOKEN` | Invalid or expired JWT token |
| `INSUFFICIENT_BALANCE` | Insufficient wallet balance |
| `SLIPPAGE_EXCEEDED` | Trade slippage exceeded maximum |
| `OPPORTUNITY_EXPIRED` | Arbitrage opportunity expired |
| `POOL_NOT_FOUND` | Yield farming pool not found |
| `BOT_ALREADY_RUNNING` | Trading bot already active |
| `RATE_LIMIT_EXCEEDED` | API rate limit exceeded |

## 📈 Rate Limits

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| Authentication | 10 requests | 1 minute |
| Trading Operations | 100 requests | 1 minute |
| Data Queries | 1000 requests | 1 minute |
| Bot Management | 50 requests | 1 minute |

## 🔔 Webhooks

Subscribe to real-time events:

```http
POST /webhooks/subscribe
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "url": "https://your-app.com/webhooks/defi",
  "events": ["arbitrage.opportunity", "trade.executed", "bot.performance"]
}
```

**Webhook Payload Example:**
```json
{
  "event": "arbitrage.opportunity",
  "data": {
    "opportunity_id": "arb-001",
    "profit_margin": "0.015",
    "expires_at": "2024-01-15T10:35:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

**🚀 Ready to integrate? Check out our [SDK documentation](SDK.md) for easier implementation!**
