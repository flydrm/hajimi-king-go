package platform

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiPlatform implements the Gemini platform
type GeminiPlatform struct {
	config *PlatformConfig
	client *genai.Client
}

// NewGeminiPlatform creates a new Gemini platform
func NewGeminiPlatform(apiKey string) (*GeminiPlatform, error) {
	config := &PlatformConfig{
		Name:        "gemini",
		Enabled:     true,
		Priority:    1,
		Description: "Google Gemini API平台",
		APIKey:      apiKey,
		BaseURL:     "https://generativelanguage.googleapis.com",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiPlatform{
		config: config,
		client: client,
	}, nil
}

// GetName returns the platform name
func (gp *GeminiPlatform) GetName() string {
	return gp.config.Name
}

// GetQueries returns search queries for Gemini
func (gp *GeminiPlatform) GetQueries() []string {
	return []string{
		"AIza",
		"generative-ai-go",
		"genai.NewClient",
		"google.generative-ai-go",
		"generativelanguage.googleapis.com",
		"GEMINI_API_KEY",
		"gemini_api_key",
		"GOOGLE_AI_API_KEY",
	}
}

// GetRegexPatterns returns regex patterns for Gemini keys
func (gp *GeminiPlatform) GetRegexPatterns() []*RegexPattern {
	return []*RegexPattern{
		{
			Name:        "Gemini API Key",
			Pattern:     `AIza[0-9A-Za-z\\-_]{35}`,
			Confidence:  0.9,
			Description: "Google Gemini API key pattern",
		},
		{
			Name:        "Gemini API Key Alternative",
			Pattern:     `AIzaSy[A-Za-z0-9\\-_]{33}`,
			Confidence:  0.8,
			Description: "Google Gemini API key alternative pattern",
		},
	}
}

// ValidateKey validates a Gemini API key using the discovered key itself
func (gp *GeminiPlatform) ValidateKey(key string) (*ValidationResult, error) {
	result := &ValidationResult{
		Key:        key,
		Platform:   gp.GetName(),
		ValidatedAt: time.Now(),
	}

	// Basic format validation
	if !gp.isValidFormat(key) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid key format")
		return result, nil
	}

	// API validation using the discovered key
	ctx, cancel := context.WithTimeout(context.Background(), gp.config.Timeout)
	defer cancel()

	// Create client with the discovered key
	clientOpts := []option.ClientOption{
		option.WithAPIKey(key),
		option.WithEndpoint("generativelanguage.googleapis.com"),
	}

	client, err := genai.NewClient(ctx, clientOpts...)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Errorf("failed to create client: %w", err)
		result.Response = err.Error()
		return result, nil
	}
	defer client.Close()

	// Test with a simple request
	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(ctx, genai.Text("hi"))
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "PermissionDenied") || strings.Contains(errStr, "Unauthenticated") {
			result.Valid = false
			result.Error = fmt.Errorf("not_authorized_key")
			result.Response = "not_authorized_key"
			return result, nil
		}
		if strings.Contains(errStr, "TooManyRequests") || strings.Contains(errStr, "429") || strings.Contains(strings.ToLower(errStr), "rate limit") {
			result.Valid = false
			result.Error = fmt.Errorf("rate_limited")
			result.Response = "rate_limited"
			return result, nil
		}
		if strings.Contains(errStr, "SERVICE_DISABLED") || strings.Contains(errStr, "API has not been used") {
			result.Valid = false
			result.Error = fmt.Errorf("disabled")
			result.Response = "disabled"
			return result, nil
		}
		result.Valid = false
		result.Error = err
		result.Response = "error:" + errStr
		return result, nil
	}

	if resp != nil {
		result.Valid = true
		result.Response = "ok"
		return result, nil
	}

	result.Valid = false
	result.Error = fmt.Errorf("unknown_error")
	result.Response = "unknown_error"
	return result, nil
}

// GetConfig returns platform configuration
func (gp *GeminiPlatform) GetConfig() *PlatformConfig {
	return gp.config
}

// IsEnabled returns whether the platform is enabled
func (gp *GeminiPlatform) IsEnabled() bool {
	return gp.config.Enabled
}

// GetPriority returns platform priority
func (gp *GeminiPlatform) GetPriority() int {
	return gp.config.Priority
}

// isValidFormat validates the key format
func (gp *GeminiPlatform) isValidFormat(key string) bool {
	// Check length
	if len(key) != 39 {
		return false
	}
	
	// Check prefix
	if !regexp.MustCompile(`^AIza`).MatchString(key) {
		return false
	}
	
	// Check character set
	pattern := regexp.MustCompile(`^AIza[0-9A-Za-z\\-_]{35}$`)
	return pattern.MatchString(key)
}

// Close closes the platform resources
func (gp *GeminiPlatform) Close() error {
	if gp.client != nil {
		return gp.client.Close()
	}
	return nil
}