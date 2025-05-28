package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)
type Config struct {
	AgentName                      string `yaml:"agent_name"`
	LogLevel                       string `yaml:"log_level"`
	LLMAPIKey                      string `yaml:"llm_api_key"`
	KafkaBrokerAddress             string `yaml:"kafka_broker_address"`
	KafkaOutputTopicSocialMediaContent string `yaml:"kafka_output_topic_social_media_content"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func main() {
	fmt.Println("Starting Social Media Content Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// Example usage of content generation and notification
	contentInfo := map[string]string{
		"type":    "new_drink",
		"name":    "Espresso Martini Twist",
		"details": "A unique blend of rich espresso, vodka, and a hint of orange zest. Perfect for an evening pick-me-up!",
	}
	theme := "organized chaos"

	socialMediaContent, err := generateSocialMediaContent(contentInfo, theme, config.LLMAPIKey)
	if err != nil {
		log.Printf("Error generating social media content: %v", err)
	} else {
		fmt.Println("\nGenerated Social Media Content:")
		fmt.Println(socialMediaContent)
		// Use Kafka to send the content
		err = sendSocialMediaContentToKafka(config.KafkaBrokerAddress, config.KafkaOutputTopicSocialMediaContent, socialMediaContent)
		if err != nil {
			log.Printf("Error sending social media content to Kafka: %v", err)
		} else {
			fmt.Println("Social media content sent to Kafka successfully.")
		}
	}

	fmt.Println("Social Media Content Agent started successfully.")
}

// generateSocialMediaContent simulates content generation. In a real scenario, this would
// integrate with an LLM using the provided API key.
func generateSocialMediaContent(contentInfo map[string]string, theme string, llmAPIKey string) (string, error) {
	// This is a placeholder. In a real application, you would call an LLM API here.
	// For example, using llmAPIKey to authenticate with OpenAI, Gemini, etc.
	// The prompt would incorporate contentInfo and theme.

	prompt := fmt.Sprintf(
		"Generate a social media post (e.g., for Instagram or Twitter) about a %s: '%s - %s'. "+
			"The coffee shop's theme is '%s'. Make it engaging and reflect the theme.",
		contentInfo["type"], contentInfo["name"], contentInfo["details"], theme,
	)

	// Simulate LLM response
	generatedContent := fmt.Sprintf(
		"✨ New Brew Alert! ✨ Dive into the delightful '"+contentInfo["name"]+"' - "+
			contentInfo["details"]+". It's a symphony of flavors amidst our signature "+
			"'"+theme+"' vibe. Come get lost in the taste! #GoCoffeeCo #NewDrink #"+
			contentInfo["name"]+" #CoffeeChaos",
	)

	log.Printf("LLM API Key (for demonstration): %s", llmAPIKey) // Log API key for demonstration
	log.Printf("LLM Prompt: %s", prompt)

	return generatedContent, nil
}

// sendSocialMediaContentToKafka sends the generated content to a Kafka topic.
func sendSocialMediaContentToKafka(brokerAddress, topic string, content string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	msg := kafka.Message{
		Value: []byte(content),
	}

	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Message sent to Kafka topic %s\n", topic)
	return nil
}
