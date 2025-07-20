package benchmark

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/cpu"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/latency"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/lockfree"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/performance/memory"
	"github.com/shopspring/decimal"
)

// BenchmarkResult holds the results of a benchmark
type BenchmarkResult struct {
	Name            string
	Duration        time.Duration
	Operations      int64
	OperationsPerSec float64
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	AllocBytes      uint64
	AllocObjects    uint64
	MemoryUsage     uint64
	CPUUsage        float64
	Timestamp       time.Time
}

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	results         []BenchmarkResult
	memoryManager   *memory.MemoryManager
	latencyTracker  *latency.LatencyTracker
	threadManager   *cpu.ThreadManager
	mutex           sync.RWMutex
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		results:        make([]BenchmarkResult, 0),
		memoryManager:  memory.GetGlobalMemoryManager(),
		latencyTracker: latency.GetGlobalLatencyTracker(),
		threadManager:  cpu.GetGlobalThreadManager(),
	}
}

// RunAllBenchmarks runs all performance benchmarks
func (bs *BenchmarkSuite) RunAllBenchmarks() []BenchmarkResult {
	benchmarks := []func() BenchmarkResult{
		bs.BenchmarkOrderCreation,
		bs.BenchmarkOrderBookOperations,
		bs.BenchmarkMemoryPooling,
		bs.BenchmarkLatencyMeasurement,
		bs.BenchmarkLockFreeOperations,
		bs.BenchmarkConcurrentOrderProcessing,
		bs.BenchmarkMarketDataProcessing,
		bs.BenchmarkRiskCalculations,
		bs.BenchmarkSerializationDeserialization,
		bs.BenchmarkNetworkOperations,
	}
	
	results := make([]BenchmarkResult, 0, len(benchmarks))
	
	for _, benchmark := range benchmarks {
		result := benchmark()
		results = append(results, result)
		
		bs.mutex.Lock()
		bs.results = append(bs.results, result)
		bs.mutex.Unlock()
		
		// Allow GC between benchmarks
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
	}
	
	return results
}

// BenchmarkOrderCreation benchmarks order creation performance
func (bs *BenchmarkSuite) BenchmarkOrderCreation() BenchmarkResult {
	const iterations = 100000
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		orderStart := time.Now()
		
		order, err := entities.NewOrder(
			entities.StrategyID(fmt.Sprintf("strategy-%d", i%10)),
			entities.Symbol("BTCUSDT"),
			entities.Exchange("binance"),
			valueobjects.OrderSideBuy,
			valueobjects.OrderTypeLimit,
			valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
			valueobjects.TimeInForceGTC,
		)
		
		latencies[i] = time.Since(orderStart)
		
		if err != nil || order == nil {
			panic(fmt.Sprintf("failed to create order: %v", err))
		}
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("OrderCreation", duration, iterations, latencies, 
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkOrderBookOperations benchmarks order book operations
func (bs *BenchmarkSuite) BenchmarkOrderBookOperations() BenchmarkResult {
	const iterations = 50000
	
	orderBook := lockfree.NewLockFreeOrderBook("BTCUSDT", "binance")
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Create test order
		order, _ := entities.NewOrder(
			entities.StrategyID("test"),
			entities.Symbol("BTCUSDT"),
			entities.Exchange("binance"),
			valueobjects.OrderSideBuy,
			valueobjects.OrderTypeLimit,
			valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
			valueobjects.Price{Decimal: decimal.NewFromFloat(50000 + float64(i%1000))},
			valueobjects.TimeInForceGTC,
		)
		
		// Add to order book
		orderBook.AddOrder(order)
		
		// Get best bid/ask
		orderBook.GetBestBid()
		orderBook.GetBestAsk()
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("OrderBookOperations", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkMemoryPooling benchmarks memory pool performance
func (bs *BenchmarkSuite) BenchmarkMemoryPooling() BenchmarkResult {
	const iterations = 200000
	
	orderPool := memory.NewOrderPool(1000)
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Get from pool
		order := orderPool.Get()
		
		// Use the order (simulate work)
		_ = order
		
		// Return to pool
		orderPool.Put(order)
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("MemoryPooling", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkLatencyMeasurement benchmarks latency measurement overhead
func (bs *BenchmarkSuite) BenchmarkLatencyMeasurement() BenchmarkResult {
	const iterations = 100000
	
	tracker := latency.NewLatencyTracker(10000)
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		traceID := fmt.Sprintf("trace-%d", i)
		trace := tracker.StartTrace(traceID)
		if trace != nil {
			tracker.RecordPoint(traceID, latency.PointSystemStart)
			tracker.RecordPoint(traceID, latency.PointSystemEnd)
			tracker.EndTrace(traceID)
		}
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("LatencyMeasurement", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkLockFreeOperations benchmarks lock-free data structure operations
func (bs *BenchmarkSuite) BenchmarkLockFreeOperations() BenchmarkResult {
	const iterations = 100000
	const goroutines = 4
	
	orderBook := lockfree.NewLockFreeOrderBook("BTCUSDT", "binance")
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	var wg sync.WaitGroup
	latencyChan := make(chan time.Duration, iterations)
	
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < iterations/goroutines; i++ {
				opStart := time.Now()
				
				// Create and add order
				order, _ := entities.NewOrder(
					entities.StrategyID(fmt.Sprintf("strategy-%d", goroutineID)),
					entities.Symbol("BTCUSDT"),
					entities.Exchange("binance"),
					valueobjects.OrderSideBuy,
					valueobjects.OrderTypeLimit,
					valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
					valueobjects.Price{Decimal: decimal.NewFromFloat(50000 + float64(i))},
					valueobjects.TimeInForceGTC,
				)
				
				orderBook.AddOrder(order)
				
				latencyChan <- time.Since(opStart)
			}
		}(g)
	}
	
	wg.Wait()
	close(latencyChan)
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	// Collect latencies
	latencies := make([]time.Duration, 0, iterations)
	for lat := range latencyChan {
		latencies = append(latencies, lat)
	}
	
	return bs.calculateResult("LockFreeOperations", duration, int64(len(latencies)), latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkConcurrentOrderProcessing benchmarks concurrent order processing
func (bs *BenchmarkSuite) BenchmarkConcurrentOrderProcessing() BenchmarkResult {
	const iterations = 50000
	const goroutines = 8
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	var wg sync.WaitGroup
	var processedOrders int64
	latencyChan := make(chan time.Duration, iterations)
	
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < iterations/goroutines; i++ {
				opStart := time.Now()
				
				// Simulate order processing
				order, _ := entities.NewOrder(
					entities.StrategyID(fmt.Sprintf("strategy-%d", goroutineID)),
					entities.Symbol("BTCUSDT"),
					entities.Exchange("binance"),
					valueobjects.OrderSideBuy,
					valueobjects.OrderTypeLimit,
					valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
					valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
					valueobjects.TimeInForceGTC,
				)
				
				// Simulate validation
				_ = order.GetID()
				_ = order.GetQuantity()
				_ = order.GetPrice()
				
				// Simulate confirmation
				order.Confirm()
				
				atomic.AddInt64(&processedOrders, 1)
				latencyChan <- time.Since(opStart)
			}
		}(g)
	}
	
	wg.Wait()
	close(latencyChan)
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	// Collect latencies
	latencies := make([]time.Duration, 0, iterations)
	for lat := range latencyChan {
		latencies = append(latencies, lat)
	}
	
	return bs.calculateResult("ConcurrentOrderProcessing", duration, processedOrders, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkMarketDataProcessing benchmarks market data processing
func (bs *BenchmarkSuite) BenchmarkMarketDataProcessing() BenchmarkResult {
	const iterations = 100000
	
	marketDataPool := memory.NewMarketDataPool(5000, 1000)
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Get tick from pool
		tick := marketDataPool.GetTick()
		
		// Simulate market data processing
		tick.Symbol = "BTCUSDT"
		tick.Exchange = "binance"
		tick.Price = 50000.0 + float64(i%1000)
		tick.Quantity = 0.1
		tick.Timestamp = time.Now().UnixNano()
		
		// Process tick (simulate)
		_ = tick.Price * tick.Quantity
		
		// Return to pool
		marketDataPool.PutTick(tick)
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("MarketDataProcessing", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkRiskCalculations benchmarks risk calculation performance
func (bs *BenchmarkSuite) BenchmarkRiskCalculations() BenchmarkResult {
	const iterations = 50000
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Simulate risk calculations
		price := decimal.NewFromFloat(50000.0 + float64(i%1000))
		quantity := decimal.NewFromFloat(0.1)
		
		// Calculate position value
		positionValue := price.Mul(quantity)
		
		// Calculate risk metrics
		var riskScore decimal.Decimal
		if positionValue.GreaterThan(decimal.NewFromFloat(1000)) {
			riskScore = positionValue.Div(decimal.NewFromFloat(10000))
		} else {
			riskScore = decimal.NewFromFloat(0.01)
		}
		
		// Simulate risk limit check
		maxRisk := decimal.NewFromFloat(0.1)
		_ = riskScore.LessThan(maxRisk)
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("RiskCalculations", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkSerializationDeserialization benchmarks JSON serialization
func (bs *BenchmarkSuite) BenchmarkSerializationDeserialization() BenchmarkResult {
	const iterations = 20000
	
	// Create test order
	order, _ := entities.NewOrder(
		entities.StrategyID("test-strategy"),
		entities.Symbol("BTCUSDT"),
		entities.Exchange("binance"),
		valueobjects.OrderSideBuy,
		valueobjects.OrderTypeLimit,
		valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
		valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
		valueobjects.TimeInForceGTC,
	)
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Simulate serialization (would use actual JSON marshaling)
		data := fmt.Sprintf(`{"id":"%s","symbol":"%s","side":"%s","quantity":"%s","price":"%s"}`,
			order.GetID(), order.GetSymbol(), order.GetSide(), 
			order.GetQuantity().String(), order.GetPrice().String())
		
		// Simulate deserialization
		_ = len(data)
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("SerializationDeserialization", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// BenchmarkNetworkOperations benchmarks network operation simulation
func (bs *BenchmarkSuite) BenchmarkNetworkOperations() BenchmarkResult {
	const iterations = 10000
	
	bufferPool := memory.NewByteBufferPool(500, 200, 100)
	
	start := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	latencies := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		opStart := time.Now()
		
		// Get buffer from pool
		buffer := bufferPool.GetBuffer(1024)
		
		// Simulate network write
		data := []byte(fmt.Sprintf("order_data_%d", i))
		*buffer = append((*buffer)[:0], data...)
		
		// Simulate network read
		_ = len(*buffer)
		
		// Return buffer to pool
		bufferPool.PutBuffer(buffer, 1024)
		
		latencies[i] = time.Since(opStart)
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)
	
	return bs.calculateResult("NetworkOperations", duration, iterations, latencies,
		memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc,
		memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// calculateResult calculates benchmark result statistics
func (bs *BenchmarkSuite) calculateResult(name string, duration time.Duration, operations int64, 
	latencies []time.Duration, allocBytes, allocObjects uint64) BenchmarkResult {
	
	if len(latencies) == 0 {
		return BenchmarkResult{Name: name, Timestamp: time.Now()}
	}
	
	// Sort latencies for percentile calculations
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	
	// Simple insertion sort (good enough for benchmarking)
	for i := 1; i < len(sortedLatencies); i++ {
		key := sortedLatencies[i]
		j := i - 1
		for j >= 0 && sortedLatencies[j] > key {
			sortedLatencies[j+1] = sortedLatencies[j]
			j--
		}
		sortedLatencies[j+1] = key
	}
	
	// Calculate statistics
	minLatency := sortedLatencies[0]
	maxLatency := sortedLatencies[len(sortedLatencies)-1]
	p50Latency := sortedLatencies[len(sortedLatencies)*50/100]
	p95Latency := sortedLatencies[len(sortedLatencies)*95/100]
	p99Latency := sortedLatencies[len(sortedLatencies)*99/100]
	
	// Calculate average
	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}
	avgLatency := totalLatency / time.Duration(len(latencies))
	
	opsPerSec := float64(operations) / duration.Seconds()
	
	return BenchmarkResult{
		Name:            name,
		Duration:        duration,
		Operations:      operations,
		OperationsPerSec: opsPerSec,
		AvgLatency:      avgLatency,
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		P50Latency:      p50Latency,
		P95Latency:      p95Latency,
		P99Latency:      p99Latency,
		AllocBytes:      allocBytes,
		AllocObjects:    allocObjects,
		MemoryUsage:     getCurrentMemoryUsage(),
		CPUUsage:        getCurrentCPUUsage(),
		Timestamp:       time.Now(),
	}
}

// GetResults returns all benchmark results
func (bs *BenchmarkSuite) GetResults() []BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	results := make([]BenchmarkResult, len(bs.results))
	copy(results, bs.results)
	return results
}

// PrintResults prints benchmark results in a formatted table
func (bs *BenchmarkSuite) PrintResults() {
	results := bs.GetResults()
	
	fmt.Printf("\n=== HFT Performance Benchmark Results ===\n\n")
	fmt.Printf("%-30s %12s %15s %12s %12s %12s %12s\n", 
		"Benchmark", "Ops/Sec", "Avg Latency", "P50", "P95", "P99", "Alloc MB")
	fmt.Printf("%s\n", string(make([]byte, 120)))
	
	for _, result := range results {
		fmt.Printf("%-30s %12.0f %15s %12s %12s %12s %12.2f\n",
			result.Name,
			result.OperationsPerSec,
			result.AvgLatency.String(),
			result.P50Latency.String(),
			result.P95Latency.String(),
			result.P99Latency.String(),
			float64(result.AllocBytes)/(1024*1024))
	}
	
	fmt.Printf("\n")
}

// Helper functions

func getCurrentMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
}

func getCurrentCPUUsage() float64 {
	// This would require platform-specific implementation
	// For now, return a placeholder
	return 0.0
}

// RunBenchmarkTests runs Go benchmark tests
func RunBenchmarkTests(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	b.Run("OrderCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			entities.NewOrder(
				entities.StrategyID("test"),
				entities.Symbol("BTCUSDT"),
				entities.Exchange("binance"),
				valueobjects.OrderSideBuy,
				valueobjects.OrderTypeLimit,
				valueobjects.Quantity{Decimal: decimal.NewFromFloat(0.1)},
				valueobjects.Price{Decimal: decimal.NewFromFloat(50000)},
				valueobjects.TimeInForceGTC,
			)
		}
	})
	
	b.Run("MemoryPooling", func(b *testing.B) {
		pool := memory.NewOrderPool(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			order := pool.Get()
			pool.Put(order)
		}
	})
	
	b.Run("LatencyMeasurement", func(b *testing.B) {
		tracker := suite.latencyTracker
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			traceID := fmt.Sprintf("trace-%d", i)
			trace := tracker.StartTrace(traceID)
			if trace != nil {
				tracker.RecordPoint(traceID, latency.PointSystemStart)
				tracker.EndTrace(traceID)
			}
		}
	})
}
