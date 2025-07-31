package config

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 配置结构体
// 包含了应用程序运行所需的所有配置参数
type Config struct {
	GitHubTokens                 []string // GitHub访问令牌列表，支持多个token轮换
	ProxyList                    []string // 代理服务器列表，支持HTTP和SOCKS5代理
	DataPath                     string   // 数据存储目录路径
	DateRangeDays                int      // 仓库年龄过滤天数
	QueriesFile                  string   // 搜索查询配置文件名
	ScannedSHAsFile              string   // 已扫描文件SHA记录文件名
	HajimiCheckModel             string   // 用于验证密钥的Gemini模型名称
	FilePathBlacklist            []string // 文件路径黑名单，用于过滤文档等文件
	ValidKeyPrefix               string   // 有效密钥文件名前缀
	RateLimitedKeyPrefix         string   // 限流密钥文件名前缀
	KeysSendPrefix               string   // 发送密钥文件名前缀
	ValidKeyDetailPrefix         string   // 有效密钥详细日志文件名前缀
	RateLimitedKeyDetailPrefix   string   // 限流密钥详细日志文件名前缀
	KeysSendDetailPrefix         string   // 发送密钥详细日志文件名前缀
	GeminiBalancerSyncEnabled    bool     // 是否启用Gemini Balancer同步
	GeminiBalancerURL            string   // Gemini Balancer服务地址
	GeminiBalancerAuth           string   // Gemini Balancer认证密码
	GPTLoadSyncEnabled           bool     // 是否启用GPT Load同步
	GPTLoadURL                   string   // GPT Load服务地址
	GPTLoadAuth                  string   // GPT Load认证令牌
	GPTLoadGroupName             string   // GPT Load组名称列表
	APIEnabled                   bool     // 是否启用API服务器
	APIPort                      int      // API服务器端口
	APIAuthKey                   string   // API访问密钥
}

var globalConfig *Config

// LoadConfig 加载配置
// 从环境变量和.env文件中加载配置，并设置默认值
func LoadConfig() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	// 只在环境变量不存在时才从.env加载值
	_ = godotenv.Load()

	config := &Config{
		GitHubTokens:                 parseStringList(os.Getenv("GITHUB_TOKENS")),
		ProxyList:                    parseStringList(os.Getenv("PROXY")),
		DataPath:                     getEnvWithDefault("DATA_PATH", "./data"),
		DateRangeDays:                getEnvIntWithDefault("DATE_RANGE_DAYS", 730),
		QueriesFile:                  getEnvWithDefault("QUERIES_FILE", "queries.txt"),
		ScannedSHAsFile:              getEnvWithDefault("SCANNED_SHAS_FILE", "scanned_shas.txt"),
		HajimiCheckModel:             getEnvWithDefault("HAJIMI_CHECK_MODEL", "gemini-2.5-flash"),
		FilePathBlacklist:            parseStringList(os.Getenv("FILE_PATH_BLACKLIST")),
		ValidKeyPrefix:               getEnvWithDefault("VALID_KEY_PREFIX", "keys/keys_valid_"),
		RateLimitedKeyPrefix:         getEnvWithDefault("RATE_LIMITED_KEY_PREFIX", "keys/key_429_"),
		KeysSendPrefix:               getEnvWithDefault("KEYS_SEND_PREFIX", "keys/keys_send_"),
		ValidKeyDetailPrefix:         getEnvWithDefault("VALID_KEY_DETAIL_PREFIX", "logs/keys_valid_detail_"),
		RateLimitedKeyDetailPrefix:   getEnvWithDefault("RATE_LIMITED_KEY_DETAIL_PREFIX", "logs/key_429_detail_"),
		KeysSendDetailPrefix:         getEnvWithDefault("KEYS_SEND_DETAIL_PREFIX", "logs/keys_send_detail_"),
		GeminiBalancerSyncEnabled:    parseBool(os.Getenv("GEMINI_BALANCER_SYNC_ENABLED")),
		GeminiBalancerURL:            os.Getenv("GEMINI_BALANCER_URL"),
		GeminiBalancerAuth:           os.Getenv("GEMINI_BALANCER_AUTH"),
		GPTLoadSyncEnabled:           parseBool(os.Getenv("GPT_LOAD_SYNC_ENABLED")),
		GPTLoadURL:                   os.Getenv("GPT_LOAD_URL"),
		GPTLoadAuth:                  os.Getenv("GPT_LOAD_AUTH"),
		GPTLoadGroupName:             os.Getenv("GPT_LOAD_GROUP_NAME"),
		APIEnabled:                   parseBool(os.Getenv("API_ENABLED")),
		APIPort:                      getEnvIntWithDefault("API_PORT", 8080),
		APIAuthKey:                   os.Getenv("API_AUTH_KEY"),
	}

	// 设置默认黑名单，用于过滤文档和示例文件
	if len(config.FilePathBlacklist) == 0 {
		config.FilePathBlacklist = []string{"readme", "docs", "doc/", ".md", "example", "sample", "tutorial", "test", "spec", "demo", "mock"}
	}

	globalConfig = config
	return config
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig
}

// GetRandomProxy 获取随机代理配置
// 从代理列表中随机选择一个代理服务器，返回适合HTTP客户端使用的代理配置
func (c *Config) GetRandomProxy() map[string]string {
	if len(c.ProxyList) == 0 {
		return nil
	}

	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())
	proxyURL := c.ProxyList[rand.Intn(len(c.ProxyList))]

	return map[string]string{
		"http":  proxyURL,
		"https": proxyURL,
	}
}

// Check 检查必要配置是否完整
func (c *Config) Check() bool {
	log.Println("🔍 检查必要配置...")

	var errors []string

	// 检查GitHub tokens
	if len(c.GitHubTokens) == 0 {
		errors = append(errors, "GitHub tokens not found. Please set GITHUB_TOKENS environment variable.")
		log.Println("❌ GitHub tokens: Missing")
	} else {
		log.Printf("✅ GitHub tokens: %d configured", len(c.GitHubTokens))
	}

	// 检查Gemini Balancer配置
	if c.GeminiBalancerSyncEnabled {
		log.Printf("✅ Gemini Balancer enabled, URL: %s", c.GeminiBalancerURL)
		if c.GeminiBalancerAuth == "" || c.GeminiBalancerURL == "" {
			log.Println("⚠️ Gemini Balancer Auth or URL Missing (Balancer功能将被禁用)")
		} else {
			log.Println("✅ Gemini Balancer Auth: ****")
		}
	} else {
		log.Println("ℹ️ Gemini Balancer: Not configured (Balancer功能将被禁用)")
	}

	// 检查GPT Load Balancer配置
	if c.GPTLoadSyncEnabled {
		log.Printf("✅ GPT Load Balancer enabled, URL: %s", c.GPTLoadURL)
		if c.GPTLoadAuth == "" || c.GPTLoadURL == "" || c.GPTLoadGroupName == "" {
			log.Println("⚠️ GPT Load Balancer Auth, URL or Group Name Missing (Load Balancer功能将被禁用)")
		} else {
			log.Println("✅ GPT Load Balancer Auth: ****")
			log.Printf("✅ GPT Load Balancer Group Name: %s", c.GPTLoadGroupName)
		}
	} else {
		log.Println("ℹ️ GPT Load Balancer: Not configured (Load Balancer功能将被禁用)")
	}

	if len(errors) > 0 {
		log.Println("❌ Configuration check failed:")
		log.Println("Please check your .env file and configuration.")
		return false
	}

	log.Println("✅ All required configurations are valid")
	return true
}

// parseStringList 解析字符串列表
func parseStringList(input string) []string {
	if input == "" {
		return []string{}
	}

	items := strings.Split(input, ",")
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getEnvWithDefault 获取环境变量，如果不存在则使用默认值
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault 获取整数环境变量，如果不存在则使用默认值
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseBool 解析布尔值
func parseBool(value string) bool {
	if value == "" {
		return false
	}
	lowerValue := strings.ToLower(value)
	return lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" || lowerValue == "on" || lowerValue == "enabled"
}