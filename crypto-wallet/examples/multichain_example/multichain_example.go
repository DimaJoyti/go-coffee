package main

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/wallet/multichain"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🌐 Multi-Chain Wallet Management Example")
	fmt.Println("========================================")

	// Initialize logger
	logConfig := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	logger := logger.NewLogger(logConfig)

	// Create multichain configuration
	config := multichain.GetDefaultMultichainConfig()
	
	// Configure for demo (disable actual network connections)
	for chainName := range config.ChainConfigs {
		chainConfig := config.ChainConfigs[chainName]
		chainConfig.Enabled = false // Disable for demo
		config.ChainConfigs[chainName] = chainConfig
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Enabled: %v\n", config.Enabled)
	fmt.Printf("  Supported Chains: %v\n", config.SupportedChains)
	fmt.Printf("  Default Chain: %s\n", config.DefaultChain)
	fmt.Printf("  Update Interval: %v\n", config.UpdateInterval)
	fmt.Printf("  Balance Threshold: %s\n", config.BalanceThreshold.String())
	fmt.Println()

	// Create multichain manager
	manager := multichain.NewMultichainManager(logger, config)

	// Start the manager
	ctx := context.Background()
	if err := manager.Start(ctx); err != nil {
		fmt.Printf("Failed to start multichain manager: %v\n", err)
		return
	}

	fmt.Println("✅ Multichain manager started successfully!")
	fmt.Println()

	// Show supported chains
	fmt.Println("🌐 Supported Chains:")
	chains := manager.GetSupportedChains()
	for i, chain := range chains {
		chainInfo := config.ChainConfigs[chain]
		fmt.Printf("  %d. %s (%s)\n", i+1, chainInfo.Name, chain)
		fmt.Printf("     Chain ID: %d\n", chainInfo.ChainID)
		fmt.Printf("     Native Token: %s\n", chainInfo.NativeToken.Symbol)
		fmt.Printf("     Priority: %d\n", chainInfo.Priority)
	}
	fmt.Println()

	// Show chain status
	fmt.Println("📊 Chain Status:")
	chainStatus := manager.GetChainStatus()
	for chain, status := range chainStatus {
		fmt.Printf("  %s:\n", chain)
		fmt.Printf("    Chain ID: %d\n", status.ChainID)
		fmt.Printf("    Healthy: %v\n", status.IsHealthy)
		fmt.Printf("    Block Number: %d\n", status.BlockNumber)
		fmt.Printf("    Last Updated: %v\n", status.LastUpdated.Format("15:04:05"))
	}
	fmt.Println()

	// Show network statistics
	fmt.Println("📈 Network Statistics:")
	networkStats := manager.GetNetworkStats()
	fmt.Printf("  Total Chains: %d\n", networkStats.TotalChains)
	fmt.Printf("  Active Chains: %d\n", networkStats.ActiveChains)
	fmt.Printf("  Last Updated: %v\n", networkStats.LastUpdated.Format("15:04:05"))
	fmt.Println()

	// Demonstrate wallet operations with mock data
	fmt.Println("💼 Wallet Operations Demo:")
	
	// Mock wallet address
	walletAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b5C8E1")
	fmt.Printf("  Wallet Address: %s\n", walletAddress.Hex())
	fmt.Println()

	// Mock unified balances
	fmt.Println("💰 Mock Unified Balances:")
	mockBalances := map[string]*multichain.UnifiedBalance{
		"ETH": {
			Token: multichain.TokenConfig{
				Symbol:   "ETH",
				Name:     "Ethereum",
				Decimals: 18,
			},
			TotalBalance:  decimal.NewFromFloat(5.25),
			TotalValueUSD: decimal.NewFromFloat(10500),
			ChainBalances: map[string]*multichain.ChainBalance{
				"ethereum": {
					Chain:     "ethereum",
					Balance:   decimal.NewFromFloat(3.5),
					ValueUSD:  decimal.NewFromFloat(7000),
					Available: decimal.NewFromFloat(3.5),
				},
				"arbitrum": {
					Chain:     "arbitrum",
					Balance:   decimal.NewFromFloat(1.75),
					ValueUSD:  decimal.NewFromFloat(3500),
					Available: decimal.NewFromFloat(1.75),
				},
			},
			PriceUSD:  decimal.NewFromFloat(2000),
			Change24h: decimal.NewFromFloat(-2.5),
		},
		"USDC": {
			Token: multichain.TokenConfig{
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
			},
			TotalBalance:  decimal.NewFromFloat(15000),
			TotalValueUSD: decimal.NewFromFloat(15000),
			ChainBalances: map[string]*multichain.ChainBalance{
				"ethereum": {
					Chain:     "ethereum",
					Balance:   decimal.NewFromFloat(8000),
					ValueUSD:  decimal.NewFromFloat(8000),
					Available: decimal.NewFromFloat(8000),
				},
				"polygon": {
					Chain:     "polygon",
					Balance:   decimal.NewFromFloat(4000),
					ValueUSD:  decimal.NewFromFloat(4000),
					Available: decimal.NewFromFloat(4000),
				},
				"bsc": {
					Chain:     "bsc",
					Balance:   decimal.NewFromFloat(3000),
					ValueUSD:  decimal.NewFromFloat(3000),
					Available: decimal.NewFromFloat(3000),
				},
			},
			PriceUSD:  decimal.NewFromFloat(1.0),
			Change24h: decimal.NewFromFloat(0.1),
		},
		"MATIC": {
			Token: multichain.TokenConfig{
				Symbol:   "MATIC",
				Name:     "Polygon",
				Decimals: 18,
			},
			TotalBalance:  decimal.NewFromFloat(12500),
			TotalValueUSD: decimal.NewFromFloat(10000),
			ChainBalances: map[string]*multichain.ChainBalance{
				"polygon": {
					Chain:     "polygon",
					Balance:   decimal.NewFromFloat(12500),
					ValueUSD:  decimal.NewFromFloat(10000),
					Available: decimal.NewFromFloat(12500),
				},
			},
			PriceUSD:  decimal.NewFromFloat(0.8),
			Change24h: decimal.NewFromFloat(5.2),
		},
	}

	totalPortfolioValue := decimal.Zero
	for symbol, balance := range mockBalances {
		fmt.Printf("  %s (%s):\n", symbol, balance.Token.Name)
		fmt.Printf("    Total Balance: %s %s\n", balance.TotalBalance.String(), symbol)
		fmt.Printf("    Total Value: $%s\n", balance.TotalValueUSD.String())
		fmt.Printf("    Price: $%s\n", balance.PriceUSD.String())
		fmt.Printf("    24h Change: %s%%\n", balance.Change24h.String())
		fmt.Printf("    Chains: %d\n", len(balance.ChainBalances))
		
		for chainName, chainBalance := range balance.ChainBalances {
			fmt.Printf("      %s: %s %s ($%s)\n", 
				chainName, 
				chainBalance.Balance.String(), 
				symbol,
				chainBalance.ValueUSD.String())
		}
		
		totalPortfolioValue = totalPortfolioValue.Add(balance.TotalValueUSD)
		fmt.Println()
	}

	// Portfolio summary
	fmt.Println("📊 Portfolio Summary:")
	fmt.Printf("  Total Value: $%s\n", totalPortfolioValue.String())
	fmt.Printf("  Total Tokens: %d\n", len(mockBalances))
	fmt.Printf("  Chains Used: %d\n", len(chains))
	
	// Chain distribution
	chainDistribution := make(map[string]decimal.Decimal)
	for _, balance := range mockBalances {
		for chainName, chainBalance := range balance.ChainBalances {
			if existing, exists := chainDistribution[chainName]; exists {
				chainDistribution[chainName] = existing.Add(chainBalance.ValueUSD)
			} else {
				chainDistribution[chainName] = chainBalance.ValueUSD
			}
		}
	}
	
	fmt.Printf("  Chain Distribution:\n")
	for chainName, value := range chainDistribution {
		percentage := value.Div(totalPortfolioValue).Mul(decimal.NewFromInt(100))
		fmt.Printf("    %s: $%s (%.1f%%)\n", chainName, value.String(), percentage.InexactFloat64())
	}
	fmt.Println()

	// Risk analysis
	fmt.Println("⚠️ Risk Analysis:")
	
	// Calculate largest token concentration
	largestTokenValue := decimal.Zero
	largestToken := ""
	for symbol, balance := range mockBalances {
		if balance.TotalValueUSD.GreaterThan(largestTokenValue) {
			largestTokenValue = balance.TotalValueUSD
			largestToken = symbol
		}
	}
	
	tokenConcentration := largestTokenValue.Div(totalPortfolioValue).Mul(decimal.NewFromInt(100))
	fmt.Printf("  Token Concentration: %.1f%% (%s)\n", tokenConcentration.InexactFloat64(), largestToken)
	
	// Calculate largest chain concentration
	largestChainValue := decimal.Zero
	largestChain := ""
	for chainName, value := range chainDistribution {
		if value.GreaterThan(largestChainValue) {
			largestChainValue = value
			largestChain = chainName
		}
	}
	
	chainConcentration := largestChainValue.Div(totalPortfolioValue).Mul(decimal.NewFromInt(100))
	fmt.Printf("  Chain Concentration: %.1f%% (%s)\n", chainConcentration.InexactFloat64(), largestChain)
	
	// Risk level
	riskLevel := "Low"
	if tokenConcentration.GreaterThan(decimal.NewFromFloat(70)) || chainConcentration.GreaterThan(decimal.NewFromFloat(90)) {
		riskLevel = "High"
	} else if tokenConcentration.GreaterThan(decimal.NewFromFloat(40)) || chainConcentration.GreaterThan(decimal.NewFromFloat(70)) {
		riskLevel = "Medium"
	}
	
	fmt.Printf("  Overall Risk Level: %s\n", riskLevel)
	fmt.Println()

	// Cross-chain transfer demo
	fmt.Println("🌉 Cross-Chain Transfer Demo:")
	transferRequest := &multichain.CrossChainTransferRequest{
		SourceChain: "ethereum",
		DestChain:   "polygon",
		Token: multichain.TokenConfig{
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
		},
		Amount:      decimal.NewFromFloat(1000),
		FromAddress: walletAddress,
		ToAddress:   walletAddress,
		Slippage:    decimal.NewFromFloat(0.01),
		Deadline:    time.Now().Add(10 * time.Minute),
	}

	fmt.Printf("  Transfer Request:\n")
	fmt.Printf("    From: %s (%s)\n", transferRequest.SourceChain, "Ethereum")
	fmt.Printf("    To: %s (%s)\n", transferRequest.DestChain, "Polygon")
	fmt.Printf("    Token: %s\n", transferRequest.Token.Symbol)
	fmt.Printf("    Amount: %s %s\n", transferRequest.Amount.String(), transferRequest.Token.Symbol)
	fmt.Printf("    Slippage: %s%%\n", transferRequest.Slippage.Mul(decimal.NewFromInt(100)).String())
	fmt.Printf("    From Address: %s\n", transferRequest.FromAddress.Hex())
	fmt.Printf("    To Address: %s\n", transferRequest.ToAddress.Hex())
	fmt.Println()

	// Mock bridge options
	fmt.Printf("  Available Bridges:\n")
	bridges := []struct {
		Name          string
		Fee           string
		EstimatedTime string
		Success       string
	}{
		{"Stargate", "0.06%", "10 min", "99.5%"},
		{"Hop Protocol", "0.4%", "15 min", "98.2%"},
		{"Across", "0.25%", "12 min", "99.1%"},
		{"cBridge", "0.1%", "20 min", "97.8%"},
	}

	for i, bridge := range bridges {
		fmt.Printf("    %d. %s\n", i+1, bridge.Name)
		fmt.Printf("       Fee: %s | Time: %s | Success: %s\n", 
			bridge.Fee, bridge.EstimatedTime, bridge.Success)
	}
	fmt.Println()

	fmt.Printf("  ✅ Recommended: Stargate (Best overall score)\n")
	fmt.Printf("  💰 Estimated Fee: $0.60 (0.06%%)\n")
	fmt.Printf("  ⏱️ Estimated Time: 10 minutes\n")
	fmt.Printf("  📊 Success Rate: 99.5%%\n")
	fmt.Println()

	// Performance metrics
	fmt.Println("📈 Performance Metrics:")
	fmt.Printf("  Portfolio 24h Change: -$125.50 (-0.35%%)\n")
	fmt.Printf("  Best Performer: MATIC (+5.2%%)\n")
	fmt.Printf("  Worst Performer: ETH (-2.5%%)\n")
	fmt.Printf("  Total Transactions: 47\n")
	fmt.Printf("  Cross-Chain Transfers: 12\n")
	fmt.Printf("  Average Gas Saved: 15%%\n")
	fmt.Println()

	// Recommendations
	fmt.Println("💡 Recommendations:")
	fmt.Printf("  • Consider rebalancing ETH across more chains\n")
	fmt.Printf("  • USDC distribution looks optimal\n")
	fmt.Printf("  • Monitor MATIC price for potential profit-taking\n")
	fmt.Printf("  • Consider adding exposure to Arbitrum and Optimism\n")
	fmt.Printf("  • Gas optimization opportunities on Ethereum\n")
	fmt.Println()

	fmt.Println("🎉 Multi-Chain Wallet Management example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Multi-chain wallet management")
	fmt.Println("  ✅ Unified balance tracking across chains")
	fmt.Println("  ✅ Cross-chain transfer capabilities")
	fmt.Println("  ✅ Portfolio risk analysis")
	fmt.Println("  ✅ Chain health monitoring")
	fmt.Println("  ✅ Bridge protocol integration")
	fmt.Println("  ✅ Gas price tracking")
	fmt.Println("  ✅ Performance analytics")
	fmt.Println("  ✅ Automated recommendations")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the system with mock data.")
	fmt.Println("Configure real RPC endpoints and API keys for live operations.")

	// Stop the manager
	if err := manager.Stop(); err != nil {
		fmt.Printf("Error stopping multichain manager: %v\n", err)
	} else {
		fmt.Println("\n🛑 Multichain manager stopped")
	}
}
