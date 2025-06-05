package defi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// OnChainAnalyzer analyzes on-chain data for trading insights
type OnChainAnalyzer struct {
	logger        *logger.Logger
	cache         redis.Client
	ethClient     *blockchain.EthereumClient
	bscClient     *blockchain.EthereumClient
	polygonClient *blockchain.EthereumClient

	// Configuration
	scanInterval time.Duration
	blockRange   uint64

	// State
	metrics         map[string]*OnChainMetrics
	whaleWatches    map[string]*WhaleWatch
	liquidityEvents map[string]*LiquidityEvent
	mutex           sync.RWMutex

	// Channels
	eventChan chan *BlockchainEvent
	stopChan  chan struct{}
}



// NewOnChainAnalyzer creates a new on-chain analyzer
func NewOnChainAnalyzer(
	logger *logger.Logger,
	cache redis.Client,
	ethClient *blockchain.EthereumClient,
	bscClient *blockchain.EthereumClient,
	polygonClient *blockchain.EthereumClient,
) *OnChainAnalyzer {
	return &OnChainAnalyzer{
		logger:          logger.Named("onchain-analyzer"),
		cache:           cache,
		ethClient:       ethClient,
		bscClient:       bscClient,
		polygonClient:   polygonClient,
		scanInterval:    time.Minute * 2, // Scan every 2 minutes
		blockRange:      100,             // Analyze last 100 blocks
		metrics:         make(map[string]*OnChainMetrics),
		whaleWatches:    make(map[string]*WhaleWatch),
		liquidityEvents: make(map[string]*LiquidityEvent),
		eventChan:       make(chan *BlockchainEvent, 200),
		stopChan:        make(chan struct{}),
	}
}

// Start begins the on-chain analysis
func (oca *OnChainAnalyzer) Start(ctx context.Context) error {
	oca.logger.Info("Starting on-chain analyzer")

	// Initialize whale watches
	oca.initializeWhaleWatches()

	// Start the main scanning loop
	go oca.scanningLoop(ctx)

	// Start event processing
	go oca.eventProcessingLoop(ctx)

	// Start metrics calculation
	go oca.metricsCalculationLoop(ctx)

	return nil
}

// Stop stops the on-chain analysis
func (oca *OnChainAnalyzer) Stop() {
	oca.logger.Info("Stopping on-chain analyzer")
	close(oca.stopChan)
}

// GetMetrics returns current on-chain metrics for a token
func (oca *OnChainAnalyzer) GetMetrics(ctx context.Context, tokenAddress string) (*OnChainMetrics, error) {
	oca.mutex.RLock()
	defer oca.mutex.RUnlock()

	metrics, exists := oca.metrics[tokenAddress]
	if !exists {
		return nil, fmt.Errorf("metrics not found for token: %s", tokenAddress)
	}

	return metrics, nil
}

// GetMarketSignals returns current market signals
func (oca *OnChainAnalyzer) GetMarketSignals(ctx context.Context) ([]*MarketSignal, error) {
	// Get signals from cache
	cacheKey := "onchain:signals"
	var signals []*MarketSignal

	// Try to get from cache (simplified - in real implementation use proper cache interface)
	cachedData, err := oca.cache.Get(ctx, cacheKey)
	if err != nil || cachedData == "" {
		// Generate new signals if not in cache
		signals = oca.generateMarketSignals(ctx)

		// Cache for 5 minutes
		if err := oca.cache.Set(ctx, cacheKey, signals, time.Minute*5); err != nil {
			oca.logger.Error("Failed to cache market signals", zap.Error(err))
		}
	}

	return signals, nil
}

// GetWhaleActivity returns recent whale activity
func (oca *OnChainAnalyzer) GetWhaleActivity(ctx context.Context) ([]*WhaleWatch, error) {
	oca.mutex.RLock()
	defer oca.mutex.RUnlock()

	whales := make([]*WhaleWatch, 0, len(oca.whaleWatches))
	for _, whale := range oca.whaleWatches {
		if whale.Active && whale.TxCount24h > 0 {
			whales = append(whales, whale)
		}
	}

	return whales, nil
}

// scanningLoop runs the main blockchain scanning loop
func (oca *OnChainAnalyzer) scanningLoop(ctx context.Context) {
	ticker := time.NewTicker(oca.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-oca.stopChan:
			return
		case <-ticker.C:
			oca.scanBlockchains(ctx)
		}
	}
}

// scanBlockchains scans all supported blockchains
func (oca *OnChainAnalyzer) scanBlockchains(ctx context.Context) {
	// Scan Ethereum
	if err := oca.scanChain(ctx, ChainEthereum, oca.ethClient); err != nil {
		oca.logger.Error("Failed to scan Ethereum", zap.Error(err))
	}

	// Scan BSC
	if err := oca.scanChain(ctx, ChainBSC, oca.bscClient); err != nil {
		oca.logger.Error("Failed to scan BSC", zap.Error(err))
	}

	// Scan Polygon
	if err := oca.scanChain(ctx, ChainPolygon, oca.polygonClient); err != nil {
		oca.logger.Error("Failed to scan Polygon", zap.Error(err))
	}
}

// scanChain scans a specific blockchain
func (oca *OnChainAnalyzer) scanChain(ctx context.Context, chain Chain, client *blockchain.EthereumClient) error {
	// Get latest block number
	latestBlockBig, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	latestBlock := latestBlockBig.Uint64()

	// Calculate scan range
	var fromBlock uint64
	if latestBlock > oca.blockRange {
		fromBlock = latestBlock - oca.blockRange
	} else {
		fromBlock = 0
	}

	oca.logger.Debug("Scanning chain",
		zap.String("chain", string(chain)),
		zap.Uint64("from_block", fromBlock),
		zap.Uint64("to_block", latestBlock))

	// Scan blocks for events
	for blockNum := fromBlock; blockNum <= latestBlock; blockNum++ {
		if err := oca.scanBlock(ctx, chain, client, blockNum); err != nil {
			oca.logger.Error("Failed to scan block",
				zap.String("chain", string(chain)),
				zap.Uint64("block", blockNum),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// scanBlock scans a specific block for events
func (oca *OnChainAnalyzer) scanBlock(ctx context.Context, chain Chain, client *blockchain.EthereumClient, blockNumber uint64) error {
	// Get block data (mock implementation)
	// In real implementation, fetch actual block and transaction data

	// Mock large transfer event
	if blockNumber%50 == 0 { // Every 50th block
		event := &BlockchainEvent{
			ID:          fmt.Sprintf("%s-%d-transfer", chain, blockNumber),
			Type:        EventTypeLargeTransfer,
			Chain:       chain,
			BlockNumber: blockNumber,
			TxHash:      fmt.Sprintf("0x%d", blockNumber),
			Token: Token{
				Address: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
				Symbol:  "USDC",
				Chain:   chain,
			},
			Amount:    decimal.NewFromFloat(1000000), // $1M transfer
			From:      "0x1234567890123456789012345678901234567890",
			To:        "0x0987654321098765432109876543210987654321",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"whale": true},
		}

		select {
		case oca.eventChan <- event:
		default:
			oca.logger.Warn("Event channel full, dropping event")
		}
	}

	return nil
}

// eventProcessingLoop processes blockchain events
func (oca *OnChainAnalyzer) eventProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-oca.stopChan:
			return
		case event := <-oca.eventChan:
			oca.processEvent(ctx, event)
		}
	}
}

// processEvent processes a blockchain event
func (oca *OnChainAnalyzer) processEvent(ctx context.Context, event *BlockchainEvent) {
	oca.logger.Debug("Processing blockchain event",
		zap.String("type", string(event.Type)),
		zap.String("token", event.Token.Symbol),
		zap.String("amount", event.Amount.String()))

	switch event.Type {
	case EventTypeLargeTransfer:
		oca.processLargeTransfer(ctx, event)
	case EventTypeLiquidityAdd:
		oca.processLiquidityEvent(ctx, event)
	case EventTypeLiquidityRemove:
		oca.processLiquidityEvent(ctx, event)
	case EventTypeSwap:
		oca.processSwapEvent(ctx, event)
	}

	// Update metrics
	oca.updateMetrics(ctx, event)
}

// processLargeTransfer processes large transfer events
func (oca *OnChainAnalyzer) processLargeTransfer(ctx context.Context, event *BlockchainEvent) {
	// Check if this is a whale address
	oca.mutex.Lock()
	defer oca.mutex.Unlock()

	if whale, exists := oca.whaleWatches[event.From]; exists {
		whale.LastTx = event.Timestamp
		whale.TxCount24h++
		whale.Volume24h = whale.Volume24h.Add(event.Amount)
	}

	if whale, exists := oca.whaleWatches[event.To]; exists {
		whale.LastTx = event.Timestamp
		whale.TxCount24h++
		whale.Volume24h = whale.Volume24h.Add(event.Amount)
	}
}

// processLiquidityEvent processes liquidity events
func (oca *OnChainAnalyzer) processLiquidityEvent(ctx context.Context, event *BlockchainEvent) {
	// Track liquidity changes for market analysis
	oca.logger.Info("Liquidity event detected",
		zap.String("type", string(event.Type)),
		zap.String("token", event.Token.Symbol),
		zap.String("amount", event.Amount.String()))
}

// processSwapEvent processes swap events
func (oca *OnChainAnalyzer) processSwapEvent(ctx context.Context, event *BlockchainEvent) {
	// Track swap volume and price impact
	oca.logger.Info("Swap event detected",
		zap.String("token", event.Token.Symbol),
		zap.String("amount", event.Amount.String()))
}

// metricsCalculationLoop calculates on-chain metrics
func (oca *OnChainAnalyzer) metricsCalculationLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // Calculate every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-oca.stopChan:
			return
		case <-ticker.C:
			oca.calculateMetrics(ctx)
		}
	}
}

// calculateMetrics calculates on-chain metrics for tracked tokens
func (oca *OnChainAnalyzer) calculateMetrics(ctx context.Context) {
	// List of tokens to track
	trackedTokens := []Token{
		{
			Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			Symbol:  "WETH",
			Chain:   ChainEthereum,
		},
		{
			Address: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1",
			Symbol:  "USDC",
			Chain:   ChainEthereum,
		},
		{
			Address: "0x0000000000000000000000000000000000000000", // Coffee token
			Symbol:  "COFFEE",
			Chain:   ChainEthereum,
		},
	}

	for _, token := range trackedTokens {
		metrics := oca.calculateTokenMetrics(ctx, token)

		oca.mutex.Lock()
		oca.metrics[token.Address] = metrics
		oca.mutex.Unlock()

		// Cache metrics
		cacheKey := fmt.Sprintf("onchain:metrics:%s", token.Address)
		if err := oca.cache.Set(ctx, cacheKey, metrics, time.Minute*10); err != nil {
			oca.logger.Error("Failed to cache metrics", zap.Error(err))
		}
	}
}

// calculateTokenMetrics calculates metrics for a specific token
func (oca *OnChainAnalyzer) calculateTokenMetrics(ctx context.Context, token Token) *OnChainMetrics {
	// Mock implementation - in real scenario, aggregate from blockchain data
	return &OnChainMetrics{
		Token:           token,
		Chain:           token.Chain,
		Price:           decimal.NewFromFloat(2500.0),       // Mock price
		Volume24h:       decimal.NewFromFloat(50000000),     // $50M volume
		Liquidity:       decimal.NewFromFloat(200000000),    // $200M liquidity
		MarketCap:       decimal.NewFromFloat(300000000000), // $300B market cap
		Holders:         150000,                             // 150k holders
		Transactions24h: 25000,                              // 25k transactions
		Volatility:      decimal.NewFromFloat(0.15),         // 15% volatility
		Timestamp:       time.Now(),
	}
}

// updateMetrics updates metrics based on blockchain events
func (oca *OnChainAnalyzer) updateMetrics(ctx context.Context, event *BlockchainEvent) {
	oca.mutex.Lock()
	defer oca.mutex.Unlock()

	metrics, exists := oca.metrics[event.Token.Address]
	if !exists {
		return
	}

	// Update transaction count
	metrics.Transactions24h++

	// Update volume
	metrics.Volume24h = metrics.Volume24h.Add(event.Amount)

	// Update timestamp
	metrics.Timestamp = time.Now()
}

// generateMarketSignals generates market signals from on-chain data
func (oca *OnChainAnalyzer) generateMarketSignals(ctx context.Context) []*MarketSignal {
	var signals []*MarketSignal

	// Whale movement signals
	whaleSignals := oca.generateWhaleSignals(ctx)
	signals = append(signals, whaleSignals...)

	// Volume spike signals
	volumeSignals := oca.generateVolumeSignals(ctx)
	signals = append(signals, volumeSignals...)

	// Liquidity shift signals
	liquiditySignals := oca.generateLiquiditySignals(ctx)
	signals = append(signals, liquiditySignals...)

	return signals
}

// generateWhaleSignals generates signals based on whale activity
func (oca *OnChainAnalyzer) generateWhaleSignals(ctx context.Context) []*MarketSignal {
	var signals []*MarketSignal

	oca.mutex.RLock()
	defer oca.mutex.RUnlock()

	for _, whale := range oca.whaleWatches {
		if whale.TxCount24h > 5 { // Active whale
			signal := &MarketSignal{
				ID:         fmt.Sprintf("whale-%s-%d", whale.Address, time.Now().Unix()),
				Type:       SignalTypeWhaleMovement,
				Token:      Token{Symbol: "MIXED", Chain: whale.Chain},
				Strength:   decimal.NewFromFloat(0.7),
				Confidence: decimal.NewFromFloat(0.8),
				Direction:  oca.determineWhaleDirection(whale),
				Timeframe:  time.Hour * 24,
				Reason:     fmt.Sprintf("Whale %s made %d transactions in 24h", whale.Label, whale.TxCount24h),
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(time.Hour * 6),
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// generateVolumeSignals generates signals based on volume spikes
func (oca *OnChainAnalyzer) generateVolumeSignals(ctx context.Context) []*MarketSignal {
	var signals []*MarketSignal

	oca.mutex.RLock()
	defer oca.mutex.RUnlock()

	for _, metrics := range oca.metrics {
		// Check for volume spike (simplified)
		avgVolume := decimal.NewFromFloat(30000000) // $30M average
		if metrics.Volume24h.GreaterThan(avgVolume.Mul(decimal.NewFromFloat(2))) {
			signal := &MarketSignal{
				ID:         fmt.Sprintf("volume-%s-%d", metrics.Token.Symbol, time.Now().Unix()),
				Type:       SignalTypeVolumeSpike,
				Token:      metrics.Token,
				Strength:   decimal.NewFromFloat(0.8),
				Confidence: decimal.NewFromFloat(0.9),
				Direction:  SignalDirectionBullish,
				Timeframe:  time.Hour * 12,
				Reason:     fmt.Sprintf("Volume spike: %s vs %s average", metrics.Volume24h, avgVolume),
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(time.Hour * 4),
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// generateLiquiditySignals generates signals based on liquidity changes
func (oca *OnChainAnalyzer) generateLiquiditySignals(ctx context.Context) []*MarketSignal {
	var signals []*MarketSignal

	// Mock liquidity signal
	signal := &MarketSignal{
		ID:         fmt.Sprintf("liquidity-%d", time.Now().Unix()),
		Type:       SignalTypeLiquidityShift,
		Token:      Token{Symbol: "COFFEE", Chain: ChainEthereum},
		Strength:   decimal.NewFromFloat(0.6),
		Confidence: decimal.NewFromFloat(0.7),
		Direction:  SignalDirectionBullish,
		Timeframe:  time.Hour * 8,
		Reason:     "Increased liquidity provision in COFFEE-ETH pool",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour * 3),
	}
	signals = append(signals, signal)

	return signals
}

// initializeWhaleWatches initializes whale address monitoring
func (oca *OnChainAnalyzer) initializeWhaleWatches() {
	oca.mutex.Lock()
	defer oca.mutex.Unlock()

	// Known whale addresses (mock data)
	whales := []*WhaleWatch{
		{
			Address:    "0x1234567890123456789012345678901234567890",
			Label:      "Whale #1",
			Chain:      ChainEthereum,
			Balance:    decimal.NewFromFloat(100000000), // $100M
			TxCount24h: 0,
			Volume24h:  decimal.Zero,
			Active:     true,
		},
		{
			Address:    "0x0987654321098765432109876543210987654321",
			Label:      "Whale #2",
			Chain:      ChainEthereum,
			Balance:    decimal.NewFromFloat(75000000), // $75M
			TxCount24h: 0,
			Volume24h:  decimal.Zero,
			Active:     true,
		},
		{
			Address:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			Label:      "DeFi Protocol Treasury",
			Chain:      ChainEthereum,
			Balance:    decimal.NewFromFloat(500000000), // $500M
			TxCount24h: 0,
			Volume24h:  decimal.Zero,
			Active:     true,
		},
	}

	for _, whale := range whales {
		oca.whaleWatches[whale.Address] = whale
	}

	oca.logger.Info("Initialized whale watches", zap.Int("count", len(whales)))
}

// Helper methods

func (oca *OnChainAnalyzer) determineWhaleDirection(whale *WhaleWatch) SignalDirection {
	// Simplified logic - in real implementation, analyze transaction patterns
	if whale.Volume24h.GreaterThan(decimal.NewFromFloat(10000000)) { // > $10M
		return SignalDirectionBearish // Large outflows might indicate selling
	} else if whale.TxCount24h > 10 {
		return SignalDirectionBullish // High activity might indicate accumulation
	}
	return SignalDirectionNeutral
}

// GetTokenAnalysis returns comprehensive analysis for a token
func (oca *OnChainAnalyzer) GetTokenAnalysis(ctx context.Context, tokenAddress string) (*TokenAnalysis, error) {
	metrics, err := oca.GetMetrics(ctx, tokenAddress)
	if err != nil {
		return nil, err
	}

	// Generate analysis
	analysis := &TokenAnalysis{
		Token:           metrics.Token,
		Metrics:         *metrics,
		Signals:         []*MarketSignal{},
		WhaleActivity:   []*WhaleWatch{},
		LiquidityEvents: []*LiquidityEvent{},
		Score:           oca.calculateTokenScore(metrics),
		Recommendation:  oca.generateRecommendation(metrics),
		UpdatedAt:       time.Now(),
	}

	return analysis, nil
}



func (oca *OnChainAnalyzer) calculateTokenScore(metrics *OnChainMetrics) decimal.Decimal {
	// Simple scoring algorithm
	score := decimal.NewFromFloat(50) // Base score

	// Volume factor
	if metrics.Volume24h.GreaterThan(decimal.NewFromFloat(10000000)) {
		score = score.Add(decimal.NewFromFloat(20))
	}

	// Liquidity factor
	if metrics.Liquidity.GreaterThan(decimal.NewFromFloat(50000000)) {
		score = score.Add(decimal.NewFromFloat(15))
	}

	// Holder count factor
	if metrics.Holders > 10000 {
		score = score.Add(decimal.NewFromFloat(10))
	}

	// Volatility penalty
	if metrics.Volatility.GreaterThan(decimal.NewFromFloat(0.2)) {
		score = score.Sub(decimal.NewFromFloat(10))
	}

	// Cap at 100
	if score.GreaterThan(decimal.NewFromFloat(100)) {
		score = decimal.NewFromFloat(100)
	}

	return score
}

func (oca *OnChainAnalyzer) generateRecommendation(metrics *OnChainMetrics) string {
	score := oca.calculateTokenScore(metrics)

	if score.GreaterThan(decimal.NewFromFloat(80)) {
		return "Strong Buy - Excellent on-chain metrics"
	} else if score.GreaterThan(decimal.NewFromFloat(60)) {
		return "Buy - Good fundamentals"
	} else if score.GreaterThan(decimal.NewFromFloat(40)) {
		return "Hold - Mixed signals"
	} else {
		return "Caution - Weak on-chain activity"
	}
}
