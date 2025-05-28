package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
"bytes"
	"encoding/json"
	"net/http"
)

type Config struct {
	AgentName          string `yaml:"agent_name"`
	LogLevel           string `yaml:"log_level"`
	ClickUpAPIKey      string `yaml:"clickup_api_key"`
	ClickUpWorkspaceID string `yaml:"clickup_workspace_id"`
	ClickUpSpaceID     string `yaml:"clickup_space_id"`
	ClickUpListID      string `yaml:"clickup_list_id"`
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
	fmt.Println("Starting Task Manager Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// Implement agent logic here
	// - Create, assign, and track tasks in ClickUp
	// - Receive new recipe proposals, schedule changes, inventory alerts

	// Example: Create a task in ClickUp
	// taskName := "Test Task from Agent"
	// taskDescription := "This is a test task created by the Task Manager Agent."
	// err = createTask(config, taskName, taskDescription)
	// if err != nil {
	// 	log.Printf("Error creating task: %v", err)
	// } else {
	// 	fmt.Println("Task created successfully in ClickUp.")
	// }

	fmt.Println("Task Manager Agent started successfully.")
}
func createTask(config *Config, taskName, taskDescription string) error {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/tasks", config.ClickUpListID)

	taskPayload := map[string]interface{}{
		"name":        taskName,
		"description": taskDescription,
		"status":      "to do", // Default status
		// Add other fields as needed, e.g., assignees, priority, due_date
	}

	jsonPayload, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", config.ClickUpAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create task, status code: %d", resp.StatusCode)
	}

	return nil
}