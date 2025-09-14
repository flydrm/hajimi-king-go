package models

import (
	"time"
)

// Checkpoint represents a checkpoint for incremental scanning
type Checkpoint struct {
	LastScanTime time.Time `json:"last_scan_time"`
	ProcessedFiles map[string]bool `json:"processed_files"`
	TotalFiles   int       `json:"total_files"`
	TotalKeys    int       `json:"total_keys"`
	ValidKeys    int       `json:"valid_keys"`
	RateLimitedKeys int    `json:"rate_limited_keys"`
}

// GitHubSearchResult represents the result of a GitHub search
type GitHubSearchResult struct {
	TotalCount int               `json:"total_count"`
	Items      []GitHubSearchItem `json:"items"`
}

// GitHubSearchItem represents a single item from GitHub search
type GitHubSearchItem struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	URL         string           `json:"url"`
	Repository  GitHubRepository `json:"repository"`
	TextMatches []TextMatch      `json:"text_matches"`
	Score       float64          `json:"score"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Language    string `json:"language"`
	Size        int    `json:"size"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TextMatch represents a text match in search results
type TextMatch struct {
	ObjectURL  string `json:"object_url"`
	ObjectType string `json:"object_type"`
	Property   string `json:"property"`
	Fragment   string `json:"fragment"`
	Matches    []Match `json:"matches"`
}

// Match represents a specific match within text
type Match struct {
	Text       string `json:"text"`
	Indices    []int  `json:"indices"`
}

// KeyInfo represents information about a discovered key
type KeyInfo struct {
	Key         string    `json:"key"`
	Platform    string    `json:"platform"`
	Repository  string    `json:"repository"`
	FilePath    string    `json:"file_path"`
	FileURL     string    `json:"file_url"`
	LineNumber  int       `json:"line_number"`
	Context     string    `json:"context"`
	Confidence  float64   `json:"confidence"`
	RiskLevel   string    `json:"risk_level"`
	IsValid     bool      `json:"is_valid"`
	IsPlaceholder bool    `json:"is_placeholder"`
	IsTestKey   bool      `json:"is_test_key"`
	DiscoveredAt time.Time `json:"discovered_at"`
	ValidatedAt  *time.Time `json:"validated_at,omitempty"`
}

// SkipStats represents statistics about skipped items
type SkipStats struct {
	TotalSkipped    int `json:"total_skipped"`
	DuplicateSkipped int `json:"duplicate_skipped"`
	SizeSkipped     int `json:"size_skipped"`
	ExtensionSkipped int `json:"extension_skipped"`
	PathSkipped     int `json:"path_skipped"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	// Processing metrics
	ProcessedFiles   int64   `json:"processed_files"`
	ProcessedKeys    int64   `json:"processed_keys"`
	ValidKeys        int64   `json:"valid_keys"`
	RateLimitedKeys  int64   `json:"rate_limited_keys"`
	
	// Performance metrics
	ThroughputKeysPerSecond float64 `json:"throughput_keys_per_second"`
	AverageResponseTime     float64 `json:"average_response_time"`
	MemoryUsageMB          float64 `json:"memory_usage_mb"`
	
	// Cache metrics
	CacheHitRate           float64 `json:"cache_hit_rate"`
	CacheMissRate          float64 `json:"cache_miss_rate"`
	
	// Detection metrics
	DetectionRate          float64 `json:"detection_rate"`
	FalsePositiveRate      float64 `json:"false_positive_rate"`
	
	// Worker pool metrics
	ActiveWorkers          int     `json:"active_workers"`
	QueueSize              int     `json:"queue_size"`
	TasksCompleted         int64   `json:"tasks_completed"`
	TasksFailed            int64   `json:"tasks_failed"`
	
	// Platform metrics
	PlatformMetrics        map[string]PlatformMetrics `json:"platform_metrics"`
	
	// Timestamps
	StartTime              time.Time `json:"start_time"`
	LastUpdateTime         time.Time `json:"last_update_time"`
}

// PlatformMetrics represents metrics for a specific platform
type PlatformMetrics struct {
	PlatformName    string  `json:"platform_name"`
	KeysFound       int64   `json:"keys_found"`
	ValidKeys       int64   `json:"valid_keys"`
	InvalidKeys     int64   `json:"invalid_keys"`
	RateLimitedKeys int64   `json:"rate_limited_keys"`
	AverageResponseTime float64 `json:"average_response_time"`
	SuccessRate     float64 `json:"success_rate"`
	LastProcessed   time.Time `json:"last_processed"`
}

// QueryTask represents a task for processing a search query
type QueryTask struct {
	ID          string
	Platform    string
	Query       string
	Priority    int
	HajimiKing  *OptimizedHajimiKing
}

// QueryResult represents the result of a query task
type QueryResult struct {
	TaskID      string
	Platform    string
	Query       string
	Items       []GitHubSearchItem
	Error       error
	ProcessedAt time.Time
}

// ValidationTask represents a task for validating a key
type ValidationTask struct {
	ID          string
	Platform    string
	Key         string
	Priority    int
	HajimiKing  *OptimizedHajimiKing
}

// ValidationResult represents the result of a validation task
type ValidationResult struct {
	TaskID      string
	Platform    string
	Key         string
	IsValid     bool
	Error       error
	ProcessedAt time.Time
}