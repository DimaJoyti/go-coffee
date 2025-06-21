package container

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/config"
	"github.com/DimaJoyti/go-coffee/internal/object-detection/monitoring"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Container holds all application dependencies
type Container struct {
	Config   *config.Config
	Logger   *zap.Logger
	DB       *sql.DB
	Redis    *redis.Client
	Metrics  *monitoring.Metrics

	// Services will be added here as we implement them
	// StreamService    domain.StreamService
	// DetectionService domain.DetectionService
	// TrackingService  domain.TrackingService
	// AlertService     domain.AlertService
	// ModelService     domain.ModelService

	// Repositories will be added here as we implement them
	// StreamRepo    domain.StreamRepository
	// DetectionRepo domain.DetectionRepository
	// TrackingRepo  domain.TrackingRepository
	// AlertRepo     domain.AlertRepository
	// ModelRepo     domain.ModelRepository
	// CacheRepo     domain.CacheRepository
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, logger *zap.Logger) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize metrics
	container.Metrics = monitoring.NewMetrics()

	// Initialize database connection
	if err := container.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis connection
	if err := container.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// TODO: Initialize repositories
	// if err := container.initRepositories(); err != nil {
	//     return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	// }

	// TODO: Initialize services
	// if err := container.initServices(); err != nil {
	//     return nil, fmt.Errorf("failed to initialize services: %w", err)
	// }

	logger.Info("Container initialized successfully")
	return container, nil
}

// initDatabase initializes the database connection
func (c *Container) initDatabase() error {
	dsn := c.Config.GetDatabaseDSN()
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.Config.Database.MaxOpenConns)
	db.SetMaxIdleConns(c.Config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(c.Config.Database.ConnMaxLifetime) * time.Second)

	// Test the connection
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.DB = db
	c.Logger.Info("Database connection established", 
		zap.String("host", c.Config.Database.Host),
		zap.Int("port", c.Config.Database.Port),
		zap.String("database", c.Config.Database.Database))

	return nil
}

// initRedis initializes the Redis connection
func (c *Container) initRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Config.GetRedisAddr(),
		Password:     c.Config.Redis.Password,
		DB:           c.Config.Redis.Database,
		PoolSize:     c.Config.Redis.PoolSize,
		MinIdleConns: c.Config.Redis.MinIdleConns,
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	c.Redis = rdb
	c.Logger.Info("Redis connection established", 
		zap.String("addr", c.Config.GetRedisAddr()),
		zap.Int("database", c.Config.Redis.Database))

	return nil
}

// TODO: Implement these methods as we add repositories and services

// initRepositories initializes all repositories
// func (c *Container) initRepositories() error {
//     // Initialize PostgreSQL repositories
//     c.StreamRepo = postgres.NewStreamRepository(c.DB, c.Logger)
//     c.DetectionRepo = postgres.NewDetectionRepository(c.DB, c.Logger)
//     c.TrackingRepo = postgres.NewTrackingRepository(c.DB, c.Logger)
//     c.AlertRepo = postgres.NewAlertRepository(c.DB, c.Logger)
//     c.ModelRepo = postgres.NewModelRepository(c.DB, c.Logger)
//     
//     // Initialize Redis cache repository
//     c.CacheRepo = redis.NewCacheRepository(c.Redis, c.Logger)
//     
//     return nil
// }

// initServices initializes all services
// func (c *Container) initServices() error {
//     // Initialize services with their dependencies
//     c.StreamService = application.NewStreamService(
//         c.StreamRepo,
//         c.CacheRepo,
//         c.Logger,
//     )
//     
//     c.DetectionService = application.NewDetectionService(
//         c.DetectionRepo,
//         c.StreamRepo,
//         c.CacheRepo,
//         c.Logger,
//         c.Config,
//     )
//     
//     c.TrackingService = application.NewTrackingService(
//         c.TrackingRepo,
//         c.CacheRepo,
//         c.Logger,
//         c.Config,
//     )
//     
//     c.AlertService = application.NewAlertService(
//         c.AlertRepo,
//         c.CacheRepo,
//         c.Logger,
//     )
//     
//     c.ModelService = application.NewModelService(
//         c.ModelRepo,
//         c.CacheRepo,
//         c.Logger,
//         c.Config,
//     )
//     
//     return nil
// }

// Close closes all connections and cleans up resources
func (c *Container) Close() error {
	var errors []error

	// Close database connection
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database: %w", err))
		}
	}

	// Close Redis connection
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errors)
	}

	c.Logger.Info("Container closed successfully")
	return nil
}

// HealthCheck performs health checks on all dependencies
func (c *Container) HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)

	// Check database
	if c.DB != nil {
		if err := c.DB.PingContext(ctx); err != nil {
			status["database"] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			status["database"] = "healthy"
		}
	} else {
		status["database"] = "not initialized"
	}

	// Check Redis
	if c.Redis != nil {
		if err := c.Redis.Ping(ctx).Err(); err != nil {
			status["redis"] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			status["redis"] = "healthy"
		}
	} else {
		status["redis"] = "not initialized"
	}

	return status
}
