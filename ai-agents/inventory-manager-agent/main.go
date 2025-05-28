package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Config defines the structure for the agent's configuration
type Config struct {
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
		BrokerAddress          string `yaml:"broker_address"`
		OutputTopicInventoryUpdate string `yaml:"output_topic_inventory_update"`
		OutputTopicLowStock    string `yaml:"output_topic_low_stock"`
	} `yaml:"kafka"`
	LowStockThresholds map[string]int `yaml:"low_stock_thresholds"`
}

// Inventory represents the current stock levels of ingredients
type Inventory struct {
	sync.Mutex
	Levels map[string]int
}

var (
	cfg         Config
	inventory   Inventory
	kafkaWriter *kafka.Writer
)

func init() {
	inventory = Inventory{
		Levels: make(map[string]int),
	}
}

// readConfig reads the configuration from the config.yaml file
func readConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// updateInventory simulates updating the inventory system
func updateInventory(ingredient string, quantity int) {
	inventory.Lock()
	defer inventory.Unlock()

	inventory.Levels[ingredient] += quantity
	log.Printf("Inventory updated: %s, new quantity: %d", ingredient, inventory.Levels[ingredient])

	// Simulate API call to Monday inventory board
	log.Printf("Simulating update to Monday inventory board (API Key: %s, Board ID: %s) for %s: %d",
		cfg.InventorySystem.APIKey, cfg.InventorySystem.BoardID, ingredient, inventory.Levels[ingredient])

	sendInventoryUpdateToKafka(ingredient, inventory.Levels[ingredient])
}

// checkStockLevels checks the current stock levels against predefined thresholds
func checkStockLevels() {
	inventory.Lock()
	defer inventory.Unlock()

	log.Println("Checking stock levels...")
	for ingredient, threshold := range cfg.LowStockThresholds {
		currentLevel, exists := inventory.Levels[ingredient]
		if !exists || currentLevel <= threshold {
			log.Printf("Low stock alert for %s: current level %d, threshold %d", ingredient, currentLevel, threshold)
			sendLowStockNotificationToKafka(ingredient, currentLevel)
		} else {
			log.Printf("Stock level for %s is healthy: %d", ingredient, currentLevel)
		}
	}
}

// sendInventoryUpdateToKafka sends inventory update information to Kafka
func sendInventoryUpdateToKafka(ingredient string, quantity int) {
	message := map[string]interface{}{
		"ingredient": ingredient,
		"quantity":   quantity,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal inventory update message: %v", err)
		return
	}

	err = kafkaWriter.WriteMessages(
		nil, // context.Background()
		kafka.Message{
			Key:   []byte(ingredient),
			Value: jsonMessage,
		},
	)
	if err != nil {
		log.Printf("Failed to write inventory update message to Kafka: %v", err)
	} else {
		log.Printf("Sent inventory update to Kafka topic '%s': %s", cfg.Kafka.OutputTopicInventoryUpdate, string(jsonMessage))
	}
}

// sendLowStockNotificationToKafka sends low stock notifications to Kafka
func sendLowStockNotificationToKafka(ingredient string, currentLevel int) {
	message := map[string]interface{}{
		"ingredient":   ingredient,
		"currentLevel": currentLevel,
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal low stock notification message: %v", err)
		return
	}

	err = kafkaWriter.WriteMessages(
		nil, // context.Background()
		kafka.Message{
			Key:   []byte(ingredient),
			Value: jsonMessage,
		},
	)
	if err != nil {
		log.Printf("Failed to write low stock notification message to Kafka: %v", err)
	} else {
		log.Printf("Sent low stock notification to Kafka topic '%s': %s", cfg.Kafka.OutputTopicLowStock, string(jsonMessage))
	}
}

func main() {
	// Read configuration
	err := readConfig("ai-agents/inventory-manager-agent/config.yaml")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	// Initialize Kafka writer
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(cfg.Kafka.BrokerAddress),
		Topic:    cfg.Kafka.OutputTopicInventoryUpdate, // Default topic, will be overridden in send functions
		Balancer: &kafka.LeastBytes{},
		BatchTimeout: 1 * time.Second,
	}
	defer kafkaWriter.Close()

	log.Println("Inventory Manager Agent started.")

	// Initial inventory setup (for demonstration)
	updateInventory("coffee_beans", 100)
	updateInventory("milk", 50)
	updateInventory("sugar", 30)

	// Simulate some usage
	updateInventory("coffee_beans", -15)
	updateInventory("milk", -10)

	// Check stock levels
	checkStockLevels()

	// Simulate new delivery
	updateInventory("coffee_beans", 50)
	updateInventory("milk", 20)

	// Check stock levels again after delivery
	checkStockLevels()
}