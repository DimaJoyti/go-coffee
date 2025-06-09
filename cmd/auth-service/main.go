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
	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/config"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/repository"
	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/security"
	grpcTransport "github.com/DimaJoyti/go-coffee/internal/auth/transport/grpc"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

const (
	serviceName = "auth-service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Auth Service...")

	// Initialize Redis client
	redisClient, err := initializeRedis(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := repository.NewRedisUserRepository(redisClient, logger)
	sessionRepo := repository.NewRedisSessionRepository(redisClient, logger)

	// Initialize security services
	jwtConfig := &security.JWTConfig{
		Secret:          cfg.Security.JWT.Secret,
		AccessTokenTTL:  cfg.Security.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.Security.JWT.RefreshTokenTTL,
		Issuer:          cfg.Security.JWT.Issuer,
		Audience:        cfg.Security.JWT.Audience,
	}
	jwtService := security.NewJWTService(jwtConfig, logger)

	passwordConfig := &security.PasswordConfig{
		BcryptCost: cfg.Security.Password.BcryptCost,
		PasswordPolicy: &application.PasswordPolicy{
			MinLength:        cfg.Security.Password.Policy.MinLength,
			RequireUppercase: cfg.Security.Password.Policy.RequireUppercase,
			RequireLowercase: cfg.Security.Password.Policy.RequireLowercase,
			RequireNumbers:   cfg.Security.Password.Policy.RequireNumbers,
			RequireSymbols:   cfg.Security.Password.Policy.RequireSymbols,
		},
	}
	passwordService := security.NewPasswordService(passwordConfig, logger)

	// Initialize security service (placeholder implementation)
	securityService := &MockSecurityService{logger: logger}

	// Initialize auth service
	authConfig := &application.AuthConfig{
		AccessTokenTTL:   cfg.Security.JWT.AccessTokenTTL,
		RefreshTokenTTL:  cfg.Security.JWT.RefreshTokenTTL,
		MaxLoginAttempts: cfg.Security.Account.MaxLoginAttempts,
		LockoutDuration:  cfg.Security.Account.LockoutDuration,
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
	httpServer := startHTTPServer(cfg, authService, logger)

	// Start gRPC server
	grpcServer := grpcTransport.NewServer(cfg, authService, logger)
	if err := grpcServer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start gRPC server")
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Auth Service is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Auth Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	// Shutdown gRPC server
	if err := grpcServer.Stop(); err != nil {
		logger.WithError(err).Error("gRPC server shutdown error")
	}

	logger.Info("✅ Auth Service stopped gracefully")
}

// initializeRedis initializes Redis client
func initializeRedis(cfg *config.Config, logger *logger.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	opt.DB = cfg.Redis.DB
	opt.MaxRetries = cfg.Redis.MaxRetries
	opt.PoolSize = cfg.Redis.PoolSize
	opt.MinIdleConns = cfg.Redis.MinIdleConns
	opt.DialTimeout = cfg.Redis.DialTimeout
	opt.ReadTimeout = cfg.Redis.ReadTimeout
	opt.WriteTimeout = cfg.Redis.WriteTimeout
	opt.PoolTimeout = cfg.Redis.PoolTimeout
	opt.IdleTimeout = cfg.Redis.IdleTimeout

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
func startHTTPServer(cfg *config.Config, authService *application.AuthServiceImpl, logger *logger.Logger) *http.Server {
	if cfg.Environment == "production" {
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
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.WithField("port", cfg.Server.HTTPPort).Info("🌐 HTTP Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	return server
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
