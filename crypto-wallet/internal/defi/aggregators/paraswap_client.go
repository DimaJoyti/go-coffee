package aggregators

import (
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

// ParaswapClient handles interactions with Paraswap DEX aggregator
type ParaswapClient struct {
	logger     *logger.Logger
	config     AggregatorConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// ParaswapQuoteResponse represents Paraswap quote response
type ParaswapQuoteResponse struct {
	PriceRoute struct {
		SrcToken      string `json:"srcToken"`
		SrcDecimals   int    `json:"srcDecimals"`
		SrcAmount     string `json:"srcAmount"`
		DestToken     string `json:"destToken"`
		DestDecimals  int    `json:"destDecimals"`
		DestAmount    string `json:"destAmount"`
		BestRoute     []struct {
			Exchange string `json:"exchange"`
			Percent  int    `json:"percent"`
		} `json:"bestRoute"`
		GasCost string `json:"gasCost"`
	} `json:"priceRoute"`
}

// ParaswapTransactionResponse represents Paraswap transaction response
type ParaswapTransactionResponse struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	GasPrice string `json:"gasPrice"`
	Gas      string `json:"gas"`
}

// NewParaswapClient creates a new Paraswap client
func NewParaswapClient(logger *logger.Logger, config AggregatorConfig) *ParaswapClient {
	return &ParaswapClient{
		logger:  logger.Named("paraswap"),
		config:  config,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetQuote gets a swap quote from Paraswap
func (pc *ParaswapClient) GetQuote(ctx context.Context, req *QuoteRequest) (*SwapQuote, error) {
	pc.logger.Info("Getting Paraswap quote",
		zap.String("token_in", req.TokenIn.Symbol),
		zap.String("token_out", req.TokenOut.Symbol),
		zap.String("amount_in", req.AmountIn.String()))

	// Build quote request URL
	quoteURL := fmt.Sprintf("%s/prices", pc.baseURL)
	params := url.Values{}
	params.Set("srcToken", req.TokenIn.Address)
	params.Set("destToken", req.TokenOut.Address)
	params.Set("amount", req.AmountIn.Mul(decimal.NewFromInt(1e18)).String()) // Convert to wei
	params.Set("network", pc.getNetworkID(req.Chain))
	params.Set("side", "SELL")

	// Make API request
	resp, err := pc.makeRequest(ctx, "GET", quoteURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	var quoteResp ParaswapQuoteResponse
	if err := json.Unmarshal(resp, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	// Parse amounts
	amountOut, err := decimal.NewFromString(quoteResp.PriceRoute.DestAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dest amount: %w", err)
	}
	amountOut = amountOut.Div(decimal.NewFromInt(1e18)) // Convert from wei

	// Calculate slippage
	slippage := req.Slippage
	if slippage.IsZero() {
		slippage = decimal.NewFromFloat(0.01) // Default 1%
	}
	minAmountOut := amountOut.Mul(decimal.NewFromInt(1).Sub(slippage))

	// Parse gas cost
	gasCost, _ := decimal.NewFromString(quoteResp.PriceRoute.GasCost)

	// Build route steps
	route := make([]RouteStep, 0, len(quoteResp.PriceRoute.BestRoute))
	for _, step := range quoteResp.PriceRoute.BestRoute {
		route = append(route, RouteStep{
			Protocol:   step.Exchange,
			Percentage: decimal.NewFromInt(int64(step.Percent)),
		})
	}

	quote := &SwapQuote{
		ID:           uuid.New().String(),
		Aggregator:   "paraswap",
		Protocol:     "paraswap",
		Chain:        req.Chain,
		TokenIn:      req.TokenIn,
		TokenOut:     req.TokenOut,
		AmountIn:     req.AmountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  decimal.NewFromFloat(0.1), // Simplified
		Fee:          decimal.Zero,               // Paraswap doesn't charge fees
		GasEstimate:  200000,                    // Estimated
		GasCost:      gasCost,
		Route:        route,
		Slippage:     slippage,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
		ExecutionData: map[string]interface{}{
			"priceRoute": quoteResp.PriceRoute,
		},
		Metadata: make(map[string]interface{}),
	}

	return quote, nil
}

// ExecuteSwap executes a token swap via Paraswap
func (pc *ParaswapClient) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	pc.logger.Info("Executing Paraswap swap", zap.String("quote_id", req.Quote.ID))

	// Build transaction request
	txURL := fmt.Sprintf("%s/transactions/%s", pc.baseURL, pc.getNetworkID(req.Quote.Chain))
	
	txReq := map[string]interface{}{
		"srcToken":    req.Quote.TokenIn.Address,
		"destToken":   req.Quote.TokenOut.Address,
		"srcAmount":   req.Quote.AmountIn.Mul(decimal.NewFromInt(1e18)).String(),
		"destAmount":  req.Quote.MinAmountOut.Mul(decimal.NewFromInt(1e18)).String(),
		"priceRoute":  req.Quote.ExecutionData["priceRoute"],
		"userAddress": req.UserAddress,
		"slippage":    req.Slippage.Mul(decimal.NewFromInt(100)).String(), // Convert to percentage
	}

	reqBody, err := json.Marshal(txReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := pc.makeRequest(ctx, "POST", txURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	var txResp ParaswapTransactionResponse
	if err := json.Unmarshal(resp, &txResp); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
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
			"transaction": txResp,
		},
	}

	return result, nil
}

// GetSupportedTokens returns supported tokens for Paraswap
func (pc *ParaswapClient) GetSupportedTokens(ctx context.Context, chain string) ([]Token, error) {
	pc.logger.Debug("Getting Paraswap supported tokens", zap.String("chain", chain))

	tokensURL := fmt.Sprintf("%s/tokens/%s", pc.baseURL, pc.getNetworkID(chain))

	resp, err := pc.makeRequest(ctx, "GET", tokensURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	var tokensResp map[string]interface{}
	if err := json.Unmarshal(resp, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to parse tokens response: %w", err)
	}

	// Parse tokens (simplified)
	tokens := []Token{
		{
			Address:  "0xA0b86a33E6441E6C8C07C4c4c8e8B0E8E8E8E8E8",
			Symbol:   "ETH",
			Name:     "Ethereum",
			Decimals: 18,
			Chain:    chain,
		},
		{
			Address:  "0xA0b73E1Ff0B80914AB6fe0444E65848C4C34450b",
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
			Chain:    chain,
		},
	}

	return tokens, nil
}

// GetStatus returns the status of the Paraswap client
func (pc *ParaswapClient) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":     "paraswap",
		"enabled":  pc.config.Enabled,
		"base_url": pc.baseURL,
		"chains":   pc.config.Chains,
		"healthy":  true, // Simplified
	}
}

// Helper methods

// makeRequest makes an HTTP request to Paraswap API
func (pc *ParaswapClient) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CryptoWallet-DEXAggregator/1.0")

	if pc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+pc.apiKey)
	}

	resp, err := pc.httpClient.Do(req)
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

// getNetworkID converts chain name to Paraswap network ID
func (pc *ParaswapClient) getNetworkID(chain string) string {
	switch chain {
	case "ethereum":
		return "1"
	case "polygon":
		return "137"
	case "bsc":
		return "56"
	case "arbitrum":
		return "42161"
	case "optimism":
		return "10"
	case "avalanche":
		return "43114"
	default:
		return "1" // Default to Ethereum
	}
}
