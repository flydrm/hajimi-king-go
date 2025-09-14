package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements a Redis-based cache
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	ctx    context.Context
}

// NewRedisCache creates a new Redis-based cache
func NewRedisCache(redisURL string, ttl time.Duration) (*RedisCache, error) {
	// Parse Redis URL and create client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	
	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
		ctx:    ctx,
	}, nil
}

// Get retrieves a value from the Redis cache
func (rc *RedisCache) Get(key string) (interface{}, bool) {
	val, err := rc.client.Get(rc.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		return nil, false
	}

	// Unmarshal the value
	var value interface{}
	if err := json.Unmarshal([]byte(val), &value); err != nil {
		return nil, false
	}

	return value, true
}

// Set stores a value in the Redis cache
func (rc *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	// Marshal the value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Set with TTL
	if err := rc.client.Set(rc.ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set Redis key: %w", err)
	}

	return nil
}

// Delete removes a value from the Redis cache
func (rc *RedisCache) Delete(key string) error {
	if err := rc.client.Del(rc.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete Redis key: %w", err)
	}
	return nil
}

// Clear clears all keys in the Redis cache
func (rc *RedisCache) Clear() error {
	// Get all keys
	keys, err := rc.client.Keys(rc.ctx, "*").Result()
	if err != nil {
		return fmt.Errorf("failed to get Redis keys: %w", err)
	}

	// Delete all keys
	if len(keys) > 0 {
		if err := rc.client.Del(rc.ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete Redis keys: %w", err)
		}
	}

	return nil
}

// Size returns the number of keys in the Redis cache
func (rc *RedisCache) Size() int {
	keys, err := rc.client.Keys(rc.ctx, "*").Result()
	if err != nil {
		return 0
	}
	return len(keys)
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}