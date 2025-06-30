# Configuration Management Package

This package provides comprehensive configuration management for the Go Coffee AI Agents, supporting multiple environments, hot reloading, validation, and environment variable overrides.

## Overview

The configuration system implements a layered approach:

1. **Default Values**: Sensible defaults for all configuration options
2. **YAML Files**: Environment-specific configuration files
3. **Environment Variables**: Runtime overrides with prefix support
4. **Validation**: Comprehensive validation with detailed error messages
5. **Hot Reloading**: Automatic configuration reloading on file changes

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Configuration   │    │ Loader          │    │ Validator       │
│ Manager         │───▶│ (YAML + Env)    │    │ (Rules Engine)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Hot Reloading   │    │ Environment     │    │ Change          │
│ & File Watching │    │ Templates       │    │ Notifications   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. Core Configuration (`config.go`)

Defines the complete configuration structure:

```go
type Config struct {
    Service       ServiceConfig       // Service metadata and server settings
    Database      DatabaseConfig      // Database connection configuration
    Kafka         KafkaConfig         // Message broker configuration
    AI            AIConfig            // AI provider configurations
    External      ExternalConfig      // External service integrations
    Security      SecurityConfig      // Security and authentication
    Observability TelemetryConfig     // Monitoring and observability
    Resilience    ResilienceConfig    // Error handling and resilience
    Features      FeatureConfig       // Feature flags
    Environment   Environment         // Deployment environment
}
```

### 2. Configuration Loader (`loader.go`)

Handles loading from multiple sources with precedence:

```go
// Load configuration with environment override
loader := NewLoader("./config.yaml", "GOCOFFEE")
config, err := loader.Load()

// Environment variable format: PREFIX_SECTION_FIELD
// Example: GOCOFFEE_DATABASE_HOST=localhost
//          GOCOFFEE_AI_PROVIDERS_GEMINI_API_KEY=your-key
```

### 3. Environment Templates (`environments.go`)

Pre-configured templates for different environments:

```go
// Get environment-specific configuration
devConfig := GetDevelopmentConfig()
prodConfig := GetProductionConfig()
stagingConfig := GetStagingConfig()
testConfig := GetTestingConfig()
```

### 4. Configuration Validation (`validator.go`)

Comprehensive validation with detailed error messages:

```go
validator := NewValidator()
if err := validator.Validate(config); err != nil {
    // Handle validation errors
}
```

### 5. Configuration Manager (`manager.go`)

Manages configuration lifecycle with hot reloading:

```go
// Initialize configuration manager
manager := NewManager("./config.yaml", "GOCOFFEE")
manager.EnableWatch(5 * time.Second)

// Add change handlers
manager.AddChangeHandler(func(change ConfigChange) {
    switch change.Type {
    case ConfigReloaded:
        log.Println("Configuration reloaded successfully")
    case ConfigError:
        log.Printf("Configuration error: %v", change.Error)
    }
})
```

## Usage Examples

### Basic Usage

```go
// Load configuration
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// Use configuration
log.Printf("Starting %s on %s", cfg.Service.Name, cfg.GetServerAddress())
```

### Global Configuration Manager

```go
// Initialize global manager
if err := config.InitGlobalManager("./config.yaml", "GOCOFFEE"); err != nil {
    log.Fatal("Failed to initialize config manager:", err)
}

// Use global configuration
serviceConfig := config.GetServiceConfig()
dbConfig := config.GetDatabaseConfig()

// Check feature flags
if config.IsFeatureEnabled("ai") {
    // AI functionality enabled
}
```

### Environment-Specific Loading

```go
// Load by environment
env := os.Getenv("ENVIRONMENT")
config := config.GetConfigForEnvironmentString(env)

// Environment detection
if config.IsProduction() {
    // Production-specific logic
}
```

## Configuration Files

### Development Configuration (`configs/development.yaml`)

```yaml
environment: development

service:
  name: beverage-inventor-agent
  port: 8080
  debug: true

database:
  host: localhost
  database: gocoffee_dev
  ssl_mode: disable

features:
  enable_ai: true
  enable_task_creation: false  # Disabled in development
```

### Production Configuration (`configs/production.yaml`)

```yaml
environment: production

service:
  name: beverage-inventor-agent
  port: 8080
  debug: false

database:
  host: ${DB_HOST}
  database: ${DB_NAME}
  ssl_mode: require

security:
  api:
    require_https: true
    enable_rate_limit: true

features:
  enable_ai: true
  enable_task_creation: true
  enable_notifications: true
```

## Environment Variables

### Database Configuration
```bash
GOCOFFEE_DATABASE_HOST=db.example.com
GOCOFFEE_DATABASE_PASSWORD=secure-password
GOCOFFEE_DATABASE_SSL_MODE=require
```

### AI Provider Configuration
```bash
GOCOFFEE_AI_PROVIDERS_GEMINI_API_KEY=your-gemini-api-key
GOCOFFEE_AI_PROVIDERS_GEMINI_ENABLED=true
```

### External Services
```bash
GOCOFFEE_EXTERNAL_CLICKUP_API_KEY=your-clickup-api-key
GOCOFFEE_EXTERNAL_SLACK_BOT_TOKEN=xoxb-your-bot-token
```

### Security Configuration
```bash
GOCOFFEE_SECURITY_JWT_SECRET_KEY=your-super-secure-secret-key
```

### Feature Flags
```bash
GOCOFFEE_FEATURES_ENABLE_AI=true
GOCOFFEE_FEATURES_ENABLE_TASK_CREATION=true
GOCOFFEE_FEATURES_ENABLE_NOTIFICATIONS=true
```

## Environment Differences

| Setting | Development | Staging | Production |
|---------|-------------|---------|------------|
| Debug Mode | ✅ | ❌ | ❌ |
| CORS Origins | `*` | Specific | Specific |
| SSL Mode | Disable | Require | Require |
| Sampling Rate | 100% | 50% | 10% |
| Rate Limits | Relaxed | Moderate | Strict |
| External Services | Disabled | Enabled | Enabled |

## Validation Rules

### Service Validation
- Service name format (lowercase, alphanumeric, hyphens)
- Semantic versioning for service version
- Valid port range (1-65535)
- Host format validation

### Database Validation
- Supported database drivers
- Connection parameter validation
- Connection pool limits
- SSL mode requirements

### Security Validation
- JWT secret key strength (minimum 32 characters)
- HTTPS requirements in production
- CORS origin validation
- Rate limiting parameters

## Hot Reloading Features

- **File Watching**: Automatic detection of configuration file changes
- **Change Notifications**: Event-driven configuration updates
- **Validation on Reload**: Ensures configuration remains valid
- **Graceful Handling**: Non-blocking reload with error recovery
- **Change Detection**: Identifies significant changes requiring restart

## Best Practices

### 1. Environment Variable Naming
- Use consistent prefix (e.g., `GOCOFFEE_`)
- Follow hierarchical structure (`SECTION_SUBSECTION_FIELD`)
- Use uppercase with underscores

### 2. Secret Management
- Never commit secrets to version control
- Use environment variables for sensitive data
- Consider using secret management systems

### 3. Configuration Validation
- Validate configuration on startup
- Implement environment-specific validation rules
- Provide clear error messages

### 4. Hot Reloading
- Enable hot reloading in development
- Use with caution in production
- Implement proper change detection

### 5. Feature Flags
- Use feature flags for gradual rollouts
- Keep flags simple and boolean
- Clean up unused flags regularly

This configuration management system provides a robust foundation for managing complex application settings across different environments while maintaining security, validation, and operational flexibility.
