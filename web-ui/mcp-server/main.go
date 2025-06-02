package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// loadEnv loads environment variables from .env file
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err()
}

type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type ScrapeResult struct {
	Content string `json:"content"`
	URL     string `json:"url"`
	Success bool   `json:"success"`
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv("../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "3001"
	}

	http.HandleFunc("/", handleMCPRequest)
	http.HandleFunc("/health", handleHealth)

	log.Printf("🚀 MCP Server starting on port %s", port)
	log.Printf("📊 Health: http://localhost:%s/health", port)
	log.Printf("🔗 MCP Endpoint: http://localhost:%s/", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start MCP server:", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"service":   "go-coffee-mcp-server",
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}
	json.NewEncoder(w).Encode(response)
}

func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, 400, "Invalid JSON request")
		return
	}

	log.Printf("📥 MCP Request: %s", req.Method)

	var result interface{}
	var err error

	switch req.Method {
	case "search_engine_Bright_Data":
		result, err = handleSearchEngine(req.Params)
	case "scrape_as_markdown_Bright_Data":
		result, err = handleScrapeMarkdown(req.Params)
	case "session_stats_Bright_Data":
		result, err = handleSessionStats(req.Params)
	case "web_data_amazon_product_Bright_Data":
		result, err = handleAmazonProduct(req.Params)
	default:
		sendError(w, 404, fmt.Sprintf("Method not found: %s", req.Method))
		return
	}

	if err != nil {
		sendError(w, 500, err.Error())
		return
	}

	response := MCPResponse{Result: result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSearchEngine(params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params for search_engine")
	}

	query, _ := paramsMap["query"].(string)
	engine, _ := paramsMap["engine"].(string)
	
	if engine == "" {
		engine = "google"
	}

	log.Printf("🔍 Search: %s on %s", query, engine)

	// Mock search results based on query
	results := []SearchResult{}
	
	if strings.Contains(strings.ToLower(query), "coffee") {
		results = append(results, SearchResult{
			Title:       "Coffee Market Prices Rise 15% in 2024",
			URL:         "https://coffeenews.com/market-prices-2024",
			Description: "Global coffee prices have increased significantly due to weather conditions in Brazil and Colombia.",
		})
		results = append(results, SearchResult{
			Title:       "Arabica Coffee Futures Hit New Highs",
			URL:         "https://markets.com/arabica-futures",
			Description: "Arabica coffee futures reached $2.10 per pound, the highest level since 2022.",
		})
		results = append(results, SearchResult{
			Title:       "Starbucks Announces New Sustainability Initiative",
			URL:         "https://starbucks.com/sustainability-2024",
			Description: "The coffee giant commits to carbon-neutral operations by 2030.",
		})
	}

	return results, nil
}

func handleScrapeMarkdown(params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params for scrape_as_markdown")
	}

	url, _ := paramsMap["url"].(string)
	log.Printf("🌐 Scraping: %s", url)

	// Mock scraping results based on URL
	var content string
	
	if strings.Contains(url, "starbucks") {
		content = `# Starbucks Menu
		
## Hot Coffees
- Pike Place Roast - $2.45
- Blonde Roast - $2.45
- Dark Roast - $2.45

## Espresso Drinks
- Latte (Grande) - $5.45
- Cappuccino (Grande) - $4.95
- Americano (Grande) - $3.65

## Cold Drinks
- Iced Coffee (Grande) - $3.25
- Cold Brew (Grande) - $3.45
- Frappuccino (Grande) - $5.95`
	} else if strings.Contains(url, "dunkin") {
		content = `# Dunkin' Menu

## Hot Coffee
- Original Blend - $2.89
- Dark Roast - $2.89
- Decaf - $2.89

## Espresso & Coffee
- Latte (Medium) - $4.59
- Cappuccino (Medium) - $4.19
- Macchiato (Medium) - $4.99

## Cold Coffee
- Iced Coffee (Medium) - $2.99
- Cold Brew (Medium) - $3.39
- Frozen Coffee (Medium) - $4.99`
	} else {
		content = fmt.Sprintf("# Scraped Content from %s\n\nThis is mock content for demonstration purposes.", url)
	}

	return ScrapeResult{
		Content: content,
		URL:     url,
		Success: true,
	}, nil
}

func handleSessionStats(params interface{}) (interface{}, error) {
	log.Printf("📊 Getting session stats")
	
	stats := map[string]interface{}{
		"total_requests":    42,
		"successful_scrapes": 38,
		"failed_scrapes":    4,
		"search_queries":    15,
		"data_transferred":  "2.5MB",
		"session_duration":  "45 minutes",
		"last_activity":     time.Now().Format(time.RFC3339),
	}

	return stats, nil
}

func handleAmazonProduct(params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params for amazon_product")
	}

	url, _ := paramsMap["url"].(string)
	log.Printf("🛒 Amazon Product: %s", url)

	product := map[string]interface{}{
		"title":       "Premium Coffee Beans - Arabica Blend",
		"price":       "$24.99",
		"rating":      4.5,
		"reviews":     1247,
		"availability": "In Stock",
		"description": "Premium quality arabica coffee beans, medium roast, perfect for espresso and drip coffee.",
	}

	return product, nil
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // MCP protocol uses 200 with error in body
	
	response := MCPResponse{
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(response)
}
