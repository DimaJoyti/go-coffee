package server

import (
	"github.com/yourusername/coffee-order-system/accounts-service/internal/service"
)

// Resolver is the root resolver for GraphQL queries and mutations
type Resolver struct {
	AccountService *service.AccountService
	VendorService  *service.VendorService
	ProductService *service.ProductService
	OrderService   *service.OrderService
}

// NewResolver creates a new resolver with the given services
func NewResolver(
	accountService *service.AccountService,
	vendorService *service.VendorService,
	productService *service.ProductService,
	orderService *service.OrderService,
) *Resolver {
	return &Resolver{
		AccountService: accountService,
		VendorService:  vendorService,
		ProductService: productService,
		OrderService:   orderService,
	}
}
