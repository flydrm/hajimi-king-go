package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/filemanager"
	"hajimi-king-go/internal/logger"
	"hajimi-king-go/internal/models"
)

// APIServer API服务器
type APIServer struct {
	config       *config.Config
	fileManager  *filemanager.FileManager
	httpServer   *http.Server
	keysCache    *KeysCache
}

// KeysCache 密钥缓存
type KeysCache struct {
	ValidKeys       []models.KeyInfo `json:"valid_keys"`
	RateLimitedKeys []models.KeyInfo `json:"rate_limited_keys"`
	LastUpdated     time.Time        `json:"last_updated"`
}

// APIResponse API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// KeysRequest 获取密钥列表的请求参数
type KeysRequest struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	KeyType    string `json:"key_type" form:"key_type"` // "valid", "rate_limited", "all"
	Search     string `json:"search" form:"search"`
	Repository string `json:"repository" form:"repository"`
}

// KeysResponse 密钥列表响应
type KeysResponse struct {
	Keys       []models.KeyInfo `json:"keys"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// StatsResponse 统计信息响应
type StatsResponse struct {
	ValidKeysCount       int `json:"valid_keys_count"`
	RateLimitedKeysCount int `json:"rate_limited_keys_count"`
	TotalKeysCount       int `json:"total_keys_count"`
}

// NewAPIServer 创建API服务器
func NewAPIServer(cfg *config.Config, fm *filemanager.FileManager) *APIServer {
	return &APIServer{
		config:      cfg,
		fileManager: fm,
		keysCache:   &KeysCache{},
	}
}

// Start 启动API服务器
func (s *APIServer) Start() error {
	mux := http.NewServeMux()
	
	// 设置路由
	mux.HandleFunc("/api/auth", s.handleAuth)
	mux.HandleFunc("/api/keys", s.authMiddleware(s.corsMiddleware(s.handleGetKeys)))
	mux.HandleFunc("/api/stats", s.authMiddleware(s.corsMiddleware(s.handleGetStats)))
	mux.HandleFunc("/api/health", s.authMiddleware(s.corsMiddleware(s.handleHealthCheck)))
	mux.HandleFunc("/", s.serveStaticFiles)
	
	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.APIPort),
		Handler: mux,
	}
	
	logger.GetLogger().Infof("🚀 API server starting on port %d", s.config.APIPort)
	
	// 启动缓存更新
	go s.startCacheUpdater()
	
	return s.httpServer.ListenAndServe()
}

// Stop 停止API服务器
func (s *APIServer) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

// handleGetKeys 处理获取密钥列表的请求
func (s *APIServer) handleGetKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 只处理GET请求
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	// 解析查询参数
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	keyType := query.Get("key_type")
	if keyType == "" {
		keyType = "all"
	}
	
	search := query.Get("search")
	repository := query.Get("repository")
	
	// 获取过滤后的密钥
	keys, total := s.getFilteredKeys(keyType, search, repository)
	
	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(keys) {
		start = len(keys)
	}
	if end > len(keys) {
		end = len(keys)
	}
	
	pagedKeys := keys[start:end]
	totalPages := (total + pageSize - 1) / pageSize
	
	response := APIResponse{
		Success: true,
		Message: "Keys retrieved successfully",
		Data: KeysResponse{
			Keys:       pagedKeys,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleGetStats 处理获取统计信息的请求
func (s *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	s.updateCacheIfNeeded()
	
	response := APIResponse{
		Success: true,
		Message: "Stats retrieved successfully",
		Data: StatsResponse{
			ValidKeysCount:       len(s.keysCache.ValidKeys),
			RateLimitedKeysCount: len(s.keysCache.RateLimitedKeys),
			TotalKeysCount:       len(s.keysCache.ValidKeys) + len(s.keysCache.RateLimitedKeys),
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleHealthCheck 处理健康检查请求
func (s *APIServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := APIResponse{
		Success: true,
		Message: "API server is running",
		Data: map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// getFilteredKeys 获取过滤后的密钥列表
func (s *APIServer) getFilteredKeys(keyType, search, repository string) ([]models.KeyInfo, int) {
	s.updateCacheIfNeeded()
	
	var allKeys []models.KeyInfo
	
	switch keyType {
	case "valid":
		allKeys = s.keysCache.ValidKeys
	case "rate_limited":
		allKeys = s.keysCache.RateLimitedKeys
	default:
		allKeys = append(s.keysCache.ValidKeys, s.keysCache.RateLimitedKeys...)
	}
	
	// 应用过滤
	filteredKeys := []models.KeyInfo{}
	for _, key := range allKeys {
		// 搜索过滤
		if search != "" {
			if !strings.Contains(strings.ToLower(key.Key), strings.ToLower(search)) &&
				!strings.Contains(strings.ToLower(key.Repository), strings.ToLower(search)) &&
				!strings.Contains(strings.ToLower(key.FilePath), strings.ToLower(search)) {
				continue
			}
		}
		
		// 仓库过滤
		if repository != "" {
			if !strings.Contains(strings.ToLower(key.Repository), strings.ToLower(repository)) {
				continue
			}
		}
		
		filteredKeys = append(filteredKeys, key)
	}
	
	return filteredKeys, len(filteredKeys)
}

// updateCacheIfNeeded 更新缓存（如果需要）
func (s *APIServer) updateCacheIfNeeded() {
	if s.keysCache.LastUpdated.IsZero() || time.Since(s.keysCache.LastUpdated) > 5*time.Minute {
		s.updateCache()
	}
}

// updateCache 更新密钥缓存
func (s *APIServer) updateCache() {
	validKeys := s.loadKeysFromFile(s.config.ValidKeyPrefix)
	rateLimitedKeys := s.loadKeysFromFile(s.config.RateLimitedKeyPrefix)
	
	// 设置rate_limited标志
	for i := range rateLimitedKeys {
		rateLimitedKeys[i].RateLimited = true
	}
	
	s.keysCache.ValidKeys = validKeys
	s.keysCache.RateLimitedKeys = rateLimitedKeys
	s.keysCache.LastUpdated = time.Now()
	
	logger.GetLogger().Infof("📊 Cache updated: %d valid keys, %d rate limited keys", 
		len(validKeys), len(rateLimitedKeys))
}

// loadKeysFromFile 从文件加载密钥
func (s *APIServer) loadKeysFromFile(prefix string) []models.KeyInfo {
	keys := []models.KeyInfo{}
	
	// 获取所有匹配的文件
	files, err := s.fileManager.GetFilesByPrefix(prefix)
	if err != nil {
		logger.GetLogger().Errorf("❌ Failed to get files with prefix %s: %v", prefix, err)
		return keys
	}
	
	for _, file := range files {
		fileKeys, err := s.parseKeyFile(file)
		if err != nil {
			logger.GetLogger().Warningf("⚠️ Failed to parse file %s: %v", file, err)
			continue
		}
		keys = append(keys, fileKeys...)
	}
	
	return keys
}

// parseKeyFile 解析密钥文件
func (s *APIServer) parseKeyFile(filePath string) ([]models.KeyInfo, error) {
	content, err := s.fileManager.ReadFileContent(filePath)
	if err != nil {
		return nil, err
	}
	
	keys := []models.KeyInfo{}
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			key := models.KeyInfo{
				Key:        parts[0],
				Repository: parts[1],
				FilePath:   parts[2],
				FileURL:    parts[3],
				FoundAt:    time.Now(), // 使用当前时间，因为文件名中包含时间戳
			}
			keys = append(keys, key)
		}
	}
	
	return keys, nil
}

// startCacheUpdater 启动缓存更新器
func (s *APIServer) startCacheUpdater() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.updateCache()
	}
}

// serveStaticFiles 提供静态文件服务
func (s *APIServer) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	// 检查是否是API请求
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.writeErrorResponse(w, http.StatusNotFound, "API endpoint not found")
		return
	}
	
	// 提供前端静态文件
	http.ServeFile(w, r, "web/index.html")
}

// corsMiddleware CORS中间件
func (s *APIServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// handleAuth 处理认证请求
func (s *APIServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	// 如果没有设置认证密钥，则直接允许访问
	if s.config.APIAuthKey == "" {
		response := APIResponse{
			Success: true,
			Message: "Authentication disabled",
			Data: map[string]string{
				"token": "no-auth-required",
			},
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// 从请求中获取认证密钥
	var authRequest struct {
		AuthKey string `json:"auth_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// 验证认证密钥
	if authRequest.AuthKey == s.config.APIAuthKey {
		response := APIResponse{
			Success: true,
			Message: "Authentication successful",
			Data: map[string]string{
				"token": s.config.APIAuthKey,
			},
		}
		json.NewEncoder(w).Encode(response)
	} else {
		s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authentication key")
	}
}

// authMiddleware 认证中间件
func (s *APIServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果没有设置认证密钥，则直接允许访问
		if s.config.APIAuthKey == "" {
			next(w, r)
			return
		}
		
		// 从请求头中获取认证信息
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
			return
		}
		
		// 检查认证格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}
		
		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		// 验证token
		if token != s.config.APIAuthKey {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		
		next(w, r)
	}
}

// writeErrorResponse 写入错误响应
func (s *APIServer) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}