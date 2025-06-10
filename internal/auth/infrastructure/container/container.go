package container

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/jwt"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/password"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/repository/postgres"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/security"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/auth/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Container holds all dependencies for the auth service
type Container struct {
	// Infrastructure
	DB           *sql.DB
	CacheService application.CacheService
	Logger       *logger.Logger

	// Repositories
	UserRepository    domain.UserRepository
	SessionRepository domain.SessionRepository

	// Services
	PasswordService application.PasswordService
	JWTService      application.JWTService
	SecurityService application.SecurityService
	AuthService     application.AuthService
	MFAService      application.MFAService

	// Transport
	HTTPServer *httpTransport.Server

	// Configuration
	Config *Config
}

// Config represents the complete configuration for the auth service
type Config struct {
	Database *DatabaseConfig       `yaml:"database"`
	Redis    *cache.Config         `yaml:"redis"`
	JWT      *jwt.Config           `yaml:"jwt"`
	Password *password.Config      `yaml:"password"`
	Security *security.Config      `yaml:"security"`
	HTTP     *httpTransport.Config `yaml:"http"`
	Logger   *logger.Config        `yaml:"logger"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_conns"`
	MinConns int    `yaml:"min_conns"`
}

// NewContainer creates a new dependency injection container
func NewContainer(config *Config) (*Container, error) {
	container := &Container{
		Config: config,
	}

	// Initialize logger first
	if err := container.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database
	if err := container.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize cache
	if err := container.initCache(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize repositories
	if err := container.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize services
	if err := container.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize transport
	if err := container.initTransport(); err != nil {
		return nil, fmt.Errorf("failed to initialize transport: %w", err)
	}

	container.Logger.Info("Container initialized successfully")
	return container, nil
}

// initLogger initializes the logger
func (c *Container) initLogger() error {
	if c.Config.Logger == nil {
		c.Config.Logger = logger.DefaultConfig()
	}

	c.Logger = logger.NewLogger(c.Config.Logger)
	return nil
}

// initDatabase initializes the database connection
func (c *Container) initDatabase() error {
	if c.Config.Database == nil {
		c.Config.Database = DefaultDatabaseConfig()
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Config.Database.Host,
		c.Config.Database.Port,
		c.Config.Database.Username,
		c.Config.Database.Password,
		c.Config.Database.Database,
		c.Config.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.Config.Database.MaxConns)
	db.SetMaxIdleConns(c.Config.Database.MinConns)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.DB = db
	c.Logger.Info("Database connection established")
	return nil
}

// initCache initializes the cache service
func (c *Container) initCache() error {
	if c.Config.Redis == nil {
		c.Config.Redis = cache.DefaultConfig()
	}

	cacheService, err := cache.NewRedisCache(c.Config.Redis, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create cache service: %w", err)
	}

	c.CacheService = cacheService
	c.Logger.Info("Cache service initialized")
	return nil
}

// initRepositories initializes the repositories
func (c *Container) initRepositories() error {
	// User repository
	c.UserRepository = postgres.NewUserRepository(c.DB, c.Logger)

	// Session repository
	c.SessionRepository = postgres.NewSessionRepository(c.DB, c.Logger)

	c.Logger.Info("Repositories initialized")
	return nil
}

// initServices initializes the application services
func (c *Container) initServices() error {
	// Password service
	if c.Config.Password == nil {
		c.Config.Password = password.DefaultConfig()
	}
	c.PasswordService = password.NewBcryptService(c.Config.Password, c.Logger)

	// JWT service
	if c.Config.JWT == nil {
		c.Config.JWT = jwt.DefaultConfig()
	}
	if err := jwt.ValidateConfig(c.Config.JWT); err != nil {
		return fmt.Errorf("invalid JWT config: %w", err)
	}
	c.JWTService = jwt.NewJWTService(c.Config.JWT, c.Logger)

	// Security service
	if c.Config.Security == nil {
		c.Config.Security = security.DefaultConfig()
	}
	c.SecurityService = security.NewSecurityService(c.CacheService, c.Config.Security, c.Logger)

	// Auth service
	c.AuthService = application.NewAuthService(
		c.UserRepository,
		c.SessionRepository,
		c.PasswordService,
		c.JWTService,
		c.SecurityService,
		c.CacheService,
		c.Logger,
	)

	// MFA service
	c.MFAService = application.NewMFAService(
		c.UserRepository,
		c.SessionRepository,
		nil, // monitoring service - not implemented yet
		nil, // SMS provider - not implemented yet
		nil, // email provider - not implemented yet
		&application.MFAConfig{
			TOTPIssuer:              "Go Coffee",
			SMSCodeLength:           6,
			EmailCodeLength:         6,
			BackupCodesCount:        10,
			BackupCodeLength:        8,
			SMSCodeExpiry:           5 * time.Minute,
			EmailCodeExpiry:         5 * time.Minute,
			MaxVerificationAttempts: 3,
		},
		c.Logger,
	)

	c.Logger.Info("Services initialized")
	return nil
}

// initTransport initializes the transport layer
func (c *Container) initTransport() error {
	if c.Config.HTTP == nil {
		c.Config.HTTP = httpTransport.DefaultConfig()
	}

	c.HTTPServer = httpTransport.NewServer(
		c.Config.HTTP,
		c.AuthService,
		c.MFAService,
		c.Logger,
	)

	c.Logger.Info("Transport layer initialized")
	return nil
}

// Close closes all connections and cleans up resources
func (c *Container) Close() error {
	var errors []error

	// Close HTTP server
	if c.HTTPServer != nil {
		// HTTP server close is handled by the server itself
	}

	// Close cache service
	if c.CacheService != nil {
		if closer, ok := c.CacheService.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close cache service: %w", err))
			}
		}
	}

	// Close database
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errors)
	}

	c.Logger.Info("Container closed successfully")
	return nil
}

// Health checks the health of all components
func (c *Container) Health() map[string]interface{} {
	health := make(map[string]interface{})

	// Database health
	if c.DB != nil {
		if err := c.DB.Ping(); err != nil {
			health["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			health["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	// Cache health
	if c.CacheService != nil {
		if healthChecker, ok := c.CacheService.(interface{ Health() error }); ok {
			if err := healthChecker.Health(); err != nil {
				health["cache"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
			} else {
				health["cache"] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	return health
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DefaultDatabaseConfig(),
		Redis:    cache.DefaultConfig(),
		JWT:      jwt.DefaultConfig(),
		Password: password.DefaultConfig(),
		Security: security.DefaultConfig(),
		HTTP:     httpTransport.DefaultConfig(),
		Logger:   logger.DefaultConfig(),
	}
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "go_coffee_auth",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
		MaxConns: 25,
		MinConns: 5,
	}
}
