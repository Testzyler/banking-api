package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/logger"
	"github.com/redis/go-redis/v9"
)

// RedisDatabase is a struct that holds the Redis client
type RedisDatabase struct {
	Client redis.Cmdable
	config *config.CacheConfig
}

func (r *RedisDatabase) GetClient() redis.Cmdable {
	return r.Client
}

func (r *RedisDatabase) GetConfig() *config.CacheConfig {
	return r.config
}

func NewCacheClient(config *config.CacheConfig) (*RedisDatabase, error) {
	var client redis.Cmdable

	if config.UseSentinel {
		// Redis Sentinel configuration
		logger.Infof("Initializing Redis client with Sentinel configuration")
		logger.Infof("Sentinel addresses: %v", config.SentinelAddrs)
		logger.Infof("Master name: %s", config.MasterName)

		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       config.MasterName,
			SentinelAddrs:    config.SentinelAddrs,
			SentinelPassword: config.SentinelPassword,
			Password:         config.Password,
			DB:               config.DB,
			MaxRetries:       3,
			DialTimeout:      5 * time.Second,
			ReadTimeout:      3 * time.Second,
			WriteTimeout:     3 * time.Second,
			PoolSize:         10,
			MinIdleConns:     config.MaxIdle,
			ConnMaxIdleTime:  config.IdleTimeout,
		})
	} else {
		// Single Redis instance configuration
		logger.Infof("Initializing Redis client with single instance configuration")
		logger.Infof("Redis address: %s:%d", config.Host, config.Port)

		client = redis.NewClient(&redis.Options{
			Addr:            config.Host + ":" + strconv.Itoa(config.Port),
			Password:        config.Password,
			DB:              config.DB,
			MaxRetries:      3,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolSize:        10,
			MinIdleConns:    config.MaxIdle,
			ConnMaxIdleTime: config.IdleTimeout,
		})
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Infof("Successfully connected to Redis")

	return &RedisDatabase{
		Client: client,
		config: config,
	}, nil
}

// Close closes the Redis connection
func (r *RedisDatabase) Close() error {
	if clientWithClose, ok := r.Client.(interface{ Close() error }); ok {
		return clientWithClose.Close()
	}
	return nil
}
