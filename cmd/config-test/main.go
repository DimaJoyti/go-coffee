package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DimaJoyti/go-coffee/pkg/config"
)

func main() {
	fmt.Println("🔧 Go Coffee Configuration Test Utility")
	fmt.Println("========================================")
	fmt.Println()

	// Автоматично завантажуємо .env файли
	fmt.Println("📁 Loading environment files...")
	if err := config.AutoLoadEnvFiles(); err != nil {
		log.Printf("Warning: %v", err)
	}
	fmt.Println()

	// Завантажуємо конфігурацію з змінних середовища
	fmt.Println("⚙️  Loading configuration from environment variables...")
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println("✅ Configuration loaded successfully!")
	fmt.Println()

	// Валідуємо конфігурацію
	fmt.Println("🔍 Validating configuration...")
	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	fmt.Println("✅ Configuration is valid!")
	fmt.Println()

	// Виводимо конфігурацію
	fmt.Println("📋 Current configuration:")
	fmt.Println()
	config.PrintConfig(cfg)
	fmt.Println()

	// Перевіряємо критичні сервіси
	fmt.Println("🔧 Service Status Check:")
	fmt.Println("========================")
	
	// Перевіряємо feature flags
	if cfg.Features.APIGatewayEnabled {
		fmt.Printf("✅ API Gateway: ENABLED (Port: %d)\n", cfg.Server.APIGatewayPort)
	} else {
		fmt.Println("❌ API Gateway: DISABLED")
	}

	if cfg.Features.ProducerServiceEnabled {
		fmt.Printf("✅ Producer Service: ENABLED (Port: %d)\n", cfg.Server.ProducerPort)
	} else {
		fmt.Println("❌ Producer Service: DISABLED")
	}

	if cfg.Features.ConsumerServiceEnabled {
		fmt.Printf("✅ Consumer Service: ENABLED (Port: %d)\n", cfg.Server.ConsumerPort)
	} else {
		fmt.Println("❌ Consumer Service: DISABLED")
	}

	if cfg.Features.StreamsServiceEnabled {
		fmt.Printf("✅ Streams Service: ENABLED (Port: %d)\n", cfg.Server.StreamsPort)
	} else {
		fmt.Println("❌ Streams Service: DISABLED")
	}

	if cfg.Features.AISearchEnabled {
		fmt.Printf("✅ AI Search Engine: ENABLED (Port: %d)\n", cfg.Server.AISearchPort)
	} else {
		fmt.Println("❌ AI Search Engine: DISABLED")
	}

	if cfg.Features.AuthModuleEnabled {
		fmt.Printf("✅ Auth Service: ENABLED (Port: %d)\n", cfg.Server.AuthServicePort)
	} else {
		fmt.Println("❌ Auth Service: DISABLED")
	}

	if cfg.Features.Web3WalletEnabled {
		fmt.Println("✅ Web3 Wallet: ENABLED")
	} else {
		fmt.Println("❌ Web3 Wallet: DISABLED")
	}

	if cfg.Features.DeFiServiceEnabled {
		fmt.Println("✅ DeFi Service: ENABLED")
	} else {
		fmt.Println("❌ DeFi Service: DISABLED")
	}

	if cfg.Features.AIAgentsEnabled {
		fmt.Println("✅ AI Agents: ENABLED")
	} else {
		fmt.Println("❌ AI Agents: DISABLED")
	}

	fmt.Println()

	// Перевіряємо підключення до зовнішніх сервісів
	fmt.Println("🌐 External Services Configuration:")
	fmt.Println("===================================")
	
	// Database
	fmt.Printf("🗄️  Database: %s@%s:%d/%s\n", 
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	
	// Redis
	fmt.Printf("🔴 Redis: %s:%d (DB: %d)\n", 
		cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
	
	// Kafka
	fmt.Printf("📨 Kafka: %v\n", cfg.Kafka.Brokers)
	fmt.Printf("   Topic: %s\n", cfg.Kafka.Topic)
	fmt.Printf("   Consumer Group: %s\n", cfg.Kafka.ConsumerGroup)
	
	fmt.Println()

	// Перевіряємо AI конфігурацію
	fmt.Println("🤖 AI Configuration:")
	fmt.Println("====================")
	
	if cfg.AI.OpenAIAPIKey != "" && cfg.AI.OpenAIAPIKey != "your-openai-api-key" {
		fmt.Println("✅ OpenAI API Key: CONFIGURED")
	} else {
		fmt.Println("❌ OpenAI API Key: NOT CONFIGURED")
	}

	if cfg.AI.GeminiAPIKey != "" && cfg.AI.GeminiAPIKey != "your-gemini-api-key" {
		fmt.Println("✅ Gemini API Key: CONFIGURED")
	} else {
		fmt.Println("❌ Gemini API Key: NOT CONFIGURED")
	}

	fmt.Printf("🔍 Search Model: %s\n", cfg.AI.SearchEmbeddingModel)
	fmt.Printf("📐 Vector Dimensions: %d\n", cfg.AI.SearchVectorDimensions)
	fmt.Printf("🎯 Similarity Threshold: %.2f\n", cfg.AI.SearchSimilarityThreshold)
	
	fmt.Println()

	// Перевіряємо Web3 конфігурацію
	fmt.Println("🌐 Web3 Configuration:")
	fmt.Println("======================")
	
	if cfg.Web3.Ethereum.RPCURL != "" && cfg.Web3.Ethereum.RPCURL != "https://mainnet.infura.io/v3/your-project-id" {
		fmt.Println("✅ Ethereum RPC: CONFIGURED")
	} else {
		fmt.Println("❌ Ethereum RPC: NOT CONFIGURED")
	}

	if cfg.Web3.Bitcoin.RPCURL != "" && cfg.Web3.Bitcoin.RPCURL != "https://your-bitcoin-node.com" {
		fmt.Println("✅ Bitcoin RPC: CONFIGURED")
	} else {
		fmt.Println("❌ Bitcoin RPC: NOT CONFIGURED")
	}

	if cfg.Web3.Solana.RPCURL != "" && cfg.Web3.Solana.RPCURL != "https://api.mainnet-beta.solana.com" {
		fmt.Println("✅ Solana RPC: CONFIGURED")
	} else {
		fmt.Println("❌ Solana RPC: NOT CONFIGURED")
	}

	fmt.Println()

	// Перевіряємо безпеку
	fmt.Println("🔒 Security Configuration:")
	fmt.Println("==========================")
	
	if cfg.Security.JWTSecret != "" && cfg.Security.JWTSecret != "your-super-secret-jwt-key" {
		fmt.Println("✅ JWT Secret: CONFIGURED")
	} else {
		fmt.Println("❌ JWT Secret: NOT CONFIGURED (SECURITY RISK!)")
	}

	if cfg.Security.EncryptionKey != "" && cfg.Security.EncryptionKey != "your-32-character-encryption-key!!" {
		fmt.Println("✅ Encryption Key: CONFIGURED")
	} else {
		fmt.Println("❌ Encryption Key: NOT CONFIGURED (SECURITY RISK!)")
	}

	fmt.Printf("⏰ JWT Expiry: %s\n", cfg.Security.JWTExpiry)
	fmt.Printf("🔄 Refresh Token Expiry: %s\n", cfg.Security.RefreshTokenExpiry)
	
	fmt.Println()

	// Перевіряємо моніторинг
	fmt.Println("📊 Monitoring Configuration:")
	fmt.Println("============================")
	
	if cfg.Monitoring.Prometheus.Enabled {
		fmt.Printf("✅ Prometheus: ENABLED (Port: %d)\n", cfg.Monitoring.Prometheus.Port)
	} else {
		fmt.Println("❌ Prometheus: DISABLED")
	}

	if cfg.Monitoring.Jaeger.Enabled {
		fmt.Printf("✅ Jaeger Tracing: ENABLED (%s)\n", cfg.Monitoring.Jaeger.Endpoint)
	} else {
		fmt.Println("❌ Jaeger Tracing: DISABLED")
	}

	if cfg.Monitoring.Sentry.DSN != "" && cfg.Monitoring.Sentry.DSN != "your-sentry-dsn" {
		fmt.Println("✅ Sentry Error Tracking: CONFIGURED")
	} else {
		fmt.Println("❌ Sentry Error Tracking: NOT CONFIGURED")
	}

	fmt.Println()

	// Підсумок
	fmt.Println("📋 Configuration Test Summary:")
	fmt.Println("==============================")
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Debug Mode: %t\n", cfg.Debug)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	
	// Рахуємо активні сервіси
	activeServices := 0
	totalServices := 9
	
	if cfg.Features.APIGatewayEnabled { activeServices++ }
	if cfg.Features.ProducerServiceEnabled { activeServices++ }
	if cfg.Features.ConsumerServiceEnabled { activeServices++ }
	if cfg.Features.StreamsServiceEnabled { activeServices++ }
	if cfg.Features.AISearchEnabled { activeServices++ }
	if cfg.Features.AuthModuleEnabled { activeServices++ }
	if cfg.Features.Web3WalletEnabled { activeServices++ }
	if cfg.Features.DeFiServiceEnabled { activeServices++ }
	if cfg.Features.AIAgentsEnabled { activeServices++ }
	
	fmt.Printf("Active Services: %d/%d\n", activeServices, totalServices)
	
	if activeServices == totalServices {
		fmt.Println("🎉 All services are enabled and ready to go!")
	} else {
		fmt.Printf("⚠️  %d services are disabled\n", totalServices-activeServices)
	}

	fmt.Println()
	fmt.Println("✅ Configuration test completed successfully!")
	
	// Якщо є аргументи командного рядка, виконуємо додаткові дії
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "validate":
			fmt.Println("\n🔍 Running extended validation...")
			// Тут можна додати додаткові перевірки
			
		case "export":
			fmt.Println("\n📤 Exporting configuration to JSON...")
			// Тут можна експортувати конфігурацію в JSON файл
			
		case "help":
			fmt.Println("\n📖 Usage:")
			fmt.Println("  go run cmd/config-test/main.go [command]")
			fmt.Println("")
			fmt.Println("Commands:")
			fmt.Println("  validate  - Run extended validation")
			fmt.Println("  export    - Export configuration to JSON")
			fmt.Println("  help      - Show this help message")
		}
	}
}
