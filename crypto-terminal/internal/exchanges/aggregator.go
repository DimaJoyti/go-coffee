package exchanges

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// AggregationService aggregates data from multiple exchanges
type AggregationService struct {
	clients       map[ExchangeType]ExchangeClient
	priceAggr     PriceAggregator
	arbDetector   ArbitrageDetector
	dataValidator DataValidator
	cacheManager  CacheManager
	eventPublisher EventPublisher
	metricsCollector MetricsCollector
	logger        *logrus.Logger
	
	// Configuration
	config *AggregationConfig
	
	// State management
	isRunning bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mutex     sync.RWMutex
	
	// Data storage
	latestTickers    map[string]map[ExchangeType]*Ticker
	latestOrderBooks map[string]map[ExchangeType]*OrderBook
	dataQuality      map[ExchangeType]map[string]*DataQualityMetrics
}

// AggregationConfig represents configuration for the aggregation service
type AggregationConfig struct {
	UpdateInterval       time.Duration `json:"update_interval"`
	ArbitrageThreshold   float64       `json:"arbitrage_threshold"`
	DataQualityThreshold float64       `json:"data_quality_threshold"`
	MaxPriceDeviation    float64       `json:"max_price_deviation"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	EnableArbitrage      bool          `json:"enable_arbitrage"`
	EnableDataValidation bool          `json:"enable_data_validation"`
	Symbols              []string      `json:"symbols"`
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(
	clients map[ExchangeType]ExchangeClient,
	config *AggregationConfig,
	logger *logrus.Logger,
) *AggregationService {
	return &AggregationService{
		clients:          clients,
		priceAggr:        NewPriceAggregator(),
		arbDetector:      NewArbitrageDetector(config.ArbitrageThreshold),
		dataValidator:    NewDataValidator(),
		config:           config,
		logger:           logger,
		stopChan:         make(chan struct{}),
		latestTickers:    make(map[string]map[ExchangeType]*Ticker),
		latestOrderBooks: make(map[string]map[ExchangeType]*OrderBook),
		dataQuality:      make(map[ExchangeType]map[string]*DataQualityMetrics),
	}
}

// Start starts the aggregation service
func (as *AggregationService) Start(ctx context.Context) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	if as.isRunning {
		return fmt.Errorf("aggregation service is already running")
	}
	
	// Connect to all exchanges
	for exchangeType, client := range as.clients {
		if err := client.Connect(ctx); err != nil {
			as.logger.Errorf("Failed to connect to %s: %v", exchangeType, err)
			continue
		}
		as.logger.Infof("Connected to %s", exchangeType)
	}
	
	as.isRunning = true
	
	// Start data collection goroutines
	as.wg.Add(1)
	go as.collectMarketData(ctx)
	
	if as.config.EnableArbitrage {
		as.wg.Add(1)
		go as.detectArbitrage(ctx)
	}
	
	as.logger.Info("Aggregation service started")
	return nil
}

// Stop stops the aggregation service
func (as *AggregationService) Stop(ctx context.Context) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	if !as.isRunning {
		return nil
	}
	
	close(as.stopChan)
	as.wg.Wait()
	
	// Disconnect from all exchanges
	for exchangeType, client := range as.clients {
		if err := client.Disconnect(ctx); err != nil {
			as.logger.Errorf("Failed to disconnect from %s: %v", exchangeType, err)
		}
	}
	
	as.isRunning = false
	as.logger.Info("Aggregation service stopped")
	return nil
}

// GetAggregatedTicker returns aggregated ticker data for a symbol
func (as *AggregationService) GetAggregatedTicker(ctx context.Context, symbol string) (*MarketSummary, error) {
	as.mutex.RLock()
	exchangeTickers, exists := as.latestTickers[symbol]
	as.mutex.RUnlock()
	
	if !exists || len(exchangeTickers) == 0 {
		return nil, fmt.Errorf("no ticker data available for symbol %s", symbol)
	}
	
	// Convert to slice for aggregation
	tickers := make([]*Ticker, 0, len(exchangeTickers))
	for _, ticker := range exchangeTickers {
		if as.config.EnableDataValidation {
			if !as.dataValidator.ValidateTicker(ticker) {
				continue
			}
		}
		tickers = append(tickers, ticker)
	}
	
	if len(tickers) == 0 {
		return nil, fmt.Errorf("no valid ticker data available for symbol %s", symbol)
	}
	
	// Calculate aggregated metrics
	weightedPrice := as.priceAggr.CalculateWeightedPrice(tickers)
	bestBid, bestAsk := as.priceAggr.CalculateBestBidAsk(tickers)
	totalVolume := as.priceAggr.CalculateTotalVolume(tickers)
	
	spread := decimal.Zero
	spreadPercent := decimal.Zero
	if bestBid != nil && bestAsk != nil {
		spread = bestAsk.Price.Sub(bestBid.Price)
		if !bestBid.Price.IsZero() {
			spreadPercent = spread.Div(bestBid.Price).Mul(decimal.NewFromInt(100))
		}
	}
	
	// Check for arbitrage opportunities
	var arbitrage *ArbitrageOpportunity
	if as.config.EnableArbitrage {
		opportunities, err := as.arbDetector.DetectOpportunities(ctx, tickers)
		if err == nil && len(opportunities) > 0 {
			arbitrage = opportunities[0] // Take the best opportunity
		}
	}
	
	// Calculate data quality score
	dataQuality := as.calculateDataQuality(symbol, exchangeTickers)
	
	return &MarketSummary{
		Symbol:         symbol,
		BestBid:        bestBid,
		BestAsk:        bestAsk,
		WeightedPrice:  weightedPrice,
		TotalVolume24h: totalVolume,
		PriceSpread:    spread,
		SpreadPercent:  spreadPercent,
		ExchangePrices: exchangeTickers,
		Arbitrage:      arbitrage,
		Timestamp:      time.Now(),
		DataQuality:    dataQuality,
	}, nil
}

// GetBestPrices returns the best bid and ask prices across all exchanges
func (as *AggregationService) GetBestPrices(ctx context.Context, symbol string) (*MarketSummary, error) {
	return as.GetAggregatedTicker(ctx, symbol)
}

// FindArbitrageOpportunities finds arbitrage opportunities across exchanges
func (as *AggregationService) FindArbitrageOpportunities(ctx context.Context, symbols []string) ([]*ArbitrageOpportunity, error) {
	var allOpportunities []*ArbitrageOpportunity
	
	for _, symbol := range symbols {
		as.mutex.RLock()
		exchangeTickers, exists := as.latestTickers[symbol]
		as.mutex.RUnlock()
		
		if !exists || len(exchangeTickers) < 2 {
			continue
		}
		
		// Convert to slice
		tickers := make([]*Ticker, 0, len(exchangeTickers))
		for _, ticker := range exchangeTickers {
			tickers = append(tickers, ticker)
		}
		
		opportunities, err := as.arbDetector.DetectOpportunities(ctx, tickers)
		if err != nil {
			as.logger.Errorf("Failed to detect arbitrage for %s: %v", symbol, err)
			continue
		}
		
		allOpportunities = append(allOpportunities, opportunities...)
	}
	
	// Sort by profit percentage
	sort.Slice(allOpportunities, func(i, j int) bool {
		return allOpportunities[i].ProfitPercent.GreaterThan(allOpportunities[j].ProfitPercent)
	})
	
	return allOpportunities, nil
}

// GetArbitrageOpportunity gets the best arbitrage opportunity for a symbol
func (as *AggregationService) GetArbitrageOpportunity(ctx context.Context, symbol string) (*ArbitrageOpportunity, error) {
	opportunities, err := as.FindArbitrageOpportunities(ctx, []string{symbol})
	if err != nil {
		return nil, err
	}
	
	if len(opportunities) == 0 {
		return nil, fmt.Errorf("no arbitrage opportunities found for %s", symbol)
	}
	
	return opportunities[0], nil
}

// GetDataQuality returns data quality metrics for an exchange and symbol
func (as *AggregationService) GetDataQuality(ctx context.Context, exchange ExchangeType, symbol string) (*DataQualityMetrics, error) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	
	if exchangeMetrics, exists := as.dataQuality[exchange]; exists {
		if metrics, exists := exchangeMetrics[symbol]; exists {
			return metrics, nil
		}
	}
	
	return nil, fmt.Errorf("no data quality metrics available for %s on %s", symbol, exchange)
}

// GetExchangeStatus returns the status of all exchanges
func (as *AggregationService) GetExchangeStatus(ctx context.Context) (map[ExchangeType]string, error) {
	status := make(map[ExchangeType]string)
	
	for exchangeType, client := range as.clients {
		status[exchangeType] = client.GetStatus()
	}
	
	return status, nil
}

// collectMarketData collects market data from all exchanges
func (as *AggregationService) collectMarketData(ctx context.Context) {
	defer as.wg.Done()
	
	ticker := time.NewTicker(as.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.updateMarketData(ctx)
		}
	}
}

// updateMarketData updates market data from all exchanges
func (as *AggregationService) updateMarketData(ctx context.Context) {
	for _, symbol := range as.config.Symbols {
		as.updateSymbolData(ctx, symbol)
	}
}

// updateSymbolData updates data for a specific symbol
func (as *AggregationService) updateSymbolData(ctx context.Context, symbol string) {
	as.mutex.Lock()
	if as.latestTickers[symbol] == nil {
		as.latestTickers[symbol] = make(map[ExchangeType]*Ticker)
	}
	as.mutex.Unlock()
	
	var wg sync.WaitGroup
	
	for exchangeType, client := range as.clients {
		wg.Add(1)
		go func(et ExchangeType, c ExchangeClient) {
			defer wg.Done()
			
			startTime := time.Now()
			ticker, err := c.GetTicker(ctx, symbol)
			latency := time.Since(startTime)
			
			if err != nil {
				as.logger.Errorf("Failed to get ticker for %s from %s: %v", symbol, et, err)
				if as.metricsCollector != nil {
					as.metricsCollector.RecordError(et, "get_ticker", err)
				}
				return
			}
			
			// Update data quality metrics
			as.updateDataQuality(et, symbol, true, latency)
			
			// Store ticker data
			as.mutex.Lock()
			as.latestTickers[symbol][et] = ticker
			as.mutex.Unlock()
			
			// Record metrics
			if as.metricsCollector != nil {
				as.metricsCollector.RecordSuccess(et, "get_ticker")
				as.metricsCollector.RecordLatency(et, "get_ticker", latency)
			}
			
			// Publish event
			if as.eventPublisher != nil {
				as.eventPublisher.PublishTicker(ctx, ticker)
			}
			
		}(exchangeType, client)
	}
	
	wg.Wait()
}

// detectArbitrage continuously detects arbitrage opportunities
func (as *AggregationService) detectArbitrage(ctx context.Context) {
	defer as.wg.Done()

	ticker := time.NewTicker(as.config.UpdateInterval * 2) // Check less frequently
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			opportunities, err := as.FindArbitrageOpportunities(ctx, as.config.Symbols)
			if err != nil {
				as.logger.Errorf("Failed to detect arbitrage opportunities: %v", err)
				continue
			}

			for _, opportunity := range opportunities {
				if opportunity.ProfitPercent.GreaterThan(decimal.NewFromFloat(as.config.ArbitrageThreshold)) {
					as.logger.Infof("Arbitrage opportunity detected: %s %.2f%% profit",
						opportunity.Symbol, opportunity.ProfitPercent)

					if as.eventPublisher != nil {
						as.eventPublisher.PublishArbitrage(ctx, opportunity)
					}

					if as.metricsCollector != nil {
						as.metricsCollector.RecordArbitrageOpportunity(opportunity)
					}
				}
			}
		}
	}
}

// updateDataQuality updates data quality metrics
func (as *AggregationService) updateDataQuality(exchange ExchangeType, symbol string, success bool, latency time.Duration) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if as.dataQuality[exchange] == nil {
		as.dataQuality[exchange] = make(map[string]*DataQualityMetrics)
	}

	metrics, exists := as.dataQuality[exchange][symbol]
	if !exists {
		metrics = &DataQualityMetrics{
			Exchange:        exchange,
			Symbol:          symbol,
			LastUpdate:      time.Now(),
			UpdateFrequency: 0,
			Latency:         latency,
			ErrorRate:       0,
			Availability:    1.0,
			QualityScore:    1.0,
		}
		as.dataQuality[exchange][symbol] = metrics
	}

	// Update metrics
	metrics.LastUpdate = time.Now()
	metrics.Latency = latency

	if success {
		metrics.Availability = 0.95*metrics.Availability + 0.05*1.0
	} else {
		metrics.Availability = 0.95*metrics.Availability + 0.05*0.0
		metrics.ErrorRate = 0.95*metrics.ErrorRate + 0.05*1.0
	}

	// Calculate quality score
	metrics.QualityScore = as.dataValidator.CalculateDataQuality(metrics)
}

// calculateDataQuality calculates overall data quality for a symbol
func (as *AggregationService) calculateDataQuality(symbol string, tickers map[ExchangeType]*Ticker) float64 {
	if len(tickers) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	count := 0

	for exchange := range tickers {
		if metrics, err := as.GetDataQuality(context.Background(), exchange, symbol); err == nil {
			totalQuality += metrics.QualityScore
			count++
		}
	}

	if count == 0 {
		return 0.5 // Default quality score
	}

	return totalQuality / float64(count)
}

// GetAggregatedOrderBook returns aggregated order book data
func (as *AggregationService) GetAggregatedOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	// Collect order books from all exchanges
	var orderBooks []*OrderBook

	for exchangeType, client := range as.clients {
		orderBook, err := client.GetOrderBook(ctx, symbol, depth)
		if err != nil {
			as.logger.Errorf("Failed to get order book for %s from %s: %v", symbol, exchangeType, err)
			continue
		}
		orderBooks = append(orderBooks, orderBook)
	}

	if len(orderBooks) == 0 {
		return nil, fmt.Errorf("no order book data available for symbol %s", symbol)
	}

	// Aggregate order books
	return as.aggregateOrderBooks(symbol, orderBooks, depth)
}

// aggregateOrderBooks aggregates multiple order books into one
func (as *AggregationService) aggregateOrderBooks(symbol string, orderBooks []*OrderBook, depth int) (*OrderBook, error) {
	var allBids []OrderBookLevel
	var allAsks []OrderBookLevel

	// Collect all bids and asks
	for _, orderBook := range orderBooks {
		allBids = append(allBids, orderBook.Bids...)
		allAsks = append(allAsks, orderBook.Asks...)
	}

	// Sort bids (highest price first)
	sort.Slice(allBids, func(i, j int) bool {
		return allBids[i].Price.GreaterThan(allBids[j].Price)
	})

	// Sort asks (lowest price first)
	sort.Slice(allAsks, func(i, j int) bool {
		return allAsks[i].Price.LessThan(allAsks[j].Price)
	})

	// Limit depth
	if depth > 0 {
		if len(allBids) > depth {
			allBids = allBids[:depth]
		}
		if len(allAsks) > depth {
			allAsks = allAsks[:depth]
		}
	}

	return &OrderBook{
		Exchange:  ExchangeType("aggregated"),
		Symbol:    symbol,
		Bids:      allBids,
		Asks:      allAsks,
		Timestamp: time.Now(),
	}, nil
}

// GetHistoricalData returns aggregated historical data
func (as *AggregationService) GetHistoricalData(ctx context.Context, symbol, interval string, startTime, endTime time.Time) ([]*Kline, error) {
	// For now, use the first available exchange
	// In a full implementation, this would aggregate data from multiple exchanges
	for _, client := range as.clients {
		klines, err := client.GetKlines(ctx, symbol, interval, &startTime, &endTime, 0)
		if err == nil {
			return klines, nil
		}
	}

	return nil, fmt.Errorf("no historical data available for symbol %s", symbol)
}

// GetAggregatedVolume returns aggregated volume for a symbol over a period
func (as *AggregationService) GetAggregatedVolume(ctx context.Context, symbol string, period time.Duration) (decimal.Decimal, error) {
	as.mutex.RLock()
	exchangeTickers, exists := as.latestTickers[symbol]
	as.mutex.RUnlock()

	if !exists {
		return decimal.Zero, fmt.Errorf("no ticker data available for symbol %s", symbol)
	}

	totalVolume := decimal.Zero
	for _, ticker := range exchangeTickers {
		totalVolume = totalVolume.Add(ticker.Volume24h)
	}

	return totalVolume, nil
}
