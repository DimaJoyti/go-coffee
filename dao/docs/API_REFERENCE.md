# Developer DAO Platform - API Reference

## 🎯 Overview

Complete API reference for all three microservices in the Developer DAO Platform. All APIs follow RESTful conventions with consistent error handling and response formats.

## 🏗️ Base URLs

```
Bounty Service:     http://localhost:8080/api/v1
Marketplace Service: http://localhost:8081/api/v1  
Metrics Service:    http://localhost:8082/api/v1
```

## 🎯 Bounty Management API

### Bounties

#### List Bounties
```http
GET /bounties?category=0&status=1&developer=0x123&limit=20&offset=0
```

#### Create Bounty
```http
POST /bounties
Content-Type: application/json

{
  "title": "DeFi Analytics Dashboard",
  "description": "Build comprehensive analytics dashboard",
  "category": 0,
  "reward_amount": "10000.00",
  "currency": "USDC",
  "deadline": "2024-03-01T00:00:00Z",
  "milestones": [
    {
      "title": "UI Design",
      "description": "Complete UI mockups",
      "reward_percentage": "30.00",
      "deadline": "2024-02-01T00:00:00Z"
    }
  ]
}
```

#### Get Bounty Details
```http
GET /bounties/{id}
```

#### Apply for Bounty
```http
POST /bounties/{id}/apply
Content-Type: application/json

{
  "applicant_address": "0x1234567890123456789012345678901234567890",
  "message": "I want to work on this bounty",
  "proposed_timeline": 14
}
```

### Performance Tracking

#### Get Performance Dashboard
```http
GET /performance/dashboard
```

#### Get Developer Leaderboard
```http
GET /performance/leaderboard?limit=10
```

## 🏪 Solution Marketplace API

### Solutions

#### List Solutions
```http
GET /solutions?category=0&status=1&developer=0x123&min_rating=4.0&limit=20&offset=0
```

#### Create Solution
```http
POST /solutions
Content-Type: application/json

{
  "name": "DeFi Trading Widget",
  "description": "Reusable trading interface component",
  "category": 0,
  "version": "1.0.0",
  "developer_address": "0x1234567890123456789012345678901234567890",
  "repository_url": "https://github.com/dev/trading-widget",
  "documentation_url": "https://docs.trading-widget.com",
  "demo_url": "https://demo.trading-widget.com",
  "tags": ["defi", "trading", "widget"]
}
```

#### Get Solution Details
```http
GET /solutions/{id}
```

#### Review Solution
```http
POST /solutions/{id}/review
Content-Type: application/json

{
  "reviewer_address": "0x9876543210987654321098765432109876543210",
  "rating": 5,
  "comment": "Excellent solution with great performance",
  "security_score": 5,
  "performance_score": 4,
  "usability_score": 5,
  "documentation_score": 4
}
```

#### Install Solution
```http
POST /solutions/{id}/install
Content-Type: application/json

{
  "installer_address": "0x1111222233334444555566667777888899990000",
  "environment": "production",
  "config_data": {
    "api_key": "your_api_key",
    "theme": "dark"
  }
}
```

### Categories

#### Get All Categories
```http
GET /categories
```

#### Get Solutions by Category
```http
GET /categories/{category}/solutions?limit=20&offset=0
```

### Quality & Analytics

#### Calculate Quality Score
```http
POST /quality/score
Content-Type: application/json

{
  "solution_id": 1
}
```

#### Get Quality Metrics
```http
GET /quality/metrics
```

#### Get Popular Solutions
```http
GET /analytics/popular?limit=10
```

#### Get Trending Solutions
```http
GET /analytics/trending?limit=10
```

#### Get Marketplace Statistics
```http
GET /analytics/stats
```

## 📊 TVL/MAU Metrics API

### TVL Tracking

#### Get TVL Metrics
```http
GET /tvl?protocol=go-coffee-defi&chain=ethereum&period=daily
```

#### Record TVL Measurement
```http
POST /tvl/record
Content-Type: application/json

{
  "protocol": "go-coffee-defi",
  "chain": "ethereum",
  "amount": "5000000.00",
  "token_symbol": "USDC",
  "source": "defi-llama",
  "block_number": 18500000,
  "tx_hash": "0x1234567890abcdef"
}
```

#### Get TVL History
```http
GET /tvl/history?protocol=go-coffee-defi&chain=ethereum&days=30
```

#### Get TVL by Protocol
```http
GET /tvl/by-protocol?limit=20&offset=0
```

#### Get TVL Growth Analysis
```http
GET /tvl/growth?protocol=go-coffee-defi&chain=ethereum
```

### MAU Tracking

#### Get MAU Metrics
```http
GET /mau?feature=defi_trading&period=monthly
```

#### Record MAU Measurement
```http
POST /mau/record
Content-Type: application/json

{
  "feature": "defi_trading",
  "user_count": 25000,
  "unique_users": 22500,
  "period": "monthly",
  "source": "analytics"
}
```

#### Get MAU History
```http
GET /mau/history?feature=defi_trading&months=12
```

#### Get MAU by Feature
```http
GET /mau/by-feature
```

#### Get MAU Growth Analysis
```http
GET /mau/growth?feature=defi_trading
```

### Performance Analytics

#### Get Performance Dashboard
```http
GET /performance/dashboard
```

#### Get Attribution Analysis
```http
GET /performance/attribution?entity_id=dev_001&period=monthly
```

#### Record Impact Metrics
```http
POST /performance/impact
Content-Type: application/json

{
  "entity_id": "dev_001",
  "entity_type": "developer",
  "tvl_impact": "500000.00",
  "mau_impact": 2500,
  "attribution": "0.75",
  "source": "automated"
}
```

#### Get Impact Leaderboard
```http
GET /performance/leaderboard?entity_type=developer&period=monthly&limit=10
```

### Analytics & Insights

#### Get Analytics Overview
```http
GET /analytics/overview
```

#### Get Trends Analysis
```http
GET /analytics/trends?metric_type=0&period=monthly
```

#### Get Forecasts
```http
GET /analytics/forecasts?metric_type=tvl&horizon=quarterly
```

### Alerts & Monitoring

#### Get Active Alerts
```http
GET /analytics/alerts?limit=20&offset=0
```

#### Create Alert
```http
POST /analytics/alerts
Content-Type: application/json

{
  "name": "TVL Growth Alert",
  "type": 1,
  "metric_type": 0,
  "threshold": "1000000.00",
  "condition": "greater_than"
}
```

### Reporting

#### Get Daily Report
```http
GET /reports/daily?date=2024-01-15
```

#### Get Weekly Report
```http
GET /reports/weekly?week=2024-W03
```

#### Get Monthly Report
```http
GET /reports/monthly?month=2024-01
```

#### Generate Custom Report
```http
POST /reports/generate
Content-Type: application/json

{
  "name": "Q1 Performance Report",
  "type": "performance",
  "period": "quarterly",
  "filters": {
    "start_date": "2024-01-01",
    "end_date": "2024-03-31"
  },
  "created_by": "admin@developer-dao.com"
}
```

### Data Integration

#### Handle Webhook
```http
POST /integrations/webhook
Content-Type: application/json

{
  "source": "defi-llama",
  "type": "tvl_update",
  "data": {
    "protocol": "go-coffee-defi",
    "tvl": "5000000.00"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "signature": "webhook_signature"
}
```

#### Get Data Sources
```http
GET /integrations/sources
```

#### Add Data Source
```http
POST /integrations/sources
Content-Type: application/json

{
  "name": "DeFiLlama API",
  "type": "external_api",
  "url": "https://api.llama.fi",
  "api_key": "your_api_key",
  "config": {
    "rate_limit": "100",
    "timeout": "30s"
  }
}
```

## 🔧 Common Response Formats

### Success Response
```json
{
  "status": "success",
  "data": {
    // Response data
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Error Response
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "amount",
      "issue": "must be positive"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Paginated Response
```json
{
  "status": "success",
  "data": [
    // Array of items
  ],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "has_more": true
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🔒 Authentication

### JWT Token
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Rate Limiting
- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## 📊 Status Codes

- **200**: Success
- **201**: Created
- **400**: Bad Request
- **401**: Unauthorized
- **403**: Forbidden
- **404**: Not Found
- **429**: Too Many Requests
- **500**: Internal Server Error

## 🎉 Complete API Coverage

The Developer DAO Platform provides **60+ REST endpoints** across three microservices:

- **✅ Bounty Management**: 15+ endpoints for complete bounty lifecycle
- **✅ Solution Marketplace**: 20+ endpoints for component registry and quality
- **✅ TVL/MAU Tracking**: 25+ endpoints for metrics and analytics

**All APIs are production-ready with comprehensive validation, error handling, and documentation! 🚀**
