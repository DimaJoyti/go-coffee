package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SimpleMetrics provides a mock metrics implementation without OpenTelemetry dependencies
type SimpleMetrics struct {
	serviceName string
	counters    map[string]int64
	gauges      map[string]float64
	histograms  map[string][]float64
	mu          sync.RWMutex
}

// Counter represents a simple counter metric
type Counter struct {
	name    string
	metrics *SimpleMetrics
}

// Gauge represents a simple gauge metric
type Gauge struct {
	name    string
	metrics *SimpleMetrics
}

// Histogram represents a simple histogram metric
type Histogram struct {
	name    string
	metrics *SimpleMetrics
}

// Metrics holds all HFT system metrics (simplified version)
type Metrics struct {
	serviceName string
	simple      *SimpleMetrics

	// Order metrics
	OrdersTotal        *Counter
	OrdersLatency      *Histogram
	OrdersFilled       *Counter
	OrdersCanceled     *Counter
	OrdersRejected     *Counter

	// Market data metrics
	TicksReceived      *Counter
	TicksLatency       *Histogram
	OrderBookUpdates   *Counter
	ConnectionStatus   *Gauge

	// Strategy metrics
	StrategyPnL        *Gauge
	StrategyTrades     *Counter
	StrategySignals    *Counter
	StrategyErrors     *Counter

	// Risk metrics
	RiskChecks         *Counter
	RiskViolations     *Counter
	PositionSize       *Gauge
	Exposure           *Gauge

	// System metrics
	GoroutineCount     *Gauge
	MemoryUsage        *Gauge
	CPUUsage           *Gauge
	NetworkLatency     *Histogram
}

// New creates a new simple metrics instance
func New(serviceName string) (*Metrics, error) {
	simple := &SimpleMetrics{
		serviceName: serviceName,
		counters:    make(map[string]int64),
		gauges:      make(map[string]float64),
		histograms:  make(map[string][]float64),
	}

	return &Metrics{
		serviceName: serviceName,
		simple:      simple,

		// Order metrics
		OrdersTotal:        &Counter{"hft_orders_total", simple},
		OrdersLatency:      &Histogram{"hft_orders_latency_seconds", simple},
		OrdersFilled:       &Counter{"hft_orders_filled_total", simple},
		OrdersCanceled:     &Counter{"hft_orders_canceled_total", simple},
		OrdersRejected:     &Counter{"hft_orders_rejected_total", simple},

		// Market data metrics
		TicksReceived:      &Counter{"hft_ticks_received_total", simple},
		TicksLatency:       &Histogram{"hft_ticks_latency_seconds", simple},
		OrderBookUpdates:   &Counter{"hft_orderbook_updates_total", simple},
		ConnectionStatus:   &Gauge{"hft_connection_status", simple},

		// Strategy metrics
		StrategyPnL:        &Gauge{"hft_strategy_pnl", simple},
		StrategyTrades:     &Counter{"hft_strategy_trades_total", simple},
		StrategySignals:    &Counter{"hft_strategy_signals_total", simple},
		StrategyErrors:     &Counter{"hft_strategy_errors_total", simple},

		// Risk metrics
		RiskChecks:         &Counter{"hft_risk_checks_total", simple},
		RiskViolations:     &Counter{"hft_risk_violations_total", simple},
		PositionSize:       &Gauge{"hft_position_size", simple},
		Exposure:           &Gauge{"hft_exposure", simple},

		// System metrics
		GoroutineCount:     &Gauge{"hft_goroutines", simple},
		MemoryUsage:        &Gauge{"hft_memory_bytes", simple},
		CPUUsage:           &Gauge{"hft_cpu_usage_percent", simple},
		NetworkLatency:     &Histogram{"hft_network_latency_seconds", simple},
	}, nil
}

// Counter methods
func (c *Counter) Add(ctx context.Context, value int64, attributes ...interface{}) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()
	c.metrics.counters[c.name] += value
}

func (c *Counter) Inc(ctx context.Context, attributes ...interface{}) {
	c.Add(ctx, 1, attributes...)
}

func (c *Counter) Get() int64 {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	return c.metrics.counters[c.name]
}

// Gauge methods
func (g *Gauge) Set(ctx context.Context, value float64, attributes ...interface{}) {
	g.metrics.mu.Lock()
	defer g.metrics.mu.Unlock()
	g.metrics.gauges[g.name] = value
}

func (g *Gauge) Add(ctx context.Context, value float64, attributes ...interface{}) {
	g.metrics.mu.Lock()
	defer g.metrics.mu.Unlock()
	g.metrics.gauges[g.name] += value
}

func (g *Gauge) Get() float64 {
	g.metrics.mu.RLock()
	defer g.metrics.mu.RUnlock()
	return g.metrics.gauges[g.name]
}

// Histogram methods
func (h *Histogram) Record(ctx context.Context, value float64, attributes ...interface{}) {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()
	if h.metrics.histograms[h.name] == nil {
		h.metrics.histograms[h.name] = make([]float64, 0)
	}
	h.metrics.histograms[h.name] = append(h.metrics.histograms[h.name], value)
	
	// Keep only last 1000 values to prevent memory growth
	if len(h.metrics.histograms[h.name]) > 1000 {
		h.metrics.histograms[h.name] = h.metrics.histograms[h.name][len(h.metrics.histograms[h.name])-1000:]
	}
}

func (h *Histogram) GetValues() []float64 {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()
	values := h.metrics.histograms[h.name]
	if values == nil {
		return []float64{}
	}
	// Return a copy to prevent race conditions
	result := make([]float64, len(values))
	copy(result, values)
	return result
}

// Metrics utility methods
func (m *Metrics) GetServiceName() string {
	return m.serviceName
}

func (m *Metrics) GetAllCounters() map[string]int64 {
	m.simple.mu.RLock()
	defer m.simple.mu.RUnlock()
	
	result := make(map[string]int64)
	for k, v := range m.simple.counters {
		result[k] = v
	}
	return result
}

func (m *Metrics) GetAllGauges() map[string]float64 {
	m.simple.mu.RLock()
	defer m.simple.mu.RUnlock()
	
	result := make(map[string]float64)
	for k, v := range m.simple.gauges {
		result[k] = v
	}
	return result
}

func (m *Metrics) GetAllHistograms() map[string][]float64 {
	m.simple.mu.RLock()
	defer m.simple.mu.RUnlock()
	
	result := make(map[string][]float64)
	for k, v := range m.simple.histograms {
		if v != nil {
			result[k] = make([]float64, len(v))
			copy(result[k], v)
		}
	}
	return result
}

// Reset all metrics
func (m *Metrics) Reset() {
	m.simple.mu.Lock()
	defer m.simple.mu.Unlock()
	
	m.simple.counters = make(map[string]int64)
	m.simple.gauges = make(map[string]float64)
	m.simple.histograms = make(map[string][]float64)
}

// Export metrics in Prometheus format for compatibility
func (m *Metrics) ExportPrometheus() string {
	m.simple.mu.RLock()
	defer m.simple.mu.RUnlock()
	
	var result string
	
	// Export counters
	for name, value := range m.simple.counters {
		result += fmt.Sprintf("# TYPE %s counter\n", name)
		result += fmt.Sprintf("%s %d\n", name, value)
	}
	
	// Export gauges
	for name, value := range m.simple.gauges {
		result += fmt.Sprintf("# TYPE %s gauge\n", name)
		result += fmt.Sprintf("%s %.6f\n", name, value)
	}
	
	// Export histograms (simplified - just count and sum)
	for name, values := range m.simple.histograms {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			result += fmt.Sprintf("# TYPE %s histogram\n", name)
			result += fmt.Sprintf("%s_count %d\n", name, len(values))
			result += fmt.Sprintf("%s_sum %.6f\n", name, sum)
		}
	}
	
	return result
}

// Helper functions for common metric operations

// RecordOrderLatency records order execution latency
func (m *Metrics) RecordOrderLatency(ctx context.Context, duration time.Duration) {
	m.OrdersLatency.Record(ctx, duration.Seconds())
}

// IncrementOrdersTotal increments total orders counter
func (m *Metrics) IncrementOrdersTotal(ctx context.Context, orderType string) {
	m.OrdersTotal.Inc(ctx, "type", orderType)
}

// SetConnectionStatus sets connection status (1 for connected, 0 for disconnected)
func (m *Metrics) SetConnectionStatus(ctx context.Context, connected bool) {
	if connected {
		m.ConnectionStatus.Set(ctx, 1.0)
	} else {
		m.ConnectionStatus.Set(ctx, 0.0)
	}
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics(ctx context.Context, goroutines int64, memoryBytes int64, cpuPercent float64) {
	m.GoroutineCount.Set(ctx, float64(goroutines))
	m.MemoryUsage.Set(ctx, float64(memoryBytes))
	m.CPUUsage.Set(ctx, cpuPercent)
}

// RecordNetworkLatency records network operation latency
func (m *Metrics) RecordNetworkLatency(ctx context.Context, duration time.Duration) {
	m.NetworkLatency.Record(ctx, duration.Seconds())
}
