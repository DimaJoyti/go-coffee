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
)

// ArbitrageDetector detects arbitrage opportunities across multiple DEXs
type ArbitrageDetector struct {
	logger        *logger.Logger
	cache         redis.Client
	uniswapClient *UniswapClient
	oneInchClient *OneInchClient

	// Configuration
	minProfitMargin decimal.Decimal
	maxGasCost      decimal.Decimal
	scanInterval    time.Duration

	// State
	exchanges     []Exchange
	watchedTokens []Token
	opportunities map[string]*ArbitrageDetection
	mutex         sync.RWMutex

	// Channels
	opportunityChan chan *ArbitrageDetection
	stopChan        chan struct{}
}

// NewArbitrageDetector creates a new arbitrage detector
func NewArbitrageDetector(
	logger *logger.Logger,
	cache redis.Client,
	uniswapClient *UniswapClient,
	oneInchClient *OneInchClient,
) *ArbitrageDetector {
	return &ArbitrageDetector{
		logger:          logger.Named("arbitrage-detector"),
		cache:           cache,
		uniswapClient:   uniswapClient,
		oneInchClient:   oneInchClient,
		minProfitMargin: decimal.NewFromFloat(0.005), // 0.5% minimum profit
		maxGasCost:      decimal.NewFromFloat(0.01),  // Max $10 gas cost
		scanInterval:    time.Second * 30,            // Scan every 30 seconds
		exchanges:       initializeExchanges(),
		watchedTokens:   initializeWatchedTokens(),
		opportunities:   make(map[string]*ArbitrageDetection),
		opportunityChan: make(chan *ArbitrageDetection, 100),
		stopChan:        make(chan struct{}),
	}
}

// Start begins the arbitrage detection process
func (ad *ArbitrageDetector) Start(ctx context.Context) error {
	ad.logger.Info("Starting arbitrage detector")

	// Start the main detection loop
	go ad.detectionLoop(ctx)

	// Start the opportunity processor
	go ad.processOpportunities(ctx)

	return nil
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
	ad.logger.Debug("Detecting arbitrage for token", "token", token.Symbol)

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
	ticker := time.NewTicker(ad.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ad.stopChan:
			return
		case <-ticker.C:
			ad.scanForOpportunities(ctx)
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
				"token", token.Symbol,
				"error", err)
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

	// Cache opportunity for external access
	cacheKey := fmt.Sprintf("arbitrage:opportunity:%s", opp.ID)
	if err := ad.cache.Set(ctx, cacheKey, opp, time.Minute*5); err != nil {
		ad.logger.Error("Failed to cache opportunity", "error", err)
	}

	ad.logger.Info("Arbitrage opportunity detected",
		"id", opp.ID,
		"token", opp.Token.Symbol,
		"profit_margin", opp.ProfitMargin,
		"net_profit", opp.NetProfit,
		"source", opp.SourceExchange.Name,
		"target", opp.TargetExchange.Name)
}

// getPricesFromExchanges gets token prices from all configured exchanges
func (ad *ArbitrageDetector) getPricesFromExchanges(ctx context.Context, token Token) (map[string]decimal.Decimal, error) {
	prices := make(map[string]decimal.Decimal)

	// Get price from Uniswap
	if uniswapPrice, err := ad.getUniswapPrice(ctx, token); err == nil {
		prices["uniswap"] = uniswapPrice
	}

	// Get price from 1inch (aggregated)
	if oneInchPrice, err := ad.getOneInchPrice(ctx, token); err == nil {
		prices["1inch"] = oneInchPrice
	}

	// Add more exchanges as needed

	return prices, nil
}

// getUniswapPrice gets token price from Uniswap
func (ad *ArbitrageDetector) getUniswapPrice(ctx context.Context, token Token) (decimal.Decimal, error) {
	// Use USDC as base currency for price comparison
	usdcAddress := "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1"

	quote, err := ad.uniswapClient.GetSwapQuote(ctx, &GetSwapQuoteRequest{
		TokenIn:  token.Address,
		TokenOut: usdcAddress,
		AmountIn: decimal.NewFromFloat(1.0),
		Chain:    token.Chain,
	})
	if err != nil {
		return decimal.Zero, err
	}

	return quote.AmountOut, nil
}

// getOneInchPrice gets token price from 1inch
func (ad *ArbitrageDetector) getOneInchPrice(ctx context.Context, token Token) (decimal.Decimal, error) {
	// Use USDC as base currency for price comparison
	usdcAddress := "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1"

	quote, err := ad.oneInchClient.GetSwapQuote(ctx, &GetSwapQuoteRequest{
		TokenIn:  token.Address,
		TokenOut: usdcAddress,
		AmountIn: decimal.NewFromFloat(1.0),
		Chain:    token.Chain,
	})
	if err != nil {
		return decimal.Zero, err
	}

	return quote.AmountOut, nil
}

// calculateArbitrageOpportunity calculates if there's a profitable arbitrage opportunity
func (ad *ArbitrageDetector) calculateArbitrageOpportunity(
	token Token,
	sourceExchange, targetExchange Exchange,
	sourcePrice, targetPrice decimal.Decimal,
) *ArbitrageDetection {
	// Calculate profit margin
	profitMargin := targetPrice.Sub(sourcePrice).Div(sourcePrice)

	// Check if profit margin meets minimum threshold
	if profitMargin.LessThan(ad.minProfitMargin) {
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
