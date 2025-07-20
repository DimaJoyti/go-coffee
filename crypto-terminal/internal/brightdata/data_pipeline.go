package brightdata

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// DataPipeline manages real-time data collection and processing using Bright Data MCP
type DataPipeline struct {
	scraper           *RealTimeScraper
	logger            *logrus.Logger
	config            *BrightDataConfig
	
	// Data channels for real-time streaming
	tradingViewChan   chan *TradingViewData
	marketSentimentChan chan string
	alertsChan        chan *MarketAlert
	
	// Control channels
	stopChan          chan struct{}
	errorChan         chan error
	
	// State management
	isRunning         bool
	mu                sync.RWMutex
	lastUpdate        time.Time
	dataQuality       float64
}

// MarketAlert represents a market alert generated from real-time data
type MarketAlert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Symbol      string                 `json:"symbol,omitempty"`
	Price       float64                `json:"price,omitempty"`
	Change      float64                `json:"change,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// NewDataPipeline creates a new data pipeline
func NewDataPipeline(config *BrightDataConfig, logger *logrus.Logger) *DataPipeline {
	return &DataPipeline{
		scraper:             NewRealTimeScraper(logger),
		logger:              logger,
		config:              config,
		tradingViewChan:     make(chan *TradingViewData, 100),
		marketSentimentChan: make(chan string, 50),
		alertsChan:          make(chan *MarketAlert, 200),
		stopChan:            make(chan struct{}),
		errorChan:           make(chan error, 10),
		dataQuality:         0.0,
	}
}

// Start begins the real-time data pipeline
func (dp *DataPipeline) Start(ctx context.Context) error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if dp.isRunning {
		return fmt.Errorf("data pipeline is already running")
	}

	dp.logger.Info("Starting Bright Data MCP real-time data pipeline")
	dp.isRunning = true

	// Start data collection goroutines
	go dp.collectTradingViewData(ctx)
	go dp.collectMarketSentiment(ctx)
	go dp.generateMarketAlerts(ctx)
	go dp.monitorDataQuality(ctx)

	dp.logger.Info("Bright Data MCP data pipeline started successfully")
	return nil
}

// Stop stops the data pipeline
func (dp *DataPipeline) Stop() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if !dp.isRunning {
		return fmt.Errorf("data pipeline is not running")
	}

	dp.logger.Info("Stopping Bright Data MCP data pipeline")
	close(dp.stopChan)
	dp.isRunning = false

	dp.logger.Info("Bright Data MCP data pipeline stopped")
	return nil
}

// collectTradingViewData continuously collects TradingView data using Bright Data MCP
func (dp *DataPipeline) collectTradingViewData(ctx context.Context) {
	ticker := time.NewTicker(dp.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dp.stopChan:
			return
		case <-ticker.C:
			dp.logger.Debug("Collecting TradingView data via Bright Data MCP")
			
			data, err := dp.scraper.ScrapeTradingViewMarketData(ctx)
			if err != nil {
				dp.logger.Errorf("Failed to scrape TradingView data: %v", err)
				dp.errorChan <- err
				continue
			}

			// Update data quality metrics
			dp.updateDataQuality(data.DataQuality)
			
			// Send data to channel for real-time processing
			select {
			case dp.tradingViewChan <- data:
				dp.logger.Debug("TradingView data sent to processing channel")
			default:
				dp.logger.Warn("TradingView data channel is full, dropping data")
			}
		}
	}
}

// collectMarketSentiment continuously analyzes market sentiment
func (dp *DataPipeline) collectMarketSentiment(ctx context.Context) {
	ticker := time.NewTicker(dp.config.UpdateInterval * 2) // Less frequent than price data
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dp.stopChan:
			return
		case <-ticker.C:
			dp.logger.Debug("Analyzing market sentiment via Bright Data MCP")
			
			sentiment, err := dp.scraper.GetRealTimeMarketSentiment(ctx)
			if err != nil {
				dp.logger.Errorf("Failed to get market sentiment: %v", err)
				dp.errorChan <- err
				continue
			}

			select {
			case dp.marketSentimentChan <- sentiment:
				dp.logger.Debugf("Market sentiment updated: %s", sentiment)
			default:
				dp.logger.Warn("Market sentiment channel is full, dropping data")
			}
		}
	}
}

// generateMarketAlerts generates alerts based on real-time data analysis
func (dp *DataPipeline) generateMarketAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dp.stopChan:
			return
		case data := <-dp.tradingViewChan:
			alerts := dp.analyzeDataForAlerts(data)
			for _, alert := range alerts {
				select {
				case dp.alertsChan <- alert:
					dp.logger.Infof("Generated market alert: %s", alert.Title)
				default:
					dp.logger.Warn("Alerts channel is full, dropping alert")
				}
			}
		}
	}
}

// analyzeDataForAlerts analyzes TradingView data for alert conditions
func (dp *DataPipeline) analyzeDataForAlerts(data *TradingViewData) []*MarketAlert {
	var alerts []*MarketAlert

	// Check for significant price movements
	for _, coin := range data.Coins {
		if coin.ChangePercent.GreaterThan(decimal.NewFromFloat(10)) {
			alerts = append(alerts, &MarketAlert{
				ID:        fmt.Sprintf("price_surge_%s_%d", coin.Symbol, time.Now().Unix()),
				Type:      "price_movement",
				Severity:  "high",
				Title:     fmt.Sprintf("%s Price Surge", coin.Symbol),
				Message:   fmt.Sprintf("%s has increased by %.2f%% in the last 24 hours", coin.Name, coin.ChangePercent.InexactFloat64()),
				Symbol:    coin.Symbol,
				Price:     coin.Price.InexactFloat64(),
				Change:    coin.ChangePercent.InexactFloat64(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"market_cap": coin.MarketCap.InexactFloat64(),
					"volume_24h": coin.Volume24h.InexactFloat64(),
				},
			})
		}

		if coin.ChangePercent.LessThan(decimal.NewFromFloat(-10)) {
			alerts = append(alerts, &MarketAlert{
				ID:        fmt.Sprintf("price_drop_%s_%d", coin.Symbol, time.Now().Unix()),
				Type:      "price_movement",
				Severity:  "medium",
				Title:     fmt.Sprintf("%s Price Drop", coin.Symbol),
				Message:   fmt.Sprintf("%s has decreased by %.2f%% in the last 24 hours", coin.Name, coin.ChangePercent.Abs().InexactFloat64()),
				Symbol:    coin.Symbol,
				Price:     coin.Price.InexactFloat64(),
				Change:    coin.ChangePercent.InexactFloat64(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"market_cap": coin.MarketCap.InexactFloat64(),
					"volume_24h": coin.Volume24h.InexactFloat64(),
				},
			})
		}
	}

	// Check for trending coins
	for _, trending := range data.TrendingCoins {
		if trending.TrendScore.GreaterThan(decimal.NewFromFloat(90)) {
			alerts = append(alerts, &MarketAlert{
				ID:        fmt.Sprintf("trending_%s_%d", trending.Symbol, time.Now().Unix()),
				Type:      "trending",
				Severity:  "info",
				Title:     fmt.Sprintf("%s Trending", trending.Symbol),
				Message:   fmt.Sprintf("%s is trending with a score of %.1f and %d mentions", trending.Name, trending.TrendScore.InexactFloat64(), trending.Mentions),
				Symbol:    trending.Symbol,
				Price:     trending.Price.InexactFloat64(),
				Change:    trending.Change24h.InexactFloat64(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"trend_score": trending.TrendScore.InexactFloat64(),
					"mentions":    trending.Mentions,
				},
			})
		}
	}

	return alerts
}

// monitorDataQuality monitors the quality of scraped data
func (dp *DataPipeline) monitorDataQuality(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dp.stopChan:
			return
		case <-ticker.C:
			dp.mu.RLock()
			quality := dp.dataQuality
			lastUpdate := dp.lastUpdate
			dp.mu.RUnlock()

			// Check if data is stale
			if time.Since(lastUpdate) > dp.config.UpdateInterval*3 {
				dp.logger.Warn("Data appears to be stale, last update was over 3 intervals ago")
				
				alert := &MarketAlert{
					ID:        fmt.Sprintf("data_quality_%d", time.Now().Unix()),
					Type:      "system",
					Severity:  "warning",
					Title:     "Data Quality Warning",
					Message:   "Market data appears to be stale or outdated",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"last_update":  lastUpdate,
						"data_quality": quality,
					},
				}

				select {
				case dp.alertsChan <- alert:
				default:
				}
			}

			// Check data quality threshold
			if quality < 0.8 {
				dp.logger.Warnf("Data quality is below threshold: %.2f", quality)
			}
		}
	}
}

// updateDataQuality updates the data quality metrics
func (dp *DataPipeline) updateDataQuality(quality float64) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	
	dp.dataQuality = quality
	dp.lastUpdate = time.Now()
}

// GetTradingViewData returns the latest TradingView data
func (dp *DataPipeline) GetTradingViewData() *TradingViewData {
	select {
	case data := <-dp.tradingViewChan:
		return data
	default:
		return nil
	}
}

// GetMarketAlerts returns pending market alerts
func (dp *DataPipeline) GetMarketAlerts() []*MarketAlert {
	var alerts []*MarketAlert
	
	// Drain the alerts channel
	for {
		select {
		case alert := <-dp.alertsChan:
			alerts = append(alerts, alert)
		default:
			return alerts
		}
	}
}

// GetDataQuality returns current data quality metrics
func (dp *DataPipeline) GetDataQuality() (float64, time.Time) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	return dp.dataQuality, dp.lastUpdate
}

// IsRunning returns whether the pipeline is currently running
func (dp *DataPipeline) IsRunning() bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	return dp.isRunning
}
