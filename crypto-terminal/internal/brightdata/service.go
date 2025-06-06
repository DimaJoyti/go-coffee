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
	}
	
	// Initialize sub-services
	service.newsCollector = NewNewsCollector(service, logger)
	service.sentimentAnalyzer = NewSentimentAnalyzer(service, logger)
	service.marketIntelligence = NewMarketIntelligence(service, logger)
	
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
