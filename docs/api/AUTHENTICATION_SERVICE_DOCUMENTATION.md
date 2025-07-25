# 🔐 Authentication Service - Complete Documentation

## 📋 Overview

The Authentication Service is responsible for user authentication, authorization, session management, and security enforcement across the Go Coffee platform. It provides JWT-based authentication with refresh tokens, role-based access control (RBAC), and multi-factor authentication (MFA) capabilities.

## 🏗️ Architecture

### **Core Components**
- **User Management**: Registration, profile management, password handling
- **Authentication Engine**: Login, logout, token generation/validation
- **Authorization System**: Role-based access control and permissions
- **Session Management**: Token lifecycle and refresh mechanisms
- **Security Features**: MFA, rate limiting, audit logging
- **Password Security**: Bcrypt hashing, complexity validation, breach detection

### **Security Features**
- JWT tokens with RS256 signing
- Refresh token rotation
- Multi-factor authentication (TOTP, SMS)
- Account lockout protection
- Password breach detection
- Audit logging for security events

## 🔗 API Endpoints

### **1. User Registration**

#### **POST /api/v1/auth/register**
Registers a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "confirm_password": "SecurePassword123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "date_of_birth": "1990-01-15",
  "marketing_consent": true,
  "terms_accepted": true
}
```

**Response:**
```json
{
  "user_id": "usr_123e4567e89b12d3",
  "email": "user@example.com",
  "status": "pending_verification",
  "verification_sent": true,
  "message": "Please check your email to verify your account"
}
```

**Validation Rules:**
- Email must be valid and unique
- Password must meet complexity requirements (8+ chars, uppercase, lowercase, number, special char)
- Phone number must be valid format
- User must be 13+ years old
- Terms acceptance is required

#### **POST /api/v1/auth/verify-email**
Verifies user email address with verification token.

**Request:**
```json
{
  "email": "user@example.com",
  "verification_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "message": "Email verified successfully",
  "user_id": "usr_123e4567e89b12d3",
  "status": "active"
}
```

### **2. Authentication**

#### **POST /api/v1/auth/login**
Authenticates user credentials and returns JWT tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "remember_me": true,
  "device_info": {
    "device_id": "dev_mobile_123",
    "device_name": "iPhone 13",
    "platform": "iOS",
    "app_version": "1.2.3"
  }
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "refresh_expires_in": 604800,
  "token_type": "Bearer",
  "user": {
    "id": "usr_123e4567e89b12d3",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "customer",
    "permissions": ["order:create", "order:read", "profile:update"],
    "mfa_enabled": false,
    "email_verified": true,
    "last_login": "2024-01-15T10:30:00Z"
  },
  "session_id": "sess_789abc123def456"
}
```

#### **POST /api/v1/auth/mfa/verify**
Verifies multi-factor authentication code during login.

**Request:**
```json
{
  "session_id": "sess_789abc123def456",
  "mfa_code": "123456",
  "mfa_type": "totp"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### **POST /api/v1/auth/refresh**
Refreshes an expired access token using a refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### **POST /api/v1/auth/logout**
Invalidates the current session and tokens.

**Request:**
```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{
  "logout_all_devices": false
}
```

**Response:**
```json
{
  "message": "Successfully logged out",
  "logged_out_sessions": 1
}
```

### **3. Password Management**

#### **POST /api/v1/auth/password/forgot**
Initiates password reset process.

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "message": "If an account with this email exists, a password reset link has been sent",
  "reset_token_expires": "2024-01-15T11:30:00Z"
}
```

#### **POST /api/v1/auth/password/reset**
Resets password using reset token.

**Request:**
```json
{
  "reset_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "new_password": "NewSecurePassword123!",
  "confirm_password": "NewSecurePassword123!"
}
```

**Response:**
```json
{
  "message": "Password reset successfully",
  "user_id": "usr_123e4567e89b12d3"
}
```

#### **POST /api/v1/auth/password/change**
Changes password for authenticated user.

**Request:**
```http
POST /api/v1/auth/password/change
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{
  "current_password": "CurrentPassword123!",
  "new_password": "NewSecurePassword123!",
  "confirm_password": "NewSecurePassword123!"
}
```

**Response:**
```json
{
  "message": "Password changed successfully",
  "logout_other_sessions": true
}
```

### **4. Multi-Factor Authentication**

#### **POST /api/v1/auth/mfa/setup**
Sets up multi-factor authentication for user.

**Request:**
```http
POST /api/v1/auth/mfa/setup
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{
  "mfa_type": "totp",
  "phone_number": "+1234567890"
}
```

**Response:**
```json
{
  "mfa_type": "totp",
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "backup_codes": [
    "12345678",
    "87654321",
    "11223344",
    "44332211",
    "55667788"
  ]
}
```

#### **POST /api/v1/auth/mfa/enable**
Enables MFA after verification.

**Request:**
```json
{
  "mfa_code": "123456",
  "backup_codes_acknowledged": true
}
```

**Response:**
```json
{
  "message": "Multi-factor authentication enabled successfully",
  "mfa_enabled": true
}
```

#### **POST /api/v1/auth/mfa/disable**
Disables multi-factor authentication.

**Request:**
```json
{
  "password": "CurrentPassword123!",
  "mfa_code": "123456"
}
```

**Response:**
```json
{
  "message": "Multi-factor authentication disabled successfully",
  "mfa_enabled": false
}
```

### **5. User Profile Management**

#### **GET /api/v1/auth/profile**
Retrieves current user profile.

**Request:**
```http
GET /api/v1/auth/profile
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "user": {
    "id": "usr_123e4567e89b12d3",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "date_of_birth": "1990-01-15",
    "role": "customer",
    "status": "active",
    "email_verified": true,
    "phone_verified": false,
    "mfa_enabled": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "last_login": "2024-01-15T10:30:00Z",
    "login_count": 42,
    "preferences": {
      "language": "en",
      "timezone": "America/New_York",
      "marketing_emails": true,
      "push_notifications": true
    }
  }
}
```

#### **PUT /api/v1/auth/profile**
Updates user profile information.

**Request:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+1234567890",
  "preferences": {
    "language": "en",
    "timezone": "America/Los_Angeles",
    "marketing_emails": false,
    "push_notifications": true
  }
}
```

**Response:**
```json
{
  "message": "Profile updated successfully",
  "user": {
    "id": "usr_123e4567e89b12d3",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Smith",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

### **6. Session Management**

#### **GET /api/v1/auth/sessions**
Lists active sessions for the user.

**Request:**
```http
GET /api/v1/auth/sessions
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "sessions": [
    {
      "session_id": "sess_789abc123def456",
      "device_info": {
        "device_name": "iPhone 13",
        "platform": "iOS",
        "app_version": "1.2.3"
      },
      "ip_address": "192.168.1.100",
      "location": "New York, NY, US",
      "created_at": "2024-01-15T10:30:00Z",
      "last_activity": "2024-01-15T10:35:00Z",
      "is_current": true
    }
  ],
  "total_sessions": 1
}
```

#### **DELETE /api/v1/auth/sessions/{session_id}**
Terminates a specific session.

**Request:**
```http
DELETE /api/v1/auth/sessions/sess_789abc123def456
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "message": "Session terminated successfully",
  "session_id": "sess_789abc123def456"
}
```

### **7. Admin Endpoints**

#### **GET /api/v1/auth/admin/users**
Lists users (admin only).

**Request:**
```http
GET /api/v1/auth/admin/users?page=1&limit=20&status=active&role=customer
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "users": [
    {
      "id": "usr_123e4567e89b12d3",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "customer",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "last_login": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### **PUT /api/v1/auth/admin/users/{user_id}/status**
Updates user status (admin only).

**Request:**
```json
{
  "status": "suspended",
  "reason": "Terms of service violation",
  "notify_user": true
}
```

**Response:**
```json
{
  "message": "User status updated successfully",
  "user_id": "usr_123e4567e89b12d3",
  "new_status": "suspended"
}
```

## 🔒 Security Features

### **JWT Token Structure**
```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-id-123"
  },
  "payload": {
    "sub": "usr_123e4567e89b12d3",
    "email": "user@example.com",
    "role": "customer",
    "permissions": ["order:create", "order:read"],
    "session_id": "sess_789abc123def456",
    "iat": 1642248600,
    "exp": 1642252200,
    "iss": "go-coffee-auth",
    "aud": "go-coffee-api"
  }
}
```

### **Password Security**
- **Hashing**: Bcrypt with cost factor 12
- **Complexity**: Minimum 8 characters, mixed case, numbers, special characters
- **Breach Detection**: Integration with HaveIBeenPwned API
- **History**: Prevents reuse of last 12 passwords
- **Expiry**: Optional password expiry (90 days for admin accounts)

### **Rate Limiting**
- **Login Attempts**: 5 attempts per 15 minutes per IP
- **Password Reset**: 3 requests per hour per email
- **Registration**: 10 registrations per hour per IP
- **MFA Attempts**: 5 attempts per 5 minutes per session

### **Account Security**
- **Lockout**: Account locked after 10 failed login attempts
- **Unlock**: Automatic unlock after 30 minutes or admin intervention
- **Suspicious Activity**: Automatic alerts for unusual login patterns
- **Device Tracking**: New device notifications

## 📊 Monitoring and Metrics

### **Security Metrics**
- Failed login attempts
- Account lockouts
- Password reset requests
- MFA setup/usage rates
- Suspicious activity alerts

### **Performance Metrics**
- Authentication response times
- Token validation latency
- Database query performance
- Cache hit rates

### **Business Metrics**
- User registration rates
- Login success rates
- Session duration
- MFA adoption rates

## 🔧 Configuration

### **Environment Variables**
```bash
# JWT Configuration
JWT_PRIVATE_KEY_PATH=/secrets/jwt-private.pem
JWT_PUBLIC_KEY_PATH=/secrets/jwt-public.pem
JWT_ACCESS_TOKEN_EXPIRY=3600
JWT_REFRESH_TOKEN_EXPIRY=604800

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/auth_db
REDIS_URL=redis://localhost:6379

# Security
BCRYPT_COST=12
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=1800
PASSWORD_MIN_LENGTH=8

# MFA
MFA_ISSUER=Go Coffee
MFA_BACKUP_CODES_COUNT=10

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=noreply@gocoffee.com
SMTP_PASSWORD=app-password

# Rate Limiting
RATE_LIMIT_LOGIN=5
RATE_LIMIT_WINDOW=900
```

## 🚀 Deployment

### **Database Schema**
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    date_of_birth DATE,
    role VARCHAR(50) DEFAULT 'customer',
    status VARCHAR(50) DEFAULT 'pending_verification',
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    last_login TIMESTAMP,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW()
);

-- Password history table
CREATE TABLE password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```
