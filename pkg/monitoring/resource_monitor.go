package monitoring

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// ResourceMetrics holds system resource metrics
type ResourceMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsagePercent float64   `json:"cpu_usage_percent"`
	MemoryUsageMB   float64   `json:"memory_usage_mb"`
	MemoryTotalMB   float64   `json:"memory_total_mb"`
	GoroutineCount  int       `json:"goroutine_count"`
	HeapSizeMB      float64   `json:"heap_size_mb"`
	HeapInUseMB     float64   `json:"heap_in_use_mb"`
	GCPauseMS       float64   `json:"gc_pause_ms"`
	GCCount         uint32    `json:"gc_count"`
}

// AlertThreshold defines when to trigger alerts
type AlertThreshold struct {
	CPUPercent    float64
	MemoryPercent float64
	GoroutineMax  int
}

// ResourceMonitor monitors system resources and triggers alerts
type ResourceMonitor struct {
	mu            sync.RWMutex
	metrics       []ResourceMetrics
	maxHistory    int
	threshold     AlertThreshold
	alertCallback func(alert string, metrics ResourceMetrics)
	ticker        *time.Ticker
	stopChan      chan struct{}
	running       bool
	lastGCStats   runtime.MemStats
}

// ResourceMonitorConfig configures the resource monitor
type ResourceMonitorConfig struct {
	MonitorInterval time.Duration
	MaxHistory      int
	Threshold       AlertThreshold
	AlertCallback   func(alert string, metrics ResourceMetrics)
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(config ResourceMonitorConfig) *ResourceMonitor {
	if config.MonitorInterval == 0 {
		config.MonitorInterval = time.Second * 30
	}
	if config.MaxHistory == 0 {
		config.MaxHistory = 100
	}
	if config.Threshold.CPUPercent == 0 {
		config.Threshold.CPUPercent = 80.0
	}
	if config.Threshold.MemoryPercent == 0 {
		config.Threshold.MemoryPercent = 85.0
	}
	if config.Threshold.GoroutineMax == 0 {
		config.Threshold.GoroutineMax = 10000
	}

	monitor := &ResourceMonitor{
		maxHistory:    config.MaxHistory,
		threshold:     config.Threshold,
		alertCallback: config.AlertCallback,
		ticker:        time.NewTicker(config.MonitorInterval),
		stopChan:      make(chan struct{}),
		metrics:       make([]ResourceMetrics, 0, config.MaxHistory),
	}

	// Get initial GC stats
	runtime.ReadMemStats(&monitor.lastGCStats)

	return monitor
}

// Start begins resource monitoring
func (rm *ResourceMonitor) Start() {
	rm.mu.Lock()
	if rm.running {
		rm.mu.Unlock()
		return
	}
	rm.running = true
	rm.mu.Unlock()

	go rm.monitor()
	log.Println("Resource monitor started")
}

// Stop stops resource monitoring
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	if !rm.running {
		rm.mu.Unlock()
		return
	}
	rm.running = false
	rm.mu.Unlock()

	close(rm.stopChan)
	rm.ticker.Stop()
	log.Println("Resource monitor stopped")
}

// monitor is the main monitoring loop
func (rm *ResourceMonitor) monitor() {
	for {
		select {
		case <-rm.ticker.C:
			rm.collectMetrics()
		case <-rm.stopChan:
			return
		}
	}
}

// collectMetrics gathers current system metrics
func (rm *ResourceMonitor) collectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := ResourceMetrics{
		Timestamp:      time.Now(),
		GoroutineCount: runtime.NumGoroutine(),
		MemoryUsageMB:  float64(memStats.Sys) / 1024 / 1024,
		HeapSizeMB:     float64(memStats.HeapSys) / 1024 / 1024,
		HeapInUseMB:    float64(memStats.HeapInuse) / 1024 / 1024,
		GCCount:        memStats.NumGC,
	}

	// Calculate GC pause time
	if memStats.NumGC > rm.lastGCStats.NumGC {
		// Get the most recent GC pause
		gcPauseIndex := (memStats.NumGC + 255) % 256
		metrics.GCPauseMS = float64(memStats.PauseNs[gcPauseIndex]) / 1e6
	}

	// Simple CPU estimation based on GC activity and goroutines
	// In a real implementation, you would use more sophisticated CPU monitoring
	metrics.CPUUsagePercent = rm.estimateCPUUsage(metrics)

	// Store memory total (approximation)
	metrics.MemoryTotalMB = metrics.MemoryUsageMB * 1.2 // Rough estimate

	// Store metrics
	rm.mu.Lock()
	rm.metrics = append(rm.metrics, metrics)
	if len(rm.metrics) > rm.maxHistory {
		rm.metrics = rm.metrics[1:]
	}
	rm.mu.Unlock()

	// Check thresholds and trigger alerts
	rm.checkThresholds(metrics)

	// Update last GC stats
	rm.lastGCStats = memStats
}

// estimateCPUUsage provides a rough CPU usage estimate
func (rm *ResourceMonitor) estimateCPUUsage(metrics ResourceMetrics) float64 {
	// This is a simplified estimation based on:
	// - Goroutine count (more goroutines = higher CPU potential)
	// - GC activity (recent GC = CPU usage)
	// - Memory pressure (high memory usage can indicate CPU activity)

	cpuEstimate := 0.0

	// Base estimate from goroutine count
	if metrics.GoroutineCount > 100 {
		cpuEstimate += float64(metrics.GoroutineCount) / 1000 * 10
	}

	// Add estimate from GC activity
	if metrics.GCPauseMS > 0 {
		cpuEstimate += metrics.GCPauseMS / 10
	}

	// Add estimate from memory pressure
	memoryPressure := metrics.HeapInUseMB / metrics.HeapSizeMB
	if memoryPressure > 0.8 {
		cpuEstimate += (memoryPressure - 0.8) * 50
	}

	// Cap at 100%
	if cpuEstimate > 100 {
		cpuEstimate = 100
	}

	return cpuEstimate
}

// checkThresholds checks if any thresholds are exceeded
func (rm *ResourceMonitor) checkThresholds(metrics ResourceMetrics) {
	alerts := make([]string, 0)

	// Check CPU threshold
	if metrics.CPUUsagePercent > rm.threshold.CPUPercent {
		alerts = append(alerts, fmt.Sprintf("High CPU usage: %.2f%%", metrics.CPUUsagePercent))
	}

	// Check memory threshold
	memoryPercent := (metrics.HeapInUseMB / metrics.MemoryTotalMB) * 100
	if memoryPercent > rm.threshold.MemoryPercent {
		alerts = append(alerts, fmt.Sprintf("High memory usage: %.2f%%", memoryPercent))
	}

	// Check goroutine threshold
	if metrics.GoroutineCount > rm.threshold.GoroutineMax {
		alerts = append(alerts, fmt.Sprintf("High goroutine count: %d", metrics.GoroutineCount))
	}

	// Trigger alerts
	for _, alert := range alerts {
		if rm.alertCallback != nil {
			rm.alertCallback(alert, metrics)
		} else {
			log.Printf("ALERT: %s", alert)
		}
	}
}

// GetCurrentMetrics returns the most recent metrics
func (rm *ResourceMonitor) GetCurrentMetrics() (ResourceMetrics, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.metrics) == 0 {
		return ResourceMetrics{}, fmt.Errorf("no metrics available")
	}

	return rm.metrics[len(rm.metrics)-1], nil
}

// GetMetricsHistory returns historical metrics
func (rm *ResourceMonitor) GetMetricsHistory(duration time.Duration) []ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	var result []ResourceMetrics

	for _, metric := range rm.metrics {
		if metric.Timestamp.After(cutoff) {
			result = append(result, metric)
		}
	}

	return result
}

// GetAverageMetrics calculates average metrics over a duration
func (rm *ResourceMonitor) GetAverageMetrics(duration time.Duration) ResourceMetrics {
	history := rm.GetMetricsHistory(duration)
	if len(history) == 0 {
		return ResourceMetrics{}
	}

	var avg ResourceMetrics
	for _, metric := range history {
		avg.CPUUsagePercent += metric.CPUUsagePercent
		avg.MemoryUsageMB += metric.MemoryUsageMB
		avg.GoroutineCount += metric.GoroutineCount
		avg.HeapSizeMB += metric.HeapSizeMB
		avg.HeapInUseMB += metric.HeapInUseMB
		avg.GCPauseMS += metric.GCPauseMS
	}

	count := float64(len(history))
	avg.CPUUsagePercent /= count
	avg.MemoryUsageMB /= count
	avg.GoroutineCount = int(float64(avg.GoroutineCount) / count)
	avg.HeapSizeMB /= count
	avg.HeapInUseMB /= count
	avg.GCPauseMS /= count
	avg.Timestamp = time.Now()

	return avg
}

// IsHealthy returns true if all metrics are within acceptable ranges
func (rm *ResourceMonitor) IsHealthy() bool {
	current, err := rm.GetCurrentMetrics()
	if err != nil {
		return false
	}

	// Check if any threshold is exceeded
	memoryPercent := (current.HeapInUseMB / current.MemoryTotalMB) * 100

	return current.CPUUsagePercent <= rm.threshold.CPUPercent &&
		memoryPercent <= rm.threshold.MemoryPercent &&
		current.GoroutineCount <= rm.threshold.GoroutineMax
}

// TriggerGC forces garbage collection and returns metrics before/after
func (rm *ResourceMonitor) TriggerGC() (before, after ResourceMetrics) {
	// Get metrics before GC
	before, _ = rm.GetCurrentMetrics()

	// Force GC
	runtime.GC()
	runtime.GC() // Run twice to ensure complete collection

	// Wait a moment for GC to complete
	time.Sleep(time.Millisecond * 100)

	// Collect new metrics
	rm.collectMetrics()
	after, _ = rm.GetCurrentMetrics()

	return before, after
}

// DefaultConfig returns default monitoring configuration
func DefaultConfig() ResourceMonitorConfig {
	return ResourceMonitorConfig{
		MonitorInterval: time.Second * 30,
		MaxHistory:      100,
		Threshold: AlertThreshold{
			CPUPercent:    80.0,
			MemoryPercent: 85.0,
			GoroutineMax:  10000,
		},
		AlertCallback: func(alert string, metrics ResourceMetrics) {
			log.Printf("RESOURCE ALERT: %s (Goroutines: %d, Memory: %.2fMB, CPU: %.2f%%)",
				alert, metrics.GoroutineCount, metrics.MemoryUsageMB, metrics.CPUUsagePercent)
		},
	}
}

// HealthCheckHandler provides a simple health check endpoint data
func (rm *ResourceMonitor) HealthCheckHandler() map[string]interface{} {
	current, err := rm.GetCurrentMetrics()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	status := "healthy"
	if !rm.IsHealthy() {
		status = "warning"
	}

	return map[string]interface{}{
		"status":         status,
		"timestamp":      current.Timestamp,
		"cpu_percent":    current.CPUUsagePercent,
		"memory_mb":      current.MemoryUsageMB,
		"goroutines":     current.GoroutineCount,
		"heap_size_mb":   current.HeapSizeMB,
		"heap_in_use_mb": current.HeapInUseMB,
		"gc_pause_ms":    current.GCPauseMS,
		"gc_count":       current.GCCount,
	}
}
