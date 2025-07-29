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

// ZeroXClient handles interactions with 0x Protocol DEX aggregator
type ZeroXClient struct {
	logger     *logger.Logger
	config     AggregatorConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// ZeroXQuoteResponse represents 0x quote response
type ZeroXQuoteResponse struct {
	Price              string `json:"price"`
	GuaranteedPrice    string `json:"guaranteedPrice"`
	To                 string `json:"to"`
	Data               string `json:"data"`
	Value              string `json:"value"`
	Gas                string `json:"gas"`
	EstimatedGas       string `json:"estimatedGas"`
	GasPrice           string `json:"gasPrice"`
	ProtocolFee        string `json:"protocolFee"`
	MinimumProtocolFee string `json:"minimumProtocolFee"`
	BuyTokenAddress    string `json:"buyTokenAddress"`
	SellTokenAddress   string `json:"sellTokenAddress"`
	BuyAmount          string `json:"buyAmount"`
	SellAmount         string `json:"sellAmount"`
	Sources            []struct {
		Name       string `json:"name"`
		Proportion string `json:"proportion"`
	} `json:"sources"`
	AllowanceTarget string `json:"allowanceTarget"`
}

// ZeroXSwapResponse represents 0x swap response
type ZeroXSwapResponse struct {
	ChainId            int    `json:"chainId"`
	Price              string `json:"price"`
	GuaranteedPrice    string `json:"guaranteedPrice"`
	To                 string `json:"to"`
	Data               string `json:"data"`
	Value              string `json:"value"`
	Gas                string `json:"gas"`
	EstimatedGas       string `json:"estimatedGas"`
	GasPrice           string `json:"gasPrice"`
	ProtocolFee        string `json:"protocolFee"`
	MinimumProtocolFee string `json:"minimumProtocolFee"`
	BuyTokenAddress    string `json:"buyTokenAddress"`
	SellTokenAddress   string `json:"sellTokenAddress"`
	BuyAmount          string `json:"buyAmount"`
	SellAmount         string `json:"sellAmount"`
	AllowanceTarget    string `json:"allowanceTarget"`
}

// NewZeroXClient creates a new 0x client
func NewZeroXClient(logger *logger.Logger, config AggregatorConfig) *ZeroXClient {
	return &ZeroXClient{
		logger:  logger.Named("0x"),
		config:  config,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetQuote gets a swap quote from 0x Protocol
func (zc *ZeroXClient) GetQuote(ctx context.Context, req *QuoteRequest) (*SwapQuote, error) {
	zc.logger.Info("Getting 0x quote",
		zap.String("token_in", req.TokenIn.Symbol),
		zap.String("token_out", req.TokenOut.Symbol),
		zap.String("amount_in", req.AmountIn.String()))

	// Build quote request URL
	quoteURL := fmt.Sprintf("%s/swap/v1/quote", zc.baseURL)
	params := url.Values{}
	params.Set("sellToken", req.TokenIn.Address)
	params.Set("buyToken", req.TokenOut.Address)
	params.Set("sellAmount", req.AmountIn.Mul(decimal.NewFromInt(1e18)).String()) // Convert to wei

	if req.Slippage.GreaterThan(decimal.Zero) {
		params.Set("slippagePercentage", req.Slippage.String())
	}

	// Make API request
	resp, err := zc.makeRequest(ctx, "GET", quoteURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	var quoteResp ZeroXQuoteResponse
	if err := json.Unmarshal(resp, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	// Parse amounts
	amountOut, err := decimal.NewFromString(quoteResp.BuyAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse buy amount: %w", err)
	}
	amountOut = amountOut.Div(decimal.NewFromInt(1e18)) // Convert from wei

	// Note: guaranteedPrice is available in quoteResp.GuaranteedPrice if needed for future use

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

	// Parse protocol fee
	protocolFee, _ := decimal.NewFromString(quoteResp.ProtocolFee)
	protocolFee = protocolFee.Div(decimal.NewFromInt(1e18))

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
		Aggregator:   "0x",
		Protocol:     "0x",
		Chain:        req.Chain,
		TokenIn:      req.TokenIn,
		TokenOut:     req.TokenOut,
		AmountIn:     req.AmountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  decimal.NewFromFloat(0.1), // Simplified
		Fee:          protocolFee,
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

// ExecuteSwap executes a token swap via 0x Protocol
func (zc *ZeroXClient) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	zc.logger.Info("Executing 0x swap", zap.String("quote_id", req.Quote.ID))

	// Build swap request URL
	swapURL := fmt.Sprintf("%s/swap/v1/quote", zc.baseURL)
	params := url.Values{}
	params.Set("sellToken", req.Quote.TokenIn.Address)
	params.Set("buyToken", req.Quote.TokenOut.Address)
	params.Set("sellAmount", req.Quote.AmountIn.Mul(decimal.NewFromInt(1e18)).String())
	params.Set("takerAddress", req.UserAddress)
	params.Set("slippagePercentage", req.Slippage.String())

	// Make API request
	resp, err := zc.makeRequest(ctx, "GET", swapURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap transaction: %w", err)
	}

	var swapResp ZeroXSwapResponse
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

// GetSupportedTokens returns supported tokens for 0x Protocol
func (zc *ZeroXClient) GetSupportedTokens(ctx context.Context, chain string) ([]Token, error) {
	zc.logger.Debug("Getting 0x supported tokens", zap.String("chain", chain))

	// 0x doesn't have a specific tokens endpoint, so we return common tokens
	tokens := []Token{
		{
			Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			Symbol:   "WETH",
			Name:     "Wrapped Ethereum",
			Decimals: 18,
			Chain:    chain,
		},
		{
			Address:  "0xA0b86a33E6441E6C8C07C4c4c8e8B0E8E8E8E8E8",
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
			Chain:    chain,
		},
		{
			Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			Symbol:   "USDT",
			Name:     "Tether USD",
			Decimals: 6,
			Chain:    chain,
		},
	}

	return tokens, nil
}

// GetStatus returns the status of the 0x client
func (zc *ZeroXClient) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":     "0x",
		"enabled":  zc.config.Enabled,
		"base_url": zc.baseURL,
		"chains":   zc.config.Chains,
		"healthy":  true, // Simplified
	}
}

// Helper methods

// makeRequest makes an HTTP request to 0x API
func (zc *ZeroXClient) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
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

	if zc.apiKey != "" {
		req.Header.Set("0x-api-key", zc.apiKey)
	}

	resp, err := zc.httpClient.Do(req)
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
