package exchanges

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeClient defines the interface that all exchange clients must implement
type ExchangeClient interface {
	// Basic Information
	GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error)
	GetServerTime(ctx context.Context) (time.Time, error)
	GetSymbols(ctx context.Context) ([]Symbol, error)
	
	// Market Data
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error)
	GetAllTickers(ctx context.Context) ([]*Ticker, error)
	GetOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error)
	GetRecentTrades(ctx context.Context, symbol string, limit int) ([]*Trade, error)
	GetKlines(ctx context.Context, symbol, interval string, startTime, endTime *time.Time, limit int) ([]*Kline, error)
	
	// Account Data (if authenticated)
	GetBalances(ctx context.Context) ([]*Balance, error)
	GetBalance(ctx context.Context, asset string) (*Balance, error)
	
	// Trading (if authenticated)
	PlaceOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, symbol, orderID string) error
	GetOrder(ctx context.Context, symbol, orderID string) (*Order, error)
	GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error)
	GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error)
	
	// WebSocket Streaming
	SubscribeToTicker(ctx context.Context, symbol string, callback func(*Ticker)) error
	SubscribeToTrades(ctx context.Context, symbol string, callback func(*Trade)) error
	SubscribeToOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error
	SubscribeToKlines(ctx context.Context, symbol, interval string, callback func(*Kline)) error
	UnsubscribeAll(ctx context.Context) error
	
	// Connection Management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetExchangeType() ExchangeType
	
	// Health and Status
	Ping(ctx context.Context) error
	GetStatus() string
	GetLastError() error
}

// StreamingClient defines the interface for real-time streaming
type StreamingClient interface {
	// Stream Management
	StartStream(ctx context.Context) error
	StopStream(ctx context.Context) error
	IsStreaming() bool
	
	// Subscription Management
	Subscribe(ctx context.Context, req *SubscriptionRequest) error
	Unsubscribe(ctx context.Context, req *SubscriptionRequest) error
	GetSubscriptions() []*SubscriptionRequest
	
	// Data Callbacks
	OnTicker(callback func(*Ticker))
	OnTrade(callback func(*Trade))
	OnOrderBook(callback func(*OrderBook))
	OnKline(callback func(*Kline))
	OnError(callback func(error))
	OnConnect(callback func())
	OnDisconnect(callback func())
	
	// Stream Data
	GetStreamChannel() <-chan *StreamData
}

// DataProvider defines the interface for data aggregation
type DataProvider interface {
	// Multi-Exchange Data
	GetAggregatedTicker(ctx context.Context, symbol string) (*MarketSummary, error)
	GetAggregatedOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)
	GetBestPrices(ctx context.Context, symbol string) (*MarketSummary, error)
	
	// Arbitrage Detection
	FindArbitrageOpportunities(ctx context.Context, symbols []string) ([]*ArbitrageOpportunity, error)
	GetArbitrageOpportunity(ctx context.Context, symbol string) (*ArbitrageOpportunity, error)
	
	// Data Quality
	GetDataQuality(ctx context.Context, exchange ExchangeType, symbol string) (*DataQualityMetrics, error)
	GetExchangeStatus(ctx context.Context) (map[ExchangeType]string, error)
	
	// Historical Data
	GetHistoricalData(ctx context.Context, symbol, interval string, startTime, endTime time.Time) ([]*Kline, error)
	GetAggregatedVolume(ctx context.Context, symbol string, period time.Duration) (decimal.Decimal, error)
}

// PriceAggregator defines the interface for price aggregation strategies
type PriceAggregator interface {
	// Price Calculation
	CalculateWeightedPrice(tickers []*Ticker) decimal.Decimal
	CalculateVWAP(tickers []*Ticker) decimal.Decimal
	CalculateMedianPrice(tickers []*Ticker) decimal.Decimal
	CalculateBestBidAsk(tickers []*Ticker) (*ExchangePrice, *ExchangePrice)
	
	// Spread Analysis
	CalculateSpread(bid, ask decimal.Decimal) decimal.Decimal
	CalculateSpreadPercent(bid, ask decimal.Decimal) decimal.Decimal
	
	// Volume Analysis
	CalculateTotalVolume(tickers []*Ticker) decimal.Decimal
	CalculateVolumeWeights(tickers []*Ticker) map[ExchangeType]decimal.Decimal
}

// ArbitrageDetector defines the interface for arbitrage detection
type ArbitrageDetector interface {
	// Opportunity Detection
	DetectOpportunities(ctx context.Context, tickers []*Ticker) ([]*ArbitrageOpportunity, error)
	CalculateProfitability(buyPrice, sellPrice, volume decimal.Decimal) decimal.Decimal
	EstimateExecutionCost(exchange ExchangeType, volume decimal.Decimal) decimal.Decimal
	
	// Risk Assessment
	AssessRisk(opportunity *ArbitrageOpportunity) float64
	CalculateConfidence(opportunity *ArbitrageOpportunity) float64
	ValidateOpportunity(ctx context.Context, opportunity *ArbitrageOpportunity) bool
}

// DataValidator defines the interface for data validation
type DataValidator interface {
	// Price Validation
	ValidatePrice(price decimal.Decimal, symbol string) bool
	ValidatePriceRange(price decimal.Decimal, reference decimal.Decimal, tolerance float64) bool
	ValidateVolume(volume decimal.Decimal) bool
	
	// Data Consistency
	ValidateTimestamp(timestamp time.Time, tolerance time.Duration) bool
	ValidateOrderBook(orderBook *OrderBook) bool
	ValidateTicker(ticker *Ticker) bool
	
	// Quality Scoring
	CalculateDataQuality(metrics *DataQualityMetrics) float64
	UpdateQualityMetrics(exchange ExchangeType, symbol string, success bool, latency time.Duration)
}

// CacheManager defines the interface for caching
type CacheManager interface {
	// Cache Operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Bulk Operations
	SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeletePattern(ctx context.Context, pattern string) error
	
	// Cache Statistics
	GetStats(ctx context.Context) (map[string]interface{}, error)
	ClearAll(ctx context.Context) error
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// Event Publishing
	PublishTicker(ctx context.Context, ticker *Ticker) error
	PublishTrade(ctx context.Context, trade *Trade) error
	PublishOrderBook(ctx context.Context, orderBook *OrderBook) error
	PublishArbitrage(ctx context.Context, opportunity *ArbitrageOpportunity) error
	
	// Custom Events
	PublishEvent(ctx context.Context, eventType string, data interface{}) error
	
	// Subscription Management
	Subscribe(ctx context.Context, eventType string, callback func(interface{})) error
	Unsubscribe(ctx context.Context, eventType string) error
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	// Performance Metrics
	RecordLatency(exchange ExchangeType, operation string, duration time.Duration)
	RecordError(exchange ExchangeType, operation string, err error)
	RecordSuccess(exchange ExchangeType, operation string)
	
	// Data Metrics
	RecordDataPoint(exchange ExchangeType, symbol string, dataType string)
	RecordVolumeProcessed(exchange ExchangeType, volume decimal.Decimal)
	RecordArbitrageOpportunity(opportunity *ArbitrageOpportunity)
	
	// System Metrics
	RecordMemoryUsage(bytes int64)
	RecordCPUUsage(percent float64)
	RecordConnectionCount(exchange ExchangeType, count int)
	
	// Export Metrics
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
	ExportPrometheus(ctx context.Context) (string, error)
}
