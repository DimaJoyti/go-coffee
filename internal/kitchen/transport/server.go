package transport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/transport/grpc"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/kitchen/transport/http"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/transport/middleware"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/transport/websocket"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Server represents the main transport server
type Server struct {
	config          *Config
	logger          *logger.Logger
	httpServer      *http.Server
	grpcServer      *grpc.Server
	websocketServer *websocket.Server

	// Services
	kitchenService      application.KitchenService
	queueService        application.QueueService
	optimizerService    application.OptimizerService
	notificationService application.NotificationService
	eventService        application.EventService
}

// Config represents server configuration
type Config struct {
	HTTPPort     string `json:"http_port" env:"HTTP_PORT" default:"8080"`
	GRPCPort     string `json:"grpc_port" env:"GRPC_PORT" default:"9090"`
	JWTSecret    string `json:"jwt_secret" env:"JWT_SECRET" default:"your-secret-key"`
	EnableCORS   bool   `json:"enable_cors" env:"ENABLE_CORS" default:"true"`
	EnableAuth   bool   `json:"enable_auth" env:"ENABLE_AUTH" default:"true"`
	ReadTimeout  int    `json:"read_timeout" env:"READ_TIMEOUT" default:"30"`
	WriteTimeout int    `json:"write_timeout" env:"WRITE_TIMEOUT" default:"30"`
	IdleTimeout  int    `json:"idle_timeout" env:"IDLE_TIMEOUT" default:"120"`
}

// NewServer creates a new transport server
func NewServer(
	config *Config,
	logger *logger.Logger,
	kitchenService application.KitchenService,
	queueService application.QueueService,
	optimizerService application.OptimizerService,
	notificationService application.NotificationService,
	eventService application.EventService,
) *Server {
	return &Server{
		config:              config,
		logger:              logger,
		kitchenService:      kitchenService,
		queueService:        queueService,
		optimizerService:    optimizerService,
		notificationService: notificationService,
		eventService:        eventService,
	}
}

// Start starts all transport servers
func (s *Server) Start() error {
	s.logger.Info("Starting kitchen transport servers")

	// Start gRPC server
	s.grpcServer = grpc.NewServer(
		s.kitchenService,
		s.queueService,
		s.optimizerService,
		s.notificationService,
		s.logger,
		s.config.GRPCPort,
	)
	go func() {
		if err := s.grpcServer.Start(); err != nil {
			s.logger.WithError(err).Error("gRPC server failed to start")
		}
	}()

	// Start WebSocket server
	s.websocketServer = websocket.NewServer(
		s.kitchenService,
		s.queueService,
		s.eventService,
		s.logger,
	)
	s.websocketServer.Start()

	// Setup HTTP server with all routes
	if err := s.setupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	// Start HTTP server
	go func() {
		s.logger.WithField("port", s.config.HTTPPort).Info("Starting HTTP server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server failed")
		}
	}()

	// Setup event handlers for WebSocket
	s.setupEventHandlers()

	s.logger.Info("All transport servers started successfully")
	return nil
}

// Stop stops all transport servers gracefully
func (s *Server) Stop() error {
	s.logger.Info("Stopping kitchen transport servers")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to shutdown HTTP server")
		}
	}

	// Stop gRPC server
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}

	// Stop WebSocket server
	if s.websocketServer != nil {
		s.websocketServer.Stop()
	}

	s.logger.Info("All transport servers stopped")
	return nil
}

// setupHTTPServer sets up the HTTP server with all routes and middleware
func (s *Server) setupHTTPServer() error {
	router := mux.NewRouter()

	// Setup middleware
	s.setupMiddleware(router)

	// Setup routes
	s.setupRoutes(router)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.HTTPPort,
		Handler:      router,
		ReadTimeout:  time.Duration(s.config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.IdleTimeout) * time.Second,
	}

	return nil
}

// setupMiddleware sets up HTTP middleware
func (s *Server) setupMiddleware(router *mux.Router) {
	// CORS middleware
	if s.config.EnableCORS {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"}, // In production, specify allowed origins
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			ExposedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
		router.Use(c.Handler)
	}

	// Logging middleware
	router.Use(s.loggingMiddleware)

	// Recovery middleware
	router.Use(s.recoveryMiddleware)

	// Authentication middleware (if enabled)
	if s.config.EnableAuth {
		authInterceptor := middleware.NewAuthInterceptor(s.logger, s.config.JWTSecret)
		router.Use(s.authMiddleware(authInterceptor))
	}
}

// setupRoutes sets up all HTTP routes
func (s *Server) setupRoutes(router *mux.Router) {
	// Create HTTP handler
	httpHandler := httpTransport.NewHandler(
		s.kitchenService,
		s.queueService,
		s.optimizerService,
		s.notificationService,
		s.logger,
	)

	// Register API routes
	httpHandler.RegisterRoutes(router)

	// WebSocket endpoint
	router.HandleFunc("/ws", s.websocketServer.HandleWebSocket)

	// Static files (if needed)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"kitchen-transport"}`))
	}).Methods("GET")
}

// setupEventHandlers sets up event handlers for WebSocket broadcasting
func (s *Server) setupEventHandlers() {
	// Subscribe to domain events and broadcast via WebSocket
	eventTypes := []string{
		"kitchen.order.added_to_queue",
		"kitchen.order.status_changed",
		"kitchen.order.completed",
		"kitchen.order.overdue",
		"kitchen.equipment.status_changed",
		"kitchen.staff.assigned",
		"kitchen.queue.status_changed",
	}

	for _, eventType := range eventTypes {
		s.eventService.SubscribeToEvents(context.Background(), []string{eventType},
			domain.EventHandlerFunc(func(event *domain.DomainEvent) error {
				switch {
				case strings.HasPrefix(event.Type, "kitchen.order."):
					s.websocketServer.HandleOrderEvent(event)
				case strings.HasPrefix(event.Type, "kitchen.equipment."):
					s.websocketServer.HandleEquipmentEvent(event)
				case strings.HasPrefix(event.Type, "kitchen.staff."):
					s.websocketServer.HandleStaffEvent(event)
				case strings.HasPrefix(event.Type, "kitchen.queue."):
					s.websocketServer.HandleQueueEvent(event)
				}
				return nil
			}))
	}
}

// Middleware functions

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		s.logger.WithFields(map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapper.statusCode,
			"duration":    duration.String(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("HTTP request completed")
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"panic":  err,
				}).Error("HTTP request panicked")

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// authMiddleware provides authentication for HTTP requests
func (s *Server) authMiddleware(authInterceptor *middleware.AuthInterceptor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public endpoints
			if s.isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract and validate token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// For HTTP, we'll do a simplified token validation
			// In production, integrate with the gRPC auth interceptor
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" || token == "invalid" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to request context (simplified)
			ctx := context.WithValue(r.Context(), "user_id", "user_123")
			ctx = context.WithValue(ctx, "role", "staff")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicEndpoint checks if an endpoint is public
func (s *Server) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/api/v1/kitchen/health",
		"/api/v1/kitchen/queue/status", // Public queue status
		"/ws",                          // WebSocket endpoint handles auth separately
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RunWithGracefulShutdown runs the server with graceful shutdown
func (s *Server) RunWithGracefulShutdown() error {
	// Start the server
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.logger.Info("Shutting down server...")

	// Gracefully shutdown the server
	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	s.logger.Info("Server exited")
	return nil
}
