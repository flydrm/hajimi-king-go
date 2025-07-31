package syncutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/logger"
)

// SyncUtils 同步工具
type SyncUtils struct {
	config           *config.Config
	BalancerEnabled  bool
	balancerQueue    []string
	GPTLoadEnabled   bool
	gptLoadQueue     []string
	lastSyncTime     time.Time
	syncInterval     time.Duration
	stopChan         chan struct{}
}

// NewSyncUtils 创建同步工具
func NewSyncUtils(cfg *config.Config) *SyncUtils {
	return &SyncUtils{
		config:          cfg,
		BalancerEnabled: cfg.GeminiBalancerSyncEnabled,
		balancerQueue:   []string{},
		GPTLoadEnabled:  cfg.GPTLoadSyncEnabled,
		gptLoadQueue:    []string{},
		syncInterval:    5 * time.Minute, // 默认5分钟同步一次
		stopChan:        make(chan struct{}),
	}
}

// Start 启动同步服务
func (su *SyncUtils) Start() {
	if !su.BalancerEnabled && !su.GPTLoadEnabled {
		logger.GetLogger().Info("ℹ️ No sync services enabled")
		return
	}

	logger.GetLogger().Info("🔗 Starting sync services...")
	go su.syncWorker()
}

// Stop 停止同步服务
func (su *SyncUtils) Stop() {
	close(su.stopChan)
	logger.GetLogger().Info("🔚 Sync services stopped")
}

// AddKeysToQueue 添加密钥到同步队列
func (su *SyncUtils) AddKeysToQueue(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	if su.BalancerEnabled {
		su.balancerQueue = append(su.balancerQueue, keys...)
		logger.GetLogger().Infof("📥 Added %d keys to balancer queue", len(keys))
	}

	if su.GPTLoadEnabled {
		su.gptLoadQueue = append(su.gptLoadQueue, keys...)
		logger.GetLogger().Infof("📥 Added %d keys to GPT Load queue", len(keys))
	}

	return nil
}

// syncWorker 同步工作协程
func (su *SyncUtils) syncWorker() {
	ticker := time.NewTicker(su.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			su.doSync()
		case <-su.stopChan:
			return
		}
	}
}

// doSync 执行同步操作
func (su *SyncUtils) doSync() {
	if len(su.balancerQueue) > 0 {
		if err := su.syncToBalancer(); err != nil {
			logger.GetLogger().Errorf("❌ Failed to sync to balancer: %v", err)
		}
	}

	if len(su.gptLoadQueue) > 0 {
		if err := su.syncToGPTLoad(); err != nil {
			logger.GetLogger().Errorf("❌ Failed to sync to GPT Load: %v", err)
		}
	}
}

// syncToBalancer 同步到Gemini Balancer
func (su *SyncUtils) syncToBalancer() error {
	if len(su.balancerQueue) == 0 {
		return nil
	}

	logger.GetLogger().Infof("🔗 Syncing %d keys to Gemini Balancer...", len(su.balancerQueue))

	// 构建请求数据
	data := map[string]interface{}{
		"keys": su.balancerQueue,
		"auth": su.config.GeminiBalancerAuth,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal balancer data: %v", err)
	}

	// 发送请求
	resp, err := http.Post(su.config.GeminiBalancerURL+"/api/add-keys", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request to balancer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("balancer returned status code: %d", resp.StatusCode)
	}

	// 清空队列
	su.balancerQueue = []string{}
	logger.GetLogger().Infof("✅ Successfully synced keys to Gemini Balancer")
	return nil
}

// syncToGPTLoad 同步到GPT Load
func (su *SyncUtils) syncToGPTLoad() error {
	if len(su.gptLoadQueue) == 0 {
		return nil
	}

	logger.GetLogger().Infof("🔗 Syncing %d keys to GPT Load...", len(su.gptLoadQueue))

	// 解析组名
	groups := strings.Split(su.config.GPTLoadGroupName, ",")
	for i, group := range groups {
		groups[i] = strings.TrimSpace(group)
	}

	// 为每个组发送密钥
	for _, group := range groups {
		if group == "" {
			continue
		}

		// 构建请求数据
		data := map[string]interface{}{
			"keys":  su.gptLoadQueue,
			"group": group,
			"auth":  su.config.GPTLoadAuth,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal GPT Load data: %v", err)
		}

		// 发送请求
		resp, err := http.Post(su.config.GPTLoadURL+"/api/add-keys", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to send request to GPT Load: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("GPT Load returned status code: %d", resp.StatusCode)
		}
	}

	// 清空队列
	su.gptLoadQueue = []string{}
	logger.GetLogger().Infof("✅ Successfully synced keys to GPT Load")
	return nil
}

// GetQueueStatus 获取队列状态
func (su *SyncUtils) GetQueueStatus() (int, int) {
	return len(su.balancerQueue), len(su.gptLoadQueue)
}

// SetSyncInterval 设置同步间隔
func (su *SyncUtils) SetSyncInterval(interval time.Duration) {
	su.syncInterval = interval
	logger.GetLogger().Infof("🕐 Sync interval set to %v", interval)
}