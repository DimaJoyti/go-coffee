package metrics

import (
	"time"

	"github.com/shopspring/decimal"
)

// MetricType represents the type of metric
type MetricType int

const (
	MetricTypeTVL MetricType = iota
	MetricTypeMAU
	MetricTypeRevenue
	MetricTypeTransactions
	MetricTypeUsers
)

func (mt MetricType) String() string {
	switch mt {
	case MetricTypeTVL:
		return "TVL"
	case MetricTypeMAU:
		return "MAU"
	case MetricTypeRevenue:
		return "REVENUE"
	case MetricTypeTransactions:
		return "TRANSACTIONS"
	case MetricTypeUsers:
		return "USERS"
	default:
		return "UNKNOWN"
	}
}

// AlertType represents the type of alert
type AlertType int

const (
	AlertTypeThreshold AlertType = iota
	AlertTypeGrowthRate
	AlertTypeAnomaly
	AlertTypeDowntime
)

func (at AlertType) String() string {
	switch at {
	case AlertTypeThreshold:
		return "THRESHOLD"
	case AlertTypeGrowthRate:
		return "GROWTH_RATE"
	case AlertTypeAnomaly:
		return "ANOMALY"
	case AlertTypeDowntime:
		return "DOWNTIME"
	default:
		return "UNKNOWN"
	}
}

// AlertStatus represents the status of an alert
type AlertStatus int

const (
	AlertStatusActive AlertStatus = iota
	AlertStatusResolved
	AlertStatusSuppressed
)

func (as AlertStatus) String() string {
	switch as {
	case AlertStatusActive:
		return "ACTIVE"
	case AlertStatusResolved:
		return "RESOLVED"
	case AlertStatusSuppressed:
		return "SUPPRESSED"
	default:
		return "UNKNOWN"
	}
}

// TVLRecord represents a Total Value Locked measurement
type TVLRecord struct {
	ID          int64           `json:"id" db:"id"`
	Protocol    string          `json:"protocol" db:"protocol"`
	Chain       string          `json:"chain" db:"chain"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	TokenSymbol string          `json:"token_symbol" db:"token_symbol"`
	Source      string          `json:"source" db:"source"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
	BlockNumber *int64          `json:"block_number" db:"block_number"`
	TxHash      string          `json:"tx_hash" db:"tx_hash"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
}

// MAURecord represents a Monthly Active Users measurement
type MAURecord struct {
	ID          int64             `json:"id" db:"id"`
	Feature     string            `json:"feature" db:"feature"`
	UserCount   int               `json:"user_count" db:"user_count"`
	UniqueUsers int               `json:"unique_users" db:"unique_users"`
	Period      string            `json:"period" db:"period"`
	Source      string            `json:"source" db:"source"`
	Timestamp   time.Time         `json:"timestamp" db:"timestamp"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
}

// ImpactRecord represents developer/solution impact metrics
type ImpactRecord struct {
	ID          int64             `json:"id" db:"id"`
	EntityID    string            `json:"entity_id" db:"entity_id"`
	EntityType  string            `json:"entity_type" db:"entity_type"`
	TVLImpact   decimal.Decimal   `json:"tvl_impact" db:"tvl_impact"`
	MAUImpact   int               `json:"mau_impact" db:"mau_impact"`
	Attribution decimal.Decimal   `json:"attribution" db:"attribution"`
	Source      string            `json:"source" db:"source"`
	Timestamp   time.Time         `json:"timestamp" db:"timestamp"`
	Verified    bool              `json:"verified" db:"verified"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
}

// Alert represents a metrics alert
type Alert struct {
	ID          int64             `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Type        AlertType         `json:"type" db:"type"`
	MetricType  MetricType        `json:"metric_type" db:"metric_type"`
	Threshold   decimal.Decimal   `json:"threshold" db:"threshold"`
	Condition   string            `json:"condition" db:"condition"`
	Status      AlertStatus       `json:"status" db:"status"`
	Message     string            `json:"message" db:"message"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	TriggeredAt *time.Time        `json:"triggered_at" db:"triggered_at"`
	ResolvedAt  *time.Time        `json:"resolved_at" db:"resolved_at"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
}

// Report represents a generated metrics report
type Report struct {
	ID          int64             `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Type        string            `json:"type" db:"type"`
	Period      string            `json:"period" db:"period"`
	Data        map[string]interface{} `json:"data" db:"data"`
	GeneratedAt time.Time         `json:"generated_at" db:"generated_at"`
	CreatedBy   string            `json:"created_by" db:"created_by"`
	Metadata    map[string]string `json:"metadata" db:"metadata"`
}

// DataSource represents an external data source
type DataSource struct {
	ID          int64             `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Type        string            `json:"type" db:"type"`
	URL         string            `json:"url" db:"url"`
	APIKey      string            `json:"api_key" db:"api_key"`
	IsActive    bool              `json:"is_active" db:"is_active"`
	LastSync    *time.Time        `json:"last_sync" db:"last_sync"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	Config      map[string]string `json:"config" db:"config"`
}

// MetricsSnapshot represents a cached metrics snapshot
type MetricsSnapshot struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// AggregatedMetrics represents aggregated metrics data
type AggregatedMetrics struct {
	Type        MetricType        `json:"type"`
	Period      string            `json:"period"`
	Value       decimal.Decimal   `json:"value"`
	Growth      decimal.Decimal   `json:"growth"`
	Timestamp   time.Time         `json:"timestamp"`
	Breakdown   map[string]decimal.Decimal `json:"breakdown"`
}

// Request/Response types for API

// RecordTVLRequest represents a request to record TVL
type RecordTVLRequest struct {
	Protocol    string          `json:"protocol" binding:"required"`
	Chain       string          `json:"chain" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	TokenSymbol string          `json:"token_symbol"`
	Source      string          `json:"source" binding:"required"`
	BlockNumber *int64          `json:"block_number"`
	TxHash      string          `json:"tx_hash"`
	Metadata    map[string]string `json:"metadata"`
}

// RecordTVLResponse represents the response after recording TVL
type RecordTVLResponse struct {
	RecordID  int64     `json:"record_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// RecordMAURequest represents a request to record MAU
type RecordMAURequest struct {
	Feature     string            `json:"feature" binding:"required"`
	UserCount   int               `json:"user_count" binding:"required"`
	UniqueUsers int               `json:"unique_users"`
	Period      string            `json:"period" binding:"required"`
	Source      string            `json:"source" binding:"required"`
	Metadata    map[string]string `json:"metadata"`
}

// RecordMAUResponse represents the response after recording MAU
type RecordMAUResponse struct {
	RecordID  int64     `json:"record_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// RecordImpactRequest represents a request to record impact
type RecordImpactRequest struct {
	EntityID    string            `json:"entity_id" binding:"required"`
	EntityType  string            `json:"entity_type" binding:"required"`
	TVLImpact   decimal.Decimal   `json:"tvl_impact"`
	MAUImpact   int               `json:"mau_impact"`
	Attribution decimal.Decimal   `json:"attribution"`
	Source      string            `json:"source" binding:"required"`
	Metadata    map[string]string `json:"metadata"`
}

// RecordImpactResponse represents the response after recording impact
type RecordImpactResponse struct {
	RecordID  int64     `json:"record_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// GetTVLMetricsRequest represents a request to get TVL metrics
type GetTVLMetricsRequest struct {
	Protocol string `json:"protocol"`
	Chain    string `json:"chain"`
	Period   string `json:"period"`
}

// TVLMetricsResponse represents TVL metrics response
type TVLMetricsResponse struct {
	CurrentTVL decimal.Decimal `json:"current_tvl"`
	Growth24h  decimal.Decimal `json:"growth_24h"`
	Growth7d   decimal.Decimal `json:"growth_7d"`
	Growth30d  decimal.Decimal `json:"growth_30d"`
	Timestamp  time.Time       `json:"timestamp"`
	Breakdown  map[string]decimal.Decimal `json:"breakdown"`
}

// GetMAUMetricsRequest represents a request to get MAU metrics
type GetMAUMetricsRequest struct {
	Feature string `json:"feature"`
	Period  string `json:"period"`
}

// MAUMetricsResponse represents MAU metrics response
type MAUMetricsResponse struct {
	CurrentMAU int             `json:"current_mau"`
	Growth30d  decimal.Decimal `json:"growth_30d"`
	Growth90d  decimal.Decimal `json:"growth_90d"`
	Retention  decimal.Decimal `json:"retention"`
	Timestamp  time.Time       `json:"timestamp"`
	Breakdown  map[string]int  `json:"breakdown"`
}

// PerformanceDashboard represents performance dashboard data
type PerformanceDashboard struct {
	TVLMetrics         *TVLMetricsResponse    `json:"tvl_metrics"`
	MAUMetrics         *MAUMetricsResponse    `json:"mau_metrics"`
	TopContributors    []ImpactLeaderboard    `json:"top_contributors"`
	RecentAlerts       []Alert                `json:"recent_alerts"`
	TrendingProtocols  []ProtocolTrend        `json:"trending_protocols"`
	LastUpdated        time.Time              `json:"last_updated"`
}

// ImpactLeaderboard represents impact leaderboard entry
type ImpactLeaderboard struct {
	EntityID    string          `json:"entity_id"`
	EntityType  string          `json:"entity_type"`
	Name        string          `json:"name"`
	TVLImpact   decimal.Decimal `json:"tvl_impact"`
	MAUImpact   int             `json:"mau_impact"`
	TotalScore  decimal.Decimal `json:"total_score"`
	Rank        int             `json:"rank"`
}

// ProtocolTrend represents protocol trend data
type ProtocolTrend struct {
	Protocol    string          `json:"protocol"`
	Chain       string          `json:"chain"`
	CurrentTVL  decimal.Decimal `json:"current_tvl"`
	Growth24h   decimal.Decimal `json:"growth_24h"`
	Growth7d    decimal.Decimal `json:"growth_7d"`
	TrendScore  decimal.Decimal `json:"trend_score"`
}

// AttributionAnalysis represents attribution analysis data
type AttributionAnalysis struct {
	EntityID       string                    `json:"entity_id"`
	EntityType     string                    `json:"entity_type"`
	TotalImpact    decimal.Decimal           `json:"total_impact"`
	Attribution    map[string]decimal.Decimal `json:"attribution"`
	Confidence     decimal.Decimal           `json:"confidence"`
	Period         string                    `json:"period"`
	LastCalculated time.Time                 `json:"last_calculated"`
}

// TrendsAnalysis represents trends analysis data
type TrendsAnalysis struct {
	MetricType     MetricType               `json:"metric_type"`
	Period         string                   `json:"period"`
	Trend          string                   `json:"trend"`
	GrowthRate     decimal.Decimal          `json:"growth_rate"`
	Seasonality    map[string]decimal.Decimal `json:"seasonality"`
	Forecast       map[string]decimal.Decimal `json:"forecast"`
	Confidence     decimal.Decimal          `json:"confidence"`
	GeneratedAt    time.Time                `json:"generated_at"`
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Name       string          `json:"name" binding:"required"`
	Type       AlertType       `json:"type" binding:"required"`
	MetricType MetricType      `json:"metric_type" binding:"required"`
	Threshold  decimal.Decimal `json:"threshold" binding:"required"`
	Condition  string          `json:"condition" binding:"required"`
	Metadata   map[string]string `json:"metadata"`
}

// CreateAlertResponse represents the response after creating an alert
type CreateAlertResponse struct {
	AlertID   int64     `json:"alert_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateReportRequest represents a request to generate a custom report
type GenerateReportRequest struct {
	Name       string            `json:"name" binding:"required"`
	Type       string            `json:"type" binding:"required"`
	Period     string            `json:"period" binding:"required"`
	Filters    map[string]string `json:"filters"`
	CreatedBy  string            `json:"created_by" binding:"required"`
}

// GenerateReportResponse represents the response after generating a report
type GenerateReportResponse struct {
	ReportID    int64     `json:"report_id"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generated_at"`
	DownloadURL string    `json:"download_url"`
}

// WebhookPayload represents webhook payload data
type WebhookPayload struct {
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature"`
}

// AddDataSourceRequest represents a request to add a data source
type AddDataSourceRequest struct {
	Name   string            `json:"name" binding:"required"`
	Type   string            `json:"type" binding:"required"`
	URL    string            `json:"url" binding:"required"`
	APIKey string            `json:"api_key"`
	Config map[string]string `json:"config"`
}

// AddDataSourceResponse represents the response after adding a data source
type AddDataSourceResponse struct {
	DataSourceID int64     `json:"data_source_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
