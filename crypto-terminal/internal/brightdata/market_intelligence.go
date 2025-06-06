package brightdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// MarketIntelligence collects market intelligence data using Bright Data MCP
type MarketIntelligence struct {
	service *Service
	logger  *logrus.Logger
	
	// Intelligence sources
	newsSources     []string
	analysisSources []string
	eventSources    []string
}

// NewMarketIntelligence creates a new market intelligence collector
func NewMarketIntelligence(service *Service, logger *logrus.Logger) *MarketIntelligence {
	newsSources := []string{
		"https://cointelegraph.com",
		"https://coindesk.com",
		"https://decrypt.co",
		"https://theblock.co",
		"https://cryptonews.com",
	}
	
	analysisSources := []string{
		"https://messari.io",
		"https://glassnode.com",
		"https://santiment.net",
		"https://cryptoquant.com",
	}
	
	eventSources := []string{
		"https://coinmarketcal.com",
		"https://cryptocalendar.pro",
		"https://coindar.org",
	}
	
	return &MarketIntelligence{
		service:         service,
		logger:          logger,
		newsSources:     newsSources,
		analysisSources: analysisSources,
		eventSources:    eventSources,
	}
}

// CollectIntelligence collects market intelligence from various sources
func (mi *MarketIntelligence) CollectIntelligence(ctx context.Context) error {
	mi.logger.Info("Starting market intelligence collection")
	
	var allInsights []*MarketInsight
	
	// Collect insights from different sources
	newsInsights, err := mi.collectNewsInsights(ctx)
	if err != nil {
		mi.logger.Errorf("Failed to collect news insights: %v", err)
	} else {
		allInsights = append(allInsights, newsInsights...)
	}
	
	technicalInsights, err := mi.collectTechnicalInsights(ctx)
	if err != nil {
		mi.logger.Errorf("Failed to collect technical insights: %v", err)
	} else {
		allInsights = append(allInsights, technicalInsights...)
	}
	
	eventInsights, err := mi.collectEventInsights(ctx)
	if err != nil {
		mi.logger.Errorf("Failed to collect event insights: %v", err)
	} else {
		allInsights = append(allInsights, eventInsights...)
	}
	
	// Update service with collected insights
	mi.service.updateInsights(allInsights)
	
	mi.logger.Infof("Collected %d market insights", len(allInsights))
	return nil
}

// collectNewsInsights collects insights from crypto news sources
func (mi *MarketIntelligence) collectNewsInsights(ctx context.Context) ([]*MarketInsight, error) {
	var insights []*MarketInsight
	
	for _, source := range mi.newsSources {
		sourceInsights, err := mi.scrapeNewsSource(ctx, source)
		if err != nil {
			mi.logger.Warnf("Failed to scrape news source %s: %v", source, err)
			continue
		}
		insights = append(insights, sourceInsights...)
	}
	
	return insights, nil
}

// collectTechnicalInsights collects technical analysis insights
func (mi *MarketIntelligence) collectTechnicalInsights(ctx context.Context) ([]*MarketInsight, error) {
	var insights []*MarketInsight
	
	// Simulate technical insights
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("tech_%d", time.Now().Unix()),
		Type:        "technical",
		Category:    "bullish",
		Title:       "Bitcoin Shows Strong Support at $40,000",
		Description: "Technical analysis indicates Bitcoin has established strong support at the $40,000 level with increasing buying volume.",
		Impact:      "medium",
		Confidence:  0.75,
		Symbols:     []string{"BTC"},
		Source:      "Technical Analysis",
		Data: map[string]interface{}{
			"support_level":    40000,
			"resistance_level": 45000,
			"rsi":             45.2,
			"volume_trend":    "increasing",
		},
		CreatedAt: time.Now(),
	})
	
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("tech_%d", time.Now().Unix()+1),
		Type:        "technical",
		Category:    "neutral",
		Title:       "Ethereum Consolidating in Range",
		Description: "Ethereum is consolidating between $2,800 and $3,200, waiting for a breakout direction.",
		Impact:      "low",
		Confidence:  0.65,
		Symbols:     []string{"ETH"},
		Source:      "Technical Analysis",
		Data: map[string]interface{}{
			"range_low":     2800,
			"range_high":    3200,
			"current_price": 3000,
			"pattern":       "consolidation",
		},
		CreatedAt: time.Now(),
	})
	
	return insights, nil
}

// collectEventInsights collects insights from crypto events and announcements
func (mi *MarketIntelligence) collectEventInsights(ctx context.Context) ([]*MarketInsight, error) {
	var insights []*MarketInsight
	
	// Simulate event insights
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:        "fundamental",
		Category:    "bullish",
		Title:       "Major Exchange Announces Bitcoin ETF Support",
		Description: "Leading cryptocurrency exchange announces support for Bitcoin ETF trading, potentially increasing institutional adoption.",
		Impact:      "high",
		Confidence:  0.85,
		Symbols:     []string{"BTC"},
		Source:      "Exchange Announcement",
		Data: map[string]interface{}{
			"event_type":     "announcement",
			"exchange":       "Major Exchange",
			"expected_impact": "positive",
			"timeline":       "Q1 2024",
		},
		CreatedAt: time.Now(),
		ExpiresAt: &[]time.Time{time.Now().Add(7 * 24 * time.Hour)}[0],
	})
	
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("event_%d", time.Now().Unix()+2),
		Type:        "fundamental",
		Category:    "bearish",
		Title:       "Regulatory Concerns in Major Market",
		Description: "Regulatory authorities express concerns about cryptocurrency trading, potentially impacting market sentiment.",
		Impact:      "medium",
		Confidence:  0.70,
		Symbols:     []string{"BTC", "ETH", "BNB"},
		Source:      "Regulatory News",
		Data: map[string]interface{}{
			"event_type":   "regulation",
			"region":       "Major Market",
			"severity":     "medium",
			"timeline":     "ongoing",
		},
		CreatedAt: time.Now(),
		ExpiresAt: &[]time.Time{time.Now().Add(14 * 24 * time.Hour)}[0],
	})
	
	return insights, nil
}

// scrapeNewsSource scrapes insights from a news source
func (mi *MarketIntelligence) scrapeNewsSource(ctx context.Context, sourceURL string) ([]*MarketInsight, error) {
	// This would use scrape_as_markdown_Bright_Data MCP function
	mi.logger.Infof("Scraping news source: %s", sourceURL)
	
	// Simulate scraped insights
	var insights []*MarketInsight
	
	// Extract domain name for source identification
	sourceName := mi.extractSourceName(sourceURL)
	
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("news_%s_%d", sourceName, time.Now().Unix()),
		Type:        "news",
		Category:    "bullish",
		Title:       "Institutional Adoption Continues to Grow",
		Description: "Major financial institutions continue to adopt cryptocurrency, signaling growing mainstream acceptance.",
		Impact:      "high",
		Confidence:  0.80,
		Symbols:     []string{"BTC", "ETH"},
		Source:      sourceName,
		URL:         sourceURL,
		Data: map[string]interface{}{
			"article_type":   "analysis",
			"publish_date":   time.Now().Format("2006-01-02"),
			"author":         "Crypto Analyst",
			"word_count":     1200,
		},
		CreatedAt: time.Now(),
	})
	
	return insights, nil
}

// extractSourceName extracts a clean source name from URL
func (mi *MarketIntelligence) extractSourceName(url string) string {
	// Remove protocol and www
	name := strings.TrimPrefix(url, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "www.")
	
	// Extract domain
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		domain := parts[0]
		// Remove .com, .org, etc.
		domainParts := strings.Split(domain, ".")
		if len(domainParts) > 0 {
			return strings.Title(domainParts[0])
		}
	}
	
	return "Unknown Source"
}

// DetectMarketEvents detects significant market events
func (mi *MarketIntelligence) DetectMarketEvents(ctx context.Context) ([]*MarketEvent, error) {
	mi.logger.Info("Detecting market events")
	
	var events []*MarketEvent
	
	// Simulate market event detection
	events = append(events, &MarketEvent{
		ID:          fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:        "announcement",
		Title:       "Major Partnership Announcement",
		Description: "Leading blockchain project announces strategic partnership with Fortune 500 company.",
		Impact:      "bullish",
		Severity:    "high",
		Symbols:     []string{"ETH", "BNB"},
		Sources:     []string{"Official Announcement", "CryptoNews"},
		EventTime:   time.Now().Add(-30 * time.Minute),
		DetectedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"partnership_type": "strategic",
			"industry":         "technology",
			"market_cap_impact": "positive",
		},
	})
	
	events = append(events, &MarketEvent{
		ID:          fmt.Sprintf("event_%d", time.Now().Unix()+1),
		Type:        "regulation",
		Title:       "New Cryptocurrency Regulations Proposed",
		Description: "Government proposes new regulations for cryptocurrency exchanges and trading.",
		Impact:      "bearish",
		Severity:    "medium",
		Symbols:     []string{"BTC", "ETH", "BNB"},
		Sources:     []string{"Government Press Release", "Financial News"},
		EventTime:   time.Now().Add(-2 * time.Hour),
		DetectedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"regulation_type": "exchange_oversight",
			"implementation_timeline": "6_months",
			"affected_regions": []string{"US", "EU"},
		},
	})
	
	return events, nil
}

// AnalyzeMarketSentiment analyzes overall market sentiment
func (mi *MarketIntelligence) AnalyzeMarketSentiment(ctx context.Context) (map[string]interface{}, error) {
	mi.logger.Info("Analyzing market sentiment")
	
	// Simulate market sentiment analysis
	sentiment := map[string]interface{}{
		"overall_sentiment": 0.15, // Slightly positive
		"sentiment_score":   57,   // 0-100 scale
		"confidence":        0.75,
		"trend":            "improving",
		"key_drivers": []string{
			"institutional_adoption",
			"technical_breakout",
			"positive_news_flow",
		},
		"risk_factors": []string{
			"regulatory_uncertainty",
			"market_volatility",
			"macroeconomic_concerns",
		},
		"sentiment_breakdown": map[string]interface{}{
			"news":        0.25,  // Positive news sentiment
			"social":      0.10,  // Slightly positive social sentiment
			"technical":   0.20,  // Positive technical indicators
			"fundamental": 0.05,  // Neutral fundamental sentiment
		},
		"last_updated": time.Now(),
	}
	
	return sentiment, nil
}

// GetInfluencerInsights gets insights from crypto influencers
func (mi *MarketIntelligence) GetInfluencerInsights(ctx context.Context) ([]*MarketInsight, error) {
	mi.logger.Info("Collecting influencer insights")
	
	var insights []*MarketInsight
	
	// Simulate influencer insights
	insights = append(insights, &MarketInsight{
		ID:          fmt.Sprintf("influencer_%d", time.Now().Unix()),
		Type:        "social",
		Category:    "bullish",
		Title:       "Crypto Influencer Bullish on DeFi",
		Description: "Leading crypto influencer expresses strong bullish sentiment on DeFi protocols and their future potential.",
		Impact:      "medium",
		Confidence:  0.70,
		Symbols:     []string{"ETH", "UNI", "AAVE"},
		Source:      "Crypto Influencer",
		Data: map[string]interface{}{
			"influencer":      "CryptoExpert",
			"followers":       500000,
			"engagement_rate": 0.08,
			"platform":        "twitter",
			"post_type":       "thread",
		},
		CreatedAt: time.Now(),
	})
	
	return insights, nil
}
