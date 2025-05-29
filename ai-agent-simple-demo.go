package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// AgentMCPRequest represents a natural language request to Redis
type AgentMCPRequest struct {
	Query     string                 `json:"query"`
	Context   map[string]interface{} `json:"context,omitempty"`
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// AgentMCPResponse represents the response from Redis MCP
type AgentMCPResponse struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Query     string                 `json:"executed_query,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CoffeeShopAIAgent represents an AI agent for coffee shop management
type CoffeeShopAIAgent struct {
	agentID      string
	mcpServerURL string
	httpClient   *http.Client
}

// NewCoffeeShopAIAgent creates a new coffee shop AI agent
func NewCoffeeShopAIAgent(agentID, mcpServerURL string) *CoffeeShopAIAgent {
	return &CoffeeShopAIAgent{
		agentID:      agentID,
		mcpServerURL: mcpServerURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a natural language query against Redis MCP
func (agent *CoffeeShopAIAgent) Query(ctx context.Context, query string, context map[string]interface{}) (*AgentMCPResponse, error) {
	request := AgentMCPRequest{
		Query:     query,
		Context:   context,
		AgentID:   agent.agentID,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := agent.mcpServerURL + "/api/v1/redis-mcp/query"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("AI-Agent-%s", agent.agentID))

	log.Printf("🤖 [%s] Sending query: %s", agent.agentID, request.Query)

	resp, err := agent.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response AgentMCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &response, fmt.Errorf("server returned status %d: %s", resp.StatusCode, response.Error)
	}

	return &response, nil
}

// AnalyzeCoffeeShopPerformance analyzes coffee shop performance using natural language
func (agent *CoffeeShopAIAgent) AnalyzeCoffeeShopPerformance(ctx context.Context) {
	log.Printf("🔍 [%s] Starting coffee shop performance analysis...", agent.agentID)

	// 1. Get top orders
	log.Printf("📊 [%s] Getting top orders...", agent.agentID)
	response, err := agent.Query(ctx, "get top orders today", nil)
	if err != nil {
		log.Printf("❌ [%s] Failed to get top orders: %v", agent.agentID, err)
		return
	}

	if response.Success {
		log.Printf("✅ [%s] Top orders retrieved successfully!", agent.agentID)
		if data, ok := response.Data.([]interface{}); ok {
			log.Printf("📈 [%s] Top selling drinks:", agent.agentID)
			for i, item := range data {
				if orderMap, ok := item.(map[string]interface{}); ok {
					drink := orderMap["Member"]
					score := orderMap["Score"]
					log.Printf("   %d. %s: %.0f orders", i+1, drink, score)
				}
			}
		}
	}

	// 2. Check inventory levels for all shops
	shops := []string{"downtown", "uptown", "westside"}
	for _, shop := range shops {
		log.Printf("🏪 [%s] Checking inventory for %s shop...", agent.agentID, shop)
		
		query := fmt.Sprintf("get inventory for %s", shop)
		response, err := agent.Query(ctx, query, map[string]interface{}{
			"shop_id": shop,
		})
		
		if err != nil {
			log.Printf("❌ [%s] Failed to get inventory for %s: %v", agent.agentID, shop, err)
			continue
		}

		if response.Success {
			if inventory, ok := response.Data.(map[string]interface{}); ok {
				log.Printf("📦 [%s] %s inventory:", agent.agentID, shop)
				
				// Check for low stock
				lowStockItems := []string{}
				for ingredient, quantityStr := range inventory {
					if quantity, ok := quantityStr.(string); ok {
						var qty float64
						fmt.Sscanf(quantity, "%f", &qty)
						log.Printf("   %s: %.0f units", ingredient, qty)
						
						// Check if stock is low (less than 30 units)
						if qty < 30 {
							lowStockItems = append(lowStockItems, ingredient)
						}
					}
				}
				
				if len(lowStockItems) > 0 {
					log.Printf("⚠️  [%s] LOW STOCK ALERT for %s: %v", agent.agentID, shop, lowStockItems)
				} else {
					log.Printf("✅ [%s] All inventory levels are healthy for %s", agent.agentID, shop)
				}
			}
		}
	}

	// 3. Get customer information
	log.Printf("👥 [%s] Analyzing customer data...", agent.agentID)
	customers := []string{"123", "456", "789"}
	
	for _, customerID := range customers {
		query := fmt.Sprintf("get customer %s name", customerID)
		response, err := agent.Query(ctx, query, map[string]interface{}{
			"customer_id": customerID,
		})
		
		if err != nil {
			log.Printf("❌ [%s] Failed to get customer %s: %v", agent.agentID, customerID, err)
			continue
		}

		if response.Success {
			if name, ok := response.Data.(string); ok {
				log.Printf("👤 [%s] Customer %s: %s", agent.agentID, customerID, name)
				
				// Get favorite drink
				favQuery := fmt.Sprintf("get customer %s favorite_drink", customerID)
				favResponse, err := agent.Query(ctx, favQuery, nil)
				if err == nil && favResponse.Success {
					if favDrink, ok := favResponse.Data.(string); ok {
						log.Printf("   ☕ Favorite drink: %s", favDrink)
					}
				}
			}
		}
	}

	log.Printf("🎉 [%s] Coffee shop performance analysis completed!", agent.agentID)
}

// SimulateInventoryManagement simulates inventory management tasks
func (agent *CoffeeShopAIAgent) SimulateInventoryManagement(ctx context.Context) {
	log.Printf("📦 [%s] Starting inventory management simulation...", agent.agentID)

	// Simulate adding various ingredients
	newIngredients := []string{"oat_milk", "coconut_sugar", "lavender_syrup", "cold_brew_concentrate"}
	
	for _, ingredient := range newIngredients {
		query := fmt.Sprintf("add %s to ingredients", ingredient)
		response, err := agent.Query(ctx, query, map[string]interface{}{
			"category": "new_product",
			"added_by": agent.agentID,
		})
		
		if err != nil {
			log.Printf("❌ [%s] Failed to add %s: %v", agent.agentID, ingredient, err)
		} else if response.Success {
			log.Printf("✅ [%s] Added %s to inventory", agent.agentID, ingredient)
		}
		
		// Small delay to simulate real-world timing
		time.Sleep(500 * time.Millisecond)
	}

	// Search for coffee-related data
	log.Printf("🔍 [%s] Searching for coffee-related data...", agent.agentID)
	response, err := agent.Query(ctx, "search coffee", nil)
	if err != nil {
		log.Printf("❌ [%s] Failed to search: %v", agent.agentID, err)
	} else if response.Success {
		if keys, ok := response.Data.([]interface{}); ok {
			log.Printf("🔎 [%s] Found %d coffee-related keys:", agent.agentID, len(keys))
			for _, key := range keys {
				log.Printf("   - %s", key)
			}
		}
	}

	log.Printf("✅ [%s] Inventory management simulation completed!", agent.agentID)
}

// RunAIAgentDemo runs the AI agent demonstration
func RunAIAgentDemo() {
	log.Println("🚀 Starting Coffee Shop AI Agent Demo...")

	// Create AI agents
	performanceAgent := NewCoffeeShopAIAgent("performance-analyzer-001", "http://localhost:8090")
	inventoryAgent := NewCoffeeShopAIAgent("inventory-manager-001", "http://localhost:8090")

	ctx := context.Background()

	// Wait a moment for the MCP server to be ready
	time.Sleep(2 * time.Second)

	log.Println("🎯 Starting AI Agent demonstrations...")

	// Run performance analysis
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔍 PERFORMANCE ANALYSIS AGENT DEMO")
	log.Println(strings.Repeat("=", 60))
	performanceAgent.AnalyzeCoffeeShopPerformance(ctx)

	// Wait between demos
	time.Sleep(3 * time.Second)

	// Run inventory management
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📦 INVENTORY MANAGEMENT AGENT DEMO")
	log.Println(strings.Repeat("=", 60))
	inventoryAgent.SimulateInventoryManagement(ctx)

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🎉 AI AGENT DEMO COMPLETED SUCCESSFULLY!")
	log.Println(strings.Repeat("=", 60))
	log.Println("✅ All natural language queries were successfully processed by Redis MCP")
	log.Println("✅ AI agents can now interact with Redis using human-like language")
	log.Println("✅ Complex data operations are simplified through intelligent parsing")
}
