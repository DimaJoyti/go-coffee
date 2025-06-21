package streaming

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NetworkMonitor monitors network conditions and performance
type NetworkMonitor struct {
	logger    *zap.Logger
	config    NetworkMonitorConfig
	isRunning bool
	mutex     sync.RWMutex
	stats     *NetworkMonitorStats
}

// NetworkMonitorConfig configures the network monitor
type NetworkMonitorConfig struct {
	MonitoringInterval    time.Duration // How often to check network conditions
	PingTargets          []string      // Targets to ping for latency measurement
	BandwidthTestEnabled bool          // Enable bandwidth testing
	BandwidthTestInterval time.Duration // How often to test bandwidth
	BandwidthTestSize    int64         // Size of bandwidth test data
	TimeoutDuration      time.Duration // Timeout for network operations
	MaxRetries           int           // Maximum retries for failed operations
}

// NetworkMonitorStats tracks monitoring performance
type NetworkMonitorStats struct {
	TotalPings        int64
	SuccessfulPings   int64
	FailedPings       int64
	AverageLatency    time.Duration
	BandwidthTests    int64
	LastBandwidthTest time.Time
	StartTime         time.Time
	LastUpdate        time.Time
	mutex             sync.RWMutex
}

// PingResult represents the result of a ping operation
type PingResult struct {
	Target    string
	Latency   time.Duration
	Success   bool
	Error     error
	Timestamp time.Time
}

// BandwidthResult represents the result of a bandwidth test
type BandwidthResult struct {
	UploadSpeed   int64 // bits per second
	DownloadSpeed int64 // bits per second
	TestDuration  time.Duration
	Success       bool
	Error         error
	Timestamp     time.Time
}

// DefaultNetworkMonitorConfig returns default configuration
func DefaultNetworkMonitorConfig() NetworkMonitorConfig {
	return NetworkMonitorConfig{
		MonitoringInterval:    5 * time.Second,
		PingTargets:          []string{"8.8.8.8", "1.1.1.1", "208.67.222.222"}, // Google, Cloudflare, OpenDNS
		BandwidthTestEnabled:  false, // Disabled by default to avoid unnecessary traffic
		BandwidthTestInterval: 60 * time.Second,
		BandwidthTestSize:     1024 * 1024, // 1MB
		TimeoutDuration:       5 * time.Second,
		MaxRetries:            3,
	}
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(logger *zap.Logger) *NetworkMonitor {
	config := DefaultNetworkMonitorConfig()
	
	return &NetworkMonitor{
		logger: logger.With(zap.String("component", "network_monitor")),
		config: config,
		stats: &NetworkMonitorStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the network monitor
func (nm *NetworkMonitor) Start(ctx context.Context) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.isRunning {
		return fmt.Errorf("network monitor is already running")
	}

	nm.logger.Info("Starting network monitor",
		zap.Duration("monitoring_interval", nm.config.MonitoringInterval),
		zap.Strings("ping_targets", nm.config.PingTargets),
		zap.Bool("bandwidth_test_enabled", nm.config.BandwidthTestEnabled))

	nm.isRunning = true

	// Start monitoring goroutine
	go nm.monitoringLoop(ctx)

	nm.logger.Info("Network monitor started")
	return nil
}

// Stop stops the network monitor
func (nm *NetworkMonitor) Stop() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if !nm.isRunning {
		return fmt.Errorf("network monitor is not running")
	}

	nm.logger.Info("Stopping network monitor")
	nm.isRunning = false

	nm.logger.Info("Network monitor stopped")
	return nil
}

// monitoringLoop continuously monitors network conditions
func (nm *NetworkMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(nm.config.MonitoringInterval)
	defer ticker.Stop()

	bandwidthTicker := time.NewTicker(nm.config.BandwidthTestInterval)
	defer bandwidthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nm.performLatencyCheck()
		case <-bandwidthTicker.C:
			if nm.config.BandwidthTestEnabled {
				nm.performBandwidthTest()
			}
		}
	}
}

// performLatencyCheck performs latency checks to configured targets
func (nm *NetworkMonitor) performLatencyCheck() {
	for _, target := range nm.config.PingTargets {
		go func(target string) {
			result := nm.pingTarget(target)
			nm.processPingResult(result)
		}(target)
	}
}

// pingTarget performs a ping to a specific target
func (nm *NetworkMonitor) pingTarget(target string) *PingResult {
	start := time.Now()
	
	// Use TCP connection as a simple "ping" alternative
	// This is more reliable than ICMP ping which may be blocked
	conn, err := net.DialTimeout("tcp", target+":80", nm.config.TimeoutDuration)
	latency := time.Since(start)
	
	result := &PingResult{
		Target:    target,
		Latency:   latency,
		Success:   err == nil,
		Error:     err,
		Timestamp: time.Now(),
	}

	if conn != nil {
		conn.Close()
	}

	return result
}

// processPingResult processes a ping result and updates statistics
func (nm *NetworkMonitor) processPingResult(result *PingResult) {
	nm.stats.mutex.Lock()
	defer nm.stats.mutex.Unlock()

	nm.stats.TotalPings++
	nm.stats.LastUpdate = time.Now()

	if result.Success {
		nm.stats.SuccessfulPings++
		
		// Update average latency
		if nm.stats.SuccessfulPings == 1 {
			nm.stats.AverageLatency = result.Latency
		} else {
			// Exponential moving average
			alpha := 0.1
			nm.stats.AverageLatency = time.Duration(
				alpha*float64(result.Latency) + (1-alpha)*float64(nm.stats.AverageLatency),
			)
		}
	} else {
		nm.stats.FailedPings++
		nm.logger.Debug("Ping failed",
			zap.String("target", result.Target),
			zap.Error(result.Error))
	}
}

// performBandwidthTest performs a bandwidth test
func (nm *NetworkMonitor) performBandwidthTest() {
	// This is a simplified bandwidth test
	// In a real implementation, you might use a dedicated bandwidth testing service
	
	nm.logger.Debug("Performing bandwidth test")
	
	start := time.Now()
	
	// Simulate bandwidth test by measuring connection establishment time
	// This is not a real bandwidth test but provides some network performance indication
	var totalLatency time.Duration
	successfulConnections := 0
	
	for _, target := range nm.config.PingTargets {
		conn, err := net.DialTimeout("tcp", target+":80", nm.config.TimeoutDuration)
		if err == nil && conn != nil {
			successfulConnections++
			totalLatency += time.Since(start)
			conn.Close()
		}
	}
	
	testDuration := time.Since(start)
	
	// Estimate bandwidth based on connection performance
	// This is a very rough estimation
	estimatedBandwidth := int64(1000000) // Default 1 Mbps
	if successfulConnections > 0 {
		avgLatency := totalLatency / time.Duration(successfulConnections)
		// Lower latency suggests better connection, estimate higher bandwidth
		if avgLatency < 50*time.Millisecond {
			estimatedBandwidth = 10000000 // 10 Mbps
		} else if avgLatency < 100*time.Millisecond {
			estimatedBandwidth = 5000000 // 5 Mbps
		} else if avgLatency < 200*time.Millisecond {
			estimatedBandwidth = 2000000 // 2 Mbps
		}
	}
	
	result := &BandwidthResult{
		UploadSpeed:   estimatedBandwidth,
		DownloadSpeed: estimatedBandwidth,
		TestDuration:  testDuration,
		Success:       successfulConnections > 0,
		Timestamp:     time.Now(),
	}
	
	nm.processBandwidthResult(result)
}

// processBandwidthResult processes a bandwidth test result
func (nm *NetworkMonitor) processBandwidthResult(result *BandwidthResult) {
	nm.stats.mutex.Lock()
	defer nm.stats.mutex.Unlock()

	nm.stats.BandwidthTests++
	nm.stats.LastBandwidthTest = result.Timestamp

	if result.Success {
		nm.logger.Debug("Bandwidth test completed",
			zap.Int64("upload_speed", result.UploadSpeed),
			zap.Int64("download_speed", result.DownloadSpeed),
			zap.Duration("test_duration", result.TestDuration))
	} else {
		nm.logger.Debug("Bandwidth test failed", zap.Error(result.Error))
	}
}

// GetCurrentNetworkMetrics returns current network metrics
func (nm *NetworkMonitor) GetCurrentNetworkMetrics() *NetworkMetrics {
	nm.stats.mutex.RLock()
	defer nm.stats.mutex.RUnlock()

	// Calculate packet loss percentage
	packetLoss := 0.0
	if nm.stats.TotalPings > 0 {
		packetLoss = float64(nm.stats.FailedPings) / float64(nm.stats.TotalPings) * 100
	}

	// Estimate bandwidth (this would be more sophisticated in a real implementation)
	estimatedBandwidth := int64(1000000) // Default 1 Mbps
	if nm.stats.AverageLatency < 50*time.Millisecond {
		estimatedBandwidth = 10000000 // 10 Mbps
	} else if nm.stats.AverageLatency < 100*time.Millisecond {
		estimatedBandwidth = 5000000 // 5 Mbps
	} else if nm.stats.AverageLatency < 200*time.Millisecond {
		estimatedBandwidth = 2000000 // 2 Mbps
	}

	// Calculate jitter (simplified as a percentage of average latency)
	jitter := time.Duration(float64(nm.stats.AverageLatency) * 0.1)

	metrics := &NetworkMetrics{
		Bandwidth:         estimatedBandwidth,
		Latency:           nm.stats.AverageLatency,
		PacketLoss:        packetLoss,
		Jitter:            jitter,
		ConnectionQuality: nm.calculateConnectionQuality(estimatedBandwidth, nm.stats.AverageLatency, packetLoss),
		LastUpdated:       time.Now(),
		BandwidthHistory:  make([]int64, 0),
		LatencyHistory:    make([]time.Duration, 0),
	}

	return metrics
}

// calculateConnectionQuality calculates overall connection quality
func (nm *NetworkMonitor) calculateConnectionQuality(bandwidth int64, latency time.Duration, packetLoss float64) float64 {
	// Bandwidth score (0-1)
	bandwidthScore := float64(bandwidth) / 10000000.0 // 10 Mbps = 1.0
	if bandwidthScore > 1.0 {
		bandwidthScore = 1.0
	}

	// Latency score (0-1, lower is better)
	latencyMs := float64(latency.Milliseconds())
	latencyScore := 1.0 - (latencyMs / 500.0) // 500ms = 0.0
	if latencyScore < 0 {
		latencyScore = 0
	}

	// Packet loss score (0-1, lower is better)
	packetLossScore := 1.0 - (packetLoss / 10.0) // 10% = 0.0
	if packetLossScore < 0 {
		packetLossScore = 0
	}

	// Weighted average
	quality := (bandwidthScore*0.5 + latencyScore*0.3 + packetLossScore*0.2)
	
	if quality > 1.0 {
		quality = 1.0
	}
	if quality < 0 {
		quality = 0
	}

	return quality
}

// GetStats returns network monitor statistics
func (nm *NetworkMonitor) GetStats() *NetworkMonitorStats {
	nm.stats.mutex.RLock()
	defer nm.stats.mutex.RUnlock()

	return &NetworkMonitorStats{
		TotalPings:        nm.stats.TotalPings,
		SuccessfulPings:   nm.stats.SuccessfulPings,
		FailedPings:       nm.stats.FailedPings,
		AverageLatency:    nm.stats.AverageLatency,
		BandwidthTests:    nm.stats.BandwidthTests,
		LastBandwidthTest: nm.stats.LastBandwidthTest,
		StartTime:         nm.stats.StartTime,
		LastUpdate:        nm.stats.LastUpdate,
	}
}

// IsRunning returns whether the monitor is running
func (nm *NetworkMonitor) IsRunning() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.isRunning
}

// UpdateConfig updates the monitor configuration
func (nm *NetworkMonitor) UpdateConfig(config NetworkMonitorConfig) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.config = config
	nm.logger.Info("Network monitor configuration updated",
		zap.Duration("monitoring_interval", config.MonitoringInterval),
		zap.Bool("bandwidth_test_enabled", config.BandwidthTestEnabled))
}

// GetConfig returns the current configuration
func (nm *NetworkMonitor) GetConfig() NetworkMonitorConfig {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.config
}
