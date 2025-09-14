package platform

import (
	"context"
	"fmt"
	"regexp"
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

// ValidateKey validates a Gemini API key
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

	// API validation
	ctx, cancel := context.WithTimeout(context.Background(), gp.config.Timeout)
	defer cancel()

	// Try to create a model to validate the key
	model := gp.client.GenerativeModel("gemini-pro")
	
	// Test with a simple request
	_, err := model.GenerateContent(ctx, genai.Text("test"))
	if err != nil {
		result.Valid = false
		result.Error = err
		result.Response = err.Error()
		return result, nil
	}

	result.Valid = true
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