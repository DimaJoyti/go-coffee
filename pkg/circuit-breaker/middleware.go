package circuitbreaker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPMiddleware provides circuit breaker middleware for HTTP requests
type HTTPMiddleware struct {
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
	config   *MiddlewareConfig
}

// MiddlewareConfig contains middleware configuration
type MiddlewareConfig struct {
	DefaultConfig    *Config
	PathConfigs      map[string]*Config
	FallbackHandler  gin.HandlerFunc
	SkipPaths        []string
	KeyGenerator     func(*gin.Context) string
	ErrorHandler     func(*gin.Context, error)
}

// NewHTTPMiddleware creates a new HTTP middleware
func NewHTTPMiddleware(logger *zap.Logger, config *MiddlewareConfig) *HTTPMiddleware {
	if config == nil {
		config = &MiddlewareConfig{
			DefaultConfig: DefaultConfig("http"),
		}
	}

	if config.DefaultConfig == nil {
		config.DefaultConfig = DefaultConfig("http")
	}

	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *gin.Context) string {
			return fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())
		}
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *gin.Context, err error) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"message": err.Error(),
				"code":    "CIRCUIT_BREAKER_OPEN",
			})
		}
	}

	return &HTTPMiddleware{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
		config:   config,
	}
}

// Middleware returns the Gin middleware function
func (m *HTTPMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if path is in skip list
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Generate circuit breaker key
		key := m.config.KeyGenerator(c)

		// Get or create circuit breaker
		breaker := m.getOrCreateBreaker(key, c.FullPath())

		// Execute request with circuit breaker protection
		err := breaker.Execute(c.Request.Context(), func() error {
			c.Next()

			// Check if response indicates failure
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}

			// Check if context was cancelled
			if c.Request.Context().Err() != nil {
				return c.Request.Context().Err()
			}

			return nil
		})

		// Handle circuit breaker errors
		if err != nil {
			// Don't override response if already written
			if !c.Writer.Written() {
				if breaker.IsOpen() && m.config.FallbackHandler != nil {
					m.logger.Info("Circuit breaker open, executing fallback",
						zap.String("key", key),
						zap.String("path", c.FullPath()),
					)
					m.config.FallbackHandler(c)
				} else {
					m.config.ErrorHandler(c, err)
				}
			}
			c.Abort()
		}
	}
}

// MiddlewareWithFallback returns middleware with custom fallback
func (m *HTTPMiddleware) MiddlewareWithFallback(fallback gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if path is in skip list
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		key := m.config.KeyGenerator(c)
		breaker := m.getOrCreateBreaker(key, c.FullPath())

		err := breaker.ExecuteWithFallback(
			c.Request.Context(),
			func() error {
				c.Next()
				if c.Writer.Status() >= 500 {
					return fmt.Errorf("server error: %d", c.Writer.Status())
				}
				return nil
			},
			func() error {
				if !c.Writer.Written() {
					fallback(c)
				}
				return nil
			},
		)

		if err != nil && !c.Writer.Written() {
			m.config.ErrorHandler(c, err)
			c.Abort()
		}
	}
}

// getOrCreateBreaker gets or creates a circuit breaker for the given key
func (m *HTTPMiddleware) getOrCreateBreaker(key, path string) *CircuitBreaker {
	if breaker, exists := m.breakers[key]; exists {
		return breaker
	}

	// Get config for this path
	config := m.getConfigForPath(path)
	config.Name = key

	// Create new circuit breaker
	breaker := NewCircuitBreaker(config, m.logger)
	
	// Set state change callback
	breaker.SetOnStateChange(func(name string, from State, to State) {
		m.logger.Info("Circuit breaker state changed",
			zap.String("name", name),
			zap.String("from", from.String()),
			zap.String("to", to.String()),
		)
	})

	m.breakers[key] = breaker
	return breaker
}

// getConfigForPath returns configuration for a specific path
func (m *HTTPMiddleware) getConfigForPath(path string) *Config {
	if m.config.PathConfigs != nil {
		if config, exists := m.config.PathConfigs[path]; exists {
			return config
		}
	}
	return m.config.DefaultConfig
}

// shouldSkip checks if path should be skipped
func (m *HTTPMiddleware) shouldSkip(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// GetStats returns statistics for all circuit breakers
func (m *HTTPMiddleware) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	for key, breaker := range m.breakers {
		stats[key] = breaker.Stats()
	}

	return stats
}

// GetBreaker returns a circuit breaker by key
func (m *HTTPMiddleware) GetBreaker(key string) *CircuitBreaker {
	return m.breakers[key]
}

// Reset resets all circuit breakers
func (m *HTTPMiddleware) Reset() {
	for _, breaker := range m.breakers {
		breaker.Reset()
	}
	m.logger.Info("All circuit breakers reset")
}

// ResetBreaker resets a specific circuit breaker
func (m *HTTPMiddleware) ResetBreaker(key string) error {
	if breaker, exists := m.breakers[key]; exists {
		breaker.Reset()
		return nil
	}
	return fmt.Errorf("circuit breaker not found: %s", key)
}

// gRPC Middleware

// GRPCMiddleware provides circuit breaker middleware for gRPC
type GRPCMiddleware struct {
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
	config   *Config
}

// NewGRPCMiddleware creates a new gRPC middleware
func NewGRPCMiddleware(logger *zap.Logger, config *Config) *GRPCMiddleware {
	if config == nil {
		config = DefaultConfig("grpc")
	}

	return &GRPCMiddleware{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
		config:   config,
	}
}

// UnaryInterceptor returns a gRPC unary interceptor
func (m *GRPCMiddleware) UnaryInterceptor() func(ctx context.Context, req interface{}, info interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
		// Extract method name from info
		methodName := fmt.Sprintf("%v", info)
		
		// Get or create circuit breaker
		breaker := m.getOrCreateGRPCBreaker(methodName)

		var result interface{}
		err := breaker.Execute(ctx, func() error {
			var handlerErr error
			result, handlerErr = handler(ctx, req)
			return handlerErr
		})

		return result, err
	}
}

// StreamInterceptor returns a gRPC stream interceptor
func (m *GRPCMiddleware) StreamInterceptor() func(srv interface{}, ss interface{}, info interface{}, handler func(interface{}, interface{}) error) error {
	return func(srv interface{}, ss interface{}, info interface{}, handler func(interface{}, interface{}) error) error {
		// Extract method name from info
		methodName := fmt.Sprintf("%v", info)
		
		// Get or create circuit breaker
		breaker := m.getOrCreateGRPCBreaker(methodName)

		return breaker.Execute(context.Background(), func() error {
			return handler(srv, ss)
		})
	}
}

// getOrCreateGRPCBreaker gets or creates a circuit breaker for gRPC method
func (m *GRPCMiddleware) getOrCreateGRPCBreaker(methodName string) *CircuitBreaker {
	if breaker, exists := m.breakers[methodName]; exists {
		return breaker
	}

	config := *m.config // Copy config
	config.Name = methodName

	breaker := NewCircuitBreaker(&config, m.logger)
	m.breakers[methodName] = breaker
	
	return breaker
}

// Client-side Circuit Breaker

// ClientCircuitBreaker provides circuit breaker for client calls
type ClientCircuitBreaker struct {
	*CircuitBreaker
}

// NewClientCircuitBreaker creates a new client circuit breaker
func NewClientCircuitBreaker(name string, config *Config, logger *zap.Logger) *ClientCircuitBreaker {
	if config == nil {
		config = DefaultConfig(name)
	}
	config.Name = name

	return &ClientCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(config, logger),
	}
}

// Call executes a client call with circuit breaker protection
func (ccb *ClientCircuitBreaker) Call(ctx context.Context, fn func() error) error {
	return ccb.Execute(ctx, fn)
}

// CallWithTimeout executes a client call with timeout
func (ccb *ClientCircuitBreaker) CallWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return ccb.Execute(ctx, func() error {
		done := make(chan error, 1)
		go func() {
			done <- fn()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
