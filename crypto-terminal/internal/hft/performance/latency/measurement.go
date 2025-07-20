package latency

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LatencyPoint represents a point in the latency measurement chain
type LatencyPoint string

const (
	// Market data flow
	PointMarketDataReceived    LatencyPoint = "market_data_received"
	PointMarketDataParsed      LatencyPoint = "market_data_parsed"
	PointMarketDataProcessed   LatencyPoint = "market_data_processed"
	PointOrderBookUpdated      LatencyPoint = "order_book_updated"
	
	// Signal generation
	PointSignalGenerated       LatencyPoint = "signal_generated"
	PointSignalValidated       LatencyPoint = "signal_validated"
	PointRiskChecked          LatencyPoint = "risk_checked"
	
	// Order management
	PointOrderCreated         LatencyPoint = "order_created"
	PointOrderValidated       LatencyPoint = "order_validated"
	PointOrderSent            LatencyPoint = "order_sent"
	PointOrderAcknowledged    LatencyPoint = "order_acknowledged"
	PointOrderFilled          LatencyPoint = "order_filled"
	
	// System points
	PointSystemStart          LatencyPoint = "system_start"
	PointSystemEnd            LatencyPoint = "system_end"
)

// LatencyMeasurement represents a single latency measurement
type LatencyMeasurement struct {
	ID          string
	Point       LatencyPoint
	Timestamp   int64 // nanoseconds since epoch
	ThreadID    int
	CPUCore     int
	Metadata    map[string]interface{}
}

// LatencyTrace represents a complete latency trace through the system
type LatencyTrace struct {
	TraceID      string
	StartTime    int64
	EndTime      int64
	TotalLatency time.Duration
	Points       []LatencyMeasurement
	Metadata     map[string]interface{}
}

// LatencyStats holds statistical information about latencies
type LatencyStats struct {
	Count       uint64
	Sum         uint64  // nanoseconds
	Min         uint64  // nanoseconds
	Max         uint64  // nanoseconds
	Mean        float64 // nanoseconds
	P50         uint64  // nanoseconds
	P95         uint64  // nanoseconds
	P99         uint64  // nanoseconds
	P999        uint64  // nanoseconds
	LastUpdated int64
}

// HighResolutionTimer provides nanosecond precision timing
type HighResolutionTimer struct {
	startTime int64
	frequency int64
}

// NewHighResolutionTimer creates a new high-resolution timer
func NewHighResolutionTimer() *HighResolutionTimer {
	return &HighResolutionTimer{
		frequency: 1000000000, // 1 GHz (nanosecond precision)
	}
}

// Now returns the current time in nanoseconds with high precision
func (t *HighResolutionTimer) Now() int64 {
	// On x86-64, we can use RDTSC for even higher precision
	// For now, we'll use Go's time.Now() which has nanosecond precision
	return time.Now().UnixNano()
}

// Start starts the timer
func (t *HighResolutionTimer) Start() {
	t.startTime = t.Now()
}

// Elapsed returns the elapsed time since Start() was called
func (t *HighResolutionTimer) Elapsed() time.Duration {
	return time.Duration(t.Now() - t.startTime)
}

// LatencyTracker tracks latency measurements across the system
type LatencyTracker struct {
	traces      sync.Map // map[string]*LatencyTrace
	stats       sync.Map // map[LatencyPoint]*LatencyStats
	timer       *HighResolutionTimer
	tracer      trace.Tracer
	
	// Configuration
	maxTraces   int32
	traceCount  int32
	enabled     int32 // atomic bool
	
	// Performance optimization
	measurementPool sync.Pool
	tracePool       sync.Pool
	
	// Statistics collection
	statsInterval   time.Duration
	lastStatsUpdate int64
}

// NewLatencyTracker creates a new latency tracker
func NewLatencyTracker(maxTraces int) *LatencyTracker {
	tracker := &LatencyTracker{
		timer:         NewHighResolutionTimer(),
		tracer:        otel.Tracer("hft.latency.tracker"),
		maxTraces:     int32(maxTraces),
		enabled:       1,
		statsInterval: 1 * time.Second,
	}
	
	// Initialize object pools
	tracker.measurementPool = sync.Pool{
		New: func() interface{} {
			return &LatencyMeasurement{
				Metadata: make(map[string]interface{}),
			}
		},
	}
	
	tracker.tracePool = sync.Pool{
		New: func() interface{} {
			return &LatencyTrace{
				Points:   make([]LatencyMeasurement, 0, 10),
				Metadata: make(map[string]interface{}),
			}
		},
	}
	
	return tracker
}

// StartTrace starts a new latency trace
func (lt *LatencyTracker) StartTrace(traceID string) *LatencyTrace {
	if atomic.LoadInt32(&lt.enabled) == 0 {
		return nil
	}
	
	// Check if we've exceeded max traces
	if atomic.LoadInt32(&lt.traceCount) >= lt.maxTraces {
		return nil
	}
	
	trace := lt.tracePool.Get().(*LatencyTrace)
	trace.TraceID = traceID
	trace.StartTime = lt.timer.Now()
	trace.Points = trace.Points[:0] // Reset slice
	
	// Clear metadata
	for k := range trace.Metadata {
		delete(trace.Metadata, k)
	}
	
	lt.traces.Store(traceID, trace)
	atomic.AddInt32(&lt.traceCount, 1)
	
	return trace
}

// RecordPoint records a latency measurement point
func (lt *LatencyTracker) RecordPoint(traceID string, point LatencyPoint, metadata ...map[string]interface{}) {
	if atomic.LoadInt32(&lt.enabled) == 0 {
		return
	}
	
	traceInterface, exists := lt.traces.Load(traceID)
	if !exists {
		return
	}
	
	trace := traceInterface.(*LatencyTrace)
	timestamp := lt.timer.Now()
	
	measurement := lt.measurementPool.Get().(*LatencyMeasurement)
	measurement.ID = fmt.Sprintf("%s_%s_%d", traceID, point, timestamp)
	measurement.Point = point
	measurement.Timestamp = timestamp
	measurement.ThreadID = getThreadID()
	measurement.CPUCore = getCPUCore()
	
	// Clear and set metadata
	for k := range measurement.Metadata {
		delete(measurement.Metadata, k)
	}
	if len(metadata) > 0 {
		for k, v := range metadata[0] {
			measurement.Metadata[k] = v
		}
	}
	
	// Add to trace
	trace.Points = append(trace.Points, *measurement)
	
	// Return measurement to pool
	lt.measurementPool.Put(measurement)
	
	// Update statistics
	lt.updatePointStats(point, timestamp, trace.StartTime)
}

// EndTrace ends a latency trace and calculates total latency
func (lt *LatencyTracker) EndTrace(traceID string) *LatencyTrace {
	if atomic.LoadInt32(&lt.enabled) == 0 {
		return nil
	}
	
	traceInterface, exists := lt.traces.LoadAndDelete(traceID)
	if !exists {
		return nil
	}
	
	trace := traceInterface.(*LatencyTrace)
	trace.EndTime = lt.timer.Now()
	trace.TotalLatency = time.Duration(trace.EndTime - trace.StartTime)
	
	atomic.AddInt32(&lt.traceCount, -1)
	
	// Create a copy for return (original goes back to pool)
	result := &LatencyTrace{
		TraceID:      trace.TraceID,
		StartTime:    trace.StartTime,
		EndTime:      trace.EndTime,
		TotalLatency: trace.TotalLatency,
		Points:       make([]LatencyMeasurement, len(trace.Points)),
		Metadata:     make(map[string]interface{}),
	}
	
	copy(result.Points, trace.Points)
	for k, v := range trace.Metadata {
		result.Metadata[k] = v
	}
	
	// Return trace to pool
	lt.tracePool.Put(trace)
	
	return result
}

// GetTrace retrieves an active trace
func (lt *LatencyTracker) GetTrace(traceID string) *LatencyTrace {
	traceInterface, exists := lt.traces.Load(traceID)
	if !exists {
		return nil
	}
	
	return traceInterface.(*LatencyTrace)
}

// GetStats returns statistics for a specific latency point
func (lt *LatencyTracker) GetStats(point LatencyPoint) *LatencyStats {
	statsInterface, exists := lt.stats.Load(point)
	if !exists {
		return nil
	}
	
	stats := statsInterface.(*LatencyStats)
	
	// Return a copy to avoid race conditions
	return &LatencyStats{
		Count:       atomic.LoadUint64(&stats.Count),
		Sum:         atomic.LoadUint64(&stats.Sum),
		Min:         atomic.LoadUint64(&stats.Min),
		Max:         atomic.LoadUint64(&stats.Max),
		Mean:        stats.Mean, // This should be atomic too in production
		P50:         atomic.LoadUint64(&stats.P50),
		P95:         atomic.LoadUint64(&stats.P95),
		P99:         atomic.LoadUint64(&stats.P99),
		P999:        atomic.LoadUint64(&stats.P999),
		LastUpdated: atomic.LoadInt64(&stats.LastUpdated),
	}
}

// GetAllStats returns statistics for all latency points
func (lt *LatencyTracker) GetAllStats() map[LatencyPoint]*LatencyStats {
	result := make(map[LatencyPoint]*LatencyStats)
	
	lt.stats.Range(func(key, value interface{}) bool {
		point := key.(LatencyPoint)
		result[point] = lt.GetStats(point)
		return true
	})
	
	return result
}

// Enable enables latency tracking
func (lt *LatencyTracker) Enable() {
	atomic.StoreInt32(&lt.enabled, 1)
}

// Disable disables latency tracking
func (lt *LatencyTracker) Disable() {
	atomic.StoreInt32(&lt.enabled, 0)
}

// IsEnabled returns true if latency tracking is enabled
func (lt *LatencyTracker) IsEnabled() bool {
	return atomic.LoadInt32(&lt.enabled) == 1
}

// ClearStats clears all statistics
func (lt *LatencyTracker) ClearStats() {
	lt.stats.Range(func(key, value interface{}) bool {
		lt.stats.Delete(key)
		return true
	})
}

// updatePointStats updates statistics for a latency point
func (lt *LatencyTracker) updatePointStats(point LatencyPoint, timestamp, startTime int64) {
	latency := uint64(timestamp - startTime)
	
	statsInterface, _ := lt.stats.LoadOrStore(point, &LatencyStats{
		Min: ^uint64(0), // Max uint64
	})
	
	stats := statsInterface.(*LatencyStats)
	
	// Update atomic counters
	atomic.AddUint64(&stats.Count, 1)
	atomic.AddUint64(&stats.Sum, latency)
	atomic.StoreInt64(&stats.LastUpdated, timestamp)
	
	// Update min
	for {
		current := atomic.LoadUint64(&stats.Min)
		if latency >= current || atomic.CompareAndSwapUint64(&stats.Min, current, latency) {
			break
		}
	}
	
	// Update max
	for {
		current := atomic.LoadUint64(&stats.Max)
		if latency <= current || atomic.CompareAndSwapUint64(&stats.Max, current, latency) {
			break
		}
	}
	
	// Update mean (simplified - should use more sophisticated algorithm in production)
	count := atomic.LoadUint64(&stats.Count)
	sum := atomic.LoadUint64(&stats.Sum)
	stats.Mean = float64(sum) / float64(count)
	
	// Percentiles would be calculated periodically using a histogram
	// For now, we'll use simplified approximations
	lt.updatePercentiles(stats, latency)
}

// updatePercentiles updates percentile statistics (simplified implementation)
func (lt *LatencyTracker) updatePercentiles(stats *LatencyStats, latency uint64) {
	// This is a simplified implementation
	// In production, you'd use a proper histogram or quantile estimation algorithm
	
	// Simple approximation: use exponential moving average
	alpha := 0.1
	
	current50 := atomic.LoadUint64(&stats.P50)
	if current50 == 0 {
		atomic.StoreUint64(&stats.P50, latency)
	} else {
		new50 := uint64(float64(current50)*(1-alpha) + float64(latency)*alpha)
		atomic.StoreUint64(&stats.P50, new50)
	}
	
	// P95 approximation
	current95 := atomic.LoadUint64(&stats.P95)
	if latency > current95 {
		new95 := uint64(float64(current95)*0.95 + float64(latency)*0.05)
		atomic.StoreUint64(&stats.P95, new95)
	}
	
	// P99 approximation
	current99 := atomic.LoadUint64(&stats.P99)
	if latency > current99 {
		new99 := uint64(float64(current99)*0.99 + float64(latency)*0.01)
		atomic.StoreUint64(&stats.P99, new99)
	}
	
	// P999 approximation
	current999 := atomic.LoadUint64(&stats.P999)
	if latency > current999 {
		new999 := uint64(float64(current999)*0.999 + float64(latency)*0.001)
		atomic.StoreUint64(&stats.P999, new999)
	}
}

// LatencyProfiler provides high-level profiling functionality
type LatencyProfiler struct {
	tracker *LatencyTracker
	tracer  trace.Tracer
}

// NewLatencyProfiler creates a new latency profiler
func NewLatencyProfiler(maxTraces int) *LatencyProfiler {
	return &LatencyProfiler{
		tracker: NewLatencyTracker(maxTraces),
		tracer:  otel.Tracer("hft.latency.profiler"),
	}
}

// ProfileFunction profiles the latency of a function execution
func (lp *LatencyProfiler) ProfileFunction(ctx context.Context, name string, fn func() error) error {
	ctx, span := lp.tracer.Start(ctx, fmt.Sprintf("profile_%s", name))
	defer span.End()
	
	traceID := fmt.Sprintf("%s_%d", name, time.Now().UnixNano())
	trace := lp.tracker.StartTrace(traceID)
	if trace == nil {
		return fn() // Tracking disabled or full
	}
	
	lp.tracker.RecordPoint(traceID, PointSystemStart)
	
	err := fn()
	
	lp.tracker.RecordPoint(traceID, PointSystemEnd)
	completedTrace := lp.tracker.EndTrace(traceID)
	
	if completedTrace != nil {
		span.SetAttributes(
			attribute.String("trace_id", traceID),
			attribute.Int64("total_latency_ns", int64(completedTrace.TotalLatency)),
			attribute.Int("measurement_points", len(completedTrace.Points)),
		)
	}
	
	return err
}

// GetTracker returns the underlying latency tracker
func (lp *LatencyProfiler) GetTracker() *LatencyTracker {
	return lp.tracker
}

// Helper functions

// getThreadID returns the current thread ID (simplified)
func getThreadID() int {
	// In a real implementation, this would use runtime.Goid() or similar
	return int(uintptr(unsafe.Pointer(&struct{}{}))) % 10000
}

// getCPUCore returns the current CPU core (simplified)
func getCPUCore() int {
	// In a real implementation, this would use sched_getcpu() on Linux
	return runtime.NumCPU() % 8 // Simplified approximation
}

// Global latency tracker instance
var globalLatencyTracker *LatencyTracker
var latencyTrackerOnce sync.Once

// GetGlobalLatencyTracker returns the global latency tracker
func GetGlobalLatencyTracker() *LatencyTracker {
	latencyTrackerOnce.Do(func() {
		globalLatencyTracker = NewLatencyTracker(10000)
	})
	return globalLatencyTracker
}

// Convenience functions for common use cases

// MeasureTickToTrade measures the complete tick-to-trade latency
func MeasureTickToTrade(traceID string) func() {
	tracker := GetGlobalLatencyTracker()
	trace := tracker.StartTrace(traceID)
	if trace == nil {
		return func() {} // No-op if tracking disabled
	}
	
	tracker.RecordPoint(traceID, PointMarketDataReceived)
	
	return func() {
		tracker.RecordPoint(traceID, PointOrderSent)
		tracker.EndTrace(traceID)
	}
}

// MeasureOrderProcessing measures order processing latency
func MeasureOrderProcessing(traceID string) func() {
	tracker := GetGlobalLatencyTracker()
	tracker.RecordPoint(traceID, PointOrderCreated)
	
	return func() {
		tracker.RecordPoint(traceID, PointOrderAcknowledged)
	}
}

// MeasureRiskCheck measures risk check latency
func MeasureRiskCheck(traceID string) func() {
	tracker := GetGlobalLatencyTracker()
	start := tracker.timer.Now()
	
	return func() {
		end := tracker.timer.Now()
		tracker.RecordPoint(traceID, PointRiskChecked, map[string]interface{}{
			"risk_check_duration_ns": end - start,
		})
	}
}
