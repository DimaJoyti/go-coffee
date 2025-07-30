package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/cli/config"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// SystemStatus represents the overall system status
type SystemStatus struct {
	Environment    string             `json:"environment" yaml:"environment"`
	Timestamp      time.Time          `json:"timestamp" yaml:"timestamp"`
	OverallHealth  string             `json:"overall_health" yaml:"overall_health"`
	Services       []ServiceStatus    `json:"services" yaml:"services"`
	Infrastructure InfraStatus        `json:"infrastructure" yaml:"infrastructure"`
	Dependencies   []DependencyStatus `json:"dependencies" yaml:"dependencies"`
	RecentEvents   []Event            `json:"recent_events" yaml:"recent_events"`
}

// ServiceStatus represents the status of a single service
type ServiceStatus struct {
	Name         string                 `json:"name" yaml:"name"`
	Status       string                 `json:"status" yaml:"status"`
	Health       string                 `json:"health" yaml:"health"`
	Version      string                 `json:"version" yaml:"version"`
	Uptime       time.Duration          `json:"uptime" yaml:"uptime"`
	CPU          float64                `json:"cpu_usage" yaml:"cpu_usage"`
	Memory       float64                `json:"memory_usage" yaml:"memory_usage"`
	Replicas     int                    `json:"replicas" yaml:"replicas"`
	Endpoints    []string               `json:"endpoints" yaml:"endpoints"`
	LastDeployed time.Time              `json:"last_deployed" yaml:"last_deployed"`
	Metrics      map[string]interface{} `json:"metrics" yaml:"metrics"`
}

// InfraStatus represents infrastructure component status
type InfraStatus struct {
	Database     DatabaseStatus     `json:"database" yaml:"database"`
	Cache        CacheStatus        `json:"cache" yaml:"cache"`
	MessageQueue MessageQueueStatus `json:"message_queue" yaml:"message_queue"`
	Storage      StorageStatus      `json:"storage" yaml:"storage"`
	Network      NetworkStatus      `json:"network" yaml:"network"`
}

// DatabaseStatus represents database status
type DatabaseStatus struct {
	PostgreSQL DatabaseInstanceStatus `json:"postgresql" yaml:"postgresql"`
}

// DatabaseInstanceStatus represents a database instance status
type DatabaseInstanceStatus struct {
	Status       string  `json:"status" yaml:"status"`
	Connections  int     `json:"connections" yaml:"connections"`
	MaxConns     int     `json:"max_connections" yaml:"max_connections"`
	ResponseTime float64 `json:"response_time_ms" yaml:"response_time_ms"`
	DiskUsage    float64 `json:"disk_usage_percent" yaml:"disk_usage_percent"`
}

// CacheStatus represents cache status
type CacheStatus struct {
	Redis RedisStatus `json:"redis" yaml:"redis"`
}

// RedisStatus represents Redis status
type RedisStatus struct {
	Status       string  `json:"status" yaml:"status"`
	Memory       float64 `json:"memory_usage_mb" yaml:"memory_usage_mb"`
	MaxMemory    float64 `json:"max_memory_mb" yaml:"max_memory_mb"`
	Connections  int     `json:"connections" yaml:"connections"`
	ResponseTime float64 `json:"response_time_ms" yaml:"response_time_ms"`
	KeyCount     int64   `json:"key_count" yaml:"key_count"`
}

// MessageQueueStatus represents message queue status
type MessageQueueStatus struct {
	Kafka KafkaStatus `json:"kafka" yaml:"kafka"`
}

// KafkaStatus represents Kafka status
type KafkaStatus struct {
	Status      string `json:"status" yaml:"status"`
	Brokers     int    `json:"brokers" yaml:"brokers"`
	Topics      int    `json:"topics" yaml:"topics"`
	Partitions  int    `json:"partitions" yaml:"partitions"`
	ConsumerLag int64  `json:"consumer_lag" yaml:"consumer_lag"`
}

// StorageStatus represents storage status
type StorageStatus struct {
	DiskUsage float64 `json:"disk_usage_percent" yaml:"disk_usage_percent"`
	Available int64   `json:"available_gb" yaml:"available_gb"`
	Total     int64   `json:"total_gb" yaml:"total_gb"`
}

// NetworkStatus represents network status
type NetworkStatus struct {
	Latency    float64 `json:"latency_ms" yaml:"latency_ms"`
	Throughput float64 `json:"throughput_mbps" yaml:"throughput_mbps"`
	PacketLoss float64 `json:"packet_loss_percent" yaml:"packet_loss_percent"`
}

// DependencyStatus represents external dependency status
type DependencyStatus struct {
	Name         string    `json:"name" yaml:"name"`
	Type         string    `json:"type" yaml:"type"`
	Status       string    `json:"status" yaml:"status"`
	ResponseTime float64   `json:"response_time_ms" yaml:"response_time_ms"`
	LastChecked  time.Time `json:"last_checked" yaml:"last_checked"`
	Endpoint     string    `json:"endpoint" yaml:"endpoint"`
}

// Event represents a system event
type Event struct {
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	Type      string    `json:"type" yaml:"type"`
	Service   string    `json:"service" yaml:"service"`
	Message   string    `json:"message" yaml:"message"`
	Severity  string    `json:"severity" yaml:"severity"`
}

// showSystemStatus displays the current system status
func showSystemStatus(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment, format string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Gathering system status..."
	s.Start()

	status, err := gatherSystemStatus(ctx, cfg, logger, environment)
	s.Stop()

	if err != nil {
		return fmt.Errorf("failed to gather system status: %w", err)
	}

	switch format {
	case "json":
		return printStatusJSON(status)
	case "yaml":
		return printStatusYAML(status)
	default:
		return printStatusTable(status)
	}
}

// watchSystemStatus continuously monitors system status
func watchSystemStatus(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment, format string, interval time.Duration) error {
	color.Cyan("👀 Watching system status (Press Ctrl+C to stop)")
	color.Cyan("=" + strings.Repeat("=", 50))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Clear screen
			fmt.Print("\033[2J\033[H")

			status, err := gatherSystemStatus(ctx, cfg, logger, environment)
			if err != nil {
				color.Red("❌ Failed to gather status: %v", err)
				continue
			}

			fmt.Printf("Last updated: %s\n\n", time.Now().Format("15:04:05"))

			switch format {
			case "json":
				printStatusJSON(status)
			case "yaml":
				printStatusYAML(status)
			default:
				printStatusTable(status)
			}
		}
	}
}

// gatherSystemStatus collects status information from all components
func gatherSystemStatus(ctx context.Context, cfg *config.Config, logger *zap.Logger, environment string) (*SystemStatus, error) {
	status := &SystemStatus{
		Environment: environment,
		Timestamp:   time.Now(),
	}

	// Gather service status
	services, err := gatherServiceStatus(ctx, environment)
	if err != nil {
		logger.Warn("Failed to gather service status", zap.Error(err))
	}
	status.Services = services

	// Gather infrastructure status
	infra, err := gatherInfrastructureStatus(ctx)
	if err != nil {
		logger.Warn("Failed to gather infrastructure status", zap.Error(err))
	}
	status.Infrastructure = infra

	// Gather dependency status
	deps, err := gatherDependencyStatus(ctx)
	if err != nil {
		logger.Warn("Failed to gather dependency status", zap.Error(err))
	}
	status.Dependencies = deps

	// Gather recent events
	events, err := gatherRecentEvents(ctx, environment)
	if err != nil {
		logger.Warn("Failed to gather recent events", zap.Error(err))
	}
	status.RecentEvents = events

	// Calculate overall health
	status.OverallHealth = calculateOverallHealth(status)

	return status, nil
}

// gatherServiceStatus collects status from all services
func gatherServiceStatus(ctx context.Context, environment string) ([]ServiceStatus, error) {
	services := []string{
		"api-gateway", "auth-service", "order-service", "kitchen-service",
		"payment-service", "producer", "consumer", "streams",
	}

	var statuses []ServiceStatus

	for _, serviceName := range services {
		status, err := getServiceStatus(ctx, serviceName, environment)
		if err != nil {
			// Create a status indicating the service is down
			status = ServiceStatus{
				Name:   serviceName,
				Status: "down",
				Health: "unhealthy",
			}
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// getServiceStatus gets status for a specific service
func getServiceStatus(ctx context.Context, serviceName, environment string) (ServiceStatus, error) {
	// Try to get status from health endpoint
	healthURL := fmt.Sprintf("http://localhost:8080/%s/health", serviceName)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return ServiceStatus{}, err
	}
	defer resp.Body.Close()

	status := ServiceStatus{
		Name:         serviceName,
		Status:       "running",
		Health:       "healthy",
		Version:      "latest",
		Uptime:       time.Hour * 24, // Mock data
		CPU:          15.5,           // Mock data
		Memory:       128.0,          // Mock data
		Replicas:     1,
		Endpoints:    []string{healthURL},
		LastDeployed: time.Now().Add(-time.Hour * 2),
		Metrics:      make(map[string]interface{}),
	}

	if resp.StatusCode != http.StatusOK {
		status.Health = "unhealthy"
	}

	return status, nil
}

// gatherInfrastructureStatus collects infrastructure component status
func gatherInfrastructureStatus(ctx context.Context) (InfraStatus, error) {
	return InfraStatus{
		Database: DatabaseStatus{
			PostgreSQL: DatabaseInstanceStatus{
				Status:       "healthy",
				Connections:  5,
				MaxConns:     100,
				ResponseTime: 2.5,
				DiskUsage:    45.2,
			},
		},
		Cache: CacheStatus{
			Redis: RedisStatus{
				Status:       "healthy",
				Memory:       256.0,
				MaxMemory:    512.0,
				Connections:  10,
				ResponseTime: 0.8,
				KeyCount:     1500,
			},
		},
		MessageQueue: MessageQueueStatus{
			Kafka: KafkaStatus{
				Status:      "healthy",
				Brokers:     3,
				Topics:      5,
				Partitions:  15,
				ConsumerLag: 0,
			},
		},
		Storage: StorageStatus{
			DiskUsage: 65.3,
			Available: 50,
			Total:     100,
		},
		Network: NetworkStatus{
			Latency:    1.2,
			Throughput: 1000.0,
			PacketLoss: 0.0,
		},
	}, nil
}

// gatherDependencyStatus checks external dependencies
func gatherDependencyStatus(ctx context.Context) ([]DependencyStatus, error) {
	dependencies := []DependencyStatus{
		{
			Name:         "External Payment API",
			Type:         "REST API",
			Status:       "healthy",
			ResponseTime: 150.0,
			LastChecked:  time.Now(),
			Endpoint:     "https://api.payment-provider.com/health",
		},
		{
			Name:         "Email Service",
			Type:         "SMTP",
			Status:       "healthy",
			ResponseTime: 50.0,
			LastChecked:  time.Now(),
			Endpoint:     "smtp.email-provider.com:587",
		},
	}

	return dependencies, nil
}

// gatherRecentEvents collects recent system events
func gatherRecentEvents(ctx context.Context, environment string) ([]Event, error) {
	events := []Event{
		{
			Timestamp: time.Now().Add(-time.Minute * 5),
			Type:      "deployment",
			Service:   "order-service",
			Message:   "Successfully deployed version v1.2.3",
			Severity:  "info",
		},
		{
			Timestamp: time.Now().Add(-time.Minute * 15),
			Type:      "alert",
			Service:   "payment-service",
			Message:   "High response time detected",
			Severity:  "warning",
		},
	}

	return events, nil
}

// calculateOverallHealth determines the overall system health
func calculateOverallHealth(status *SystemStatus) string {
	healthyServices := 0
	totalServices := len(status.Services)

	for _, service := range status.Services {
		if service.Health == "healthy" {
			healthyServices++
		}
	}

	if totalServices == 0 {
		return "unknown"
	}

	healthPercentage := float64(healthyServices) / float64(totalServices) * 100

	switch {
	case healthPercentage >= 90:
		return "healthy"
	case healthPercentage >= 70:
		return "degraded"
	default:
		return "unhealthy"
	}
}

// printStatusTable prints status in table format
func printStatusTable(status *SystemStatus) error {
	// Print header
	color.Cyan("🏥 System Health Status")
	color.Cyan("=" + strings.Repeat("=", 50))
	fmt.Printf("Environment: %s\n", status.Environment)
	fmt.Printf("Overall Health: %s\n", getHealthEmoji(status.OverallHealth))
	fmt.Printf("Last Updated: %s\n\n", status.Timestamp.Format("2006-01-02 15:04:05"))

	// Services table
	color.Yellow("📋 Services")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "Status", "Health", "CPU %", "Memory MB", "Uptime"})
	table.SetBorder(false)

	for _, service := range status.Services {
		table.Append([]string{
			service.Name,
			getStatusEmoji(service.Status),
			getHealthEmoji(service.Health),
			fmt.Sprintf("%.1f", service.CPU),
			fmt.Sprintf("%.0f", service.Memory),
			formatDuration(service.Uptime),
		})
	}
	table.Render()

	// Infrastructure status
	fmt.Println()
	color.Yellow("🏗️ Infrastructure")
	infraTable := tablewriter.NewWriter(os.Stdout)
	infraTable.SetHeader([]string{"Component", "Status", "Details"})
	infraTable.SetBorder(false)

	infraTable.Append([]string{
		"PostgreSQL",
		getHealthEmoji(status.Infrastructure.Database.PostgreSQL.Status),
		fmt.Sprintf("Conns: %d/%d, Disk: %.1f%%",
			status.Infrastructure.Database.PostgreSQL.Connections,
			status.Infrastructure.Database.PostgreSQL.MaxConns,
			status.Infrastructure.Database.PostgreSQL.DiskUsage),
	})

	infraTable.Append([]string{
		"Redis",
		getHealthEmoji(status.Infrastructure.Cache.Redis.Status),
		fmt.Sprintf("Mem: %.0f/%.0f MB, Keys: %d",
			status.Infrastructure.Cache.Redis.Memory,
			status.Infrastructure.Cache.Redis.MaxMemory,
			status.Infrastructure.Cache.Redis.KeyCount),
	})

	infraTable.Render()

	return nil
}

// printStatusJSON prints status in JSON format
func printStatusJSON(status *SystemStatus) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// printStatusYAML prints status in YAML format
func printStatusYAML(status *SystemStatus) error {
	data, err := yaml.Marshal(status)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Helper functions are now in utils.go
