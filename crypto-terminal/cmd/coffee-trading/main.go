package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/api"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/trading"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("☕ Starting Coffee Trading Platform...")

	// Initialize strategy engine
	engineConfig := &trading.EngineConfig{
		MaxConcurrentStrategies: 10,
		SignalBufferSize:        100,
		ExecutionBufferSize:     100,
		TickInterval:            time.Second * 5,
		MaxPortfolioRisk:        decimal.NewFromFloat(0.02), // 2%
		EmergencyStopEnabled:    true,
		CoffeeCorrelationWeight: decimal.NewFromFloat(0.1),  // 10%
	}

	strategyEngine := trading.NewStrategyEngine(engineConfig, logger)

	// Initialize API server
	serverConfig := &api.ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		EnableCORS:   true,
		EnableTLS:    false,
	}

	server := api.NewServer(strategyEngine, serverConfig, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start strategy engine
	if err := strategyEngine.Start(ctx); err != nil {
		logger.Fatalf("Failed to start strategy engine: %v", err)
	}

	// Start API server
	if err := server.Start(ctx); err != nil {
		logger.Fatalf("Failed to start API server: %v", err)
	}

	// Create some example strategies
	if err := createExampleStrategies(strategyEngine, logger); err != nil {
		logger.Errorf("Failed to create example strategies: %v", err)
	}

	// Start example WebSocket client (for demonstration)
	go startExampleWebSocketClient(logger)

	// Print startup information
	printStartupInfo(logger)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("☕ Shutting down Coffee Trading Platform...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop strategy engine
	if err := strategyEngine.Stop(); err != nil {
		logger.Errorf("Error stopping strategy engine: %v", err)
	}

	// Stop API server
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Errorf("Error stopping API server: %v", err)
	}

	logger.Info("☕ Coffee Trading Platform stopped gracefully. Thanks for trading with us!")
}

// createExampleStrategies creates some example coffee strategies
func createExampleStrategies(engine *trading.StrategyEngine, logger *logrus.Logger) error {
	factory := trading.NewCoffeeStrategyFactory()

	// Create Espresso strategy for BTC
	espressoStrategy := factory.CreateEspressoStrategy("BTCUSDT")
	espressoStrategy.Name = "Bitcoin Espresso Scalper"
	if err := engine.AddStrategy(espressoStrategy); err != nil {
		return fmt.Errorf("failed to add espresso strategy: %w", err)
	}
	logger.Info("☕ Created Bitcoin Espresso strategy")

	// Create Latte strategy for ETH
	latteStrategy := factory.CreateLatteStrategy("ETHUSDT")
	latteStrategy.Name = "Ethereum Latte Swing"
	if err := engine.AddStrategy(latteStrategy); err != nil {
		return fmt.Errorf("failed to add latte strategy: %w", err)
	}
	logger.Info("🥛 Created Ethereum Latte strategy")

	// Create Cold Brew strategy for ADA
	coldBrewStrategy := factory.CreateColdBrewStrategy("ADAUSDT")
	coldBrewStrategy.Name = "Cardano Cold Brew Position"
	if err := engine.AddStrategy(coldBrewStrategy); err != nil {
		return fmt.Errorf("failed to add cold brew strategy: %w", err)
	}
	logger.Info("🧊 Created Cardano Cold Brew strategy")

	// Create Cappuccino strategy for SOL
	cappuccinoStrategy := factory.CreateCappuccinoStrategy("SOLUSDT")
	cappuccinoStrategy.Name = "Solana Cappuccino Momentum"
	if err := engine.AddStrategy(cappuccinoStrategy); err != nil {
		return fmt.Errorf("failed to add cappuccino strategy: %w", err)
	}
	logger.Info("🫖 Created Solana Cappuccino strategy")

	return nil
}

// startExampleWebSocketClient starts an example WebSocket client for demonstration
func startExampleWebSocketClient(logger *logrus.Logger) {
	// This would typically be in a separate client application
	// For now, we'll just log that it would connect
	time.Sleep(2 * time.Second) // Wait for server to start
	logger.Info("📡 Example WebSocket client would connect to ws://localhost:8080/ws/coffee-trading")
	logger.Info("📡 Client would subscribe to channels: price_updates, signal_alerts, trade_executions")
}

// printStartupInfo prints helpful startup information
func printStartupInfo(logger *logrus.Logger) {
	logger.Info("☕ ================================")
	logger.Info("☕ Coffee Trading Platform Started!")
	logger.Info("☕ ================================")
	logger.Info("")
	logger.Info("🌐 API Server: http://localhost:8080")
	logger.Info("📡 WebSocket: ws://localhost:8080/ws/coffee-trading")
	logger.Info("📚 Documentation: http://localhost:8080/docs")
	logger.Info("❤️ Health Check: http://localhost:8080/health")
	logger.Info("")
	logger.Info("☕ Coffee Strategy Menu:")
	logger.Info("   • Espresso (☕): High-frequency scalping")
	logger.Info("   • Latte (🥛): Smooth swing trading")
	logger.Info("   • Cold Brew (🧊): Patient position trading")
	logger.Info("   • Cappuccino (🫖): Frothy momentum trading")
	logger.Info("")
	logger.Info("🚀 Quick Start Commands:")
	logger.Info("   curl http://localhost:8080/api/v1/coffee-trading/coffee/menu")
	logger.Info("   curl http://localhost:8080/api/v1/coffee-trading/strategies")
	logger.Info("   curl http://localhost:8080/api/v1/coffee-trading/analytics/dashboard")
	logger.Info("")
	logger.Info("📡 WebSocket Example:")
	logger.Info(`   {
     "type": "subscribe",
     "data": {
       "channels": ["price_updates", "signal_alerts", "trade_executions"]
     }
   }`)
	logger.Info("")
	logger.Info("☕ Happy Trading! May your profits be as rich as your coffee!")
	logger.Info("☕ ================================")
}
