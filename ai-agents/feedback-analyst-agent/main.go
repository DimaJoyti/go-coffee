package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AgentName    string `yaml:"agent_name"`
	LogLevel     string `yaml:"log_level"`
	GoogleSheets struct {
		APIKey        string `yaml:"api_key"`
		SpreadsheetID string `yaml:"spreadsheet_id"`
		Range         string `yaml:"range"`
	} `yaml:"google_sheets"`
	Airtable struct {
		APIKey    string `yaml:"api_key"`
		BaseID    string `yaml:"base_id"`
		TableName string `yaml:"table_name"`
	} `yaml:"airtable"`
	NotifierAgent struct {
		URL string `yaml:"url"`
	} `yaml:"notifier_agent"`
	Kafka struct {
		BrokerAddress              string `yaml:"broker_address"`
		OutputTopicFeedbackSummary string `yaml:"output_topic_feedback_summary"`
	} `yaml:"kafka"`
}

type Feedback struct {
	ID      string
	Rating  int
	Comment string
	Item    string
}

type AirtableRecord struct {
	Fields map[string]interface{} `json:"fields"`
}

type AirtableResponse struct {
	Records []AirtableRecord `json:"records"`
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

func fetchFeedbackFromGoogleSheets(config *Config) ([]Feedback, error) {
	log.Printf("Fetching feedback from Google Sheets (simulated)... SpreadsheetID: %s, Range: %s\n", config.GoogleSheets.SpreadsheetID, config.GoogleSheets.Range)
	// Simulate fetching data from Google Sheets
	// In a real scenario, you would use the Google Sheets API client
	feedbackData := []Feedback{
		{ID: "1", Rating: 5, Comment: "Great coffee, loved the new blend!", Item: "Espresso"},
		{ID: "2", Rating: 4, Comment: "The crazy combination of lavender and espresso was surprisingly good!", Item: "Lavender Espresso"},
		{ID: "3", Rating: 2, Comment: "Too sweet, didn't like the caramel macchiato.", Item: "Caramel Macchiato"},
		{ID: "4", Rating: 5, Comment: "Amazing! The chili chocolate mocha is a must-try.", Item: "Chili Chocolate Mocha"},
		{ID: "5", Rating: 3, Comment: "Decent, but the cold brew was a bit watery.", Item: "Cold Brew"},
	}
	return feedbackData, nil
}

func analyzeFeedback(feedback []Feedback) (map[string]int, []string) {
	log.Println("Analyzing feedback...")
	trends := make(map[string]int)
	crazyCombinations := []string{}

	for _, fb := range feedback {
		// Simple sentiment analysis (very basic)
		if strings.Contains(strings.ToLower(fb.Comment), "great") || strings.Contains(strings.ToLower(fb.Comment), "loved") || strings.Contains(strings.ToLower(fb.Comment), "amazing") {
			trends["positive_feedback"]++
		} else if strings.Contains(strings.ToLower(fb.Comment), "sweet") || strings.Contains(strings.ToLower(fb.Comment), "watery") || strings.Contains(strings.ToLower(fb.Comment), "didn't like") {
			trends["negative_feedback"]++
		}

		// Identify "crazy combinations"
		if strings.Contains(strings.ToLower(fb.Comment), "crazy combination") || strings.Contains(strings.ToLower(fb.Comment), "surprisingly good") || strings.Contains(strings.ToLower(fb.Comment), "must-try") {
			crazyCombinations = append(crazyCombinations, fb.Item)
		}

		// Item popularity
		trends[fb.Item]++
	}
	return trends, crazyCombinations
}

func updateAirtable(config *Config, crazyCombinations []string) error {
	log.Printf("Updating Airtable... BaseID: %s, TableName: %s\n", config.Airtable.BaseID, config.Airtable.TableName)

	// Check if Airtable configuration is provided
	if config.Airtable.APIKey == "" || config.Airtable.BaseID == "" || config.Airtable.TableName == "" {
		log.Println("Airtable configuration incomplete. Simulating update...")
		for _, combo := range crazyCombinations {
			log.Printf("Simulating Airtable update for: %s\n", combo)
		}
		return nil
	}

	// Real Airtable API integration
	for _, combo := range crazyCombinations {
		log.Printf("Updating Airtable for combination: %s\n", combo)

		url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", config.Airtable.BaseID, config.Airtable.TableName)
		record := AirtableRecord{
			Fields: map[string]interface{}{
				"CombinationName": combo,
				"Status":          "Working",
				"LastUpdated":     time.Now().Format(time.RFC3339),
				"Source":          "Feedback Analysis",
			},
		}

		body, err := json.Marshal(map[string][]AirtableRecord{"records": {record}})
		if err != nil {
			log.Printf("Error marshaling Airtable record: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Error creating HTTP request: %v", err)
			continue
		}

		req.Header.Add("Authorization", "Bearer "+config.Airtable.APIKey)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request to Airtable: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			log.Printf("Airtable API returned status %d for combination %s", resp.StatusCode, combo)
			continue
		}

		log.Printf("Successfully updated Airtable for combination: %s", combo)
	}
	return nil
}

func sendFeedbackSummaryToKafka(config *Config, summary string) error {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.Kafka.BrokerAddress),
		Topic:    config.Kafka.OutputTopicFeedbackSummary,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   []byte(config.AgentName),
		Value: []byte(summary),
		Time:  time.Now(),
	}

	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Feedback summary sent to Kafka topic %s successfully.\n", config.Kafka.OutputTopicFeedbackSummary)
	return nil
}

func sendFeedbackSummary(config *Config, trends map[string]int, crazyCombinations []string) error {
	log.Println("Sending feedback summary...")
	summary := fmt.Sprintf("Feedback Analysis Summary:\nTrends: %+v\nCrazy Combinations that work: %+v\n", trends, crazyCombinations)

	if config.Kafka.BrokerAddress != "" && config.Kafka.OutputTopicFeedbackSummary != "" {
		return sendFeedbackSummaryToKafka(config, summary)
	} else {
		log.Println("Kafka configuration not complete. Printing summary to console:")
		fmt.Println(summary)
	}
	return nil
}

func main() {
	fmt.Println("Starting Feedback Analyst Agent...")

	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	feedbackData, err := fetchFeedbackFromGoogleSheets(config)
	if err != nil {
		log.Fatalf("Error fetching feedback: %v", err)
	}

	trends, crazyCombinations := analyzeFeedback(feedbackData)

	err = updateAirtable(config, crazyCombinations)
	if err != nil {
		log.Fatalf("Error updating Airtable: %v", err)
	}

	err = sendFeedbackSummary(config, trends, crazyCombinations)
	if err != nil {
		log.Fatalf("Error sending feedback summary: %v", err)
	}

	fmt.Println("Feedback Analyst Agent finished successfully.")
}
