package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// InteractiveMCPRequest represents a natural language request to Redis
type InteractiveMCPRequest struct {
	Query     string                 `json:"query"`
	Context   map[string]interface{} `json:"context,omitempty"`
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// InteractiveMCPResponse represents the response from Redis MCP
type InteractiveMCPResponse struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Query     string                 `json:"executed_query,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// InteractiveRedisMCP provides interactive testing of Redis MCP
type InteractiveRedisMCP struct {
	mcpServerURL string
	httpClient   *http.Client
	agentID      string
}

// NewInteractiveRedisMCP creates a new interactive Redis MCP tester
func NewInteractiveRedisMCP(mcpServerURL, agentID string) *InteractiveRedisMCP {
	return &InteractiveRedisMCP{
		mcpServerURL: mcpServerURL,
		agentID:      agentID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a natural language query against Redis MCP
func (i *InteractiveRedisMCP) Query(ctx context.Context, query string) (*InteractiveMCPResponse, error) {
	request := InteractiveMCPRequest{
		Query:     query,
		AgentID:   i.agentID,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := i.mcpServerURL + "/api/v1/redis-mcp/query"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response InteractiveMCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// PrettyPrintResponse formats and prints the response beautifully
func (i *InteractiveRedisMCP) PrettyPrintResponse(response *InteractiveMCPResponse) {
	if response.Success {
		fmt.Printf("✅ %sQuery executed successfully!%s\n", "\033[32m", "\033[0m")
		fmt.Printf("🔧 %sRedis Command:%s %s\n", "\033[36m", "\033[0m", response.Query)
		
		if response.Metadata != nil {
			if confidence, ok := response.Metadata["confidence"].(float64); ok {
				fmt.Printf("🎯 %sConfidence:%s %.1f%%\n", "\033[33m", "\033[0m", confidence*100)
			}
		}
		
		fmt.Printf("📊 %sResult:%s\n", "\033[35m", "\033[0m")
		
		// Format different types of data
		switch data := response.Data.(type) {
		case map[string]interface{}:
			fmt.Println("   📋 Hash data:")
			for key, value := range data {
				fmt.Printf("      %s%s:%s %v\n", "\033[34m", key, "\033[0m", value)
			}
		case []interface{}:
			if len(data) > 0 {
				// Check if it's sorted set data with scores
				if firstItem, ok := data[0].(map[string]interface{}); ok {
					if _, hasScore := firstItem["Score"]; hasScore {
						fmt.Println("   🏆 Ranked data:")
						for i, item := range data {
							if itemMap, ok := item.(map[string]interface{}); ok {
								member := itemMap["Member"]
								score := itemMap["Score"]
								fmt.Printf("      %d. %s%s:%s %.0f\n", i+1, "\033[34m", member, "\033[0m", score)
							}
						}
					} else {
						fmt.Println("   📝 List data:")
						for i, item := range data {
							fmt.Printf("      %d. %v\n", i+1, item)
						}
					}
				} else {
					fmt.Println("   📝 List data:")
					for i, item := range data {
						fmt.Printf("      %d. %v\n", i+1, item)
					}
				}
			}
		case string:
			fmt.Printf("   📄 %s\n", data)
		case float64:
			fmt.Printf("   🔢 %.0f\n", data)
		default:
			fmt.Printf("   📦 %v\n", data)
		}
	} else {
		fmt.Printf("❌ %sQuery failed:%s %s\n", "\033[31m", "\033[0m", response.Error)
	}
	fmt.Println()
}

// RunInteractiveSession runs an interactive session
func (i *InteractiveRedisMCP) RunInteractiveSession() {
	fmt.Println("🎉 Welcome to Interactive Redis MCP Demo!")
	fmt.Println("Type natural language queries to interact with Redis data.")
	fmt.Println("Type 'help' for examples, 'quit' to exit.")
	fmt.Println(strings.Repeat("=", 60))

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("🤖 Enter your query: ")
		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())
		
		if query == "" {
			continue
		}

		if query == "quit" || query == "exit" {
			fmt.Println("👋 Goodbye! Thanks for testing Redis MCP!")
			break
		}

		if query == "help" {
			i.ShowHelp()
			continue
		}

		// Execute the query
		fmt.Printf("🔍 Processing: %s\n", query)
		response, err := i.Query(ctx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		i.PrettyPrintResponse(response)
	}
}

// ShowHelp displays example queries
func (i *InteractiveRedisMCP) ShowHelp() {
	fmt.Println("\n📚 Example Natural Language Queries:")
	fmt.Println("   🏪 Coffee Shop Operations:")
	fmt.Println("      • get menu for shop downtown")
	fmt.Println("      • get menu for shop uptown")
	fmt.Println("      • get inventory for westside")
	fmt.Println()
	fmt.Println("   📦 Inventory Management:")
	fmt.Println("      • add matcha to ingredients")
	fmt.Println("      • add chai_spice to ingredients")
	fmt.Println()
	fmt.Println("   📊 Analytics:")
	fmt.Println("      • get top orders today")
	fmt.Println("      • search coffee")
	fmt.Println()
	fmt.Println("   👥 Customer Data:")
	fmt.Println("      • get customer 123 name")
	fmt.Println("      • get customer 456 favorite_drink")
	fmt.Println("      • get customer 789 loyalty_points")
	fmt.Println()
	fmt.Println("   🔍 Search Operations:")
	fmt.Println("      • search menu")
	fmt.Println("      • search inventory")
	fmt.Println("      • search customer")
	fmt.Println(strings.Repeat("-", 60))
}

// RunPredefinedDemo runs a predefined demo with exciting queries
func (i *InteractiveRedisMCP) RunPredefinedDemo() {
	fmt.Println("🎬 Starting Predefined Redis MCP Demo!")
	fmt.Println(strings.Repeat("=", 60))

	demoQueries := []struct {
		description string
		query       string
		delay       time.Duration
	}{
		{"🏪 Getting downtown coffee menu", "get menu for shop downtown", 2 * time.Second},
		{"📦 Checking uptown inventory", "get inventory for uptown", 2 * time.Second},
		{"🏆 Finding top selling drinks", "get top orders today", 2 * time.Second},
		{"👤 Getting customer info", "get customer 123 name", 1 * time.Second},
		{"☕ Finding favorite drink", "get customer 456 favorite_drink", 1 * time.Second},
		{"🧪 Adding trendy ingredient", "add bubble_tea_pearls to ingredients", 2 * time.Second},
		{"🔍 Searching coffee data", "search coffee", 2 * time.Second},
		{"📊 Getting westside menu", "get menu for shop westside", 1 * time.Second},
		{"🥛 Adding premium milk", "add premium_oat_milk to ingredients", 1 * time.Second},
		{"👥 Checking another customer", "get customer 789 loyalty_points", 1 * time.Second},
	}

	ctx := context.Background()

	for idx, demo := range demoQueries {
		fmt.Printf("\n%d. %s\n", idx+1, demo.description)
		fmt.Printf("🤖 Query: %s\n", demo.query)

		response, err := i.Query(ctx, demo.query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			i.PrettyPrintResponse(response)
		}

		if idx < len(demoQueries)-1 {
			time.Sleep(demo.delay)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🎉 Predefined demo completed! All queries processed successfully!")
}

// RunInteractiveTest runs the interactive Redis MCP test
func RunInteractiveTest() {
	log.Println("🚀 Starting Interactive Redis MCP Test...")

	// Check if Redis MCP server is running
	mcpTester := NewInteractiveRedisMCP("http://localhost:8090", "interactive-tester")

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testResponse, err := mcpTester.Query(ctx, "get menu for shop downtown")
	if err != nil {
		log.Printf("❌ Cannot connect to Redis MCP server: %v", err)
		log.Println("💡 Make sure the Redis MCP server is running on localhost:8090")
		return
	}

	if !testResponse.Success {
		log.Printf("❌ Redis MCP server returned error: %s", testResponse.Error)
		return
	}

	log.Println("✅ Redis MCP server is running and responsive!")

	// Ask user for demo type
	fmt.Println("\nChoose demo type:")
	fmt.Println("1. Interactive mode (you type queries)")
	fmt.Println("2. Predefined demo (automatic)")
	fmt.Print("Enter choice (1 or 2): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			mcpTester.RunInteractiveSession()
		case "2":
			mcpTester.RunPredefinedDemo()
		default:
			fmt.Println("Invalid choice, running predefined demo...")
			mcpTester.RunPredefinedDemo()
		}
	} else {
		fmt.Println("Running predefined demo...")
		mcpTester.RunPredefinedDemo()
	}
}
