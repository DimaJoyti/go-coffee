# 🏗️ Auth Service Architecture Guide

<div align="center">

![Clean Architecture](https://img.shields.io/badge/Clean-Architecture-blue?style=for-the-badge)
![DDD](https://img.shields.io/badge/Domain%20Driven-Design-green?style=for-the-badge)
![SOLID](https://img.shields.io/badge/SOLID-Principles-orange?style=for-the-badge)

**Enterprise-grade microservice architecture following Clean Architecture and DDD principles**

</div>

---

## 🎯 Architecture Overview

The Auth Service follows **Clean Architecture** principles with clear separation of concerns and dependency inversion. The architecture is designed to be:

- **🔄 Maintainable** - Easy to modify and extend
- **🧪 Testable** - High test coverage with isolated units
- **🔌 Flexible** - Pluggable components and interfaces
- **📈 Scalable** - Horizontal scaling capabilities

---

## 🏛️ Layer Architecture

```mermaid
graph TB
    subgraph "🌐 Transport Layer"
        HTTP[HTTP REST API]
        GRPC[gRPC API]
        WS[WebSocket]
    end
    
    subgraph "🎯 Application Layer"
        AS[Auth Service]
        DTO[DTOs]
        INT[Interfaces]
    end
    
    subgraph "🏢 Domain Layer"
        USER[User Entity]
        SESSION[Session Entity]
        TOKEN[Token Entity]
        REPO[Repository Interfaces]
    end
    
    subgraph "🔧 Infrastructure Layer"
        REDIS[Redis Repositories]
        JWT[JWT Service]
        PWD[Password Service]
        SEC[Security Service]
    end
    
    HTTP --> AS
    GRPC --> AS
    AS --> USER
    AS --> SESSION
    AS --> TOKEN
    AS --> REDIS
    AS --> JWT
    AS --> PWD
    AS --> SEC
    
    style HTTP fill:#e1f5fe
    style GRPC fill:#e1f5fe
    style AS fill:#f3e5f5
    style USER fill:#e8f5e8
    style SESSION fill:#e8f5e8
    style TOKEN fill:#e8f5e8
    style REDIS fill:#fff3e0
    style JWT fill:#fff3e0
```

---

## 📦 Domain Layer

### 🎯 Core Entities

<table>
<tr>
<td width="33%">

**👤 User Entity**
```go
type User struct {
    ID                string
    Email             string
    PasswordHash      string
    Role              UserRole
    Status            UserStatus
    FailedLoginCount  int
    LastLoginAt       *time.Time
    LockedUntil       *time.Time
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

</td>
<td width="33%">

**🔐 Session Entity**
```go
type Session struct {
    ID               string
    UserID           string
    AccessToken      string
    RefreshToken     string
    Status           SessionStatus
    ExpiresAt        time.Time
    RefreshExpiresAt time.Time
    DeviceInfo       *DeviceInfo
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

</td>
<td width="33%">

**🎫 Token Entity**
```go
type Token struct {
    ID        string
    UserID    string
    SessionID string
    Type      TokenType
    Status    TokenStatus
    Value     string
    ExpiresAt time.Time
    IssuedAt  time.Time
    CreatedAt time.Time
}
```

</td>
</tr>
</table>

### 🔄 Business Rules

<details>
<summary><b>👤 User Business Rules</b></summary>

- **Email Validation**: Must be valid email format
- **Password Policy**: Minimum 8 chars, mixed case, numbers, symbols
- **Account Locking**: Lock after 5 failed login attempts
- **Status Management**: Active, inactive, locked, suspended states
- **Role Assignment**: User, admin roles with different permissions

</details>

<details>
<summary><b>🔐 Session Business Rules</b></summary>

- **Single User, Multiple Sessions**: Users can have multiple active sessions
- **Session Expiration**: Access tokens expire in 15 minutes
- **Refresh Token Lifecycle**: Refresh tokens expire in 7 days
- **Device Tracking**: Track device information for security
- **Session Revocation**: Can revoke individual or all sessions

</details>

<details>
<summary><b>🎫 Token Business Rules</b></summary>

- **JWT Structure**: Header, payload, signature with HS256
- **Claims Validation**: User ID, email, role, session ID, expiration
- **Token Types**: Access tokens for API calls, refresh for renewal
- **Blacklisting**: Revoked tokens added to blacklist
- **Signature Verification**: All tokens must have valid signatures

</details>

---

## 🎯 Application Layer

### 🔧 Service Architecture

```mermaid
graph LR
    subgraph "🎯 Application Services"
        AS[AuthService]
        VS[ValidationService]
        NS[NotificationService]
    end
    
    subgraph "🔌 Infrastructure Services"
        JS[JWTService]
        PS[PasswordService]
        SS[SecurityService]
        CS[CacheService]
    end
    
    subgraph "🗄️ Repositories"
        UR[UserRepository]
        SR[SessionRepository]
        TR[TokenRepository]
    end
    
    AS --> JS
    AS --> PS
    AS --> SS
    AS --> UR
    AS --> SR
    AS --> TR
    AS --> VS
    AS --> NS
    AS --> CS
```

### 📋 Use Cases

<table>
<tr>
<td width="50%">

**🔑 Authentication Use Cases**
- ✅ Register new user
- ✅ Login with credentials
- ✅ Logout and revoke session
- ✅ Change password
- ✅ Validate user status

</td>
<td width="50%">

**🎫 Token Management Use Cases**
- ✅ Generate JWT token pair
- ✅ Validate access token
- ✅ Refresh expired token
- ✅ Revoke tokens
- ✅ Blacklist management

</td>
</tr>
<tr>
<td width="50%">

**🔐 Session Management Use Cases**
- ✅ Create user session
- ✅ Track session activity
- ✅ Manage multiple sessions
- ✅ Session cleanup
- ✅ Device tracking

</td>
<td width="50%">

**🛡️ Security Use Cases**
- ✅ Rate limiting
- ✅ Account lockout
- ✅ Security event logging
- ✅ Failed login tracking
- ✅ Suspicious activity detection

</td>
</tr>
</table>

---

## 🔧 Infrastructure Layer

### 🗄️ Data Storage Architecture

```mermaid
graph TB
    subgraph "📊 Redis Data Store"
        subgraph "👤 User Data"
            U1[auth:users:{userID}]
            U2[auth:users:email:{email}]
        end
        
        subgraph "🔐 Session Data"
            S1[auth:sessions:{sessionID}]
            S2[auth:access_tokens:{token}]
            S3[auth:refresh_tokens:{token}]
            S4[auth:user_sessions:{userID}]
        end
        
        subgraph "🛡️ Security Data"
            SEC1[auth:failed_login:{email}]
            SEC2[auth:blacklist:{tokenID}]
            SEC3[auth:rate_limit:{key}]
        end
    end
    
    subgraph "🔌 Repository Layer"
        UR[UserRepository]
        SR[SessionRepository]
        TR[TokenRepository]
    end
    
    UR --> U1
    UR --> U2
    SR --> S1
    SR --> S2
    SR --> S3
    SR --> S4
    TR --> SEC2
```

### 🔐 Security Services

<table>
<tr>
<td width="33%">

**🔑 JWT Service**
- Token generation
- Token validation
- Claims parsing
- Signature verification
- Expiration handling

</td>
<td width="33%">

**🔒 Password Service**
- bcrypt hashing
- Password validation
- Policy enforcement
- Strength checking
- Common password detection

</td>
<td width="33%">

**🛡️ Security Service**
- Rate limiting
- Event logging
- Account lockout
- Failed login tracking
- Suspicious activity detection

</td>
</tr>
</table>

---

## 🌐 Transport Layer

### 🔄 API Architecture

```mermaid
graph TB
    subgraph "🌐 External Clients"
        WEB[Web Application]
        MOBILE[Mobile App]
        API[API Clients]
        SERVICE[Other Services]
    end
    
    subgraph "🚪 API Gateway"
        GATEWAY[Load Balancer / API Gateway]
    end
    
    subgraph "🎯 Auth Service"
        HTTP[HTTP Server :8080]
        GRPC[gRPC Server :50053]
        MIDDLEWARE[Middleware Stack]
    end
    
    WEB --> GATEWAY
    MOBILE --> GATEWAY
    API --> GATEWAY
    SERVICE --> GRPC
    
    GATEWAY --> HTTP
    HTTP --> MIDDLEWARE
    GRPC --> MIDDLEWARE
    
    MIDDLEWARE --> AS[Auth Service]
```

### 🔧 Middleware Stack

<table>
<tr>
<td width="25%">

**📊 Logging**
- Request/response logging
- Structured logging
- Correlation IDs
- Performance metrics

</td>
<td width="25%">

**🛡️ Security**
- CORS handling
- Security headers
- Rate limiting
- Input validation

</td>
<td width="25%">

**📈 Monitoring**
- Prometheus metrics
- Health checks
- Distributed tracing
- Error tracking

</td>
<td width="25%">

**🔄 Recovery**
- Panic recovery
- Graceful degradation
- Circuit breakers
- Retry mechanisms

</td>
</tr>
</table>

---

## 🔄 Data Flow

### 🔐 Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant H as HTTP Handler
    participant A as Auth Service
    participant U as User Repository
    participant S as Session Repository
    participant J as JWT Service
    participant R as Redis
    
    C->>H: POST /auth/login
    H->>A: Login(request)
    A->>U: GetUserByEmail(email)
    U->>R: GET auth:users:email:{email}
    R-->>U: userID
    U->>R: GET auth:users:{userID}
    R-->>U: user data
    U-->>A: User entity
    A->>A: Validate password
    A->>J: GenerateTokenPair(user)
    J-->>A: access_token, refresh_token
    A->>S: CreateSession(session)
    S->>R: SET auth:sessions:{sessionID}
    S->>R: SET auth:access_tokens:{token}
    S->>R: SET auth:refresh_tokens:{token}
    A-->>H: LoginResponse
    H-->>C: 200 OK + tokens
```

### 🔄 Token Refresh Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant H as HTTP Handler
    participant A as Auth Service
    participant J as JWT Service
    participant S as Session Repository
    participant R as Redis
    
    C->>H: POST /auth/refresh
    H->>A: RefreshToken(request)
    A->>J: ValidateRefreshToken(token)
    J-->>A: token claims
    A->>S: GetSessionByRefreshToken(token)
    S->>R: GET auth:refresh_tokens:{token}
    R-->>S: sessionID
    S->>R: GET auth:sessions:{sessionID}
    R-->>S: session data
    S-->>A: Session entity
    A->>A: Validate session
    A->>J: GenerateTokenPair(user)
    J-->>A: new tokens
    A->>S: UpdateSession(session)
    S->>R: SET auth:sessions:{sessionID}
    A-->>H: RefreshResponse
    H-->>C: 200 OK + new tokens
```

---

## 🔌 Dependency Injection

### 🏗️ Service Construction

```go
// Main service initialization
func InitializeAuthService() *AuthService {
    // Infrastructure
    redisClient := initRedis()
    logger := initLogger()
    
    // Repositories
    userRepo := repository.NewRedisUserRepository(redisClient, logger)
    sessionRepo := repository.NewRedisSessionRepository(redisClient, logger)
    
    // Services
    jwtService := security.NewJWTService(jwtConfig, logger)
    passwordService := security.NewPasswordService(passwordConfig, logger)
    securityService := security.NewSecurityService(securityConfig, logger)
    
    // Application service
    return application.NewAuthService(
        userRepo,
        sessionRepo,
        jwtService,
        passwordService,
        securityService,
        authConfig,
        logger,
    )
}
```

### 🔄 Interface Segregation

```go
// Small, focused interfaces
type UserRepository interface {
    CreateUser(ctx context.Context, user *User) error
    GetUserByEmail(ctx context.Context, email string) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
}

type JWTService interface {
    GenerateAccessToken(ctx context.Context, user *User, sessionID string) (string, error)
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type PasswordService interface {
    HashPassword(password string) (string, error)
    VerifyPassword(hashedPassword, password string) error
}
```

---

## 📊 Scalability Considerations

### 🔄 Horizontal Scaling

<table>
<tr>
<td width="50%">

**✅ Stateless Design**
- No server-side sessions
- JWT tokens contain all needed info
- Redis for shared state
- Load balancer friendly

</td>
<td width="50%">

**📈 Performance Optimizations**
- Connection pooling
- Redis clustering
- Token caching
- Async operations

</td>
</tr>
<tr>
<td width="50%">

**🔄 High Availability**
- Multiple service instances
- Redis replication
- Health checks
- Graceful shutdowns

</td>
<td width="50%">

**📊 Monitoring & Observability**
- Prometheus metrics
- Distributed tracing
- Structured logging
- Performance dashboards

</td>
</tr>
</table>

---

## 🧪 Testing Strategy

### 🏗️ Testing Pyramid

```mermaid
graph TB
    subgraph "🧪 Testing Levels"
        E2E[E2E Tests<br/>API Integration]
        INT[Integration Tests<br/>Repository + Redis]
        UNIT[Unit Tests<br/>Domain + Application]
    end
    
    style E2E fill:#ffcdd2
    style INT fill:#fff3e0
    style UNIT fill:#e8f5e8
```

### 📋 Test Coverage

- **Unit Tests**: Domain entities, application services
- **Integration Tests**: Repository implementations, external services
- **API Tests**: HTTP endpoints, request/response validation
- **Security Tests**: Authentication flows, authorization checks
- **Performance Tests**: Load testing, stress testing

---

<div align="center">

**🏗️ Architecture Documentation**

[🏠 Main README](./README.md) • [📖 API Reference](./api-reference.md) • [🛡️ Security](./security.md) • [🚀 Deployment](./deployment.md)

</div>
