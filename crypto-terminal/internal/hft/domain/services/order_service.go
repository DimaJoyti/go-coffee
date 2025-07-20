package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
)

// OrderService defines the domain service interface for order management
type OrderService interface {
	// Order lifecycle management
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*entities.Order, error)
	ValidateOrder(ctx context.Context, order *entities.Order) error
	CalculateOrderValue(order *entities.Order) (valueobjects.Price, error)
	CalculateCommission(order *entities.Order, exchange entities.Exchange) (valueobjects.Commission, error)
	
	// Order execution
	ExecuteOrder(ctx context.Context, order *entities.Order) error
	CancelOrder(ctx context.Context, orderID entities.OrderID, reason string) error
	
	// Order queries
	CanPlaceOrder(ctx context.Context, strategyID entities.StrategyID, order *entities.Order) (bool, error)
	GetOrdersByStrategy(ctx context.Context, strategyID entities.StrategyID) ([]*entities.Order, error)
	GetActiveOrders(ctx context.Context) ([]*entities.Order, error)
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	StrategyID  entities.StrategyID      `json:"strategy_id"`
	Symbol      entities.Symbol          `json:"symbol"`
	Exchange    entities.Exchange        `json:"exchange"`
	Side        valueobjects.OrderSide   `json:"side"`
	OrderType   valueobjects.OrderType   `json:"order_type"`
	Quantity    valueobjects.Quantity    `json:"quantity"`
	Price       valueobjects.Price       `json:"price"`
	StopPrice   *valueobjects.Price      `json:"stop_price,omitempty"`
	TimeInForce valueobjects.TimeInForce `json:"time_in_force"`
	ExpiresAt   *time.Time               `json:"expires_at,omitempty"`
}

// Validate validates the create order request
func (req CreateOrderRequest) Validate() error {
	if req.StrategyID == "" {
		return fmt.Errorf("strategy ID is required")
	}
	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if req.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if !req.Side.IsValid() {
		return fmt.Errorf("invalid order side: %s", req.Side)
	}
	if !req.OrderType.IsValid() {
		return fmt.Errorf("invalid order type: %s", req.OrderType)
	}
	if !req.Quantity.IsValid() {
		return fmt.Errorf("invalid quantity: %s", req.Quantity)
	}
	if req.OrderType.RequiresPrice() && !req.Price.IsValid() {
		return fmt.Errorf("price is required for order type %s", req.OrderType)
	}
	if !req.TimeInForce.IsValid() {
		return fmt.Errorf("invalid time in force: %s", req.TimeInForce)
	}
	if req.TimeInForce == valueobjects.TimeInForceGTD && req.ExpiresAt == nil {
		return fmt.Errorf("expires_at is required for GTD orders")
	}
	return nil
}

// OrderDomainService implements the order domain service
type OrderDomainService struct {
	riskService RiskService
}

// NewOrderDomainService creates a new order domain service
func NewOrderDomainService(riskService RiskService) OrderService {
	return &OrderDomainService{
		riskService: riskService,
	}
}

// CreateOrder creates a new order with domain validation
func (s *OrderDomainService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*entities.Order, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create order request: %w", err)
	}

	// Create order entity
	order, err := entities.NewOrder(
		req.StrategyID,
		req.Symbol,
		req.Exchange,
		req.Side,
		req.OrderType,
		req.Quantity,
		req.Price,
		req.TimeInForce,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order entity: %w", err)
	}

	// Set optional fields
	if req.StopPrice != nil {
		// Note: This would require adding a SetStopPrice method to the Order entity
		// order.SetStopPrice(*req.StopPrice)
	}

	// Validate order against domain rules
	if err := s.ValidateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	return order, nil
}

// ValidateOrder validates an order against domain rules
func (s *OrderDomainService) ValidateOrder(ctx context.Context, order *entities.Order) error {
	// Check risk limits
	if err := s.riskService.ValidateOrderRisk(ctx, order); err != nil {
		return fmt.Errorf("risk validation failed: %w", err)
	}

	// Validate order parameters
	if order.GetQuantity().IsZero() {
		return fmt.Errorf("order quantity cannot be zero")
	}

	if order.GetOrderType().RequiresPrice() && order.GetPrice().IsZero() {
		return fmt.Errorf("price is required for order type %s", order.GetOrderType())
	}

	// Validate symbol format (basic validation)
	symbol := string(order.GetSymbol())
	if len(symbol) < 3 {
		return fmt.Errorf("invalid symbol format: %s", symbol)
	}

	return nil
}

// CalculateOrderValue calculates the total value of an order
func (s *OrderDomainService) CalculateOrderValue(order *entities.Order) (valueobjects.Price, error) {
	if order.GetOrderType() == valueobjects.OrderTypeMarket {
		// For market orders, we can't calculate exact value without current market price
		return valueobjects.Price{}, fmt.Errorf("cannot calculate value for market orders without current price")
	}

	// For limit orders, value = quantity * price
	value := order.GetQuantity().Mul(order.GetPrice().Decimal)
	return valueobjects.Price{Decimal: value}, nil
}

// CalculateCommission calculates the commission for an order
func (s *OrderDomainService) CalculateCommission(order *entities.Order, exchange entities.Exchange) (valueobjects.Commission, error) {
	// This would typically involve exchange-specific commission rates
	// For now, we'll use a simple percentage-based calculation
	
	orderValue, err := s.CalculateOrderValue(order)
	if err != nil {
		return valueobjects.Commission{}, fmt.Errorf("failed to calculate order value: %w", err)
	}

	// Default commission rate (0.1%)
	commissionRate, err := valueobjects.NewPriceFromFloat(0.001)
	if err != nil {
		return valueobjects.Commission{}, fmt.Errorf("failed to create commission rate: %w", err)
	}
	if commissionRate.Decimal.IsZero() {
		return valueobjects.Commission{}, fmt.Errorf("invalid commission rate")
	}

	commissionAmount := orderValue.Mul(commissionRate.Decimal)
	
	// Determine commission asset (typically the quote currency)
	commissionAsset := "USDT" // Default, would be determined by symbol parsing
	
	return valueobjects.NewCommission(commissionAmount, commissionAsset)
}

// ExecuteOrder executes an order (placeholder implementation)
func (s *OrderDomainService) ExecuteOrder(ctx context.Context, order *entities.Order) error {
	// This would involve sending the order to the exchange
	// For now, we'll just confirm the order
	return order.Confirm()
}

// CancelOrder cancels an order
func (s *OrderDomainService) CancelOrder(ctx context.Context, orderID entities.OrderID, reason string) error {
	// This would typically involve:
	// 1. Retrieving the order from repository
	// 2. Validating cancellation is allowed
	// 3. Sending cancel request to exchange
	// 4. Updating order status
	
	// Placeholder implementation
	return fmt.Errorf("cancel order not implemented")
}

// CanPlaceOrder checks if an order can be placed for a strategy
func (s *OrderDomainService) CanPlaceOrder(ctx context.Context, strategyID entities.StrategyID, order *entities.Order) (bool, error) {
	// Check risk limits
	if err := s.riskService.ValidateOrderRisk(ctx, order); err != nil {
		return false, err
	}

	// Check strategy-specific limits
	// This would involve checking:
	// - Strategy is active
	// - Strategy has sufficient capital
	// - Strategy hasn't exceeded order limits
	
	return true, nil
}

// GetOrdersByStrategy retrieves orders for a specific strategy
func (s *OrderDomainService) GetOrdersByStrategy(ctx context.Context, strategyID entities.StrategyID) ([]*entities.Order, error) {
	// This would involve querying the order repository
	// Placeholder implementation
	return nil, fmt.Errorf("get orders by strategy not implemented")
}

// GetActiveOrders retrieves all active orders
func (s *OrderDomainService) GetActiveOrders(ctx context.Context) ([]*entities.Order, error) {
	// This would involve querying the order repository for active orders
	// Placeholder implementation
	return nil, fmt.Errorf("get active orders not implemented")
}

// RiskService defines the interface for risk management
type RiskService interface {
	ValidateOrderRisk(ctx context.Context, order *entities.Order) error
	CheckPositionLimits(ctx context.Context, strategyID entities.StrategyID) error
	CheckDailyLossLimits(ctx context.Context, strategyID entities.StrategyID) error
}
