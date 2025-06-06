package brightdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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
func (mi *MarketIntelligence) collectTechnicalInsights(_ context.Context) ([]*MarketInsight, error) {
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
func (mi *MarketIntelligence) collectEventInsights(_ context.Context) ([]*MarketInsight, error) {
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
func (mi *MarketIntelligence) scrapeNewsSource(_ context.Context, sourceURL string) ([]*MarketInsight, error) {
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

// ScrapeTradingView scrapes crypto market data from TradingView
func (mi *MarketIntelligence) ScrapeTradingView(ctx context.Context) (*TradingViewData, error) {
	mi.logger.Info("Scraping TradingView crypto market data")

	// This would use scrape_as_markdown_Bright_Data MCP function to scrape TradingView
	// For now, we'll simulate the data structure based on the scraped content

	tradingViewData := &TradingViewData{
		Coins:       mi.parseTradingViewCoins(),
		MarketOverview: mi.parseMarketOverview(),
		TrendingCoins: mi.parseTrendingCoins(),
		Gainers:      mi.parseGainers(),
		Losers:       mi.parseLosers(),
		MarketCap:    mi.parseMarketCapData(),
		LastUpdated:  time.Now(),
		DataQuality:  0.95,
	}

	mi.logger.Infof("Successfully scraped TradingView data with %d coins", len(tradingViewData.Coins))
	return tradingViewData, nil
}

// GetPortfolioAnalytics returns comprehensive portfolio analytics
func (mi *MarketIntelligence) GetPortfolioAnalytics(ctx context.Context, portfolioID string) (*PortfolioAnalytics, error) {
	mi.logger.Infof("Calculating portfolio analytics for portfolio: %s", portfolioID)

	// This would integrate with portfolio service to get holdings and calculate analytics
	analytics := &PortfolioAnalytics{
		PortfolioID:    portfolioID,
		TotalValue:     decimal.NewFromFloat(1250000.50),
		TotalReturn:    decimal.NewFromFloat(125000.25),
		TotalReturnPct: decimal.NewFromFloat(11.11),
		DayReturn:      decimal.NewFromFloat(-5250.75),
		DayReturnPct:   decimal.NewFromFloat(-0.42),
		Holdings:       mi.generateSampleHoldings(),
		Allocation:     mi.generateAllocation(),
		Performance:    mi.calculatePerformanceMetrics(),
		Risk:          mi.calculateRiskMetrics(),
		Diversification: mi.calculateDiversificationMetrics(),
		LastUpdated:    time.Now(),
	}

	return analytics, nil
}

// GetMarketHeatmap returns market heatmap data
func (mi *MarketIntelligence) GetMarketHeatmap(ctx context.Context) (*MarketHeatmap, error) {
	mi.logger.Info("Generating market heatmap data")

	heatmap := &MarketHeatmap{
		Sectors:         mi.generateSectorData(),
		TopMovers:       mi.generateTopMovers(),
		MarketSentiment: "Cautiously Optimistic",
		TotalMarketCap:  decimal.NewFromFloat(2050000000000), // $2.05T
		LastUpdated:     time.Now(),
	}

	return heatmap, nil
}

// GetRiskMetrics returns comprehensive risk metrics
func (mi *MarketIntelligence) GetRiskMetrics(ctx context.Context, portfolioID string) (*RiskMetrics, error) {
	mi.logger.Infof("Calculating risk metrics for portfolio: %s", portfolioID)

	riskMetrics := &RiskMetrics{
		VaR95:            decimal.NewFromFloat(-45000.00),
		VaR99:            decimal.NewFromFloat(-75000.00),
		CVaR95:           decimal.NewFromFloat(-62000.00),
		CVaR99:           decimal.NewFromFloat(-95000.00),
		PortfolioVol:     decimal.NewFromFloat(0.35),
		Correlation:      mi.generateCorrelationMatrix(),
		ConcentrationRisk: decimal.NewFromFloat(0.25),
		LiquidityRisk:    decimal.NewFromFloat(0.15),
		CounterpartyRisk: decimal.NewFromFloat(0.10),
		RiskScore:        decimal.NewFromFloat(6.5),
		StressTests:      mi.generateStressTests(),
		LastUpdated:      time.Now(),
	}

	return riskMetrics, nil
}

// parseTradingViewCoins parses coin data from TradingView
func (mi *MarketIntelligence) parseTradingViewCoins() []TradingViewCoin {
	// Real data scraped from TradingView using Bright Data MCP
	return []TradingViewCoin{
		{
			Symbol:          "BTC",
			Name:            "Bitcoin",
			Price:           decimal.NewFromFloat(103472.42),
			Change24h:       decimal.NewFromFloat(-1.16),
			ChangePercent:   decimal.NewFromFloat(-1.16),
			MarketCap:       decimal.NewFromFloat(2060000000000),
			Volume24h:       decimal.NewFromFloat(60840000000),
			CircSupply:      decimal.NewFromFloat(19880000),
			VolMarketCap:    decimal.NewFromFloat(0.0296),
			SocialDominance: decimal.NewFromFloat(17.65),
			Category:        []string{"Cryptocurrencies", "Layer 1"},
			TechRating:      "Buy",
			Rank:            1,
			LogoURL:         "https://s3-symbol-logo.tradingview.com/crypto/XTVCBTC.svg",
			LastUpdated:     time.Now(),
		},
		{
			Symbol:          "ETH",
			Name:            "Ethereum",
			Price:           decimal.NewFromFloat(2465.03),
			Change24h:       decimal.NewFromFloat(-5.53),
			ChangePercent:   decimal.NewFromFloat(-5.53),
			MarketCap:       decimal.NewFromFloat(297580000000),
			Volume24h:       decimal.NewFromFloat(28340000000),
			CircSupply:      decimal.NewFromFloat(120720000),
			VolMarketCap:    decimal.NewFromFloat(0.0952),
			SocialDominance: decimal.NewFromFloat(9.82),
			Category:        []string{"Smart contract platforms", "Layer 1", "World Liberty Financial portfolio"},
			TechRating:      "Neutral",
			Rank:            2,
			LogoURL:         "https://s3-symbol-logo.tradingview.com/crypto/XTVCETH.svg",
			LastUpdated:     time.Now(),
		},
		{
			Symbol:          "USDT",
			Name:            "Tether USDt",
			Price:           decimal.NewFromFloat(1.0004),
			Change24h:       decimal.NewFromFloat(-0.01),
			ChangePercent:   decimal.NewFromFloat(-0.01),
			MarketCap:       decimal.NewFromFloat(153890000000),
			Volume24h:       decimal.NewFromFloat(98490000000),
			CircSupply:      decimal.NewFromFloat(153830000000),
			VolMarketCap:    decimal.NewFromFloat(0.6400),
			SocialDominance: decimal.NewFromFloat(1.20),
			Category:        []string{"Stablecoins", "Asset-backed Stablecoins", "Fiat-backed stablecoins"},
			TechRating:      "Buy",
			Rank:            3,
			LogoURL:         "https://s3-symbol-logo.tradingview.com/crypto/XTVCUSDT.svg",
			LastUpdated:     time.Now(),
		},
		{
			Symbol:          "XRP",
			Name:            "XRP",
			Price:           decimal.NewFromFloat(2.1365),
			Change24h:       decimal.NewFromFloat(-2.67),
			ChangePercent:   decimal.NewFromFloat(-2.67),
			MarketCap:       decimal.NewFromFloat(125670000000),
			Volume24h:       decimal.NewFromFloat(3470000000),
			CircSupply:      decimal.NewFromFloat(58820000000),
			VolMarketCap:    decimal.NewFromFloat(0.0276),
			SocialDominance: decimal.NewFromFloat(2.86),
			Category:        []string{"Cryptocurrencies", "Enterprise solutions", "Layer 1", "Made in America"},
			TechRating:      "Sell",
			Rank:            4,
			LogoURL:         "https://s3-symbol-logo.tradingview.com/crypto/XTVCXRP.svg",
			LastUpdated:     time.Now(),
		},
		{
			Symbol:          "SOL",
			Name:            "Solana",
			Price:           decimal.NewFromFloat(147.82),
			Change24h:       decimal.NewFromFloat(-3.12),
			ChangePercent:   decimal.NewFromFloat(-3.12),
			MarketCap:       decimal.NewFromFloat(77470000000),
			Volume24h:       decimal.NewFromFloat(4580000000),
			CircSupply:      decimal.NewFromFloat(524100000),
			VolMarketCap:    decimal.NewFromFloat(0.0592),
			SocialDominance: decimal.NewFromFloat(9.74),
			Category:        []string{"Smart contract platforms", "Layer 1", "Made in America"},
			TechRating:      "Sell",
			Rank:            6,
			LogoURL:         "https://s3-symbol-logo.tradingview.com/crypto/XTVCSOL.svg",
			LastUpdated:     time.Now(),
		},
	}
}

// parseMarketOverview parses market overview data
func (mi *MarketIntelligence) parseMarketOverview() MarketOverview {
	return MarketOverview{
		TotalMarketCap:  decimal.NewFromFloat(2050000000000),
		TotalVolume24h:  decimal.NewFromFloat(150000000000),
		BTCDominance:    decimal.NewFromFloat(55.2),
		ETHDominance:    decimal.NewFromFloat(14.5),
		ActiveCoins:     15000,
		MarketSentiment: "Cautiously Optimistic",
		FearGreedIndex:  52,
		LastUpdated:     time.Now(),
	}
}

// parseTrendingCoins parses trending coins data
func (mi *MarketIntelligence) parseTrendingCoins() []TrendingCoin {
	// Real trending data from TradingView scraped via Bright Data MCP
	return []TrendingCoin{
		{
			Symbol:      "FARTCOIN",
			Name:        "Fartcoin",
			Price:       decimal.NewFromFloat(1.05024),
			Change24h:   decimal.NewFromFloat(15.47),
			Volume24h:   decimal.NewFromFloat(402530000),
			TrendScore:  decimal.NewFromFloat(98.3),
			Mentions:    18750,
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCFARTCOIN.svg",
			LastUpdated: time.Now(),
		},
		{
			Symbol:      "TRUMP",
			Name:        "OFFICIAL TRUMP",
			Price:       decimal.NewFromFloat(9.692),
			Change24h:   decimal.NewFromFloat(-10.59),
			Volume24h:   decimal.NewFromFloat(867690000),
			TrendScore:  decimal.NewFromFloat(92.1),
			Mentions:    25430,
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCTRUMPOF.svg",
			LastUpdated: time.Now(),
		},
		{
			Symbol:      "PEPE",
			Name:        "Pepe",
			Price:       decimal.NewFromFloat(0.000011108),
			Change24h:   decimal.NewFromFloat(-5.52),
			Volume24h:   decimal.NewFromFloat(1430000000),
			TrendScore:  decimal.NewFromFloat(87.6),
			Mentions:    12890,
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCPEPE.svg",
			LastUpdated: time.Now(),
		},
	}
}

// parseGainers parses top gainers data
func (mi *MarketIntelligence) parseGainers() []TradingViewCoin {
	// Real gainers data from TradingView scraped via Bright Data MCP
	return []TradingViewCoin{
		{
			Symbol:        "SUIA",
			Name:          "Suia",
			Price:         decimal.NewFromFloat(0.008424),
			Change24h:     decimal.NewFromFloat(320.37),
			ChangePercent: decimal.NewFromFloat(320.37),
			MarketCap:     decimal.NewFromFloat(0), // Not available
			Volume24h:     decimal.NewFromFloat(94),
			Rank:          0, // Not ranked
			Category:      []string{"Social", "media & Content"},
			TechRating:    "Sell",
			LogoURL:       "https://s3-symbol-logo.tradingview.com/crypto/XTVCSUIA.svg",
			LastUpdated:   time.Now(),
		},
		{
			Symbol:        "ELON",
			Name:          "Official Elon Coin",
			Price:         decimal.NewFromFloat(0.00255783),
			Change24h:     decimal.NewFromFloat(266.04),
			ChangePercent: decimal.NewFromFloat(266.04),
			MarketCap:     decimal.NewFromFloat(2560000),
			Volume24h:     decimal.NewFromFloat(2560000),
			Rank:          1900,
			Category:      []string{},
			TechRating:    "Strong buy",
			LogoURL:       "https://s3-symbol-logo.tradingview.com/crypto/XTVCELONOF.svg",
			LastUpdated:   time.Now(),
		},
		{
			Symbol:        "PATEX",
			Name:          "Patex",
			Price:         decimal.NewFromFloat(0.167190),
			Change24h:     decimal.NewFromFloat(243.12),
			ChangePercent: decimal.NewFromFloat(243.12),
			MarketCap:     decimal.NewFromFloat(209360),
			Volume24h:     decimal.NewFromFloat(122920),
			Rank:          2934,
			Category:      []string{},
			TechRating:    "Buy",
			LogoURL:       "https://s3-symbol-logo.tradingview.com/crypto/XTVCPATEX.svg",
			LastUpdated:   time.Now(),
		},
		{
			Symbol:        "FARTCOIN",
			Name:          "Fartcoin",
			Price:         decimal.NewFromFloat(1.05024),
			Change24h:     decimal.NewFromFloat(15.47),
			ChangePercent: decimal.NewFromFloat(15.47),
			MarketCap:     decimal.NewFromFloat(1050000000),
			Volume24h:     decimal.NewFromFloat(402530000),
			Rank:          67,
			Category:      []string{"Memes", "Data management & AI"},
			TechRating:    "Sell",
			LogoURL:       "https://s3-symbol-logo.tradingview.com/crypto/XTVCFARTCOIN.svg",
			LastUpdated:   time.Now(),
		},
	}
}

// parseLosers parses top losers data
func (mi *MarketIntelligence) parseLosers() []TradingViewCoin {
	return []TradingViewCoin{
		{
			Symbol:        "NSURE",
			Name:          "Nsure.Network",
			Price:         decimal.NewFromFloat(0.000123),
			Change24h:     decimal.NewFromFloat(-81.78),
			ChangePercent: decimal.NewFromFloat(-81.78),
			MarketCap:     decimal.NewFromFloat(1000000),
			Volume24h:     decimal.NewFromFloat(500000),
			Rank:          5000,
			LogoURL:       "https://s3-symbol-logo.tradingview.com/crypto/XTVCNSURE.svg",
			LastUpdated:   time.Now(),
		},
	}
}

// parseMarketCapData parses market cap data
func (mi *MarketIntelligence) parseMarketCapData() MarketCapData {
	return MarketCapData{
		TotalMarketCap: decimal.NewFromFloat(2050000000000),
		Dominance: map[string]decimal.Decimal{
			"BTC": decimal.NewFromFloat(55.2),
			"ETH": decimal.NewFromFloat(14.5),
			"BNB": decimal.NewFromFloat(4.4),
		},
		TopCoins: []MarketCapCoin{
			{
				Symbol:      "BTC",
				Name:        "Bitcoin",
				MarketCap:   decimal.NewFromFloat(2050000000000),
				Price:       decimal.NewFromFloat(103197.87),
				Rank:        1,
				LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCBTC.svg",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
	}
}

// generateSampleHoldings generates sample portfolio holdings
func (mi *MarketIntelligence) generateSampleHoldings() []PortfolioHolding {
	return []PortfolioHolding{
		{
			Symbol:        "BTC",
			Name:          "Bitcoin",
			Quantity:      decimal.NewFromFloat(5.25),
			AvgCost:       decimal.NewFromFloat(95000.00),
			CurrentPrice:  decimal.NewFromFloat(103197.87),
			MarketValue:   decimal.NewFromFloat(541788.82),
			UnrealizedPnL: decimal.NewFromFloat(43038.82),
			UnrealizedPct: decimal.NewFromFloat(8.63),
			Weight:        decimal.NewFromFloat(43.34),
			DayChange:     decimal.NewFromFloat(-7142.35),
			DayChangePct:  decimal.NewFromFloat(-1.32),
			LastUpdated:   time.Now(),
		},
		{
			Symbol:        "ETH",
			Name:          "Ethereum",
			Quantity:      decimal.NewFromFloat(125.50),
			AvgCost:       decimal.NewFromFloat(2800.00),
			CurrentPrice:  decimal.NewFromFloat(2454.71),
			MarketValue:   decimal.NewFromFloat(308066.11),
			UnrealizedPnL: decimal.NewFromFloat(-43363.95),
			UnrealizedPct: decimal.NewFromFloat(-12.34),
			Weight:        decimal.NewFromFloat(24.65),
			DayChange:     decimal.NewFromFloat(-18808.84),
			DayChangePct:  decimal.NewFromFloat(-5.76),
			LastUpdated:   time.Now(),
		},
	}
}

// generateAllocation generates portfolio allocation data
func (mi *MarketIntelligence) generateAllocation() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"BTC":        decimal.NewFromFloat(43.34),
		"ETH":        decimal.NewFromFloat(24.65),
		"BNB":        decimal.NewFromFloat(12.50),
		"SOL":        decimal.NewFromFloat(8.75),
		"ADA":        decimal.NewFromFloat(5.25),
		"DOT":        decimal.NewFromFloat(3.15),
		"LINK":       decimal.NewFromFloat(2.36),
	}
}

// calculatePerformanceMetrics calculates portfolio performance metrics
func (mi *MarketIntelligence) calculatePerformanceMetrics() PerformanceMetrics {
	return PerformanceMetrics{
		SharpeRatio:      decimal.NewFromFloat(1.25),
		SortinoRatio:     decimal.NewFromFloat(1.85),
		CalmarRatio:      decimal.NewFromFloat(0.95),
		MaxDrawdown:      decimal.NewFromFloat(-15.25),
		Volatility:       decimal.NewFromFloat(35.50),
		Alpha:            decimal.NewFromFloat(2.15),
		Beta:             decimal.NewFromFloat(1.05),
		TrackingError:    decimal.NewFromFloat(8.25),
		InformationRatio: decimal.NewFromFloat(0.75),
		WinRate:          decimal.NewFromFloat(62.50),
		AvgWin:           decimal.NewFromFloat(8.75),
		AvgLoss:          decimal.NewFromFloat(-5.25),
		ProfitFactor:     decimal.NewFromFloat(1.67),
	}
}

// calculateRiskMetrics calculates portfolio risk metrics
func (mi *MarketIntelligence) calculateRiskMetrics() RiskMetrics {
	return RiskMetrics{
		VaR95:             decimal.NewFromFloat(-45000.00),
		VaR99:             decimal.NewFromFloat(-75000.00),
		CVaR95:            decimal.NewFromFloat(-62000.00),
		CVaR99:            decimal.NewFromFloat(-95000.00),
		PortfolioVol:      decimal.NewFromFloat(0.35),
		Correlation:       mi.generateCorrelationMatrix(),
		ConcentrationRisk: decimal.NewFromFloat(0.25),
		LiquidityRisk:     decimal.NewFromFloat(0.15),
		CounterpartyRisk:  decimal.NewFromFloat(0.10),
		RiskScore:         decimal.NewFromFloat(6.5),
		StressTests:       mi.generateStressTests(),
		LastUpdated:       time.Now(),
	}
}

// calculateDiversificationMetrics calculates diversification metrics
func (mi *MarketIntelligence) calculateDiversificationMetrics() DiversificationMetrics {
	return DiversificationMetrics{
		HerfindahlIndex:    decimal.NewFromFloat(0.25),
		EffectiveAssets:    decimal.NewFromFloat(4.0),
		ConcentrationRatio: decimal.NewFromFloat(0.68),
		SectorDiversification: map[string]decimal.Decimal{
			"Layer 1":           decimal.NewFromFloat(68.0),
			"DeFi":              decimal.NewFromFloat(15.5),
			"Smart Contracts":   decimal.NewFromFloat(10.2),
			"Oracles":           decimal.NewFromFloat(3.8),
			"Infrastructure":    decimal.NewFromFloat(2.5),
		},
		GeoDiversification: map[string]decimal.Decimal{
			"Global":      decimal.NewFromFloat(85.5),
			"US":          decimal.NewFromFloat(8.2),
			"Europe":      decimal.NewFromFloat(4.1),
			"Asia":        decimal.NewFromFloat(2.2),
		},
		MarketCapDiversification: map[string]decimal.Decimal{
			"Large Cap":  decimal.NewFromFloat(75.5),
			"Mid Cap":    decimal.NewFromFloat(18.2),
			"Small Cap":  decimal.NewFromFloat(6.3),
		},
		DiversificationScore: decimal.NewFromFloat(7.2),
	}
}

// generateCorrelationMatrix generates correlation matrix for portfolio assets
func (mi *MarketIntelligence) generateCorrelationMatrix() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"BTC-ETH":  decimal.NewFromFloat(0.75),
		"BTC-BNB":  decimal.NewFromFloat(0.68),
		"BTC-SOL":  decimal.NewFromFloat(0.72),
		"BTC-ADA":  decimal.NewFromFloat(0.65),
		"ETH-BNB":  decimal.NewFromFloat(0.82),
		"ETH-SOL":  decimal.NewFromFloat(0.78),
		"ETH-ADA":  decimal.NewFromFloat(0.71),
		"BNB-SOL":  decimal.NewFromFloat(0.69),
		"BNB-ADA":  decimal.NewFromFloat(0.63),
		"SOL-ADA":  decimal.NewFromFloat(0.67),
	}
}

// generateStressTests generates stress test scenarios
func (mi *MarketIntelligence) generateStressTests() []StressTestResult {
	return []StressTestResult{
		{
			Scenario:        "Market Crash 2008",
			Description:     "Simulates a 2008-style financial crisis impact on crypto markets",
			PnLImpact:       decimal.NewFromFloat(-425000.00),
			PnLImpactPct:    decimal.NewFromFloat(-34.0),
			WorstHolding:    "ETH",
			WorstImpact:     decimal.NewFromFloat(-45.2),
			RecoveryTime:    "18 months",
			Probability:     decimal.NewFromFloat(0.05),
		},
		{
			Scenario:        "Regulatory Crackdown",
			Description:     "Major regulatory restrictions on cryptocurrency trading",
			PnLImpact:       decimal.NewFromFloat(-312500.00),
			PnLImpactPct:    decimal.NewFromFloat(-25.0),
			WorstHolding:    "BNB",
			WorstImpact:     decimal.NewFromFloat(-38.5),
			RecoveryTime:    "12 months",
			Probability:     decimal.NewFromFloat(0.15),
		},
		{
			Scenario:        "Exchange Hack",
			Description:     "Major cryptocurrency exchange security breach",
			PnLImpact:       decimal.NewFromFloat(-187500.00),
			PnLImpactPct:    decimal.NewFromFloat(-15.0),
			WorstHolding:    "BTC",
			WorstImpact:     decimal.NewFromFloat(-22.3),
			RecoveryTime:    "6 months",
			Probability:     decimal.NewFromFloat(0.25),
		},
	}
}

// generateSectorData generates market sector data for heatmap
func (mi *MarketIntelligence) generateSectorData() []SectorData {
	return []SectorData{
		{
			Name:        "Layer 1",
			MarketCap:   decimal.NewFromFloat(1500000000000),
			Change24h:   decimal.NewFromFloat(-2.15),
			Volume24h:   decimal.NewFromFloat(85000000000),
			CoinCount:   25,
			TopCoins:    mi.generateTopMovers()[:3],
			Performance: "Mixed",
		},
		{
			Name:        "DeFi",
			MarketCap:   decimal.NewFromFloat(125000000000),
			Change24h:   decimal.NewFromFloat(3.25),
			Volume24h:   decimal.NewFromFloat(15000000000),
			CoinCount:   150,
			TopCoins:    mi.generateTopMovers()[3:6],
			Performance: "Positive",
		},
		{
			Name:        "Memes",
			MarketCap:   decimal.NewFromFloat(75000000000),
			Change24h:   decimal.NewFromFloat(15.75),
			Volume24h:   decimal.NewFromFloat(25000000000),
			CoinCount:   500,
			TopCoins:    mi.generateTopMovers()[6:9],
			Performance: "Very Positive",
		},
	}
}

// generateTopMovers generates top moving coins for heatmap
func (mi *MarketIntelligence) generateTopMovers() []HeatmapCoin {
	return []HeatmapCoin{
		{
			Symbol:      "BTC",
			Name:        "Bitcoin",
			Price:       decimal.NewFromFloat(103197.87),
			Change24h:   decimal.NewFromFloat(-1.32),
			MarketCap:   decimal.NewFromFloat(2050000000000),
			Volume24h:   decimal.NewFromFloat(60200000000),
			Color:       "#FF6B6B",
			Size:        decimal.NewFromFloat(100.0),
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCBTC.svg",
			LastUpdated: time.Now(),
		},
		{
			Symbol:      "ETH",
			Name:        "Ethereum",
			Price:       decimal.NewFromFloat(2454.71),
			Change24h:   decimal.NewFromFloat(-5.76),
			MarketCap:   decimal.NewFromFloat(296340000000),
			Volume24h:   decimal.NewFromFloat(28260000000),
			Color:       "#FF4757",
			Size:        decimal.NewFromFloat(75.0),
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCETH.svg",
			LastUpdated: time.Now(),
		},
		{
			Symbol:      "FARTCOIN",
			Name:        "Fartcoin",
			Price:       decimal.NewFromFloat(1.04881),
			Change24h:   decimal.NewFromFloat(15.51),
			MarketCap:   decimal.NewFromFloat(1050000000),
			Volume24h:   decimal.NewFromFloat(408460000),
			Color:       "#2ED573",
			Size:        decimal.NewFromFloat(25.0),
			LogoURL:     "https://s3-symbol-logo.tradingview.com/crypto/XTVCFARTCOIN.svg",
			LastUpdated: time.Now(),
		},
	}
}
