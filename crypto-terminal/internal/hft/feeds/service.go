package feeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// FeedProvider interface for market data providers
type FeedProvider interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(symbols []string) error
	Unsubscribe(symbols []string) error
	GetTickChannel() <-chan *models.MarketDataTick
	GetOrderBookChannel() <-chan *models.OrderBook
	IsConnected() bool
	GetLatency() time.Duration
}

// Service manages ultra-low latency market data feeds
type Service struct {
	config    *config.Config
	db        *sql.DB
	redis     *redis.Client
	providers map[string]FeedProvider
	
	// Channels for distributing data
	tickChan      chan *models.MarketDataTick
	orderBookChan chan *models.OrderBook
	
	// Subscribers
	tickSubscribers      map[string][]chan *models.MarketDataTick
	orderBookSubscribers map[string][]chan *models.OrderBook
	
	// State management
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	
	// Performance metrics
	tickCount     uint64
	lastTickTime  time.Time
	avgLatency    time.Duration
	maxLatency    time.Duration
	minLatency    time.Duration
}

// NewService creates a new market data feed service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	s := &Service{
		config:               cfg,
		db:                   db,
		redis:                redis,
		providers:            make(map[string]FeedProvider),
		tickChan:             make(chan *models.MarketDataTick, 10000),
		orderBookChan:        make(chan *models.OrderBook, 1000),
		tickSubscribers:      make(map[string][]chan *models.MarketDataTick),
		orderBookSubscribers: make(map[string][]chan *models.OrderBook),
		stopChan:             make(chan struct{}),
		minLatency:           time.Hour, // Initialize to high value
	}

	// Initialize providers
	if err := s.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return s, nil
}

// Start starts the market data feed service
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("service is already running")
	}

	logrus.Info("Starting HFT market data feed service")

	// Connect to all providers
	for name, provider := range s.providers {
		if err := provider.Connect(ctx); err != nil {
			logrus.WithError(err).Errorf("Failed to connect to provider %s", name)
			continue
		}
		logrus.Infof("Connected to market data provider: %s", name)
	}

	// Start data processing goroutines
	s.wg.Add(3)
	go s.processTickData(ctx)
	go s.processOrderBookData(ctx)
	go s.performanceMonitor(ctx)

	s.isRunning = true
	logrus.Info("HFT market data feed service started successfully")

	return nil
}

// Stop stops the market data feed service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping HFT market data feed service")

	// Signal stop
	close(s.stopChan)

	// Disconnect from providers
	for name, provider := range s.providers {
		if err := provider.Disconnect(); err != nil {
			logrus.WithError(err).Errorf("Failed to disconnect from provider %s", name)
		}
	}

	// Wait for goroutines to finish
	s.wg.Wait()

	s.isRunning = false
	logrus.Info("HFT market data feed service stopped")

	return nil
}

// SubscribeToTicks subscribes to tick data for specific symbols
func (s *Service) SubscribeToTicks(symbols []string) <-chan *models.MarketDataTick {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *models.MarketDataTick, 1000)
	
	for _, symbol := range symbols {
		s.tickSubscribers[symbol] = append(s.tickSubscribers[symbol], ch)
		
		// Subscribe to provider feeds
		for _, provider := range s.providers {
			if provider.IsConnected() {
				provider.Subscribe([]string{symbol})
			}
		}
	}

	return ch
}

// SubscribeToOrderBook subscribes to order book data for specific symbols
func (s *Service) SubscribeToOrderBook(symbols []string) <-chan *models.OrderBook {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *models.OrderBook, 100)
	
	for _, symbol := range symbols {
		s.orderBookSubscribers[symbol] = append(s.orderBookSubscribers[symbol], ch)
		
		// Subscribe to provider feeds
		for _, provider := range s.providers {
			if provider.IsConnected() {
				provider.Subscribe([]string{symbol})
			}
		}
	}

	return ch
}

// GetLatencyStats returns current latency statistics
func (s *Service) GetLatencyStats() map[string]time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]time.Duration{
		"avg": s.avgLatency,
		"min": s.minLatency,
		"max": s.maxLatency,
	}
}

// GetTickCount returns the total number of ticks processed
func (s *Service) GetTickCount() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tickCount
}

// GetMetrics returns service metrics
func (s *Service) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"tick_count":   s.tickCount,
		"avg_latency":  s.avgLatency,
		"min_latency":  s.minLatency,
		"max_latency":  s.maxLatency,
		"providers":    len(s.providers),
		"subscribers":  len(s.tickSubscribers) + len(s.orderBookSubscribers),
	}
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isRunning {
		return false
	}

	// Check if we've received data recently
	if time.Since(s.lastTickTime) > 30*time.Second {
		return false
	}

	// Check provider connections
	connectedProviders := 0
	for _, provider := range s.providers {
		if provider.IsConnected() {
			connectedProviders++
		}
	}

	return connectedProviders > 0
}

// processTickData processes incoming tick data from providers
func (s *Service) processTickData(ctx context.Context) {
	defer s.wg.Done()

	// Collect tick channels from all providers
	var tickChannels []<-chan *models.MarketDataTick
	for _, provider := range s.providers {
		tickChannels = append(tickChannels, provider.GetTickChannel())
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			// Process ticks from all providers
			for _, tickChan := range tickChannels {
				select {
				case tick := <-tickChan:
					if tick != nil {
						s.handleTick(tick)
					}
				default:
					// Non-blocking
				}
			}
		}
	}
}

// processOrderBookData processes incoming order book data from providers
func (s *Service) processOrderBookData(ctx context.Context) {
	defer s.wg.Done()

	// Collect order book channels from all providers
	var orderBookChannels []<-chan *models.OrderBook
	for _, provider := range s.providers {
		orderBookChannels = append(orderBookChannels, provider.GetOrderBookChannel())
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			// Process order books from all providers
			for _, obChan := range orderBookChannels {
				select {
				case orderBook := <-obChan:
					if orderBook != nil {
						s.handleOrderBook(orderBook)
					}
				default:
					// Non-blocking
				}
			}
		}
	}
}

// handleTick processes a single tick and distributes it to subscribers
func (s *Service) handleTick(tick *models.MarketDataTick) {
	// Update processing time and latency
	tick.ProcessTime = time.Now()
	tick.Latency = tick.ProcessTime.Sub(tick.ReceiveTime)

	// Update performance metrics
	s.updateLatencyMetrics(tick.Latency)
	s.tickCount++
	s.lastTickTime = tick.ProcessTime

	// Store in Redis for fast access
	s.storeTick(tick)

	// Distribute to subscribers
	s.mu.RLock()
	subscribers := s.tickSubscribers[tick.Symbol]
	s.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- tick:
		default:
			// Channel is full, skip
		}
	}
}

// handleOrderBook processes order book data and distributes it to subscribers
func (s *Service) handleOrderBook(orderBook *models.OrderBook) {
	// Store in Redis
	s.storeOrderBook(orderBook)

	// Distribute to subscribers
	s.mu.RLock()
	subscribers := s.orderBookSubscribers[orderBook.Symbol]
	s.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- orderBook:
		default:
			// Channel is full, skip
		}
	}
}

// updateLatencyMetrics updates latency performance metrics
func (s *Service) updateLatencyMetrics(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if latency < s.minLatency {
		s.minLatency = latency
	}
	if latency > s.maxLatency {
		s.maxLatency = latency
	}

	// Simple moving average for now
	if s.avgLatency == 0 {
		s.avgLatency = latency
	} else {
		s.avgLatency = (s.avgLatency + latency) / 2
	}
}

// storeTick stores tick data in Redis for fast access
func (s *Service) storeTick(tick *models.MarketDataTick) {
	ctx := context.Background()

	// Store latest tick
	key := fmt.Sprintf("hft:tick:latest:%s:%s", tick.Exchange, tick.Symbol)
	data, err := json.Marshal(tick)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal tick data")
		return
	}

	s.redis.Set(ctx, key, data, 5*time.Minute)

	// Store in time series for historical analysis
	tsKey := fmt.Sprintf("hft:tick:ts:%s:%s", tick.Exchange, tick.Symbol)
	s.redis.ZAdd(ctx, tsKey, redis.Z{
		Score:  float64(tick.Timestamp.UnixNano()),
		Member: data,
	})

	// Keep only last 1000 ticks
	s.redis.ZRemRangeByRank(ctx, tsKey, 0, -1001)
}

// storeOrderBook stores order book data in Redis
func (s *Service) storeOrderBook(orderBook *models.OrderBook) {
	ctx := context.Background()

	key := fmt.Sprintf("hft:orderbook:%s:%s", orderBook.Exchange, orderBook.Symbol)
	data, err := json.Marshal(orderBook)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal order book data")
		return
	}

	s.redis.Set(ctx, key, data, 1*time.Minute)
}

// performanceMonitor monitors and logs performance metrics
func (s *Service) performanceMonitor(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.logPerformanceMetrics()
		}
	}
}

// logPerformanceMetrics logs current performance metrics
func (s *Service) logPerformanceMetrics() {
	s.mu.RLock()
	tickCount := s.tickCount
	avgLatency := s.avgLatency
	minLatency := s.minLatency
	maxLatency := s.maxLatency
	s.mu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"tick_count":   tickCount,
		"avg_latency":  avgLatency,
		"min_latency":  minLatency,
		"max_latency":  maxLatency,
		"providers":    len(s.providers),
	}).Info("HFT feed performance metrics")
}

// initializeProviders initializes market data feed providers
func (s *Service) initializeProviders() error {
	// Initialize Binance provider
	binanceProvider, err := NewBinanceProvider(s.config)
	if err != nil {
		return fmt.Errorf("failed to create Binance provider: %w", err)
	}
	s.providers["binance"] = binanceProvider

	// Initialize Coinbase provider
	coinbaseProvider, err := NewCoinbaseProvider(s.config)
	if err != nil {
		return fmt.Errorf("failed to create Coinbase provider: %w", err)
	}
	s.providers["coinbase"] = coinbaseProvider

	// Initialize Kraken provider
	krakenProvider, err := NewKrakenProvider(s.config)
	if err != nil {
		return fmt.Errorf("failed to create Kraken provider: %w", err)
	}
	s.providers["kraken"] = krakenProvider

	if len(s.providers) == 0 {
		return fmt.Errorf("no market data providers configured")
	}

	logrus.Infof("Initialized %d market data providers", len(s.providers))
	return nil
}
