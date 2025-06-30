package observability

import (
	"time"
)

// TelemetryConfig configures OpenTelemetry components
type TelemetryConfig struct {
	ServiceName    string         `yaml:"service_name"`
	ServiceVersion string         `yaml:"service_version"`
	Environment    string         `yaml:"environment"`
	Tracing        TracingConfig  `yaml:"tracing"`
	Metrics        MetricsConfig  `yaml:"metrics"`
	Logging        LoggingConfig  `yaml:"logging"`
	Exporters      ExportersConfig `yaml:"exporters"`
}

// TracingConfig configures distributed tracing
type TracingConfig struct {
	Enabled        bool    `yaml:"enabled"`
	SamplingRate   float64 `yaml:"sampling_rate"`
	MaxSpansPerTrace int   `yaml:"max_spans_per_trace"`
	SpanProcessors []SpanProcessorConfig `yaml:"span_processors"`
}

// SpanProcessorConfig configures span processors
type SpanProcessorConfig struct {
	Type       string        `yaml:"type"` // batch, simple
	BatchSize  int           `yaml:"batch_size"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxQueue   int           `yaml:"max_queue"`
}

// MetricsConfig configures metrics collection
type MetricsConfig struct {
	Enabled         bool          `yaml:"enabled"`
	CollectInterval time.Duration `yaml:"collect_interval"`
	Readers         []MetricReaderConfig `yaml:"readers"`
	Views           []MetricViewConfig   `yaml:"views"`
}

// MetricReaderConfig configures metric readers
type MetricReaderConfig struct {
	Type     string        `yaml:"type"` // periodic, manual
	Interval time.Duration `yaml:"interval"`
}

// MetricViewConfig configures metric views
type MetricViewConfig struct {
	Name        string   `yaml:"name"`
	Instrument  string   `yaml:"instrument"`
	Aggregation string   `yaml:"aggregation"`
	Attributes  []string `yaml:"attributes"`
}

// LoggingConfig configures structured logging
type LoggingConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Level         string `yaml:"level"`
	Format        string `yaml:"format"` // json, text
	IncludeTrace  bool   `yaml:"include_trace"`
	IncludeSpan   bool   `yaml:"include_span"`
	CorrelationID bool   `yaml:"correlation_id"`
}

// ExportersConfig configures telemetry exporters
type ExportersConfig struct {
	OTLP       OTLPConfig       `yaml:"otlp"`
	Jaeger     JaegerConfig     `yaml:"jaeger"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Console    ConsoleConfig    `yaml:"console"`
	File       FileConfig       `yaml:"file"`
}

// OTLPConfig configures OTLP exporter
type OTLPConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Endpoint  string            `yaml:"endpoint"`
	Headers   map[string]string `yaml:"headers"`
	Insecure  bool              `yaml:"insecure"`
	Timeout   time.Duration     `yaml:"timeout"`
	Retry     RetryConfig       `yaml:"retry"`
}

// JaegerConfig configures Jaeger exporter
type JaegerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// PrometheusConfig configures Prometheus exporter
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// ConsoleConfig configures console exporter
type ConsoleConfig struct {
	Enabled bool `yaml:"enabled"`
	Pretty  bool `yaml:"pretty"`
}

// FileConfig configures file exporter
type FileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	TracePath string `yaml:"trace_path"`
	MetricPath string `yaml:"metric_path"`
}

// RetryConfig configures retry behavior for exporters
type RetryConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxAttempts int           `yaml:"max_attempts"`
	InitialDelay time.Duration `yaml:"initial_delay"`
	MaxDelay    time.Duration `yaml:"max_delay"`
}

// DefaultTelemetryConfig returns a default telemetry configuration
func DefaultTelemetryConfig(serviceName string) TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    serviceName,
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Tracing: TracingConfig{
			Enabled:      true,
			SamplingRate: 1.0, // 100% sampling in development
			MaxSpansPerTrace: 1000,
			SpanProcessors: []SpanProcessorConfig{
				{
					Type:      "batch",
					BatchSize: 512,
					Timeout:   5 * time.Second,
					MaxQueue:  2048,
				},
			},
		},
		Metrics: MetricsConfig{
			Enabled:         true,
			CollectInterval: 30 * time.Second,
			Readers: []MetricReaderConfig{
				{
					Type:     "periodic",
					Interval: 30 * time.Second,
				},
			},
		},
		Logging: LoggingConfig{
			Enabled:       true,
			Level:         "info",
			Format:        "json",
			IncludeTrace:  true,
			IncludeSpan:   true,
			CorrelationID: true,
		},
		Exporters: ExportersConfig{
			OTLP: OTLPConfig{
				Enabled:  false,
				Endpoint: "http://localhost:4317",
				Insecure: true,
				Timeout:  10 * time.Second,
				Retry: RetryConfig{
					Enabled:      true,
					MaxAttempts:  3,
					InitialDelay: 1 * time.Second,
					MaxDelay:     30 * time.Second,
				},
			},
			Jaeger: JaegerConfig{
				Enabled:  true,
				Endpoint: "http://localhost:14268/api/traces",
			},
			Prometheus: PrometheusConfig{
				Enabled: true,
				Port:    9090,
				Path:    "/metrics",
			},
			Console: ConsoleConfig{
				Enabled: true,
				Pretty:  true,
			},
			File: FileConfig{
				Enabled:    false,
				TracePath:  "/tmp/traces.json",
				MetricPath: "/tmp/metrics.json",
			},
		},
	}
}

// ProductionTelemetryConfig returns a production-ready telemetry configuration
func ProductionTelemetryConfig(serviceName string) TelemetryConfig {
	config := DefaultTelemetryConfig(serviceName)
	
	// Production-specific settings
	config.Environment = "production"
	config.Tracing.SamplingRate = 0.1 // 10% sampling in production
	config.Logging.Level = "warn"
	
	// Enable OTLP for production
	config.Exporters.OTLP.Enabled = true
	config.Exporters.Console.Enabled = false
	
	// Disable Jaeger in favor of OTLP
	config.Exporters.Jaeger.Enabled = false
	
	return config
}

// StagingTelemetryConfig returns a staging telemetry configuration
func StagingTelemetryConfig(serviceName string) TelemetryConfig {
	config := DefaultTelemetryConfig(serviceName)
	
	// Staging-specific settings
	config.Environment = "staging"
	config.Tracing.SamplingRate = 0.5 // 50% sampling in staging
	config.Logging.Level = "info"
	
	return config
}

// GetConfigForEnvironment returns configuration for a specific environment
func GetConfigForEnvironment(serviceName, environment string) TelemetryConfig {
	switch environment {
	case "production", "prod":
		return ProductionTelemetryConfig(serviceName)
	case "staging", "stage":
		return StagingTelemetryConfig(serviceName)
	case "development", "dev", "local":
		return DefaultTelemetryConfig(serviceName)
	default:
		return DefaultTelemetryConfig(serviceName)
	}
}

// ResourceAttributes defines standard resource attributes
type ResourceAttributes struct {
	ServiceName      string
	ServiceVersion   string
	ServiceNamespace string
	ServiceInstance  string
	Environment      string
	Region           string
	AvailabilityZone string
	CloudProvider    string
	ContainerName    string
	ContainerID      string
	PodName          string
	NodeName         string
	ClusterName      string
}

// DefaultResourceAttributes returns default resource attributes
func DefaultResourceAttributes(serviceName string) ResourceAttributes {
	return ResourceAttributes{
		ServiceName:      serviceName,
		ServiceVersion:   "1.0.0",
		ServiceNamespace: "go-coffee",
		Environment:      "development",
		Region:           "us-west-2",
		CloudProvider:    "aws",
	}
}

// InstrumentationConfig configures automatic instrumentation
type InstrumentationConfig struct {
	HTTP     HTTPInstrumentationConfig     `yaml:"http"`
	GRPC     GRPCInstrumentationConfig     `yaml:"grpc"`
	Database DatabaseInstrumentationConfig `yaml:"database"`
	Kafka    KafkaInstrumentationConfig    `yaml:"kafka"`
	Redis    RedisInstrumentationConfig    `yaml:"redis"`
}

// HTTPInstrumentationConfig configures HTTP instrumentation
type HTTPInstrumentationConfig struct {
	Enabled           bool     `yaml:"enabled"`
	CaptureHeaders    bool     `yaml:"capture_headers"`
	CaptureBody       bool     `yaml:"capture_body"`
	SensitiveHeaders  []string `yaml:"sensitive_headers"`
	IgnoreRoutes      []string `yaml:"ignore_routes"`
	MaxBodySize       int      `yaml:"max_body_size"`
}

// GRPCInstrumentationConfig configures gRPC instrumentation
type GRPCInstrumentationConfig struct {
	Enabled        bool `yaml:"enabled"`
	CaptureMessage bool `yaml:"capture_message"`
	MaxMessageSize int  `yaml:"max_message_size"`
}

// DatabaseInstrumentationConfig configures database instrumentation
type DatabaseInstrumentationConfig struct {
	Enabled           bool     `yaml:"enabled"`
	CaptureStatements bool     `yaml:"capture_statements"`
	SanitizeSQL       bool     `yaml:"sanitize_sql"`
	IgnoreTables      []string `yaml:"ignore_tables"`
}

// KafkaInstrumentationConfig configures Kafka instrumentation
type KafkaInstrumentationConfig struct {
	Enabled        bool     `yaml:"enabled"`
	CaptureMessage bool     `yaml:"capture_message"`
	IgnoreTopics   []string `yaml:"ignore_topics"`
	MaxMessageSize int      `yaml:"max_message_size"`
}

// RedisInstrumentationConfig configures Redis instrumentation
type RedisInstrumentationConfig struct {
	Enabled         bool     `yaml:"enabled"`
	CaptureCommands bool     `yaml:"capture_commands"`
	SanitizeCommands bool    `yaml:"sanitize_commands"`
	IgnoreCommands  []string `yaml:"ignore_commands"`
}

// DefaultInstrumentationConfig returns default instrumentation configuration
func DefaultInstrumentationConfig() InstrumentationConfig {
	return InstrumentationConfig{
		HTTP: HTTPInstrumentationConfig{
			Enabled:        true,
			CaptureHeaders: true,
			CaptureBody:    false,
			SensitiveHeaders: []string{
				"authorization",
				"cookie",
				"x-api-key",
				"x-auth-token",
			},
			IgnoreRoutes: []string{
				"/health",
				"/metrics",
				"/favicon.ico",
			},
			MaxBodySize: 1024,
		},
		GRPC: GRPCInstrumentationConfig{
			Enabled:        true,
			CaptureMessage: false,
			MaxMessageSize: 1024,
		},
		Database: DatabaseInstrumentationConfig{
			Enabled:           true,
			CaptureStatements: true,
			SanitizeSQL:       true,
			IgnoreTables: []string{
				"health_check",
				"migrations",
			},
		},
		Kafka: KafkaInstrumentationConfig{
			Enabled:        true,
			CaptureMessage: false,
			IgnoreTopics: []string{
				"__consumer_offsets",
				"_schemas",
			},
			MaxMessageSize: 1024,
		},
		Redis: RedisInstrumentationConfig{
			Enabled:         true,
			CaptureCommands: true,
			SanitizeCommands: true,
			IgnoreCommands: []string{
				"ping",
				"info",
			},
		},
	}
}
