package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/repository"
)

// ProductService handles business logic for products
type ProductService struct {
	productRepo repository.ProductRepository
	vendorRepo  repository.VendorRepository
}

// NewProductService creates a new product service
func NewProductService(productRepo repository.ProductRepository, vendorRepo repository.VendorRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		vendorRepo:  vendorRepo,
	}
}

// Create creates a new product
func (s *ProductService) Create(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	// Check if vendor exists
	vendor, err := s.vendorRepo.GetByID(ctx, input.VendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}

	// Create the product
	product := models.NewProduct(input)
	product.Vendor = vendor

	// Save the product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetByID retrieves a product by ID
func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Load vendor
	vendor, err := s.vendorRepo.GetByID(ctx, product.VendorID)
	if err == nil {
		product.Vendor = vendor
	}

	return product, nil
}

// List retrieves all products with optional pagination
func (s *ProductService) List(ctx context.Context, offset, limit int) ([]*models.Product, error) {
	products, err := s.productRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Load vendors for each product
	for _, product := range products {
		vendor, err := s.vendorRepo.GetByID(ctx, product.VendorID)
		if err == nil {
			product.Vendor = vendor
		}
	}

	return products, nil
}

// ListByVendor retrieves all products for a specific vendor
func (s *ProductService) ListByVendor(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]*models.Product, error) {
	// Check if vendor exists
	vendor, err := s.vendorRepo.GetByID(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}

	products, err := s.productRepo.ListByVendor(ctx, vendorID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by vendor: %w", err)
	}

	// Set vendor for each product
	for _, product := range products {
		product.Vendor = vendor
	}

	return products, nil
}

// Update updates an existing product
func (s *ProductService) Update(ctx context.Context, id uuid.UUID, input models.ProductInput) (*models.Product, error) {
	// Get the existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update the product fields
	if input.Name != "" {
		product.Name = input.Name
	}

	if input.Description != nil {
		product.Description = *input.Description
	}

	if input.Price > 0 {
		product.Price = input.Price
	}

	if input.IsAvailable != nil {
		product.IsAvailable = *input.IsAvailable
	}

	// Save the updated product
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Load vendor
	vendor, err := s.vendorRepo.GetByID(ctx, product.VendorID)
	if err == nil {
		product.Vendor = vendor
	}

	return product, nil
}

// Delete deletes a product by ID
func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// Count returns the total number of products
func (s *ProductService) Count(ctx context.Context) (int, error) {
	count, err := s.productRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

// CountByVendor returns the total number of products for a specific vendor
func (s *ProductService) CountByVendor(ctx context.Context, vendorID uuid.UUID) (int, error) {
	// Check if vendor exists
	_, err := s.vendorRepo.GetByID(ctx, vendorID)
	if err != nil {
		return 0, fmt.Errorf("vendor not found: %w", err)
	}

	count, err := s.productRepo.CountByVendor(ctx, vendorID)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by vendor: %w", err)
	}
	return count, nil
}

// Search searches for products by name
func (s *ProductService) Search(ctx context.Context, query string, offset, limit int) ([]*models.Product, error) {
	products, err := s.productRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Load vendors for each product
	for _, product := range products {
		vendor, err := s.vendorRepo.GetByID(ctx, product.VendorID)
		if err == nil {
			product.Vendor = vendor
		}
	}

	return products, nil
}
