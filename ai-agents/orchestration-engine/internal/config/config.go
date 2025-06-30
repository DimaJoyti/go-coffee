package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Cache configuration
	Cache CacheConfig `json:"cache"`

	// Messaging configuration
	Messaging MessagingConfig `json:"messaging"`

	// Agent configuration
	Agents AgentConfig `json:"agents"`

	// Analytics configuration
	Analytics AnalyticsConfig `json:"analytics"`

	// Security configuration
	Security SecurityConfig `json:"security"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// Monitoring configuration
	Monitoring MonitoringConfig `json:"monitoring"`

	// Observability configuration
	Observability ObservabilityConfig `json:"observability"`

	// Dashboard configuration
	Dashboard DashboardConfig `json:"dashboard"`

	// Feature flags
	Features FeatureConfig `json:"features"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port                string        `json:"port"`
	Host                string        `json:"host"`
	ReadTimeout         time.Duration `json:"read_timeout"`
	WriteTimeout        time.Duration `json:"write_timeout"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	MaxHeaderBytes      int           `json:"max_header_bytes"`
	EnableTLS           bool          `json:"enable_tls"`
	TLSCertFile         string        `json:"tls_cert_file"`
	TLSKeyFile          string        `json:"tls_key_file"`
	EnableCORS          bool          `json:"enable_cors"`
	CORSAllowedOrigins  []string      `json:"cors_allowed_origins"`
	EnableCompression   bool          `json:"enable_compression"`
	EnableRateLimiting  bool          `json:"enable_rate_limiting"`
	RateLimitRPS        int           `json:"rate_limit_rps"`
	RateLimitBurst      int           `json:"rate_limit_burst"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	URL                 string        `json:"url"`
	Driver              string        `json:"driver"`
	MaxOpenConnections  int           `json:"max_open_connections"`
	MaxIdleConnections  int           `json:"max_idle_connections"`
	ConnectionLifetime  time.Duration `json:"connection_lifetime"`
	ConnectionTimeout   time.Duration `json:"connection_timeout"`
	QueryTimeout        time.Duration `json:"query_timeout"`
	EnableMigrations    bool          `json:"enable_migrations"`
	MigrationsPath      string        `json:"migrations_path"`
	EnableSSL           bool          `json:"enable_ssl"`
	SSLMode             string        `json:"ssl_mode"`
}

// CacheConfig contains cache-related configuration
type CacheConfig struct {
	URL                 string        `json:"url"`
	Password            string        `json:"password"`
	Database            int           `json:"database"`
	MaxRetries          int           `json:"max_retries"`
	MinRetryBackoff     time.Duration `json:"min_retry_backoff"`
	MaxRetryBackoff     time.Duration `json:"max_retry_backoff"`
	DialTimeout         time.Duration `json:"dial_timeout"`
	ReadTimeout         time.Duration `json:"read_timeout"`
	WriteTimeout        time.Duration `json:"write_timeout"`
	PoolSize            int           `json:"pool_size"`
	MinIdleConnections  int           `json:"min_idle_connections"`
	MaxConnectionAge    time.Duration `json:"max_connection_age"`
	PoolTimeout         time.Duration `json:"pool_timeout"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	IdleCheckFrequency  time.Duration `json:"idle_check_frequency"`
	EnableTLS           bool          `json:"enable_tls"`
}

// MessagingConfig contains messaging-related configuration
type MessagingConfig struct {
	Brokers             []string      `json:"brokers"`
	ClientID            string        `json:"client_id"`
	GroupID             string        `json:"group_id"`
	EnableSASL          bool          `json:"enable_sasl"`
	SASLMechanism       string        `json:"sasl_mechanism"`
	SASLUsername        string        `json:"sasl_username"`
	SASLPassword        string        `json:"sasl_password"`
	EnableTLS           bool          `json:"enable_tls"`
	TLSInsecureSkipVerify bool        `json:"tls_insecure_skip_verify"`
	ProducerConfig      ProducerConfig `json:"producer"`
	ConsumerConfig      ConsumerConfig `json:"consumer"`
}

// ProducerConfig contains Kafka producer configuration
type ProducerConfig struct {
	BatchSize           int           `json:"batch_size"`
	BatchTimeout        time.Duration `json:"batch_timeout"`
	MaxMessageBytes     int           `json:"max_message_bytes"`
	RequiredAcks        int           `json:"required_acks"`
	Compression         string        `json:"compression"`
	FlushFrequency      time.Duration `json:"flush_frequency"`
	RetryMax            int           `json:"retry_max"`
	RetryBackoff        time.Duration `json:"retry_backoff"`
	EnableIdempotent    bool          `json:"enable_idempotent"`
}

// ConsumerConfig contains Kafka consumer configuration
type ConsumerConfig struct {
	MinBytes            int           `json:"min_bytes"`
	MaxBytes            int           `json:"max_bytes"`
	MaxWait             time.Duration `json:"max_wait"`
	CommitInterval      time.Duration `json:"commit_interval"`
	StartOffset         string        `json:"start_offset"`
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	RebalanceTimeout    time.Duration `json:"rebalance_timeout"`
	EnableAutoCommit    bool          `json:"enable_auto_commit"`
}

// AgentConfig contains agent-related configuration
type AgentConfig struct {
	Endpoints           map[string]string `json:"endpoints"`
	Timeout             time.Duration     `json:"timeout"`
	MaxRetries          int               `json:"max_retries"`
	RetryBackoff        time.Duration     `json:"retry_backoff"`
	HealthCheckInterval time.Duration     `json:"health_check_interval"`
	EnableCircuitBreaker bool             `json:"enable_circuit_breaker"`
	CircuitBreakerConfig CircuitBreakerConfig `json:"circuit_breaker"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests         uint32        `json:"max_requests"`
	Interval            time.Duration `json:"interval"`
	Timeout             time.Duration `json:"timeout"`
	ReadyToTrip         func(counts map[string]uint64) bool `json:"-"`
	OnStateChange       func(name string, from, to string) `json:"-"`
}

// AnalyticsConfig contains analytics-related configuration
type AnalyticsConfig struct {
	EnableRealTimeMetrics bool          `json:"enable_real_time_metrics"`
	MetricsInterval       time.Duration `json:"metrics_interval"`
	CacheExpiry           time.Duration `json:"cache_expiry"`
	MaxMetricsHistory     int           `json:"max_metrics_history"`
	EnableAlerts          bool          `json:"enable_alerts"`
	AlertThresholds       AlertThresholds `json:"alert_thresholds"`
	EnablePredictiveAnalytics bool      `json:"enable_predictive_analytics"`
}

// AlertThresholds contains alert threshold configuration
type AlertThresholds struct {
	WorkflowErrorRate     float64 `json:"workflow_error_rate"`
	AgentResponseTime     time.Duration `json:"agent_response_time"`
	SystemCPUUsage        float64 `json:"system_cpu_usage"`
	SystemMemoryUsage     float64 `json:"system_memory_usage"`
	QueueDepth            int64   `json:"queue_depth"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	EnableAuthentication  bool          `json:"enable_authentication"`
	EnableAuthorization   bool          `json:"enable_authorization"`
	JWTSecret             string        `json:"jwt_secret"`
	JWTExpiration         time.Duration `json:"jwt_expiration"`
	EnableAPIKeys         bool          `json:"enable_api_keys"`
	APIKeyHeader          string        `json:"api_key_header"`
	EnableEncryption      bool          `json:"enable_encryption"`
	EncryptionKey         string        `json:"encryption_key"`
	EnableAuditLogging    bool          `json:"enable_audit_logging"`
	AuditLogLevel         string        `json:"audit_log_level"`
	EnableCSRFProtection  bool          `json:"enable_csrf_protection"`
	CSRFTokenLength       int           `json:"csrf_token_length"`
	
	// Additional security features
	EnableRateLimit       bool          `json:"enable_rate_limit"`
	EnableInputValidation bool          `json:"enable_input_validation"`
	EnableSecurityHeaders bool          `json:"enable_security_headers"`
	EnableCORS            bool          `json:"enable_cors"`
	
	// Threat detection settings
	MaxFailedAttempts     int           `json:"max_failed_attempts"`
	ThreatScoreThreshold  float64       `json:"threat_score_threshold"`
	BlockDuration         time.Duration `json:"block_duration"`
	MonitoringWindow      time.Duration `json:"monitoring_window"`
}

// LoggingConfig contains logging-related configuration
type LoggingConfig struct {
	Level               string `json:"level"`
	Format              string `json:"format"`
	Output              string `json:"output"`
	EnableStructured    bool   `json:"enable_structured"`
	EnableColors        bool   `json:"enable_colors"`
	EnableCaller        bool   `json:"enable_caller"`
	EnableStackTrace    bool   `json:"enable_stack_trace"`
	MaxSize             int    `json:"max_size"`
	MaxBackups          int    `json:"max_backups"`
	MaxAge              int    `json:"max_age"`
	Compress            bool   `json:"compress"`
}

// MonitoringConfig contains monitoring-related configuration
type MonitoringConfig struct {
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsPath         string        `json:"metrics_path"`
	EnableTracing       bool          `json:"enable_tracing"`
	TracingEndpoint     string        `json:"tracing_endpoint"`
	TracingSampleRate   float64       `json:"tracing_sample_rate"`
	EnableProfiling     bool          `json:"enable_profiling"`
	ProfilingPath       string        `json:"profiling_path"`
	HealthCheckPath     string        `json:"health_check_path"`
	ReadinessCheckPath  string        `json:"readiness_check_path"`
	EnableStatusPage    bool          `json:"enable_status_page"`
	StatusPagePath      string        `json:"status_page_path"`
	
	// Dashboard configuration
	MaxDashboards             int    `json:"max_dashboards"`
	DefaultRefreshRateSeconds int    `json:"default_refresh_rate_seconds"`
	EnableAutoRefresh         bool   `json:"enable_auto_refresh"`
	EnableAlerts              bool   `json:"enable_alerts"`
	EnableTemplates           bool   `json:"enable_templates"`
	StorageBackend            string `json:"storage_backend"`
	CacheTimeoutSeconds       int    `json:"cache_timeout_seconds"`
	MaxWidgetsPerDashboard    int    `json:"max_widgets_per_dashboard"`
	EnableSharing             bool   `json:"enable_sharing"`
	EnableExport              bool   `json:"enable_export"`
}

// ObservabilityConfig contains observability and telemetry configuration
type ObservabilityConfig struct {
	EnableTracing           bool          `json:"enable_tracing"`
	TracingEndpoint         string        `json:"tracing_endpoint"`
	TracingSampleRate       float64       `json:"tracing_sample_rate"`
	EnableMetrics           bool          `json:"enable_metrics"`
	MetricsEndpoint         string        `json:"metrics_endpoint"`
	MetricsInterval         time.Duration `json:"metrics_interval"`
	EnableHealthChecks      bool          `json:"enable_health_checks"`
	HealthCheckInterval     time.Duration `json:"health_check_interval"`
	EnableInstrumentation   bool          `json:"enable_instrumentation"`
	InstrumentationLevel    string        `json:"instrumentation_level"`
	EnableErrorTracking     bool          `json:"enable_error_tracking"`
	ErrorSamplingRate       float64       `json:"error_sampling_rate"`
	EnablePerformanceMonitoring bool      `json:"enable_performance_monitoring"`
	ServiceName             string        `json:"service_name"`
	ServiceVersion          string        `json:"service_version"`
	Environment             string        `json:"environment"`
	ResourceAttributes      map[string]string `json:"resource_attributes"`
}

// DashboardConfig contains real-time dashboard configuration
type DashboardConfig struct {
	Port                  int      `json:"port"`
	UpdateIntervalSeconds int      `json:"update_interval_seconds"`
	MaxEvents             int      `json:"max_events"`
	EnableWebSocket       bool     `json:"enable_websocket"`
	EnableHTTPAPI         bool     `json:"enable_http_api"`
	EnableStaticFiles     bool     `json:"enable_static_files"`
	StaticFilesPath       string   `json:"static_files_path"`
	CORSEnabled           bool     `json:"cors_enabled"`
	AllowedOrigins        []string `json:"allowed_origins"`
	AuthEnabled           bool     `json:"auth_enabled"`
	AuthToken             string   `json:"auth_token"`
	EnableSSL             bool     `json:"enable_ssl"`
	SSLCertFile           string   `json:"ssl_cert_file"`
	SSLKeyFile            string   `json:"ssl_key_file"`
	MaxConnections        int      `json:"max_connections"`
	ReadTimeout           int      `json:"read_timeout_seconds"`
	WriteTimeout          int      `json:"write_timeout_seconds"`
	IdleTimeout           int      `json:"idle_timeout_seconds"`
}

// FeatureConfig contains feature flag configuration
type FeatureConfig struct {
	EnableWebSocket         bool `json:"enable_websocket"`
	EnableGraphQL           bool `json:"enable_graphql"`
	EnableWorkflowTemplates bool `json:"enable_workflow_templates"`
	EnableAdvancedAnalytics bool `json:"enable_advanced_analytics"`
	EnableAIOptimization    bool `json:"enable_ai_optimization"`
	EnableMultiTenancy      bool `json:"enable_multi_tenancy"`
	EnableWorkflowVersioning bool `json:"enable_workflow_versioning"`
	EnableScheduledWorkflows bool `json:"enable_scheduled_workflows"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:                getEnv("PORT", "8080"),
			Host:                getEnv("HOST", "0.0.0.0"),
			ReadTimeout:         getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:        getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:         getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
			MaxHeaderBytes:      getIntEnv("MAX_HEADER_BYTES", 1<<20), // 1MB
			EnableTLS:           getBoolEnv("ENABLE_TLS", false),
			TLSCertFile:         getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:          getEnv("TLS_KEY_FILE", ""),
			EnableCORS:          getBoolEnv("ENABLE_CORS", true),
			CORSAllowedOrigins:  getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
			EnableCompression:   getBoolEnv("ENABLE_COMPRESSION", true),
			EnableRateLimiting:  getBoolEnv("ENABLE_RATE_LIMITING", true),
			RateLimitRPS:        getIntEnv("RATE_LIMIT_RPS", 100),
			RateLimitBurst:      getIntEnv("RATE_LIMIT_BURST", 200),
		},
		Database: DatabaseConfig{
			URL:                 getEnv("DATABASE_URL", "postgres://localhost/orchestration"),
			Driver:              getEnv("DATABASE_DRIVER", "postgres"),
			MaxOpenConnections:  getIntEnv("DB_MAX_OPEN_CONNECTIONS", 25),
			MaxIdleConnections:  getIntEnv("DB_MAX_IDLE_CONNECTIONS", 5),
			ConnectionLifetime:  getDurationEnv("DB_CONNECTION_LIFETIME", 5*time.Minute),
			ConnectionTimeout:   getDurationEnv("DB_CONNECTION_TIMEOUT", 10*time.Second),
			QueryTimeout:        getDurationEnv("DB_QUERY_TIMEOUT", 30*time.Second),
			EnableMigrations:    getBoolEnv("DB_ENABLE_MIGRATIONS", true),
			MigrationsPath:      getEnv("DB_MIGRATIONS_PATH", "migrations"),
			EnableSSL:           getBoolEnv("DB_ENABLE_SSL", false),
			SSLMode:             getEnv("DB_SSL_MODE", "disable"),
		},
		Cache: CacheConfig{
			URL:                 getEnv("REDIS_URL", "redis://localhost:6379"),
			Password:            getEnv("REDIS_PASSWORD", ""),
			Database:            getIntEnv("REDIS_DATABASE", 0),
			MaxRetries:          getIntEnv("REDIS_MAX_RETRIES", 3),
			MinRetryBackoff:     getDurationEnv("REDIS_MIN_RETRY_BACKOFF", 8*time.Millisecond),
			MaxRetryBackoff:     getDurationEnv("REDIS_MAX_RETRY_BACKOFF", 512*time.Millisecond),
			DialTimeout:         getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:         getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:        getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:            getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConnections:  getIntEnv("REDIS_MIN_IDLE_CONNECTIONS", 2),
			MaxConnectionAge:    getDurationEnv("REDIS_MAX_CONNECTION_AGE", 30*time.Minute),
			PoolTimeout:         getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:         getDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
			IdleCheckFrequency:  getDurationEnv("REDIS_IDLE_CHECK_FREQUENCY", 1*time.Minute),
			EnableTLS:           getBoolEnv("REDIS_ENABLE_TLS", false),
		},
		Messaging: MessagingConfig{
			Brokers:             getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			ClientID:            getEnv("KAFKA_CLIENT_ID", "orchestration-engine"),
			GroupID:             getEnv("KAFKA_GROUP_ID", "orchestration-group"),
			EnableSASL:          getBoolEnv("KAFKA_ENABLE_SASL", false),
			SASLMechanism:       getEnv("KAFKA_SASL_MECHANISM", "PLAIN"),
			SASLUsername:        getEnv("KAFKA_SASL_USERNAME", ""),
			SASLPassword:        getEnv("KAFKA_SASL_PASSWORD", ""),
			EnableTLS:           getBoolEnv("KAFKA_ENABLE_TLS", false),
			TLSInsecureSkipVerify: getBoolEnv("KAFKA_TLS_INSECURE_SKIP_VERIFY", false),
			ProducerConfig: ProducerConfig{
				BatchSize:        getIntEnv("KAFKA_PRODUCER_BATCH_SIZE", 100),
				BatchTimeout:     getDurationEnv("KAFKA_PRODUCER_BATCH_TIMEOUT", 10*time.Millisecond),
				MaxMessageBytes:  getIntEnv("KAFKA_PRODUCER_MAX_MESSAGE_BYTES", 1000000),
				RequiredAcks:     getIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", 1),
				Compression:      getEnv("KAFKA_PRODUCER_COMPRESSION", "snappy"),
				FlushFrequency:   getDurationEnv("KAFKA_PRODUCER_FLUSH_FREQUENCY", 1*time.Second),
				RetryMax:         getIntEnv("KAFKA_PRODUCER_RETRY_MAX", 3),
				RetryBackoff:     getDurationEnv("KAFKA_PRODUCER_RETRY_BACKOFF", 100*time.Millisecond),
				EnableIdempotent: getBoolEnv("KAFKA_PRODUCER_ENABLE_IDEMPOTENT", true),
			},
			ConsumerConfig: ConsumerConfig{
				MinBytes:          getIntEnv("KAFKA_CONSUMER_MIN_BYTES", 1),
				MaxBytes:          getIntEnv("KAFKA_CONSUMER_MAX_BYTES", 10e6),
				MaxWait:           getDurationEnv("KAFKA_CONSUMER_MAX_WAIT", 1*time.Second),
				CommitInterval:    getDurationEnv("KAFKA_CONSUMER_COMMIT_INTERVAL", 1*time.Second),
				StartOffset:       getEnv("KAFKA_CONSUMER_START_OFFSET", "latest"),
				HeartbeatInterval: getDurationEnv("KAFKA_CONSUMER_HEARTBEAT_INTERVAL", 3*time.Second),
				SessionTimeout:    getDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", 30*time.Second),
				RebalanceTimeout:  getDurationEnv("KAFKA_CONSUMER_REBALANCE_TIMEOUT", 30*time.Second),
				EnableAutoCommit:  getBoolEnv("KAFKA_CONSUMER_ENABLE_AUTO_COMMIT", true),
			},
		},
		Agents: AgentConfig{
			Endpoints: map[string]string{
				"social-media-content":  getEnv("SOCIAL_MEDIA_AGENT_URL", "http://localhost:8081"),
				"feedback-analyst":      getEnv("FEEDBACK_ANALYST_AGENT_URL", "http://localhost:8082"),
				"beverage-inventor":     getEnv("BEVERAGE_INVENTOR_AGENT_URL", "http://localhost:8083"),
				"inventory":             getEnv("INVENTORY_AGENT_URL", "http://localhost:8084"),
				"notifier":              getEnv("NOTIFIER_AGENT_URL", "http://localhost:8085"),
			},
			Timeout:             getDurationEnv("AGENT_TIMEOUT", 30*time.Second),
			MaxRetries:          getIntEnv("AGENT_MAX_RETRIES", 3),
			RetryBackoff:        getDurationEnv("AGENT_RETRY_BACKOFF", 1*time.Second),
			HealthCheckInterval: getDurationEnv("AGENT_HEALTH_CHECK_INTERVAL", 30*time.Second),
			EnableCircuitBreaker: getBoolEnv("AGENT_ENABLE_CIRCUIT_BREAKER", true),
		},
		Analytics: AnalyticsConfig{
			EnableRealTimeMetrics: getBoolEnv("ANALYTICS_ENABLE_REAL_TIME_METRICS", true),
			MetricsInterval:       getDurationEnv("ANALYTICS_METRICS_INTERVAL", 10*time.Second),
			CacheExpiry:           getDurationEnv("ANALYTICS_CACHE_EXPIRY", 5*time.Minute),
			MaxMetricsHistory:     getIntEnv("ANALYTICS_MAX_METRICS_HISTORY", 1000),
			EnableAlerts:          getBoolEnv("ANALYTICS_ENABLE_ALERTS", true),
			AlertThresholds: AlertThresholds{
				WorkflowErrorRate:  getFloatEnv("ALERT_WORKFLOW_ERROR_RATE", 10.0),
				AgentResponseTime:  getDurationEnv("ALERT_AGENT_RESPONSE_TIME", 5*time.Second),
				SystemCPUUsage:     getFloatEnv("ALERT_SYSTEM_CPU_USAGE", 80.0),
				SystemMemoryUsage:  getFloatEnv("ALERT_SYSTEM_MEMORY_USAGE", 85.0),
				QueueDepth:         getInt64Env("ALERT_QUEUE_DEPTH", 100),
			},
			EnablePredictiveAnalytics: getBoolEnv("ANALYTICS_ENABLE_PREDICTIVE", false),
		},
		Security: SecurityConfig{
			EnableAuthentication: getBoolEnv("SECURITY_ENABLE_AUTHENTICATION", false),
			EnableAuthorization:  getBoolEnv("SECURITY_ENABLE_AUTHORIZATION", false),
			JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpiration:        getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			EnableAPIKeys:        getBoolEnv("SECURITY_ENABLE_API_KEYS", false),
			APIKeyHeader:         getEnv("SECURITY_API_KEY_HEADER", "X-API-Key"),
			EnableEncryption:     getBoolEnv("SECURITY_ENABLE_ENCRYPTION", false),
			EncryptionKey:        getEnv("SECURITY_ENCRYPTION_KEY", ""),
			EnableAuditLogging:   getBoolEnv("SECURITY_ENABLE_AUDIT_LOGGING", true),
			AuditLogLevel:        getEnv("SECURITY_AUDIT_LOG_LEVEL", "info"),
			EnableCSRFProtection: getBoolEnv("SECURITY_ENABLE_CSRF_PROTECTION", false),
			CSRFTokenLength:      getIntEnv("SECURITY_CSRF_TOKEN_LENGTH", 32),
			
			// Additional security features
			EnableRateLimit:       getBoolEnv("SECURITY_ENABLE_RATE_LIMIT", true),
			EnableInputValidation: getBoolEnv("SECURITY_ENABLE_INPUT_VALIDATION", true),
			EnableSecurityHeaders: getBoolEnv("SECURITY_ENABLE_SECURITY_HEADERS", true),
			EnableCORS:            getBoolEnv("SECURITY_ENABLE_CORS", true),
			
			// Threat detection settings
			MaxFailedAttempts:     getIntEnv("SECURITY_MAX_FAILED_ATTEMPTS", 5),
			ThreatScoreThreshold:  getFloatEnv("SECURITY_THREAT_SCORE_THRESHOLD", 80.0),
			BlockDuration:         getDurationEnv("SECURITY_BLOCK_DURATION", 15*time.Minute),
			MonitoringWindow:      getDurationEnv("SECURITY_MONITORING_WINDOW", 1*time.Hour),
		},
		Logging: LoggingConfig{
			Level:            getEnv("LOG_LEVEL", "info"),
			Format:           getEnv("LOG_FORMAT", "json"),
			Output:           getEnv("LOG_OUTPUT", "stdout"),
			EnableStructured: getBoolEnv("LOG_ENABLE_STRUCTURED", true),
			EnableColors:     getBoolEnv("LOG_ENABLE_COLORS", false),
			EnableCaller:     getBoolEnv("LOG_ENABLE_CALLER", true),
			EnableStackTrace: getBoolEnv("LOG_ENABLE_STACK_TRACE", false),
			MaxSize:          getIntEnv("LOG_MAX_SIZE", 100),
			MaxBackups:       getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAge:           getIntEnv("LOG_MAX_AGE", 28),
			Compress:         getBoolEnv("LOG_COMPRESS", true),
		},
		Monitoring: MonitoringConfig{
			EnableMetrics:       getBoolEnv("MONITORING_ENABLE_METRICS", true),
			MetricsPath:         getEnv("MONITORING_METRICS_PATH", "/metrics"),
			EnableTracing:       getBoolEnv("MONITORING_ENABLE_TRACING", false),
			TracingEndpoint:     getEnv("MONITORING_TRACING_ENDPOINT", ""),
			TracingSampleRate:   getFloatEnv("MONITORING_TRACING_SAMPLE_RATE", 0.1),
			EnableProfiling:     getBoolEnv("MONITORING_ENABLE_PROFILING", false),
			ProfilingPath:       getEnv("MONITORING_PROFILING_PATH", "/debug/pprof"),
			HealthCheckPath:     getEnv("MONITORING_HEALTH_CHECK_PATH", "/health"),
			ReadinessCheckPath:  getEnv("MONITORING_READINESS_CHECK_PATH", "/ready"),
			EnableStatusPage:    getBoolEnv("MONITORING_ENABLE_STATUS_PAGE", true),
			StatusPagePath:      getEnv("MONITORING_STATUS_PAGE_PATH", "/status"),
			
			// Dashboard configuration
			MaxDashboards:             getIntEnv("MONITORING_MAX_DASHBOARDS", 100),
			DefaultRefreshRateSeconds: getIntEnv("MONITORING_DEFAULT_REFRESH_RATE_SECONDS", 30),
			EnableAutoRefresh:         getBoolEnv("MONITORING_ENABLE_AUTO_REFRESH", true),
			EnableAlerts:              getBoolEnv("MONITORING_ENABLE_ALERTS", true),
			EnableTemplates:           getBoolEnv("MONITORING_ENABLE_TEMPLATES", true),
			StorageBackend:            getEnv("MONITORING_STORAGE_BACKEND", "memory"),
			CacheTimeoutSeconds:       getIntEnv("MONITORING_CACHE_TIMEOUT_SECONDS", 300),
			MaxWidgetsPerDashboard:    getIntEnv("MONITORING_MAX_WIDGETS_PER_DASHBOARD", 50),
			EnableSharing:             getBoolEnv("MONITORING_ENABLE_SHARING", true),
			EnableExport:              getBoolEnv("MONITORING_ENABLE_EXPORT", true),
		},
		Observability: ObservabilityConfig{
			EnableTracing:              getBoolEnv("OBSERVABILITY_ENABLE_TRACING", true),
			TracingEndpoint:            getEnv("OBSERVABILITY_TRACING_ENDPOINT", "http://localhost:14268/api/traces"),
			TracingSampleRate:          getFloatEnv("OBSERVABILITY_TRACING_SAMPLE_RATE", 0.1),
			EnableMetrics:              getBoolEnv("OBSERVABILITY_ENABLE_METRICS", true),
			MetricsEndpoint:            getEnv("OBSERVABILITY_METRICS_ENDPOINT", "http://localhost:9090/api/v1/write"),
			MetricsInterval:            getDurationEnv("OBSERVABILITY_METRICS_INTERVAL", 15*time.Second),
			EnableHealthChecks:         getBoolEnv("OBSERVABILITY_ENABLE_HEALTH_CHECKS", true),
			HealthCheckInterval:        getDurationEnv("OBSERVABILITY_HEALTH_CHECK_INTERVAL", 30*time.Second),
			EnableInstrumentation:      getBoolEnv("OBSERVABILITY_ENABLE_INSTRUMENTATION", true),
			InstrumentationLevel:       getEnv("OBSERVABILITY_INSTRUMENTATION_LEVEL", "detailed"),
			EnableErrorTracking:        getBoolEnv("OBSERVABILITY_ENABLE_ERROR_TRACKING", true),
			ErrorSamplingRate:          getFloatEnv("OBSERVABILITY_ERROR_SAMPLING_RATE", 1.0),
			EnablePerformanceMonitoring: getBoolEnv("OBSERVABILITY_ENABLE_PERFORMANCE_MONITORING", true),
			ServiceName:                getEnv("OBSERVABILITY_SERVICE_NAME", "orchestration-engine"),
			ServiceVersion:             getEnv("OBSERVABILITY_SERVICE_VERSION", "1.0.0"),
			Environment:                getEnv("OBSERVABILITY_ENVIRONMENT", "development"),
			ResourceAttributes:         map[string]string{
				"service.name":    getEnv("OBSERVABILITY_SERVICE_NAME", "orchestration-engine"),
				"service.version": getEnv("OBSERVABILITY_SERVICE_VERSION", "1.0.0"),
				"deployment.environment": getEnv("OBSERVABILITY_ENVIRONMENT", "development"),
			},
		},
		Dashboard: DashboardConfig{
			Port:                  getIntEnv("DASHBOARD_PORT", 8090),
			UpdateIntervalSeconds: getIntEnv("DASHBOARD_UPDATE_INTERVAL_SECONDS", 5),
			MaxEvents:             getIntEnv("DASHBOARD_MAX_EVENTS", 1000),
			EnableWebSocket:       getBoolEnv("DASHBOARD_ENABLE_WEBSOCKET", true),
			EnableHTTPAPI:         getBoolEnv("DASHBOARD_ENABLE_HTTP_API", true),
			EnableStaticFiles:     getBoolEnv("DASHBOARD_ENABLE_STATIC_FILES", true),
			StaticFilesPath:       getEnv("DASHBOARD_STATIC_FILES_PATH", "./web/dashboard"),
			CORSEnabled:           getBoolEnv("DASHBOARD_CORS_ENABLED", true),
			AllowedOrigins:        getSliceEnv("DASHBOARD_ALLOWED_ORIGINS", []string{"*"}),
			AuthEnabled:           getBoolEnv("DASHBOARD_AUTH_ENABLED", false),
			AuthToken:             getEnv("DASHBOARD_AUTH_TOKEN", ""),
			EnableSSL:             getBoolEnv("DASHBOARD_ENABLE_SSL", false),
			SSLCertFile:           getEnv("DASHBOARD_SSL_CERT_FILE", ""),
			SSLKeyFile:            getEnv("DASHBOARD_SSL_KEY_FILE", ""),
			MaxConnections:        getIntEnv("DASHBOARD_MAX_CONNECTIONS", 100),
			ReadTimeout:           getIntEnv("DASHBOARD_READ_TIMEOUT_SECONDS", 30),
			WriteTimeout:          getIntEnv("DASHBOARD_WRITE_TIMEOUT_SECONDS", 30),
			IdleTimeout:           getIntEnv("DASHBOARD_IDLE_TIMEOUT_SECONDS", 120),
		},
		Features: FeatureConfig{
			EnableWebSocket:         getBoolEnv("FEATURE_ENABLE_WEBSOCKET", true),
			EnableGraphQL:           getBoolEnv("FEATURE_ENABLE_GRAPHQL", false),
			EnableWorkflowTemplates: getBoolEnv("FEATURE_ENABLE_WORKFLOW_TEMPLATES", true),
			EnableAdvancedAnalytics: getBoolEnv("FEATURE_ENABLE_ADVANCED_ANALYTICS", true),
			EnableAIOptimization:    getBoolEnv("FEATURE_ENABLE_AI_OPTIMIZATION", false),
			EnableMultiTenancy:      getBoolEnv("FEATURE_ENABLE_MULTI_TENANCY", false),
			EnableWorkflowVersioning: getBoolEnv("FEATURE_ENABLE_WORKFLOW_VERSIONING", true),
			EnableScheduledWorkflows: getBoolEnv("FEATURE_ENABLE_SCHEDULED_WORKFLOWS", true),
		},
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if c.Cache.URL == "" {
		return fmt.Errorf("cache URL is required")
	}

	if len(c.Messaging.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}

	if len(c.Agents.Endpoints) == 0 {
		return fmt.Errorf("at least one agent endpoint is required")
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
