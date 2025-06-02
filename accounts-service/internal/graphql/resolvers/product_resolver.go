package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/models"
)

// Product represents a GraphQL product resolver
type Product struct {
	product  *models.Product
	resolver *Resolver
}

// ID resolves the ID field of a product
func (r *Product) ID() string {
	return r.product.ID.String()
}

// VendorID resolves the vendorId field of a product
func (r *Product) VendorID() string {
	return r.product.VendorID.String()
}

// Name resolves the name field of a product
func (r *Product) Name() string {
	return r.product.Name
}

// Description resolves the description field of a product
func (r *Product) Description() *string {
	if r.product.Description == "" {
		return nil
	}
	return &r.product.Description
}

// Price resolves the price field of a product
func (r *Product) Price() float64 {
	return r.product.Price
}

// IsAvailable resolves the isAvailable field of a product
func (r *Product) IsAvailable() bool {
	return r.product.IsAvailable
}

// CreatedAt resolves the createdAt field of a product
func (r *Product) CreatedAt() string {
	return r.product.CreatedAt.Format(time.RFC3339)
}

// UpdatedAt resolves the updatedAt field of a product
func (r *Product) UpdatedAt() string {
	return r.product.UpdatedAt.Format(time.RFC3339)
}

// Vendor resolves the vendor field of a product
func (r *Product) Vendor(ctx context.Context) (*Vendor, error) {
	if r.product.Vendor != nil {
		return &Vendor{
			vendor:   r.product.Vendor,
			resolver: r.resolver,
		}, nil
	}

	vendor, err := r.resolver.vendorService.GetByID(ctx, r.product.VendorID)
	if err != nil {
		return nil, err
	}

	return &Vendor{
		vendor:   vendor,
		resolver: r.resolver,
	}, nil
}

// Product resolves the product query
func (r *Resolver) Product(ctx context.Context, args struct{ ID string }) (*Product, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	product, err := r.productService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Product{
		product:  product,
		resolver: r,
	}, nil
}

// Products resolves the products query
func (r *Resolver) Products(ctx context.Context, args struct {
	Offset *int32
	Limit  *int32
}) ([]*Product, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	products, err := r.productService.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var productResolvers []*Product
	for _, product := range products {
		productResolvers = append(productResolvers, &Product{
			product:  product,
			resolver: r,
		})
	}

	return productResolvers, nil
}

// ProductsByVendor resolves the productsByVendor query
func (r *Resolver) ProductsByVendor(ctx context.Context, args struct {
	VendorID string
	Offset   *int32
	Limit    *int32
}) ([]*Product, error) {
	vendorID, err := uuid.Parse(args.VendorID)
	if err != nil {
		return nil, fmt.Errorf("invalid vendor ID: %w", err)
	}

	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	products, err := r.productService.ListByVendor(ctx, vendorID, offset, limit)
	if err != nil {
		return nil, err
	}

	var productResolvers []*Product
	for _, product := range products {
		productResolvers = append(productResolvers, &Product{
			product:  product,
			resolver: r,
		})
	}

	return productResolvers, nil
}

// ProductsCount resolves the productsCount query
func (r *Resolver) ProductsCount(ctx context.Context) (int32, error) {
	count, err := r.productService.Count(ctx)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

// SearchProducts resolves the searchProducts query
func (r *Resolver) SearchProducts(ctx context.Context, args struct {
	Query  string
	Offset *int32
	Limit  *int32
}) ([]*Product, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	products, err := r.productService.Search(ctx, args.Query, offset, limit)
	if err != nil {
		return nil, err
	}

	var productResolvers []*Product
	for _, product := range products {
		productResolvers = append(productResolvers, &Product{
			product:  product,
			resolver: r,
		})
	}

	return productResolvers, nil
}

// CreateProduct resolves the createProduct mutation
func (r *Resolver) CreateProduct(ctx context.Context, args struct {
	Input struct {
		VendorID    string
		Name        string
		Description *string
		Price       float64
		IsAvailable *bool
	}
}) (*Product, error) {
	vendorID, err := uuid.Parse(args.Input.VendorID)
	if err != nil {
		return nil, fmt.Errorf("invalid vendor ID: %w", err)
	}

	input := models.ProductInput{
		VendorID:    vendorID,
		Name:        args.Input.Name,
		Description: args.Input.Description,
		Price:       args.Input.Price,
		IsAvailable: args.Input.IsAvailable,
	}

	product, err := r.productService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return &Product{
		product:  product,
		resolver: r,
	}, nil
}

// UpdateProduct resolves the updateProduct mutation
func (r *Resolver) UpdateProduct(ctx context.Context, args struct {
	ID    string
	Input struct {
		Name        *string
		Description *string
		Price       *float64
		IsAvailable *bool
	}
}) (*Product, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// Get the existing product
	product, err := r.productService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	input := models.ProductInput{
		VendorID:    product.VendorID,
		Description: args.Input.Description,
		IsAvailable: args.Input.IsAvailable,
	}

	if args.Input.Name != nil {
		input.Name = *args.Input.Name
	} else {
		input.Name = product.Name
	}

	if args.Input.Price != nil {
		input.Price = *args.Input.Price
	} else {
		input.Price = product.Price
	}

	updatedProduct, err := r.productService.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return &Product{
		product:  updatedProduct,
		resolver: r,
	}, nil
}

// DeleteProduct resolves the deleteProduct mutation
func (r *Resolver) DeleteProduct(ctx context.Context, args struct{ ID string }) (bool, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return false, fmt.Errorf("invalid ID: %w", err)
	}

	err = r.productService.Delete(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}
