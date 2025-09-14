package cache

import (
	"testing"
	"time"
)

func TestMultiLevelCache(t *testing.T) {
	config := &CacheConfig{
		L1MaxSize:       100,
		L1TTL:           5 * time.Minute,
		L2TTL:           1 * time.Hour,
		L3TTL:           24 * time.Hour,
		EnableL3:        false,
		CleanupInterval: 10 * time.Minute,
		L2Path:          "./test_cache",
	}
	
	cache, err := NewMultiLevelCache(config)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	
	// Test Set and Get
	key := "test_key"
	value := "test_value"
	ttl := 1 * time.Minute
	
	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}
	
	retrievedValue, exists := cache.Get(key)
	if !exists {
		t.Error("Expected value to exist in cache")
	}
	
	if retrievedValue != value {
		t.Errorf("Expected value '%s', got '%s'", value, retrievedValue)
	}
	
	// Test GetHitRate
	hitRate := cache.GetHitRate()
	if hitRate < 0 || hitRate > 1 {
		t.Errorf("Invalid hit rate: %f", hitRate)
	}
	
	// Test Size
	size := cache.Size()
	if size < 0 {
		t.Errorf("Invalid cache size: %d", size)
	}
	
	// Test Delete
	err = cache.Delete(key)
	if err != nil {
		t.Errorf("Failed to delete cache value: %v", err)
	}
	
	_, exists = cache.Get(key)
	if exists {
		t.Error("Expected value to be deleted from cache")
	}
}