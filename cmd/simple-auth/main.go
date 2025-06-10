package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// SimpleAuthService provides a minimal auth service for testing
type SimpleAuthService struct {
	logger *logger.Logger
}

// NewSimpleAuthService creates a new simple auth service
func NewSimpleAuthService(logger *logger.Logger) *SimpleAuthService {
	return &SimpleAuthService{
		logger: logger,
	}
}

// RegisterHandler handles user registration
func (s *SimpleAuthService) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Registration request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{
		"success": true,
		"message": "User registered successfully",
		"user": {
			"id": "user_123",
			"email": "test@example.com",
			"role": "user"
		},
		"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"refresh_token": "refresh_token_123",
		"expires_in": 900,
		"token_type": "Bearer"
	}`))
}

// LoginHandler handles user login
func (s *SimpleAuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Login request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"success": true,
		"message": "Login successful",
		"user": {
			"id": "user_123",
			"email": "test@example.com",
			"role": "user"
		},
		"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"refresh_token": "refresh_token_123",
		"expires_in": 900,
		"token_type": "Bearer"
	}`))
}

// ValidateHandler handles token validation
func (s *SimpleAuthService) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Token validation request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"valid": true,
		"user_id": "user_123",
		"role": "user",
		"session_id": "session_123",
		"user": {
			"id": "user_123",
			"email": "test@example.com",
			"role": "user"
		}
	}`))
}

// UserInfoHandler handles getting user info
func (s *SimpleAuthService) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("User info request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"user": {
			"id": "user_123",
			"email": "test@example.com",
			"role": "user",
			"status": "active",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	}`))
}

// LogoutHandler handles user logout
func (s *SimpleAuthService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Logout request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"success": true,
		"message": "Logged out successfully"
	}`))
}

// RefreshHandler handles token refresh
func (s *SimpleAuthService) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Token refresh request received")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"refresh_token": "refresh_token_456",
		"expires_in": 900,
		"token_type": "Bearer"
	}`))
}

// HealthHandler handles health checks
func (s *SimpleAuthService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status": "healthy",
		"service": "auth-service",
		"version": "1.0.0",
		"timestamp": "` + time.Now().Format(time.RFC3339) + `"
	}`))
}

// SetupRoutes sets up the HTTP routes
func (s *SimpleAuthService) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1/auth").Subrouter()
	api.HandleFunc("/register", s.RegisterHandler).Methods("POST")
	api.HandleFunc("/login", s.LoginHandler).Methods("POST")
	api.HandleFunc("/logout", s.LogoutHandler).Methods("POST")
	api.HandleFunc("/validate", s.ValidateHandler).Methods("POST")
	api.HandleFunc("/refresh", s.RefreshHandler).Methods("POST")
	api.HandleFunc("/me", s.UserInfoHandler).Methods("GET")

	// Health check
	router.HandleFunc("/health", s.HealthHandler).Methods("GET")

	return router
}

func main() {
	fmt.Println("🚀 Starting Simple Auth Service")
	fmt.Println("===============================")

	// Initialize logger
	log := logger.New("simple-auth-service")

	// Create simple auth service
	authService := NewSimpleAuthService(log)

	// Setup routes
	router := authService.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.ErrorWithFields("Server error", logger.Error(err))
	case sig := <-sigChan:
		log.InfoWithFields("Received shutdown signal", logger.String("signal", sig.String()))
	}

	// Graceful shutdown
	log.Info("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.ErrorWithFields("Failed to shutdown HTTP server gracefully", logger.Error(err))
	}

	fmt.Println()
	fmt.Println("✅ Simple Auth Service stopped")
	fmt.Println()
	fmt.Println("📚 Available endpoints:")
	fmt.Println("  POST /api/v1/auth/register   - Register new user")
	fmt.Println("  POST /api/v1/auth/login      - Login user")
	fmt.Println("  POST /api/v1/auth/logout     - Logout user")
	fmt.Println("  POST /api/v1/auth/validate   - Validate token")
	fmt.Println("  GET  /api/v1/auth/me         - Get user info")
	fmt.Println("  POST /api/v1/auth/refresh    - Refresh token")
	fmt.Println("  GET  /health                 - Health check")
	fmt.Println()
	fmt.Println("🧪 Test with curl:")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"SecurePass123!","role":"user"}'`)
}
