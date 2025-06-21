package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisSessionRepository implements SessionRepository using Redis
type RedisSessionRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisSessionRepository creates a new Redis session repository
func NewRedisSessionRepository(client *redis.Client, logger *logger.Logger) *RedisSessionRepository {
	return &RedisSessionRepository{
		client: client,
		logger: logger,
	}
}

// Redis key patterns for sessions
const (
	sessionKeyPattern        = "auth:sessions:%s"           // auth:sessions:{sessionID}
	accessTokenKeyPattern    = "auth:access_tokens:%s"      // auth:access_tokens:{accessToken}
	refreshTokenKeyPattern   = "auth:refresh_tokens:%s"     // auth:refresh_tokens:{refreshToken}
	userSessionsKeyPattern   = "auth:user_sessions:%s"      // auth:user_sessions:{userID}
	expiredSessionsKey       = "auth:expired_sessions"      // Set of expired session IDs
)

// CreateSession creates a new session in Redis
func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	sessionKey := fmt.Sprintf(sessionKeyPattern, session.ID)
	accessTokenKey := fmt.Sprintf(accessTokenKeyPattern, session.AccessToken)
	refreshTokenKey := fmt.Sprintf(refreshTokenKeyPattern, session.RefreshToken)
	userSessionsKey := fmt.Sprintf(userSessionsKeyPattern, session.UserID)

	// Check if session already exists
	exists, err := r.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		r.logger.Error("Failed to check session existence: %v", err)
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("session already exists")
	}

	// Serialize session data
	sessionData, err := json.Marshal(session)
	if err != nil {
		r.logger.Error("Failed to marshal session data: %v", err)
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Calculate TTL for session
	sessionTTL := time.Until(session.RefreshExpiresAt)
	if sessionTTL <= 0 {
		return fmt.Errorf("session is already expired")
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	
	// Store session data with TTL
	pipe.Set(ctx, sessionKey, sessionData, sessionTTL)
	
	// Map access token to session ID
	pipe.Set(ctx, accessTokenKey, session.ID, time.Until(session.ExpiresAt))
	
	// Map refresh token to session ID
	pipe.Set(ctx, refreshTokenKey, session.ID, sessionTTL)
	
	// Add session to user's session set
	pipe.SAdd(ctx, userSessionsKey, session.ID)
	pipe.Expire(ctx, userSessionsKey, sessionTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to create session: %v", err)
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Info("Session created successfully")
	return nil
}

// GetSessionByID retrieves a session by ID from Redis
func (r *RedisSessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	sessionKey := fmt.Sprintf(sessionKeyPattern, sessionID)

	sessionData, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.Error("Failed to get session by ID: %v", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		r.logger.Error("Failed to unmarshal session data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

// GetSessionByAccessToken retrieves a session by access token from Redis
func (r *RedisSessionRepository) GetSessionByAccessToken(ctx context.Context, accessToken string) (*domain.Session, error) {
	accessTokenKey := fmt.Sprintf(accessTokenKeyPattern, accessToken)

	// Get session ID from access token mapping
	sessionID, err := r.client.Get(ctx, accessTokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.Error("Failed to get session ID by access token: %v", err)
		return nil, fmt.Errorf("failed to get session ID by access token: %w", err)
	}

	// Get session by ID
	return r.GetSessionByID(ctx, sessionID)
}

// GetSessionByRefreshToken retrieves a session by refresh token from Redis
func (r *RedisSessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	refreshTokenKey := fmt.Sprintf(refreshTokenKeyPattern, refreshToken)

	// Get session ID from refresh token mapping
	sessionID, err := r.client.Get(ctx, refreshTokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		r.logger.Error("Failed to get session ID by refresh token: %v", err)
		return nil, fmt.Errorf("failed to get session ID by refresh token: %w", err)
	}

	// Get session by ID
	return r.GetSessionByID(ctx, sessionID)
}

// UpdateSession updates a session in Redis
func (r *RedisSessionRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	sessionKey := fmt.Sprintf(sessionKeyPattern, session.ID)

	// Check if session exists
	exists, err := r.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		r.logger.Error("Failed to check session existence: %v", err)
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists == 0 {
		return domain.ErrSessionNotFound
	}

	// Update timestamp
	session.UpdatedAt = time.Now()

	// Serialize session data
	sessionData, err := json.Marshal(session)
	if err != nil {
		r.logger.Error("Failed to marshal session data: %v", err)
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Calculate TTL for session
	sessionTTL := time.Until(session.RefreshExpiresAt)
	if sessionTTL <= 0 {
		// Session is expired, mark it as expired
		session.Expire()
		sessionTTL = time.Hour // Keep expired session for a short time for audit
	}

	// Update session data
	err = r.client.Set(ctx, sessionKey, sessionData, sessionTTL).Err()
	if err != nil {
		r.logger.Error("Failed to update session: %v", err)
		return fmt.Errorf("failed to update session: %w", err)
	}

	r.logger.Info("Session updated successfully")
	return nil
}

// DeleteSession deletes a session from Redis
func (r *RedisSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	// Get session first to get tokens and user ID for cleanup
	session, err := r.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return nil // Already deleted
		}
		return err
	}

	sessionKey := fmt.Sprintf(sessionKeyPattern, sessionID)
	accessTokenKey := fmt.Sprintf(accessTokenKeyPattern, session.AccessToken)
	refreshTokenKey := fmt.Sprintf(refreshTokenKeyPattern, session.RefreshToken)
	userSessionsKey := fmt.Sprintf(userSessionsKeyPattern, session.UserID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.Del(ctx, accessTokenKey)
	pipe.Del(ctx, refreshTokenKey)
	pipe.SRem(ctx, userSessionsKey, sessionID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete session: %v", err)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	r.logger.Info("Session deleted successfully")
	return nil
}

// GetUserSessions retrieves all sessions for a user from Redis
func (r *RedisSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	userSessionsKey := fmt.Sprintf(userSessionsKeyPattern, userID)

	// Get all session IDs for the user
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		r.logger.Error("Failed to get user session IDs: %v", err)
		return nil, fmt.Errorf("failed to get user session IDs: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*domain.Session{}, nil
	}

	// Get all sessions
	sessions := make([]*domain.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := r.GetSessionByID(ctx, sessionID)
		if err != nil {
			if err == domain.ErrSessionNotFound {
				// Session was deleted, remove from user sessions set
				r.client.SRem(ctx, userSessionsKey, sessionID)
				continue
			}
			r.logger.Error("Failed to get session: %v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteUserSessions deletes all sessions for a user from Redis
func (r *RedisSessionRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	sessions, err := r.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := r.DeleteSession(ctx, session.ID); err != nil {
			r.logger.Error("Failed to delete user session: %v", err)
		}
	}

	r.logger.Info("All user sessions deleted")
	return nil
}

// DeleteExpiredSessions deletes expired sessions from Redis
func (r *RedisSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	// This is a cleanup operation that would typically be run periodically
	// Redis TTL will automatically clean up expired keys, but we can also
	// implement manual cleanup for better control

	// Get all session keys
	pattern := fmt.Sprintf(sessionKeyPattern, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("Failed to get session keys for cleanup: %v", err)
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	expiredCount := 0
	for _, key := range keys {
		// Check TTL
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// If TTL is negative, key is expired or doesn't exist
		if ttl < 0 {
			sessionID := key[len("auth:sessions:"):]
			if err := r.DeleteSession(ctx, sessionID); err != nil {
				r.logger.Error("Failed to delete expired session: %v", err)
			} else {
				expiredCount++
			}
		}
	}

	r.logger.Info("Expired sessions cleanup completed")
	return nil
}

// RevokeSession revokes a session
func (r *RedisSessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	session, err := r.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Revoke()
	return r.UpdateSession(ctx, session)
}

// RevokeUserSessions revokes all sessions for a user
func (r *RedisSessionRepository) RevokeUserSessions(ctx context.Context, userID string) error {
	sessions, err := r.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		session.Revoke()
		if err := r.UpdateSession(ctx, session); err != nil {
			r.logger.Error("Failed to revoke user session: %v", err)
		}
	}

	r.logger.Info("All user sessions revoked")
	return nil
}

// IsSessionValid checks if a session is valid
func (r *RedisSessionRepository) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
	session, err := r.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return false, nil
		}
		return false, err
	}

	return session.IsValid(), nil
}

// UpdateSessionLastUsed updates the last used timestamp of a session
func (r *RedisSessionRepository) UpdateSessionLastUsed(ctx context.Context, sessionID string) error {
	session, err := r.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	session.UpdateLastUsed()
	return r.UpdateSession(ctx, session)
}
