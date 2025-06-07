package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"crypto-terminal/internal/brightdata"
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
func (h *TradingSignalsHandler) RegisterRoutes(router *mux.Router) {
	// Trading Signals endpoints
	router.HandleFunc("/api/v2/trading/signals", h.GetTradingSignals).Methods("GET")
	router.HandleFunc("/api/v2/trading/signals/{symbol}", h.GetTradingSignalsBySymbol).Methods("GET")
	router.HandleFunc("/api/v2/trading/signals/search", h.SearchTradingSignals).Methods("GET")
	
	// Trading Bots endpoints
	router.HandleFunc("/api/v2/trading/bots", h.GetTradingBots).Methods("GET")
	router.HandleFunc("/api/v2/trading/bots/top", h.GetTopTradingBots).Methods("GET")
	
	// Technical Analysis endpoints
	router.HandleFunc("/api/v2/trading/analysis", h.GetTechnicalAnalysis).Methods("GET")
	router.HandleFunc("/api/v2/trading/analysis/{symbol}", h.GetTechnicalAnalysisBySymbol).Methods("GET")
	
	// Active Deals endpoints
	router.HandleFunc("/api/v2/trading/deals", h.GetActiveDeals).Methods("GET")
	router.HandleFunc("/api/v2/trading/deals/active", h.GetActiveDealsOnly).Methods("GET")
	
	// 3commas specific endpoints
	router.HandleFunc("/api/v2/3commas/bots", h.GetCommasBots).Methods("GET")
	router.HandleFunc("/api/v2/3commas/signals", h.GetCommasSignals).Methods("GET")
	router.HandleFunc("/api/v2/3commas/deals", h.GetCommasDeals).Methods("GET")
}

// GetTradingSignals returns all trading signals
func (h *TradingSignalsHandler) GetTradingSignals(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting all trading signals")
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	symbolsStr := r.URL.Query().Get("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	
	signals, err := h.brightDataService.GetTradingSignals(r.Context(), symbols, limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals: %v", err)
		http.Error(w, "Failed to get trading signals", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"signals": signals,
		"count":   len(signals),
		"limit":   limit,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTradingSignalsBySymbol returns trading signals for a specific symbol
func (h *TradingSignalsHandler) GetTradingSignalsBySymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := strings.ToUpper(vars["symbol"])
	
	h.logger.Infof("Getting trading signals for symbol: %s", symbol)
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	signals, err := h.brightDataService.GetTradingSignals(r.Context(), []string{symbol}, limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals for %s: %v", symbol, err)
		http.Error(w, "Failed to get trading signals", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"symbol":  symbol,
		"signals": signals,
		"count":   len(signals),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchTradingSignals searches trading signals by criteria
func (h *TradingSignalsHandler) SearchTradingSignals(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Searching trading signals")
	
	// Parse search parameters
	source := r.URL.Query().Get("source")
	signalType := r.URL.Query().Get("type")
	riskLevel := r.URL.Query().Get("risk_level")
	minConfidence := r.URL.Query().Get("min_confidence")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all signals first
	signals, err := h.brightDataService.GetTradingSignals(r.Context(), nil, 0)
	if err != nil {
		h.logger.Errorf("Failed to get trading signals: %v", err)
		http.Error(w, "Failed to search trading signals", http.StatusInternalServerError)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTradingBots returns all trading bots
func (h *TradingSignalsHandler) GetTradingBots(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting trading bots")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	bots, err := h.brightDataService.GetTradingBots(r.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get trading bots: %v", err)
		http.Error(w, "Failed to get trading bots", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"bots":  bots,
		"count": len(bots),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTopTradingBots returns top performing trading bots
func (h *TradingSignalsHandler) GetTopTradingBots(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting top trading bots")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	bots, err := h.brightDataService.GetTradingBots(r.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get top trading bots: %v", err)
		http.Error(w, "Failed to get top trading bots", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"top_bots": bots,
		"count":    len(bots),
		"criteria": "sorted by profit percentage",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnicalAnalysis returns technical analysis for all symbols
func (h *TradingSignalsHandler) GetTechnicalAnalysis(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting technical analysis")
	
	symbolsStr := r.URL.Query().Get("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	
	analysis, err := h.brightDataService.GetTechnicalAnalysis(r.Context(), symbols)
	if err != nil {
		h.logger.Errorf("Failed to get technical analysis: %v", err)
		http.Error(w, "Failed to get technical analysis", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"analysis": analysis,
		"count":    len(analysis),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnicalAnalysisBySymbol returns technical analysis for a specific symbol
func (h *TradingSignalsHandler) GetTechnicalAnalysisBySymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := strings.ToUpper(vars["symbol"])
	
	h.logger.Infof("Getting technical analysis for symbol: %s", symbol)
	
	analysis, err := h.brightDataService.GetTechnicalAnalysis(r.Context(), []string{symbol})
	if err != nil {
		h.logger.Errorf("Failed to get technical analysis for %s: %v", symbol, err)
		http.Error(w, "Failed to get technical analysis", http.StatusInternalServerError)
		return
	}
	
	if len(analysis) == 0 {
		http.Error(w, "No technical analysis found for symbol", http.StatusNotFound)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveDeals returns all active trading deals
func (h *TradingSignalsHandler) GetActiveDeals(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting active trading deals")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	deals, err := h.brightDataService.GetActiveDeals(r.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get active deals: %v", err)
		http.Error(w, "Failed to get active deals", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"deals": deals,
		"count": len(deals),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveDealsOnly returns only active deals (status = "active")
func (h *TradingSignalsHandler) GetActiveDealsOnly(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting active deals only")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	deals, err := h.brightDataService.GetActiveDeals(r.Context(), limit)
	if err != nil {
		h.logger.Errorf("Failed to get active deals: %v", err)
		http.Error(w, "Failed to get active deals", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"active_deals": deals,
		"count":        len(deals),
		"status":       "active_only",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCommasBots returns 3commas specific bots
func (h *TradingSignalsHandler) GetCommasBots(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting 3commas bots")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all bots and filter for 3commas
	allBots, err := h.brightDataService.GetTradingBots(r.Context(), 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas bots: %v", err)
		http.Error(w, "Failed to get 3commas bots", http.StatusInternalServerError)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCommasSignals returns 3commas specific signals
func (h *TradingSignalsHandler) GetCommasSignals(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting 3commas signals")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all signals and filter for 3commas
	allSignals, err := h.brightDataService.GetTradingSignals(r.Context(), nil, 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas signals: %v", err)
		http.Error(w, "Failed to get 3commas signals", http.StatusInternalServerError)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCommasDeals returns 3commas specific deals
func (h *TradingSignalsHandler) GetCommasDeals(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting 3commas deals")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	// Get all deals and filter for 3commas
	allDeals, err := h.brightDataService.GetActiveDeals(r.Context(), 0)
	if err != nil {
		h.logger.Errorf("Failed to get 3commas deals: %v", err)
		http.Error(w, "Failed to get 3commas deals", http.StatusInternalServerError)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
