package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/lib/pq"
)

// UserRepository implements the UserRepository interface using PostgreSQL
type UserRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB, logger *logger.Logger) domain.UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone_number,
			role, status, is_email_verified, is_phone_verified,
			mfa_enabled, mfa_method, mfa_secret, mfa_backup_codes,
			failed_login_attempts, last_failed_login, last_login_at,
			last_password_change, security_level, risk_score,
			device_fingerprints, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23
		)`

	// Convert complex fields to JSON
	mfaBackupCodesJSON, err := json.Marshal(user.MFABackupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal MFA backup codes: %w", err)
	}

	deviceFingerprintsJSON, err := json.Marshal(user.DeviceFingerprints)
	if err != nil {
		return fmt.Errorf("failed to marshal device fingerprints: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.PhoneNumber, user.Role, user.Status, user.IsEmailVerified,
		user.IsPhoneVerified, user.MFAEnabled, user.MFAMethod, user.MFASecret,
		mfaBackupCodesJSON, user.FailedLoginCount, user.LastFailedLoginAt,
		user.LastLoginAt, user.LastPasswordChange, user.SecurityLevel,
		user.RiskScore, deviceFingerprintsJSON, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return domain.ErrEmailAlreadyExists
				}
			}
		}
		r.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithField("user_id", user.ID).Info("User created successfully")
	return nil
}

// GetUser retrieves a user by ID
func (r *UserRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone_number,
			   role, status, is_email_verified, is_phone_verified,
			   mfa_enabled, mfa_method, mfa_secret, mfa_backup_codes,
			   failed_login_attempts, last_failed_login, last_login_at,
			   last_password_change, security_level, risk_score,
			   device_fingerprints, created_at, updated_at
		FROM users WHERE id = $1`

	user := &domain.User{}
	var mfaBackupCodesJSON, deviceFingerprintsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.PhoneNumber, &user.Role, &user.Status,
		&user.IsEmailVerified, &user.IsPhoneVerified, &user.MFAEnabled,
		&user.MFAMethod, &user.MFASecret, &mfaBackupCodesJSON,
		&user.FailedLoginCount, &user.LastFailedLoginAt, &user.LastLoginAt,
		&user.LastPasswordChange, &user.SecurityLevel, &user.RiskScore,
		&deviceFingerprintsJSON, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(mfaBackupCodesJSON, &user.MFABackupCodes); err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to unmarshal MFA backup codes")
		return nil, fmt.Errorf("failed to unmarshal MFA backup codes: %w", err)
	}

	if err := json.Unmarshal(deviceFingerprintsJSON, &user.DeviceFingerprints); err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to unmarshal device fingerprints")
		return nil, fmt.Errorf("failed to unmarshal device fingerprints: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone_number,
			   role, status, is_email_verified, is_phone_verified,
			   mfa_enabled, mfa_method, mfa_secret, mfa_backup_codes,
			   failed_login_attempts, last_failed_login, last_login_at,
			   last_password_change, security_level, risk_score,
			   device_fingerprints, created_at, updated_at
		FROM users WHERE email = $1`

	user := &domain.User{}
	var mfaBackupCodesJSON, deviceFingerprintsJSON []byte

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.PhoneNumber, &user.Role, &user.Status,
		&user.IsEmailVerified, &user.IsPhoneVerified, &user.MFAEnabled,
		&user.MFAMethod, &user.MFASecret, &mfaBackupCodesJSON,
		&user.FailedLoginCount, &user.LastFailedLoginAt, &user.LastLoginAt,
		&user.LastPasswordChange, &user.SecurityLevel, &user.RiskScore,
		&deviceFingerprintsJSON, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		r.logger.WithError(err).WithField("email", email).Error("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(mfaBackupCodesJSON, &user.MFABackupCodes); err != nil {
		r.logger.WithError(err).WithField("email", email).Error("Failed to unmarshal MFA backup codes")
		return nil, fmt.Errorf("failed to unmarshal MFA backup codes: %w", err)
	}

	if err := json.Unmarshal(deviceFingerprintsJSON, &user.DeviceFingerprints); err != nil {
		r.logger.WithError(err).WithField("email", email).Error("Failed to unmarshal device fingerprints")
		return nil, fmt.Errorf("failed to unmarshal device fingerprints: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			email = $2, password_hash = $3, first_name = $4, last_name = $5,
			phone_number = $6, role = $7, status = $8, is_email_verified = $9,
			is_phone_verified = $10, mfa_enabled = $11, mfa_method = $12,
			mfa_secret = $13, mfa_backup_codes = $14, failed_login_attempts = $15,
			last_failed_login = $16, last_login_at = $17, last_password_change = $18,
			security_level = $19, risk_score = $20, device_fingerprints = $21,
			updated_at = $22
		WHERE id = $1`

	// Convert complex fields to JSON
	mfaBackupCodesJSON, err := json.Marshal(user.MFABackupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal MFA backup codes: %w", err)
	}

	deviceFingerprintsJSON, err := json.Marshal(user.DeviceFingerprints)
	if err != nil {
		return fmt.Errorf("failed to marshal device fingerprints: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.PhoneNumber, user.Role, user.Status, user.IsEmailVerified,
		user.IsPhoneVerified, user.MFAEnabled, user.MFAMethod, user.MFASecret,
		mfaBackupCodesJSON, user.FailedLoginCount, user.LastFailedLoginAt,
		user.LastLoginAt, user.LastPasswordChange, user.SecurityLevel,
		user.RiskScore, deviceFingerprintsJSON, user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return domain.ErrEmailAlreadyExists
				}
			}
		}
		r.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", user.ID).Info("User updated successfully")
	return nil
}

// DeleteUser deletes a user (soft delete by updating status)
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, domain.UserStatusInactive, time.Now(), userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("User deleted successfully")
	return nil
}

// ExistsUser checks if a user exists by ID
func (r *UserRepository) ExistsUser(ctx context.Context, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to check user existence")
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// ExistsUserByEmail checks if a user exists by email
func (r *UserRepository) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).WithField("email", email).Error("Failed to check user existence by email")
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}
	return exists, nil
}

// Missing interface methods

// GetUserByID retrieves a user by ID (alias for GetUser)
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return r.GetUser(ctx, userID)
}

// UpdateUserPassword updates a user's password
func (r *UserRepository) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, last_password_change = $2, updated_at = $3 WHERE id = $4`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, passwordHash, now, now, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user password")
		return fmt.Errorf("failed to update user password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("User password updated successfully")
	return nil
}

// UpdateUserStatus updates a user's status
func (r *UserRepository) UpdateUserStatus(ctx context.Context, userID string, status domain.UserStatus) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, now, userID)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id": userID,
			"status":  status,
		}).Error("Failed to update user status")
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"status":  status,
	}).Info("User status updated successfully")
	return nil
}

// UpdateUserRole updates a user's role
func (r *UserRepository) UpdateUserRole(ctx context.Context, userID string, role domain.UserRole) error {
	query := `UPDATE users SET role = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, role, now, userID)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id": userID,
			"role":    role,
		}).Error("Failed to update user role")
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}).Info("User role updated successfully")
	return nil
}

// IncrementFailedLoginAttempts increments the failed login attempts counter
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	query := `UPDATE users SET failed_login_count = failed_login_count + 1, last_failed_login_at = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to increment failed login attempts")
		return fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("Failed login attempts incremented")
	return nil
}

// ResetFailedLoginAttempts resets the failed login attempts counter
func (r *UserRepository) ResetFailedLoginAttempts(ctx context.Context, userID string) error {
	query := `UPDATE users SET failed_login_count = 0, last_failed_login_at = NULL, updated_at = $1 WHERE id = $2`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to reset failed login attempts")
		return fmt.Errorf("failed to reset failed login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("Failed login attempts reset")
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to update last login")
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("Last login updated")
	return nil
}

// GetFailedLoginCount gets the failed login count for a user
func (r *UserRepository) GetFailedLoginCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT failed_login_count FROM users WHERE id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrUserNotFound
		}
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to get failed login count")
		return 0, fmt.Errorf("failed to get failed login count: %w", err)
	}

	return count, nil
}

// Interface methods with different names

// IncrementFailedLogin increments failed login count (interface method)
func (r *UserRepository) IncrementFailedLogin(ctx context.Context, userID string) error {
	return r.IncrementFailedLoginAttempts(ctx, userID)
}

// ResetFailedLogin resets failed login count (interface method)
func (r *UserRepository) ResetFailedLogin(ctx context.Context, userID string) error {
	return r.ResetFailedLoginAttempts(ctx, userID)
}

// SetFailedLoginCount sets the failed login count with expiry (cache-based method)
func (r *UserRepository) SetFailedLoginCount(ctx context.Context, email string, count int, expiry time.Duration) error {
	// This method is typically implemented by a cache service, not database
	// For now, we'll implement it as a database update
	query := `UPDATE users SET failed_login_count = $1, updated_at = $2 WHERE email = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, count, now, email)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"email": email,
			"count": count,
		}).Error("Failed to set failed login count")
		return fmt.Errorf("failed to set failed login count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithFields(map[string]interface{}{
		"email": email,
		"count": count,
	}).Info("Failed login count set")
	return nil
}

// LockUser locks a user account until a specific time
func (r *UserRepository) LockUser(ctx context.Context, userID string, until time.Time) error {
	query := `UPDATE users SET status = $1, locked_until = $2, updated_at = $3 WHERE id = $4`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, domain.UserStatusLocked, until, now, userID)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id": userID,
			"until":   until,
		}).Error("Failed to lock user")
		return fmt.Errorf("failed to lock user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"until":   until,
	}).Info("User locked successfully")
	return nil
}

// UnlockUser unlocks a user account
func (r *UserRepository) UnlockUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET status = $1, locked_until = NULL, updated_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, domain.UserStatusActive, now, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to unlock user")
		return fmt.Errorf("failed to unlock user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.WithField("user_id", userID).Info("User unlocked successfully")
	return nil
}

// UserExists checks if a user exists (interface method)
func (r *UserRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	return r.ExistsUser(ctx, userID)
}

// UserExistsByID checks if a user exists by ID (interface method)
func (r *UserRepository) UserExistsByID(ctx context.Context, userID string) (bool, error) {
	return r.ExistsUser(ctx, userID)
}

// UserExistsByEmail checks if a user exists by email (interface method)
func (r *UserRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.ExistsUserByEmail(ctx, email)
}
