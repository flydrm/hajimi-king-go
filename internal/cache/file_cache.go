package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileCache implements a file-based cache
type FileCache struct {
	basePath string
	ttl      time.Duration
	mutex    sync.RWMutex
}

// FileCacheItem represents a cached item in file cache
type FileCacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// NewFileCache creates a new file-based cache
func NewFileCache(basePath string, ttl time.Duration) (*FileCache, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &FileCache{
		basePath: basePath,
		ttl:      ttl,
	}, nil
}

// Get retrieves a value from the file cache
func (fc *FileCache) Get(key string) (interface{}, bool) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()

	filePath := fc.getFilePath(key)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, false
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}

	// Unmarshal cache item
	var item FileCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		// Remove expired file
		os.Remove(filePath)
		return nil, false
	}

	return item.Value, true
}

// Set stores a value in the file cache
func (fc *FileCache) Set(key string, value interface{}, ttl time.Duration) error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	filePath := fc.getFilePath(key)
	
	// Create cache item
	item := FileCacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Delete removes a value from the file cache
func (fc *FileCache) Delete(key string) error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	filePath := fc.getFilePath(key)
	return os.Remove(filePath)
}

// Clear clears all files in the cache
func (fc *FileCache) Clear() error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	// Read directory
	entries, err := os.ReadDir(fc.basePath)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	// Remove all files
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(fc.basePath, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// Size returns the number of files in the cache
func (fc *FileCache) Size() int {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()

	entries, err := os.ReadDir(fc.basePath)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

// getFilePath returns the file path for a given key
func (fc *FileCache) getFilePath(key string) string {
	// Use key as filename (you might want to hash it for security)
	return filepath.Join(fc.basePath, key+".cache")
}