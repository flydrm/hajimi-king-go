package filemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/logger"
	"hajimi-king-go/internal/models"
)

// FileManager 文件管理器
type FileManager struct {
	config *config.Config
}

// NewFileManager 创建文件管理器
func NewFileManager(cfg *config.Config) *FileManager {
	return &FileManager{
		config: cfg,
	}
}

// Check 检查文件管理器是否就绪
func (fm *FileManager) Check() bool {
	// 创建数据目录
	if err := os.MkdirAll(fm.config.DataPath, 0755); err != nil {
		logger.GetLogger().Errorf("❌ Failed to create data directory: %v", err)
		return false
	}

	// 创建子目录
	subdirs := []string{"keys", "logs"}
	for _, subdir := range subdirs {
		fullPath := filepath.Join(fm.config.DataPath, subdir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			logger.GetLogger().Errorf("❌ Failed to create directory %s: %v", fullPath, err)
			return false
		}
	}

	logger.GetLogger().Infof("✅ Data directories created: %s", fm.config.DataPath)
	return true
}

// GetSearchQueries 获取搜索查询列表
func (fm *FileManager) GetSearchQueries() []string {
	queriesFile := filepath.Join(fm.config.DataPath, fm.config.QueriesFile)
	content, err := os.ReadFile(queriesFile)
	if err != nil {
		logger.GetLogger().Errorf("❌ Failed to read queries file: %v", err)
		return []string{}
	}

	lines := strings.Split(string(content), "\n")
	var queries []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			queries = append(queries, line)
		}
	}

	logger.GetLogger().Infof("📋 Loaded %d search queries", len(queries))
	return queries
}

// SaveCheckpoint 保存检查点
func (fm *FileManager) SaveCheckpoint(checkpoint *models.Checkpoint) error {
	checkpointFile := filepath.Join(fm.config.DataPath, fm.config.ScannedSHAsFile)
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %v", err)
	}

	if err := os.WriteFile(checkpointFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %v", err)
	}

	logger.GetLogger().Infof("💾 Checkpoint saved: %s", checkpointFile)
	return nil
}

// LoadCheckpoint 加载检查点
func (fm *FileManager) LoadCheckpoint() (*models.Checkpoint, error) {
	checkpointFile := filepath.Join(fm.config.DataPath, fm.config.ScannedSHAsFile)
	
	// 如果文件不存在，返回空的检查点
	if _, err := os.Stat(checkpointFile); os.IsNotExist(err) {
		return &models.Checkpoint{
			LastScanTime:     "",
			ScannedSHAs:      []string{},
			ProcessedQueries: []string{},
			WaitSendBalancer: []string{},
			WaitSendGPTLoad:  []string{},
		}, nil
	}

	content, err := os.ReadFile(checkpointFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %v", err)
	}

	var checkpoint models.Checkpoint
	if err := json.Unmarshal(content, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %v", err)
	}

	logger.GetLogger().Infof("💾 Checkpoint loaded: %d scanned SHAs, %d processed queries", 
		len(checkpoint.ScannedSHAs), len(checkpoint.ProcessedQueries))
	return &checkpoint, nil
}

// SaveValidKeys 保存有效密钥
func (fm *FileManager) SaveValidKeys(repoName, filePath, fileURL string, keys []string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s%s.txt", fm.config.ValidKeyPrefix, timestamp)
	fullPath := filepath.Join(fm.config.DataPath, filename)

	// 创建文件内容
	var content strings.Builder
	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s|%s|%s|%s\n", key, repoName, filePath, fileURL))
	}

	if err := os.WriteFile(fullPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to save valid keys: %v", err)
	}

	// 保存详细信息
	detailFilename := fmt.Sprintf("%s%s.log", fm.config.ValidKeyDetailPrefix, timestamp)
	detailFullPath := filepath.Join(fm.config.DataPath, detailFilename)
	detailContent := fmt.Sprintf("[%s] Found %d valid keys in %s/%s\n%s\n", 
		time.Now().Format("2006-01-02 15:04:05"), len(keys), repoName, filePath, content.String())

	if err := os.WriteFile(detailFullPath, []byte(detailContent), 0644); err != nil {
		logger.GetLogger().Warningf("⚠️ Failed to save detail log: %v", err)
	}

	logger.GetLogger().Infof("💾 Saved %d valid keys to %s", len(keys), filename)
	return nil
}

// SaveRateLimitedKeys 保存被限流的密钥
func (fm *FileManager) SaveRateLimitedKeys(repoName, filePath, fileURL string, keys []string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s%s.txt", fm.config.RateLimitedKeyPrefix, timestamp)
	fullPath := filepath.Join(fm.config.DataPath, filename)

	// 创建文件内容
	var content strings.Builder
	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s|%s|%s|%s\n", key, repoName, filePath, fileURL))
	}

	if err := os.WriteFile(fullPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to save rate limited keys: %v", err)
	}

	// 保存详细信息
	detailFilename := fmt.Sprintf("%s%s.log", fm.config.RateLimitedKeyDetailPrefix, timestamp)
	detailFullPath := filepath.Join(fm.config.DataPath, detailFilename)
	detailContent := fmt.Sprintf("[%s] Found %d rate limited keys in %s/%s\n%s\n", 
		time.Now().Format("2006-01-02 15:04:05"), len(keys), repoName, filePath, content.String())

	if err := os.WriteFile(detailFullPath, []byte(detailContent), 0644); err != nil {
		logger.GetLogger().Warningf("⚠️ Failed to save detail log: %v", err)
	}

	logger.GetLogger().Infof("💾 Saved %d rate limited keys to %s", len(keys), filename)
	return nil
}

// UpdateDynamicFilenames 更新动态文件名
func (fm *FileManager) UpdateDynamicFilenames() {
	// 这个函数可以用于实现动态文件名更新逻辑
	// 目前保持为空，根据需要实现
}

// NormalizeQuery 规范化查询字符串
func (fm *FileManager) NormalizeQuery(query string) string {
	query = strings.Join(strings.Fields(query), " ")

	var parts []string
	i := 0
	for i < len(query) {
		if query[i] == '"' {
			endQuote := strings.Index(query[i+1:], "\"")
			if endQuote != -1 {
				parts = append(parts, query[i:i+endQuote+2])
				i += endQuote + 2
			} else {
				parts = append(parts, string(query[i]))
				i++
			}
		} else if query[i] == ' ' {
			i++
		} else {
			start := i
			for i < len(query) && query[i] != ' ' {
				i++
			}
			parts = append(parts, query[start:i])
		}
	}

	var quotedStrings, languageParts, filenameParts, pathParts, otherParts []string
	for _, part := range parts {
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			quotedStrings = append(quotedStrings, part)
		} else if strings.HasPrefix(part, "language:") {
			languageParts = append(languageParts, part)
		} else if strings.HasPrefix(part, "filename:") {
			filenameParts = append(filenameParts, part)
		} else if strings.HasPrefix(part, "path:") {
			pathParts = append(pathParts, part)
		} else if strings.TrimSpace(part) != "" {
			otherParts = append(otherParts, part)
		}
	}

	// 排序并重新组合
	var normalizedParts []string
	normalizedParts = append(normalizedParts, quotedStrings...)
	normalizedParts = append(normalizedParts, otherParts...)
	normalizedParts = append(normalizedParts, languageParts...)
	normalizedParts = append(normalizedParts, filenameParts...)
	normalizedParts = append(normalizedParts, pathParts...)

	return strings.Join(normalizedParts, " ")
}