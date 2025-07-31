package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	jwtSecret    string
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
	// 生成JWT密钥
	jwtSecret := cfg.APIAuthKey
	if jwtSecret == "" {
		jwtSecret = "hajimi-king-default-secret"
	}
	
	return &APIServer{
		config:      cfg,
		fileManager: fm,
		keysCache:   &KeysCache{},
		jwtSecret:   jwtSecret,
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
	mux.HandleFunc("/api/debug/files", s.authMiddleware(s.corsMiddleware(s.handleDebugFiles)))
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
	
	// 按发现时间倒序排序（最新的在前）
	sort.Slice(filteredKeys, func(i, j int) bool {
		return filteredKeys[i].FoundAt.After(filteredKeys[j].FoundAt)
	})
	
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

// extractTimeFromFilename 从文件名中提取时间戳
func (s *APIServer) extractTimeFromFilename(filePath string) time.Time {
	// 获取文件名
	filename := filepath.Base(filePath)
	
	// 定义正则表达式匹配时间戳模式
	// 匹配格式：keys_valid_YYYYMMDD_HHMMSS.txt 或 key_429_YYYYMMDD_HHMMSS.txt
	re := regexp.MustCompile(`_(\d{8})_(\d{6})\.txt$`)
	matches := re.FindStringSubmatch(filename)
	
	if len(matches) == 3 {
		dateStr := matches[1]
		timeStr := matches[2]
		
		// 解析时间：YYYYMMDD_HHMMSS，使用本地时区
		layout := "20060102 150405"
		timeStrFull := dateStr + " " + timeStr
		
		// 直接使用本地时区解析时间，不进行UTC转换
		parsedTime, err := time.ParseInLocation(layout, timeStrFull, time.Local)
		if err == nil {
			return parsedTime
		}
	}
	
	// 如果无法解析时间，返回文件的修改时间
	if fileInfo, err := os.Stat(filePath); err == nil {
		return fileInfo.ModTime()
	}
	
	// 如果都无法获取，返回当前时间
	return time.Now()
}

// parseKeyFile 解析密钥文件
func (s *APIServer) parseKeyFile(filePath string) ([]models.KeyInfo, error) {
	content, err := s.fileManager.ReadFileContent(filePath)
	if err != nil {
		return nil, err
	}
	
	keys := []models.KeyInfo{}
	lines := strings.Split(string(content), "\n")
	
	// logger.GetLogger().Infof("📄 Parsing file %s with %d lines", filePath, len(lines))
	
	for lineIndex, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 尝试用 | 分割
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			key := models.KeyInfo{
				Key:        strings.TrimSpace(parts[0]),
				Repository: strings.TrimSpace(parts[1]),
				FilePath:   strings.TrimSpace(parts[2]),
				FileURL:    strings.TrimSpace(parts[3]),
				FoundAt:    s.extractTimeFromFilename(filePath), // 从文件名中提取准确的时间戳
			}
			keys = append(keys, key)
		} else {
			// 如果分割失败，尝试其他格式
			logger.GetLogger().Warningf("⚠️ Invalid format in %s line %d: %s", filePath, lineIndex+1, line)
		}
	}
	
	// logger.GetLogger().Infof("📄 Parsed %d keys from file %s", len(keys), filePath)
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
				"expires_in": "0",
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
		// 生成JWT令牌，有效期24小时
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		})
		
		tokenString, err := token.SignedString([]byte(s.jwtSecret))
		if err != nil {
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}
		
		response := APIResponse{
			Success: true,
			Message: "Authentication successful",
			Data: map[string]string{
				"token": tokenString,
				"expires_in": "86400",
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
		if token == "no-auth-required" {
			next(w, r)
			return
		}
		
		// 验证JWT令牌
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.jwtSecret), nil
		})
		
		if err != nil || !parsedToken.Valid {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}
		
		next(w, r)
	}
}

// handleDebugFiles 处理调试文件信息
func (s *APIServer) handleDebugFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 获取有效密钥文件
	validFiles, err := s.fileManager.GetFilesByPrefix(s.config.ValidKeyPrefix)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get valid key files: %v", err))
		return
	}
	
	// 获取限流密钥文件
	rateLimitedFiles, err := s.fileManager.GetFilesByPrefix(s.config.RateLimitedKeyPrefix)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get rate limited key files: %v", err))
		return
	}
	
	// 手动更新缓存
	s.updateCache()
	
	debugInfo := map[string]interface{}{
		"data_path":                s.config.DataPath,
		"valid_key_prefix":         s.config.ValidKeyPrefix,
		"rate_limited_key_prefix":  s.config.RateLimitedKeyPrefix,
		"valid_files":             validFiles,
		"rate_limited_files":       rateLimitedFiles,
		"valid_files_count":       len(validFiles),
		"rate_limited_files_count": len(rateLimitedFiles),
		"cached_valid_keys":       len(s.keysCache.ValidKeys),
		"cached_rate_limited_keys": len(s.keysCache.RateLimitedKeys),
		"cache_last_updated":      s.keysCache.LastUpdated.Format(time.RFC3339),
	}
	
	response := APIResponse{
		Success: true,
		Message: "Debug information retrieved successfully",
		Data:    debugInfo,
	}
	
	json.NewEncoder(w).Encode(response)
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