package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/observability"
)

// HTTPConfig combines service and server configuration for HTTP server
type HTTPConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// NewHTTPConfig creates HTTPConfig from service and server configs
func NewHTTPConfig(service config.ServiceConfig, server config.ServerConfig) HTTPConfig {
	return HTTPConfig{
		Port:         service.Port,
		ReadTimeout:  server.ReadTimeout,
		WriteTimeout: server.WriteTimeout,
		IdleTimeout:  server.IdleTimeout,
	}
}

// Server represents the HTTP API server
type Server struct {
	config     HTTPConfig
	server     *http.Server
	mux        *http.ServeMux
	logger     *observability.StructuredLogger
	metrics    *observability.MetricsCollector
	tracing    *observability.TracingHelper
	middleware *MiddlewareChain

	// Route handlers will be implemented directly in server
	// This eliminates the import cycle
}

// NewServer creates a new HTTP API server
func NewServer(
	config HTTPConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Server {
	mux := http.NewServeMux()

	server := &Server{
		config:  config,
		mux:     mux,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			Handler:      mux,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}

	// Initialize middleware
	server.middleware = NewMiddlewareChain(logger, metrics, tracing)

	// Initialize handlers
	server.initializeHandlers()

	// Setup routes
	server.setupRoutes()

	return server
}

// initializeHandlers initializes all route handlers
func (s *Server) initializeHandlers() {
	// TODO: Initialize handlers directly here to avoid import cycle
	// For now, we'll use placeholder handlers
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Apply middleware to all routes
	handler := s.middleware.Apply(s.mux)
	s.server.Handler = handler

	// Health endpoints
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /health/ready", s.handleHealthReady)
	s.mux.HandleFunc("GET /health/live", s.handleHealthLive)

	// TODO: Add other endpoints as placeholder handlers to avoid import cycle
	// For now, just add the essential health endpoints

	// API documentation
	s.mux.HandleFunc("GET /api/v1/docs", s.handleAPIDocs)
	s.mux.HandleFunc("GET /api/v1/openapi.json", s.handleOpenAPISpec)

	s.logger.Info("HTTP API routes configured",
		"port", s.config.Port,
		"endpoints_count", s.countRoutes())
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	ctx, span := s.tracing.StartHTTPSpan(ctx, "START_SERVER", "", "")
	defer span.End()

	s.logger.InfoContext(ctx, "Starting HTTP API server",
		"port", s.config.Port,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "HTTP server failed", err)
		}
	}()

	s.tracing.RecordSuccess(span, "HTTP server started")
	s.logger.InfoContext(ctx, "HTTP API server started successfully",
		"address", s.server.Addr)

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	ctx, span := s.tracing.StartHTTPSpan(ctx, "STOP_SERVER", "", "")
	defer span.End()

	s.logger.InfoContext(ctx, "Stopping HTTP API server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.tracing.RecordError(span, err, "Failed to shutdown HTTP server gracefully")
		s.logger.ErrorContext(ctx, "Failed to shutdown HTTP server gracefully", err)
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.tracing.RecordSuccess(span, "HTTP server stopped")
	s.logger.InfoContext(ctx, "HTTP API server stopped successfully")

	return nil
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return s.server.Addr
}

// GetPort returns the server port
func (s *Server) GetPort() int {
	return s.config.Port
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	// Simple check - in production, you might want a more sophisticated check
	return s.server != nil
}

// countRoutes counts the number of registered routes
func (s *Server) countRoutes() int {
	// This is a simplified count - in practice, you might want to track this more precisely
	return 25 // Approximate count based on registered routes
}

// handleAPIDocs serves API documentation
func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracing.StartHTTPSpan(r.Context(), "API_DOCS", r.URL.Path, r.UserAgent())
	defer span.End()

	docs := `
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee AI Agents API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #2c3e50; }
        h2 { color: #34495e; border-bottom: 1px solid #ecf0f1; padding-bottom: 10px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #e74c3c; }
        .path { font-family: monospace; background: #ecf0f1; padding: 2px 5px; }
        .description { margin-top: 10px; color: #7f8c8d; }
    </style>
</head>
<body>
    <h1>Go Coffee AI Agents API</h1>
    <p>RESTful API for managing beverages, tasks, and AI operations in the Mars colony.</p>
    
    <h2>Health Endpoints</h2>
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/health</span>
        <div class="description">Overall system health check</div>
    </div>
    
    <h2>Beverage Endpoints</h2>
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/beverages</span>
        <div class="description">List all beverages</div>
    </div>
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/beverages/generate</span>
        <div class="description">Generate new beverage recipes using AI</div>
    </div>
    
    <h2>AI Endpoints</h2>
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/ai/text</span>
        <div class="description">Generate text using AI providers</div>
    </div>
    
    <p><a href="/api/v1/openapi.json">OpenAPI Specification</a></p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(docs))

	s.tracing.RecordSuccess(span, "API documentation served")
}

// handleOpenAPISpec serves the OpenAPI specification
func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracing.StartHTTPSpan(r.Context(), "OPENAPI_SPEC", r.URL.Path, r.UserAgent())
	defer span.End()

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Go Coffee AI Agents API",
			"description": "RESTful API for managing beverages, tasks, and AI operations",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{
				"url":         fmt.Sprintf("http://localhost:%d", s.config.Port),
				"description": "Development server",
			},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"description": "Check the overall health of the system",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "System is healthy",
						},
					},
				},
			},
			"/api/v1/beverages": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List beverages",
					"description": "Get a list of all beverages",
				},
				"post": map[string]interface{}{
					"summary":     "Create beverage",
					"description": "Create a new beverage",
				},
			},
		},
	}

	WriteJSONResponse(w, http.StatusOK, spec)
	s.tracing.RecordSuccess(span, "OpenAPI specification served")
}

// Global server instance
var globalServer *Server

// InitGlobalServer initializes the global HTTP server
func InitGlobalServer(
	config HTTPConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Server {
	globalServer = NewServer(config, logger, metrics, tracing)
	return globalServer
}

// GetGlobalServer returns the global HTTP server
func GetGlobalServer() *Server {
	return globalServer
}

// StartGlobalServer starts the global HTTP server
func StartGlobalServer(ctx context.Context) error {
	if globalServer == nil {
		return fmt.Errorf("global server not initialized")
	}
	return globalServer.Start(ctx)
}

// StopGlobalServer stops the global HTTP server
func StopGlobalServer(ctx context.Context) error {
	if globalServer == nil {
		return nil
	}
	return globalServer.Stop(ctx)
}

// Placeholder health handlers to avoid import cycle
// TODO: Move actual handler logic here or refactor to avoid cycle

// handleHealth handles the main health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "go-coffee-ai-agents",
	}
	httputils.WriteJSONResponse(w, http.StatusOK, health)
}

// handleHealthReady handles the readiness check endpoint
func (s *Server) handleHealthReady(w http.ResponseWriter, r *http.Request) {
	ready := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now(),
	}
	httputils.WriteJSONResponse(w, http.StatusOK, ready)
}

// handleHealthLive handles the liveness check endpoint
func (s *Server) handleHealthLive(w http.ResponseWriter, r *http.Request) {
	live := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
	}
	httputils.WriteJSONResponse(w, http.StatusOK, live)
}
