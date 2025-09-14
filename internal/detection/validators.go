package detection

import (
	"regexp"
	"strings"
)

// GeminiValidator validates Gemini API keys
type GeminiValidator struct{}

func (gv *GeminiValidator) Validate(key string) bool {
	// Gemini keys start with AIza and are 39 characters long
	if len(key) != 39 {
		return false
	}
	if !strings.HasPrefix(key, "AIza") {
		return false
	}
	
	// Check if it contains only valid characters
	pattern := regexp.MustCompile(`^AIza[0-9A-Za-z\\-_]{35}$`)
	return pattern.MatchString(key)
}

func (gv *GeminiValidator) GetName() string {
	return "GeminiValidator"
}

// OpenRouterValidator validates OpenRouter API keys
type OpenRouterValidator struct{}

func (ov *OpenRouterValidator) Validate(key string) bool {
	// OpenRouter keys start with sk-or- and are 35 characters long
	if len(key) != 35 {
		return false
	}
	if !strings.HasPrefix(key, "sk-or-") {
		return false
	}
	
	// Check if it contains only valid characters
	pattern := regexp.MustCompile(`^sk-or-[0-9A-Za-z]{32}$`)
	return pattern.MatchString(key)
}

func (ov *OpenRouterValidator) GetName() string {
	return "OpenRouterValidator"
}

// SiliconFlowValidator validates SiliconFlow API keys
type SiliconFlowValidator struct{}

func (sv *SiliconFlowValidator) Validate(key string) bool {
	// SiliconFlow keys start with sk- and are 35 characters long
	if len(key) != 35 {
		return false
	}
	if !strings.HasPrefix(key, "sk-") {
		return false
	}
	
	// Check if it contains only valid characters
	pattern := regexp.MustCompile(`^sk-[0-9A-Za-z]{32}$`)
	return pattern.MatchString(key)
}

func (sv *SiliconFlowValidator) GetName() string {
	return "SiliconFlowValidator"
}