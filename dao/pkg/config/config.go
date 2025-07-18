package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// BigUint64 is a custom type that can unmarshal large numbers from strings
type BigUint64 uint64

// UnmarshalYAML implements yaml.Unmarshaler for BigUint64
func (b *BigUint64) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		// Try to decode as uint64 directly
		var num uint64
		if err := value.Decode(&num); err != nil {
			return err
		}
		*b = BigUint64(num)
		return nil
	}

	num, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	*b = BigUint64(num)
	return nil
}

// Config represents the Developer DAO configuration
type Config struct {
	Environment string `yaml:"environment"`

	// Server configuration
	Server ServerConfig `yaml:"server"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// Redis configuration
	Redis RedisConfig `yaml:"redis"`

	// Blockchain configuration
	Blockchain BlockchainConfig `yaml:"blockchain"`

	// Smart contract addresses
	Contracts ContractsConfig `yaml:"contracts"`

	// DAO configuration
	DAO DAOConfig `yaml:"dao"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics"`

	// Revenue sharing configuration
	RevenueSharing RevenueSharingConfig `yaml:"revenue_sharing"`

	// API configuration
	API APIConfig `yaml:"api"`

	// gRPC configuration
	GRPC GRPCConfig `yaml:"grpc"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`

	// Monitoring configuration
	Monitoring MonitoringConfig `yaml:"monitoring"`

	// Security configuration
	Security SecurityConfig `yaml:"security"`

	// Integration configuration
	Integrations IntegrationsConfig `yaml:"integrations"`

	// Feature flags
	Features FeaturesConfig `yaml:"features"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	MigrationPath   string        `yaml:"migration_path"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	PoolTimeout  time.Duration `yaml:"pool_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// BlockchainConfig represents blockchain configuration
type BlockchainConfig struct {
	Ethereum EthereumConfig `yaml:"ethereum"`
	BSC      BSCConfig      `yaml:"bsc"`
	Polygon  PolygonConfig  `yaml:"polygon"`
}

// EthereumConfig represents Ethereum configuration
type EthereumConfig struct {
	RPCURL             string `yaml:"rpc_url"`
	ChainID            int    `yaml:"chain_id"`
	GasLimit           uint64 `yaml:"gas_limit"`
	GasPrice           uint64 `yaml:"gas_price"`
	ConfirmationBlocks int    `yaml:"confirmation_blocks"`
}

// BSCConfig represents BSC configuration
type BSCConfig struct {
	RPCURL             string `yaml:"rpc_url"`
	ChainID            int    `yaml:"chain_id"`
	GasLimit           uint64 `yaml:"gas_limit"`
	GasPrice           uint64 `yaml:"gas_price"`
	ConfirmationBlocks int    `yaml:"confirmation_blocks"`
}

// PolygonConfig represents Polygon configuration
type PolygonConfig struct {
	RPCURL             string `yaml:"rpc_url"`
	ChainID            int    `yaml:"chain_id"`
	GasLimit           uint64 `yaml:"gas_limit"`
	GasPrice           uint64 `yaml:"gas_price"`
	ConfirmationBlocks int    `yaml:"confirmation_blocks"`
}

// ContractsConfig represents smart contract addresses
type ContractsConfig struct {
	CoffeeToken        string `yaml:"coffee_token"`
	DAOGovernor        string `yaml:"dao_governor"`
	BountyManager      string `yaml:"bounty_manager"`
	RevenueSharing     string `yaml:"revenue_sharing"`
	SolutionRegistry   string `yaml:"solution_registry"`
	TimelockController string `yaml:"timelock_controller"`
}

// DAOConfig represents DAO configuration
type DAOConfig struct {
	VotingDelay       int       `yaml:"voting_delay"`
	VotingPeriod      int       `yaml:"voting_period"`
	ProposalThreshold BigUint64 `yaml:"proposal_threshold"`
	QuorumPercentage  int       `yaml:"quorum_percentage"`
	TimelockDelay     int       `yaml:"timelock_delay"`
	MinBountyReward   BigUint64 `yaml:"min_bounty_reward"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	OracleAddress             string        `yaml:"oracle_address"`
	UpdateInterval            time.Duration `yaml:"update_interval"`
	TVLThreshold              BigUint64     `yaml:"tvl_threshold"`
	MAUThreshold              int           `yaml:"mau_threshold"`
	PerformanceBonusThreshold float64       `yaml:"performance_bonus_threshold"`
}

// RevenueSharingConfig represents revenue sharing configuration
type RevenueSharingConfig struct {
	DeveloperShareBPS     int           `yaml:"developer_share_bps"`
	CommunityShareBPS     int           `yaml:"community_share_bps"`
	TreasuryShareBPS      int           `yaml:"treasury_share_bps"`
	DistributionInterval  time.Duration `yaml:"distribution_interval"`
	MinDistributionAmount BigUint64     `yaml:"min_distribution_amount"`
}

// APIConfig represents API configuration
type APIConfig struct {
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
}

// GRPCConfig represents gRPC configuration
type GRPCConfig struct {
	Port              int             `yaml:"port"`
	MaxRecvMsgSize    int             `yaml:"max_recv_msg_size"`
	MaxSendMsgSize    int             `yaml:"max_send_msg_size"`
	ConnectionTimeout time.Duration   `yaml:"connection_timeout"`
	Keepalive         KeepaliveConfig `yaml:"keepalive"`
}

// KeepaliveConfig represents keepalive configuration
type KeepaliveConfig struct {
	Time                time.Duration `yaml:"time"`
	Timeout             time.Duration `yaml:"timeout"`
	PermitWithoutStream bool          `yaml:"permit_without_stream"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Prometheus  PrometheusConfig  `yaml:"prometheus"`
	Jaeger      JaegerConfig      `yaml:"jaeger"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// JaegerConfig represents Jaeger configuration
type JaegerConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Endpoint    string `yaml:"endpoint"`
	ServiceName string `yaml:"service_name"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWT          JWTConfig          `yaml:"jwt"`
	RateLimiting RateLimitingConfig `yaml:"rate_limiting"`
	CORS         CORSSecurityConfig `yaml:"cors"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	Expiration time.Duration `yaml:"expiration"`
	Issuer     string        `yaml:"issuer"`
}

// RateLimitingConfig represents rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// CORSSecurityConfig represents CORS security configuration
type CORSSecurityConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// IntegrationsConfig represents integration configuration
type IntegrationsConfig struct {
	ExistingServices ExistingServicesConfig `yaml:"existing_services"`
}

// ExistingServicesConfig represents existing services configuration
type ExistingServicesConfig struct {
	DeFiService DeFiServiceConfig `yaml:"defi_service"`
	AIAgents    AIAgentsConfig    `yaml:"ai_agents"`
	APIGateway  APIGatewayConfig  `yaml:"api_gateway"`
}

// DeFiServiceConfig represents DeFi service configuration
type DeFiServiceConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// AIAgentsConfig represents AI agents configuration
type AIAgentsConfig struct {
	OrchestratorHost string        `yaml:"orchestrator_host"`
	OrchestratorPort int           `yaml:"orchestrator_port"`
	Timeout          time.Duration `yaml:"timeout"`
}

// APIGatewayConfig represents API gateway configuration
type APIGatewayConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	BountySystem        bool `yaml:"bounty_system"`
	RevenueSharing      bool `yaml:"revenue_sharing"`
	SolutionMarketplace bool `yaml:"solution_marketplace"`
	AIIntegration       bool `yaml:"ai_integration"`
	PerformanceTracking bool `yaml:"performance_tracking"`
	GovernanceVoting    bool `yaml:"governance_voting"`
}

// Load loads configuration from a YAML file
func Load(configPath string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	setDefaults(&config)

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8090
	}

	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}

	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}

	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}

	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
}
