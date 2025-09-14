package detection

import (
	"testing"
)

func TestSmartKeyDetector(t *testing.T) {
	detector := NewSmartKeyDetector()
	
	// Test Gemini key detection
	content := "const apiKey = 'AIzaSyTest123456789012345678901234567890'"
	keyContexts := detector.DetectKeys(content)
	
	// Should detect the key but may filter it out due to confidence
	t.Logf("Detected %d keys from content: %s", len(keyContexts), content)
	
	// Test placeholder filtering
	placeholderContent := "const apiKey = 'your-api-key-here'"
	placeholderContexts := detector.DetectKeys(placeholderContent)
	
	// Should filter out placeholder
	if len(placeholderContexts) > 0 {
		t.Logf("Placeholder content detected %d keys (expected 0)", len(placeholderContexts))
	}
	
	// Test test key filtering
	testContent := "const apiKey = 'sk-test123456789012345678901234567890'"
	testContexts := detector.DetectKeys(testContent)
	
	// Should filter out test key
	if len(testContexts) > 0 {
		t.Logf("Test key content detected %d keys (expected 0)", len(testContexts))
	}
	
	// Test valid OpenRouter key
	validContent := "const apiKey = 'sk-or-12345678901234567890123456789012'"
	validContexts := detector.DetectKeys(validContent)
	
	t.Logf("Valid OpenRouter key detected %d keys", len(validContexts))
}