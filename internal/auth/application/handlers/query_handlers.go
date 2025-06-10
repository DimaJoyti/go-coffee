package handlers

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application/queries"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// QueryBus represents a query bus for handling queries
type QueryBus interface {
	Handle(ctx context.Context, query queries.Query) (interface{}, error)
}

// UserQueryHandler handles user-related queries
type UserQueryHandler struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	logger      *logger.Logger
}

// NewUserQueryHandler creates a new user query handler
func NewUserQueryHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	logger *logger.Logger,
) *UserQueryHandler {
	return &UserQueryHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// HandleGetUserByID handles get user by ID query
func (h *UserQueryHandler) HandleGetUserByID(ctx context.Context, query queries.GetUserByIDQuery) (*queries.QueryResult[*UserDTO], error) {
	user, err := h.userRepo.GetUserByID(ctx, query.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return queries.NewQueryResult[*UserDTO](nil, false, "User not found"), err
		}
		h.logger.ErrorWithFields("Failed to get user by ID", logger.Error(err), logger.String("user_id", query.UserID))
		return queries.NewQueryResult[*UserDTO](nil, false, "Internal server error"), err
	}

	userDTO := ToUserDTO(user)
	return queries.NewQueryResult(userDTO, true, "User retrieved successfully"), nil
}

// HandleGetUserByEmail handles get user by email query
func (h *UserQueryHandler) HandleGetUserByEmail(ctx context.Context, query queries.GetUserByEmailQuery) (*queries.QueryResult[*UserDTO], error) {
	user, err := h.userRepo.GetUserByEmail(ctx, query.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return queries.NewQueryResult[*UserDTO](nil, false, "User not found"), err
		}
		h.logger.ErrorWithFields("Failed to get user by email", logger.Error(err), logger.String("email", query.Email))
		return queries.NewQueryResult[*UserDTO](nil, false, "Internal server error"), err
	}

	userDTO := ToUserDTO(user)
	return queries.NewQueryResult(userDTO, true, "User retrieved successfully"), nil
}

// HandleGetUserProfile handles get user profile query
func (h *UserQueryHandler) HandleGetUserProfile(ctx context.Context, query queries.GetUserProfileQuery) (*queries.QueryResult[*UserProfileDTO], error) {
	user, err := h.userRepo.GetUserByID(ctx, query.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return queries.NewQueryResult[*UserProfileDTO](nil, false, "User not found"), err
		}
		h.logger.ErrorWithFields("Failed to get user profile", logger.Error(err), logger.String("user_id", query.UserID))
		return queries.NewQueryResult[*UserProfileDTO](nil, false, "Internal server error"), err
	}

	profileDTO := ToUserProfileDTO(user)
	return queries.NewQueryResult(profileDTO, true, "User profile retrieved successfully"), nil
}

// HandleGetUsers handles get users query with pagination
func (h *UserQueryHandler) HandleGetUsers(ctx context.Context, query queries.GetUsersQuery) (*queries.PaginatedResult[*UserDTO], error) {
	// Set default pagination values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// Calculate offset
	offset := (query.Page - 1) * query.PageSize

	// Get users with filters (this would be implemented in the repository)
	users, total, err := h.getUsersWithPagination(ctx, query, offset)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get users", logger.Error(err))
		result := &queries.PaginatedResult[*UserDTO]{}
		result.AddError("Failed to retrieve users")
		return result, err
	}

	// Convert to DTOs
	userDTOs := make([]*UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = ToUserDTO(user)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))
	pagination := queries.Pagination{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return queries.NewPaginatedResult(userDTOs, pagination), nil
}

// HandleGetUserSessions handles get user sessions query
func (h *UserQueryHandler) HandleGetUserSessions(ctx context.Context, query queries.GetUserSessionsQuery) (*queries.PaginatedResult[*SessionDTO], error) {
	// Set default pagination values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// Calculate offset
	offset := (query.Page - 1) * query.PageSize

	// Get sessions with filters
	sessions, total, err := h.getUserSessionsWithPagination(ctx, query, offset)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get user sessions", logger.Error(err))
		result := &queries.PaginatedResult[*SessionDTO]{}
		result.AddError("Failed to retrieve sessions")
		return result, err
	}

	// Convert to DTOs
	sessionDTOs := make([]*SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionDTOs[i] = ToSessionDTO(session)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))
	pagination := queries.Pagination{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return queries.NewPaginatedResult(sessionDTOs, pagination), nil
}

// HandleGetSessionByID handles get session by ID query
func (h *UserQueryHandler) HandleGetSessionByID(ctx context.Context, query queries.GetSessionByIDQuery) (*queries.QueryResult[*SessionDTO], error) {
	session, err := h.sessionRepo.GetSessionByID(ctx, query.SessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return queries.NewQueryResult[*SessionDTO](nil, false, "Session not found"), err
		}
		h.logger.ErrorWithFields("Failed to get session by ID", logger.Error(err), logger.String("session_id", query.SessionID))
		return queries.NewQueryResult[*SessionDTO](nil, false, "Internal server error"), err
	}

	sessionDTO := ToSessionDTO(session)
	return queries.NewQueryResult(sessionDTO, true, "Session retrieved successfully"), nil
}

// Helper methods (these would be implemented based on your repository interface)

func (h *UserQueryHandler) getUsersWithPagination(ctx context.Context, query queries.GetUsersQuery, offset int) ([]*domain.User, int64, error) {
	// This is a placeholder - implement based on your repository interface
	// The repository should support filtering, sorting, and pagination

	// For now, return empty results
	return []*domain.User{}, 0, nil
}

func (h *UserQueryHandler) getUserSessionsWithPagination(ctx context.Context, query queries.GetUserSessionsQuery, offset int) ([]*domain.Session, int64, error) {
	// This is a placeholder - implement based on your repository interface
	// The repository should support filtering and pagination for user sessions

	// For now, return empty results
	return []*domain.Session{}, 0, nil
}

// DTO conversion functions

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name,omitempty"`
	LastName        string     `json:"last_name,omitempty"`
	PhoneNumber     string     `json:"phone_number,omitempty"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	IsEmailVerified bool       `json:"is_email_verified"`
	IsPhoneVerified bool       `json:"is_phone_verified"`
	MFAEnabled      bool       `json:"mfa_enabled"`
	SecurityLevel   string     `json:"security_level"`
	RiskScore       float64    `json:"risk_score"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UserProfileDTO represents a detailed user profile
type UserProfileDTO struct {
	*UserDTO
	MFAMethod          string     `json:"mfa_method,omitempty"`
	DeviceCount        int        `json:"device_count"`
	ActiveSessionCount int        `json:"active_session_count"`
	LastPasswordChange *time.Time `json:"last_password_change,omitempty"`
	FailedLoginCount   int        `json:"failed_login_count"`
	LockedUntil        *time.Time `json:"locked_until,omitempty"`
}

// SessionDTO represents a session data transfer object
type SessionDTO struct {
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	Status     string             `json:"status"`
	ExpiresAt  time.Time          `json:"expires_at"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
	IPAddress  string             `json:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
}

// ToUserDTO converts a domain User to UserDTO
func ToUserDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:              user.ID,
		Email:           user.Email,
		FirstName:       "", // Add FirstName field to domain.User if needed
		LastName:        "", // Add LastName field to domain.User if needed
		PhoneNumber:     user.PhoneNumber,
		Role:            string(user.Role),
		Status:          string(user.Status),
		IsEmailVerified: user.IsEmailVerified,
		IsPhoneVerified: user.IsPhoneVerified,
		MFAEnabled:      user.MFAEnabled,
		SecurityLevel:   string(user.SecurityLevel),
		RiskScore:       user.RiskScore,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// ToUserProfileDTO converts a domain User to UserProfileDTO
func ToUserProfileDTO(user *domain.User) *UserProfileDTO {
	userDTO := ToUserDTO(user)
	return &UserProfileDTO{
		UserDTO:            userDTO,
		MFAMethod:          string(user.MFAMethod),
		DeviceCount:        len(user.DeviceFingerprints),
		FailedLoginCount:   user.FailedLoginCount,
		LastPasswordChange: user.LastPasswordChange,
		LockedUntil:        user.LockedUntil,
		// ActiveSessionCount would be populated by a separate query
	}
}

// ToSessionDTO converts a domain Session to SessionDTO
func ToSessionDTO(session *domain.Session) *SessionDTO {
	if session == nil {
		return nil
	}
	return &SessionDTO{
		ID:         session.ID,
		UserID:     session.UserID,
		Status:     string(session.Status),
		ExpiresAt:  session.ExpiresAt,
		DeviceInfo: session.DeviceInfo,
		IPAddress:  session.IPAddress,
		UserAgent:  session.UserAgent,
		CreatedAt:  session.CreatedAt,
		LastUsedAt: session.LastUsedAt,
	}
}
