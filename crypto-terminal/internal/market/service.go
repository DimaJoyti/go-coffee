package market

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Service handles market data operations
type Service struct {
	config      *config.Config
	db          *sql.DB
	redis       *redis.Client
	httpClient  *http.Client
	isHealthy   bool
	mu          sync.RWMutex
	stopChan    chan struct{}
	providers   map[string]Provider
}

// Provider interface for market data providers
type Provider interface {
	GetPrice(ctx context.Context, symbol string) (*models.Price, error)
	GetPrices(ctx context.Context, symbols []string) ([]*models.Price, error)
	GetPriceHistory(ctx context.Context, symbol, timeframe string, limit int) ([]*models.OHLCV, error)
	GetMarketData(ctx context.Context, symbol string) (*models.MarketData, error)
	IsHealthy() bool
}

// NewService creates a new market service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	service := &Service{
		config:     cfg,
		db:         db,
		redis:      redis,
		httpClient: httpClient,
		isHealthy:  true,
		stopChan:   make(chan struct{}),
		providers:  make(map[string]Provider),
	}

	// Initialize providers
	if err := service.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return service, nil
}

// Start starts the market service
func (s *Service) Start(ctx context.Context) error {
	logrus.Info("Starting market data service")

	// Start price update goroutine
	go s.startPriceUpdates(ctx)

	// Start health check goroutine
	go s.startHealthCheck(ctx)

	logrus.Info("Market data service started")
	return nil
}

// Stop stops the market service
func (s *Service) Stop() error {
	logrus.Info("Stopping market data service")
	close(s.stopChan)
	return nil
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isHealthy
}

// GetAllPrices returns prices for all tracked cryptocurrencies
func (s *Service) GetAllPrices(ctx context.Context) ([]*models.Price, error) {
	// Try to get from cache first
	cacheKey := "market:prices:all"
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var prices []*models.Price
		if err := json.Unmarshal([]byte(cached), &prices); err == nil {
			return prices, nil
		}
	}

	// Get from primary provider (CoinGecko)
	symbols := []string{"bitcoin", "ethereum", "binancecoin", "cardano", "solana", "polkadot", "dogecoin", "avalanche-2", "polygon", "chainlink"}
	
	var prices []*models.Price
	for _, provider := range s.providers {
		if providerPrices, err := provider.GetPrices(ctx, symbols); err == nil {
			prices = append(prices, providerPrices...)
			break
		}
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("failed to get prices from any provider")
	}

	// Cache the result
	if data, err := json.Marshal(prices); err == nil {
		s.redis.Set(ctx, cacheKey, data, s.config.MarketData.Cache.PriceTTL)
	}

	return prices, nil
}

// GetPrice returns the current price for a specific symbol
func (s *Service) GetPrice(ctx context.Context, symbol string) (*models.Price, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("market:price:%s", symbol)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var price models.Price
		if err := json.Unmarshal([]byte(cached), &price); err == nil {
			return &price, nil
		}
	}

	// Get from providers
	for _, provider := range s.providers {
		if price, err := provider.GetPrice(ctx, symbol); err == nil {
			// Cache the result
			if data, err := json.Marshal(price); err == nil {
				s.redis.Set(ctx, cacheKey, data, s.config.MarketData.Cache.PriceTTL)
			}
			return price, nil
		}
	}

	return nil, fmt.Errorf("failed to get price for %s from any provider", symbol)
}

// GetPriceHistory returns historical price data
func (s *Service) GetPriceHistory(ctx context.Context, symbol, timeframe, limitStr string) ([]*models.OHLCV, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	// Try cache first
	cacheKey := fmt.Sprintf("market:history:%s:%s:%d", symbol, timeframe, limit)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var history []*models.OHLCV
		if err := json.Unmarshal([]byte(cached), &history); err == nil {
			return history, nil
		}
	}

	// Get from providers
	for _, provider := range s.providers {
		if history, err := provider.GetPriceHistory(ctx, symbol, timeframe, limit); err == nil {
			// Cache the result
			if data, err := json.Marshal(history); err == nil {
				s.redis.Set(ctx, cacheKey, data, s.config.MarketData.Cache.MarketDataTTL)
			}
			return history, nil
		}
	}

	return nil, fmt.Errorf("failed to get price history for %s from any provider", symbol)
}

// GetTechnicalIndicators returns technical indicators for a symbol
func (s *Service) GetTechnicalIndicators(ctx context.Context, symbol, timeframe string) ([]*models.TechnicalIndicator, error) {
	// For now, return mock data
	indicators := []*models.TechnicalIndicator{
		{
			Symbol:     symbol,
			Timeframe:  timeframe,
			Indicator:  "RSI",
			Value:      decimal.NewFromFloat(65.5),
			Timestamp:  time.Now(),
			Signal:     "NEUTRAL",
			Confidence: decimal.NewFromFloat(0.7),
		},
		{
			Symbol:     symbol,
			Timeframe:  timeframe,
			Indicator:  "MACD",
			Value:      decimal.NewFromFloat(0.025),
			Timestamp:  time.Now(),
			Signal:     "BUY",
			Confidence: decimal.NewFromFloat(0.8),
		},
	}

	return indicators, nil
}

// GetMarketOverview returns overall market statistics
func (s *Service) GetMarketOverview(ctx context.Context) (*models.MarketOverview, error) {
	// Mock data for now
	overview := &models.MarketOverview{
		TotalMarketCap:         decimal.NewFromFloat(2500000000000), // $2.5T
		TotalVolume24h:         decimal.NewFromFloat(85000000000),   // $85B
		MarketCapChange24h:     decimal.NewFromFloat(2.5),           // +2.5%
		ActiveCryptocurrencies: 10000,
		Markets:                50000,
		BTCDominance:           decimal.NewFromFloat(42.5),
		ETHDominance:           decimal.NewFromFloat(18.2),
		FearGreedIndex:         65, // Greed
		LastUpdated:            time.Now(),
	}

	return overview, nil
}

// GetTopGainers returns top gaining cryptocurrencies
func (s *Service) GetTopGainers(ctx context.Context) ([]*models.TopGainer, error) {
	// Mock data for now
	gainers := []*models.TopGainer{
		{
			Symbol:    "SOL",
			Name:      "Solana",
			Price:     decimal.NewFromFloat(125.50),
			Change24h: decimal.NewFromFloat(15.2),
			Volume24h: decimal.NewFromFloat(2500000000),
		},
		{
			Symbol:    "AVAX",
			Name:      "Avalanche",
			Price:     decimal.NewFromFloat(42.80),
			Change24h: decimal.NewFromFloat(12.8),
			Volume24h: decimal.NewFromFloat(850000000),
		},
	}

	return gainers, nil
}

// GetTopLosers returns top losing cryptocurrencies
func (s *Service) GetTopLosers(ctx context.Context) ([]*models.TopLoser, error) {
	// Mock data for now
	losers := []*models.TopLoser{
		{
			Symbol:    "DOGE",
			Name:      "Dogecoin",
			Price:     decimal.NewFromFloat(0.085),
			Change24h: decimal.NewFromFloat(-8.5),
			Volume24h: decimal.NewFromFloat(450000000),
		},
		{
			Symbol:    "ADA",
			Name:      "Cardano",
			Price:     decimal.NewFromFloat(0.52),
			Change24h: decimal.NewFromFloat(-6.2),
			Volume24h: decimal.NewFromFloat(320000000),
		},
	}

	return losers, nil
}

// GetTrendingCoins returns trending cryptocurrencies
func (s *Service) GetTrendingCoins(ctx context.Context) ([]*models.MarketData, error) {
	// Mock data for now
	trending := []*models.MarketData{
		{
			Symbol:        "BTC",
			Name:          "Bitcoin",
			CurrentPrice:  decimal.NewFromFloat(65000),
			MarketCap:     decimal.NewFromFloat(1280000000000),
			MarketCapRank: 1,
			Volume24h:     decimal.NewFromFloat(25000000000),
			Change24h:     decimal.NewFromFloat(3.2),
			LastUpdated:   time.Now(),
		},
		{
			Symbol:        "ETH",
			Name:          "Ethereum",
			CurrentPrice:  decimal.NewFromFloat(3200),
			MarketCap:     decimal.NewFromFloat(385000000000),
			MarketCapRank: 2,
			Volume24h:     decimal.NewFromFloat(15000000000),
			Change24h:     decimal.NewFromFloat(2.8),
			LastUpdated:   time.Now(),
		},
	}

	return trending, nil
}

// initializeProviders initializes market data providers
func (s *Service) initializeProviders() error {
	// Initialize CoinGecko provider
	if s.config.MarketData.Providers.CoinGecko.APIKey != "" {
		provider := NewCoinGeckoProvider(s.config.MarketData.Providers.CoinGecko, s.httpClient)
		s.providers["coingecko"] = provider
	}

	// Initialize Binance provider
	provider := NewBinanceProvider(s.config.MarketData.Providers.Binance, s.httpClient)
	s.providers["binance"] = provider

	if len(s.providers) == 0 {
		return fmt.Errorf("no market data providers configured")
	}

	return nil
}

// startPriceUpdates starts the price update goroutine
func (s *Service) startPriceUpdates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			// Update prices in background
			go s.updatePrices(ctx)
		}
	}
}

// updatePrices updates prices from providers
func (s *Service) updatePrices(ctx context.Context) {
	// Implementation for background price updates
	// This would fetch latest prices and update cache/database
}

// startHealthCheck starts the health check goroutine
func (s *Service) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkHealth(ctx)
		}
	}
}

// checkHealth checks the health of providers
func (s *Service) checkHealth(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	healthy := false
	for _, provider := range s.providers {
		if provider.IsHealthy() {
			healthy = true
			break
		}
	}

	s.isHealthy = healthy
}
