package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/defi"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/shopspring/decimal"
)

// Handler represents the HTTP handler for DeFi service
type Handler struct {
	defiService *defi.Service
	logger      *logger.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(defiService *defi.Service, logger *logger.Logger) *Handler {
	return &Handler{
		defiService: defiService,
		logger:      logger,
	}
}

// SetupRoutes configures the HTTP routes for the DeFi service
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", h.methodHandler("GET", h.handleHealthCheck))

	// Token price endpoints
	mux.HandleFunc("/api/v1/tokens/price", h.methodHandler("POST", h.handleGetTokenPrice))

	// Swap endpoints
	mux.HandleFunc("/api/v1/swaps/quote", h.methodHandler("POST", h.handleGetSwapQuote))
	mux.HandleFunc("/api/v1/swaps/execute", h.methodHandler("POST", h.handleExecuteSwap))

	// Liquidity pool endpoints
	mux.HandleFunc("/api/v1/pools", h.methodHandler("GET", h.handleGetLiquidityPools))

	// Arbitrage endpoints
	mux.HandleFunc("/api/v1/arbitrage/opportunities", h.methodHandler("GET", h.handleGetArbitrageOpportunities))

	// Yield farming endpoints
	mux.HandleFunc("/api/v1/yield/opportunities", h.methodHandler("GET", h.handleGetYieldOpportunities))

	// Trading bot endpoints
	mux.HandleFunc("/api/v1/bots", h.handleTradingBots)
	mux.HandleFunc("/api/v1/bots/", h.handleTradingBotWithID)

	// On-chain analysis endpoints
	mux.HandleFunc("/api/v1/analysis/signals", h.methodHandler("GET", h.handleGetMarketSignals))
	mux.HandleFunc("/api/v1/analysis/whales", h.methodHandler("GET", h.handleGetWhaleActivity))
	mux.HandleFunc("/api/v1/analysis/tokens/", h.handleTokenAnalysis)
}

// methodHandler is a middleware that checks HTTP method
func (h *Handler) methodHandler(allowedMethod string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handler(w, r)
	}
}

// handleHealthCheck handles health check requests
func (h *Handler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "defi-service",
		"timestamp": "2024-01-01T00:00:00Z", // This would be time.Now().UTC() in real implementation
	}
	h.writeSuccessResponse(w, response)
}

// handleGetTokenPrice handles token price requests
func (h *Handler) handleGetTokenPrice(w http.ResponseWriter, r *http.Request) {
	var req defi.GetTokenPriceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	resp, err := h.defiService.GetTokenPrice(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get token price")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleGetSwapQuote handles swap quote requests
func (h *Handler) handleGetSwapQuote(w http.ResponseWriter, r *http.Request) {
	var req defi.GetSwapQuoteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	resp, err := h.defiService.GetSwapQuote(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get swap quote")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleExecuteSwap handles swap execution requests
func (h *Handler) handleExecuteSwap(w http.ResponseWriter, r *http.Request) {
	var req defi.ExecuteSwapRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	resp, err := h.defiService.ExecuteSwap(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute swap")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleGetLiquidityPools handles liquidity pool requests
func (h *Handler) handleGetLiquidityPools(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	req := &defi.GetLiquidityPoolsRequest{
		Chain:    defi.Chain(query.Get("chain")),
		Protocol: defi.ProtocolType(query.Get("protocol")),
		Token0:   query.Get("token0"),
		Token1:   query.Get("token1"),
	}

	// Parse optional parameters
	if minTVL := query.Get("min_tvl"); minTVL != "" {
		if tvl, err := decimal.NewFromString(minTVL); err == nil {
			req.MinTVL = tvl
		}
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	resp, err := h.defiService.GetLiquidityPools(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get liquidity pools")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, resp)
}

// handleGetArbitrageOpportunities handles arbitrage opportunity requests
func (h *Handler) handleGetArbitrageOpportunities(w http.ResponseWriter, r *http.Request) {
	opportunities, err := h.defiService.GetArbitrageOpportunities(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get arbitrage opportunities")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"opportunities": opportunities,
		"count":         len(opportunities),
	}
	h.writeSuccessResponse(w, response)
}

// handleGetYieldOpportunities handles yield farming opportunity requests
func (h *Handler) handleGetYieldOpportunities(w http.ResponseWriter, r *http.Request) {
	limit := 10 // default limit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	opportunities, err := h.defiService.GetBestYieldOpportunities(r.Context(), limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get yield opportunities")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"opportunities": opportunities,
		"count":         len(opportunities),
	}
	h.writeSuccessResponse(w, response)
}

// handleTradingBots handles trading bot operations without ID
func (h *Handler) handleTradingBots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreateTradingBot(w, r)
	case http.MethodGet:
		h.handleGetAllTradingBots(w, r)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleTradingBotWithID handles trading bot operations with ID parameter
func (h *Handler) handleTradingBotWithID(w http.ResponseWriter, r *http.Request) {
	// Extract bot ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/bots/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Bot ID is required")
		return
	}

	botID := parts[0]

	switch r.Method {
	case http.MethodGet:
		if len(parts) >= 2 && parts[1] == "performance" {
			h.handleGetTradingBotPerformance(w, r, botID)
		} else {
			h.handleGetTradingBot(w, r, botID)
		}
	case http.MethodPost:
		if len(parts) >= 2 {
			action := parts[1]
			switch action {
			case "start":
				h.handleStartTradingBot(w, r, botID)
			case "stop":
				h.handleStopTradingBot(w, r, botID)
			default:
				h.writeErrorResponse(w, http.StatusNotFound, "Action not found")
			}
		} else {
			h.writeErrorResponse(w, http.StatusNotFound, "Action is required")
		}
	case http.MethodDelete:
		h.handleDeleteTradingBot(w, r, botID)
	default:
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleTokenAnalysis handles token analysis requests
func (h *Handler) handleTokenAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract token address from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/analysis/tokens/")
	if path == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Token address is required")
		return
	}

	analysis, err := h.defiService.GetTokenAnalysis(r.Context(), path)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get token analysis")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, analysis)
}

// Helper methods for response handling
func (h *Handler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	h.writeSuccessResponseWithStatus(w, http.StatusOK, data)
}

func (h *Handler) writeSuccessResponseWithStatus(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Trading bot handler methods

// handleCreateTradingBot handles trading bot creation requests
func (h *Handler) handleCreateTradingBot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string                   `json:"name"`
		Strategy defi.TradingStrategyType `json:"strategy"`
		Config   defi.TradingBotConfig    `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Name == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	bot, err := h.defiService.CreateTradingBot(r.Context(), req.Name, req.Strategy, req.Config)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create trading bot")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponseWithStatus(w, http.StatusCreated, bot)
}

// handleGetAllTradingBots handles requests to get all trading bots
func (h *Handler) handleGetAllTradingBots(w http.ResponseWriter, r *http.Request) {
	bots, err := h.defiService.GetAllTradingBots(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trading bots")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"bots":  bots,
		"count": len(bots),
	}
	h.writeSuccessResponse(w, response)
}

// handleGetTradingBot handles requests to get a specific trading bot
func (h *Handler) handleGetTradingBot(w http.ResponseWriter, r *http.Request, botID string) {
	bot, err := h.defiService.GetTradingBot(r.Context(), botID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trading bot")
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, bot)
}

// handleStartTradingBot handles requests to start a trading bot
func (h *Handler) handleStartTradingBot(w http.ResponseWriter, r *http.Request, botID string) {
	err := h.defiService.StartTradingBot(r.Context(), botID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start trading bot")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Trading bot started successfully"}
	h.writeSuccessResponse(w, response)
}

// handleStopTradingBot handles requests to stop a trading bot
func (h *Handler) handleStopTradingBot(w http.ResponseWriter, r *http.Request, botID string) {
	err := h.defiService.StopTradingBot(r.Context(), botID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to stop trading bot")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Trading bot stopped successfully"}
	h.writeSuccessResponse(w, response)
}

// handleDeleteTradingBot handles requests to delete a trading bot
func (h *Handler) handleDeleteTradingBot(w http.ResponseWriter, r *http.Request, botID string) {
	err := h.defiService.DeleteTradingBot(r.Context(), botID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete trading bot")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]string{"message": "Trading bot deleted successfully"}
	h.writeSuccessResponse(w, response)
}

// handleGetTradingBotPerformance handles requests to get trading bot performance
func (h *Handler) handleGetTradingBotPerformance(w http.ResponseWriter, r *http.Request, botID string) {
	performance, err := h.defiService.GetTradingBotPerformance(r.Context(), botID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trading bot performance")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccessResponse(w, performance)
}

// Analysis handler methods

// handleGetMarketSignals handles requests to get market signals
func (h *Handler) handleGetMarketSignals(w http.ResponseWriter, r *http.Request) {
	signals, err := h.defiService.GetMarketSignals(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get market signals")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"signals": signals,
		"count":   len(signals),
	}
	h.writeSuccessResponse(w, response)
}

// handleGetWhaleActivity handles requests to get whale activity
func (h *Handler) handleGetWhaleActivity(w http.ResponseWriter, r *http.Request) {
	whales, err := h.defiService.GetWhaleActivity(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get whale activity")
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"whales": whales,
		"count":  len(whales),
	}
	h.writeSuccessResponse(w, response)
}
