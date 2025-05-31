# 💡 Auth Service Examples

<div align="center">

![Examples](https://img.shields.io/badge/Examples-Ready%20to%20Use-green?style=for-the-badge)
![cURL](https://img.shields.io/badge/cURL-Commands-blue?style=for-the-badge)
![Postman](https://img.shields.io/badge/Postman-Collection-orange?style=for-the-badge)

**Complete examples and code snippets for Auth Service integration**

</div>

---

## 🚀 Quick Start Examples

### 🔐 Complete Authentication Flow

```bash
#!/bin/bash
# Complete authentication flow example

BASE_URL="http://localhost:8080"

echo "🔐 Auth Service Example Flow"
echo "=========================="

# 1. Register a new user
echo "📝 1. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }')

echo "✅ Registration Response:"
echo "$REGISTER_RESPONSE" | jq .

# Extract access token
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token')

echo "🎫 Access Token: ${ACCESS_TOKEN:0:50}..."
echo "🔄 Refresh Token: ${REFRESH_TOKEN:0:50}..."

# 2. Get user info
echo -e "\n👤 2. Getting user info..."
USER_INFO=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "✅ User Info:"
echo "$USER_INFO" | jq .

# 3. Refresh token
echo -e "\n🔄 3. Refreshing token..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

echo "✅ Refresh Response:"
echo "$REFRESH_RESPONSE" | jq .

# Extract new tokens
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')

# 4. Validate token
echo -e "\n✅ 4. Validating token..."
VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$NEW_ACCESS_TOKEN\"}")

echo "✅ Validation Response:"
echo "$VALIDATE_RESPONSE" | jq .

# 5. Logout
echo -e "\n🚪 5. Logging out..."
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"all_sessions": false}')

echo "✅ Logout Response:"
echo "$LOGOUT_RESPONSE" | jq .

echo -e "\n🎉 Authentication flow completed!"
```

---

## 📋 Individual API Examples

### 👤 User Registration

<details>
<summary><b>🔐 Basic Registration</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

**Response:**
```json
{
  "user": {
    "id": "user_1705123456789",
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
<summary><b>👑 Admin Registration</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "AdminPass123!",
    "role": "admin"
  }'
```

</details>

### 🔐 User Login

<details>
<summary><b>🚪 Basic Login</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

</details>

<details>
<summary><b>📱 Login with Device Info</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "remember_me": true,
    "device_info": {
      "device_id": "device_12345",
      "device_type": "desktop",
      "os": "macOS",
      "browser": "Chrome",
      "app_version": "1.0.0"
    }
  }'
```

</details>

### 🔄 Token Management

<details>
<summary><b>🔄 Refresh Token</b></summary>

```bash
# Store refresh token from login response
REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}"
```

</details>

<details>
<summary><b>✅ Validate Token</b></summary>

```bash
# Store access token from login response
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}"
```

</details>

### 👤 User Management

<details>
<summary><b>👤 Get User Info</b></summary>

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

</details>

<details>
<summary><b>🔒 Change Password</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

</details>

### 🚪 Logout

<details>
<summary><b>🚪 Single Session Logout</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session_12345",
    "all_sessions": false
  }'
```

</details>

<details>
<summary><b>🚪 All Sessions Logout</b></summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "all_sessions": true
  }'
```

</details>

---

## 🐍 Python Examples

### 🔐 Python Authentication Client

```python
import requests
import json
from typing import Optional, Dict, Any

class AuthClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
    
    def register(self, email: str, password: str, role: str = "user") -> Dict[str, Any]:
        """Register a new user"""
        response = requests.post(
            f"{self.base_url}/api/v1/auth/register",
            json={
                "email": email,
                "password": password,
                "role": role
            }
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data["access_token"]
        self.refresh_token = data["refresh_token"]
        
        return data
    
    def login(self, email: str, password: str, remember_me: bool = False) -> Dict[str, Any]:
        """Login user"""
        response = requests.post(
            f"{self.base_url}/api/v1/auth/login",
            json={
                "email": email,
                "password": password,
                "remember_me": remember_me
            }
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data["access_token"]
        self.refresh_token = data["refresh_token"]
        
        return data
    
    def refresh_token_method(self) -> Dict[str, Any]:
        """Refresh access token"""
        if not self.refresh_token:
            raise ValueError("No refresh token available")
        
        response = requests.post(
            f"{self.base_url}/api/v1/auth/refresh",
            json={"refresh_token": self.refresh_token}
        )
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data["access_token"]
        self.refresh_token = data["refresh_token"]
        
        return data
    
    def get_user_info(self) -> Dict[str, Any]:
        """Get current user information"""
        if not self.access_token:
            raise ValueError("No access token available")
        
        response = requests.get(
            f"{self.base_url}/api/v1/auth/me",
            headers={"Authorization": f"Bearer {self.access_token}"}
        )
        response.raise_for_status()
        
        return response.json()
    
    def change_password(self, current_password: str, new_password: str) -> Dict[str, Any]:
        """Change user password"""
        if not self.access_token:
            raise ValueError("No access token available")
        
        response = requests.post(
            f"{self.base_url}/api/v1/auth/change-password",
            headers={"Authorization": f"Bearer {self.access_token}"},
            json={
                "current_password": current_password,
                "new_password": new_password
            }
        )
        response.raise_for_status()
        
        return response.json()
    
    def logout(self, all_sessions: bool = False) -> Dict[str, Any]:
        """Logout user"""
        if not self.access_token:
            raise ValueError("No access token available")
        
        response = requests.post(
            f"{self.base_url}/api/v1/auth/logout",
            headers={"Authorization": f"Bearer {self.access_token}"},
            json={"all_sessions": all_sessions}
        )
        response.raise_for_status()
        
        # Clear tokens after logout
        self.access_token = None
        self.refresh_token = None
        
        return response.json()

# Usage example
if __name__ == "__main__":
    client = AuthClient()
    
    # Register new user
    print("🔐 Registering user...")
    register_response = client.register("demo@example.com", "SecurePass123!")
    print(f"✅ User registered: {register_response['user']['email']}")
    
    # Get user info
    print("\n👤 Getting user info...")
    user_info = client.get_user_info()
    print(f"✅ User info: {user_info['user']['email']}")
    
    # Refresh token
    print("\n🔄 Refreshing token...")
    refresh_response = client.refresh_token_method()
    print(f"✅ Token refreshed, expires in: {refresh_response['expires_in']}s")
    
    # Logout
    print("\n🚪 Logging out...")
    logout_response = client.logout()
    print(f"✅ Logged out: {logout_response['message']}")
```

---

## 🟨 JavaScript Examples

### 🌐 JavaScript Authentication Client

```javascript
class AuthClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.accessToken = null;
        this.refreshToken = null;
    }

    async register(email, password, role = 'user') {
        const response = await fetch(`${this.baseUrl}/api/v1/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password, role }),
        });

        if (!response.ok) {
            throw new Error(`Registration failed: ${response.statusText}`);
        }

        const data = await response.json();
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token;

        return data;
    }

    async login(email, password, rememberMe = false, deviceInfo = null) {
        const response = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email,
                password,
                remember_me: rememberMe,
                device_info: deviceInfo,
            }),
        });

        if (!response.ok) {
            throw new Error(`Login failed: ${response.statusText}`);
        }

        const data = await response.json();
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token;

        return data;
    }

    async refreshTokenMethod() {
        if (!this.refreshToken) {
            throw new Error('No refresh token available');
        }

        const response = await fetch(`${this.baseUrl}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refresh_token: this.refreshToken }),
        });

        if (!response.ok) {
            throw new Error(`Token refresh failed: ${response.statusText}`);
        }

        const data = await response.json();
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token;

        return data;
    }

    async getUserInfo() {
        if (!this.accessToken) {
            throw new Error('No access token available');
        }

        const response = await fetch(`${this.baseUrl}/api/v1/auth/me`, {
            headers: {
                'Authorization': `Bearer ${this.accessToken}`,
            },
        });

        if (!response.ok) {
            throw new Error(`Get user info failed: ${response.statusText}`);
        }

        return await response.json();
    }

    async logout(allSessions = false) {
        if (!this.accessToken) {
            throw new Error('No access token available');
        }

        const response = await fetch(`${this.baseUrl}/api/v1/auth/logout`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.accessToken}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ all_sessions: allSessions }),
        });

        if (!response.ok) {
            throw new Error(`Logout failed: ${response.statusText}`);
        }

        // Clear tokens after logout
        this.accessToken = null;
        this.refreshToken = null;

        return await response.json();
    }
}

// Usage example
(async () => {
    const client = new AuthClient();

    try {
        // Register new user
        console.log('🔐 Registering user...');
        const registerResponse = await client.register('demo@example.com', 'SecurePass123!');
        console.log('✅ User registered:', registerResponse.user.email);

        // Get user info
        console.log('\n👤 Getting user info...');
        const userInfo = await client.getUserInfo();
        console.log('✅ User info:', userInfo.user.email);

        // Logout
        console.log('\n🚪 Logging out...');
        const logoutResponse = await client.logout();
        console.log('✅ Logged out:', logoutResponse.message);

    } catch (error) {
        console.error('❌ Error:', error.message);
    }
})();
```

---

## 📊 Monitoring Examples

### 🔍 Health Check

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl "http://localhost:8080/health?detailed=true"
```

### 📈 Metrics

```bash
# Get Prometheus metrics
curl http://localhost:8080/metrics

# Filter specific metrics
curl http://localhost:8080/metrics | grep auth_requests_total
```

---

## 🧪 Testing Examples

### 🔬 Load Testing with Apache Bench

```bash
# Test registration endpoint
ab -n 1000 -c 10 -T 'application/json' -p register.json \
   http://localhost:8080/api/v1/auth/register

# Test login endpoint
ab -n 1000 -c 10 -T 'application/json' -p login.json \
   http://localhost:8080/api/v1/auth/login
```

### 📋 Test Data Files

**register.json:**
```json
{
  "email": "test@example.com",
  "password": "TestPass123!",
  "role": "user"
}
```

**login.json:**
```json
{
  "email": "test@example.com",
  "password": "TestPass123!"
}
```

---

<div align="center">

**💡 More Examples Coming Soon!**

[🏠 Main README](./README.md) • [📖 API Reference](./api-reference.md) • [🏗️ Architecture](./architecture.md)

</div>
