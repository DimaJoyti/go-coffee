package aggregators

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultDEXAggregatorConfig returns default DEX aggregator configuration
func GetDefaultDEXAggregatorConfig() DEXAggregatorConfig {
	return DEXAggregatorConfig{
		Enabled:             true,
		DefaultSlippage:     decimal.NewFromFloat(0.01), // 1%
		MaxSlippage:         decimal.NewFromFloat(0.05), // 5%
		QuoteTimeout:        30 * time.Second,
		CacheTimeout:        5 * time.Minute,
		MaxRetries:          3,
		HealthCheckInterval: 1 * time.Minute,
		Aggregators: map[string]AggregatorConfig{
			"1inch": {
				Enabled:   true,
				BaseURL:   "https://api.1inch.io",
				Timeout:   30 * time.Second,
				RateLimit: 10,
				Priority:  1,
				Chains:    []string{"ethereum", "bsc", "polygon", "optimism", "arbitrum"},
				MinAmount: decimal.NewFromFloat(0.001),
				MaxAmount: decimal.NewFromFloat(1000000),
			},
			"paraswap": {
				Enabled:   true,
				BaseURL:   "https://apiv5.paraswap.io",
				Timeout:   30 * time.Second,
				RateLimit: 10,
				Priority:  2,
				Chains:    []string{"ethereum", "bsc", "polygon", "avalanche"},
				MinAmount: decimal.NewFromFloat(0.001),
				MaxAmount: decimal.NewFromFloat(1000000),
			},
			"0x": {
				Enabled:   true,
				BaseURL:   "https://api.0x.org",
				Timeout:   30 * time.Second,
				RateLimit: 10,
				Priority:  3,
				Chains:    []string{"ethereum", "bsc", "polygon"},
				MinAmount: decimal.NewFromFloat(0.001),
				MaxAmount: decimal.NewFromFloat(1000000),
			},
			"matcha": {
				Enabled:   true,
				BaseURL:   "https://api.matcha.xyz",
				Timeout:   30 * time.Second,
				RateLimit: 10,
				Priority:  4,
				Chains:    []string{"ethereum", "bsc", "polygon"},
				MinAmount: decimal.NewFromFloat(0.001),
				MaxAmount: decimal.NewFromFloat(1000000),
			},
		},
		RoutingConfig: RoutingConfig{
			Strategy:             RoutingStrategyBestPrice,
			ParallelQuotes:       true,
			FallbackEnabled:      true,
			PriceImpactThreshold: decimal.NewFromFloat(0.03), // 3%
			GasOptimization:      true,
			MinSavingsThreshold:  decimal.NewFromFloat(0.001), // 0.1%
		},
	}
}

// GetSupportedChains returns all supported chains across aggregators
func GetSupportedChains() []string {
	return []string{
		"ethereum",
		"bsc",
		"polygon",
		"optimism",
		"arbitrum",
		"avalanche",
		"fantom",
		"gnosis",
		"klaytn",
		"aurora",
	}
}

// GetChainInfo returns information about supported chains
func GetChainInfo() map[string]ChainInfo {
	return map[string]ChainInfo{
		"ethereum": {
			ChainID:     1,
			Name:        "Ethereum",
			Symbol:      "ETH",
			RPC:         "https://mainnet.infura.io/v3/",
			Explorer:    "https://etherscan.io",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch", "paraswap", "0x", "matcha"},
		},
		"bsc": {
			ChainID:     56,
			Name:        "Binance Smart Chain",
			Symbol:      "BNB",
			RPC:         "https://bsc-dataseed.binance.org/",
			Explorer:    "https://bscscan.com",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch", "paraswap", "0x", "matcha"},
		},
		"polygon": {
			ChainID:     137,
			Name:        "Polygon",
			Symbol:      "MATIC",
			RPC:         "https://polygon-rpc.com/",
			Explorer:    "https://polygonscan.com",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch", "paraswap", "0x", "matcha"},
		},
		"optimism": {
			ChainID:     10,
			Name:        "Optimism",
			Symbol:      "ETH",
			RPC:         "https://mainnet.optimism.io",
			Explorer:    "https://optimistic.etherscan.io",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch"},
		},
		"arbitrum": {
			ChainID:     42161,
			Name:        "Arbitrum One",
			Symbol:      "ETH",
			RPC:         "https://arb1.arbitrum.io/rpc",
			Explorer:    "https://arbiscan.io",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch"},
		},
		"avalanche": {
			ChainID:     43114,
			Name:        "Avalanche",
			Symbol:      "AVAX",
			RPC:         "https://api.avax.network/ext/bc/C/rpc",
			Explorer:    "https://snowtrace.io",
			NativeToken: "0x0000000000000000000000000000000000000000",
			Aggregators: []string{"1inch", "paraswap"},
		},
	}
}

// ChainInfo holds information about a blockchain
type ChainInfo struct {
	ChainID     int      `json:"chain_id"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	RPC         string   `json:"rpc"`
	Explorer    string   `json:"explorer"`
	NativeToken string   `json:"native_token"`
	Aggregators []string `json:"aggregators"`
}

// GetCommonTokens returns common tokens for each chain
func GetCommonTokens() map[string][]Token {
	return map[string][]Token{
		"ethereum": {
			{
				Address:  "0x0000000000000000000000000000000000000000",
				Symbol:   "ETH",
				Name:     "Ethereum",
				Decimals: 18,
				Chain:    "ethereum",
			},
			{
				Address:  "0xA0b86a33E6441E6C8C07C4c4c8e8B0E8E8E8E8E8",
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
				Chain:    "ethereum",
			},
			{
				Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				Symbol:   "USDT",
				Name:     "Tether USD",
				Decimals: 6,
				Chain:    "ethereum",
			},
			{
				Address:  "0x6B175474E89094C44Da98b954EedeAC495271d0F",
				Symbol:   "DAI",
				Name:     "Dai Stablecoin",
				Decimals: 18,
				Chain:    "ethereum",
			},
			{
				Address:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
				Symbol:   "WETH",
				Name:     "Wrapped Ethereum",
				Decimals: 18,
				Chain:    "ethereum",
			},
		},
		"bsc": {
			{
				Address:  "0x0000000000000000000000000000000000000000",
				Symbol:   "BNB",
				Name:     "Binance Coin",
				Decimals: 18,
				Chain:    "bsc",
			},
			{
				Address:  "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 18,
				Chain:    "bsc",
			},
			{
				Address:  "0x55d398326f99059fF775485246999027B3197955",
				Symbol:   "USDT",
				Name:     "Tether USD",
				Decimals: 18,
				Chain:    "bsc",
			},
		},
		"polygon": {
			{
				Address:  "0x0000000000000000000000000000000000000000",
				Symbol:   "MATIC",
				Name:     "Polygon",
				Decimals: 18,
				Chain:    "polygon",
			},
			{
				Address:  "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
				Chain:    "polygon",
			},
			{
				Address:  "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
				Symbol:   "USDT",
				Name:     "Tether USD",
				Decimals: 6,
				Chain:    "polygon",
			},
		},
	}
}

// GetRoutingStrategies returns available routing strategies
func GetRoutingStrategies() []RoutingStrategy {
	return []RoutingStrategy{
		RoutingStrategyBestPrice,
		RoutingStrategyLowestGas,
		RoutingStrategyBestValue,
		RoutingStrategyFastest,
		RoutingStrategyMostLiquid,
		RoutingStrategyBalanced,
	}
}

// GetStrategyDescription returns description for routing strategies
func GetStrategyDescription() map[RoutingStrategy]string {
	return map[RoutingStrategy]string{
		RoutingStrategyBestPrice:  "Maximizes output token amount",
		RoutingStrategyLowestGas:  "Minimizes gas costs",
		RoutingStrategyBestValue:  "Optimizes for best value (output - gas)",
		RoutingStrategyFastest:    "Prioritizes execution speed",
		RoutingStrategyMostLiquid: "Minimizes price impact",
		RoutingStrategyBalanced:   "Balances all factors with weighted scoring",
	}
}

// ValidateDEXAggregatorConfig validates DEX aggregator configuration
func ValidateDEXAggregatorConfig(config DEXAggregatorConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if config.DefaultSlippage.LessThan(decimal.Zero) || config.DefaultSlippage.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("default slippage must be between 0 and 1")
	}

	if config.MaxSlippage.LessThan(config.DefaultSlippage) {
		return fmt.Errorf("max slippage must be greater than or equal to default slippage")
	}

	if config.QuoteTimeout <= 0 {
		return fmt.Errorf("quote timeout must be positive")
	}

	if config.CacheTimeout <= 0 {
		return fmt.Errorf("cache timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	// Validate aggregator configs
	for name, aggConfig := range config.Aggregators {
		if err := validateAggregatorConfig(name, aggConfig); err != nil {
			return fmt.Errorf("invalid config for %s: %w", name, err)
		}
	}

	return nil
}

// validateAggregatorConfig validates individual aggregator configuration
func validateAggregatorConfig(name string, config AggregatorConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}

	if config.Priority <= 0 {
		return fmt.Errorf("priority must be positive")
	}

	if len(config.Chains) == 0 {
		return fmt.Errorf("at least one chain must be specified")
	}

	if config.MinAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("min amount cannot be negative")
	}

	if config.MaxAmount.LessThan(config.MinAmount) {
		return fmt.Errorf("max amount must be greater than or equal to min amount")
	}

	return nil
}
