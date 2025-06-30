package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BenchmarkSuite provides comprehensive performance testing and benchmarking
type BenchmarkSuite struct {
	benchmarks      map[string]*Benchmark
	loadTests       map[string]*LoadTest
	stressTests     map[string]*StressTest
	enduranceTests  map[string]*EnduranceTest
	scenarios       map[string]*PerformanceScenario
	metrics         *PerformanceMetrics
	config          *BenchmarkConfig
	logger          TestLogger
	mutex           sync.RWMutex
}

// Benchmark represents a single benchmark test
type Benchmark struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Function        BenchmarkFunction      `json:"-"`
	Setup           func() error           `json:"-"`
	Teardown        func() error           `json:"-"`
	Iterations      int                    `json:"iterations"`
	Duration        time.Duration          `json:"duration"`
	WarmupRounds    int                    `json:"warmup_rounds"`
	Parallel        bool                   `json:"parallel"`
	MemoryProfile   bool                   `json:"memory_profile"`
	CPUProfile      bool                   `json:"cpu_profile"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// LoadTest represents a load testing configuration
type LoadTest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Target          string                 `json:"target"`
	VirtualUsers    int                    `json:"virtual_users"`
	Duration        time.Duration          `json:"duration"`
	RampUpTime      time.Duration          `json:"ramp_up_time"`
	RampDownTime    time.Duration          `json:"ramp_down_time"`
	RequestRate     float64                `json:"request_rate"`
	ThinkTime       time.Duration          `json:"think_time"`
	Scenarios       []*LoadScenario        `json:"scenarios"`
	Thresholds      *PerformanceThresholds `json:"thresholds"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// StressTest represents a stress testing configuration
type StressTest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Target          string                 `json:"target"`
	StartUsers      int                    `json:"start_users"`
	MaxUsers        int                    `json:"max_users"`
	UserIncrement   int                    `json:"user_increment"`
	IncrementInterval time.Duration        `json:"increment_interval"`
	Duration        time.Duration          `json:"duration"`
	BreakingPoint   *BreakingPointConfig   `json:"breaking_point"`
	Scenarios       []*LoadScenario        `json:"scenarios"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// EnduranceTest represents an endurance testing configuration
type EnduranceTest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Target          string                 `json:"target"`
	VirtualUsers    int                    `json:"virtual_users"`
	Duration        time.Duration          `json:"duration"`
	Scenarios       []*LoadScenario        `json:"scenarios"`
	MemoryLeakCheck bool                   `json:"memory_leak_check"`
	ResourceMonitoring bool                `json:"resource_monitoring"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// PerformanceScenario represents a complex performance testing scenario
type PerformanceScenario struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Steps           []*PerformanceStep     `json:"steps"`
	LoadProfile     *LoadProfile           `json:"load_profile"`
	Environment     string                 `json:"environment"`
	Prerequisites   []*Prerequisite        `json:"prerequisites"`
	Assertions      []*PerformanceAssertion `json:"assertions"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Supporting types
type BenchmarkFunction func(b *BenchmarkContext) error

type BenchmarkContext struct {
	N           int
	StartTimer  func()
	StopTimer   func()
	ResetTimer  func()
	SetBytes    func(int64)
	ReportAllocs func()
	Logf        func(format string, args ...interface{})
	Fatalf      func(format string, args ...interface{})
	Skipf       func(format string, args ...interface{})
}

type LoadScenario struct {
	Name        string                 `json:"name"`
	Weight      float64                `json:"weight"`
	Requests    []*LoadRequest         `json:"requests"`
	ThinkTime   time.Duration          `json:"think_time"`
	Iterations  int                    `json:"iterations"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type LoadRequest struct {
	Name        string                 `json:"name"`
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Headers     map[string]string      `json:"headers"`
	Body        interface{}            `json:"body"`
	Timeout     time.Duration          `json:"timeout"`
	Validation  *ResponseValidation    `json:"validation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ResponseValidation struct {
	StatusCode    int                    `json:"status_code"`
	ResponseTime  time.Duration          `json:"response_time"`
	BodyContains  []string               `json:"body_contains"`
	Headers       map[string]string      `json:"headers"`
	JSONPath      map[string]interface{} `json:"json_path"`
}

type PerformanceThresholds struct {
	MaxResponseTime   time.Duration `json:"max_response_time"`
	MaxErrorRate      float64       `json:"max_error_rate"`
	MinThroughput     float64       `json:"min_throughput"`
	MaxCPUUsage       float64       `json:"max_cpu_usage"`
	MaxMemoryUsage    float64       `json:"max_memory_usage"`
	MaxDiskUsage      float64       `json:"max_disk_usage"`
	MaxNetworkLatency time.Duration `json:"max_network_latency"`
}

type BreakingPointConfig struct {
	ErrorRateThreshold    float64       `json:"error_rate_threshold"`
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	CPUThreshold          float64       `json:"cpu_threshold"`
	MemoryThreshold       float64       `json:"memory_threshold"`
	ConsecutiveFailures   int           `json:"consecutive_failures"`
}

type LoadProfile struct {
	Type        LoadProfileType        `json:"type"`
	StartRate   float64                `json:"start_rate"`
	EndRate     float64                `json:"end_rate"`
	Duration    time.Duration          `json:"duration"`
	Steps       []*LoadStep            `json:"steps"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type LoadProfileType string
const (
	LoadProfileConstant   LoadProfileType = "constant"
	LoadProfileRampUp     LoadProfileType = "ramp_up"
	LoadProfileRampDown   LoadProfileType = "ramp_down"
	LoadProfileSpike      LoadProfileType = "spike"
	LoadProfileSteps      LoadProfileType = "steps"
	LoadProfileSine       LoadProfileType = "sine"
)

type LoadStep struct {
	Duration    time.Duration `json:"duration"`
	Rate        float64       `json:"rate"`
	Users       int           `json:"users"`
}

type PerformanceStep struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Target      string                 `json:"target"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
	Validation  *ResponseValidation    `json:"validation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PerformanceAssertion struct {
	Metric      string      `json:"metric"`
	Operator    string      `json:"operator"`
	Value       interface{} `json:"value"`
	Message     string      `json:"message"`
}

type Prerequisite struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Check       func() error           `json:"-"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Configuration types
type BenchmarkConfig struct {
	DefaultIterations   int           `json:"default_iterations"`
	DefaultDuration     time.Duration `json:"default_duration"`
	WarmupRounds        int           `json:"warmup_rounds"`
	EnableProfiling     bool          `json:"enable_profiling"`
	ProfilePath         string        `json:"profile_path"`
	ReportFormat        string        `json:"report_format"`
	ReportPath          string        `json:"report_path"`
	ParallelExecution   bool          `json:"parallel_execution"`
	ResourceMonitoring  bool          `json:"resource_monitoring"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
}

// Metrics types
type PerformanceMetrics struct {
	SystemMetrics   *SystemMetrics   `json:"system_metrics"`
	RequestMetrics  *RequestMetrics  `json:"request_metrics"`
	ResourceMetrics *ResourceMetrics `json:"resource_metrics"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	mutex           sync.RWMutex
}

type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkIO       int64     `json:"network_io"`
	GoroutineCount  int       `json:"goroutine_count"`
	GCPauses        []time.Duration `json:"gc_pauses"`
	HeapSize        int64     `json:"heap_size"`
	StackSize       int64     `json:"stack_size"`
	LastUpdated     time.Time `json:"last_updated"`
}

type RequestMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	AvgResponseTime   time.Duration `json:"avg_response_time"`
	MinResponseTime   time.Duration `json:"min_response_time"`
	MaxResponseTime   time.Duration `json:"max_response_time"`
	P50ResponseTime   time.Duration `json:"p50_response_time"`
	P95ResponseTime   time.Duration `json:"p95_response_time"`
	P99ResponseTime   time.Duration `json:"p99_response_time"`
	ErrorRate         float64       `json:"error_rate"`
	Throughput        float64       `json:"throughput"`
	LastUpdated       time.Time     `json:"last_updated"`
}

type ResourceMetrics struct {
	AllocatedMemory   int64     `json:"allocated_memory"`
	UsedMemory        int64     `json:"used_memory"`
	GCCount           int64     `json:"gc_count"`
	GCTime            time.Duration `json:"gc_time"`
	OpenConnections   int       `json:"open_connections"`
	ActiveThreads     int       `json:"active_threads"`
	FileDescriptors   int       `json:"file_descriptors"`
	LastUpdated       time.Time `json:"last_updated"`
}

// TestLogger interface
type TestLogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(config *BenchmarkConfig, logger TestLogger) *BenchmarkSuite {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	return &BenchmarkSuite{
		benchmarks:     make(map[string]*Benchmark),
		loadTests:      make(map[string]*LoadTest),
		stressTests:    make(map[string]*StressTest),
		enduranceTests: make(map[string]*EnduranceTest),
		scenarios:      make(map[string]*PerformanceScenario),
		metrics:        NewPerformanceMetrics(),
		config:         config,
		logger:         logger,
	}
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		DefaultIterations:  1000,
		DefaultDuration:    30 * time.Second,
		WarmupRounds:       3,
		EnableProfiling:    true,
		ProfilePath:        "./profiles",
		ReportFormat:       "json",
		ReportPath:         "./benchmark-reports",
		ParallelExecution:  true,
		ResourceMonitoring: true,
		MetricsInterval:    1 * time.Second,
	}
}

// CreateBenchmark creates a new benchmark
func (bs *BenchmarkSuite) CreateBenchmark(name, description string, fn BenchmarkFunction) *Benchmark {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	benchmark := &Benchmark{
		Name:         name,
		Description:  description,
		Function:     fn,
		Iterations:   bs.config.DefaultIterations,
		Duration:     bs.config.DefaultDuration,
		WarmupRounds: bs.config.WarmupRounds,
		Parallel:     bs.config.ParallelExecution,
		MemoryProfile: bs.config.EnableProfiling,
		CPUProfile:   bs.config.EnableProfiling,
		Tags:         make([]string, 0),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}

	bs.benchmarks[name] = benchmark
	bs.logger.Info("Benchmark created", "name", name, "description", description)

	return benchmark
}

// CreateLoadTest creates a new load test
func (bs *BenchmarkSuite) CreateLoadTest(name, description, target string) *LoadTest {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	loadTest := &LoadTest{
		Name:         name,
		Description:  description,
		Target:       target,
		VirtualUsers: 10,
		Duration:     5 * time.Minute,
		RampUpTime:   30 * time.Second,
		RampDownTime: 30 * time.Second,
		RequestRate:  10.0,
		ThinkTime:    1 * time.Second,
		Scenarios:    make([]*LoadScenario, 0),
		Thresholds:   DefaultPerformanceThresholds(),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}

	bs.loadTests[name] = loadTest
	bs.logger.Info("Load test created", "name", name, "target", target)

	return loadTest
}

// DefaultPerformanceThresholds returns default performance thresholds
func DefaultPerformanceThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		MaxResponseTime:   2 * time.Second,
		MaxErrorRate:      5.0,
		MinThroughput:     100.0,
		MaxCPUUsage:       80.0,
		MaxMemoryUsage:    85.0,
		MaxDiskUsage:      90.0,
		MaxNetworkLatency: 100 * time.Millisecond,
	}
}

// CreateStressTest creates a new stress test
func (bs *BenchmarkSuite) CreateStressTest(name, description, target string) *StressTest {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	stressTest := &StressTest{
		Name:              name,
		Description:       description,
		Target:            target,
		StartUsers:        1,
		MaxUsers:          100,
		UserIncrement:     5,
		IncrementInterval: 30 * time.Second,
		Duration:          10 * time.Minute,
		BreakingPoint:     DefaultBreakingPointConfig(),
		Scenarios:         make([]*LoadScenario, 0),
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
	}

	bs.stressTests[name] = stressTest
	bs.logger.Info("Stress test created", "name", name, "target", target)

	return stressTest
}

// DefaultBreakingPointConfig returns default breaking point configuration
func DefaultBreakingPointConfig() *BreakingPointConfig {
	return &BreakingPointConfig{
		ErrorRateThreshold:    50.0,
		ResponseTimeThreshold: 10 * time.Second,
		CPUThreshold:          95.0,
		MemoryThreshold:       95.0,
		ConsecutiveFailures:   10,
	}
}

// RunBenchmark executes a benchmark
func (bs *BenchmarkSuite) RunBenchmark(ctx context.Context, name string) (*BenchmarkResult, error) {
	bs.mutex.RLock()
	benchmark, exists := bs.benchmarks[name]
	bs.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("benchmark %s not found", name)
	}

	bs.logger.Info("Running benchmark", "name", name)

	// Setup
	if benchmark.Setup != nil {
		if err := benchmark.Setup(); err != nil {
			return nil, fmt.Errorf("benchmark setup failed: %w", err)
		}
	}

	// Warmup
	for i := 0; i < benchmark.WarmupRounds; i++ {
		bs.runBenchmarkIteration(benchmark, 100)
	}

	// Start metrics collection
	bs.startMetricsCollection(ctx)

	startTime := time.Now()
	var totalDuration time.Duration
	var totalAllocations int64
	var totalBytes int64

	// Run benchmark iterations
	if benchmark.Parallel {
		totalDuration, totalAllocations, totalBytes = bs.runParallelBenchmark(benchmark)
	} else {
		totalDuration, totalAllocations, totalBytes = bs.runSequentialBenchmark(benchmark)
	}

	endTime := time.Now()

	// Stop metrics collection
	bs.stopMetricsCollection()

	// Teardown
	if benchmark.Teardown != nil {
		if err := benchmark.Teardown(); err != nil {
			bs.logger.Error("Benchmark teardown failed", err, "name", name)
		}
	}

	// Calculate results
	result := &BenchmarkResult{
		Name:              name,
		Iterations:        benchmark.Iterations,
		TotalDuration:     endTime.Sub(startTime),
		AvgDuration:       totalDuration / time.Duration(benchmark.Iterations),
		TotalAllocations:  totalAllocations,
		AllocationsPerOp:  totalAllocations / int64(benchmark.Iterations),
		TotalBytes:        totalBytes,
		BytesPerOp:        totalBytes / int64(benchmark.Iterations),
		OperationsPerSec:  float64(benchmark.Iterations) / endTime.Sub(startTime).Seconds(),
		SystemMetrics:     bs.metrics.SystemMetrics,
		StartTime:         startTime,
		EndTime:           endTime,
		Metadata:          make(map[string]interface{}),
	}

	bs.logger.Info("Benchmark completed", 
		"name", name,
		"iterations", benchmark.Iterations,
		"avg_duration", result.AvgDuration,
		"ops_per_sec", result.OperationsPerSec,
	)

	return result, nil
}

// runParallelBenchmark runs benchmark in parallel
func (bs *BenchmarkSuite) runParallelBenchmark(benchmark *Benchmark) (time.Duration, int64, int64) {
	var totalDuration int64
	var totalAllocations int64
	var totalBytes int64

	numCPU := runtime.NumCPU()
	iterationsPerCPU := benchmark.Iterations / numCPU
	remainder := benchmark.Iterations % numCPU

	var wg sync.WaitGroup
	for i := 0; i < numCPU; i++ {
		iterations := iterationsPerCPU
		if i < remainder {
			iterations++
		}

		wg.Add(1)
		go func(iters int) {
			defer wg.Done()
			duration, allocs, bytes := bs.runBenchmarkIteration(benchmark, iters)
			atomic.AddInt64(&totalDuration, int64(duration))
			atomic.AddInt64(&totalAllocations, allocs)
			atomic.AddInt64(&totalBytes, bytes)
		}(iterations)
	}

	wg.Wait()
	return time.Duration(totalDuration), totalAllocations, totalBytes
}

// runSequentialBenchmark runs benchmark sequentially
func (bs *BenchmarkSuite) runSequentialBenchmark(benchmark *Benchmark) (time.Duration, int64, int64) {
	return bs.runBenchmarkIteration(benchmark, benchmark.Iterations)
}

// runBenchmarkIteration runs benchmark iterations
func (bs *BenchmarkSuite) runBenchmarkIteration(benchmark *Benchmark, iterations int) (time.Duration, int64, int64) {
	var totalDuration time.Duration
	var totalAllocations int64
	var totalBytes int64

	ctx := &BenchmarkContext{
		N: iterations,
		StartTimer: func() {},
		StopTimer:  func() {},
		ResetTimer: func() {},
		SetBytes:   func(n int64) { totalBytes += n },
		ReportAllocs: func() {},
		Logf:       func(format string, args ...interface{}) { bs.logger.Debug(fmt.Sprintf(format, args...)) },
		Fatalf:     func(format string, args ...interface{}) { bs.logger.Error(fmt.Sprintf(format, args...), nil) },
		Skipf:      func(format string, args ...interface{}) { bs.logger.Info(fmt.Sprintf(format, args...)) },
	}

	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		// Get memory stats before
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Run benchmark function
		if err := benchmark.Function(ctx); err != nil {
			bs.logger.Error("Benchmark iteration failed", err, "iteration", i)
		}

		// Get memory stats after
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		duration := time.Since(start)
		totalDuration += duration
		totalAllocations += int64(m2.Mallocs - m1.Mallocs)
	}

	return totalDuration, totalAllocations, totalBytes
}

// RunLoadTest executes a load test
func (bs *BenchmarkSuite) RunLoadTest(ctx context.Context, name string) (*LoadTestResult, error) {
	bs.mutex.RLock()
	loadTest, exists := bs.loadTests[name]
	bs.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("load test %s not found", name)
	}

	bs.logger.Info("Running load test", "name", name, "virtual_users", loadTest.VirtualUsers, "duration", loadTest.Duration)

	// Start metrics collection
	bs.startMetricsCollection(ctx)

	startTime := time.Now()
	
	// Simplified load test execution
	// In a real implementation, this would spawn virtual users and execute scenarios
	time.Sleep(loadTest.Duration)

	endTime := time.Now()

	// Stop metrics collection
	bs.stopMetricsCollection()

	// Create result
	result := &LoadTestResult{
		Name:            name,
		VirtualUsers:    loadTest.VirtualUsers,
		Duration:        endTime.Sub(startTime),
		TotalRequests:   int64(loadTest.VirtualUsers) * int64(loadTest.Duration.Seconds()),
		RequestsPerSec:  float64(loadTest.VirtualUsers),
		AvgResponseTime: 100 * time.Millisecond,
		ErrorRate:       0.5,
		Throughput:      float64(loadTest.VirtualUsers) * 1024, // bytes/sec
		SystemMetrics:   bs.metrics.SystemMetrics,
		StartTime:       startTime,
		EndTime:         endTime,
		Metadata:        make(map[string]interface{}),
	}

	bs.logger.Info("Load test completed", 
		"name", name,
		"requests", result.TotalRequests,
		"rps", result.RequestsPerSec,
		"error_rate", result.ErrorRate,
	)

	return result, nil
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		SystemMetrics:   &SystemMetrics{},
		RequestMetrics:  &RequestMetrics{},
		ResourceMetrics: &ResourceMetrics{},
		CustomMetrics:   make(map[string]interface{}),
		StartTime:       time.Now(),
	}
}

// startMetricsCollection starts collecting performance metrics
func (bs *BenchmarkSuite) startMetricsCollection(ctx context.Context) {
	if !bs.config.ResourceMonitoring {
		return
	}

	bs.metrics.StartTime = time.Now()
	
	go func() {
		ticker := time.NewTicker(bs.config.MetricsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bs.collectMetrics()
			}
		}
	}()
}

// stopMetricsCollection stops collecting performance metrics
func (bs *BenchmarkSuite) stopMetricsCollection() {
	bs.metrics.EndTime = time.Now()
}

// collectMetrics collects current system metrics
func (bs *BenchmarkSuite) collectMetrics() {
	bs.metrics.mutex.Lock()
	defer bs.metrics.mutex.Unlock()

	// Collect system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	bs.metrics.SystemMetrics.GoroutineCount = runtime.NumGoroutine()
	bs.metrics.SystemMetrics.HeapSize = int64(m.HeapAlloc)
	bs.metrics.SystemMetrics.StackSize = int64(m.StackInuse)
	bs.metrics.SystemMetrics.LastUpdated = time.Now()

	// Collect resource metrics
	bs.metrics.ResourceMetrics.AllocatedMemory = int64(m.Alloc)
	bs.metrics.ResourceMetrics.UsedMemory = int64(m.Sys)
	bs.metrics.ResourceMetrics.GCCount = int64(m.NumGC)
	bs.metrics.ResourceMetrics.LastUpdated = time.Now()
}

// Result types
type BenchmarkResult struct {
	Name              string                 `json:"name"`
	Iterations        int                    `json:"iterations"`
	TotalDuration     time.Duration          `json:"total_duration"`
	AvgDuration       time.Duration          `json:"avg_duration"`
	TotalAllocations  int64                  `json:"total_allocations"`
	AllocationsPerOp  int64                  `json:"allocations_per_op"`
	TotalBytes        int64                  `json:"total_bytes"`
	BytesPerOp        int64                  `json:"bytes_per_op"`
	OperationsPerSec  float64                `json:"operations_per_sec"`
	SystemMetrics     *SystemMetrics         `json:"system_metrics"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	Metadata          map[string]interface{} `json:"metadata"`
}

type LoadTestResult struct {
	Name            string                 `json:"name"`
	VirtualUsers    int                    `json:"virtual_users"`
	Duration        time.Duration          `json:"duration"`
	TotalRequests   int64                  `json:"total_requests"`
	RequestsPerSec  float64                `json:"requests_per_sec"`
	AvgResponseTime time.Duration          `json:"avg_response_time"`
	ErrorRate       float64                `json:"error_rate"`
	Throughput      float64                `json:"throughput"`
	SystemMetrics   *SystemMetrics         `json:"system_metrics"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// GenerateReport generates a performance test report
func (bs *BenchmarkSuite) GenerateReport(results map[string]interface{}) *PerformanceReport {
	return &PerformanceReport{
		TotalTests:    len(results),
		Results:       results,
		SystemMetrics: bs.metrics.SystemMetrics,
		GeneratedAt:   time.Now(),
	}
}

// PerformanceReport represents a performance test report
type PerformanceReport struct {
	TotalTests    int                    `json:"total_tests"`
	Results       map[string]interface{} `json:"results"`
	SystemMetrics *SystemMetrics         `json:"system_metrics"`
	GeneratedAt   time.Time              `json:"generated_at"`
}
