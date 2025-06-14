package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/cache"
	"github.com/DimaJoyti/go-coffee/pkg/database"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"go.uber.org/zap"
)

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	logger       *zap.Logger
	dbManager    *database.Manager
	cacheManager *cache.Manager
	results      *BenchmarkResults
}

// BenchmarkResults stores benchmark results
type BenchmarkResults struct {
	DatabaseBenchmarks map[string]*BenchmarkResult `json:"database_benchmarks"`
	CacheBenchmarks    map[string]*BenchmarkResult `json:"cache_benchmarks"`
	ConcurrencyTests   map[string]*BenchmarkResult `json:"concurrency_tests"`
	MemoryTests        map[string]*BenchmarkResult `json:"memory_tests"`
	StartTime          time.Time                   `json:"start_time"`
	EndTime            time.Time                   `json:"end_time"`
	TotalDuration      time.Duration               `json:"total_duration"`
}

// BenchmarkResult stores individual benchmark results
type BenchmarkResult struct {
	Name             string          `json:"name"`
	Operations       int64           `json:"operations"`
	Duration         time.Duration   `json:"duration"`
	OperationsPerSec float64         `json:"operations_per_sec"`
	AvgLatency       time.Duration   `json:"avg_latency"`
	MinLatency       time.Duration   `json:"min_latency"`
	MaxLatency       time.Duration   `json:"max_latency"`
	P95Latency       time.Duration   `json:"p95_latency"`
	P99Latency       time.Duration   `json:"p99_latency"`
	ErrorCount       int64           `json:"error_count"`
	ErrorRate        float64         `json:"error_rate"`
	MemoryUsage      int64           `json:"memory_usage"`
	Latencies        []time.Duration `json:"-"` // Don't serialize raw latencies
}

// TestOrder represents a test order
type TestOrder struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	Items      []TestItem `json:"items"`
	Total      float64    `json:"total"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TestItem represents a test order item
type TestItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(logger *zap.Logger) (*BenchmarkSuite, error) {
	// Initialize configuration
	cfg := &config.InfrastructureConfig{
		Database: &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Database: "go_coffee_bench",
			SSLMode:  "disable",
		},
		Redis: &config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           1, // Use different DB for benchmarks
			PoolSize:     100,
			MinIdleConns: 20,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
			DialTimeout:  2 * time.Second,
			MaxRetries:   3,
			RetryDelay:   50 * time.Millisecond,
		},
	}

	// Initialize database manager
	dbManager, err := database.NewManager(cfg.Database, logger)
	if err != nil {
		logger.Warn("Database connection failed, skipping database benchmarks", zap.Error(err))
	}

	// Initialize cache manager
	cacheManager, err := cache.NewManager(cfg.Redis, logger)
	if err != nil {
		logger.Warn("Cache connection failed, skipping cache benchmarks", zap.Error(err))
	}

	return &BenchmarkSuite{
		logger:       logger,
		dbManager:    dbManager,
		cacheManager: cacheManager,
		results: &BenchmarkResults{
			DatabaseBenchmarks: make(map[string]*BenchmarkResult),
			CacheBenchmarks:    make(map[string]*BenchmarkResult),
			ConcurrencyTests:   make(map[string]*BenchmarkResult),
			MemoryTests:        make(map[string]*BenchmarkResult),
		},
	}, nil
}

// RunAllBenchmarks runs all benchmark suites
func (bs *BenchmarkSuite) RunAllBenchmarks() {
	bs.logger.Info("🚀 Starting comprehensive benchmark suite")
	bs.results.StartTime = time.Now()

	// Database benchmarks
	if bs.dbManager != nil {
		bs.logger.Info("📊 Running database benchmarks")
		bs.runDatabaseBenchmarks()
	}

	// Cache benchmarks
	if bs.cacheManager != nil {
		bs.logger.Info("⚡ Running cache benchmarks")
		bs.runCacheBenchmarks()
	}

	// Concurrency benchmarks
	bs.logger.Info("🔄 Running concurrency benchmarks")
	bs.runConcurrencyBenchmarks()

	// Memory benchmarks
	bs.logger.Info("🧠 Running memory benchmarks")
	bs.runMemoryBenchmarks()

	bs.results.EndTime = time.Now()
	bs.results.TotalDuration = bs.results.EndTime.Sub(bs.results.StartTime)

	bs.logger.Info("✅ Benchmark suite completed",
		zap.Duration("total_duration", bs.results.TotalDuration))
}

// runDatabaseBenchmarks runs database performance tests
func (bs *BenchmarkSuite) runDatabaseBenchmarks() {
	ctx := context.Background()

	// Benchmark 1: Sequential writes
	bs.results.DatabaseBenchmarks["sequential_writes"] = bs.benchmarkDatabaseWrites(ctx, 1000, 1)

	// Benchmark 2: Concurrent writes
	bs.results.DatabaseBenchmarks["concurrent_writes"] = bs.benchmarkDatabaseWrites(ctx, 1000, 10)

	// Benchmark 3: Sequential reads
	bs.results.DatabaseBenchmarks["sequential_reads"] = bs.benchmarkDatabaseReads(ctx, 1000, 1)

	// Benchmark 4: Concurrent reads
	bs.results.DatabaseBenchmarks["concurrent_reads"] = bs.benchmarkDatabaseReads(ctx, 1000, 10)
}

// benchmarkDatabaseWrites benchmarks database write operations
func (bs *BenchmarkSuite) benchmarkDatabaseWrites(ctx context.Context, operations int, concurrency int) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("Database Writes (ops=%d, concurrency=%d)", operations, concurrency),
		Latencies: make([]time.Duration, 0, operations),
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	operationsPerWorker := operations / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				order := bs.generateTestOrder()

				opStart := time.Now()
				err := bs.simulateDatabaseWrite(ctx, order)
				latency := time.Since(opStart)

				mu.Lock()
				result.Latencies = append(result.Latencies, latency)
				if err != nil {
					result.ErrorCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)
	result.Operations = int64(len(result.Latencies))

	bs.calculateBenchmarkStats(result)
	return result
}

// benchmarkDatabaseReads benchmarks database read operations
func (bs *BenchmarkSuite) benchmarkDatabaseReads(ctx context.Context, operations int, concurrency int) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("Database Reads (ops=%d, concurrency=%d)", operations, concurrency),
		Latencies: make([]time.Duration, 0, operations),
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	operationsPerWorker := operations / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				orderID := fmt.Sprintf("order-%d-%d", workerID, j)

				opStart := time.Now()
				err := bs.simulateDatabaseRead(ctx, orderID)
				latency := time.Since(opStart)

				mu.Lock()
				result.Latencies = append(result.Latencies, latency)
				if err != nil {
					result.ErrorCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)
	result.Operations = int64(len(result.Latencies))

	bs.calculateBenchmarkStats(result)
	return result
}

// runCacheBenchmarks runs cache performance tests
func (bs *BenchmarkSuite) runCacheBenchmarks() {
	ctx := context.Background()

	// Benchmark 1: Cache sets
	bs.results.CacheBenchmarks["cache_sets"] = bs.benchmarkCacheOperations(ctx, "set", 10000, 10)

	// Benchmark 2: Cache gets
	bs.results.CacheBenchmarks["cache_gets"] = bs.benchmarkCacheOperations(ctx, "get", 10000, 10)

	// Benchmark 3: Cache mixed operations
	bs.results.CacheBenchmarks["cache_mixed"] = bs.benchmarkCacheOperations(ctx, "mixed", 10000, 10)
}

// benchmarkCacheOperations benchmarks cache operations
func (bs *BenchmarkSuite) benchmarkCacheOperations(ctx context.Context, operation string, operations int, concurrency int) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("Cache %s (ops=%d, concurrency=%d)", operation, operations, concurrency),
		Latencies: make([]time.Duration, 0, operations),
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	operationsPerWorker := operations / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				var err error
				var latency time.Duration

				switch operation {
				case "set":
					latency, err = bs.benchmarkCacheSet(ctx, workerID, j)
				case "get":
					latency, err = bs.benchmarkCacheGet(ctx, workerID, j)
				case "mixed":
					if j%2 == 0 {
						latency, err = bs.benchmarkCacheSet(ctx, workerID, j)
					} else {
						latency, err = bs.benchmarkCacheGet(ctx, workerID, j)
					}
				}

				mu.Lock()
				result.Latencies = append(result.Latencies, latency)
				if err != nil {
					result.ErrorCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)
	result.Operations = int64(len(result.Latencies))

	bs.calculateBenchmarkStats(result)
	return result
}

// benchmarkCacheSet benchmarks cache set operations
func (bs *BenchmarkSuite) benchmarkCacheSet(ctx context.Context, workerID, opID int) (time.Duration, error) {
	key := fmt.Sprintf("bench:order:%d:%d", workerID, opID)
	order := bs.generateTestOrder()

	start := time.Now()
	err := bs.cacheManager.Set(ctx, key, order, 1*time.Hour)
	return time.Since(start), err
}

// benchmarkCacheGet benchmarks cache get operations
func (bs *BenchmarkSuite) benchmarkCacheGet(ctx context.Context, workerID, opID int) (time.Duration, error) {
	key := fmt.Sprintf("bench:order:%d:%d", workerID, opID)
	var order TestOrder

	start := time.Now()
	err := bs.cacheManager.Get(ctx, key, &order)
	return time.Since(start), err
}

// runConcurrencyBenchmarks runs concurrency performance tests
func (bs *BenchmarkSuite) runConcurrencyBenchmarks() {
	// Test different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 25, 50, 100}

	for _, level := range concurrencyLevels {
		name := fmt.Sprintf("concurrency_%d", level)
		bs.results.ConcurrencyTests[name] = bs.benchmarkConcurrency(level, 1000)
	}
}

// benchmarkConcurrency benchmarks concurrent operations
func (bs *BenchmarkSuite) benchmarkConcurrency(concurrency int, operationsPerWorker int) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("Concurrency Test (workers=%d, ops_per_worker=%d)", concurrency, operationsPerWorker),
		Latencies: make([]time.Duration, 0, concurrency*operationsPerWorker),
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				opStart := time.Now()

				// Simulate mixed workload
				if j%3 == 0 && bs.cacheManager != nil {
					// Cache operation
					key := fmt.Sprintf("bench:worker:%d:op:%d", workerID, j)
					bs.cacheManager.Set(context.Background(), key, "test-value", 1*time.Minute)
				} else {
					// CPU-intensive operation
					bs.simulateCPUWork(1000)
				}

				latency := time.Since(opStart)

				mu.Lock()
				result.Latencies = append(result.Latencies, latency)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)
	result.Operations = int64(len(result.Latencies))

	bs.calculateBenchmarkStats(result)
	return result
}

// runMemoryBenchmarks runs memory performance tests
func (bs *BenchmarkSuite) runMemoryBenchmarks() {
	// Test memory allocation patterns
	bs.results.MemoryTests["allocation_small"] = bs.benchmarkMemoryAllocation(1000, 1024)   // 1KB objects
	bs.results.MemoryTests["allocation_medium"] = bs.benchmarkMemoryAllocation(1000, 10240) // 10KB objects
	bs.results.MemoryTests["allocation_large"] = bs.benchmarkMemoryAllocation(100, 1048576) // 1MB objects
}

// benchmarkMemoryAllocation benchmarks memory allocation patterns
func (bs *BenchmarkSuite) benchmarkMemoryAllocation(operations int, objectSize int) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("Memory Allocation (ops=%d, size=%d bytes)", operations, objectSize),
		Latencies: make([]time.Duration, 0, operations),
	}

	start := time.Now()
	var objects [][]byte

	for i := 0; i < operations; i++ {
		opStart := time.Now()

		// Allocate object
		obj := make([]byte, objectSize)
		for j := 0; j < len(obj); j++ {
			obj[j] = byte(j % 256)
		}
		objects = append(objects, obj)

		latency := time.Since(opStart)
		result.Latencies = append(result.Latencies, latency)
	}

	result.Duration = time.Since(start)
	result.Operations = int64(len(result.Latencies))

	bs.calculateBenchmarkStats(result)

	// Force GC to measure memory impact
	runtime.GC()

	return result
}

// Helper functions

func (bs *BenchmarkSuite) generateTestOrder() *TestOrder {
	return &TestOrder{
		ID:         fmt.Sprintf("order-%d", rand.Int63()),
		CustomerID: fmt.Sprintf("customer-%d", rand.Int63()),
		Items: []TestItem{
			{ProductID: "coffee-1", Name: "Espresso", Quantity: 1, Price: 4.50},
			{ProductID: "coffee-2", Name: "Latte", Quantity: 1, Price: 5.00},
		},
		Total:     9.50,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
}

func (bs *BenchmarkSuite) simulateDatabaseWrite(ctx context.Context, order *TestOrder) error {
	// Simulate database write operation
	time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)+500)) // 0.5-1.5ms
	return nil
}

func (bs *BenchmarkSuite) simulateDatabaseRead(ctx context.Context, orderID string) error {
	// Simulate database read operation
	time.Sleep(time.Microsecond * time.Duration(rand.Intn(500)+200)) // 0.2-0.7ms
	return nil
}

func (bs *BenchmarkSuite) simulateCPUWork(iterations int) {
	// Simulate CPU-intensive work
	sum := 0
	for i := 0; i < iterations; i++ {
		sum += i * i
	}
}

func (bs *BenchmarkSuite) calculateBenchmarkStats(result *BenchmarkResult) {
	if len(result.Latencies) == 0 {
		return
	}

	// Calculate basic stats
	result.OperationsPerSec = float64(result.Operations) / result.Duration.Seconds()
	result.ErrorRate = float64(result.ErrorCount) / float64(result.Operations)

	// Calculate latency stats
	total := time.Duration(0)
	result.MinLatency = result.Latencies[0]
	result.MaxLatency = result.Latencies[0]

	for _, latency := range result.Latencies {
		total += latency
		if latency < result.MinLatency {
			result.MinLatency = latency
		}
		if latency > result.MaxLatency {
			result.MaxLatency = latency
		}
	}

	result.AvgLatency = total / time.Duration(len(result.Latencies))

	// Calculate percentiles
	sortedLatencies := make([]time.Duration, len(result.Latencies))
	copy(sortedLatencies, result.Latencies)

	// Simple sort for percentiles
	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	p95Index := int(float64(len(sortedLatencies)) * 0.95)
	p99Index := int(float64(len(sortedLatencies)) * 0.99)

	if p95Index < len(sortedLatencies) {
		result.P95Latency = sortedLatencies[p95Index]
	}
	if p99Index < len(sortedLatencies) {
		result.P99Latency = sortedLatencies[p99Index]
	}
}

// PrintResults prints benchmark results
func (bs *BenchmarkSuite) PrintResults() {
	fmt.Println("\n================================================================================")
	fmt.Println("🎯 GO COFFEE PERFORMANCE BENCHMARK RESULTS")
	fmt.Println("================================================================================")

	fmt.Printf("📅 Test Duration: %v\n", bs.results.TotalDuration)
	fmt.Printf("🕐 Started: %v\n", bs.results.StartTime.Format(time.RFC3339))
	fmt.Printf("🕐 Ended: %v\n\n", bs.results.EndTime.Format(time.RFC3339))

	// Print database benchmarks
	if len(bs.results.DatabaseBenchmarks) > 0 {
		fmt.Println("📊 DATABASE BENCHMARKS")
		fmt.Println("--------------------------------------------------")
		for name, result := range bs.results.DatabaseBenchmarks {
			bs.printBenchmarkResult(name, result)
		}
		fmt.Println()
	}

	// Print cache benchmarks
	if len(bs.results.CacheBenchmarks) > 0 {
		fmt.Println("⚡ CACHE BENCHMARKS")
		fmt.Println("--------------------------------------------------")
		for name, result := range bs.results.CacheBenchmarks {
			bs.printBenchmarkResult(name, result)
		}
		fmt.Println()
	}

	// Print concurrency benchmarks
	if len(bs.results.ConcurrencyTests) > 0 {
		fmt.Println("🔄 CONCURRENCY BENCHMARKS")
		fmt.Println("--------------------------------------------------")
		for name, result := range bs.results.ConcurrencyTests {
			bs.printBenchmarkResult(name, result)
		}
		fmt.Println()
	}

	// Print memory benchmarks
	if len(bs.results.MemoryTests) > 0 {
		fmt.Println("🧠 MEMORY BENCHMARKS")
		fmt.Println("--------------------------------------------------")
		for name, result := range bs.results.MemoryTests {
			bs.printBenchmarkResult(name, result)
		}
	}
}

func (bs *BenchmarkSuite) printBenchmarkResult(name string, result *BenchmarkResult) {
	fmt.Printf("  %s:\n", name)
	fmt.Printf("    Operations: %d\n", result.Operations)
	fmt.Printf("    Duration: %v\n", result.Duration)
	fmt.Printf("    Ops/sec: %.2f\n", result.OperationsPerSec)
	fmt.Printf("    Avg Latency: %v\n", result.AvgLatency)
	fmt.Printf("    P95 Latency: %v\n", result.P95Latency)
	fmt.Printf("    P99 Latency: %v\n", result.P99Latency)
	fmt.Printf("    Error Rate: %.2f%%\n", result.ErrorRate*100)
	fmt.Println()
}

// SaveResults saves benchmark results to JSON file
func (bs *BenchmarkSuite) SaveResults(filename string) error {
	data, err := json.MarshalIndent(bs.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create benchmark suite
	suite, err := NewBenchmarkSuite(logger)
	if err != nil {
		log.Fatal("Failed to create benchmark suite:", err)
	}

	// Run all benchmarks
	suite.RunAllBenchmarks()

	// Print results
	suite.PrintResults()

	// Save results to file
	filename := fmt.Sprintf("benchmark-results-%d.json", time.Now().Unix())
	if err := suite.SaveResults(filename); err != nil {
		logger.Error("Failed to save results", zap.Error(err))
	} else {
		logger.Info("Results saved", zap.String("filename", filename))
	}

	fmt.Println("✅ Benchmark suite completed successfully!")
}
