package infrastructure

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
	"github.com/DimaJoyti/go-coffee/infrastructure/middleware"
	"github.com/DimaJoyti/go-coffee/infrastructure/persistence"
)

// InfrastructureConfig holds configuration for infrastructure components
type InfrastructureConfig struct {
	Database DatabaseConfig `json:"database" yaml:"database"`
	Tenant   TenantConfig   `json:"tenant" yaml:"tenant"`
	Security SecurityConfig `json:"security" yaml:"security"`
	Logging  LoggingConfig  `json:"logging" yaml:"logging"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver               string                        `json:"driver" yaml:"driver"`
	ConnectionString     string                        `json:"connection_string" yaml:"connection_string"`
	MaxOpenConnections   int                           `json:"max_open_connections" yaml:"max_open_connections"`
	MaxIdleConnections   int                           `json:"max_idle_connections" yaml:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration                `json:"connection_max_lifetime" yaml:"connection_max_lifetime"`
	IsolationLevel       shared.TenantIsolationLevel   `json:"isolation_level" yaml:"isolation_level"`
	TenantConnections    map[string]string             `json:"tenant_connections" yaml:"tenant_connections"`
}

// TenantConfig holds tenant-specific configuration
type TenantConfig struct {
	DefaultIsolationLevel shared.TenantIsolationLevel `json:"default_isolation_level" yaml:"default_isolation_level"`
	SchemaPrefix          string                      `json:"schema_prefix" yaml:"schema_prefix"`
	EnableMetrics         bool                        `json:"enable_metrics" yaml:"enable_metrics"`
	CacheEnabled          bool                        `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL              time.Duration               `json:"cache_ttl" yaml:"cache_ttl"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret           string        `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration       time.Duration `json:"jwt_expiration" yaml:"jwt_expiration"`
	EnableRateLimit     bool          `json:"enable_rate_limit" yaml:"enable_rate_limit"`
	RateLimitRequests   int           `json:"rate_limit_requests" yaml:"rate_limit_requests"`
	RateLimitWindow     time.Duration `json:"rate_limit_window" yaml:"rate_limit_window"`
	EnableCORS          bool          `json:"enable_cors" yaml:"enable_cors"`
	AllowedOrigins      []string      `json:"allowed_origins" yaml:"allowed_origins"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	EnableJSON bool   `json:"enable_json" yaml:"enable_json"`
}

// DefaultInfrastructureConfig returns default infrastructure configuration
func DefaultInfrastructureConfig() *InfrastructureConfig {
	return &InfrastructureConfig{
		Database: DatabaseConfig{
			Driver:                "postgres",
			ConnectionString:      "postgres://localhost/go_coffee?sslmode=disable",
			MaxOpenConnections:    25,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: 5 * time.Minute,
			IsolationLevel:        shared.SharedDatabase,
			TenantConnections:     make(map[string]string),
		},
		Tenant: TenantConfig{
			DefaultIsolationLevel: shared.SharedDatabase,
			SchemaPrefix:          "tenant_",
			EnableMetrics:         true,
			CacheEnabled:          true,
			CacheTTL:              10 * time.Minute,
		},
		Security: SecurityConfig{
			JWTSecret:         "your-secret-key",
			JWTExpiration:     24 * time.Hour,
			EnableRateLimit:   true,
			RateLimitRequests: 100,
			RateLimitWindow:   time.Minute,
			EnableCORS:        true,
			AllowedOrigins:    []string{"*"},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			EnableJSON: false,
		},
	}
}

// InfrastructureContainer holds all infrastructure components
type InfrastructureContainer struct {
	config                    *InfrastructureConfig
	database                  *sql.DB
	tenantDB                  persistence.TenantAwareDB
	tenantRepository          tenant.TenantRepository
	tenantContextMiddleware   *middleware.TenantContextMiddleware
	tenantIsolationMiddleware *middleware.TenantIsolationMiddleware
	eventPublisher            shared.DomainEventPublisher
}

// NewInfrastructureContainer creates a new infrastructure container
func NewInfrastructureContainer(config *InfrastructureConfig) (*InfrastructureContainer, error) {
	container := &InfrastructureContainer{
		config: config,
	}

	if err := container.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := container.initializeTenantComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize tenant components: %w", err)
	}

	if err := container.initializeEventSystem(); err != nil {
		return nil, fmt.Errorf("failed to initialize event system: %w", err)
	}

	return container, nil
}

// initializeDatabase initializes the database connection
func (c *InfrastructureContainer) initializeDatabase() error {
	db, err := sql.Open(c.config.Database.Driver, c.config.Database.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.config.Database.MaxOpenConnections)
	db.SetMaxIdleConns(c.config.Database.MaxIdleConnections)
	db.SetConnMaxLifetime(c.config.Database.ConnectionMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.database = db

	// Initialize tenant-aware database
	c.tenantDB = persistence.NewMultiTenantDB(c.config.Database.IsolationLevel, db)

	return nil
}

// initializeTenantComponents initializes tenant-related components
func (c *InfrastructureContainer) initializeTenantComponents() error {
	// Initialize tenant repository (this would be implemented based on your specific needs)
	// For now, we'll use a placeholder
	// c.tenantRepository = persistence.NewTenantRepository(c.tenantDB)

	// Initialize middleware
	if c.tenantRepository != nil {
		c.tenantContextMiddleware = middleware.NewTenantContextMiddleware(c.tenantRepository)
	}
	c.tenantIsolationMiddleware = middleware.NewTenantIsolationMiddleware()

	return nil
}

// initializeEventSystem initializes the event system
func (c *InfrastructureContainer) initializeEventSystem() error {
	c.eventPublisher = shared.NewInMemoryDomainEventPublisher()
	return nil
}

// GetDatabase returns the database connection
func (c *InfrastructureContainer) GetDatabase() *sql.DB {
	return c.database
}

// GetTenantDB returns the tenant-aware database
func (c *InfrastructureContainer) GetTenantDB() persistence.TenantAwareDB {
	return c.tenantDB
}

// GetTenantRepository returns the tenant repository
func (c *InfrastructureContainer) GetTenantRepository() tenant.TenantRepository {
	return c.tenantRepository
}

// GetTenantContextMiddleware returns the tenant context middleware
func (c *InfrastructureContainer) GetTenantContextMiddleware() *middleware.TenantContextMiddleware {
	return c.tenantContextMiddleware
}

// GetTenantIsolationMiddleware returns the tenant isolation middleware
func (c *InfrastructureContainer) GetTenantIsolationMiddleware() *middleware.TenantIsolationMiddleware {
	return c.tenantIsolationMiddleware
}

// GetEventPublisher returns the event publisher
func (c *InfrastructureContainer) GetEventPublisher() shared.DomainEventPublisher {
	return c.eventPublisher
}

// GetConfig returns the infrastructure configuration
func (c *InfrastructureContainer) GetConfig() *InfrastructureConfig {
	return c.config
}

// Close closes all infrastructure resources
func (c *InfrastructureContainer) Close() error {
	if c.database != nil {
		return c.database.Close()
	}
	return nil
}

// HealthCheck performs a health check on infrastructure components
func (c *InfrastructureContainer) HealthCheck() error {
	// Check database connection
	if c.database != nil {
		if err := c.database.Ping(); err != nil {
			return fmt.Errorf("database health check failed: %w", err)
		}
	}

	// Add more health checks as needed
	return nil
}

// Metrics returns infrastructure metrics
func (c *InfrastructureContainer) Metrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	if c.database != nil {
		stats := c.database.Stats()
		metrics["database"] = map[string]interface{}{
			"open_connections":     stats.OpenConnections,
			"in_use":              stats.InUse,
			"idle":                stats.Idle,
			"wait_count":          stats.WaitCount,
			"wait_duration":       stats.WaitDuration,
			"max_idle_closed":     stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		}
	}

	return metrics
}
