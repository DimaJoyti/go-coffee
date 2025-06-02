package service

import (
	"context"

	"github.com/DimaJoyti/go-coffee/accounts-service/internal/events"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
)

// AccountServiceInterface defines the interface for account service operations
type AccountServiceInterface interface {
	// Event handlers
	HandleAccountCreated(event events.Event) error
	HandleAccountUpdated(event events.Event) error
	HandleAccountDeleted(event events.Event) error

	// Account operations
	Create(ctx context.Context, input models.AccountInput) (*models.Account, error)
	GetByID(ctx context.Context, id string) (*models.Account, error)
	Update(ctx context.Context, id string, input models.AccountInput) (*models.Account, error)
	Delete(ctx context.Context, id string) error
}

// OrderServiceInterface defines the interface for order service operations
type OrderServiceInterface interface {
	// Event handlers
	HandleOrderCreated(event events.Event) error
	HandleOrderStatusChanged(event events.Event) error
	HandleOrderDeleted(event events.Event) error

	// Order operations
	GetByID(ctx context.Context, id string) (*models.Order, error)
	ListByAccount(ctx context.Context, accountID string) ([]*models.Order, error)
	ListByStatus(ctx context.Context, status models.OrderStatus) ([]*models.Order, error)
}

// ProductServiceInterface defines the interface for product service operations
type ProductServiceInterface interface {
	// Event handlers
	HandleProductCreated(event events.Event) error
	HandleProductUpdated(event events.Event) error
	HandleProductDeleted(event events.Event) error

	// Product operations
	GetByID(ctx context.Context, id string) (*models.Product, error)
	ListByVendor(ctx context.Context, vendorID string) ([]*models.Product, error)
	Search(ctx context.Context, query string) ([]*models.Product, error)
}

// VendorServiceInterface defines the interface for vendor service operations
type VendorServiceInterface interface {
	// Event handlers
	HandleVendorCreated(event events.Event) error
	HandleVendorUpdated(event events.Event) error
	HandleVendorDeleted(event events.Event) error

	// Vendor operations
	GetByID(ctx context.Context, id string) (*models.Vendor, error)
	List(ctx context.Context) ([]*models.Vendor, error)
	Search(ctx context.Context, query string) ([]*models.Vendor, error)
}
