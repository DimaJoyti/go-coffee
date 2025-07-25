package observability

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryConfig holds configuration for OpenTelemetry
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	PrometheusPort int
	SampleRate     float64
	EnableTracing  bool
	EnableMetrics  bool
	EnableLogging  bool
}

// TelemetryProvider manages OpenTelemetry providers
type TelemetryProvider struct {
	config TelemetryConfig
	tracer trace.Tracer
	meter  metric.Meter

	// HFT-specific metrics (will be initialized when meter is available)
	orderLatencyHistogram  metric.Float64Histogram
	orderThroughputCounter metric.Int64Counter
	strategyPnLGauge       metric.Float64Gauge
	riskViolationCounter   metric.Int64Counter
	marketDataLatencyHist  metric.Float64Histogram
	fillRateGauge          metric.Float64Gauge
	positionSizeGauge      metric.Float64Gauge
	errorCounter           metric.Int64Counter
}

// NewTelemetryProvider creates a new telemetry provider
func NewTelemetryProvider(config TelemetryConfig) (*TelemetryProvider, error) {
	tp := &TelemetryProvider{
		config: config,
	}

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize tracer and meter with global providers
	if config.EnableTracing {
		tp.tracer = otel.Tracer(
			config.ServiceName,
			trace.WithInstrumentationVersion(config.ServiceVersion),
		)
	}

	if config.EnableMetrics {
		tp.meter = otel.Meter(
			config.ServiceName,
			metric.WithInstrumentationVersion(config.ServiceVersion),
		)

		// Initialize HFT-specific metrics
		if err := tp.initHFTMetrics(); err != nil {
			log.Printf("Warning: Failed to initialize HFT metrics: %v", err)
		}
	}

	return tp, nil
}

// initHFTMetrics initializes HFT-specific metrics
func (tp *TelemetryProvider) initHFTMetrics() error {
	if tp.meter == nil {
		return fmt.Errorf("meter not initialized")
	}

	var err error

	// Order latency histogram (microseconds)
	tp.orderLatencyHistogram, err = tp.meter.Float64Histogram(
		"hft_order_latency_microseconds",
		metric.WithDescription("Order processing latency in microseconds"),
		metric.WithUnit("μs"),
	)
	if err != nil {
		return fmt.Errorf("failed to create order latency histogram: %w", err)
	}

	// Order throughput counter
	tp.orderThroughputCounter, err = tp.meter.Int64Counter(
		"hft_orders_total",
		metric.WithDescription("Total number of orders processed"),
	)
	if err != nil {
		return fmt.Errorf("failed to create order throughput counter: %w", err)
	}

	// Strategy PnL gauge
	tp.strategyPnLGauge, err = tp.meter.Float64Gauge(
		"hft_strategy_pnl",
		metric.WithDescription("Strategy profit and loss"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		return fmt.Errorf("failed to create strategy PnL gauge: %w", err)
	}

	// Risk violation counter
	tp.riskViolationCounter, err = tp.meter.Int64Counter(
		"hft_risk_violations_total",
		metric.WithDescription("Total number of risk violations"),
	)
	if err != nil {
		return fmt.Errorf("failed to create risk violation counter: %w", err)
	}

	// Market data latency histogram
	tp.marketDataLatencyHist, err = tp.meter.Float64Histogram(
		"hft_market_data_latency_microseconds",
		metric.WithDescription("Market data processing latency in microseconds"),
		metric.WithUnit("μs"),
	)
	if err != nil {
		return fmt.Errorf("failed to create market data latency histogram: %w", err)
	}

	// Fill rate gauge
	tp.fillRateGauge, err = tp.meter.Float64Gauge(
		"hft_fill_rate_percent",
		metric.WithDescription("Order fill rate percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return fmt.Errorf("failed to create fill rate gauge: %w", err)
	}

	// Position size gauge
	tp.positionSizeGauge, err = tp.meter.Float64Gauge(
		"hft_position_size",
		metric.WithDescription("Current position size"),
	)
	if err != nil {
		return fmt.Errorf("failed to create position size gauge: %w", err)
	}

	// Error counter
	tp.errorCounter, err = tp.meter.Int64Counter(
		"hft_errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create error counter: %w", err)
	}

	return nil
}

// GetTracer returns the tracer
func (tp *TelemetryProvider) GetTracer() trace.Tracer {
	return tp.tracer
}

// GetMeter returns the meter
func (tp *TelemetryProvider) GetMeter() metric.Meter {
	return tp.meter
}

// RecordOrderLatency records order processing latency
func (tp *TelemetryProvider) RecordOrderLatency(ctx context.Context, latency time.Duration, attributes ...attribute.KeyValue) {
	if tp.orderLatencyHistogram != nil {
		tp.orderLatencyHistogram.Record(ctx, float64(latency.Microseconds()), metric.WithAttributes(attributes...))
	}
}

// IncrementOrderCount increments the order counter
func (tp *TelemetryProvider) IncrementOrderCount(ctx context.Context, attributes ...attribute.KeyValue) {
	if tp.orderThroughputCounter != nil {
		tp.orderThroughputCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	}
}

// RecordStrategyPnL records strategy profit and loss
func (tp *TelemetryProvider) RecordStrategyPnL(ctx context.Context, pnl float64, attributes ...attribute.KeyValue) {
	if tp.strategyPnLGauge != nil {
		tp.strategyPnLGauge.Record(ctx, pnl, metric.WithAttributes(attributes...))
	}
}

// IncrementRiskViolation increments the risk violation counter
func (tp *TelemetryProvider) IncrementRiskViolation(ctx context.Context, attributes ...attribute.KeyValue) {
	if tp.riskViolationCounter != nil {
		tp.riskViolationCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	}
}

// RecordMarketDataLatency records market data processing latency
func (tp *TelemetryProvider) RecordMarketDataLatency(ctx context.Context, latency time.Duration, attributes ...attribute.KeyValue) {
	if tp.marketDataLatencyHist != nil {
		tp.marketDataLatencyHist.Record(ctx, float64(latency.Microseconds()), metric.WithAttributes(attributes...))
	}
}

// RecordFillRate records order fill rate
func (tp *TelemetryProvider) RecordFillRate(ctx context.Context, fillRate float64, attributes ...attribute.KeyValue) {
	if tp.fillRateGauge != nil {
		tp.fillRateGauge.Record(ctx, fillRate, metric.WithAttributes(attributes...))
	}
}

// RecordPositionSize records current position size
func (tp *TelemetryProvider) RecordPositionSize(ctx context.Context, size float64, attributes ...attribute.KeyValue) {
	if tp.positionSizeGauge != nil {
		tp.positionSizeGauge.Record(ctx, size, metric.WithAttributes(attributes...))
	}
}

// IncrementErrorCount increments the error counter
func (tp *TelemetryProvider) IncrementErrorCount(ctx context.Context, attributes ...attribute.KeyValue) {
	if tp.errorCounter != nil {
		tp.errorCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	}
}

// Shutdown gracefully shuts down the telemetry provider
func (tp *TelemetryProvider) Shutdown(ctx context.Context) error {
	// In this simplified version, we don't manage providers directly
	// The global providers should be managed by the main application
	log.Printf("Telemetry provider shutdown requested")
	return nil
}

// HFTMetrics provides a convenient interface for HFT-specific metrics
type HFTMetrics struct {
	provider *TelemetryProvider
}

// NewHFTMetrics creates a new HFT metrics instance
func NewHFTMetrics(provider *TelemetryProvider) *HFTMetrics {
	return &HFTMetrics{
		provider: provider,
	}
}

// RecordOrderPlaced records an order placement event
func (m *HFTMetrics) RecordOrderPlaced(ctx context.Context, strategyID, symbol, exchange, side, orderType string, latency time.Duration) {
	attributes := []attribute.KeyValue{
		attribute.String("strategy_id", strategyID),
		attribute.String("symbol", symbol),
		attribute.String("exchange", exchange),
		attribute.String("side", side),
		attribute.String("order_type", orderType),
		attribute.String("event", "order_placed"),
	}

	m.provider.RecordOrderLatency(ctx, latency, attributes...)
	m.provider.IncrementOrderCount(ctx, attributes...)
}

// RecordOrderFilled records an order fill event
func (m *HFTMetrics) RecordOrderFilled(ctx context.Context, strategyID, symbol, exchange string, fillLatency time.Duration, fillPrice, fillQuantity float64) {
	attributes := []attribute.KeyValue{
		attribute.String("strategy_id", strategyID),
		attribute.String("symbol", symbol),
		attribute.String("exchange", exchange),
		attribute.String("event", "order_filled"),
		attribute.Float64("fill_price", fillPrice),
		attribute.Float64("fill_quantity", fillQuantity),
	}

	m.provider.RecordOrderLatency(ctx, fillLatency, attributes...)
	m.provider.IncrementOrderCount(ctx, attributes...)
}

// RecordMarketDataReceived records market data reception
func (m *HFTMetrics) RecordMarketDataReceived(ctx context.Context, symbol, exchange string, latency time.Duration, price float64) {
	attributes := []attribute.KeyValue{
		attribute.String("symbol", symbol),
		attribute.String("exchange", exchange),
		attribute.String("event", "market_data_received"),
		attribute.Float64("price", price),
	}

	m.provider.RecordMarketDataLatency(ctx, latency, attributes...)
}

// RecordStrategyPerformance records strategy performance metrics
func (m *HFTMetrics) RecordStrategyPerformance(ctx context.Context, strategyID string, pnl, fillRate, positionSize float64) {
	attributes := []attribute.KeyValue{
		attribute.String("strategy_id", strategyID),
	}

	m.provider.RecordStrategyPnL(ctx, pnl, attributes...)
	m.provider.RecordFillRate(ctx, fillRate, attributes...)
	m.provider.RecordPositionSize(ctx, positionSize, attributes...)
}

// RecordRiskEvent records a risk management event
func (m *HFTMetrics) RecordRiskEvent(ctx context.Context, strategyID, riskType, severity string) {
	attributes := []attribute.KeyValue{
		attribute.String("strategy_id", strategyID),
		attribute.String("risk_type", riskType),
		attribute.String("severity", severity),
	}

	if severity == "violation" {
		m.provider.IncrementRiskViolation(ctx, attributes...)
	}
}

// RecordError records an error event
func (m *HFTMetrics) RecordError(ctx context.Context, component, errorType, errorMessage string) {
	attributes := []attribute.KeyValue{
		attribute.String("component", component),
		attribute.String("error_type", errorType),
		attribute.String("error_message", errorMessage),
	}

	m.provider.IncrementErrorCount(ctx, attributes...)
}

// DefaultTelemetryConfig returns a default telemetry configuration
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    "hft-system",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLPEndpoint:   "http://localhost:4318/v1/traces",
		PrometheusPort: 9090,
		SampleRate:     1.0,
		EnableTracing:  true,
		EnableMetrics:  true,
		EnableLogging:  true,
	}
}

// InitializeTelemetry initializes telemetry with default configuration
func InitializeTelemetry() (*TelemetryProvider, error) {
	config := DefaultTelemetryConfig()
	return NewTelemetryProvider(config)
}

// LogTelemetryInfo logs telemetry initialization information
func LogTelemetryInfo(config TelemetryConfig) {
	log.Printf("Initializing HFT Telemetry:")
	log.Printf("  Service: %s v%s", config.ServiceName, config.ServiceVersion)
	log.Printf("  Environment: %s", config.Environment)
	log.Printf("  Tracing: %v (OTLP: %s)", config.EnableTracing, config.OTLPEndpoint)
	log.Printf("  Metrics: %v (Prometheus: %d)", config.EnableMetrics, config.PrometheusPort)
	log.Printf("  Sample Rate: %.2f", config.SampleRate)
}
