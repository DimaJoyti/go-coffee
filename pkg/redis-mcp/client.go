package redismcp

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	// Parse Redis URL if provided
	var opts *redis.Options
	var err error
	
	if cfg.URL != "" {
		opts, err = redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
	} else {
		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		}
	}

	// Set additional options
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.DialTimeout != "" {
		if timeout, err := time.ParseDuration(cfg.DialTimeout); err == nil {
			opts.DialTimeout = timeout
		}
	}
	if cfg.ReadTimeout != "" {
		if timeout, err := time.ParseDuration(cfg.ReadTimeout); err == nil {
			opts.ReadTimeout = timeout
		}
	}
	if cfg.WriteTimeout != "" {
		if timeout, err := time.ParseDuration(cfg.WriteTimeout); err == nil {
			opts.WriteTimeout = timeout
		}
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		logger: logger.New("redis-client"),
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

// Set sets a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

// Del deletes one or more keys
func (r *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Exists checks if keys exist
func (r *RedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Exists(ctx, keys...)
}

// HSet sets field in hash
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// HGet gets field from hash
func (r *RedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return r.client.HGet(ctx, key, field)
}

// HGetAll gets all fields from hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return r.client.HGetAll(ctx, key)
}

// LPush pushes elements to the head of list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.LPush(ctx, key, values...)
}

// RPush pushes elements to the tail of list
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.RPush(ctx, key, values...)
}

// LPop pops element from the head of list
func (r *RedisClient) LPop(ctx context.Context, key string) *redis.StringCmd {
	return r.client.LPop(ctx, key)
}

// RPop pops element from the tail of list
func (r *RedisClient) RPop(ctx context.Context, key string) *redis.StringCmd {
	return r.client.RPop(ctx, key)
}

// LRange gets range of elements from list
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return r.client.LRange(ctx, key, start, stop)
}

// SAdd adds members to set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return r.client.SAdd(ctx, key, members...)
}

// SMembers gets all members of set
func (r *RedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return r.client.SMembers(ctx, key)
}

// SIsMember checks if member is in set
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return r.client.SIsMember(ctx, key, member)
}

// ZAdd adds members to sorted set
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	return r.client.ZAdd(ctx, key, members...)
}

// ZRange gets range of members from sorted set
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return r.client.ZRange(ctx, key, start, stop)
}

// ZRangeByScore gets members by score range from sorted set
func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return r.client.ZRangeByScore(ctx, key, opt)
}

// Publish publishes message to channel
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return r.client.Publish(ctx, channel, message)
}

// Subscribe subscribes to channels
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Pipeline creates a pipeline for batch operations
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline creates a transaction pipeline
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Eval executes Lua script
func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.Eval(ctx, script, keys, args...)
}

// EvalSha executes Lua script by SHA
func (r *RedisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.EvalSha(ctx, sha1, keys, args...)
}

// ScriptLoad loads Lua script
func (r *RedisClient) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	return r.client.ScriptLoad(ctx, script)
}

// Keys finds keys matching pattern
func (r *RedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	return r.client.Keys(ctx, pattern)
}

// Scan scans keys with cursor
func (r *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return r.client.Scan(ctx, cursor, match, count)
}

// TTL gets time to live for key
func (r *RedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return r.client.TTL(ctx, key)
}

// Expire sets expiration for key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// Ping pings Redis server
func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}

// Info gets Redis server info
func (r *RedisClient) Info(ctx context.Context, section ...string) *redis.StringCmd {
	return r.client.Info(ctx, section...)
}

// FlushDB flushes current database
func (r *RedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	return r.client.FlushDB(ctx)
}

// FlushAll flushes all databases
func (r *RedisClient) FlushAll(ctx context.Context) *redis.StatusCmd {
	return r.client.FlushAll(ctx)
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
