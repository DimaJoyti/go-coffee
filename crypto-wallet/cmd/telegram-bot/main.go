package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/defi"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/telegram"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/wallet"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.New("telegram-bot")
	appLogger.Info("Starting Telegram Bot for Web3 Coffee Platform")

	// Initialize Redis client
	redisClient, err := redis.NewClientFromConfig(&cfg.Redis)
	if err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
	defer redisClient.Close()

	// Initialize AI service
	aiService, err := ai.NewService(cfg.AI, appLogger, redisClient)
	if err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to initialize AI service: %v", err))
	}
	defer aiService.Close()

	// Initialize wallet service
	walletService, err := createWalletService(cfg, appLogger, redisClient)
	if err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to initialize wallet service: %v", err))
	}

	// Initialize DeFi service
	defiService, err := createDeFiService(cfg, appLogger, redisClient)
	if err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to initialize DeFi service: %v", err))
	}

	// Initialize Telegram bot
	bot, err := telegram.NewBot(
		cfg.Telegram,
		appLogger,
		redisClient,
		aiService,
		walletService,
		defiService,
	)
	if err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to initialize Telegram bot: %v", err))
	}

	// Start health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		})
		log.Println("Health check server starting on :8087")
		if err := http.ListenAndServe(":8087", nil); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the bot
	if err := bot.Start(ctx); err != nil {
		appLogger.Fatal(fmt.Sprintf("Failed to start Telegram bot: %v", err))
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	appLogger.Info("Telegram bot is running. Press Ctrl+C to stop.")
	<-sigChan

	appLogger.Info("Shutting down Telegram bot...")
	cancel()

	// Stop the bot
	if err := bot.Stop(); err != nil {
		appLogger.Error(fmt.Sprintf("Error stopping bot: %v", err))
	}

	appLogger.Info("Telegram bot stopped successfully")
}

// createWalletService creates a wallet service
func createWalletService(cfg *config.Config, logger *logger.Logger, redisClient redis.Client) (*wallet.Service, error) {
	logger.Info("Creating wallet service")

	// For now, create a simplified wallet service
	// In production, this would initialize proper database connections and blockchain clients

	// Create a mock wallet service that can handle basic operations
	// This will be replaced with proper implementation once all dependencies are resolved
	return createMockWalletService(cfg, logger)
}

// createDeFiService creates a DeFi service
func createDeFiService(cfg *config.Config, logger *logger.Logger, redisClient redis.Client) (*defi.Service, error) {
	logger.Info("Creating DeFi service")

	// For now, create a simplified DeFi service
	// In production, this would initialize proper blockchain clients and dependencies
	return createMockDeFiService(cfg, logger)
}

// createMockWalletService creates a mock wallet service for testing
func createMockWalletService(cfg *config.Config, logger *logger.Logger) (*wallet.Service, error) {
	logger.Info("Creating mock wallet service")
	// Return nil for now - the bot will handle nil services gracefully
	return nil, nil
}

// createMockDeFiService creates a mock DeFi service for testing
func createMockDeFiService(cfg *config.Config, logger *logger.Logger) (*defi.Service, error) {
	logger.Info("Creating mock DeFi service")
	// Return nil for now - the bot will handle nil services gracefully
	return nil, nil
}

// Environment variables setup guide:
/*
Required environment variables:

1. TELEGRAM_BOT_TOKEN - Your Telegram bot token from @BotFather
   Example: export TELEGRAM_BOT_TOKEN="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"

2. GEMINI_API_KEY - Your Google Gemini API key
   Example: export GEMINI_API_KEY="AIzaSyC..."

3. TELEGRAM_WEBHOOK_URL (optional) - For webhook mode
   Example: export TELEGRAM_WEBHOOK_URL="https://yourdomain.com/webhook"

4. Redis configuration (if using external Redis):
   - REDIS_HOST
   - REDIS_PORT
   - REDIS_PASSWORD

Setup instructions:

1. Create a Telegram bot:
   - Message @BotFather on Telegram
   - Use /newbot command
   - Follow instructions to get your bot token
   - Set the token in TELEGRAM_BOT_TOKEN environment variable

2. Get Gemini API key:
   - Go to https://makersuite.google.com/app/apikey
   - Create a new API key
   - Set it in GEMINI_API_KEY environment variable

3. Install and run Ollama (optional, for local AI):
   - Download from https://ollama.ai/
   - Install and run: ollama serve
   - Pull a model: ollama pull llama3.1

4. Start Redis:
   - Using Docker: docker run -d -p 6379:6379 redis:alpine
   - Or install locally

5. Run the bot:
   - go run cmd/telegram-bot/main.go

Bot commands:
/start - Start the bot and setup wallet
/wallet - Manage your Web3 wallet
/balance - Check your crypto balance
/coffee - Order coffee with crypto payment
/menu - View coffee menu
/orders - View your orders
/settings - Bot settings
/help - Show help

Features:
- AI-powered coffee ordering with Gemini and Ollama
- Web3 wallet integration
- Crypto payments (BTC, ETH, USDC, USDT)
- Natural language processing for orders
- Personalized recommendations
- Multi-language support (Ukrainian/English)
- Secure wallet management
- Real-time balance checking
- Order tracking
*/
