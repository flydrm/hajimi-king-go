package syncutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/logger"
)

// GroupInfo GPT Load Group信息
type GroupInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GroupsResponse GPT Load Groups API响应
type GroupsResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []GroupInfo `json:"data"`
}

// AddKeysResponse GPT Load Add Keys API响应
type AddKeysResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

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
	
	// GPT Load Group ID缓存
	groupIDCache     map[string]int
	groupIDCacheTime map[string]time.Time
	groupIDCacheTTL  time.Duration
	cacheMutex       sync.RWMutex
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
		
		// 初始化GPT Load Group ID缓存
		groupIDCache:     make(map[string]int),
		groupIDCacheTime: make(map[string]time.Time),
		groupIDCacheTTL:  15 * time.Minute, // 15分钟缓存
		cacheMutex:       sync.RWMutex{},
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

	// 1. 获取当前配置
	configURL := su.config.GeminiBalancerURL + "/api/config"
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create config request: %v", err)
	}
	
	// 设置认证头 - 使用Cookie认证
	req.Header.Set("Cookie", "auth_token="+su.config.GeminiBalancerAuth)
	req.Header.Set("User-Agent", "HajimiKing/1.0")
	
	logger.GetLogger().Infof("📥 Fetching current config from: %s", configURL)
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get config: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get config: HTTP %d", resp.StatusCode)
	}
	
	// 解析配置响应
	var configData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&configData); err != nil {
		return fmt.Errorf("failed to decode config response: %v", err)
	}
	
	// 2. 获取当前的API_KEYS数组
	currentAPIKeys, ok := configData["API_KEYS"].([]interface{})
	if !ok {
		currentAPIKeys = []interface{}{}
	}
	
	// 3. 合并新keys（去重）
	existingKeysSet := make(map[string]bool)
	for _, key := range currentAPIKeys {
		if keyStr, ok := key.(string); ok {
			existingKeysSet[keyStr] = true
		}
	}
	
	newAddKeysSet := make(map[string]bool)
	for _, key := range su.balancerQueue {
		if !existingKeysSet[key] {
			existingKeysSet[key] = true
			newAddKeysSet[key] = true
		}
	}
	
	if len(newAddKeysSet) == 0 {
		logger.GetLogger().Infof("ℹ️ All %d key(s) already exist in balancer", len(su.balancerQueue))
		// 清空队列
		su.balancerQueue = []string{}
		return nil
	}
	
	// 4. 更新配置中的API_KEYS
	var updatedKeys []string
	for key := range existingKeysSet {
		updatedKeys = append(updatedKeys, key)
	}
	configData["API_KEYS"] = updatedKeys
	
	logger.GetLogger().Infof("📝 Updating gemini balancer config with %d new key(s)...", len(newAddKeysSet))
	
	// 5. 发送更新后的配置到服务器
	jsonData, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %v", err)
	}
	
	req, err = http.NewRequest("PUT", configURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create update request: %v", err)
	}
	
	// 设置认证头
	req.Header.Set("Cookie", "auth_token="+su.config.GeminiBalancerAuth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HajimiKing/1.0")
	
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update config: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to update config: HTTP %d", resp.StatusCode)
	}
	
	// 6. 验证是否添加成功
	var updatedConfig map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&updatedConfig); err != nil {
		return fmt.Errorf("failed to decode update response: %v", err)
	}
	
	updatedAPIKeys, ok := updatedConfig["API_KEYS"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid API_KEYS format in response")
	}
	
	updatedKeysSet := make(map[string]bool)
	for _, key := range updatedAPIKeys {
		if keyStr, ok := key.(string); ok {
			updatedKeysSet[keyStr] = true
		}
	}
	
	var failedToAdd []string
	for key := range newAddKeysSet {
		if !updatedKeysSet[key] {
			failedToAdd = append(failedToAdd, key)
		}
	}
	
	if len(failedToAdd) > 0 {
		return fmt.Errorf("failed to add %d keys", len(failedToAdd))
	}
	
	logger.GetLogger().Infof("✅ All %d new key(s) successfully added to balancer", len(newAddKeysSet))
	
	// 清空队列
	su.balancerQueue = []string{}
	return nil
}

// getGPTLoadGroupID 获取GPT Load Group ID，带缓存功能
func (su *SyncUtils) getGPTLoadGroupID(groupName string) (int, error) {
	// 检查缓存
	su.cacheMutex.RLock()
	cachedID, hasCache := su.groupIDCache[groupName]
	cacheTime, hasTime := su.groupIDCacheTime[groupName]
	su.cacheMutex.RUnlock()
	
	if hasCache && hasTime && time.Since(cacheTime) < su.groupIDCacheTTL {
		logger.GetLogger().Infof("📋 Using cached group ID for '%s': %d", groupName, cachedID)
		return cachedID, nil
	}
	
	// 缓存不存在或过期，重新获取
	groupsURL := su.config.GPTLoadURL + "/api/groups"
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}
	
	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+su.config.GPTLoadAuth)
	req.Header.Set("User-Agent", "HajimiKing/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get groups: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("groups API returned status code: %d", resp.StatusCode)
	}
	
	var groupsResp GroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groupsResp); err != nil {
		return 0, fmt.Errorf("failed to decode groups response: %v", err)
	}
	
	if groupsResp.Code != 0 {
		return 0, fmt.Errorf("groups API returned error: %s", groupsResp.Message)
	}
	
	// 查找指定group的ID
	for _, group := range groupsResp.Data {
		if group.Name == groupName {
			// 更新缓存
			su.cacheMutex.Lock()
			su.groupIDCache[groupName] = group.ID
			su.groupIDCacheTime[groupName] = time.Now()
			su.cacheMutex.Unlock()
			
			logger.GetLogger().Infof("✅ Found and cached group '%s' with ID: %d", groupName, group.ID)
			return group.ID, nil
		}
	}
	
	return 0, fmt.Errorf("group '%s' not found", groupName)
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

	// 为每个组并发发送密钥
	var wg sync.WaitGroup
	failedGroupsChan := make(chan string, len(groups))

	for _, group := range groups {
		if group == "" {
			continue
		}
		wg.Add(1)
		go func(g string) {
			defer wg.Done()
			
			logger.GetLogger().Infof("📝 Processing group: %s", g)
			
			// 1. 获取group ID
			groupID, err := su.getGPTLoadGroupID(g)
			if err != nil {
				logger.GetLogger().Errorf("❌ Failed to get group ID for '%s': %v", g, err)
				failedGroupsChan <- g
				return
			}
			
			// 2. 发送keys到指定group
			addKeysURL := su.config.GPTLoadURL + "/api/keys/add-async"
			keysText := strings.Join(su.gptLoadQueue, ",")
			
			data := map[string]interface{}{
				"group_id":   groupID,
				"keys_text":  keysText,
			}
			
			jsonData, err := json.Marshal(data)
			if err != nil {
				logger.GetLogger().Errorf("❌ Failed to marshal GPT Load data for group '%s': %v", g, err)
				failedGroupsChan <- g
				return
			}
			
			client := &http.Client{
				Timeout: 60 * time.Second,
			}
			
			req, err := http.NewRequest("POST", addKeysURL, bytes.NewBuffer(jsonData))
			if err != nil {
				logger.GetLogger().Errorf("❌ Failed to create request for group '%s': %v", g, err)
				failedGroupsChan <- g
				return
			}
			
			// 设置认证头
			req.Header.Set("Authorization", "Bearer "+su.config.GPTLoadAuth)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "HajimiKing/1.0")
			
			resp, err := client.Do(req)
			if err != nil {
				logger.GetLogger().Errorf("❌ Failed to send request to GPT Load for group '%s': %v", g, err)
				failedGroupsChan <- g
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != 200 {
				logger.GetLogger().Errorf("❌ GPT Load returned status code %d for group '%s'", resp.StatusCode, g)
				failedGroupsChan <- g
				return
			}
			
			var addResp AddKeysResponse
			if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
				logger.GetLogger().Errorf("❌ Failed to decode add keys response for group '%s': %v", g, err)
				failedGroupsChan <- g
				return
			}
			
			// 检查响应状态
			if addResp.Code != "0" && addResp.Code != "success" {
				logger.GetLogger().Errorf("❌ Add keys API returned error for group '%s': %s", g, addResp.Message)
				failedGroupsChan <- g
				return
			}
			
			// 检查任务数据
			if taskData, ok := addResp.Data.(map[string]interface{}); ok {
				taskType := taskData["task_type"]
				isRunning := taskData["is_running"]
				total := taskData["total"]
				responseGroupName := taskData["group_name"]
				
				logger.GetLogger().Infof("✅ Keys addition task started successfully for group '%s':", g)
				logger.GetLogger().Infof("   Task Type: %v", taskType)
				logger.GetLogger().Infof("   Is Running: %v", isRunning)
				logger.GetLogger().Infof("   Total Keys: %v", total)
				logger.GetLogger().Infof("   Group Name: %v", responseGroupName)
			}
		}(group)
	}

	wg.Wait()
	close(failedGroupsChan)

	var failedGroups []string
	for g := range failedGroupsChan {
		failedGroups = append(failedGroups, g)
	}
	
	// 根据结果返回状态
	if len(failedGroups) == 0 {
		logger.GetLogger().Infof("✅ Successfully sent keys to all %d group(s)", len(groups))
		// 清空队列
		su.gptLoadQueue = []string{}
		return nil
	} else {
		logger.GetLogger().Errorf("❌ Failed to send keys to %d group(s): %s", len(failedGroups), strings.Join(failedGroups, ", "))
		return fmt.Errorf("failed to send keys to %d groups", len(failedGroups))
	}
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