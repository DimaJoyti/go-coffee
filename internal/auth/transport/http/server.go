package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Server represents HTTP server for auth service
type Server struct {
	handler     *Handler
	server      *http.Server
	logger      *logger.Logger
	port        int
}

// Config represents HTTP server configuration
type Config struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// NewServer creates a new HTTP server
func NewServer(
	config *Config,
	authService application.AuthService,
	mfaService application.MFAService,
	logger *logger.Logger,
) *Server {
	handler := NewHandler(authService, mfaService, logger)
	
	return &Server{
		handler: handler,
		logger:  logger,
		port:    config.Port,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(s.handler.corsMiddleware)
	router.Use(s.handler.loggingMiddleware)
	router.Use(s.handler.recoveryMiddleware)
	router.Use(s.handler.securityHeadersMiddleware)
	router.Use(s.handler.requestIDMiddleware)
	router.Use(s.handler.deviceFingerprintMiddleware)

	// Apply content type validation for POST/PUT requests
	router.Use(s.handler.contentTypeMiddleware([]string{
		"application/json",
		"application/x-www-form-urlencoded",
	}))

	// Register routes
	s.handler.RegisterRoutes(router)

	// Set router to server
	s.server.Handler = router

	s.logger.WithField("port", s.port).Info("Starting HTTP server")
	
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop stops the HTTP server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	
	return s.server.Shutdown(ctx)
}

// GetPort returns the server port
func (s *Server) GetPort() int {
	return s.port
}

// GetHandler returns the HTTP handler
func (s *Server) GetHandler() *Handler {
	return s.handler
}

// Health check endpoint for load balancers
func (s *Server) HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"auth-service"}`))
	}
}

// Metrics endpoint for monitoring
func (s *Server) Metrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metrics":"placeholder"}`))
	}
}

// Ready endpoint for readiness probes
func (s *Server) Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add actual readiness checks (database connectivity, etc.)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","service":"auth-service"}`))
	}
}

// DefaultConfig returns default HTTP server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:         8080,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// ServerWithGracefulShutdown wraps the server with graceful shutdown capability
type ServerWithGracefulShutdown struct {
	*Server
	shutdownTimeout time.Duration
}

// NewServerWithGracefulShutdown creates a server with graceful shutdown
func NewServerWithGracefulShutdown(
	config *Config,
	authService application.AuthService,
	mfaService application.MFAService,
	logger *logger.Logger,
	shutdownTimeout time.Duration,
) *ServerWithGracefulShutdown {
	server := NewServer(config, authService, mfaService, logger)
	
	return &ServerWithGracefulShutdown{
		Server:          server,
		shutdownTimeout: shutdownTimeout,
	}
}

// StartWithGracefulShutdown starts the server and handles graceful shutdown
func (s *ServerWithGracefulShutdown) StartWithGracefulShutdown(ctx context.Context) error {
	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		s.logger.Info("Received shutdown signal")
		
		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		
		// Attempt graceful shutdown
		if err := s.Stop(shutdownCtx); err != nil {
			s.logger.WithError(err).Error("Failed to shutdown server gracefully")
			return err
		}
		
		s.logger.Info("Server shutdown completed")
		return nil
	}
}

// SetupRoutes sets up additional routes for the server
func (s *Server) SetupRoutes(setupFunc func(*mux.Router)) {
	if s.server.Handler != nil {
		if router, ok := s.server.Handler.(*mux.Router); ok {
			setupFunc(router)
		}
	}
}

// AddMiddleware adds middleware to the server
func (s *Server) AddMiddleware(middleware ...mux.MiddlewareFunc) {
	if s.server.Handler != nil {
		if router, ok := s.server.Handler.(*mux.Router); ok {
			router.Use(middleware...)
		}
	}
}

// GetRouter returns the underlying router
func (s *Server) GetRouter() *mux.Router {
	if s.server.Handler != nil {
		if router, ok := s.server.Handler.(*mux.Router); ok {
			return router
		}
	}
	return nil
}

// ServerInfo represents server information
type ServerInfo struct {
	Port      int    `json:"port"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

// GetServerInfo returns server information
func (s *Server) GetServerInfo() *ServerInfo {
	return &ServerInfo{
		Port:      s.port,
		Status:    "running",
		Version:   "1.0.0", // TODO: Get from build info
		BuildTime: time.Now().Format(time.RFC3339),
	}
}

// EnableProfiling enables Go profiling endpoints
func (s *Server) EnableProfiling() {
	if router := s.GetRouter(); router != nil {
		// Add pprof routes for debugging
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}
}

// EnableCORS enables CORS with custom configuration
func (s *Server) EnableCORS(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) {
	// This would be implemented with a more sophisticated CORS middleware
	// For now, the basic CORS middleware in handlers.go is sufficient
}
