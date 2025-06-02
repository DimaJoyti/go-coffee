package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// MCPClient handles communication with MCP servers
type MCPClient struct {
	client  *http.Client
	baseURL string
}

// MCPRequest represents a request to an MCP server
type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from an MCP server
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error from an MCP server
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient() *MCPClient {
	return &MCPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: os.Getenv("MCP_SERVER_URL"),
	}
}

// CallMCP makes a call to an MCP server
func (c *MCPClient) CallMCP(method string, params interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		Method: method,
		Params: params,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	return &mcpResp, nil
}

// BrightDataMCPService integrates with Bright Data through MCP
type BrightDataMCPService struct {
	mcpClient *MCPClient
}

// NewBrightDataMCPService creates a new Bright Data MCP service
func NewBrightDataMCPService() *BrightDataMCPService {
	return &BrightDataMCPService{
		mcpClient: NewMCPClient(),
	}
}

// ScrapeURL scrapes a URL using Bright Data MCP
func (s *BrightDataMCPService) ScrapeURL(url string) (*ScrapingResponse, error) {
	params := map[string]interface{}{
		"url": url,
	}

	resp, err := s.mcpClient.CallMCP("scrape_as_markdown_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}

// SearchEngine performs a search using Bright Data MCP
func (s *BrightDataMCPService) SearchEngine(query string, engine string) (*ScrapingResponse, error) {
	params := map[string]interface{}{
		"query":  query,
		"engine": engine, // google, bing, yandex
	}

	resp, err := s.mcpClient.CallMCP("search_engine_Bright_Data", params)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}

// GetSessionStats gets session statistics from Bright Data MCP
func (s *BrightDataMCPService) GetSessionStats() (*ScrapingResponse, error) {
	resp, err := s.mcpClient.CallMCP("session_stats_Bright_Data", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	return &ScrapingResponse{
		Success: true,
		Data:    resp.Result,
	}, nil
}
