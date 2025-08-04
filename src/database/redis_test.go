package database

import (
	"context"
	"testing"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCacheClient_SingleInstance(t *testing.T) {
	// Test single instance configuration
	config := &config.CacheConfig{
		Host:        "localhost",
		Port:        6379,
		Password:    "",
		DB:          0,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		UseSentinel: false,
	}

	redisDB, err := NewCacheClient(config)

	// Note: This test will fail if Redis is not running locally
	// In a real test environment, you would use a test container or mock
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, redisDB)
	require.NotNil(t, redisDB.Client)

	// Test basic operations using direct client access
	ctx := context.Background()

	// Test Set and Get
	err = redisDB.Client.Set(ctx, "test_key", "test_value", time.Minute).Err()
	assert.NoError(t, err)

	val, err := redisDB.Client.Get(ctx, "test_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "test_value", val)

	// Test TTL
	ttl, err := redisDB.Client.TTL(ctx, "test_key").Result()
	assert.NoError(t, err)
	assert.True(t, ttl > 0)

	// Test Del
	deleted, err := redisDB.Client.Del(ctx, "test_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// Test Ping
	pong, err := redisDB.Client.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)

	// Clean up
	err = redisDB.Close()
	assert.NoError(t, err)
}

func TestNewCacheClient_Sentinel(t *testing.T) {
	// Test Sentinel configuration
	config := &config.CacheConfig{
		Host:             "",
		Port:             0,
		Password:         "",
		DB:               0,
		MaxIdle:          10,
		IdleTimeout:      240 * time.Second,
		UseSentinel:      true,
		SentinelAddrs:    []string{"localhost:26379", "localhost:26380", "localhost:26381"},
		SentinelPassword: "",
		MasterName:       "mymaster",
	}

	redisDB, err := NewCacheClient(config)

	// Note: This test will fail if Redis Sentinel is not running locally
	// In a real test environment, you would use a test container or mock
	if err != nil {
		t.Skipf("Redis Sentinel not available for testing: %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, redisDB)

	// Test basic operations
	ctx := context.Background()

	// Test Ping
	pong, err := redisDB.Client.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)

	// Clean up
	err = redisDB.Close()
	assert.NoError(t, err)
}

func TestRedisDatabase_HashOperations(t *testing.T) {
	config := &config.CacheConfig{
		Host:        "localhost",
		Port:        6379,
		Password:    "",
		DB:          0,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		UseSentinel: false,
	}

	redisDB, err := NewCacheClient(config)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
		return
	}
	defer redisDB.Close()

	ctx := context.Background()
	hashKey := "test_hash"

	// Test HSet
	err = redisDB.Client.HSet(ctx, hashKey, "field1", "value1", "field2", "value2").Err()
	assert.NoError(t, err)

	// Test HGet
	val, err := redisDB.Client.HGet(ctx, hashKey, "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test HGetAll
	all, err := redisDB.Client.HGetAll(ctx, hashKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"field1": "value1", "field2": "value2"}, all)

	// Test HDel
	deleted, err := redisDB.Client.HDel(ctx, hashKey, "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// Clean up
	redisDB.Client.Del(ctx, hashKey)
}

func TestRedisDatabase_ListOperations(t *testing.T) {
	config := &config.CacheConfig{
		Host:        "localhost",
		Port:        6379,
		Password:    "",
		DB:          0,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		UseSentinel: false,
	}

	redisDB, err := NewCacheClient(config)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
		return
	}
	defer redisDB.Close()

	ctx := context.Background()
	listKey := "test_list"

	// Test LPush
	length, err := redisDB.Client.LPush(ctx, listKey, "item1", "item2").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), length)

	// Test RPush
	length, err = redisDB.Client.RPush(ctx, listKey, "item3").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// Test LLen
	length, err = redisDB.Client.LLen(ctx, listKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// Test LPop
	val, err := redisDB.Client.LPop(ctx, listKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, "item2", val)

	// Test RPop
	val, err = redisDB.Client.RPop(ctx, listKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, "item3", val)

	// Clean up
	redisDB.Client.Del(ctx, listKey)
}
