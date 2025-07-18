package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Client interfaces for external data sources

// DefiLlamaClientInterface defines the interface for DeFiLlama API client
type DefiLlamaClientInterface interface {
	GetProtocolTVL(ctx context.Context, protocol string) (decimal.Decimal, error)
	GetChainTVL(ctx context.Context, chain string) (decimal.Decimal, error)
	GetHistoricalTVL(ctx context.Context, protocol string, days int) ([]TVLDataPoint, error)
	GetProtocolList(ctx context.Context) ([]ProtocolInfo, error)
}

// AnalyticsClientInterface defines the interface for analytics data client
type AnalyticsClientInterface interface {
	GetUserMetrics(ctx context.Context, feature string, period string) (*UserMetrics, error)
	GetEngagementMetrics(ctx context.Context, period string) (*EngagementMetrics, error)
	GetRetentionMetrics(ctx context.Context, period string) (*RetentionMetrics, error)
	TrackEvent(ctx context.Context, event *AnalyticsEvent) error
}

// BlockchainClientInterface defines the interface for blockchain data client
type BlockchainClientInterface interface {
	GetLatestBlock(ctx context.Context) (int64, error)
	GetTVLFromContract(ctx context.Context, contractAddress string) (decimal.Decimal, error)
	GetTransactionCount(ctx context.Context, address string) (int64, error)
	MonitorContract(ctx context.Context, contractAddress string, callback func(*ContractEvent)) error
}

// Data structures for external APIs

// TVLDataPoint represents a TVL data point
type TVLDataPoint struct {
	Timestamp time.Time       `json:"timestamp"`
	TVL       decimal.Decimal `json:"tvl"`
}

// ProtocolInfo represents protocol information
type ProtocolInfo struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Symbol   string          `json:"symbol"`
	Chain    string          `json:"chain"`
	Category string          `json:"category"`
	TVL      decimal.Decimal `json:"tvl"`
}

// UserMetrics represents user analytics metrics
type UserMetrics struct {
	Feature        string `json:"feature"`
	ActiveUsers    int    `json:"active_users"`
	NewUsers       int    `json:"new_users"`
	ReturningUsers int    `json:"returning_users"`
	Period         string `json:"period"`
}

// EngagementMetrics represents user engagement metrics
type EngagementMetrics struct {
	PageViews       int64           `json:"page_views"`
	SessionDuration decimal.Decimal `json:"session_duration"`
	BounceRate      decimal.Decimal `json:"bounce_rate"`
	ConversionRate  decimal.Decimal `json:"conversion_rate"`
	Period          string          `json:"period"`
}

// RetentionMetrics represents user retention metrics
type RetentionMetrics struct {
	Day1Retention  decimal.Decimal `json:"day1_retention"`
	Day7Retention  decimal.Decimal `json:"day7_retention"`
	Day30Retention decimal.Decimal `json:"day30_retention"`
	CohortSize     int             `json:"cohort_size"`
	Period         string          `json:"period"`
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	EventType  string                 `json:"event_type"`
	UserID     string                 `json:"user_id"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ContractEvent represents a blockchain contract event
type ContractEvent struct {
	ContractAddress string                 `json:"contract_address"`
	EventName       string                 `json:"event_name"`
	BlockNumber     int64                  `json:"block_number"`
	TxHash          string                 `json:"tx_hash"`
	Data            map[string]interface{} `json:"data"`
	Timestamp       time.Time              `json:"timestamp"`
}

// Client implementations

// DefiLlamaClient implements DefiLlamaClientInterface
type DefiLlamaClient struct {
	logger *logger.Logger
}

// NewDefiLlamaClient creates a new DeFiLlama client
func NewDefiLlamaClient(logger *logger.Logger) DefiLlamaClientInterface {
	return &DefiLlamaClient{
		logger: logger,
	}
}

func (c *DefiLlamaClient) GetProtocolTVL(ctx context.Context, protocol string) (decimal.Decimal, error) {
	c.logger.Info("Getting protocol TVL from DeFiLlama",
		zap.String("protocol", protocol))

	// Mock TVL data - in real implementation, this would call DeFiLlama API
	tvl := decimal.NewFromFloat(5000000) // $5M

	c.logger.Info("Retrieved protocol TVL",
		zap.String("protocol", protocol),
		zap.String("tvl", tvl.String()))

	return tvl, nil
}

func (c *DefiLlamaClient) GetChainTVL(ctx context.Context, chain string) (decimal.Decimal, error) {
	c.logger.Info("Getting chain TVL from DeFiLlama",
		zap.String("chain", chain))

	// Mock TVL data
	tvl := decimal.NewFromFloat(25000000) // $25M

	return tvl, nil
}

func (c *DefiLlamaClient) GetHistoricalTVL(ctx context.Context, protocol string, days int) ([]TVLDataPoint, error) {
	c.logger.Info("Getting historical TVL from DeFiLlama",
		zap.String("protocol", protocol),
		zap.Int("days", days))

	// Mock historical data
	var dataPoints []TVLDataPoint
	baseTime := time.Now().AddDate(0, 0, -days)
	baseTVL := decimal.NewFromFloat(4000000)

	for i := 0; i < days; i++ {
		growth := decimal.NewFromFloat(float64(i) * 0.01) // 1% daily growth
		tvl := baseTVL.Mul(decimal.NewFromFloat(1).Add(growth))

		dataPoints = append(dataPoints, TVLDataPoint{
			Timestamp: baseTime.AddDate(0, 0, i),
			TVL:       tvl,
		})
	}

	return dataPoints, nil
}

func (c *DefiLlamaClient) GetProtocolList(ctx context.Context) ([]ProtocolInfo, error) {
	c.logger.Info("Getting protocol list from DeFiLlama")

	// Mock protocol list
	protocols := []ProtocolInfo{
		{
			ID:       "go-coffee-defi",
			Name:     "Go Coffee DeFi",
			Symbol:   "GCDEFI",
			Chain:    "ethereum",
			Category: "DEX",
			TVL:      decimal.NewFromFloat(5000000),
		},
		{
			ID:       "go-coffee-lending",
			Name:     "Go Coffee Lending",
			Symbol:   "GCLEND",
			Chain:    "ethereum",
			Category: "Lending",
			TVL:      decimal.NewFromFloat(3000000),
		},
	}

	return protocols, nil
}

// AnalyticsClient implements AnalyticsClientInterface
type AnalyticsClient struct {
	logger *logger.Logger
}

// NewAnalyticsClient creates a new analytics client
func NewAnalyticsClient(logger *logger.Logger) AnalyticsClientInterface {
	return &AnalyticsClient{
		logger: logger,
	}
}

func (c *AnalyticsClient) GetUserMetrics(ctx context.Context, feature string, period string) (*UserMetrics, error) {
	c.logger.Info("Getting user metrics",
		zap.String("feature", feature),
		zap.String("period", period))

	// Mock user metrics
	metrics := &UserMetrics{
		Feature:        feature,
		ActiveUsers:    25000,
		NewUsers:       2500,
		ReturningUsers: 22500,
		Period:         period,
	}

	return metrics, nil
}

func (c *AnalyticsClient) GetEngagementMetrics(ctx context.Context, period string) (*EngagementMetrics, error) {
	c.logger.Info("Getting engagement metrics",
		zap.String("period", period))

	// Mock engagement metrics
	metrics := &EngagementMetrics{
		PageViews:       150000,
		SessionDuration: decimal.NewFromFloat(8.5),  // 8.5 minutes
		BounceRate:      decimal.NewFromFloat(0.35), // 35%
		ConversionRate:  decimal.NewFromFloat(0.12), // 12%
		Period:          period,
	}

	return metrics, nil
}

func (c *AnalyticsClient) GetRetentionMetrics(ctx context.Context, period string) (*RetentionMetrics, error) {
	c.logger.Info("Getting retention metrics",
		zap.String("period", period))

	// Mock retention metrics
	metrics := &RetentionMetrics{
		Day1Retention:  decimal.NewFromFloat(0.85), // 85%
		Day7Retention:  decimal.NewFromFloat(0.65), // 65%
		Day30Retention: decimal.NewFromFloat(0.45), // 45%
		CohortSize:     1000,
		Period:         period,
	}

	return metrics, nil
}

func (c *AnalyticsClient) TrackEvent(ctx context.Context, event *AnalyticsEvent) error {
	c.logger.Info("Tracking analytics event",
		zap.String("eventType", event.EventType),
		zap.String("userID", event.UserID))

	// In real implementation, this would send event to analytics service
	return nil
}

// BlockchainClient implements BlockchainClientInterface
type BlockchainClient struct {
	client *ethclient.Client
	chain  string
	logger *logger.Logger
}

// NewBlockchainClient creates a new blockchain client
func NewBlockchainClient(client *ethclient.Client, chain string, logger *logger.Logger) BlockchainClientInterface {
	return &BlockchainClient{
		client: client,
		chain:  chain,
		logger: logger,
	}
}

func (c *BlockchainClient) GetLatestBlock(ctx context.Context) (int64, error) {
	c.logger.Info("Getting latest block number",
		zap.String("chain", c.chain))

	// Mock block number
	blockNumber := int64(18500000)

	return blockNumber, nil
}

func (c *BlockchainClient) GetTVLFromContract(ctx context.Context, contractAddress string) (decimal.Decimal, error) {
	c.logger.Info("Getting TVL from contract",
		zap.String("chain", c.chain),
		zap.String("contract", contractAddress))

	// Mock TVL from contract
	tvl := decimal.NewFromFloat(2500000) // $2.5M

	return tvl, nil
}

func (c *BlockchainClient) GetTransactionCount(ctx context.Context, address string) (int64, error) {
	c.logger.Info("Getting transaction count",
		zap.String("chain", c.chain),
		zap.String("address", address))

	// Mock transaction count
	txCount := int64(1250)

	return txCount, nil
}

func (c *BlockchainClient) MonitorContract(ctx context.Context, contractAddress string, callback func(*ContractEvent)) error {
	c.logger.Info("Starting contract monitoring",
		zap.String("chain", c.chain),
		zap.String("contract", contractAddress))

	// In real implementation, this would set up event monitoring
	// For now, we'll just log that monitoring started
	return nil
}

// Additional helper functions

// CalculateGrowthRate calculates growth rate between two values
func CalculateGrowthRate(current, previous decimal.Decimal) decimal.Decimal {
	if previous.IsZero() {
		return decimal.Zero
	}

	diff := current.Sub(previous)
	growth := diff.Div(previous).Mul(decimal.NewFromFloat(100))

	return growth
}

// CalculateMovingAverage calculates moving average for a series of values
func CalculateMovingAverage(values []decimal.Decimal, window int) []decimal.Decimal {
	if len(values) < window {
		return []decimal.Decimal{}
	}

	var averages []decimal.Decimal

	for i := window - 1; i < len(values); i++ {
		sum := decimal.Zero
		for j := i - window + 1; j <= i; j++ {
			sum = sum.Add(values[j])
		}
		average := sum.Div(decimal.NewFromInt(int64(window)))
		averages = append(averages, average)
	}

	return averages
}

// DetectAnomaly detects anomalies in metric values using simple threshold
func DetectAnomaly(current, average, threshold decimal.Decimal) bool {
	diff := current.Sub(average).Abs()
	return diff.GreaterThan(threshold)
}

// FormatMetricValue formats metric values for display
func FormatMetricValue(value decimal.Decimal, metricType MetricType) string {
	switch metricType {
	case MetricTypeTVL, MetricTypeRevenue:
		return fmt.Sprintf("$%s", value.StringFixed(2))
	case MetricTypeMAU, MetricTypeUsers, MetricTypeTransactions:
		return value.StringFixed(0)
	default:
		return value.StringFixed(2)
	}
}
