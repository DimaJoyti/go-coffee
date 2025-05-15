package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
)

// VendorRepository implements repository.VendorRepository using PostgreSQL
type VendorRepository struct {
	db *Database
}

// NewVendorRepository creates a new PostgreSQL vendor repository
func NewVendorRepository(db *Database) repository.VendorRepository {
	return &VendorRepository{db: db}
}

// Create creates a new vendor
func (r *VendorRepository) Create(ctx context.Context, vendor *models.Vendor) error {
	query := `
		INSERT INTO vendors (
			id, name, description, contact_email, contact_phone, address, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		vendor.ID,
		vendor.Name,
		vendor.Description,
		vendor.ContactEmail,
		vendor.ContactPhone,
		vendor.Address,
		vendor.IsActive,
		vendor.CreatedAt,
		vendor.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create vendor: %w", err)
	}

	return nil
}

// GetByID retrieves a vendor by ID
func (r *VendorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vendor, error) {
	query := `
		SELECT id, name, description, contact_email, contact_phone, address, is_active, created_at, updated_at
		FROM vendors
		WHERE id = $1
	`

	var vendor models.Vendor
	err := r.db.GetDB().GetContext(ctx, &vendor, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("vendor not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}

	return &vendor, nil
}

// List retrieves all vendors with optional pagination
func (r *VendorRepository) List(ctx context.Context, offset, limit int) ([]*models.Vendor, error) {
	query := `
		SELECT id, name, description, contact_email, contact_phone, address, is_active, created_at, updated_at
		FROM vendors
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	var vendors []*models.Vendor
	err := r.db.GetDB().SelectContext(ctx, &vendors, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors: %w", err)
	}

	return vendors, nil
}

// Update updates an existing vendor
func (r *VendorRepository) Update(ctx context.Context, vendor *models.Vendor) error {
	query := `
		UPDATE vendors
		SET name = $1, description = $2, contact_email = $3, contact_phone = $4, address = $5, is_active = $6
		WHERE id = $7
	`

	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		vendor.Name,
		vendor.Description,
		vendor.ContactEmail,
		vendor.ContactPhone,
		vendor.Address,
		vendor.IsActive,
		vendor.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update vendor: %w", err)
	}

	return nil
}

// Delete deletes a vendor by ID
func (r *VendorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vendors WHERE id = $1`

	_, err := r.db.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete vendor: %w", err)
	}

	return nil
}

// Count returns the total number of vendors
func (r *VendorRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM vendors`

	var count int
	err := r.db.GetDB().GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count vendors: %w", err)
	}

	return count, nil
}

// Search searches for vendors by name
func (r *VendorRepository) Search(ctx context.Context, query string, offset, limit int) ([]*models.Vendor, error) {
	sqlQuery := `
		SELECT id, name, description, contact_email, contact_phone, address, is_active, created_at, updated_at
		FROM vendors
		WHERE name ILIKE $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	var vendors []*models.Vendor
	err := r.db.GetDB().SelectContext(ctx, &vendors, sqlQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search vendors: %w", err)
	}

	return vendors, nil
}
