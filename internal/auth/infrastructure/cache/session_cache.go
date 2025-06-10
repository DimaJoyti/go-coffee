package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// SessionCacheService provides Redis-based session caching
type SessionCacheService struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
	ttl    time.Duration
}

// NewSessionCacheService creates a new session cache service
func NewSessionCacheService(client *redis.Client, logger *logger.Logger) *SessionCacheService {
	return &SessionCacheService{
		client: client,
		logger: logger,
		prefix: "auth:sessions:",
		ttl:    24 * time.Hour, // Default TTL
	}
}

// StoreSession stores a session in Redis
func (scs *SessionCacheService) StoreSession(ctx context.Context, session *domain.Session) error {
	key := scs.getSessionKey(session.ID)
	
	// Serialize session
	data, err := json.Marshal(session)
	if err != nil {
		scs.logger.ErrorWithFields("Failed to marshal session", 
			logger.Error(err), 
			logger.String("session_id", session.ID))
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Calculate TTL based on session expiry
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = scs.ttl
	}

	// Store session
	err = scs.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		scs.logger.ErrorWithFields("Failed to store session", 
			logger.Error(err), 
			logger.String("session_id", session.ID))
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Add to user sessions set
	userSessionsKey := scs.getUserSessionsKey(session.UserID)
	err = scs.client.SAdd(ctx, userSessionsKey, session.ID).Err()
	if err != nil {
		scs.logger.ErrorWithFields("Failed to add session to user sessions", 
			logger.Error(err), 
			logger.String("session_id", session.ID),
			logger.String("user_id", session.UserID))
		// Don't return error for this operation
	}

	// Set expiry for user sessions set
	scs.client.Expire(ctx, userSessionsKey, 30*24*time.Hour) // 30 days

	// Store access token mapping
	if session.AccessToken != "" {
		accessTokenKey := scs.getAccessTokenKey(session.AccessToken)
		err = scs.client.Set(ctx, accessTokenKey, session.ID, ttl).Err()
		if err != nil {
			scs.logger.ErrorWithFields("Failed to store access token mapping", 
				logger.Error(err), 
				logger.String("session_id", session.ID))
			// Don't return error for this operation
		}
	}

	// Store refresh token mapping
	if session.RefreshToken != "" {
		refreshTokenKey := scs.getRefreshTokenKey(session.RefreshToken)
		refreshTTL := time.Until(session.RefreshExpiresAt)
		if refreshTTL <= 0 {
			refreshTTL = 7 * 24 * time.Hour // Default refresh TTL
		}
		
		err = scs.client.Set(ctx, refreshTokenKey, session.ID, refreshTTL).Err()
		if err != nil {
			scs.logger.ErrorWithFields("Failed to store refresh token mapping", 
				logger.Error(err), 
				logger.String("session_id", session.ID))
			// Don't return error for this operation
		}
	}

	scs.logger.InfoWithFields("Session stored successfully", 
		logger.String("session_id", session.ID),
		logger.String("user_id", session.UserID))

	return nil
}

// GetSession retrieves a session from Redis
func (scs *SessionCacheService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := scs.getSessionKey(sessionID)
	
	data, err := scs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		scs.logger.ErrorWithFields("Failed to get session", 
			logger.Error(err), 
			logger.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session domain.Session
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		scs.logger.ErrorWithFields("Failed to unmarshal session", 
			logger.Error(err), 
			logger.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetSessionByAccessToken retrieves a session by access token
func (scs *SessionCacheService) GetSessionByAccessToken(ctx context.Context, accessToken string) (*domain.Session, error) {
	accessTokenKey := scs.getAccessTokenKey(accessToken)
	
	sessionID, err := scs.client.Get(ctx, accessTokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		scs.logger.ErrorWithFields("Failed to get session ID by access token", logger.Error(err))
		return nil, fmt.Errorf("failed to get session ID by access token: %w", err)
	}

	return scs.GetSession(ctx, sessionID)
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (scs *SessionCacheService) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	refreshTokenKey := scs.getRefreshTokenKey(refreshToken)
	
	sessionID, err := scs.client.Get(ctx, refreshTokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		scs.logger.ErrorWithFields("Failed to get session ID by refresh token", logger.Error(err))
		return nil, fmt.Errorf("failed to get session ID by refresh token: %w", err)
	}

	return scs.GetSession(ctx, sessionID)
}

// GetUserSessions retrieves all sessions for a user
func (scs *SessionCacheService) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	userSessionsKey := scs.getUserSessionsKey(userID)
	
	sessionIDs, err := scs.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		scs.logger.ErrorWithFields("Failed to get user session IDs", 
			logger.Error(err), 
			logger.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user session IDs: %w", err)
	}

	sessions := make([]*domain.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := scs.GetSession(ctx, sessionID)
		if err != nil {
			if err == domain.ErrSessionNotFound {
				// Remove expired session from set
				scs.client.SRem(ctx, userSessionsKey, sessionID)
				continue
			}
			scs.logger.ErrorWithFields("Failed to get user session", 
				logger.Error(err), 
				logger.String("session_id", sessionID))
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateSession updates a session in Redis
func (scs *SessionCacheService) UpdateSession(ctx context.Context, session *domain.Session) error {
	// Remove old token mappings if tokens changed
	oldSession, err := scs.GetSession(ctx, session.ID)
	if err == nil {
		if oldSession.AccessToken != session.AccessToken && oldSession.AccessToken != "" {
			oldAccessTokenKey := scs.getAccessTokenKey(oldSession.AccessToken)
			scs.client.Del(ctx, oldAccessTokenKey)
		}
		if oldSession.RefreshToken != session.RefreshToken && oldSession.RefreshToken != "" {
			oldRefreshTokenKey := scs.getRefreshTokenKey(oldSession.RefreshToken)
			scs.client.Del(ctx, oldRefreshTokenKey)
		}
	}

	// Store updated session
	return scs.StoreSession(ctx, session)
}

// DeleteSession deletes a session from Redis
func (scs *SessionCacheService) DeleteSession(ctx context.Context, sessionID string) error {
	// Get session first to clean up token mappings
	session, err := scs.GetSession(ctx, sessionID)
	if err != nil && err != domain.ErrSessionNotFound {
		scs.logger.ErrorWithFields("Failed to get session for deletion", 
			logger.Error(err), 
			logger.String("session_id", sessionID))
	}

	// Delete session
	key := scs.getSessionKey(sessionID)
	err = scs.client.Del(ctx, key).Err()
	if err != nil {
		scs.logger.ErrorWithFields("Failed to delete session", 
			logger.Error(err), 
			logger.String("session_id", sessionID))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if session != nil {
		// Remove from user sessions set
		userSessionsKey := scs.getUserSessionsKey(session.UserID)
		scs.client.SRem(ctx, userSessionsKey, sessionID)

		// Delete token mappings
		if session.AccessToken != "" {
			accessTokenKey := scs.getAccessTokenKey(session.AccessToken)
			scs.client.Del(ctx, accessTokenKey)
		}
		if session.RefreshToken != "" {
			refreshTokenKey := scs.getRefreshTokenKey(session.RefreshToken)
			scs.client.Del(ctx, refreshTokenKey)
		}
	}

	scs.logger.InfoWithFields("Session deleted successfully", 
		logger.String("session_id", sessionID))

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (scs *SessionCacheService) DeleteUserSessions(ctx context.Context, userID string) error {
	sessions, err := scs.GetUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, session := range sessions {
		if err := scs.DeleteSession(ctx, session.ID); err != nil {
			scs.logger.ErrorWithFields("Failed to delete user session", 
				logger.Error(err), 
				logger.String("session_id", session.ID),
				logger.String("user_id", userID))
		}
	}

	// Clean up user sessions set
	userSessionsKey := scs.getUserSessionsKey(userID)
	scs.client.Del(ctx, userSessionsKey)

	scs.logger.InfoWithFields("All user sessions deleted", 
		logger.String("user_id", userID),
		logger.Int("session_count", len(sessions)))

	return nil
}

// IsSessionActive checks if a session is active
func (scs *SessionCacheService) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	key := scs.getSessionKey(sessionID)
	exists, err := scs.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return exists > 0, nil
}

// ExtendSession extends the TTL of a session
func (scs *SessionCacheService) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	key := scs.getSessionKey(sessionID)
	err := scs.client.Expire(ctx, key, duration).Err()
	if err != nil {
		scs.logger.ErrorWithFields("Failed to extend session", 
			logger.Error(err), 
			logger.String("session_id", sessionID))
		return fmt.Errorf("failed to extend session: %w", err)
	}
	return nil
}

// Helper methods

func (scs *SessionCacheService) getSessionKey(sessionID string) string {
	return fmt.Sprintf("%ssession:%s", scs.prefix, sessionID)
}

func (scs *SessionCacheService) getUserSessionsKey(userID string) string {
	return fmt.Sprintf("%suser:%s:sessions", scs.prefix, userID)
}

func (scs *SessionCacheService) getAccessTokenKey(accessToken string) string {
	return fmt.Sprintf("%saccess_token:%s", scs.prefix, accessToken)
}

func (scs *SessionCacheService) getRefreshTokenKey(refreshToken string) string {
	return fmt.Sprintf("%srefresh_token:%s", scs.prefix, refreshToken)
}
