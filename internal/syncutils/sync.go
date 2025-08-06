package syncutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/filemanager"
	"hajimi-king-go/internal/logger"
	"hajimi-king-go/internal/models"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SyncUtils 同步工具类，负责异步发送keys到外部应用
type SyncUtils struct {
	config *config.Config

	// Gemini Balancer
	balancerEnabled bool
	balancerURL     string
	balancerAuth    string

	// GPT Load Balancer
	gptLoadEnabled    bool
	gptLoadURL        string
	gptLoadAuth       string
	gptLoadGroupNames []string
	groupIDCache      map[string]int
	groupIDCacheTime  map[string]time.Time
	groupIDCacheTTL   time.Duration
	groupIDCacheMutex sync.RWMutex

	// 异步执行
	executor      chan func()
	shutdownFlag  bool
	shutdownMutex sync.Mutex

	// 周期性发送
	batchInterval time.Duration
	batchTimer    *time.Timer

	// 检查点
	checkpoint      *models.Checkpoint
	checkpointMutex sync.Mutex
	fileManager     *filemanager.FileManager

	httpClient *http.Client
}

// NewSyncUtils 创建一个新的 SyncUtils 实例
func NewSyncUtils(cfg *config.Config, cp *models.Checkpoint, fm *filemanager.FileManager) *SyncUtils {
	su := &SyncUtils{
		config: cfg,

		balancerURL:  strings.TrimRight(cfg.GeminiBalancerURL, "/"),
		balancerAuth: cfg.GeminiBalancerAuth,

		gptLoadURL:        strings.TrimRight(cfg.GPTLoadURL, "/"),
		gptLoadAuth:       cfg.GPTLoadAuth,
		gptLoadGroupNames: parseGroupNames(cfg.GPTLoadGroupName),

		groupIDCache:     make(map[string]int),
		groupIDCacheTime: make(map[string]time.Time),
		groupIDCacheTTL:  15 * time.Minute,

		executor:      make(chan func(), 2), // 类似Python的ThreadPoolExecutor(max_workers=2)
		shutdownFlag:  false,
		batchInterval: 60 * time.Second,
		checkpoint:    cp,
		fileManager:   fm,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	su.balancerEnabled = su.balancerURL != "" && su.balancerAuth != "" && cfg.GeminiBalancerSyncEnabled
	su.gptLoadEnabled = su.gptLoadURL != "" && su.gptLoadAuth != "" && len(su.gptLoadGroupNames) > 0 && cfg.GPTLoadSyncEnabled

	if !su.balancerEnabled {
		logger.GetLogger().Warning("🚫 Gemini Balancer sync disabled - URL or AUTH not configured or sync disabled")
	} else {
		logger.GetLogger().Infof("🔗 Gemini Balancer enabled - URL: %s", su.balancerURL)
	}

	if !su.gptLoadEnabled {
		logger.GetLogger().Warning("🚫 GPT Load Balancer sync disabled - URL, AUTH, GROUP_NAME not configured or sync disabled")
	} else {
		logger.GetLogger().Infof("🔗 GPT Load Balancer enabled - URL: %s, Groups: %s", su.gptLoadURL, strings.Join(su.gptLoadGroupNames, ", "))
	}

	return su
}

// IsBalancerEnabled 返回 Gemin Balancer 是否启用
func (su *SyncUtils) IsBalancerEnabled() bool {
	return su.balancerEnabled
}

// IsGPTLoadEnabled 返回 GPT Load Balancer 是否启用
func (su *SyncUtils) IsGPTLoadEnabled() bool {
	return su.gptLoadEnabled
}

// GetQueueStatus 返回队列状态
func (su *SyncUtils) GetQueueStatus() (int, int) {
	su.checkpointMutex.Lock()
	defer su.checkpointMutex.Unlock()
	return len(su.checkpoint.WaitSendBalancer), len(su.checkpoint.WaitSendGPTLoad)
}

func parseGroupNames(groupNames string) []string {
	if groupNames == "" {
		return []string{}
	}
	names := strings.Split(groupNames, ",")
	var result []string
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Start 启动同步服务
func (su *SyncUtils) Start() {
	go su.startExecutor()
	su.startBatchSender()
}

// Stop 停止同步服务
func (su *SyncUtils) Stop() {
	su.shutdownMutex.Lock()
	defer su.shutdownMutex.Unlock()

	if su.shutdownFlag {
		return
	}

	su.shutdownFlag = true
	if su.batchTimer != nil {
		su.batchTimer.Stop()
	}
	close(su.executor)
	logger.GetLogger().Info("🔚 SyncUtils shutdown complete")
}

func (su *SyncUtils) startExecutor() {
	for task := range su.executor {
		go task()
	}
}

// AddKeysToQueue 将keys同时添加到balancer和GPT load的发送队列
func (su *SyncUtils) AddKeysToQueue(keys []string) {
	if len(keys) == 0 {
		return
	}

	su.checkpointMutex.Lock()
	defer su.checkpointMutex.Unlock()

	if su.balancerEnabled {
		initialCount := len(su.checkpoint.WaitSendBalancer)
		su.checkpoint.WaitSendBalancer = appendUnique(su.checkpoint.WaitSendBalancer, keys)
		addedCount := len(su.checkpoint.WaitSendBalancer) - initialCount
		logger.GetLogger().Infof("📥 Added %d key(s) to gemini balancer queue (total: %d)", addedCount, len(su.checkpoint.WaitSendBalancer))
	} else {
		logger.GetLogger().Infof("🚫 Gemini Balancer disabled, skipping %d key(s) for gemini balancer queue", len(keys))
	}

	if su.gptLoadEnabled {
		initialCount := len(su.checkpoint.WaitSendGPTLoad)
		su.checkpoint.WaitSendGPTLoad = appendUnique(su.checkpoint.WaitSendGPTLoad, keys)
		addedCount := len(su.checkpoint.WaitSendGPTLoad) - initialCount
		logger.GetLogger().Infof("📥 Added %d key(s) to GPT load balancer queue (total: %d)", addedCount, len(su.checkpoint.WaitSendGPTLoad))
	} else {
		logger.GetLogger().Infof("🚫 GPT Load Balancer disabled, skipping %d key(s) for GPT load balancer queue", len(keys))
	}

	if err := su.fileManager.SaveCheckpoint(su.checkpoint); err != nil {
		logger.GetLogger().Errorf("Error saving checkpoint: %v", err)
	}
}

func appendUnique(slice []string, items []string) []string {
	set := make(map[string]struct{})
	for _, s := range slice {
		set[s] = struct{}{}
	}
	for _, item := range items {
		if _, ok := set[item]; !ok {
			slice = append(slice, item)
			set[item] = struct{}{}
		}
	}
	return slice
}

func (su *SyncUtils) sendBalancerWorker(keys []string) string {
	logger.GetLogger().Infof("🔄 Sending %d key(s) to balancer...", len(keys))

	// 1. 获取当前配置
	configURL := fmt.Sprintf("%s/api/config", su.balancerURL)
	req, err := http.NewRequest("GET", configURL, nil)
	if err != nil {
		return "create_request_failed"
	}
	req.Header.Set("Cookie", fmt.Sprintf("auth_token=%s", su.balancerAuth))
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	resp, err := su.httpClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("❌ Request timeout when connecting to balancer: %v", err)
		return "timeout"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.GetLogger().Errorf("Failed to get config: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
		return "get_config_failed_not_200"
	}

	var configData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&configData); err != nil {
		logger.GetLogger().Errorf("❌ Invalid JSON response from balancer: %v", err)
		return "json_decode_error"
	}

	// 2. 合并新keys
	currentAPIKeys, _ := configData["API_KEYS"].([]interface{})
	existingKeysSet := make(map[string]struct{})
	for _, key := range currentAPIKeys {
		if keyStr, ok := key.(string); ok {
			existingKeysSet[keyStr] = struct{}{}
		}
	}

	newAddKeysSet := make(map[string]struct{})
	for _, key := range keys {
		if _, exists := existingKeysSet[key]; !exists {
			existingKeysSet[key] = struct{}{}
			newAddKeysSet[key] = struct{}{}
		}
	}

	if len(newAddKeysSet) == 0 {
		logger.GetLogger().Infof("ℹ️ All %d key(s) already exist in balancer", len(keys))
		return "ok"
	}

	// 3. 更新配置
	var updatedKeys []string
	for key := range existingKeysSet {
		updatedKeys = append(updatedKeys, key)
	}
	configData["API_KEYS"] = updatedKeys

	jsonData, err := json.Marshal(configData)
	if err != nil {
		return "json_marshal_error"
	}

	updateReq, err := http.NewRequest("PUT", configURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "create_request_failed"
	}
	updateReq.Header.Set("Cookie", fmt.Sprintf("auth_token=%s", su.balancerAuth))
	updateReq.Header.Set("User-Agent", "HajimiKing/1.0")
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := su.httpClient.Do(updateReq)
	if err != nil {
		logger.GetLogger().Errorf("❌ Failed to update config: %v", err)
		return "update_config_failed"
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(updateResp.Body)
		logger.GetLogger().Errorf("Failed to update config: HTTP %d - %s", updateResp.StatusCode, string(bodyBytes))
		return "update_config_failed_not_200"
	}

	logger.GetLogger().Infof("✅ All %d new key(s) successfully added to balancer.", len(newAddKeysSet))
	return "ok"
}

func (su *SyncUtils) getGPTLoadGroupID(groupName string) (int, error) {
	su.groupIDCacheMutex.RLock()
	cachedID, hasCache := su.groupIDCache[groupName]
	cacheTime, hasTime := su.groupIDCacheTime[groupName]
	su.groupIDCacheMutex.RUnlock()

	if hasCache && hasTime && time.Since(cacheTime) < su.groupIDCacheTTL {
		logger.GetLogger().Infof("📋 Using cached group ID for '%s': %d", groupName, cachedID)
		return cachedID, nil
	}

	groupsURL := fmt.Sprintf("%s/api/groups", su.gptLoadURL)
	req, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", su.gptLoadAuth))
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	resp, err := su.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("groups API returned status code: %d", resp.StatusCode)
	}

	var groupsResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groupsResp); err != nil {
		return 0, err
	}

	if groupsResp.Code != 0 {
		return 0, fmt.Errorf("groups API returned error: %s", groupsResp.Message)
	}

	for _, group := range groupsResp.Data {
		if group.Name == groupName {
			su.groupIDCacheMutex.Lock()
			su.groupIDCache[groupName] = group.ID
			su.groupIDCacheTime[groupName] = time.Now()
			su.groupIDCacheMutex.Unlock()
			logger.GetLogger().Infof("✅ Found and cached group '%s' with ID: %d", groupName, group.ID)
			return group.ID, nil
		}
	}

	return 0, fmt.Errorf("group '%s' not found", groupName)
}

func (su *SyncUtils) sendGPTLoadWorker(keys []string) string {
	logger.GetLogger().Infof("🔄 Sending %d key(s) to GPT load balancer for %d group(s)...", len(keys), len(su.gptLoadGroupNames))

	var wg sync.WaitGroup
	failedGroups := make(chan string, len(su.gptLoadGroupNames))

	for _, groupName := range su.gptLoadGroupNames {
		wg.Add(1)
		go func(groupName string) {
			defer wg.Done()
			logger.GetLogger().Infof("📝 Processing group: %s", groupName)

			groupID, err := su.getGPTLoadGroupID(groupName)
			if err != nil {
				logger.GetLogger().Errorf("Failed to get group ID for '%s': %v", groupName, err)
				failedGroups <- groupName
				return
			}

			addKeysURL := fmt.Sprintf("%s/api/keys/add-async", su.gptLoadURL)
			payload := map[string]interface{}{
				"group_id":  groupID,
				"keys_text": strings.Join(keys, ","),
			}
			jsonData, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", addKeysURL, bytes.NewBuffer(jsonData))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", su.gptLoadAuth))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "HajimiKing/1.0")

			resp, err := su.httpClient.Do(req)
			if err != nil {
				logger.GetLogger().Errorf("❌ Exception when adding keys to group '%s': %v", groupName, err)
				failedGroups <- groupName
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				logger.GetLogger().Errorf("Failed to add keys to group '%s': HTTP %d - %s", groupName, resp.StatusCode, string(bodyBytes))
				failedGroups <- groupName
				return
			}

			var addData struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&addData); err != nil || addData.Code != 0 {
				logger.GetLogger().Errorf("Add keys API returned error for group '%s': %s", groupName, addData.Message)
				failedGroups <- groupName
				return
			}
			logger.GetLogger().Infof("✅ Keys addition task started successfully for group '%s'", groupName)
		}(groupName)
	}

	wg.Wait()
	close(failedGroups)

	if len(failedGroups) > 0 {
		var failedGroupNames []string
		for name := range failedGroups {
			failedGroupNames = append(failedGroupNames, name)
		}
		logger.GetLogger().Errorf("❌ Failed to send keys to %d group(s): %s", len(failedGroupNames), strings.Join(failedGroupNames, ", "))
		return "partial_failure"
	}

	logger.GetLogger().Infof("✅ Successfully sent keys to all %d group(s)", len(su.gptLoadGroupNames))
	return "ok"
}

func (su *SyncUtils) startBatchSender() {
	su.shutdownMutex.Lock()
	if su.shutdownFlag {
		su.shutdownMutex.Unlock()
		return
	}

	su.executor <- su.batchSendWorker

	su.batchTimer = time.AfterFunc(su.batchInterval, su.startBatchSender)
	su.shutdownMutex.Unlock()
}

func (su *SyncUtils) batchSendWorker() {
	su.checkpointMutex.Lock()
	defer su.checkpointMutex.Unlock()

	logger.GetLogger().Infof("📥 Starting batch sending, wait_send_balancer length: %d, wait_send_gpt_load length: %d", len(su.checkpoint.WaitSendBalancer), len(su.checkpoint.WaitSendGPTLoad))

	if len(su.checkpoint.WaitSendBalancer) > 0 && su.balancerEnabled {
		keys := su.checkpoint.WaitSendBalancer
		logger.GetLogger().Infof("🔄 Processing %d key(s) from gemini balancer queue", len(keys))
		if resultCode := su.sendBalancerWorker(keys); resultCode == "ok" {
			su.checkpoint.WaitSendBalancer = []string{}
			logger.GetLogger().Infof("✅ Gemini balancer queue processed successfully, cleared %d key(s)", len(keys))
		} else {
			logger.GetLogger().Errorf("❌ Gemini balancer queue processing failed with code: %s", resultCode)
		}
	}

	if len(su.checkpoint.WaitSendGPTLoad) > 0 && su.gptLoadEnabled {
		keys := su.checkpoint.WaitSendGPTLoad
		logger.GetLogger().Infof("🔄 Processing %d key(s) from GPT load balancer queue", len(keys))
		if resultCode := su.sendGPTLoadWorker(keys); resultCode == "ok" {
			su.checkpoint.WaitSendGPTLoad = []string{}
			logger.GetLogger().Infof("✅ GPT load balancer queue processed successfully, cleared %d key(s)", len(keys))
		} else {
			logger.GetLogger().Errorf("❌ GPT load balancer queue processing failed with code: %s", resultCode)
		}
	}

	if err := su.fileManager.SaveCheckpoint(su.checkpoint); err != nil {
		logger.GetLogger().Errorf("Error saving checkpoint: %v", err)
	}
}
