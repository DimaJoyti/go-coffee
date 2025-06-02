package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
)

// ProductRepository implements repository.ProductRepository using PostgreSQL
type ProductRepository struct {
	db *Database
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(db *Database) repository.ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (
			id, vendor_id, name, description, price, is_available, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		product.ID,
		product.VendorID,
		product.Name,
		product.Description,
		product.Price,
		product.IsAvailable,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, vendor_id, name, description, price, is_available, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := r.db.GetDB().GetContext(ctx, &product, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves all products with optional pagination
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*models.Product, error) {
	query := `
		SELECT id, vendor_id, name, description, price, is_available, created_at, updated_at
		FROM products
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	var products []*models.Product
	err := r.db.GetDB().SelectContext(ctx, &products, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// ListByVendor retrieves all products for a specific vendor
func (r *ProductRepository) ListByVendor(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*models.Product, error) {
	query := `
		SELECT id, vendor_id, name, description, price, is_available, created_at, updated_at
		FROM products
		WHERE vendor_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	var products []*models.Product
	err := r.db.GetDB().SelectContext(ctx, &products, query, vendorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by vendor: %w", err)
	}

	return products, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, is_available = $4
		WHERE id = $5
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.IsAvailable,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product by ID
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// Count returns the total number of products
func (r *ProductRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// CountByVendor returns the total number of products for a specific vendor
func (r *ProductRepository) CountByVendor(ctx context.Context, vendorID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE vendor_id = $1`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query, vendorID)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by vendor: %w", err)
	}

	return count, nil
}

// Search searches for products by name
func (r *ProductRepository) Search(ctx context.Context, query string, offset, limit int) ([]*models.Product, error) {
	sqlQuery := `
		SELECT id, vendor_id, name, description, price, is_available, created_at, updated_at
		FROM products
		WHERE name ILIKE $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	var products []*models.Product
	err := r.db.GetDB().SelectContext(ctx, &products, sqlQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}
