package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AgentName                      string `yaml:"agent_name"`
	LogLevel                       string `yaml:"log_level"`
	CalendarAPIKey                 string `yaml:"calendar_api_key"`
	TaskManagerAgentURL            string `yaml:"task_manager_agent_url"`
	DefaultTastingDurationMinutes  int    `yaml:"default_tasting_duration_minutes"`
	DefaultRescheduleBufferHours   int    `yaml:"default_reschedule_buffer_hours"`
	KafkaBrokerAddress             string `yaml:"kafka_broker_address"`
	KafkaOutputTopicTasting        string `yaml:"kafka_output_topic_tasting"`
	KafkaOutputTopicScheduleChange string `yaml:"kafka_output_topic_schedule_change"`
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
	fmt.Println("Starting Scheduler Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)
	fmt.Printf("Calendar API Key: %s\n", config.CalendarAPIKey)
	fmt.Printf("Task Manager Agent URL: %s\n", config.TaskManagerAgentURL)
	fmt.Printf("Default Tasting Duration: %d minutes\n", config.DefaultTastingDurationMinutes)
	fmt.Printf("Default Reschedule Buffer: %d hours\n", config.DefaultRescheduleBufferHours)
	fmt.Printf("Kafka Broker Address: %s\n", config.KafkaBrokerAddress)
	fmt.Printf("Kafka Output Topic (Tasting): %s\n", config.KafkaOutputTopicTasting)
	fmt.Printf("Kafka Output Topic (Schedule Change): %s\n", config.KafkaOutputTopicScheduleChange)

	// Demonstrate the improved flow: receive request -> execute logic -> publish to Kafka

	// Scenario 1: Schedule a new tasting session
	fmt.Println("\n--- Scenario 1: Scheduling a new tasting session ---")
	beverageInfo := map[string]string{
		"name":        "Ethiopian Yirgacheffe Single Origin",
		"description": "A bright and floral coffee with notes of lemon and jasmine.",
	}
	locations := []string{"Main Roastery", "Innovation Lab"}

	err = scheduleTastingSession(beverageInfo, locations, config)
	if err != nil {
		log.Printf("Error scheduling tasting session: %v", err)
	}

	// Scenario 2: Adjust schedule due to ingredient shortage
	fmt.Println("\n--- Scenario 2: Adjusting schedule for ingredient shortage ---")
	ingredientShortageInfo := map[string]string{
		"ingredient": "Organic Vanilla Syrup",
		"location":   "All Branches",
		"reason":     "Supplier delay due to logistics issues",
	}
	err = adjustScheduleForIngredientShortage(ingredientShortageInfo, config)
	if err != nil {
		log.Printf("Error adjusting schedule for ingredient shortage: %v", err)
	}

	fmt.Println("\nScheduler Agent demonstration complete.")
}

type TastingSession struct {
	BeverageInfo    map[string]string `json:"beverage_info"`
	Locations       []string          `json:"locations"`
	ScheduledTime   time.Time         `json:"scheduled_time"`
	DurationMinutes int               `json:"duration_minutes"`
}

type ScheduleChange struct {
	Ingredient string `json:"ingredient"`
	Location   string `json:"location"`
	Reason     string `json:"reason"`
	Impact     string `json:"impact"`
	Proposal   string `json:"proposal"`
}

func sendTastingSessionToKafka(session TastingSession, config *Config) error {
	if config.KafkaBrokerAddress == "" || config.KafkaOutputTopicTasting == "" {
		log.Println("Kafka configuration incomplete. Skipping Kafka message.")
		return nil
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{config.KafkaBrokerAddress},
		Topic:    config.KafkaOutputTopicTasting,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting session JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(config.AgentName),
		Value: sessionJSON,
		Time:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to write tasting session message to Kafka: %w", err)
	}

	log.Printf("Tasting session for '%s' sent to Kafka topic '%s'.\n", session.BeverageInfo["name"], config.KafkaOutputTopicTasting)
	return nil
}

func sendScheduleChangeToKafka(change ScheduleChange, config *Config) error {
	if config.KafkaBrokerAddress == "" || config.KafkaOutputTopicScheduleChange == "" {
		log.Println("Kafka configuration incomplete. Skipping Kafka message.")
		return nil
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{config.KafkaBrokerAddress},
		Topic:    config.KafkaOutputTopicScheduleChange,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	changeJSON, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule change JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(config.AgentName),
		Value: changeJSON,
		Time:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to write schedule change message to Kafka: %w", err)
	}

	log.Printf("Schedule change for '%s' sent to Kafka topic '%s'.\n", change.Ingredient, config.KafkaOutputTopicScheduleChange)
	return nil
}

func scheduleTastingSession(beverageInfo map[string]string, locations []string, config *Config) error {
	log.Printf("Attempting to schedule tasting session for beverage '%s' at locations %v\n", beverageInfo["name"], locations)

	// 1. Determine available time (simulated complex logic)
	// In a real scenario, this would query a calendar API for available slots,
	// considering existing events, staff availability, and location capacity.
	// For simulation, we'll find the next available slot after a buffer.
	desiredStartTime := time.Now().Add(time.Duration(config.DefaultRescheduleBufferHours) * time.Hour)
	availableTime, err := findAvailableTime(desiredStartTime, config.DefaultTastingDurationMinutes, locations)
	if err != nil {
		return fmt.Errorf("failed to find available time for tasting session: %w", err)
	}

	session := TastingSession{
		BeverageInfo:    beverageInfo,
		Locations:       locations,
		ScheduledTime:   availableTime,
		DurationMinutes: config.DefaultTastingDurationMinutes,
	}

	// 2. Formulate a detailed event description for the calendar (for logging/internal use)
	eventDescription := fmt.Sprintf(
		"Tasting Session: %s\n\nDescription: %s\n\nLocations: %s\n\nParticipants: Marketing Team, Product Development, Quality Control\n\nNotes: Please provide detailed feedback on aroma, flavor, and mouthfeel. Focus on alignment with target customer profile.",
		beverageInfo["name"],
		beverageInfo["description"],
		fmt.Sprintf("%v", locations),
	)

	log.Printf("Tasting session scheduled for %s at %v for %d minutes.\nDetails:\n%s\n",
		session.ScheduledTime.Format(time.RFC3339), session.Locations, session.DurationMinutes, eventDescription)

	// 3. Send tasting session information to Kafka
	err = sendTastingSessionToKafka(session, config)
	if err != nil {
		return fmt.Errorf("failed to send tasting session to Kafka: %w", err)
	}

	log.Println("Tasting session information sent to Kafka successfully.")
	return nil
}

func adjustScheduleForIngredientShortage(shortageInfo map[string]string, config *Config) error {
	log.Printf("Adjusting schedule due to shortage of '%s' at '%s'. Reason: %s\n",
		shortageInfo["ingredient"], shortageInfo["location"], shortageInfo["reason"])

	// 1. Determine the impact of ingredient shortage on current orders/schedules (simulated complex logic)
	// In a real scenario, this would query inventory, order management, and production schedules
	// to identify affected items, production batches, and customer orders.
	affectedOrders := []string{"Order #12345", "Order #12346"} // Simulated affected orders
	impactDetails := fmt.Sprintf(
		"Shortage of %s at %s due to %s. Affected orders: %v. Estimated delay: 24-48 hours.",
		shortageInfo["ingredient"], shortageInfo["location"], shortageInfo["reason"], affectedOrders,
	)
	log.Println("Impact analysis:", impactDetails)

	// 2. Propose rescheduling or cancellation of events
	// This would involve checking existing schedules and suggesting alternatives.
	// For simulation, we'll propose a reschedule.
	proposedRescheduleTime := time.Now().Add(time.Duration(config.DefaultRescheduleBufferHours*2) * time.Hour) // Propose 2x buffer
	proposal := fmt.Sprintf(
		"Proposal: Reschedule affected production runs and deliveries to %s. Notify customers of potential delays. Consider alternative suppliers if shortage persists.",
		proposedRescheduleTime.Format(time.RFC3339),
	)
	log.Println("Adjustment proposal:", proposal)

	// 3. Send schedule change information to Kafka
	scheduleChange := ScheduleChange{
		Ingredient: shortageInfo["ingredient"],
		Location:   shortageInfo["location"],
		Reason:     shortageInfo["reason"],
		Impact:     impactDetails,
		Proposal:   proposal,
	}

	err := sendScheduleChangeToKafka(scheduleChange, config)
	if err != nil {
		return fmt.Errorf("failed to send schedule change to Kafka: %w", err)
	}

	log.Println("Ingredient shortage adjustment information sent to Kafka successfully.")
	return nil
}

// findAvailableTime simulates finding an available slot in a calendar.
// In a real application, this would involve API calls to a calendar service.
func findAvailableTime(startTime time.Time, durationMinutes int, locations []string) (time.Time, error) {
	// Simulate checking calendar for conflicts.
	// For simplicity, we'll just return the startTime for now, assuming it's available.
	// In a real scenario, you'd iterate through potential slots until an open one is found.
	log.Printf("Simulating finding available time for %d minutes starting from %s at locations %v\n", durationMinutes, startTime.Format(time.RFC3339), locations)
	return startTime, nil
}
