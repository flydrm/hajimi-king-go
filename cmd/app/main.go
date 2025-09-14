package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hajimi-king-go-v2/internal/cache"
	"hajimi-king-go-v2/internal/concurrent"
	"hajimi-king-go-v2/internal/config"
	"hajimi-king-go-v2/internal/detection"
	"hajimi-king-go-v2/internal/filemanager"
	"hajimi-king-go-v2/internal/github"
	"hajimi-king-go-v2/internal/logger"
	"hajimi-king-go-v2/internal/metrics"
	"hajimi-king-go-v2/internal/models"
	"hajimi-king-go-v2/internal/platform"
	"hajimi-king-go-v2/internal/syncutils"
)

// OptimizedHajimiKing represents the optimized main application
type OptimizedHajimiKing struct {
	config           *config.Config
	logger           *logger.Logger
	workerPool       *concurrent.WorkerPool
	cacheManager     *cache.MultiLevelCache
	smartDetector    *detection.SmartKeyDetector
	platformManager  *platform.PlatformManager
	githubClient     *github.Client
	fileManager      *filemanager.FileManager
	syncUtils        *syncutils.SyncUtils
	metrics          *metrics.SystemMetrics
	totalKeysFound   int64
	totalRateLimited int64
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewOptimizedHajimiKing creates a new optimized HajimiKing instance
func NewOptimizedHajimiKing() (*OptimizedHajimiKing, error) {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger, err := logger.NewLogger(logger.InfoLevel, "logs/hajimi-king.log")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize cache manager
	cacheConfig := &cache.CacheConfig{
		L1MaxSize:       cfg.CacheConfig.L1MaxSize,
		L1TTL:           cfg.CacheConfig.L1TTL,
		L2TTL:           cfg.CacheConfig.L2TTL,
		L3TTL:           cfg.CacheConfig.L3TTL,
		EnableL3:        cfg.CacheConfig.EnableL3,
		CleanupInterval: cfg.CacheConfig.CleanupInterval,
		L2Path:          "./cache/l2",
		L3RedisURL:      "redis://localhost:6379",
	}

	cacheManager, err := cache.NewMultiLevelCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// Initialize smart detector
	smartDetector := detection.NewSmartKeyDetector()

	// Initialize platform manager
	platformManager := platform.NewPlatformManager()

	// Initialize GitHub client
	githubClient, err := github.NewClient(cfg.GitHubToken, cfg.GitHubProxy, cfg.GitHubBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
	}

	// Initialize file manager
	fileManager, err := filemanager.NewFileManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file manager: %w", err)
	}

	// Initialize sync utils
	syncUtils, err := syncutils.NewSyncUtils(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sync utils: %w", err)
	}

	// Initialize metrics
	systemMetrics := metrics.NewSystemMetrics()

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	hk := &OptimizedHajimiKing{
		config:          cfg,
		logger:          logger,
		cacheManager:    cacheManager,
		smartDetector:   smartDetector,
		platformManager: platformManager,
		githubClient:    githubClient,
		fileManager:     fileManager,
		syncUtils:       syncUtils,
		metrics:         systemMetrics,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize platforms (no longer need pre-configured API keys)
	if err := hk.initializePlatforms(); err != nil {
		return nil, fmt.Errorf("failed to initialize platforms: %w", err)
	}

	// Initialize worker pool
	hk.workerPool = concurrent.NewWorkerPool(cfg.WorkerPoolSize)

	return hk, nil
}

// initializePlatforms initializes all platforms (no longer need pre-configured API keys)
func (hk *OptimizedHajimiKing) initializePlatforms() error {
	// Initialize Gemini platform (no API key needed for initialization)
	geminiPlatform, err := platform.NewGeminiPlatform("")
	if err != nil {
		hk.logger.Warn("Failed to initialize Gemini platform", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		if err := hk.platformManager.RegisterPlatform(geminiPlatform); err != nil {
			hk.logger.Warn("Failed to register Gemini platform", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			hk.logger.Info("Gemini platform registered successfully")
		}
	}

	// Initialize OpenRouter platform (no API key needed for initialization)
	openrouterPlatform, err := platform.NewOpenRouterPlatform("")
	if err != nil {
		hk.logger.Warn("Failed to initialize OpenRouter platform", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		if err := hk.platformManager.RegisterPlatform(openrouterPlatform); err != nil {
			hk.logger.Warn("Failed to register OpenRouter platform", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			hk.logger.Info("OpenRouter platform registered successfully")
		}
	}

	// Initialize SiliconFlow platform (no API key needed for initialization)
	siliconflowPlatform, err := platform.NewSiliconFlowPlatform("")
	if err != nil {
		hk.logger.Warn("Failed to initialize SiliconFlow platform", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		if err := hk.platformManager.RegisterPlatform(siliconflowPlatform); err != nil {
			hk.logger.Warn("Failed to register SiliconFlow platform", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			hk.logger.Info("SiliconFlow platform registered successfully")
		}
	}

	return nil
}

// Run starts the optimized HajimiKing application
func (hk *OptimizedHajimiKing) Run() error {
	hk.logger.Info("Starting Hajimi King Go v2.0", map[string]interface{}{
		"version": "2.0.0",
		"platforms": hk.platformManager.GetPlatformNames(),
	})

	// Start worker pool
	if err := hk.workerPool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer hk.workerPool.Stop()

	// Start result processing goroutine
	go hk.processResults()

	// Start main processing loop
	go hk.optimizedMainLoop()

	// Wait for shutdown signal
	hk.waitForShutdown()

	return nil
}

// optimizedMainLoop runs the optimized main processing loop
func (hk *OptimizedHajimiKing) optimizedMainLoop() {
	hk.logger.Info("Starting optimized main loop")

	enabledPlatforms := hk.config.GetEnabledPlatforms()
	if len(enabledPlatforms) == 0 {
		hk.logger.Warn("No platforms enabled, exiting")
		return
	}

	hk.logger.Info("Enabled platforms", map[string]interface{}{
		"platforms": enabledPlatforms,
	})

	loopCount := 0
	ticker := time.NewTicker(hk.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hk.ctx.Done():
			hk.logger.Info("Main loop stopped")
			return
		case <-ticker.C:
			loopCount++
			hk.logger.Debug("Starting scan cycle", map[string]interface{}{
				"cycle": loopCount,
			})

			// Process platforms concurrently
			var wg sync.WaitGroup
			for _, platformName := range enabledPlatforms {
				wg.Add(1)
				go func(platform string) {
					defer wg.Done()
					hk.processPlatformConcurrently(platform, loopCount)
				}(platformName)
			}
			wg.Wait()

			// Update metrics
			hk.updateMetrics()
		}
	}
}

// processPlatformConcurrently processes a platform concurrently
func (hk *OptimizedHajimiKing) processPlatformConcurrently(platformName string, loopCount int) {
	platform, exists := hk.platformManager.GetPlatform(platformName)
	if !exists {
		hk.logger.Warn("Platform not found", map[string]interface{}{
			"platform": platformName,
		})
		return
	}

	hk.logger.Debug("Processing platform", map[string]interface{}{
		"platform": platformName,
		"cycle":    loopCount,
	})

	queries := platform.GetQueries()
	for i, query := range queries {
		task := &QueryTask{
			ID:         fmt.Sprintf("%s-%d-%d", platformName, loopCount, i),
			Platform:   platformName,
			Query:      query,
			Priority:   platform.GetPriority(),
			HajimiKing: hk,
		}

		if err := hk.workerPool.SubmitTask(task); err != nil {
			hk.logger.Error("Failed to submit query task", map[string]interface{}{
				"platform": platformName,
				"query":    query,
				"error":    err.Error(),
			})
		}
	}
}

// processResults processes results from the worker pool
func (hk *OptimizedHajimiKing) processResults() {
	for result := range hk.workerPool.GetResult() {
		switch r := result.(type) {
		case *QueryResult:
			hk.handleQueryResult(r)
		case *ValidationResult:
			hk.handleValidationResult(r)
		default:
			hk.logger.Warn("Unknown result type", map[string]interface{}{
				"type": fmt.Sprintf("%T", result),
			})
		}
	}
}

// handleQueryResult handles query results
func (hk *OptimizedHajimiKing) handleQueryResult(result *QueryResult) {
	if result.Error != nil {
		hk.logger.Error("Query failed", map[string]interface{}{
			"platform": result.Platform,
			"query":    result.Query,
			"error":    result.Error.Error(),
		})
		return
	}

	hk.logger.Debug("Query completed", map[string]interface{}{
		"platform": result.Platform,
		"query":    result.Query,
		"items":    len(result.Items),
	})

	// Process items concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, hk.config.MaxConcurrentFiles)

	for _, item := range result.Items {
		wg.Add(1)
		go func(item models.GitHubSearchItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			hk.processItemWithSmartDetection(item, result.Platform)
		}(item)
	}
	wg.Wait()
}

// handleValidationResult handles validation results
func (hk *OptimizedHajimiKing) handleValidationResult(result *ValidationResult) {
	hk.logger.LogKeyValidation(result.Platform, result.Key, result.IsValid, "")
	
	if result.IsValid {
		hk.metrics.IncrementValidKeys()
	} else {
		hk.metrics.IncrementRateLimitedKeys()
	}
}

// processItemWithSmartDetection processes an item with smart detection
func (hk *OptimizedHajimiKing) processItemWithSmartDetection(item models.GitHubSearchItem, platformName string) {
	// Check cache first
	cacheKey := fmt.Sprintf("file_content_%s_%s", item.Repository.FullName, item.Path)
	if cached, exists := hk.cacheManager.Get(cacheKey); exists {
		hk.logger.Debug("Using cached content", map[string]interface{}{
			"repository": item.Repository.FullName,
			"path":       item.Path,
		})
		hk.processContentWithSmartDetection(cached.(string), item, platformName)
		return
	}

	// Fetch file content
	content, err := hk.githubClient.GetFileContent(item.Repository.FullName, item.Path)
	if err != nil {
		hk.logger.Error("Failed to fetch file content", map[string]interface{}{
			"repository": item.Repository.FullName,
			"path":       item.Path,
			"error":      err.Error(),
		})
		return
	}

	// Cache the content
	hk.cacheManager.Set(cacheKey, content, hk.config.CacheConfig.L1TTL)

	// Process content
	hk.processContentWithSmartDetection(content, item, platformName)
}

// processContentWithSmartDetection processes content with smart detection
func (hk *OptimizedHajimiKing) processContentWithSmartDetection(content string, item models.GitHubSearchItem, platformName string) {
	// Use smart detector
	keyContexts := hk.smartDetector.DetectKeys(content)
	
	hk.logger.Debug("Smart detection completed", map[string]interface{}{
		"repository": item.Repository.FullName,
		"path":       item.Path,
		"keys_found": len(keyContexts),
	})

	// Process detected keys
	for _, keyContext := range keyContexts {
		hk.processKeyContext(keyContext, item, platformName)
	}
}

// processKeyContext processes a detected key context
func (hk *OptimizedHajimiKing) processKeyContext(keyContext *detection.KeyContext, item models.GitHubSearchItem, platformName string) {
	// Create validation task
	task := &ValidationTask{
		ID:         fmt.Sprintf("validation_%s_%d", platformName, time.Now().UnixNano()),
		Platform:   platformName,
		Key:        keyContext.Key,
		Priority:   1,
		HajimiKing: hk,
	}

	// Submit validation task
	if err := hk.workerPool.SubmitTask(task); err != nil {
		hk.logger.Error("Failed to submit validation task", map[string]interface{}{
			"platform": platformName,
			"key":      keyContext.Key,
			"error":    err.Error(),
		})
		return
	}

	// Log key discovery
	hk.logger.LogKeyDiscovery(platformName, keyContext.Key, item.Repository.FullName, item.Path, true, keyContext.Confidence)
}

// updateMetrics updates system metrics
func (hk *OptimizedHajimiKing) updateMetrics() {
	// Update cache metrics
	cacheMetrics := hk.cacheManager.GetMetrics()
	hk.metrics.UpdateCacheMetrics(cacheMetrics.GetHitRate(), 1.0-cacheMetrics.GetHitRate())

	// Update detection metrics
	detectionMetrics := hk.smartDetector.GetMetrics()
	hk.metrics.UpdateDetectionMetrics(detectionMetrics.DetectionRate, detectionMetrics.FalsePositiveRate)

	// Update worker pool metrics
	workerPoolMetrics := hk.workerPool.GetMetrics()
	hk.metrics.UpdateWorkerPoolMetrics(
		workerPoolMetrics.ActiveWorkers,
		workerPoolMetrics.QueueSize,
		workerPoolMetrics.TasksCompleted,
		workerPoolMetrics.TasksFailed,
	)

	// Log metrics periodically
	hk.logger.LogPerformance("throughput", hk.metrics.GetTotalThroughput(), "keys/second")
	hk.logger.LogPerformance("cache_hit_rate", cacheMetrics.GetHitRate(), "percentage")
	hk.logger.LogPerformance("detection_rate", detectionMetrics.DetectionRate, "percentage")
}

// waitForShutdown waits for shutdown signal
func (hk *OptimizedHajimiKing) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	hk.logger.Info("Shutdown signal received")

	// Cancel context
	hk.cancel()

	// Wait for graceful shutdown
	time.Sleep(5 * time.Second)

	hk.logger.Info("Application stopped")
}

// QueryTask implements the Task interface
type QueryTask struct {
	ID         string
	Platform   string
	Query      string
	Priority   int
	HajimiKing *OptimizedHajimiKing
}

func (qt *QueryTask) Execute() concurrent.Result {
	// Execute GitHub search
	items, err := qt.HajimiKing.githubClient.SearchCode(qt.Query)
	if err != nil {
		return &QueryResult{
			TaskID:      qt.ID,
			Platform:    qt.Platform,
			Query:       qt.Query,
			Error:       err,
			ProcessedAt: time.Now(),
		}
	}

	// Convert GitHub items to models
	var modelItems []models.GitHubSearchItem
	for _, item := range items {
		modelItems = append(modelItems, models.GitHubSearchItem{
			Name:        item.Name,
			Path:        item.Path,
			URL:         item.URL,
			Repository:  models.GitHubRepository{
				ID:          item.Repository.ID,
				Name:        item.Repository.Name,
				FullName:    item.Repository.FullName,
				Description: item.Repository.Description,
				URL:         item.Repository.URL,
				HTMLURL:     item.Repository.HTMLURL,
				CloneURL:    item.Repository.CloneURL,
				Language:    item.Repository.Language,
				Size:        item.Repository.Size,
				Stars:       item.Repository.Stars,
				Forks:       item.Repository.Forks,
				CreatedAt:   item.Repository.CreatedAt,
				UpdatedAt:   item.Repository.UpdatedAt,
			},
			TextMatches: convertTextMatches(item.TextMatches),
			Score:       item.Score,
		})
	}

	return &QueryResult{
		TaskID:      qt.ID,
		Platform:    qt.Platform,
		Query:       qt.Query,
		Items:       modelItems,
		ProcessedAt: time.Now(),
	}
}

func (qt *QueryTask) GetID() string {
	return qt.ID
}

func (qt *QueryTask) GetPriority() int {
	return qt.Priority
}

// ValidationTask implements the Task interface
type ValidationTask struct {
	ID         string
	Platform   string
	Key        string
	Priority   int
	HajimiKing *OptimizedHajimiKing
}

func (vt *ValidationTask) Execute() concurrent.Result {
	// Validate key
	platform, exists := vt.HajimiKing.platformManager.GetPlatform(vt.Platform)
	if !exists {
		return &ValidationResult{
			TaskID:      vt.ID,
			Platform:    vt.Platform,
			Key:         vt.Key,
			IsValid:     false,
			Error:       fmt.Errorf("platform not found"),
			ProcessedAt: time.Now(),
		}
	}

	validationResult, err := platform.ValidateKey(vt.Key)
	if err != nil {
		return &ValidationResult{
			TaskID:      vt.ID,
			Platform:    vt.Platform,
			Key:         vt.Key,
			IsValid:     false,
			Error:       err,
			ProcessedAt: time.Now(),
		}
	}

	return &ValidationResult{
		TaskID:      vt.ID,
		Platform:    vt.Platform,
		Key:         vt.Key,
		IsValid:     validationResult.Valid,
		ProcessedAt: time.Now(),
	}
}

func (vt *ValidationTask) GetID() string {
	return vt.ID
}

func (vt *ValidationTask) GetPriority() int {
	return vt.Priority
}

// QueryResult implements the Result interface
type QueryResult struct {
	TaskID      string
	Platform    string
	Query       string
	Items       []models.GitHubSearchItem
	Error       error
	ProcessedAt time.Time
}

func (qr *QueryResult) GetTaskID() string {
	return qr.TaskID
}

func (qr *QueryResult) GetError() error {
	return qr.Error
}

func (qr *QueryResult) GetData() interface{} {
	return qr
}

// ValidationResult implements the Result interface
type ValidationResult struct {
	TaskID      string
	Platform    string
	Key         string
	IsValid     bool
	Error       error
	ProcessedAt time.Time
}

func (vr *ValidationResult) GetTaskID() string {
	return vr.TaskID
}

func (vr *ValidationResult) GetError() error {
	return vr.Error
}

func (vr *ValidationResult) GetData() interface{} {
	return vr
}

// convertTextMatches converts GitHub text matches to models
func convertTextMatches(matches []github.TextMatch) []models.TextMatch {
	var modelMatches []models.TextMatch
	for _, match := range matches {
		var modelMatchIndices []models.Match
		for _, m := range match.Matches {
			modelMatchIndices = append(modelMatchIndices, models.Match{
				Text:    m.Text,
				Indices: m.Indices,
			})
		}
		modelMatches = append(modelMatches, models.TextMatch{
			ObjectURL:  match.ObjectURL,
			ObjectType: match.ObjectType,
			Property:   match.Property,
			Fragment:   match.Fragment,
			Matches:    modelMatchIndices,
		})
	}
	return modelMatches
}

func main() {
	// Create HajimiKing instance
	hk, err := NewOptimizedHajimiKing()
	if err != nil {
		log.Fatalf("Failed to create HajimiKing instance: %v", err)
	}

	// Run the application
	if err := hk.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}