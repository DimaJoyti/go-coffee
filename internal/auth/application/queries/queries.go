package queries

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// Query represents a query in the CQRS pattern
type Query interface {
	QueryType() string
}

// QueryHandler represents a query handler
type QueryHandler[T Query, R any] interface {
	Handle(query T) (R, error)
}

// User Queries

// GetUserByIDQuery represents a query to get user by ID
type GetUserByIDQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetUserByIDQuery) QueryType() string { return "GetUserByID" }

// GetUserByEmailQuery represents a query to get user by email
type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

func (q GetUserByEmailQuery) QueryType() string { return "GetUserByEmail" }

// GetUserProfileQuery represents a query to get user profile
type GetUserProfileQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetUserProfileQuery) QueryType() string { return "GetUserProfile" }

// GetUsersQuery represents a query to get users with pagination
type GetUsersQuery struct {
	Page     int                 `json:"page" validate:"min=1"`
	PageSize int                 `json:"page_size" validate:"min=1,max=100"`
	Role     domain.UserRole     `json:"role,omitempty"`
	Status   domain.UserStatus   `json:"status,omitempty"`
	Search   string              `json:"search,omitempty"`
	SortBy   string              `json:"sort_by,omitempty"`
	SortDir  string              `json:"sort_dir,omitempty"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

func (q GetUsersQuery) QueryType() string { return "GetUsers" }

// Session Queries

// GetUserSessionsQuery represents a query to get user sessions
type GetUserSessionsQuery struct {
	UserID   string `json:"user_id" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Active   *bool  `json:"active,omitempty"`
}

func (q GetUserSessionsQuery) QueryType() string { return "GetUserSessions" }

// GetSessionByIDQuery represents a query to get session by ID
type GetSessionByIDQuery struct {
	SessionID string `json:"session_id" validate:"required"`
}

func (q GetSessionByIDQuery) QueryType() string { return "GetSessionByID" }

// GetActiveSessionsQuery represents a query to get active sessions
type GetActiveSessionsQuery struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

func (q GetActiveSessionsQuery) QueryType() string { return "GetActiveSessions" }

// Token Queries

// ValidateTokenQuery represents a query to validate a token
type ValidateTokenQuery struct {
	Token string `json:"token" validate:"required"`
}

func (q ValidateTokenQuery) QueryType() string { return "ValidateToken" }

// GetTokenInfoQuery represents a query to get token information
type GetTokenInfoQuery struct {
	TokenID string `json:"token_id" validate:"required"`
}

func (q GetTokenInfoQuery) QueryType() string { return "GetTokenInfo" }

// Security Queries

// GetSecurityEventsQuery represents a query to get security events
type GetSecurityEventsQuery struct {
	UserID    string    `json:"user_id,omitempty"`
	EventType string    `json:"event_type,omitempty"`
	Severity  string    `json:"severity,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

func (q GetSecurityEventsQuery) QueryType() string { return "GetSecurityEvents" }

// GetUserRiskScoreQuery represents a query to get user risk score
type GetUserRiskScoreQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetUserRiskScoreQuery) QueryType() string { return "GetUserRiskScore" }

// GetUserDevicesQuery represents a query to get user devices
type GetUserDevicesQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetUserDevicesQuery) QueryType() string { return "GetUserDevices" }

// CheckAccountLockStatusQuery represents a query to check account lock status
type CheckAccountLockStatusQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q CheckAccountLockStatusQuery) QueryType() string { return "CheckAccountLockStatus" }

// MFA Queries

// GetMFAStatusQuery represents a query to get MFA status
type GetMFAStatusQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetMFAStatusQuery) QueryType() string { return "GetMFAStatus" }

// GetBackupCodesQuery represents a query to get backup codes
type GetBackupCodesQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

func (q GetBackupCodesQuery) QueryType() string { return "GetBackupCodes" }

// Audit Queries

// GetUserAuditLogQuery represents a query to get user audit log
type GetUserAuditLogQuery struct {
	UserID    string    `json:"user_id" validate:"required"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Action    string    `json:"action,omitempty"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

func (q GetUserAuditLogQuery) QueryType() string { return "GetUserAuditLog" }

// GetSystemAuditLogQuery represents a query to get system audit log
type GetSystemAuditLogQuery struct {
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Event     string    `json:"event,omitempty"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

func (q GetSystemAuditLogQuery) QueryType() string { return "GetSystemAuditLog" }

// Analytics Queries

// GetLoginStatsQuery represents a query to get login statistics
type GetLoginStatsQuery struct {
	UserID    string    `json:"user_id,omitempty"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
	Granularity string  `json:"granularity,omitempty"` // hour, day, week, month
}

func (q GetLoginStatsQuery) QueryType() string { return "GetLoginStats" }

// GetUserActivityQuery represents a query to get user activity
type GetUserActivityQuery struct {
	UserID    string    `json:"user_id" validate:"required"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

func (q GetUserActivityQuery) QueryType() string { return "GetUserActivity" }

// GetSecurityMetricsQuery represents a query to get security metrics
type GetSecurityMetricsQuery struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
	MetricType string   `json:"metric_type,omitempty"` // failed_logins, locked_accounts, mfa_usage
}

func (q GetSecurityMetricsQuery) QueryType() string { return "GetSecurityMetrics" }

// Health Queries

// GetHealthStatusQuery represents a query to get health status
type GetHealthStatusQuery struct {
	Component string `json:"component,omitempty"` // database, redis, external_services
}

func (q GetHealthStatusQuery) QueryType() string { return "GetHealthStatus" }

// Query Results

// QueryResult represents the result of a query execution
type QueryResult[T any] struct {
	Data      T                      `json:"data"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewQueryResult creates a new query result
func NewQueryResult[T any](data T, success bool, message string) *QueryResult[T] {
	return &QueryResult[T]{
		Data:      data,
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// AddError adds an error to the query result
func (qr *QueryResult[T]) AddError(err string) {
	if qr.Errors == nil {
		qr.Errors = make([]string, 0)
	}
	qr.Errors = append(qr.Errors, err)
	qr.Success = false
}

// AddMetadata adds metadata to the query result
func (qr *QueryResult[T]) AddMetadata(key string, value interface{}) {
	if qr.Metadata == nil {
		qr.Metadata = make(map[string]interface{})
	}
	qr.Metadata[key] = value
}

// Pagination represents pagination information
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResult represents a paginated query result
type PaginatedResult[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Errors     []string   `json:"errors,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](items []T, pagination Pagination) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Items:      items,
		Pagination: pagination,
		Success:    true,
		Timestamp:  time.Now(),
	}
}

// AddError adds an error to the paginated result
func (pr *PaginatedResult[T]) AddError(err string) {
	if pr.Errors == nil {
		pr.Errors = make([]string, 0)
	}
	pr.Errors = append(pr.Errors, err)
	pr.Success = false
}
