package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

type Location struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type Config struct {
	AgentName                          string     `yaml:"agent_name"`
	LogLevel                           string     `yaml:"log_level"`
	InventoryManagerURL                string     `yaml:"inventory_manager_url"`
	SchedulerURL                       string     `yaml:"scheduler_url"`
	TaskManagerURL                     string     `yaml:"task_manager_url"`
	NotifierURL                        string     `yaml:"notifier_url"`
	KafkaBrokerAddress                 string     `yaml:"kafka_broker_address"`
	KafkaInputTopicInventoryUpdate     string     `yaml:"kafka_input_topic_inventory_update"`
	KafkaInputTopicScheduleChange      string     `yaml:"kafka_input_topic_schedule_change"`
	KafkaOutputTopicCoordination       string     `yaml:"kafka_output_topic_coordination"`
	KafkaOutputTopicConflictResolution string     `yaml:"kafka_output_topic_conflict_resolution"`
	Locations                          []Location `yaml:"locations"`
}

// InventoryNeed represents an ingredient or beverage needed by a location.
type InventoryNeed struct {
	LocationID string `json:"location_id"`
	Item       string `json:"item"`
	Quantity   int    `json:"quantity"`
}

// SurplusInventory represents an ingredient or beverage available at a location.
type SurplusInventory struct {
	LocationID string `json:"location_id"`
	Item       string `json:"item"`
	Quantity   int    `json:"quantity"`
}

// DeliveryInstruction represents a detailed instruction for a delivery.
type DeliveryInstruction struct {
	FromLocationID string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	Item           string `json:"item"`
	Quantity       int    `json:"quantity"`
	Route          string `json:"route"` // Simulated optimal route
	ETA            string `json:"eta"`   // Estimated Time of Arrival
}

// ScheduleConflict represents a conflict in scheduling.
type ScheduleConflict struct {
	LocationID  string `json:"location_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// ResourceAllocation represents a proposed solution for resource distribution.
type ResourceAllocation struct {
	ConflictID  string `json:"conflict_id"`
	Description string `json:"description"`
	Action      string `json:"action"`
	AssignedTo  string `json:"assigned_to"`
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
	fmt.Println("Starting Inter-Location Coordinator Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// New flow: Fetch information -> Coordinate/Resolve -> Interact with other agents
	inventoryNeeds, err := getInventoryNeeds(config.InventoryManagerURL)
	if err != nil {
		log.Printf("Error fetching inventory needs: %v", err)
	} else {
		log.Printf("Received inventory needs: %+v", inventoryNeeds)
		coordinateDeliveries(config, inventoryNeeds)
	}

	scheduleConflicts, err := getScheduleConflicts(config.SchedulerURL)
	if err != nil {
		log.Printf("Error fetching schedule conflicts: %v", err)
	} else {
		log.Printf("Received schedule conflicts: %+v", scheduleConflicts)
		resolveInterLocationConflicts(config, scheduleConflicts)
	}

	fmt.Println("Inter-Location Coordinator Agent started successfully.")
}

// getInventoryNeeds fetches inventory needs from the Inventory Manager Agent.
func getInventoryNeeds(url string) ([]InventoryNeed, error) {
	log.Printf("Fetching inventory needs from %s...", url)
	resp, err := http.Get(url + "/inventory-needs") // Assuming an endpoint for inventory needs
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory needs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch inventory needs, status: %s", resp.Status)
	}

	var needs []InventoryNeed
	if err := json.NewDecoder(resp.Body).Decode(&needs); err != nil {
		return nil, fmt.Errorf("failed to decode inventory needs: %w", err)
	}
	return needs, nil
}

// getSurplusInventory fetches surplus inventory from the Inventory Manager Agent.
func getSurplusInventory(url string) ([]SurplusInventory, error) {
	log.Printf("Fetching surplus inventory from %s...", url)
	resp, err := http.Get(url + "/surplus-inventory") // Assuming an endpoint for surplus inventory
	if err != nil {
		return nil, fmt.Errorf("failed to fetch surplus inventory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch surplus inventory, status: %s", resp.Status)
	}

	var surplus []SurplusInventory
	if err := json.NewDecoder(resp.Body).Decode(&surplus); err != nil {
		return nil, fmt.Errorf("failed to decode surplus inventory: %w", err)
	}
	return surplus, nil
}

// getScheduleConflicts fetches schedule conflicts from the Scheduler Agent.
func getScheduleConflicts(url string) ([]ScheduleConflict, error) {
	log.Printf("Fetching schedule conflicts from %s...", url)
	resp, err := http.Get(url + "/schedule-conflicts") // Assuming an endpoint for schedule conflicts
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule conflicts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch schedule conflicts, status: %s", resp.Status)
	}

	var conflicts []ScheduleConflict
	if err := json.NewDecoder(resp.Body).Decode(&conflicts); err != nil {
		return nil, fmt.Errorf("failed to decode schedule conflicts: %w", err)
	}
	return conflicts, nil
}

// coordinateDeliveries implements more complex logic for coordinating deliveries.
func coordinateDeliveries(config *Config, inventoryNeeds []InventoryNeed) {
	log.Printf("Coordinating deliveries for agent: %s", config.AgentName)

	surplusInventory, err := getSurplusInventory(config.InventoryManagerURL)
	if err != nil {
		log.Printf("Error fetching surplus inventory: %v", err)
		return
	}
	log.Printf("Received surplus inventory: %+v", surplusInventory)

	// Convert surplus inventory to a more accessible map for quick lookups
	surplusMap := make(map[string]map[string]int)
	for _, s := range surplusInventory {
		if _, ok := surplusMap[s.LocationID]; !ok {
			surplusMap[s.LocationID] = make(map[string]int)
		}
		surplusMap[s.LocationID][s.Item] += s.Quantity
	}

	for _, need := range inventoryNeeds {
		log.Printf("Processing need: %+v", need)
		foundSource := false
		for _, surplusLoc := range config.Locations {
			if need.LocationID == surplusLoc.ID {
				continue // Don't transfer to self
			}

			if availableQty, ok := surplusMap[surplusLoc.ID][need.Item]; ok && availableQty >= need.Quantity {
				log.Printf("Coordinating transfer: %d units of %s from %s to %s", need.Quantity, need.Item, surplusLoc.ID, need.LocationID)

				// Simulate optimal route determination and ETA
				route := fmt.Sprintf("Route from %s to %s via main highway", surplusLoc.Name, getLocationName(config, need.LocationID))
				eta := "2 hours" // Simulated ETA

				deliveryInstruction := DeliveryInstruction{
					FromLocationID: surplusLoc.ID,
					ToLocationID:   need.LocationID,
					Item:           need.Item,
					Quantity:       need.Quantity,
					Route:          route,
					ETA:            eta,
				}

				taskDescription := fmt.Sprintf("Coordinate delivery of %d %s from %s to %s. Route: %s, ETA: %s",
					deliveryInstruction.Quantity, deliveryInstruction.Item, deliveryInstruction.FromLocationID,
					deliveryInstruction.ToLocationID, deliveryInstruction.Route, deliveryInstruction.ETA)
				sendTask(config.TaskManagerURL, taskDescription)

				notificationMsg := fmt.Sprintf("New delivery coordinated: %d %s from %s to %s. ETA: %s",
					deliveryInstruction.Quantity, deliveryInstruction.Item, deliveryInstruction.FromLocationID,
					deliveryInstruction.ToLocationID, deliveryInstruction.ETA)
				sendNotification(config.NotifierURL, notificationMsg)

				// Send coordination message to Kafka
				coordinationMsg := fmt.Sprintf("Delivery coordination: %d %s from %s to %s, Route: %s, ETA: %s",
					deliveryInstruction.Quantity, deliveryInstruction.Item, deliveryInstruction.FromLocationID,
					deliveryInstruction.ToLocationID, deliveryInstruction.Route, deliveryInstruction.ETA)
				if err := sendCoordinationMessageToKafka(config, coordinationMsg); err != nil {
					log.Printf("Error sending coordination message to Kafka: %v", err)
				}

				foundSource = true
				break // Move to next need after finding a source
			}
		}
		if !foundSource {
			log.Printf("No immediate surplus found for %d units of %s at %s. Notifying for procurement.", need.Quantity, need.Item, need.LocationID)
			sendNotification(config.NotifierURL, fmt.Sprintf("Urgent: %d units of %s needed at %s. No immediate surplus found.", need.Quantity, need.Item, need.LocationID))
			sendTask(config.TaskManagerURL, fmt.Sprintf("Procure %d %s for %s", need.Quantity, need.Item, need.LocationID))
		}
	}
	log.Println("Delivery coordination complete.")
}

// resolveInterLocationConflicts implements more complex logic for resolving conflicts.
func resolveInterLocationConflicts(config *Config, scheduleConflicts []ScheduleConflict) {
	log.Printf("Resolving inter-location conflicts for agent: %s", config.AgentName)

	for _, conflict := range scheduleConflicts {
		log.Printf("Analyzing conflict: %+v", conflict)

		// More sophisticated conflict analysis
		resolutionProposal := ""
		taskDescription := ""
		notificationMsg := ""

		switch conflict.Description {
		case "barista shortage":
			resolutionProposal = fmt.Sprintf("Proposing to reallocate a barista from another location or hire temporary staff for %s.", conflict.LocationID)
			taskDescription = fmt.Sprintf("Resolve barista shortage at %s: %s", conflict.LocationID, resolutionProposal)
			notificationMsg = fmt.Sprintf("Conflict resolved (proposed): Barista shortage at %s. %s", conflict.LocationID, resolutionProposal)
		case "delivery truck double-booked":
			resolutionProposal = fmt.Sprintf("Proposing to reschedule one delivery or assign an alternative vehicle for %s.", conflict.LocationID)
			taskDescription = fmt.Sprintf("Resolve delivery truck double-booking at %s: %s", conflict.LocationID, resolutionProposal)
			notificationMsg = fmt.Sprintf("Conflict resolved (proposed): Delivery truck double-booked at %s. %s", conflict.LocationID, resolutionProposal)
		default:
			resolutionProposal = fmt.Sprintf("Generic conflict resolution for %s: Investigate and propose solution.", conflict.Description)
			taskDescription = fmt.Sprintf("Investigate and resolve conflict: %s at %s", conflict.Description, conflict.LocationID)
			notificationMsg = fmt.Sprintf("Conflict detected: %s at %s. Investigation initiated.", conflict.Description, conflict.LocationID)
		}

		sendNotification(config.NotifierURL, notificationMsg)
		sendTask(config.TaskManagerURL, taskDescription)

		// Send conflict resolution message to Kafka
		conflictResolutionMsg := fmt.Sprintf("Conflict Resolution: %s at %s - %s",
			conflict.Description, conflict.LocationID, resolutionProposal)
		if err := sendConflictResolutionToKafka(config, conflictResolutionMsg); err != nil {
			log.Printf("Error sending conflict resolution message to Kafka: %v", err)
		}
	}
	log.Println("Conflict resolution complete.")
}

// getLocationName is a helper function to get location name by ID.
func getLocationName(config *Config, locationID string) string {
	for _, loc := range config.Locations {
		if loc.ID == locationID {
			return loc.Name
		}
	}
	return "Unknown Location"
}

// sendTask sends a task to the Task Manager Agent.
func sendTask(url, taskDescription string) {
	log.Printf("Sending task to Task Manager: %s", taskDescription)
	payload := map[string]string{"task": taskDescription}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error sending task to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Task sent to %s, status: %s", url, resp.Status)
}

// sendNotification sends a notification to the Notifier Agent.
func sendNotification(url, message string) {
	log.Printf("Sending notification to Notifier: %s", message)
	payload := map[string]string{"message": message}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error sending notification to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Notification sent to %s, status: %s", url, resp.Status)
}

// sendCoordinationMessageToKafka sends coordination messages to Kafka
func sendCoordinationMessageToKafka(config *Config, message string) error {
	if config.KafkaBrokerAddress == "" || config.KafkaOutputTopicCoordination == "" {
		log.Println("Kafka configuration incomplete. Skipping Kafka message.")
		return nil
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.KafkaBrokerAddress),
		Topic:    config.KafkaOutputTopicCoordination,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   []byte(config.AgentName),
		Value: []byte(message),
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write coordination message to Kafka: %w", err)
	}

	log.Printf("Coordination message sent to Kafka topic %s successfully.", config.KafkaOutputTopicCoordination)
	return nil
}

// sendConflictResolutionToKafka sends conflict resolution messages to Kafka
func sendConflictResolutionToKafka(config *Config, message string) error {
	if config.KafkaBrokerAddress == "" || config.KafkaOutputTopicConflictResolution == "" {
		log.Println("Kafka configuration incomplete. Skipping Kafka message.")
		return nil
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.KafkaBrokerAddress),
		Topic:    config.KafkaOutputTopicConflictResolution,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   []byte(config.AgentName),
		Value: []byte(message),
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write conflict resolution message to Kafka: %w", err)
	}

	log.Printf("Conflict resolution message sent to Kafka topic %s successfully.", config.KafkaOutputTopicConflictResolution)
	return nil
}
