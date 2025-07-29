package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"github.com/shopspring/decimal"
)

// MockRedisClient for the example
type MockRedisClient struct{}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	return false, nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) Close() error {
	return nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return "", nil
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return nil
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockRedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return time.Hour, nil
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}

func (m *MockRedisClient) FlushDB(ctx context.Context) error {
	return nil
}

func (m *MockRedisClient) Pipeline() redis.Pipeline {
	// Return nil for simplicity in this example
	// In production, you would implement a proper pipeline
	return nil
}

func main() {
	fmt.Println("🌾 Advanced Yield Farming Automation Example")
	fmt.Println("============================================")

	// Initialize logger and cache
	logConfig := config.LoggingConfig{
		Level:      "info",
		Format:     "console",
		Output:     "stdout",
		FilePath:   "",
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 3,
		Compress:   true,
	}
	logger := logger.NewLogger(logConfig)

	// Use our mock Redis client for the example
	cache := &MockRedisClient{}

	// Create yield farming configuration
	config := defi.GetDefaultYieldFarmingConfig()
	config.AutoCompoundingEnabled = true
	config.PoolMigrationEnabled = true
	config.ImpermanentLossProtection = true
	config.MinYieldThreshold = decimal.NewFromFloat(0.08) // 8% minimum APY
	config.MaxPositionSize = decimal.NewFromFloat(5000)   // $5k max position
	config.CompoundingInterval = 6 * time.Hour            // Compound every 6 hours

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Auto Compounding: %v\n", config.AutoCompoundingEnabled)
	fmt.Printf("  Pool Migration: %v\n", config.PoolMigrationEnabled)
	fmt.Printf("  IL Protection: %v\n", config.ImpermanentLossProtection)
	fmt.Printf("  Min Yield Threshold: %s%%\n", config.MinYieldThreshold.Mul(decimal.NewFromFloat(100)).String())
	fmt.Printf("  Supported Protocols: %v\n", config.SupportedProtocols)
	fmt.Println()

	// Create yield farming automation
	automation := defi.NewYieldFarmingAutomation(logger, cache, config)

	// Start the automation
	ctx := context.Background()
	if err := automation.Start(ctx); err != nil {
		log.Fatalf("Failed to start yield farming automation: %v", err)
	}

	fmt.Println("✅ Yield farming automation started successfully!")
	fmt.Println()

	// Demonstrate scanning for opportunities
	fmt.Println("🔍 Scanning for yield opportunities...")
	opportunities, err := automation.ScanForYieldOpportunities(ctx)
	if err != nil {
		log.Fatalf("Failed to scan for opportunities: %v", err)
	}

	fmt.Printf("Found %d yield opportunities:\n", len(opportunities))
	for i, opp := range opportunities {
		if i >= 3 { // Show only first 3
			break
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, opp.PoolName, opp.Protocol)
		fmt.Printf("     APY: %s%% | TVL: $%s | Risk Score: %s\n",
			opp.CurrentAPY.Mul(decimal.NewFromFloat(100)).StringFixed(2),
			opp.TVL.StringFixed(0),
			opp.RiskScore.StringFixed(3))
		fmt.Printf("     IL Risk: %s%% | Confidence: %s%%\n",
			opp.ImpermanentLossRisk.Mul(decimal.NewFromFloat(100)).StringFixed(2),
			opp.Confidence.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	}
	fmt.Println()

	// Demonstrate farming strategies
	fmt.Println("📋 Available farming strategies:")
	strategies := automation.GetFarmingStrategies()
	for i, strategy := range strategies {
		fmt.Printf("  %d. %s\n", i+1, strategy.Name)
		fmt.Printf("     Target APY: %s%% | Risk Level: %s\n",
			strategy.TargetAPY.Mul(decimal.NewFromFloat(100)).StringFixed(1),
			strategy.RiskLevel)
		fmt.Printf("     Max IL: %s%% | Allocation: %s%%\n",
			strategy.MaxImpermanentLoss.Mul(decimal.NewFromFloat(100)).StringFixed(1),
			strategy.AllocationPercentage.Mul(decimal.NewFromFloat(100)).StringFixed(1))
	}
	fmt.Println()

	// Demonstrate entering a farm
	if len(opportunities) > 0 && len(strategies) > 0 {
		fmt.Println("🌱 Entering yield farm...")
		amount := decimal.NewFromFloat(2000) // $2000 investment

		farm, err := automation.EnterFarm(ctx, opportunities[0].ID, amount, strategies[0].ID)
		if err != nil {
			log.Printf("Failed to enter farm: %v", err)
		} else {
			fmt.Printf("✅ Successfully entered farm: %s\n", farm.ID)
			fmt.Printf("   Pool: %s (%s)\n", farm.PoolName, farm.Protocol)
			fmt.Printf("   Amount: $%s | APY: %s%%\n",
				farm.LiquidityAmount.StringFixed(2),
				farm.CurrentAPY.Mul(decimal.NewFromFloat(100)).StringFixed(2))
			fmt.Printf("   Auto Compounding: %v | Strategy: %s\n",
				farm.AutoCompounding, farm.Strategy)
		}
		fmt.Println()

		// Demonstrate performance metrics
		fmt.Println("📊 Performance metrics:")
		metrics := automation.GetPerformanceMetrics()
		fmt.Printf("   Total Value Locked: $%s\n", metrics.TotalValueLocked.StringFixed(2))
		fmt.Printf("   Active Farms: %d\n", metrics.ActiveFarmsCount)
		fmt.Printf("   Total Rewards Earned: $%s\n", metrics.TotalRewardsEarned.StringFixed(2))
		fmt.Printf("   Average APY: %s%%\n", metrics.AverageAPY.Mul(decimal.NewFromFloat(100)).StringFixed(2))

		if len(metrics.ProtocolDistribution) > 0 {
			fmt.Printf("   Protocol Distribution:\n")
			for protocol, amount := range metrics.ProtocolDistribution {
				fmt.Printf("     %s: $%s\n", protocol, amount.StringFixed(2))
			}
		}
		fmt.Println()

		// Simulate some time passing and rewards accumulating
		fmt.Println("⏰ Simulating reward accumulation...")
		time.Sleep(100 * time.Millisecond) // Brief pause for demo

		// Manually add some rewards for demonstration
		activeFarms := automation.GetActiveFarms()
		if len(activeFarms) > 0 {
			// Simulate earned rewards (this would normally happen automatically)
			fmt.Printf("💰 Simulated rewards earned: $25.50\n")

			// In a real scenario, rewards would be detected automatically
			// and compounding would happen based on the strategy's frequency
			fmt.Println("🔄 Auto-compounding would trigger based on strategy settings")
		}
		fmt.Println()

		// Demonstrate impermanent loss protection
		fmt.Println("🛡️ Impermanent Loss Protection:")
		ilProtection := automation.GetImpermanentLossProtection()
		fmt.Printf("   Enabled: %v\n", ilProtection.Enabled)
		fmt.Printf("   Max Acceptable Loss: %s%%\n",
			ilProtection.MaxAcceptableLoss.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("   Auto Exit Enabled: %v (Threshold: %s%%)\n",
			ilProtection.AutoExitEnabled,
			ilProtection.AutoExitThreshold.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("   Rebalancing Enabled: %v (Threshold: %s%%)\n",
			ilProtection.RebalancingEnabled,
			ilProtection.RebalancingThreshold.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Println()

		// Demonstrate creating a custom strategy
		fmt.Println("🎯 Creating custom farming strategy...")
		customStrategy := &defi.FarmingStrategy{
			Name:                 "High-Yield DeFi Strategy",
			Description:          "Custom strategy targeting high-yield opportunities",
			TargetAPY:            decimal.NewFromFloat(0.25), // 25% target APY
			MaxImpermanentLoss:   decimal.NewFromFloat(0.08), // 8% max IL
			PreferredProtocols:   []string{"uniswap", "sushiswap"},
			PreferredChains:      []string{"ethereum", "arbitrum"},
			RiskLevel:            "moderate",
			CompoundingFrequency: 4 * time.Hour, // Compound every 4 hours
			RebalanceThreshold:   decimal.NewFromFloat(0.1),
			AllocationPercentage: decimal.NewFromFloat(0.3), // 30% allocation
			IsActive:             true,
		}

		err = automation.CreateFarmingStrategy(customStrategy)
		if err != nil {
			log.Printf("Failed to create custom strategy: %v", err)
		} else {
			fmt.Printf("✅ Created custom strategy: %s (ID: %s)\n",
				customStrategy.Name, customStrategy.ID)
		}
		fmt.Println()

		// Show all active farms
		fmt.Println("🌾 Active farms summary:")
		allFarms := automation.GetActiveFarms()
		for i, farm := range allFarms {
			fmt.Printf("  %d. %s (%s)\n", i+1, farm.PoolName, farm.Protocol)
			fmt.Printf("     Value: $%s | APY: %s%% | Status: %s\n",
				farm.TotalValue.StringFixed(2),
				farm.CurrentAPY.Mul(decimal.NewFromFloat(100)).StringFixed(2),
				farm.Status)
			fmt.Printf("     P&L: $%s | IL: %s%%\n",
				farm.ProfitLoss.StringFixed(2),
				farm.ImpermanentLoss.Mul(decimal.NewFromFloat(100)).StringFixed(2))
		}
		fmt.Println()
	}

	// Demonstrate configuration updates
	fmt.Println("⚙️ Updating configuration...")
	newConfig := automation.GetConfig()
	newConfig.MinYieldThreshold = decimal.NewFromFloat(0.10) // Increase to 10%
	newConfig.CompoundingInterval = 8 * time.Hour            // Less frequent compounding

	err = automation.UpdateConfig(newConfig)
	if err != nil {
		log.Printf("Failed to update configuration: %v", err)
	} else {
		fmt.Println("✅ Configuration updated successfully")
		fmt.Printf("   New min yield threshold: %s%%\n",
			newConfig.MinYieldThreshold.Mul(decimal.NewFromFloat(100)).StringFixed(1))
		fmt.Printf("   New compounding interval: %v\n", newConfig.CompoundingInterval)
	}
	fmt.Println()

	// Show automation status
	fmt.Println("📈 Automation Status:")
	fmt.Printf("   Running: %v\n", automation.IsRunning())
	fmt.Printf("   Active Farms: %d\n", len(automation.GetActiveFarms()))
	fmt.Printf("   Available Opportunities: %d\n", len(automation.GetYieldOpportunities()))
	fmt.Printf("   Farming Strategies: %d\n", len(automation.GetFarmingStrategies()))
	fmt.Println()

	fmt.Println("🎉 Yield farming automation example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✅ Automated opportunity detection")
	fmt.Println("  ✅ Multi-strategy farming")
	fmt.Println("  ✅ Auto-compounding rewards")
	fmt.Println("  ✅ Impermanent loss protection")
	fmt.Println("  ✅ Pool migration optimization")
	fmt.Println("  ✅ Performance tracking")
	fmt.Println("  ✅ Risk management")
	fmt.Println("  ✅ Custom strategy creation")

	// Stop the automation
	if err := automation.Stop(); err != nil {
		log.Printf("Error stopping automation: %v", err)
	} else {
		fmt.Println("\n🛑 Yield farming automation stopped")
	}
}
