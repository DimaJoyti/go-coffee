package brightdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Service provides Bright Data integration for enhanced market intelligence
type Service struct {
	config           *BrightDataConfig
	redis            *redis.Client
	logger           *logrus.Logger
	newsCollector    *NewsCollector
	sentimentAnalyzer *SentimentAnalyzer
	marketIntelligence *MarketIntelligence

	// New enhanced scrapers
	commasScraper    *CommasScraper
	tradingViewEnhanced *TradingViewEnhancedScraper
	socialScraper    *SocialScraper

	// State management
	isRunning        bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	mutex            sync.RWMutex

	// Data storage
	latestNews       map[string][]*NewsArticle
	latestSentiment  map[string]*SentimentAnalysis
	latestInsights   []*MarketInsight
	trendingTopics   []*TrendingTopic
	lastUpdate       time.Time

	// New data storage for enhanced features
	tradingSignals   map[string][]*TradingSignal
	tradingBots      []*TradingBot
	technicalAnalysis map[string]*TechnicalAnalysis
	activeDeals      []*TradingDeal

	// Quality metrics
	qualityMetrics   map[string]*DataQualityMetrics
}

// NewService creates a new Bright Data service
func NewService(config *BrightDataConfig, redis *redis.Client, logger *logrus.Logger) *Service {
	service := &Service{
		config:           config,
		redis:            redis,
		logger:           logger,
		stopChan:         make(chan struct{}),
		latestNews:       make(map[string][]*NewsArticle),
		latestSentiment:  make(map[string]*SentimentAnalysis),
		latestInsights:   make([]*MarketInsight, 0),
		trendingTopics:   make([]*TrendingTopic, 0),
		qualityMetrics:   make(map[string]*DataQualityMetrics),

		// Initialize new data storage
		tradingSignals:   make(map[string][]*TradingSignal),
		tradingBots:      make([]*TradingBot, 0),
		technicalAnalysis: make(map[string]*TechnicalAnalysis),
		activeDeals:      make([]*TradingDeal, 0),
	}

	// Initialize sub-services
	service.newsCollector = NewNewsCollector(service, logger)
	service.sentimentAnalyzer = NewSentimentAnalyzer(service, logger)
	service.marketIntelligence = NewMarketIntelligence(service, logger)

	// Initialize new enhanced scrapers
	commasConfig := &CommasScraperConfig{
		BaseURL:         "https://3commas.io",
		UpdateInterval:  config.UpdateInterval,
		MaxConcurrent:   config.MaxConcurrent,
		RateLimitRPS:    config.RateLimitRPS,
		EnableBots:      true,
		EnableSignals:   true,
		EnableDeals:     true,
		TargetExchanges: []string{"binance", "coinbase", "kraken"},
		TargetPairs:     []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
	}
	service.commasScraper = NewCommasScraper(commasConfig, logger)

	tradingViewConfig := &TradingViewConfig{
		BaseURL:          "https://tradingview.com",
		UpdateInterval:   config.UpdateInterval,
		MaxConcurrent:    config.MaxConcurrent,
		RateLimitRPS:     config.RateLimitRPS,
		EnableIdeas:      true,
		EnableScreeners:  true,
		EnableIndicators: true,
		TargetSymbols:    []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "SOLUSDT"},
		TimeFrames:       []string{"1h", "4h", "1d"},
	}
	service.tradingViewEnhanced = NewTradingViewEnhancedScraper(tradingViewConfig, logger)

	socialConfig := &SocialScraperConfig{
		UpdateInterval:     config.UpdateInterval,
		MaxConcurrent:      config.MaxConcurrent,
		RateLimitRPS:       config.RateLimitRPS,
		EnableTwitter:      true,
		EnableReddit:       true,
		EnableTelegram:     true,
		TwitterKeywords:    []string{"bitcoin", "ethereum", "crypto", "defi"},
		RedditSubreddits:   []string{"cryptocurrency", "bitcoin", "ethereum", "defi"},
		TelegramChannels:   []string{"cryptosignals", "tradingview"},
		InfluencerAccounts: []string{"elonmusk", "VitalikButerin", "cz_binance"},
		MinFollowers:       10000,
	}
	service.socialScraper = NewSocialScraper(socialConfig, logger)

	return service
}

// Start starts the Bright Data service
func (s *Service) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("bright data service is already running")
	}
	
	if !s.config.Enabled {
		s.logger.Info("Bright Data service is disabled")
		return nil
	}
	
	s.isRunning = true
	
	// Start data collection goroutines
	if s.config.EnableNews {
		s.wg.Add(1)
		go s.collectNews(ctx)
	}

	if s.config.EnableSentiment {
		s.wg.Add(1)
		go s.analyzeSentiment(ctx)
	}

	if s.config.EnableEvents {
		s.wg.Add(1)
		go s.collectMarketIntelligence(ctx)
	}

	// Start new enhanced scrapers
	s.wg.Add(1)
	go s.collectCommasData(ctx)

	s.wg.Add(1)
	go s.collectTradingViewEnhanced(ctx)

	if s.config.EnableSocial {
		s.wg.Add(1)
		go s.collectSocialData(ctx)
	}

	s.wg.Add(1)
	go s.updateQualityMetrics(ctx)
	
	s.logger.Info("Bright Data service started")
	return nil
}

// Stop stops the Bright Data service
func (s *Service) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return nil
	}
	
	close(s.stopChan)
	s.wg.Wait()
	
	s.isRunning = false
	s.logger.Info("Bright Data service stopped")
	return nil
}

// GetNews returns latest news articles for symbols
func (s *Service) GetNews(ctx context.Context, symbols []string, limit int) ([]*NewsArticle, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var allNews []*NewsArticle
	
	if len(symbols) == 0 {
		// Return all news
		for _, articles := range s.latestNews {
			allNews = append(allNews, articles...)
		}
	} else {
		// Return news for specific symbols
		for _, symbol := range symbols {
			if articles, exists := s.latestNews[symbol]; exists {
				allNews = append(allNews, articles...)
			}
		}
	}
	
	// Sort by published date (most recent first)
	for i := 0; i < len(allNews)-1; i++ {
		for j := i + 1; j < len(allNews); j++ {
			if allNews[i].PublishedAt.Before(allNews[j].PublishedAt) {
				allNews[i], allNews[j] = allNews[j], allNews[i]
			}
		}
	}
	
	// Apply limit
	if limit > 0 && len(allNews) > limit {
		allNews = allNews[:limit]
	}
	
	return allNews, nil
}

// GetSentiment returns sentiment analysis for a symbol
func (s *Service) GetSentiment(ctx context.Context, symbol string) (*SentimentAnalysis, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if sentiment, exists := s.latestSentiment[symbol]; exists {
		return sentiment, nil
	}
	
	return nil, fmt.Errorf("no sentiment data available for symbol %s", symbol)
}

// GetAllSentiment returns sentiment analysis for all symbols
func (s *Service) GetAllSentiment(ctx context.Context) (map[string]*SentimentAnalysis, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	result := make(map[string]*SentimentAnalysis)
	for symbol, sentiment := range s.latestSentiment {
		result[symbol] = sentiment
	}
	
	return result, nil
}

// GetMarketInsights returns latest market insights
func (s *Service) GetMarketInsights(ctx context.Context, limit int) ([]*MarketInsight, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	insights := make([]*MarketInsight, len(s.latestInsights))
	copy(insights, s.latestInsights)
	
	// Apply limit
	if limit > 0 && len(insights) > limit {
		insights = insights[:limit]
	}
	
	return insights, nil
}

// GetTrendingTopics returns current trending topics
func (s *Service) GetTrendingTopics(ctx context.Context, limit int) ([]*TrendingTopic, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	topics := make([]*TrendingTopic, len(s.trendingTopics))
	copy(topics, s.trendingTopics)
	
	// Apply limit
	if limit > 0 && len(topics) > limit {
		topics = topics[:limit]
	}
	
	return topics, nil
}

// GetQualityMetrics returns data quality metrics
func (s *Service) GetQualityMetrics(ctx context.Context) (map[string]*DataQualityMetrics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	result := make(map[string]*DataQualityMetrics)
	for source, metrics := range s.qualityMetrics {
		result[source] = metrics
	}
	
	return result, nil
}

// collectNews collects news articles using Bright Data
func (s *Service) collectNews(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.newsCollector.CollectNews(ctx); err != nil {
				s.logger.Errorf("Failed to collect news: %v", err)
			}
		}
	}
}

// analyzeSentiment analyzes social media sentiment
func (s *Service) analyzeSentiment(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.sentimentAnalyzer.AnalyzeSentiment(ctx); err != nil {
				s.logger.Errorf("Failed to analyze sentiment: %v", err)
			}
		}
	}
}

// collectMarketIntelligence collects market intelligence data
func (s *Service) collectMarketIntelligence(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.config.UpdateInterval * 2) // Less frequent updates
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.marketIntelligence.CollectIntelligence(ctx); err != nil {
				s.logger.Errorf("Failed to collect market intelligence: %v", err)
			}
		}
	}
}

// updateQualityMetrics updates data quality metrics
func (s *Service) updateQualityMetrics(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.calculateQualityMetrics()
		}
	}
}

// calculateQualityMetrics calculates quality metrics for all data sources
func (s *Service) calculateQualityMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	now := time.Now()
	
	// Calculate metrics for news sources
	for _, source := range s.config.NewsSources {
		if !source.Enabled {
			continue
		}
		
		metrics := &DataQualityMetrics{
			Source:          source.Name,
			LastUpdate:      now,
			UpdateFrequency: 1.0, // Default frequency
			SuccessRate:     0.95, // Default success rate
			AverageLatency:  2 * time.Second,
			ErrorCount:      0,
			DataFreshness:   time.Since(s.lastUpdate),
			QualityScore:    0.9, // Default quality score
		}
		
		s.qualityMetrics[source.Name] = metrics
	}
	
	// Calculate metrics for social sources
	for _, source := range s.config.SocialSources {
		if !source.Enabled {
			continue
		}
		
		metrics := &DataQualityMetrics{
			Source:          source.Platform,
			LastUpdate:      now,
			UpdateFrequency: 2.0, // Higher frequency for social
			SuccessRate:     0.90,
			AverageLatency:  3 * time.Second,
			ErrorCount:      0,
			DataFreshness:   time.Since(s.lastUpdate),
			QualityScore:    0.85,
		}
		
		s.qualityMetrics[source.Platform] = metrics
	}
}

// updateNews updates news data in memory and cache
func (s *Service) updateNews(symbol string, articles []*NewsArticle) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.latestNews[symbol] = articles
	s.lastUpdate = time.Now()
	
	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(articles)
		cacheKey := fmt.Sprintf("brightdata:news:%s", symbol)
		s.redis.Set(context.Background(), cacheKey, data, s.config.CacheTTL)
	}
}

// updateSentiment updates sentiment data in memory and cache
func (s *Service) updateSentiment(symbol string, sentiment *SentimentAnalysis) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.latestSentiment[symbol] = sentiment
	s.lastUpdate = time.Now()
	
	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(sentiment)
		cacheKey := fmt.Sprintf("brightdata:sentiment:%s", symbol)
		s.redis.Set(context.Background(), cacheKey, data, s.config.CacheTTL)
	}
}

// updateInsights updates market insights in memory and cache
func (s *Service) updateInsights(insights []*MarketInsight) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.latestInsights = insights
	s.lastUpdate = time.Now()
	
	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(insights)
		s.redis.Set(context.Background(), "brightdata:insights", data, s.config.CacheTTL)
	}
}

// SearchCryptoNews searches for crypto news using Bright Data
func (s *Service) SearchCryptoNews(ctx context.Context, query string) ([]*SearchResult, error) {
	// Use search_engine_Bright_Data to search for crypto news
	return s.newsCollector.SearchNews(ctx, query)
}

// ScrapeURL scrapes content from a specific URL
func (s *Service) ScrapeURL(ctx context.Context, url string) (*ScrapedContent, error) {
	// Use scrape_as_markdown_Bright_Data to scrape content
	return s.newsCollector.ScrapeContent(ctx, url)
}

// ScrapeTradingViewData scrapes crypto market data from TradingView
func (s *Service) ScrapeTradingViewData(ctx context.Context) (*TradingViewData, error) {
	// Scrape TradingView crypto market data
	return s.marketIntelligence.ScrapeTradingView(ctx)
}

// GetPortfolioAnalytics returns portfolio analytics and risk metrics
func (s *Service) GetPortfolioAnalytics(ctx context.Context, portfolioID string) (*PortfolioAnalytics, error) {
	return s.marketIntelligence.GetPortfolioAnalytics(ctx, portfolioID)
}

// GetMarketHeatmap returns market heatmap data
func (s *Service) GetMarketHeatmap(ctx context.Context) (*MarketHeatmap, error) {
	return s.marketIntelligence.GetMarketHeatmap(ctx)
}

// GetRiskMetrics returns comprehensive risk metrics
func (s *Service) GetRiskMetrics(ctx context.Context, portfolioID string) (*RiskMetrics, error) {
	return s.marketIntelligence.GetRiskMetrics(ctx, portfolioID)
}

// GetTradingSignals returns trading signals for symbols
func (s *Service) GetTradingSignals(ctx context.Context, symbols []string, limit int) ([]*TradingSignal, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var allSignals []*TradingSignal

	if len(symbols) == 0 {
		// Return all signals
		for _, signals := range s.tradingSignals {
			allSignals = append(allSignals, signals...)
		}
	} else {
		// Return signals for specific symbols
		for _, symbol := range symbols {
			if signals, exists := s.tradingSignals[symbol]; exists {
				allSignals = append(allSignals, signals...)
			}
		}
	}

	// Sort by confidence and created date
	for i := 0; i < len(allSignals)-1; i++ {
		for j := i + 1; j < len(allSignals); j++ {
			if allSignals[i].Confidence.LessThan(allSignals[j].Confidence) ||
				(allSignals[i].Confidence.Equal(allSignals[j].Confidence) &&
				 allSignals[i].CreatedAt.Before(allSignals[j].CreatedAt)) {
				allSignals[i], allSignals[j] = allSignals[j], allSignals[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(allSignals) > limit {
		allSignals = allSignals[:limit]
	}

	return allSignals, nil
}

// GetTradingBots returns top performing trading bots
func (s *Service) GetTradingBots(ctx context.Context, limit int) ([]*TradingBot, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bots := make([]*TradingBot, len(s.tradingBots))
	copy(bots, s.tradingBots)

	// Sort by profit percentage
	for i := 0; i < len(bots)-1; i++ {
		for j := i + 1; j < len(bots); j++ {
			if bots[i].TotalProfitPct.LessThan(bots[j].TotalProfitPct) {
				bots[i], bots[j] = bots[j], bots[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(bots) > limit {
		bots = bots[:limit]
	}

	return bots, nil
}

// GetTechnicalAnalysis returns technical analysis for symbols
func (s *Service) GetTechnicalAnalysis(ctx context.Context, symbols []string) (map[string]*TechnicalAnalysis, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*TechnicalAnalysis)

	if len(symbols) == 0 {
		// Return all technical analysis
		for symbol, analysis := range s.technicalAnalysis {
			result[symbol] = analysis
		}
	} else {
		// Return analysis for specific symbols
		for _, symbol := range symbols {
			if analysis, exists := s.technicalAnalysis[symbol]; exists {
				result[symbol] = analysis
			}
		}
	}

	return result, nil
}

// GetActiveDeals returns active trading deals
func (s *Service) GetActiveDeals(ctx context.Context, limit int) ([]*TradingDeal, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	deals := make([]*TradingDeal, 0)

	// Filter for active deals only
	for _, deal := range s.activeDeals {
		if deal.Status == "active" {
			deals = append(deals, deal)
		}
	}

	// Sort by unrealized PnL
	for i := 0; i < len(deals)-1; i++ {
		for j := i + 1; j < len(deals); j++ {
			if deals[i].UnrealizedPnL.LessThan(deals[j].UnrealizedPnL) {
				deals[i], deals[j] = deals[j], deals[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(deals) > limit {
		deals = deals[:limit]
	}

	return deals, nil
}

// ScrapeCommasData scrapes data from 3commas
func (s *Service) ScrapeCommasData(ctx context.Context) error {
	s.logger.Info("Scraping 3commas data")

	// Scrape trading bots
	bots, err := s.commasScraper.ScrapeTopBots(ctx)
	if err != nil {
		s.logger.Errorf("Failed to scrape 3commas bots: %v", err)
	} else {
		s.updateTradingBots(bots)
	}

	// Scrape trading signals
	signals, err := s.commasScraper.ScrapeTradingSignals(ctx)
	if err != nil {
		s.logger.Errorf("Failed to scrape 3commas signals: %v", err)
	} else {
		s.updateTradingSignals("3commas", signals)
	}

	// Scrape active deals
	deals, err := s.commasScraper.ScrapeActiveDeals(ctx)
	if err != nil {
		s.logger.Errorf("Failed to scrape 3commas deals: %v", err)
	} else {
		s.updateActiveDeals(deals)
	}

	return nil
}

// ScrapeTradingViewEnhanced scrapes enhanced TradingView data
func (s *Service) ScrapeTradingViewEnhanced(ctx context.Context, symbols []string) error {
	s.logger.Info("Scraping enhanced TradingView data")

	// Scrape technical analysis
	analysis, err := s.tradingViewEnhanced.ScrapeTechnicalAnalysis(ctx, symbols)
	if err != nil {
		s.logger.Errorf("Failed to scrape TradingView technical analysis: %v", err)
	} else {
		s.updateTechnicalAnalysis(analysis)
	}

	// Scrape trader ideas
	ideas, err := s.tradingViewEnhanced.ScrapeTraderIdeas(ctx, symbols)
	if err != nil {
		s.logger.Errorf("Failed to scrape TradingView ideas: %v", err)
	} else {
		s.updateTradingSignals("tradingview", ideas)
	}

	return nil
}

// ScrapeSocialMedia scrapes social media sentiment and trends
func (s *Service) ScrapeSocialMedia(ctx context.Context, symbols []string) error {
	s.logger.Info("Scraping social media data")

	// Scrape crypto sentiment
	sentiment, err := s.socialScraper.ScrapeCryptoSentiment(ctx, symbols)
	if err != nil {
		s.logger.Errorf("Failed to scrape social sentiment: %v", err)
	} else {
		s.updateSocialSentiment(sentiment)
	}

	// Scrape trending topics
	topics, err := s.socialScraper.ScrapeTrendingTopics(ctx)
	if err != nil {
		s.logger.Errorf("Failed to scrape trending topics: %v", err)
	} else {
		s.updateTrendingTopics(topics)
	}

	return nil
}

// updateTradingBots updates trading bots data in memory and cache
func (s *Service) updateTradingBots(bots []TradingBot) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tradingBots = make([]*TradingBot, len(bots))
	for i := range bots {
		s.tradingBots[i] = &bots[i]
	}
	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(s.tradingBots)
		s.redis.Set(context.Background(), "brightdata:trading_bots", data, s.config.CacheTTL)
	}
}

// updateTradingSignals updates trading signals data in memory and cache
func (s *Service) updateTradingSignals(source string, signals []TradingSignal) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Group signals by symbol
	signalsBySymbol := make(map[string][]*TradingSignal)
	for i := range signals {
		symbol := signals[i].Symbol
		signalsBySymbol[symbol] = append(signalsBySymbol[symbol], &signals[i])
	}

	// Update signals for each symbol
	for symbol, symbolSignals := range signalsBySymbol {
		key := fmt.Sprintf("%s_%s", source, symbol)
		s.tradingSignals[key] = symbolSignals
	}

	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(s.tradingSignals)
		cacheKey := fmt.Sprintf("brightdata:trading_signals:%s", source)
		s.redis.Set(context.Background(), cacheKey, data, s.config.CacheTTL)
	}
}

// updateTechnicalAnalysis updates technical analysis data in memory and cache
func (s *Service) updateTechnicalAnalysis(analysis map[string]TechnicalAnalysis) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for symbol, ta := range analysis {
		s.technicalAnalysis[symbol] = &ta
	}
	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(s.technicalAnalysis)
		s.redis.Set(context.Background(), "brightdata:technical_analysis", data, s.config.CacheTTL)
	}
}

// updateActiveDeals updates active deals data in memory and cache
func (s *Service) updateActiveDeals(deals []TradingDeal) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.activeDeals = make([]*TradingDeal, len(deals))
	for i := range deals {
		s.activeDeals[i] = &deals[i]
	}
	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(s.activeDeals)
		s.redis.Set(context.Background(), "brightdata:active_deals", data, s.config.CacheTTL)
	}
}

// updateSocialSentiment updates social sentiment data in memory and cache
func (s *Service) updateSocialSentiment(sentiment map[string]SentimentAnalysis) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for symbol, sa := range sentiment {
		s.latestSentiment[symbol] = &sa
	}
	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		for symbol, sa := range sentiment {
			data, _ := json.Marshal(sa)
			cacheKey := fmt.Sprintf("brightdata:social_sentiment:%s", symbol)
			s.redis.Set(context.Background(), cacheKey, data, s.config.CacheTTL)
		}
	}
}

// updateTrendingTopics updates trending topics data in memory and cache
func (s *Service) updateTrendingTopics(topics []TrendingTopic) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.trendingTopics = make([]*TrendingTopic, len(topics))
	for i := range topics {
		s.trendingTopics[i] = &topics[i]
	}
	s.lastUpdate = time.Now()

	// Cache in Redis
	if s.redis != nil {
		data, _ := json.Marshal(s.trendingTopics)
		s.redis.Set(context.Background(), "brightdata:trending_topics", data, s.config.CacheTTL)
	}
}

// collectCommasData collects data from 3commas periodically
func (s *Service) collectCommasData(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.UpdateInterval * 3) // Less frequent updates
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.ScrapeCommasData(ctx); err != nil {
				s.logger.Errorf("Failed to collect 3commas data: %v", err)
			}
		}
	}
}

// collectTradingViewEnhanced collects enhanced TradingView data periodically
func (s *Service) collectTradingViewEnhanced(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.UpdateInterval * 2) // Moderate frequency
	defer ticker.Stop()

	// Target symbols for analysis
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "SOLUSDT", "MATICUSDT"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.ScrapeTradingViewEnhanced(ctx, symbols); err != nil {
				s.logger.Errorf("Failed to collect enhanced TradingView data: %v", err)
			}
		}
	}
}

// collectSocialData collects social media data periodically
func (s *Service) collectSocialData(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.UpdateInterval) // Regular frequency
	defer ticker.Stop()

	// Target symbols for sentiment analysis
	symbols := []string{"BTC", "ETH", "ADA", "SOL", "MATIC", "DOGE", "SHIB"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.ScrapeSocialMedia(ctx, symbols); err != nil {
				s.logger.Errorf("Failed to collect social media data: %v", err)
			}
		}
	}
}
