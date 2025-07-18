package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/brightdata"
)

// TradingSignalsHandler handles trading signals API endpoints
type TradingSignalsHandler struct {
	brightDataService *brightdata.Service
	logger           *logrus.Logger
}

// NewTradingSignalsHandler creates a new trading signals handler
func NewTradingSignalsHandler(brightDataService *brightdata.Service, logger *logrus.Logger) *TradingSignalsHandler {
	return &TradingSignalsHandler{
		brightDataService: brightDataService,
		logger:           logger,
	}
}

// RegisterRoutes registers trading signals routes
func (h *TradingSignalsHandler) RegisterRoutes(router *gin.Engine) {
	// Trading Signals endpoints
	router.GET("/api/v2/trading/signals", h.GetTradingSignals)
	router.GET("/api/v2/trading/signals/:symbol", h.GetTradingSignalsBySymbol)
	router.GET("/api/v2/trading/signals/search", h.SearchTradingSignals)
	
	// Trading Bots endpoints
	router.GET("/api/v2/trading/bots", h.GetTradingBots)
	router.GET("/api/v2/trading/bots/top", h.GetTopTradingBots)
	
	// Technical Analysis endpoints
	router.GET("/api/v2/trading/analysis", h.GetTechnicalAnalysis)
	router.GET("/api/v2/trading/analysis/:symbol", h.GetTechnicalAnalysisBySymbol)
	
	// Active Deals endpoints
	router.GET("/api/v2/trading/deals", h.GetActiveDeals)
	router.GET("/api/v2/trading/deals/active", h.GetActiveDealsOnly)
	
	// 3commas specific endpoints
	router.GET("/api/v2/3commas/bots", h.GetCommasBots)
	router.GET("/api/v2/3commas/signals", h.GetCommasSignals)
	router.GET("/api/v2/3commas/deals", h.GetCommasDeals)
}

// GetTradingSignals returns all trading signals
func (h *TradingSignalsHandler) GetTradingSignals(c *gin.Context) {
	h.logger.Info("Getting all trading signals")
	
	// Parse query parameters
	limitStr := c.Query("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	symbolsStr := c.Query("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	
	signals, err := h.brightDataService.GetTradingSignals(c.Request.Context(), symbols, limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trading signals"})
		return
	}
	
	response := map[string]interface{}{
		"signals": signals,
		"count":   len(signals),
		"limit":   limit,
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTradingSignalsBySymbol returns trading signals for a specific symbol
func (h *TradingSignalsHandler) GetTradingSignalsBySymbol(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	
	h.logger.Infof("Getting trading signals for symbol: %s", symbol)
	
	limitStr := c.Query("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	signals, err := h.brightDataService.GetTradingSignals(c.Request.Context(), []string{symbol}, limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trading signals"})
		return
	}
	
	response := map[string]interface{}{
		"symbol":  symbol,
		"signals": signals,
		"count":   len(signals),
	}
	
	c.JSON(http.StatusOK, response)
}

// SearchTradingSignals searches trading signals by criteria
func (h *TradingSignalsHandler) SearchTradingSignals(c *gin.Context) {
	h.logger.Info("Searching trading signals")
	
	// Parse search parameters
	source := c.Query("source")
	signalType := c.Query("type")
	riskLevel := c.Query("risk_level")
	minConfidence := c.Query("min_confidence")
	
	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all signals first
	signals, err := h.brightDataService.GetTradingSignals(c.Request.Context(), nil, 0)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search trading signals"})
		return
	}
	
	// Filter signals based on criteria
	var filteredSignals []*brightdata.TradingSignal
	for _, signal := range signals {
		if source != "" && signal.Source != source {
			continue
		}
		if signalType != "" && signal.Type != signalType {
			continue
		}
		if riskLevel != "" && signal.RiskLevel != riskLevel {
			continue
		}
		if minConfidence != "" {
			if minConf, err := strconv.ParseFloat(minConfidence, 64); err == nil {
				if signal.Confidence.InexactFloat64() < minConf {
					continue
				}
			}
		}
		
		filteredSignals = append(filteredSignals, signal)
		if len(filteredSignals) >= limit {
			break
		}
	}
	
	response := map[string]interface{}{
		"signals": filteredSignals,
		"count":   len(filteredSignals),
		"filters": map[string]string{
			"source":         source,
			"type":           signalType,
			"risk_level":     riskLevel,
			"min_confidence": minConfidence,
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTradingBots returns all trading bots
func (h *TradingSignalsHandler) GetTradingBots(c *gin.Context) {
	h.logger.Info("Getting trading bots")
	
	limitStr := c.Query("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	bots, err := h.brightDataService.GetTradingBots(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading bots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trading bots"})
		return
	}
	
	response := map[string]interface{}{
		"bots":  bots,
		"count": len(bots),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTopTradingBots returns top performing trading bots
func (h *TradingSignalsHandler) GetTopTradingBots(c *gin.Context) {
	h.logger.Info("Getting top trading bots")
	
	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	bots, err := h.brightDataService.GetTradingBots(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get top trading bots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top trading bots"})
		return
	}
	
	response := map[string]interface{}{
		"top_bots": bots,
		"count":    len(bots),
		"criteria": "sorted by profit percentage",
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTechnicalAnalysis returns technical analysis for all symbols
func (h *TradingSignalsHandler) GetTechnicalAnalysis(c *gin.Context) {
	h.logger.Info("Getting technical analysis")
	
	symbolsStr := c.Query("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	
	analysis, err := h.brightDataService.GetTechnicalAnalysis(c.Request.Context(), symbols)
	if err != nil {
		h.logger.Errorf("Failed to get technical analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get technical analysis"})
		return
	}
	
	response := map[string]interface{}{
		"analysis": analysis,
		"count":    len(analysis),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetTechnicalAnalysisBySymbol returns technical analysis for a specific symbol
func (h *TradingSignalsHandler) GetTechnicalAnalysisBySymbol(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	
	h.logger.Infof("Getting technical analysis for symbol: %s", symbol)
	
	analysis, err := h.brightDataService.GetTechnicalAnalysis(c.Request.Context(), []string{symbol})
	if err != nil {
		h.logger.Errorf("Failed to get technical analysis for %s: %v", symbol, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get technical analysis"})
		return
	}
	
	if len(analysis) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No technical analysis found for symbol"})
		return
	}
	
	// Get the analysis for the symbol (there might be multiple timeframes)
	var symbolAnalysis []*brightdata.TechnicalAnalysis
	for key, ta := range analysis {
		if strings.Contains(key, symbol) {
			symbolAnalysis = append(symbolAnalysis, ta)
		}
	}
	
	response := map[string]interface{}{
		"symbol":   symbol,
		"analysis": symbolAnalysis,
		"count":    len(symbolAnalysis),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetActiveDeals returns all active trading deals
func (h *TradingSignalsHandler) GetActiveDeals(c *gin.Context) {
	h.logger.Info("Getting active trading deals")
	
	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	deals, err := h.brightDataService.GetActiveDeals(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get active deals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active deals"})
		return
	}
	
	response := map[string]interface{}{
		"deals": deals,
		"count": len(deals),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetActiveDealsOnly returns only active deals (status = "active")
func (h *TradingSignalsHandler) GetActiveDealsOnly(c *gin.Context) {
	h.logger.Info("Getting active deals only")
	
	limitStr := c.Query("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	deals, err := h.brightDataService.GetActiveDeals(c.Request.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get active deals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active deals"})
		return
	}
	
	response := map[string]interface{}{
		"active_deals": deals,
		"count":        len(deals),
		"status":       "active_only",
	}
	
	c.JSON(http.StatusOK, response)
}

// GetCommasBots returns 3commas specific bots
func (h *TradingSignalsHandler) GetCommasBots(c *gin.Context) {
	h.logger.Info("Getting 3commas bots")
	
	limitStr := c.Query("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all bots and filter for 3commas
	allBots, err := h.brightDataService.GetTradingBots(c.Request.Context(), 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas bots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get 3commas bots"})
		return
	}
	
	var commasBots []*brightdata.TradingBot
	for _, bot := range allBots {
		if strings.Contains(strings.ToLower(bot.ID), "3commas") {
			commasBots = append(commasBots, bot)
			if len(commasBots) >= limit {
				break
			}
		}
	}
	
	response := map[string]interface{}{
		"bots":   commasBots,
		"count":  len(commasBots),
		"source": "3commas",
	}
	
	c.JSON(http.StatusOK, response)
}

// GetCommasSignals returns 3commas specific signals
func (h *TradingSignalsHandler) GetCommasSignals(c *gin.Context) {
	h.logger.Info("Getting 3commas signals")
	
	limitStr := c.Query("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all signals and filter for 3commas
	allSignals, err := h.brightDataService.GetTradingSignals(c.Request.Context(), nil, 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas signals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get 3commas signals"})
		return
	}
	
	var commasSignals []*brightdata.TradingSignal
	for _, signal := range allSignals {
		if signal.Source == "3commas" {
			commasSignals = append(commasSignals, signal)
			if len(commasSignals) >= limit {
				break
			}
		}
	}
	
	response := map[string]interface{}{
		"signals": commasSignals,
		"count":   len(commasSignals),
		"source":  "3commas",
	}
	
	c.JSON(http.StatusOK, response)
}

// GetCommasDeals returns 3commas specific deals
func (h *TradingSignalsHandler) GetCommasDeals(c *gin.Context) {
	h.logger.Info("Getting 3commas deals")
	
	limitStr := c.Query("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all deals and filter for 3commas
	allDeals, err := h.brightDataService.GetActiveDeals(c.Request.Context(), 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas deals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get 3commas deals"})
		return
	}
	
	var commasDeals []*brightdata.TradingDeal
	for _, deal := range allDeals {
		if strings.Contains(strings.ToLower(deal.ID), "3commas") {
			commasDeals = append(commasDeals, deal)
			if len(commasDeals) >= limit {
				break
			}
		}
	}
	
	response := map[string]interface{}{
		"deals":  commasDeals,
		"count":  len(commasDeals),
		"source": "3commas",
	}
	
	c.JSON(http.StatusOK, response)
}
