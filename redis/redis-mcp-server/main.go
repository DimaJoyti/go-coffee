package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

	aisimple "github.com/DimaJoyti/go-coffee/pkg/ai-simple"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

func main() {
	// Initialize logger
	logger := logger.New("redis-mcp-server")
	logger.Info("🚀 Starting Redis MCP Server...")

	// Get configuration from environment
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	// AI configuration (not used with simple AI service)
	_ = os.Getenv("GEMINI_API_KEY")
	_ = os.Getenv("OLLAMA_BASE_URL")

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8090"
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.ErrorMap("Failed to parse Redis URL", map[string]interface{}{"error": err})
		return
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.ErrorMap("Failed to connect to Redis", map[string]interface{}{"error": err})
		return
	}

	logger.Info("✅ Connected to Redis successfully")

	// Initialize simple AI service
	aiService := aisimple.NewService("redis-mcp-ai")

	logger.Info("✅ AI service initialized successfully")

	// Initialize Redis MCP server
	mcpServer := redismcp.NewMCPServer(redisClient, aiService, logger)

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Starting Redis MCP Server on port " + serverPort)
		if err := mcpServer.Start(serverPort); err != nil {
			logger.ErrorMap("Failed to start Redis MCP server", map[string]interface{}{"error": err})
		}
	}()

	// Initialize sample data in Redis
	go initializeSampleData(redisClient, logger)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Redis MCP Server is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Redis MCP Server...")
	mcpServer.Stop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.ErrorMap("Error closing Redis connection", map[string]interface{}{"error": err})
	}

	logger.Info("✅ Redis MCP Server stopped gracefully")
}

// initializeSampleData populates Redis with sample coffee shop data
func initializeSampleData(client *redis.Client, logger *logger.Logger) {
	ctx := context.Background()

	logger.Info("🏪 Initializing sample coffee shop data in Redis...")

	// Sample coffee shop menus
	shops := map[string]map[string]string{
		"downtown": {
			"latte":      "4.50",
			"cappuccino": "4.00",
			"espresso":   "2.50",
			"americano":  "3.00",
			"macchiato":  "4.25",
		},
		"uptown": {
			"latte":      "4.75",
			"cappuccino": "4.25",
			"espresso":   "2.75",
			"americano":  "3.25",
			"mocha":      "5.00",
		},
		"westside": {
			"latte":      "4.25",
			"cappuccino": "3.75",
			"espresso":   "2.25",
			"americano":  "2.75",
			"flat_white": "4.50",
		},
	}

	// Set coffee shop menus
	for shopID, menu := range shops {
		menuKey := "coffee:menu:" + shopID
		for item, price := range menu {
			if err := client.HSet(ctx, menuKey, item, price).Err(); err != nil {
				logger.ErrorMap("Failed to set menu item", map[string]interface{}{
					"error": err, "shop": shopID, "item": item,
				})
			}
		}
		logger.InfoMap("✅ Menu set for shop", map[string]interface{}{"shop": shopID})
	}

	// Sample inventory data
	inventories := map[string]map[string]string{
		"downtown": {
			"coffee_beans": "150",
			"milk":         "75",
			"sugar":        "50",
			"oat_milk":     "25",
			"almond_milk":  "20",
		},
		"uptown": {
			"coffee_beans": "200",
			"milk":         "100",
			"sugar":        "60",
			"oat_milk":     "30",
			"coconut_milk": "15",
		},
		"westside": {
			"coffee_beans": "120",
			"milk":         "60",
			"sugar":        "40",
			"oat_milk":     "20",
			"soy_milk":     "25",
		},
	}

	// Set inventory data
	for shopID, inventory := range inventories {
		inventoryKey := "coffee:inventory:" + shopID
		for ingredient, quantity := range inventory {
			if err := client.HSet(ctx, inventoryKey, ingredient, quantity).Err(); err != nil {
				logger.ErrorMap("Failed to set inventory item", map[string]interface{}{
					"error": err, "shop": shopID, "ingredient": ingredient,
				})
			}
		}
		logger.InfoMap("✅ Inventory set for shop", map[string]interface{}{"shop": shopID})
	}

	// Sample available ingredients set
	ingredients := []string{
		"coffee_beans", "milk", "sugar", "oat_milk", "almond_milk",
		"coconut_milk", "soy_milk", "vanilla_syrup", "caramel_syrup",
		"chocolate_syrup", "whipped_cream", "cinnamon", "nutmeg",
	}

	for _, ingredient := range ingredients {
		if err := client.SAdd(ctx, "ingredients:available", ingredient).Err(); err != nil {
			logger.ErrorMap("Failed to add ingredient", map[string]interface{}{
				"error": err, "ingredient": ingredient,
			})
		}
	}
	logger.Info("✅ Available ingredients set")

	// Sample daily orders (sorted set)
	orders := map[string]float64{
		"latte":      150,
		"cappuccino": 120,
		"americano":  100,
		"espresso":   80,
		"macchiato":  60,
		"mocha":      45,
		"flat_white": 30,
	}

	for drink, count := range orders {
		if err := client.ZAdd(ctx, "coffee:orders:today", &redis.Z{
			Score:  count,
			Member: drink,
		}).Err(); err != nil {
			logger.ErrorMap("Failed to add order count", map[string]interface{}{
				"error": err, "drink": drink,
			})
		}
	}
	logger.Info("✅ Daily orders data set")

	// Sample customer data
	customers := map[string]map[string]string{
		"customer:123": {
			"name":           "John Doe",
			"email":          "john@example.com",
			"favorite_drink": "latte",
			"loyalty_points": "150",
			"visits":         "25",
		},
		"customer:456": {
			"name":           "Jane Smith",
			"email":          "jane@example.com",
			"favorite_drink": "cappuccino",
			"loyalty_points": "200",
			"visits":         "30",
		},
		"customer:789": {
			"name":           "Bob Johnson",
			"email":          "bob@example.com",
			"favorite_drink": "americano",
			"loyalty_points": "75",
			"visits":         "12",
		},
	}

	for customerKey, data := range customers {
		for field, value := range data {
			if err := client.HSet(ctx, customerKey, field, value).Err(); err != nil {
				logger.ErrorMap("Failed to set customer data", map[string]interface{}{
					"error": err, "customer": customerKey, "field": field,
				})
			}
		}
		logger.InfoMap("✅ Customer data set", map[string]interface{}{"customer": customerKey})
	}

	// Sample analytics data
	analytics := map[string]float64{
		"revenue_today":     1250.75,
		"orders_today":      85,
		"avg_order_value":   14.71,
		"customer_satisfaction": 4.6,
	}

	for metric, value := range analytics {
		if err := client.ZAdd(ctx, "analytics:daily", &redis.Z{
			Score:  value,
			Member: metric,
		}).Err(); err != nil {
			logger.ErrorMap("Failed to add analytics", map[string]interface{}{
				"error": err, "metric": metric,
			})
		}
	}
	logger.Info("✅ Analytics data set")

	logger.Info("🎉 Sample data initialization completed successfully!")
}
