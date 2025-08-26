package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/web3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the Web3 service
type Handler struct {
	service *web3.Service
	logger  *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(service *web3.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreatePayment handles payment creation requests
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req web3.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	response, err := h.service.CreatePayment(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create payment", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to create payment")
		return
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetPaymentStatus handles payment status requests
func (h *Handler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payment/status/")
	paymentID := strings.Split(path, "/")[0]
	
	if paymentID == "" {
		h.writeError(w, http.StatusBadRequest, "Payment ID is required")
		return
	}

	ctx := r.Context()
	payment, err := h.service.GetPaymentStatus(ctx, paymentID)
	if err != nil {
		h.logger.Error("Failed to get payment status", zap.Error(err))
		h.writeError(w, http.StatusNotFound, "Payment not found")
		return
	}

	h.writeJSON(w, http.StatusOK, payment)
}

// ConfirmPayment handles payment confirmation requests
func (h *Handler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		PaymentID       string `json:"payment_id"`
		TransactionHash string `json:"transaction_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PaymentID == "" || req.TransactionHash == "" {
		h.writeError(w, http.StatusBadRequest, "Payment ID and transaction hash are required")
		return
	}

	ctx := r.Context()
	if err := h.service.ConfirmPayment(ctx, req.PaymentID, req.TransactionHash); err != nil {
		h.logger.Error("Failed to confirm payment", zap.Error(err))
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

// CancelPayment handles payment cancellation requests
func (h *Handler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		PaymentID string `json:"payment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PaymentID == "" {
		h.writeError(w, http.StatusBadRequest, "Payment ID is required")
		return
	}

	ctx := r.Context()
	if err := h.service.CancelPayment(ctx, req.PaymentID); err != nil {
		h.logger.Error("Failed to cancel payment", zap.Error(err))
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetWalletBalance handles wallet balance requests
func (h *Handler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract wallet address from URL path
	path := strings.TrimPrefix(r.URL.Path, "/wallet/balance/")
	address := strings.Split(path, "/")[0]
	
	if address == "" {
		h.writeError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	// Mock response - in production, this would query blockchain
	balance := map[string]interface{}{
		"address": address,
		"balances": map[string]string{
			"ETH":    "1.5",
			"USDC":   "1000.0",
			"COFFEE": "50.0",
		},
		"total_usd": "3500.00",
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, balance)
}

// GetWalletTransactions handles wallet transaction history requests
func (h *Handler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract wallet address from URL path
	path := strings.TrimPrefix(r.URL.Path, "/wallet/transactions/")
	address := strings.Split(path, "/")[0]
	
	if address == "" {
		h.writeError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	// Mock response - in production, this would query blockchain
	transactions := map[string]interface{}{
		"address": address,
		"transactions": []map[string]interface{}{
			{
				"hash":      "0x1234567890abcdef",
				"type":      "payment",
				"amount":    "5.0",
				"currency":  "USDC",
				"status":    "confirmed",
				"timestamp": time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339),
			},
			{
				"hash":      "0xabcdef1234567890",
				"type":      "payment",
				"amount":    "0.1",
				"currency":  "ETH",
				"status":    "confirmed",
				"timestamp": time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
			},
		},
		"total": 2,
	}

	h.writeJSON(w, http.StatusOK, transactions)
}

// GetTokenPrice handles token price requests
func (h *Handler) GetTokenPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract token symbol from URL path
	path := strings.TrimPrefix(r.URL.Path, "/token/price/")
	symbol := strings.Split(path, "/")[0]
	
	if symbol == "" {
		h.writeError(w, http.StatusBadRequest, "Token symbol is required")
		return
	}

	// Mock prices - in production, this would fetch from price oracles
	prices := map[string]string{
		"ETH":    "2500.00",
		"BNB":    "300.00",
		"MATIC":  "0.80",
		"SOL":    "100.00",
		"USDC":   "1.00",
		"USDT":   "1.00",
		"COFFEE": "0.50",
	}

	price, exists := prices[strings.ToUpper(symbol)]
	if !exists {
		h.writeError(w, http.StatusNotFound, "Token not found")
		return
	}

	response := map[string]interface{}{
		"symbol":     strings.ToUpper(symbol),
		"price_usd":  price,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// SwapTokens handles token swap requests
func (h *Handler) SwapTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		FromToken string          `json:"from_token"`
		ToToken   string          `json:"to_token"`
		Amount    decimal.Decimal `json:"amount"`
		Chain     string          `json:"chain"`
		Slippage  float64         `json:"slippage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Mock swap response - in production, this would execute actual swap
	response := map[string]interface{}{
		"swap_id":        "swap_123456",
		"from_token":     req.FromToken,
		"to_token":       req.ToToken,
		"from_amount":    req.Amount.String(),
		"to_amount":      req.Amount.Mul(decimal.NewFromFloat(0.95)).String(), // Mock 5% slippage
		"estimated_gas":  "0.005 ETH",
		"status":         "pending",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetYieldOpportunities handles yield farming opportunity requests
func (h *Handler) GetYieldOpportunities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Mock yield opportunities - in production, this would fetch from DeFi protocols
	opportunities := map[string]interface{}{
		"opportunities": []map[string]interface{}{
			{
				"protocol":     "Uniswap V3",
				"pair":         "ETH/USDC",
				"apy":          "12.5%",
				"tvl":          "$1,250,000",
				"risk_level":   "medium",
				"min_deposit":  "0.1 ETH",
			},
			{
				"protocol":     "Aave",
				"token":        "USDC",
				"apy":          "8.2%",
				"tvl":          "$500,000,000",
				"risk_level":   "low",
				"min_deposit":  "100 USDC",
			},
		},
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, opportunities)
}

// StakeTokens handles token staking requests
func (h *Handler) StakeTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Token    string          `json:"token"`
		Amount   decimal.Decimal `json:"amount"`
		Protocol string          `json:"protocol"`
		Duration int             `json:"duration_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Mock staking response
	response := map[string]interface{}{
		"stake_id":       "stake_123456",
		"token":          req.Token,
		"amount":         req.Amount.String(),
		"protocol":       req.Protocol,
		"duration_days":  req.Duration,
		"estimated_apy":  "15.2%",
		"status":         "pending",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// UnstakeTokens handles token unstaking requests
func (h *Handler) UnstakeTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		StakeID string `json:"stake_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Mock unstaking response
	response := map[string]interface{}{
		"stake_id":       req.StakeID,
		"status":         "unstaking",
		"unstake_time":   time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339),
		"penalty":        "0%",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"service":   "web3-payment",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ReadinessCheck handles readiness check requests
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ready",
		"service":   "web3-payment",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]string{
			"ethereum": "ok",
			"bsc":      "ok",
			"polygon":  "ok",
			"solana":   "ok",
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
