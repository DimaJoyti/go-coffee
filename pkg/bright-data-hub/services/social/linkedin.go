package social

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/core"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// LinkedInHandler handles LinkedIn-specific operations
type LinkedInHandler struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
}

// NewLinkedInHandler creates a new LinkedIn handler
func NewLinkedInHandler(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) *LinkedInHandler {
	return &LinkedInHandler{
		client: client,
		config: cfg,
		logger: log,
	}
}

// GetPersonProfile retrieves LinkedIn person profile
func (h *LinkedInHandler) GetPersonProfile(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting LinkedIn person profile")
	
	response, err := h.client.CallMCP(ctx, "web_data_linkedin_person_profile_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get LinkedIn person profile: %w", err)
	}
	
	return response.Result, nil
}

// GetCompanyProfile retrieves LinkedIn company profile
func (h *LinkedInHandler) GetCompanyProfile(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Getting LinkedIn company profile")
	
	response, err := h.client.CallMCP(ctx, "web_data_linkedin_company_profile_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get LinkedIn company profile: %w", err)
	}
	
	return response.Result, nil
}
