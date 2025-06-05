package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/application"
	"github.com/DimaJoyti/go-coffee/internal/security-gateway/infrastructure"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/security-gateway/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	pkgLogger "github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/security/encryption"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		Host         string        `mapstructure:"host"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`

	Security struct {
		Encryption encryption.Config `mapstructure:"encryption"`
		Validation validation.Config `mapstructure:"validation"`
		Monitoring monitoring.Config `mapstructure:"monitoring"`

		ThreatDetector struct {
			SuspiciousIPThreshold int           `mapstructure:"suspicious_ip_threshold"`
			BlockDuration         time.Duration `mapstructure:"block_duration"`
			AllowedUserAgents    []string      `mapstructure:"allowed_user_agents"`
			BlockedIPRanges      []string      `mapstructure:"blocked_ip_ranges"`
		} `mapstructure:"threat_detector"`

		RateLimit struct {
			Enabled           bool          `mapstructure:"enabled"`
			RequestsPerMinute int           `mapstructure:"requests_per_minute"`
			BurstSize         int           `mapstructure:"burst_size"`
			CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
		} `mapstructure:"rate_limit"`

		WAF struct {
			Enabled           bool     `mapstructure:"enabled"`
			BlockSuspiciousIP bool     `mapstructure:"block_suspicious_ip"`
			AllowedCountries  []string `mapstructure:"allowed_countries"`
			BlockedCountries  []string `mapstructure:"blocked_countries"`
			MaxRequestSize    int64    `mapstructure:"max_request_size"`
		} `mapstructure:"waf"`

		CORS struct {
			AllowedOrigins   []string `mapstructure:"allowed_origins"`
			AllowedMethods   []string `mapstructure:"allowed_methods"`
			AllowedHeaders   []string `mapstructure:"allowed_headers"`
			ExposedHeaders   []string `mapstructure:"exposed_headers"`
			AllowCredentials bool     `mapstructure:"allow_credentials"`
			MaxAge           int      `mapstructure:"max_age"`
		} `mapstructure:"cors"`
	} `mapstructure:"security"`

	Redis struct {
		URL      string `mapstructure:"url"`
		DB       int    `mapstructure:"db"`
		Password string `mapstructure:"password"`
	} `mapstructure:"redis"`

	Services struct {
		AuthService    string `mapstructure:"auth_service"`
		OrderService   string `mapstructure:"order_service"`
		PaymentService string `mapstructure:"payment_service"`
		UserService    string `mapstructure:"user_service"`
	} `mapstructure:"services"`

	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"logging"`

	Environment string `mapstructure:"environment"`
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := pkgLogger.New("security-gateway")
	defer logger.Sync()

	logger.Info("Starting Security Gateway Service - Environment: %s", config.Environment)

	// Initialize services
	services, err := initializeServices(config, logger)
	if err != nil {
		logger.Fatal("Failed to initialize services: %v", err)
	}

	// Initialize HTTP server
	router := setupRouter(config, services, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on address: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./cmd/security-gateway/config")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("environment", "development")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Enable environment variable support
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

type Services struct {
	EncryptionService *encryption.EncryptionService
	ValidationService *validation.ValidationService
	MonitoringService *monitoring.SecurityMonitoringService
	GatewayService    *application.SecurityGatewayService
	WAFService        *application.WAFService
	RateLimitService  *application.RateLimitService
}

func initializeServices(config *Config, logger *logger.Logger) (*Services, error) {
	// Initialize encryption service
	encryptionService, err := encryption.NewEncryptionService(&config.Security.Encryption)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	// Initialize validation service
	validationService := validation.NewValidationService(&config.Security.Validation, logger)

	// Initialize Redis services
	redisServices, err := infrastructure.NewRedisServices(&infrastructure.RedisConfig{
		URL:      config.Redis.URL,
		DB:       config.Redis.DB,
		Password: config.Redis.Password,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis services: %w", err)
	}

	// Create adapters for monitoring service
	eventStoreAdapter := infrastructure.NewEventStoreAdapter(redisServices.EventStore)
	alertManagerAdapter := infrastructure.NewAlertManagerAdapter(redisServices.AlertMgr)
	threatDetectorAdapter := infrastructure.NewThreatDetectorAdapter()

	// Initialize monitoring service
	monitoringService := monitoring.NewSecurityMonitoringService(
		&config.Security.Monitoring,
		logger,
		eventStoreAdapter,
		alertManagerAdapter,
		threatDetectorAdapter,
	)

	// Initialize rate limit service
	rateLimitService := application.NewRateLimitService(
		&application.RateLimitConfig{
			Enabled:           config.Security.RateLimit.Enabled,
			RequestsPerMinute: config.Security.RateLimit.RequestsPerMinute,
			BurstSize:         config.Security.RateLimit.BurstSize,
			CleanupInterval:   config.Security.RateLimit.CleanupInterval,
		},
		redisServices.Client,
		logger,
	)

	// Initialize WAF service
	wafService := application.NewWAFService(
		&application.WAFConfig{
			Enabled:           config.Security.WAF.Enabled,
			BlockSuspiciousIP: config.Security.WAF.BlockSuspiciousIP,
			AllowedCountries:  config.Security.WAF.AllowedCountries,
			BlockedCountries:  config.Security.WAF.BlockedCountries,
			MaxRequestSize:    config.Security.WAF.MaxRequestSize,
		},
		validationService,
		monitoringService,
		logger,
	)

	// Initialize gateway service
	gatewayService := application.NewSecurityGatewayService(
		&application.GatewayConfig{
			Services: map[string]string{
				"auth":    config.Services.AuthService,
				"order":   config.Services.OrderService,
				"payment": config.Services.PaymentService,
				"user":    config.Services.UserService,
			},
		},
		encryptionService,
		validationService,
		monitoringService,
		rateLimitService,
		wafService,
		logger,
	)

	return &Services{
		EncryptionService: encryptionService,
		ValidationService: validationService,
		MonitoringService: monitoringService,
		GatewayService:    gatewayService,
		WAFService:        wafService,
		RateLimitService:  rateLimitService,
	}, nil
}

func setupRouter(config *Config, services *Services, logger *logger.Logger) *gin.Engine {
	// Set Gin mode
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())

	// Add custom middleware
	router.Use(httpTransport.LoggingMiddleware(logger))
	router.Use(httpTransport.CORSMiddleware(&httpTransport.CORSConfig{
		AllowedOrigins:   config.Security.CORS.AllowedOrigins,
		AllowedMethods:   config.Security.CORS.AllowedMethods,
		AllowedHeaders:   config.Security.CORS.AllowedHeaders,
		ExposedHeaders:   config.Security.CORS.ExposedHeaders,
		AllowCredentials: config.Security.CORS.AllowCredentials,
		MaxAge:           config.Security.CORS.MaxAge,
	}))
	router.Use(httpTransport.SecurityHeadersMiddleware())

	if config.Security.RateLimit.Enabled {
		router.Use(httpTransport.RateLimitMiddleware(services.RateLimitService))
	}

	if config.Security.WAF.Enabled {
		router.Use(httpTransport.WAFMiddleware(services.WAFService))
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "security-gateway",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", httpTransport.MetricsHandler(services.MonitoringService))

	// API routes
	api := router.Group("/api/v1")
	{
		// Security endpoints
		security := api.Group("/security")
		{
			security.POST("/validate", httpTransport.ValidateHandler(services.ValidationService))
			security.GET("/metrics", httpTransport.SecurityMetricsHandler(services.MonitoringService))
			security.GET("/alerts", httpTransport.AlertsHandler(services.MonitoringService))
		}

		// Gateway endpoints (proxy to other services)
		gateway := api.Group("/gateway")
		gateway.Use(httpTransport.GatewayMiddleware(services.GatewayService))
		{
			gateway.Any("/auth/*path", httpTransport.ProxyHandler("auth", services.GatewayService))
			gateway.Any("/order/*path", httpTransport.ProxyHandler("order", services.GatewayService))
			gateway.Any("/payment/*path", httpTransport.ProxyHandler("payment", services.GatewayService))
			gateway.Any("/user/*path", httpTransport.ProxyHandler("user", services.GatewayService))
		}
	}

	return router
}
