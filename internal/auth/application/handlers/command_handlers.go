package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application/commands"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// CommandBus represents a command bus for handling commands
type CommandBus interface {
	Handle(ctx context.Context, cmd commands.Command) (*commands.CommandResult, error)
}

// UserCommandHandler handles user-related commands
type UserCommandHandler struct {
	userRepo        domain.UserRepository
	sessionRepo     domain.SessionRepository
	passwordService PasswordService
	jwtService      JWTService
	eventPublisher  EventPublisher
	logger          *logger.Logger
}

// NewUserCommandHandler creates a new user command handler
func NewUserCommandHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	passwordService PasswordService,
	jwtService JWTService,
	eventPublisher EventPublisher,
	logger *logger.Logger,
) *UserCommandHandler {
	return &UserCommandHandler{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

// HandleRegisterUser handles user registration command
func (h *UserCommandHandler) HandleRegisterUser(ctx context.Context, cmd commands.RegisterUserCommand) (*commands.CommandResult, error) {
	// Check if user already exists
	exists, err := h.userRepo.UserExists(ctx, cmd.Email)
	if err != nil {
		h.logger.ErrorWithFields("Failed to check user existence", logger.Error(err), logger.String("email", cmd.Email))
		return commands.NewCommandResult(false, "Internal server error", nil), err
	}

	if exists {
		return commands.NewCommandResult(false, "User already exists", nil), domain.ErrUserExists
	}

	// Hash password
	passwordHash, err := h.passwordService.HashPassword(cmd.Password)
	if err != nil {
		h.logger.ErrorWithFields("Failed to hash password", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to process password", nil), err
	}

	// Create user aggregate
	userAggregate, err := domain.NewUserAggregate(cmd.Email, passwordHash, cmd.Role)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create user aggregate", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to create user", nil), err
	}

	// Set additional fields
	if cmd.FirstName != "" {
		userAggregate.User.FirstName = cmd.FirstName
	}
	if cmd.LastName != "" {
		userAggregate.User.LastName = cmd.LastName
	}
	if cmd.PhoneNumber != "" {
		if err := domain.ValidatePhoneNumber(cmd.PhoneNumber); err != nil {
			return commands.NewCommandResult(false, "Invalid phone number", nil), err
		}
		userAggregate.User.PhoneNumber = cmd.PhoneNumber
	}
	if cmd.Metadata != nil {
		userAggregate.User.Metadata = cmd.Metadata
	}

	// Save user
	if err := h.userRepo.CreateUser(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to save user", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to save user", nil), err
	}

	// Publish events
	if err := h.publishEvents(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to publish events", logger.Error(err))
		// Don't fail the command for event publishing errors
	}

	result := commands.NewCommandResult(true, "User registered successfully", map[string]interface{}{
		"user_id": userAggregate.User.ID,
		"email":   userAggregate.User.Email,
	})

	return result, nil
}

// HandleLoginUser handles user login command
func (h *UserCommandHandler) HandleLoginUser(ctx context.Context, cmd commands.LoginUserCommand) (*commands.CommandResult, error) {
	// Get user by email
	user, err := h.userRepo.GetUserByEmail(ctx, cmd.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return commands.NewCommandResult(false, "Invalid credentials", nil), domain.ErrInvalidCredentials
		}
		h.logger.ErrorWithFields("Failed to get user", logger.Error(err))
		return commands.NewCommandResult(false, "Internal server error", nil), err
	}

	// Create user aggregate
	userAggregate := domain.LoadUserAggregate(user)

	// Verify password
	if err := h.passwordService.VerifyPassword(user.PasswordHash, cmd.Password); err != nil {
		// Record failed login
		userAggregate.RecordFailedLogin(cmd.IPAddress, cmd.UserAgent)
		
		// Save updated user
		if saveErr := h.userRepo.UpdateUser(ctx, userAggregate.User); saveErr != nil {
			h.logger.ErrorWithFields("Failed to update user after failed login", logger.Error(saveErr))
		}

		// Publish events
		if publishErr := h.publishEvents(ctx, userAggregate.User); publishErr != nil {
			h.logger.ErrorWithFields("Failed to publish events", logger.Error(publishErr))
		}

		return commands.NewCommandResult(false, "Invalid credentials", nil), domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, _, _, err := h.jwtService.GenerateTokenPair(ctx, user, "")
	if err != nil {
		h.logger.ErrorWithFields("Failed to generate tokens", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to generate tokens", nil), err
	}

	// Create session
	session := domain.NewSession(user.ID, accessToken, refreshToken, 15*time.Minute, 7*24*time.Hour)
	if cmd.DeviceInfo != nil {
		session.SetDeviceInfo(cmd.DeviceInfo)
	}
	session.IPAddress = cmd.IPAddress
	session.UserAgent = cmd.UserAgent

	// Extend session if remember me is enabled
	if cmd.RememberMe {
		extendedTTL := 30 * 24 * time.Hour // 30 days
		session.RefreshExpiresAt = time.Now().Add(extendedTTL)
	}

	// Save session
	if err := h.sessionRepo.CreateSession(ctx, session); err != nil {
		h.logger.ErrorWithFields("Failed to create session", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to create session", nil), err
	}

	// Record successful login
	if err := userAggregate.AttemptLogin(cmd.IPAddress, cmd.UserAgent, session.ID, false); err != nil {
		h.logger.ErrorWithFields("Failed to record login", logger.Error(err))
		// Continue with login even if recording fails
	}

	// Save updated user
	if err := h.userRepo.UpdateUser(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to update user", logger.Error(err))
		// Continue with login even if update fails
	}

	// Publish events
	if err := h.publishEvents(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to publish events", logger.Error(err))
		// Don't fail the command for event publishing errors
	}

	if err := h.publishEvents(ctx, session); err != nil {
		h.logger.ErrorWithFields("Failed to publish session events", logger.Error(err))
		// Don't fail the command for event publishing errors
	}

	result := commands.NewCommandResult(true, "Login successful", map[string]interface{}{
		"user_id":       user.ID,
		"session_id":    session.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(15 * 60), // 15 minutes in seconds
		"token_type":    "Bearer",
	})

	return result, nil
}

// HandleLogoutUser handles user logout command
func (h *UserCommandHandler) HandleLogoutUser(ctx context.Context, cmd commands.LogoutUserCommand) (*commands.CommandResult, error) {
	// Get session
	session, err := h.sessionRepo.GetSessionByID(ctx, cmd.SessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return commands.NewCommandResult(false, "Session not found", nil), err
		}
		h.logger.ErrorWithFields("Failed to get session", logger.Error(err))
		return commands.NewCommandResult(false, "Internal server error", nil), err
	}

	// Verify session belongs to user
	if session.UserID != cmd.UserID {
		return commands.NewCommandResult(false, "Unauthorized", nil), domain.ErrUnauthorized
	}

	// Revoke session
	session.Revoke()

	// Save session
	if err := h.sessionRepo.UpdateSession(ctx, session); err != nil {
		h.logger.ErrorWithFields("Failed to update session", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to logout", nil), err
	}

	// Publish events
	if err := h.publishEvents(ctx, session); err != nil {
		h.logger.ErrorWithFields("Failed to publish events", logger.Error(err))
		// Don't fail the command for event publishing errors
	}

	result := commands.NewCommandResult(true, "Logout successful", map[string]interface{}{
		"session_id": session.ID,
	})

	return result, nil
}

// HandleChangePassword handles password change command
func (h *UserCommandHandler) HandleChangePassword(ctx context.Context, cmd commands.ChangePasswordCommand) (*commands.CommandResult, error) {
	// Get user
	user, err := h.userRepo.GetUserByID(ctx, cmd.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return commands.NewCommandResult(false, "User not found", nil), err
		}
		h.logger.ErrorWithFields("Failed to get user", logger.Error(err))
		return commands.NewCommandResult(false, "Internal server error", nil), err
	}

	// Create user aggregate
	userAggregate := domain.LoadUserAggregate(user)

	// Verify old password (unless forced)
	if !cmd.Forced {
		if err := h.passwordService.VerifyPassword(user.PasswordHash, cmd.OldPassword); err != nil {
			return commands.NewCommandResult(false, "Invalid current password", nil), domain.ErrInvalidCredentials
		}
	}

	// Hash new password
	newPasswordHash, err := h.passwordService.HashPassword(cmd.NewPassword)
	if err != nil {
		h.logger.ErrorWithFields("Failed to hash new password", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to process new password", nil), err
	}

	// Change password using aggregate
	if err := userAggregate.ChangePassword(newPasswordHash, cmd.Forced); err != nil {
		return commands.NewCommandResult(false, err.Error(), nil), err
	}

	// Save user
	if err := h.userRepo.UpdateUser(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to update user", logger.Error(err))
		return commands.NewCommandResult(false, "Failed to save password change", nil), err
	}

	// Publish events
	if err := h.publishEvents(ctx, userAggregate.User); err != nil {
		h.logger.ErrorWithFields("Failed to publish events", logger.Error(err))
		// Don't fail the command for event publishing errors
	}

	result := commands.NewCommandResult(true, "Password changed successfully", nil)
	return result, nil
}

// Helper methods

// publishEvents publishes domain events
func (h *UserCommandHandler) publishEvents(ctx context.Context, aggregate interface{}) error {
	if eventSource, ok := aggregate.(interface{ GetEvents() []domain.DomainEvent }); ok {
		events := eventSource.GetEvents()
		for _, event := range events {
			if err := h.eventPublisher.Publish(ctx, &event); err != nil {
				return fmt.Errorf("failed to publish event %s: %w", event.Type, err)
			}
		}
		
		// Clear events after publishing
		if eventClearer, ok := aggregate.(interface{ ClearEvents() }); ok {
			eventClearer.ClearEvents()
		}
	}
	return nil
}

// Service interfaces (to be implemented in infrastructure layer)

// PasswordService defines password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

// JWTService defines JWT operations
type JWTService interface {
	GenerateTokenPair(ctx context.Context, user *domain.User, sessionID string) (accessToken, refreshToken string, accessClaims, refreshClaims *domain.TokenClaims, err error)
	ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)
}

// EventPublisher defines event publishing operations
type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DomainEvent) error
}
