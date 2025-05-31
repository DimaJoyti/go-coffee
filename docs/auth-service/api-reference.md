# 📖 Auth Service API Reference

<div align="center">

![API Version](https://img.shields.io/badge/API%20Version-v1-blue?style=for-the-badge)
![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-green?style=for-the-badge)
![REST](https://img.shields.io/badge/REST-API-orange?style=for-the-badge)

**Complete API documentation for the Auth Service**

</div>

---

## 🌐 Base Information

| **Property** | **Value** |
|-------------|-----------|
| **Base URL** | `http://localhost:8080` |
| **API Version** | `v1` |
| **Content Type** | `application/json` |
| **Authentication** | Bearer Token (JWT) |

---

## 🔑 Authentication Endpoints

### 👤 Register User

**Create a new user account**

```http
POST /api/v1/auth/register
```

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "role": "user"
}
```

#### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | `string` | ✅ | Valid email address |
| `password` | `string` | ✅ | Password (min 8 chars, mixed case, numbers, symbols) |
| `role` | `string` | ❌ | User role (`user`, `admin`) - defaults to `user` |

#### Response

<details>
<summary><b>✅ 201 Created</b></summary>

```json
{
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com",
    "role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

</details>

<details>
<summary><b>❌ 400 Bad Request</b></summary>

```json
{
  "error": "validation_failed",
  "message": "Password does not meet security requirements",
  "code": 400
}
```

</details>

<details>
<summary><b>❌ 409 Conflict</b></summary>

```json
{
  "error": "user_exists",
  "message": "User with this email already exists",
  "code": 409
}
```

</details>

---

### 🔐 Login User

**Authenticate user and create session**

```http
POST /api/v1/auth/login
```

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "remember_me": true,
  "device_info": {
    "device_id": "device_123",
    "device_type": "desktop",
    "os": "macOS",
    "browser": "Chrome",
    "app_version": "1.0.0"
  }
}
```

#### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | `string` | ✅ | User email address |
| `password` | `string` | ✅ | User password |
| `remember_me` | `boolean` | ❌ | Extend session duration |
| `device_info` | `object` | ❌ | Device information for tracking |

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com",
    "role": "user",
    "status": "active",
    "last_login_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

</details>

<details>
<summary><b>❌ 401 Unauthorized</b></summary>

```json
{
  "error": "invalid_credentials",
  "message": "Invalid email or password",
  "code": 401
}
```

</details>

<details>
<summary><b>❌ 423 Locked</b></summary>

```json
{
  "error": "account_locked",
  "message": "Account is locked due to too many failed login attempts",
  "code": 423
}
```

</details>

---

### 🚪 Logout User

**Terminate user session**

```http
POST /api/v1/auth/logout
```

#### Headers

```
Authorization: Bearer <access_token>
```

#### Request Body

```json
{
  "session_id": "session_123",
  "all_sessions": false
}
```

#### Request Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `session_id` | `string` | ❌ | Specific session to logout |
| `all_sessions` | `boolean` | ❌ | Logout from all sessions |

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "message": "Logged out successfully",
  "success": true
}
```

</details>

---

### 🔄 Refresh Token

**Get new access token using refresh token**

```http
POST /api/v1/auth/refresh
```

#### Request Body

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

</details>

<details>
<summary><b>❌ 401 Unauthorized</b></summary>

```json
{
  "error": "invalid_refresh_token",
  "message": "Refresh token is invalid or expired",
  "code": 401
}
```

</details>

---

### ✅ Validate Token

**Validate access token and get user info**

```http
POST /api/v1/auth/validate
```

#### Request Body

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "valid": true,
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com",
    "role": "user",
    "status": "active"
  },
  "claims": {
    "user_id": "user_1234567890",
    "email": "user@example.com",
    "role": "user",
    "session_id": "session_123",
    "token_id": "token_456",
    "type": "access",
    "issued_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-01-15T10:45:00Z"
  }
}
```

</details>

<details>
<summary><b>✅ 200 OK (Invalid Token)</b></summary>

```json
{
  "valid": false,
  "message": "Token is expired"
}
```

</details>

---

### 🔒 Change Password

**Change user password**

```http
POST /api/v1/auth/change-password
```

#### Headers

```
Authorization: Bearer <access_token>
```

#### Request Body

```json
{
  "current_password": "OldPass123!",
  "new_password": "NewSecurePass456!"
}
```

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "message": "Password changed successfully",
  "success": true
}
```

</details>

<details>
<summary><b>❌ 400 Bad Request</b></summary>

```json
{
  "error": "invalid_current_password",
  "message": "Current password is incorrect",
  "code": 400
}
```

</details>

---

### 👤 Get User Info

**Get current user information**

```http
GET /api/v1/auth/me
```

#### Headers

```
Authorization: Bearer <access_token>
```

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com",
    "role": "user",
    "status": "active",
    "last_login_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "metadata": {
      "preferred_language": "en",
      "timezone": "UTC"
    }
  }
}
```

</details>

---

## 📊 System Endpoints

### 🏥 Health Check

**Check service health**

```http
GET /health
```

#### Response

<details>
<summary><b>✅ 200 OK</b></summary>

```json
{
  "status": "healthy",
  "service": "auth-service",
  "time": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "dependencies": {
    "redis": "healthy",
    "database": "healthy"
  }
}
```

</details>

---

### 📈 Metrics

**Prometheus metrics**

```http
GET /metrics
```

#### Response

```
# HELP auth_requests_total Total number of authentication requests
# TYPE auth_requests_total counter
auth_requests_total{method="login",status="success"} 1234
auth_requests_total{method="login",status="failed"} 56

# HELP auth_active_sessions Current number of active sessions
# TYPE auth_active_sessions gauge
auth_active_sessions 789
```

---

## 🔧 Error Handling

### Error Response Format

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "code": 400,
  "details": {
    "field": "Additional error details"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| `400` | `validation_failed` | Request validation failed |
| `401` | `invalid_credentials` | Invalid email/password |
| `401` | `invalid_token` | Invalid or expired token |
| `403` | `insufficient_permissions` | User lacks required permissions |
| `409` | `user_exists` | User already exists |
| `423` | `account_locked` | Account is locked |
| `429` | `rate_limit_exceeded` | Too many requests |
| `500` | `internal_error` | Internal server error |

---

## 🔐 Authentication

### Bearer Token

Include the access token in the Authorization header:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Lifecycle

1. **Access Token**: 15 minutes TTL, used for API requests
2. **Refresh Token**: 7 days TTL, used to get new access tokens
3. **Automatic Refresh**: Client should refresh tokens before expiry

---

## 📝 Examples

### Complete Authentication Flow

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}'

# 2. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}'

# 3. Use access token
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8080/api/v1/auth/me

# 4. Refresh token when needed
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

---

<div align="center">

**📚 More Documentation**

[🏠 Main README](./README.md) • [🏗️ Architecture](./architecture.md) • [🛡️ Security](./security.md) • [🚀 Deployment](./deployment.md)

</div>
