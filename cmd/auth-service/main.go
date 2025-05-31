package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/repository"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/security"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

const (
	serviceName = "auth-service"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		HTTPPort         int           `mapstructure:"http_port"`
		GRPCPort         int           `mapstructure:"grpc_port"`
		Host             string        `mapstructure:"host"`
		ReadTimeout      time.Duration `mapstructure:"read_timeout"`
		WriteTimeout     time.Duration `mapstructure:"write_timeout"`
		IdleTimeout      time.Duration `mapstructure:"idle_timeout"`
		ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
	} `mapstructure:"server"`

	Redis struct {
		URL             string        `mapstructure:"url"`
		DB              int           `mapstructure:"db"`
		MaxRetries      int           `mapstructure:"max_retries"`
		PoolSize        int           `mapstructure:"pool_size"`
		MinIdleConns    int           `mapstructure:"min_idle_conns"`
		DialTimeout     time.Duration `mapstructure:"dial_timeout"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
		IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"redis"`

	Security struct {
		JWTSecret        string        `mapstructure:"jwt_secret"`
		AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
		RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
		BcryptCost       int           `mapstructure:"bcrypt_cost"`
		MaxLoginAttempts int           `mapstructure:"max_login_attempts"`
		LockoutDuration  time.Duration `mapstructure:"lockout_duration"`
		PasswordPolicy   struct {
			MinLength        int  `mapstructure:"min_length"`
			RequireUppercase bool `mapstructure:"require_uppercase"`
			RequireLowercase bool `mapstructure:"require_lowercase"`
			RequireNumbers   bool `mapstructure:"require_numbers"`
			RequireSymbols   bool `mapstructure:"require_symbols"`
		} `mapstructure:"password_policy"`
	} `mapstructure:"security"`

	RateLimiting struct {
		Enabled           bool          `mapstructure:"enabled"`
		RequestsPerMinute int           `mapstructure:"requests_per_minute"`
		BurstSize         int           `mapstructure:"burst_size"`
		CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
	} `mapstructure:"rate_limiting"`

	CORS struct {
		AllowedOrigins   []string `mapstructure:"allowed_origins"`
		AllowedMethods   []string `mapstructure:"allowed_methods"`
		AllowedHeaders   []string `mapstructure:"allowed_headers"`
		ExposeHeaders    []string `mapstructure:"expose_headers"`
		AllowCredentials bool     `mapstructure:"allow_credentials"`
		MaxAge           int      `mapstructure:"max_age"`
	} `mapstructure:"cors"`

	Logging struct {
		Level      string `mapstructure:"level"`
		Format     string `mapstructure:"format"`
		Output     string `mapstructure:"output"`
		FilePath   string `mapstructure:"file_path"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxAge     int    `mapstructure:"max_age"`
		MaxBackups int    `mapstructure:"max_backups"`
		Compress   bool   `mapstructure:"compress"`
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
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Auth Service...")

	// Initialize Redis client
	redisClient, err := initializeRedis(config, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := repository.NewRedisUserRepository(redisClient, logger)
	sessionRepo := repository.NewRedisSessionRepository(redisClient, logger)

	// Initialize security services
	jwtConfig := &security.JWTConfig{
		Secret:          config.Security.JWTSecret,
		AccessTokenTTL:  config.Security.AccessTokenTTL,
		RefreshTokenTTL: config.Security.RefreshTokenTTL,
		Issuer:          serviceName,
		Audience:        "auth-service-users",
	}
	jwtService := security.NewJWTService(jwtConfig, logger)

	passwordConfig := &security.PasswordConfig{
		BcryptCost: config.Security.BcryptCost,
		PasswordPolicy: &application.PasswordPolicy{
			MinLength:        config.Security.PasswordPolicy.MinLength,
			RequireUppercase: config.Security.PasswordPolicy.RequireUppercase,
			RequireLowercase: config.Security.PasswordPolicy.RequireLowercase,
			RequireNumbers:   config.Security.PasswordPolicy.RequireNumbers,
			RequireSymbols:   config.Security.PasswordPolicy.RequireSymbols,
		},
	}
	passwordService := security.NewPasswordService(passwordConfig, logger)

	// Initialize security service (placeholder implementation)
	securityService := &MockSecurityService{logger: logger}

	// Initialize auth service
	authConfig := &application.AuthConfig{
		AccessTokenTTL:   config.Security.AccessTokenTTL,
		RefreshTokenTTL:  config.Security.RefreshTokenTTL,
		MaxLoginAttempts: config.Security.MaxLoginAttempts,
		LockoutDuration:  config.Security.LockoutDuration,
	}
	authService := application.NewAuthService(
		userRepo,
		sessionRepo,
		jwtService,
		passwordService,
		securityService,
		authConfig,
		logger,
	)

	// Start HTTP server
	httpServer := startHTTPServer(config, authService, logger)

	// Start gRPC server
	grpcServer := startGRPCServer(config, authService, logger)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Auth Service is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Auth Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Info("✅ Auth Service stopped gracefully")
}

// loadConfig loads configuration from file and environment variables
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cmd/auth-service/config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

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

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.grpc_port", 50053)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	viper.SetDefault("security.access_token_ttl", "15m")
	viper.SetDefault("security.refresh_token_ttl", "168h")
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("security.max_login_attempts", 5)
	viper.SetDefault("security.lockout_duration", "30m")

	viper.SetDefault("environment", "development")
}

// initializeRedis initializes Redis client
func initializeRedis(config *Config, logger *logger.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(config.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	opt.DB = config.Redis.DB
	opt.MaxRetries = config.Redis.MaxRetries
	opt.PoolSize = config.Redis.PoolSize
	opt.MinIdleConns = config.Redis.MinIdleConns
	opt.DialTimeout = config.Redis.DialTimeout
	opt.ReadTimeout = config.Redis.ReadTimeout
	opt.WriteTimeout = config.Redis.WriteTimeout
	opt.PoolTimeout = config.Redis.PoolTimeout
	opt.IdleTimeout = config.Redis.IdleTimeout

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("✅ Connected to Redis successfully")
	return client, nil
}

// MockSecurityService is a placeholder implementation
type MockSecurityService struct {
	logger *logger.Logger
}

func (m *MockSecurityService) LogSecurityEvent(ctx context.Context, userID string, eventType domain.SecurityEventType, severity domain.SecuritySeverity, description string, metadata map[string]string) error {
	m.logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"event_type":  string(eventType),
		"severity":    string(severity),
		"description": description,
	}).Info("Security event logged")
	return nil
}

func (m *MockSecurityService) CheckRateLimit(ctx context.Context, key string) error {
	// Placeholder implementation
	return nil
}

func (m *MockSecurityService) TrackFailedLogin(ctx context.Context, email string) error {
	// Placeholder implementation
	return nil
}

func (m *MockSecurityService) ResetFailedLoginCount(ctx context.Context, email string) error {
	// Placeholder implementation
	return nil
}

func (m *MockSecurityService) IsAccountLocked(ctx context.Context, email string) (bool, error) {
	// Placeholder implementation
	return false, nil
}

func (m *MockSecurityService) CheckAccountSecurity(ctx context.Context, userID string) error {
	// Placeholder implementation
	return nil
}

func (m *MockSecurityService) LockAccount(ctx context.Context, userID string, reason string) error {
	// Placeholder implementation
	m.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"reason":  reason,
	}).Info("Account locked")
	return nil
}

func (m *MockSecurityService) UnlockAccount(ctx context.Context, userID string) error {
	// Placeholder implementation
	m.logger.WithField("user_id", userID).Info("Account unlocked")
	return nil
}

func (m *MockSecurityService) IncrementRateLimit(ctx context.Context, key string) error {
	// Placeholder implementation
	return nil
}

func (m *MockSecurityService) GetSecurityEvents(ctx context.Context, userID string, limit int) ([]*application.SecurityEventDTO, error) {
	// Placeholder implementation
	return []*application.SecurityEventDTO{}, nil
}

// startHTTPServer starts the HTTP server
func startHTTPServer(config *Config, authService *application.AuthServiceImpl, logger *logger.Logger) *http.Server {
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": serviceName,
			"time":    time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister(authService, logger))
			auth.POST("/login", handleLogin(authService, logger))
			auth.POST("/logout", handleLogout(authService, logger))
			auth.POST("/refresh", handleRefreshToken(authService, logger))
			auth.POST("/validate", handleValidateToken(authService, logger))
			auth.POST("/change-password", handleChangePassword(authService, logger))
			auth.GET("/me", handleGetUserInfo(authService, logger))
		}
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	go func() {
		logger.WithField("port", config.Server.HTTPPort).Info("🌐 HTTP Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	return server
}

// startGRPCServer starts the gRPC server
func startGRPCServer(config *Config, authService *application.AuthServiceImpl, logger *logger.Logger) *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Server.Host, config.Server.GRPCPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen for gRPC")
	}

	grpcServer := grpc.NewServer()

	// Register services here when gRPC handlers are implemented
	// auth.RegisterAuthServiceServer(grpcServer, grpcHandler)

	// Enable reflection for development
	if config.Environment == "development" {
		reflection.Register(grpcServer)
	}

	go func() {
		logger.WithField("port", config.Server.GRPCPort).Info("🌐 gRPC Server listening")
		if err := grpcServer.Serve(listener); err != nil {
			logger.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()

	return grpcServer
}

// HTTP Handlers (placeholder implementations)

func handleRegister(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := authService.Register(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Registration failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func handleLogin(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := authService.Login(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Login failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleLogout(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Extract user ID from token (placeholder)
		userID := "user-123" // This should be extracted from JWT token

		resp, err := authService.Logout(c.Request.Context(), userID, &req)
		if err != nil {
			logger.WithError(err).Error("Logout failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleRefreshToken(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := authService.RefreshToken(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Token refresh failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleValidateToken(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.ValidateTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := authService.ValidateToken(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Token validation failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleChangePassword(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Extract user ID from token (placeholder)
		userID := "user-123" // This should be extracted from JWT token

		resp, err := authService.ChangePassword(c.Request.Context(), userID, &req)
		if err != nil {
			logger.WithError(err).Error("Password change failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleGetUserInfo(authService *application.AuthServiceImpl, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID from token (placeholder)
		userID := "user-123" // This should be extracted from JWT token

		req := &application.GetUserInfoRequest{UserID: userID}
		resp, err := authService.GetUserInfo(c.Request.Context(), req)
		if err != nil {
			logger.WithError(err).Error("Get user info failed")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
