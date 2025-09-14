package syncutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hajimi-king-go-v2/internal/config"
)

// SyncUtils handles external synchronization
type SyncUtils struct {
	config *config.Config
	client *http.Client
}

// SyncRequest represents a sync request
type SyncRequest struct {
	Keys     []KeyInfo `json:"keys"`
	Platform string    `json:"platform"`
	Timestamp time.Time `json:"timestamp"`
}

// KeyInfo represents key information for sync
type KeyInfo struct {
	Key        string `json:"key"`
	Platform   string `json:"platform"`
	Repository string `json:"repository"`
	FilePath   string `json:"file_path"`
	IsValid    bool   `json:"is_valid"`
}

// NewSyncUtils creates a new sync utils instance
func NewSyncUtils(cfg *config.Config) (*SyncUtils, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &SyncUtils{
		config: cfg,
		client: client,
	}, nil
}

// SyncToGeminiBalancer syncs keys to Gemini Balancer
func (su *SyncUtils) SyncToGeminiBalancer(keys []KeyInfo) error {
	if !su.config.SyncToGeminiBalancer {
		return nil
	}

	// Filter Gemini keys
	var geminiKeys []KeyInfo
	for _, key := range keys {
		if key.Platform == "gemini" && key.IsValid {
			geminiKeys = append(geminiKeys, key)
		}
	}

	if len(geminiKeys) == 0 {
		return nil
	}

	// Create sync request
	request := SyncRequest{
		Keys:      geminiKeys,
		Platform:  "gemini",
		Timestamp: time.Now(),
	}

	return su.sendSyncRequest(request, "/api/gemini/keys")
}

// SyncToGPTLoad syncs keys to GPT Load Balancer
func (su *SyncUtils) SyncToGPTLoad(keys []KeyInfo) error {
	if !su.config.SyncToGPTLoad {
		return nil
	}

	// Filter valid keys
	var validKeys []KeyInfo
	for _, key := range keys {
		if key.IsValid {
			validKeys = append(validKeys, key)
		}
	}

	if len(validKeys) == 0 {
		return nil
	}

	// Create sync request
	request := SyncRequest{
		Keys:      validKeys,
		Platform:  "all",
		Timestamp: time.Now(),
	}

	return su.sendSyncRequest(request, "/api/gpt/keys")
}

// sendSyncRequest sends a sync request
func (su *SyncUtils) sendSyncRequest(request SyncRequest, endpoint string) error {
	if su.config.SyncEndpoint == "" {
		return fmt.Errorf("sync endpoint not configured")
	}

	// Marshal request
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal sync request: %w", err)
	}

	// Create HTTP request
	url := su.config.SyncEndpoint + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create sync request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if su.config.SyncToken != "" {
		req.Header.Set("Authorization", "Bearer "+su.config.SyncToken)
	}

	// Send request
	resp, err := su.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sync request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != 200 {
		return fmt.Errorf("sync request failed with status: %d", resp.StatusCode)
	}

	return nil
}