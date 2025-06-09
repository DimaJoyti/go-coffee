package integration

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	orderPb "github.com/DimaJoyti/go-coffee/proto/order"
)

// OrderServiceClient represents a client for the order service
type OrderServiceClient struct {
	client orderPb.OrderServiceClient
	conn   *grpc.ClientConn
	logger *logger.Logger
}

// NewOrderServiceClient creates a new order service client
func NewOrderServiceClient(address string, logger *logger.Logger) (*OrderServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	client := orderPb.NewOrderServiceClient(conn)

	return &OrderServiceClient{
		client: client,
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the connection to the order service
func (c *OrderServiceClient) Close() error {
	return c.conn.Close()
}

// GetOrder retrieves an order from the order service
func (c *OrderServiceClient) GetOrder(ctx context.Context, orderID string) (*OrderInfo, error) {
	req := &orderPb.GetOrderRequest{
		Id: orderID,
	}

	resp, err := c.client.GetOrder(ctx, req)
	if err != nil {
		c.logger.WithError(err).WithField("order_id", orderID).Error("Failed to get order from order service")
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return c.convertOrderFromProto(resp), nil
}

// UpdateOrderStatus updates the order status in the order service
func (c *OrderServiceClient) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	req := &orderPb.UpdateOrderStatusRequest{
		Id:     orderID,
		Status: c.convertOrderStatusToProto(status),
	}

	_, err := c.client.UpdateOrderStatus(ctx, req)
	if err != nil {
		c.logger.WithError(err).WithFields(map[string]interface{}{
			"order_id": orderID,
			"status":   status,
		}).Error("Failed to update order status in order service")
		return fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"status":   status,
	}).Info("Order status updated in order service")

	return nil
}

// NotifyOrderReady notifies the order service that an order is ready
func (c *OrderServiceClient) NotifyOrderReady(ctx context.Context, orderID string, estimatedTime int32) error {
	req := &orderPb.NotifyOrderReadyRequest{
		Id:            orderID,
		EstimatedTime: estimatedTime,
		ReadyAt:       time.Now().Unix(),
	}

	_, err := c.client.NotifyOrderReady(ctx, req)
	if err != nil {
		c.logger.WithError(err).WithField("order_id", orderID).Error("Failed to notify order ready")
		return fmt.Errorf("failed to notify order ready: %w", err)
	}

	c.logger.WithField("order_id", orderID).Info("Order ready notification sent")
	return nil
}

// NotifyOrderCompleted notifies the order service that an order is completed
func (c *OrderServiceClient) NotifyOrderCompleted(ctx context.Context, orderID string, actualTime int32) error {
	req := &orderPb.NotifyOrderCompletedRequest{
		Id:          orderID,
		ActualTime:  actualTime,
		CompletedAt: time.Now().Unix(),
	}

	_, err := c.client.NotifyOrderCompleted(ctx, req)
	if err != nil {
		c.logger.WithError(err).WithField("order_id", orderID).Error("Failed to notify order completed")
		return fmt.Errorf("failed to notify order completed: %w", err)
	}

	c.logger.WithField("order_id", orderID).Info("Order completed notification sent")
	return nil
}

// ListPendingOrders retrieves pending orders from the order service
func (c *OrderServiceClient) ListPendingOrders(ctx context.Context) ([]*OrderInfo, error) {
	req := &orderPb.ListOrdersRequest{
		Status: orderPb.OrderStatus_ORDER_STATUS_PENDING,
		Limit:  100, // Reasonable limit
	}

	resp, err := c.client.ListOrders(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list pending orders")
		return nil, fmt.Errorf("failed to list pending orders: %w", err)
	}

	orders := make([]*OrderInfo, len(resp.Orders))
	for i, order := range resp.Orders {
		orders[i] = c.convertOrderFromProto(order)
	}

	c.logger.WithField("order_count", len(orders)).Info("Retrieved pending orders")
	return orders, nil
}

// OrderInfo represents order information from the order service
type OrderInfo struct {
	ID          string
	CustomerID  string
	Items       []*OrderItemInfo
	Status      domain.OrderStatus
	Priority    domain.OrderPriority
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalAmount float64
	Currency    string
}

// OrderItemInfo represents order item information
type OrderItemInfo struct {
	ID           string
	Name         string
	Quantity     int32
	Price        float64
	Instructions string
	Metadata     map[string]string
}

// Conversion methods

func (c *OrderServiceClient) convertOrderFromProto(order *orderPb.OrderResponse) *OrderInfo {
	items := make([]*OrderItemInfo, len(order.Items))
	for i, item := range order.Items {
		items[i] = &OrderItemInfo{
			ID:           item.Id,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Price:        item.Price,
			Instructions: item.Instructions,
			Metadata:     item.Metadata,
		}
	}

	return &OrderInfo{
		ID:          order.Id,
		CustomerID:  order.CustomerId,
		Items:       items,
		Status:      c.convertOrderStatusFromProto(order.Status),
		Priority:    c.convertOrderPriorityFromProto(order.Priority),
		CreatedAt:   order.CreatedAt.AsTime(),
		UpdatedAt:   order.UpdatedAt.AsTime(),
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
	}
}

func (c *OrderServiceClient) convertOrderStatusFromProto(status orderPb.OrderStatus) domain.OrderStatus {
	switch status {
	case orderPb.OrderStatus_ORDER_STATUS_PENDING:
		return domain.OrderStatusPending
	case orderPb.OrderStatus_ORDER_STATUS_PROCESSING:
		return domain.OrderStatusProcessing
	case orderPb.OrderStatus_ORDER_STATUS_COMPLETED:
		return domain.OrderStatusCompleted
	case orderPb.OrderStatus_ORDER_STATUS_CANCELLED:
		return domain.OrderStatusCancelled
	default:
		return domain.OrderStatusUnknown
	}
}

func (c *OrderServiceClient) convertOrderStatusToProto(status domain.OrderStatus) orderPb.OrderStatus {
	switch status {
	case domain.OrderStatusPending:
		return orderPb.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusProcessing:
		return orderPb.OrderStatus_ORDER_STATUS_PROCESSING
	case domain.OrderStatusCompleted:
		return orderPb.OrderStatus_ORDER_STATUS_COMPLETED
	case domain.OrderStatusCancelled:
		return orderPb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderPb.OrderStatus_ORDER_STATUS_UNKNOWN
	}
}

func (c *OrderServiceClient) convertOrderPriorityFromProto(priority orderPb.OrderPriority) domain.OrderPriority {
	switch priority {
	case orderPb.OrderPriority_ORDER_PRIORITY_LOW:
		return domain.OrderPriorityLow
	case orderPb.OrderPriority_ORDER_PRIORITY_NORMAL:
		return domain.OrderPriorityNormal
	case orderPb.OrderPriority_ORDER_PRIORITY_HIGH:
		return domain.OrderPriorityHigh
	case orderPb.OrderPriority_ORDER_PRIORITY_URGENT:
		return domain.OrderPriorityUrgent
	default:
		return domain.OrderPriorityNormal
	}
}

// OrderServiceIntegration provides integration with the order service
type OrderServiceIntegration struct {
	client         *OrderServiceClient
	kitchenService application.KitchenService
	logger         *logger.Logger
}

// NewOrderServiceIntegration creates a new order service integration
func NewOrderServiceIntegration(
	client *OrderServiceClient,
	kitchenService application.KitchenService,
	logger *logger.Logger,
) *OrderServiceIntegration {
	return &OrderServiceIntegration{
		client:         client,
		kitchenService: kitchenService,
		logger:         logger,
	}
}

// SyncPendingOrders synchronizes pending orders from the order service
func (i *OrderServiceIntegration) SyncPendingOrders(ctx context.Context) error {
	i.logger.Info("Starting order synchronization")

	// Get pending orders from order service
	orders, err := i.client.ListPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending orders: %w", err)
	}

	// Convert and add orders to kitchen queue
	for _, orderInfo := range orders {
		kitchenOrder, err := i.convertToKitchenOrder(orderInfo)
		if err != nil {
			i.logger.WithError(err).WithField("order_id", orderInfo.ID).Error("Failed to convert order")
			continue
		}

		// Add to kitchen queue
		_, err = i.kitchenService.AddOrderToQueue(ctx, kitchenOrder)
		if err != nil {
			i.logger.WithError(err).WithField("order_id", orderInfo.ID).Error("Failed to add order to kitchen queue")
			continue
		}

		i.logger.WithField("order_id", orderInfo.ID).Info("Order synchronized to kitchen")
	}

	i.logger.WithField("orders_synced", len(orders)).Info("Order synchronization completed")
	return nil
}

// HandleOrderStatusUpdate handles order status updates from kitchen to order service
func (i *OrderServiceIntegration) HandleOrderStatusUpdate(ctx context.Context, orderID string, status domain.OrderStatus) error {
	return i.client.UpdateOrderStatus(ctx, orderID, status)
}

// HandleOrderCompleted handles order completion notifications
func (i *OrderServiceIntegration) HandleOrderCompleted(ctx context.Context, order *domain.KitchenOrder) error {
	return i.client.NotifyOrderCompleted(ctx, order.ID(), order.ActualTime())
}

// convertToKitchenOrder converts order service order to kitchen order
func (i *OrderServiceIntegration) convertToKitchenOrder(orderInfo *OrderInfo) (*application.AddOrderRequest, error) {
	items := make([]*application.OrderItemRequest, len(orderInfo.Items))
	for j, item := range orderInfo.Items {
		// Map order items to kitchen requirements (simplified mapping)
		requirements := i.mapItemToStationRequirements(item.Name)
		
		items[j] = &application.OrderItemRequest{
			ID:           item.ID,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Instructions: item.Instructions,
			Requirements: requirements,
			Metadata:     item.Metadata,
		}
	}

	return &application.AddOrderRequest{
		ID:         orderInfo.ID,
		CustomerID: orderInfo.CustomerID,
		Items:      items,
		Priority:   orderInfo.Priority,
	}, nil
}

// mapItemToStationRequirements maps order items to kitchen station requirements
func (i *OrderServiceIntegration) mapItemToStationRequirements(itemName string) []domain.StationType {
	// Simplified mapping - in production, this would be more sophisticated
	itemMappings := map[string][]domain.StationType{
		"espresso":    {domain.StationTypeEspresso, domain.StationTypeGrinder},
		"cappuccino":  {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"latte":       {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"americano":   {domain.StationTypeEspresso, domain.StationTypeGrinder},
		"macchiato":   {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer},
		"mocha":       {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer, domain.StationTypeAssembly},
		"frappuccino": {domain.StationTypeEspresso, domain.StationTypeGrinder, domain.StationTypeSteamer, domain.StationTypeAssembly},
	}

	// Check for exact matches first
	if requirements, exists := itemMappings[itemName]; exists {
		return requirements
	}

	// Check for partial matches (case-insensitive)
	for key, requirements := range itemMappings {
		if contains(itemName, key) {
			return requirements
		}
	}

	// Default requirements for unknown items
	return []domain.StationType{domain.StationTypeAssembly}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
