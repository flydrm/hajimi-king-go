package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	// GitHub configuration
	GitHubToken    string `json:"github_token"`
	GitHubProxy    string `json:"github_proxy"`
	GitHubBaseURL  string `json:"github_base_url"`
	
	// Data paths
	DataPath       string `json:"data_path"`
	QueriesPath    string `json:"queries_path"`
	CheckpointPath string `json:"checkpoint_path"`
	
	// API configuration
	GeminiAPIKey     string `json:"gemini_api_key"`
	OpenRouterAPIKey string `json:"openrouter_api_key"`
	SiliconFlowAPIKey string `json:"siliconflow_api_key"`
	
	// File prefixes
	ValidKeyPrefix     string `json:"valid_key_prefix"`
	ValidKeyDetailPrefix string `json:"valid_key_detail_prefix"`
	RateLimitedPrefix  string `json:"rate_limited_prefix"`
	
	// Processing configuration
	MaxConcurrentFiles int           `json:"max_concurrent_files"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	ScanInterval       time.Duration `json:"scan_interval"`
	
	// Cache configuration
	CacheConfig CacheConfig `json:"cache_config"`
	
	// Platform switches
	PlatformSwitches PlatformSwitches `json:"platform_switches"`
	
	// Worker pool configuration
	WorkerPoolSize int `json:"worker_pool_size"`
	
	// API server configuration
	APIServerPort int    `json:"api_server_port"`
	JWTSecret     string `json:"jwt_secret"`
	
	// External sync configuration
	SyncToGeminiBalancer bool   `json:"sync_to_gemini_balancer"`
	SyncToGPTLoad        bool   `json:"sync_to_gpt_load"`
	SyncEndpoint         string `json:"sync_endpoint"`
	SyncToken            string `json:"sync_token"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	L1MaxSize       int           `json:"l1_max_size"`
	L1TTL           time.Duration `json:"l1_ttl"`
	L2TTL           time.Duration `json:"l2_ttl"`
	L3TTL           time.Duration `json:"l3_ttl"`
	EnableL3        bool          `json:"enable_l3"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// PlatformSwitches represents platform control configuration
type PlatformSwitches struct {
	GlobalEnabled     bool                      `json:"global_enabled"`
	ExecutionMode     string                    `json:"execution_mode"` // "all", "single", "selected"
	SelectedPlatforms []string                  `json:"selected_platforms"`
	Platforms         map[string]PlatformConfig `json:"platforms"`
}

// PlatformConfig represents individual platform configuration
type PlatformConfig struct {
	Enabled     bool   `json:"enabled"`
	Priority    int    `json:"priority"`
	Description string `json:"description"`
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	config := &Config{
		// GitHub configuration
		GitHubToken:    getEnvWithDefault("GITHUB_TOKEN", ""),
		GitHubProxy:    getEnvWithDefault("GITHUB_PROXY", ""),
		GitHubBaseURL:  getEnvWithDefault("GITHUB_BASE_URL", "https://api.github.com"),
		
		// Data paths
		DataPath:       getEnvWithDefault("DATA_PATH", "./data"),
		QueriesPath:    getEnvWithDefault("QUERIES_PATH", "./queries.txt"),
		CheckpointPath: getEnvWithDefault("CHECKPOINT_PATH", "./checkpoint.json"),
		
		// API configuration
		GeminiAPIKey:     getEnvWithDefault("GEMINI_API_KEY", ""),
		OpenRouterAPIKey: getEnvWithDefault("OPENROUTER_API_KEY", ""),
		SiliconFlowAPIKey: getEnvWithDefault("SILICONFLOW_API_KEY", ""),
		
		// File prefixes
		ValidKeyPrefix:     getEnvWithDefault("VALID_KEY_PREFIX", "keys_valid"),
		ValidKeyDetailPrefix: getEnvWithDefault("VALID_KEY_DETAIL_PREFIX", "keys_valid_detail"),
		RateLimitedPrefix:  getEnvWithDefault("RATE_LIMITED_PREFIX", "key_429"),
		
		// Processing configuration
		MaxConcurrentFiles: parseIntWithDefault("MAX_CONCURRENT_FILES", 10),
		MaxRetries:         parseIntWithDefault("MAX_RETRIES", 3),
		RetryDelay:         parseDurationWithDefault("RETRY_DELAY", "5s"),
		ScanInterval:       parseDurationWithDefault("SCAN_INTERVAL", "1m"),
		
		// Cache configuration
		CacheConfig: CacheConfig{
			L1MaxSize:       parseIntWithDefault("CACHE_L1_MAX_SIZE", 1000),
			L1TTL:           parseDurationWithDefault("CACHE_L1_TTL", "5m"),
			L2TTL:           parseDurationWithDefault("CACHE_L2_TTL", "1h"),
			L3TTL:           parseDurationWithDefault("CACHE_L3_TTL", "24h"),
			EnableL3:        parseBoolWithDefault("CACHE_ENABLE_L3", false),
			CleanupInterval: parseDurationWithDefault("CACHE_CLEANUP_INTERVAL", "10m"),
		},
		
		// Platform switches
		PlatformSwitches: PlatformSwitches{
			GlobalEnabled:     parseBoolWithDefault("PLATFORM_GLOBAL_ENABLED", true),
			ExecutionMode:     getEnvWithDefault("PLATFORM_EXECUTION_MODE", "all"),
			SelectedPlatforms: parseStringList(getEnvWithDefault("PLATFORM_SELECTED", "")),
			Platforms: map[string]PlatformConfig{
				"gemini": {
					Enabled:     parseBoolWithDefault("PLATFORM_GEMINI_ENABLED", true),
					Priority:    1,
					Description: "Google Gemini API平台",
				},
				"openrouter": {
					Enabled:     parseBoolWithDefault("PLATFORM_OPENROUTER_ENABLED", false),
					Priority:    2,
					Description: "OpenRouter API平台",
				},
				"siliconflow": {
					Enabled:     parseBoolWithDefault("PLATFORM_SILICONFLOW_ENABLED", false),
					Priority:    3,
					Description: "SiliconFlow API平台",
				},
			},
		},
		
		// Worker pool configuration
		WorkerPoolSize: parseIntWithDefault("WORKER_POOL_SIZE", 8),
		
		// API server configuration
		APIServerPort: parseIntWithDefault("API_SERVER_PORT", 8080),
		JWTSecret:     getEnvWithDefault("JWT_SECRET", "hajimi-king-secret-key"),
		
		// External sync configuration
		SyncToGeminiBalancer: parseBoolWithDefault("SYNC_TO_GEMINI_BALANCER", false),
		SyncToGPTLoad:        parseBoolWithDefault("SYNC_TO_GPT_LOAD", false),
		SyncEndpoint:         getEnvWithDefault("SYNC_ENDPOINT", ""),
		SyncToken:            getEnvWithDefault("SYNC_TOKEN", ""),
	}

	return config
}

// GetEnabledPlatforms returns the list of enabled platforms based on configuration
func (c *Config) GetEnabledPlatforms() []string {
	if !c.PlatformSwitches.GlobalEnabled {
		return []string{}
	}

	var platforms []string
	
	switch c.PlatformSwitches.ExecutionMode {
	case "single":
		// Return only the first enabled platform
		for name, config := range c.PlatformSwitches.Platforms {
			if config.Enabled {
				return []string{name}
			}
		}
	case "selected":
		// Return only selected platforms
		for _, name := range c.PlatformSwitches.SelectedPlatforms {
			if config, exists := c.PlatformSwitches.Platforms[name]; exists && config.Enabled {
				platforms = append(platforms, name)
			}
		}
	default: // "all"
		// Return all enabled platforms
		for name, config := range c.PlatformSwitches.Platforms {
			if config.Enabled {
				platforms = append(platforms, name)
			}
		}
	}

	// Sort platforms by priority
	return c.sortPlatformsByPriority(platforms)
}

// sortPlatformsByPriority sorts platforms by their priority
func (c *Config) sortPlatformsByPriority(platforms []string) []string {
	// Simple bubble sort by priority
	for i := 0; i < len(platforms)-1; i++ {
		for j := 0; j < len(platforms)-i-1; j++ {
			priority1 := c.PlatformSwitches.Platforms[platforms[j]].Priority
			priority2 := c.PlatformSwitches.Platforms[platforms[j+1]].Priority
			if priority1 > priority2 {
				platforms[j], platforms[j+1] = platforms[j+1], platforms[j]
			}
		}
	}
	return platforms
}

// Helper functions
func loadEnvFile() error {
	// Try to load .env file
	if _, err := os.Stat(".env"); err == nil {
		// Load .env file using godotenv
		// This would require importing github.com/joho/godotenv
		// For now, we'll skip this and rely on environment variables
	}
	return nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func parseDurationWithDefault(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 5 * time.Second // fallback
}

func parseStringList(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}