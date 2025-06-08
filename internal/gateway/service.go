package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Service handles API gateway operations
type Service struct {
	config  *config.Config
	logger  *logger.Logger
	client  *http.Client
	services map[string]string
}

// ServiceEndpoint represents a service endpoint configuration
type ServiceEndpoint struct {
	Name    string
	BaseURL string
	Health  string
}

// NewService creates a new API gateway service
func NewService(cfg *config.Config, log *logger.Logger) (*Service, error) {
	// HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Service registry - maps service names to their base URLs
	services := map[string]string{
		"auth":    fmt.Sprintf("http://localhost:%d", cfg.Server.AuthServicePort),
		"payment": fmt.Sprintf("http://localhost:%d", cfg.Server.PaymentServicePort),
		"order":   fmt.Sprintf("http://localhost:%d", cfg.Server.OrderServicePort),
		"kitchen": fmt.Sprintf("http://localhost:%d", cfg.Server.KitchenServicePort),
	}

	return &Service{
		config:   cfg,
		logger:   log,
		client:   client,
		services: services,
	}, nil
}

// ProxyRequest proxies a request to the appropriate service
func (s *Service) ProxyRequest(ctx context.Context, serviceName, path string, method string, body []byte, headers map[string]string) (*http.Response, error) {
	baseURL, exists := s.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Construct the full URL
	url := baseURL + path

	s.logger.Info("Proxying request", "service", serviceName, "method", method, "url", url)

	// Create request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add gateway headers
	req.Header.Set("X-Gateway", "go-coffee-api-gateway")
	req.Header.Set("X-Request-ID", generateRequestID())

	// Make request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to proxy request", "error", err, "service", serviceName, "url", url)
		return nil, fmt.Errorf("failed to proxy request: %w", err)
	}

	s.logger.Info("Request proxied successfully", "service", serviceName, "status", resp.StatusCode)
	return resp, nil
}

// CheckServiceHealth checks the health of a service
func (s *Service) CheckServiceHealth(ctx context.Context, serviceName string) (bool, error) {
	baseURL, exists := s.services[serviceName]
	if !exists {
		return false, fmt.Errorf("service %s not found", serviceName)
	}

	url := baseURL + "/health"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetServiceStatus returns the status of all services
func (s *Service) GetServiceStatus(ctx context.Context) map[string]interface{} {
	status := make(map[string]interface{})
	
	for serviceName := range s.services {
		healthy, err := s.CheckServiceHealth(ctx, serviceName)
		serviceStatus := map[string]interface{}{
			"healthy": healthy,
			"url":     s.services[serviceName],
		}
		
		if err != nil {
			serviceStatus["error"] = err.Error()
		}
		
		status[serviceName] = serviceStatus
	}
	
	return status
}

// RouteRequest determines which service should handle the request
func (s *Service) RouteRequest(path string) (string, string, error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	// Split path into segments
	segments := strings.Split(path, "/")
	if len(segments) < 3 {
		return "", "", fmt.Errorf("invalid path format: %s", path)
	}

	// Expected format: /api/v1/{service}/{endpoint}
	if segments[0] != "api" || segments[1] != "v1" {
		return "", "", fmt.Errorf("invalid API version: %s", path)
	}

	serviceName := segments[2]
	
	// Map service names to internal service names
	switch serviceName {
	case "auth":
		return "auth", "/" + strings.Join(segments, "/"), nil
	case "payment":
		return "payment", "/" + strings.Join(segments, "/"), nil
	case "order", "orders":
		return "order", "/" + strings.Join(segments, "/"), nil
	case "kitchen":
		return "kitchen", "/" + strings.Join(segments, "/"), nil
	default:
		return "", "", fmt.Errorf("unknown service: %s", serviceName)
	}
}

// LoadBalancer (placeholder for future implementation)
type LoadBalancer struct {
	services map[string][]string
}

// GetServiceInstance returns a service instance (round-robin for now)
func (lb *LoadBalancer) GetServiceInstance(serviceName string) (string, error) {
	instances, exists := lb.services[serviceName]
	if !exists || len(instances) == 0 {
		return "", fmt.Errorf("no instances available for service %s", serviceName)
	}
	
	// Simple round-robin (in production, use more sophisticated algorithms)
	return instances[0], nil
}

// Middleware functions

// LoggingMiddleware logs all requests
func (s *Service) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		s.logger.Info("Incoming request", 
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
		
		next(w, r)
		
		duration := time.Since(start)
		s.logger.Info("Request completed", 
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration.String(),
		)
	}
}

// CORSMiddleware handles CORS
func (s *Service) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// AuthMiddleware validates JWT tokens (placeholder)
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and auth endpoints
		if strings.HasSuffix(r.URL.Path, "/health") || strings.Contains(r.URL.Path, "/auth/") {
			next(w, r)
			return
		}
		
		// Extract and validate JWT token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		// In a real implementation, validate the JWT token here
		// For now, just check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		next(w, r)
	}
}

// RateLimitMiddleware implements rate limiting (placeholder)
func (s *Service) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, implement rate limiting here
		// For now, just pass through
		next(w, r)
	}
}

// Helper functions

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
