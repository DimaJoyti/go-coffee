package infrastructure

import (
	"net/http"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/infrastructure/persistence"
)

// ExampleHTTPServer demonstrates how to use the infrastructure components
func ExampleHTTPServer() {
	// Initialize infrastructure
	config := DefaultInfrastructureConfig()
	container, err := NewInfrastructureContainer(config)
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

	// Add tenant isolation middleware
	// Wrap the existing handler with tenant isolation middleware
	originalHandler := handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant context
		tenantCtx, err := GetTenantContext(r)
		if err != nil {
			http.Error(w, "Tenant context required", http.StatusUnauthorized)
			return
		}
		
		// Extract tenant ID from URL if present (simplified implementation)
		// In a real app, you would extract the tenant ID from the URL path and verify permissions
		_ = tenantCtx // Use the tenant context or implement actual tenant validation here
		
		// If all checks pass, proceed to the next handler
		originalHandler.ServeHTTP(w, r)
	})

	// Add feature requirement middleware for specific routes
	if container.GetTenantContextMiddleware() != nil {
		// Convert Gin middleware to HTTP middleware for premium features
		premiumHandler := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract tenant context
				tenantCtx, err := GetTenantContext(r)
				if err != nil {
					http.Error(w, "Tenant context required", http.StatusUnauthorized)
					return
				}
				
				// Check if tenant has premium features
				if !tenantCtx.HasFeature("premium_features") {
					http.Error(w, "Premium features not available", http.StatusForbidden)
					return
				}
				
				next.ServeHTTP(w, r)
			})
		}
		mux.Handle("/api/premium/", premiumHandler(http.HandlerFunc(handlePremiumEndpoint)))

		// Convert Gin middleware to HTTP middleware for enterprise subscription
		enterpriseHandler := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract tenant context
				tenantCtx, err := GetTenantContext(r)
				if err != nil {
					http.Error(w, "Tenant context required", http.StatusUnauthorized)
					return
				}
				
				// Check if tenant has enterprise subscription
				if tenantCtx.Subscription() != shared.SubscriptionEnterprise {
					http.Error(w, "Enterprise subscription required", http.StatusForbidden)
					return
				}
				
				next.ServeHTTP(w, r)
			})
		}
		mux.Handle("/api/enterprise/", enterpriseHandler(http.HandlerFunc(handleEnterpriseEndpoint)))

	}

	// Add regular endpoints
	mux.HandleFunc("/api/health", handleHealth(container))
	mux.HandleFunc("/api/metrics", handleMetrics(container))
	mux.HandleFunc("/api/orders", handleOrders)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// In a real application, you would handle graceful shutdown
	server.ListenAndServe()
}

// handleHealth returns health status
func handleHealth(container *InfrastructureContainer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := container.HealthCheck(); err != nil {
			http.Error(w, "Health check failed: "+err.Error(), http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}
}

// handleMetrics returns infrastructure metrics
func handleMetrics(container *InfrastructureContainer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := container.Metrics()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// In a real implementation, you would properly serialize the metrics
		w.Write([]byte(`{"metrics": "available"}`))
		_ = metrics // Use metrics here
	}
}

// handleOrders handles order-related requests
func handleOrders(w http.ResponseWriter, r *http.Request) {
	// Extract tenant context from request
	tenantCtx, err := GetTenantContext(r)
	if err != nil {
		http.Error(w, "Tenant context required", http.StatusUnauthorized)
		return
	}

	// Use tenant context for business logic
	tenantID := tenantCtx.TenantID()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Orders for tenant ` + tenantID.Value() + `"}`))
}

// handlePremiumEndpoint handles premium feature requests
func handlePremiumEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Premium feature accessed"}`))
}

// handleEnterpriseEndpoint handles enterprise feature requests
func handleEnterpriseEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Enterprise feature accessed"}`))
}

// GetTenantContext is a helper function to extract tenant context from request
func GetTenantContext(r *http.Request) (*shared.TenantContext, error) {
	return shared.FromContext(r.Context())
}

// ExampleTenantAwareRepository demonstrates how to use the tenant-aware repository
func ExampleTenantAwareRepository(container *InfrastructureContainer) {
	// Get tenant-aware database
	tenantDB := container.GetTenantDB()
	
	// Create a repository for a specific entity
	orderRepo := NewExampleOrderRepository(tenantDB)
	
	// Use the repository with tenant context
	// tenantID := shared.NewTenantID("tenant-123")
	// orders, err := orderRepo.FindOrdersByTenant(context.Background(), tenantID)
	
	_ = orderRepo // Use repository here
}

// ExampleOrderRepository demonstrates a tenant-aware repository implementation
type ExampleOrderRepository struct {
	*persistence.BaseTenantAwareRepository
}

// NewExampleOrderRepository creates a new example order repository
func NewExampleOrderRepository(db persistence.TenantAwareDB) *ExampleOrderRepository {
	base := persistence.NewBaseTenantAwareRepository(db, "orders", "Order")
	return &ExampleOrderRepository{
		BaseTenantAwareRepository: base,
	}
}

// Example methods would be implemented here
// func (r *ExampleOrderRepository) FindOrdersByTenant(ctx context.Context, tenantID shared.TenantID) ([]Order, error) {
//     query, args := r.BuildSelectQuery(tenantID, []string{"*"}, "", nil)
//     rows, err := r.ExecuteQuery(ctx, tenantID, query, args)
//     if err != nil {
//         return nil, err
//     }
//     defer rows.Close()
//     
//     // Parse rows into Order structs
//     return orders, nil
// }

// ExampleEventHandling demonstrates how to use the event system
func ExampleEventHandling(container *InfrastructureContainer) {
	eventPublisher := container.GetEventPublisher()
	
	// Create and subscribe event handlers
	// tenantHandler := tenant.NewTenantEventHandler()
	// eventPublisher.Subscribe(tenantHandler, 
	//     tenant.TenantCreatedEventType,
	//     tenant.TenantActivatedEventType,
	//     tenant.TenantSuspendedEventType,
	// )
	
	_ = eventPublisher // Use event publisher here
}

// ExampleConfiguration demonstrates different configuration options
func ExampleConfiguration() {
	// Default configuration
	defaultConfig := DefaultInfrastructureConfig()
	
	// Custom configuration for schema-per-tenant isolation
	schemaConfig := DefaultInfrastructureConfig()
	schemaConfig.Database.IsolationLevel = shared.SchemaPerTenant
	schemaConfig.Tenant.DefaultIsolationLevel = shared.SchemaPerTenant
	schemaConfig.Tenant.SchemaPrefix = "tenant_"
	
	// Custom configuration for database-per-tenant isolation
	dbConfig := DefaultInfrastructureConfig()
	dbConfig.Database.IsolationLevel = shared.DatabasePerTenant
	dbConfig.Tenant.DefaultIsolationLevel = shared.DatabasePerTenant
	dbConfig.Database.TenantConnections = map[string]string{
		"tenant-1": "postgres://localhost/tenant_1?sslmode=disable",
		"tenant-2": "postgres://localhost/tenant_2?sslmode=disable",
	}
	
	_ = defaultConfig
	_ = schemaConfig
	_ = dbConfig
}
