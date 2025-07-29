package aggregators

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
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MatchaClient handles interactions with Matcha DEX aggregator
type MatchaClient struct {
	logger     *logger.Logger
	config     AggregatorConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// MatchaQuoteResponse represents Matcha quote response
type MatchaQuoteResponse struct {
	Price            string `json:"price"`
	GuaranteedPrice  string `json:"guaranteedPrice"`
	To               string `json:"to"`
	Data             string `json:"data"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	EstimatedGas     string `json:"estimatedGas"`
	GasPrice         string `json:"gasPrice"`
	BuyTokenAddress  string `json:"buyTokenAddress"`
	SellTokenAddress string `json:"sellTokenAddress"`
	BuyAmount        string `json:"buyAmount"`
	SellAmount       string `json:"sellAmount"`
	Sources          []struct {
		Name       string `json:"name"`
		Proportion string `json:"proportion"`
	} `json:"sources"`
	AllowanceTarget string `json:"allowanceTarget"`
}

// MatchaTokenResponse represents Matcha token response
type MatchaTokenResponse struct {
	Records []struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"records"`
}

// NewMatchaClient creates a new Matcha client
func NewMatchaClient(logger *logger.Logger, config AggregatorConfig) *MatchaClient {
	return &MatchaClient{
		logger:  logger.Named("matcha"),
		config:  config,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetQuote gets a swap quote from Matcha
func (mc *MatchaClient) GetQuote(ctx context.Context, req *QuoteRequest) (*SwapQuote, error) {
	mc.logger.Info("Getting Matcha quote",
		zap.String("token_in", req.TokenIn.Symbol),
		zap.String("token_out", req.TokenOut.Symbol),
		zap.String("amount_in", req.AmountIn.String()))

	// Build quote request URL
	quoteURL := fmt.Sprintf("%s/swap/v1/quote", mc.baseURL)
	params := url.Values{}
	params.Set("sellToken", req.TokenIn.Address)
	params.Set("buyToken", req.TokenOut.Address)
	params.Set("sellAmount", req.AmountIn.Mul(decimal.NewFromInt(1e18)).String()) // Convert to wei

	if req.Slippage.GreaterThan(decimal.Zero) {
		params.Set("slippagePercentage", req.Slippage.String())
	}

	// Make API request
	resp, err := mc.makeRequest(ctx, "GET", quoteURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	var quoteResp MatchaQuoteResponse
	if err := json.Unmarshal(resp, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	// Parse amounts
	amountOut, err := decimal.NewFromString(quoteResp.BuyAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse buy amount: %w", err)
	}
	amountOut = amountOut.Div(decimal.NewFromInt(1e18)) // Convert from wei

	// Calculate slippage
	slippage := req.Slippage
	if slippage.IsZero() {
		slippage = decimal.NewFromFloat(0.01) // Default 1%
	}
	minAmountOut := amountOut.Mul(decimal.NewFromInt(1).Sub(slippage))

	// Parse gas estimate
	gasEstimate, _ := decimal.NewFromString(quoteResp.EstimatedGas)
	gasPrice, _ := decimal.NewFromString(quoteResp.GasPrice)
	gasCost := gasEstimate.Mul(gasPrice).Div(decimal.NewFromInt(1e18))

	// Build route steps
	route := make([]RouteStep, 0, len(quoteResp.Sources))
	for _, source := range quoteResp.Sources {
		proportion, _ := decimal.NewFromString(source.Proportion)
		route = append(route, RouteStep{
			Protocol:   source.Name,
			Percentage: proportion.Mul(decimal.NewFromInt(100)),
		})
	}

	quote := &SwapQuote{
		ID:           uuid.New().String(),
		Aggregator:   "matcha",
		Protocol:     "matcha",
		Chain:        req.Chain,
		TokenIn:      req.TokenIn,
		TokenOut:     req.TokenOut,
		AmountIn:     req.AmountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  decimal.NewFromFloat(0.1), // Simplified
		Fee:          decimal.Zero,              // Matcha doesn't charge fees
		GasEstimate:  uint64(gasEstimate.IntPart()),
		GasPrice:     gasPrice,
		GasCost:      gasCost,
		Route:        route,
		Slippage:     slippage,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
		ExecutionData: map[string]interface{}{
			"to":              quoteResp.To,
			"data":            quoteResp.Data,
			"value":           quoteResp.Value,
			"allowanceTarget": quoteResp.AllowanceTarget,
		},
		Metadata: make(map[string]interface{}),
	}

	return quote, nil
}

// ExecuteSwap executes a token swap via Matcha
func (mc *MatchaClient) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	mc.logger.Info("Executing Matcha swap", zap.String("quote_id", req.Quote.ID))

	// Build swap request URL
	swapURL := fmt.Sprintf("%s/swap/v1/quote", mc.baseURL)
	params := url.Values{}
	params.Set("sellToken", req.Quote.TokenIn.Address)
	params.Set("buyToken", req.Quote.TokenOut.Address)
	params.Set("sellAmount", req.Quote.AmountIn.Mul(decimal.NewFromInt(1e18)).String())
	params.Set("takerAddress", req.UserAddress)
	params.Set("slippagePercentage", req.Slippage.String())

	// Make API request
	resp, err := mc.makeRequest(ctx, "GET", swapURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap transaction: %w", err)
	}

	var swapResp MatchaQuoteResponse
	if err := json.Unmarshal(resp, &swapResp); err != nil {
		return nil, fmt.Errorf("failed to parse swap response: %w", err)
	}

	// In a real implementation, you would sign and send the transaction
	// For now, return a mock result
	result := &SwapResult{
		TransactionHash: "0x" + uuid.New().String(),
		Status:          "pending",
		AmountIn:        req.Quote.AmountIn,
		AmountOut:       req.Quote.AmountOut,
		GasUsed:         req.Quote.GasEstimate,
		GasCost:         req.Quote.GasCost,
		Metadata: map[string]interface{}{
			"transaction": swapResp,
		},
	}

	return result, nil
}

// GetSupportedTokens returns supported tokens for Matcha
func (mc *MatchaClient) GetSupportedTokens(ctx context.Context, chain string) ([]Token, error) {
	mc.logger.Debug("Getting Matcha supported tokens", zap.String("chain", chain))

	tokensURL := fmt.Sprintf("%s/tokens", mc.baseURL)

	resp, err := mc.makeRequest(ctx, "GET", tokensURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	var tokensResp MatchaTokenResponse
	if err := json.Unmarshal(resp, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to parse tokens response: %w", err)
	}

	// Convert to our Token format
	tokens := make([]Token, 0, len(tokensResp.Records))
	for _, token := range tokensResp.Records {
		tokens = append(tokens, Token{
			Address:  token.Address,
			Symbol:   token.Symbol,
			Name:     token.Name,
			Decimals: token.Decimals,
			Chain:    chain,
			LogoURI:  token.LogoURI,
		})
	}

	return tokens, nil
}

// GetStatus returns the status of the Matcha client
func (mc *MatchaClient) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":     "matcha",
		"enabled":  mc.config.Enabled,
		"base_url": mc.baseURL,
		"chains":   mc.config.Chains,
		"healthy":  true, // Simplified
	}
}

// Helper methods

// makeRequest makes an HTTP request to Matcha API
func (mc *MatchaClient) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CryptoWallet-DEXAggregator/1.0")

	if mc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+mc.apiKey)
	}

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
