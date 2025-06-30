package external

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/external/auth"
	"go-coffee-ai-agents/internal/external/clickup"
	"go-coffee-ai-agents/internal/external/interfaces"
	"go-coffee-ai-agents/internal/observability"
)

// Manager manages all external system integrations
type Manager struct {
	// Authentication
	authManager *auth.AuthManager
	
	// Service providers
	taskManagers     map[string]interfaces.TaskManager
	messagingProviders map[string]interfaces.MessagingProvider
	spreadsheetProviders map[string]interfaces.SpreadsheetProvider
	socialMediaProviders map[string]interfaces.SocialMediaProvider
	
	// Configuration
	config *Config
	
	// Observability
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
	
	// Thread safety
	mutex sync.RWMutex
}

// Config holds external systems configuration
type Config struct {
	// Authentication configuration
	Auth *auth.Config `yaml:"auth" json:"auth"`
	
	// Service configurations
	ClickUp     *clickup.Config     `yaml:"clickup" json:"clickup"`
	Slack       *SlackConfig        `yaml:"slack" json:"slack"`
	GoogleSheets *GoogleSheetsConfig `yaml:"google_sheets" json:"google_sheets"`
	Airtable    *AirtableConfig     `yaml:"airtable" json:"airtable"`
	Twitter     *TwitterConfig      `yaml:"twitter" json:"twitter"`
	Instagram   *InstagramConfig    `yaml:"instagram" json:"instagram"`
	Facebook    *FacebookConfig     `yaml:"facebook" json:"facebook"`
	
	// Global settings
	DefaultTimeout    time.Duration `yaml:"default_timeout" json:"default_timeout"`
	RetryAttempts     int           `yaml:"retry_attempts" json:"retry_attempts"`
	EnableMetrics     bool          `yaml:"enable_metrics" json:"enable_metrics"`
	EnableTracing     bool          `yaml:"enable_tracing" json:"enable_tracing"`
	
	// Rate limiting
	GlobalRateLimit   int           `yaml:"global_rate_limit" json:"global_rate_limit"`
	RateLimitWindow   time.Duration `yaml:"rate_limit_window" json:"rate_limit_window"`
}

// Service-specific configurations
type SlackConfig struct {
	BotToken    string `yaml:"bot_token" json:"bot_token"`
	AppToken    string `yaml:"app_token" json:"app_token"`
	SigningSecret string `yaml:"signing_secret" json:"signing_secret"`
	WebhookURL  string `yaml:"webhook_url" json:"webhook_url"`
}

type GoogleSheetsConfig struct {
	ServiceAccountKey string   `yaml:"service_account_key" json:"service_account_key"`
	Scopes           []string `yaml:"scopes" json:"scopes"`
}

type AirtableConfig struct {
	APIKey  string `yaml:"api_key" json:"api_key"`
	BaseID  string `yaml:"base_id" json:"base_id"`
}

type TwitterConfig struct {
	APIKey       string `yaml:"api_key" json:"api_key"`
	APISecret    string `yaml:"api_secret" json:"api_secret"`
	AccessToken  string `yaml:"access_token" json:"access_token"`
	AccessSecret string `yaml:"access_secret" json:"access_secret"`
	BearerToken  string `yaml:"bearer_token" json:"bearer_token"`
}

type InstagramConfig struct {
	AccessToken string `yaml:"access_token" json:"access_token"`
	AppID       string `yaml:"app_id" json:"app_id"`
	AppSecret   string `yaml:"app_secret" json:"app_secret"`
}

type FacebookConfig struct {
	AppID       string `yaml:"app_id" json:"app_id"`
	AppSecret   string `yaml:"app_secret" json:"app_secret"`
	AccessToken string `yaml:"access_token" json:"access_token"`
	PageID      string `yaml:"page_id" json:"page_id"`
}

// NewManager creates a new external systems manager
func NewManager(
	config *Config,
	authManager *auth.AuthManager,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Manager {
	return &Manager{
		authManager:          authManager,
		taskManagers:         make(map[string]interfaces.TaskManager),
		messagingProviders:   make(map[string]interfaces.MessagingProvider),
		spreadsheetProviders: make(map[string]interfaces.SpreadsheetProvider),
		socialMediaProviders: make(map[string]interfaces.SocialMediaProvider),
		config:               config,
		logger:               logger,
		metrics:              metrics,
		tracing:              tracing,
	}
}

// Initialize initializes all configured external systems
func (m *Manager) Initialize(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Initialize ClickUp if configured
	if m.config.ClickUp != nil && m.config.ClickUp.APIKey != "" {
		client := clickup.NewClient(m.config.ClickUp)
		m.taskManagers["clickup"] = client
		m.logger.Info("Initialized ClickUp integration")
	}
	
	// TODO: Initialize other services
	// - Slack
	// - Google Sheets
	// - Airtable
	// - Social media platforms
	
	m.logger.Info("External systems manager initialized",
		"task_managers", len(m.taskManagers),
		"messaging_providers", len(m.messagingProviders),
		"spreadsheet_providers", len(m.spreadsheetProviders),
		"social_media_providers", len(m.socialMediaProviders),
	)
	
	return nil
}

// GetTaskManager returns a task manager by name
func (m *Manager) GetTaskManager(name string) (interfaces.TaskManager, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	manager, exists := m.taskManagers[name]
	if !exists {
		return nil, fmt.Errorf("task manager %s not found", name)
	}
	
	return manager, nil
}

// GetMessagingProvider returns a messaging provider by name
func (m *Manager) GetMessagingProvider(name string) (interfaces.MessagingProvider, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	provider, exists := m.messagingProviders[name]
	if !exists {
		return nil, fmt.Errorf("messaging provider %s not found", name)
	}
	
	return provider, nil
}

// GetSpreadsheetProvider returns a spreadsheet provider by name
func (m *Manager) GetSpreadsheetProvider(name string) (interfaces.SpreadsheetProvider, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	provider, exists := m.spreadsheetProviders[name]
	if !exists {
		return nil, fmt.Errorf("spreadsheet provider %s not found", name)
	}
	
	return provider, nil
}

// GetSocialMediaProvider returns a social media provider by name
func (m *Manager) GetSocialMediaProvider(name string) (interfaces.SocialMediaProvider, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	provider, exists := m.socialMediaProviders[name]
	if !exists {
		return nil, fmt.Errorf("social media provider %s not found", name)
	}
	
	return provider, nil
}

// ListProviders returns information about all available providers
func (m *Manager) ListProviders() *ProvidersInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	info := &ProvidersInfo{
		TaskManagers:         make(map[string]*interfaces.ProviderInfo),
		MessagingProviders:   make(map[string]*interfaces.ProviderInfo),
		SpreadsheetProviders: make(map[string]*interfaces.ProviderInfo),
		SocialMediaProviders: make(map[string]*interfaces.ProviderInfo),
	}
	
	// Collect task manager info
	for name, manager := range m.taskManagers {
		info.TaskManagers[name] = manager.GetProviderInfo()
	}
	
	// Collect messaging provider info
	for name, provider := range m.messagingProviders {
		info.MessagingProviders[name] = provider.GetProviderInfo()
	}
	
	// Collect spreadsheet provider info
	for name, provider := range m.spreadsheetProviders {
		info.SpreadsheetProviders[name] = provider.GetProviderInfo()
	}
	
	// Collect social media provider info
	for name, provider := range m.socialMediaProviders {
		info.SocialMediaProviders[name] = provider.GetProviderInfo()
	}
	
	return info
}

// ProvidersInfo contains information about all providers
type ProvidersInfo struct {
	TaskManagers         map[string]*interfaces.ProviderInfo `json:"task_managers"`
	MessagingProviders   map[string]*interfaces.ProviderInfo `json:"messaging_providers"`
	SpreadsheetProviders map[string]*interfaces.ProviderInfo `json:"spreadsheet_providers"`
	SocialMediaProviders map[string]*interfaces.ProviderInfo `json:"social_media_providers"`
}

// Health check methods
func (m *Manager) HealthCheck(ctx context.Context) *HealthStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	status := &HealthStatus{
		Overall:              "healthy",
		TaskManagers:         make(map[string]string),
		MessagingProviders:   make(map[string]string),
		SpreadsheetProviders: make(map[string]string),
		SocialMediaProviders: make(map[string]string),
		CheckedAt:            time.Now(),
	}
	
	// Check task managers
	for name := range m.taskManagers {
		// TODO: Implement actual health checks
		status.TaskManagers[name] = "healthy"
	}
	
	// Check messaging providers
	for name := range m.messagingProviders {
		status.MessagingProviders[name] = "healthy"
	}
	
	// Check spreadsheet providers
	for name := range m.spreadsheetProviders {
		status.SpreadsheetProviders[name] = "healthy"
	}
	
	// Check social media providers
	for name := range m.socialMediaProviders {
		status.SocialMediaProviders[name] = "healthy"
	}
	
	return status
}

// HealthStatus represents the health status of all external systems
type HealthStatus struct {
	Overall              string            `json:"overall"`
	TaskManagers         map[string]string `json:"task_managers"`
	MessagingProviders   map[string]string `json:"messaging_providers"`
	SpreadsheetProviders map[string]string `json:"spreadsheet_providers"`
	SocialMediaProviders map[string]string `json:"social_media_providers"`
	CheckedAt            time.Time         `json:"checked_at"`
}

// Shutdown gracefully shuts down all external system connections
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.logger.Info("Shutting down external systems manager")
	
	// TODO: Implement graceful shutdown for each provider type
	// This would include:
	// - Closing HTTP connections
	// - Stopping webhook listeners
	// - Cleaning up resources
	
	m.logger.Info("External systems manager shutdown complete")
	return nil
}

// Metrics and monitoring
func (m *Manager) GetMetrics() *ExternalSystemsMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return &ExternalSystemsMetrics{
		TaskManagerCount:         len(m.taskManagers),
		MessagingProviderCount:   len(m.messagingProviders),
		SpreadsheetProviderCount: len(m.spreadsheetProviders),
		SocialMediaProviderCount: len(m.socialMediaProviders),
		LastHealthCheck:          time.Now(),
	}
}

// ExternalSystemsMetrics contains metrics about external systems
type ExternalSystemsMetrics struct {
	TaskManagerCount         int       `json:"task_manager_count"`
	MessagingProviderCount   int       `json:"messaging_provider_count"`
	SpreadsheetProviderCount int       `json:"spreadsheet_provider_count"`
	SocialMediaProviderCount int       `json:"social_media_provider_count"`
	LastHealthCheck          time.Time `json:"last_health_check"`
}

// Helper methods for common operations
func (m *Manager) CreateTaskInClickUp(ctx context.Context, req *interfaces.CreateTaskRequest) (*interfaces.Task, error) {
	clickup, err := m.GetTaskManager("clickup")
	if err != nil {
		return nil, fmt.Errorf("failed to get ClickUp client: %w", err)
	}
	
	return clickup.CreateTask(ctx, req)
}

func (m *Manager) SendSlackMessage(ctx context.Context, req *interfaces.SendMessageRequest) (*interfaces.Message, error) {
	slack, err := m.GetMessagingProvider("slack")
	if err != nil {
		return nil, fmt.Errorf("failed to get Slack client: %w", err)
	}
	
	return slack.SendMessage(ctx, req)
}

// TODO: Add more helper methods for common operations across different providers
