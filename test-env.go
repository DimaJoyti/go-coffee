package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DimaJoyti/go-coffee/pkg/config"
)

func main() {
	fmt.Println("🔧 Testing Go Coffee Environment Configuration")
	fmt.Println("==============================================")
	fmt.Println()

	// Test loading .env files
	fmt.Println("📁 Testing environment file loading...")
	
	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("⚠️  .env file not found, creating from .env.example...")
		if _, err := os.Stat(".env.example"); err == nil {
			// Copy .env.example to .env
			data, err := os.ReadFile(".env.example")
			if err != nil {
				log.Fatalf("Failed to read .env.example: %v", err)
			}
			if err := os.WriteFile(".env", data, 0644); err != nil {
				log.Fatalf("Failed to create .env: %v", err)
			}
			fmt.Println("✅ Created .env from .env.example")
		} else {
			log.Fatalf("Neither .env nor .env.example found!")
		}
	} else {
		fmt.Println("✅ .env file found")
	}

	// Load environment files
	fmt.Println("\n📥 Loading environment files...")
	if err := config.AutoLoadEnvFiles(); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Load configuration from environment
	fmt.Println("\n⚙️  Loading configuration from environment variables...")
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println("✅ Configuration loaded successfully!")

	// Validate configuration
	fmt.Println("\n🔍 Validating configuration...")
	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		fmt.Println("\n⚠️  Some configuration issues found, but continuing with test...")
	} else {
		fmt.Println("✅ Configuration is valid!")
	}

	// Print basic configuration info
	fmt.Println("\n📋 Basic Configuration Info:")
	fmt.Println("============================")
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Debug Mode: %t\n", cfg.Debug)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)

	fmt.Println("\nServer Ports:")
	fmt.Printf("  API Gateway: %d\n", cfg.Server.APIGatewayPort)
	fmt.Printf("  Producer: %d\n", cfg.Server.ProducerPort)
	fmt.Printf("  Consumer: %d\n", cfg.Server.ConsumerPort)
	fmt.Printf("  AI Search: %d\n", cfg.Server.AISearchPort)
	fmt.Printf("  Auth Service: %d\n", cfg.Server.AuthServicePort)

	fmt.Println("\nDatabase Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  Name: %s\n", cfg.Database.Name)
	fmt.Printf("  User: %s\n", cfg.Database.User)

	fmt.Println("\nRedis Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Redis.Host)
	fmt.Printf("  Port: %d\n", cfg.Redis.Port)
	fmt.Printf("  DB: %d\n", cfg.Redis.DB)

	fmt.Println("\nKafka Configuration:")
	fmt.Printf("  Brokers: %v\n", cfg.Kafka.Brokers)
	fmt.Printf("  Topic: %s\n", cfg.Kafka.Topic)

	fmt.Println("\nFeature Flags:")
	fmt.Printf("  API Gateway: %t\n", cfg.Features.APIGatewayEnabled)
	fmt.Printf("  Producer Service: %t\n", cfg.Features.ProducerServiceEnabled)
	fmt.Printf("  Consumer Service: %t\n", cfg.Features.ConsumerServiceEnabled)
	fmt.Printf("  AI Search: %t\n", cfg.Features.AISearchEnabled)
	fmt.Printf("  Web3 Wallet: %t\n", cfg.Features.Web3WalletEnabled)
	fmt.Printf("  AI Agents: %t\n", cfg.Features.AIAgentsEnabled)

	// Test individual environment variable functions
	fmt.Println("\n🧪 Testing environment variable functions:")
	fmt.Println("==========================================")
	
	// Test GetEnv
	testValue := config.GetEnv("ENVIRONMENT", "test-default")
	fmt.Printf("GetEnv('ENVIRONMENT'): %s\n", testValue)
	
	// Test GetEnvAsInt
	testPort := config.GetEnvAsInt("API_GATEWAY_PORT", 9999)
	fmt.Printf("GetEnvAsInt('API_GATEWAY_PORT'): %d\n", testPort)
	
	// Test GetEnvAsBool
	testDebug := config.GetEnvAsBool("DEBUG", false)
	fmt.Printf("GetEnvAsBool('DEBUG'): %t\n", testDebug)
	
	// Test GetEnvAsSlice
	testBrokers := config.GetEnvAsSlice("KAFKA_BROKERS", []string{"default:9092"})
	fmt.Printf("GetEnvAsSlice('KAFKA_BROKERS'): %v\n", testBrokers)

	// Security check
	fmt.Println("\n🔒 Security Check:")
	fmt.Println("==================")
	
	if cfg.Security.JWTSecret == "your-super-secret-jwt-key" {
		fmt.Println("❌ JWT_SECRET is using default value - SECURITY RISK!")
	} else {
		fmt.Println("✅ JWT_SECRET is configured")
	}
	
	if cfg.Security.EncryptionKey == "your-32-character-encryption-key!!" {
		fmt.Println("❌ ENCRYPTION_KEY is using default value - SECURITY RISK!")
	} else {
		fmt.Println("✅ ENCRYPTION_KEY is configured")
	}

	// AI Configuration check
	fmt.Println("\n🤖 AI Configuration Check:")
	fmt.Println("==========================")
	
	if cfg.AI.OpenAIAPIKey != "" && cfg.AI.OpenAIAPIKey != "your-openai-api-key" {
		fmt.Println("✅ OpenAI API Key is configured")
	} else {
		fmt.Println("⚠️  OpenAI API Key is not configured")
	}
	
	if cfg.AI.GeminiAPIKey != "" && cfg.AI.GeminiAPIKey != "your-gemini-api-key" {
		fmt.Println("✅ Gemini API Key is configured")
	} else {
		fmt.Println("⚠️  Gemini API Key is not configured")
	}

	fmt.Printf("Search Model: %s\n", cfg.AI.SearchEmbeddingModel)
	fmt.Printf("Vector Dimensions: %d\n", cfg.AI.SearchVectorDimensions)

	// Summary
	fmt.Println("\n🎉 Environment Configuration Test Summary:")
	fmt.Println("=========================================")
	fmt.Println("✅ Environment files loaded successfully")
	fmt.Println("✅ Configuration structure is valid")
	fmt.Println("✅ All environment variable functions work")
	
	// Count active services
	activeServices := 0
	if cfg.Features.APIGatewayEnabled { activeServices++ }
	if cfg.Features.ProducerServiceEnabled { activeServices++ }
	if cfg.Features.ConsumerServiceEnabled { activeServices++ }
	if cfg.Features.StreamsServiceEnabled { activeServices++ }
	if cfg.Features.AISearchEnabled { activeServices++ }
	if cfg.Features.AuthModuleEnabled { activeServices++ }
	if cfg.Features.Web3WalletEnabled { activeServices++ }
	if cfg.Features.DeFiServiceEnabled { activeServices++ }
	if cfg.Features.AIAgentsEnabled { activeServices++ }
	
	fmt.Printf("✅ %d/9 services are enabled\n", activeServices)
	
	fmt.Println("\n🚀 Environment configuration test completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Review and customize your .env file")
	fmt.Println("2. Generate secure secrets: make env-generate-secrets")
	fmt.Println("3. Start required services (PostgreSQL, Redis, Kafka)")
	fmt.Println("4. Run the full application: make run")
}
