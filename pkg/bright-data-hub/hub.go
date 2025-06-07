package brightdatahub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/core"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/services/analytics"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/services/ecommerce"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/services/search"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/services/social"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// BrightDataHub is the main service that orchestrates all Bright Data operations
type BrightDataHub struct {
	config  *config.BrightDataHubConfig
	client  *core.MCPClient
	logger  *logger.Logger
	
	// Service modules
	socialService    *social.Service
	ecommerceService *ecommerce.Service
	searchService    *search.Service
	analyticsService *analytics.Service
	
	// State management
	isRunning bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// HubResponse represents a standardized response from the hub
type HubResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Metadata  *Metadata   `json:"metadata,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Metadata contains additional information about the response
type Metadata struct {
	Source       string        `json:"source"`
	CacheHit     bool          `json:"cache_hit"`
	ProcessTime  time.Duration `json:"process_time"`
	RequestID    string        `json:"request_id"`
	DataQuality  float64       `json:"data_quality,omitempty"`
	Confidence   float64       `json:"confidence,omitempty"`
}

// NewBrightDataHub creates a new Bright Data Hub instance
func NewBrightDataHub(cfg *config.BrightDataHubConfig, log *logger.Logger) (*BrightDataHub, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	
	// Create MCP client
	client, err := core.NewMCPClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	
	// Create context for the hub
	ctx, cancel := context.WithCancel(context.Background())
	
	hub := &BrightDataHub{
		config: cfg,
		client: client,
		logger: log,
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize service modules
	if err := hub.initializeServices(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	return hub, nil
}

// initializeServices initializes all service modules
func (h *BrightDataHub) initializeServices() error {
	var err error
	
	// Initialize social service
	if h.config.EnableSocial {
		h.socialService, err = social.NewService(h.client, h.config, h.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize social service: %w", err)
		}
	}
	
	// Initialize ecommerce service
	if h.config.EnableEcommerce {
		h.ecommerceService, err = ecommerce.NewService(h.client, h.config, h.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize ecommerce service: %w", err)
		}
	}
	
	// Initialize search service
	if h.config.EnableSearch {
		h.searchService, err = search.NewService(h.client, h.config, h.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize search service: %w", err)
		}
	}
	
	// Initialize analytics service
	if h.config.EnableAnalytics {
		h.analyticsService, err = analytics.NewService(h.client, h.config, h.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize analytics service: %w", err)
		}
	}
	
	return nil
}

// Start starts the Bright Data Hub
func (h *BrightDataHub) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.isRunning {
		return fmt.Errorf("hub is already running")
	}
	
	if !h.config.Enabled {
		h.logger.Info("Bright Data Hub is disabled")
		return nil
	}
	
	h.logger.Info("Starting Bright Data Hub")
	h.isRunning = true
	
	// Start service modules
	if h.socialService != nil {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			if err := h.socialService.Start(h.ctx); err != nil {
				h.logger.Error("Social service error: %v", err)
			}
		}()
	}
	
	if h.ecommerceService != nil {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			if err := h.ecommerceService.Start(h.ctx); err != nil {
				h.logger.Error("Ecommerce service error: %v", err)
			}
		}()
	}
	
	if h.searchService != nil {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			if err := h.searchService.Start(h.ctx); err != nil {
				h.logger.Error("Search service error: %v", err)
			}
		}()
	}
	
	if h.analyticsService != nil {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			if err := h.analyticsService.Start(h.ctx); err != nil {
				h.logger.Error("Analytics service error: %v", err)
			}
		}()
	}
	
	h.logger.Info("Bright Data Hub started successfully")
	return nil
}

// Stop gracefully stops the Bright Data Hub
func (h *BrightDataHub) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if !h.isRunning {
		return nil
	}
	
	h.logger.Info("Stopping Bright Data Hub")
	
	// Cancel context to signal all goroutines to stop
	h.cancel()
	
	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()
	
	// Wait with timeout
	select {
	case <-done:
		h.logger.Info("All services stopped gracefully")
	case <-time.After(30 * time.Second):
		h.logger.Warn("Timeout waiting for services to stop")
	}
	
	// Close MCP client
	if err := h.client.Close(); err != nil {
		h.logger.Error("Error closing MCP client: %v", err)
	}
	
	h.isRunning = false
	h.logger.Info("Bright Data Hub stopped")
	
	return nil
}

// ExecuteFunction executes any Bright Data MCP function
func (h *BrightDataHub) ExecuteFunction(ctx context.Context, function string, params interface{}) (*HubResponse, error) {
	startTime := time.Now()
	
	// Route to appropriate service
	var result interface{}
	var err error
	var source string
	
	switch {
	case h.isSocialFunction(function):
		if h.socialService == nil {
			return h.createErrorResponse("Social service not enabled", startTime), nil
		}
		result, err = h.socialService.ExecuteFunction(ctx, function, params)
		source = "social"
		
	case h.isEcommerceFunction(function):
		if h.ecommerceService == nil {
			return h.createErrorResponse("Ecommerce service not enabled", startTime), nil
		}
		result, err = h.ecommerceService.ExecuteFunction(ctx, function, params)
		source = "ecommerce"
		
	case h.isSearchFunction(function):
		if h.searchService == nil {
			return h.createErrorResponse("Search service not enabled", startTime), nil
		}
		result, err = h.searchService.ExecuteFunction(ctx, function, params)
		source = "search"
		
	default:
		// Direct MCP call for unsupported functions
		mcpResp, mcpErr := h.client.CallMCP(ctx, function, params)
		if mcpErr != nil {
			return h.createErrorResponse(mcpErr.Error(), startTime), nil
		}
		result = mcpResp.Result
		source = "direct"
	}
	
	if err != nil {
		return h.createErrorResponse(err.Error(), startTime), nil
	}
	
	// Create successful response
	response := &HubResponse{
		Success:   true,
		Data:      result,
		Timestamp: time.Now(),
		Metadata: &Metadata{
			Source:      source,
			ProcessTime: time.Since(startTime),
		},
	}
	
	return response, nil
}

// GetStatus returns the current status of the hub
func (h *BrightDataHub) GetStatus() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":   h.isRunning,
		"config":    h.config,
		"metrics":   h.client.GetMetrics(),
		"services": map[string]bool{
			"social":    h.socialService != nil,
			"ecommerce": h.ecommerceService != nil,
			"search":    h.searchService != nil,
			"analytics": h.analyticsService != nil,
		},
	}
	
	return status
}

// Helper methods
func (h *BrightDataHub) createErrorResponse(errorMsg string, startTime time.Time) *HubResponse {
	return &HubResponse{
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
		Metadata: &Metadata{
			ProcessTime: time.Since(startTime),
		},
	}
}

func (h *BrightDataHub) isSocialFunction(function string) bool {
	socialFunctions := []string{
		"web_data_instagram_profiles_Bright_Data",
		"web_data_instagram_posts_Bright_Data",
		"web_data_instagram_reels_Bright_Data",
		"web_data_instagram_comments_Bright_Data",
		"web_data_facebook_posts_Bright_Data",
		"web_data_facebook_marketplace_listings_Bright_Data",
		"web_data_facebook_company_reviews_Bright_Data",
		"web_data_x_posts_Bright_Data",
		"web_data_linkedin_person_profile_Bright_Data",
		"web_data_linkedin_company_profile_Bright_Data",
	}
	
	for _, sf := range socialFunctions {
		if sf == function {
			return true
		}
	}
	return false
}

func (h *BrightDataHub) isEcommerceFunction(function string) bool {
	ecommerceFunctions := []string{
		"web_data_amazon_product_Bright_Data",
		"web_data_amazon_product_reviews_Bright_Data",
		"web_data_booking_hotel_listings_Bright_Data",
		"web_data_zillow_properties_listing_Bright_Data",
	}
	
	for _, ef := range ecommerceFunctions {
		if ef == function {
			return true
		}
	}
	return false
}

func (h *BrightDataHub) isSearchFunction(function string) bool {
	searchFunctions := []string{
		"search_engine_Bright_Data",
		"scrape_as_markdown_Bright_Data",
		"scrape_as_html_Bright_Data",
	}
	
	for _, sf := range searchFunctions {
		if sf == function {
			return true
		}
	}
	return false
}
