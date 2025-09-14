package filemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hajimi-king-go-v2/internal/config"
	"hajimi-king-go-v2/internal/models"
)

// FileManager manages file operations
type FileManager struct {
	config *config.Config
}

// NewFileManager creates a new file manager
func NewFileManager(cfg *config.Config) (*FileManager, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileManager{
		config: cfg,
	}, nil
}

// SaveValidKeysForPlatform saves valid keys for a specific platform
func (fm *FileManager) SaveValidKeysForPlatform(platform, repoName, filePath, fileURL string, keys []string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.txt", fm.config.ValidKeyPrefix, platform, timestamp)
	fullPath := filepath.Join(fm.config.DataPath, filename)

	// Create content
	content := fmt.Sprintf("Platform: %s\nRepository: %s\nFile: %s\nURL: %s\nKeys:\n", platform, repoName, filePath, fileURL)
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write valid keys file: %w", err)
	}

	// Save detailed log
	detailFilename := fmt.Sprintf("%s_%s_%s.log", fm.config.ValidKeyDetailPrefix, platform, timestamp)
	detailPath := filepath.Join(fm.config.DataPath, detailFilename)

	detailContent := fmt.Sprintf("[%s] Valid keys found for platform %s\n", time.Now().Format(time.RFC3339), platform)
	detailContent += fmt.Sprintf("Repository: %s\n", repoName)
	detailContent += fmt.Sprintf("File: %s\n", filePath)
	detailContent += fmt.Sprintf("URL: %s\n", fileURL)
	detailContent += fmt.Sprintf("Keys count: %d\n", len(keys))
	detailContent += "---\n"

	if err := os.WriteFile(detailPath, []byte(detailContent), 0644); err != nil {
		return fmt.Errorf("failed to write detail log: %w", err)
	}

	return nil
}

// SaveRateLimitedKeysForPlatform saves rate limited keys for a specific platform
func (fm *FileManager) SaveRateLimitedKeysForPlatform(platform, repoName, filePath, fileURL string, keys []string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.txt", fm.config.RateLimitedPrefix, platform, timestamp)
	fullPath := filepath.Join(fm.config.DataPath, filename)

	// Create content
	content := fmt.Sprintf("Platform: %s\nRepository: %s\nFile: %s\nURL: %s\nRate Limited Keys:\n", platform, repoName, filePath, fileURL)
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write rate limited keys file: %w", err)
	}

	return nil
}

// LoadCheckpoint loads checkpoint data
func (fm *FileManager) LoadCheckpoint() (*models.Checkpoint, error) {
	if _, err := os.Stat(fm.config.CheckpointPath); os.IsNotExist(err) {
		return &models.Checkpoint{
			LastScanTime:  time.Now(),
			ProcessedFiles: make(map[string]bool),
		}, nil
	}

	data, err := os.ReadFile(fm.config.CheckpointPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint models.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// SaveCheckpoint saves checkpoint data
func (fm *FileManager) SaveCheckpoint(checkpoint *models.Checkpoint) error {
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(fm.config.CheckpointPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// LoadSearchQueries loads search queries from file
func (fm *FileManager) LoadSearchQueries() ([]string, error) {
	if _, err := os.Stat(fm.config.QueriesPath); os.IsNotExist(err) {
		// Create default queries file
		defaultQueries := []string{
			"AIza",
			"sk-or-",
			"sk-",
			"api_key",
			"API_KEY",
			"secret",
			"token",
		}
		return defaultQueries, nil
	}

	data, err := os.ReadFile(fm.config.QueriesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read queries file: %w", err)
	}

	var queries []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			queries = append(queries, line)
		}
	}

	return queries, nil
}

// GetSearchQueries returns search queries
func (fm *FileManager) GetSearchQueries() []string {
	queries, err := fm.LoadSearchQueries()
	if err != nil {
		return []string{"AIza", "sk-or-", "sk-"}
	}
	return queries
}