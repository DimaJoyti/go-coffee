package defi

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// UniswapPriceProvider implements PriceProvider interface for Uniswap
type UniswapPriceProvider struct {
	client *UniswapClient
	logger *zap.Logger
}

// NewUniswapPriceProvider creates a new Uniswap price provider
func NewUniswapPriceProvider(client *UniswapClient) *UniswapPriceProvider {
	return &UniswapPriceProvider{
		client: client,
		logger: zap.L().Named("uniswap-price-provider"),
	}
}

// GetPrice implements PriceProvider interface
func (upp *UniswapPriceProvider) GetPrice(ctx context.Context, token Token) (decimal.Decimal, error) {
	upp.logger.Info("Getting price from Uniswap",
		zap.String("token", token.Address),
		zap.String("symbol", token.Symbol),
		zap.String("chain", string(token.Chain)))

	// Use USDC as base token for price calculation
	usdcAddress := upp.getUSDCAddress(token.Chain)
	if usdcAddress == "" {
		return decimal.Zero, fmt.Errorf("USDC not supported on chain %s", token.Chain)
	}

	// If token is already USDC, return 1.0
	if token.Address == usdcAddress {
		return decimal.NewFromFloat(1.0), nil
	}

	// Create swap quote request to get price
	// Use 1 token as amount to get price per token
	amountIn := decimal.NewFromFloat(1.0)
	
	req := &GetSwapQuoteRequest{
		TokenIn:  token.Address,
		TokenOut: usdcAddress,
		AmountIn: amountIn,
		Chain:    token.Chain,
		Slippage: decimal.NewFromFloat(0.01), // 1% slippage
	}

	quote, err := upp.client.GetSwapQuote(ctx, req)
	if err != nil {
		upp.logger.Error("Failed to get Uniswap quote", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get price from Uniswap: %w", err)
	}

	// Price is AmountOut / AmountIn
	price := quote.AmountOut.Div(amountIn)
	
	upp.logger.Info("Got price from Uniswap",
		zap.String("token", token.Symbol),
		zap.String("price", price.String()))

	return price, nil
}

// GetExchangeInfo implements PriceProvider interface
func (upp *UniswapPriceProvider) GetExchangeInfo() Exchange {
	return Exchange{
		ID:       "uniswap-v3",
		Name:     "Uniswap V3",
		Type:     ExchangeTypeDEX,
		Chain:    ChainEthereum,
		Protocol: ProtocolTypeUniswap,
		Fee:      decimal.NewFromFloat(0.003), // 0.3% fee
		Active:   true,
	}
}

// IsHealthy implements PriceProvider interface
func (upp *UniswapPriceProvider) IsHealthy(ctx context.Context) bool {
	// Try to get a simple quote to test connectivity
	testToken := Token{
		Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Symbol:  "WETH",
		Chain:   ChainEthereum,
	}

	_, err := upp.GetPrice(ctx, testToken)
	if err != nil {
		upp.logger.Warn("Uniswap health check failed", zap.Error(err))
		return false
	}

	return true
}

// getUSDCAddress returns USDC address for the given chain
func (upp *UniswapPriceProvider) getUSDCAddress(chain Chain) string {
	switch chain {
	case ChainEthereum:
		return "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1" // USDC on Ethereum
	case ChainBSC:
		return "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d" // USDC on BSC
	case ChainPolygon:
		return "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174" // USDC on Polygon
	default:
		return ""
	}
}

// OneInchPriceProvider implements PriceProvider interface for 1inch
type OneInchPriceProvider struct {
	client *OneInchClient
	logger *zap.Logger
}

// NewOneInchPriceProvider creates a new 1inch price provider
func NewOneInchPriceProvider(client *OneInchClient) *OneInchPriceProvider {
	return &OneInchPriceProvider{
		client: client,
		logger: zap.L().Named("1inch-price-provider"),
	}
}

// GetPrice implements PriceProvider interface
func (oipp *OneInchPriceProvider) GetPrice(ctx context.Context, token Token) (decimal.Decimal, error) {
	oipp.logger.Info("Getting price from 1inch",
		zap.String("token", token.Address),
		zap.String("symbol", token.Symbol),
		zap.String("chain", string(token.Chain)))

	// Use USDC as base token for price calculation
	usdcAddress := oipp.getUSDCAddress(token.Chain)
	if usdcAddress == "" {
		return decimal.Zero, fmt.Errorf("USDC not supported on chain %s", token.Chain)
	}

	// If token is already USDC, return 1.0
	if token.Address == usdcAddress {
		return decimal.NewFromFloat(1.0), nil
	}

	// Create swap quote request to get price
	// Use 1 token as amount to get price per token
	amountIn := decimal.NewFromFloat(1.0)
	
	req := &GetSwapQuoteRequest{
		TokenIn:  token.Address,
		TokenOut: usdcAddress,
		AmountIn: amountIn,
		Chain:    token.Chain,
		Slippage: decimal.NewFromFloat(0.01), // 1% slippage
	}

	quote, err := oipp.client.GetSwapQuote(ctx, req)
	if err != nil {
		oipp.logger.Error("Failed to get 1inch quote", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get price from 1inch: %w", err)
	}

	// Price is AmountOut / AmountIn
	price := quote.AmountOut.Div(amountIn)
	
	oipp.logger.Info("Got price from 1inch",
		zap.String("token", token.Symbol),
		zap.String("price", price.String()))

	return price, nil
}

// GetExchangeInfo implements PriceProvider interface
func (oipp *OneInchPriceProvider) GetExchangeInfo() Exchange {
	return Exchange{
		ID:       "1inch-aggregator",
		Name:     "1inch Aggregator",
		Type:     ExchangeTypeDEX,
		Chain:    ChainEthereum,
		Protocol: ProtocolType1inch,
		Fee:      decimal.Zero, // 1inch doesn't charge fees
		Active:   true,
	}
}

// IsHealthy implements PriceProvider interface
func (oipp *OneInchPriceProvider) IsHealthy(ctx context.Context) bool {
	// Try to get a simple quote to test connectivity
	testToken := Token{
		Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Symbol:  "WETH",
		Chain:   ChainEthereum,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := oipp.GetPrice(ctx, testToken)
	if err != nil {
		oipp.logger.Warn("1inch health check failed", zap.Error(err))
		return false
	}

	return true
}

// getUSDCAddress returns USDC address for the given chain
func (oipp *OneInchPriceProvider) getUSDCAddress(chain Chain) string {
	switch chain {
	case ChainEthereum:
		return "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1" // USDC on Ethereum
	case ChainBSC:
		return "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d" // USDC on BSC
	case ChainPolygon:
		return "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174" // USDC on Polygon
	default:
		return ""
	}
}

// ChainlinkPriceProvider implements PriceProvider interface for Chainlink
type ChainlinkPriceProvider struct {
	client *ChainlinkClient
	logger *zap.Logger
}

// NewChainlinkPriceProvider creates a new Chainlink price provider
func NewChainlinkPriceProvider(client *ChainlinkClient) *ChainlinkPriceProvider {
	return &ChainlinkPriceProvider{
		client: client,
		logger: zap.L().Named("chainlink-price-provider"),
	}
}

// GetPrice implements PriceProvider interface
func (cpp *ChainlinkPriceProvider) GetPrice(ctx context.Context, token Token) (decimal.Decimal, error) {
	cpp.logger.Info("Getting price from Chainlink",
		zap.String("token", token.Address),
		zap.String("symbol", token.Symbol))

	price, err := cpp.client.GetTokenPrice(ctx, token.Address)
	if err != nil {
		cpp.logger.Error("Failed to get Chainlink price", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get price from Chainlink: %w", err)
	}

	cpp.logger.Info("Got price from Chainlink",
		zap.String("token", token.Symbol),
		zap.String("price", price.String()))

	return price, nil
}

// GetExchangeInfo implements PriceProvider interface
func (cpp *ChainlinkPriceProvider) GetExchangeInfo() Exchange {
	return Exchange{
		ID:       "chainlink-oracle",
		Name:     "Chainlink Price Feeds",
		Type:     ExchangeTypeOracle,
		Chain:    ChainEthereum,
		Protocol: "chainlink",
		Fee:      decimal.Zero,
		Active:   true,
	}
}

// IsHealthy implements PriceProvider interface
func (cpp *ChainlinkPriceProvider) IsHealthy(ctx context.Context) bool {
	// Try to get ETH price as a health check
	testToken := Token{
		Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Symbol:  "WETH",
		Chain:   ChainEthereum,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := cpp.GetPrice(ctx, testToken)
	if err != nil {
		cpp.logger.Warn("Chainlink health check failed", zap.Error(err))
		return false
	}

	return true
}
