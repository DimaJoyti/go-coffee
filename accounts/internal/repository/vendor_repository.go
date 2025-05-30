package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
)

// VendorRepository defines the interface for vendor repository operations
type VendorRepository interface {
	// Create creates a new vendor
	Create(ctx context.Context, vendor *models.Vendor) error
	
	// GetByID retrieves a vendor by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Vendor, error)
	
	// List retrieves all vendors with optional pagination
	List(ctx context.Context, offset, limit int) ([]*models.Vendor, error)
	
	// Update updates an existing vendor
	Update(ctx context.Context, vendor *models.Vendor) error
	
	// Delete deletes a vendor by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Count returns the total number of vendors
	Count(ctx context.Context) (int, error)
	
	// Search searches for vendors by name
	Search(ctx context.Context, query string, offset, limit int) ([]*models.Vendor, error)
}
