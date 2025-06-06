package api

import (
	"net/http"
	"strconv"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/brightdata"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TradingViewHandlers handles TradingView-related API endpoints
type TradingViewHandlers struct {
	brightDataService *brightdata.Service
	logger            *logrus.Logger
}

// NewTradingViewHandlers creates new TradingView handlers
func NewTradingViewHandlers(brightDataService *brightdata.Service, logger *logrus.Logger) *TradingViewHandlers {
	return &TradingViewHandlers{
		brightDataService: brightDataService,
		logger:            logger,
	}
}

// RegisterRoutes registers TradingView API routes
func (h *TradingViewHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// TradingView data endpoints
	router.GET("/tradingview/market-data", h.GetTradingViewData)
	router.GET("/tradingview/coins", h.GetTradingViewCoins)
	router.GET("/tradingview/trending", h.GetTrendingCoins)
	router.GET("/tradingview/gainers", h.GetTopGainers)
	router.GET("/tradingview/losers", h.GetTopLosers)
	router.GET("/tradingview/market-overview", h.GetMarketOverview)

	// Portfolio analytics endpoints
	router.GET("/portfolio/:portfolioId/analytics", h.GetPortfolioAnalytics)
	router.GET("/portfolio/:portfolioId/risk-metrics", h.GetRiskMetrics)
	router.GET("/portfolio/:portfolioId/performance", h.GetPerformanceMetrics)

	// Market visualization endpoints
	router.GET("/market/heatmap", h.GetMarketHeatmap)
	router.GET("/market/sectors", h.GetSectorData)
	router.GET("/market/correlation", h.GetCorrelationMatrix)
}

// GetTradingViewData gets comprehensive TradingView market data
func (h *TradingViewHandlers) GetTradingViewData(c *gin.Context) {
	h.logger.Info("Getting TradingView market data")

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get TradingView data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get TradingView data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetTradingViewCoins gets cryptocurrency data from TradingView
func (h *TradingViewHandlers) GetTradingViewCoins(c *gin.Context) {
	h.logger.Info("Getting TradingView coins data")

	// Parse query parameters
	limitStr := c.Query("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get TradingView coins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get TradingView coins",
		})
		return
	}

	// Apply limit
	coins := data.Coins
	if len(coins) > limit {
		coins = coins[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"coins":        coins,
			"total_count":  len(data.Coins),
			"last_updated": data.LastUpdated,
		},
	})
}

// GetTrendingCoins gets trending cryptocurrencies
func (h *TradingViewHandlers) GetTrendingCoins(c *gin.Context) {
	h.logger.Info("Getting trending coins")

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get trending coins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get trending coins",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trending_coins": data.TrendingCoins,
			"last_updated":   data.LastUpdated,
		},
	})
}

// GetTopGainers gets top gaining cryptocurrencies
func (h *TradingViewHandlers) GetTopGainers(c *gin.Context) {
	h.logger.Info("Getting top gainers")

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get top gainers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get top gainers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"gainers":      data.Gainers,
			"last_updated": data.LastUpdated,
		},
	})
}

// GetTopLosers gets top losing cryptocurrencies
func (h *TradingViewHandlers) GetTopLosers(c *gin.Context) {
	h.logger.Info("Getting top losers")

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get top losers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get top losers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"losers":       data.Losers,
			"last_updated": data.LastUpdated,
		},
	})
}

// GetMarketOverview gets market overview data
func (h *TradingViewHandlers) GetMarketOverview(c *gin.Context) {
	h.logger.Info("Getting market overview")

	data, err := h.brightDataService.ScrapeTradingViewData(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get market overview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get market overview",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"market_overview": data.MarketOverview,
			"market_cap":      data.MarketCap,
			"last_updated":    data.LastUpdated,
		},
	})
}

// GetPortfolioAnalytics gets comprehensive portfolio analytics
func (h *TradingViewHandlers) GetPortfolioAnalytics(c *gin.Context) {
	portfolioID := c.Param("portfolioId")

	h.logger.Infof("Getting portfolio analytics for portfolio: %s", portfolioID)

	analytics, err := h.brightDataService.GetPortfolioAnalytics(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Errorf("Failed to get portfolio analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get portfolio analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetRiskMetrics gets portfolio risk metrics
func (h *TradingViewHandlers) GetRiskMetrics(c *gin.Context) {
	portfolioID := c.Param("portfolioId")

	h.logger.Infof("Getting risk metrics for portfolio: %s", portfolioID)

	riskMetrics, err := h.brightDataService.GetRiskMetrics(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Errorf("Failed to get risk metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get risk metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    riskMetrics,
	})
}

// GetPerformanceMetrics gets portfolio performance metrics
func (h *TradingViewHandlers) GetPerformanceMetrics(c *gin.Context) {
	portfolioID := c.Param("portfolioId")

	h.logger.Infof("Getting performance metrics for portfolio: %s", portfolioID)

	analytics, err := h.brightDataService.GetPortfolioAnalytics(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Errorf("Failed to get performance metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get performance metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"performance":      analytics.Performance,
			"total_return":     analytics.TotalReturn,
			"total_return_pct": analytics.TotalReturnPct,
			"day_return":       analytics.DayReturn,
			"day_return_pct":   analytics.DayReturnPct,
			"last_updated":     analytics.LastUpdated,
		},
	})
}

// GetMarketHeatmap gets market heatmap data
func (h *TradingViewHandlers) GetMarketHeatmap(c *gin.Context) {
	h.logger.Info("Getting market heatmap")

	heatmap, err := h.brightDataService.GetMarketHeatmap(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get market heatmap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get market heatmap",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    heatmap,
	})
}

// GetSectorData gets market sector data
func (h *TradingViewHandlers) GetSectorData(c *gin.Context) {
	h.logger.Info("Getting sector data")

	heatmap, err := h.brightDataService.GetMarketHeatmap(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get sector data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get sector data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sectors":      heatmap.Sectors,
			"last_updated": heatmap.LastUpdated,
		},
	})
}

// GetCorrelationMatrix gets asset correlation matrix
func (h *TradingViewHandlers) GetCorrelationMatrix(c *gin.Context) {
	h.logger.Info("Getting correlation matrix")

	// For now, return a sample correlation matrix
	// In a real implementation, this would calculate correlations from historical price data
	correlationMatrix := map[string]any{
		"BTC-ETH": 0.75,
		"BTC-BNB": 0.68,
		"BTC-SOL": 0.72,
		"BTC-ADA": 0.65,
		"ETH-BNB": 0.82,
		"ETH-SOL": 0.78,
		"ETH-ADA": 0.71,
		"BNB-SOL": 0.69,
		"BNB-ADA": 0.63,
		"SOL-ADA": 0.67,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"correlation_matrix": correlationMatrix,
			"last_updated":       "2024-01-15T10:30:00Z",
		},
	})
}
