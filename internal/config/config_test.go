package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Setenv("PLATFORM_GEMINI_ENABLED", "true")
	os.Setenv("PLATFORM_OPENROUTER_ENABLED", "false")
	os.Setenv("PLATFORM_SILICONFLOW_ENABLED", "false")
	os.Setenv("PLATFORM_EXECUTION_MODE", "all")
	
	config := LoadConfig()
	
	if config.GitHubToken != "test_token" {
		t.Errorf("Expected GitHubToken to be 'test_token', got '%s'", config.GitHubToken)
	}
	
	if !config.PlatformSwitches.Platforms["gemini"].Enabled {
		t.Error("Expected Gemini platform to be enabled")
	}
	
	if config.PlatformSwitches.Platforms["openrouter"].Enabled {
		t.Error("Expected OpenRouter platform to be disabled")
	}
	
	if config.PlatformSwitches.Platforms["siliconflow"].Enabled {
		t.Error("Expected SiliconFlow platform to be disabled")
	}
}

func TestGetEnabledPlatforms(t *testing.T) {
	config := &Config{
		PlatformSwitches: PlatformSwitches{
			GlobalEnabled: true,
			ExecutionMode: "all",
			Platforms: map[string]PlatformConfig{
				"gemini": {
					Enabled:  true,
					Priority: 1,
				},
				"openrouter": {
					Enabled:  false,
					Priority: 2,
				},
				"siliconflow": {
					Enabled:  true,
					Priority: 3,
				},
			},
		},
	}
	
	enabledPlatforms := config.GetEnabledPlatforms()
	
	if len(enabledPlatforms) != 2 {
		t.Errorf("Expected 2 enabled platforms, got %d", len(enabledPlatforms))
	}
	
	if enabledPlatforms[0] != "gemini" {
		t.Errorf("Expected first platform to be 'gemini', got '%s'", enabledPlatforms[0])
	}
	
	if enabledPlatforms[1] != "siliconflow" {
		t.Errorf("Expected second platform to be 'siliconflow', got '%s'", enabledPlatforms[1])
	}
}