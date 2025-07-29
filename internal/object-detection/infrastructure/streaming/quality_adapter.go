package streaming

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QualityAdapter manages adaptive streaming quality based on network conditions
type QualityAdapter struct {
	logger         *zap.Logger
	clients        map[string]*ClientQualityProfile
	config         QualityAdapterConfig
	networkMonitor *NetworkMonitor
	isRunning      bool
	mutex          sync.RWMutex
	stats          *QualityAdapterStats
}

// QualityAdapterConfig configures the quality adapter
type QualityAdapterConfig struct {
	EnableAdaptation         bool           // Enable adaptive quality
	MonitoringInterval       time.Duration  // How often to check network conditions
	QualityLevels            []QualityLevel // Available quality levels
	DefaultQualityLevel      int            // Default quality level index
	AdaptationSensitivity    float64        // How quickly to adapt (0.1-1.0)
	MinStabilityPeriod       time.Duration  // Minimum time before quality changes
	BandwidthSmoothingWindow int            // Window size for bandwidth smoothing
	LatencySmoothingWindow   int            // Window size for latency smoothing
	EnablePredictiveScaling  bool           // Enable predictive quality scaling
}

// QualityLevel defines a streaming quality configuration
type QualityLevel struct {
	Name          string     `json:"name"`
	Level         int        `json:"level"`
	MaxFrameRate  float64    `json:"max_frame_rate"`
	JpegQuality   int        `json:"jpeg_quality"`
	MaxResolution Resolution `json:"max_resolution"`
	MaxBitrate    int64      `json:"max_bitrate"`   // bits per second
	MinBandwidth  int64      `json:"min_bandwidth"` // required bandwidth
	Description   string     `json:"description"`
}

// Resolution represents video resolution
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ClientQualityProfile tracks quality metrics for a specific client
type ClientQualityProfile struct {
	ClientID            string
	CurrentQualityLevel int
	TargetQualityLevel  int
	NetworkMetrics      *NetworkMetrics
	QualityHistory      []QualityChange
	LastAdaptation      time.Time
	AdaptationCount     int64
	StabilityScore      float64
	mutex               sync.RWMutex
}

// NetworkMetrics tracks network performance for a client
type NetworkMetrics struct {
	Bandwidth         int64         // Current bandwidth (bits/sec)
	Latency           time.Duration // Round-trip latency
	PacketLoss        float64       // Packet loss percentage (0-100)
	Jitter            time.Duration // Network jitter
	ConnectionQuality float64       // Overall quality score (0-1)
	LastUpdated       time.Time
	BandwidthHistory  []int64         // Historical bandwidth measurements
	LatencyHistory    []time.Duration // Historical latency measurements
	mutex             sync.RWMutex
}

// QualityChange represents a quality level change event
type QualityChange struct {
	Timestamp    time.Time
	FromLevel    int
	ToLevel      int
	Reason       string
	NetworkState *NetworkMetrics
}

// QualityAdapterStats tracks adapter performance
type QualityAdapterStats struct {
	TotalAdaptations    int64
	QualityUpgrades     int64
	QualityDowngrades   int64
	ClientsMonitored    int
	AverageQualityLevel float64
	AdaptationRate      float64 // Adaptations per minute
	StartTime           time.Time
	LastAdaptation      time.Time
	mutex               sync.RWMutex
}

// DefaultQualityAdapterConfig returns default configuration
func DefaultQualityAdapterConfig() QualityAdapterConfig {
	return QualityAdapterConfig{
		EnableAdaptation:         true,
		MonitoringInterval:       2 * time.Second,
		DefaultQualityLevel:      2, // Medium quality
		AdaptationSensitivity:    0.7,
		MinStabilityPeriod:       5 * time.Second,
		BandwidthSmoothingWindow: 5,
		LatencySmoothingWindow:   5,
		EnablePredictiveScaling:  true,
		QualityLevels: []QualityLevel{
			{
				Name:          "Low",
				Level:         0,
				MaxFrameRate:  5.0,
				JpegQuality:   30,
				MaxResolution: Resolution{Width: 320, Height: 240},
				MaxBitrate:    100000, // 100 Kbps
				MinBandwidth:  150000, // 150 Kbps required
				Description:   "Low quality for poor connections",
			},
			{
				Name:          "Medium-Low",
				Level:         1,
				MaxFrameRate:  10.0,
				JpegQuality:   50,
				MaxResolution: Resolution{Width: 480, Height: 360},
				MaxBitrate:    300000, // 300 Kbps
				MinBandwidth:  450000, // 450 Kbps required
				Description:   "Medium-low quality for limited bandwidth",
			},
			{
				Name:          "Medium",
				Level:         2,
				MaxFrameRate:  15.0,
				JpegQuality:   70,
				MaxResolution: Resolution{Width: 640, Height: 480},
				MaxBitrate:    800000,  // 800 Kbps
				MinBandwidth:  1200000, // 1.2 Mbps required
				Description:   "Medium quality for standard connections",
			},
			{
				Name:          "High",
				Level:         3,
				MaxFrameRate:  25.0,
				JpegQuality:   85,
				MaxResolution: Resolution{Width: 1280, Height: 720},
				MaxBitrate:    2000000, // 2 Mbps
				MinBandwidth:  3000000, // 3 Mbps required
				Description:   "High quality for good connections",
			},
			{
				Name:          "Ultra",
				Level:         4,
				MaxFrameRate:  30.0,
				JpegQuality:   95,
				MaxResolution: Resolution{Width: 1920, Height: 1080},
				MaxBitrate:    5000000, // 5 Mbps
				MinBandwidth:  7500000, // 7.5 Mbps required
				Description:   "Ultra quality for excellent connections",
			},
		},
	}
}

// NewQualityAdapter creates a new quality adapter
func NewQualityAdapter(logger *zap.Logger, config QualityAdapterConfig) *QualityAdapter {
	return &QualityAdapter{
		logger:         logger.With(zap.String("component", "quality_adapter")),
		clients:        make(map[string]*ClientQualityProfile),
		config:         config,
		networkMonitor: NewNetworkMonitor(logger),
		stats: &QualityAdapterStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the quality adapter
func (qa *QualityAdapter) Start(ctx context.Context) error {
	qa.mutex.Lock()
	defer qa.mutex.Unlock()

	if qa.isRunning {
		return fmt.Errorf("quality adapter is already running")
	}

	qa.logger.Info("Starting quality adapter",
		zap.Bool("adaptation_enabled", qa.config.EnableAdaptation),
		zap.Duration("monitoring_interval", qa.config.MonitoringInterval),
		zap.Int("quality_levels", len(qa.config.QualityLevels)))

	qa.isRunning = true

	// Start network monitor
	if err := qa.networkMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start network monitor: %w", err)
	}

	// Start monitoring goroutine
	go qa.monitoringLoop(ctx)

	qa.logger.Info("Quality adapter started")
	return nil
}

// Stop stops the quality adapter
func (qa *QualityAdapter) Stop() error {
	qa.mutex.Lock()
	defer qa.mutex.Unlock()

	if !qa.isRunning {
		return fmt.Errorf("quality adapter is not running")
	}

	qa.logger.Info("Stopping quality adapter")

	qa.isRunning = false

	// Stop network monitor
	if err := qa.networkMonitor.Stop(); err != nil {
		qa.logger.Error("Failed to stop network monitor", zap.Error(err))
	}

	qa.logger.Info("Quality adapter stopped")
	return nil
}

// RegisterClient registers a new client for quality monitoring
func (qa *QualityAdapter) RegisterClient(clientID string) {
	qa.mutex.Lock()
	defer qa.mutex.Unlock()

	if _, exists := qa.clients[clientID]; exists {
		return
	}

	profile := &ClientQualityProfile{
		ClientID:            clientID,
		CurrentQualityLevel: qa.config.DefaultQualityLevel,
		TargetQualityLevel:  qa.config.DefaultQualityLevel,
		NetworkMetrics: &NetworkMetrics{
			ConnectionQuality: 0.5, // Start with medium quality assumption
			LastUpdated:       time.Now(),
			BandwidthHistory:  make([]int64, 0, qa.config.BandwidthSmoothingWindow),
			LatencyHistory:    make([]time.Duration, 0, qa.config.LatencySmoothingWindow),
		},
		QualityHistory: make([]QualityChange, 0),
		LastAdaptation: time.Now(),
		StabilityScore: 1.0,
	}

	qa.clients[clientID] = profile

	qa.logger.Info("Client registered for quality adaptation",
		zap.String("client_id", clientID),
		zap.Int("initial_quality", qa.config.DefaultQualityLevel))
}

// UnregisterClient removes a client from quality monitoring
func (qa *QualityAdapter) UnregisterClient(clientID string) {
	qa.mutex.Lock()
	defer qa.mutex.Unlock()

	delete(qa.clients, clientID)

	qa.logger.Info("Client unregistered from quality adaptation",
		zap.String("client_id", clientID))
}

// UpdateNetworkMetrics updates network metrics for a client
func (qa *QualityAdapter) UpdateNetworkMetrics(clientID string, metrics *NetworkMetrics) {
	qa.mutex.RLock()
	profile, exists := qa.clients[clientID]
	qa.mutex.RUnlock()

	if !exists {
		return
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	// Update current metrics
	profile.NetworkMetrics.Bandwidth = metrics.Bandwidth
	profile.NetworkMetrics.Latency = metrics.Latency
	profile.NetworkMetrics.PacketLoss = metrics.PacketLoss
	profile.NetworkMetrics.Jitter = metrics.Jitter
	profile.NetworkMetrics.LastUpdated = time.Now()

	// Add to history for smoothing
	profile.NetworkMetrics.BandwidthHistory = append(profile.NetworkMetrics.BandwidthHistory, metrics.Bandwidth)
	if len(profile.NetworkMetrics.BandwidthHistory) > qa.config.BandwidthSmoothingWindow {
		profile.NetworkMetrics.BandwidthHistory = profile.NetworkMetrics.BandwidthHistory[1:]
	}

	profile.NetworkMetrics.LatencyHistory = append(profile.NetworkMetrics.LatencyHistory, metrics.Latency)
	if len(profile.NetworkMetrics.LatencyHistory) > qa.config.LatencySmoothingWindow {
		profile.NetworkMetrics.LatencyHistory = profile.NetworkMetrics.LatencyHistory[1:]
	}

	// Calculate connection quality score
	profile.NetworkMetrics.ConnectionQuality = qa.calculateConnectionQuality(profile.NetworkMetrics)

	qa.logger.Debug("Network metrics updated",
		zap.String("client_id", clientID),
		zap.Int64("bandwidth", metrics.Bandwidth),
		zap.Duration("latency", metrics.Latency),
		zap.Float64("quality_score", profile.NetworkMetrics.ConnectionQuality))
}

// GetOptimalQuality returns the optimal quality level for a client
func (qa *QualityAdapter) GetOptimalQuality(clientID string) *QualityLevel {
	qa.mutex.RLock()
	profile, exists := qa.clients[clientID]
	qa.mutex.RUnlock()

	if !exists {
		// Return default quality for unknown clients
		return &qa.config.QualityLevels[qa.config.DefaultQualityLevel]
	}

	profile.mutex.RLock()
	qualityLevel := profile.CurrentQualityLevel
	profile.mutex.RUnlock()

	if qualityLevel < 0 || qualityLevel >= len(qa.config.QualityLevels) {
		qualityLevel = qa.config.DefaultQualityLevel
	}

	return &qa.config.QualityLevels[qualityLevel]
}

// GetClientProfile returns the quality profile for a client
func (qa *QualityAdapter) GetClientProfile(clientID string) *ClientQualityProfile {
	qa.mutex.RLock()
	defer qa.mutex.RUnlock()

	profile, exists := qa.clients[clientID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	profile.mutex.RLock()
	defer profile.mutex.RUnlock()

	profileCopy := &ClientQualityProfile{
		ClientID:            profile.ClientID,
		CurrentQualityLevel: profile.CurrentQualityLevel,
		TargetQualityLevel:  profile.TargetQualityLevel,
		LastAdaptation:      profile.LastAdaptation,
		AdaptationCount:     profile.AdaptationCount,
		StabilityScore:      profile.StabilityScore,
	}

	// Copy network metrics
	if profile.NetworkMetrics != nil {
		profileCopy.NetworkMetrics = &NetworkMetrics{
			Bandwidth:         profile.NetworkMetrics.Bandwidth,
			Latency:           profile.NetworkMetrics.Latency,
			PacketLoss:        profile.NetworkMetrics.PacketLoss,
			Jitter:            profile.NetworkMetrics.Jitter,
			ConnectionQuality: profile.NetworkMetrics.ConnectionQuality,
			LastUpdated:       profile.NetworkMetrics.LastUpdated,
		}
	}

	return profileCopy
}

// monitoringLoop continuously monitors and adapts quality for all clients
func (qa *QualityAdapter) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(qa.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if qa.config.EnableAdaptation {
				qa.adaptAllClients()
			}
		}
	}
}

// adaptAllClients adapts quality for all registered clients
func (qa *QualityAdapter) adaptAllClients() {
	qa.mutex.RLock()
	clients := make(map[string]*ClientQualityProfile)
	for id, profile := range qa.clients {
		clients[id] = profile
	}
	qa.mutex.RUnlock()

	for clientID, profile := range clients {
		qa.adaptClientQuality(clientID, profile)
	}
}

// adaptClientQuality adapts quality for a specific client
func (qa *QualityAdapter) adaptClientQuality(clientID string, profile *ClientQualityProfile) {
	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	// Check if enough time has passed since last adaptation
	if time.Since(profile.LastAdaptation) < qa.config.MinStabilityPeriod {
		return
	}

	// Calculate optimal quality level
	optimalLevel := qa.calculateOptimalQualityLevel(profile.NetworkMetrics)

	// Check if adaptation is needed
	if optimalLevel == profile.CurrentQualityLevel {
		return
	}

	// Apply adaptation sensitivity
	levelDiff := optimalLevel - profile.CurrentQualityLevel
	adaptedDiff := int(float64(levelDiff) * qa.config.AdaptationSensitivity)

	if adaptedDiff == 0 {
		if levelDiff > 0 {
			adaptedDiff = 1
		} else if levelDiff < 0 {
			adaptedDiff = -1
		}
	}

	newLevel := profile.CurrentQualityLevel + adaptedDiff

	// Ensure level is within bounds
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel >= len(qa.config.QualityLevels) {
		newLevel = len(qa.config.QualityLevels) - 1
	}

	// Apply quality change
	if newLevel != profile.CurrentQualityLevel {
		qa.applyQualityChange(clientID, profile, newLevel, "network_adaptation")
	}
}

// calculateOptimalQualityLevel calculates the optimal quality level based on network metrics
func (qa *QualityAdapter) calculateOptimalQualityLevel(metrics *NetworkMetrics) int {
	if metrics == nil {
		return qa.config.DefaultQualityLevel
	}

	// Get smoothed bandwidth
	smoothedBandwidth := qa.getSmoothedBandwidth(metrics)

	// Find the highest quality level that fits the bandwidth
	for i := len(qa.config.QualityLevels) - 1; i >= 0; i-- {
		level := qa.config.QualityLevels[i]

		// Check bandwidth requirement
		if smoothedBandwidth >= level.MinBandwidth {
			// Additional checks for latency and packet loss
			if metrics.Latency < 200*time.Millisecond && metrics.PacketLoss < 5.0 {
				return i
			}
			// Downgrade for poor network conditions
			if i > 0 {
				return i - 1
			}
		}
	}

	return 0 // Fallback to lowest quality
}

// getSmoothedBandwidth calculates smoothed bandwidth from history
func (qa *QualityAdapter) getSmoothedBandwidth(metrics *NetworkMetrics) int64 {
	if len(metrics.BandwidthHistory) == 0 {
		return metrics.Bandwidth
	}

	// Calculate weighted average (more recent values have higher weight)
	var weightedSum, totalWeight float64

	for i, bandwidth := range metrics.BandwidthHistory {
		weight := float64(i+1) / float64(len(metrics.BandwidthHistory))
		weightedSum += float64(bandwidth) * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return int64(weightedSum / totalWeight)
	}

	return metrics.Bandwidth
}

// calculateConnectionQuality calculates overall connection quality score
func (qa *QualityAdapter) calculateConnectionQuality(metrics *NetworkMetrics) float64 {
	if metrics == nil {
		return 0.5
	}

	// Bandwidth score (0-1)
	bandwidthScore := math.Min(float64(metrics.Bandwidth)/5000000.0, 1.0) // 5 Mbps = 1.0

	// Latency score (0-1, lower is better)
	latencyMs := float64(metrics.Latency.Milliseconds())
	latencyScore := math.Max(0, 1.0-latencyMs/500.0) // 500ms = 0.0

	// Packet loss score (0-1, lower is better)
	packetLossScore := math.Max(0, 1.0-metrics.PacketLoss/10.0) // 10% = 0.0

	// Jitter score (0-1, lower is better)
	jitterMs := float64(metrics.Jitter.Milliseconds())
	jitterScore := math.Max(0, 1.0-jitterMs/100.0) // 100ms = 0.0

	// Weighted average
	quality := (bandwidthScore*0.4 + latencyScore*0.3 + packetLossScore*0.2 + jitterScore*0.1)

	return math.Max(0, math.Min(1, quality))
}

// applyQualityChange applies a quality level change for a client
func (qa *QualityAdapter) applyQualityChange(clientID string, profile *ClientQualityProfile, newLevel int, reason string) {
	oldLevel := profile.CurrentQualityLevel
	profile.CurrentQualityLevel = newLevel
	profile.TargetQualityLevel = newLevel
	profile.LastAdaptation = time.Now()
	profile.AdaptationCount++

	// Record quality change
	change := QualityChange{
		Timestamp:    time.Now(),
		FromLevel:    oldLevel,
		ToLevel:      newLevel,
		Reason:       reason,
		NetworkState: profile.NetworkMetrics,
	}
	profile.QualityHistory = append(profile.QualityHistory, change)

	// Limit history size
	if len(profile.QualityHistory) > 100 {
		profile.QualityHistory = profile.QualityHistory[1:]
	}

	// Update stability score
	qa.updateStabilityScore(profile)

	// Update global stats
	qa.updateStats(oldLevel, newLevel)

	qa.logger.Info("Quality level adapted",
		zap.String("client_id", clientID),
		zap.Int("from_level", oldLevel),
		zap.Int("to_level", newLevel),
		zap.String("reason", reason),
		zap.String("quality_name", qa.config.QualityLevels[newLevel].Name))
}

// updateStabilityScore updates the stability score for a client
func (qa *QualityAdapter) updateStabilityScore(profile *ClientQualityProfile) {
	// Calculate stability based on recent adaptations
	recentChanges := 0
	cutoff := time.Now().Add(-5 * time.Minute)

	for _, change := range profile.QualityHistory {
		if change.Timestamp.After(cutoff) {
			recentChanges++
		}
	}

	// Higher stability score for fewer recent changes
	profile.StabilityScore = math.Max(0, 1.0-float64(recentChanges)/10.0)
}

// updateStats updates global adapter statistics
func (qa *QualityAdapter) updateStats(oldLevel, newLevel int) {
	qa.stats.mutex.Lock()
	defer qa.stats.mutex.Unlock()

	qa.stats.TotalAdaptations++
	qa.stats.LastAdaptation = time.Now()

	if newLevel > oldLevel {
		qa.stats.QualityUpgrades++
	} else if newLevel < oldLevel {
		qa.stats.QualityDowngrades++
	}

	// Calculate adaptation rate (adaptations per minute)
	elapsed := time.Since(qa.stats.StartTime).Minutes()
	if elapsed > 0 {
		qa.stats.AdaptationRate = float64(qa.stats.TotalAdaptations) / elapsed
	}
}

// GetStats returns adapter statistics
func (qa *QualityAdapter) GetStats() *QualityAdapterStats {
	qa.stats.mutex.RLock()
	defer qa.stats.mutex.RUnlock()

	qa.mutex.RLock()
	clientsMonitored := len(qa.clients)

	// Calculate average quality level
	var totalQuality float64
	for _, profile := range qa.clients {
		profile.mutex.RLock()
		totalQuality += float64(profile.CurrentQualityLevel)
		profile.mutex.RUnlock()
	}
	qa.mutex.RUnlock()

	averageQuality := 0.0
	if clientsMonitored > 0 {
		averageQuality = totalQuality / float64(clientsMonitored)
	}

	return &QualityAdapterStats{
		TotalAdaptations:    qa.stats.TotalAdaptations,
		QualityUpgrades:     qa.stats.QualityUpgrades,
		QualityDowngrades:   qa.stats.QualityDowngrades,
		ClientsMonitored:    clientsMonitored,
		AverageQualityLevel: averageQuality,
		AdaptationRate:      qa.stats.AdaptationRate,
		StartTime:           qa.stats.StartTime,
		LastAdaptation:      qa.stats.LastAdaptation,
	}
}

// IsRunning returns whether the adapter is running
func (qa *QualityAdapter) IsRunning() bool {
	qa.mutex.RLock()
	defer qa.mutex.RUnlock()
	return qa.isRunning
}
