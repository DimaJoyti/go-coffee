package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client represents a Redis client wrapper
type Client struct {
	client redis.UniversalClient
	config *config.RedisConfig
	logger *logger.Logger
}

// ClientInterface defines the Redis client interface
type ClientInterface interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Hash operations
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int64, error)

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)

	// Sorted set operations
	ZAdd(ctx context.Context, key string, members ...redis.Z) error
	ZRem(ctx context.Context, key string, members ...interface{}) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)

	// Pub/Sub operations
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub

	// Transaction operations
	TxPipeline() redis.Pipeliner
	Pipeline() redis.Pipeliner

	// Utility operations
	Ping(ctx context.Context) error
	FlushDB(ctx context.Context) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)

	// Connection management
	Close() error
	PoolStats() *redis.PoolStats

	// Raw client access for advanced operations
	GetClient() redis.UniversalClient
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig, logger *logger.Logger) (ClientInterface, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Redis configuration: %w", err)
	}

	var client redis.UniversalClient

	if cfg.ClusterMode {
		client = createClusterClient(cfg)
	} else if cfg.SentinelMode {
		client = createSentinelClient(cfg)
	} else {
		client = createStandaloneClient(cfg)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")

	return &Client{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// createStandaloneClient creates a standalone Redis client
func createStandaloneClient(cfg *config.RedisConfig) redis.UniversalClient {
	opts := &redis.Options{
		Addr:            cfg.GetRedisAddr(),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = createTLSConfig(cfg)
	}

	return redis.NewClient(opts)
}

// createClusterClient creates a Redis cluster client
func createClusterClient(cfg *config.RedisConfig) redis.UniversalClient {
	opts := &redis.ClusterOptions{
		Addrs:           cfg.ClusterHosts,
		Password:        cfg.Password,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = createTLSConfig(cfg)
	}

	return redis.NewClusterClient(opts)
}

// createSentinelClient creates a Redis sentinel client
func createSentinelClient(cfg *config.RedisConfig) redis.UniversalClient {
	opts := &redis.FailoverOptions{
		MasterName:       cfg.SentinelMaster,
		SentinelAddrs:    cfg.SentinelHosts,
		SentinelPassword: cfg.SentinelPassword,
		Password:         cfg.Password,
		DB:               cfg.DB,
		PoolSize:         cfg.PoolSize,
		MinIdleConns:     cfg.MinIdleConns,
		MaxRetries:       cfg.MaxRetries,
		DialTimeout:      cfg.DialTimeout,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
		ConnMaxIdleTime:  cfg.IdleTimeout,
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = createTLSConfig(cfg)
	}

	return redis.NewFailoverClient(opts)
}

// createTLSConfig creates TLS configuration
func createTLSConfig(cfg *config.RedisConfig) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSSkipVerify,
	}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return tlsConfig
}

// addKeyPrefix adds the configured key prefix
func (c *Client) addKeyPrefix(key string) string {
	if c.config.KeyPrefix == "" {
		return key
	}
	return c.config.KeyPrefix + key
}

// Basic operations implementation
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.addKeyPrefix(key), value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.addKeyPrefix(key)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.addKeyPrefix(key)
	}
	return c.client.Exists(ctx, prefixedKeys...).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, c.addKeyPrefix(key), expiration).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.addKeyPrefix(key)).Result()
}

// Hash operations implementation
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, c.addKeyPrefix(key), field).Result()
}

func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, c.addKeyPrefix(key), values...).Err()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, c.addKeyPrefix(key), fields...).Err()
}

func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, c.addKeyPrefix(key), field).Result()
}

func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	return c.client.HLen(ctx, c.addKeyPrefix(key)).Result()
}

// List operations implementation
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, c.addKeyPrefix(key), values...).Err()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, c.addKeyPrefix(key), values...).Err()
}

func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, c.addKeyPrefix(key), start, stop).Result()
}

// Set operations implementation
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, c.addKeyPrefix(key), members...).Err()
}

func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, c.addKeyPrefix(key), members...).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, c.addKeyPrefix(key), member).Result()
}

func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, c.addKeyPrefix(key)).Result()
}

// Sorted set operations implementation
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return c.client.ZAdd(ctx, c.addKeyPrefix(key), members...).Err()
}

func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, c.addKeyPrefix(key), members...).Err()
}

func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(ctx, c.addKeyPrefix(key), start, stop).Result()
}

func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRangeWithScores(ctx, c.addKeyPrefix(key), start, stop).Result()
}

func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.client.ZCard(ctx, c.addKeyPrefix(key)).Result()
}

func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.client.ZScore(ctx, c.addKeyPrefix(key), member).Result()
}

// Pub/Sub operations implementation
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.client.Publish(ctx, c.addKeyPrefix(channel), message).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	prefixedChannels := make([]string, len(channels))
	for i, channel := range channels {
		prefixedChannels[i] = c.addKeyPrefix(channel)
	}
	return c.client.Subscribe(ctx, prefixedChannels...)
}

func (c *Client) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	prefixedPatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		prefixedPatterns[i] = c.addKeyPrefix(pattern)
	}
	return c.client.PSubscribe(ctx, prefixedPatterns...)
}

// Transaction operations implementation
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// Utility operations implementation
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, c.addKeyPrefix(pattern)).Result()
}

func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.client.Scan(ctx, cursor, c.addKeyPrefix(match), count).Result()
}

// Connection management implementation
func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

func (c *Client) GetClient() redis.UniversalClient {
	return c.client
}
