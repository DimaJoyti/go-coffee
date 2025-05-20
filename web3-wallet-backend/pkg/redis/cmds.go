package redis

import (
	"github.com/go-redis/redis/v8"
)

// redisCmder implements the Cmder interface
type redisCmder struct {
	cmd redis.Cmder
}

// Name returns the command name
func (c *redisCmder) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisCmder) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisCmder) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisCmder) String() string {
	return c.cmd.String()
}

// redisStringCmd implements the StringCmd interface
type redisStringCmd struct {
	cmd *redis.StringCmd
}

// Name returns the command name
func (c *redisStringCmd) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisStringCmd) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisStringCmd) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisStringCmd) String() string {
	return c.cmd.String()
}

// Result returns the command result
func (c *redisStringCmd) Result() (string, error) {
	return c.cmd.Result()
}

// redisStatusCmd implements the StatusCmd interface
type redisStatusCmd struct {
	cmd *redis.StatusCmd
}

// Name returns the command name
func (c *redisStatusCmd) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisStatusCmd) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisStatusCmd) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisStatusCmd) String() string {
	return c.cmd.String()
}

// Result returns the command result
func (c *redisStatusCmd) Result() (string, error) {
	return c.cmd.Result()
}

// redisIntCmd implements the IntCmd interface
type redisIntCmd struct {
	cmd *redis.IntCmd
}

// Name returns the command name
func (c *redisIntCmd) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisIntCmd) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisIntCmd) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisIntCmd) String() string {
	return c.cmd.String()
}

// Result returns the command result
func (c *redisIntCmd) Result() (int64, error) {
	return c.cmd.Result()
}

// redisBoolCmd implements the BoolCmd interface
type redisBoolCmd struct {
	cmd *redis.BoolCmd
}

// Name returns the command name
func (c *redisBoolCmd) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisBoolCmd) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisBoolCmd) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisBoolCmd) String() string {
	return c.cmd.String()
}

// Result returns the command result
func (c *redisBoolCmd) Result() (bool, error) {
	return c.cmd.Result()
}

// redisStringStringMapCmd implements the StringStringMapCmd interface
type redisStringStringMapCmd struct {
	cmd *redis.StringStringMapCmd
}

// Name returns the command name
func (c *redisStringStringMapCmd) Name() string {
	return c.cmd.Name()
}

// Args returns the command arguments
func (c *redisStringStringMapCmd) Args() []interface{} {
	return c.cmd.Args()
}

// Err returns the command error
func (c *redisStringStringMapCmd) Err() error {
	return c.cmd.Err()
}

// String returns the command string
func (c *redisStringStringMapCmd) String() string {
	return c.cmd.String()
}

// Result returns the command result
func (c *redisStringStringMapCmd) Result() (map[string]string, error) {
	return c.cmd.Result()
}
