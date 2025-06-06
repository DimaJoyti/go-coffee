package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/exchanges"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/market"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// EnhancedHandlers provides enhanced API endpoints for crypto data
type EnhancedHandlers struct {
	marketService *market.EnhancedService
	logger        *logrus.Logger
}

// NewEnhancedHandlers creates new enhanced API handlers
func NewEnhancedHandlers(marketService *market.EnhancedService, logger *logrus.Logger) *EnhancedHandlers {
	return &EnhancedHandlers{
		marketService: marketService,
		logger:        logger,
	}
}

// RegisterRoutes registers enhanced API routes
func (h *EnhancedHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// Enhanced market data routes
	v2 := router.Group("/v2")
	{
		// Multi-exchange aggregated data
		market := v2.Group("/market")
		{
			market.GET("/aggregated/:symbol", h.GetAggregatedPrice)
			market.GET("/best-prices/:symbol", h.GetBestPrices)
			market.GET("/summary/:symbol", h.GetMarketSummary)
			market.GET("/orderbook/:symbol", h.GetAggregatedOrderBook)
			market.GET("/exchanges/status", h.GetExchangeStatus)
			market.GET("/data-quality", h.GetDataQuality)
		}
		
		// Arbitrage opportunities
		arbitrage := v2.Group("/arbitrage")
		{
			arbitrage.GET("/opportunities", h.GetArbitrageOpportunities)
			arbitrage.GET("/opportunities/:symbol", h.GetArbitrageOpportunity)
		}
		
		// Analytics endpoints
		analytics := v2.Group("/analytics")
		{
			analytics.GET("/volume/:symbol", h.GetVolumeAnalytics)
			analytics.GET("/spread/:symbol", h.GetSpreadAnalytics)
			analytics.GET("/liquidity/:symbol", h.GetLiquidityAnalytics)
		}
		
		// Real-time streaming endpoints
		stream := v2.Group("/stream")
		{
			stream.GET("/prices", h.StreamPrices)
			stream.GET("/arbitrage", h.StreamArbitrage)
			stream.GET("/market-summary", h.StreamMarketSummary)
		}
	}
}

// GetAggregatedPrice returns aggregated price data from multiple exchanges
func (h *EnhancedHandlers) GetAggregatedPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	price, err := h.marketService.GetAggregatedPrice(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get aggregated price for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      price,
		Timestamp: time.Now(),
	})
}

// GetBestPrices returns the best bid and ask prices across exchanges
func (h *EnhancedHandlers) GetBestPrices(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	summary, err := h.marketService.GetBestPrices(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get best prices for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	response := map[string]interface{}{
		"symbol":          summary.Symbol,
		"best_bid":        summary.BestBid,
		"best_ask":        summary.BestAsk,
		"weighted_price":  summary.WeightedPrice,
		"price_spread":    summary.PriceSpread,
		"spread_percent":  summary.SpreadPercent,
		"total_volume":    summary.TotalVolume24h,
		"data_quality":    summary.DataQuality,
		"timestamp":       summary.Timestamp,
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		Timestamp: time.Now(),
	})
}

// GetMarketSummary returns comprehensive market summary
func (h *EnhancedHandlers) GetMarketSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	summary, err := h.marketService.GetMarketSummary(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get market summary for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      summary,
		Timestamp: time.Now(),
	})
}

// GetAggregatedOrderBook returns aggregated order book from multiple exchanges
func (h *EnhancedHandlers) GetAggregatedOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	depthStr := c.DefaultQuery("depth", "20")
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		depth = 20
	}
	
	orderBook, err := h.marketService.GetAggregatedOrderBook(c.Request.Context(), symbol, depth)
	if err != nil {
		h.logger.Errorf("Failed to get aggregated order book for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      orderBook,
		Timestamp: time.Now(),
	})
}

// GetExchangeStatus returns the status of all connected exchanges
func (h *EnhancedHandlers) GetExchangeStatus(c *gin.Context) {
	status, err := h.marketService.GetExchangeStatus(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get exchange status: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      status,
		Timestamp: time.Now(),
	})
}

// GetDataQuality returns data quality metrics for exchanges
func (h *EnhancedHandlers) GetDataQuality(c *gin.Context) {
	quality, err := h.marketService.GetDataQuality(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get data quality metrics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      quality,
		Timestamp: time.Now(),
	})
}

// GetArbitrageOpportunities returns current arbitrage opportunities
func (h *EnhancedHandlers) GetArbitrageOpportunities(c *gin.Context) {
	opportunities, err := h.marketService.GetArbitrageOpportunities(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get arbitrage opportunities: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	// Filter by minimum profit if specified
	minProfitStr := c.Query("min_profit")
	if minProfitStr != "" {
		if minProfit, err := strconv.ParseFloat(minProfitStr, 64); err == nil {
			filtered := make([]*exchanges.ArbitrageOpportunity, 0)
			for _, opp := range opportunities {
				if profitFloat, _ := opp.ProfitPercent.Float64(); profitFloat >= minProfit {
					filtered = append(filtered, opp)
				}
			}
			opportunities = filtered
		}
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      opportunities,
		Timestamp: time.Now(),
	})
}

// GetArbitrageOpportunity returns arbitrage opportunity for a specific symbol
func (h *EnhancedHandlers) GetArbitrageOpportunity(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}
	
	opportunities, err := h.marketService.GetArbitrageOpportunities(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get arbitrage opportunities: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	
	// Find opportunity for the specified symbol
	var symbolOpportunity *exchanges.ArbitrageOpportunity
	for _, opp := range opportunities {
		if opp.Symbol == symbol {
			symbolOpportunity = opp
			break
		}
	}
	
	if symbolOpportunity == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success:   false,
			Error:     "No arbitrage opportunity found for symbol",
			Timestamp: time.Now(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      symbolOpportunity,
		Timestamp: time.Now(),
	})
}

// GetVolumeAnalytics returns volume analytics for a symbol
func (h *EnhancedHandlers) GetVolumeAnalytics(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}

	summary, err := h.marketService.GetMarketSummary(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get market summary for volume analytics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// Calculate volume analytics
	analytics := map[string]interface{}{
		"symbol":              summary.Symbol,
		"total_volume_24h":    summary.TotalVolume24h,
		"exchange_volumes":    make(map[string]interface{}),
		"volume_distribution": make(map[string]float64),
		"timestamp":           summary.Timestamp,
	}

	// Calculate volume distribution across exchanges
	totalVolume := summary.TotalVolume24h
	exchangeVolumes := make(map[string]interface{})
	volumeDistribution := make(map[string]float64)

	for exchange, ticker := range summary.ExchangePrices {
		exchangeVolumes[string(exchange)] = map[string]interface{}{
			"volume_24h":       ticker.Volume24h,
			"volume_quote_24h": ticker.VolumeQuote24h,
			"last_update":      ticker.Timestamp,
		}

		if !totalVolume.IsZero() {
			percentage, _ := ticker.Volume24h.Div(totalVolume).Mul(decimal.NewFromInt(100)).Float64()
			volumeDistribution[string(exchange)] = percentage
		}
	}

	analytics["exchange_volumes"] = exchangeVolumes
	analytics["volume_distribution"] = volumeDistribution

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      analytics,
		Timestamp: time.Now(),
	})
}

// GetSpreadAnalytics returns spread analytics for a symbol
func (h *EnhancedHandlers) GetSpreadAnalytics(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}

	summary, err := h.marketService.GetMarketSummary(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Errorf("Failed to get market summary for spread analytics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// Calculate spread analytics
	analytics := map[string]interface{}{
		"symbol":           summary.Symbol,
		"best_bid":         summary.BestBid,
		"best_ask":         summary.BestAsk,
		"price_spread":     summary.PriceSpread,
		"spread_percent":   summary.SpreadPercent,
		"exchange_spreads": make(map[string]interface{}),
		"timestamp":        summary.Timestamp,
	}

	// Calculate spreads for each exchange
	exchangeSpreads := make(map[string]interface{})
	for exchange, ticker := range summary.ExchangePrices {
		spread := ticker.AskPrice.Sub(ticker.BidPrice)
		spreadPercent := decimal.Zero
		if !ticker.BidPrice.IsZero() {
			midPrice := ticker.BidPrice.Add(ticker.AskPrice).Div(decimal.NewFromInt(2))
			if !midPrice.IsZero() {
				spreadPercent = spread.Div(midPrice).Mul(decimal.NewFromInt(100))
			}
		}

		exchangeSpreads[string(exchange)] = map[string]interface{}{
			"bid_price":       ticker.BidPrice,
			"ask_price":       ticker.AskPrice,
			"spread":          spread,
			"spread_percent":  spreadPercent,
			"last_update":     ticker.Timestamp,
		}
	}

	analytics["exchange_spreads"] = exchangeSpreads

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      analytics,
		Timestamp: time.Now(),
	})
}

// GetLiquidityAnalytics returns liquidity analytics for a symbol
func (h *EnhancedHandlers) GetLiquidityAnalytics(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     "Symbol parameter is required",
			Timestamp: time.Now(),
		})
		return
	}

	depthStr := c.DefaultQuery("depth", "20")
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		depth = 20
	}

	orderBook, err := h.marketService.GetAggregatedOrderBook(c.Request.Context(), symbol, depth)
	if err != nil {
		h.logger.Errorf("Failed to get order book for liquidity analytics: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// Calculate liquidity metrics
	bidLiquidity := decimal.Zero
	askLiquidity := decimal.Zero

	for _, bid := range orderBook.Bids {
		bidLiquidity = bidLiquidity.Add(bid.Price.Mul(bid.Quantity))
	}

	for _, ask := range orderBook.Asks {
		askLiquidity = askLiquidity.Add(ask.Price.Mul(ask.Quantity))
	}

	totalLiquidity := bidLiquidity.Add(askLiquidity)

	analytics := map[string]interface{}{
		"symbol":          orderBook.Symbol,
		"bid_liquidity":   bidLiquidity,
		"ask_liquidity":   askLiquidity,
		"total_liquidity": totalLiquidity,
		"bid_levels":      len(orderBook.Bids),
		"ask_levels":      len(orderBook.Asks),
		"depth_analyzed":  depth,
		"timestamp":       orderBook.Timestamp,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      analytics,
		Timestamp: time.Now(),
	})
}

// StreamPrices provides real-time price streaming via WebSocket
func (h *EnhancedHandlers) StreamPrices(c *gin.Context) {
	// For now, return a placeholder response
	// In a full implementation, this would upgrade to WebSocket
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "WebSocket streaming endpoint for real-time prices",
			"upgrade": "Use WebSocket connection for real-time data",
		},
		Timestamp: time.Now(),
	})
}

// StreamArbitrage provides real-time arbitrage opportunity streaming
func (h *EnhancedHandlers) StreamArbitrage(c *gin.Context) {
	// For now, return a placeholder response
	// In a full implementation, this would upgrade to WebSocket
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "WebSocket streaming endpoint for real-time arbitrage opportunities",
			"upgrade": "Use WebSocket connection for real-time data",
		},
		Timestamp: time.Now(),
	})
}

// StreamMarketSummary provides real-time market summary streaming
func (h *EnhancedHandlers) StreamMarketSummary(c *gin.Context) {
	// For now, return a placeholder response
	// In a full implementation, this would upgrade to WebSocket
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "WebSocket streaming endpoint for real-time market summaries",
			"upgrade": "Use WebSocket connection for real-time data",
		},
		Timestamp: time.Now(),
	})
}
