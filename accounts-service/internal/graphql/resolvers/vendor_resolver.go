package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/models"
)

// Vendor represents a GraphQL vendor resolver
type Vendor struct {
	vendor   *models.Vendor
	resolver *Resolver
}

// ID resolves the ID field of a vendor
func (r *Vendor) ID() string {
	return r.vendor.ID.String()
}

// Name resolves the name field of a vendor
func (r *Vendor) Name() string {
	return r.vendor.Name
}

// Description resolves the description field of a vendor
func (r *Vendor) Description() *string {
	if r.vendor.Description == "" {
		return nil
	}
	return &r.vendor.Description
}

// ContactEmail resolves the contactEmail field of a vendor
func (r *Vendor) ContactEmail() *string {
	if r.vendor.ContactEmail == "" {
		return nil
	}
	return &r.vendor.ContactEmail
}

// ContactPhone resolves the contactPhone field of a vendor
func (r *Vendor) ContactPhone() *string {
	if r.vendor.ContactPhone == "" {
		return nil
	}
	return &r.vendor.ContactPhone
}

// Address resolves the address field of a vendor
func (r *Vendor) Address() *string {
	if r.vendor.Address == "" {
		return nil
	}
	return &r.vendor.Address
}

// IsActive resolves the isActive field of a vendor
func (r *Vendor) IsActive() bool {
	return r.vendor.IsActive
}

// CreatedAt resolves the createdAt field of a vendor
func (r *Vendor) CreatedAt() string {
	return r.vendor.CreatedAt.Format(time.RFC3339)
}

// UpdatedAt resolves the updatedAt field of a vendor
func (r *Vendor) UpdatedAt() string {
	return r.vendor.UpdatedAt.Format(time.RFC3339)
}

// Products resolves the products field of a vendor
func (r *Vendor) Products(ctx context.Context) ([]*Product, error) {
	products, err := r.resolver.productService.ListByVendor(ctx, r.vendor.ID, 0, 100)
	if err != nil {
		return nil, err
	}

	var productResolvers []*Product
	for _, product := range products {
		productResolvers = append(productResolvers, &Product{
			product:  product,
			resolver: r.resolver,
		})
	}

	return productResolvers, nil
}

// Vendor resolves the vendor query
func (r *Resolver) Vendor(ctx context.Context, args struct{ ID string }) (*Vendor, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	vendor, err := r.vendorService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Vendor{
		vendor:   vendor,
		resolver: r,
	}, nil
}

// Vendors resolves the vendors query
func (r *Resolver) Vendors(ctx context.Context, args struct {
	Offset *int32
	Limit  *int32
}) ([]*Vendor, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	vendors, err := r.vendorService.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var vendorResolvers []*Vendor
	for _, vendor := range vendors {
		vendorResolvers = append(vendorResolvers, &Vendor{
			vendor:   vendor,
			resolver: r,
		})
	}

	return vendorResolvers, nil
}

// VendorsCount resolves the vendorsCount query
func (r *Resolver) VendorsCount(ctx context.Context) (int32, error) {
	count, err := r.vendorService.Count(ctx)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

// SearchVendors resolves the searchVendors query
func (r *Resolver) SearchVendors(ctx context.Context, args struct {
	Query  string
	Offset *int32
	Limit  *int32
}) ([]*Vendor, error) {
	offset := 0
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	limit := 10
	if args.Limit != nil {
		limit = int(*args.Limit)
	}

	vendors, err := r.vendorService.Search(ctx, args.Query, offset, limit)
	if err != nil {
		return nil, err
	}

	var vendorResolvers []*Vendor
	for _, vendor := range vendors {
		vendorResolvers = append(vendorResolvers, &Vendor{
			vendor:   vendor,
			resolver: r,
		})
	}

	return vendorResolvers, nil
}

// CreateVendor resolves the createVendor mutation
func (r *Resolver) CreateVendor(ctx context.Context, args struct {
	Input struct {
		Name         string
		Description  *string
		ContactEmail *string
		ContactPhone *string
		Address      *string
		IsActive     *bool
	}
}) (*Vendor, error) {
	input := models.VendorInput{
		Name:         args.Input.Name,
		Description:  args.Input.Description,
		ContactEmail: args.Input.ContactEmail,
		ContactPhone: args.Input.ContactPhone,
		Address:      args.Input.Address,
		IsActive:     args.Input.IsActive,
	}

	vendor, err := r.vendorService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return &Vendor{
		vendor:   vendor,
		resolver: r,
	}, nil
}

// UpdateVendor resolves the updateVendor mutation
func (r *Resolver) UpdateVendor(ctx context.Context, args struct {
	ID    string
	Input struct {
		Name         *string
		Description  *string
		ContactEmail *string
		ContactPhone *string
		Address      *string
		IsActive     *bool
	}
}) (*Vendor, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	input := models.VendorInput{
		Description:  args.Input.Description,
		ContactEmail: args.Input.ContactEmail,
		ContactPhone: args.Input.ContactPhone,
		Address:      args.Input.Address,
		IsActive:     args.Input.IsActive,
	}

	if args.Input.Name != nil {
		input.Name = *args.Input.Name
	}

	vendor, err := r.vendorService.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return &Vendor{
		vendor:   vendor,
		resolver: r,
	}, nil
}

// DeleteVendor resolves the deleteVendor mutation
func (r *Resolver) DeleteVendor(ctx context.Context, args struct{ ID string }) (bool, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return false, fmt.Errorf("invalid ID: %w", err)
	}

	err = r.vendorService.Delete(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}
