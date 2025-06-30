package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// ChangeType represents the type of configuration change
type ChangeType string

const (
	ConfigLoaded   ChangeType = "loaded"
	ConfigReloaded ChangeType = "reloaded"
	ConfigChanged  ChangeType = "changed"
	ConfigError    ChangeType = "error"
)

// ConfigChange represents a configuration change event
type ConfigChange struct {
	Type      ChangeType
	Config    *Config
	Error     error
	Timestamp time.Time
	Source    string // file, env, etc.
}

// ConfigChangeHandler is a function that handles configuration changes
type ConfigChangeHandler func(change ConfigChange)

// Manager manages configuration loading, validation, and hot reloading
type Manager struct {
	config         *Config
	configPath     string
	envPrefix      string
	loader         *Loader
	validator      *ConfigValidator
	handlers       []ConfigChangeHandler
	watchEnabled   bool
	watchInterval  time.Duration
	lastModTime    time.Time
	mutex          sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewManager creates a new configuration manager
func NewManager(configPath, envPrefix string) *Manager {
	return &Manager{
		configPath:    configPath,
		envPrefix:     envPrefix,
		loader:        NewLoader(configPath, envPrefix),
		validator:     NewConfigValidator(),
		handlers:      make([]ConfigChangeHandler, 0),
		watchEnabled:  false,
		watchInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// Load loads the configuration initially
func (m *Manager) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	config, err := m.loader.Load()
	if err != nil {
		m.notifyHandlers(ConfigChange{
			Type:      ConfigError,
			Error:     err,
			Timestamp: time.Now(),
			Source:    "initial_load",
		})
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := m.validator.Validate(config); err != nil {
		m.notifyHandlers(ConfigChange{
			Type:      ConfigError,
			Error:     err,
			Timestamp: time.Now(),
			Source:    "validation",
		})
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config

	// Update last modification time
	if m.configPath != "" {
		if stat, err := os.Stat(m.configPath); err == nil {
			m.lastModTime = stat.ModTime()
		}
	}

	// Notify handlers
	m.notifyHandlers(ConfigChange{
		Type:      ConfigLoaded,
		Config:    config,
		Timestamp: time.Now(),
		Source:    "file",
	})

	return nil
}

// Get returns the current configuration (thread-safe)
func (m *Manager) Get() *Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.config
}

// Reload reloads the configuration
func (m *Manager) Reload() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	config, err := m.loader.Load()
	if err != nil {
		m.notifyHandlers(ConfigChange{
			Type:      ConfigError,
			Error:     err,
			Timestamp: time.Now(),
			Source:    "reload",
		})
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Validate configuration
	if err := m.validator.Validate(config); err != nil {
		m.notifyHandlers(ConfigChange{
			Type:      ConfigError,
			Error:     err,
			Timestamp: time.Now(),
			Source:    "validation",
		})
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	oldConfig := m.config
	m.config = config

	// Update last modification time
	if m.configPath != "" {
		if stat, err := os.Stat(m.configPath); err == nil {
			m.lastModTime = stat.ModTime()
		}
	}

	// Notify handlers
	m.notifyHandlers(ConfigChange{
		Type:      ConfigReloaded,
		Config:    config,
		Timestamp: time.Now(),
		Source:    "file",
	})

	// Check for significant changes
	if m.hasSignificantChanges(oldConfig, config) {
		m.notifyHandlers(ConfigChange{
			Type:      ConfigChanged,
			Config:    config,
			Timestamp: time.Now(),
			Source:    "comparison",
		})
	}

	return nil
}

// AddChangeHandler adds a configuration change handler
func (m *Manager) AddChangeHandler(handler ConfigChangeHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.handlers = append(m.handlers, handler)
}

// EnableWatch enables configuration file watching
func (m *Manager) EnableWatch(interval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.watchEnabled {
		return // Already watching
	}

	m.watchEnabled = true
	m.watchInterval = interval

	m.wg.Add(1)
	go m.watchLoop()
}

// DisableWatch disables configuration file watching
func (m *Manager) DisableWatch() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.watchEnabled {
		return // Not watching
	}

	m.watchEnabled = false
	close(m.stopChan)
	m.wg.Wait()

	// Recreate stop channel for potential future use
	m.stopChan = make(chan struct{})
}

// watchLoop watches for configuration file changes
func (m *Manager) watchLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.checkForChanges(); err != nil {
				log.Printf("Error checking for config changes: %v", err)
			}
		case <-m.stopChan:
			return
		}
	}
}

// checkForChanges checks if the configuration file has changed
func (m *Manager) checkForChanges() error {
	if m.configPath == "" {
		return nil // No file to watch
	}

	stat, err := os.Stat(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	m.mutex.RLock()
	lastModTime := m.lastModTime
	m.mutex.RUnlock()

	if stat.ModTime().After(lastModTime) {
		log.Printf("Configuration file changed, reloading...")
		if err := m.Reload(); err != nil {
			return fmt.Errorf("failed to reload configuration: %w", err)
		}
	}

	return nil
}

// notifyHandlers notifies all registered change handlers
func (m *Manager) notifyHandlers(change ConfigChange) {
	for _, handler := range m.handlers {
		go func(h ConfigChangeHandler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Configuration change handler panicked: %v", r)
				}
			}()
			h(change)
		}(handler)
	}
}

// hasSignificantChanges checks if there are significant changes between configurations
func (m *Manager) hasSignificantChanges(old, new *Config) bool {
	if old == nil || new == nil {
		return true
	}

	// Check for significant changes that might require service restart
	significantChanges := []bool{
		old.Service.Port != new.Service.Port,
		old.Service.Host != new.Service.Host,
		old.Database.Host != new.Database.Host,
		old.Database.Port != new.Database.Port,
		old.Database.Database != new.Database.Database,
		old.Kafka.GetBrokerList() != new.Kafka.GetBrokerList(),
		old.AI.DefaultProvider != new.AI.DefaultProvider,
		old.Environment != new.Environment,
	}

	for _, changed := range significantChanges {
		if changed {
			return true
		}
	}

	return false
}

// Validate validates the current configuration
func (m *Manager) Validate() error {
	m.mutex.RLock()
	config := m.config
	m.mutex.RUnlock()

	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return m.validator.Validate(config)
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetEnvPrefix returns the environment variable prefix
func (m *Manager) GetEnvPrefix() string {
	return m.envPrefix
}

// IsWatchEnabled returns whether file watching is enabled
func (m *Manager) IsWatchEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.watchEnabled
}

// GetLastModTime returns the last modification time of the config file
func (m *Manager) GetLastModTime() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastModTime
}

// Shutdown gracefully shuts down the configuration manager
func (m *Manager) Shutdown(ctx context.Context) error {
	m.DisableWatch()

	// Wait for all handlers to complete with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ConfigSnapshot represents a point-in-time configuration snapshot
type ConfigSnapshot struct {
	Config    *Config
	Timestamp time.Time
	Source    string
	Hash      string
}

// CreateSnapshot creates a configuration snapshot
func (m *Manager) CreateSnapshot() *ConfigSnapshot {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &ConfigSnapshot{
		Config:    m.config,
		Timestamp: time.Now(),
		Source:    m.configPath,
		Hash:      m.calculateConfigHash(),
	}
}

// calculateConfigHash calculates a hash of the current configuration
func (m *Manager) calculateConfigHash() string {
	// Simple hash calculation - in production, use a proper hash function
	if m.config == nil {
		return ""
	}
	return fmt.Sprintf("%x", time.Now().UnixNano()) // Placeholder
}

// Global configuration manager instance
var globalManager *Manager
var managerOnce sync.Once

// InitGlobalManager initializes the global configuration manager
func InitGlobalManager(configPath, envPrefix string) error {
	var err error
	managerOnce.Do(func() {
		globalManager = NewManager(configPath, envPrefix)
		err = globalManager.Load()
	})
	return err
}

// GetGlobalManager returns the global configuration manager
func GetGlobalManager() *Manager {
	return globalManager
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *Config {
	if globalManager == nil {
		return nil
	}
	return globalManager.Get()
}

// ReloadGlobalConfig reloads the global configuration
func ReloadGlobalConfig() error {
	if globalManager == nil {
		return fmt.Errorf("global configuration manager not initialized")
	}
	return globalManager.Reload()
}

// Convenience functions for common configuration access patterns

// GetServiceConfig returns the service configuration
func GetServiceConfig() ServiceConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.Service
	}
	return ServiceConfig{}
}

// GetDatabaseConfig returns the database configuration
func GetDatabaseConfig() DatabaseConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.Database
	}
	return DatabaseConfig{}
}

// GetKafkaConfig returns the Kafka configuration
func GetKafkaConfig() KafkaConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.Kafka
	}
	return KafkaConfig{}
}

// GetAIConfig returns the AI configuration
func GetAIConfig() AIConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.AI
	}
	return AIConfig{}
}

// GetSecurityConfig returns the security configuration
func GetSecurityConfig() SecurityConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.Security
	}
	return SecurityConfig{}
}

// GetFeatureConfig returns the feature configuration
func GetFeatureConfig() FeatureConfig {
	if config := GetGlobalConfig(); config != nil {
		return config.Features
	}
	return FeatureConfig{}
}

// IsFeatureEnabled checks if a specific feature is enabled
func IsFeatureEnabled(feature string) bool {
	features := GetFeatureConfig()
	switch feature {
	case "ai":
		return features.EnableAI
	case "task_creation":
		return features.EnableTaskCreation
	case "notifications":
		return features.EnableNotifications
	case "metrics":
		return features.EnableMetrics
	case "tracing":
		return features.EnableTracing
	case "audit_logging":
		return features.EnableAuditLogging
	case "caching":
		return features.EnableCaching
	case "rate_limiting":
		return features.EnableRateLimiting
	case "circuit_breaker":
		return features.EnableCircuitBreaker
	case "retry":
		return features.EnableRetry
	case "health_checks":
		return features.EnableHealthChecks
	case "graceful_shutdown":
		return features.EnableGracefulShutdown
	default:
		return false
	}
}
