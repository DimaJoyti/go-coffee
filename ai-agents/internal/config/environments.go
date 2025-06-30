package config

import (
	"time"

	"go-coffee-ai-agents/internal/observability"
	"go-coffee-ai-agents/internal/resilience"
)

// GetDevelopmentConfig returns a development environment configuration
func GetDevelopmentConfig() *Config {
	return &Config{
		Environment: Development,
		Service: ServiceConfig{
			Name:        "beverage-inventor-agent",
			Version:     "1.0.0",
			Description: "AI-powered beverage invention agent",
			Port:        8080,
			Host:        "localhost",
			BasePath:    "/api/v1",
			Debug:       true,
		},
		Server: ServerConfig{
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1MB
			EnableCORS:      true,
			CORSOrigins:     []string{"*"},
			EnableMetrics:   true,
			MetricsPath:     "/metrics",
			HealthPath:      "/health",
		},
		Database: DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "gocoffee_dev",
			Username:        "gocoffee",
			Password:        "password",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			MigrationsPath:  "./migrations",
			EnableLogging:   true,
		},
		Kafka: KafkaConfig{
			Brokers:        []string{"localhost:9092"},
			ClientID:       "beverage-inventor-dev",
			GroupID:        "beverage-inventor-group-dev",
			ConnectTimeout: 10 * time.Second,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			BatchSize:      100,
			BatchTimeout:   100 * time.Millisecond,
			RetryMax:       3,
			RetryBackoff:   100 * time.Millisecond,
			Topics: TopicsConfig{
				BeverageCreated:    "beverage.created.dev",
				BeverageUpdated:    "beverage.updated.dev",
				TaskCreated:        "task.created.dev",
				TaskUpdated:        "task.updated.dev",
				NotificationSent:   "notification.sent.dev",
				AIRequestCompleted: "ai.request.completed.dev",
				SystemEvent:        "system.event.dev",
			},
		},
		AI: AIConfig{
			DefaultProvider: "gemini",
			Providers: map[string]AIProviderConfig{
				"gemini": {
					Enabled:     true,
					APIKey:      "dev-api-key",
					BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
					Model:       "gemini-pro",
					MaxTokens:   2048,
					Temperature: 0.7,
					Timeout:     30 * time.Second,
				},
				"openai": {
					Enabled:     false,
					APIKey:      "dev-api-key",
					BaseURL:     "https://api.openai.com/v1",
					Model:       "gpt-3.5-turbo",
					MaxTokens:   2048,
					Temperature: 0.7,
					Timeout:     30 * time.Second,
				},
			},
			RateLimits: AIRateLimitsConfig{
				RequestsPerMinute: 100,
				TokensPerMinute:   10000,
				BurstSize:         20,
				CooldownPeriod:    1 * time.Minute,
			},
			Timeouts: AITimeoutsConfig{
				AnalyzeIngredients:  15 * time.Second,
				GenerateDescription: 20 * time.Second,
				SuggestImprovements: 25 * time.Second,
				GenerateRecipe:      30 * time.Second,
			},
		},
		External: ExternalConfig{
			ClickUp: ClickUpConfig{
				Enabled:    false, // Disabled in development
				APIKey:     "dev-clickup-key",
				BaseURL:    "https://api.clickup.com/api/v2",
				Timeout:    15 * time.Second,
				RetryCount: 3,
				RateLimit:  100,
			},
			Slack: SlackConfig{
				Enabled:        false, // Disabled in development
				BotToken:       "dev-slack-bot-token",
				DefaultChannel: "#dev-notifications",
				Timeout:        10 * time.Second,
				RetryCount:     3,
			},
			GoogleSheets: GoogleSheetsConfig{
				Enabled:          false, // Disabled in development
				CredentialsPath:  "./credentials/google-sheets-dev.json",
				DefaultSheetName: "Development Data",
				Timeout:          15 * time.Second,
				RetryCount:       3,
			},
			Email: EmailConfig{
				Enabled:   false, // Disabled in development
				Provider:  "smtp",
				SMTPHost:  "localhost",
				SMTPPort:  1025, // MailHog for development
				FromEmail: "dev@gocoffee.local",
				FromName:  "Go Coffee Dev",
				EnableTLS: false,
			},
		},
		Security: SecurityConfig{
			JWT: JWTConfig{
				SecretKey:      "dev-secret-key-change-in-production",
				Issuer:         "go-coffee-dev",
				Audience:       "go-coffee-users",
				ExpirationTime: 24 * time.Hour,
				RefreshTime:    7 * 24 * time.Hour,
				Algorithm:      "HS256",
			},
			API: APISecurityConfig{
				EnableAPIKeys:   false,
				RequireHTTPS:    false,
				EnableRateLimit: false,
				MaxRequestSize:  10 << 20, // 10MB
			},
			CORS: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           86400,
			},
			RateLimit: RateLimitConfig{
				Enabled:        false,
				RequestsPerMin: 1000,
				BurstSize:      100,
				CleanupPeriod:  1 * time.Minute,
				Storage:        "memory",
			},
		},
		Observability: observability.DefaultTelemetryConfig("beverage-inventor-agent"),
		Resilience:    resilience.DefaultResilienceConfig(),
		Features: FeatureConfig{
			EnableAI:               true,
			EnableTaskCreation:     false, // Disabled in development
			EnableNotifications:    false, // Disabled in development
			EnableMetrics:          true,
			EnableTracing:          true,
			EnableAuditLogging:     true,
			EnableCaching:          false,
			EnableRateLimiting:     false,
			EnableCircuitBreaker:   true,
			EnableRetry:            true,
			EnableHealthChecks:     true,
			EnableGracefulShutdown: true,
		},
	}
}

// GetStagingConfig returns a staging environment configuration
func GetStagingConfig() *Config {
	config := GetDevelopmentConfig()
	
	// Override staging-specific settings
	config.Environment = Staging
	config.Service.Debug = false
	config.Service.Host = "0.0.0.0"
	
	// Database
	config.Database.Database = "gocoffee_staging"
	config.Database.EnableLogging = false
	config.Database.MaxOpenConns = 20
	
	// Kafka
	config.Kafka.Topics = TopicsConfig{
		BeverageCreated:    "beverage.created.staging",
		BeverageUpdated:    "beverage.updated.staging",
		TaskCreated:        "task.created.staging",
		TaskUpdated:        "task.updated.staging",
		NotificationSent:   "notification.sent.staging",
		AIRequestCompleted: "ai.request.completed.staging",
		SystemEvent:        "system.event.staging",
	}
	
	// AI rate limits (more restrictive)
	config.AI.RateLimits.RequestsPerMinute = 60
	config.AI.RateLimits.TokensPerMinute = 5000
	
	// Enable external services for testing
	config.External.ClickUp.Enabled = true
	config.External.Slack.Enabled = true
	config.External.Slack.DefaultChannel = "#staging-notifications"
	
	// Security
	config.Security.API.RequireHTTPS = true
	config.Security.API.EnableRateLimit = true
	config.Security.RateLimit.Enabled = true
	config.Security.RateLimit.RequestsPerMin = 500
	
	// Observability (reduced sampling)
	config.Observability = observability.StagingTelemetryConfig("beverage-inventor-agent")
	
	// Features
	config.Features.EnableTaskCreation = true
	config.Features.EnableNotifications = true
	config.Features.EnableRateLimiting = true
	
	return config
}

// GetProductionConfig returns a production environment configuration
func GetProductionConfig() *Config {
	config := GetStagingConfig()
	
	// Override production-specific settings
	config.Environment = Production
	config.Service.Debug = false
	
	// Database
	config.Database.Database = "gocoffee_production"
	config.Database.SSLMode = "require"
	config.Database.MaxOpenConns = 50
	config.Database.MaxIdleConns = 10
	
	// Kafka
	config.Kafka.Topics = TopicsConfig{
		BeverageCreated:    "beverage.created",
		BeverageUpdated:    "beverage.updated",
		TaskCreated:        "task.created",
		TaskUpdated:        "task.updated",
		NotificationSent:   "notification.sent",
		AIRequestCompleted: "ai.request.completed",
		SystemEvent:        "system.event",
	}
	config.Kafka.EnableSASL = true
	config.Kafka.EnableTLS = true
	
	// AI rate limits (production limits)
	config.AI.RateLimits.RequestsPerMinute = 30
	config.AI.RateLimits.TokensPerMinute = 2000
	config.AI.RateLimits.BurstSize = 5
	
	// Security (production hardening)
	config.Security.JWT.SecretKey = "" // Must be set via environment
	config.Security.API.EnableAPIKeys = true
	config.Security.API.RequireHTTPS = true
	config.Security.API.EnableRateLimit = true
	config.Security.RateLimit.RequestsPerMin = 100
	config.Security.RateLimit.Storage = "redis"
	config.Security.CORS.AllowedOrigins = []string{
		"https://gocoffee.com",
		"https://app.gocoffee.com",
	}
	
	// Observability (production settings)
	config.Observability = observability.ProductionTelemetryConfig("beverage-inventor-agent")
	
	// All features enabled in production
	config.Features.EnableTaskCreation = true
	config.Features.EnableNotifications = true
	config.Features.EnableCaching = true
	config.Features.EnableRateLimiting = true
	
	return config
}

// GetTestingConfig returns a testing environment configuration
func GetTestingConfig() *Config {
	config := GetDevelopmentConfig()
	
	// Override testing-specific settings
	config.Environment = Testing
	config.Service.Port = 0 // Random port for testing
	
	// Database (in-memory or test database)
	config.Database.Database = "gocoffee_test"
	config.Database.MaxOpenConns = 5
	config.Database.MaxIdleConns = 2
	
	// Kafka (use test topics)
	config.Kafka.Topics = TopicsConfig{
		BeverageCreated:    "beverage.created.test",
		BeverageUpdated:    "beverage.updated.test",
		TaskCreated:        "task.created.test",
		TaskUpdated:        "task.updated.test",
		NotificationSent:   "notification.sent.test",
		AIRequestCompleted: "ai.request.completed.test",
		SystemEvent:        "system.event.test",
	}
	
	// Disable external services in tests
	config.External.ClickUp.Enabled = false
	config.External.Slack.Enabled = false
	config.External.GoogleSheets.Enabled = false
	config.External.Email.Enabled = false
	
	// Disable some features for faster tests
	config.Features.EnableMetrics = false
	config.Features.EnableTracing = false
	config.Features.EnableAuditLogging = false
	
	return config
}

// GetConfigForEnvironment returns configuration for a specific environment
func GetConfigForEnvironment(env Environment) *Config {
	switch env {
	case Development:
		return GetDevelopmentConfig()
	case Staging:
		return GetStagingConfig()
	case Production:
		return GetProductionConfig()
	case Testing:
		return GetTestingConfig()
	default:
		return GetDevelopmentConfig()
	}
}

// GetConfigForEnvironmentString returns configuration for an environment string
func GetConfigForEnvironmentString(env string) *Config {
	switch Environment(env) {
	case Development:
		return GetDevelopmentConfig()
	case Staging:
		return GetStagingConfig()
	case Production:
		return GetProductionConfig()
	case Testing:
		return GetTestingConfig()
	default:
		return GetDevelopmentConfig()
	}
}
