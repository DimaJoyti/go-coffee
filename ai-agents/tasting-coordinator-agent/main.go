package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AgentName string `yaml:"agent_name"`
	LogLevel  string `yaml:"log_level"`
	// Add other configuration parameters as needed
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
	fmt.Println("Starting Tasting Coordinator Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// TODO: Implement agent logic here
	// - Receive new recipe proposals
	// - Access location availability and staff schedules
	// - Schedule and manage tasting sessions
	// - Collect and consolidate feedback from sessions

	fmt.Println("Tasting Coordinator Agent started successfully.")
}