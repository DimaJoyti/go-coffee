package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

func TestLoggingMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.LoggingMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	handler := func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.RecoveryMiddleware(handler)
	wrappedHandler(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.CORSMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called")
	}
	
	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS Allow-Origin header not set correctly")
	}
	
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS Allow-Methods header not set")
	}
	
	// Check for X-API-Key in allowed headers
	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowedHeaders == "" {
		t.Error("CORS Allow-Headers header not set")
	}
}

func TestCORSOptionsRequest(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.CORSMiddleware(handler)
	wrappedHandler(w, req)
	
	// Handler should not be called for OPTIONS request
	if called {
		t.Error("Handler should not be called for OPTIONS request")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	// Test health check bypass
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.AuthMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for health check")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Test public endpoint bypass
	called = false
	req = httptest.NewRequest(http.MethodPost, "/api/v1/tokens/price", nil)
	w = httptest.NewRecorder()
	
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for public endpoint")
	}
	
	// Test protected endpoint without auth
	called = false
	req = httptest.NewRequest(http.MethodPost, "/api/v1/bots", nil)
	w = httptest.NewRecorder()
	
	wrappedHandler(w, req)
	
	if called {
		t.Error("Handler should not be called for protected endpoint without auth")
	}
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
	
	// Test protected endpoint with API key
	called = false
	req = httptest.NewRequest(http.MethodPost, "/api/v1/bots", nil)
	req.Header.Set("X-API-Key", "test-key")
	w = httptest.NewRecorder()
	
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for protected endpoint with API key")
	}
	
	// Test protected endpoint with Authorization header
	called = false
	req = httptest.NewRequest(http.MethodPost, "/api/v1/bots", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w = httptest.NewRecorder()
	
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for protected endpoint with Authorization header")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.SecurityHeadersMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called")
	}
	
	// Check security headers
	expectedHeaders := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
	}
	
	for header, expectedValue := range expectedHeaders {
		if w.Header().Get(header) != expectedValue {
			t.Errorf("Expected %s header to be %s, got %s", header, expectedValue, w.Header().Get(header))
		}
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	// Test health check bypass
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.RateLimitMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for health check")
	}
	
	// Test regular endpoint
	called = false
	req = httptest.NewRequest(http.MethodGet, "/api/v1/pools", nil)
	w = httptest.NewRecorder()
	
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called for regular endpoint")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler := middleware.MetricsMiddleware(handler)
	wrappedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChainMiddleware(t *testing.T) {
	logger := logger.New("test")
	middleware := NewMiddleware(logger)
	
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	// Chain multiple middleware
	chainedHandler := middleware.Chain(
		handler,
		middleware.LoggingMiddleware,
		middleware.RecoveryMiddleware,
		middleware.CORSMiddleware,
		middleware.SecurityHeadersMiddleware,
	)
	
	chainedHandler(w, req)
	
	if !called {
		t.Error("Handler was not called")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Check that CORS and security headers are set
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS headers not set in chain")
	}
	
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Security headers not set in chain")
	}
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	
	wrapper.WriteHeader(http.StatusCreated)
	
	if wrapper.statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, wrapper.statusCode)
	}
	
	if w.Code != http.StatusCreated {
		t.Errorf("Expected response writer status %d, got %d", http.StatusCreated, w.Code)
	}
}
