package models

import "time"

// Checkpoint 用于跟踪扫描进度的检查点结构
// 这个结构体保存了程序的运行状态，支持断点续传功能
type Checkpoint struct {
	LastScanTime     string   `json:"last_scan_time"`     // 最后扫描时间，用于增量扫描
	ScannedSHAs      []string `json:"scanned_shas"`       // 已扫描的文件SHA列表，避免重复处理
	ProcessedQueries []string `json:"processed_queries"`  // 已处理的查询列表，避免重复查询
	WaitSendBalancer []string `json:"wait_send_balancer"` // 等待发送到Balancer的密钥队列
	WaitSendGPTLoad  []string `json:"wait_send_gpt_load"`  // 等待发送到GPT Load的密钥队列
}

// GitHubSearchResult 表示GitHub搜索结果
// 包含从GitHub API返回的搜索结果数据
type GitHubSearchResult struct {
	TotalCount       int               `json:"total_count"`        // 搜索结果总数
	IncompleteResults bool             `json:"incomplete_results"` // 结果是否完整（可能因为分页限制）
	Items            []GitHubSearchItem `json:"items"`             // 搜索结果项列表
}

// GitHubSearchItem 表示GitHub搜索的单个结果项
// 包含了单个代码文件的详细信息
type GitHubSearchItem struct {
	SHA        string           `json:"sha"`        // 文件的SHA哈希值，用于唯一标识
	Path       string           `json:"path"`       // 文件在仓库中的路径
	HTMLURL    string           `json:"html_url"`   // 文件的HTML页面URL
	Repository GitHubRepository `json:"repository"` // 所属仓库信息
}

// GitHubRepository 表示GitHub仓库信息
// 包含了仓库的基本信息
type GitHubRepository struct {
	FullName  string `json:"full_name"` // 仓库完整名称（用户名/仓库名）
	PushedAt  string `json:"pushed_at"` // 最后推送时间，用于过滤旧仓库
}

// KeyInfo 表示找到的密钥信息
// 这个结构体用于存储发现的API密钥及其相关信息
type KeyInfo struct {
	Key        string    `json:"key"`         // 密钥内容
	Valid      bool      `json:"valid"`       // 密钥是否有效（通过验证）
	RateLimited bool     `json:"rate_limited"` // 密钥是否被限流
	Repository string    `json:"repository"`  // 发现密钥的仓库名称
	FilePath   string    `json:"file_path"`   // 发现密钥的文件路径
	FileURL    string    `json:"file_url"`    // 发现密钥的文件URL
	FoundAt    time.Time `json:"found_at"`    // 发现密钥的时间
}

// SkipStats 跳过统计信息
// 用于跟踪各种原因跳过的文件数量统计
type SkipStats struct {
	TimeFilter   int `json:"time_filter"`   // 因时间过滤跳过的文件数
	SHADuplicate int `json:"sha_duplicate"` // 因SHA重复跳过的文件数
	AgeFilter    int `json:"age_filter"`    // 因仓库年龄过滤跳过的文件数
	DocFilter    int `json:"doc_filter"`    // 因文档类型过滤跳过的文件数
}