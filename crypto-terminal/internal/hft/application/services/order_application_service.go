package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/entities"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/repositories"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/services"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/hft/domain/valueobjects"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OrderApplicationService handles order-related use cases
type OrderApplicationService struct {
	orderRepo       repositories.OrderRepository
	strategyRepo    repositories.StrategyRepository
	eventStore      repositories.EventStore
	orderService    services.OrderService
	riskService     services.RiskService
	tracer          trace.Tracer
}

// NewOrderApplicationService creates a new order application service
func NewOrderApplicationService(
	orderRepo repositories.OrderRepository,
	strategyRepo repositories.StrategyRepository,
	eventStore repositories.EventStore,
	orderService services.OrderService,
	riskService services.RiskService,
) *OrderApplicationService {
	return &OrderApplicationService{
		orderRepo:    orderRepo,
		strategyRepo: strategyRepo,
		eventStore:   eventStore,
		orderService: orderService,
		riskService:  riskService,
		tracer:       otel.Tracer("hft.order.application"),
	}
}

// PlaceOrderCommand represents a command to place an order
type PlaceOrderCommand struct {
	StrategyID  string  `json:"strategy_id" validate:"required"`
	Symbol      string  `json:"symbol" validate:"required"`
	Exchange    string  `json:"exchange" validate:"required"`
	Side        string  `json:"side" validate:"required,oneof=buy sell"`
	OrderType   string  `json:"order_type" validate:"required"`
	Quantity    string  `json:"quantity" validate:"required"`
	Price       string  `json:"price"`
	StopPrice   *string `json:"stop_price,omitempty"`
	TimeInForce string  `json:"time_in_force" validate:"required"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// PlaceOrderResult represents the result of placing an order
type PlaceOrderResult struct {
	OrderID     string    `json:"order_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	EstimatedValue string `json:"estimated_value,omitempty"`
	Commission  string    `json:"commission,omitempty"`
}

// PlaceOrder handles the place order use case
func (s *OrderApplicationService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) (*PlaceOrderResult, error) {
	ctx, span := s.tracer.Start(ctx, "OrderApplicationService.PlaceOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("strategy_id", cmd.StrategyID),
		attribute.String("symbol", cmd.Symbol),
		attribute.String("exchange", cmd.Exchange),
		attribute.String("side", cmd.Side),
		attribute.String("order_type", cmd.OrderType),
		attribute.String("quantity", cmd.Quantity),
	)

	// Validate strategy exists and is active
	strategy, err := s.strategyRepo.FindByID(ctx, entities.StrategyID(cmd.StrategyID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find strategy: %w", err)
	}

	if !strategy.IsRunning() {
		err := fmt.Errorf("strategy %s is not running, current status: %s", cmd.StrategyID, strategy.GetStatus())
		span.RecordError(err)
		return nil, err
	}

	// Convert command to domain types
	createOrderReq, err := s.commandToCreateOrderRequest(cmd)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to convert command: %w", err)
	}

	// Create order through domain service
	order, err := s.orderService.CreateOrder(ctx, createOrderReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Check if order can be placed (risk checks, etc.)
	canPlace, err := s.orderService.CanPlaceOrder(ctx, entities.StrategyID(cmd.StrategyID), order)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to validate order placement: %w", err)
	}

	if !canPlace {
		err := fmt.Errorf("order cannot be placed due to risk or strategy limits")
		span.RecordError(err)
		return nil, err
	}

	// Calculate order value and commission
	orderValue, err := s.orderService.CalculateOrderValue(order)
	if err != nil {
		// Log warning but don't fail the order
		span.AddEvent("failed to calculate order value", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	commission, err := s.orderService.CalculateCommission(order, entities.Exchange(cmd.Exchange))
	if err != nil {
		// Log warning but don't fail the order
		span.AddEvent("failed to calculate commission", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	// Save order to repository
	if err := s.orderRepo.Save(ctx, order); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Save order events to event store
	if err := s.eventStore.SaveEvents(ctx, string(order.GetID()), s.convertOrderEventsToGeneric(order.GetEvents())); err != nil {
		// Log error but don't fail the order placement
		span.AddEvent("failed to save order events", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	// Execute order (send to exchange)
	if err := s.orderService.ExecuteOrder(ctx, order); err != nil {
		span.RecordError(err)
		// Update order status to rejected
		order.Reject(fmt.Sprintf("execution failed: %s", err.Error()))
		s.orderRepo.Update(ctx, order)
		return nil, fmt.Errorf("failed to execute order: %w", err)
	}

	// Update order in repository after execution
	if err := s.orderRepo.Update(ctx, order); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update order after execution: %w", err)
	}

	span.SetAttributes(
		attribute.String("order_id", string(order.GetID())),
		attribute.String("order_status", string(order.GetStatus())),
	)

	// Build result
	result := &PlaceOrderResult{
		OrderID:   string(order.GetID()),
		Status:    string(order.GetStatus()),
		Message:   "Order placed successfully",
		CreatedAt: order.GetCreatedAt(),
	}

	if !orderValue.IsZero() {
		result.EstimatedValue = orderValue.String()
	}

	if commission.IsValid() {
		result.Commission = fmt.Sprintf("%s %s", commission.Amount.String(), commission.Asset)
	}

	return result, nil
}

// CancelOrderCommand represents a command to cancel an order
type CancelOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	Reason  string `json:"reason"`
}

// CancelOrderResult represents the result of canceling an order
type CancelOrderResult struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CanceledAt time.Time `json:"canceled_at"`
}

// CancelOrder handles the cancel order use case
func (s *OrderApplicationService) CancelOrder(ctx context.Context, cmd CancelOrderCommand) (*CancelOrderResult, error) {
	ctx, span := s.tracer.Start(ctx, "OrderApplicationService.CancelOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", cmd.OrderID),
		attribute.String("reason", cmd.Reason),
	)

	// Find order
	order, err := s.orderRepo.FindByID(ctx, entities.OrderID(cmd.OrderID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// Check if order can be canceled
	if !order.IsActive() {
		err := fmt.Errorf("order %s cannot be canceled, current status: %s", cmd.OrderID, order.GetStatus())
		span.RecordError(err)
		return nil, err
	}

	// Cancel order through domain service
	if err := s.orderService.CancelOrder(ctx, entities.OrderID(cmd.OrderID), cmd.Reason); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Cancel order entity
	if err := order.Cancel(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to cancel order entity: %w", err)
	}

	// Update order in repository
	if err := s.orderRepo.Update(ctx, order); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update canceled order: %w", err)
	}

	// Save order events to event store
	if err := s.eventStore.SaveEvents(ctx, string(order.GetID()), s.convertOrderEventsToGeneric(order.GetEvents())); err != nil {
		// Log error but don't fail the cancellation
		span.AddEvent("failed to save order events", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	span.SetAttributes(
		attribute.String("order_status", string(order.GetStatus())),
	)

	return &CancelOrderResult{
		OrderID:    string(order.GetID()),
		Status:     string(order.GetStatus()),
		Message:    "Order canceled successfully",
		CanceledAt: order.GetUpdatedAt(),
	}, nil
}

// GetOrderQuery represents a query to get order details
type GetOrderQuery struct {
	OrderID string `json:"order_id" validate:"required"`
}

// OrderDTO represents order data transfer object
type OrderDTO struct {
	ID              string    `json:"id"`
	StrategyID      string    `json:"strategy_id"`
	Symbol          string    `json:"symbol"`
	Exchange        string    `json:"exchange"`
	Side            string    `json:"side"`
	OrderType       string    `json:"order_type"`
	Quantity        string    `json:"quantity"`
	Price           string    `json:"price"`
	Status          string    `json:"status"`
	FilledQuantity  string    `json:"filled_quantity"`
	RemainingQty    string    `json:"remaining_quantity"`
	AvgFillPrice    string    `json:"avg_fill_price"`
	Commission      string    `json:"commission"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Latency         string    `json:"latency"`
}

// GetOrder handles the get order query
func (s *OrderApplicationService) GetOrder(ctx context.Context, query GetOrderQuery) (*OrderDTO, error) {
	ctx, span := s.tracer.Start(ctx, "OrderApplicationService.GetOrder")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", query.OrderID))

	// Find order
	order, err := s.orderRepo.FindByID(ctx, entities.OrderID(query.OrderID))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// Convert to DTO
	dto := s.orderToDTO(order)
	return dto, nil
}

// Helper methods

// commandToCreateOrderRequest converts command to domain create order request
func (s *OrderApplicationService) commandToCreateOrderRequest(cmd PlaceOrderCommand) (services.CreateOrderRequest, error) {
	// Parse quantity
	quantity, err := valueobjects.NewQuantityFromString(cmd.Quantity)
	if err != nil {
		return services.CreateOrderRequest{}, fmt.Errorf("invalid quantity: %w", err)
	}

	// Parse price
	var price valueobjects.Price
	if cmd.Price != "" {
		price, err = valueobjects.NewPriceFromString(cmd.Price)
		if err != nil {
			return services.CreateOrderRequest{}, fmt.Errorf("invalid price: %w", err)
		}
	}

	// Parse stop price if provided
	var stopPrice *valueobjects.Price
	if cmd.StopPrice != nil {
		sp, err := valueobjects.NewPriceFromString(*cmd.StopPrice)
		if err != nil {
			return services.CreateOrderRequest{}, fmt.Errorf("invalid stop price: %w", err)
		}
		stopPrice = &sp
	}

	return services.CreateOrderRequest{
		StrategyID:  entities.StrategyID(cmd.StrategyID),
		Symbol:      entities.Symbol(cmd.Symbol),
		Exchange:    entities.Exchange(cmd.Exchange),
		Side:        valueobjects.OrderSide(cmd.Side),
		OrderType:   valueobjects.OrderType(cmd.OrderType),
		Quantity:    quantity,
		Price:       price,
		StopPrice:   stopPrice,
		TimeInForce: valueobjects.TimeInForce(cmd.TimeInForce),
		ExpiresAt:   cmd.ExpiresAt,
	}, nil
}

// orderToDTO converts order entity to DTO
func (s *OrderApplicationService) orderToDTO(order *entities.Order) *OrderDTO {
	commission := order.GetCommission()
	commissionStr := ""
	if commission.IsValid() {
		commissionStr = fmt.Sprintf("%s %s", commission.Amount.String(), commission.Asset)
	}

	return &OrderDTO{
		ID:             string(order.GetID()),
		StrategyID:     string(order.GetStrategyID()),
		Symbol:         string(order.GetSymbol()),
		Exchange:       string(order.GetExchange()),
		Side:           string(order.GetSide()),
		OrderType:      string(order.GetOrderType()),
		Quantity:       order.GetQuantity().String(),
		Price:          order.GetPrice().String(),
		Status:         string(order.GetStatus()),
		FilledQuantity: order.GetFilledQuantity().String(),
		RemainingQty:   order.GetRemainingQuantity().String(),
		AvgFillPrice:   order.GetAvgFillPrice().String(),
		Commission:     commissionStr,
		CreatedAt:      order.GetCreatedAt(),
		UpdatedAt:      order.GetUpdatedAt(),
		Latency:        order.GetLatency().String(),
	}
}

// convertOrderEventsToGeneric converts order events to generic domain events
func (s *OrderApplicationService) convertOrderEventsToGeneric(events []valueobjects.OrderEvent) []repositories.DomainEvent {
	genericEvents := make([]repositories.DomainEvent, len(events))
	for i, event := range events {
		genericEvents[i] = repositories.DomainEvent{
			ID:          fmt.Sprintf("order_event_%d", i),
			EventType:   string(event.Type),
			EventData:   event.Data,
			Timestamp:   event.Timestamp,
			Version:     i + 1,
		}
	}
	return genericEvents
}
