package aggregators

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// DEXAggregator provides unified access to multiple DEX aggregators
type DEXAggregator struct {
	logger *logger.Logger
	config DEXAggregatorConfig

	// Aggregator clients
	oneInchClient  *OneInchClient
	paraswapClient *ParaswapClient
	zeroXClient    *ZeroXClient
	matchaClient   *MatchaClient

	// State management
	mutex         sync.RWMutex
	isRunning     bool
	lastUpdate    time.Time
	healthStatus  map[string]bool
	priceCache    map[string]*CachedQuote
	routingEngine *RoutingEngine
}

// DEXAggregatorConfig holds configuration for DEX aggregator
type DEXAggregatorConfig struct {
	Enabled             bool                        `json:"enabled" yaml:"enabled"`
	DefaultSlippage     decimal.Decimal             `json:"default_slippage" yaml:"default_slippage"`
	MaxSlippage         decimal.Decimal             `json:"max_slippage" yaml:"max_slippage"`
	QuoteTimeout        time.Duration               `json:"quote_timeout" yaml:"quote_timeout"`
	CacheTimeout        time.Duration               `json:"cache_timeout" yaml:"cache_timeout"`
	MaxRetries          int                         `json:"max_retries" yaml:"max_retries"`
	HealthCheckInterval time.Duration               `json:"health_check_interval" yaml:"health_check_interval"`
	Aggregators         map[string]AggregatorConfig `json:"aggregators" yaml:"aggregators"`
	RoutingConfig       RoutingConfig               `json:"routing_config" yaml:"routing_config"`
}

// AggregatorConfig holds configuration for individual aggregators
type AggregatorConfig struct {
	Enabled   bool            `json:"enabled" yaml:"enabled"`
	APIKey    string          `json:"api_key" yaml:"api_key"`
	BaseURL   string          `json:"base_url" yaml:"base_url"`
	Timeout   time.Duration   `json:"timeout" yaml:"timeout"`
	RateLimit int             `json:"rate_limit" yaml:"rate_limit"`
	Priority  int             `json:"priority" yaml:"priority"`
	Chains    []string        `json:"chains" yaml:"chains"`
	MinAmount decimal.Decimal `json:"min_amount" yaml:"min_amount"`
	MaxAmount decimal.Decimal `json:"max_amount" yaml:"max_amount"`
}

// RoutingConfig holds routing configuration
type RoutingConfig struct {
	Strategy             RoutingStrategy `json:"strategy" yaml:"strategy"`
	ParallelQuotes       bool            `json:"parallel_quotes" yaml:"parallel_quotes"`
	FallbackEnabled      bool            `json:"fallback_enabled" yaml:"fallback_enabled"`
	PriceImpactThreshold decimal.Decimal `json:"price_impact_threshold" yaml:"price_impact_threshold"`
	GasOptimization      bool            `json:"gas_optimization" yaml:"gas_optimization"`
	MinSavingsThreshold  decimal.Decimal `json:"min_savings_threshold" yaml:"min_savings_threshold"`
}

// RoutingStrategy defines routing strategies
type RoutingStrategy string

const (
	RoutingStrategyBestPrice  RoutingStrategy = "best_price"
	RoutingStrategyLowestGas  RoutingStrategy = "lowest_gas"
	RoutingStrategyBestValue  RoutingStrategy = "best_value"
	RoutingStrategyFastest    RoutingStrategy = "fastest"
	RoutingStrategyMostLiquid RoutingStrategy = "most_liquid"
	RoutingStrategyBalanced   RoutingStrategy = "balanced"
)

// AggregatedQuote represents a quote from multiple aggregators
type AggregatedQuote struct {
	ID                string                 `json:"id"`
	TokenIn           Token                  `json:"token_in"`
	TokenOut          Token                  `json:"token_out"`
	AmountIn          decimal.Decimal        `json:"amount_in"`
	BestQuote         *SwapQuote             `json:"best_quote"`
	AllQuotes         []*SwapQuote           `json:"all_quotes"`
	Savings           decimal.Decimal        `json:"savings"`
	SavingsPercentage decimal.Decimal        `json:"savings_percentage"`
	RecommendedRoute  *RouteRecommendation   `json:"recommended_route"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
	ExpiresAt         time.Time              `json:"expires_at"`
}

// SwapQuote represents a swap quote from an aggregator
type SwapQuote struct {
	ID            string                 `json:"id"`
	Aggregator    string                 `json:"aggregator"`
	Protocol      string                 `json:"protocol"`
	Chain         string                 `json:"chain"`
	TokenIn       Token                  `json:"token_in"`
	TokenOut      Token                  `json:"token_out"`
	AmountIn      decimal.Decimal        `json:"amount_in"`
	AmountOut     decimal.Decimal        `json:"amount_out"`
	MinAmountOut  decimal.Decimal        `json:"min_amount_out"`
	PriceImpact   decimal.Decimal        `json:"price_impact"`
	Fee           decimal.Decimal        `json:"fee"`
	GasEstimate   uint64                 `json:"gas_estimate"`
	GasPrice      decimal.Decimal        `json:"gas_price"`
	GasCost       decimal.Decimal        `json:"gas_cost"`
	Route         []RouteStep            `json:"route"`
	Slippage      decimal.Decimal        `json:"slippage"`
	ExpiresAt     time.Time              `json:"expires_at"`
	CreatedAt     time.Time              `json:"created_at"`
	ExecutionData map[string]interface{} `json:"execution_data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Token represents a token
type Token struct {
	Address  string          `json:"address"`
	Symbol   string          `json:"symbol"`
	Name     string          `json:"name"`
	Decimals int             `json:"decimals"`
	Chain    string          `json:"chain"`
	LogoURI  string          `json:"logo_uri"`
	Price    decimal.Decimal `json:"price"`
}

// RouteStep represents a step in a swap route
type RouteStep struct {
	Protocol   string          `json:"protocol"`
	Pool       string          `json:"pool"`
	TokenIn    Token           `json:"token_in"`
	TokenOut   Token           `json:"token_out"`
	AmountIn   decimal.Decimal `json:"amount_in"`
	AmountOut  decimal.Decimal `json:"amount_out"`
	Fee        decimal.Decimal `json:"fee"`
	Percentage decimal.Decimal `json:"percentage"`
}

// RouteRecommendation provides routing recommendations
type RouteRecommendation struct {
	Strategy     RoutingStrategy        `json:"strategy"`
	Reason       string                 `json:"reason"`
	Confidence   decimal.Decimal        `json:"confidence"`
	Alternatives []*SwapQuote           `json:"alternatives"`
	RiskLevel    string                 `json:"risk_level"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedQuote represents a cached quote
type CachedQuote struct {
	Quote     *AggregatedQuote `json:"quote"`
	ExpiresAt time.Time        `json:"expires_at"`
}

// QuoteRequest represents a quote request
type QuoteRequest struct {
	TokenIn         Token                  `json:"token_in"`
	TokenOut        Token                  `json:"token_out"`
	AmountIn        decimal.Decimal        `json:"amount_in"`
	Chain           string                 `json:"chain"`
	Slippage        decimal.Decimal        `json:"slippage"`
	UserAddress     string                 `json:"user_address"`
	GasPrice        decimal.Decimal        `json:"gas_price"`
	Deadline        time.Time              `json:"deadline"`
	PreferredRoute  string                 `json:"preferred_route"`
	ExcludeRoutes   []string               `json:"exclude_routes"`
	IncludeGasPrice bool                   `json:"include_gas_price"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SwapRequest represents a swap execution request
type SwapRequest struct {
	Quote       *SwapQuote             `json:"quote"`
	UserAddress string                 `json:"user_address"`
	Slippage    decimal.Decimal        `json:"slippage"`
	Deadline    time.Time              `json:"deadline"`
	GasPrice    decimal.Decimal        `json:"gas_price"`
	GasLimit    uint64                 `json:"gas_limit"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SwapResult represents the result of a swap execution
type SwapResult struct {
	TransactionHash string                 `json:"transaction_hash"`
	Status          string                 `json:"status"`
	AmountIn        decimal.Decimal        `json:"amount_in"`
	AmountOut       decimal.Decimal        `json:"amount_out"`
	GasUsed         uint64                 `json:"gas_used"`
	GasCost         decimal.Decimal        `json:"gas_cost"`
	ExecutionTime   time.Duration          `json:"execution_time"`
	BlockNumber     uint64                 `json:"block_number"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewDEXAggregator creates a new DEX aggregator
func NewDEXAggregator(logger *logger.Logger, config DEXAggregatorConfig) *DEXAggregator {
	aggregator := &DEXAggregator{
		logger:       logger.Named("dex-aggregator"),
		config:       config,
		healthStatus: make(map[string]bool),
		priceCache:   make(map[string]*CachedQuote),
	}

	// Initialize aggregator clients
	if config.Aggregators["1inch"].Enabled {
		aggregator.oneInchClient = NewOneInchClient(logger, config.Aggregators["1inch"])
	}
	if config.Aggregators["paraswap"].Enabled {
		aggregator.paraswapClient = NewParaswapClient(logger, config.Aggregators["paraswap"])
	}
	if config.Aggregators["0x"].Enabled {
		aggregator.zeroXClient = NewZeroXClient(logger, config.Aggregators["0x"])
	}
	if config.Aggregators["matcha"].Enabled {
		aggregator.matchaClient = NewMatchaClient(logger, config.Aggregators["matcha"])
	}

	// Initialize routing engine
	aggregator.routingEngine = NewRoutingEngine(logger, config.RoutingConfig)

	return aggregator
}

// Start starts the DEX aggregator
func (da *DEXAggregator) Start(ctx context.Context) error {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	if da.isRunning {
		return fmt.Errorf("DEX aggregator is already running")
	}

	if !da.config.Enabled {
		da.logger.Info("DEX aggregator is disabled")
		return nil
	}

	da.logger.Info("Starting DEX aggregator")

	// Start health monitoring
	go da.monitorHealth(ctx)

	// Start cache cleanup
	go da.cleanupCache(ctx)

	da.isRunning = true
	da.lastUpdate = time.Now()

	da.logger.Info("DEX aggregator started successfully")
	return nil
}

// Stop stops the DEX aggregator
func (da *DEXAggregator) Stop() error {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	if !da.isRunning {
		return nil
	}

	da.logger.Info("Stopping DEX aggregator")
	da.isRunning = false

	da.logger.Info("DEX aggregator stopped")
	return nil
}

// GetAggregatedQuote gets quotes from multiple aggregators and returns the best options
func (da *DEXAggregator) GetAggregatedQuote(ctx context.Context, req *QuoteRequest) (*AggregatedQuote, error) {
	da.logger.Info("Getting aggregated quote",
		zap.String("token_in", req.TokenIn.Symbol),
		zap.String("token_out", req.TokenOut.Symbol),
		zap.String("amount_in", req.AmountIn.String()),
		zap.String("chain", req.Chain))

	// Check cache first
	cacheKey := da.generateCacheKey(req)
	if cached := da.getCachedQuote(cacheKey); cached != nil {
		da.logger.Debug("Returning cached quote", zap.String("cache_key", cacheKey))
		return cached, nil
	}

	// Get quotes from all enabled aggregators
	quotes, err := da.getQuotesFromAllAggregators(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes available")
	}

	// Find best quote using routing engine
	bestQuote := da.routingEngine.SelectBestQuote(quotes, req)

	// Calculate savings
	savings, savingsPercentage := da.calculateSavings(quotes, bestQuote)

	// Generate recommendation
	recommendation := da.routingEngine.GenerateRecommendation(quotes, bestQuote, req)

	aggregatedQuote := &AggregatedQuote{
		ID:                da.generateQuoteID(),
		TokenIn:           req.TokenIn,
		TokenOut:          req.TokenOut,
		AmountIn:          req.AmountIn,
		BestQuote:         bestQuote,
		AllQuotes:         quotes,
		Savings:           savings,
		SavingsPercentage: savingsPercentage,
		RecommendedRoute:  recommendation,
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(da.config.CacheTimeout),
	}

	// Cache the result
	da.cacheQuote(cacheKey, aggregatedQuote)

	da.logger.Info("Generated aggregated quote",
		zap.String("quote_id", aggregatedQuote.ID),
		zap.String("best_aggregator", bestQuote.Aggregator),
		zap.String("amount_out", bestQuote.AmountOut.String()),
		zap.String("savings", savings.String()))

	return aggregatedQuote, nil
}

// ExecuteSwap executes a swap using the specified quote
func (da *DEXAggregator) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	da.logger.Info("Executing swap",
		zap.String("quote_id", req.Quote.ID),
		zap.String("aggregator", req.Quote.Aggregator),
		zap.String("user_address", req.UserAddress))

	startTime := time.Now()

	var result *SwapResult
	var err error

	// Route to appropriate aggregator
	switch req.Quote.Aggregator {
	case "1inch":
		if da.oneInchClient != nil {
			result, err = da.oneInchClient.ExecuteSwap(ctx, req)
		} else {
			err = fmt.Errorf("1inch client not available")
		}
	case "paraswap":
		if da.paraswapClient != nil {
			result, err = da.paraswapClient.ExecuteSwap(ctx, req)
		} else {
			err = fmt.Errorf("paraswap client not available")
		}
	case "0x":
		if da.zeroXClient != nil {
			result, err = da.zeroXClient.ExecuteSwap(ctx, req)
		} else {
			err = fmt.Errorf("0x client not available")
		}
	case "matcha":
		if da.matchaClient != nil {
			result, err = da.matchaClient.ExecuteSwap(ctx, req)
		} else {
			err = fmt.Errorf("matcha client not available")
		}
	default:
		err = fmt.Errorf("unsupported aggregator: %s", req.Quote.Aggregator)
	}

	if err != nil {
		da.logger.Error("Swap execution failed",
			zap.String("aggregator", req.Quote.Aggregator),
			zap.Error(err))
		return nil, fmt.Errorf("swap execution failed: %w", err)
	}

	result.ExecutionTime = time.Since(startTime)

	da.logger.Info("Swap executed successfully",
		zap.String("transaction_hash", result.TransactionHash),
		zap.String("amount_out", result.AmountOut.String()),
		zap.Duration("execution_time", result.ExecutionTime))

	return result, nil
}

// GetSupportedTokens returns supported tokens for all aggregators
func (da *DEXAggregator) GetSupportedTokens(ctx context.Context, chain string) ([]Token, error) {
	da.logger.Debug("Getting supported tokens", zap.String("chain", chain))

	tokenMap := make(map[string]Token)

	// Collect tokens from all aggregators
	if da.oneInchClient != nil {
		tokens, err := da.oneInchClient.GetSupportedTokens(ctx, chain)
		if err == nil {
			for _, token := range tokens {
				tokenMap[token.Address] = token
			}
		}
	}

	if da.paraswapClient != nil {
		tokens, err := da.paraswapClient.GetSupportedTokens(ctx, chain)
		if err == nil {
			for _, token := range tokens {
				tokenMap[token.Address] = token
			}
		}
	}

	if da.zeroXClient != nil {
		tokens, err := da.zeroXClient.GetSupportedTokens(ctx, chain)
		if err == nil {
			for _, token := range tokens {
				tokenMap[token.Address] = token
			}
		}
	}

	if da.matchaClient != nil {
		tokens, err := da.matchaClient.GetSupportedTokens(ctx, chain)
		if err == nil {
			for _, token := range tokens {
				tokenMap[token.Address] = token
			}
		}
	}

	// Convert map to slice
	tokens := make([]Token, 0, len(tokenMap))
	for _, token := range tokenMap {
		tokens = append(tokens, token)
	}

	da.logger.Info("Retrieved supported tokens",
		zap.String("chain", chain),
		zap.Int("token_count", len(tokens)))

	return tokens, nil
}

// GetAggregatorStatus returns the status of all aggregators
func (da *DEXAggregator) GetAggregatorStatus() map[string]interface{} {
	da.mutex.RLock()
	defer da.mutex.RUnlock()

	status := map[string]interface{}{
		"is_running":    da.isRunning,
		"last_update":   da.lastUpdate,
		"cache_size":    len(da.priceCache),
		"health_status": da.healthStatus,
		"aggregators":   make(map[string]interface{}),
	}

	// Add individual aggregator status
	if da.oneInchClient != nil {
		status["aggregators"].(map[string]interface{})["1inch"] = da.oneInchClient.GetStatus()
	}
	if da.paraswapClient != nil {
		status["aggregators"].(map[string]interface{})["paraswap"] = da.paraswapClient.GetStatus()
	}
	if da.zeroXClient != nil {
		status["aggregators"].(map[string]interface{})["0x"] = da.zeroXClient.GetStatus()
	}
	if da.matchaClient != nil {
		status["aggregators"].(map[string]interface{})["matcha"] = da.matchaClient.GetStatus()
	}

	return status
}

// Helper methods

// getQuotesFromAllAggregators gets quotes from all enabled aggregators
func (da *DEXAggregator) getQuotesFromAllAggregators(ctx context.Context, req *QuoteRequest) ([]*SwapQuote, error) {
	var quotes []*SwapQuote
	quoteChan := make(chan *SwapQuote, 4)
	errorChan := make(chan error, 4)
	activeRequests := 0

	// Get quotes from all enabled aggregators in parallel
	if da.oneInchClient != nil && da.config.Aggregators["1inch"].Enabled {
		activeRequests++
		go func() {
			quote, err := da.oneInchClient.GetQuote(ctx, req)
			if err != nil {
				errorChan <- fmt.Errorf("1inch: %w", err)
			} else {
				quoteChan <- quote
			}
		}()
	}

	if da.paraswapClient != nil && da.config.Aggregators["paraswap"].Enabled {
		activeRequests++
		go func() {
			quote, err := da.paraswapClient.GetQuote(ctx, req)
			if err != nil {
				errorChan <- fmt.Errorf("paraswap: %w", err)
			} else {
				quoteChan <- quote
			}
		}()
	}

	if da.zeroXClient != nil && da.config.Aggregators["0x"].Enabled {
		activeRequests++
		go func() {
			quote, err := da.zeroXClient.GetQuote(ctx, req)
			if err != nil {
				errorChan <- fmt.Errorf("0x: %w", err)
			} else {
				quoteChan <- quote
			}
		}()
	}

	if da.matchaClient != nil && da.config.Aggregators["matcha"].Enabled {
		activeRequests++
		go func() {
			quote, err := da.matchaClient.GetQuote(ctx, req)
			if err != nil {
				errorChan <- fmt.Errorf("matcha: %w", err)
			} else {
				quoteChan <- quote
			}
		}()
	}

	// Collect results
	for i := 0; i < activeRequests; i++ {
		select {
		case quote := <-quoteChan:
			quotes = append(quotes, quote)
		case err := <-errorChan:
			da.logger.Warn("Failed to get quote from aggregator", zap.Error(err))
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return quotes, nil
}

// calculateSavings calculates savings compared to other quotes
func (da *DEXAggregator) calculateSavings(quotes []*SwapQuote, bestQuote *SwapQuote) (decimal.Decimal, decimal.Decimal) {
	if len(quotes) <= 1 {
		return decimal.Zero, decimal.Zero
	}

	// Find second best quote
	var secondBest *SwapQuote
	for _, quote := range quotes {
		if quote.ID != bestQuote.ID {
			if secondBest == nil || quote.AmountOut.GreaterThan(secondBest.AmountOut) {
				secondBest = quote
			}
		}
	}

	if secondBest == nil {
		return decimal.Zero, decimal.Zero
	}

	savings := bestQuote.AmountOut.Sub(secondBest.AmountOut)
	savingsPercentage := savings.Div(secondBest.AmountOut).Mul(decimal.NewFromInt(100))

	return savings, savingsPercentage
}

// generateCacheKey generates a cache key for a quote request
func (da *DEXAggregator) generateCacheKey(req *QuoteRequest) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s",
		req.TokenIn.Address,
		req.TokenOut.Address,
		req.AmountIn.String(),
		req.Chain,
		req.Slippage.String())
}

// getCachedQuote retrieves a cached quote
func (da *DEXAggregator) getCachedQuote(key string) *AggregatedQuote {
	da.mutex.RLock()
	defer da.mutex.RUnlock()

	cached, exists := da.priceCache[key]
	if !exists {
		return nil
	}

	if time.Now().After(cached.ExpiresAt) {
		delete(da.priceCache, key)
		return nil
	}

	return cached.Quote
}

// cacheQuote caches a quote
func (da *DEXAggregator) cacheQuote(key string, quote *AggregatedQuote) {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	da.priceCache[key] = &CachedQuote{
		Quote:     quote,
		ExpiresAt: quote.ExpiresAt,
	}
}

// generateQuoteID generates a unique quote ID
func (da *DEXAggregator) generateQuoteID() string {
	return fmt.Sprintf("quote_%d", time.Now().UnixNano())
}

// monitorHealth monitors the health of aggregators
func (da *DEXAggregator) monitorHealth(ctx context.Context) {
	ticker := time.NewTicker(da.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			da.performHealthCheck()
		}
	}
}

// performHealthCheck performs health checks on all aggregators
func (da *DEXAggregator) performHealthCheck() {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	// Check 1inch
	if da.oneInchClient != nil {
		da.healthStatus["1inch"] = da.checkAggregatorHealth("1inch")
	}

	// Check Paraswap
	if da.paraswapClient != nil {
		da.healthStatus["paraswap"] = da.checkAggregatorHealth("paraswap")
	}

	// Check 0x
	if da.zeroXClient != nil {
		da.healthStatus["0x"] = da.checkAggregatorHealth("0x")
	}

	// Check Matcha
	if da.matchaClient != nil {
		da.healthStatus["matcha"] = da.checkAggregatorHealth("matcha")
	}

	da.lastUpdate = time.Now()
}

// checkAggregatorHealth checks the health of a specific aggregator
func (da *DEXAggregator) checkAggregatorHealth(aggregator string) bool {
	// Simplified health check - in production would make actual API calls
	return true
}

// cleanupCache cleans up expired cache entries
func (da *DEXAggregator) cleanupCache(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			da.performCacheCleanup()
		}
	}
}

// performCacheCleanup removes expired cache entries
func (da *DEXAggregator) performCacheCleanup() {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	now := time.Now()
	for key, cached := range da.priceCache {
		if now.After(cached.ExpiresAt) {
			delete(da.priceCache, key)
		}
	}
}
