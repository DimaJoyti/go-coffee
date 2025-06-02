package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/google/uuid"
)

// Order represents a GraphQL order resolver
type Order struct {
	order    *models.Order
	resolver *Resolver
}

// ID resolves the ID field of an order
func (r *Order) ID() string {
	return r.order.ID.String()
}

// AccountID resolves the accountId field of an order
func (r *Order) AccountID() string {
	return r.order.AccountID.String()
}

// Status resolves the status field of an order
func (r *Order) Status() string {
	return string(r.order.Status)
}

// TotalAmount resolves the totalAmount field of an order
func (r *Order) TotalAmount() float64 {
	return r.order.TotalAmount
}

// CreatedAt resolves the createdAt field of an order
func (r *Order) CreatedAt() string {
	return r.order.CreatedAt.Format(time.RFC3339)
}

// UpdatedAt resolves the updatedAt field of an order
func (r *Order) UpdatedAt() string {
	return r.order.UpdatedAt.Format(time.RFC3339)
}

// Account resolves the account field of an order
func (r *Order) Account(ctx context.Context) (*Account, error) {
	if r.order.Account != nil {
		return &Account{
			account:  r.order.Account,
			resolver: r.resolver,
		}, nil
	}

	account, err := r.resolver.accountService.GetByID(ctx, r.order.AccountID)
	if err != nil {
		return nil, err
	}

	return &Account{
		account:  account,
		resolver: r.resolver,
	}, nil
}

// Items resolves the items field of an order
func (r *Order) Items(ctx context.Context) ([]*OrderItem, error) {
	if r.order.Items != nil {
		var itemResolvers []*OrderItem
		for i := range r.order.Items {
			// Get pointer to the item since it's stored by value in r.order.Items
			item := &r.order.Items[i]
			itemResolvers = append(itemResolvers, &OrderItem{
				orderItem: item,
				resolver:  r.resolver,
			})
		}
		return itemResolvers, nil
	}

	// Return empty array if no items
	return []*OrderItem{}, nil
}

// OrderItem represents a GraphQL order item resolver
type OrderItem struct {
	orderItem *models.OrderItem
	resolver  *Resolver
}

// ID resolves the ID field of an order item
func (r *OrderItem) ID() string {
	return r.orderItem.ID.String()
}

// OrderID resolves the orderId field of an order item
func (r *OrderItem) OrderID() string {
	return r.orderItem.OrderID.String()
}

// ProductID resolves the productId field of an order item
func (r *OrderItem) ProductID() string {
	return r.orderItem.ProductID.String()
}

// Quantity resolves the quantity field of an order item
func (r *OrderItem) Quantity() int32 {
	return int32(r.orderItem.Quantity)
}

// UnitPrice resolves the unitPrice field of an order item
func (r *OrderItem) UnitPrice() float64 {
	return r.orderItem.UnitPrice
}

// TotalPrice resolves the totalPrice field of an order item
func (r *OrderItem) TotalPrice() float64 {
	return r.orderItem.TotalPrice
}

// CreatedAt resolves the createdAt field of an order item
func (r *OrderItem) CreatedAt() string {
	return r.orderItem.CreatedAt.Format(time.RFC3339)
}

// Product resolves the product field of an order item
func (r *OrderItem) Product(ctx context.Context) (*Product, error) {
	if r.orderItem.Product != nil {
		return &Product{
			product:  r.orderItem.Product,
			resolver: r.resolver,
		}, nil
	}

	product, err := r.resolver.productService.GetByID(ctx, r.orderItem.ProductID)
	if err != nil {
		return nil, err
	}

	return &Product{
		product:  product,
		resolver: r.resolver,
	}, nil
}

// Order resolves the order query
func (r *Resolver) Order(ctx context.Context, args struct{ ID string }) (*Order, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	order, err := r.orderService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Order{
		order:    order,
		resolver: r,
	}, nil
}

// Orders resolves the orders query
func (r *Resolver) Orders(ctx context.Context, args struct {
	Offset *int32
	Limit  *int32
}) ([]*Order, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	orders, err := r.orderService.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var orderResolvers []*Order
	for _, order := range orders {
		orderResolvers = append(orderResolvers, &Order{
			order:    order,
			resolver: r,
		})
	}

	return orderResolvers, nil
}

// OrdersByAccount resolves the ordersByAccount query
func (r *Resolver) OrdersByAccount(ctx context.Context, args struct {
	AccountID string
	Offset    *int32
	Limit     *int32
}) ([]*Order, error) {
	accountID, err := uuid.Parse(args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}

	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	orders, err := r.orderService.ListByAccount(ctx, accountID, offset, limit)
	if err != nil {
		return nil, err
	}

	var orderResolvers []*Order
	for _, order := range orders {
		orderResolvers = append(orderResolvers, &Order{
			order:    order,
			resolver: r,
		})
	}

	return orderResolvers, nil
}

// OrdersByStatus resolves the ordersByStatus query
func (r *Resolver) OrdersByStatus(ctx context.Context, args struct {
	Status string
	Offset *int32
	Limit  *int32
}) ([]*Order, error) {
	status := models.OrderStatus(args.Status)

	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	orders, err := r.orderService.ListByStatus(ctx, status, offset, limit)
	if err != nil {
		return nil, err
	}

	var orderResolvers []*Order
	for _, order := range orders {
		orderResolvers = append(orderResolvers, &Order{
			order:    order,
			resolver: r,
		})
	}

	return orderResolvers, nil
}

// OrdersCount resolves the ordersCount query
func (r *Resolver) OrdersCount(ctx context.Context) (int32, error) {
	count, err := r.orderService.Count(ctx)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

// CreateOrder resolves the createOrder mutation
func (r *Resolver) CreateOrder(ctx context.Context, args struct {
	Input struct {
		AccountID string
		Items     []struct {
			ProductID string
			Quantity  int32
		}
	}
}) (*Order, error) {
	accountID, err := uuid.Parse(args.Input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}

	var items []models.OrderItemInput
	for _, item := range args.Input.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}

		items = append(items, models.OrderItemInput{
			ProductID: productID,
			Quantity:  int(item.Quantity),
		})
	}

	input := models.OrderInput{
		AccountID: accountID,
		Items:     items,
	}

	order, err := r.orderService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return &Order{
		order:    order,
		resolver: r,
	}, nil
}

// UpdateOrderStatus resolves the updateOrderStatus mutation
func (r *Resolver) UpdateOrderStatus(ctx context.Context, args struct {
	ID     string
	Status string
}) (*Order, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	status := models.OrderStatus(args.Status)

	order, err := r.orderService.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	return &Order{
		order:    order,
		resolver: r,
	}, nil
}

// DeleteOrder resolves the deleteOrder mutation
func (r *Resolver) DeleteOrder(ctx context.Context, args struct{ ID string }) (bool, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return false, fmt.Errorf("invalid ID: %w", err)
	}

	err = r.orderService.Delete(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}
