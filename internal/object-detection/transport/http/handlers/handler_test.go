package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestHandler() *Handler {
	logger := zap.NewNop() // No-op logger for tests
	cfg := &config.Config{
		Environment: "test",
		Server: config.ServerConfig{
			Port: 8080,
		},
		Monitoring: config.MonitoringConfig{
			Enabled:     true,
			MetricsPath: "/metrics",
		},
		WebSocket: config.WebSocketConfig{
			Enabled: true,
		},
	}
	return NewHandler(logger, cfg)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := setupTestHandler()
	handler.SetupRoutes(router)
	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "object-detection-service", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotNil(t, response["timestamp"])
}

func TestReadinessCheck(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/ready", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ready", response["status"])
	assert.NotNil(t, response["checks"])
	
	checks := response["checks"].(map[string]interface{})
	assert.Equal(t, "ok", checks["database"])
	assert.Equal(t, "ok", checks["redis"])
	assert.Equal(t, "ok", checks["model"])
}

func TestMetricsEndpoint(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestGetStreams_Empty(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/streams", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Streams retrieved successfully", response.Message)
	
	// Data should be an empty array
	data := response.Data.([]interface{})
	assert.Empty(t, data)
}

func TestGetStream_NotFound(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/streams/nonexistent", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "stream_not_found", response.Error)
	assert.Equal(t, "Stream not found", response.Message)
}

func TestGetModels_Empty(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/models", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Models retrieved successfully", response.Message)
	
	// Data should be an empty array
	data := response.Data.([]interface{})
	assert.Empty(t, data)
}

func TestGetDetectionResults_WithPagination(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/detection/results?limit=10&offset=0", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Detection results retrieved successfully", response.Message)
	
	data := response.Data.(map[string]interface{})
	assert.NotNil(t, data["results"])
	assert.NotNil(t, data["pagination"])
	
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(0), pagination["offset"])
	assert.Equal(t, float64(0), pagination["total"])
}

func TestGetDetectionResults_InvalidPagination(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/detection/results?limit=invalid", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_request", response.Error)
	assert.Equal(t, "Invalid limit parameter", response.Message)
}

func TestGetDetectionStats_MissingStreamID(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/detection/stats", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_request", response.Error)
	assert.Equal(t, "Stream ID is required", response.Message)
}

func TestGetDetectionStats_ValidStreamID(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/detection/stats?stream_id=test-stream", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Detection stats retrieved successfully", response.Message)
	assert.NotNil(t, response.Data)
}

func TestGetAlerts_WithFilters(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/alerts?stream_id=test&acknowledged=false&limit=20", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Alerts retrieved successfully", response.Message)
	
	data := response.Data.(map[string]interface{})
	assert.NotNil(t, data["alerts"])
	assert.NotNil(t, data["pagination"])
	assert.NotNil(t, data["filters"])
	
	filters := data["filters"].(map[string]interface{})
	assert.Equal(t, "test", filters["stream_id"])
	assert.Equal(t, false, filters["acknowledged"])
}

func TestGetActiveTracking_ValidStreamID(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/tracking/active/test-stream", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Active tracking retrieved successfully", response.Message)
	
	// Data should be an empty array (placeholder implementation)
	data := response.Data.([]interface{})
	assert.Empty(t, data)
}

func TestGetTrackingHistory_NotFound(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/api/v1/tracking/history/nonexistent", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "tracking_not_found", response.Error)
	assert.Equal(t, "Tracking data not found", response.Message)
}
