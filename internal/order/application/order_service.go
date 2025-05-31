package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, orderID string) (*domain.Order, error)
	GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, orderID string) error
	List(ctx context.Context, filters OrderFilters) ([]*domain.Order, error)
}

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, paymentID string) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	List(ctx context.Context, filters PaymentFilters) ([]*domain.Payment, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DomainEvent) error
	PublishBatch(ctx context.Context, events []*domain.DomainEvent) error
}

// KitchenService defines the interface for kitchen service integration
type KitchenService interface {
	SubmitOrder(ctx context.Context, order *domain.Order) error
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
	GetEstimatedTime(ctx context.Context, items []*domain.OrderItem) (int32, error)
}

// AuthService defines the interface for authentication service integration
type AuthService interface {
	ValidateUser(ctx context.Context, userID string) (*UserInfo, error)
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
}

// OrderService implements the order management use cases
type OrderService struct {
	orderRepo      OrderRepository
	paymentRepo    PaymentRepository
	eventPublisher EventPublisher
	kitchenService KitchenService
	authService    AuthService
	logger         *logger.Logger
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo OrderRepository,
	paymentRepo PaymentRepository,
	eventPublisher EventPublisher,
	kitchenService KitchenService,
	authService AuthService,
	logger *logger.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		eventPublisher: eventPublisher,
		kitchenService: kitchenService,
		authService:    authService,
		logger:         logger,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Validate user
	userInfo, err := s.authService.ValidateUser(ctx, req.CustomerID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to validate user")
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	// Convert request to domain items
	items := make([]*domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		customizations := make([]*domain.Customization, len(item.Customizations))
		for j, custom := range item.Customizations {
			customizations[j] = &domain.Customization{
				ID:         custom.ID,
				Name:       custom.Name,
				Value:      custom.Value,
				ExtraPrice: custom.ExtraPrice,
			}
		}

		items[i] = &domain.OrderItem{
			ID:             generateItemID(),
			ProductID:      item.ProductID,
			Name:           item.Name,
			Description:    item.Description,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			Customizations: customizations,
			Metadata:       item.Metadata,
		}
	}

	// Create order
	order, err := domain.NewOrder(req.CustomerID, items)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create order")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Set additional order properties
	if req.SpecialInstructions != "" {
		order.SpecialInstructions = req.SpecialInstructions
	}
	
	if req.DeliveryAddress != nil {
		order.DeliveryAddress = &domain.Address{
			Street:     req.DeliveryAddress.Street,
			City:       req.DeliveryAddress.City,
			State:      req.DeliveryAddress.State,
			PostalCode: req.DeliveryAddress.PostalCode,
			Country:    req.DeliveryAddress.Country,
			Latitude:   req.DeliveryAddress.Latitude,
			Longitude:  req.DeliveryAddress.Longitude,
		}
		order.IsDelivery = true
	}

	// Get estimated preparation time from kitchen service
	estimatedTime, err := s.kitchenService.GetEstimatedTime(ctx, items)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get estimated time from kitchen service, using default")
		estimatedTime = 300 // 5 minutes default
	}
	order.EstimatedTime = estimatedTime

	// Save order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to save order")
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order created event
	event := domain.NewOrderCreatedEvent(order)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish order created event")
		// Don't fail the request for event publishing errors
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"total":       order.TotalAmount,
	}).Info("Order created successfully")

	return &CreateOrderResponse{
		OrderID:       order.ID,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		EstimatedTime: order.EstimatedTime,
		Status:        order.Status.String(),
		CreatedAt:     order.CreatedAt,
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get order")
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if user has permission to view this order
	if req.CustomerID != "" && order.CustomerID != req.CustomerID {
		return nil, errors.New("unauthorized to view this order")
	}

	return s.buildOrderResponse(order), nil
}

// ConfirmOrder confirms an order and submits it to the kitchen
func (s *OrderService) ConfirmOrder(ctx context.Context, req *ConfirmOrderRequest) (*ConfirmOrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be confirmed
	if !order.CanTransitionTo(domain.OrderStatusConfirmed) {
		return nil, errors.New("order cannot be confirmed in current status")
	}

	// Update order status
	previousStatus := order.Status
	if err := order.UpdateStatus(domain.OrderStatusConfirmed); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Set payment method if provided
	if req.PaymentMethod != domain.PaymentMethodUnknown {
		order.PaymentMethod = req.PaymentMethod
	}

	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Submit order to kitchen service
	if err := s.kitchenService.SubmitOrder(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to submit order to kitchen service")
		// Don't fail the confirmation for kitchen service errors
	}

	// Publish order confirmed event
	event := domain.NewOrderStatusChangedEvent(order, previousStatus)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish order confirmed event")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
	}).Info("Order confirmed successfully")

	return &ConfirmOrderResponse{
		OrderID:       order.ID,
		Status:        order.Status.String(),
		EstimatedTime: order.EstimatedTime,
		UpdatedAt:     order.UpdatedAt,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *UpdateOrderStatusRequest) (*UpdateOrderStatusResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Parse status
	var newStatus domain.OrderStatus
	switch req.Status {
	case "PREPARING":
		newStatus = domain.OrderStatusPreparing
	case "READY":
		newStatus = domain.OrderStatusReady
	case "COMPLETED":
		newStatus = domain.OrderStatusCompleted
	case "CANCELLED":
		newStatus = domain.OrderStatusCancelled
	default:
		return nil, errors.New("invalid order status")
	}

	// Update order status
	previousStatus := order.Status
	if err := order.UpdateStatus(newStatus); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Set actual time if order is completed
	if newStatus == domain.OrderStatusCompleted && order.ConfirmedAt != nil {
		order.ActualTime = int32(time.Since(*order.ConfirmedAt).Seconds())
	}

	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish order status changed event
	event := domain.NewOrderStatusChangedEvent(order, previousStatus)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish order status changed event")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":   order.ID,
		"new_status": newStatus.String(),
		"old_status": previousStatus.String(),
	}).Info("Order status updated successfully")

	return &UpdateOrderStatusResponse{
		OrderID:   order.ID,
		Status:    order.Status.String(),
		UpdatedAt: order.UpdatedAt,
	}, nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if user has permission to cancel this order
	if req.CustomerID != "" && order.CustomerID != req.CustomerID {
		return nil, errors.New("unauthorized to cancel this order")
	}

	// Check if order can be cancelled
	if !order.CanTransitionTo(domain.OrderStatusCancelled) {
		return nil, errors.New("order cannot be cancelled in current status")
	}

	// Update order status
	previousStatus := order.Status
	if err := order.UpdateStatus(domain.OrderStatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish order cancelled event
	event := domain.NewOrderStatusChangedEvent(order, previousStatus)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish order cancelled event")
	}

	s.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"reason":      req.Reason,
	}).Info("Order cancelled successfully")

	return &CancelOrderResponse{
		OrderID:   order.ID,
		Status:    order.Status.String(),
		UpdatedAt: order.UpdatedAt,
	}, nil
}

// Helper methods

// buildOrderResponse builds an order response from domain order
func (s *OrderService) buildOrderResponse(order *domain.Order) *GetOrderResponse {
	items := make([]*OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		customizations := make([]*CustomizationResponse, len(item.Customizations))
		for j, custom := range item.Customizations {
			customizations[j] = &CustomizationResponse{
				ID:         custom.ID,
				Name:       custom.Name,
				Value:      custom.Value,
				ExtraPrice: custom.ExtraPrice,
			}
		}

		items[i] = &OrderItemResponse{
			ID:             item.ID,
			ProductID:      item.ProductID,
			Name:           item.Name,
			Description:    item.Description,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			TotalPrice:     item.TotalPrice,
			Customizations: customizations,
		}
	}

	var deliveryAddress *AddressResponse
	if order.DeliveryAddress != nil {
		deliveryAddress = &AddressResponse{
			Street:     order.DeliveryAddress.Street,
			City:       order.DeliveryAddress.City,
			State:      order.DeliveryAddress.State,
			PostalCode: order.DeliveryAddress.PostalCode,
			Country:    order.DeliveryAddress.Country,
			Latitude:   order.DeliveryAddress.Latitude,
			Longitude:  order.DeliveryAddress.Longitude,
		}
	}

	return &GetOrderResponse{
		OrderID:             order.ID,
		CustomerID:          order.CustomerID,
		Items:               items,
		Status:              order.Status.String(),
		Priority:            order.Priority,
		TotalAmount:         order.TotalAmount,
		Currency:            order.Currency,
		PaymentMethod:       order.PaymentMethod,
		EstimatedTime:       order.EstimatedTime,
		ActualTime:          order.ActualTime,
		SpecialInstructions: order.SpecialInstructions,
		DeliveryAddress:     deliveryAddress,
		IsDelivery:          order.IsDelivery,
		CreatedAt:           order.CreatedAt,
		UpdatedAt:           order.UpdatedAt,
		ConfirmedAt:         order.ConfirmedAt,
		CompletedAt:         order.CompletedAt,
	}
}

// generateItemID generates a unique item ID
func generateItemID() string {
	return "item_" + time.Now().Format("20060102150405") + "_" + generateRandomString(6)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
