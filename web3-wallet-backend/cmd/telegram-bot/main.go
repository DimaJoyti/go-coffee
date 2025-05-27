package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	logger := logger.New("telegram-bot")
	logger.Info("Starting Telegram Bot for Web3 Coffee Platform")

	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
	defer redisClient.Close()

	// Initialize AI service
	aiService, err := ai.NewService(cfg.AI, logger, redisClient)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize AI service: %v", err))
	}
	defer aiService.Close()

	// Initialize wallet service (mock for now)
	walletService := createMockWalletService(logger)

	// Initialize DeFi service (mock for now)
	defiService := createMockDeFiService(logger)

	// Initialize Telegram bot
	bot, err := telegram.NewBot(
		cfg.Telegram,
		logger,
		redisClient,
		aiService,
		walletService,
		defiService,
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize Telegram bot: %v", err))
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
		logger.Fatal(fmt.Sprintf("Failed to start Telegram bot: %v", err))
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Telegram bot is running. Press Ctrl+C to stop.")
	<-sigChan

	logger.Info("Shutting down Telegram bot...")
	cancel()

	// Stop the bot
	if err := bot.Stop(); err != nil {
		logger.Error(fmt.Sprintf("Error stopping bot: %v", err))
	}

	logger.Info("Telegram bot stopped successfully")
}

// createMockWalletService creates a mock wallet service for testing
func createMockWalletService(logger *logger.Logger) *wallet.Service {
	// In a real implementation, you would initialize the actual wallet service
	// with proper repository, blockchain clients, etc.
	logger.Info("Creating mock wallet service")
	return nil // Return nil for now, will be properly implemented later
}

// createMockDeFiService creates a mock DeFi service for testing
func createMockDeFiService(logger *logger.Logger) *defi.Service {
	// In a real implementation, you would initialize the actual DeFi service
	// with proper configuration and dependencies
	logger.Info("Creating mock DeFi service")
	return nil // Return nil for now, will be properly implemented later
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
