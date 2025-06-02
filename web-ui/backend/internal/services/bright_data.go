package services

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type BrightDataService struct {
	apiToken  string
	baseURL   string
	client    *http.Client
	mcpClient *BrightDataMCPService
}

type ScrapingRequest struct {
	URL     string            `json:"url"`
	Format  string            `json:"format"` // "markdown" or "html"
	Options map[string]string `json:"options,omitempty"`
}

type ScrapingResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

type MarketDataItem struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Price       *float64  `json:"price,omitempty"`
	Change      *float64  `json:"change,omitempty"`
	URL         string    `json:"url"`
	LastUpdated time.Time `json:"lastUpdated"`
	Category    string    `json:"category"`
	Content     string    `json:"content,omitempty"`
}

func NewBrightDataService() *BrightDataService {
	return &BrightDataService{
		apiToken:  os.Getenv("BRIGHT_DATA_API_TOKEN"),
		baseURL:   "https://api.brightdata.com",
		mcpClient: NewBrightDataMCPService(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *BrightDataService) ScrapeURL(url string, format string) (*ScrapingResponse, error) {
	// This would integrate with Bright Data's actual API
	// For now, we'll return mock data

	// Simulate different types of scraped data based on URL
	var mockData interface{}
	var category string

	switch {
	case contains(url, "starbucks"):
		category = "competitors"
		mockData = map[string]interface{}{
			"title": "Starbucks Menu Prices",
			"prices": []map[string]interface{}{
				{"item": "Grande Latte", "price": 5.45, "change": 0.15},
				{"item": "Venti Americano", "price": 3.95, "change": -0.05},
			},
		}
	case contains(url, "coffee") && contains(url, "futures"):
		category = "coffee-prices"
		mockData = map[string]interface{}{
			"title":         "Coffee Futures Prices",
			"arabica_price": 1.85,
			"robusta_price": 1.23,
			"change_24h":    2.3,
		}
	case contains(url, "news"):
		category = "news"
		mockData = map[string]interface{}{
			"title": "Coffee Industry News",
			"articles": []map[string]interface{}{
				{
					"headline": "New Sustainable Coffee Farming Practices",
					"summary":  "Industry adopts new eco-friendly methods",
					"date":     time.Now().Format("2006-01-02"),
				},
			},
		}
	default:
		category = "general"
		mockData = map[string]interface{}{
			"title":   "Scraped Content",
			"content": "Mock scraped content from " + url,
		}
	}

	return &ScrapingResponse{
		Success: true,
		Data: map[string]interface{}{
			"url":        url,
			"category":   category,
			"data":       mockData,
			"scraped_at": time.Now(),
		},
	}, nil
}

func (s *BrightDataService) GetMarketData() ([]MarketDataItem, error) {
	var allItems []MarketDataItem

	// Get competitor data
	competitorItems, err := s.ScrapeCompetitorPrices()
	if err != nil {
		fmt.Printf("Warning: Failed to get competitor data: %v\n", err)
		// Continue with other sources
	} else {
		allItems = append(allItems, competitorItems...)
	}

	// Get market news
	newsItems, err := s.ScrapeMarketNews()
	if err != nil {
		fmt.Printf("Warning: Failed to get market news: %v\n", err)
		// Continue with other sources
	} else {
		allItems = append(allItems, newsItems...)
	}

	// Get coffee futures data
	futuresItems, err := s.ScrapeCoffeeFutures()
	if err != nil {
		fmt.Printf("Warning: Failed to get futures data: %v\n", err)
		// Continue with other sources
	} else {
		allItems = append(allItems, futuresItems...)
	}

	// Get social media trends
	socialItems, err := s.ScrapeSocialTrends()
	if err != nil {
		fmt.Printf("Warning: Failed to get social trends: %v\n", err)
		// Continue with other sources
	} else {
		allItems = append(allItems, socialItems...)
	}

	// If we have no data at all, return fallback mock data
	if len(allItems) == 0 {
		fmt.Println("Warning: No real data available, returning fallback mock data")
		return s.getFallbackMarketData(), nil
	}

	return allItems, nil
}

func (s *BrightDataService) RefreshMarketData() error {
	// This would trigger fresh scraping of all configured sources
	// For now, we'll just simulate a refresh
	time.Sleep(2 * time.Second) // Simulate processing time
	return nil
}

func (s *BrightDataService) GetDataSources() ([]map[string]interface{}, error) {
	sources := []map[string]interface{}{
		{
			"id":          "starbucks",
			"name":        "Starbucks",
			"url":         "https://starbucks.com",
			"category":    "competitors",
			"status":      "active",
			"last_update": time.Now().Add(-10 * time.Minute),
		},
		{
			"id":          "dunkin",
			"name":        "Dunkin'",
			"url":         "https://dunkin.com",
			"category":    "competitors",
			"status":      "active",
			"last_update": time.Now().Add(-15 * time.Minute),
		},
		{
			"id":          "coffee-futures",
			"name":        "Coffee Futures",
			"url":         "https://markets.com/coffee",
			"category":    "coffee-prices",
			"status":      "active",
			"last_update": time.Now().Add(-5 * time.Minute),
		},
		{
			"id":          "coffee-news",
			"name":        "Coffee News Daily",
			"url":         "https://coffeenews.com",
			"category":    "news",
			"status":      "active",
			"last_update": time.Now().Add(-30 * time.Minute),
		},
		{
			"id":          "twitter-coffee",
			"name":        "Twitter Coffee Trends",
			"url":         "https://twitter.com/search?q=coffee",
			"category":    "social",
			"status":      "active",
			"last_update": time.Now().Add(-2 * time.Minute),
		},
	}

	return sources, nil
}

// ScrapeCompetitorPrices scrapes competitor pricing data using MCP integration
func (s *BrightDataService) ScrapeCompetitorPrices() ([]MarketDataItem, error) {
	competitors := []struct {
		name string
		url  string
	}{
		{"Starbucks", "https://www.starbucks.com/menu/drinks"},
		{"Dunkin'", "https://www.dunkindonuts.com/en/menu"},
		{"Costa Coffee", "https://www.costa.co.uk/menu"},
	}

	var items []MarketDataItem

	for i, competitor := range competitors {
		// Use real MCP to scrape competitor data
		resp, err := s.mcpClient.ScrapeURL(competitor.url)
		if err != nil {
			fmt.Printf("Failed to scrape %s: %v\n", competitor.name, err)
			continue
		}

		// Parse the scraped data
		item := MarketDataItem{
			ID:          fmt.Sprintf("comp_%d", i+1),
			Source:      competitor.name,
			Title:       fmt.Sprintf("Menu prices from %s", competitor.name),
			URL:         competitor.url,
			LastUpdated: time.Now(),
			Category:    "competitors",
		}

		// Try to extract content from response
		if resp.Data != nil {
			if dataMap, ok := resp.Data.(map[string]interface{}); ok {
				if content, exists := dataMap["content"]; exists {
					if contentStr, ok := content.(string); ok {
						item.Content = contentStr
					}
				}
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// ScrapeMarketNews scrapes coffee market news using MCP integration
func (s *BrightDataService) ScrapeMarketNews() ([]MarketDataItem, error) {
	// Search for coffee market news using Bright Data MCP
	queries := []string{
		"coffee market prices 2024",
		"arabica coffee futures",
		"coffee industry news",
		"coffee commodity prices",
	}

	var items []MarketDataItem

	for i, query := range queries {
		// Use real MCP to search for news
		resp, err := s.mcpClient.SearchEngine(query, "google")
		if err != nil {
			fmt.Printf("Failed to search for '%s': %v\n", query, err)
			continue
		}

		// Parse search results
		if resp.Data != nil {
			if results, ok := resp.Data.([]interface{}); ok {
				for j, result := range results {
					if resultMap, ok := result.(map[string]interface{}); ok {
						item := MarketDataItem{
							ID:          fmt.Sprintf("news_%d_%d", i+1, j+1),
							Source:      "Search Results",
							LastUpdated: time.Now(),
							Category:    "news",
						}

						if title, exists := resultMap["title"]; exists {
							if titleStr, ok := title.(string); ok {
								item.Title = titleStr
							}
						}

						if url, exists := resultMap["url"]; exists {
							if urlStr, ok := url.(string); ok {
								item.URL = urlStr
							}
						}

						if description, exists := resultMap["description"]; exists {
							if descStr, ok := description.(string); ok {
								item.Content = descStr
							}
						}

						items = append(items, item)
					}
				}
			}
		}
	}

	return items, nil
}

// ScrapeCoffeeFutures scrapes coffee futures data
func (s *BrightDataService) ScrapeCoffeeFutures() ([]MarketDataItem, error) {
	// Search for coffee futures data
	resp, err := s.mcpClient.SearchEngine("coffee futures prices arabica robusta", "google")
	if err != nil {
		return nil, fmt.Errorf("failed to search coffee futures: %w", err)
	}

	var items []MarketDataItem

	// Parse search results for futures data
	if resp.Data != nil {
		if results, ok := resp.Data.([]interface{}); ok {
			for i, result := range results {
				if resultMap, ok := result.(map[string]interface{}); ok {
					item := MarketDataItem{
						ID:          fmt.Sprintf("futures_%d", i+1),
						Source:      "Coffee Futures",
						LastUpdated: time.Now(),
						Category:    "coffee-prices",
					}

					if title, exists := resultMap["title"]; exists {
						if titleStr, ok := title.(string); ok {
							item.Title = titleStr
						}
					}

					if url, exists := resultMap["url"]; exists {
						if urlStr, ok := url.(string); ok {
							item.URL = urlStr
						}
					}

					items = append(items, item)
				}
			}
		}
	}

	return items, nil
}

// ScrapeSocialTrends scrapes social media trends about coffee
func (s *BrightDataService) ScrapeSocialTrends() ([]MarketDataItem, error) {
	// Search for coffee trends on social media
	resp, err := s.mcpClient.SearchEngine("coffee trends social media twitter", "google")
	if err != nil {
		return nil, fmt.Errorf("failed to search social trends: %w", err)
	}

	var items []MarketDataItem

	// Parse search results for social trends
	if resp.Data != nil {
		if results, ok := resp.Data.([]interface{}); ok {
			for i, result := range results {
				if resultMap, ok := result.(map[string]interface{}); ok {
					item := MarketDataItem{
						ID:          fmt.Sprintf("social_%d", i+1),
						Source:      "Social Media",
						LastUpdated: time.Now(),
						Category:    "social",
					}

					if title, exists := resultMap["title"]; exists {
						if titleStr, ok := title.(string); ok {
							item.Title = titleStr
						}
					}

					if url, exists := resultMap["url"]; exists {
						if urlStr, ok := url.(string); ok {
							item.URL = urlStr
						}
					}

					items = append(items, item)
				}
			}
		}
	}

	return items, nil
}

// getFallbackMarketData returns mock data as fallback when real data is unavailable
func (s *BrightDataService) getFallbackMarketData() []MarketDataItem {
	// Helper function to create float64 pointers
	floatPtr := func(f float64) *float64 { return &f }

	return []MarketDataItem{
		{
			ID:          "fallback_1",
			Source:      "Starbucks (Fallback)",
			Title:       "Grande Latte Price Update",
			Price:       floatPtr(5.45),
			Change:      floatPtr(0.15),
			URL:         "https://starbucks.com",
			LastUpdated: time.Now().Add(-10 * time.Minute),
			Category:    "competitors",
		},
		{
			ID:          "fallback_2",
			Source:      "Coffee Futures (Fallback)",
			Title:       "Arabica Coffee Futures Rise 2.3%",
			Price:       floatPtr(1.85),
			Change:      floatPtr(2.3),
			URL:         "https://markets.com",
			LastUpdated: time.Now().Add(-15 * time.Minute),
			Category:    "coffee-prices",
		},
		{
			ID:          "fallback_3",
			Source:      "Coffee News (Fallback)",
			Title:       "New Sustainable Coffee Farming Practices",
			URL:         "https://coffeenews.com",
			LastUpdated: time.Now().Add(-30 * time.Minute),
			Category:    "news",
		},
	}
}

// Helper function to extract domain from URL
func extractDomain(rawURL string) string {
	// Remove protocol
	if strings.HasPrefix(rawURL, "https://") {
		rawURL = rawURL[8:]
	} else if strings.HasPrefix(rawURL, "http://") {
		rawURL = rawURL[7:]
	}

	// Remove www prefix
	rawURL = strings.TrimPrefix(rawURL, "www.")

	// Find first slash and return everything before it
	if idx := strings.Index(rawURL, "/"); idx != -1 {
		return rawURL[:idx]
	}

	return rawURL
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
