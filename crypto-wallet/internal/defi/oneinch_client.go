package defi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// 1inch API endpoints
const (
	OneInchAPIBaseURL = "https://api.1inch.io/v5.0"
	OneInchEthereumChainID = "1"
	OneInchBSCChainID     = "56"
	OneInchPolygonChainID = "137"
)

// OneInchClient handles interactions with 1inch DEX aggregator
type OneInchClient struct {
	apiKey     string
	httpClient *http.Client
	logger     *logger.Logger
}

// NewOneInchClient creates a new 1inch client
func NewOneInchClient(apiKey string, logger *logger.Logger) *OneInchClient {
	return &OneInchClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.Named("1inch"),
	}
}

// GetSwapQuote gets a swap quote from 1inch
func (oc *OneInchClient) GetSwapQuote(ctx context.Context, req *GetSwapQuoteRequest) (*SwapQuote, error) {
	oc.logger.Info("Getting 1inch swap quote", 
		"tokenIn", req.TokenIn, 
		"tokenOut", req.TokenOut, 
		"amountIn", req.AmountIn,
		"chain", req.Chain)

	chainID := oc.getChainID(req.Chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", req.Chain)
	}

	// Build quote request URL
	quoteURL := fmt.Sprintf("%s/%s/quote", OneInchAPIBaseURL, chainID)
	params := url.Values{}
	params.Set("fromTokenAddress", req.TokenIn)
	params.Set("toTokenAddress", req.TokenOut)
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

	// Convert response to SwapQuote
	amountOut := decimal.NewFromBigInt(quoteResp.ToTokenAmount, -18)
	
	// Calculate slippage and minimum amount out
	slippage := req.Slippage
	if slippage.IsZero() {
		slippage = decimal.NewFromFloat(0.01) // 1% default
	}
	minAmountOut := amountOut.Mul(decimal.NewFromInt(1).Sub(slippage))

	quote := &SwapQuote{
		ID:           uuid.New().String(),
		Protocol:     ProtocolType1inch,
		Chain:        req.Chain,
		TokenIn: Token{
			Address: req.TokenIn,
			Chain:   req.Chain,
		},
		TokenOut: Token{
			Address: req.TokenOut,
			Chain:   req.Chain,
		},
		AmountIn:     req.AmountIn,
		AmountOut:    amountOut,
		MinAmountOut: minAmountOut,
		PriceImpact:  decimal.NewFromFloat(0.1), // Simplified
		Fee:          decimal.Zero, // 1inch doesn't charge fees
		GasEstimate:  uint64(quoteResp.EstimatedGas),
		Route:        []string{req.TokenIn, req.TokenOut}, // Simplified
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		CreatedAt:    time.Now(),
	}

	return quote, nil
}

// ExecuteSwap executes a token swap via 1inch
func (oc *OneInchClient) ExecuteSwap(ctx context.Context, quote *SwapQuote, walletID, passphrase string) (string, error) {
	oc.logger.Info("Executing 1inch swap", "quoteID", quote.ID)

	chainID := oc.getChainID(quote.Chain)
	if chainID == "" {
		return "", fmt.Errorf("unsupported chain: %s", quote.Chain)
	}

	// In a real implementation, you would:
	// 1. Get user's wallet address and private key
	// 2. Build swap transaction using 1inch API
	// 3. Sign and send transaction

	// Build swap request URL
	swapURL := fmt.Sprintf("%s/%s/swap", OneInchAPIBaseURL, chainID)
	params := url.Values{}
	params.Set("fromTokenAddress", quote.TokenIn.Address)
	params.Set("toTokenAddress", quote.TokenOut.Address)
	params.Set("amount", quote.AmountIn.Mul(decimal.NewFromInt(1e18)).String())
	params.Set("fromAddress", "0x742d35Cc6634C0532925a3b8D4C9db96590e4265") // Mock address
	params.Set("slippage", "1") // 1%

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", swapURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get swap transaction: %w", err)
	}

	var swapResp OneInchSwapResponse
	if err := json.Unmarshal(resp, &swapResp); err != nil {
		return "", fmt.Errorf("failed to parse swap response: %w", err)
	}

	// For now, return the transaction hash from the response
	// In a real implementation, you would sign and send the transaction
	return swapResp.Tx.Hash, nil
}

// GetSupportedTokens retrieves supported tokens for a chain
func (oc *OneInchClient) GetSupportedTokens(ctx context.Context, chain Chain) ([]Token, error) {
	oc.logger.Info("Getting supported tokens from 1inch", "chain", chain)

	chainID := oc.getChainID(chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}

	// Build tokens request URL
	tokensURL := fmt.Sprintf("%s/%s/tokens", OneInchAPIBaseURL, chainID)

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", tokensURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	var tokensResp OneInchTokensResponse
	if err := json.Unmarshal(resp, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to parse tokens response: %w", err)
	}

	// Convert response to Token slice
	var tokens []Token
	for address, tokenInfo := range tokensResp.Tokens {
		token := Token{
			Address:  address,
			Symbol:   tokenInfo.Symbol,
			Name:     tokenInfo.Name,
			Decimals: tokenInfo.Decimals,
			Chain:    chain,
			LogoURL:  tokenInfo.LogoURI,
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// GetLiquiditySources retrieves available liquidity sources
func (oc *OneInchClient) GetLiquiditySources(ctx context.Context, chain Chain) ([]string, error) {
	oc.logger.Info("Getting liquidity sources from 1inch", "chain", chain)

	chainID := oc.getChainID(chain)
	if chainID == "" {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}

	// Build liquidity sources request URL
	sourcesURL := fmt.Sprintf("%s/%s/liquidity-sources", OneInchAPIBaseURL, chainID)

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", sourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get liquidity sources: %w", err)
	}

	var sourcesResp OneInchLiquiditySourcesResponse
	if err := json.Unmarshal(resp, &sourcesResp); err != nil {
		return nil, fmt.Errorf("failed to parse liquidity sources response: %w", err)
	}

	var sources []string
	for _, protocol := range sourcesResp.Protocols {
		sources = append(sources, protocol.ID)
	}

	return sources, nil
}

// GetSpender retrieves the spender address for token approvals
func (oc *OneInchClient) GetSpender(ctx context.Context, chain Chain) (string, error) {
	oc.logger.Info("Getting spender address from 1inch", "chain", chain)

	chainID := oc.getChainID(chain)
	if chainID == "" {
		return "", fmt.Errorf("unsupported chain: %s", chain)
	}

	// Build spender request URL
	spenderURL := fmt.Sprintf("%s/%s/approve/spender", OneInchAPIBaseURL, chainID)

	// Make API request
	resp, err := oc.makeRequest(ctx, "GET", spenderURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get spender: %w", err)
	}

	var spenderResp OneInchSpenderResponse
	if err := json.Unmarshal(resp, &spenderResp); err != nil {
		return "", fmt.Errorf("failed to parse spender response: %w", err)
	}

	return spenderResp.Address, nil
}

// Helper methods

// makeRequest makes an HTTP request to 1inch API
func (oc *OneInchClient) makeRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Add API key if available
	if oc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+oc.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// getChainID maps Chain to 1inch chain ID
func (oc *OneInchClient) getChainID(chain Chain) string {
	switch chain {
	case ChainEthereum:
		return OneInchEthereumChainID
	case ChainBSC:
		return OneInchBSCChainID
	case ChainPolygon:
		return OneInchPolygonChainID
	default:
		return ""
	}
}

// Data structures for 1inch API responses

// OneInchQuoteResponse represents a quote response from 1inch
type OneInchQuoteResponse struct {
	FromToken     OneInchToken `json:"fromToken"`
	ToToken       OneInchToken `json:"toToken"`
	ToTokenAmount string       `json:"toTokenAmount"`
	FromTokenAmount string     `json:"fromTokenAmount"`
	EstimatedGas  int          `json:"estimatedGas"`
}

// OneInchSwapResponse represents a swap response from 1inch
type OneInchSwapResponse struct {
	FromToken OneInchToken `json:"fromToken"`
	ToToken   OneInchToken `json:"toToken"`
	Tx        OneInchTx    `json:"tx"`
}

// OneInchTx represents transaction data from 1inch
type OneInchTx struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Hash     string `json:"hash"`
}

// OneInchToken represents token data from 1inch
type OneInchToken struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI"`
}

// OneInchTokensResponse represents tokens response from 1inch
type OneInchTokensResponse struct {
	Tokens map[string]OneInchToken `json:"tokens"`
}

// OneInchLiquiditySourcesResponse represents liquidity sources response from 1inch
type OneInchLiquiditySourcesResponse struct {
	Protocols []OneInchProtocol `json:"protocols"`
}

// OneInchProtocol represents a protocol from 1inch
type OneInchProtocol struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Image string `json:"img"`
}

// OneInchSpenderResponse represents spender response from 1inch
type OneInchSpenderResponse struct {
	Address string `json:"address"`
}
