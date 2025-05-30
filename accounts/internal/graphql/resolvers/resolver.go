package resolvers

import (
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/service"
)

// Resolver is the root resolver for GraphQL queries and mutations
type Resolver struct {
	accountService *service.AccountService
	vendorService  *service.VendorService
	productService *service.ProductService
	orderService   *service.OrderService
}

// NewResolver creates a new resolver
func NewResolver(
	accountService *service.AccountService,
	vendorService *service.VendorService,
	productService *service.ProductService,
	orderService *service.OrderService,
) *Resolver {
	return &Resolver{
		accountService: accountService,
		vendorService:  vendorService,
		productService: productService,
		orderService:   orderService,
	}
}

// NewResolverFromRepositories creates a new resolver from repositories
func NewResolverFromRepositories(
	accountRepo repository.AccountRepository,
	vendorRepo repository.VendorRepository,
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
) *Resolver {
	accountService := service.NewAccountService(accountRepo)
	vendorService := service.NewVendorService(vendorRepo)
	productService := service.NewProductService(productRepo, vendorRepo)
	orderService := service.NewOrderService(orderRepo, orderItemRepo, accountRepo, productRepo)

	return &Resolver{
		accountService: accountService,
		vendorService:  vendorService,
		productService: productService,
		orderService:   orderService,
	}
}
