package schema

import (
	"fmt"
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	// Connection settings
	Driver   string `yaml:"driver" json:"driver" validate:"required,db_driver" default:"postgres"`
	Host     string `yaml:"host" json:"host" validate:"required,host" default:"localhost"`
	Port     int    `yaml:"port" json:"port" validate:"required,port" default:"5432"`
	Database string `yaml:"database" json:"database" validate:"required" default:"go_coffee_ai_agents"`
	Username string `yaml:"username" json:"username" validate:"required" default:"postgres"`
	Password string `yaml:"password" json:"password" validate:"secret"`
	
	// Secret management
	PasswordPath string `yaml:"password_path" json:"password_path"` // Path to secret in Vault
	
	// SSL/TLS settings
	SSLMode         string `yaml:"ssl_mode" json:"ssl_mode" validate:"oneof=disable require verify-ca verify-full" default:"require"`
	SSLCert         string `yaml:"ssl_cert" json:"ssl_cert"`
	SSLKey          string `yaml:"ssl_key" json:"ssl_key"`
	SSLRootCert     string `yaml:"ssl_root_cert" json:"ssl_root_cert"`
	
	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns" validate:"min=1" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns" validate:"min=1" default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime" default:"1h"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time" default:"15m"`
	
	// Query settings
	QueryTimeout    time.Duration `yaml:"query_timeout" json:"query_timeout" default:"30s"`
	
	// Migration settings
	Migration MigrationConfig `yaml:"migration" json:"migration"`
	
	// Read replica settings
	ReadReplicas []ReplicaConfig `yaml:"read_replicas" json:"read_replicas"`
	
	// Backup settings
	Backup BackupConfig `yaml:"backup" json:"backup"`
	
	// Monitoring settings
	Monitoring DatabaseMonitoringConfig `yaml:"monitoring" json:"monitoring"`
}

// MigrationConfig holds database migration configuration
type MigrationConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled" default:"true"`
	MigrationsPath  string `yaml:"migrations_path" json:"migrations_path" default:"./migrations"`
	MigrationsTable string `yaml:"migrations_table" json:"migrations_table" default:"schema_migrations"`
	AutoMigrate     bool   `yaml:"auto_migrate" json:"auto_migrate" default:"false"`
	
	// Migration behavior
	AllowDirty      bool `yaml:"allow_dirty" json:"allow_dirty" default:"false"`
	LockTimeout     time.Duration `yaml:"lock_timeout" json:"lock_timeout" default:"15m"`
	
	// Rollback settings
	EnableRollback  bool `yaml:"enable_rollback" json:"enable_rollback" default:"true"`
	MaxRollbacks    int  `yaml:"max_rollbacks" json:"max_rollbacks" validate:"min=0" default:"5"`
}

// ReplicaConfig holds read replica configuration
type ReplicaConfig struct {
	Name     string `yaml:"name" json:"name" validate:"required"`
	Host     string `yaml:"host" json:"host" validate:"required,host"`
	Port     int    `yaml:"port" json:"port" validate:"required,port"`
	Database string `yaml:"database" json:"database" validate:"required"`
	Username string `yaml:"username" json:"username" validate:"required"`
	Password string `yaml:"password" json:"password" validate:"secret"`
	
	// Secret management
	PasswordPath string `yaml:"password_path" json:"password_path"`
	
	// Connection settings
	MaxOpenConns int           `yaml:"max_open_conns" json:"max_open_conns" validate:"min=1" default:"10"`
	MaxIdleConns int           `yaml:"max_idle_conns" json:"max_idle_conns" validate:"min=1" default:"2"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime" default:"1h"`
	
	// Load balancing
	Weight   int  `yaml:"weight" json:"weight" validate:"min=1" default:"1"`
	Enabled  bool `yaml:"enabled" json:"enabled" default:"true"`
	
	// Health check
	HealthCheck HealthCheckConfig `yaml:"health_check" json:"health_check"`
}

// BackupConfig holds database backup configuration
type BackupConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled" default:"false"`
	Schedule        string        `yaml:"schedule" json:"schedule"` // Cron expression
	RetentionDays   int           `yaml:"retention_days" json:"retention_days" validate:"min=1" default:"30"`
	BackupPath      string        `yaml:"backup_path" json:"backup_path" default:"./backups"`
	Compression     bool          `yaml:"compression" json:"compression" default:"true"`
	
	// Backup types
	FullBackup      bool `yaml:"full_backup" json:"full_backup" default:"true"`
	IncrementalBackup bool `yaml:"incremental_backup" json:"incremental_backup" default:"false"`
	
	// Storage settings
	StorageType     string `yaml:"storage_type" json:"storage_type" validate:"oneof=local s3 gcs azure" default:"local"`
	StorageConfig   map[string]interface{} `yaml:"storage_config" json:"storage_config"`
	
	// Encryption
	Encryption      bool   `yaml:"encryption" json:"encryption" default:"false"`
	EncryptionKey   string `yaml:"encryption_key" json:"encryption_key"`
	EncryptionKeyPath string `yaml:"encryption_key_path" json:"encryption_key_path"`
}

// DatabaseMonitoringConfig holds database monitoring configuration
type DatabaseMonitoringConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled" default:"true"`
	MetricsInterval time.Duration `yaml:"metrics_interval" json:"metrics_interval" default:"30s"`
	
	// Query monitoring
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold" json:"slow_query_threshold" default:"1s"`
	LogSlowQueries     bool          `yaml:"log_slow_queries" json:"log_slow_queries" default:"true"`
	
	// Connection monitoring
	MonitorConnections bool `yaml:"monitor_connections" json:"monitor_connections" default:"true"`
	ConnectionThreshold int `yaml:"connection_threshold" json:"connection_threshold" validate:"min=1" default:"20"`
	
	// Performance monitoring
	MonitorPerformance bool `yaml:"monitor_performance" json:"monitor_performance" default:"true"`
	PerformanceMetrics []string `yaml:"performance_metrics" json:"performance_metrics"`
	
	// Health check
	HealthCheck HealthCheckConfig `yaml:"health_check" json:"health_check"`
	
	// Alerting
	Alerting AlertingConfig `yaml:"alerting" json:"alerting"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled" default:"true"`
	Interval        time.Duration `yaml:"interval" json:"interval" default:"30s"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout" default:"5s"`
	FailureThreshold int          `yaml:"failure_threshold" json:"failure_threshold" validate:"min=1" default:"3"`
	SuccessThreshold int          `yaml:"success_threshold" json:"success_threshold" validate:"min=1" default:"1"`
	
	// Health check query
	Query           string `yaml:"query" json:"query" default:"SELECT 1"`
	
	// Circuit breaker integration
	CircuitBreaker  bool `yaml:"circuit_breaker" json:"circuit_breaker" default:"true"`
}

// AlertingConfig holds alerting configuration
type AlertingConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled" default:"false"`
	Channels    []string `yaml:"channels" json:"channels"` // email, slack, webhook
	
	// Alert thresholds
	ConnectionThreshold float64 `yaml:"connection_threshold" json:"connection_threshold" default:"0.8"` // 80% of max connections
	QueryTimeThreshold  time.Duration `yaml:"query_time_threshold" json:"query_time_threshold" default:"5s"`
	ErrorRateThreshold  float64 `yaml:"error_rate_threshold" json:"error_rate_threshold" default:"0.05"` // 5% error rate
	
	// Alert settings
	CooldownPeriod  time.Duration `yaml:"cooldown_period" json:"cooldown_period" default:"15m"`
	MaxAlerts       int           `yaml:"max_alerts" json:"max_alerts" validate:"min=1" default:"10"`
}

// GetConnectionString builds the database connection string
func (dc *DatabaseConfig) GetConnectionString() string {
	switch dc.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dc.Host, dc.Port, dc.Username, dc.Password, dc.Database, dc.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			dc.Username, dc.Password, dc.Host, dc.Port, dc.Database)
	case "sqlite":
		return dc.Database // For SQLite, database is the file path
	default:
		return ""
	}
}

// GetReadOnlyConnectionString builds a read-only connection string
func (dc *DatabaseConfig) GetReadOnlyConnectionString() string {
	// For most databases, this would be the same as the main connection
	// but with read-only parameters
	connStr := dc.GetConnectionString()
	
	switch dc.Driver {
	case "postgres":
		return connStr + "&default_transaction_read_only=on"
	case "mysql":
		return connStr + "&readOnly=true"
	default:
		return connStr
	}
}

// GetReplicaConnectionString builds a connection string for a specific replica
func (rc *ReplicaConfig) GetConnectionString(driver string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
			rc.Host, rc.Port, rc.Username, rc.Password, rc.Database)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&readOnly=true",
			rc.Username, rc.Password, rc.Host, rc.Port, rc.Database)
	default:
		return ""
	}
}

// IsEnabled checks if the replica is enabled and healthy
func (rc *ReplicaConfig) IsEnabled() bool {
	return rc.Enabled
}

// Validate validates the database configuration
func (dc *DatabaseConfig) Validate() error {
	if dc.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	
	if dc.Driver == "sqlite" {
		if dc.Database == "" {
			return fmt.Errorf("database file path is required for SQLite")
		}
		return nil
	}
	
	if dc.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if dc.Port <= 0 || dc.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	
	if dc.Database == "" {
		return fmt.Errorf("database name is required")
	}
	
	if dc.Username == "" {
		return fmt.Errorf("database username is required")
	}
	
	// Validate connection pool settings
	if dc.MaxOpenConns < dc.MaxIdleConns {
		return fmt.Errorf("max_open_conns must be greater than or equal to max_idle_conns")
	}
	
	// Validate read replicas
	for i, replica := range dc.ReadReplicas {
		if replica.Name == "" {
			return fmt.Errorf("replica %d: name is required", i)
		}
		if replica.Host == "" {
			return fmt.Errorf("replica %s: host is required", replica.Name)
		}
		if replica.Port <= 0 || replica.Port > 65535 {
			return fmt.Errorf("replica %s: port must be between 1 and 65535", replica.Name)
		}
	}
	
	return nil
}

// GetDefaultDatabaseConfig returns default database configuration
func GetDefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "go_coffee_ai_agents",
		Username: "postgres",
		SSLMode:  "require",
		
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 15 * time.Minute,
		QueryTimeout:    30 * time.Second,
		
		Migration: MigrationConfig{
			Enabled:         true,
			MigrationsPath:  "./migrations",
			MigrationsTable: "schema_migrations",
			AutoMigrate:     false,
			AllowDirty:      false,
			LockTimeout:     15 * time.Minute,
			EnableRollback:  true,
			MaxRollbacks:    5,
		},
		
		Backup: BackupConfig{
			Enabled:       false,
			RetentionDays: 30,
			BackupPath:    "./backups",
			Compression:   true,
			FullBackup:    true,
			StorageType:   "local",
			Encryption:    false,
		},
		
		Monitoring: DatabaseMonitoringConfig{
			Enabled:            true,
			MetricsInterval:    30 * time.Second,
			SlowQueryThreshold: time.Second,
			LogSlowQueries:     true,
			MonitorConnections: true,
			ConnectionThreshold: 20,
			MonitorPerformance: true,
			
			HealthCheck: HealthCheckConfig{
				Enabled:          true,
				Interval:         30 * time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
				SuccessThreshold: 1,
				Query:            "SELECT 1",
				CircuitBreaker:   true,
			},
			
			Alerting: AlertingConfig{
				Enabled:             false,
				ConnectionThreshold: 0.8,
				QueryTimeThreshold:  5 * time.Second,
				ErrorRateThreshold:  0.05,
				CooldownPeriod:      15 * time.Minute,
				MaxAlerts:           10,
			},
		},
	}
}
