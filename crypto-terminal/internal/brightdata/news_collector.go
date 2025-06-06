package brightdata

import (
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// NewsCollector collects crypto news using Bright Data MCP
type NewsCollector struct {
	service *Service
	logger  *logrus.Logger
	
	// Crypto-related keywords for filtering
	cryptoKeywords []string
	symbolRegex    *regexp.Regexp
}

// NewNewsCollector creates a new news collector
func NewNewsCollector(service *Service, logger *logrus.Logger) *NewsCollector {
	cryptoKeywords := []string{
		"bitcoin", "btc", "ethereum", "eth", "crypto", "cryptocurrency",
		"blockchain", "defi", "nft", "altcoin", "binance", "coinbase",
		"trading", "hodl", "mining", "staking", "yield", "liquidity",
		"smart contract", "dapp", "web3", "metaverse", "dao",
	}
	
	// Regex to find crypto symbols (3-5 uppercase letters)
	symbolRegex := regexp.MustCompile(`\b[A-Z]{3,5}\b`)
	
	return &NewsCollector{
		service:        service,
		logger:         logger,
		cryptoKeywords: cryptoKeywords,
		symbolRegex:    symbolRegex,
	}
}

// CollectNews collects crypto news from various sources
func (nc *NewsCollector) CollectNews(ctx context.Context) error {
	nc.logger.Info("Starting news collection")
	
	// Search for crypto news using different queries
	queries := []string{
		"bitcoin cryptocurrency news",
		"ethereum defi news",
		"crypto market analysis",
		"blockchain technology news",
		"cryptocurrency trading news",
	}
	
	allArticles := make(map[string][]*NewsArticle)
	
	for _, query := range queries {
		articles, err := nc.searchAndProcessNews(ctx, query)
		if err != nil {
			nc.logger.Errorf("Failed to search news for query '%s': %v", query, err)
			continue
		}
		
		// Group articles by symbol
		for _, article := range articles {
			for _, symbol := range article.Symbols {
				if allArticles[symbol] == nil {
					allArticles[symbol] = make([]*NewsArticle, 0)
				}
				allArticles[symbol] = append(allArticles[symbol], article)
			}
			
			// Also add to general crypto news
			if allArticles["CRYPTO"] == nil {
				allArticles["CRYPTO"] = make([]*NewsArticle, 0)
			}
			allArticles["CRYPTO"] = append(allArticles["CRYPTO"], article)
		}
	}
	
	// Update service with collected news
	for symbol, articles := range allArticles {
		// Limit to latest 50 articles per symbol
		if len(articles) > 50 {
			articles = articles[:50]
		}
		nc.service.updateNews(symbol, articles)
	}
	
	nc.logger.Infof("Collected news for %d symbols", len(allArticles))
	return nil
}

// searchAndProcessNews searches for news and processes the results
func (nc *NewsCollector) searchAndProcessNews(ctx context.Context, query string) ([]*NewsArticle, error) {
	// Use Bright Data search engine
	searchResults, err := nc.SearchNews(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search news: %w", err)
	}
	
	var articles []*NewsArticle
	
	for _, result := range searchResults {
		// Filter for crypto relevance
		if !nc.isCryptoRelevant(result.Title + " " + result.Description) {
			continue
		}
		
		// Scrape full content
		content, err := nc.ScrapeContent(ctx, result.URL)
		if err != nil {
			nc.logger.Warnf("Failed to scrape content from %s: %v", result.URL, err)
			continue
		}
		
		// Create news article
		article := &NewsArticle{
			ID:          nc.generateID(result.URL),
			Title:       result.Title,
			Summary:     result.Description,
			Content:     content.Content,
			URL:         result.URL,
			Source:      result.Source,
			PublishedAt: result.Timestamp,
			Symbols:     nc.extractSymbols(result.Title + " " + result.Description + " " + content.Content),
			Tags:        nc.extractTags(content.Content),
			Sentiment:   nc.calculateSentiment(content.Content),
			Relevance:   result.Relevance,
			CreatedAt:   time.Now(),
		}
		
		articles = append(articles, article)
	}
	
	return articles, nil
}

// SearchNews searches for news using Bright Data search engine
func (nc *NewsCollector) SearchNews(ctx context.Context, query string) ([]*SearchResult, error) {
	// This would use the search_engine_Bright_Data MCP function
	// For now, we'll simulate the call
	
	nc.logger.Infof("Searching news for query: %s", query)
	
	// Simulate search results
	results := []*SearchResult{
		{
			Title:       "Bitcoin Reaches New All-Time High Amid Institutional Adoption",
			URL:         "https://example.com/bitcoin-ath",
			Description: "Bitcoin surged to a new record high as major institutions continue to adopt cryptocurrency",
			Source:      "CryptoNews",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			Relevance:   0.95,
		},
		{
			Title:       "Ethereum 2.0 Staking Rewards Attract More Validators",
			URL:         "https://example.com/eth-staking",
			Description: "Ethereum's proof-of-stake mechanism continues to attract validators with attractive rewards",
			Source:      "DeFi Daily",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Relevance:   0.90,
		},
		{
			Title:       "DeFi Protocol Launches New Yield Farming Opportunities",
			URL:         "https://example.com/defi-yield",
			Description: "New decentralized finance protocol offers high-yield farming opportunities for liquidity providers",
			Source:      "DeFi Pulse",
			Timestamp:   time.Now().Add(-3 * time.Hour),
			Relevance:   0.85,
		},
	}
	
	return results, nil
}

// ScrapeContent scrapes content from a URL using Bright Data
func (nc *NewsCollector) ScrapeContent(ctx context.Context, url string) (*ScrapedContent, error) {
	// This would use the scrape_as_markdown_Bright_Data MCP function
	// For now, we'll simulate the call
	
	nc.logger.Infof("Scraping content from: %s", url)
	
	// Simulate scraped content
	content := &ScrapedContent{
		URL:         url,
		Title:       "Sample Crypto News Article",
		Content:     "This is sample content about cryptocurrency market movements and blockchain technology developments. Bitcoin and Ethereum continue to show strong performance in the market.",
		Metadata:    map[string]interface{}{"author": "Crypto Reporter", "published": time.Now().Format(time.RFC3339)},
		ScrapedAt:   time.Now(),
		ContentType: "article",
		WordCount:   150,
		Language:    "en",
	}
	
	return content, nil
}

// isCryptoRelevant checks if content is relevant to cryptocurrency
func (nc *NewsCollector) isCryptoRelevant(content string) bool {
	content = strings.ToLower(content)
	
	for _, keyword := range nc.cryptoKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	
	return false
}

// extractSymbols extracts cryptocurrency symbols from text
func (nc *NewsCollector) extractSymbols(text string) []string {
	matches := nc.symbolRegex.FindAllString(text, -1)
	
	// Filter for known crypto symbols
	knownSymbols := map[string]bool{
		"BTC": true, "ETH": true, "BNB": true, "ADA": true, "SOL": true,
		"XRP": true, "DOT": true, "DOGE": true, "AVAX": true, "MATIC": true,
		"LINK": true, "UNI": true, "LTC": true, "ATOM": true, "XLM": true,
		"VET": true, "FIL": true, "TRX": true, "ETC": true, "THETA": true,
	}
	
	var symbols []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if knownSymbols[match] && !seen[match] {
			symbols = append(symbols, match)
			seen[match] = true
		}
	}
	
	return symbols
}

// extractTags extracts relevant tags from content
func (nc *NewsCollector) extractTags(content string) []string {
	content = strings.ToLower(content)
	
	tagKeywords := map[string]string{
		"defi":        "DeFi",
		"nft":         "NFT",
		"mining":      "Mining",
		"staking":     "Staking",
		"trading":     "Trading",
		"regulation":  "Regulation",
		"adoption":    "Adoption",
		"partnership": "Partnership",
		"hack":        "Security",
		"upgrade":     "Technology",
		"fork":        "Technology",
		"yield":       "DeFi",
		"liquidity":   "DeFi",
		"governance":  "DAO",
		"metaverse":   "Metaverse",
		"web3":        "Web3",
	}
	
	var tags []string
	seen := make(map[string]bool)
	
	for keyword, tag := range tagKeywords {
		if strings.Contains(content, keyword) && !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}
	
	return tags
}

// calculateSentiment calculates sentiment score for content
func (nc *NewsCollector) calculateSentiment(content string) float64 {
	content = strings.ToLower(content)
	
	positiveWords := []string{
		"bullish", "positive", "growth", "increase", "rise", "surge", "pump",
		"adoption", "partnership", "upgrade", "breakthrough", "success",
		"profit", "gain", "rally", "moon", "optimistic", "confident",
	}
	
	negativeWords := []string{
		"bearish", "negative", "decline", "decrease", "fall", "crash", "dump",
		"hack", "scam", "regulation", "ban", "concern", "risk", "loss",
		"fear", "uncertainty", "doubt", "pessimistic", "worried",
	}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		positiveCount += strings.Count(content, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(content, word)
	}
	
	total := positiveCount + negativeCount
	if total == 0 {
		return 0.0 // Neutral
	}
	
	// Calculate sentiment score between -1 and 1
	sentiment := float64(positiveCount-negativeCount) / float64(total)
	
	// Normalize to -1 to 1 range
	if sentiment > 1 {
		sentiment = 1
	} else if sentiment < -1 {
		sentiment = -1
	}
	
	return sentiment
}

// generateID generates a unique ID for an article
func (nc *NewsCollector) generateID(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x", hash)[:16]
}
