package benchmarking

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	logger Logger
	config *BenchmarkConfig
	results map[string]*BenchmarkResult
	mutex   sync.RWMutex
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// BenchmarkConfig contains benchmark configuration
type BenchmarkConfig struct {
	Duration        time.Duration `json:"duration"`
	Concurrency     int           `json:"concurrency"`
	WarmupDuration  time.Duration `json:"warmup_duration"`
	CooldownDuration time.Duration `json:"cooldown_duration"`
	SampleInterval  time.Duration `json:"sample_interval"`
	EnableProfiling bool          `json:"enable_profiling"`
}

// BenchmarkResult contains benchmark results
type BenchmarkResult struct {
	Name            string        `json:"name"`
	Duration        time.Duration `json:"duration"`
	TotalOperations int64         `json:"total_operations"`
	OperationsPerSec float64      `json:"operations_per_sec"`
	AverageLatency  time.Duration `json:"average_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	P50Latency      time.Duration `json:"p50_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	ErrorCount      int64         `json:"error_count"`
	ErrorRate       float64       `json:"error_rate"`
	MemoryUsage     uint64        `json:"memory_usage"`
	CPUUsage        float64       `json:"cpu_usage"`
	GCCount         uint32        `json:"gc_count"`
	GCPauseTime     time.Duration `json:"gc_pause_time"`
	Timestamp       time.Time     `json:"timestamp"`
}

// BenchmarkFunction represents a function to benchmark
type BenchmarkFunction func(ctx context.Context) error

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(config *BenchmarkConfig, logger Logger) *BenchmarkSuite {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	return &BenchmarkSuite{
		logger:  logger,
		config:  config,
		results: make(map[string]*BenchmarkResult),
	}
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Duration:         30 * time.Second,
		Concurrency:      10,
		WarmupDuration:   5 * time.Second,
		CooldownDuration: 2 * time.Second,
		SampleInterval:   100 * time.Millisecond,
		EnableProfiling:  false,
	}
}

// RunBenchmark runs a benchmark for a specific function
func (bs *BenchmarkSuite) RunBenchmark(ctx context.Context, name string, fn BenchmarkFunction) (*BenchmarkResult, error) {
	bs.logger.Info("Starting benchmark", "name", name, "duration", bs.config.Duration, "concurrency", bs.config.Concurrency)

	// Warmup phase
	if bs.config.WarmupDuration > 0 {
		bs.logger.Debug("Starting warmup phase", "duration", bs.config.WarmupDuration)
		bs.runWarmup(ctx, fn)
	}

	// Collect initial memory stats
	var initialMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)

	// Prepare benchmark
	result := &BenchmarkResult{
		Name:        name,
		MinLatency:  time.Hour, // Initialize to high value
		Timestamp:   time.Now(),
	}

	// Channels for coordination
	startCh := make(chan struct{})
	doneCh := make(chan struct{})
	latencies := make(chan time.Duration, 10000)
	errors := make(chan error, 1000)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bs.config.Concurrency; i++ {
		wg.Add(1)
		go bs.benchmarkWorker(ctx, fn, startCh, doneCh, latencies, errors, &wg)
	}

	// Start benchmark
	benchmarkStart := time.Now()
	close(startCh)

	// Stop benchmark after duration
	go func() {
		time.Sleep(bs.config.Duration)
		close(doneCh)
	}()

	// Collect metrics during benchmark
	go bs.collectMetrics(ctx, result, doneCh)

	// Wait for all workers to complete
	wg.Wait()
	close(latencies)
	close(errors)

	// Calculate results
	result.Duration = time.Since(benchmarkStart)
	bs.calculateResults(result, latencies, errors)

	// Collect final memory stats
	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)
	result.MemoryUsage = finalMemStats.Alloc - initialMemStats.Alloc
	result.GCCount = finalMemStats.NumGC - initialMemStats.NumGC

	// Cooldown phase
	if bs.config.CooldownDuration > 0 {
		bs.logger.Debug("Starting cooldown phase", "duration", bs.config.CooldownDuration)
		time.Sleep(bs.config.CooldownDuration)
	}

	// Store result
	bs.mutex.Lock()
	bs.results[name] = result
	bs.mutex.Unlock()

	bs.logger.Info("Benchmark completed",
		"name", name,
		"operations_per_sec", result.OperationsPerSec,
		"average_latency", result.AverageLatency,
		"error_rate", result.ErrorRate)

	return result, nil
}

// benchmarkWorker runs benchmark operations in a worker goroutine
func (bs *BenchmarkSuite) benchmarkWorker(
	ctx context.Context,
	fn BenchmarkFunction,
	startCh, doneCh <-chan struct{},
	latencies chan<- time.Duration,
	errors chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Wait for start signal
	<-startCh

	for {
		select {
		case <-doneCh:
			return
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			err := fn(ctx)
			latency := time.Since(start)

			select {
			case latencies <- latency:
			default:
				// Channel full, skip this latency
			}

			if err != nil {
				select {
				case errors <- err:
				default:
					// Channel full, skip this error
				}
			}
		}
	}
}

// runWarmup runs warmup operations
func (bs *BenchmarkSuite) runWarmup(ctx context.Context, fn BenchmarkFunction) {
	warmupCtx, cancel := context.WithTimeout(ctx, bs.config.WarmupDuration)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < bs.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-warmupCtx.Done():
					return
				default:
					fn(warmupCtx)
				}
			}
		}()
	}

	wg.Wait()
}

// collectMetrics collects system metrics during benchmark
func (bs *BenchmarkSuite) collectMetrics(ctx context.Context, result *BenchmarkResult, doneCh <-chan struct{}) {
	ticker := time.NewTicker(bs.config.SampleInterval)
	defer ticker.Stop()

	var cpuSamples []float64
	var gcSamples []time.Duration

	for {
		select {
		case <-doneCh:
			// Calculate averages
			if len(cpuSamples) > 0 {
				var total float64
				for _, sample := range cpuSamples {
					total += sample
				}
				result.CPUUsage = total / float64(len(cpuSamples))
			}

			if len(gcSamples) > 0 {
				var total time.Duration
				for _, sample := range gcSamples {
					total += sample
				}
				result.GCPauseTime = total / time.Duration(len(gcSamples))
			}
			return

		case <-ticker.C:
			// Collect CPU usage (simplified)
			cpuSamples = append(cpuSamples, bs.getCPUUsage())

			// Collect GC metrics
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if len(m.PauseNs) > 0 {
				recentPause := m.PauseNs[(m.NumGC+255)%256]
				gcSamples = append(gcSamples, time.Duration(recentPause))
			}
		}
	}
}

// getCPUUsage returns current CPU usage (simplified implementation)
func (bs *BenchmarkSuite) getCPUUsage() float64 {
	// In a real implementation, this would measure actual CPU usage
	// For now, return a mock value
	return 45.0
}

// calculateResults calculates benchmark results from collected data
func (bs *BenchmarkSuite) calculateResults(result *BenchmarkResult, latencies <-chan time.Duration, errors <-chan error) {
	var latencySlice []time.Duration
	var totalLatency time.Duration

	// Collect all latencies
	for latency := range latencies {
		latencySlice = append(latencySlice, latency)
		totalLatency += latency
		result.TotalOperations++

		// Update min/max
		if latency < result.MinLatency {
			result.MinLatency = latency
		}
		if latency > result.MaxLatency {
			result.MaxLatency = latency
		}
	}

	// Count errors
	for range errors {
		result.ErrorCount++
	}

	// Calculate metrics
	if result.TotalOperations > 0 {
		result.AverageLatency = totalLatency / time.Duration(result.TotalOperations)
		result.OperationsPerSec = float64(result.TotalOperations) / result.Duration.Seconds()
		result.ErrorRate = float64(result.ErrorCount) / float64(result.TotalOperations) * 100

		// Calculate percentiles
		if len(latencySlice) > 0 {
			bs.sortDurations(latencySlice)
			result.P50Latency = bs.percentile(latencySlice, 50)
			result.P95Latency = bs.percentile(latencySlice, 95)
			result.P99Latency = bs.percentile(latencySlice, 99)
		}
	}
}

// sortDurations sorts a slice of durations
func (bs *BenchmarkSuite) sortDurations(durations []time.Duration) {
	// Simple bubble sort for demonstration
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

// percentile calculates the percentile value from sorted durations
func (bs *BenchmarkSuite) percentile(sortedDurations []time.Duration, percentile float64) time.Duration {
	if len(sortedDurations) == 0 {
		return 0
	}

	index := int(float64(len(sortedDurations)) * percentile / 100.0)
	if index >= len(sortedDurations) {
		index = len(sortedDurations) - 1
	}

	return sortedDurations[index]
}

// GetResults returns all benchmark results
func (bs *BenchmarkSuite) GetResults() map[string]*BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	results := make(map[string]*BenchmarkResult)
	for name, result := range bs.results {
		resultCopy := *result
		results[name] = &resultCopy
	}

	return results
}

// GetResult returns a specific benchmark result
func (bs *BenchmarkSuite) GetResult(name string) (*BenchmarkResult, bool) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	result, exists := bs.results[name]
	if !exists {
		return nil, false
	}

	resultCopy := *result
	return &resultCopy, true
}

// ClearResults clears all benchmark results
func (bs *BenchmarkSuite) ClearResults() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	bs.results = make(map[string]*BenchmarkResult)
	bs.logger.Info("Benchmark results cleared")
}

// GenerateReport generates a comprehensive benchmark report
func (bs *BenchmarkSuite) GenerateReport() *BenchmarkReport {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	report := &BenchmarkReport{
		Timestamp:    time.Now(),
		TotalBenchmarks: len(bs.results),
		Results:      make([]*BenchmarkResult, 0, len(bs.results)),
		Summary:      &BenchmarkSummary{},
	}

	var totalOps int64
	var totalDuration time.Duration
	var totalErrors int64

	for _, result := range bs.results {
		resultCopy := *result
		report.Results = append(report.Results, &resultCopy)

		totalOps += result.TotalOperations
		totalDuration += result.Duration
		totalErrors += result.ErrorCount
	}

	// Calculate summary
	if len(bs.results) > 0 {
		report.Summary.AverageOpsPerSec = float64(totalOps) / totalDuration.Seconds()
		report.Summary.TotalOperations = totalOps
		report.Summary.TotalErrors = totalErrors
		report.Summary.OverallErrorRate = float64(totalErrors) / float64(totalOps) * 100
	}

	return report
}

// BenchmarkReport represents a comprehensive benchmark report
type BenchmarkReport struct {
	Timestamp       time.Time          `json:"timestamp"`
	TotalBenchmarks int                `json:"total_benchmarks"`
	Results         []*BenchmarkResult `json:"results"`
	Summary         *BenchmarkSummary  `json:"summary"`
}

// BenchmarkSummary represents a summary of all benchmarks
type BenchmarkSummary struct {
	AverageOpsPerSec float64 `json:"average_ops_per_sec"`
	TotalOperations  int64   `json:"total_operations"`
	TotalErrors      int64   `json:"total_errors"`
	OverallErrorRate float64 `json:"overall_error_rate"`
}

// LoadTester provides load testing capabilities
type LoadTester struct {
	benchmarkSuite *BenchmarkSuite
	logger         Logger
}

// NewLoadTester creates a new load tester
func NewLoadTester(config *BenchmarkConfig, logger Logger) *LoadTester {
	return &LoadTester{
		benchmarkSuite: NewBenchmarkSuite(config, logger),
		logger:         logger,
	}
}

// RunLoadTest runs a load test with increasing concurrency
func (lt *LoadTester) RunLoadTest(ctx context.Context, name string, fn BenchmarkFunction, maxConcurrency int) ([]*BenchmarkResult, error) {
	lt.logger.Info("Starting load test", "name", name, "max_concurrency", maxConcurrency)

	var results []*BenchmarkResult
	
	// Test with increasing concurrency levels
	for concurrency := 1; concurrency <= maxConcurrency; concurrency *= 2 {
		testName := fmt.Sprintf("%s_concurrency_%d", name, concurrency)
		
		// Update config for this test
		lt.benchmarkSuite.config.Concurrency = concurrency
		
		result, err := lt.benchmarkSuite.RunBenchmark(ctx, testName, fn)
		if err != nil {
			return results, fmt.Errorf("load test failed at concurrency %d: %w", concurrency, err)
		}
		
		results = append(results, result)
		
		lt.logger.Info("Load test step completed",
			"concurrency", concurrency,
			"ops_per_sec", result.OperationsPerSec,
			"error_rate", result.ErrorRate)
	}

	return results, nil
}

// StressTester provides stress testing capabilities
type StressTester struct {
	benchmarkSuite *BenchmarkSuite
	logger         Logger
}

// NewStressTester creates a new stress tester
func NewStressTester(config *BenchmarkConfig, logger Logger) *StressTester {
	return &StressTester{
		benchmarkSuite: NewBenchmarkSuite(config, logger),
		logger:         logger,
	}
}

// RunStressTest runs a stress test to find breaking points
func (st *StressTester) RunStressTest(ctx context.Context, name string, fn BenchmarkFunction) (*StressTestResult, error) {
	st.logger.Info("Starting stress test", "name", name)

	result := &StressTestResult{
		Name:      name,
		Timestamp: time.Now(),
		Steps:     make([]*BenchmarkResult, 0),
	}

	concurrency := 1
	maxErrorRate := 10.0 // 10% error rate threshold

	for {
		testName := fmt.Sprintf("%s_stress_%d", name, concurrency)
		st.benchmarkSuite.config.Concurrency = concurrency

		benchResult, err := st.benchmarkSuite.RunBenchmark(ctx, testName, fn)
		if err != nil {
			return result, fmt.Errorf("stress test failed at concurrency %d: %w", concurrency, err)
		}

		result.Steps = append(result.Steps, benchResult)

		// Check if we've hit the breaking point
		if benchResult.ErrorRate > maxErrorRate {
			result.BreakingPoint = concurrency
			result.MaxOpsPerSec = benchResult.OperationsPerSec
			st.logger.Info("Stress test breaking point found",
				"concurrency", concurrency,
				"error_rate", benchResult.ErrorRate)
			break
		}

		// Update max ops per sec
		if benchResult.OperationsPerSec > result.MaxOpsPerSec {
			result.MaxOpsPerSec = benchResult.OperationsPerSec
		}

		// Increase concurrency
		concurrency += 10
		if concurrency > 1000 { // Safety limit
			st.logger.Warn("Stress test reached safety limit", "max_concurrency", 1000)
			break
		}
	}

	return result, nil
}

// StressTestResult represents stress test results
type StressTestResult struct {
	Name          string             `json:"name"`
	Timestamp     time.Time          `json:"timestamp"`
	BreakingPoint int                `json:"breaking_point"`
	MaxOpsPerSec  float64            `json:"max_ops_per_sec"`
	Steps         []*BenchmarkResult `json:"steps"`
}
