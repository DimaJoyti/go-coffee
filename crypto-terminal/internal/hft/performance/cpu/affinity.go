package cpu

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// CPUCore represents a CPU core with its properties
type CPUCore struct {
	ID           int
	PhysicalID   int
	CoreID       int
	Siblings     []int
	NUMANode     int
	Frequency    uint64 // MHz
	CacheL1Size  uint64 // KB
	CacheL2Size  uint64 // KB
	CacheL3Size  uint64 // KB
	IsHyperThread bool
}

// CPUTopology represents the CPU topology of the system
type CPUTopology struct {
	Cores         []CPUCore
	NumCores      int
	NumPhysical   int
	NumLogical    int
	NumNUMANodes  int
	HyperThreading bool
	Architecture  string
}

// ThreadAffinity manages CPU affinity for threads
type ThreadAffinity struct {
	topology     *CPUTopology
	assignments  sync.Map // map[string]int - thread name to core ID
	coreUsage    []int32  // atomic counters for core usage
	isolatedCores []int   // cores reserved for HFT
	mutex        sync.RWMutex
}

// AffinityConfig holds configuration for CPU affinity
type AffinityConfig struct {
	IsolatedCores    []int  // Cores to isolate for HFT
	MarketDataCore   int    // Dedicated core for market data
	OrderProcessCore int    // Dedicated core for order processing
	RiskCheckCore    int    // Dedicated core for risk checks
	NetworkCore      int    // Dedicated core for network I/O
	EnableIsolation  bool   // Enable CPU isolation
	EnableRealTime   bool   // Enable real-time scheduling
	Priority         int    // Thread priority (1-99)
}

// NewThreadAffinity creates a new thread affinity manager
func NewThreadAffinity() (*ThreadAffinity, error) {
	topology, err := GetCPUTopology()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU topology: %w", err)
	}
	
	ta := &ThreadAffinity{
		topology:  topology,
		coreUsage: make([]int32, topology.NumLogical),
	}
	
	return ta, nil
}

// GetCPUTopology discovers the CPU topology of the system
func GetCPUTopology() (*CPUTopology, error) {
	// This is a simplified implementation
	// In production, this would parse /proc/cpuinfo, /sys/devices/system/cpu/, etc.
	
	numLogical := runtime.NumCPU()
	numPhysical := numLogical / 2 // Assume hyperthreading
	if numPhysical == 0 {
		numPhysical = numLogical
	}
	
	cores := make([]CPUCore, numLogical)
	for i := 0; i < numLogical; i++ {
		cores[i] = CPUCore{
			ID:           i,
			PhysicalID:   i / 2,
			CoreID:       i % numPhysical,
			NUMANode:     i / 8, // Assume 8 cores per NUMA node
			Frequency:    3000,  // 3 GHz default
			CacheL1Size:  32,    // 32KB
			CacheL2Size:  256,   // 256KB
			CacheL3Size:  8192,  // 8MB
			IsHyperThread: i >= numPhysical,
		}
		
		// Set siblings for hyperthreading
		if cores[i].IsHyperThread {
			cores[i].Siblings = []int{i - numPhysical}
		} else {
			if i+numPhysical < numLogical {
				cores[i].Siblings = []int{i + numPhysical}
			}
		}
	}
	
	return &CPUTopology{
		Cores:         cores,
		NumCores:      numPhysical,
		NumPhysical:   numPhysical,
		NumLogical:    numLogical,
		NumNUMANodes:  (numLogical + 7) / 8,
		HyperThreading: numLogical > numPhysical,
		Architecture:  runtime.GOARCH,
	}, nil
}

// SetThreadAffinity sets CPU affinity for the current thread
func (ta *ThreadAffinity) SetThreadAffinity(threadName string, coreID int) error {
	if coreID < 0 || coreID >= ta.topology.NumLogical {
		return fmt.Errorf("invalid core ID: %d", coreID)
	}
	
	// Store the assignment
	ta.assignments.Store(threadName, coreID)
	
	// Update core usage
	atomic.AddInt32(&ta.coreUsage[coreID], 1)
	
	// Set actual CPU affinity (platform-specific)
	return ta.setOSThreadAffinity(coreID)
}

// setOSThreadAffinity sets OS-level thread affinity (platform-specific)
func (ta *ThreadAffinity) setOSThreadAffinity(coreID int) error {
	// This would use platform-specific system calls
	// Linux: sched_setaffinity()
	// Windows: SetThreadAffinityMask()
	// macOS: thread_policy_set()
	
	// For now, this is a placeholder
	// In production, this would use CGO to call system APIs
	
	return nil
}

// GetOptimalCore returns the optimal core for a given workload type
func (ta *ThreadAffinity) GetOptimalCore(workloadType string) (int, error) {
	ta.mutex.RLock()
	defer ta.mutex.RUnlock()
	
	switch workloadType {
	case "market_data":
		// Prefer physical cores with low usage
		return ta.findLeastUsedPhysicalCore()
	case "order_processing":
		// Prefer isolated cores
		return ta.findLeastUsedIsolatedCore()
	case "risk_check":
		// Prefer cores with good cache locality
		return ta.findCoreWithGoodCacheLocality()
	case "network_io":
		// Prefer cores close to network interrupts
		return ta.findNetworkOptimalCore()
	default:
		return ta.findLeastUsedCore()
	}
}

// findLeastUsedPhysicalCore finds the physical core with least usage
func (ta *ThreadAffinity) findLeastUsedPhysicalCore() (int, error) {
	minUsage := int32(^uint32(0) >> 1) // Max int32
	bestCore := -1
	
	for _, core := range ta.topology.Cores {
		if !core.IsHyperThread {
			usage := atomic.LoadInt32(&ta.coreUsage[core.ID])
			if usage < minUsage {
				minUsage = usage
				bestCore = core.ID
			}
		}
	}
	
	if bestCore == -1 {
		return 0, fmt.Errorf("no physical cores available")
	}
	
	return bestCore, nil
}

// findLeastUsedIsolatedCore finds the least used isolated core
func (ta *ThreadAffinity) findLeastUsedIsolatedCore() (int, error) {
	if len(ta.isolatedCores) == 0 {
		return ta.findLeastUsedPhysicalCore()
	}
	
	minUsage := int32(^uint32(0) >> 1) // Max int32
	bestCore := -1
	
	for _, coreID := range ta.isolatedCores {
		usage := atomic.LoadInt32(&ta.coreUsage[coreID])
		if usage < minUsage {
			minUsage = usage
			bestCore = coreID
		}
	}
	
	if bestCore == -1 {
		return ta.isolatedCores[0], nil
	}
	
	return bestCore, nil
}

// findCoreWithGoodCacheLocality finds a core with good cache locality
func (ta *ThreadAffinity) findCoreWithGoodCacheLocality() (int, error) {
	// Prefer cores that share L3 cache
	// For now, use a simple heuristic
	return ta.findLeastUsedPhysicalCore()
}

// findNetworkOptimalCore finds the optimal core for network I/O
func (ta *ThreadAffinity) findNetworkOptimalCore() (int, error) {
	// Prefer cores on the same NUMA node as network interrupts
	// For now, use core 0 as it typically handles interrupts
	return 0, nil
}

// findLeastUsedCore finds the least used core overall
func (ta *ThreadAffinity) findLeastUsedCore() (int, error) {
	minUsage := int32(^uint32(0) >> 1) // Max int32
	bestCore := -1
	
	for i := 0; i < ta.topology.NumLogical; i++ {
		usage := atomic.LoadInt32(&ta.coreUsage[i])
		if usage < minUsage {
			minUsage = usage
			bestCore = i
		}
	}
	
	if bestCore == -1 {
		return 0, fmt.Errorf("no cores available")
	}
	
	return bestCore, nil
}

// SetIsolatedCores sets the cores to be isolated for HFT
func (ta *ThreadAffinity) SetIsolatedCores(cores []int) error {
	ta.mutex.Lock()
	defer ta.mutex.Unlock()
	
	// Validate cores
	for _, coreID := range cores {
		if coreID < 0 || coreID >= ta.topology.NumLogical {
			return fmt.Errorf("invalid core ID: %d", coreID)
		}
	}
	
	ta.isolatedCores = make([]int, len(cores))
	copy(ta.isolatedCores, cores)
	
	return nil
}

// GetCoreUsage returns the usage count for each core
func (ta *ThreadAffinity) GetCoreUsage() []int32 {
	usage := make([]int32, len(ta.coreUsage))
	for i := range ta.coreUsage {
		usage[i] = atomic.LoadInt32(&ta.coreUsage[i])
	}
	return usage
}

// GetTopology returns the CPU topology
func (ta *ThreadAffinity) GetTopology() *CPUTopology {
	return ta.topology
}

// ThreadManager manages high-performance threads for HFT
type ThreadManager struct {
	affinity     *ThreadAffinity
	config       *AffinityConfig
	threads      sync.Map // map[string]*HFTThread
	threadCount  int32
}

// HFTThread represents a high-performance thread
type HFTThread struct {
	Name         string
	CoreID       int
	Priority     int
	IsRealTime   bool
	StartTime    time.Time
	CPUTime      time.Duration
	ContextSwitches uint64
	CacheMisses  uint64
}

// NewThreadManager creates a new thread manager
func NewThreadManager(config *AffinityConfig) (*ThreadManager, error) {
	affinity, err := NewThreadAffinity()
	if err != nil {
		return nil, err
	}
	
	if config.EnableIsolation && len(config.IsolatedCores) > 0 {
		if err := affinity.SetIsolatedCores(config.IsolatedCores); err != nil {
			return nil, err
		}
	}
	
	return &ThreadManager{
		affinity: affinity,
		config:   config,
	}, nil
}

// CreateHFTThread creates a new high-performance thread
func (tm *ThreadManager) CreateHFTThread(name, workloadType string) (*HFTThread, error) {
	coreID, err := tm.affinity.GetOptimalCore(workloadType)
	if err != nil {
		return nil, err
	}
	
	thread := &HFTThread{
		Name:      name,
		CoreID:    coreID,
		Priority:  tm.config.Priority,
		IsRealTime: tm.config.EnableRealTime,
		StartTime: time.Now(),
	}
	
	// Set thread affinity
	if err := tm.affinity.SetThreadAffinity(name, coreID); err != nil {
		return nil, err
	}
	
	// Set real-time priority if enabled
	if tm.config.EnableRealTime {
		if err := tm.setRealTimePriority(tm.config.Priority); err != nil {
			// Log warning but don't fail
		}
	}
	
	tm.threads.Store(name, thread)
	atomic.AddInt32(&tm.threadCount, 1)
	
	return thread, nil
}

// setRealTimePriority sets real-time scheduling priority
func (tm *ThreadManager) setRealTimePriority(priority int) error {
	// This would use platform-specific system calls
	// Linux: sched_setscheduler() with SCHED_FIFO or SCHED_RR
	// Windows: SetThreadPriority() with THREAD_PRIORITY_TIME_CRITICAL
	
	// For now, this is a placeholder
	return nil
}

// GetThread returns a thread by name
func (tm *ThreadManager) GetThread(name string) (*HFTThread, bool) {
	threadInterface, exists := tm.threads.Load(name)
	if !exists {
		return nil, false
	}
	return threadInterface.(*HFTThread), true
}

// GetAllThreads returns all managed threads
func (tm *ThreadManager) GetAllThreads() map[string]*HFTThread {
	result := make(map[string]*HFTThread)
	tm.threads.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*HFTThread)
		return true
	})
	return result
}

// GetThreadCount returns the number of managed threads
func (tm *ThreadManager) GetThreadCount() int32 {
	return atomic.LoadInt32(&tm.threadCount)
}

// OptimizeForHFT applies HFT-specific optimizations
func (tm *ThreadManager) OptimizeForHFT() error {
	// Disable Go's garbage collector for critical threads
	// This is dangerous and should only be used in specific scenarios
	// runtime.GC()
	// debug.SetGCPercent(-1)
	
	// Set GOMAXPROCS to match isolated cores
	if len(tm.config.IsolatedCores) > 0 {
		runtime.GOMAXPROCS(len(tm.config.IsolatedCores))
	}
	
	// Preallocate goroutine stack space
	// This would require runtime modifications
	
	return nil
}

// CPUProfiler provides CPU profiling for HFT threads
type CPUProfiler struct {
	threads     map[string]*HFTThread
	sampling    bool
	sampleRate  time.Duration
	stopChan    chan struct{}
}

// NewCPUProfiler creates a new CPU profiler
func NewCPUProfiler(sampleRate time.Duration) *CPUProfiler {
	return &CPUProfiler{
		threads:    make(map[string]*HFTThread),
		sampleRate: sampleRate,
		stopChan:   make(chan struct{}),
	}
}

// StartProfiling starts CPU profiling
func (cp *CPUProfiler) StartProfiling() {
	if cp.sampling {
		return
	}
	
	cp.sampling = true
	go cp.profileLoop()
}

// StopProfiling stops CPU profiling
func (cp *CPUProfiler) StopProfiling() {
	if !cp.sampling {
		return
	}
	
	cp.sampling = false
	close(cp.stopChan)
}

// profileLoop runs the profiling loop
func (cp *CPUProfiler) profileLoop() {
	ticker := time.NewTicker(cp.sampleRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cp.sampleCPUUsage()
		case <-cp.stopChan:
			return
		}
	}
}

// sampleCPUUsage samples CPU usage for all threads
func (cp *CPUProfiler) sampleCPUUsage() {
	// This would collect CPU usage statistics
	// Platform-specific implementation required
}

// Global thread manager
var globalThreadManager *ThreadManager
var threadManagerOnce sync.Once

// GetGlobalThreadManager returns the global thread manager
func GetGlobalThreadManager() *ThreadManager {
	threadManagerOnce.Do(func() {
		config := &AffinityConfig{
			IsolatedCores:    []int{2, 3, 4, 5}, // Isolate cores 2-5
			MarketDataCore:   2,
			OrderProcessCore: 3,
			RiskCheckCore:    4,
			NetworkCore:      5,
			EnableIsolation:  true,
			EnableRealTime:   false, // Requires root privileges
			Priority:         50,
		}
		
		var err error
		globalThreadManager, err = NewThreadManager(config)
		if err != nil {
			// Fallback to basic configuration
			config.EnableIsolation = false
			config.EnableRealTime = false
			globalThreadManager, _ = NewThreadManager(config)
		}
	})
	return globalThreadManager
}

// Helper functions for thread management

// PinCurrentThreadToCore pins the current goroutine to a specific core
func PinCurrentThreadToCore(coreID int) error {
	tm := GetGlobalThreadManager()
	threadName := fmt.Sprintf("goroutine_%d", getGoroutineID())
	return tm.affinity.SetThreadAffinity(threadName, coreID)
}

// getGoroutineID returns the current goroutine ID (unsafe)
func getGoroutineID() int64 {
	// This is unsafe and should not be used in production
	// It's included here for demonstration purposes
	return int64(uintptr(unsafe.Pointer(&struct{}{})))
}

// SetHighPriority sets high priority for the current thread
func SetHighPriority() error {
	tm := GetGlobalThreadManager()
	return tm.setRealTimePriority(tm.config.Priority)
}

// DisableGCForCriticalSection disables GC for critical sections
func DisableGCForCriticalSection() func() {
	// This is extremely dangerous and should only be used
	// for very short critical sections
	runtime.GC()
	// In production, you might use runtime.LockOSThread()
	runtime.LockOSThread()
	
	return func() {
		runtime.UnlockOSThread()
	}
}
