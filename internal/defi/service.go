package defi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/redis"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Service provides DeFi protocol operations
type Service struct {
	ethClient     blockchain.EthereumClient
	bscClient     blockchain.EthereumClient
	polygonClient blockchain.EthereumClient
	solanaClient  blockchain.SolanaClient
	cache         redis.Client
	logger        *logger.Logger
	config        config.DeFiConfig

	// Protocol clients
	uniswapClient   *UniswapClient
	aaveClient      *AaveClient
	chainlinkClient *ChainlinkClient
	oneInchClient   *OneInchClient

	// Solana DeFi clients
	raydiumClient *RaydiumClient
	jupiterClient *JupiterClient

	// Trading components
	arbitrageDetector *ArbitrageDetector
	yieldAggregator   *YieldAggregator
	onchainAnalyzer   *OnChainAnalyzer
	tradingBots       map[string]*TradingBotImpl

	// State
	mutex sync.RWMutex
}

// NewService creates a new DeFi service with the expected signature for tests
func NewService(
	ethClient blockchain.EthereumClient,
	bscClient blockchain.EthereumClient,
	polygonClient blockchain.EthereumClient,
	cache redis.Client,
	logger *logger.Logger,
	config config.DeFiConfig,
) *Service {
	service := &Service{
		ethClient:     ethClient,
		bscClient:     bscClient,
		polygonClient: polygonClient,
		cache:         cache,
		logger:        logger.Named("defi"),
		config:        config,
		tradingBots:   make(map[string]*TradingBotImpl),
	}

	// Initialize protocol clients
	service.uniswapClient = NewUniswapClient(ethClient, logger)
	service.aaveClient = NewAaveClient(ethClient, logger)
	service.chainlinkClient = NewChainlinkClient(ethClient, logger)
	service.oneInchClient = NewOneInchClient(config.OneInchAPIKey, logger)

	// Initialize trading components
	// Create price providers for arbitrage detection
	priceProviders := []PriceProvider{
		NewUniswapPriceProvider(service.uniswapClient),
		NewOneInchPriceProvider(service.oneInchClient),
	}

	service.arbitrageDetector = NewArbitrageDetector(
		logger,
		cache,
		priceProviders,
	)

	service.yieldAggregator = NewYieldAggregator(
		logger,
		cache,
		service.uniswapClient,
		service.aaveClient,
	)

	service.onchainAnalyzer = NewOnChainAnalyzer(
		logger,
		cache,
		ethClient,
		bscClient,
		polygonClient,
	)

	return service
}

// NewServiceWithSolana creates a new DeFi service with Solana support
func NewServiceWithSolana(
	ethClient blockchain.EthereumClient,
	bscClient blockchain.EthereumClient,
	polygonClient blockchain.EthereumClient,
	solanaClient blockchain.SolanaClient,
	raydiumClient *RaydiumClient,
	jupiterClient *JupiterClient,
	cache redis.Client,
	logger *logger.Logger,
	config config.DeFiConfig,
) *Service {
	service := NewService(ethClient, bscClient, polygonClient, cache, logger, config)
	
	// Add Solana support
	service.solanaClient = solanaClient
	service.raydiumClient = raydiumClient
	service.jupiterClient = jupiterClient

	return service
}

// GetTokenPrice retrieves the current price of a token
func (s *Service) GetTokenPrice(ctx context.Context, req *GetTokenPriceRequest) (*GetTokenPriceResponse, error) {
	s.logger.Info("Getting token price for token %s on chain %s", req.TokenAddress, string(req.Chain))

	// Check cache first
	cacheKey := fmt.Sprintf("token_price:%s:%s", req.Chain, req.TokenAddress)
	if cachedPrice, err := s.cache.Get(ctx, cacheKey); err == nil {
		var price decimal.Decimal
		if err := price.UnmarshalText([]byte(cachedPrice)); err == nil {
			token := Token{
				Address: req.TokenAddress,
				Chain:   req.Chain,
			}
			return &GetTokenPriceResponse{
				Token: token,
				Price: price,
			}, nil
		}
	}

	// Get price from Chainlink oracle
	price, err := s.chainlinkClient.GetTokenPrice(ctx, req.TokenAddress)
	if err != nil {
		s.logger.Error("Failed to get token price from Chainlink: %v", err)
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	// Cache the price for 1 minute
	priceBytes, _ := price.MarshalText()
	s.cache.Set(ctx, cacheKey, string(priceBytes), time.Minute)

	token := Token{
		Address: req.TokenAddress,
		Chain:   req.Chain,
		Price:   price,
	}

	return &GetTokenPriceResponse{
		Token: token,
		Price: price,
	}, nil
}

// GetSwapQuote gets a quote for token swap
func (s *Service) GetSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*GetSwapQuoteResponse, error) {
	s.logger.Info("Getting swap quote for %s -> %s, amount: %s on chain %s",
		req.TokenIn, req.TokenOut, req.AmountIn.String(), string(req.Chain))

	var quote *SwapQuote
	var err error

	switch req.Chain {
	case ChainEthereum:
		// Try Uniswap first
		quote, err = s.uniswapClient.GetSwapQuote(ctx, req)
		if err != nil {
			s.logger.Warn("Uniswap quote failed, trying 1inch: %v", err)
			// Fallback to 1inch
			quote, err = s.oneInchClient.GetSwapQuote(ctx, req)
		}
	case ChainBSC:
		// Use PancakeSwap for BSC
		quote, err = s.getPancakeSwapQuote(ctx, req)
	case ChainPolygon:
		// Use QuickSwap for Polygon
		quote, err = s.getQuickSwapQuote(ctx, req)
	case ChainSolana:
		// Use Jupiter for Solana
		quote, err = s.getSolanaSwapQuote(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported chain: %s", req.Chain)
	}

	if err != nil {
		s.logger.Error("Failed to get swap quote: %v", err)
		return nil, fmt.Errorf("failed to get swap quote: %w", err)
	}

	return &GetSwapQuoteResponse{
		Quote: *quote,
	}, nil
}

// ExecuteSwap executes a token swap
func (s *Service) ExecuteSwap(ctx context.Context, req *ExecuteSwapRequest) (*ExecuteSwapResponse, error) {
	s.logger.Info("Executing swap for quote %s, user %s", req.QuoteID, req.UserID)

	// Get quote from cache
	quote, err := s.getQuoteFromCache(ctx, req.QuoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Check if quote is still valid
	if time.Now().After(quote.ExpiresAt) {
		return nil, fmt.Errorf("quote has expired")
	}

	// Execute swap based on protocol
	var txHash string
	switch quote.Protocol {
	case ProtocolTypeUniswap:
		txHash, err = s.uniswapClient.ExecuteSwap(ctx, quote, req.WalletID, req.Passphrase)
	case ProtocolType1inch:
		txHash, err = s.oneInchClient.ExecuteSwap(ctx, quote, req.WalletID, req.Passphrase)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", quote.Protocol)
	}

	if err != nil {
		s.logger.Error("Failed to execute swap: %v", err)
		return nil, fmt.Errorf("failed to execute swap: %w", err)
	}

	return &ExecuteSwapResponse{
		TransactionHash: txHash,
		Status:          "pending",
	}, nil
}

// GetLiquidityPools retrieves available liquidity pools
func (s *Service) GetLiquidityPools(ctx context.Context, req *GetLiquidityPoolsRequest) (*GetLiquidityPoolsResponse, error) {
	s.logger.Info("Getting liquidity pools for chain %s, protocol %s", string(req.Chain), string(req.Protocol))

	var pools []LiquidityPool
	var err error

	switch req.Protocol {
	case ProtocolTypeUniswap:
		pools, err = s.uniswapClient.GetLiquidityPools(ctx, req)
	case ProtocolTypePancakeSwap:
		pools, err = s.getPancakeSwapPools(ctx, req)
	case ProtocolTypeQuickSwap:
		pools, err = s.getQuickSwapPools(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", req.Protocol)
	}

	if err != nil {
		s.logger.Error("Failed to get liquidity pools: %v", err)
		return nil, fmt.Errorf("failed to get liquidity pools: %w", err)
	}

	return &GetLiquidityPoolsResponse{
		Pools: pools,
		Total: len(pools),
	}, nil
}

// Start starts all trading components
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting DeFi service with trading components")

	// Start arbitrage detector
	if err := s.arbitrageDetector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start arbitrage detector: %w", err)
	}

	// Start yield aggregator
	if err := s.yieldAggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start yield aggregator: %w", err)
	}

	// Start on-chain analyzer
	if err := s.onchainAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start on-chain analyzer: %w", err)
	}

	s.logger.Info("All trading components started successfully")
	return nil
}

// Stop stops all trading components
func (s *Service) Stop() {
	s.logger.Info("Stopping DeFi service")

	// Stop arbitrage detector
	s.arbitrageDetector.Stop()

	// Stop yield aggregator
	s.yieldAggregator.Stop()

	// Stop on-chain analyzer
	s.onchainAnalyzer.Stop()

	// Stop all trading bots
	s.mutex.Lock()
	for _, bot := range s.tradingBots {
		bot.Stop()
	}
	s.mutex.Unlock()

	s.logger.Info("DeFi service stopped")
}

// Arbitrage Methods

// GetArbitrageOpportunities returns current arbitrage opportunities
func (s *Service) GetArbitrageOpportunities(ctx context.Context) ([]*ArbitrageDetection, error) {
	return s.arbitrageDetector.GetOpportunities(ctx)
}

// DetectArbitrageForToken detects arbitrage opportunities for a specific token
func (s *Service) DetectArbitrageForToken(ctx context.Context, token Token) ([]*ArbitrageDetection, error) {
	return s.arbitrageDetector.DetectArbitrageForToken(ctx, token)
}

// Yield Farming Methods

// GetBestYieldOpportunities returns the best yield farming opportunities
func (s *Service) GetBestYieldOpportunities(ctx context.Context, limit int) ([]*YieldFarmingOpportunity, error) {
	return s.yieldAggregator.GetBestOpportunities(ctx, limit)
}

// GetOptimalYieldStrategy returns the optimal yield strategy for given parameters
func (s *Service) GetOptimalYieldStrategy(ctx context.Context, req *OptimalStrategyRequest) (*YieldStrategy, error) {
	return s.yieldAggregator.GetOptimalStrategy(ctx, req)
}

// On-Chain Analysis Methods

// GetOnChainMetrics returns on-chain metrics for a token
func (s *Service) GetOnChainMetrics(ctx context.Context, tokenAddress string) (*OnChainMetrics, error) {
	return s.onchainAnalyzer.GetMetrics(ctx, tokenAddress)
}

// GetMarketSignals returns current market signals
func (s *Service) GetMarketSignals(ctx context.Context) ([]*MarketSignal, error) {
	return s.onchainAnalyzer.GetMarketSignals(ctx)
}

// GetWhaleActivity returns recent whale activity
func (s *Service) GetWhaleActivity(ctx context.Context) ([]*WhaleWatch, error) {
	return s.onchainAnalyzer.GetWhaleActivity(ctx)
}

// GetTokenAnalysis returns comprehensive analysis for a token
func (s *Service) GetTokenAnalysis(ctx context.Context, tokenAddress string) (*TokenAnalysis, error) {
	return s.onchainAnalyzer.GetTokenAnalysis(ctx, tokenAddress)
}

// Trading Bot Methods

// CreateTradingBot creates a new trading bot
func (s *Service) CreateTradingBot(ctx context.Context, name string, strategy TradingStrategyType, config TradingBotConfig) (*TradingBotImpl, error) {
	s.logger.Info("Creating trading bot %s with strategy %s", name, string(strategy))

	bot := NewTradingBot(
		name,
		strategy,
		config,
		s.logger,
		s.cache,
		s.arbitrageDetector,
		s.yieldAggregator,
		s.uniswapClient,
		s.oneInchClient,
		s.aaveClient,
	)

	s.mutex.Lock()
	s.tradingBots[bot.ID] = bot
	s.mutex.Unlock()

	return bot, nil
}

// GetTradingBot returns a trading bot by ID
func (s *Service) GetTradingBot(ctx context.Context, botID string) (*TradingBotImpl, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bot, exists := s.tradingBots[botID]
	if !exists {
		return nil, fmt.Errorf("trading bot not found: %s", botID)
	}

	return bot, nil
}

// GetAllTradingBots returns all trading bots
func (s *Service) GetAllTradingBots(ctx context.Context) ([]*TradingBotImpl, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bots := make([]*TradingBotImpl, 0, len(s.tradingBots))
	for _, bot := range s.tradingBots {
		bots = append(bots, bot)
	}

	return bots, nil
}

// StartTradingBot starts a trading bot
func (s *Service) StartTradingBot(ctx context.Context, botID string) error {
	bot, err := s.GetTradingBot(ctx, botID)
	if err != nil {
		return err
	}

	return bot.Start(ctx)
}

// StopTradingBot stops a trading bot
func (s *Service) StopTradingBot(ctx context.Context, botID string) error {
	bot, err := s.GetTradingBot(ctx, botID)
	if err != nil {
		return err
	}

	return bot.Stop()
}

// DeleteTradingBot deletes a trading bot
func (s *Service) DeleteTradingBot(ctx context.Context, botID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bot, exists := s.tradingBots[botID]
	if !exists {
		return fmt.Errorf("trading bot not found: %s", botID)
	}

	// Stop the bot first
	bot.Stop()

	// Remove from map
	delete(s.tradingBots, botID)

	s.logger.Info("Trading bot deleted: %s", botID)
	return nil
}

// GetTradingBotPerformance returns performance metrics for a trading bot
func (s *Service) GetTradingBotPerformance(ctx context.Context, botID string) (*TradingPerformance, error) {
	bot, err := s.GetTradingBot(ctx, botID)
	if err != nil {
		return nil, err
	}

	performance := bot.GetPerformance()
	return &performance, nil
}

// Helper methods for different protocols

// getQuoteFromCache retrieves a quote from cache
func (s *Service) getQuoteFromCache(ctx context.Context, quoteID string) (*SwapQuote, error) {
	// Mock implementation - in real scenario, get from Redis cache
	return nil, fmt.Errorf("quote not found: %s", quoteID)
}

// getPancakeSwapQuote gets a quote from PancakeSwap
func (s *Service) getPancakeSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*SwapQuote, error) {
	// Mock implementation for PancakeSwap
	return &SwapQuote{
		ID:           uuid.New().String(),
		Protocol:     ProtocolTypePancakeSwap,
		Chain:        req.Chain,
		AmountOut:    req.AmountIn.Mul(decimal.NewFromFloat(0.997)), // 0.3% fee
		MinAmountOut: req.AmountIn.Mul(decimal.NewFromFloat(0.99)),   // 1% slippage
		PriceImpact:  decimal.NewFromFloat(0.003),
		Fee:          decimal.NewFromFloat(0.0025),
		GasEstimate:  120000,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
	}, nil
}

// getQuickSwapQuote gets a quote from QuickSwap
func (s *Service) getQuickSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*SwapQuote, error) {
	// Mock implementation for QuickSwap
	return &SwapQuote{
		ID:           uuid.New().String(),
		Protocol:     ProtocolTypeQuickSwap,
		Chain:        req.Chain,
		AmountOut:    req.AmountIn.Mul(decimal.NewFromFloat(0.997)), // 0.3% fee
		MinAmountOut: req.AmountIn.Mul(decimal.NewFromFloat(0.99)),   // 1% slippage
		PriceImpact:  decimal.NewFromFloat(0.003),
		Fee:          decimal.NewFromFloat(0.003),
		GasEstimate:  100000,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
	}, nil
}

// getSolanaSwapQuote gets a quote for Solana swaps
func (s *Service) getSolanaSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*SwapQuote, error) {
	if s.jupiterClient == nil {
		return nil, fmt.Errorf("Jupiter client not initialized")
	}

	// Mock implementation for Jupiter
	return &SwapQuote{
		ID:           uuid.New().String(),
		Protocol:     ProtocolType1inch, // Using 1inch as placeholder
		Chain:        req.Chain,
		AmountOut:    req.AmountIn.Mul(decimal.NewFromFloat(0.995)), // 0.5% slippage
		MinAmountOut: req.AmountIn.Mul(decimal.NewFromFloat(0.99)),   // 1% slippage
		PriceImpact:  decimal.NewFromFloat(0.005),
		Fee:          decimal.Zero,
		GasEstimate:  5000, // Solana compute units
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
	}, nil
}

// getPancakeSwapPools gets liquidity pools from PancakeSwap
func (s *Service) getPancakeSwapPools(ctx context.Context, req *GetLiquidityPoolsRequest) ([]LiquidityPool, error) {
	// Mock implementation
	return []LiquidityPool{}, nil
}

// getQuickSwapPools gets liquidity pools from QuickSwap
func (s *Service) getQuickSwapPools(ctx context.Context, req *GetLiquidityPoolsRequest) ([]LiquidityPool, error) {
	// Mock implementation
	return []LiquidityPool{}, nil
}
