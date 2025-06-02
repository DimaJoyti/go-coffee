package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
)

// VendorService handles business logic for vendors
type VendorService struct {
	vendorRepo repository.VendorRepository
}

// NewVendorService creates a new vendor service
func NewVendorService(vendorRepo repository.VendorRepository) *VendorService {
	return &VendorService{
		vendorRepo: vendorRepo,
	}
}

// Create creates a new vendor
func (s *VendorService) Create(ctx context.Context, input models.VendorInput) (*models.Vendor, error) {
	// Create the vendor
	vendor := models.NewVendor(input)

	// Save the vendor
	if err := s.vendorRepo.Create(ctx, vendor); err != nil {
		return nil, fmt.Errorf("failed to create vendor: %w", err)
	}

	return vendor, nil
}

// GetByID retrieves a vendor by ID
func (s *VendorService) GetByID(ctx context.Context, id uuid.UUID) (*models.Vendor, error) {
	vendor, err := s.vendorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}
	return vendor, nil
}

// List retrieves all vendors with optional pagination
func (s *VendorService) List(ctx context.Context, offset, limit int) ([]*models.Vendor, error) {
	vendors, err := s.vendorRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors: %w", err)
	}
	return vendors, nil
}

// Update updates an existing vendor
func (s *VendorService) Update(ctx context.Context, id uuid.UUID, input models.VendorInput) (*models.Vendor, error) {
	// Get the existing vendor
	vendor, err := s.vendorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}

	// Update the vendor fields
	if input.Name != "" {
		vendor.Name = input.Name
	}

	if input.Description != nil {
		vendor.Description = *input.Description
	}

	if input.ContactEmail != nil {
		vendor.ContactEmail = *input.ContactEmail
	}

	if input.ContactPhone != nil {
		vendor.ContactPhone = *input.ContactPhone
	}

	if input.Address != nil {
		vendor.Address = *input.Address
	}

	if input.IsActive != nil {
		vendor.IsActive = *input.IsActive
	}

	// Save the updated vendor
	if err := s.vendorRepo.Update(ctx, vendor); err != nil {
		return nil, fmt.Errorf("failed to update vendor: %w", err)
	}

	return vendor, nil
}

// Delete deletes a vendor by ID
func (s *VendorService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.vendorRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete vendor: %w", err)
	}
	return nil
}

// Count returns the total number of vendors
func (s *VendorService) Count(ctx context.Context) (int, error) {
	count, err := s.vendorRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count vendors: %w", err)
	}
	return count, nil
}

// Search searches for vendors by name
func (s *VendorService) Search(ctx context.Context, query string, offset, limit int) ([]*models.Vendor, error) {
	vendors, err := s.vendorRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search vendors: %w", err)
	}
	return vendors, nil
}
