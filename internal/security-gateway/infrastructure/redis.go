package infrastructure

import (
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
)

type RedisServices struct {
	Client      *redis.Client
	EventStore  *RedisEventStore
	AlertMgr    *RedisAlertManager
}

func NewRedisServices(config *RedisConfig, logger *logger.Logger) (*RedisServices, error) {
	client, err := NewRedisClient(config)
	if err != nil {
		return nil, err
	}

	return &RedisServices{
		Client:      client,
		EventStore:  NewRedisEventStore(client, logger),
		AlertMgr:    NewRedisAlertManager(client, logger),
	}, nil
}
