package marketplace

import (
	"time"

	"github.com/shopspring/decimal"
)

// SolutionCategory represents the category of a solution
type SolutionCategory int

const (
	SolutionCategoryDeFi SolutionCategory = iota
	SolutionCategoryNFT
	SolutionCategoryDAO
	SolutionCategoryAnalytics
	SolutionCategoryInfrastructure
	SolutionCategorySecurity
	SolutionCategoryUI
	SolutionCategoryIntegration
)

func (sc SolutionCategory) String() string {
	switch sc {
	case SolutionCategoryDeFi:
		return "DEFI"
	case SolutionCategoryNFT:
		return "NFT"
	case SolutionCategoryDAO:
		return "DAO"
	case SolutionCategoryAnalytics:
		return "ANALYTICS"
	case SolutionCategoryInfrastructure:
		return "INFRASTRUCTURE"
	case SolutionCategorySecurity:
		return "SECURITY"
	case SolutionCategoryUI:
		return "UI"
	case SolutionCategoryIntegration:
		return "INTEGRATION"
	default:
		return "UNKNOWN"
	}
}

// SolutionStatus represents the status of a solution
type SolutionStatus int

const (
	SolutionStatusPending SolutionStatus = iota
	SolutionStatusApproved
	SolutionStatusRejected
	SolutionStatusDeprecated
)

func (ss SolutionStatus) String() string {
	switch ss {
	case SolutionStatusPending:
		return "PENDING"
	case SolutionStatusApproved:
		return "APPROVED"
	case SolutionStatusRejected:
		return "REJECTED"
	case SolutionStatusDeprecated:
		return "DEPRECATED"
	default:
		return "UNKNOWN"
	}
}

// InstallationStatus represents the status of a solution installation
type InstallationStatus int

const (
	InstallationStatusActive InstallationStatus = iota
	InstallationStatusInactive
	InstallationStatusFailed
)

func (is InstallationStatus) String() string {
	switch is {
	case InstallationStatusActive:
		return "ACTIVE"
	case InstallationStatusInactive:
		return "INACTIVE"
	case InstallationStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// Solution represents a solution in the marketplace
type Solution struct {
	ID                 int64            `json:"id" db:"id"`
	SolutionID         uint64           `json:"solution_id" db:"solution_id"`
	Name               string           `json:"name" db:"name"`
	Description        string           `json:"description" db:"description"`
	Category           SolutionCategory `json:"category" db:"category"`
	Version            string           `json:"version" db:"version"`
	DeveloperAddress   string           `json:"developer_address" db:"developer_address"`
	Status             SolutionStatus   `json:"status" db:"status"`
	RepositoryURL      string           `json:"repository_url" db:"repository_url"`
	DocumentationURL   string           `json:"documentation_url" db:"documentation_url"`
	DemoURL            string           `json:"demo_url" db:"demo_url"`
	Tags               []string         `json:"tags" db:"tags"`
	InstallCount       int              `json:"install_count" db:"install_count"`
	AverageRating      decimal.Decimal  `json:"average_rating" db:"average_rating"`
	ReviewCount        int              `json:"review_count" db:"review_count"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
	ApprovedAt         *time.Time       `json:"approved_at" db:"approved_at"`
	ApproverAddress    *string          `json:"approver_address" db:"approver_address"`
	TransactionHash    string           `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber        *int64           `json:"block_number" db:"block_number"`
}

// Review represents a solution review
type Review struct {
	ID                 int64     `json:"id" db:"id"`
	SolutionID         uint64    `json:"solution_id" db:"solution_id"`
	ReviewerAddress    string    `json:"reviewer_address" db:"reviewer_address"`
	Rating             int       `json:"rating" db:"rating"`
	Comment            string    `json:"comment" db:"comment"`
	SecurityScore      int       `json:"security_score" db:"security_score"`
	PerformanceScore   int       `json:"performance_score" db:"performance_score"`
	UsabilityScore     int       `json:"usability_score" db:"usability_score"`
	DocumentationScore int       `json:"documentation_score" db:"documentation_score"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Installation represents a solution installation
type Installation struct {
	ID               int64              `json:"id" db:"id"`
	SolutionID       uint64             `json:"solution_id" db:"solution_id"`
	InstallerAddress string             `json:"installer_address" db:"installer_address"`
	Environment      string             `json:"environment" db:"environment"`
	Version          string             `json:"version" db:"version"`
	Status           InstallationStatus `json:"status" db:"status"`
	InstalledAt      time.Time          `json:"installed_at" db:"installed_at"`
	LastUsed         *time.Time         `json:"last_used" db:"last_used"`
	ConfigData       map[string]string  `json:"config_data" db:"config_data"`
}

// QualityScore represents the quality metrics of a solution
type QualityScore struct {
	ID                 int64     `json:"id" db:"id"`
	SolutionID         uint64    `json:"solution_id" db:"solution_id"`
	OverallScore       float64   `json:"overall_score" db:"overall_score"`
	SecurityScore      float64   `json:"security_score" db:"security_score"`
	PerformanceScore   float64   `json:"performance_score" db:"performance_score"`
	UsabilityScore     float64   `json:"usability_score" db:"usability_score"`
	DocumentationScore float64   `json:"documentation_score" db:"documentation_score"`
	ReviewCount        int       `json:"review_count" db:"review_count"`
	LastUpdated        time.Time `json:"last_updated" db:"last_updated"`
}

// Category represents a solution category
type Category struct {
	ID          int64            `json:"id" db:"id"`
	Category    SolutionCategory `json:"category" db:"category"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	IconURL     string           `json:"icon_url" db:"icon_url"`
	SolutionCount int            `json:"solution_count" db:"solution_count"`
}

// CompatibilityResult represents compatibility check results
type CompatibilityResult struct {
	SolutionID   uint64    `json:"solution_id"`
	Environment  string    `json:"environment"`
	IsCompatible bool      `json:"is_compatible"`
	Issues       []string  `json:"issues"`
	CheckedAt    time.Time `json:"checked_at"`
}

// SolutionDetails represents detailed solution information
type SolutionDetails struct {
	Solution      *Solution             `json:"solution"`
	Reviews       []*Review             `json:"reviews"`
	Installations []*Installation       `json:"installations"`
	QualityScore  *QualityScore         `json:"quality_score"`
}

// Request/Response types for API

// CreateSolutionRequest represents a request to create a new solution
type CreateSolutionRequest struct {
	Name             string           `json:"name" binding:"required"`
	Description      string           `json:"description" binding:"required"`
	Category         SolutionCategory `json:"category" binding:"required"`
	Version          string           `json:"version" binding:"required"`
	DeveloperAddress string           `json:"developer_address" binding:"required"`
	RepositoryURL    string           `json:"repository_url" binding:"required"`
	DocumentationURL string           `json:"documentation_url"`
	DemoURL          string           `json:"demo_url"`
	Tags             []string         `json:"tags"`
}

// CreateSolutionResponse represents the response after creating a solution
type CreateSolutionResponse struct {
	SolutionID      uint64        `json:"solution_id"`
	TransactionHash string        `json:"transaction_hash"`
	Status          SolutionStatus `json:"status"`
	QualityScore    *QualityScore `json:"quality_score"`
}

// UpdateSolutionRequest represents a request to update a solution
type UpdateSolutionRequest struct {
	SolutionID       uint64   `json:"solution_id" binding:"required"`
	Version          string   `json:"version"`
	Description      string   `json:"description"`
	RepositoryURL    string   `json:"repository_url"`
	DocumentationURL string   `json:"documentation_url"`
	DemoURL          string   `json:"demo_url"`
	Tags             []string `json:"tags"`
}

// UpdateSolutionResponse represents the response after updating a solution
type UpdateSolutionResponse struct {
	SolutionID   uint64        `json:"solution_id"`
	Status       SolutionStatus `json:"status"`
	QualityScore *QualityScore `json:"quality_score"`
}

// ReviewSolutionRequest represents a request to review a solution
type ReviewSolutionRequest struct {
	SolutionID         uint64 `json:"solution_id" binding:"required"`
	ReviewerAddress    string `json:"reviewer_address" binding:"required"`
	Rating             int    `json:"rating" binding:"required,min=1,max=5"`
	Comment            string `json:"comment"`
	SecurityScore      int    `json:"security_score" binding:"min=1,max=5"`
	PerformanceScore   int    `json:"performance_score" binding:"min=1,max=5"`
	UsabilityScore     int    `json:"usability_score" binding:"min=1,max=5"`
	DocumentationScore int    `json:"documentation_score" binding:"min=1,max=5"`
}

// ReviewSolutionResponse represents the response after reviewing a solution
type ReviewSolutionResponse struct {
	ReviewID     int64         `json:"review_id"`
	QualityScore *QualityScore `json:"quality_score"`
}

// ApproveSolutionRequest represents a request to approve a solution
type ApproveSolutionRequest struct {
	SolutionID      uint64 `json:"solution_id" binding:"required"`
	ApproverAddress string `json:"approver_address" binding:"required"`
}

// ApproveSolutionResponse represents the response after approving a solution
type ApproveSolutionResponse struct {
	TransactionHash string        `json:"transaction_hash"`
	Status          SolutionStatus `json:"status"`
}

// InstallSolutionRequest represents a request to install a solution
type InstallSolutionRequest struct {
	SolutionID       uint64            `json:"solution_id" binding:"required"`
	InstallerAddress string            `json:"installer_address" binding:"required"`
	Environment      string            `json:"environment" binding:"required"`
	ConfigData       map[string]string `json:"config_data"`
}

// InstallSolutionResponse represents the response after installing a solution
type InstallSolutionResponse struct {
	InstallationID int64              `json:"installation_id"`
	Status         InstallationStatus `json:"status"`
	Version        string             `json:"version"`
}

// GetSolutionsRequest represents a request to get solutions
type GetSolutionsRequest struct {
	Category        *SolutionCategory `json:"category"`
	Status          *SolutionStatus   `json:"status"`
	DeveloperAddress string           `json:"developer_address"`
	Tags            []string          `json:"tags"`
	MinRating       float64           `json:"min_rating"`
	Limit           int               `json:"limit"`
	Offset          int               `json:"offset"`
	SortBy          string            `json:"sort_by"`
	SortOrder       string            `json:"sort_order"`
}

// GetSolutionsResponse represents the response with solutions list
type GetSolutionsResponse struct {
	Solutions []*Solution `json:"solutions"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}

// QualityScoreRequest represents a request to calculate quality score
type QualityScoreRequest struct {
	SolutionID uint64 `json:"solution_id" binding:"required"`
}

// QualityScoreResponse represents the response with quality score
type QualityScoreResponse struct {
	QualityScore *QualityScore `json:"quality_score"`
}

// CompatibilityRequest represents a request to check compatibility
type CompatibilityRequest struct {
	SolutionID  uint64 `json:"solution_id" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}

// CompatibilityResponse represents the response with compatibility results
type CompatibilityResponse struct {
	Result *CompatibilityResult `json:"result"`
}

// MarketplaceStats represents marketplace statistics
type MarketplaceStats struct {
	TotalSolutions     int             `json:"total_solutions"`
	ApprovedSolutions  int             `json:"approved_solutions"`
	TotalInstallations int             `json:"total_installations"`
	TotalReviews       int             `json:"total_reviews"`
	AverageRating      decimal.Decimal `json:"average_rating"`
	TopCategories      []CategoryStats `json:"top_categories"`
	RecentActivity     []ActivityItem  `json:"recent_activity"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category      SolutionCategory `json:"category"`
	Name          string           `json:"name"`
	SolutionCount int              `json:"solution_count"`
	InstallCount  int              `json:"install_count"`
	AverageRating decimal.Decimal  `json:"average_rating"`
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	Type        string    `json:"type"`
	SolutionID  uint64    `json:"solution_id"`
	SolutionName string   `json:"solution_name"`
	UserAddress string    `json:"user_address"`
	Timestamp   time.Time `json:"timestamp"`
	Details     string    `json:"details"`
}

// SolutionFilter represents filters for querying solutions
type SolutionFilter struct {
	Category        *SolutionCategory
	Status          *SolutionStatus
	DeveloperAddress string
	Tags            []string
	MinRating       float64
	Limit           int
	Offset          int
	SortBy          string
	SortOrder       string
}

// Blockchain request types
type SolutionSubmissionRequest struct {
	Name        string
	Description string
	Category    SolutionCategory
	Version     string
	Developer   string
}
