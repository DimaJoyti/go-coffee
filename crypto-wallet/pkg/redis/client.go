package redis

import (
	"context"
	"time"
)

// Config represents Redis configuration
type Config struct {
	Addresses              []string      // Redis server addresses (host:port)
	Host                   string        // Redis host (for single instance)
	Port                   int           // Redis port (for single instance)
	Password               string        // Redis password
	DB                     int           // Redis database
	PoolSize               int           // Connection pool size
	MinIdleConns           int           // Minimum number of idle connections
	DialTimeout            time.Duration // Dial timeout
	ReadTimeout            time.Duration // Read timeout
	WriteTimeout           time.Duration // Write timeout
	PoolTimeout            time.Duration // Pool timeout
	IdleTimeout            time.Duration // Idle timeout
	IdleCheckFrequency     time.Duration // Idle check frequency
	MaxRetries             int           // Maximum number of retries
	MinRetryBackoff        time.Duration // Minimum retry backoff
	MaxRetryBackoff        time.Duration // Maximum retry backoff
	EnableCluster          bool          // Whether to use Redis cluster
	RouteByLatency         bool          // Whether to route by latency
	RouteRandomly          bool          // Whether to route randomly
	EnableReadFromReplicas bool          // Whether to enable read from replicas
}

// Client represents a Redis client
type Client interface {
	// Get gets a value from Redis
	Get(ctx context.Context, key string) (string, error)

	// Set sets a value in Redis
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Del deletes keys from Redis
	Del(ctx context.Context, keys ...string) error

	// Exists checks if keys exist in Redis
	Exists(ctx context.Context, keys ...string) (bool, error)

	// Incr increments a key in Redis
	Incr(ctx context.Context, key string) (int64, error)

	// HGet gets a field from a hash in Redis
	HGet(ctx context.Context, key, field string) (string, error)

	// HSet sets fields in a hash in Redis
	HSet(ctx context.Context, key string, values ...interface{}) error

	// HGetAll gets all fields from a hash in Redis
	HGetAll(ctx context.Context, key string) (map[string]string, error)

	// HDel deletes fields from a hash in Redis
	HDel(ctx context.Context, key string, fields ...string) error

	// Expire sets an expiration on a key in Redis
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Pipeline returns a Redis pipeline
	Pipeline() Pipeline

	// Close closes the Redis client
	Close() error

	// Ping checks the Redis connection
	Ping(ctx context.Context) error
}

// Pipeline represents a Redis pipeline
type Pipeline interface {
	// Get gets a value from Redis
	Get(ctx context.Context, key string) StringCmd

	// Set sets a value in Redis
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd

	// Del deletes keys from Redis
	Del(ctx context.Context, keys ...string) IntCmd

	// Exists checks if keys exist in Redis
	Exists(ctx context.Context, keys ...string) IntCmd

	// Incr increments a key in Redis
	Incr(ctx context.Context, key string) IntCmd

	// HGet gets a field from a hash in Redis
	HGet(ctx context.Context, key, field string) StringCmd

	// HSet sets fields in a hash in Redis
	HSet(ctx context.Context, key string, values ...interface{}) IntCmd

	// HGetAll gets all fields from a hash in Redis
	HGetAll(ctx context.Context, key string) StringStringMapCmd

	// HDel deletes fields from a hash in Redis
	HDel(ctx context.Context, key string, fields ...string) IntCmd

	// Expire sets an expiration on a key in Redis
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd

	// Exec executes the pipeline
	Exec(ctx context.Context) ([]Cmder, error)
}

// Cmder represents a Redis command
type Cmder interface {
	// Name returns the command name
	Name() string

	// Args returns the command arguments
	Args() []interface{}

	// Err returns the command error
	Err() error

	// String returns the command string
	String() string
}

// StringCmd represents a Redis string command
type StringCmd interface {
	Cmder

	// Result returns the command result
	Result() (string, error)
}

// StatusCmd represents a Redis status command
type StatusCmd interface {
	Cmder

	// Result returns the command result
	Result() (string, error)
}

// IntCmd represents a Redis integer command
type IntCmd interface {
	Cmder

	// Result returns the command result
	Result() (int64, error)
}

// BoolCmd represents a Redis boolean command
type BoolCmd interface {
	Cmder

	// Result returns the command result
	Result() (bool, error)
}

// StringStringMapCmd represents a Redis string string map command
type StringStringMapCmd interface {
	Cmder

	// Result returns the command result
	Result() (map[string]string, error)
}
