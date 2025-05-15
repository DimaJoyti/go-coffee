package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
)

// ProductRepository defines the interface for product repository operations
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *models.Product) error
	
	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	
	// List retrieves all products with optional pagination
	List(ctx context.Context, offset, limit int) ([]*models.Product, error)
	
	// ListByVendor retrieves all products for a specific vendor
	ListByVendor(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*models.Product, error)
	
	// Update updates an existing product
	Update(ctx context.Context, product *models.Product) error
	
	// Delete deletes a product by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Count returns the total number of products
	Count(ctx context.Context) (int, error)
	
	// CountByVendor returns the total number of products for a specific vendor
	CountByVendor(ctx context.Context, vendorID uuid.UUID) (int, error)
	
	// Search searches for products by name
	Search(ctx context.Context, query string, offset, limit int) ([]*models.Product, error)
}
