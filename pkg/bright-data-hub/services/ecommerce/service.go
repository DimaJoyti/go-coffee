package ecommerce

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/core"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Service handles all e-commerce data collection
type Service struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
	
	// Platform handlers
	amazon  *AmazonHandler
	booking *BookingHandler
	zillow  *ZillowHandler
}

// NewService creates a new e-commerce service
func NewService(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) (*Service, error) {
	service := &Service{
		client: client,
		config: cfg,
		logger: log,
	}
	
	// Initialize platform handlers
	if cfg.Ecommerce.Amazon.Enabled {
		service.amazon = NewAmazonHandler(client, cfg, log)
	}
	
	if cfg.Ecommerce.Booking.Enabled {
		service.booking = NewBookingHandler(client, cfg, log)
	}
	
	if cfg.Ecommerce.Zillow.Enabled {
		service.zillow = NewZillowHandler(client, cfg, log)
	}
	
	return service, nil
}

// Start starts the e-commerce service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting e-commerce service")
	return nil
}

// ExecuteFunction executes an e-commerce function
func (s *Service) ExecuteFunction(ctx context.Context, function string, params interface{}) (interface{}, error) {
	s.logger.Debug("Executing e-commerce function: %s", function)
	
	switch function {
	// Amazon functions
	case "web_data_amazon_product_Bright_Data":
		if s.amazon == nil {
			return nil, fmt.Errorf("Amazon handler not enabled")
		}
		return s.amazon.GetProduct(ctx, params)
		
	case "web_data_amazon_product_reviews_Bright_Data":
		if s.amazon == nil {
			return nil, fmt.Errorf("Amazon handler not enabled")
		}
		return s.amazon.GetProductReviews(ctx, params)
		
	// Booking functions
	case "web_data_booking_hotel_listings_Bright_Data":
		if s.booking == nil {
			return nil, fmt.Errorf("Booking handler not enabled")
		}
		return s.booking.GetHotelListings(ctx, params)
		
	// Zillow functions
	case "web_data_zillow_properties_listing_Bright_Data":
		if s.zillow == nil {
			return nil, fmt.Errorf("Zillow handler not enabled")
		}
		return s.zillow.GetPropertyListings(ctx, params)
		
	default:
		return nil, fmt.Errorf("unsupported e-commerce function: %s", function)
	}
}

// AmazonHandler handles Amazon-specific operations
type AmazonHandler struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
}

// NewAmazonHandler creates a new Amazon handler
func NewAmazonHandler(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) *AmazonHandler {
	return &AmazonHandler{
		client: client,
		config: cfg,
		logger: log,
	}
}

// GetProduct retrieves Amazon product data
func (h *AmazonHandler) GetProduct(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting Amazon product")
	
	response, err := h.client.CallMCP(ctx, "web_data_amazon_product_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Amazon product: %w", err)
	}
	
	return response.Result, nil
}

// GetProductReviews retrieves Amazon product reviews
func (h *AmazonHandler) GetProductReviews(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting Amazon product reviews")
	
	response, err := h.client.CallMCP(ctx, "web_data_amazon_product_reviews_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Amazon product reviews: %w", err)
	}
	
	return response.Result, nil
}

// BookingHandler handles Booking.com-specific operations
type BookingHandler struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
}

// NewBookingHandler creates a new Booking handler
func NewBookingHandler(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) *BookingHandler {
	return &BookingHandler{
		client: client,
		config: cfg,
		logger: log,
	}
}

// GetHotelListings retrieves Booking.com hotel listings
func (h *BookingHandler) GetHotelListings(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting Booking hotel listings")
	
	response, err := h.client.CallMCP(ctx, "web_data_booking_hotel_listings_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Booking hotel listings: %w", err)
	}
	
	return response.Result, nil
}

// ZillowHandler handles Zillow-specific operations
type ZillowHandler struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
}

// NewZillowHandler creates a new Zillow handler
func NewZillowHandler(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) *ZillowHandler {
	return &ZillowHandler{
		client: client,
		config: cfg,
		logger: log,
	}
}

// GetPropertyListings retrieves Zillow property listings
func (h *ZillowHandler) GetPropertyListings(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting Zillow property listings")
	
	response, err := h.client.CallMCP(ctx, "web_data_zillow_properties_listing_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Zillow property listings: %w", err)
	}
	
	return response.Result, nil
}
