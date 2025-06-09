package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Simple test for HTTP layer without application dependencies
func TestMethodHandler(t *testing.T) {
	logger := logger.New("test")
	handler := &Handler{logger: logger}

	called := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	// Test allowed method
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler := handler.methodHandler("POST", testHandler)
	wrappedHandler(w, req)

	if !called {
		t.Error("Handler was not called for allowed method")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test disallowed method
	called = false
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()

	wrappedHandler(w, req)

	if called {
		t.Error("Handler should not be called for disallowed method")
	}

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestWriteErrorResponse(t *testing.T) {
	logger := logger.New("test")
	handler := &Handler{logger: logger}

	w := httptest.NewRecorder()
	handler.writeErrorResponse(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	logger := logger.New("test")
	handler := &Handler{logger: logger}

	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}
	handler.writeSuccessResponse(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
}
