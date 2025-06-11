# Infrastructure Layer - Fixed and Enhanced

This document describes the fixes and enhancements made to the infrastructure layer of the Go Coffee platform.

## 🔧 Issues Fixed

### 1. **Dependency Management**
- ✅ Removed conflicting external module dependencies
- ✅ Fixed Go module version conflicts
- ✅ Made infrastructure components framework-agnostic
- ✅ Resolved compilation errors

### 2. **Domain Layer Completion**
- ✅ Added missing `TenantIsolationLevel` enum to shared domain
- ✅ Fixed domain event implementations
- ✅ Completed tenant aggregate and repository interfaces
- ✅ Added proper value object definitions

### 3. **Middleware Layer**
- ✅ Created framework-agnostic HTTP middleware (`tenant_context_http.go`)
- ✅ Removed Gin framework dependency for better flexibility
- ✅ Implemented proper tenant context extraction and validation
- ✅ Added feature and subscription-based access control

### 4. **Persistence Layer**
- ✅ Fixed type assertion issues in tenant-aware repository
- ✅ Improved multi-tenant database abstraction
- ✅ Added proper error handling and transaction support
- ✅ Implemented tenant isolation strategies

## 🏗️ Architecture Overview

```
infrastructure/
├── middleware/
│   ├── tenant_context.go          # Original Gin-based middleware (deprecated)
│   └── tenant_context_http.go     # New framework-agnostic middleware
├── persistence/
│   └── tenant_aware_repository.go # Multi-tenant database abstraction
├── config.go                      # Infrastructure configuration
├── example_usage.go               # Usage examples
└── README.md                      # This file
```

## 🚀 Key Features

### **Multi-Tenant Support**
- **Shared Database**: All tenants share tables with `tenant_id` column
- **Schema Per Tenant**: Each tenant has its own database schema
- **Database Per Tenant**: Each tenant has its own dedicated database

### **Middleware Components**
- **Tenant Context Extraction**: From subdomain, headers, JWT, query params, or URL path
- **Feature-Based Access Control**: Restrict access based on subscription features
- **Subscription-Based Access Control**: Restrict access based on subscription plans
- **Tenant Isolation**: Ensure data isolation between tenants

### **Repository Pattern**
- **Tenant-Aware Queries**: Automatically add tenant filters to SQL queries
- **Transaction Support**: Tenant-aware database transactions
- **Query Builders**: Helper methods for building tenant-aware SQL queries

## 📖 Usage Examples

### Basic HTTP Server Setup

```go
package main

import (
    "net/http"
    "github.com/DimaJoyti/go-coffee/infrastructure"
)

func main() {
    // Initialize infrastructure
    config := infrastructure.DefaultInfrastructureConfig()
    container, err := infrastructure.NewInfrastructureContainer(config)
    if err != nil {
        panic(err)
    }
    defer container.Close()

    // Create HTTP server with middleware
    mux := http.NewServeMux()
    
    // Add tenant context middleware
    var handler http.Handler = mux
    if container.GetTenantContextMiddleware() != nil {
        handler = container.GetTenantContextMiddleware().HTTPTenantContext(handler)
    }
    
    // Add your routes
    mux.HandleFunc("/api/orders", handleOrders)
    
    // Start server
    server := &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }
    server.ListenAndServe()
}
```

### Tenant-Aware Repository

```go
// Create a tenant-aware repository
type OrderRepository struct {
    *persistence.BaseTenantAwareRepository
}

func NewOrderRepository(db persistence.TenantAwareDB) *OrderRepository {
    base := persistence.NewBaseTenantAwareRepository(db, "orders", "Order")
    return &OrderRepository{BaseTenantAwareRepository: base}
}

func (r *OrderRepository) FindOrdersByTenant(ctx context.Context, tenantID shared.TenantID) ([]Order, error) {
    query, args := r.BuildSelectQuery(tenantID, []string{"*"}, "", nil)
    rows, err := r.ExecuteQuery(ctx, tenantID, query, args)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // Parse rows into Order structs
    return orders, nil
}
```

### Feature-Based Access Control

```go
// Require premium features
premiumHandler := container.GetTenantContextMiddleware().RequireFeatureHTTP("premium_features")
mux.Handle("/api/premium/", premiumHandler(http.HandlerFunc(handlePremiumEndpoint)))

// Require enterprise subscription
enterpriseHandler := container.GetTenantContextMiddleware().RequireSubscriptionHTTP(shared.SubscriptionEnterprise)
mux.Handle("/api/enterprise/", enterpriseHandler(http.HandlerFunc(handleEnterpriseEndpoint)))
```

## ⚙️ Configuration

### Database Configuration

```go
config := infrastructure.DefaultInfrastructureConfig()

// For schema-per-tenant isolation
config.Database.IsolationLevel = shared.SchemaPerTenant
config.Tenant.SchemaPrefix = "tenant_"

// For database-per-tenant isolation
config.Database.IsolationLevel = shared.DatabasePerTenant
config.Database.TenantConnections = map[string]string{
    "tenant-1": "postgres://localhost/tenant_1?sslmode=disable",
    "tenant-2": "postgres://localhost/tenant_2?sslmode=disable",
}
```

### Security Configuration

```go
config.Security.JWTSecret = "your-secret-key"
config.Security.JWTExpiration = 24 * time.Hour
config.Security.EnableRateLimit = true
config.Security.RateLimitRequests = 100
config.Security.RateLimitWindow = time.Minute
```

## 🔍 Tenant Context Extraction

The middleware supports multiple methods for extracting tenant information:

1. **Subdomain**: `tenant1.api.example.com`
2. **Custom Header**: `X-Tenant-ID: tenant-123`
3. **JWT Token**: Bearer token with tenant claim
4. **Query Parameter**: `?tenant_id=tenant-123`
5. **URL Path**: `/api/v1/tenants/tenant-123/orders`

## 🏥 Health Checks and Metrics

```go
// Health check endpoint
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := container.HealthCheck(); err != nil {
        http.Error(w, "Health check failed", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "healthy"}`))
})

// Metrics endpoint
mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metrics := container.Metrics()
    // Serialize and return metrics
})
```

## 🧪 Testing

The infrastructure components are designed to be easily testable:

```go
func TestTenantContextMiddleware(t *testing.T) {
    // Create test configuration
    config := infrastructure.DefaultInfrastructureConfig()
    container, err := infrastructure.NewInfrastructureContainer(config)
    require.NoError(t, err)
    defer container.Close()
    
    // Test middleware functionality
    // ...
}
```

## 🔄 Migration from Old Infrastructure

If you're migrating from the old Gin-based infrastructure:

1. Replace `gin.HandlerFunc` with `http.HandlerFunc`
2. Use `tenant_context_http.go` instead of the old middleware
3. Update your route handlers to use standard HTTP patterns
4. Configure the infrastructure container instead of individual components

## 📝 Next Steps

1. **Implement Concrete Repository**: Create specific repository implementations for your entities
2. **Add Caching Layer**: Implement Redis-based caching for tenant data
3. **Enhance Security**: Add JWT token validation and more sophisticated authentication
4. **Add Monitoring**: Integrate with Prometheus/Grafana for metrics collection
5. **Performance Optimization**: Add connection pooling and query optimization

## 🤝 Contributing

When contributing to the infrastructure layer:

1. Maintain framework-agnostic design
2. Add comprehensive tests for new components
3. Update documentation for any API changes
4. Follow the established patterns for tenant awareness
5. Ensure backward compatibility where possible
