package config

import (
	"fmt"
	"time"
)

// InfrastructureConfig holds all infrastructure configuration
type InfrastructureConfig struct {
	Redis    *RedisConfig    `yaml:"redis" json:"redis"`
	Database *DatabaseConfig `yaml:"database" json:"database"`
	Security *SecurityConfig `yaml:"security" json:"security"`
	Events   *EventsConfig   `yaml:"events" json:"events"`
	Cache    *CacheConfig    `yaml:"cache" json:"cache"`
	Metrics  *MetricsConfig  `yaml:"metrics" json:"metrics"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	// Connection settings
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
	
	// Pool settings
	PoolSize        int           `yaml:"pool_size" json:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns" json:"min_idle_conns"`
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	
	// Timeout settings
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	
	// Cluster settings
	ClusterMode bool     `yaml:"cluster_mode" json:"cluster_mode"`
	ClusterHosts []string `yaml:"cluster_hosts" json:"cluster_hosts"`
	
	// Sentinel settings
	SentinelMode      bool     `yaml:"sentinel_mode" json:"sentinel_mode"`
	SentinelHosts     []string `yaml:"sentinel_hosts" json:"sentinel_hosts"`
	SentinelMaster    string   `yaml:"sentinel_master" json:"sentinel_master"`
	SentinelPassword  string   `yaml:"sentinel_password" json:"sentinel_password"`
	
	// Key prefix for namespacing
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix"`
	
	// SSL/TLS settings
	TLSEnabled   bool   `yaml:"tls_enabled" json:"tls_enabled"`
	TLSCertFile  string `yaml:"tls_cert_file" json:"tls_cert_file"`
	TLSKeyFile   string `yaml:"tls_key_file" json:"tls_key_file"`
	TLSCAFile    string `yaml:"tls_ca_file" json:"tls_ca_file"`
	TLSSkipVerify bool  `yaml:"tls_skip_verify" json:"tls_skip_verify"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	// Connection settings
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	
	// Pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	
	// Timeout settings
	ConnectTimeout time.Duration `yaml:"connect_timeout" json:"connect_timeout"`
	QueryTimeout   time.Duration `yaml:"query_timeout" json:"query_timeout"`
	
	// Migration settings
	MigrationsPath   string `yaml:"migrations_path" json:"migrations_path"`
	MigrationsTable  string `yaml:"migrations_table" json:"migrations_table"`
	AutoMigrate      bool   `yaml:"auto_migrate" json:"auto_migrate"`
	
	// Monitoring
	LogQueries      bool          `yaml:"log_queries" json:"log_queries"`
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold" json:"slow_query_threshold"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	// JWT settings
	JWT *JWTConfig `yaml:"jwt" json:"jwt"`
	
	// Encryption settings
	Encryption *EncryptionConfig `yaml:"encryption" json:"encryption"`
	
	// Rate limiting
	RateLimit *RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	
	// CORS settings
	CORS *CORSConfig `yaml:"cors" json:"cors"`
	
	// Security headers
	SecurityHeaders *SecurityHeadersConfig `yaml:"security_headers" json:"security_headers"`
	
	// Session settings
	Session *SessionConfig `yaml:"session" json:"session"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey        string        `yaml:"secret_key" json:"secret_key"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl" json:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl" json:"refresh_token_ttl"`
	Issuer           string        `yaml:"issuer" json:"issuer"`
	Audience         string        `yaml:"audience" json:"audience"`
	Algorithm        string        `yaml:"algorithm" json:"algorithm"`
	RefreshThreshold time.Duration `yaml:"refresh_threshold" json:"refresh_threshold"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	AESKey    string `yaml:"aes_key" json:"aes_key"`
	Algorithm string `yaml:"algorithm" json:"algorithm"`
	KeySize   int    `yaml:"key_size" json:"key_size"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	RequestsPerMinute int     `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize   int           `yaml:"burst_size" json:"burst_size"`
	WindowSize  time.Duration `yaml:"window_size" json:"window_size"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age"`
}

// SecurityHeadersConfig represents security headers configuration
type SecurityHeadersConfig struct {
	ContentTypeNosniff    bool   `yaml:"content_type_nosniff" json:"content_type_nosniff"`
	FrameDeny             bool   `yaml:"frame_deny" json:"frame_deny"`
	ContentSecurityPolicy string `yaml:"content_security_policy" json:"content_security_policy"`
	ReferrerPolicy        string `yaml:"referrer_policy" json:"referrer_policy"`
	XSSProtection         bool   `yaml:"xss_protection" json:"xss_protection"`
	HSTSMaxAge            int    `yaml:"hsts_max_age" json:"hsts_max_age"`
}

// SessionConfig represents session configuration
type SessionConfig struct {
	CookieName     string        `yaml:"cookie_name" json:"cookie_name"`
	CookieDomain   string        `yaml:"cookie_domain" json:"cookie_domain"`
	CookiePath     string        `yaml:"cookie_path" json:"cookie_path"`
	CookieSecure   bool          `yaml:"cookie_secure" json:"cookie_secure"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only" json:"cookie_http_only"`
	CookieSameSite string        `yaml:"cookie_same_site" json:"cookie_same_site"`
	MaxAge         time.Duration `yaml:"max_age" json:"max_age"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// EventsConfig represents events configuration
type EventsConfig struct {
	// Event store settings
	Store *EventStoreConfig `yaml:"store" json:"store"`
	
	// Publisher settings
	Publisher *EventPublisherConfig `yaml:"publisher" json:"publisher"`
	
	// Subscriber settings
	Subscriber *EventSubscriberConfig `yaml:"subscriber" json:"subscriber"`
	
	// Retry settings
	Retry *EventRetryConfig `yaml:"retry" json:"retry"`
}

// EventStoreConfig represents event store configuration
type EventStoreConfig struct {
	Type           string        `yaml:"type" json:"type"` // redis, postgres, kafka
	RetentionDays  int           `yaml:"retention_days" json:"retention_days"`
	BatchSize      int           `yaml:"batch_size" json:"batch_size"`
	FlushInterval  time.Duration `yaml:"flush_interval" json:"flush_interval"`
	Compression    bool          `yaml:"compression" json:"compression"`
}

// EventPublisherConfig represents event publisher configuration
type EventPublisherConfig struct {
	BufferSize    int           `yaml:"buffer_size" json:"buffer_size"`
	Workers       int           `yaml:"workers" json:"workers"`
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`
	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// EventSubscriberConfig represents event subscriber configuration
type EventSubscriberConfig struct {
	Workers        int           `yaml:"workers" json:"workers"`
	BufferSize     int           `yaml:"buffer_size" json:"buffer_size"`
	AckTimeout     time.Duration `yaml:"ack_timeout" json:"ack_timeout"`
	MaxRetries     int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay" json:"retry_delay"`
	DeadLetterQueue bool         `yaml:"dead_letter_queue" json:"dead_letter_queue"`
}

// EventRetryConfig represents event retry configuration
type EventRetryConfig struct {
	MaxAttempts   int           `yaml:"max_attempts" json:"max_attempts"`
	InitialDelay  time.Duration `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay      time.Duration `yaml:"max_delay" json:"max_delay"`
	Multiplier    float64       `yaml:"multiplier" json:"multiplier"`
	Jitter        bool          `yaml:"jitter" json:"jitter"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	DefaultTTL      time.Duration `yaml:"default_ttl" json:"default_ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	MaxSize         int64         `yaml:"max_size" json:"max_size"`
	Compression     bool          `yaml:"compression" json:"compression"`
	Serialization   string        `yaml:"serialization" json:"serialization"` // json, msgpack, gob
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	Path       string        `yaml:"path" json:"path"`
	Namespace  string        `yaml:"namespace" json:"namespace"`
	Subsystem  string        `yaml:"subsystem" json:"subsystem"`
	Interval   time.Duration `yaml:"interval" json:"interval"`
	Buckets    []float64     `yaml:"buckets" json:"buckets"`
}

// GetRedisAddr returns the Redis address
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetDatabaseDSN returns the database DSN
func (d *DatabaseConfig) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode)
}

// Validate validates the infrastructure configuration
func (c *InfrastructureConfig) Validate() error {
	if c.Redis != nil {
		if err := c.Redis.Validate(); err != nil {
			return fmt.Errorf("redis config validation failed: %w", err)
		}
	}
	
	if c.Database != nil {
		if err := c.Database.Validate(); err != nil {
			return fmt.Errorf("database config validation failed: %w", err)
		}
	}
	
	if c.Security != nil {
		if err := c.Security.Validate(); err != nil {
			return fmt.Errorf("security config validation failed: %w", err)
		}
	}
	
	return nil
}

// Validate validates Redis configuration
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("redis port must be between 1 and 65535")
	}
	if r.PoolSize <= 0 {
		r.PoolSize = 10 // default
	}
	if r.MinIdleConns < 0 {
		r.MinIdleConns = 0
	}
	return nil
}

// Validate validates Database configuration
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if d.Port <= 0 || d.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if d.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if d.Username == "" {
		return fmt.Errorf("database username is required")
	}
	return nil
}

// Validate validates Security configuration
func (s *SecurityConfig) Validate() error {
	if s.JWT != nil {
		if s.JWT.SecretKey == "" {
			return fmt.Errorf("JWT secret key is required")
		}
		if s.JWT.AccessTokenTTL <= 0 {
			s.JWT.AccessTokenTTL = 15 * time.Minute // default
		}
		if s.JWT.RefreshTokenTTL <= 0 {
			s.JWT.RefreshTokenTTL = 24 * time.Hour // default
		}
	}
	return nil
}

// DefaultInfrastructureConfig returns default infrastructure configuration
func DefaultInfrastructureConfig() *InfrastructureConfig {
	return &InfrastructureConfig{
		Redis: &RedisConfig{
			Host:            "localhost",
			Port:            6379,
			DB:              0,
			PoolSize:        10,
			MinIdleConns:    5,
			MaxRetries:      3,
			RetryDelay:      100 * time.Millisecond,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			IdleTimeout:     5 * time.Minute,
			KeyPrefix:       "go-coffee:",
		},
		Database: &DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Database:        "go_coffee",
			Username:        "postgres",
			Password:        "postgres",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
			ConnectTimeout:  10 * time.Second,
			QueryTimeout:    30 * time.Second,
			MigrationsPath:  "./migrations",
			MigrationsTable: "schema_migrations",
			AutoMigrate:     false,
			LogQueries:      false,
			SlowQueryThreshold: 1 * time.Second,
		},
		Security: &SecurityConfig{
			JWT: &JWTConfig{
				AccessTokenTTL:   15 * time.Minute,
				RefreshTokenTTL:  24 * time.Hour,
				Issuer:           "go-coffee",
				Algorithm:        "HS256",
				RefreshThreshold: 5 * time.Minute,
			},
			RateLimit: &RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 100,
				BurstSize:         10,
				WindowSize:        1 * time.Minute,
				CleanupInterval:   5 * time.Minute,
			},
			CORS: &CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
				MaxAge:           86400,
			},
			SecurityHeaders: &SecurityHeadersConfig{
				ContentTypeNosniff:    true,
				FrameDeny:             true,
				ContentSecurityPolicy: "default-src 'self'",
				ReferrerPolicy:        "strict-origin-when-cross-origin",
				XSSProtection:         true,
				HSTSMaxAge:            31536000,
			},
			Session: &SessionConfig{
				CookieName:     "session_id",
				CookiePath:     "/",
				CookieSecure:   false,
				CookieHTTPOnly: true,
				CookieSameSite: "Lax",
				MaxAge:         24 * time.Hour,
				IdleTimeout:    30 * time.Minute,
			},
		},
		Events: &EventsConfig{
			Store: &EventStoreConfig{
				Type:          "redis",
				RetentionDays: 30,
				BatchSize:     100,
				FlushInterval: 1 * time.Second,
				Compression:   false,
			},
			Publisher: &EventPublisherConfig{
				BufferSize:    1000,
				Workers:       5,
				FlushInterval: 100 * time.Millisecond,
				MaxRetries:    3,
				RetryDelay:    1 * time.Second,
			},
			Subscriber: &EventSubscriberConfig{
				Workers:         5,
				BufferSize:      1000,
				AckTimeout:      30 * time.Second,
				MaxRetries:      3,
				RetryDelay:      1 * time.Second,
				DeadLetterQueue: true,
			},
			Retry: &EventRetryConfig{
				MaxAttempts:  5,
				InitialDelay: 1 * time.Second,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
		},
		Cache: &CacheConfig{
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
			MaxSize:         100 * 1024 * 1024, // 100MB
			Compression:     false,
			Serialization:   "json",
		},
		Metrics: &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "go_coffee",
			Interval:  15 * time.Second,
			Buckets:   []float64{0.1, 0.3, 1.2, 5.0},
		},
	}
}
