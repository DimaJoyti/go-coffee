package supply

import (
	"context"
)

// Repository represents a supply repository
type Repository interface {
	// GetSupply gets a supply by ID
	GetSupply(ctx context.Context, id string) (*Supply, error)
	
	// CreateSupply creates a new supply
	CreateSupply(ctx context.Context, supply *Supply) error
	
	// UpdateSupply updates a supply
	UpdateSupply(ctx context.Context, id string, supply *Supply) error
	
	// DeleteSupply deletes a supply
	DeleteSupply(ctx context.Context, id string) error
	
	// ListSupplies lists supplies
	ListSupplies(ctx context.Context, userID, currency, status string, page, pageSize int) ([]*Supply, int, error)
}
