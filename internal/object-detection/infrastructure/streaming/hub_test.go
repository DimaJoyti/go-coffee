package streaming

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHub(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()

	hub := NewHub(logger, config)

	assert.NotNil(t, hub)
	assert.Equal(t, config, hub.config)
	assert.Empty(t, hub.clients)
	assert.False(t, hub.IsRunning())
}

func TestDefaultHubConfig(t *testing.T) {
	config := DefaultHubConfig()

	assert.Equal(t, 1024, config.ReadBufferSize)
	assert.Equal(t, 1024, config.WriteBufferSize)
	assert.Equal(t, 10*time.Second, config.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.PongTimeout)
	assert.Equal(t, 54*time.Second, config.PingPeriod)
	assert.Equal(t, int64(512*1024), config.MaxMessageSize)
	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Equal(t, 1000, config.MaxConnections)
	assert.True(t, config.EnableCompression)
}

func TestHub_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	ctx := context.Background()

	// Test start
	err := hub.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, hub.IsRunning())

	// Test start when already running
	err = hub.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = hub.Stop()
	assert.NoError(t, err)
	assert.False(t, hub.IsRunning())

	// Test stop when not running
	err = hub.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestHub_HandleWebSocket_NotRunning(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	// Make request to non-running hub
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHub_Broadcast_NotRunning(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	message := &Message{
		Type:      MessageTypeStatus,
		Timestamp: time.Now(),
		Data:      "test",
	}

	err := hub.Broadcast(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestHub_BroadcastToStream_NotRunning(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	message := &Message{
		Type:      MessageTypeStatus,
		StreamID:  "test-stream",
		Timestamp: time.Now(),
		Data:      "test",
	}

	err := hub.BroadcastToStream("test-stream", message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestHub_GetConnectedClients(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	// Initially no clients
	assert.Equal(t, 0, hub.GetConnectedClients())

	ctx := context.Background()
	err := hub.Start(ctx)
	require.NoError(t, err)
	defer hub.Stop()

	// Still no clients after start
	assert.Equal(t, 0, hub.GetConnectedClients())
}

func TestHub_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	stats := hub.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.ConnectedClients)
	assert.Equal(t, int64(0), stats.TotalConnections)
	assert.Equal(t, int64(0), stats.MessagesSent)
	assert.Equal(t, int64(0), stats.MessagesReceived)
}

func TestHub_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	newConfig := HubConfig{
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
		WriteTimeout:      20 * time.Second,
		PongTimeout:       120 * time.Second,
		PingPeriod:        108 * time.Second,
		MaxMessageSize:    1024 * 1024,
		AllowedOrigins:    []string{"localhost"},
		MaxConnections:    500,
		EnableCompression: false,
	}

	hub.UpdateConfig(newConfig)
	assert.Equal(t, newConfig, hub.GetConfig())
}

func TestMessage_Types(t *testing.T) {
	assert.Equal(t, "detection", string(MessageTypeDetection))
	assert.Equal(t, "tracking", string(MessageTypeTracking))
	assert.Equal(t, "frame", string(MessageTypeFrame))
	assert.Equal(t, "status", string(MessageTypeStatus))
	assert.Equal(t, "error", string(MessageTypeError))
	assert.Equal(t, "subscribe", string(MessageTypeSubscribe))
	assert.Equal(t, "unsubscribe", string(MessageTypeUnsubscribe))
	assert.Equal(t, "ping", string(MessageTypePing))
	assert.Equal(t, "pong", string(MessageTypePong))
}

// Integration test with actual WebSocket connection
func TestHub_WebSocketIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	config := DefaultHubConfig()
	config.MaxConnections = 2 // Limit for testing
	hub := NewHub(logger, config)

	ctx := context.Background()
	err := hub.Start(ctx)
	require.NoError(t, err)
	defer hub.Stop()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Wait a bit for connection to be registered
	time.Sleep(100 * time.Millisecond)

	// Check connected clients
	assert.Equal(t, 1, hub.GetConnectedClients())

	// Test broadcast
	message := &Message{
		Type:      MessageTypeStatus,
		Timestamp: time.Now(),
		Data:      "test broadcast",
	}

	err = hub.Broadcast(message)
	assert.NoError(t, err)

	// Read message from WebSocket
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(receivedData), "test broadcast")
}

func TestHub_MaxConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	config := DefaultHubConfig()
	config.MaxConnections = 1 // Very low limit for testing
	hub := NewHub(logger, config)

	ctx := context.Background()
	err := hub.Start(ctx)
	require.NoError(t, err)
	defer hub.Stop()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// First connection should succeed
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn1.Close()

	// Wait for connection to be registered
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, hub.GetConnectedClients())

	// Second connection should fail due to max connections
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestHub_ConcurrentOperations(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	hub := NewHub(logger, config)

	ctx := context.Background()
	err := hub.Start(ctx)
	require.NoError(t, err)
	defer hub.Stop()

	// Test concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Concurrent reads
			hub.GetConnectedClients()
			hub.GetStats()
			hub.IsRunning()

			// Concurrent broadcasts
			message := &Message{
				Type:      MessageTypeStatus,
				Timestamp: time.Now(),
				Data:      fmt.Sprintf("test %d", id),
			}
			hub.Broadcast(message)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

func TestHub_OriginCheck(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultHubConfig()
	config.AllowedOrigins = []string{"https://example.com"}
	hub := NewHub(logger, config)

	// Test origin check in upgrader
	upgrader := hub.upgrader
	
	// Create request with allowed origin
	req1 := httptest.NewRequest("GET", "/ws", nil)
	req1.Header.Set("Origin", "https://example.com")
	assert.True(t, upgrader.CheckOrigin(req1))

	// Create request with disallowed origin
	req2 := httptest.NewRequest("GET", "/ws", nil)
	req2.Header.Set("Origin", "https://malicious.com")
	assert.False(t, upgrader.CheckOrigin(req2))

	// Test wildcard origin
	config.AllowedOrigins = []string{"*"}
	hub = NewHub(logger, config)
	upgrader = hub.upgrader
	
	req3 := httptest.NewRequest("GET", "/ws", nil)
	req3.Header.Set("Origin", "https://any-origin.com")
	assert.True(t, upgrader.CheckOrigin(req3))
}
