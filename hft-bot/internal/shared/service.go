package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/hft-bot/pkg/config"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// SimpleLogger implements Logger interface using standard log package
type SimpleLogger struct {
	prefix string
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger(prefix string) Logger {
	return &SimpleLogger{prefix: prefix}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("INFO", msg, keysAndValues...)
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("ERROR", msg, keysAndValues...)
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("WARN", msg, keysAndValues...)
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("DEBUG", msg, keysAndValues...)
}

// logWithLevel logs a message with the specified level
func (l *SimpleLogger) logWithLevel(level, msg string, keysAndValues ...interface{}) {
	logMsg := fmt.Sprintf("[%s] %s: %s", level, l.prefix, msg)
	
	// Add key-value pairs
	if len(keysAndValues) > 0 {
		logMsg += " |"
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
	}
	
	log.Println(logMsg)
}

// Service represents a base service with common functionality
type Service struct {
	name     string
	config   *config.Config
	logger   Logger
	server   *http.Server
	metrics  *http.Server
	health   *http.Server
	
	// Lifecycle management
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdownCh chan struct{}
	
	// Health status
	healthy    bool
	ready      bool
	mu         sync.RWMutex
}

// ServiceOption represents a configuration option for the service
type ServiceOption func(*Service)

// NewService creates a new service instance
func NewService(name string, cfg *config.Config, logger Logger, opts ...ServiceOption) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	
	s := &Service{
		name:       name,
		config:     cfg,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
		healthy:    false,
		ready:      false,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(s)
	}
	
	return s
}

// WithMetricsServer adds a metrics server to the service
func WithMetricsServer(port int) ServiceOption {
	return func(s *Service) {
		mux := http.NewServeMux()
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "# Simple metrics endpoint\n")
			fmt.Fprintf(w, "service_status{name=\"%s\"} 1\n", s.name)
			fmt.Fprintf(w, "service_healthy{name=\"%s\"} %d\n", s.name, boolToInt(s.IsHealthy()))
			fmt.Fprintf(w, "service_ready{name=\"%s\"} %d\n", s.name, boolToInt(s.IsReady()))
		})
		
		s.metrics = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}
	}
}

// WithHealthServer adds a health check server to the service
func WithHealthServer(port int) ServiceOption {
	return func(s *Service) {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", s.healthHandler)
		mux.HandleFunc("/ready", s.readyHandler)
		
		s.health = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}
	}
}

// WithHTTPServer adds an HTTP server to the service
func WithHTTPServer(port int, handler http.Handler) ServiceOption {
	return func(s *Service) {
		s.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		}
	}
}

// Start starts the service and all its components
func (s *Service) Start() error {
	s.logger.Info("Starting service", "name", s.name)
	
	// Start metrics server if configured
	if s.metrics != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Starting metrics server", "addr", s.metrics.Addr)
			if err := s.metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("Metrics server error", "error", err)
			}
		}()
	}
	
	// Start health server if configured
	if s.health != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Starting health server", "addr", s.health.Addr)
			if err := s.health.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("Health server error", "error", err)
			}
		}()
	}
	
	// Start main HTTP server if configured
	if s.server != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("HTTP server error", "error", err)
			}
		}()
	}
	
	// Mark as healthy and ready
	s.setHealthy(true)
	s.setReady(true)
	
	s.logger.Info("Service started successfully", "name", s.name)
	return nil
}

// Stop gracefully stops the service
func (s *Service) Stop() error {
	s.logger.Info("Stopping service", "name", s.name)
	
	// Mark as not ready
	s.setReady(false)
	
	// Cancel context
	s.cancel()
	
	// Shutdown servers with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Shutdown HTTP server
	if s.server != nil {
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down HTTP server", "error", err)
		}
	}
	
	// Shutdown metrics server
	if s.metrics != nil {
		if err := s.metrics.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down metrics server", "error", err)
		}
	}
	
	// Shutdown health server
	if s.health != nil {
		if err := s.health.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down health server", "error", err)
		}
	}
	
	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		s.logger.Info("All goroutines stopped")
	case <-shutdownCtx.Done():
		s.logger.Warn("Shutdown timeout reached")
	}
	
	// Mark as not healthy
	s.setHealthy(false)
	
	s.logger.Info("Service stopped", "name", s.name)
	return nil
}

// Run starts the service and waits for shutdown signals
func (s *Service) Run() error {
	// Start the service
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	
	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case sig := <-sigCh:
		s.logger.Info("Received shutdown signal", "signal", sig)
	case <-s.ctx.Done():
		s.logger.Info("Context cancelled")
	}
	
	// Stop the service
	return s.Stop()
}

// Context returns the service context
func (s *Service) Context() context.Context {
	return s.ctx
}

// Logger returns the service logger
func (s *Service) Logger() Logger {
	return s.logger
}

// Config returns the service configuration
func (s *Service) Config() *config.Config {
	return s.config
}

// IsHealthy returns the health status
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthy
}

// IsReady returns the readiness status
func (s *Service) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

// setHealthy sets the health status
func (s *Service) setHealthy(healthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthy = healthy
}

// setReady sets the readiness status
func (s *Service) setReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
}

// healthHandler handles health check requests
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	if s.IsHealthy() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"%s","timestamp":"%s"}`, 
			s.name, time.Now().Format(time.RFC3339))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","service":"%s","timestamp":"%s"}`, 
			s.name, time.Now().Format(time.RFC3339))
	}
}

// readyHandler handles readiness check requests
func (s *Service) readyHandler(w http.ResponseWriter, r *http.Request) {
	if s.IsReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"%s","timestamp":"%s"}`, 
			s.name, time.Now().Format(time.RFC3339))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","service":"%s","timestamp":"%s"}`, 
			s.name, time.Now().Format(time.RFC3339))
	}
}

// boolToInt converts boolean to integer for metrics
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
