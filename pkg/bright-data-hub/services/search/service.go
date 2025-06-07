package search

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/core"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Service handles all search engine operations
type Service struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
}

// NewService creates a new search service
func NewService(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) (*Service, error) {
	return &Service{
		client: client,
		config: cfg,
		logger: log,
	}, nil
}

// Start starts the search service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting search service")
	return nil
}

// ExecuteFunction executes a search function
func (s *Service) ExecuteFunction(ctx context.Context, function string, params interface{}) (interface{}, error) {
	s.logger.Debug("Executing search function: %s", function)
	
	switch function {
	case "search_engine_Bright_Data":
		return s.SearchEngine(ctx, params)
	case "scrape_as_markdown_Bright_Data":
		return s.ScrapeAsMarkdown(ctx, params)
	case "scrape_as_html_Bright_Data":
		return s.ScrapeAsHTML(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported search function: %s", function)
	}
}

// SearchEngine performs a search using specified engine
func (s *Service) SearchEngine(ctx context.Context, params interface{}) (interface{}, error) {
	s.logger.Debug("Performing search engine query")
	
	response, err := s.client.CallMCP(ctx, "search_engine_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	
	return response.Result, nil
}

// ScrapeAsMarkdown scrapes a URL and returns markdown content
func (s *Service) ScrapeAsMarkdown(ctx context.Context, params interface{}) (interface{}, error) {
	s.logger.Debug("Scraping URL as markdown")
	
	response, err := s.client.CallMCP(ctx, "scrape_as_markdown_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape as markdown: %w", err)
	}
	
	return response.Result, nil
}

// ScrapeAsHTML scrapes a URL and returns HTML content
func (s *Service) ScrapeAsHTML(ctx context.Context, params interface{}) (interface{}, error) {
	s.logger.Debug("Scraping URL as HTML")
	
	response, err := s.client.CallMCP(ctx, "scrape_as_html_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape as HTML: %w", err)
	}
	
	return response.Result, nil
}
