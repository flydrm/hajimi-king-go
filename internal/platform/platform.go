package platform

import (
	"context"
	"fmt"
	"time"
)

// Platform represents a platform interface for API key discovery and validation
type Platform interface {
	GetName() string
	GetQueries() []string
	GetRegexPatterns() []*RegexPattern
	ValidateKey(key string) (*ValidationResult, error)
	GetConfig() *PlatformConfig
	IsEnabled() bool
	GetPriority() int
}

// RegexPattern represents a regex pattern for key detection
type RegexPattern struct {
	Name        string
	Pattern     string
	Confidence  float64
	Description string
}

// ValidationResult represents the result of key validation
type ValidationResult struct {
	Valid      bool
	Key        string
	Platform   string
	Error      error
	StatusCode int
	Response   string
	ValidatedAt time.Time
}

// PlatformConfig represents platform-specific configuration
type PlatformConfig struct {
	Name        string
	Enabled     bool
	Priority    int
	Description string
	APIKey      string
	BaseURL     string
	Timeout     time.Duration
	MaxRetries  int
}

// PlatformManager manages multiple platforms
type PlatformManager struct {
	platforms map[string]Platform
	configs   map[string]*PlatformConfig
}

// NewPlatformManager creates a new platform manager
func NewPlatformManager() *PlatformManager {
	return &PlatformManager{
		platforms: make(map[string]Platform),
		configs:   make(map[string]*PlatformConfig),
	}
}

// RegisterPlatform registers a platform
func (pm *PlatformManager) RegisterPlatform(platform Platform) error {
	name := platform.GetName()
	if _, exists := pm.platforms[name]; exists {
		return fmt.Errorf("platform %s already registered", name)
	}
	
	pm.platforms[name] = platform
	pm.configs[name] = platform.GetConfig()
	return nil
}

// GetPlatform returns a platform by name
func (pm *PlatformManager) GetPlatform(name string) (Platform, bool) {
	platform, exists := pm.platforms[name]
	return platform, exists
}

// GetEnabledPlatforms returns all enabled platforms
func (pm *PlatformManager) GetEnabledPlatforms() []Platform {
	var enabled []Platform
	for _, platform := range pm.platforms {
		if platform.IsEnabled() {
			enabled = append(enabled, platform)
		}
	}
	return enabled
}

// GetAllPlatforms returns all registered platforms
func (pm *PlatformManager) GetAllPlatforms() []Platform {
	var all []Platform
	for _, platform := range pm.platforms {
		all = append(all, platform)
	}
	return all
}

// GetPlatformNames returns all platform names
func (pm *PlatformManager) GetPlatformNames() []string {
	var names []string
	for name := range pm.platforms {
		names = append(names, name)
	}
	return names
}

// ValidateKey validates a key against a specific platform
func (pm *PlatformManager) ValidateKey(platformName, key string) (*ValidationResult, error) {
	platform, exists := pm.GetPlatform(platformName)
	if !exists {
		return nil, fmt.Errorf("platform %s not found", platformName)
	}
	
	return platform.ValidateKey(key)
}

// ValidateKeyAll validates a key against all enabled platforms
func (pm *PlatformManager) ValidateKeyAll(key string) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)
	
	for _, platform := range pm.GetEnabledPlatforms() {
		result, err := platform.ValidateKey(key)
		if err != nil {
			result = &ValidationResult{
				Valid:      false,
				Key:        key,
				Platform:   platform.GetName(),
				Error:      err,
				ValidatedAt: time.Now(),
			}
		}
		results[platform.GetName()] = result
	}
	
	return results
}

// GetPlatformConfig returns platform configuration
func (pm *PlatformManager) GetPlatformConfig(name string) (*PlatformConfig, bool) {
	config, exists := pm.configs[name]
	return config, exists
}

// UpdatePlatformConfig updates platform configuration
func (pm *PlatformManager) UpdatePlatformConfig(name string, config *PlatformConfig) error {
	if _, exists := pm.configs[name]; !exists {
		return fmt.Errorf("platform %s not found", name)
	}
	
	pm.configs[name] = config
	return nil
}