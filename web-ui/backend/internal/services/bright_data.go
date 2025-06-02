package services

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type BrightDataService struct {
	apiToken string
	baseURL  string
	client   *http.Client
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
		apiToken: os.Getenv("BRIGHT_DATA_API_TOKEN"),
		baseURL:  "https://api.brightdata.com",
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
	// Mock market data that would come from various scraped sources

	// Helper function to create float64 pointers
	floatPtr := func(f float64) *float64 { return &f }

	items := []MarketDataItem{
		{
			ID:          "1",
			Source:      "Starbucks",
			Title:       "Grande Latte Price Update",
			Price:       floatPtr(5.45),
			Change:      floatPtr(0.15),
			URL:         "https://starbucks.com",
			LastUpdated: time.Now().Add(-10 * time.Minute),
			Category:    "competitors",
		},
		{
			ID:          "2",
			Source:      "Coffee Futures",
			Title:       "Arabica Coffee Futures Rise 2.3%",
			Price:       floatPtr(1.85),
			Change:      floatPtr(2.3),
			URL:         "https://markets.com",
			LastUpdated: time.Now().Add(-15 * time.Minute),
			Category:    "coffee-prices",
		},
		{
			ID:          "3",
			Source:      "Coffee News Daily",
			Title:       "New Sustainable Coffee Farming Practices",
			URL:         "https://coffeenews.com",
			LastUpdated: time.Now().Add(-30 * time.Minute),
			Category:    "news",
		},
		{
			ID:          "4",
			Source:      "Dunkin",
			Title:       "Medium Coffee Price Analysis",
			Price:       floatPtr(2.89),
			Change:      floatPtr(-0.05),
			URL:         "https://dunkin.com",
			LastUpdated: time.Now().Add(-20 * time.Minute),
			Category:    "competitors",
		},
		{
			ID:          "5",
			Source:      "Twitter",
			Title:       "Coffee trends gaining momentum #CoffeeLovers",
			URL:         "https://twitter.com",
			LastUpdated: time.Now().Add(-5 * time.Minute),
			Category:    "social",
		},
	}

	return items, nil
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
	competitors := []string{
		"https://www.starbucks.com/menu",
		"https://www.dunkindonuts.com/en/menu",
		"https://www.costacoffe.com/menu",
	}

	var items []MarketDataItem

	for _, url := range competitors {
		// In a real implementation, we would use MCP to scrape each competitor
		// For now, we'll use mock data but show the integration structure

		// Mock data for demonstration
		item := MarketDataItem{
			ID:          fmt.Sprintf("comp_%d", len(items)+1),
			Source:      extractDomain(url),
			Title:       fmt.Sprintf("Menu prices from %s", extractDomain(url)),
			URL:         url,
			LastUpdated: time.Now(),
			Category:    "competitors",
		}

		items = append(items, item)
	}

	return items, nil
}

// ScrapeMarketNews scrapes coffee market news using MCP integration
func (s *BrightDataService) ScrapeMarketNews() ([]MarketDataItem, error) {
	// In a real implementation, we would use MCP to search for coffee market news
	// For now, we'll return mock news data

	items := []MarketDataItem{
		{
			ID:          "news_1",
			Source:      "Coffee Market News",
			Title:       "Global Coffee Prices Rise Due to Weather Concerns",
			URL:         "https://coffeenews.com/article1",
			LastUpdated: time.Now().Add(-1 * time.Hour),
			Category:    "news",
		},
		{
			ID:          "news_2",
			Source:      "Reuters",
			Title:       "Brazil Coffee Harvest Expected to Increase 15%",
			URL:         "https://reuters.com/coffee-harvest",
			LastUpdated: time.Now().Add(-2 * time.Hour),
			Category:    "news",
		},
	}

	return items, nil
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
