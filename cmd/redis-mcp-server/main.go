package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

func main() {
	log.Println("🚀 Starting Redis MCP Server...")

	// Initialize logger
	logger := logger.New("redis-mcp-server")

	// Get configuration from environment
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8090"
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	logger.Info("✅ Connected to Redis successfully")

	// Initialize Redis MCP server
	mcpServer := redismcp.NewMCPServer(redisClient, logger)

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Starting Redis MCP Server on port %s", serverPort)
		if err := mcpServer.Start(serverPort); err != nil {
			log.Fatalf("Failed to start Redis MCP server: %v", err)
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
		logger.WithFields(map[string]interface{}{"error": err}).Error("Error closing Redis connection")
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
				logger.WithFields(map[string]interface{}{
					"error": err, "shop": shopID, "item": item,
				}).Error("Failed to set menu item")
			}
		}
		logger.WithFields(map[string]interface{}{"shop": shopID}).Info("✅ Menu set for shop")
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
				logger.WithFields(map[string]interface{}{
					"error": err, "shop": shopID, "ingredient": ingredient,
				}).Error("Failed to set inventory item")
			}
		}
		logger.WithFields(map[string]interface{}{"shop": shopID}).Info("✅ Inventory set for shop")
	}

	// Sample available ingredients set
	ingredients := []string{
		"coffee_beans", "milk", "sugar", "oat_milk", "almond_milk",
		"coconut_milk", "soy_milk", "vanilla_syrup", "caramel_syrup",
		"chocolate_syrup", "whipped_cream", "cinnamon", "nutmeg",
	}

	for _, ingredient := range ingredients {
		if err := client.SAdd(ctx, "ingredients:available", ingredient).Err(); err != nil {
			logger.WithFields(map[string]interface{}{
				"error": err, "ingredient": ingredient,
			}).Error("Failed to add ingredient")
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
			logger.WithFields(map[string]interface{}{
				"error": err, "drink": drink,
			}).Error("Failed to add order count")
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
				logger.WithFields(map[string]interface{}{
					"error": err, "customer": customerKey, "field": field,
				}).Error("Failed to set customer data")
			}
		}
		logger.WithFields(map[string]interface{}{"customer": customerKey}).Info("✅ Customer data set")
	}

	// Sample analytics data
	analytics := map[string]float64{
		"revenue_today":         1250.75,
		"orders_today":          85,
		"avg_order_value":       14.71,
		"customer_satisfaction": 4.6,
	}

	for metric, value := range analytics {
		if err := client.ZAdd(ctx, "analytics:daily", &redis.Z{
			Score:  value,
			Member: metric,
		}).Err(); err != nil {
			logger.WithFields(map[string]interface{}{
				"error": err, "metric": metric,
			}).Error("Failed to add analytics")
		}
	}
	logger.Info("✅ Analytics data set")

	logger.Info("🎉 Sample data initialization completed successfully!")
}
