package defi

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// Chainlink price feed addresses on Ethereum mainnet
var ChainlinkPriceFeeds = map[string]string{
	"ETH/USD":  "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419",
	"BTC/USD":  "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c",
	"USDC/USD": "0x8fFfFfd4AfB6115b954Bd326cbe7B4BA576818f6",
	"USDT/USD": "0x3E7d1eAB13ad0104d2750B8863b489D65364e32D",
	"LINK/USD": "0x2c1d072e956AFFC0D435Cb7AC38EF18d24d9127c",
	"UNI/USD":  "0x553303d460EE0afB37EdFf9bE42922D8FF63220e",
	"AAVE/USD": "0x547a514d5e3769680Ce22B2361c10Ea13619e8a9",
}

// ChainlinkClient handles interactions with Chainlink price feeds
type ChainlinkClient struct {
	client *blockchain.EthereumClient
	logger *logger.Logger
	
	// Contract ABI
	aggregatorABI abi.ABI
}

// NewChainlinkClient creates a new Chainlink client
func NewChainlinkClient(client *blockchain.EthereumClient, logger *logger.Logger) *ChainlinkClient {
	cc := &ChainlinkClient{
		client: client,
		logger: logger.Named("chainlink"),
	}

	// Load contract ABI
	cc.loadABI()

	return cc
}

// GetTokenPrice retrieves token price from Chainlink price feed
func (cc *ChainlinkClient) GetTokenPrice(ctx context.Context, tokenAddress string) (decimal.Decimal, error) {
	cc.logger.Info("Getting token price from Chainlink", "tokenAddress", tokenAddress)

	// Map token address to price feed
	priceFeedAddress, err := cc.getPriceFeedAddress(tokenAddress)
	if err != nil {
		return decimal.Zero, err
	}

	// Get latest price from price feed
	price, err := cc.getLatestPrice(ctx, priceFeedAddress)
	if err != nil {
		return decimal.Zero, err
	}

	cc.logger.Info("Retrieved token price", 
		"tokenAddress", tokenAddress, 
		"price", price,
		"priceFeed", priceFeedAddress)

	return price, nil
}

// GetLatestRoundData retrieves the latest round data from a price feed
func (cc *ChainlinkClient) GetLatestRoundData(ctx context.Context, priceFeedAddress string) (*ChainlinkRoundData, error) {
	cc.logger.Info("Getting latest round data from Chainlink", "priceFeed", priceFeedAddress)

	// In a real implementation, you would call latestRoundData() on the price feed contract
	// For now, return mock data
	roundData := &ChainlinkRoundData{
		RoundID:         big.NewInt(18446744073709562300),
		Answer:          big.NewInt(250000000000), // $2500.00 with 8 decimals
		StartedAt:       big.NewInt(time.Now().Unix()),
		UpdatedAt:       big.NewInt(time.Now().Unix()),
		AnsweredInRound: big.NewInt(18446744073709562300),
	}

	return roundData, nil
}

// GetHistoricalPrice retrieves historical price data
func (cc *ChainlinkClient) GetHistoricalPrice(ctx context.Context, tokenAddress string, timestamp time.Time) (decimal.Decimal, error) {
	cc.logger.Info("Getting historical price from Chainlink", 
		"tokenAddress", tokenAddress, 
		"timestamp", timestamp)

	// Map token address to price feed
	priceFeedAddress, err := cc.getPriceFeedAddress(tokenAddress)
	if err != nil {
		return decimal.Zero, err
	}

	// In a real implementation, you would:
	// 1. Find the round closest to the timestamp
	// 2. Call getRoundData() with that round ID
	// For now, return current price with some variation
	currentPrice, err := cc.getLatestPrice(ctx, priceFeedAddress)
	if err != nil {
		return decimal.Zero, err
	}

	// Add some historical variation (mock)
	variation := decimal.NewFromFloat(0.95) // 5% lower for historical data
	historicalPrice := currentPrice.Mul(variation)

	return historicalPrice, nil
}

// GetPriceFeedInfo retrieves information about a price feed
func (cc *ChainlinkClient) GetPriceFeedInfo(ctx context.Context, priceFeedAddress string) (*ChainlinkPriceFeedInfo, error) {
	cc.logger.Info("Getting price feed info from Chainlink", "priceFeed", priceFeedAddress)

	// In a real implementation, you would call various methods on the price feed contract
	// For now, return mock data
	feedInfo := &ChainlinkPriceFeedInfo{
		Description: "ETH / USD",
		Decimals:    8,
		Version:     4,
		Address:     priceFeedAddress,
	}

	return feedInfo, nil
}

// GetMultiplePrices retrieves prices for multiple tokens
func (cc *ChainlinkClient) GetMultiplePrices(ctx context.Context, tokenAddresses []string) (map[string]decimal.Decimal, error) {
	cc.logger.Info("Getting multiple token prices from Chainlink", "count", len(tokenAddresses))

	prices := make(map[string]decimal.Decimal)

	for _, tokenAddress := range tokenAddresses {
		price, err := cc.GetTokenPrice(ctx, tokenAddress)
		if err != nil {
			cc.logger.Warn("Failed to get price for token", 
				"tokenAddress", tokenAddress, 
				"error", err)
			continue
		}
		prices[tokenAddress] = price
	}

	return prices, nil
}

// SubscribeToPriceFeed subscribes to price feed updates (WebSocket)
func (cc *ChainlinkClient) SubscribeToPriceFeed(ctx context.Context, priceFeedAddress string, callback func(*ChainlinkRoundData)) error {
	cc.logger.Info("Subscribing to price feed updates", "priceFeed", priceFeedAddress)

	// In a real implementation, you would:
	// 1. Subscribe to AnswerUpdated events
	// 2. Listen for new rounds
	// 3. Call callback with new data

	// For now, simulate periodic updates
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				roundData, err := cc.GetLatestRoundData(ctx, priceFeedAddress)
				if err != nil {
					cc.logger.Error("Failed to get latest round data", "error", err)
					continue
				}
				callback(roundData)
			}
		}
	}()

	return nil
}

// Helper methods

// getLatestPrice retrieves the latest price from a price feed
func (cc *ChainlinkClient) getLatestPrice(ctx context.Context, priceFeedAddress string) (decimal.Decimal, error) {
	// In a real implementation, you would call latestRoundData() on the price feed contract
	// For now, return mock prices based on the feed address

	mockPrices := map[string]decimal.Decimal{
		ChainlinkPriceFeeds["ETH/USD"]:  decimal.NewFromFloat(2500.00),
		ChainlinkPriceFeeds["BTC/USD"]:  decimal.NewFromFloat(45000.00),
		ChainlinkPriceFeeds["USDC/USD"]: decimal.NewFromFloat(1.00),
		ChainlinkPriceFeeds["USDT/USD"]: decimal.NewFromFloat(1.00),
		ChainlinkPriceFeeds["LINK/USD"]: decimal.NewFromFloat(15.50),
		ChainlinkPriceFeeds["UNI/USD"]:  decimal.NewFromFloat(8.75),
		ChainlinkPriceFeeds["AAVE/USD"]: decimal.NewFromFloat(95.25),
	}

	if price, exists := mockPrices[priceFeedAddress]; exists {
		return price, nil
	}

	// Default price for unknown feeds
	return decimal.NewFromFloat(100.00), nil
}

// getPriceFeedAddress maps token address to price feed address
func (cc *ChainlinkClient) getPriceFeedAddress(tokenAddress string) (string, error) {
	// Normalize address
	addr := common.HexToAddress(tokenAddress).Hex()

	// Map common token addresses to price feeds
	tokenToPriceFeed := map[string]string{
		"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2": ChainlinkPriceFeeds["ETH/USD"],  // WETH
		"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599": ChainlinkPriceFeeds["BTC/USD"],  // WBTC
		"0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1": ChainlinkPriceFeeds["USDC/USD"], // USDC
		"0xdAC17F958D2ee523a2206206994597C13D831ec7": ChainlinkPriceFeeds["USDT/USD"], // USDT
		"0x514910771AF9Ca656af840dff83E8264EcF986CA": ChainlinkPriceFeeds["LINK/USD"], // LINK
		"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984": ChainlinkPriceFeeds["UNI/USD"],  // UNI
		"0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9": ChainlinkPriceFeeds["AAVE/USD"], // AAVE
	}

	if priceFeed, exists := tokenToPriceFeed[addr]; exists {
		return priceFeed, nil
	}

	return "", fmt.Errorf("price feed not found for token: %s", tokenAddress)
}

// loadABI loads the Chainlink aggregator ABI
func (cc *ChainlinkClient) loadABI() {
	// Chainlink Aggregator ABI (simplified)
	aggregatorABIJSON := `[
		{"inputs":[],"name":"latestRoundData","outputs":[{"internalType":"uint80","name":"roundId","type":"uint80"},{"internalType":"int256","name":"answer","type":"int256"},{"internalType":"uint256","name":"startedAt","type":"uint256"},{"internalType":"uint256","name":"updatedAt","type":"uint256"},{"internalType":"uint80","name":"answeredInRound","type":"uint80"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"internalType":"uint80","name":"_roundId","type":"uint80"}],"name":"getRoundData","outputs":[{"internalType":"uint80","name":"roundId","type":"uint80"},{"internalType":"int256","name":"answer","type":"int256"},{"internalType":"uint256","name":"startedAt","type":"uint256"},{"internalType":"uint256","name":"updatedAt","type":"uint256"},{"internalType":"uint80","name":"answeredInRound","type":"uint80"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"description","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"version","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
	]`

	var err error
	cc.aggregatorABI, err = abi.JSON(strings.NewReader(aggregatorABIJSON))
	if err != nil {
		cc.logger.Error("Failed to parse aggregator ABI", "error", err)
	}
}

// Data structures for Chainlink

// ChainlinkRoundData represents round data from a Chainlink price feed
type ChainlinkRoundData struct {
	RoundID         *big.Int `json:"round_id"`
	Answer          *big.Int `json:"answer"`
	StartedAt       *big.Int `json:"started_at"`
	UpdatedAt       *big.Int `json:"updated_at"`
	AnsweredInRound *big.Int `json:"answered_in_round"`
}

// ChainlinkPriceFeedInfo represents information about a Chainlink price feed
type ChainlinkPriceFeedInfo struct {
	Description string `json:"description"`
	Decimals    uint8  `json:"decimals"`
	Version     uint64 `json:"version"`
	Address     string `json:"address"`
}
