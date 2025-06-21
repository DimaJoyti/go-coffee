package streaming

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewQualityAdapter(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()

	adapter := NewQualityAdapter(logger, config)

	assert.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
	assert.Empty(t, adapter.clients)
	assert.False(t, adapter.IsRunning())
}

func TestDefaultQualityAdapterConfig(t *testing.T) {
	config := DefaultQualityAdapterConfig()

	assert.True(t, config.EnableAdaptation)
	assert.Equal(t, 2*time.Second, config.MonitoringInterval)
	assert.Equal(t, 2, config.DefaultQualityLevel)
	assert.Equal(t, 0.7, config.AdaptationSensitivity)
	assert.Equal(t, 5*time.Second, config.MinStabilityPeriod)
	assert.Equal(t, 5, config.BandwidthSmoothingWindow)
	assert.Equal(t, 5, config.LatencySmoothingWindow)
	assert.True(t, config.EnablePredictiveScaling)
	assert.Len(t, config.QualityLevels, 5)
}

func TestQualityAdapter_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	ctx := context.Background()

	// Test start
	err := adapter.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, adapter.IsRunning())

	// Test start when already running
	err = adapter.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = adapter.Stop()
	assert.NoError(t, err)
	assert.False(t, adapter.IsRunning())

	// Test stop when not running
	err = adapter.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestQualityAdapter_RegisterUnregisterClient(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	clientID := "test-client-1"

	// Register client
	adapter.RegisterClient(clientID)

	// Check client exists
	profile := adapter.GetClientProfile(clientID)
	assert.NotNil(t, profile)
	assert.Equal(t, clientID, profile.ClientID)
	assert.Equal(t, config.DefaultQualityLevel, profile.CurrentQualityLevel)

	// Register same client again (should not create duplicate)
	adapter.RegisterClient(clientID)
	assert.Len(t, adapter.clients, 1)

	// Unregister client
	adapter.UnregisterClient(clientID)
	profile = adapter.GetClientProfile(clientID)
	assert.Nil(t, profile)
}

func TestQualityAdapter_GetOptimalQuality(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	// Test unknown client (should return default)
	quality := adapter.GetOptimalQuality("unknown-client")
	assert.NotNil(t, quality)
	assert.Equal(t, config.QualityLevels[config.DefaultQualityLevel], *quality)

	// Test registered client
	clientID := "test-client"
	adapter.RegisterClient(clientID)

	quality = adapter.GetOptimalQuality(clientID)
	assert.NotNil(t, quality)
	assert.Equal(t, config.QualityLevels[config.DefaultQualityLevel], *quality)
}

func TestQualityAdapter_UpdateNetworkMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	clientID := "test-client"
	adapter.RegisterClient(clientID)

	// Update network metrics
	metrics := &NetworkMetrics{
		Bandwidth:  5000000, // 5 Mbps
		Latency:    50 * time.Millisecond,
		PacketLoss: 1.0, // 1%
		Jitter:     10 * time.Millisecond,
	}

	adapter.UpdateNetworkMetrics(clientID, metrics)

	// Check updated profile
	profile := adapter.GetClientProfile(clientID)
	assert.NotNil(t, profile)
	assert.NotNil(t, profile.NetworkMetrics)
	assert.Equal(t, metrics.Bandwidth, profile.NetworkMetrics.Bandwidth)
	assert.Equal(t, metrics.Latency, profile.NetworkMetrics.Latency)
	assert.Equal(t, metrics.PacketLoss, profile.NetworkMetrics.PacketLoss)
	assert.Greater(t, profile.NetworkMetrics.ConnectionQuality, 0.0)
}

func TestQualityAdapter_CalculateConnectionQuality(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	tests := []struct {
		name     string
		metrics  *NetworkMetrics
		expected float64
	}{
		{
			name: "excellent connection",
			metrics: &NetworkMetrics{
				Bandwidth:  10000000, // 10 Mbps
				Latency:    20 * time.Millisecond,
				PacketLoss: 0.0,
				Jitter:     5 * time.Millisecond,
			},
			expected: 0.9, // Should be high quality
		},
		{
			name: "poor connection",
			metrics: &NetworkMetrics{
				Bandwidth:  500000, // 500 Kbps
				Latency:    300 * time.Millisecond,
				PacketLoss: 5.0, // 5%
				Jitter:     50 * time.Millisecond,
			},
			expected: 0.3, // Should be low quality
		},
		{
			name: "medium connection",
			metrics: &NetworkMetrics{
				Bandwidth:  2000000, // 2 Mbps
				Latency:    100 * time.Millisecond,
				PacketLoss: 2.0, // 2%
				Jitter:     20 * time.Millisecond,
			},
			expected: 0.6, // Should be medium quality
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := adapter.calculateConnectionQuality(tt.metrics)
			assert.InDelta(t, tt.expected, quality, 0.2, "Quality score should be within expected range")
			assert.GreaterOrEqual(t, quality, 0.0, "Quality should be >= 0")
			assert.LessOrEqual(t, quality, 1.0, "Quality should be <= 1")
		})
	}
}

func TestQualityAdapter_CalculateOptimalQualityLevel(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	tests := []struct {
		name            string
		metrics         *NetworkMetrics
		expectedMinLevel int
		expectedMaxLevel int
	}{
		{
			name: "high bandwidth",
			metrics: &NetworkMetrics{
				Bandwidth:  8000000, // 8 Mbps
				Latency:    30 * time.Millisecond,
				PacketLoss: 0.5,
			},
			expectedMinLevel: 3, // Should get high quality
			expectedMaxLevel: 4,
		},
		{
			name: "low bandwidth",
			metrics: &NetworkMetrics{
				Bandwidth:  200000, // 200 Kbps
				Latency:    100 * time.Millisecond,
				PacketLoss: 2.0,
			},
			expectedMinLevel: 0, // Should get low quality
			expectedMaxLevel: 1,
		},
		{
			name: "medium bandwidth",
			metrics: &NetworkMetrics{
				Bandwidth:  1500000, // 1.5 Mbps
				Latency:    80 * time.Millisecond,
				PacketLoss: 1.5,
			},
			expectedMinLevel: 1, // Should get medium quality
			expectedMaxLevel: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := adapter.calculateOptimalQualityLevel(tt.metrics)
			assert.GreaterOrEqual(t, level, tt.expectedMinLevel, "Quality level should be >= expected minimum")
			assert.LessOrEqual(t, level, tt.expectedMaxLevel, "Quality level should be <= expected maximum")
			assert.GreaterOrEqual(t, level, 0, "Quality level should be >= 0")
			assert.Less(t, level, len(config.QualityLevels), "Quality level should be < number of levels")
		})
	}
}

func TestQualityAdapter_GetSmoothedBandwidth(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	// Test with no history
	metrics := &NetworkMetrics{
		Bandwidth:        1000000,
		BandwidthHistory: []int64{},
	}
	smoothed := adapter.getSmoothedBandwidth(metrics)
	assert.Equal(t, metrics.Bandwidth, smoothed)

	// Test with history
	metrics.BandwidthHistory = []int64{500000, 750000, 1000000, 1250000, 1500000}
	smoothed = adapter.getSmoothedBandwidth(metrics)
	assert.Greater(t, smoothed, int64(500000), "Smoothed bandwidth should be > minimum")
	assert.Less(t, smoothed, int64(1500000), "Smoothed bandwidth should be < maximum")
}

func TestQualityAdapter_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	// Initial stats
	stats := adapter.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalAdaptations)
	assert.Equal(t, 0, stats.ClientsMonitored)
	assert.Equal(t, 0.0, stats.AverageQualityLevel)

	// Register clients
	adapter.RegisterClient("client1")
	adapter.RegisterClient("client2")

	stats = adapter.GetStats()
	assert.Equal(t, 2, stats.ClientsMonitored)
	assert.Equal(t, float64(config.DefaultQualityLevel), stats.AverageQualityLevel)
}

func TestQualityLevels(t *testing.T) {
	config := DefaultQualityAdapterConfig()

	// Verify quality levels are properly ordered
	for i := 1; i < len(config.QualityLevels); i++ {
		prev := config.QualityLevels[i-1]
		curr := config.QualityLevels[i]

		assert.Less(t, prev.Level, curr.Level, "Quality levels should be ordered")
		assert.LessOrEqual(t, prev.MaxFrameRate, curr.MaxFrameRate, "Frame rate should increase with quality")
		assert.LessOrEqual(t, prev.JpegQuality, curr.JpegQuality, "JPEG quality should increase with quality")
		assert.LessOrEqual(t, prev.MaxBitrate, curr.MaxBitrate, "Bitrate should increase with quality")
		assert.LessOrEqual(t, prev.MinBandwidth, curr.MinBandwidth, "Required bandwidth should increase with quality")
	}

	// Verify all levels have required fields
	for _, level := range config.QualityLevels {
		assert.NotEmpty(t, level.Name, "Quality level should have a name")
		assert.GreaterOrEqual(t, level.Level, 0, "Quality level should be >= 0")
		assert.Greater(t, level.MaxFrameRate, 0.0, "Max frame rate should be > 0")
		assert.Greater(t, level.JpegQuality, 0, "JPEG quality should be > 0")
		assert.LessOrEqual(t, level.JpegQuality, 100, "JPEG quality should be <= 100")
		assert.Greater(t, level.MaxBitrate, int64(0), "Max bitrate should be > 0")
		assert.Greater(t, level.MinBandwidth, int64(0), "Min bandwidth should be > 0")
		assert.Greater(t, level.MaxResolution.Width, 0, "Resolution width should be > 0")
		assert.Greater(t, level.MaxResolution.Height, 0, "Resolution height should be > 0")
	}
}

func TestQualityAdapter_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultQualityAdapterConfig()
	adapter := NewQualityAdapter(logger, config)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Test concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			clientID := fmt.Sprintf("client-%d", id)

			// Concurrent client operations
			adapter.RegisterClient(clientID)
			adapter.GetOptimalQuality(clientID)
			adapter.GetClientProfile(clientID)

			// Update metrics
			metrics := &NetworkMetrics{
				Bandwidth:  1000000 + int64(id*100000),
				Latency:    time.Duration(50+id*10) * time.Millisecond,
				PacketLoss: float64(id),
			}
			adapter.UpdateNetworkMetrics(clientID, metrics)

			adapter.UnregisterClient(clientID)
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

	// Check final state
	stats := adapter.GetStats()
	assert.Equal(t, 0, stats.ClientsMonitored, "All clients should be unregistered")
}
