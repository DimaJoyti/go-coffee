package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisUserRepository implements UserRepository using Redis
type RedisUserRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisUserRepository creates a new Redis user repository
func NewRedisUserRepository(client *redis.Client, logger *logger.Logger) *RedisUserRepository {
	return &RedisUserRepository{
		client: client,
		logger: logger,
	}
}

// Redis key patterns
const (
	userKeyPattern        = "auth:users:%s"        // auth:users:{userID}
	userEmailKeyPattern   = "auth:users:email:%s"  // auth:users:email:{email}
	failedLoginKeyPattern = "auth:failed_login:%s" // auth:failed_login:{email}
)

// CreateUser creates a new user in Redis
func (r *RedisUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	userKey := fmt.Sprintf(userKeyPattern, user.ID)
	emailKey := fmt.Sprintf(userEmailKeyPattern, user.Email)

	// Check if user already exists
	exists, err := r.client.Exists(ctx, userKey).Result()
	if err != nil {
		r.logger.Error("Failed to check user existence", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists > 0 {
		return domain.ErrUserExists
	}

	// Check if email already exists
	exists, err = r.client.Exists(ctx, emailKey).Result()
	if err != nil {
		r.logger.Error("Failed to check email existence", zap.Error(err), zap.String("email", user.Email))
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists > 0 {
		return domain.ErrUserExists
	}

	// Serialize user data
	userData, err := json.Marshal(user)
	if err != nil {
		r.logger.Error("Failed to marshal user data", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	pipe.Set(ctx, userKey, userData, 0)
	pipe.Set(ctx, emailKey, user.ID, 0) // Map email to user ID

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created successfully", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return nil
}

// GetUserByID retrieves a user by ID from Redis
func (r *RedisUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	userKey := fmt.Sprintf(userKeyPattern, userID)

	userData, err := r.client.Get(ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user domain.User
	if err := json.Unmarshal([]byte(userData), &user); err != nil {
		r.logger.Error("Failed to unmarshal user data", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email from Redis
func (r *RedisUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	emailKey := fmt.Sprintf(userEmailKeyPattern, email)

	// Get user ID from email mapping
	userID, err := r.client.Get(ctx, emailKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("Failed to get user ID by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user ID by email: %w", err)
	}

	// Get user by ID
	return r.GetUserByID(ctx, userID)
}

// UpdateUser updates a user in Redis
func (r *RedisUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	userKey := fmt.Sprintf(userKeyPattern, user.ID)

	// Check if user exists
	exists, err := r.client.Exists(ctx, userKey).Result()
	if err != nil {
		r.logger.Error("Failed to check user existence", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists == 0 {
		return domain.ErrUserNotFound
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Serialize user data
	userData, err := json.Marshal(user)
	if err != nil {
		r.logger.Error("Failed to marshal user data", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Update user data
	err = r.client.Set(ctx, userKey, userData, 0).Err()
	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Info("User updated successfully", zap.String("user_id", user.ID))
	return nil
}

// DeleteUser deletes a user from Redis
func (r *RedisUserRepository) DeleteUser(ctx context.Context, userID string) error {
	// Get user first to get email for cleanup
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	userKey := fmt.Sprintf(userKeyPattern, userID)
	emailKey := fmt.Sprintf(userEmailKeyPattern, user.Email)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	pipe.Del(ctx, userKey)
	pipe.Del(ctx, emailKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.Info("User deleted successfully", zap.String("user_id", userID))
	return nil
}

// LockUser locks a user account until the specified time
func (r *RedisUserRepository) LockUser(ctx context.Context, userID string, until time.Time) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Lock(until, "Account locked due to security policy")
	return r.UpdateUser(ctx, user)
}

// UnlockUser unlocks a user account
func (r *RedisUserRepository) UnlockUser(ctx context.Context, userID string) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Unlock()
	return r.UpdateUser(ctx, user)
}

// IncrementFailedLogin increments the failed login count for a user
func (r *RedisUserRepository) IncrementFailedLogin(ctx context.Context, userID string) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IncrementFailedLogin()
	return r.UpdateUser(ctx, user)
}

// ResetFailedLogin resets the failed login count for a user
func (r *RedisUserRepository) ResetFailedLogin(ctx context.Context, userID string) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.ResetFailedLogin()
	return r.UpdateUser(ctx, user)
}

// GetFailedLoginCount gets the failed login count for an email
func (r *RedisUserRepository) GetFailedLoginCount(ctx context.Context, email string) (int, error) {
	failedLoginKey := fmt.Sprintf(failedLoginKeyPattern, email)

	countStr, err := r.client.Get(ctx, failedLoginKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		r.logger.Error("Failed to get failed login count", zap.Error(err), zap.String("email", email))
		return 0, fmt.Errorf("failed to get failed login count: %w", err)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		r.logger.Error("Failed to parse failed login count", zap.Error(err), zap.String("email", email))
		return 0, fmt.Errorf("failed to parse failed login count: %w", err)
	}

	return count, nil
}

// SetFailedLoginCount sets the failed login count for an email with expiry
func (r *RedisUserRepository) SetFailedLoginCount(ctx context.Context, email string, count int, expiry time.Duration) error {
	failedLoginKey := fmt.Sprintf(failedLoginKeyPattern, email)

	err := r.client.Set(ctx, failedLoginKey, count, expiry).Err()
	if err != nil {
		r.logger.Error("Failed to set failed login count", zap.Error(err), zap.String("email", email))
		return fmt.Errorf("failed to set failed login count: %w", err)
	}

	return nil
}

// UserExists checks if a user exists by email
func (r *RedisUserRepository) UserExists(ctx context.Context, email string) (bool, error) {
	emailKey := fmt.Sprintf(userEmailKeyPattern, email)

	exists, err := r.client.Exists(ctx, emailKey).Result()
	if err != nil {
		r.logger.Error("Failed to check user existence by email", zap.Error(err), zap.String("email", email))
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists > 0, nil
}

// UserExistsByID checks if a user exists by ID
func (r *RedisUserRepository) UserExistsByID(ctx context.Context, userID string) (bool, error) {
	userKey := fmt.Sprintf(userKeyPattern, userID)

	exists, err := r.client.Exists(ctx, userKey).Result()
	if err != nil {
		r.logger.Error("Failed to check user existence by ID", zap.Error(err), zap.String("user_id", userID))
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists > 0, nil
}
