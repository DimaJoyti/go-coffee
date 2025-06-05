package oms

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// OrderManager interface for order management operations
type OrderManager interface {
	PlaceOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	ModifyOrder(ctx context.Context, orderID string, newPrice, newQuantity decimal.Decimal) error
	GetOrder(ctx context.Context, orderID string) (*models.Order, error)
	GetActiveOrders(ctx context.Context, strategyID string) ([]*models.Order, error)
	GetOrderHistory(ctx context.Context, strategyID string, limit int) ([]*models.Order, error)
}

// PositionManager interface for position management operations
type PositionManager interface {
	GetPosition(ctx context.Context, strategyID, symbol, exchange string) (*models.Position, error)
	GetAllPositions(ctx context.Context, strategyID string) ([]*models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	ClosePosition(ctx context.Context, strategyID, symbol, exchange string) error
}

// Service manages order lifecycle and position tracking
type Service struct {
	config          *config.Config
	db              *sql.DB
	redis           *redis.Client
	orderManager    OrderManager
	positionManager PositionManager
	
	// Order tracking
	activeOrders map[string]*models.Order
	orderHistory []string
	
	// Position tracking
	positions map[string]*models.Position
	
	// Event channels
	orderUpdateChan chan *models.Order
	fillChan        chan *models.Fill
	positionChan    chan *models.Position
	
	// State management
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	
	// Performance metrics
	totalOrders     uint64
	filledOrders    uint64
	canceledOrders  uint64
	rejectedOrders  uint64
	avgFillTime     time.Duration
}

// NewService creates a new Order Management Service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	s := &Service{
		config:          cfg,
		db:              db,
		redis:           redis,
		activeOrders:    make(map[string]*models.Order),
		positions:       make(map[string]*models.Position),
		orderUpdateChan: make(chan *models.Order, 1000),
		fillChan:        make(chan *models.Fill, 1000),
		positionChan:    make(chan *models.Position, 1000),
		stopChan:        make(chan struct{}),
	}

	// Initialize managers
	s.orderManager = NewOrderManager(cfg, db, redis)
	s.positionManager = NewPositionManager(cfg, db, redis)

	return s, nil
}

// Start starts the OMS service
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("OMS service is already running")
	}

	logrus.Info("Starting HFT Order Management Service")

	// Load active orders from database
	if err := s.loadActiveOrders(ctx); err != nil {
		return fmt.Errorf("failed to load active orders: %w", err)
	}

	// Load positions from database
	if err := s.loadPositions(ctx); err != nil {
		return fmt.Errorf("failed to load positions: %w", err)
	}

	// Start processing goroutines
	s.wg.Add(3)
	go s.processOrderUpdates(ctx)
	go s.processFills(ctx)
	go s.processPositionUpdates(ctx)

	s.isRunning = true
	logrus.Info("HFT Order Management Service started successfully")

	return nil
}

// Stop stops the OMS service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping HFT Order Management Service")

	// Signal stop
	close(s.stopChan)

	// Wait for goroutines to finish
	s.wg.Wait()

	s.isRunning = false
	logrus.Info("HFT Order Management Service stopped")

	return nil
}

// PlaceOrder places a new order
func (s *Service) PlaceOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Validate order
	if err := s.validateOrder(order); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Set order timestamps
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = models.OrderStatusPending

	// Generate order ID if not provided
	if order.ID == "" {
		order.ID = s.generateOrderID()
	}

	// Place order through order manager
	placedOrder, err := s.orderManager.PlaceOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	// Track order
	s.mu.Lock()
	s.activeOrders[placedOrder.ID] = placedOrder
	s.totalOrders++
	s.mu.Unlock()

	// Send order update
	select {
	case s.orderUpdateChan <- placedOrder:
	default:
		// Channel is full, skip
	}

	logrus.WithFields(logrus.Fields{
		"order_id":    placedOrder.ID,
		"strategy_id": placedOrder.StrategyID,
		"symbol":      placedOrder.Symbol,
		"side":        placedOrder.Side,
		"quantity":    placedOrder.Quantity,
		"price":       placedOrder.Price,
	}).Info("Order placed successfully")

	return placedOrder, nil
}

// CancelOrder cancels an existing order
func (s *Service) CancelOrder(ctx context.Context, orderID string) error {
	s.mu.RLock()
	order, exists := s.activeOrders[orderID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Cancel order through order manager
	if err := s.orderManager.CancelOrder(ctx, orderID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Update order status
	order.Status = models.OrderStatusCanceled
	order.UpdatedAt = time.Now()

	// Remove from active orders
	s.mu.Lock()
	delete(s.activeOrders, orderID)
	s.canceledOrders++
	s.mu.Unlock()

	// Send order update
	select {
	case s.orderUpdateChan <- order:
	default:
		// Channel is full, skip
	}

	logrus.WithField("order_id", orderID).Info("Order canceled successfully")
	return nil
}

// GetActiveOrders returns all active orders for a strategy
func (s *Service) GetActiveOrders(ctx context.Context, strategyID string) ([]*models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var orders []*models.Order
	for _, order := range s.activeOrders {
		if order.StrategyID == strategyID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// GetPosition returns the current position for a symbol
func (s *Service) GetPosition(ctx context.Context, strategyID, symbol, exchange string) (*models.Position, error) {
	positionKey := fmt.Sprintf("%s:%s:%s", strategyID, symbol, exchange)
	
	s.mu.RLock()
	position, exists := s.positions[positionKey]
	s.mu.RUnlock()

	if !exists {
		return s.positionManager.GetPosition(ctx, strategyID, symbol, exchange)
	}

	return position, nil
}

// GetAllPositions returns all positions for a strategy
func (s *Service) GetAllPositions(ctx context.Context, strategyID string) ([]*models.Position, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var positions []*models.Position
	for _, position := range s.positions {
		if position.StrategyID == strategyID {
			positions = append(positions, position)
		}
	}

	return positions, nil
}

// GetOrderUpdateChannel returns the order update channel
func (s *Service) GetOrderUpdateChannel() <-chan *models.Order {
	return s.orderUpdateChan
}

// GetFillChannel returns the fill channel
func (s *Service) GetFillChannel() <-chan *models.Fill {
	return s.fillChan
}

// GetPositionChannel returns the position update channel
func (s *Service) GetPositionChannel() <-chan *models.Position {
	return s.positionChan
}

// IsHealthy returns the health status of the service
func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetMetrics returns OMS performance metrics
func (s *Service) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_orders":    s.totalOrders,
		"filled_orders":   s.filledOrders,
		"canceled_orders": s.canceledOrders,
		"rejected_orders": s.rejectedOrders,
		"active_orders":   len(s.activeOrders),
		"positions":       len(s.positions),
		"avg_fill_time":   s.avgFillTime,
	}
}

// validateOrder validates order parameters
func (s *Service) validateOrder(order *models.Order) error {
	if order.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if order.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if order.Quantity.IsZero() || order.Quantity.IsNegative() {
		return fmt.Errorf("quantity must be positive")
	}
	if order.Type == models.OrderTypeLimit && (order.Price.IsZero() || order.Price.IsNegative()) {
		return fmt.Errorf("price must be positive for limit orders")
	}
	return nil
}

// generateOrderID generates a unique order ID
func (s *Service) generateOrderID() string {
	return fmt.Sprintf("hft_%d", time.Now().UnixNano())
}

// loadActiveOrders loads active orders from database
func (s *Service) loadActiveOrders(ctx context.Context) error {
	// Placeholder implementation - would query database for active orders
	logrus.Info("Loading active orders from database")
	return nil
}

// loadPositions loads positions from database
func (s *Service) loadPositions(ctx context.Context) error {
	// Placeholder implementation - would query database for positions
	logrus.Info("Loading positions from database")
	return nil
}

// processOrderUpdates processes order update events
func (s *Service) processOrderUpdates(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case order := <-s.orderUpdateChan:
			s.handleOrderUpdate(order)
		}
	}
}

// processFills processes fill events
func (s *Service) processFills(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case fill := <-s.fillChan:
			s.handleFill(fill)
		}
	}
}

// processPositionUpdates processes position update events
func (s *Service) processPositionUpdates(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case position := <-s.positionChan:
			s.handlePositionUpdate(position)
		}
	}
}

// handleOrderUpdate handles order status updates
func (s *Service) handleOrderUpdate(order *models.Order) {
	logrus.WithFields(logrus.Fields{
		"order_id": order.ID,
		"status":   order.Status,
	}).Debug("Processing order update")

	// Update order in memory
	s.mu.Lock()
	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCanceled {
		delete(s.activeOrders, order.ID)
		if order.Status == models.OrderStatusFilled {
			s.filledOrders++
		}
	} else {
		s.activeOrders[order.ID] = order
	}
	s.mu.Unlock()

	// Store order update in database
	s.storeOrderUpdate(order)
}

// handleFill handles trade fill events
func (s *Service) handleFill(fill *models.Fill) {
	logrus.WithFields(logrus.Fields{
		"order_id": fill.OrderID,
		"trade_id": fill.TradeID,
		"quantity": fill.Quantity,
		"price":    fill.Price,
	}).Debug("Processing fill")

	// Update position
	s.updatePositionFromFill(fill)

	// Store fill in database
	s.storeFill(fill)
}

// handlePositionUpdate handles position updates
func (s *Service) handlePositionUpdate(position *models.Position) {
	positionKey := fmt.Sprintf("%s:%s:%s", position.StrategyID, position.Symbol, position.Exchange)

	s.mu.Lock()
	s.positions[positionKey] = position
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"strategy_id": position.StrategyID,
		"symbol":      position.Symbol,
		"size":        position.Size,
		"pnl":         position.UnrealizedPnL,
	}).Debug("Position updated")
}

// updatePositionFromFill updates position based on fill
func (s *Service) updatePositionFromFill(fill *models.Fill) {
	// Placeholder implementation - would calculate position updates
	logrus.WithField("fill_id", fill.ID).Debug("Updating position from fill")
}

// storeOrderUpdate stores order update in database
func (s *Service) storeOrderUpdate(order *models.Order) {
	// Placeholder implementation - would store in database
	logrus.WithField("order_id", order.ID).Debug("Storing order update")
}

// storeFill stores fill in database
func (s *Service) storeFill(fill *models.Fill) {
	// Placeholder implementation - would store in database
	logrus.WithField("fill_id", fill.ID).Debug("Storing fill")
}
