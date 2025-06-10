package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/go-redis/redis/v8"
)

// redisClient implements the Client interface
type redisClient struct {
	client redis.UniversalClient
	config *Config
}

// NewClient creates a new Redis client
func NewClient(config *Config) (Client, error) {
	// Set default addresses if not provided
	if len(config.Addresses) == 0 && config.Host != "" {
		config.Addresses = []string{fmt.Sprintf("%s:%d", config.Host, config.Port)}
	}
	if len(config.Addresses) == 0 {
		config.Addresses = []string{"localhost:6379"}
	}
	var client redis.UniversalClient

	if config.EnableCluster {
		// Create a Redis cluster client
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:              config.Addresses,
			Password:           config.Password,
			PoolSize:           config.PoolSize,
			MinIdleConns:       config.MinIdleConns,
			DialTimeout:        config.DialTimeout,
			ReadTimeout:        config.ReadTimeout,
			WriteTimeout:       config.WriteTimeout,
			PoolTimeout:        config.PoolTimeout,
			IdleTimeout:        config.IdleTimeout,
			IdleCheckFrequency: config.IdleCheckFrequency,
			MaxRetries:         config.MaxRetries,
			MinRetryBackoff:    config.MinRetryBackoff,
			MaxRetryBackoff:    config.MaxRetryBackoff,
			RouteByLatency:     config.RouteByLatency,
			RouteRandomly:      config.RouteRandomly,
		})
	} else {
		// Create a Redis client
		client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:              config.Addresses,
			DB:                 config.DB,
			Password:           config.Password,
			PoolSize:           config.PoolSize,
			MinIdleConns:       config.MinIdleConns,
			DialTimeout:        config.DialTimeout,
			ReadTimeout:        config.ReadTimeout,
			WriteTimeout:       config.WriteTimeout,
			PoolTimeout:        config.PoolTimeout,
			IdleTimeout:        config.IdleTimeout,
			IdleCheckFrequency: config.IdleCheckFrequency,
			MaxRetries:         config.MaxRetries,
			MinRetryBackoff:    config.MinRetryBackoff,
			MaxRetryBackoff:    config.MaxRetryBackoff,
			ReadOnly:           config.EnableReadFromReplicas,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisClient{
		client: client,
		config: config,
	}, nil
}

// Get gets a value from Redis
func (c *redisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set sets a value in Redis
func (c *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del deletes keys from Redis
func (c *redisClient) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist in Redis
func (c *redisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	result, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Incr increments a key in Redis
func (c *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// HGet gets a field from a hash in Redis
func (c *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HSet sets fields in a hash in Redis
func (c *redisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGetAll gets all fields from a hash in Redis
func (c *redisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from a hash in Redis
func (c *redisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// Expire sets an expiration on a key in Redis
func (c *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// Pipeline returns a Redis pipeline
func (c *redisClient) Pipeline() Pipeline {
	return &redisPipeline{
		pipeline: c.client.Pipeline(),
	}
}

// Close closes the Redis client
func (c *redisClient) Close() error {
	return c.client.Close()
}

// Ping checks the Redis connection
func (c *redisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// redisPipeline implements the Pipeline interface
type redisPipeline struct {
	pipeline redis.Pipeliner
}

// Get gets a value from Redis
func (p *redisPipeline) Get(ctx context.Context, key string) StringCmd {
	return &redisStringCmd{
		cmd: p.pipeline.Get(ctx, key),
	}
}

// Set sets a value in Redis
func (p *redisPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	return &redisStatusCmd{
		cmd: p.pipeline.Set(ctx, key, value, expiration),
	}
}

// Del deletes keys from Redis
func (p *redisPipeline) Del(ctx context.Context, keys ...string) IntCmd {
	return &redisIntCmd{
		cmd: p.pipeline.Del(ctx, keys...),
	}
}

// Exists checks if keys exist in Redis
func (p *redisPipeline) Exists(ctx context.Context, keys ...string) IntCmd {
	return &redisIntCmd{
		cmd: p.pipeline.Exists(ctx, keys...),
	}
}

// Incr increments a key in Redis
func (p *redisPipeline) Incr(ctx context.Context, key string) IntCmd {
	return &redisIntCmd{
		cmd: p.pipeline.Incr(ctx, key),
	}
}

// HGet gets a field from a hash in Redis
func (p *redisPipeline) HGet(ctx context.Context, key, field string) StringCmd {
	return &redisStringCmd{
		cmd: p.pipeline.HGet(ctx, key, field),
	}
}

// HSet sets fields in a hash in Redis
func (p *redisPipeline) HSet(ctx context.Context, key string, values ...interface{}) IntCmd {
	return &redisIntCmd{
		cmd: p.pipeline.HSet(ctx, key, values...),
	}
}

// HGetAll gets all fields from a hash in Redis
func (p *redisPipeline) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	return &redisStringStringMapCmd{
		cmd: p.pipeline.HGetAll(ctx, key),
	}
}

// HDel deletes fields from a hash in Redis
func (p *redisPipeline) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	return &redisIntCmd{
		cmd: p.pipeline.HDel(ctx, key, fields...),
	}
}

// Expire sets an expiration on a key in Redis
func (p *redisPipeline) Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return &redisBoolCmd{
		cmd: p.pipeline.Expire(ctx, key, expiration),
	}
}

// Exec executes the pipeline
func (p *redisPipeline) Exec(ctx context.Context) ([]Cmder, error) {
	cmds, err := p.pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Cmder, len(cmds))
	for i, cmd := range cmds {
		result[i] = &redisCmder{
			cmd: cmd,
		}
	}

	return result, nil
}

// NewClientFromConfig creates a new Redis client from config.RedisConfig
func NewClientFromConfig(cfg *config.RedisConfig) (Client, error) {
	redisConfig := &Config{
		Host:                   cfg.Host,
		Port:                   cfg.Port,
		Password:               cfg.Password,
		DB:                     cfg.DB,
		PoolSize:               cfg.PoolSize,
		MinIdleConns:           cfg.MinIdleConns,
		DialTimeout:            cfg.DialTimeout,
		ReadTimeout:            cfg.ReadTimeout,
		WriteTimeout:           cfg.WriteTimeout,
		PoolTimeout:            cfg.PoolTimeout,
		IdleTimeout:            cfg.IdleTimeout,
		IdleCheckFrequency:     cfg.IdleCheckFrequency,
		MaxRetries:             cfg.MaxRetries,
		MinRetryBackoff:        cfg.MinRetryBackoff,
		MaxRetryBackoff:        cfg.MaxRetryBackoff,
		EnableCluster:          cfg.EnableCluster,
		RouteByLatency:         cfg.RouteByLatency,
		RouteRandomly:          cfg.RouteRandomly,
		EnableReadFromReplicas: cfg.EnableReadFromReplicas,
	}

	// Copy addresses if provided
	if len(cfg.Addresses) > 0 {
		redisConfig.Addresses = make([]string, len(cfg.Addresses))
		copy(redisConfig.Addresses, cfg.Addresses)
	}

	return NewClient(redisConfig)
}
