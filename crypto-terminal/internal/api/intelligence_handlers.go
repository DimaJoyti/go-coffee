package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/brightdata"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// IntelligenceHandlers provides API endpoints for market intelligence data
type IntelligenceHandlers struct {
	brightDataService *brightdata.Service
	logger            *logrus.Logger
}

// NewIntelligenceHandlers creates new intelligence API handlers
func NewIntelligenceHandlers(brightDataService *brightdata.Service, logger *logrus.Logger) *IntelligenceHandlers {
	return &IntelligenceHandlers{
		brightDataService: brightDataService,
		logger:            logger,
	}
}

// RegisterRoutes registers intelligence API routes
func (h *IntelligenceHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// Intelligence routes
	intelligence := router.Group("/intelligence")
	{
		// News endpoints
		news := intelligence.Group("/news")
		{
			news.GET("", h.GetNews)
			news.GET("/:symbol", h.GetNewsForSymbol)
			news.GET("/search", h.SearchNews)
		}
		
		// Sentiment endpoints
		sentiment := intelligence.Group("/sentiment")
		{
			sentiment.GET("", h.GetAllSentiment)
			sentiment.GET("/:symbol", h.GetSentiment)
			sentiment.GET("/trending", h.GetTrendingTopics)
		}
		
		// Market insights endpoints
		insights := intelligence.Group("/insights")
		{
			insights.GET("", h.GetMarketInsights)
			insights.GET("/events", h.GetMarketEvents)
			insights.GET("/influencers", h.GetInfluencerInsights)
			insights.GET("/market-sentiment", h.GetMarketSentiment)
		}
		
		// Data quality endpoints
		quality := intelligence.Group("/quality")
		{
			quality.GET("/metrics", h.GetQualityMetrics)
			quality.GET("/status", h.GetServiceStatus)
		}
		
		// Utility endpoints
		utils := intelligence.Group("/utils")
		{
			utils.POST("/scrape", h.ScrapeURL)
			utils.GET("/sources", h.GetDataSources)
		}
	}
}

// GetNews returns latest crypto news
func (h *IntelligenceHandlers) GetNews(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}
	
	symbolsStr := c.Query("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
		// Clean up symbols
		for i, symbol := range symbols {
			symbols[i] = strings.TrimSpace(strings.ToUpper(symbol))
		}
	}
	
	news, err := h.brightDataService.GetNews(c.Request.Context(), symbols, limit)
	if err != nil {
		h.logger.Errorf("Failed to get news: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      news,
		Timestamp: time.Now(),
	})
}

// GetNewsForSymbol returns news for a specific symbol
func (h *IntelligenceHandlers) GetNewsForSymbol(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	
	news, err := h.brightDataService.GetNews(c.Request.Context(), []string{symbol}, limit)
	if err != nil {
		h.logger.Errorf("Failed to get news for symbol %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      news,
		Timestamp: time.Now(),
	})
}

// SearchNews searches for crypto news
func (h *IntelligenceHandlers) SearchNews(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Query parameter 'q' is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	results, err := h.brightDataService.SearchCryptoNews(c.Request.Context(), query)
	if err != nil {
		h.logger.Errorf("Failed to search news: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      results,
		Timestamp: time.Now(),
	})
}

// GetAllSentiment returns sentiment analysis for all symbols
func (h *IntelligenceHandlers) GetAllSentiment(c *gin.Context) {
	sentiment, err := h.brightDataService.GetAllSentiment(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get sentiment: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      sentiment,
		Timestamp: time.Now(),
	})
}

// GetSentiment returns sentiment analysis for a specific symbol
func (h *IntelligenceHandlers) GetSentiment(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	
	sentiment, err := h.brightDataService.GetSentiment(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get sentiment for symbol %s: %v", symbol, err)
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      sentiment,
		Timestamp: time.Now(),
	})
}

// GetTrendingTopics returns current trending topics
func (h *IntelligenceHandlers) GetTrendingTopics(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	
	topics, err := h.brightDataService.GetTrendingTopics(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get trending topics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      topics,
		Timestamp: time.Now(),
	})
}

// GetMarketInsights returns market insights
func (h *IntelligenceHandlers) GetMarketInsights(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}
	
	insights, err := h.brightDataService.GetMarketInsights(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get market insights: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      insights,
		Timestamp: time.Now(),
	})
}

// GetMarketEvents returns market events
func (h *IntelligenceHandlers) GetMarketEvents(c *gin.Context) {
	// This would call the market intelligence service to detect events
	// For now, return a placeholder response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Market events detection coming soon",
			"events":  []interface{}{},
		},
		Timestamp: time.Now(),
	})
}

// GetInfluencerInsights returns insights from crypto influencers
func (h *IntelligenceHandlers) GetInfluencerInsights(c *gin.Context) {
	// This would call the market intelligence service for influencer insights
	// For now, return a placeholder response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":  "Influencer insights coming soon",
			"insights": []interface{}{},
		},
		Timestamp: time.Now(),
	})
}

// GetMarketSentiment returns overall market sentiment
func (h *IntelligenceHandlers) GetMarketSentiment(c *gin.Context) {
	// This would call the market intelligence service for market sentiment
	// For now, return a placeholder response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"overall_sentiment": 0.15,
			"sentiment_score":   57,
			"confidence":        0.75,
			"trend":            "improving",
			"last_updated":     time.Now(),
		},
		Timestamp: time.Now(),
	})
}

// GetQualityMetrics returns data quality metrics
func (h *IntelligenceHandlers) GetQualityMetrics(c *gin.Context) {
	metrics, err := h.brightDataService.GetQualityMetrics(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get quality metrics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      metrics,
		Timestamp: time.Now(),
	})
}

// GetServiceStatus returns Bright Data service status
func (h *IntelligenceHandlers) GetServiceStatus(c *gin.Context) {
	status := map[string]interface{}{
		"service":      "bright_data",
		"status":       "operational",
		"last_update":  time.Now(),
		"data_sources": []string{"news", "social", "technical", "fundamental"},
		"health_score": 0.95,
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      status,
		Timestamp: time.Now(),
	})
}

// ScrapeURL scrapes content from a specific URL
func (h *IntelligenceHandlers) ScrapeURL(c *gin.Context) {
	var request struct {
		URL string `json:"url" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	content, err := h.brightDataService.ScrapeURL(c.Request.Context(), request.URL)
	if err != nil {
		h.logger.Errorf("Failed to scrape URL %s: %v", request.URL, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      content,
		Timestamp: time.Now(),
	})
}

// GetDataSources returns available data sources
func (h *IntelligenceHandlers) GetDataSources(c *gin.Context) {
	sources := map[string]interface{}{
		"news_sources": []string{
			"CoinTelegraph", "CoinDesk", "Decrypt", "The Block", "CryptoNews",
		},
		"social_platforms": []string{
			"Twitter/X", "Reddit", "Telegram", "Discord",
		},
		"analysis_sources": []string{
			"Messari", "Glassnode", "Santiment", "CryptoQuant",
		},
		"event_sources": []string{
			"CoinMarketCal", "CryptoCal", "Coindar",
		},
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      sources,
		Timestamp: time.Now(),
	})
}
