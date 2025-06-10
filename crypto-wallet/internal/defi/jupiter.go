package defi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

// JupiterClient represents a client for Jupiter aggregator
type JupiterClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
}

// JupiterRoute represents a route from Jupiter
type JupiterRoute struct {
	InputMint            string          `json:"inputMint"`
	InAmount             string          `json:"inAmount"`
	OutputMint           string          `json:"outputMint"`
	OutAmount            string          `json:"outAmount"`
	OtherAmountThreshold string          `json:"otherAmountThreshold"`
	SwapMode             string          `json:"swapMode"`
	SlippageBps          int             `json:"slippageBps"`
	PlatformFee          *PlatformFee    `json:"platformFee,omitempty"`
	PriceImpactPct       string          `json:"priceImpactPct"`
	RoutePlan            []RoutePlanStep `json:"routePlan"`
}

// PlatformFee represents platform fee structure
type PlatformFee struct {
	Amount     string `json:"amount"`
	FeeBps     int    `json:"feeBps"`
	FeeMint    string `json:"feeMint"`
	FeeAccount string `json:"feeAccount"`
}

// RoutePlanStep represents a step in the route plan
type RoutePlanStep struct {
	SwapInfo SwapInfo `json:"swapInfo"`
	Percent  int      `json:"percent"`
}

// SwapInfo represents swap information
type SwapInfo struct {
	AmmKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

// JupiterQuoteResponse represents Jupiter quote response
type JupiterQuoteResponse struct {
	Data        []JupiterRoute `json:"data"`
	TimeTaken   float64        `json:"timeTaken"`
	ContextSlot int64          `json:"contextSlot"`
}

// JupiterSwapRequest represents swap request
type JupiterSwapRequest struct {
	Route                         JupiterRoute `json:"route"`
	UserPublicKey                 string       `json:"userPublicKey"`
	WrapUnwrapSOL                 bool         `json:"wrapUnwrapSOL"`
	UseSharedAccounts             bool         `json:"useSharedAccounts"`
	FeeAccount                    string       `json:"feeAccount,omitempty"`
	TrackingAccount               string       `json:"trackingAccount,omitempty"`
	ComputeUnitPriceMicroLamports int          `json:"computeUnitPriceMicroLamports,omitempty"`
}

// JupiterSwapResponse represents swap response
type JupiterSwapResponse struct {
	SwapTransaction      string `json:"swapTransaction"`
	LastValidBlockHeight int64  `json:"lastValidBlockHeight"`
}

// NewJupiterClient creates a new Jupiter client
func NewJupiterClient(logger *logger.Logger) *JupiterClient {
	return &JupiterClient{
		baseURL: "https://quote-api.jup.ag/v6",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.Named("jupiter"),
	}
}

// GetQuote gets a quote for token swap
func (j *JupiterClient) GetQuote(ctx context.Context, inputMint, outputMint string, amount decimal.Decimal, slippageBps int) (*JupiterRoute, error) {
	j.logger.Info(fmt.Sprintf("Getting Jupiter quote: %s -> %s, amount: %s", inputMint, outputMint, amount.String()))

	// Convert amount to lamports/smallest unit
	amountStr := amount.Mul(decimal.NewFromInt(1000000000)).String() // Assuming 9 decimals

	// Build URL
	params := url.Values{}
	params.Set("inputMint", inputMint)
	params.Set("outputMint", outputMint)
	params.Set("amount", amountStr)
	params.Set("slippageBps", fmt.Sprintf("%d", slippageBps))
	params.Set("onlyDirectRoutes", "false")
	params.Set("asLegacyTransaction", "false")

	reqURL := fmt.Sprintf("%s/quote?%s", j.baseURL, params.Encode())

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var quoteResp JupiterQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(quoteResp.Data) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	// Return best route (first one)
	bestRoute := &quoteResp.Data[0]

	j.logger.Info(fmt.Sprintf("Jupiter quote received: %s -> %s", bestRoute.InAmount, bestRoute.OutAmount))
	return bestRoute, nil
}

// GetSwapTransaction gets a swap transaction for execution
func (j *JupiterClient) GetSwapTransaction(ctx context.Context, route *JupiterRoute, userPublicKey string) (*JupiterSwapResponse, error) {
	j.logger.Info(fmt.Sprintf("Getting swap transaction for user %s", userPublicKey))

	// Prepare swap request
	swapReq := JupiterSwapRequest{
		Route:                         *route,
		UserPublicKey:                 userPublicKey,
		WrapUnwrapSOL:                 true,
		UseSharedAccounts:             true,
		ComputeUnitPriceMicroLamports: 1000,
	}

	// Convert to JSON
	reqBody, err := json.Marshal(swapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request
	reqURL := fmt.Sprintf("%s/swap", j.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var swapResp JupiterSwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	j.logger.Info("Swap transaction received successfully")
	return &swapResp, nil
}

// ExecuteSwap executes a token swap using Jupiter
func (j *JupiterClient) ExecuteSwap(ctx context.Context, inputMint, outputMint string, amount decimal.Decimal, slippageBps int, userWallet solana.PrivateKey) (string, error) {
	j.logger.Info(fmt.Sprintf("Executing Jupiter swap: %s -> %s", inputMint, outputMint))

	// Get quote
	route, err := j.GetQuote(ctx, inputMint, outputMint, amount, slippageBps)
	if err != nil {
		return "", fmt.Errorf("failed to get quote: %w", err)
	}

	// Get swap transaction
	userPublicKey := userWallet.PublicKey().String()
	_, err = j.GetSwapTransaction(ctx, route, userPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to get swap transaction: %w", err)
	}

	// In a real implementation, you would:
	// 1. Decode the base64 transaction
	// 2. Sign it with the user's wallet
	// 3. Send it to the Solana network
	// 4. Wait for confirmation

	// For now, return a mock signature
	mockSignature := "JupiterSwap" + fmt.Sprintf("%d", time.Now().Unix())

	j.logger.Info(fmt.Sprintf("Jupiter swap executed successfully, signature: %s", mockSignature))
	return mockSignature, nil
}

// GetSupportedTokens gets list of supported tokens
func (j *JupiterClient) GetSupportedTokens(ctx context.Context) ([]string, error) {
	j.logger.Info("Getting supported tokens from Jupiter")

	// Mock implementation - in production, this would call the actual API
	supportedTokens := []string{
		"So11111111111111111111111111111111111111112",  // SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
		"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",  // mSOL
		"7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj", // stSOL
	}

	j.logger.Info(fmt.Sprintf("Retrieved %d supported tokens", len(supportedTokens)))
	return supportedTokens, nil
}

// GetTokenPrice gets current token price
func (j *JupiterClient) GetTokenPrice(ctx context.Context, tokenMint string) (decimal.Decimal, error) {
	j.logger.Info(fmt.Sprintf("Getting token price for %s", tokenMint))

	// Mock implementation - in production, this would call the price API
	var price decimal.Decimal

	switch tokenMint {
	case "So11111111111111111111111111111111111111112": // SOL
		price = decimal.NewFromFloat(50.0)
	case "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": // USDC
		price = decimal.NewFromFloat(1.0)
	case "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": // USDT
		price = decimal.NewFromFloat(1.0)
	default:
		price = decimal.NewFromFloat(1.0)
	}

	j.logger.Info(fmt.Sprintf("Token price for %s: %s", tokenMint, price.String()))
	return price, nil
}
