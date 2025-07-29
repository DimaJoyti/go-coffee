package multichain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetDefaultMultichainConfig returns default multichain configuration
func GetDefaultMultichainConfig() MultichainConfig {
	return MultichainConfig{
		Enabled:          true,
		SupportedChains:  []string{"ethereum", "bsc", "polygon", "arbitrum", "optimism", "avalanche"},
		DefaultChain:     "ethereum",
		UpdateInterval:   30 * time.Second,
		BalanceThreshold: decimal.NewFromFloat(0.001),
		ChainConfigs: map[string]ChainConfig{
			"ethereum": {
				ChainID:      1,
				Name:         "Ethereum",
				Symbol:       "ETH",
				RPCEndpoints: []string{"https://mainnet.infura.io/v3/YOUR_PROJECT_ID"},
				WSEndpoints:  []string{"wss://mainnet.infura.io/ws/v3/YOUR_PROJECT_ID"},
				ExplorerURL:  "https://etherscan.io",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "ETH",
					Name:        "Ethereum",
					Decimals:    18,
					CoingeckoID: "ethereum",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0xA0b86a33E6441E6C8C07C4c4c8e8B0E8E8E8E8E8",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    6,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
					{
						Address:     "0xdAC17F958D2ee523a2206206994597C13D831ec7",
						Symbol:      "USDT",
						Name:        "Tether USD",
						Decimals:    6,
						CoingeckoID: "tether",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.1),
				MaxGasPrice:        decimal.NewFromFloat(100),
				ConfirmationBlocks: 12,
				Enabled:            true,
				Priority:           1,
			},
			"bsc": {
				ChainID:      56,
				Name:         "Binance Smart Chain",
				Symbol:       "BNB",
				RPCEndpoints: []string{"https://bsc-dataseed.binance.org/"},
				WSEndpoints:  []string{"wss://bsc-ws-node.nariox.org:443/ws"},
				ExplorerURL:  "https://bscscan.com",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "BNB",
					Name:        "Binance Coin",
					Decimals:    18,
					CoingeckoID: "binancecoin",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    18,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.1),
				MaxGasPrice:        decimal.NewFromFloat(20),
				ConfirmationBlocks: 3,
				Enabled:            true,
				Priority:           2,
			},
			"polygon": {
				ChainID:      137,
				Name:         "Polygon",
				Symbol:       "MATIC",
				RPCEndpoints: []string{"https://polygon-rpc.com/"},
				WSEndpoints:  []string{"wss://polygon-rpc.com/"},
				ExplorerURL:  "https://polygonscan.com",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "MATIC",
					Name:        "Polygon",
					Decimals:    18,
					CoingeckoID: "matic-network",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    6,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.2),
				MaxGasPrice:        decimal.NewFromFloat(500),
				ConfirmationBlocks: 3,
				Enabled:            true,
				Priority:           3,
			},
			"arbitrum": {
				ChainID:      42161,
				Name:         "Arbitrum One",
				Symbol:       "ETH",
				RPCEndpoints: []string{"https://arb1.arbitrum.io/rpc"},
				WSEndpoints:  []string{"wss://arb1.arbitrum.io/ws"},
				ExplorerURL:  "https://arbiscan.io",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "ETH",
					Name:        "Ethereum",
					Decimals:    18,
					CoingeckoID: "ethereum",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    6,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.1),
				MaxGasPrice:        decimal.NewFromFloat(10),
				ConfirmationBlocks: 1,
				Enabled:            true,
				Priority:           4,
			},
			"optimism": {
				ChainID:      10,
				Name:         "Optimism",
				Symbol:       "ETH",
				RPCEndpoints: []string{"https://mainnet.optimism.io"},
				WSEndpoints:  []string{"wss://mainnet.optimism.io/ws"},
				ExplorerURL:  "https://optimistic.etherscan.io",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "ETH",
					Name:        "Ethereum",
					Decimals:    18,
					CoingeckoID: "ethereum",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0x7F5c764cBc14f9669B88837ca1490cCa17c31607",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    6,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.1),
				MaxGasPrice:        decimal.NewFromFloat(10),
				ConfirmationBlocks: 1,
				Enabled:            true,
				Priority:           5,
			},
			"avalanche": {
				ChainID:      43114,
				Name:         "Avalanche",
				Symbol:       "AVAX",
				RPCEndpoints: []string{"https://api.avax.network/ext/bc/C/rpc"},
				WSEndpoints:  []string{"wss://api.avax.network/ext/bc/C/ws"},
				ExplorerURL:  "https://snowtrace.io",
				NativeToken: TokenConfig{
					Address:     "0x0000000000000000000000000000000000000000",
					Symbol:      "AVAX",
					Name:        "Avalanche",
					Decimals:    18,
					CoingeckoID: "avalanche-2",
					IsNative:    true,
					Enabled:     true,
				},
				Tokens: []TokenConfig{
					{
						Address:     "0xA7D7079b0FEaD91F3e65f86E8915Cb59c1a4C664",
						Symbol:      "USDC",
						Name:        "USD Coin",
						Decimals:    6,
						CoingeckoID: "usd-coin",
						IsStable:    true,
						Enabled:     true,
					},
				},
				GasMultiplier:      decimal.NewFromFloat(1.1),
				MaxGasPrice:        decimal.NewFromFloat(100),
				ConfirmationBlocks: 1,
				Enabled:            true,
				Priority:           6,
			},
		},
		BridgeConfig: BridgeConfig{
			Enabled:          true,
			SupportedBridges: []string{"stargate", "hop", "across", "cbridge"},
			DefaultBridge:    "stargate",
			BridgeConfigs: map[string]BridgeProtocolConfig{
				"stargate": {
					Name:            "Stargate",
					Enabled:         true,
					APIEndpoint:     "https://api.stargate.finance",
					SupportedChains: []string{"ethereum", "bsc", "polygon", "arbitrum", "optimism", "avalanche"},
					SupportedTokens: []string{"USDC", "USDT", "ETH"},
					MinAmount:       decimal.NewFromFloat(1),
					MaxAmount:       decimal.NewFromFloat(1000000),
					Fee:             decimal.NewFromFloat(0.0006),
					EstimatedTime:   10 * time.Minute,
					Priority:        1,
				},
				"hop": {
					Name:            "Hop Protocol",
					Enabled:         true,
					APIEndpoint:     "https://api.hop.exchange",
					SupportedChains: []string{"ethereum", "polygon", "arbitrum", "optimism"},
					SupportedTokens: []string{"USDC", "USDT", "ETH", "MATIC"},
					MinAmount:       decimal.NewFromFloat(0.1),
					MaxAmount:       decimal.NewFromFloat(100000),
					Fee:             decimal.NewFromFloat(0.004),
					EstimatedTime:   15 * time.Minute,
					Priority:        2,
				},
			},
			MaxSlippage: decimal.NewFromFloat(0.03),
			Timeout:     5 * time.Minute,
		},
		GasConfig: GasConfig{
			Enabled:        true,
			UpdateInterval: 1 * time.Minute,
			ChainConfigs: map[string]GasChainConfig{
				"ethereum": {
					Enabled:         true,
					GasStation:      "https://ethgasstation.info/api/ethgasAPI.json",
					DefaultGasPrice: decimal.NewFromFloat(20),
					MaxGasPrice:     decimal.NewFromFloat(100),
					GasMultiplier:   decimal.NewFromFloat(1.1),
					Priority: map[string]decimal.Decimal{
						"slow":   decimal.NewFromFloat(1.0),
						"normal": decimal.NewFromFloat(1.2),
						"fast":   decimal.NewFromFloat(1.5),
					},
				},
			},
			AlertThresholds: map[string]decimal.Decimal{
				"ethereum": decimal.NewFromFloat(50),
				"bsc":      decimal.NewFromFloat(10),
				"polygon":  decimal.NewFromFloat(100),
			},
		},
		PriceOracleConfig: PriceOracleConfig{
			Enabled:         true,
			Provider:        "coingecko",
			UpdateInterval:  5 * time.Minute,
			CacheTimeout:    10 * time.Minute,
			SupportedTokens: []string{"ethereum", "bitcoin", "usd-coin", "tether", "binancecoin", "matic-network", "avalanche-2"},
			Endpoints: map[string]string{
				"coingecko": "https://api.coingecko.com/api/v3",
			},
			RateLimit: 50,
		},
		PortfolioConfig: PortfolioConfig{
			Enabled:             true,
			UpdateInterval:      1 * time.Hour,
			HistoryRetention:    30 * 24 * time.Hour, // 30 days
			RiskCalculation:     true,
			PerformanceTracking: true,
			AlertsEnabled:       true,
		},
	}
}

// ValidateMultichainConfig validates multichain configuration
func ValidateMultichainConfig(config MultichainConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if len(config.SupportedChains) == 0 {
		return fmt.Errorf("at least one supported chain must be specified")
	}

	if config.DefaultChain == "" {
		return fmt.Errorf("default chain is required")
	}

	if config.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if config.BalanceThreshold.LessThan(decimal.Zero) {
		return fmt.Errorf("balance threshold cannot be negative")
	}

	// Validate chain configs
	for chainName, chainConfig := range config.ChainConfigs {
		if err := validateChainConfig(chainName, chainConfig); err != nil {
			return fmt.Errorf("invalid config for chain %s: %w", chainName, err)
		}
	}

	// Validate bridge config
	if err := validateBridgeConfig(config.BridgeConfig); err != nil {
		return fmt.Errorf("invalid bridge config: %w", err)
	}

	return nil
}

// validateChainConfig validates individual chain configuration
func validateChainConfig(name string, config ChainConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.ChainID <= 0 {
		return fmt.Errorf("chain ID must be positive")
	}

	if config.Name == "" {
		return fmt.Errorf("chain name is required")
	}

	if len(config.RPCEndpoints) == 0 {
		return fmt.Errorf("at least one RPC endpoint is required")
	}

	if config.GasMultiplier.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("gas multiplier must be positive")
	}

	if config.MaxGasPrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max gas price must be positive")
	}

	if config.ConfirmationBlocks < 0 {
		return fmt.Errorf("confirmation blocks cannot be negative")
	}

	return nil
}

// validateBridgeConfig validates bridge configuration
func validateBridgeConfig(config BridgeConfig) error {
	if !config.Enabled {
		return nil
	}

	if len(config.SupportedBridges) == 0 {
		return fmt.Errorf("at least one supported bridge must be specified")
	}

	if config.DefaultBridge == "" {
		return fmt.Errorf("default bridge is required")
	}

	if config.MaxSlippage.LessThan(decimal.Zero) || config.MaxSlippage.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("max slippage must be between 0 and 1")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}
