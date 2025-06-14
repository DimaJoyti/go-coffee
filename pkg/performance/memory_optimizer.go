package performance

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryOptimizer provides advanced memory management and optimization
type MemoryOptimizer struct {
	logger        *zap.Logger
	config        *MemoryConfig
	pools         map[string]*ObjectPool
	gcController  *GCController
	memoryMonitor *MemoryMonitor
	leakDetector  *LeakDetector
	mu            sync.RWMutex
}

// MemoryConfig contains memory optimization settings
type MemoryConfig struct {
	// GC settings
	GCPercent          int
	MaxGCPause         time.Duration
	GCTriggerThreshold float64

	// Pool settings
	PoolCleanupInterval time.Duration
	MaxPoolSize         int
	PoolIdleTimeout     time.Duration

	// Monitoring settings
	MonitorInterval      time.Duration
	MemoryThreshold      float64
	LeakDetectionEnabled bool

	// Optimization settings
	EnableAutoGC    bool
	EnablePooling   bool
	EnableProfiling bool
}

// ObjectPool provides efficient object reuse
type ObjectPool struct {
	name        string
	pool        sync.Pool
	metrics     *PoolMetrics
	maxSize     int
	idleTimeout time.Duration
	lastCleanup time.Time
	mu          sync.RWMutex
}

// PoolMetrics tracks pool performance
type PoolMetrics struct {
	Gets        int64
	Puts        int64
	News        int64
	Cleanups    int64
	CurrentSize int64
	MaxSize     int64
	HitRatio    float64
	mu          sync.RWMutex
}

// GCController manages garbage collection optimization
type GCController struct {
	optimizer *MemoryOptimizer
	gcPercent int
	maxPause  time.Duration
	lastGC    time.Time
	gcStats   *GCStats
	mu        sync.RWMutex
}

// GCStats tracks garbage collection statistics
type GCStats struct {
	NumGC       uint32
	TotalPause  time.Duration
	LastPause   time.Duration
	AvgPause    time.Duration
	HeapSize    uint64
	HeapInUse   uint64
	HeapObjects uint64
}

// MemoryMonitor tracks memory usage and triggers optimizations
type MemoryMonitor struct {
	optimizer     *MemoryOptimizer
	threshold     float64
	interval      time.Duration
	lastCheck     time.Time
	memoryStats   *MemoryStats
	alertCallback func(*MemoryStats)
	mu            sync.RWMutex
}

// MemoryStats contains current memory statistics
type MemoryStats struct {
	Alloc         uint64
	TotalAlloc    uint64
	Sys           uint64
	Lookups       uint64
	Mallocs       uint64
	Frees         uint64
	HeapAlloc     uint64
	HeapSys       uint64
	HeapIdle      uint64
	HeapInuse     uint64
	HeapReleased  uint64
	HeapObjects   uint64
	StackInuse    uint64
	StackSys      uint64
	MSpanInuse    uint64
	MSpanSys      uint64
	MCacheInuse   uint64
	MCacheSys     uint64
	BuckHashSys   uint64
	GCSys         uint64
	OtherSys      uint64
	NextGC        uint64
	LastGC        uint64
	PauseTotalNs  uint64
	NumGC         uint32
	NumForcedGC   uint32
	GCCPUFraction float64
	Timestamp     time.Time
}

// LeakDetector identifies potential memory leaks
type LeakDetector struct {
	optimizer    *MemoryOptimizer
	enabled      bool
	snapshots    []*MemoryStats
	maxSnapshots int
	interval     time.Duration
	mu           sync.RWMutex
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(config *MemoryConfig, logger *zap.Logger) *MemoryOptimizer {
	optimizer := &MemoryOptimizer{
		logger: logger,
		config: config,
		pools:  make(map[string]*ObjectPool),
	}

	// Initialize GC controller
	optimizer.gcController = NewGCController(optimizer, config.GCPercent, config.MaxGCPause)

	// Initialize memory monitor
	optimizer.memoryMonitor = NewMemoryMonitor(optimizer, config.MemoryThreshold, config.MonitorInterval)

	// Initialize leak detector if enabled
	if config.LeakDetectionEnabled {
		optimizer.leakDetector = NewLeakDetector(optimizer, config.MonitorInterval)
	}

	// Start optimization routines
	go optimizer.startOptimization()

	return optimizer
}

// CreatePool creates a new object pool
func (mo *MemoryOptimizer) CreatePool(name string, newFunc func() interface{}) *ObjectPool {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	pool := &ObjectPool{
		name:        name,
		maxSize:     mo.config.MaxPoolSize,
		idleTimeout: mo.config.PoolIdleTimeout,
		lastCleanup: time.Now(),
		metrics:     &PoolMetrics{},
	}

	pool.pool = sync.Pool{
		New: func() interface{} {
			pool.metrics.mu.Lock()
			pool.metrics.News++
			pool.metrics.mu.Unlock()
			return newFunc()
		},
	}

	mo.pools[name] = pool
	return pool
}

// Get retrieves an object from the pool
func (p *ObjectPool) Get() interface{} {
	p.metrics.mu.Lock()
	p.metrics.Gets++
	p.metrics.mu.Unlock()

	obj := p.pool.Get()

	// Update hit ratio
	p.updateHitRatio()

	return obj
}

// Put returns an object to the pool
func (p *ObjectPool) Put(obj interface{}) {
	if obj == nil {
		return
	}

	p.metrics.mu.Lock()
	p.metrics.Puts++
	p.metrics.mu.Unlock()

	p.pool.Put(obj)
}

// updateHitRatio calculates the current hit ratio
func (p *ObjectPool) updateHitRatio() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	if p.metrics.Gets > 0 {
		hits := p.metrics.Gets - p.metrics.News
		p.metrics.HitRatio = float64(hits) / float64(p.metrics.Gets)
	}
}

// GetMetrics returns pool metrics
func (p *ObjectPool) GetMetrics() PoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	return *p.metrics
}

// NewGCController creates a new GC controller
func NewGCController(optimizer *MemoryOptimizer, gcPercent int, maxPause time.Duration) *GCController {
	controller := &GCController{
		optimizer: optimizer,
		gcPercent: gcPercent,
		maxPause:  maxPause,
		gcStats:   &GCStats{},
	}

	// Set initial GC percent
	debug.SetGCPercent(gcPercent)

	return controller
}

// OptimizeGC optimizes garbage collection based on current conditions
func (gc *GCController) OptimizeGC() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Get current memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update GC stats
	gc.gcStats.NumGC = m.NumGC
	gc.gcStats.HeapSize = m.HeapSys
	gc.gcStats.HeapInUse = m.HeapInuse
	gc.gcStats.HeapObjects = m.HeapObjects

	// Calculate average pause time
	if m.NumGC > 0 {
		gc.gcStats.TotalPause = time.Duration(m.PauseTotalNs)
		gc.gcStats.AvgPause = gc.gcStats.TotalPause / time.Duration(m.NumGC)
		gc.gcStats.LastPause = time.Duration(m.PauseNs[(m.NumGC+255)%256])
	}

	// Adjust GC percent based on conditions
	if gc.gcStats.AvgPause > gc.maxPause {
		// Reduce GC percent to trigger more frequent collections
		newPercent := gc.gcPercent - 10
		if newPercent < 50 {
			newPercent = 50
		}
		if newPercent != gc.gcPercent {
			gc.gcPercent = newPercent
			debug.SetGCPercent(newPercent)
			gc.optimizer.logger.Info("Adjusted GC percent",
				zap.Int("new_percent", newPercent),
				zap.Duration("avg_pause", gc.gcStats.AvgPause),
			)
		}
	}

	// Force GC if memory usage is high
	heapUsage := float64(m.HeapInuse) / float64(m.HeapSys)
	if heapUsage > 0.8 {
		runtime.GC()
		gc.optimizer.logger.Debug("Forced garbage collection",
			zap.Float64("heap_usage", heapUsage),
		)
	}
}

// GetGCStats returns current GC statistics
func (gc *GCController) GetGCStats() GCStats {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return *gc.gcStats
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(optimizer *MemoryOptimizer, threshold float64, interval time.Duration) *MemoryMonitor {
	return &MemoryMonitor{
		optimizer:   optimizer,
		threshold:   threshold,
		interval:    interval,
		memoryStats: &MemoryStats{},
	}
}

// StartMonitoring starts memory monitoring
func (mm *MemoryMonitor) StartMonitoring() {
	ticker := time.NewTicker(mm.interval)
	defer ticker.Stop()

	for range ticker.C {
		mm.checkMemory()
	}
}

// checkMemory checks current memory usage and triggers optimizations
func (mm *MemoryMonitor) checkMemory() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update memory stats
	mm.memoryStats = &MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		Timestamp:     time.Now(),
	}

	// Check if memory usage exceeds threshold
	memoryUsage := float64(m.Alloc) / float64(m.Sys)
	if memoryUsage > mm.threshold {
		mm.optimizer.logger.Warn("High memory usage detected",
			zap.Float64("usage", memoryUsage),
			zap.Float64("threshold", mm.threshold),
			zap.Uint64("alloc", m.Alloc),
			zap.Uint64("sys", m.Sys),
		)

		// Trigger optimizations
		mm.optimizer.triggerOptimizations()

		// Call alert callback if set
		if mm.alertCallback != nil {
			go mm.alertCallback(mm.memoryStats)
		}
	}

	mm.lastCheck = time.Now()
}

// GetMemoryStats returns current memory statistics
func (mm *MemoryMonitor) GetMemoryStats() MemoryStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return *mm.memoryStats
}

// SetAlertCallback sets a callback for memory alerts
func (mm *MemoryMonitor) SetAlertCallback(callback func(*MemoryStats)) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.alertCallback = callback
}

// NewLeakDetector creates a new leak detector
func NewLeakDetector(optimizer *MemoryOptimizer, interval time.Duration) *LeakDetector {
	return &LeakDetector{
		optimizer:    optimizer,
		enabled:      true,
		snapshots:    make([]*MemoryStats, 0),
		maxSnapshots: 10,
		interval:     interval,
	}
}

// StartDetection starts leak detection
func (ld *LeakDetector) StartDetection() {
	if !ld.enabled {
		return
	}

	ticker := time.NewTicker(ld.interval)
	defer ticker.Stop()

	for range ticker.C {
		ld.takeSnapshot()
		ld.analyzeLeaks()
	}
}

// takeSnapshot takes a memory snapshot
func (ld *LeakDetector) takeSnapshot() {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	stats := ld.optimizer.memoryMonitor.GetMemoryStats()
	ld.snapshots = append(ld.snapshots, &stats)

	// Keep only recent snapshots
	if len(ld.snapshots) > ld.maxSnapshots {
		ld.snapshots = ld.snapshots[1:]
	}
}

// analyzeLeaks analyzes snapshots for potential memory leaks
func (ld *LeakDetector) analyzeLeaks() {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	if len(ld.snapshots) < 3 {
		return
	}

	// Simple leak detection: check for consistent memory growth
	recent := ld.snapshots[len(ld.snapshots)-1]
	older := ld.snapshots[len(ld.snapshots)-3]

	growthRate := float64(recent.Alloc-older.Alloc) / float64(older.Alloc)
	if growthRate > 0.1 { // 10% growth
		ld.optimizer.logger.Warn("Potential memory leak detected",
			zap.Float64("growth_rate", growthRate),
			zap.Uint64("current_alloc", recent.Alloc),
			zap.Uint64("previous_alloc", older.Alloc),
		)
	}
}

// startOptimization starts the main optimization loop
func (mo *MemoryOptimizer) startOptimization() {
	ticker := time.NewTicker(mo.config.MonitorInterval)
	defer ticker.Stop()

	// Start sub-components
	go mo.memoryMonitor.StartMonitoring()
	if mo.leakDetector != nil {
		go mo.leakDetector.StartDetection()
	}

	for range ticker.C {
		mo.runOptimizations()
	}
}

// runOptimizations runs periodic optimizations
func (mo *MemoryOptimizer) runOptimizations() {
	if mo.config.EnableAutoGC {
		mo.gcController.OptimizeGC()
	}

	if mo.config.EnablePooling {
		mo.cleanupPools()
	}
}

// triggerOptimizations triggers immediate optimizations
func (mo *MemoryOptimizer) triggerOptimizations() {
	// Force garbage collection
	runtime.GC()

	// Clean up pools
	mo.cleanupPools()

	// Return memory to OS
	debug.FreeOSMemory()
}

// cleanupPools cleans up idle objects in pools
func (mo *MemoryOptimizer) cleanupPools() {
	mo.mu.RLock()
	defer mo.mu.RUnlock()

	for name, pool := range mo.pools {
		if time.Since(pool.lastCleanup) > mo.config.PoolCleanupInterval {
			// Simple cleanup: create new pool (Go's sync.Pool doesn't support direct cleanup)
			oldMetrics := pool.GetMetrics()
			pool.pool = sync.Pool{New: pool.pool.New}
			pool.lastCleanup = time.Now()

			pool.metrics.mu.Lock()
			pool.metrics.Cleanups++
			pool.metrics.mu.Unlock()

			mo.logger.Debug("Cleaned up object pool",
				zap.String("pool", name),
				zap.Int64("gets", oldMetrics.Gets),
				zap.Int64("puts", oldMetrics.Puts),
				zap.Float64("hit_ratio", oldMetrics.HitRatio),
			)
		}
	}
}

// GetOptimizationStats returns current optimization statistics
func (mo *MemoryOptimizer) GetOptimizationStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Memory stats
	stats["memory"] = mo.memoryMonitor.GetMemoryStats()

	// GC stats
	stats["gc"] = mo.gcController.GetGCStats()

	// Pool stats
	poolStats := make(map[string]PoolMetrics)
	mo.mu.RLock()
	for name, pool := range mo.pools {
		poolStats[name] = pool.GetMetrics()
	}
	mo.mu.RUnlock()
	stats["pools"] = poolStats

	return stats
}

// Close shuts down the memory optimizer
func (mo *MemoryOptimizer) Close() {
	// Cleanup resources
	mo.triggerOptimizations()
}
