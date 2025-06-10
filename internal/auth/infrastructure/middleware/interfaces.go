package middleware

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application/queries"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
)

// JWTService defines the interface for JWT token operations
type JWTService interface {
	ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error)
}

// SessionCacheService defines the interface for session cache operations
type SessionCacheService interface {
	GetSessionByAccessToken(ctx context.Context, accessToken string) (*domain.Session, error)
	UpdateSession(ctx context.Context, session *domain.Session) error
}

// QueryBus interface for handling queries
type QueryBus interface {
	Handle(ctx context.Context, query queries.Query) (interface{}, error)
}

// RateLimiter defines the interface for rate limiting operations
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
