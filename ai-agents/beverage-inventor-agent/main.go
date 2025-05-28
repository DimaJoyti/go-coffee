package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Config struct to hold configuration parameters
type Config struct {
	LLM struct {
		Provider string `yaml:"provider"`
		APIKey   string `yaml:"api_key"`
		Model    string `yaml:"model"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"llm"`
	Ingredients []struct {
		Name   string `yaml:"name"`
		Source string `yaml:"source"`
	} `yaml:"ingredients"`
	Kafka struct {
		Enabled     bool   `yaml:"enabled"`
		BrokerAddress string `yaml:"broker_address"`
		OutputTopic string `yaml:"output_topic"`
	} `yaml:"kafka"`
	TaskManager struct {
		Enabled    bool   `yaml:"enabled"`
		APIEndpoint string `yaml:"api_endpoint"`
		APIKey     string `yaml:"api_key"`
	} `yaml:"task_manager"`
}

// BeverageIdea struct to hold the generated beverage idea
type BeverageIdea struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ingredient  string `json:"ingredient"`
	Theme       string `json:"theme"`
}

var cfg Config

// Simple "dictionary" for beverage generation
var (
	adjectives = []string{"Exotic", "Mystical", "Vibrant", "Bold", "Smooth", "Crisp", "Sparkling", "Dreamy", "Enchanted", "Galactic"}
	beverageTypes = []string{"Latte", "Elixir", "Brew", "Infusion", "Concoction", "Nectar", "Blend", "Quencher", "Potion", "Essence"}
	themes = map[string][]string{
		"Mars Base": {
			"perfect for a Martian sunrise",
			"a taste of the red planet",
			"fuels your interplanetary journey",
			"inspired by the Martian landscape",
		},
		"Lunar Mining Corp": {
			"a true taste of the lunar surface",
			"energizes your moon rock excavation",
			"crafted for the lunar explorer",
			"reflects the stark beauty of the moon",
		},
		"Interstellar Trade Federation": {
			"transcends galaxies",
			"a cosmic concoction",
			"for the discerning space traveler",
			"unites flavors from across the cosmos",
		},
	}
)

func init() {
	rand.Seed(time.Now().UnixNano()) // Initialize random seed
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

// generateBeverageIdea generates a creative beverage idea based on an ingredient and theme
func generateBeverageIdea(ingredient, theme string) BeverageIdea {
	// Select random adjective and beverage type
	adj := adjectives[rand.Intn(len(adjectives))]
	bevType := beverageTypes[rand.Intn(len(beverageTypes))]

	var beverageName, description string

	// Get theme-specific phrases
	themePhrases, ok := themes[theme]
	if !ok || len(themePhrases) == 0 {
		themePhrases = []string{"a delightful new addition to our menu", "a unique blend"}
	}
	themePhrase := themePhrases[rand.Intn(len(themePhrases))]

	// Generate name and description using templates
	beverageName = fmt.Sprintf("%s %s %s", adj, ingredient, bevType)
	description = fmt.Sprintf("A %s featuring %s, %s. %s", bevType, ingredient, adj, themePhrase)

	return BeverageIdea{
		Name:        beverageName,
		Description: description,
		Ingredient:  ingredient,
		Theme:       theme,
	}
}

// formulateBaristaTask creates a structured description for the barista
func formulateBaristaTask(idea BeverageIdea) string {
	task := fmt.Sprintf(`
New Beverage Idea: %s
Description: %s
Key Ingredient: %s
Coffee Shop Theme: %s

Instructions for Barista:
1. Experiment with %s to create a unique flavor profile.
2. Consider presentation that aligns with the "%s" theme.
3. Document the recipe and preparation steps.
4. Prepare a sample for tasting and feedback.
`, idea.Name, idea.Description, idea.Ingredient, idea.Theme, idea.Ingredient, idea.Theme)

	return task
}

// sendBeverageIdeaToKafka sends the beverage idea to Kafka
func sendBeverageIdeaToKafka(idea BeverageIdea) error {
	if !cfg.Kafka.Enabled {
		log.Println("Kafka integration is disabled. Skipping sending beverage idea to Kafka.")
		return nil
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.BrokerAddress},
		Topic:   cfg.Kafka.OutputTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// Convert BeverageIdea to JSON
	messageValue, err := json.Marshal(idea)
	if err != nil {
		return fmt.Errorf("failed to marshal beverage idea to JSON: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(idea.Name),
		Value: messageValue,
	}

	err = writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	fmt.Printf("Successfully sent beverage idea '%s' to Kafka topic '%s'\n", idea.Name, cfg.Kafka.OutputTopic)
	return nil
}

func main() {
	configPath := "ai-agents/beverage-inventor-agent/config.yaml"
	err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	fmt.Println("Beverage Inventor Agent Started.")
	fmt.Printf("LLM Provider: %s, Model: %s\n", cfg.LLM.Provider, cfg.LLM.Model)
	fmt.Printf("Kafka Integration Enabled: %t, Broker: %s, Topic: %s\n", cfg.Kafka.Enabled, cfg.Kafka.BrokerAddress, cfg.Kafka.OutputTopic)

	// Simulate new ingredient discovery
	if len(cfg.Ingredients) == 0 {
		log.Println("No ingredients defined in config.yaml to simulate discovery.")
		return
	}

	// Pick a random ingredient for demonstration
	randomIndex := rand.Intn(len(cfg.Ingredients))
	discoveredIngredient := cfg.Ingredients[randomIndex]

	fmt.Printf("\nDiscovered new ingredient: %s from %s\n", discoveredIngredient.Name, discoveredIngredient.Source)

	// Generate beverage idea
	// For demonstration, we'll use a fixed theme or derive it from the ingredient source
	theme := "Mars Base" // Default theme
	if discoveredIngredient.Source == "Lunar Mining Corp" {
		theme = "Lunar Mining Corp"
	} else if discoveredIngredient.Source == "Interstellar Trade Federation" {
		theme = "Interstellar Trade Federation"
	}

	idea := generateBeverageIdea(discoveredIngredient.Name, theme)
	fmt.Printf("\nGenerated Beverage Idea:\n  Name: %s\n  Description: %s\n", idea.Name, idea.Description)

	// Formulate task for barista (optional, for logging/display purposes)
	baristaTask := formulateBaristaTask(idea)
	fmt.Println("\n--- Task for Barista ---")
	fmt.Println(baristaTask)
	fmt.Println("------------------------")

	// Send beverage idea to Kafka
	err = sendBeverageIdeaToKafka(idea)
	if err != nil {
		log.Fatalf("Failed to send beverage idea to Kafka: %v", err)
	}
}