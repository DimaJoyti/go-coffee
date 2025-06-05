package oms

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// OrderManagerImpl implements the OrderManager interface
type OrderManagerImpl struct {
	config *config.Config
	db     *sql.DB
	redis  *redis.Client
}

// NewOrderManager creates a new order manager
func NewOrderManager(cfg *config.Config, db *sql.DB, redis *redis.Client) OrderManager {
	return &OrderManagerImpl{
		config: cfg,
		db:     db,
		redis:  redis,
	}
}

// PlaceOrder places a new order on the exchange
func (om *OrderManagerImpl) PlaceOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Simulate order placement latency
	startTime := time.Now()

	// Validate order parameters
	if err := om.validateOrderParams(order); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Route order to appropriate exchange
	placedOrder, err := om.routeOrder(ctx, order)
	if err != nil {
		order.Status = models.OrderStatusRejected
		order.ErrorMessage = err.Error()
		return order, fmt.Errorf("order routing failed: %w", err)
	}

	// Calculate latency
	placedOrder.Latency = time.Since(startTime)

	// Store order in database
	if err := om.storeOrder(ctx, placedOrder); err != nil {
		logrus.WithError(err).Error("Failed to store order in database")
	}

	logrus.WithFields(logrus.Fields{
		"order_id":    placedOrder.ID,
		"exchange":    placedOrder.Exchange,
		"symbol":      placedOrder.Symbol,
		"latency_ms":  placedOrder.Latency.Milliseconds(),
	}).Info("Order placed successfully")

	return placedOrder, nil
}

// CancelOrder cancels an existing order
func (om *OrderManagerImpl) CancelOrder(ctx context.Context, orderID string) error {
	// Get order from database
	order, err := om.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Route cancel request to exchange
	if err := om.routeCancelOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to cancel order on exchange: %w", err)
	}

	// Update order status
	order.Status = models.OrderStatusCanceled
	order.UpdatedAt = time.Now()

	// Update order in database
	if err := om.updateOrder(ctx, order); err != nil {
		logrus.WithError(err).Error("Failed to update order in database")
	}

	logrus.WithField("order_id", orderID).Info("Order canceled successfully")
	return nil
}

// ModifyOrder modifies an existing order
func (om *OrderManagerImpl) ModifyOrder(ctx context.Context, orderID string, newPrice, newQuantity decimal.Decimal) error {
	// Get order from database
	order, err := om.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Route modify request to exchange
	if err := om.routeModifyOrder(ctx, order, newPrice, newQuantity); err != nil {
		return fmt.Errorf("failed to modify order on exchange: %w", err)
	}

	// Update order parameters
	order.Price = newPrice
	order.Quantity = newQuantity
	order.UpdatedAt = time.Now()

	// Update order in database
	if err := om.updateOrder(ctx, order); err != nil {
		logrus.WithError(err).Error("Failed to update order in database")
	}

	logrus.WithFields(logrus.Fields{
		"order_id":     orderID,
		"new_price":    newPrice,
		"new_quantity": newQuantity,
	}).Info("Order modified successfully")

	return nil
}

// GetOrder retrieves an order by ID
func (om *OrderManagerImpl) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	// Try Redis cache first
	order, err := om.getOrderFromCache(ctx, orderID)
	if err == nil {
		return order, nil
	}

	// Get from database
	return om.getOrderFromDB(ctx, orderID)
}

// GetActiveOrders returns all active orders for a strategy
func (om *OrderManagerImpl) GetActiveOrders(ctx context.Context, strategyID string) ([]*models.Order, error) {
	// Query database for active orders
	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type, 
		       quantity, price, stop_price, time_in_force, status, filled_quantity,
		       remaining_quantity, avg_fill_price, commission, commission_asset,
		       created_at, updated_at, expires_at, exchange_order_id, error_message
		FROM hft_orders 
		WHERE strategy_id = $1 AND status IN ('new', 'partially_filled')
		ORDER BY created_at DESC
	`

	rows, err := om.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID, &order.ClientOrderID, &order.StrategyID, &order.Symbol,
			&order.Exchange, &order.Side, &order.Type, &order.Quantity,
			&order.Price, &order.StopPrice, &order.TimeInForce, &order.Status,
			&order.FilledQuantity, &order.RemainingQty, &order.AvgFillPrice,
			&order.Commission, &order.CommissionAsset, &order.CreatedAt,
			&order.UpdatedAt, &order.ExpiresAt, &order.ExchangeOrderID,
			&order.ErrorMessage,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan order row")
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderHistory returns order history for a strategy
func (om *OrderManagerImpl) GetOrderHistory(ctx context.Context, strategyID string, limit int) ([]*models.Order, error) {
	// Query database for order history
	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type, 
		       quantity, price, stop_price, time_in_force, status, filled_quantity,
		       remaining_quantity, avg_fill_price, commission, commission_asset,
		       created_at, updated_at, expires_at, exchange_order_id, error_message
		FROM hft_orders 
		WHERE strategy_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := om.db.QueryContext(ctx, query, strategyID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query order history: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID, &order.ClientOrderID, &order.StrategyID, &order.Symbol,
			&order.Exchange, &order.Side, &order.Type, &order.Quantity,
			&order.Price, &order.StopPrice, &order.TimeInForce, &order.Status,
			&order.FilledQuantity, &order.RemainingQty, &order.AvgFillPrice,
			&order.Commission, &order.CommissionAsset, &order.CreatedAt,
			&order.UpdatedAt, &order.ExpiresAt, &order.ExchangeOrderID,
			&order.ErrorMessage,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan order row")
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// validateOrderParams validates order parameters
func (om *OrderManagerImpl) validateOrderParams(order *models.Order) error {
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

// routeOrder routes order to the appropriate exchange
func (om *OrderManagerImpl) routeOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Simulate order routing based on exchange
	switch order.Exchange {
	case "binance":
		return om.routeToBinance(ctx, order)
	case "coinbase":
		return om.routeToCoinbase(ctx, order)
	case "kraken":
		return om.routeToKraken(ctx, order)
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", order.Exchange)
	}
}

// routeToBinance routes order to Binance
func (om *OrderManagerImpl) routeToBinance(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Simulate Binance order placement
	order.Status = models.OrderStatusNew
	order.ExchangeOrderID = fmt.Sprintf("binance_%d", time.Now().UnixNano())
	order.RemainingQty = order.Quantity
	return order, nil
}

// routeToCoinbase routes order to Coinbase
func (om *OrderManagerImpl) routeToCoinbase(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Simulate Coinbase order placement
	order.Status = models.OrderStatusNew
	order.ExchangeOrderID = fmt.Sprintf("coinbase_%d", time.Now().UnixNano())
	order.RemainingQty = order.Quantity
	return order, nil
}

// routeToKraken routes order to Kraken
func (om *OrderManagerImpl) routeToKraken(ctx context.Context, order *models.Order) (*models.Order, error) {
	// Simulate Kraken order placement
	order.Status = models.OrderStatusNew
	order.ExchangeOrderID = fmt.Sprintf("kraken_%d", time.Now().UnixNano())
	order.RemainingQty = order.Quantity
	return order, nil
}

// routeCancelOrder routes cancel request to exchange
func (om *OrderManagerImpl) routeCancelOrder(ctx context.Context, order *models.Order) error {
	// Simulate cancel routing based on exchange
	logrus.WithFields(logrus.Fields{
		"order_id":          order.ID,
		"exchange_order_id": order.ExchangeOrderID,
		"exchange":          order.Exchange,
	}).Info("Routing cancel order to exchange")
	return nil
}

// routeModifyOrder routes modify request to exchange
func (om *OrderManagerImpl) routeModifyOrder(ctx context.Context, order *models.Order, newPrice, newQuantity decimal.Decimal) error {
	// Simulate modify routing based on exchange
	logrus.WithFields(logrus.Fields{
		"order_id":     order.ID,
		"new_price":    newPrice,
		"new_quantity": newQuantity,
		"exchange":     order.Exchange,
	}).Info("Routing modify order to exchange")
	return nil
}

// storeOrder stores order in database
func (om *OrderManagerImpl) storeOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO hft_orders (
			id, client_order_id, strategy_id, symbol, exchange, side, type,
			quantity, price, stop_price, time_in_force, status, filled_quantity,
			remaining_quantity, avg_fill_price, commission, commission_asset,
			created_at, updated_at, expires_at, exchange_order_id, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
	`

	_, err := om.db.ExecContext(ctx, query,
		order.ID, order.ClientOrderID, order.StrategyID, order.Symbol,
		order.Exchange, order.Side, order.Type, order.Quantity,
		order.Price, order.StopPrice, order.TimeInForce, order.Status,
		order.FilledQuantity, order.RemainingQty, order.AvgFillPrice,
		order.Commission, order.CommissionAsset, order.CreatedAt,
		order.UpdatedAt, order.ExpiresAt, order.ExchangeOrderID,
		order.ErrorMessage,
	)

	return err
}

// updateOrder updates order in database
func (om *OrderManagerImpl) updateOrder(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE hft_orders SET
			status = $2, filled_quantity = $3, remaining_quantity = $4,
			avg_fill_price = $5, commission = $6, commission_asset = $7,
			updated_at = $8, error_message = $9
		WHERE id = $1
	`

	_, err := om.db.ExecContext(ctx, query,
		order.ID, order.Status, order.FilledQuantity, order.RemainingQty,
		order.AvgFillPrice, order.Commission, order.CommissionAsset,
		order.UpdatedAt, order.ErrorMessage,
	)

	return err
}

// getOrderFromCache retrieves order from Redis cache
func (om *OrderManagerImpl) getOrderFromCache(ctx context.Context, orderID string) (*models.Order, error) {
	// Placeholder implementation - would retrieve from Redis
	return nil, fmt.Errorf("order not found in cache")
}

// getOrderFromDB retrieves order from database
func (om *OrderManagerImpl) getOrderFromDB(ctx context.Context, orderID string) (*models.Order, error) {
	query := `
		SELECT id, client_order_id, strategy_id, symbol, exchange, side, type,
		       quantity, price, stop_price, time_in_force, status, filled_quantity,
		       remaining_quantity, avg_fill_price, commission, commission_asset,
		       created_at, updated_at, expires_at, exchange_order_id, error_message
		FROM hft_orders
		WHERE id = $1
	`

	order := &models.Order{}
	err := om.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.ClientOrderID, &order.StrategyID, &order.Symbol,
		&order.Exchange, &order.Side, &order.Type, &order.Quantity,
		&order.Price, &order.StopPrice, &order.TimeInForce, &order.Status,
		&order.FilledQuantity, &order.RemainingQty, &order.AvgFillPrice,
		&order.Commission, &order.CommissionAsset, &order.CreatedAt,
		&order.UpdatedAt, &order.ExpiresAt, &order.ExchangeOrderID,
		&order.ErrorMessage,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get order from database: %w", err)
	}

	return order, nil
}
