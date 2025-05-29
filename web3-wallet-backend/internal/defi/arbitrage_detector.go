// Package defi provides DeFi-related functionality including arbitrage detection,
// yield farming, and liquidity management across multiple blockchain networks.
//
// This package implements Clean Architecture principles with clear separation
// between business logic, data access, and external integrations.
package defi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ArbitrageDetectorInterface defines the contract for arbitrage detection services.
// This interface follows the Dependency Inversion Principle, allowing for easy
// testing and different implementations.
type ArbitrageDetectorInterface interface {
	// Start begins the arbitrage detection process
	Start(ctx context.Context) error

	// Stop stops the arbitrage detection process
	Stop()

	// GetOpportunities returns current arbitrage opportunities
	GetOpportunities(ctx context.Context) ([]*ArbitrageDetection, error)

	// DetectArbitrageForToken detects arbitrage opportunities for a specific token
	DetectArbitrageForToken(ctx context.Context, token Token) ([]*ArbitrageDetection, error)

	// SetConfiguration updates detector configuration
	SetConfiguration(config ArbitrageConfig) error

	// GetMetrics returns performance metrics
	GetMetrics() ArbitrageMetrics
}

// PriceProvider defines the interface for getting token prices from exchanges
type PriceProvider interface {
	GetPrice(ctx context.Context, token Token) (decimal.Decimal, error)
	GetExchangeInfo() Exchange
	IsHealthy(ctx context.Context) bool
}

// ArbitrageConfig holds configuration for the arbitrage detector
type ArbitrageConfig struct {
	MinProfitMargin decimal.Decimal `json:"min_profit_margin" yaml:"min_profit_margin"`
	MaxGasCost      decimal.Decimal `json:"max_gas_cost" yaml:"max_gas_cost"`
	ScanInterval    time.Duration   `json:"scan_interval" yaml:"scan_interval"`
	MaxOpportunities int            `json:"max_opportunities" yaml:"max_opportunities"`
	EnabledChains   []string        `json:"enabled_chains" yaml:"enabled_chains"`
}

// ArbitrageMetrics holds performance metrics for the arbitrage detector
type ArbitrageMetrics struct {
	TotalOpportunities   int64         `json:"total_opportunities"`
	ProfitableOpportunities int64      `json:"profitable_opportunities"`
	AverageProfitMargin  decimal.Decimal `json:"average_profit_margin"`
	LastScanDuration     time.Duration `json:"last_scan_duration"`
	ErrorCount           int64         `json:"error_count"`
	LastError            string        `json:"last_error,omitempty"`
	Uptime               time.Duration `json:"uptime"`
}

// ArbitrageDetector detects arbitrage opportunities across multiple DEXs.
// It implements the ArbitrageDetectorInterface and follows Clean Architecture principles.
type ArbitrageDetector struct {
	// Dependencies (injected)
	logger         *logger.Logger
	cache          redis.Client
	priceProviders []PriceProvider

	// Configuration
	config ArbitrageConfig

	// State management
	exchanges     []Exchange
	watchedTokens []Token
	opportunities map[string]*ArbitrageDetection
	metrics       ArbitrageMetrics
	startTime     time.Time

	// Concurrency control
	mutex sync.RWMutex

	// Channels for async processing
	opportunityChan chan *ArbitrageDetection
	stopChan        chan struct{}

	// Internal state
	isRunning bool
}

// NewArbitrageDetector creates a new arbitrage detector with dependency injection.
// This constructor follows Clean Architecture principles by accepting interfaces
// rather than concrete implementations.
func NewArbitrageDetector(
	logger *logger.Logger,
	cache redis.Client,
	priceProviders []PriceProvider,
) *ArbitrageDetector {
	config := ArbitrageConfig{
		MinProfitMargin:  decimal.NewFromFloat(0.005), // 0.5% minimum profit
		MaxGasCost:       decimal.NewFromFloat(0.01),  // Max $10 gas cost
		ScanInterval:     time.Second * 30,            // Scan every 30 seconds
		MaxOpportunities: 100,
		EnabledChains:    []string{"ethereum", "bsc", "polygon"},
	}

	return &ArbitrageDetector{
		logger:          logger.Named("arbitrage-detector"),
		cache:           cache,
		priceProviders:  priceProviders,
		config:          config,
		exchanges:       initializeExchanges(),
		watchedTokens:   initializeWatchedTokens(),
		opportunities:   make(map[string]*ArbitrageDetection),
		metrics:         ArbitrageMetrics{},
		opportunityChan: make(chan *ArbitrageDetection, 100),
		stopChan:        make(chan struct{}),
		isRunning:       false,
	}
}

// NewArbitrageDetectorWithConfig creates a new arbitrage detector with custom configuration
func NewArbitrageDetectorWithConfig(
	logger *logger.Logger,
	cache redis.Client,
	priceProviders []PriceProvider,
	config ArbitrageConfig,
) *ArbitrageDetector {
	detector := NewArbitrageDetector(logger, cache, priceProviders)
	detector.config = config
	return detector
}

// Start begins the arbitrage detection process
func (ad *ArbitrageDetector) Start(ctx context.Context) error {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()

	if ad.isRunning {
		return fmt.Errorf("arbitrage detector is already running")
	}

	ad.logger.Info("Starting arbitrage detector")
	ad.startTime = time.Now()
	ad.isRunning = true

	// Start the main detection loop
	go ad.detectionLoop(ctx)

	// Start the opportunity processor
	go ad.processOpportunities(ctx)

	return nil
}

// SetConfiguration updates detector configuration
func (ad *ArbitrageDetector) SetConfiguration(config ArbitrageConfig) error {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()

	ad.config = config
	ad.logger.Info("Configuration updated",
		zap.String("min_profit_margin", config.MinProfitMargin.String()),
		zap.String("max_gas_cost", config.MaxGasCost.String()),
		zap.Duration("scan_interval", config.ScanInterval))

	return nil
}

// GetMetrics returns performance metrics
func (ad *ArbitrageDetector) GetMetrics() ArbitrageMetrics {
	ad.mutex.RLock()
	defer ad.mutex.RUnlock()

	metrics := ad.metrics
	if ad.isRunning {
		metrics.Uptime = time.Since(ad.startTime)
	}

	return metrics
}

// Stop stops the arbitrage detection process
func (ad *ArbitrageDetector) Stop() {
	ad.logger.Info("Stopping arbitrage detector")
	close(ad.stopChan)
}

// GetOpportunities returns current arbitrage opportunities
func (ad *ArbitrageDetector) GetOpportunities(ctx context.Context) ([]*ArbitrageDetection, error) {
	ad.mutex.RLock()
	defer ad.mutex.RUnlock()

	opportunities := make([]*ArbitrageDetection, 0, len(ad.opportunities))
	for _, opp := range ad.opportunities {
		if opp.Status == OpportunityStatusDetected && time.Now().Before(opp.ExpiresAt) {
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

// DetectArbitrageForToken detects arbitrage opportunities for a specific token
func (ad *ArbitrageDetector) DetectArbitrageForToken(ctx context.Context, token Token) ([]*ArbitrageDetection, error) {
	ad.logger.Debug("Detecting arbitrage for token", zap.String("token", token.Symbol))

	var opportunities []*ArbitrageDetection

	// Get prices from different exchanges
	prices, err := ad.getPricesFromExchanges(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	// Find arbitrage opportunities
	for i, sourceExchange := range ad.exchanges {
		for j, targetExchange := range ad.exchanges {
			if i == j || sourceExchange.Chain != targetExchange.Chain {
				continue
			}

			sourcePrice, exists := prices[sourceExchange.ID]
			if !exists {
				continue
			}

			targetPrice, exists := prices[targetExchange.ID]
			if !exists {
				continue
			}

			// Calculate potential profit
			if opportunity := ad.calculateArbitrageOpportunity(
				token, sourceExchange, targetExchange, sourcePrice, targetPrice,
			); opportunity != nil {
				opportunities = append(opportunities, opportunity)
			}
		}
	}

	return opportunities, nil
}

// detectionLoop runs the main detection loop
func (ad *ArbitrageDetector) detectionLoop(ctx context.Context) {
	ticker := time.NewTicker(ad.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ad.mutex.Lock()
			ad.isRunning = false
			ad.mutex.Unlock()
			return
		case <-ad.stopChan:
			ad.mutex.Lock()
			ad.isRunning = false
			ad.mutex.Unlock()
			return
		case <-ticker.C:
			start := time.Now()
			ad.scanForOpportunities(ctx)

			// Update metrics
			ad.mutex.Lock()
			ad.metrics.LastScanDuration = time.Since(start)
			ad.mutex.Unlock()
		}
	}
}

// scanForOpportunities scans all watched tokens for arbitrage opportunities
func (ad *ArbitrageDetector) scanForOpportunities(ctx context.Context) {
	ad.logger.Debug("Scanning for arbitrage opportunities")

	for _, token := range ad.watchedTokens {
		opportunities, err := ad.DetectArbitrageForToken(ctx, token)
		if err != nil {
			ad.logger.Error("Failed to detect arbitrage for token",
				zap.String("token", token.Symbol),
				zap.Error(err))
			continue
		}

		// Send opportunities to processing channel
		for _, opp := range opportunities {
			select {
			case ad.opportunityChan <- opp:
			default:
				ad.logger.Warn("Opportunity channel full, dropping opportunity")
			}
		}
	}
}

// processOpportunities processes detected opportunities
func (ad *ArbitrageDetector) processOpportunities(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ad.stopChan:
			return
		case opp := <-ad.opportunityChan:
			ad.handleOpportunity(ctx, opp)
		}
	}
}

// handleOpportunity handles a detected arbitrage opportunity
func (ad *ArbitrageDetector) handleOpportunity(ctx context.Context, opp *ArbitrageDetection) {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()

	// Store opportunity
	ad.opportunities[opp.ID] = opp

	// Update metrics
	ad.metrics.TotalOpportunities++
	if opp.NetProfit.GreaterThan(decimal.Zero) {
		ad.metrics.ProfitableOpportunities++

		// Update average profit margin
		if ad.metrics.ProfitableOpportunities == 1 {
			ad.metrics.AverageProfitMargin = opp.ProfitMargin
		} else {
			// Calculate running average
			currentSum := ad.metrics.AverageProfitMargin.Mul(decimal.NewFromInt(ad.metrics.ProfitableOpportunities - 1))
			newSum := currentSum.Add(opp.ProfitMargin)
			ad.metrics.AverageProfitMargin = newSum.Div(decimal.NewFromInt(ad.metrics.ProfitableOpportunities))
		}
	}

	// Cache opportunity for external access
	cacheKey := fmt.Sprintf("arbitrage:opportunity:%s", opp.ID)
	if err := ad.cache.Set(ctx, cacheKey, opp, time.Minute*5); err != nil {
		ad.logger.Error("Failed to cache opportunity", zap.Error(err))
		ad.metrics.ErrorCount++
		ad.metrics.LastError = fmt.Sprintf("Cache error: %v", err)
	}

	ad.logger.Info("Arbitrage opportunity detected",
		zap.String("id", opp.ID),
		zap.String("token", opp.Token.Symbol),
		zap.String("profit_margin", opp.ProfitMargin.String()),
		zap.String("net_profit", opp.NetProfit.String()),
		zap.String("source", opp.SourceExchange.Name),
		zap.String("target", opp.TargetExchange.Name),
		zap.String("confidence", opp.Confidence.String()),
		zap.String("risk", string(opp.Risk)))
}

// getPricesFromExchanges gets token prices from all configured price providers
func (ad *ArbitrageDetector) getPricesFromExchanges(ctx context.Context, token Token) (map[string]decimal.Decimal, error) {
	prices := make(map[string]decimal.Decimal)

	// Get prices from all configured price providers
	for _, provider := range ad.priceProviders {
		// Check if provider is healthy before requesting price
		if !provider.IsHealthy(ctx) {
			ad.logger.Warn("Price provider is unhealthy, skipping",
				zap.String("exchange", provider.GetExchangeInfo().Name))
			continue
		}

		price, err := provider.GetPrice(ctx, token)
		if err != nil {
			ad.logger.Error("Failed to get price from provider",
				zap.String("exchange", provider.GetExchangeInfo().Name),
				zap.String("token", token.Symbol),
				zap.Error(err))

			// Update error metrics
			ad.mutex.Lock()
			ad.metrics.ErrorCount++
			ad.metrics.LastError = fmt.Sprintf("Price fetch error from %s: %v",
				provider.GetExchangeInfo().Name, err)
			ad.mutex.Unlock()
			continue
		}

		exchangeInfo := provider.GetExchangeInfo()
		prices[exchangeInfo.ID] = price

		ad.logger.Debug("Got price from provider",
			zap.String("exchange", exchangeInfo.Name),
			zap.String("token", token.Symbol),
			zap.String("price", price.String()))
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices available for token %s", token.Symbol)
	}

	return prices, nil
}

// Note: getUniswapPrice and getOneInchPrice methods have been replaced
// by the PriceProvider interface pattern for better architecture

// calculateArbitrageOpportunity calculates if there's a profitable arbitrage opportunity
func (ad *ArbitrageDetector) calculateArbitrageOpportunity(
	token Token,
	sourceExchange, targetExchange Exchange,
	sourcePrice, targetPrice decimal.Decimal,
) *ArbitrageDetection {
	// Calculate profit margin
	profitMargin := targetPrice.Sub(sourcePrice).Div(sourcePrice)

	// Check if profit margin meets minimum threshold
	if profitMargin.LessThan(ad.config.MinProfitMargin) {
		return nil
	}

	// Estimate gas cost (simplified)
	gasCost := decimal.NewFromFloat(0.005) // $5 estimated gas cost

	// Calculate volume (simplified - use 1 token for now)
	volume := decimal.NewFromFloat(1.0)
	grossProfit := profitMargin.Mul(sourcePrice).Mul(volume)
	netProfit := grossProfit.Sub(gasCost)

	// Check if net profit is positive
	if netProfit.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// Calculate confidence and risk
	confidence := ad.calculateConfidence(profitMargin, volume)
	risk := ad.calculateRisk(profitMargin, volume, gasCost)

	return &ArbitrageDetection{
		ID:             uuid.New().String(),
		Token:          token,
		SourceExchange: sourceExchange,
		TargetExchange: targetExchange,
		SourcePrice:    sourcePrice,
		TargetPrice:    targetPrice,
		ProfitMargin:   profitMargin,
		Volume:         volume,
		GasCost:        gasCost,
		NetProfit:      netProfit,
		Confidence:     confidence,
		Risk:           risk,
		ExecutionTime:  time.Second * 30, // Estimated execution time
		ExpiresAt:      time.Now().Add(time.Minute * 5),
		Status:         OpportunityStatusDetected,
		CreatedAt:      time.Now(),
	}
}

// calculateConfidence calculates confidence score for the opportunity
func (ad *ArbitrageDetector) calculateConfidence(profitMargin, volume decimal.Decimal) decimal.Decimal {
	// Simple confidence calculation based on profit margin and volume
	confidence := profitMargin.Mul(decimal.NewFromFloat(10)) // Scale profit margin
	if volume.GreaterThan(decimal.NewFromFloat(10)) {
		confidence = confidence.Mul(decimal.NewFromFloat(1.2)) // Boost for higher volume
	}

	// Cap confidence at 1.0
	if confidence.GreaterThan(decimal.NewFromFloat(1.0)) {
		confidence = decimal.NewFromFloat(1.0)
	}

	return confidence
}

// calculateRisk calculates risk level for the opportunity
func (ad *ArbitrageDetector) calculateRisk(profitMargin, volume, gasCost decimal.Decimal) RiskLevel {
	// Calculate risk based on profit margin and gas cost ratio
	riskRatio := gasCost.Div(profitMargin.Mul(volume))

	if riskRatio.LessThan(decimal.NewFromFloat(0.1)) {
		return RiskLevelLow
	} else if riskRatio.LessThan(decimal.NewFromFloat(0.3)) {
		return RiskLevelMedium
	}

	return RiskLevelHigh
}

// initializeExchanges initializes the list of supported exchanges
func initializeExchanges() []Exchange {
	return []Exchange{
		{
			ID:       "uniswap-v3",
			Name:     "Uniswap V3",
			Type:     ExchangeTypeDEX,
			Chain:    ChainEthereum,
			Protocol: ProtocolTypeUniswap,
			Address:  "0xE592427A0AEce92De3Edee1F18E0157C05861564", // Uniswap V3 Router
			Fee:      decimal.NewFromFloat(0.003),                  // 0.3% fee
			Active:   true,
		},
		{
			ID:       "pancakeswap-v3",
			Name:     "PancakeSwap V3",
			Type:     ExchangeTypeDEX,
			Chain:    ChainBSC,
			Protocol: ProtocolTypePancakeSwap,
			Address:  "0x13f4EA83D0bd40E75C8222255bc855a974568Dd4", // PancakeSwap V3 Router
			Fee:      decimal.NewFromFloat(0.0025),                 // 0.25% fee
			Active:   true,
		},
		{
			ID:       "quickswap-v3",
			Name:     "QuickSwap V3",
			Type:     ExchangeTypeDEX,
			Chain:    ChainPolygon,
			Protocol: ProtocolTypeQuickSwap,
			Address:  "0xf5b509bB0909a69B1c207E495f687a596C168E12", // QuickSwap V3 Router
			Fee:      decimal.NewFromFloat(0.003),                  // 0.3% fee
			Active:   true,
		},
		{
			ID:       "1inch-aggregator",
			Name:     "1inch Aggregator",
			Type:     ExchangeTypeDEX,
			Chain:    ChainEthereum,
			Protocol: ProtocolType1inch,
			Address:  "0x1111111254EEB25477B68fb85Ed929f73A960582", // 1inch V5 Router
			Fee:      decimal.Zero,                                 // No direct fee
			Active:   true,
		},
	}
}

// initializeWatchedTokens initializes the list of tokens to watch for arbitrage
func initializeWatchedTokens() []Token {
	return []Token{
		{
			Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			Symbol:   "WETH",
			Name:     "Wrapped Ether",
			Decimals: 18,
			Chain:    ChainEthereum,
		},
		{
			Address:  "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
			Chain:    ChainEthereum,
		},
		{
			Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			Symbol:   "USDT",
			Name:     "Tether USD",
			Decimals: 6,
			Chain:    ChainEthereum,
		},
		{
			Address:  "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
			Symbol:   "WBTC",
			Name:     "Wrapped Bitcoin",
			Decimals: 8,
			Chain:    ChainEthereum,
		},
		{
			Address:  "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
			Symbol:   "UNI",
			Name:     "Uniswap",
			Decimals: 18,
			Chain:    ChainEthereum,
		},
		// Coffee Token
		{
			Address:  "0x0000000000000000000000000000000000000000", // Will be updated with actual deployment
			Symbol:   "COFFEE",
			Name:     "Coffee Token",
			Decimals: 18,
			Chain:    ChainEthereum,
		},
	}
}
