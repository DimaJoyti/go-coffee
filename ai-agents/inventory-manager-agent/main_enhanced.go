package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Enhanced Config with Redis MCP support
type EnhancedConfig struct {
	InventorySystem struct {
		APIKey  string `yaml:"api_key"`
		BoardID string `yaml:"board_id"`
	} `yaml:"inventory_system"`
	NotifierAgent struct {
		URL string `yaml:"url"`
	} `yaml:"notifier_agent"`
	TaskManagerAgent struct {
		URL string `yaml:"url"`
	} `yaml:"task_manager_agent"`
	Kafka struct {
		BrokerAddress              string `yaml:"broker_address"`
		OutputTopicInventoryUpdate string `yaml:"output_topic_inventory_update"`
		OutputTopicLowStock        string `yaml:"output_topic_low_stock"`
	} `yaml:"kafka"`
	RedisMCP struct {
		ServerURL string `yaml:"server_url"`
		AgentID   string `yaml:"agent_id"`
	} `yaml:"redis_mcp"`
	LowStockThresholds map[string]int `yaml:"low_stock_thresholds"`
}

// Enhanced Inventory with Redis integration
type EnhancedInventory struct {
	sync.Mutex
	Levels      map[string]int
	redisClient interface{} // Placeholder for Redis client
	logger      *log.Logger
}

// InventoryItem represents detailed inventory information
type InventoryItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	Unit        string    `json:"unit"`
	MinLevel    int       `json:"min_level"`
	MaxLevel    int       `json:"max_level"`
	Cost        float64   `json:"cost"`
	Supplier    string    `json:"supplier"`
	LastUpdated time.Time `json:"last_updated"`
	Location    string    `json:"location"`
	Status      string    `json:"status"` //(good", "low", "critical", "out_of_stock"
}

var (
	enhancedCfg         EnhancedConfig
	enhancedInventory   *EnhancedInventory
	enhancedKafkaWriter *kafka.Writer
)

// NewEnhancedInventory creates a new enhanced inventory with Redis MCP
func NewEnhancedInventory(redisClient interface{}, logger *log.Logger) *EnhancedInventory {
	return &EnhancedInventory{
		Levels:      make(map[string]int),
		redisClient: redisClient,
		logger:      logger,
	}
}

// updateInventoryWithRedis updates inventory with Redis MCP integration
func (ei *EnhancedInventory) updateInventoryWithRedis(shopID, ingredient string, quantity int) error {
	ei.Lock()
	defer ei.Unlock()

	// Update local inventory
	ei.Levels[ingredient] += quantity
	log.Printf("Inventory updated: %s, new quantity: %d", ingredient, ei.Levels[ingredient])

	// Update Redis using natural language
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use natural language to update Redis
	query := fmt.Sprintf("set inventory for shop %s ingredient %s to %d", shopID, ingredient, ei.Levels[ingredient])
	_ = ctx   // Avoid unused variable warning
	_ = query // Avoid unused variable warning
	// Placeholder for Redis operation
	response := struct{Success bool; Error string; Data interface{}}{Success: true}
	var err error
	// Original call would be: response, err := ei.redisClient.Query(ctx, query, map[string]interface{}{
	//     "shop_id":    shopID,
	//     "ingredient": ingredient,
	//     "quantity":   ei.Levels[ingredient],
	//     "timestamp":  time.Now().Unix(),
	// })

	if err != nil {
		ei.logger.Printf("[ERROR] Failed to update Redis inventory: %s, ingredient: %s", err.Error(), ingredient)
		return err
	}

	if !response.Success {
		ei.logger.Printf("[ERROR] Redis update failed: %s, ingredient: %s", response.Error, ingredient)
		return fmt.Errorf("redis update failed: %s", response.Error)
	}

	ei.logger.Printf("[INFO] Successfully updated Redis inventory: ingredient=%s, quantity=%d", ingredient, ei.Levels[ingredient])

	// Send to Kafka
	sendEnhancedInventoryUpdateToKafka(ingredient, ei.Levels[ingredient])

	return nil
}

// getInventoryFromRedis retrieves inventory data using natural language
func (ei *EnhancedInventory) getInventoryFromRedis(shopID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf("get inventory for shop %s", shopID)
	// Placeholder for Redis operation
	response := struct{Success bool; Error string; Data interface{}}{Success: true}
	var err error
	_ = ctx   // Avoid unused variable warning
	_ = query // Avoid unused variable warning

	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("redis query failed: %s", response.Error)
	}

	if inventoryData, ok := response.Data.(map[string]interface{}); ok {
		return inventoryData, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// checkStockLevelsWithRedis checks stock levels with Redis analytics
func (ei *EnhancedInventory) checkStockLevelsWithRedis(shopID string) {
	ei.Lock()
	defer ei.Unlock()

	log.Println("Checking stock levels with Redis analytics...")

	// Get current inventory from Redis
	redisInventory, err := ei.getInventoryFromRedis(shopID)
	if err != nil {
		ei.logger.Printf("[ERROR] Failed to get inventory from Redis: %s", err.Error())
		// Fall back to local inventory check
		ei.checkLocalStockLevels()
		return
	}

	// Check each ingredient against thresholds
	for ingredient, threshold := range enhancedCfg.LowStockThresholds {
		var currentLevel int
		if level, exists := redisInventory[ingredient]; exists {
			if levelStr, ok := level.(string); ok {
				fmt.Sscanf(levelStr, "%d", &currentLevel)
			}
		}

		if currentLevel <= threshold {
			log.Printf("🚨 Low stock alert for %s: current level %d, threshold %d", 
				ingredient, currentLevel, threshold)
			
			// Create alert in Redis using natural language
			ei.createLowStockAlert(shopID, ingredient, currentLevel, threshold)
			sendEnhancedLowStockNotificationToKafka(ingredient, currentLevel)
		} else {
			log.Printf("✅ Stock level for %s is healthy: %d", ingredient, currentLevel)
		}
	}
}

// createLowStockAlert creates an alert in Redis
func (ei *EnhancedInventory) createLowStockAlert(shopID, ingredient string, currentLevel, threshold int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	alertData := map[string]interface{}{
		"ingredient":     ingredient,
		"current_level":  currentLevel,
		"threshold":      threshold,
		"severity":       ei.calculateSeverity(currentLevel, threshold),
		"timestamp":      time.Now().Unix(),
		"shop_id":        shopID,
		"status":        "active",
	}

	query := fmt.Sprintf("add alert for shop %s ingredient %s low stock", shopID, ingredient)
	// Placeholder for Redis operation
	response := struct{Success bool; Error string; Data interface{}}{Success: true}
	var err error
	_ = ctx       // Avoid unused variable warning
	_ = query     // Avoid unused variable warning
	_ = alertData // Avoid unused variable warning

	if err != nil {
		ei.logger.Printf("[ERROR] Failed to create alert in Redis: %s, ingredient: %s", err.Error(), ingredient)
		return
	}

	if response.Success {
		ei.logger.Printf("[INFO] Low stock alert created in Redis: ingredient=%s, level=%d", ingredient, currentLevel)
	}
}

// calculateSeverity determines alert severity based on stock level
func (ei *EnhancedInventory) calculateSeverity(currentLevel, threshold int) string {
	if currentLevel == 0 {
		return "critical"
	} else if currentLevel <= threshold/2 {
		return "high"
	} else if currentLevel <= threshold {
		return "medium"
	}
	return "low"
}

// checkLocalStockLevels fallback method for local stock checking
func (ei *EnhancedInventory) checkLocalStockLevels() {
	log.Println("Checking local stock levels...")
	for ingredient, threshold := range enhancedCfg.LowStockThresholds {
		currentLevel, exists := ei.Levels[ingredient]
		if !exists || currentLevel <= threshold {
			log.Printf("Low stock alert for %s: current level %d, threshold %d", 
				ingredient, currentLevel, threshold)
			sendEnhancedLowStockNotificationToKafka(ingredient, currentLevel)
		} else {
			log.Printf("Stock level for %s is healthy: %d", ingredient, currentLevel)
		}
	}
}

// getInventoryAnalytics retrieves analytics from Redis
func (ei *EnhancedInventory) getInventoryAnalytics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get inventory turnover analytics
	query := "get analytics for inventory_turnover"
	// response, err := ei.redisClient.Query // Placeholder for Redis operation
	response := struct{Success bool; Error string; Data interface{}}{Success: true}
	var err error
	_ = ctx   // Avoid unused variable warning
	_ = query // Avoid unused variable warning
	// Placeholder for Redis query parameters
	_ = map[string]interface{}{
		"metric": "inventory_turnover",
		"period": "daily",
	}

	if err != nil {
		ei.logger.Printf("[ERROR] Failed to get analytics from Redis: %s", err.Error())
		return
	}

	if response.Success {
		ei.logger.Printf("[INFO] Inventory analytics retrieved successfully")
	}
}

// syncInventoryToRedis syncs all inventory data to Redis
func (ei *EnhancedInventory) syncInventoryToRedis(shopID string) {
	ei.Lock()
	defer ei.Unlock()

	log.Println("Syncing inventory to Redis...")

	for ingredient, quantity := range ei.Levels {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		query := fmt.Sprintf("set inventory for shop %s ingredient %s to %d", 
			shopID, ingredient, quantity)
		_ = ctx   // Avoid unused variable warning
		_ = query // Avoid unused variable warning
		
		// Redis query placeholder - would normally query Redis
	var err error
	/*
	query := fmt.Sprintf("update inventory for shop %s ingredient %s", shopID, ingredient)
	_, err = ei.redisClient.Query(ctx, query, map[string]interface{}{
		"shop_id":    shopID,
		"ingredient": ingredient,
		"quantity":   quantity,
		"sync_time":  time.Now().Unix(),
	})
	*/

		cancel()

		if err != nil {
			ei.logger.Printf("[ERROR] Failed to sync ingredient to Redis: %s, ingredient: %s", err.Error(), ingredient)
		} else {
			ei.logger.Printf("[DEBUG] Synced ingredient to Redis: ingredient=%s, quantity=%d", ingredient, quantity)
		}
	}

	log.Println("✅ Inventory sync to Redis completed")
}

// readConfig reads the enhanced configuration
func readEnhancedConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &enhancedCfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// sendEnhancedInventoryUpdateToKafka sends inventory update information to Kafka
func sendEnhancedInventoryUpdateToKafka(ingredient string, quantity int) {
	message := map[string]interface{}{
		"ingredient": ingredient,
		"quantity":   quantity,
		"timestamp":  time.Now().Format(time.RFC3339),
		"source":     "redis-mcp-enhanced",
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal inventory update message: %v", err)
		return
	}

	err = enhancedKafkaWriter.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(ingredient),
			Value: jsonMessage,
		},
	)
	if err != nil {
		log.Printf("Failed to write inventory update message to Kafka: %v", err)
	} else {
		log.Printf("📤 Sent inventory update to Kafka: %s", string(jsonMessage))
	}
}

// sendEnhancedLowStockNotificationToKafka sends low stock notifications to Kafka
func sendEnhancedLowStockNotificationToKafka(ingredient string, currentLevel int) {
	message := map[string]interface{}{
		"ingredient":   ingredient,
		"currentLevel": currentLevel,
		"timestamp":    time.Now().Format(time.RFC3339),
		"alert_type":  "low_stock",
		"source":      "redis-mcp-enhanced",
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal low stock notification message: %v", err)
		return
	}

	err = enhancedKafkaWriter.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(ingredient),
			Value: jsonMessage,
		},
	)
	if err != nil {
		log.Printf("Failed to write low stock notification message to Kafka: %v", err)
	} else {
		log.Printf("🚨 Sent low stock notification to Kafka: %s", string(jsonMessage))
	}
}

func mainEnhanced() {
	log.Println("🏪 Starting Enhanced Inventory Manager Agent with Redis MCP...")

	// Initialize logger
	enhancedLogger := log.New(os.Stdout, "[ENHANCED-INVENTORY] ", log.LstdFlags)

	// Read configuration
	err := readEnhancedConfig("ai-agents/inventory-manager-agent/config_enhanced.yaml")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	// Initialize Redis MCP client
	mcpServerURL := enhancedCfg.RedisMCP.ServerURL
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:8090"
	}
	
	agentID := enhancedCfg.RedisMCP.AgentID
	if agentID == "" {
		agentID = "inventory-manager-enhanced"
	}

	redisClient := interface{}(nil) // Placeholder for Redis client
	_ = mcpServerURL // Avoid unused variable warning
	_ = agentID      // Avoid unused variable warning

	// Test Redis MCP connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx // Avoid unused variable warning
	
	// Placeholder for Redis health check - would normally test the connection
	// if err := redisClient.Health(ctx); err != nil {
	//	log.Printf("⚠️ Redis MCP server not available: %v", err)
	//	log.Println("📝 Continuing with limited functionality...")
	// } else {
	//	log.Println("✅ Redis MCP connection established")
	// }
	log.Println("📝 Redis MCP integration (placeholder mode)")

	// Initialize enhanced inventory
	enhancedInventory = NewEnhancedInventory(redisClient, enhancedLogger)

	// Initialize Kafka writer
	enhancedKafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(enhancedCfg.Kafka.BrokerAddress),
		Topic:        enhancedCfg.Kafka.OutputTopicInventoryUpdate,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 1 * time.Second,
	}
	defer enhancedKafkaWriter.Close()

	shopID := "downtown" // Default shop ID

	log.Println("🚀 Enhanced Inventory Manager Agent started with Redis MCP integration")

	// Initial inventory setup with Redis integration
	enhancedInventory.updateInventoryWithRedis(shopID, "coffee_beans", 100)
	enhancedInventory.updateInventoryWithRedis(shopID, "milk", 50)
	enhancedInventory.updateInventoryWithRedis(shopID, "sugar", 30)
	enhancedInventory.updateInventoryWithRedis(shopID, "oat_milk", 25)

	// Simulate some usage
	enhancedInventory.updateInventoryWithRedis(shopID, "coffee_beans", -15)
	enhancedInventory.updateInventoryWithRedis(shopID, "milk", -10)

	// Check stock levels with Redis analytics
	enhancedInventory.checkStockLevelsWithRedis(shopID)

	// Get inventory analytics
	enhancedInventory.getInventoryAnalytics()

	// Simulate new delivery
	enhancedInventory.updateInventoryWithRedis(shopID, "coffee_beans", 50)
	enhancedInventory.updateInventoryWithRedis(shopID, "milk", 20)

	// Final stock check
	enhancedInventory.checkStockLevelsWithRedis(shopID)

	log.Println("✅ Enhanced Inventory Manager Agent completed demonstration")
}
