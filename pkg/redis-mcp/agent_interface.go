package redismcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// AgentRedisInterface provides Redis MCP interface for AI agents
type AgentRedisInterface struct {
	mcpServerURL string
	httpClient   *http.Client
	logger       *logger.Logger
	agentID      string
}

// NewAgentRedisInterface creates a new agent Redis interface
func NewAgentRedisInterface(mcpServerURL, agentID string, logger *logger.Logger) *AgentRedisInterface {
	return &AgentRedisInterface{
		mcpServerURL: mcpServerURL,
		agentID:      agentID,
		logger:       logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a natural language query against Redis
func (ari *AgentRedisInterface) Query(ctx context.Context, query string, context map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		Query:     query,
		Context:   context,
		AgentID:   ari.agentID,
		Timestamp: time.Now(),
	}

	return ari.sendRequest(ctx, "/api/v1/redis-mcp/query", request)
}

// BatchQuery executes multiple queries in a single request
func (ari *AgentRedisInterface) BatchQuery(ctx context.Context, queries []string, context map[string]interface{}) ([]MCPResponse, error) {
	requests := make([]MCPRequest, len(queries))
	for i, query := range queries {
		requests[i] = MCPRequest{
			Query:     query,
			Context:   context,
			AgentID:   ari.agentID,
			Timestamp: time.Now(),
		}
	}

	response, err := ari.sendBatchRequest(ctx, "/api/v1/redis-mcp/batch", requests)
	if err != nil {
		return nil, err
	}

	var batchResponse struct {
		Responses []MCPResponse `json:"responses"`
		Total     int           `json:"total"`
		Success   int           `json:"success"`
		Failed    int           `json:"failed"`
	}

	if err := json.Unmarshal(response, &batchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	return batchResponse.Responses, nil
}

// Coffee Shop specific helper methods

// GetMenu retrieves the menu for a specific coffee shop
func (ari *AgentRedisInterface) GetMenu(ctx context.Context, shopID string) (map[string]string, error) {
	query := fmt.Sprintf("get menu for shop %s", shopID)
	response, err := ari.Query(ctx, query, map[string]interface{}{"shop_id": shopID})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("query failed: %s", response.Error)
	}

	menu, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	result := make(map[string]string)
	for k, v := range menu {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}

	return result, nil
}

// GetInventory retrieves inventory for a specific coffee shop
func (ari *AgentRedisInterface) GetInventory(ctx context.Context, shopID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("get inventory for %s", shopID)
	response, err := ari.Query(ctx, query, map[string]interface{}{"shop_id": shopID})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("query failed: %s", response.Error)
	}

	inventory, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return inventory, nil
}

// AddIngredient adds a new ingredient to the available ingredients set
func (ari *AgentRedisInterface) AddIngredient(ctx context.Context, ingredient string) error {
	query := fmt.Sprintf("add %s to ingredients", ingredient)
	response, err := ari.Query(ctx, query, map[string]interface{}{"ingredient": ingredient})
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("query failed: %s", response.Error)
	}

	return nil
}

// GetTopOrders retrieves top coffee orders for a specific period
func (ari *AgentRedisInterface) GetTopOrders(ctx context.Context, period string, limit int) ([]OrderInfo, error) {
	query := fmt.Sprintf("get top %d coffee orders for %s", limit, period)
	response, err := ari.Query(ctx, query, map[string]interface{}{
		"period": period,
		"limit":  limit,
	})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("query failed: %s", response.Error)
	}

	// Parse the response data
	orders := make([]OrderInfo, 0)
	if data, ok := response.Data.([]interface{}); ok {
		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				order := OrderInfo{
					Item:  data[i].(string),
					Score: data[i+1].(string),
				}
				orders = append(orders, order)
			}
		}
	}

	return orders, nil
}

// SetCustomerData sets customer profile data
func (ari *AgentRedisInterface) SetCustomerData(ctx context.Context, customerID, field, value string) error {
	query := fmt.Sprintf("set customer %s %s to %s", customerID, field, value)
	response, err := ari.Query(ctx, query, map[string]interface{}{
		"customer_id": customerID,
		"field":       field,
		"value":       value,
	})
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("query failed: %s", response.Error)
	}

	return nil
}

// GetCustomerData retrieves customer profile data
func (ari *AgentRedisInterface) GetCustomerData(ctx context.Context, customerID, field string) (string, error) {
	query := fmt.Sprintf("get customer %s %s", customerID, field)
	response, err := ari.Query(ctx, query, map[string]interface{}{
		"customer_id": customerID,
		"field":       field,
	})
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("query failed: %s", response.Error)
	}

	if value, ok := response.Data.(string); ok {
		return value, nil
	}

	return "", fmt.Errorf("unexpected response format")
}

// SearchProducts searches for products containing specific terms
func (ari *AgentRedisInterface) SearchProducts(ctx context.Context, searchTerm string) ([]string, error) {
	query := fmt.Sprintf("search products containing %s", searchTerm)
	response, err := ari.Query(ctx, query, map[string]interface{}{
		"search_term": searchTerm,
	})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("query failed: %s", response.Error)
	}

	if products, ok := response.Data.([]interface{}); ok {
		result := make([]string, len(products))
		for i, product := range products {
			result[i] = product.(string)
		}
		return result, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// GetAnalytics retrieves analytics data for a specific metric
func (ari *AgentRedisInterface) GetAnalytics(ctx context.Context, metric string) ([]AnalyticsData, error) {
	query := fmt.Sprintf("get analytics for %s", metric)
	response, err := ari.Query(ctx, query, map[string]interface{}{
		"metric": metric,
	})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("query failed: %s", response.Error)
	}

	// Parse analytics data
	analytics := make([]AnalyticsData, 0)
	if data, ok := response.Data.([]interface{}); ok {
		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				analytic := AnalyticsData{
					Key:   data[i].(string),
					Value: data[i+1].(string),
				}
				analytics = append(analytics, analytic)
			}
		}
	}

	return analytics, nil
}

// Helper data structures
type OrderInfo struct {
	Item  string `json:"item"`
	Score string `json:"score"`
}

type AnalyticsData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// sendRequest sends a single request to the MCP server
func (ari *AgentRedisInterface) sendRequest(ctx context.Context, endpoint string, request MCPRequest) (*MCPResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := ari.mcpServerURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("AI-Agent-%s", ari.agentID))

	ari.logger.Debug("Sending MCP request",
		zap.String("url", url),
		zap.String("agent_id", ari.agentID),
		zap.String("query", request.Query),
	)

	resp, err := ari.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response MCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &response, fmt.Errorf("server returned status %d: %s", resp.StatusCode, response.Error)
	}

	return &response, nil
}

// sendBatchRequest sends a batch request to the MCP server
func (ari *AgentRedisInterface) sendBatchRequest(ctx context.Context, endpoint string, requests []MCPRequest) ([]byte, error) {
	jsonData, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	url := ari.mcpServerURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("AI-Agent-%s", ari.agentID))

	ari.logger.Debug("Sending MCP batch request",
		zap.String("url", url),
		zap.String("agent_id", ari.agentID),
		zap.Int("query_count", len(requests)),
	)

	resp, err := ari.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send batch request: %w", err)
	}
	defer resp.Body.Close()

	var responseData bytes.Buffer
	if _, err := responseData.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return responseData.Bytes(), nil
}

// Health checks the health of the MCP server
func (ari *AgentRedisInterface) Health(ctx context.Context) error {
	url := ari.mcpServerURL + "/api/v1/redis-mcp/health"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := ari.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send health request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCP server unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
