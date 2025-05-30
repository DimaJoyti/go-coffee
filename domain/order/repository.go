package order

import (
	"context"
	"errors"

	"github.com/DimaJoyti/go-coffee/domain/shared"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// Save saves an order aggregate
	Save(ctx context.Context, order *Order) error

	// FindByID finds an order by ID within tenant context
	FindByID(ctx context.Context, id shared.AggregateID, tenantID shared.TenantID) (*Order, error)
	
	// FindByOrderNumber finds an order by order number within tenant context
	FindByOrderNumber(ctx context.Context, orderNumber string, tenantID shared.TenantID) (*Order, error)
	
	// FindByCustomerID finds orders by customer ID within tenant context
	FindByCustomerID(ctx context.Context, customerID shared.AggregateID, tenantID shared.TenantID) ([]*Order, error)
	
	// FindByLocationID finds orders by location ID within tenant context
	FindByLocationID(ctx context.Context, locationID shared.AggregateID, tenantID shared.TenantID) ([]*Order, error)
	
	// FindByStatus finds orders by status within tenant context
	FindByStatus(ctx context.Context, status OrderStatus, tenantID shared.TenantID) ([]*Order, error)
	
	// FindByPriority finds orders by priority within tenant context
	FindByPriority(ctx context.Context, priority OrderPriority, tenantID shared.TenantID) ([]*Order, error)
	
	// FindActiveOrders finds all active orders within tenant context
	FindActiveOrders(ctx context.Context, tenantID shared.TenantID) ([]*Order, error)
	
	// FindOrdersInDateRange finds orders within a date range for a tenant
	FindOrdersInDateRange(ctx context.Context, dateRange shared.DateRange, tenantID shared.TenantID) ([]*Order, error)
	
	// CountByStatus counts orders by status within tenant context
	CountByStatus(ctx context.Context, status OrderStatus, tenantID shared.TenantID) (int64, error)
	
	// FindAll finds all orders with pagination within tenant context
	FindAll(ctx context.Context, offset, limit int, tenantID shared.TenantID) ([]*Order, int64, error)
	
	// ExistsByOrderNumber checks if an order exists with the given order number within tenant context
	ExistsByOrderNumber(ctx context.Context, orderNumber string, tenantID shared.TenantID) (bool, error)
}

// OrderSpecification defines specifications for order queries
type OrderSpecification interface {
	shared.Specification
}

// ActiveOrderSpecification checks if an order is active
type ActiveOrderSpecification struct {
	*shared.BaseSpecification
}

// NewActiveOrderSpecification creates a new active order specification
func NewActiveOrderSpecification() *ActiveOrderSpecification {
	return &ActiveOrderSpecification{
		BaseSpecification: &shared.BaseSpecification{},
	}
}

// IsSatisfiedBy checks if the order is active
func (spec *ActiveOrderSpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	return order.IsActive()
}

// OrderStatusSpecification checks if an order has a specific status
type OrderStatusSpecification struct {
	*shared.BaseSpecification
	status OrderStatus
}

// NewOrderStatusSpecification creates a new order status specification
func NewOrderStatusSpecification(status OrderStatus) *OrderStatusSpecification {
	return &OrderStatusSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		status:            status,
	}
}

// IsSatisfiedBy checks if the order has the specified status
func (spec *OrderStatusSpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	return order.Status() == spec.status
}

// OrderPrioritySpecification checks if an order has a specific priority
type OrderPrioritySpecification struct {
	*shared.BaseSpecification
	priority OrderPriority
}

// NewOrderPrioritySpecification creates a new order priority specification
func NewOrderPrioritySpecification(priority OrderPriority) *OrderPrioritySpecification {
	return &OrderPrioritySpecification{
		BaseSpecification: &shared.BaseSpecification{},
		priority:          priority,
	}
}

// IsSatisfiedBy checks if the order has the specified priority
func (spec *OrderPrioritySpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	return order.Priority() == spec.priority
}

// CustomerOrderSpecification checks if an order belongs to a specific customer
type CustomerOrderSpecification struct {
	*shared.BaseSpecification
	customerID shared.AggregateID
}

// NewCustomerOrderSpecification creates a new customer order specification
func NewCustomerOrderSpecification(customerID shared.AggregateID) *CustomerOrderSpecification {
	return &CustomerOrderSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		customerID:        customerID,
	}
}

// IsSatisfiedBy checks if the order belongs to the specified customer
func (spec *CustomerOrderSpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	return order.Customer().ID().Equals(spec.customerID)
}

// LocationOrderSpecification checks if an order belongs to a specific location
type LocationOrderSpecification struct {
	*shared.BaseSpecification
	locationID shared.AggregateID
}

// NewLocationOrderSpecification creates a new location order specification
func NewLocationOrderSpecification(locationID shared.AggregateID) *LocationOrderSpecification {
	return &LocationOrderSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		locationID:        locationID,
	}
}

// IsSatisfiedBy checks if the order belongs to the specified location
func (spec *LocationOrderSpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	return order.LocationID().Equals(spec.locationID)
}

// OrderAmountRangeSpecification checks if an order amount is within a range
type OrderAmountRangeSpecification struct {
	*shared.BaseSpecification
	minAmount shared.Money
	maxAmount shared.Money
}

// NewOrderAmountRangeSpecification creates a new order amount range specification
func NewOrderAmountRangeSpecification(minAmount, maxAmount shared.Money) *OrderAmountRangeSpecification {
	return &OrderAmountRangeSpecification{
		BaseSpecification: &shared.BaseSpecification{},
		minAmount:         minAmount,
		maxAmount:         maxAmount,
	}
}

// IsSatisfiedBy checks if the order amount is within the specified range
func (spec *OrderAmountRangeSpecification) IsSatisfiedBy(candidate interface{}) bool {
	order, ok := candidate.(*Order)
	if !ok {
		return false
	}
	
	orderAmount := order.FinalAmount()
	return orderAmount.Amount() >= spec.minAmount.Amount() && 
		   orderAmount.Amount() <= spec.maxAmount.Amount() &&
		   orderAmount.Currency() == spec.minAmount.Currency()
}

// OrderQueryBuilder helps build complex order queries
type OrderQueryBuilder struct {
	specifications []OrderSpecification
}

// NewOrderQueryBuilder creates a new order query builder
func NewOrderQueryBuilder() *OrderQueryBuilder {
	return &OrderQueryBuilder{
		specifications: make([]OrderSpecification, 0),
	}
}

// WithActiveStatus adds active status specification
func (qb *OrderQueryBuilder) WithActiveStatus() *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewActiveOrderSpecification())
	return qb
}

// WithStatus adds status specification
func (qb *OrderQueryBuilder) WithStatus(status OrderStatus) *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewOrderStatusSpecification(status))
	return qb
}

// WithPriority adds priority specification
func (qb *OrderQueryBuilder) WithPriority(priority OrderPriority) *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewOrderPrioritySpecification(priority))
	return qb
}

// WithCustomer adds customer specification
func (qb *OrderQueryBuilder) WithCustomer(customerID shared.AggregateID) *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewCustomerOrderSpecification(customerID))
	return qb
}

// WithLocation adds location specification
func (qb *OrderQueryBuilder) WithLocation(locationID shared.AggregateID) *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewLocationOrderSpecification(locationID))
	return qb
}

// WithAmountRange adds amount range specification
func (qb *OrderQueryBuilder) WithAmountRange(minAmount, maxAmount shared.Money) *OrderQueryBuilder {
	qb.specifications = append(qb.specifications, NewOrderAmountRangeSpecification(minAmount, maxAmount))
	return qb
}

// Build builds the final specification
func (qb *OrderQueryBuilder) Build() OrderSpecification {
	if len(qb.specifications) == 0 {
		return &AlwaysTrueOrderSpecification{}
	}
	
	if len(qb.specifications) == 1 {
		return qb.specifications[0]
	}
	
	// Combine all specifications with AND logic
	result := qb.specifications[0]
	for i := 1; i < len(qb.specifications); i++ {
		result = result.And(qb.specifications[i])
	}
	
	return result
}

// AlwaysTrueOrderSpecification always returns true
type AlwaysTrueOrderSpecification struct {
	*shared.BaseSpecification
}

// IsSatisfiedBy always returns true
func (spec *AlwaysTrueOrderSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return true
}

// OrderDomainService provides domain services for order operations
type OrderDomainService struct {
	repository OrderRepository
}

// NewOrderDomainService creates a new order domain service
func NewOrderDomainService(repository OrderRepository) *OrderDomainService {
	return &OrderDomainService{
		repository: repository,
	}
}

// ValidateOrderNumber validates that order number is unique within tenant
func (s *OrderDomainService) ValidateOrderNumber(ctx context.Context, orderNumber string, tenantID shared.TenantID, excludeID *shared.AggregateID) error {
	exists, err := s.repository.ExistsByOrderNumber(ctx, orderNumber, tenantID)
	if err != nil {
		return err
	}
	
	if exists {
		// If we're updating an existing order, check if it's the same order
		if excludeID != nil {
			existingOrder, err := s.repository.FindByOrderNumber(ctx, orderNumber, tenantID)
			if err != nil {
				return err
			}
			if existingOrder != nil && existingOrder.ID().Equals(*excludeID) {
				return nil // Same order, uniqueness is maintained
			}
		}
		return errors.New("order with this number already exists")
	}
	
	return nil
}

// GetOrdersBySpecification finds orders matching a specification
func (s *OrderDomainService) GetOrdersBySpecification(ctx context.Context, spec OrderSpecification, tenantID shared.TenantID) ([]*Order, error) {
	// This would typically be implemented with database queries
	// For now, we'll get all orders and filter (not efficient for production)
	allOrders, _, err := s.repository.FindAll(ctx, 0, 1000, tenantID)
	if err != nil {
		return nil, err
	}
	
	filteredOrders := make([]*Order, 0)
	for _, order := range allOrders {
		if spec.IsSatisfiedBy(order) {
			filteredOrders = append(filteredOrders, order)
		}
	}
	
	return filteredOrders, nil
}

// CalculateOrderMetrics calculates metrics for order analysis
func (s *OrderDomainService) CalculateOrderMetrics(ctx context.Context, tenantID shared.TenantID) (*OrderMetrics, error) {
	activeCount, err := s.repository.CountByStatus(ctx, OrderStatusPending, tenantID)
	if err != nil {
		return nil, err
	}
	
	completedCount, err := s.repository.CountByStatus(ctx, OrderStatusCompleted, tenantID)
	if err != nil {
		return nil, err
	}
	
	cancelledCount, err := s.repository.CountByStatus(ctx, OrderStatusCancelled, tenantID)
	if err != nil {
		return nil, err
	}
	
	metrics := &OrderMetrics{
		TotalActiveOrders:    activeCount,
		TotalCompletedOrders: completedCount,
		TotalCancelledOrders: cancelledCount,
		CompletionRate:       float64(completedCount) / float64(completedCount + cancelledCount) * 100,
	}
	
	return metrics, nil
}

// OrderMetrics holds order-related metrics
type OrderMetrics struct {
	TotalActiveOrders    int64
	TotalCompletedOrders int64
	TotalCancelledOrders int64
	CompletionRate       float64
	AverageOrderValue    float64
	AverageCompletionTime float64
}
