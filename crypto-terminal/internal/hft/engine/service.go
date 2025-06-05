package engine

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// StrategyInterface defines the interface that all trading strategies must implement
type StrategyInterface interface {
	Initialize(ctx context.Context, config map[string]any) error
	Start(ctx context.Context) error
	Stop() error
	OnTick(tick *models.MarketDataTick) error
	OnOrderBook(orderBook *models.OrderBook) error
	OnOrderUpdate(order *models.Order) error
	OnFill(fill *models.Fill) error
	GetSignals() <-chan *models.Signal
	GetMetrics() map[string]any
	IsHealthy() bool
	GetID() string
	GetName() string
	GetType() models.StrategyType
	GetStatus() models.StrategyStatus
}

// Service manages the strategy engine and running strategies
type Service struct {
	config     *config.Config
	db         *sql.DB
	redis      *redis.Client
	
	// Strategy management
	strategies     map[string]StrategyInterface
	strategyConfigs map[string]*models.Strategy
	
	// Data channels
	tickChan      chan *models.MarketDataTick
	orderBookChan chan *models.OrderBook
	orderChan     chan *models.Order
	fillChan      chan *models.Fill
	signalChan    chan *models.Signal
	
	// State management
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	
	// Performance metrics
	totalSignals    uint64
	executedSignals uint64
	strategyCount   int
}

// NewService creates a new strategy engine service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	return &Service{
		config:          cfg,
		db:              db,
		redis:           redis,
		strategies:      make(map[string]StrategyInterface),
		strategyConfigs: make(map[string]*models.Strategy),
		tickChan:        make(chan *models.MarketDataTick, 10000),
		orderBookChan:   make(chan *models.OrderBook, 1000),
		orderChan:       make(chan *models.Order, 1000),
		fillChan:        make(chan *models.Fill, 1000),
		signalChan:      make(chan *models.Signal, 1000),
		stopChan:        make(chan struct{}),
	}, nil
}

// Start starts the strategy engine
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("strategy engine is already running")
	}

	logrus.Info("Starting HFT Strategy Engine")

	// Load strategy configurations from database
	if err := s.loadStrategyConfigs(ctx); err != nil {
		return fmt.Errorf("failed to load strategy configs: %w", err)
	}

	// Initialize and start strategies
	if err := s.initializeStrategies(ctx); err != nil {
		return fmt.Errorf("failed to initialize strategies: %w", err)
	}

	// Start data processing goroutines
	s.wg.Add(5)
	go s.processTickData(ctx)
	go s.processOrderBookData(ctx)
	go s.processOrderUpdates(ctx)
	go s.processFills(ctx)
	go s.processSignals(ctx)

	s.isRunning = true
	logrus.Info("HFT Strategy Engine started successfully")

	return nil
}

// Stop stops the strategy engine
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping HFT Strategy Engine")

	// Stop all strategies
	for _, strategy := range s.strategies {
		if err := strategy.Stop(); err != nil {
			logrus.WithError(err).Errorf("Failed to stop strategy %s", strategy.GetID())
		}
	}

	// Signal stop
	close(s.stopChan)

	// Wait for goroutines to finish
	s.wg.Wait()

	s.isRunning = false
	logrus.Info("HFT Strategy Engine stopped")

	return nil
}

// RegisterStrategy registers a new strategy
func (s *Service) RegisterStrategy(strategy StrategyInterface) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	strategyID := strategy.GetID()
	if _, exists := s.strategies[strategyID]; exists {
		return fmt.Errorf("strategy %s already registered", strategyID)
	}

	s.strategies[strategyID] = strategy
	s.strategyCount++

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategyID,
		"strategy_name": strategy.GetName(),
		"strategy_type": strategy.GetType(),
	}).Info("Strategy registered successfully")

	return nil
}

// UnregisterStrategy unregisters a strategy
func (s *Service) UnregisterStrategy(strategyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	strategy, exists := s.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy %s not found", strategyID)
	}

	// Stop strategy if running
	if err := strategy.Stop(); err != nil {
		logrus.WithError(err).Errorf("Failed to stop strategy %s", strategyID)
	}

	delete(s.strategies, strategyID)
	delete(s.strategyConfigs, strategyID)
	s.strategyCount--

	logrus.WithField("strategy_id", strategyID).Info("Strategy unregistered successfully")

	return nil
}

// StartStrategy starts a specific strategy
func (s *Service) StartStrategy(ctx context.Context, strategyID string) error {
	s.mu.RLock()
	strategy, exists := s.strategies[strategyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("strategy %s not found", strategyID)
	}

	if err := strategy.Start(ctx); err != nil {
		return fmt.Errorf("failed to start strategy %s: %w", strategyID, err)
	}

	// Update strategy status in database
	s.updateStrategyStatus(ctx, strategyID, models.StrategyStatusRunning)

	logrus.WithField("strategy_id", strategyID).Info("Strategy started successfully")

	return nil
}

// StopStrategy stops a specific strategy
func (s *Service) StopStrategy(ctx context.Context, strategyID string) error {
	s.mu.RLock()
	strategy, exists := s.strategies[strategyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("strategy %s not found", strategyID)
	}

	if err := strategy.Stop(); err != nil {
		return fmt.Errorf("failed to stop strategy %s: %w", strategyID, err)
	}

	// Update strategy status in database
	s.updateStrategyStatus(ctx, strategyID, models.StrategyStatusStopped)

	logrus.WithField("strategy_id", strategyID).Info("Strategy stopped successfully")

	return nil
}

// GetStrategies returns all registered strategies
func (s *Service) GetStrategies() map[string]StrategyInterface {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make(map[string]StrategyInterface)
	for id, strategy := range s.strategies {
		strategies[id] = strategy
	}

	return strategies
}

// GetStrategy returns a specific strategy
func (s *Service) GetStrategy(strategyID string) (StrategyInterface, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategy, exists := s.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", strategyID)
	}

	return strategy, nil
}

// SendTick sends market data tick to all strategies
func (s *Service) SendTick(tick *models.MarketDataTick) {
	select {
	case s.tickChan <- tick:
	default:
		// Channel is full, skip
	}
}

// SendOrderBook sends order book data to all strategies
func (s *Service) SendOrderBook(orderBook *models.OrderBook) {
	select {
	case s.orderBookChan <- orderBook:
	default:
		// Channel is full, skip
	}
}

// SendOrderUpdate sends order update to all strategies
func (s *Service) SendOrderUpdate(order *models.Order) {
	select {
	case s.orderChan <- order:
	default:
		// Channel is full, skip
	}
}

// SendFill sends fill data to all strategies
func (s *Service) SendFill(fill *models.Fill) {
	select {
	case s.fillChan <- fill:
	default:
		// Channel is full, skip
	}
}

// GetSignalChannel returns the signal channel
func (s *Service) GetSignalChannel() <-chan *models.Signal {
	return s.signalChan
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isRunning {
		return false
	}

	// Check if at least one strategy is healthy
	healthyStrategies := 0
	for _, strategy := range s.strategies {
		if strategy.IsHealthy() {
			healthyStrategies++
		}
	}

	return healthyStrategies > 0
}

// GetMetrics returns strategy engine metrics
func (s *Service) GetMetrics() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategyMetrics := make(map[string]any)
	for id, strategy := range s.strategies {
		strategyMetrics[id] = strategy.GetMetrics()
	}

	return map[string]any{
		"total_signals":     s.totalSignals,
		"executed_signals":  s.executedSignals,
		"strategy_count":    s.strategyCount,
		"running_strategies": len(s.strategies),
		"strategy_metrics":  strategyMetrics,
	}
}

// loadStrategyConfigs loads strategy configurations from database
func (s *Service) loadStrategyConfigs(ctx context.Context) error {
	// Placeholder implementation - would query database for strategy configs
	logrus.Info("Loading strategy configurations from database")
	return nil
}

// initializeStrategies initializes all configured strategies
func (s *Service) initializeStrategies(ctx context.Context) error {
	// Register built-in strategies
	if err := s.registerBuiltInStrategies(); err != nil {
		return fmt.Errorf("failed to register built-in strategies: %w", err)
	}

	// Initialize strategies with their configurations
	for strategyID, config := range s.strategyConfigs {
		strategy, exists := s.strategies[strategyID]
		if !exists {
			logrus.WithField("strategy_id", strategyID).Warn("Strategy not found for configuration")
			continue
		}

		if err := strategy.Initialize(ctx, config.Parameters); err != nil {
			logrus.WithError(err).Errorf("Failed to initialize strategy %s", strategyID)
			continue
		}

		logrus.WithField("strategy_id", strategyID).Info("Strategy initialized successfully")
	}

	return nil
}

// registerBuiltInStrategies registers built-in trading strategies
func (s *Service) registerBuiltInStrategies() error {
	// Register Market Making strategy
	marketMaker := NewMarketMakingStrategy()
	if err := s.RegisterStrategy(marketMaker); err != nil {
		return fmt.Errorf("failed to register market making strategy: %w", err)
	}

	// Register Arbitrage strategy
	arbitrage := NewArbitrageStrategy()
	if err := s.RegisterStrategy(arbitrage); err != nil {
		return fmt.Errorf("failed to register arbitrage strategy: %w", err)
	}

	// Register Momentum strategy
	momentum := NewMomentumStrategy()
	if err := s.RegisterStrategy(momentum); err != nil {
		return fmt.Errorf("failed to register momentum strategy: %w", err)
	}

	return nil
}

// processTickData processes incoming tick data and distributes to strategies
func (s *Service) processTickData(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case tick := <-s.tickChan:
			s.distributeTickToStrategies(tick)
		}
	}
}

// processOrderBookData processes incoming order book data and distributes to strategies
func (s *Service) processOrderBookData(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case orderBook := <-s.orderBookChan:
			s.distributeOrderBookToStrategies(orderBook)
		}
	}
}

// processOrderUpdates processes order updates and distributes to strategies
func (s *Service) processOrderUpdates(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case order := <-s.orderChan:
			s.distributeOrderUpdateToStrategies(order)
		}
	}
}

// processFills processes fill data and distributes to strategies
func (s *Service) processFills(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case fill := <-s.fillChan:
			s.distributeFillToStrategies(fill)
		}
	}
}

// processSignals processes signals from strategies
func (s *Service) processSignals(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			// Collect signals from all strategies
			s.collectSignalsFromStrategies()
		}
	}
}

// distributeTickToStrategies distributes tick data to all strategies
func (s *Service) distributeTickToStrategies(tick *models.MarketDataTick) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, strategy := range s.strategies {
		if err := strategy.OnTick(tick); err != nil {
			logrus.WithError(err).Errorf("Strategy %s failed to process tick", strategy.GetID())
		}
	}
}

// distributeOrderBookToStrategies distributes order book data to all strategies
func (s *Service) distributeOrderBookToStrategies(orderBook *models.OrderBook) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, strategy := range s.strategies {
		if err := strategy.OnOrderBook(orderBook); err != nil {
			logrus.WithError(err).Errorf("Strategy %s failed to process order book", strategy.GetID())
		}
	}
}

// distributeOrderUpdateToStrategies distributes order updates to relevant strategies
func (s *Service) distributeOrderUpdateToStrategies(order *models.Order) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, strategy := range s.strategies {
		if strategy.GetID() == order.StrategyID {
			if err := strategy.OnOrderUpdate(order); err != nil {
				logrus.WithError(err).Errorf("Strategy %s failed to process order update", strategy.GetID())
			}
		}
	}
}

// distributeFillToStrategies distributes fill data to relevant strategies
func (s *Service) distributeFillToStrategies(fill *models.Fill) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find the strategy that owns this order
	for _, strategy := range s.strategies {
		// In a real implementation, we'd look up the order to find the strategy ID
		if err := strategy.OnFill(fill); err != nil {
			logrus.WithError(err).Errorf("Strategy %s failed to process fill", strategy.GetID())
		}
	}
}

// collectSignalsFromStrategies collects signals from all strategies
func (s *Service) collectSignalsFromStrategies() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, strategy := range s.strategies {
		signalChan := strategy.GetSignals()
		select {
		case signal := <-signalChan:
			if signal != nil {
				s.totalSignals++
				select {
				case s.signalChan <- signal:
				default:
					// Signal channel is full, skip
				}
			}
		default:
			// No signals available
		}
	}
}

// updateStrategyStatus updates strategy status in database
func (s *Service) updateStrategyStatus(ctx context.Context, strategyID string, status models.StrategyStatus) {
	// Placeholder implementation - would update database
	logrus.WithFields(logrus.Fields{
		"strategy_id": strategyID,
		"status":      status,
	}).Debug("Updating strategy status")
}
