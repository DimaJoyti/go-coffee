package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/sirupsen/logrus"
)

// BaseStrategy provides common functionality for all strategies
type BaseStrategy struct {
	id          string
	name        string
	strategyType models.StrategyType
	status      models.StrategyStatus
	config      map[string]any
	signalChan  chan *models.Signal
	isRunning   bool
	mu          sync.RWMutex
	stopChan    chan struct{}
	metrics     map[string]any
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(id, name string, strategyType models.StrategyType) *BaseStrategy {
	return &BaseStrategy{
		id:           id,
		name:         name,
		strategyType: strategyType,
		status:       models.StrategyStatusStopped,
		signalChan:   make(chan *models.Signal, 100),
		stopChan:     make(chan struct{}),
		metrics:      make(map[string]any),
	}
}

// GetID returns the strategy ID
func (bs *BaseStrategy) GetID() string {
	return bs.id
}

// GetName returns the strategy name
func (bs *BaseStrategy) GetName() string {
	return bs.name
}

// GetType returns the strategy type
func (bs *BaseStrategy) GetType() models.StrategyType {
	return bs.strategyType
}

// GetStatus returns the strategy status
func (bs *BaseStrategy) GetStatus() models.StrategyStatus {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.status
}

// GetSignals returns the signal channel
func (bs *BaseStrategy) GetSignals() <-chan *models.Signal {
	return bs.signalChan
}

// GetMetrics returns strategy metrics
func (bs *BaseStrategy) GetMetrics() map[string]any {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	
	metrics := make(map[string]any)
	for k, v := range bs.metrics {
		metrics[k] = v
	}
	return metrics
}

// IsHealthy returns the health status
func (bs *BaseStrategy) IsHealthy() bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.isRunning && bs.status == models.StrategyStatusRunning
}

// Initialize initializes the strategy with configuration
func (bs *BaseStrategy) Initialize(ctx context.Context, config map[string]any) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	bs.config = config
	bs.status = models.StrategyStatusStopped
	
	logrus.WithField("strategy_id", bs.id).Info("Strategy initialized")
	return nil
}

// Start starts the strategy
func (bs *BaseStrategy) Start(ctx context.Context) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	if bs.isRunning {
		return fmt.Errorf("strategy is already running")
	}
	
	bs.isRunning = true
	bs.status = models.StrategyStatusRunning
	
	logrus.WithField("strategy_id", bs.id).Info("Strategy started")
	return nil
}

// Stop stops the strategy
func (bs *BaseStrategy) Stop() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	if !bs.isRunning {
		return nil
	}
	
	close(bs.stopChan)
	bs.isRunning = false
	bs.status = models.StrategyStatusStopped
	
	logrus.WithField("strategy_id", bs.id).Info("Strategy stopped")
	return nil
}

// EmitSignal emits a trading signal
func (bs *BaseStrategy) EmitSignal(signal *models.Signal) {
	select {
	case bs.signalChan <- signal:
	default:
		// Channel is full, skip
	}
}

// MarketMakingStrategy implements a market making strategy
type MarketMakingStrategy struct {
	*BaseStrategy
}

// NewMarketMakingStrategy creates a new market making strategy
func NewMarketMakingStrategy() StrategyInterface {
	base := NewBaseStrategy("market_maker_1", "Market Making Strategy", models.StrategyTypeMarketMaking)
	return &MarketMakingStrategy{BaseStrategy: base}
}

// OnTick processes market data ticks
func (mms *MarketMakingStrategy) OnTick(tick *models.MarketDataTick) error {
	// Placeholder implementation for market making logic
	logrus.WithFields(logrus.Fields{
		"strategy": mms.GetID(),
		"symbol":   tick.Symbol,
		"price":    tick.Price,
	}).Debug("Processing tick for market making")
	return nil
}

// OnOrderBook processes order book updates
func (mms *MarketMakingStrategy) OnOrderBook(orderBook *models.OrderBook) error {
	// Placeholder implementation for order book analysis
	logrus.WithFields(logrus.Fields{
		"strategy": mms.GetID(),
		"symbol":   orderBook.Symbol,
	}).Debug("Processing order book for market making")
	return nil
}

// OnOrderUpdate processes order updates
func (mms *MarketMakingStrategy) OnOrderUpdate(order *models.Order) error {
	logrus.WithFields(logrus.Fields{
		"strategy": mms.GetID(),
		"order_id": order.ID,
		"status":   order.Status,
	}).Debug("Processing order update for market making")
	return nil
}

// OnFill processes fill events
func (mms *MarketMakingStrategy) OnFill(fill *models.Fill) error {
	logrus.WithFields(logrus.Fields{
		"strategy": mms.GetID(),
		"fill_id":  fill.ID,
		"quantity": fill.Quantity,
		"price":    fill.Price,
	}).Debug("Processing fill for market making")
	return nil
}

// ArbitrageStrategy implements an arbitrage strategy
type ArbitrageStrategy struct {
	*BaseStrategy
}

// NewArbitrageStrategy creates a new arbitrage strategy
func NewArbitrageStrategy() StrategyInterface {
	base := NewBaseStrategy("arbitrage_1", "Arbitrage Strategy", models.StrategyTypeArbitrage)
	return &ArbitrageStrategy{BaseStrategy: base}
}

// OnTick processes market data ticks
func (as *ArbitrageStrategy) OnTick(tick *models.MarketDataTick) error {
	logrus.WithFields(logrus.Fields{
		"strategy": as.GetID(),
		"symbol":   tick.Symbol,
		"exchange": tick.Exchange,
		"price":    tick.Price,
	}).Debug("Processing tick for arbitrage")
	return nil
}

// OnOrderBook processes order book updates
func (as *ArbitrageStrategy) OnOrderBook(orderBook *models.OrderBook) error {
	logrus.WithFields(logrus.Fields{
		"strategy": as.GetID(),
		"symbol":   orderBook.Symbol,
		"exchange": orderBook.Exchange,
	}).Debug("Processing order book for arbitrage")
	return nil
}

// OnOrderUpdate processes order updates
func (as *ArbitrageStrategy) OnOrderUpdate(order *models.Order) error {
	logrus.WithFields(logrus.Fields{
		"strategy": as.GetID(),
		"order_id": order.ID,
		"status":   order.Status,
	}).Debug("Processing order update for arbitrage")
	return nil
}

// OnFill processes fill events
func (as *ArbitrageStrategy) OnFill(fill *models.Fill) error {
	logrus.WithFields(logrus.Fields{
		"strategy": as.GetID(),
		"fill_id":  fill.ID,
		"quantity": fill.Quantity,
		"price":    fill.Price,
	}).Debug("Processing fill for arbitrage")
	return nil
}

// MomentumStrategy implements a momentum strategy
type MomentumStrategy struct {
	*BaseStrategy
}

// NewMomentumStrategy creates a new momentum strategy
func NewMomentumStrategy() StrategyInterface {
	base := NewBaseStrategy("momentum_1", "Momentum Strategy", models.StrategyTypeMomentum)
	return &MomentumStrategy{BaseStrategy: base}
}

// OnTick processes market data ticks
func (ms *MomentumStrategy) OnTick(tick *models.MarketDataTick) error {
	logrus.WithFields(logrus.Fields{
		"strategy": ms.GetID(),
		"symbol":   tick.Symbol,
		"price":    tick.Price,
	}).Debug("Processing tick for momentum")
	return nil
}

// OnOrderBook processes order book updates
func (ms *MomentumStrategy) OnOrderBook(orderBook *models.OrderBook) error {
	logrus.WithFields(logrus.Fields{
		"strategy": ms.GetID(),
		"symbol":   orderBook.Symbol,
	}).Debug("Processing order book for momentum")
	return nil
}

// OnOrderUpdate processes order updates
func (ms *MomentumStrategy) OnOrderUpdate(order *models.Order) error {
	logrus.WithFields(logrus.Fields{
		"strategy": ms.GetID(),
		"order_id": order.ID,
		"status":   order.Status,
	}).Debug("Processing order update for momentum")
	return nil
}

// OnFill processes fill events
func (ms *MomentumStrategy) OnFill(fill *models.Fill) error {
	logrus.WithFields(logrus.Fields{
		"strategy": ms.GetID(),
		"fill_id":  fill.ID,
		"quantity": fill.Quantity,
		"price":    fill.Price,
	}).Debug("Processing fill for momentum")
	return nil
}
