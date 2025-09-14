package detection

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// SmartKeyDetector implements intelligent key detection
type SmartKeyDetector struct {
	patterns           map[string]*RegexPattern
	validators         map[string]KeyValidator
	contextAnalyzer    *ContextAnalyzer
	confidenceThreshold float64
	mutex              sync.RWMutex
	metrics            *DetectionMetrics
}

// RegexPattern represents a regex pattern for key detection
type RegexPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Platform    string
	Confidence  float64
	Description string
}

// KeyValidator interface for validating detected keys
type KeyValidator interface {
	Validate(key string) bool
	GetName() string
}

// ContextAnalyzer analyzes the context around detected keys
type ContextAnalyzer struct {
	placeholderPatterns []*regexp.Regexp
	testKeyPatterns     []*regexp.Regexp
	contextWindow       int
}

// KeyContext represents the context of a detected key
type KeyContext struct {
	Key           string
	Confidence    float64
	IsPlaceholder bool
	IsTestKey     bool
	Platform      string
	RiskLevel     string
	Context       string
	LineNumber    int
	StartPos      int
	EndPos        int
}

// DetectionMetrics represents detection performance metrics
type DetectionMetrics struct {
	TotalDetections    int64
	ValidDetections    int64
	PlaceholderFiltered int64
	TestKeyFiltered    int64
	LowConfidenceFiltered int64
	DetectionRate      float64
	FalsePositiveRate  float64
}

// NewSmartKeyDetector creates a new smart key detector
func NewSmartKeyDetector() *SmartKeyDetector {
	detector := &SmartKeyDetector{
		patterns:           make(map[string]*RegexPattern),
		validators:         make(map[string]KeyValidator),
		contextAnalyzer:    NewContextAnalyzer(),
		confidenceThreshold: 0.7,
		metrics:            &DetectionMetrics{},
	}

	// Initialize platform patterns
	detector.initializePatterns()
	detector.initializeValidators()

	return detector
}

// DetectKeys detects API keys in content using smart analysis
func (skd *SmartKeyDetector) DetectKeys(content string) []*KeyContext {
	skd.mutex.Lock()
	defer skd.mutex.Unlock()

	var results []*KeyContext
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		for patternName, pattern := range skd.patterns {
			matches := pattern.Pattern.FindAllStringIndex(line, -1)
			for _, match := range matches {
				key := line[match[0]:match[1]]
				
				// Analyze key context
				context := skd.analyzeKeyContext(key, line, pattern)
				context.LineNumber = lineNum + 1
				context.StartPos = match[0]
				context.EndPos = match[1]

				// Apply smart filtering
				if skd.isValidKey(context) {
					results = append(results, context)
					skd.metrics.ValidDetections++
				} else {
					skd.metrics.TotalDetections++
					if context.IsPlaceholder {
						skd.metrics.PlaceholderFiltered++
					}
					if context.IsTestKey {
						skd.metrics.TestKeyFiltered++
					}
					if context.Confidence < skd.confidenceThreshold {
						skd.metrics.LowConfidenceFiltered++
					}
				}
			}
		}
	}

	skd.metrics.TotalDetections += int64(len(results))
	return results
}

// analyzeKeyContext analyzes the context around a detected key
func (skd *SmartKeyDetector) analyzeKeyContext(key, content string, pattern *RegexPattern) *KeyContext {
	context := &KeyContext{
		Key:      key,
		Platform: pattern.Platform,
		Context:  content,
	}

	// Check for placeholder patterns
	context.IsPlaceholder = skd.contextAnalyzer.IsPlaceholder(key, content)
	
	// Check for test key patterns
	context.IsTestKey = skd.contextAnalyzer.IsTestKey(key, content)
	
	// Calculate confidence score
	context.Confidence = skd.calculateConfidence(key, content, pattern)
	
	// Determine risk level
	context.RiskLevel = skd.determineRiskLevel(context)

	return context
}

// isValidKey determines if a detected key is valid based on context analysis
func (skd *SmartKeyDetector) isValidKey(context *KeyContext) bool {
	// Filter out placeholders
	if context.IsPlaceholder {
		return false
	}
	
	// Filter out test keys
	if context.IsTestKey {
		return false
	}
	
	// Check confidence threshold
	if context.Confidence < skd.confidenceThreshold {
		return false
	}
	
	// Additional validation if validator exists
	if validator, exists := skd.validators[context.Platform]; exists {
		return validator.Validate(context.Key)
	}
	
	return true
}

// calculateConfidence calculates confidence score for a detected key
func (skd *SmartKeyDetector) calculateConfidence(key, content string, pattern *RegexPattern) float64 {
	confidence := pattern.Confidence
	
	// Adjust based on context
	if strings.Contains(strings.ToLower(content), "api") {
		confidence += 0.1
	}
	if strings.Contains(strings.ToLower(content), "key") {
		confidence += 0.1
	}
	if strings.Contains(strings.ToLower(content), "secret") {
		confidence += 0.1
	}
	if strings.Contains(strings.ToLower(content), "token") {
		confidence += 0.1
	}
	
	// Adjust based on key characteristics
	if len(key) > 20 {
		confidence += 0.1
	}
	if strings.Contains(key, "-") || strings.Contains(key, "_") {
		confidence += 0.05
	}
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// determineRiskLevel determines the risk level of a detected key
func (skd *SmartKeyDetector) determineRiskLevel(context *KeyContext) string {
	if context.Confidence >= 0.9 {
		return "high"
	} else if context.Confidence >= 0.7 {
		return "medium"
	} else {
		return "low"
	}
}

// initializePatterns initializes regex patterns for different platforms
func (skd *SmartKeyDetector) initializePatterns() {
	// Gemini API key pattern
	geminiPattern, _ := regexp.Compile(`AIza[0-9A-Za-z\\-_]{35}`)
	skd.patterns["gemini"] = &RegexPattern{
		Name:        "Gemini API Key",
		Pattern:     geminiPattern,
		Platform:    "gemini",
		Confidence:  0.9,
		Description: "Google Gemini API key pattern",
	}

	// OpenRouter API key pattern
	openrouterPattern, _ := regexp.Compile(`sk-or-[0-9A-Za-z]{32}`)
	skd.patterns["openrouter"] = &RegexPattern{
		Name:        "OpenRouter API Key",
		Pattern:     openrouterPattern,
		Platform:    "openrouter",
		Confidence:  0.9,
		Description: "OpenRouter API key pattern",
	}

	// SiliconFlow API key pattern
	siliconflowPattern, _ := regexp.Compile(`sk-[0-9A-Za-z]{32}`)
	skd.patterns["siliconflow"] = &RegexPattern{
		Name:        "SiliconFlow API Key",
		Pattern:     siliconflowPattern,
		Platform:    "siliconflow",
		Confidence:  0.9,
		Description: "SiliconFlow API key pattern",
	}

	// Generic API key patterns
	genericPattern1, _ := regexp.Compile(`[A-Za-z0-9]{32,}`)
	skd.patterns["generic_long"] = &RegexPattern{
		Name:        "Generic Long Key",
		Pattern:     genericPattern1,
		Platform:    "unknown",
		Confidence:  0.3,
		Description: "Generic long alphanumeric key",
	}

	genericPattern2, _ := regexp.Compile(`[A-Za-z0-9\\-_]{20,}`)
	skd.patterns["generic_medium"] = &RegexPattern{
		Name:        "Generic Medium Key",
		Pattern:     genericPattern2,
		Platform:    "unknown",
		Confidence:  0.2,
		Description: "Generic medium alphanumeric key",
	}
}

// initializeValidators initializes key validators for different platforms
func (skd *SmartKeyDetector) initializeValidators() {
	skd.validators["gemini"] = &GeminiValidator{}
	skd.validators["openrouter"] = &OpenRouterValidator{}
	skd.validators["siliconflow"] = &SiliconFlowValidator{}
}

// GetDetectionRate returns the detection rate
func (skd *SmartKeyDetector) GetDetectionRate() float64 {
	skd.mutex.RLock()
	defer skd.mutex.RUnlock()

	if skd.metrics.TotalDetections == 0 {
		return 0.0
	}
	return float64(skd.metrics.ValidDetections) / float64(skd.metrics.TotalDetections)
}

// GetMetrics returns detection metrics
func (skd *SmartKeyDetector) GetMetrics() *DetectionMetrics {
	skd.mutex.RLock()
	defer skd.mutex.RUnlock()

	// Calculate rates
	if skd.metrics.TotalDetections > 0 {
		skd.metrics.DetectionRate = float64(skd.metrics.ValidDetections) / float64(skd.metrics.TotalDetections)
		skd.metrics.FalsePositiveRate = float64(skd.metrics.PlaceholderFiltered+skd.metrics.TestKeyFiltered+skd.metrics.LowConfidenceFiltered) / float64(skd.metrics.TotalDetections)
	}

	return &DetectionMetrics{
		TotalDetections:      skd.metrics.TotalDetections,
		ValidDetections:      skd.metrics.ValidDetections,
		PlaceholderFiltered:  skd.metrics.PlaceholderFiltered,
		TestKeyFiltered:      skd.metrics.TestKeyFiltered,
		LowConfidenceFiltered: skd.metrics.LowConfidenceFiltered,
		DetectionRate:        skd.metrics.DetectionRate,
		FalsePositiveRate:    skd.metrics.FalsePositiveRate,
	}
}

// NewContextAnalyzer creates a new context analyzer
func NewContextAnalyzer() *ContextAnalyzer {
	ca := &ContextAnalyzer{
		contextWindow: 50, // characters around the key
	}

	// Initialize placeholder patterns
	ca.placeholderPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(your[_-]?api[_-]?key|api[_-]?key[_-]?here|replace[_-]?with[_-]?your[_-]?key)`),
		regexp.MustCompile(`(?i)(placeholder|example|sample|test[_-]?key)`),
		regexp.MustCompile(`(?i)(xxx+|aaa+|123+|000+)`),
		regexp.MustCompile(`(?i)(key[_-]?goes[_-]?here|insert[_-]?key[_-]?here)`),
	}

	// Initialize test key patterns
	ca.testKeyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(test[_-]?key|demo[_-]?key|sample[_-]?key)`),
		regexp.MustCompile(`(?i)(sk[_-]?test|api[_-]?test|key[_-]?test)`),
		regexp.MustCompile(`(?i)(1234567890|abcdefghij|test123|demo123)`),
	}

	return ca
}

// IsPlaceholder checks if a key is a placeholder
func (ca *ContextAnalyzer) IsPlaceholder(key, context string) bool {
	contextLower := strings.ToLower(context)
	for _, pattern := range ca.placeholderPatterns {
		if pattern.MatchString(contextLower) {
			return true
		}
	}
	return false
}

// IsTestKey checks if a key is a test key
func (ca *ContextAnalyzer) IsTestKey(key, context string) bool {
	contextLower := strings.ToLower(context)
	keyLower := strings.ToLower(key)
	
	for _, pattern := range ca.testKeyPatterns {
		if pattern.MatchString(contextLower) || pattern.MatchString(keyLower) {
			return true
		}
	}
	return false
}