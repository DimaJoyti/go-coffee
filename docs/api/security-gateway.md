# Security Gateway API Documentation

## Overview

The Security Gateway provides a comprehensive API for managing security features, monitoring threats, and protecting microservices. This document covers all available endpoints, request/response formats, and authentication requirements.

## Base URL

```
Production: https://api.go-coffee.com
Development: http://localhost:8080
```

## Authentication

All API endpoints require authentication unless otherwise specified. The Security Gateway supports multiple authentication methods:

### API Key Authentication
```http
Authorization: Bearer your-api-key
```

### JWT Token Authentication
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Service-to-Service Authentication
```http
X-Service-Token: service-specific-token
X-Service-Name: auth-service
```

## Rate Limiting

All endpoints are subject to rate limiting:
- **Standard**: 100 requests per minute
- **Premium**: 1000 requests per minute
- **Service-to-Service**: 10000 requests per minute

Rate limit headers are included in all responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Error Handling

The API uses standard HTTP status codes and returns errors in JSON format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input provided",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    },
    "request_id": "req_123456789"
  }
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMITED` | 429 | Rate limit exceeded |
| `INTERNAL_ERROR` | 500 | Server error |

## Endpoints

### Health & Status

#### Health Check
Check the health status of the Security Gateway.

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "security-gateway",
  "version": "1.0.0",
  "checks": {
    "redis": "healthy",
    "database": "healthy",
    "external_services": "healthy"
  }
}
```

#### Service Status
Get detailed status of all security components.

```http
GET /api/v1/status
```

**Response:**
```json
{
  "gateway": {
    "status": "operational",
    "uptime": "72h30m15s",
    "requests_processed": 1234567
  },
  "waf": {
    "status": "operational",
    "rules_loaded": 150,
    "last_update": "2024-01-15T09:00:00Z"
  },
  "rate_limiter": {
    "status": "operational",
    "active_limits": 45,
    "blocked_requests": 123
  },
  "fraud_detector": {
    "status": "operational",
    "model_version": "v2.1.0",
    "accuracy": 0.987
  }
}
```

### Security Validation

#### Validate Input
Validate user input for security threats.

```http
POST /api/v1/security/validate
```

**Request Body:**
```json
{
  "type": "input|email|password|url|ip",
  "value": "string to validate",
  "context": {
    "user_id": "user123",
    "session_id": "session456"
  }
}
```

**Response:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": ["Potentially suspicious pattern detected"],
  "sanitized_value": "cleaned input",
  "threat_level": "low|medium|high|critical",
  "confidence": 0.95,
  "checks_performed": [
    "sql_injection",
    "xss",
    "path_traversal",
    "command_injection"
  ]
}
```

#### Bulk Validation
Validate multiple inputs in a single request.

```http
POST /api/v1/security/validate/bulk
```

**Request Body:**
```json
{
  "inputs": [
    {
      "id": "input1",
      "type": "email",
      "value": "user@example.com"
    },
    {
      "id": "input2",
      "type": "input",
      "value": "user input text"
    }
  ]
}
```

**Response:**
```json
{
  "results": [
    {
      "id": "input1",
      "valid": true,
      "threat_level": "low"
    },
    {
      "id": "input2",
      "valid": false,
      "errors": ["Potential XSS detected"],
      "threat_level": "high"
    }
  ]
}
```

### Security Metrics

#### Get Security Metrics
Retrieve real-time security metrics.

```http
GET /api/v1/security/metrics
```

**Query Parameters:**
- `timeframe`: `1h|24h|7d|30d` (default: `1h`)
- `format`: `json|prometheus` (default: `json`)

**Response:**
```json
{
  "timeframe": "1h",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "total_requests": 12345,
    "blocked_requests": 123,
    "allowed_requests": 12222,
    "threat_detections": 45,
    "rate_limit_violations": 67,
    "waf_blocks": 89,
    "fraud_detections": 12,
    "average_response_time_ms": 15.5,
    "error_rate": 0.001
  },
  "breakdown": {
    "by_country": {
      "US": 5000,
      "CA": 2000,
      "GB": 1500
    },
    "by_threat_type": {
      "sql_injection": 20,
      "xss": 15,
      "bot_traffic": 30
    }
  }
}
```

#### Get Prometheus Metrics
Retrieve metrics in Prometheus format.

```http
GET /metrics
```

**Response:**
```
# HELP security_gateway_total_requests Total number of requests
# TYPE security_gateway_total_requests counter
security_gateway_total_requests 12345

# HELP security_gateway_blocked_requests Total number of blocked requests
# TYPE security_gateway_blocked_requests counter
security_gateway_blocked_requests 123
```

### Security Alerts

#### List Security Alerts
Retrieve security alerts with filtering options.

```http
GET /api/v1/security/alerts
```

**Query Parameters:**
- `limit`: Number of alerts to return (default: 50, max: 1000)
- `offset`: Pagination offset (default: 0)
- `severity`: Filter by severity (`low|medium|high|critical`)
- `status`: Filter by status (`open|investigating|resolved`)
- `start_time`: Start time (ISO 8601 format)
- `end_time`: End time (ISO 8601 format)
- `source`: Filter by source system

**Response:**
```json
{
  "alerts": [
    {
      "id": "alert_123456",
      "type": "threat_detection",
      "severity": "high",
      "title": "Multiple failed login attempts detected",
      "description": "Detected 15 failed login attempts from IP 192.168.1.100 in 5 minutes",
      "source": "auth-service",
      "ip_address": "192.168.1.100",
      "user_id": "user123",
      "timestamp": "2024-01-15T10:25:00Z",
      "status": "open",
      "metadata": {
        "attempts": 15,
        "time_window": "5m",
        "user_agent": "Mozilla/5.0..."
      }
    }
  ],
  "total": 1,
  "pagination": {
    "limit": 50,
    "offset": 0,
    "has_more": false
  }
}
```

#### Get Alert Details
Retrieve detailed information about a specific alert.

```http
GET /api/v1/security/alerts/{alert_id}
```

**Response:**
```json
{
  "id": "alert_123456",
  "type": "threat_detection",
  "severity": "high",
  "title": "Multiple failed login attempts detected",
  "description": "Detailed description of the security event",
  "source": "auth-service",
  "ip_address": "192.168.1.100",
  "user_id": "user123",
  "timestamp": "2024-01-15T10:25:00Z",
  "status": "open",
  "timeline": [
    {
      "timestamp": "2024-01-15T10:25:00Z",
      "action": "alert_created",
      "details": "Alert automatically generated"
    }
  ],
  "related_events": [
    {
      "id": "event_789",
      "type": "failed_login",
      "timestamp": "2024-01-15T10:24:30Z"
    }
  ],
  "recommendations": [
    "Block IP address temporarily",
    "Notify user of suspicious activity"
  ]
}
```

#### Update Alert Status
Update the status of a security alert.

```http
PATCH /api/v1/security/alerts/{alert_id}
```

**Request Body:**
```json
{
  "status": "investigating|resolved",
  "assignee": "security_analyst_1",
  "notes": "Investigated and confirmed as false positive",
  "resolution": "false_positive|mitigated|escalated"
}
```

### WAF Management

#### Get WAF Rules
Retrieve current WAF rules.

```http
GET /api/v1/security/waf/rules
```

**Response:**
```json
{
  "rules": [
    {
      "id": "sql_001",
      "name": "SQL Injection - Union Select",
      "pattern": "(?i)(union\\s+select|union\\s+all\\s+select)",
      "action": "block",
      "severity": "high",
      "enabled": true,
      "last_triggered": "2024-01-15T09:30:00Z",
      "trigger_count": 15
    }
  ],
  "total": 150,
  "enabled": 145,
  "disabled": 5
}
```

#### Update WAF Rule
Enable, disable, or modify a WAF rule.

```http
PATCH /api/v1/security/waf/rules/{rule_id}
```

**Request Body:**
```json
{
  "enabled": true,
  "action": "block|log|challenge",
  "severity": "low|medium|high|critical"
}
```

### Rate Limiting

#### Get Rate Limit Status
Check current rate limit status for a key.

```http
GET /api/v1/security/rate-limit/status
```

**Query Parameters:**
- `key`: Rate limit key (IP, user ID, etc.)
- `type`: Rate limit type (`ip|user|endpoint|global`)

**Response:**
```json
{
  "key": "ip:192.168.1.100",
  "limit": 100,
  "remaining": 75,
  "reset_time": "2024-01-15T10:31:00Z",
  "window_size": "1m",
  "blocked": false
}
```

#### Reset Rate Limit
Reset rate limit for a specific key.

```http
DELETE /api/v1/security/rate-limit/{key}
```

**Response:**
```json
{
  "success": true,
  "message": "Rate limit reset successfully",
  "key": "ip:192.168.1.100"
}
```

### Fraud Detection

#### Analyze Transaction
Analyze a transaction for fraud indicators.

```http
POST /api/v1/security/fraud/analyze
```

**Request Body:**
```json
{
  "transaction": {
    "id": "txn_123456",
    "amount": 5500,
    "currency": "USD",
    "payment_method": "credit_card",
    "customer_id": "customer_789"
  },
  "metadata": {
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "device_id": "device_456",
    "session_id": "session_123",
    "geo_location": {
      "country": "US",
      "region": "CA",
      "city": "San Francisco",
      "latitude": 37.7749,
      "longitude": -122.4194
    }
  }
}
```

**Response:**
```json
{
  "risk_score": 0.25,
  "risk_level": "low",
  "decision": "allow",
  "requires_mfa": false,
  "requires_review": false,
  "confidence": 0.92,
  "indicators": [
    {
      "type": "velocity",
      "severity": "low",
      "description": "Normal transaction velocity",
      "value": 3,
      "threshold": 10
    }
  ],
  "recommendations": [
    "Allow transaction to proceed"
  ]
}
```

### Gateway Proxying

#### Proxy to Auth Service
Proxy requests to the authentication service.

```http
ANY /api/v1/gateway/auth/*
```

#### Proxy to Order Service
Proxy requests to the order service.

```http
ANY /api/v1/gateway/order/*
```

#### Proxy to Payment Service
Proxy requests to the payment service.

```http
ANY /api/v1/gateway/payment/*
```

#### Proxy to User Service
Proxy requests to the user service.

```http
ANY /api/v1/gateway/user/*
```

## WebSocket API

### Real-time Security Events
Subscribe to real-time security events via WebSocket.

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/security/events');

ws.onmessage = function(event) {
  const securityEvent = JSON.parse(event.data);
  console.log('Security event:', securityEvent);
};
```

**Event Format:**
```json
{
  "type": "threat_detected|alert_created|rule_triggered",
  "timestamp": "2024-01-15T10:30:00Z",
  "severity": "high",
  "source": "waf",
  "data": {
    "ip_address": "192.168.1.100",
    "rule_id": "sql_001",
    "blocked": true
  }
}
```

## SDK Examples

### JavaScript/Node.js
```javascript
const SecurityGateway = require('@go-coffee/security-gateway-sdk');

const client = new SecurityGateway({
  baseUrl: 'http://localhost:8080',
  apiKey: 'your-api-key'
});

// Validate input
const result = await client.validate({
  type: 'input',
  value: 'user input'
});

// Get metrics
const metrics = await client.getMetrics({ timeframe: '1h' });
```

### Python
```python
from go_coffee_security import SecurityGatewayClient

client = SecurityGatewayClient(
    base_url='http://localhost:8080',
    api_key='your-api-key'
)

# Validate input
result = client.validate(type='input', value='user input')

# Get alerts
alerts = client.get_alerts(severity='high', limit=10)
```

### Go
```go
package main

import (
    "github.com/DimaJoyti/go-coffee/pkg/security/client"
)

func main() {
    client := client.New(&client.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "your-api-key",
    })

    // Validate input
    result, err := client.Validate(ctx, &client.ValidateRequest{
        Type:  "input",
        Value: "user input",
    })
}
```

## Testing

### Health Check
```bash
curl -X GET http://localhost:8080/health
```

### Validate Input
```bash
curl -X POST http://localhost:8080/api/v1/security/validate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "type": "input",
    "value": "test input"
  }'
```

### Get Metrics
```bash
curl -X GET http://localhost:8080/api/v1/security/metrics \
  -H "Authorization: Bearer your-api-key"
```

### Multi-Factor Authentication (MFA)

#### Setup MFA
Initialize MFA for a user.

```http
POST /api/v1/auth/mfa/setup
```

**Request Body:**
```json
{
  "user_id": "user123",
  "method": "totp|sms|email",
  "phone_number": "+1234567890"
}
```

**Response:**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "backup_codes": [
    "12345678",
    "87654321"
  ],
  "setup_token": "setup_token_123"
}
```

#### Verify MFA Setup
Verify MFA setup with a code.

```http
POST /api/v1/auth/mfa/verify-setup
```

**Request Body:**
```json
{
  "user_id": "user123",
  "code": "123456",
  "setup_token": "setup_token_123"
}
```

#### Create MFA Challenge
Create an MFA challenge for authentication.

```http
POST /api/v1/auth/mfa/challenge
```

**Request Body:**
```json
{
  "user_id": "user123",
  "method": "totp|sms|email"
}
```

**Response:**
```json
{
  "challenge_id": "challenge_123",
  "method": "sms",
  "expires_at": "2024-01-15T10:35:00Z",
  "message": "Verification code sent to your phone"
}
```

#### Verify MFA Challenge
Verify an MFA challenge.

```http
POST /api/v1/auth/mfa/verify
```

**Request Body:**
```json
{
  "user_id": "user123",
  "challenge_id": "challenge_123",
  "code": "123456"
}
```

**Response:**
```json
{
  "verified": true,
  "message": "MFA verification successful",
  "session_token": "session_token_123"
}
```

### Device Management

#### Register Device
Register a new device for a user.

```http
POST /api/v1/auth/devices/register
```

**Request Body:**
```json
{
  "user_id": "user123",
  "device_fingerprint": "device_fingerprint_hash",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.100",
  "location": "San Francisco, CA"
}
```

#### Trust Device
Mark a device as trusted.

```http
POST /api/v1/auth/devices/{device_id}/trust
```

#### List User Devices
Get all devices for a user.

```http
GET /api/v1/auth/devices?user_id=user123
```

**Response:**
```json
{
  "devices": [
    {
      "id": "device_123",
      "fingerprint": "device_fingerprint_hash",
      "user_agent": "Mozilla/5.0...",
      "ip_address": "192.168.1.100",
      "location": "San Francisco, CA",
      "trusted": true,
      "last_used": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-10T09:00:00Z"
    }
  ]
}
```

## Changelog

### v1.0.0 (2024-01-15)
- Initial release
- WAF protection with OWASP Top 10 coverage
- Advanced rate limiting with multiple strategies
- ML-powered fraud detection
- Multi-factor authentication (TOTP, SMS, Email)
- Device fingerprinting and trust management
- Real-time security monitoring and alerting
- Comprehensive API for security operations
- Production-ready deployment configurations

### Upcoming Features (v1.1.0)
- Behavioral biometrics
- Advanced threat intelligence integration
- Zero-trust architecture components
- Enhanced machine learning models
- API rate limiting per endpoint
- Advanced geo-location analysis

---

For more information, see:
- [Security Architecture](../SECURITY-ARCHITECTURE.md)
- [Configuration Reference](../configuration/security-gateway.md)
- [Deployment Guide](../deployment/security-gateway.md)
