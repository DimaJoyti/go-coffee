# Tenant Infrastructure Middleware

This package provides comprehensive multi-tenant middleware for Go applications, supporting both Gin and standard HTTP handlers.

## Overview

The tenant infrastructure has been completely refactored to provide:

1. **Clean separation of concerns** between Gin and HTTP middleware
2. **Proper JWT token handling** with configurable validation
3. **Complete tenant isolation** with URL parameter validation
4. **Standardized error handling** across all middleware
5. **Comprehensive feature and subscription management**

## Components

### Core Components

- `TenantContextMiddleware` - Core tenant context extraction and validation
- `HTTPTenantMiddleware` - HTTP-specific middleware functions
- `TenantJWTExtractor` - JWT-based tenant ID extraction
- `TenantIsolationMiddleware` - Data isolation enforcement
- `TenantMetricsMiddleware` - Tenant-specific metrics collection

### Files

- `tenant_context.go` - Core tenant middleware (Gin-focused)
- `tenant_http.go` - HTTP-specific middleware
- `tenant_jwt.go` - JWT token handling

## Usage Examples

### Basic HTTP Setup

```go
package main

import (
    "net/http"
    "github.com/DimaJoyti/go-coffee/infrastructure/middleware"
    "github.com/DimaJoyti/go-coffee/infrastructure/persistence"
)

func main() {
    // Setup tenant repository
    tenantRepo := persistence.NewTenantRepository(db)
    
    // Create HTTP tenant middleware
    httpMiddleware := middleware.NewHTTPTenantMiddleware(tenantRepo, nil)
    
    // Setup routes with tenant context
    mux := http.NewServeMux()
    
    // Protected route that requires tenant context
    protectedHandler := httpMiddleware.TenantContext(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantCtx, _ := middleware.GetHTTPTenantContext(r)
            w.Write([]byte("Hello " + tenantCtx.TenantName()))
        }),
    )
    
    mux.Handle("/api/protected", protectedHandler)
    
    http.ListenAndServe(":8080", mux)
}
```

### JWT-Based Authentication

```go
package main

import (
    "crypto/rsa"
    "net/http"
    "time"
    "github.com/DimaJoyti/go-coffee/infrastructure/middleware"
)

func main() {
    // Setup JWT extractor
    jwtConfig := middleware.TenantJWTConfig{
        PublicKey:   publicKey, // Your RSA public key
        Issuer:      "your-auth-service",
        Audience:    "your-api",
        ClockSkew:   5 * time.Minute,
        TenantClaim: "tenant_id",
    }
    jwtExtractor := middleware.NewTenantJWTExtractor(jwtConfig)
    
    // Create HTTP middleware with JWT support
    httpMiddleware := middleware.NewHTTPTenantMiddleware(tenantRepo, jwtExtractor)
    
    // Setup JWT-protected route
    mux := http.NewServeMux()
    
    jwtProtectedHandler := httpMiddleware.JWTTenantContext(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantCtx, _ := middleware.GetHTTPTenantContext(r)
            w.Write([]byte("JWT Authenticated: " + tenantCtx.TenantName()))
        }),
    )
    
    mux.Handle("/api/jwt-protected", jwtProtectedHandler)
    
    http.ListenAndServe(":8080", mux)
}
```

### Gin Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/DimaJoyti/go-coffee/infrastructure/middleware"
)

func main() {
    r := gin.Default()
    
    // Setup tenant middleware
    tenantMiddleware := middleware.NewTenantContextMiddleware(tenantRepo)
    
    // Apply tenant context middleware
    r.Use(tenantMiddleware.GinTenantContext())
    
    // Protected routes
    api := r.Group("/api")
    {
        // Require specific feature
        api.Use(tenantMiddleware.RequireFeature("advanced_analytics"))
        api.GET("/analytics", func(c *gin.Context) {
            tenantCtx, _ := middleware.GetTenantContext(c)
            c.JSON(200, gin.H{"tenant": tenantCtx.TenantName()})
        })
        
        // Require specific subscription
        premium := api.Group("/premium")
        premium.Use(tenantMiddleware.RequireSubscription(shared.SubscriptionPlanPremium))
        premium.GET("/features", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "Premium features"})
        })
    }
    
    r.Run(":8080")
}
```

### Feature and Subscription Control

```go
// HTTP middleware for feature requirements
featureMiddleware := httpMiddleware.RequireFeature("advanced_reporting")
subscriptionMiddleware := httpMiddleware.RequireSubscription(shared.SubscriptionPlanEnterprise)

// Chain middlewares
handler := featureMiddleware(
    subscriptionMiddleware(
        http.HandlerFunc(yourHandler),
    ),
)

mux.Handle("/api/enterprise-feature", handler)
```

### Tenant Isolation

```go
// Ensure tenant can only access their own resources
isolationHandler := httpMiddleware.ValidateTenantAccess(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // This handler will only execute if the tenant in the URL
        // matches the tenant in the context
        w.Write([]byte("Access granted to tenant resource"))
    }),
)

// URL pattern: /api/tenants/{tenant_id}/orders
mux.Handle("/api/tenants/", isolationHandler)
```

## Tenant ID Extraction Methods

The middleware supports multiple methods for extracting tenant IDs:

1. **Subdomain**: `tenant1.api.example.com`
2. **Custom Header**: `X-Tenant-ID: tenant1`
3. **JWT Token**: Extracted from `Authorization: Bearer <token>`
4. **Query Parameter**: `?tenant_id=tenant1`
5. **URL Path**: `/api/v1/tenants/tenant1/orders`

## Error Handling

All middleware components provide consistent JSON error responses:

```json
{
  "error": "Unauthorized",
  "message": "Tenant not found"
}
```

For feature/subscription errors:

```json
{
  "error": "Forbidden",
  "message": "Feature not available for current subscription",
  "feature": "advanced_analytics"
}
```

## Metrics Collection

Enable tenant-specific metrics collection:

```go
metricsHandler := middleware.NewTenantMetricsHandler()

// HTTP
handler = metricsHandler.CollectMetrics(handler)

// Gin
r.Use(middleware.NewTenantMetricsMiddleware().CollectMetrics())
```

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_PUBLIC_KEY_PATH=/path/to/public.key
JWT_ISSUER=your-auth-service
JWT_AUDIENCE=your-api
JWT_CLOCK_SKEW=300s

# Tenant Configuration
TENANT_ISOLATION_LEVEL=shared_database
TENANT_DEFAULT_SUBSCRIPTION=basic
```

## Best Practices

1. **Always validate tenant context** before accessing tenant-specific resources
2. **Use JWT extraction** for stateless authentication
3. **Implement proper error handling** for all tenant-related operations
4. **Monitor tenant metrics** for usage patterns and billing
5. **Test tenant isolation** thoroughly to prevent data leaks

## Migration from Old Implementation

If you were using the old `tenant_context_http.go` file:

1. Replace `NewTenantContextMiddleware` with `NewHTTPTenantMiddleware`
2. Use `HTTPTenantMiddleware.TenantContext()` instead of `HTTPTenantContext()`
3. Update error handling to use the new JSON response format
4. Consider adding JWT support for better security

## Testing

See the test files for comprehensive examples of testing tenant middleware functionality.
