package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// OrderService handles business logic for orders
type OrderService struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	accountRepo   repository.AccountRepository
	productRepo   repository.ProductRepository
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	accountRepo repository.AccountRepository,
	productRepo repository.ProductRepository,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		accountRepo:   accountRepo,
		productRepo:   productRepo,
	}
}

// Create creates a new order
func (s *OrderService) Create(ctx context.Context, input models.OrderInput) (*models.Order, error) {
	// Check if account exists
	account, err := s.accountRepo.GetByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Calculate total amount and create order items
	var totalAmount float64
	var orderItems []*models.OrderItem

	for _, itemInput := range input.Items {
		// Get product
		product, err := s.productRepo.GetByID(ctx, itemInput.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		// Check if product is available
		if !product.IsAvailable {
			return nil, fmt.Errorf("product %s is not available", product.Name)
		}

		// Calculate item total
		itemTotal := float64(itemInput.Quantity) * product.Price
		totalAmount += itemTotal
	}

	// Create the order
	order := models.NewOrder(input, totalAmount)
	order.Account = account

	// Save the order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for _, itemInput := range input.Items {
		// Get product
		product, err := s.productRepo.GetByID(ctx, itemInput.ProductID)
		if err != nil {
			// This shouldn't happen since we already checked above
			continue
		}

		// Create order item
		orderItem := models.NewOrderItem(
			order.ID,
			itemInput.ProductID,
			itemInput.Quantity,
			product.Price,
		)
		orderItem.Product = product

		// Save order item
		if err := s.orderItemRepo.Create(ctx, orderItem); err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		orderItems = append(orderItems, orderItem)
	}

	order.Items = orderItems

	return order, nil
}

// GetByID retrieves an order by ID
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Load account
	account, err := s.accountRepo.GetByID(ctx, order.AccountID)
	if err == nil {
		order.Account = account
	}

	// Load order items
	items, err := s.orderItemRepo.ListByOrder(ctx, order.ID)
	if err == nil {
		// Load products for each item
		for _, item := range items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err == nil {
				item.Product = product
			}
		}
		order.Items = items
	}

	return order, nil
}

// List retrieves all orders with optional pagination
func (s *OrderService) List(ctx context.Context, offset, limit int) ([]*models.Order, error) {
	orders, err := s.orderRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	// Load related data for each order
	for _, order := range orders {
		// Load account
		account, err := s.accountRepo.GetByID(ctx, order.AccountID)
		if err == nil {
			order.Account = account
		}

		// Load order items
		items, err := s.orderItemRepo.ListByOrder(ctx, order.ID)
		if err == nil {
			order.Items = items
		}
	}

	return orders, nil
}

// ListByAccount retrieves all orders for a specific account
func (s *OrderService) ListByAccount(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]*models.Order, error) {
	// Check if account exists
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	orders, err := s.orderRepo.ListByAccount(ctx, accountID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by account: %w", err)
	}

	// Set account for each order
	for _, order := range orders {
		order.Account = account

		// Load order items
		items, err := s.orderItemRepo.ListByOrder(ctx, order.ID)
		if err == nil {
			order.Items = items
		}
	}

	return orders, nil
}

// ListByStatus retrieves all orders with a specific status
func (s *OrderService) ListByStatus(ctx context.Context, status models.OrderStatus, offset, limit int) ([]*models.Order, error) {
	orders, err := s.orderRepo.ListByStatus(ctx, status, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders by status: %w", err)
	}

	// Load related data for each order
	for _, order := range orders {
		// Load account
		account, err := s.accountRepo.GetByID(ctx, order.AccountID)
		if err == nil {
			order.Account = account
		}

		// Load order items
		items, err := s.orderItemRepo.ListByOrder(ctx, order.ID)
		if err == nil {
			order.Items = items
		}
	}

	return orders, nil
}

// UpdateStatus updates the status of an order
func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) (*models.Order, error) {
	// Get the existing order
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Update the status
	order.UpdateStatus(status)

	// Save the updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Load account
	account, err := s.accountRepo.GetByID(ctx, order.AccountID)
	if err == nil {
		order.Account = account
	}

	// Load order items
	items, err := s.orderItemRepo.ListByOrder(ctx, order.ID)
	if err == nil {
		// Load products for each item
		for _, item := range items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err == nil {
				item.Product = product
			}
		}
		order.Items = items
	}

	return order, nil
}

// Delete deletes an order by ID
func (s *OrderService) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete order items first
	if err := s.orderItemRepo.DeleteByOrder(ctx, id); err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Delete the order
	if err := s.orderRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

// Count returns the total number of orders
func (s *OrderService) Count(ctx context.Context) (int, error) {
	count, err := s.orderRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}
	return count, nil
}

// CountByAccount returns the total number of orders for a specific account
func (s *OrderService) CountByAccount(ctx context.Context, accountID uuid.UUID) (int, error) {
	// Check if account exists
	_, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("account not found: %w", err)
	}

	count, err := s.orderRepo.CountByAccount(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by account: %w", err)
	}
	return count, nil
}

// CountByStatus returns the total number of orders with a specific status
func (s *OrderService) CountByStatus(ctx context.Context, status models.OrderStatus) (int, error) {
	count, err := s.orderRepo.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by status: %w", err)
	}
	return count, nil
}
