package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SessionRepository implements the SessionRepository interface using PostgreSQL
type SessionRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSessionRepository creates a new PostgreSQL session repository
func NewSessionRepository(db *sql.DB, logger *logger.Logger) domain.SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// CreateSession creates a new session in the database
func (r *SessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, refresh_token_hash, device_info,
			ip_address, user_agent, expires_at, refresh_expires_at,
			is_active, last_activity, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	// Convert device info to JSON
	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.RefreshTokenHash,
		deviceInfoJSON, session.IPAddress, session.UserAgent, session.ExpiresAt,
		session.RefreshExpiresAt, session.IsActive, session.LastActivity,
		session.CreatedAt, session.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to create session")
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"session_id": session.ID,
		"user_id":    session.UserID,
	}).Info("Session created successfully")
	return nil
}

// GetSession retrieves a session by ID
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, user_agent, expires_at, refresh_expires_at,
			   is_active, last_activity, created_at, updated_at
		FROM sessions WHERE id = $1`

	session := &domain.Session{}
	var deviceInfoJSON []byte

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
		&deviceInfoJSON, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.RefreshExpiresAt, &session.IsActive, &session.LastActivity,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Unmarshal device info
	if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to unmarshal device info")
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return session, nil
}

// GetSessionByToken retrieves a session by token hash
func (r *SessionRepository) GetSessionByToken(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, user_agent, expires_at, refresh_expires_at,
			   is_active, last_activity, created_at, updated_at
		FROM sessions WHERE token_hash = $1 AND is_active = true`

	session := &domain.Session{}
	var deviceInfoJSON []byte

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
		&deviceInfoJSON, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.RefreshExpiresAt, &session.IsActive, &session.LastActivity,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.WithError(err).WithField("token_hash", tokenHash[:10]+"...").Error("Failed to get session by token")
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	// Unmarshal device info
	if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to unmarshal device info")
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token hash
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, user_agent, expires_at, refresh_expires_at,
			   is_active, last_activity, created_at, updated_at
		FROM sessions WHERE refresh_token_hash = $1 AND is_active = true`

	session := &domain.Session{}
	var deviceInfoJSON []byte

	err := r.db.QueryRowContext(ctx, query, refreshTokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
		&deviceInfoJSON, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.RefreshExpiresAt, &session.IsActive, &session.LastActivity,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.WithError(err).WithField("refresh_token_hash", refreshTokenHash[:10]+"...").Error("Failed to get session by refresh token")
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	// Unmarshal device info
	if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to unmarshal device info")
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return session, nil
}

// UpdateSession updates an existing session
func (r *SessionRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions SET
			token_hash = $2, refresh_token_hash = $3, device_info = $4,
			ip_address = $5, user_agent = $6, expires_at = $7,
			refresh_expires_at = $8, is_active = $9, last_activity = $10,
			updated_at = $11
		WHERE id = $1`

	// Convert device info to JSON
	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		session.ID, session.TokenHash, session.RefreshTokenHash, deviceInfoJSON,
		session.IPAddress, session.UserAgent, session.ExpiresAt,
		session.RefreshExpiresAt, session.IsActive, session.LastActivity,
		session.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	r.logger.WithField("session_id", session.ID).Info("Session updated successfully")
	return nil
}

// RevokeSession revokes a session by setting it as inactive
func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET is_active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to revoke session")
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	r.logger.WithField("session_id", sessionID).Info("Session revoked successfully")
	return nil
}

// RevokeUserSessions revokes all sessions for a user
func (r *SessionRepository) RevokeUserSessions(ctx context.Context, userID string) error {
	query := `UPDATE sessions SET is_active = false, updated_at = $1 WHERE user_id = $2 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to revoke user sessions")
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"user_id":          userID,
		"sessions_revoked": rowsAffected,
	}).Info("User sessions revoked successfully")
	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *SessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, device_info,
			   ip_address, user_agent, expires_at, refresh_expires_at,
			   is_active, last_activity, created_at, updated_at
		FROM sessions 
		WHERE user_id = $1 AND is_active = true
		ORDER BY last_activity DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user sessions")
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		var deviceInfoJSON []byte

		err := rows.Scan(
			&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
			&deviceInfoJSON, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
			&session.RefreshExpiresAt, &session.IsActive, &session.LastActivity,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan session row")
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		// Unmarshal device info
		if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
			r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to unmarshal device info")
			continue
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating session rows")
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1 OR refresh_expires_at < $1`

	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		r.logger.WithError(err).Error("Failed to cleanup expired sessions")
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithField("sessions_cleaned", rowsAffected).Info("Expired sessions cleaned up")
	return nil
}

// GetActiveSessionCount returns the number of active sessions for a user
func (r *SessionRepository) GetActiveSessionCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND is_active = true AND expires_at > $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, time.Now()).Scan(&count)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to get active session count")
		return 0, fmt.Errorf("failed to get active session count: %w", err)
	}

	return count, nil
}

// UpdateLastActivity updates the last activity timestamp for a session
func (r *SessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET last_activity = $1, updated_at = $1 WHERE id = $2 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to update last activity")
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// Missing interface methods

// GetSessionByID retrieves a session by ID (alias for GetSession)
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	return r.GetSession(ctx, sessionID)
}

// GetSessionByAccessToken retrieves a session by access token (alias for GetSessionByToken)
func (r *SessionRepository) GetSessionByAccessToken(ctx context.Context, accessToken string) (*domain.Session, error) {
	return r.GetSessionByToken(ctx, accessToken)
}

// DeleteSession deletes a session by ID
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	r.logger.WithField("session_id", sessionID).Info("Session deleted successfully")
	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (r *SessionRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to delete user sessions")
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"user_id":          userID,
		"sessions_deleted": rowsAffected,
	}).Info("User sessions deleted successfully")
	return nil
}

// DeleteExpiredSessions deletes expired sessions (alias for CleanupExpiredSessions)
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.CleanupExpiredSessions(ctx)
}

// IsSessionValid checks if a session is valid
func (r *SessionRepository) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
	query := `SELECT COUNT(*) FROM sessions WHERE id = $1 AND is_active = true AND expires_at > $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, sessionID, time.Now()).Scan(&count)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to check session validity")
		return false, fmt.Errorf("failed to check session validity: %w", err)
	}

	return count > 0, nil
}

// UpdateSessionLastUsed updates the last used timestamp for a session (alias for UpdateLastActivity)
func (r *SessionRepository) UpdateSessionLastUsed(ctx context.Context, sessionID string) error {
	return r.UpdateLastActivity(ctx, sessionID)
}
