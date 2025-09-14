package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
)

// MultiLevelCache implements a multi-level caching system
type MultiLevelCache struct {
	l1Cache    *lru.Cache[string, *CacheItem]
	l2Cache    *FileCache
	l3Cache    *RedisCache
	config     *CacheConfig
	mutex      sync.RWMutex
	metrics    *CacheMetrics
}

// CacheItem represents a cached item
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	L1MaxSize       int           `json:"l1_max_size"`
	L1TTL           time.Duration `json:"l1_ttl"`
	L2TTL           time.Duration `json:"l2_ttl"`
	L3TTL           time.Duration `json:"l3_ttl"`
	EnableL3        bool          `json:"enable_l3"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	L2Path          string        `json:"l2_path"`
	L3RedisURL      string        `json:"l3_redis_url"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	L1Hits       int64
	L1Misses     int64
	L2Hits       int64
	L2Misses     int64
	L3Hits       int64
	L3Misses     int64
	TotalHits    int64
	TotalMisses  int64
	Evictions    int64
	SetOperations int64
	GetOperations int64
	DeleteOperations int64
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(config *CacheConfig) (*MultiLevelCache, error) {
	// Create L1 cache (in-memory LRU)
	l1Cache, err := lru.New[string, *CacheItem](config.L1MaxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 cache: %w", err)
	}

	// Create L2 cache (file-based)
	l2Cache, err := NewFileCache(config.L2Path, config.L2TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create L2 cache: %w", err)
	}

	// Create L3 cache (Redis) if enabled
	var l3Cache *RedisCache
	if config.EnableL3 {
		l3Cache, err = NewRedisCache(config.L3RedisURL, config.L3TTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create L3 cache: %w", err)
		}
	}

	mlc := &MultiLevelCache{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
		l3Cache: l3Cache,
		config:  config,
		metrics: &CacheMetrics{},
	}

	// Start cleanup goroutine
	go mlc.startCleanup()

	return mlc, nil
}

// Get retrieves a value from the cache
func (mlc *MultiLevelCache) Get(key string) (interface{}, bool) {
	mlc.mutex.Lock()
	defer mlc.mutex.Unlock()

	// Try L1 cache first
	if item, ok := mlc.l1Cache.Get(key); ok {
		if time.Now().Before(item.ExpiresAt) {
			item.AccessCount++
			mlc.metrics.L1Hits++
			mlc.metrics.TotalHits++
			return item.Value, true
		} else {
			// Expired, remove from L1
			mlc.l1Cache.Remove(key)
		}
	}
	mlc.metrics.L1Misses++

	// Try L2 cache
	if value, ok := mlc.l2Cache.Get(key); ok {
		// Promote to L1
		item := &CacheItem{
			Value:      value,
			ExpiresAt:  time.Now().Add(mlc.config.L1TTL),
			CreatedAt:  time.Now(),
			AccessCount: 1,
		}
		mlc.l1Cache.Add(key, item)
		mlc.metrics.L2Hits++
		mlc.metrics.TotalHits++
		return value, true
	}
	mlc.metrics.L2Misses++

	// Try L3 cache if enabled
	if mlc.l3Cache != nil {
		if value, ok := mlc.l3Cache.Get(key); ok {
			// Promote to L1 and L2
			item := &CacheItem{
				Value:      value,
				ExpiresAt:  time.Now().Add(mlc.config.L1TTL),
				CreatedAt:  time.Now(),
				AccessCount: 1,
			}
			mlc.l1Cache.Add(key, item)
			mlc.l2Cache.Set(key, value, mlc.config.L2TTL)
			mlc.metrics.L3Hits++
			mlc.metrics.TotalHits++
			return value, true
		}
		mlc.metrics.L3Misses++
	}

	mlc.metrics.TotalMisses++
	return nil, false
}

// Set stores a value in the cache
func (mlc *MultiLevelCache) Set(key string, value interface{}, ttl time.Duration) error {
	mlc.mutex.Lock()
	defer mlc.mutex.Unlock()

	// Create cache item
	item := &CacheItem{
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		AccessCount: 0,
	}

	// Store in L1
	mlc.l1Cache.Add(key, item)

	// Store in L2
	if err := mlc.l2Cache.Set(key, value, mlc.config.L2TTL); err != nil {
		return fmt.Errorf("failed to set L2 cache: %w", err)
	}

	// Store in L3 if enabled
	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Set(key, value, mlc.config.L3TTL); err != nil {
			return fmt.Errorf("failed to set L3 cache: %w", err)
		}
	}

	mlc.metrics.SetOperations++
	return nil
}

// Delete removes a value from the cache
func (mlc *MultiLevelCache) Delete(key string) error {
	mlc.mutex.Lock()
	defer mlc.mutex.Unlock()

	// Remove from L1
	mlc.l1Cache.Remove(key)

	// Remove from L2
	if err := mlc.l2Cache.Delete(key); err != nil {
		return fmt.Errorf("failed to delete from L2 cache: %w", err)
	}

	// Remove from L3 if enabled
	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Delete(key); err != nil {
			return fmt.Errorf("failed to delete from L3 cache: %w", err)
		}
	}

	mlc.metrics.DeleteOperations++
	return nil
}

// Clear clears all caches
func (mlc *MultiLevelCache) Clear() error {
	mlc.mutex.Lock()
	defer mlc.mutex.Unlock()

	// Clear L1
	mlc.l1Cache.Purge()

	// Clear L2
	if err := mlc.l2Cache.Clear(); err != nil {
		return fmt.Errorf("failed to clear L2 cache: %w", err)
	}

	// Clear L3 if enabled
	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Clear(); err != nil {
			return fmt.Errorf("failed to clear L3 cache: %w", err)
		}
	}

	return nil
}

// Size returns the total number of items in all caches
func (mlc *MultiLevelCache) Size() int {
	mlc.mutex.RLock()
	defer mlc.mutex.RUnlock()

	size := mlc.l1Cache.Len() + mlc.l2Cache.Size()
	if mlc.l3Cache != nil {
		size += mlc.l3Cache.Size()
	}
	return size
}

// GetHitRate returns the cache hit rate
func (mlc *MultiLevelCache) GetHitRate() float64 {
	mlc.mutex.RLock()
	defer mlc.mutex.RUnlock()

	total := mlc.metrics.TotalHits + mlc.metrics.TotalMisses
	if total == 0 {
		return 0.0
	}
	return float64(mlc.metrics.TotalHits) / float64(total)
}

// GetMetrics returns cache metrics
func (mlc *MultiLevelCache) GetMetrics() *CacheMetrics {
	mlc.mutex.RLock()
	defer mlc.mutex.RUnlock()

	// Return a copy of metrics
	return &CacheMetrics{
		L1Hits:           mlc.metrics.L1Hits,
		L1Misses:         mlc.metrics.L1Misses,
		L2Hits:           mlc.metrics.L2Hits,
		L2Misses:         mlc.metrics.L2Misses,
		L3Hits:           mlc.metrics.L3Hits,
		L3Misses:         mlc.metrics.L3Misses,
		TotalHits:        mlc.metrics.TotalHits,
		TotalMisses:      mlc.metrics.TotalMisses,
		Evictions:        mlc.metrics.Evictions,
		SetOperations:    mlc.metrics.SetOperations,
		GetOperations:    mlc.metrics.GetOperations,
		DeleteOperations: mlc.metrics.DeleteOperations,
	}
}

// startCleanup starts the cleanup goroutine
func (mlc *MultiLevelCache) startCleanup() {
	ticker := time.NewTicker(mlc.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		mlc.cleanup()
	}
}

// cleanup removes expired items from caches
func (mlc *MultiLevelCache) cleanup() {
	mlc.mutex.Lock()
	defer mlc.mutex.Unlock()

	now := time.Now()
	
	// Cleanup L1 cache
	keys := mlc.l1Cache.Keys()
	for _, key := range keys {
		if item, ok := mlc.l1Cache.Peek(key); ok {
			if now.After(item.ExpiresAt) {
				mlc.l1Cache.Remove(key)
				mlc.metrics.Evictions++
			}
		}
	}

	// L2 and L3 caches handle their own cleanup
}