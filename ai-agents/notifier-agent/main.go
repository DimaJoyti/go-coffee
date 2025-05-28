package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AgentName        string `yaml:"agent_name"`
	LogLevel         string `yaml:"log_level"`
	SlackWebhookURL  string `yaml:"slack_webhook_url"`
	KafkaBrokerAddress string `yaml:"kafka_broker_address"`
	KafkaInputTopic  string `yaml:"kafka_input_topic"`
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
	fmt.Println("Starting Notifier Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// Initialize Kafka consumer
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{config.KafkaBrokerAddress},
		Topic:    config.KafkaInputTopic,
		GroupID:  "notifier-agent-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	fmt.Printf("Listening for messages on Kafka topic: %s\n", config.KafkaInputTopic)
	consumeNotificationsFromKafka(r, config.SlackWebhookURL)

	fmt.Println("Notifier Agent stopped.")
}

func consumeNotificationsFromKafka(r *kafka.Reader, slackWebhookURL string) {
	for {
		m, err := r.FetchMessage(context.Background())
		if err != nil {
			log.Printf("Error fetching message from Kafka: %v", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		fmt.Printf("Received message from Kafka: topic=%s, partition=%d, offset=%d, key=%s, value=%s\n",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		var notification struct {
			Type string            `json:"type"`
			Data map[string]string `json:"data"`
		}

		if err := json.Unmarshal(m.Value, &notification); err != nil {
			log.Printf("Error unmarshaling Kafka message: %v", err)
			if err := r.CommitMessages(context.Background(), m); err != nil {
				log.Printf("Error committing message after unmarshaling failure: %v", err)
			}
			continue
		}

		handleNotification(slackWebhookURL, notification.Type, notification.Data)

		if err := r.CommitMessages(context.Background(), m); err != nil {
			log.Printf("Error committing message to Kafka: %v", err)
		}
	}
}

func handleNotification(webhookURL, notificationType string, data map[string]string) {
	var payload interface{}

	switch notificationType {
	case "new_beverage":
		payload = map[string]interface{}{
			"text": fmt.Sprintf("New Beverage Alert: %s - %s", data["name"], data["description"]),
			"blocks": []map[string]interface{}{
				{
					"type": "section",
					"text": map[string]string{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*New Beverage Alert!* :coffee:\n\n*Name:* %s\n*Description:* %s", data["name"], data["description"]),
					},
				},
			},
		}
	case "low_stock":
		payload = map[string]interface{}{
			"text": fmt.Sprintf("Low Stock Warning: %s - %s remaining!", data["item"], data["quantity"]),
			"blocks": []map[string]interface{}{
				{
					"type": "section",
					"text": map[string]string{
						"type": "mrkdwn",
						"text": fmt.Sprintf(":warning: *Low Stock Alert!* :warning:\n\n*Item:* %s\n*Remaining:* %s", data["item"], data["quantity"]),
					},
				},
			},
		}
	case "schedule_change":
		payload = map[string]interface{}{
			"text": fmt.Sprintf("Schedule Change: %s - %s", data["date"], data["details"]),
			"blocks": []map[string]interface{}{
				{
					"type": "section",
					"text": map[string]string{
						"type": "mrkdwn",
						"text": fmt.Sprintf(":calendar: *Schedule Update!* :calendar:\n\n*Date:* %s\n*Details:* %s", data["date"], data["details"]),
					},
				},
			},
		}
	default:
		payload = map[string]string{"text": fmt.Sprintf("Unhandled Notification Type: %s - Data: %v", notificationType, data)}
	}

	err := sendSlackNotification(webhookURL, payload)
	if err != nil {
		log.Printf("Error sending Slack notification for type %s: %v", notificationType, err)
	} else {
		fmt.Printf("Slack notification for type '%s' sent successfully.\n", notificationType)
	}
}

func sendSlackNotification(webhookURL string, payload interface{}) error {
	jsonMessage, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Slack: %s", resp.Status)
	}

	return nil
}