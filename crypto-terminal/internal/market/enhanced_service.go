package market

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/exchanges"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// EnhancedService provides advanced market data operations with multi-exchange support
type EnhancedService struct {
	config           *config.Config
	db               *sql.DB
	redis            *redis.Client
	logger           *logrus.Logger
	aggregator       *exchanges.AggregationService
	
	// State management
	isRunning        bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	mutex            sync.RWMutex
	
	// Data storage
	marketSummaries  map[string]*exchanges.MarketSummary
	arbitrageOpps    []*exchanges.ArbitrageOpportunity
	lastUpdate       time.Time
}

// NewEnhancedService creates a new enhanced market service
func NewEnhancedService(
	cfg *config.Config,
	db *sql.DB,
	redis *redis.Client,
	logger *logrus.Logger,
) (*EnhancedService, error) {
	
	// Initialize exchange clients
	clients := make(map[exchanges.ExchangeType]exchanges.ExchangeClient)
	
	// Add Binance client
	if cfg.MarketData.Exchanges.Binance.APIKey != "" {
		binanceConfig := &exchanges.BinanceConfig{
			APIKey:    cfg.MarketData.Exchanges.Binance.APIKey,
			SecretKey: cfg.MarketData.Exchanges.Binance.SecretKey,
			Testnet:   cfg.MarketData.Exchanges.Binance.Testnet,
		}
		clients[exchanges.ExchangeBinance] = exchanges.NewBinanceClient(binanceConfig, logger)
	}
	
	// Add Coinbase client
	if cfg.MarketData.Exchanges.Coinbase.APIKey != "" {
		coinbaseConfig := &exchanges.CoinbaseConfig{
			APIKey:     cfg.MarketData.Exchanges.Coinbase.APIKey,
			SecretKey:  cfg.MarketData.Exchanges.Coinbase.SecretKey,
			Passphrase: cfg.MarketData.Exchanges.Coinbase.Passphrase,
			Sandbox:    cfg.MarketData.Exchanges.Coinbase.Sandbox,
		}
		clients[exchanges.ExchangeCoinbase] = exchanges.NewCoinbaseClient(coinbaseConfig, logger)
	}
	
	if len(clients) == 0 {
		return nil, fmt.Errorf("no exchange clients configured")
	}
	
	// Create aggregation configuration
	aggConfig := &exchanges.AggregationConfig{
		UpdateInterval:       30 * time.Second,
		ArbitrageThreshold:   0.5, // 0.5% minimum profit
		DataQualityThreshold: 0.7,
		MaxPriceDeviation:    0.1, // 10% max deviation
		CacheTTL:             5 * time.Minute,
		EnableArbitrage:      true,
		EnableDataValidation: true,
		Symbols:              []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"},
	}
	
	// Create aggregation service
	aggregator := exchanges.NewAggregationService(clients, aggConfig, logger)
	
	return &EnhancedService{
		config:          cfg,
		db:              db,
		redis:           redis,
		logger:          logger,
		aggregator:      aggregator,
		stopChan:        make(chan struct{}),
		marketSummaries: make(map[string]*exchanges.MarketSummary),
	}, nil
}

// Start starts the enhanced market service
func (es *EnhancedService) Start(ctx context.Context) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	
	if es.isRunning {
		return fmt.Errorf("enhanced market service is already running")
	}
	
	// Start aggregation service
	if err := es.aggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start aggregation service: %w", err)
	}
	
	es.isRunning = true
	
	// Start data collection and processing
	es.wg.Add(1)
	go es.collectMarketData(ctx)
	
	es.wg.Add(1)
	go es.processArbitrageOpportunities(ctx)
	
	es.logger.Info("Enhanced market service started")
	return nil
}

// Stop stops the enhanced market service
func (es *EnhancedService) Stop(ctx context.Context) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	
	if !es.isRunning {
		return nil
	}
	
	close(es.stopChan)
	es.wg.Wait()
	
	// Stop aggregation service
	if err := es.aggregator.Stop(ctx); err != nil {
		es.logger.Errorf("Failed to stop aggregation service: %v", err)
	}
	
	es.isRunning = false
	es.logger.Info("Enhanced market service stopped")
	return nil
}

// GetAggregatedPrice returns aggregated price data from multiple exchanges
func (es *EnhancedService) GetAggregatedPrice(ctx context.Context, symbol string) (*models.Price, error) {
	// Get aggregated ticker from exchanges
	summary, err := es.aggregator.GetAggregatedTicker(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated ticker: %w", err)
	}
	
	// Convert to models.Price format
	price := &models.Price{
		Symbol:    symbol,
		Price:     summary.WeightedPrice,
		Volume24h: summary.TotalVolume24h,
		Change24h: decimal.Zero, // Calculate from exchange data
		Timestamp: summary.Timestamp,
		Source:    "aggregated",
	}
	
	// Calculate 24h change if we have exchange data
	if len(summary.ExchangePrices) > 0 {
		totalChange := decimal.Zero
		count := 0
		for _, ticker := range summary.ExchangePrices {
			totalChange = totalChange.Add(ticker.ChangePercent24h)
			count++
		}
		if count > 0 {
			price.Change24h = totalChange.Div(decimal.NewFromInt(int64(count)))
		}
	}
	
	return price, nil
}

// GetBestPrices returns the best bid and ask prices across all exchanges
func (es *EnhancedService) GetBestPrices(ctx context.Context, symbol string) (*exchanges.MarketSummary, error) {
	return es.aggregator.GetBestPrices(ctx, symbol)
}

// GetArbitrageOpportunities returns current arbitrage opportunities
func (es *EnhancedService) GetArbitrageOpportunities(ctx context.Context) ([]*exchanges.ArbitrageOpportunity, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()
	
	// Return cached opportunities
	opportunities := make([]*exchanges.ArbitrageOpportunity, len(es.arbitrageOpps))
	copy(opportunities, es.arbitrageOpps)
	
	return opportunities, nil
}

// GetMarketSummary returns comprehensive market summary for a symbol
func (es *EnhancedService) GetMarketSummary(ctx context.Context, symbol string) (*exchanges.MarketSummary, error) {
	es.mutex.RLock()
	summary, exists := es.marketSummaries[symbol]
	es.mutex.RUnlock()
	
	if exists && time.Since(summary.Timestamp) < 30*time.Second {
		return summary, nil
	}
	
	// Get fresh data from aggregator
	return es.aggregator.GetAggregatedTicker(ctx, symbol)
}

// GetExchangeStatus returns the status of all connected exchanges
func (es *EnhancedService) GetExchangeStatus(ctx context.Context) (map[exchanges.ExchangeType]string, error) {
	return es.aggregator.GetExchangeStatus(ctx)
}

// GetDataQuality returns data quality metrics for exchanges
func (es *EnhancedService) GetDataQuality(ctx context.Context) (map[string]interface{}, error) {
	quality := make(map[string]interface{})
	
	// Get quality metrics for each exchange and symbol
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	exchanges := []exchanges.ExchangeType{exchanges.ExchangeBinance, exchanges.ExchangeCoinbase}
	
	for _, exchange := range exchanges {
		exchangeQuality := make(map[string]interface{})
		for _, symbol := range symbols {
			if metrics, err := es.aggregator.GetDataQuality(ctx, exchange, symbol); err == nil {
				exchangeQuality[symbol] = map[string]interface{}{
					"quality_score":    metrics.QualityScore,
					"availability":     metrics.Availability,
					"latency_ms":       metrics.Latency.Milliseconds(),
					"error_rate":       metrics.ErrorRate,
					"last_update":      metrics.LastUpdate,
				}
			}
		}
		quality[string(exchange)] = exchangeQuality
	}
	
	return quality, nil
}

// GetAggregatedOrderBook returns aggregated order book from multiple exchanges
func (es *EnhancedService) GetAggregatedOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	return es.aggregator.GetAggregatedOrderBook(ctx, symbol, depth)
}

// collectMarketData continuously collects and caches market data
func (es *EnhancedService) collectMarketData(ctx context.Context) {
	defer es.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-es.stopChan:
			return
		case <-ticker.C:
			es.updateMarketSummaries(ctx, symbols)
		}
	}
}

// updateMarketSummaries updates market summaries for all symbols
func (es *EnhancedService) updateMarketSummaries(ctx context.Context, symbols []string) {
	for _, symbol := range symbols {
		summary, err := es.aggregator.GetAggregatedTicker(ctx, symbol)
		if err != nil {
			es.logger.Errorf("Failed to get market summary for %s: %v", symbol, err)
			continue
		}
		
		es.mutex.Lock()
		es.marketSummaries[symbol] = summary
		es.lastUpdate = time.Now()
		es.mutex.Unlock()
		
		// Cache in Redis
		if data, err := json.Marshal(summary); err == nil {
			cacheKey := fmt.Sprintf("market:summary:%s", symbol)
			es.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		}
	}
}

// processArbitrageOpportunities continuously monitors for arbitrage opportunities
func (es *EnhancedService) processArbitrageOpportunities(ctx context.Context) {
	defer es.wg.Done()
	
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-es.stopChan:
			return
		case <-ticker.C:
			opportunities, err := es.aggregator.FindArbitrageOpportunities(ctx, symbols)
			if err != nil {
				es.logger.Errorf("Failed to find arbitrage opportunities: %v", err)
				continue
			}
			
			es.mutex.Lock()
			es.arbitrageOpps = opportunities
			es.mutex.Unlock()
			
			// Log significant opportunities
			for _, opp := range opportunities {
				if opp.ProfitPercent.GreaterThan(decimal.NewFromFloat(1.0)) {
					es.logger.Infof("Significant arbitrage opportunity: %s %.2f%% profit between %s and %s",
						opp.Symbol, opp.ProfitPercent, opp.BuyExchange, opp.SellExchange)
				}
			}
			
			// Cache opportunities
			if data, err := json.Marshal(opportunities); err == nil {
				es.redis.Set(ctx, "market:arbitrage:opportunities", data, 2*time.Minute)
			}
		}
	}
}
