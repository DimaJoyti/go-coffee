package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration
type Config struct {
	LogLevel   string           `mapstructure:"log_level"`
	ConfigFile string           `mapstructure:"config_file"`
	Telemetry  TelemetryConfig  `mapstructure:"telemetry"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	Cloud      CloudConfig      `mapstructure:"cloud"`
	Services   ServicesConfig   `mapstructure:"services"`
	Security   SecurityConfig   `mapstructure:"security"`
	GitOps     GitOpsConfig     `mapstructure:"gitops"`
}

// TelemetryConfig holds OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled     bool              `mapstructure:"enabled"`
	ServiceName string            `mapstructure:"service_name"`
	Endpoint    string            `mapstructure:"endpoint"`
	Headers     map[string]string `mapstructure:"headers"`
}

// KubernetesConfig holds Kubernetes-related configuration
type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	Context    string `mapstructure:"context"`
	Namespace  string `mapstructure:"namespace"`
	Timeout    string `mapstructure:"timeout"`
}

// CloudConfig holds multi-cloud provider configuration
type CloudConfig struct {
	Provider    string                 `mapstructure:"provider"`
	Region      string                 `mapstructure:"region"`
	Project     string                 `mapstructure:"project"`
	Settings    map[string]interface{} `mapstructure:"settings"`
	MultiCloud  MultiCloudConfig       `mapstructure:"multi_cloud"`
	EdgeNodes   EdgeConfig             `mapstructure:"edge_nodes"`
	CostControl CostControlConfig      `mapstructure:"cost_control"`
}

// MultiCloudConfig holds configuration for multiple cloud providers
type MultiCloudConfig struct {
	Enabled   bool                           `mapstructure:"enabled"`
	Primary   string                         `mapstructure:"primary"`
	Secondary string                         `mapstructure:"secondary"`
	Providers map[string]CloudProviderConfig `mapstructure:"providers"`
	Strategy  string                         `mapstructure:"strategy"` // active-passive, active-active, burst
}

// CloudProviderConfig holds provider-specific configuration
type CloudProviderConfig struct {
	Enabled     bool                   `mapstructure:"enabled"`
	Region      string                 `mapstructure:"region"`
	Zones       []string               `mapstructure:"zones"`
	Credentials CredentialsConfig      `mapstructure:"credentials"`
	Networking  NetworkingConfig       `mapstructure:"networking"`
	Compute     ComputeConfig          `mapstructure:"compute"`
	Storage     StorageConfig          `mapstructure:"storage"`
	Settings    map[string]interface{} `mapstructure:"settings"`
}

// CredentialsConfig holds authentication configuration
type CredentialsConfig struct {
	Type        string            `mapstructure:"type"` // service-account, access-key, managed-identity
	File        string            `mapstructure:"file"`
	Environment map[string]string `mapstructure:"environment"`
	Vault       VaultConfig       `mapstructure:"vault"`
}

// VaultConfig holds HashiCorp Vault configuration
type VaultConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Address   string `mapstructure:"address"`
	Token     string `mapstructure:"token"`
	Path      string `mapstructure:"path"`
	Role      string `mapstructure:"role"`
	Namespace string `mapstructure:"namespace"`
}

// NetworkingConfig holds networking configuration
type NetworkingConfig struct {
	VPC            string             `mapstructure:"vpc"`
	Subnets        []string           `mapstructure:"subnets"`
	SecurityGroups []string           `mapstructure:"security_groups"`
	LoadBalancer   LoadBalancerConfig `mapstructure:"load_balancer"`
	CDN            CDNConfig          `mapstructure:"cdn"`
}

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	Type        string            `mapstructure:"type"`   // application, network, classic
	Scheme      string            `mapstructure:"scheme"` // internet-facing, internal
	Listeners   []ListenerConfig  `mapstructure:"listeners"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
}

// ListenerConfig holds listener configuration
type ListenerConfig struct {
	Port     int    `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
	SSL      bool   `mapstructure:"ssl"`
	CertArn  string `mapstructure:"cert_arn"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Path               string `mapstructure:"path"`
	Port               int    `mapstructure:"port"`
	Protocol           string `mapstructure:"protocol"`
	HealthyThreshold   int    `mapstructure:"healthy_threshold"`
	UnhealthyThreshold int    `mapstructure:"unhealthy_threshold"`
	Timeout            int    `mapstructure:"timeout"`
	Interval           int    `mapstructure:"interval"`
}

// CDNConfig holds CDN configuration
type CDNConfig struct {
	Enabled      bool                   `mapstructure:"enabled"`
	Provider     string                 `mapstructure:"provider"` // cloudflare, cloudfront, azure-cdn
	Distribution string                 `mapstructure:"distribution"`
	Origins      []OriginConfig         `mapstructure:"origins"`
	Behaviors    []BehaviorConfig       `mapstructure:"behaviors"`
	Settings     map[string]interface{} `mapstructure:"settings"`
}

// OriginConfig holds CDN origin configuration
type OriginConfig struct {
	Name   string `mapstructure:"name"`
	Domain string `mapstructure:"domain"`
	Path   string `mapstructure:"path"`
	HTTPS  bool   `mapstructure:"https"`
}

// BehaviorConfig holds CDN behavior configuration
type BehaviorConfig struct {
	PathPattern string            `mapstructure:"path_pattern"`
	TTL         int               `mapstructure:"ttl"`
	Compress    bool              `mapstructure:"compress"`
	Headers     map[string]string `mapstructure:"headers"`
}

// ComputeConfig holds compute configuration
type ComputeConfig struct {
	InstanceTypes    []string          `mapstructure:"instance_types"`
	AutoScaling      AutoScalingConfig `mapstructure:"auto_scaling"`
	SpotInstances    SpotConfig        `mapstructure:"spot_instances"`
	ContainerRuntime string            `mapstructure:"container_runtime"`
	GPU              GPUConfig         `mapstructure:"gpu"`
}

// AutoScalingConfig holds auto-scaling configuration
type AutoScalingConfig struct {
	Enabled  bool             `mapstructure:"enabled"`
	MinNodes int              `mapstructure:"min_nodes"`
	MaxNodes int              `mapstructure:"max_nodes"`
	Metrics  []MetricConfig   `mapstructure:"metrics"`
	Policies []PolicyConfig   `mapstructure:"policies"`
	Schedule []ScheduleConfig `mapstructure:"schedule"`
}

// MetricConfig holds scaling metric configuration
type MetricConfig struct {
	Name      string  `mapstructure:"name"`
	Target    float64 `mapstructure:"target"`
	Type      string  `mapstructure:"type"` // cpu, memory, custom
	Threshold float64 `mapstructure:"threshold"`
}

// PolicyConfig holds scaling policy configuration
type PolicyConfig struct {
	Name       string `mapstructure:"name"`
	Type       string `mapstructure:"type"` // scale-up, scale-down
	Adjustment int    `mapstructure:"adjustment"`
	Cooldown   int    `mapstructure:"cooldown"`
	MetricName string `mapstructure:"metric_name"`
}

// ScheduleConfig holds scheduled scaling configuration
type ScheduleConfig struct {
	Name     string `mapstructure:"name"`
	Cron     string `mapstructure:"cron"`
	MinNodes int    `mapstructure:"min_nodes"`
	MaxNodes int    `mapstructure:"max_nodes"`
	Timezone string `mapstructure:"timezone"`
}

// SpotConfig holds spot instance configuration
type SpotConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	MaxPrice         float64 `mapstructure:"max_price"`
	Percentage       int     `mapstructure:"percentage"`
	DiversifyAZs     bool    `mapstructure:"diversify_azs"`
	FallbackOnDemand bool    `mapstructure:"fallback_on_demand"`
}

// GPUConfig holds GPU configuration
type GPUConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	Types      []string `mapstructure:"types"` // nvidia-tesla-t4, nvidia-tesla-v100, nvidia-a100
	Sharing    bool     `mapstructure:"sharing"`
	Monitoring bool     `mapstructure:"monitoring"`
	Drivers    string   `mapstructure:"drivers"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Classes     []StorageClassConfig `mapstructure:"classes"`
	Backup      BackupConfig         `mapstructure:"backup"`
	Encryption  EncryptionConfig     `mapstructure:"encryption"`
	Replication ReplicationConfig    `mapstructure:"replication"`
}

// StorageClassConfig holds storage class configuration
type StorageClassConfig struct {
	Name       string            `mapstructure:"name"`
	Type       string            `mapstructure:"type"` // ssd, hdd, nvme
	IOPS       int               `mapstructure:"iops"`
	Throughput int               `mapstructure:"throughput"`
	Encrypted  bool              `mapstructure:"encrypted"`
	Parameters map[string]string `mapstructure:"parameters"`
}

// BackupConfig holds backup configuration
type BackupConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	Schedule      string   `mapstructure:"schedule"`
	Retention     int      `mapstructure:"retention"`
	CrossRegion   bool     `mapstructure:"cross_region"`
	Encryption    bool     `mapstructure:"encryption"`
	Destinations  []string `mapstructure:"destinations"`
	Notifications []string `mapstructure:"notifications"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Algorithm string `mapstructure:"algorithm"`
	KeySource string `mapstructure:"key_source"` // cloud-kms, vault, local
	KeyID     string `mapstructure:"key_id"`
	Rotation  bool   `mapstructure:"rotation"`
}

// ReplicationConfig holds replication configuration
type ReplicationConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Type        string   `mapstructure:"type"` // sync, async
	Regions     []string `mapstructure:"regions"`
	Consistency string   `mapstructure:"consistency"` // strong, eventual
}

// EdgeConfig holds edge computing configuration
type EdgeConfig struct {
	Enabled    bool                 `mapstructure:"enabled"`
	Provider   string               `mapstructure:"provider"` // aws-wavelength, azure-edge, gcp-edge
	Locations  []EdgeLocationConfig `mapstructure:"locations"`
	Workloads  []EdgeWorkloadConfig `mapstructure:"workloads"`
	Networking EdgeNetworkingConfig `mapstructure:"networking"`
}

// EdgeLocationConfig holds edge location configuration
type EdgeLocationConfig struct {
	Name     string         `mapstructure:"name"`
	Region   string         `mapstructure:"region"`
	Zone     string         `mapstructure:"zone"`
	Capacity CapacityConfig `mapstructure:"capacity"`
	Latency  LatencyConfig  `mapstructure:"latency"`
	Services []string       `mapstructure:"services"`
}

// CapacityConfig holds capacity configuration
type CapacityConfig struct {
	CPU     string `mapstructure:"cpu"`
	Memory  string `mapstructure:"memory"`
	Storage string `mapstructure:"storage"`
	Network string `mapstructure:"network"`
}

// LatencyConfig holds latency requirements
type LatencyConfig struct {
	Target     int  `mapstructure:"target"` // milliseconds
	SLA        int  `mapstructure:"sla"`    // milliseconds
	Monitoring bool `mapstructure:"monitoring"`
}

// EdgeWorkloadConfig holds edge workload configuration
type EdgeWorkloadConfig struct {
	Name      string         `mapstructure:"name"`
	Type      string         `mapstructure:"type"` // cdn, compute, ai-inference
	Replicas  int            `mapstructure:"replicas"`
	Resources ResourceConfig `mapstructure:"resources"`
	Affinity  AffinityConfig `mapstructure:"affinity"`
}

// ResourceConfig holds resource requirements
type ResourceConfig struct {
	CPU     string `mapstructure:"cpu"`
	Memory  string `mapstructure:"memory"`
	Storage string `mapstructure:"storage"`
	GPU     int    `mapstructure:"gpu"`
}

// AffinityConfig holds affinity rules
type AffinityConfig struct {
	NodeAffinity []NodeAffinityRule `mapstructure:"node_affinity"`
	PodAffinity  []PodAffinityRule  `mapstructure:"pod_affinity"`
}

// NodeAffinityRule holds node affinity rule
type NodeAffinityRule struct {
	Key      string   `mapstructure:"key"`
	Operator string   `mapstructure:"operator"`
	Values   []string `mapstructure:"values"`
	Weight   int      `mapstructure:"weight"`
}

// PodAffinityRule holds pod affinity rule
type PodAffinityRule struct {
	LabelSelector map[string]string `mapstructure:"label_selector"`
	Topology      string            `mapstructure:"topology"`
	Weight        int               `mapstructure:"weight"`
}

// EdgeNetworkingConfig holds edge networking configuration
type EdgeNetworkingConfig struct {
	CDN          bool           `mapstructure:"cdn"`
	LoadBalancer bool           `mapstructure:"load_balancer"`
	Mesh         bool           `mapstructure:"mesh"`
	Security     SecurityConfig `mapstructure:"security"`
}

// CostControlConfig holds cost control configuration
type CostControlConfig struct {
	Enabled      bool               `mapstructure:"enabled"`
	Budget       BudgetConfig       `mapstructure:"budget"`
	Alerts       []CostAlertConfig  `mapstructure:"alerts"`
	Policies     []CostPolicyConfig `mapstructure:"policies"`
	Optimization OptimizationConfig `mapstructure:"optimization"`
	Reporting    ReportingConfig    `mapstructure:"reporting"`
}

// BudgetConfig holds budget configuration
type BudgetConfig struct {
	Monthly   float64            `mapstructure:"monthly"`
	Quarterly float64            `mapstructure:"quarterly"`
	Annual    float64            `mapstructure:"annual"`
	Currency  string             `mapstructure:"currency"`
	Breakdown map[string]float64 `mapstructure:"breakdown"`
}

// CostAlertConfig holds cost alert configuration
type CostAlertConfig struct {
	Name       string   `mapstructure:"name"`
	Threshold  float64  `mapstructure:"threshold"`
	Type       string   `mapstructure:"type"`   // absolute, percentage
	Period     string   `mapstructure:"period"` // daily, weekly, monthly
	Recipients []string `mapstructure:"recipients"`
	Actions    []string `mapstructure:"actions"`
}

// CostPolicyConfig holds cost policy configuration
type CostPolicyConfig struct {
	Name        string           `mapstructure:"name"`
	Rules       []CostRuleConfig `mapstructure:"rules"`
	Actions     []string         `mapstructure:"actions"`
	Enforcement string           `mapstructure:"enforcement"` // warn, block, auto-remediate
}

// CostRuleConfig holds cost rule configuration
type CostRuleConfig struct {
	Resource  string  `mapstructure:"resource"`
	Condition string  `mapstructure:"condition"`
	Value     float64 `mapstructure:"value"`
	Period    string  `mapstructure:"period"`
}

// OptimizationConfig holds cost optimization configuration
type OptimizationConfig struct {
	Enabled           bool                      `mapstructure:"enabled"`
	RightSizing       RightSizingConfig         `mapstructure:"right_sizing"`
	ReservedInstances ReservedInstancesConfig   `mapstructure:"reserved_instances"`
	SpotInstances     SpotOptimizationConfig    `mapstructure:"spot_instances"`
	Storage           StorageOptimizationConfig `mapstructure:"storage"`
}

// RightSizingConfig holds right-sizing configuration
type RightSizingConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	Threshold float64 `mapstructure:"threshold"`
	Period    string  `mapstructure:"period"`
	AutoApply bool    `mapstructure:"auto_apply"`
}

// ReservedInstancesConfig holds reserved instances configuration
type ReservedInstancesConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	Utilization   float64 `mapstructure:"utilization"`
	Term          string  `mapstructure:"term"`           // 1year, 3year
	PaymentOption string  `mapstructure:"payment_option"` // no-upfront, partial-upfront, all-upfront
}

// SpotOptimizationConfig holds spot optimization configuration
type SpotOptimizationConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	MaxPrice  float64 `mapstructure:"max_price"`
	Diversify bool    `mapstructure:"diversify"`
	Fallback  bool    `mapstructure:"fallback"`
}

// StorageOptimizationConfig holds storage optimization configuration
type StorageOptimizationConfig struct {
	Enabled       bool            `mapstructure:"enabled"`
	Lifecycle     LifecycleConfig `mapstructure:"lifecycle"`
	Compression   bool            `mapstructure:"compression"`
	Deduplication bool            `mapstructure:"deduplication"`
}

// LifecycleConfig holds lifecycle configuration
type LifecycleConfig struct {
	Rules []LifecycleRuleConfig `mapstructure:"rules"`
}

// LifecycleRuleConfig holds lifecycle rule configuration
type LifecycleRuleConfig struct {
	Name         string `mapstructure:"name"`
	Prefix       string `mapstructure:"prefix"`
	Days         int    `mapstructure:"days"`
	StorageClass string `mapstructure:"storage_class"`
	Action       string `mapstructure:"action"` // transition, delete
}

// ReportingConfig holds cost reporting configuration
type ReportingConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Schedule    string   `mapstructure:"schedule"`
	Recipients  []string `mapstructure:"recipients"`
	Format      string   `mapstructure:"format"`      // json, csv, pdf
	Granularity string   `mapstructure:"granularity"` // daily, weekly, monthly
}

// ServicesConfig holds service-related configuration
type ServicesConfig struct {
	DefaultPort     int               `mapstructure:"default_port"`
	HealthCheckPath string            `mapstructure:"health_check_path"`
	MetricsPath     string            `mapstructure:"metrics_path"`
	Endpoints       map[string]string `mapstructure:"endpoints"`
	Database        DatabaseConfig    `mapstructure:"database"`
	Cache           CacheConfig       `mapstructure:"cache"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	TLSEnabled    bool   `mapstructure:"tls_enabled"`
	CertPath      string `mapstructure:"cert_path"`
	KeyPath       string `mapstructure:"key_path"`
	CAPath        string `mapstructure:"ca_path"`
	TokenPath     string `mapstructure:"token_path"`
	PolicyEnabled bool   `mapstructure:"policy_enabled"`
}

// GitOpsConfig holds GitOps-related configuration
type GitOpsConfig struct {
	Provider   string `mapstructure:"provider"`
	Repository string `mapstructure:"repository"`
	Branch     string `mapstructure:"branch"`
	Path       string `mapstructure:"path"`
	Token      string `mapstructure:"token"`
}

// Load loads configuration from various sources
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config name and paths
	v.SetConfigName("gocoffee")
	v.SetConfigType("yaml")

	// Add config paths
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.gocoffee")
	v.AddConfigPath("/etc/gocoffee")

	// Environment variables
	v.SetEnvPrefix("GOCOFFEE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// General defaults
	v.SetDefault("log_level", "info")

	// Telemetry defaults
	v.SetDefault("telemetry.enabled", true)
	v.SetDefault("telemetry.service_name", "gocoffee")
	v.SetDefault("telemetry.endpoint", "")

	// Kubernetes defaults
	v.SetDefault("kubernetes.config_path", getDefaultKubeConfig())
	v.SetDefault("kubernetes.context", "")
	v.SetDefault("kubernetes.namespace", "default")
	v.SetDefault("kubernetes.timeout", "30s")

	// Cloud defaults
	v.SetDefault("cloud.provider", "gcp")
	v.SetDefault("cloud.region", "us-central1")

	// Services defaults
	v.SetDefault("services.default_port", 8080)
	v.SetDefault("services.health_check_path", "/health")
	v.SetDefault("services.metrics_path", "/metrics")

	// Security defaults
	v.SetDefault("security.tls_enabled", true)
	v.SetDefault("security.policy_enabled", true)

	// GitOps defaults
	v.SetDefault("gitops.provider", "github")
	v.SetDefault("gitops.branch", "main")
	v.SetDefault("gitops.path", "deployments")
}

// getDefaultKubeConfig returns the default kubeconfig path
func getDefaultKubeConfig() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Convert config to map
	configMap := map[string]interface{}{
		"log_level":  c.LogLevel,
		"telemetry":  c.Telemetry,
		"kubernetes": c.Kubernetes,
		"cloud":      c.Cloud,
		"services":   c.Services,
		"security":   c.Security,
		"gitops":     c.GitOps,
	}

	for key, value := range configMap {
		v.Set(key, value)
	}

	return v.WriteConfig()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.LogLevel == "" {
		return fmt.Errorf("log_level is required")
	}

	if c.Telemetry.ServiceName == "" {
		return fmt.Errorf("telemetry.service_name is required")
	}

	if c.Kubernetes.ConfigPath == "" {
		return fmt.Errorf("kubernetes.config_path is required")
	}

	return nil
}
