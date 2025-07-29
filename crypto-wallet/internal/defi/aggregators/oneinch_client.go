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

// OneInchClient handles interactions with 1inch DEX aggregator
type OneInchClient struct {
	logger     *logger.Logger
	config     AggregatorConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// OneInchQuoteResponse represents 1inch quote response
type OneInchQuoteResponse struct {
	ToTokenAmount   string `json:"toTokenAmount"`
	FromTokenAmount string `json:"fromTokenAmount"`
	Protocols       [][]struct {
		Name             string `json:"name"`
		Part             int    `json:"part"`
		FromTokenAddress string `json:"fromTokenAddress"`
		ToTokenAddress   string `json:"toTokenAddress"`
	} `json:"protocols"`
	EstimatedGas string `json:"estimatedGas"`
}

// OneInchSwapResponse represents 1inch swap response
type OneInchSwapResponse struct {
	FromToken struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"fromToken"`
	ToToken struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"toToken"`
	ToTokenAmount   string `json:"toTokenAmount"`
	FromTokenAmount string `json:"fromTokenAmount"`
	Protocols       [][]struct {
		Name string `json:"name"`
		Part int    `json:"part"`
	} `json:"protocols"`
	Tx struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Data     string `json:"data"`
		Value    string `json:"value"`
		GasPrice string `json:"gasPrice"`
		Gas      string `json:"gas"`
	} `json:"tx"`
}

// OneInchTokensResponse represents 1inch tokens response
type OneInchTokensResponse struct {
	Tokens map[string]struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"tokens"`
}

// NewOneInchClient creates a new 1inch client
func NewOneInchClient(logger *logger.Logger, config AggregatorConfig) *OneInchClient {
	return &OneInchClient{
		logger:  logger.Named("1inch"),
		config:  config,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetQuote gets a swap quote from 1inch
func (oc *OneInchClient) GetQuote(ctx context.Context, req *QuoteRequest) (*SwapQuote, error) {
	oc.logger.Info("Getting 1inch quote",
		zap.String("token_in", req.TokenIn.Symbol),
		zap.String("token_out", req.TokenOut.Symbol),
		zap.String("amount_in", req.AmountIn.String()))

	chainID := oc.getChainID(req.Chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", req.Chain)
	}

	// Build quote request URL
	quoteURL := fmt.Sprintf("%s/v5.0/%s/quote", oc.baseURL, chainID)
	params := url.Values{}
	params.Set("fromTokenAddress", req.TokenIn.Address)
	params.Set("toTokenAddress", req.TokenOut.Address)
	params.Set("amount", req.AmountIn.Mul(decimal.NewFromInt(1e18)).String()) // Convert to wei

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", quoteURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	var quoteResp OneInchQuoteResponse
	if err := json.Unmarshal(resp, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	// Parse amounts
	amountOut, err := decimal.NewFromString(quoteResp.ToTokenAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse to token amount: %w", err)
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

	// Build route steps from protocols
	route := make([]RouteStep, 0)
	for _, protocolGroup := range quoteResp.Protocols {
		for _, protocol := range protocolGroup {
			route = append(route, RouteStep{
				Protocol:   protocol.Name,
				Percentage: decimal.NewFromInt(int64(protocol.Part)),
			})
		}
	}

	quote := &SwapQuote{
		ID:           uuid.New().String(),
		Aggregator:   "1inch",
		Protocol:     "1inch",
		Chain:        req.Chain,
		TokenIn:      req.TokenIn,
		TokenOut:     req.TokenOut,
		AmountIn:     req.AmountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  decimal.NewFromFloat(0.1), // Simplified
		Fee:          decimal.Zero,              // 1inch doesn't charge fees
		GasEstimate:  uint64(gasEstimate.IntPart()),
		Route:        route,
		Slippage:     slippage,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
		ExecutionData: map[string]interface{}{
			"protocols": quoteResp.Protocols,
		},
		Metadata: make(map[string]interface{}),
	}

	return quote, nil
}

// ExecuteSwap executes a token swap via 1inch
func (oc *OneInchClient) ExecuteSwap(ctx context.Context, req *SwapRequest) (*SwapResult, error) {
	oc.logger.Info("Executing 1inch swap", zap.String("quote_id", req.Quote.ID))

	chainID := oc.getChainID(req.Quote.Chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", req.Quote.Chain)
	}

	// Build swap request URL
	swapURL := fmt.Sprintf("%s/v5.0/%s/swap", oc.baseURL, chainID)
	params := url.Values{}
	params.Set("fromTokenAddress", req.Quote.TokenIn.Address)
	params.Set("toTokenAddress", req.Quote.TokenOut.Address)
	params.Set("amount", req.Quote.AmountIn.Mul(decimal.NewFromInt(1e18)).String())
	params.Set("fromAddress", req.UserAddress)
	params.Set("slippage", req.Slippage.Mul(decimal.NewFromInt(100)).String()) // Convert to percentage

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", swapURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap transaction: %w", err)
	}

	var swapResp OneInchSwapResponse
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
			"transaction": swapResp.Tx,
		},
	}

	return result, nil
}

// GetSupportedTokens returns supported tokens for 1inch
func (oc *OneInchClient) GetSupportedTokens(ctx context.Context, chain string) ([]Token, error) {
	oc.logger.Debug("Getting 1inch supported tokens", zap.String("chain", chain))

	chainID := oc.getChainID(chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}

	tokensURL := fmt.Sprintf("%s/v5.0/%s/tokens", oc.baseURL, chainID)

	resp, err := oc.makeRequest(ctx, "GET", tokensURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	var tokensResp OneInchTokensResponse
	if err := json.Unmarshal(resp, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to parse tokens response: %w", err)
	}

	// Convert to our Token format
	tokens := make([]Token, 0, len(tokensResp.Tokens))
	for _, token := range tokensResp.Tokens {
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

// GetStatus returns the status of the 1inch client
func (oc *OneInchClient) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":     "1inch",
		"enabled":  oc.config.Enabled,
		"base_url": oc.baseURL,
		"chains":   oc.config.Chains,
		"healthy":  true, // Simplified
	}
}

// Helper methods

// makeRequest makes an HTTP request to 1inch API
func (oc *OneInchClient) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
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

	if oc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+oc.apiKey)
	}

	resp, err := oc.httpClient.Do(req)
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

// getChainID converts chain name to 1inch chain ID
func (oc *OneInchClient) getChainID(chain string) string {
	switch chain {
	case "ethereum":
		return "1"
	case "bsc":
		return "56"
	case "polygon":
		return "137"
	case "optimism":
		return "10"
	case "arbitrum":
		return "42161"
	case "gnosis":
		return "100"
	case "avalanche":
		return "43114"
	case "fantom":
		return "250"
	case "klaytn":
		return "8217"
	case "aurora":
		return "1313161554"
	default:
		return "" // Unsupported chain
	}
}
