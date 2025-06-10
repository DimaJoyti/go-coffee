package infrastructure

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/cache"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/database"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/events"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/redis"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/security"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/session"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Container holds all infrastructure dependencies
type Container struct {
	// Configuration
	Config *config.InfrastructureConfig

	// Core infrastructure
	Logger   *logger.Logger
	Database database.DatabaseInterface
	Redis    redis.ClientInterface
	Cache    cache.Cache

	// Security services
	JWTService        security.JWTService
	EncryptionService security.EncryptionService

	// Event infrastructure
	EventStore      events.EventStore
	EventPublisher  events.EventPublisher
	EventSubscriber events.EventSubscriber

	// Managers
	DatabaseManager *database.DatabaseManager
	CacheManager    *cache.CacheManager
	SessionManager  *session.Manager

	// State
	initialized bool
}

// ContainerInterface defines the container interface
type ContainerInterface interface {
	// Initialization
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	IsInitialized() bool

	// Getters
	GetLogger() *logger.Logger
	GetDatabase() database.DatabaseInterface
	GetRedis() redis.ClientInterface
	GetCache() cache.Cache
	GetJWTService() security.JWTService
	GetEncryptionService() security.EncryptionService
	GetSessionManager() *session.Manager
	GetEventStore() events.EventStore
	GetEventPublisher() events.EventPublisher
	GetEventSubscriber() events.EventSubscriber

	// Health checks
	HealthCheck(ctx context.Context) (*HealthStatus, error)
}

// HealthStatus represents the health status of infrastructure components
type HealthStatus struct {
	Overall   string           `json:"overall"`
	Database  map[string]error `json:"database"`
	Redis     error            `json:"redis"`
	Cache     error            `json:"cache"`
	Events    map[string]error `json:"events"`
	Timestamp string           `json:"timestamp"`
}

// NewContainer creates a new infrastructure container
func NewContainer(cfg *config.InfrastructureConfig, logger *logger.Logger) ContainerInterface {
	return &Container{
		Config:          cfg,
		Logger:          logger,
		DatabaseManager: database.NewDatabaseManager(logger),
		CacheManager:    cache.NewCacheManager(logger),
	}
}

// Initialize initializes all infrastructure components
func (c *Container) Initialize(ctx context.Context) error {
	if c.initialized {
		return fmt.Errorf("container is already initialized")
	}

	c.Logger.Info("Initializing infrastructure container")

	// Validate configuration
	if err := c.Config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize Redis
	if err := c.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize Database
	if err := c.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize Database: %w", err)
	}

	// Initialize Cache
	if err := c.initCache(); err != nil {
		return fmt.Errorf("failed to initialize Cache: %w", err)
	}

	// Initialize Security services
	if err := c.initSecurity(); err != nil {
		return fmt.Errorf("failed to initialize Security: %w", err)
	}

	// Initialize Event infrastructure
	if err := c.initEvents(ctx); err != nil {
		return fmt.Errorf("failed to initialize Events: %w", err)
	}

	// Initialize Session management
	if err := c.initSessionManager(); err != nil {
		return fmt.Errorf("failed to initialize Session Manager: %w", err)
	}

	c.initialized = true
	c.Logger.Info("Infrastructure container initialized successfully")

	return nil
}

// Shutdown gracefully shuts down all infrastructure components
func (c *Container) Shutdown(ctx context.Context) error {
	if !c.initialized {
		return nil
	}

	c.Logger.Info("Shutting down infrastructure container")

	var lastErr error

	// Stop session manager
	if c.SessionManager != nil {
		c.SessionManager.Shutdown()
		c.Logger.Info("Session manager stopped")
	}

	// Stop event infrastructure
	if c.EventPublisher != nil {
		if err := c.EventPublisher.Stop(ctx); err != nil {
			c.Logger.WithError(err).Error("Failed to stop event publisher")
			lastErr = err
		}
	}

	if c.EventSubscriber != nil {
		if err := c.EventSubscriber.Stop(ctx); err != nil {
			c.Logger.WithError(err).Error("Failed to stop event subscriber")
			lastErr = err
		}
	}

	// Close cache connections
	if err := c.CacheManager.CloseAll(); err != nil {
		c.Logger.WithError(err).Error("Failed to close cache connections")
		lastErr = err
	}

	// Close database connections
	if err := c.DatabaseManager.CloseAll(); err != nil {
		c.Logger.WithError(err).Error("Failed to close database connections")
		lastErr = err
	}

	// Close Redis connection
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.Logger.WithError(err).Error("Failed to close Redis connection")
			lastErr = err
		}
	}

	c.initialized = false
	c.Logger.Info("Infrastructure container shut down")

	return lastErr
}

// IsInitialized returns whether the container is initialized
func (c *Container) IsInitialized() bool {
	return c.initialized
}

// initRedis initializes Redis client
func (c *Container) initRedis() error {
	if c.Config.Redis == nil {
		c.Logger.Info("Redis configuration not provided, skipping Redis initialization")
		return nil
	}

	redisClient, err := redis.NewClient(c.Config.Redis, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	c.Redis = redisClient
	c.Logger.Info("Redis client initialized")

	return nil
}

// initDatabase initializes database connection
func (c *Container) initDatabase() error {
	if c.Config.Database == nil {
		c.Logger.Info("Database configuration not provided, skipping database initialization")
		return nil
	}

	db, err := database.NewDatabase(c.Config.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	c.Database = db
	c.DatabaseManager.AddDatabase("default", db)
	c.Logger.Info("Database connection initialized")

	return nil
}

// initCache initializes cache service
func (c *Container) initCache() error {
	if c.Redis == nil {
		c.Logger.Info("Redis not available, skipping cache initialization")
		return nil
	}

	cacheService := cache.NewRedisCache(c.Redis, c.Config.Cache, c.Logger)
	c.Cache = cacheService
	c.CacheManager.AddCache("default", cacheService)
	c.Logger.Info("Cache service initialized")

	return nil
}

// initSecurity initializes security services
func (c *Container) initSecurity() error {
	if c.Config.Security == nil {
		c.Logger.Info("Security configuration not provided, skipping security initialization")
		return nil
	}

	// Initialize JWT service
	if c.Config.Security.JWT != nil {
		jwtService := security.NewJWTService(c.Config.Security.JWT, c.Logger)
		c.JWTService = jwtService
		c.Logger.Info("JWT service initialized")
	}

	// Initialize encryption service
	if c.Config.Security.Encryption != nil {
		encryptionService, err := security.NewEncryptionService(c.Config.Security.Encryption, c.Logger)
		if err != nil {
			return fmt.Errorf("failed to create encryption service: %w", err)
		}
		c.EncryptionService = encryptionService
		c.Logger.Info("Encryption service initialized")
	}

	return nil
}

// initEvents initializes event infrastructure
func (c *Container) initEvents(ctx context.Context) error {
	if c.Config.Events == nil {
		c.Logger.Info("Events configuration not provided, skipping events initialization")
		return nil
	}

	// Initialize event store
	if c.Redis != nil && c.Config.Events.Store != nil {
		eventStore := events.NewRedisEventStore(c.Redis, c.Config.Events.Store, c.Logger)
		c.EventStore = eventStore
		c.Logger.Info("Event store initialized")
	}

	// Initialize event publisher
	if c.Redis != nil && c.Config.Events.Publisher != nil {
		eventPublisher := events.NewRedisEventPublisher(c.Redis, c.Config.Events.Publisher, c.Logger)
		if err := eventPublisher.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event publisher: %w", err)
		}
		c.EventPublisher = eventPublisher
		c.Logger.Info("Event publisher initialized")
	}

	// Initialize event subscriber
	if c.Redis != nil && c.Config.Events.Subscriber != nil {
		eventSubscriber := events.NewRedisEventSubscriber(c.Redis, c.Config.Events.Subscriber, c.Logger)
		if err := eventSubscriber.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event subscriber: %w", err)
		}
		c.EventSubscriber = eventSubscriber
		c.Logger.Info("Event subscriber initialized")
	}

	return nil
}

// initSessionManager initializes session management
func (c *Container) initSessionManager() error {
	if c.Cache == nil {
		c.Logger.Info("Cache not available, skipping session manager initialization")
		return nil
	}

	// Create session manager with cache and event publisher
	sessionManager := session.NewManager(c.Cache, c.EventPublisher, c.Logger, nil)
	c.SessionManager = sessionManager
	c.Logger.Info("Session manager initialized")

	return nil
}

// Getter methods
func (c *Container) GetLogger() *logger.Logger {
	return c.Logger
}

func (c *Container) GetDatabase() database.DatabaseInterface {
	return c.Database
}

func (c *Container) GetRedis() redis.ClientInterface {
	return c.Redis
}

func (c *Container) GetCache() cache.Cache {
	return c.Cache
}

func (c *Container) GetJWTService() security.JWTService {
	return c.JWTService
}

func (c *Container) GetEncryptionService() security.EncryptionService {
	return c.EncryptionService
}

func (c *Container) GetEventStore() events.EventStore {
	return c.EventStore
}

func (c *Container) GetEventPublisher() events.EventPublisher {
	return c.EventPublisher
}

func (c *Container) GetEventSubscriber() events.EventSubscriber {
	return c.EventSubscriber
}

func (c *Container) GetSessionManager() *session.Manager {
	return c.SessionManager
}

// HealthCheck performs health checks on all infrastructure components
func (c *Container) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Overall:   "healthy",
		Database:  make(map[string]error),
		Events:    make(map[string]error),
		Timestamp: fmt.Sprintf("%d", ctx.Value("timestamp")),
	}

	// Check database health
	if c.DatabaseManager != nil {
		status.Database = c.DatabaseManager.HealthCheck(ctx)
		for _, err := range status.Database {
			if err != nil {
				status.Overall = "unhealthy"
			}
		}
	}

	// Check Redis health
	if c.Redis != nil {
		status.Redis = c.Redis.Ping(ctx)
		if status.Redis != nil {
			status.Overall = "unhealthy"
		}
	}

	// Check cache health
	if c.Cache != nil {
		status.Cache = c.Cache.Ping(ctx)
		if status.Cache != nil {
			status.Overall = "unhealthy"
		}
	}

	// Check event infrastructure health
	if c.EventStore != nil {
		status.Events["store"] = c.EventStore.HealthCheck(ctx)
		if status.Events["store"] != nil {
			status.Overall = "unhealthy"
		}
	}

	if c.EventPublisher != nil {
		status.Events["publisher"] = c.EventPublisher.HealthCheck(ctx)
		if status.Events["publisher"] != nil {
			status.Overall = "unhealthy"
		}
	}

	if c.EventSubscriber != nil {
		status.Events["subscriber"] = c.EventSubscriber.HealthCheck(ctx)
		if status.Events["subscriber"] != nil {
			status.Overall = "unhealthy"
		}
	}

	return status, nil
}
