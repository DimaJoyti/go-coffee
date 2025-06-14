package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// FaultInjector provides chaos engineering capabilities
type FaultInjector struct {
	logger    *zap.Logger
	config    *ChaosConfig
	scenarios map[string]*ChaosScenario
	active    int32 // atomic boolean
	mu        sync.RWMutex
}

// ChaosConfig contains chaos engineering configuration
type ChaosConfig struct {
	Enabled             bool                     `json:"enabled"`
	GlobalFailureRate   float64                  `json:"global_failure_rate"`   // 0.0 to 1.0
	Scenarios           map[string]*ScenarioConfig `json:"scenarios"`
	SafeMode            bool                     `json:"safe_mode"`             // Prevents destructive operations
	MaxConcurrentFaults int                      `json:"max_concurrent_faults"`
	MonitoringInterval  time.Duration            `json:"monitoring_interval"`
}

// ScenarioConfig defines a chaos scenario
type ScenarioConfig struct {
	Name            string        `json:"name"`
	Enabled         bool          `json:"enabled"`
	FailureRate     float64       `json:"failure_rate"`     // 0.0 to 1.0
	Duration        time.Duration `json:"duration"`         // How long to run
	Interval        time.Duration `json:"interval"`         // How often to trigger
	TargetEndpoints []string      `json:"target_endpoints"` // Specific endpoints to target
	FaultType       string        `json:"fault_type"`       // "latency", "error", "timeout", "memory", "cpu"
	Parameters      map[string]interface{} `json:"parameters"` // Fault-specific parameters
}

// ChaosScenario represents an active chaos scenario
type ChaosScenario struct {
	config      *ScenarioConfig
	injector    *FaultInjector
	active      int32 // atomic boolean
	startTime   time.Time
	faultCount  int64
	lastFault   time.Time
	mu          sync.RWMutex
}

// FaultType represents different types of faults
type FaultType string

const (
	FaultTypeLatency  FaultType = "latency"
	FaultTypeError    FaultType = "error"
	FaultTypeTimeout  FaultType = "timeout"
	FaultTypeMemory   FaultType = "memory"
	FaultTypeCPU      FaultType = "cpu"
	FaultTypeNetwork  FaultType = "network"
	FaultTypeDatabase FaultType = "database"
	FaultTypeCache    FaultType = "cache"
)

// ChaosMetrics tracks chaos engineering metrics
type ChaosMetrics struct {
	ActiveScenarios   int                        `json:"active_scenarios"`
	TotalFaults       int64                      `json:"total_faults"`
	FaultsByType      map[string]int64           `json:"faults_by_type"`
	FaultsByScenario  map[string]int64           `json:"faults_by_scenario"`
	LastFaultTime     time.Time                  `json:"last_fault_time"`
	ScenarioMetrics   map[string]*ScenarioMetrics `json:"scenario_metrics"`
}

// ScenarioMetrics tracks metrics for individual scenarios
type ScenarioMetrics struct {
	Name         string        `json:"name"`
	Active       bool          `json:"active"`
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration"`
	FaultCount   int64         `json:"fault_count"`
	LastFault    time.Time     `json:"last_fault"`
	FailureRate  float64       `json:"failure_rate"`
}

// NewFaultInjector creates a new fault injector
func NewFaultInjector(config *ChaosConfig, logger *zap.Logger) *FaultInjector {
	fi := &FaultInjector{
		logger:    logger,
		config:    config,
		scenarios: make(map[string]*ChaosScenario),
	}

	// Initialize scenarios
	for name, scenarioConfig := range config.Scenarios {
		scenario := &ChaosScenario{
			config:   scenarioConfig,
			injector: fi,
		}
		fi.scenarios[name] = scenario
	}

	return fi
}

// Start starts the fault injector
func (fi *FaultInjector) Start() error {
	if !fi.config.Enabled {
		fi.logger.Info("Chaos engineering is disabled")
		return nil
	}

	if !atomic.CompareAndSwapInt32(&fi.active, 0, 1) {
		return fmt.Errorf("fault injector is already running")
	}

	fi.logger.Info("Starting chaos engineering fault injector",
		zap.Bool("safe_mode", fi.config.SafeMode),
		zap.Float64("global_failure_rate", fi.config.GlobalFailureRate))

	// Start enabled scenarios
	for name, scenario := range fi.scenarios {
		if scenario.config.Enabled {
			go scenario.start()
			fi.logger.Info("Started chaos scenario", zap.String("scenario", name))
		}
	}

	// Start monitoring
	go fi.monitor()

	return nil
}

// Stop stops the fault injector
func (fi *FaultInjector) Stop() error {
	if !atomic.CompareAndSwapInt32(&fi.active, 1, 0) {
		return nil
	}

	fi.logger.Info("Stopping chaos engineering fault injector")

	// Stop all scenarios
	for name, scenario := range fi.scenarios {
		scenario.stop()
		fi.logger.Info("Stopped chaos scenario", zap.String("scenario", name))
	}

	return nil
}

// HTTPMiddleware returns HTTP middleware for fault injection
func (fi *FaultInjector) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !fi.config.Enabled || atomic.LoadInt32(&fi.active) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Check if we should inject a fault
			if fi.shouldInjectFault(r) {
				fault := fi.selectFault(r)
				if fault != nil {
					fi.injectHTTPFault(w, r, fault)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// shouldInjectFault determines if a fault should be injected
func (fi *FaultInjector) shouldInjectFault(r *http.Request) bool {
	// Check global failure rate
	if rand.Float64() > fi.config.GlobalFailureRate {
		return false
	}

	// Check if any scenario targets this endpoint
	for _, scenario := range fi.scenarios {
		if atomic.LoadInt32(&scenario.active) == 1 && scenario.targetsEndpoint(r.URL.Path) {
			if rand.Float64() <= scenario.config.FailureRate {
				return true
			}
		}
	}

	return false
}

// selectFault selects an appropriate fault for the request
func (fi *FaultInjector) selectFault(r *http.Request) *ChaosScenario {
	var candidates []*ChaosScenario

	for _, scenario := range fi.scenarios {
		if atomic.LoadInt32(&scenario.active) == 1 && scenario.targetsEndpoint(r.URL.Path) {
			candidates = append(candidates, scenario)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Select random candidate
	return candidates[rand.Intn(len(candidates))]
}

// injectHTTPFault injects a fault into HTTP response
func (fi *FaultInjector) injectHTTPFault(w http.ResponseWriter, r *http.Request, scenario *ChaosScenario) {
	atomic.AddInt64(&scenario.faultCount, 1)
	scenario.mu.Lock()
	scenario.lastFault = time.Now()
	scenario.mu.Unlock()

	fi.logger.Debug("Injecting chaos fault",
		zap.String("scenario", scenario.config.Name),
		zap.String("fault_type", scenario.config.FaultType),
		zap.String("endpoint", r.URL.Path))

	switch FaultType(scenario.config.FaultType) {
	case FaultTypeLatency:
		fi.injectLatencyFault(w, r, scenario)
	case FaultTypeError:
		fi.injectErrorFault(w, r, scenario)
	case FaultTypeTimeout:
		fi.injectTimeoutFault(w, r, scenario)
	default:
		fi.logger.Warn("Unknown fault type", zap.String("fault_type", scenario.config.FaultType))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// injectLatencyFault injects artificial latency
func (fi *FaultInjector) injectLatencyFault(w http.ResponseWriter, r *http.Request, scenario *ChaosScenario) {
	// Get latency parameters
	minLatency := fi.getParameterDuration(scenario, "min_latency", 100*time.Millisecond)
	maxLatency := fi.getParameterDuration(scenario, "max_latency", 2*time.Second)

	// Calculate random latency
	latencyRange := maxLatency - minLatency
	latency := minLatency + time.Duration(rand.Int63n(int64(latencyRange)))

	fi.logger.Debug("Injecting latency fault", 
		zap.Duration("latency", latency),
		zap.String("endpoint", r.URL.Path))

	// Sleep to simulate latency
	time.Sleep(latency)

	// Return error or continue with high latency
	if fi.getParameterBool(scenario, "return_error", false) {
		http.Error(w, "Request Timeout", http.StatusRequestTimeout)
	} else {
		w.Header().Set("X-Chaos-Latency", latency.String())
		http.Error(w, "Slow Response", http.StatusOK)
	}
}

// injectErrorFault injects HTTP errors
func (fi *FaultInjector) injectErrorFault(w http.ResponseWriter, r *http.Request, scenario *ChaosScenario) {
	// Get error parameters
	statusCode := fi.getParameterInt(scenario, "status_code", http.StatusInternalServerError)
	errorMessage := fi.getParameterString(scenario, "error_message", "Chaos Engineering Fault Injection")

	fi.logger.Debug("Injecting error fault",
		zap.Int("status_code", statusCode),
		zap.String("endpoint", r.URL.Path))

	w.Header().Set("X-Chaos-Error", "true")
	http.Error(w, errorMessage, statusCode)
}

// injectTimeoutFault injects timeout behavior
func (fi *FaultInjector) injectTimeoutFault(w http.ResponseWriter, r *http.Request, scenario *ChaosScenario) {
	timeout := fi.getParameterDuration(scenario, "timeout", 30*time.Second)

	fi.logger.Debug("Injecting timeout fault",
		zap.Duration("timeout", timeout),
		zap.String("endpoint", r.URL.Path))

	// Sleep for timeout duration then return timeout error
	time.Sleep(timeout)
	w.Header().Set("X-Chaos-Timeout", "true")
	http.Error(w, "Request Timeout", http.StatusRequestTimeout)
}

// InjectMemoryPressure injects memory pressure
func (fi *FaultInjector) InjectMemoryPressure(ctx context.Context, scenario *ChaosScenario) {
	if fi.config.SafeMode {
		fi.logger.Warn("Memory pressure injection skipped (safe mode enabled)")
		return
	}

	size := fi.getParameterInt(scenario, "memory_size", 100*1024*1024) // 100MB default
	duration := fi.getParameterDuration(scenario, "duration", 30*time.Second)

	fi.logger.Info("Injecting memory pressure",
		zap.Int("size_bytes", size),
		zap.Duration("duration", duration))

	// Allocate memory
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Hold memory for duration
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}

	// Release memory
	data = nil
	runtime.GC()

	fi.logger.Info("Memory pressure injection completed")
}

// InjectCPUStress injects CPU stress
func (fi *FaultInjector) InjectCPUStress(ctx context.Context, scenario *ChaosScenario) {
	if fi.config.SafeMode {
		fi.logger.Warn("CPU stress injection skipped (safe mode enabled)")
		return
	}

	workers := fi.getParameterInt(scenario, "cpu_workers", runtime.NumCPU())
	duration := fi.getParameterDuration(scenario, "duration", 30*time.Second)

	fi.logger.Info("Injecting CPU stress",
		zap.Int("workers", workers),
		zap.Duration("duration", duration))

	// Start CPU stress workers
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					// Busy loop to consume CPU
					for j := 0; j < 1000000; j++ {
						_ = j * j
					}
				}
			}
		}()
	}

	// Run for specified duration
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}

	// Stop workers
	close(stopCh)
	wg.Wait()

	fi.logger.Info("CPU stress injection completed")
}

// Scenario methods

// start starts a chaos scenario
func (cs *ChaosScenario) start() {
	if !atomic.CompareAndSwapInt32(&cs.active, 0, 1) {
		return
	}

	cs.startTime = time.Now()

	// Run scenario for specified duration
	if cs.config.Duration > 0 {
		time.AfterFunc(cs.config.Duration, func() {
			cs.stop()
		})
	}
}

// stop stops a chaos scenario
func (cs *ChaosScenario) stop() {
	atomic.StoreInt32(&cs.active, 0)
}

// targetsEndpoint checks if scenario targets a specific endpoint
func (cs *ChaosScenario) targetsEndpoint(endpoint string) bool {
	if len(cs.config.TargetEndpoints) == 0 {
		return true // Target all endpoints if none specified
	}

	for _, target := range cs.config.TargetEndpoints {
		if target == endpoint || target == "*" {
			return true
		}
	}

	return false
}

// Helper methods for parameter extraction

func (fi *FaultInjector) getParameterDuration(scenario *ChaosScenario, key string, defaultValue time.Duration) time.Duration {
	if val, ok := scenario.config.Parameters[key]; ok {
		if str, ok := val.(string); ok {
			if duration, err := time.ParseDuration(str); err == nil {
				return duration
			}
		}
	}
	return defaultValue
}

func (fi *FaultInjector) getParameterInt(scenario *ChaosScenario, key string, defaultValue int) int {
	if val, ok := scenario.config.Parameters[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

func (fi *FaultInjector) getParameterString(scenario *ChaosScenario, key string, defaultValue string) string {
	if val, ok := scenario.config.Parameters[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (fi *FaultInjector) getParameterBool(scenario *ChaosScenario, key string, defaultValue bool) bool {
	if val, ok := scenario.config.Parameters[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// monitor runs periodic monitoring
func (fi *FaultInjector) monitor() {
	ticker := time.NewTicker(fi.config.MonitoringInterval)
	defer ticker.Stop()

	for atomic.LoadInt32(&fi.active) == 1 {
		select {
		case <-ticker.C:
			fi.logMetrics()
		}
	}
}

// logMetrics logs current chaos metrics
func (fi *FaultInjector) logMetrics() {
	metrics := fi.GetMetrics()
	
	fi.logger.Info("Chaos engineering metrics",
		zap.Int("active_scenarios", metrics.ActiveScenarios),
		zap.Int64("total_faults", metrics.TotalFaults),
		zap.Time("last_fault", metrics.LastFaultTime))
}

// GetMetrics returns current chaos metrics
func (fi *FaultInjector) GetMetrics() *ChaosMetrics {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	metrics := &ChaosMetrics{
		FaultsByType:     make(map[string]int64),
		FaultsByScenario: make(map[string]int64),
		ScenarioMetrics:  make(map[string]*ScenarioMetrics),
	}

	var totalFaults int64
	var lastFaultTime time.Time

	for name, scenario := range fi.scenarios {
		isActive := atomic.LoadInt32(&scenario.active) == 1
		if isActive {
			metrics.ActiveScenarios++
		}

		faultCount := atomic.LoadInt64(&scenario.faultCount)
		totalFaults += faultCount

		scenario.mu.RLock()
		if scenario.lastFault.After(lastFaultTime) {
			lastFaultTime = scenario.lastFault
		}
		scenario.mu.RUnlock()

		metrics.FaultsByScenario[name] = faultCount
		metrics.FaultsByType[scenario.config.FaultType] += faultCount

		metrics.ScenarioMetrics[name] = &ScenarioMetrics{
			Name:        name,
			Active:      isActive,
			StartTime:   scenario.startTime,
			Duration:    time.Since(scenario.startTime),
			FaultCount:  faultCount,
			LastFault:   scenario.lastFault,
			FailureRate: scenario.config.FailureRate,
		}
	}

	metrics.TotalFaults = totalFaults
	metrics.LastFaultTime = lastFaultTime

	return metrics
}
