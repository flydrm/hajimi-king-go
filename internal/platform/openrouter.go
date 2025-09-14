package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// OpenRouterPlatform implements the OpenRouter platform
type OpenRouterPlatform struct {
	config *PlatformConfig
	client *http.Client
}

// OpenRouterRequest represents a request to OpenRouter API
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	MaxTokens int      `json:"max_tokens"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents a response from OpenRouter API
type OpenRouterResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage   `json:"usage"`
	Error   *Error  `json:"error,omitempty"`
}

// Choice represents a choice in the response
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Error represents an API error
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenRouterPlatform creates a new OpenRouter platform
func NewOpenRouterPlatform(apiKey string) (*OpenRouterPlatform, error) {
	config := &PlatformConfig{
		Name:        "openrouter",
		Enabled:     true,
		Priority:    2,
		Description: "OpenRouter API平台",
		APIKey:      apiKey,
		BaseURL:     "https://openrouter.ai/api/v1",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &OpenRouterPlatform{
		config: config,
		client: client,
	}, nil
}

// GetName returns the platform name
func (op *OpenRouterPlatform) GetName() string {
	return op.config.Name
}

// GetQueries returns search queries for OpenRouter
func (op *OpenRouterPlatform) GetQueries() []string {
	return []string{
		"sk-or-",
		"openrouter.ai",
		"OPENROUTER_API_KEY",
		"openrouter_api_key",
		"openrouter",
		"OpenRouter",
		"openrouter.ai/api/v1",
	}
}

// GetRegexPatterns returns regex patterns for OpenRouter keys
func (op *OpenRouterPlatform) GetRegexPatterns() []*RegexPattern {
	return []*RegexPattern{
		{
			Name:        "OpenRouter API Key",
			Pattern:     `sk-or-[0-9A-Za-z]{32}`,
			Confidence:  0.9,
			Description: "OpenRouter API key pattern",
		},
	}
}

// ValidateKey validates an OpenRouter API key using the discovered key itself
func (op *OpenRouterPlatform) ValidateKey(key string) (*ValidationResult, error) {
	result := &ValidationResult{
		Key:        key,
		Platform:   op.GetName(),
		ValidatedAt: time.Now(),
	}

	// Basic format validation
	if !op.isValidFormat(key) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid key format")
		return result, nil
	}

	// API validation using the discovered key
	ctx, cancel := context.WithTimeout(context.Background(), op.config.Timeout)
	defer cancel()

	// Create test request
	reqBody := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: "test",
			},
		},
		MaxTokens: 1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Errorf("failed to marshal request: %w", err)
		return result, nil
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", op.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Valid = false
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result, nil
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://hajimi-king-go-v2")
	req.Header.Set("X-Title", "Hajimi King Go v2.0")

	// Send request
	resp, err := op.client.Do(req)
	if err != nil {
		result.Valid = false
		result.Error = err
		result.Response = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Parse response
	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		result.Valid = false
		result.Error = fmt.Errorf("failed to decode response: %w", err)
		return result, nil
	}

	// Check for API errors
	if response.Error != nil {
		errStr := response.Error.Message
		if strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "Invalid API key") {
			result.Valid = false
			result.Error = fmt.Errorf("not_authorized_key")
			result.Response = "not_authorized_key"
			return result, nil
		}
		if strings.Contains(errStr, "Too Many Requests") || strings.Contains(errStr, "rate limit") {
			result.Valid = false
			result.Error = fmt.Errorf("rate_limited")
			result.Response = "rate_limited"
			return result, nil
		}
		result.Valid = false
		result.Error = fmt.Errorf("API error: %s", errStr)
		result.Response = errStr
		return result, nil
	}

	// Check status code
	if resp.StatusCode == 401 {
		result.Valid = false
		result.Error = fmt.Errorf("not_authorized_key")
		result.Response = "not_authorized_key"
		return result, nil
	}
	if resp.StatusCode == 429 {
		result.Valid = false
		result.Error = fmt.Errorf("rate_limited")
		result.Response = "rate_limited"
		return result, nil
	}
	if resp.StatusCode != 200 {
		result.Valid = false
		result.Error = fmt.Errorf("HTTP error: %d", resp.StatusCode)
		result.Response = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result, nil
	}

	result.Valid = true
	result.Response = "ok"
	return result, nil
}

// GetConfig returns platform configuration
func (op *OpenRouterPlatform) GetConfig() *PlatformConfig {
	return op.config
}

// IsEnabled returns whether the platform is enabled
func (op *OpenRouterPlatform) IsEnabled() bool {
	return op.config.Enabled
}

// GetPriority returns platform priority
func (op *OpenRouterPlatform) GetPriority() int {
	return op.config.Priority
}

// isValidFormat validates the key format
func (op *OpenRouterPlatform) isValidFormat(key string) bool {
	// Check length
	if len(key) != 35 {
		return false
	}
	
	// Check prefix
	if !regexp.MustCompile(`^sk-or-`).MatchString(key) {
		return false
	}
	
	// Check character set
	pattern := regexp.MustCompile(`^sk-or-[0-9A-Za-z]{32}$`)
	return pattern.MatchString(key)
}