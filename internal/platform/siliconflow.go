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

// SiliconFlowPlatform implements the SiliconFlow platform
type SiliconFlowPlatform struct {
	config *PlatformConfig
	client *http.Client
}

// SiliconFlowRequest represents a request to SiliconFlow API
type SiliconFlowRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	MaxTokens int      `json:"max_tokens"`
}

// SiliconFlowResponse represents a response from SiliconFlow API
type SiliconFlowResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage   `json:"usage"`
	Error   *Error  `json:"error,omitempty"`
}

// NewSiliconFlowPlatform creates a new SiliconFlow platform
func NewSiliconFlowPlatform(apiKey string) (*SiliconFlowPlatform, error) {
	config := &PlatformConfig{
		Name:        "siliconflow",
		Enabled:     true,
		Priority:    3,
		Description: "SiliconFlow API平台",
		APIKey:      apiKey,
		BaseURL:     "https://api.siliconflow.cn/v1",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &SiliconFlowPlatform{
		config: config,
		client: client,
	}, nil
}

// GetName returns the platform name
func (sp *SiliconFlowPlatform) GetName() string {
	return sp.config.Name
}

// GetQueries returns search queries for SiliconFlow
func (sp *SiliconFlowPlatform) GetQueries() []string {
	return []string{
		"sk-",
		"siliconflow.cn",
		"SILICONFLOW_API_KEY",
		"siliconflow_api_key",
		"siliconflow",
		"SiliconFlow",
		"api.siliconflow.cn",
	}
}

// GetRegexPatterns returns regex patterns for SiliconFlow keys
func (sp *SiliconFlowPlatform) GetRegexPatterns() []*RegexPattern {
	return []*RegexPattern{
		{
			Name:        "SiliconFlow API Key",
			Pattern:     `sk-[0-9A-Za-z]{32}`,
			Confidence:  0.9,
			Description: "SiliconFlow API key pattern",
		},
	}
}

// ValidateKey validates a SiliconFlow API key using the discovered key itself
func (sp *SiliconFlowPlatform) ValidateKey(key string) (*ValidationResult, error) {
	result := &ValidationResult{
		Key:        key,
		Platform:   sp.GetName(),
		ValidatedAt: time.Now(),
	}

	// Basic format validation
	if !sp.isValidFormat(key) {
		result.Valid = false
		result.Error = fmt.Errorf("invalid key format")
		return result, nil
	}

	// API validation using the discovered key
	ctx, cancel := context.WithTimeout(context.Background(), sp.config.Timeout)
	defer cancel()

	// Create test request
	reqBody := SiliconFlowRequest{
		Model: "Qwen/Qwen2.5-7B-Instruct",
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
	req, err := http.NewRequestWithContext(ctx, "POST", sp.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Valid = false
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result, nil
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := sp.client.Do(req)
	if err != nil {
		result.Valid = false
		result.Error = err
		result.Response = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Parse response
	var response SiliconFlowResponse
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
func (sp *SiliconFlowPlatform) GetConfig() *PlatformConfig {
	return sp.config
}

// IsEnabled returns whether the platform is enabled
func (sp *SiliconFlowPlatform) IsEnabled() bool {
	return sp.config.Enabled
}

// GetPriority returns platform priority
func (sp *SiliconFlowPlatform) GetPriority() int {
	return sp.config.Priority
}

// isValidFormat validates the key format
func (sp *SiliconFlowPlatform) isValidFormat(key string) bool {
	// Check length
	if len(key) != 35 {
		return false
	}
	
	// Check prefix
	if !regexp.MustCompile(`^sk-`).MatchString(key) {
		return false
	}
	
	// Check character set
	pattern := regexp.MustCompile(`^sk-[0-9A-Za-z]{32}$`)
	return pattern.MatchString(key)
}