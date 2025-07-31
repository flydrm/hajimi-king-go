package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/logger"
	"hajimi-king-go/internal/models"
)

// GitHubSearchResult 表示GitHub搜索结果
type GitHubSearchResult struct {
	TotalCount       int               `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items            []GitHubSearchItem `json:"items"`
}

// GitHubSearchItem 表示GitHub搜索的单个结果项
type GitHubSearchItem struct {
	SHA        string           `json:"sha"`
	Path       string           `json:"path"`
	HTMLURL    string           `json:"html_url"`
	Repository GitHubRepository `json:"repository"`
}

// GitHubRepository 表示GitHub仓库信息
type GitHubRepository struct {
	FullName  string `json:"full_name"`
	PushedAt  string `json:"pushed_at"`
}

// Client GitHub客户端
type Client struct {
	tokens    []string
	tokenPtr  int
	client    *http.Client
}

// NewClient 创建GitHub客户端
func NewClient(tokens []string) *Client {
	return &Client{
		tokens:   tokens,
		tokenPtr: 0,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchForKeys 搜索GitHub代码中的密钥
func (c *Client) SearchForKeys(query string) (*models.GitHubSearchResult, error) {
	allItems := []GitHubSearchItem{}
	totalCount := 0
	expectedTotal := 0
	pagesProcessed := 0

	// 统计信息
	totalRequests := 0
	failedRequests := 0
	rateLimitHits := 0

	for page := 1; page <= 10; page++ {
		var pageResult *GitHubSearchResult
		pageSuccess := false

		for attempt := 1; attempt <= 5; attempt++ {
			currentToken := c.nextToken()

			headers := map[string]string{
				"Accept":     "application/vnd.github.v3+json",
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
			}

			if currentToken != "" {
				headers["Authorization"] = "token " + currentToken
			}

			params := url.Values{}
			params.Set("q", query)
			params.Set("per_page", "100")
			params.Set("page", strconv.Itoa(page))

			apiURL := "https://api.github.com/search/code?" + params.Encode()

			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				continue
			}

			for key, value := range headers {
				req.Header.Set(key, value)
			}

			totalRequests++
			
			// 获取随机代理配置
			var proxyConfig map[string]string
			if cfg := config.GetConfig(); cfg != nil {
				proxyConfig = cfg.GetRandomProxy()
			}

			var resp *http.Response
			if proxyConfig != nil {
				// 使用代理发送请求
				proxyURL := proxyConfig["http"]
				proxy, err := url.Parse(proxyURL)
				if err == nil {
					transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
					client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
					resp, err = client.Do(req)
				} else {
					resp, err = c.client.Do(req)
				}
			} else {
				resp, err = c.client.Do(req)
			}

			if err != nil {
				failedRequests++
				shifted := 2 << attempt
				wait := minFloat(float64(shifted)+rand.Float64(), 60)
				if attempt >= 3 {
					logger.GetLogger().Warningf("❌ Network error after %d attempts on page %d: %v", attempt, page, err)
				}
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			}
			defer resp.Body.Close()

			// 检查rate limit
			rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
			if rateLimitRemaining != "" {
				if remaining, err := strconv.Atoi(rateLimitRemaining); err == nil && remaining < 3 {
					logger.GetLogger().Warningf("⚠️ Rate limit low: %d remaining, token: %s", remaining, currentToken)
				}
			}

			if resp.StatusCode == 403 || resp.StatusCode == 429 {
				rateLimitHits++
				shifted := 2 << attempt
				wait := minFloat(float64(shifted)+rand.Float64(), 60)
				if attempt >= 3 {
					logger.GetLogger().Warningf("❌ Rate limit hit, status:%d (attempt %d/%d) - waiting %.1fs", resp.StatusCode, attempt, 5, wait)
				}
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			}

			if resp.StatusCode >= 400 {
				failedRequests++
				if attempt == 5 {
					logger.GetLogger().Errorf("❌ HTTP %d error after %d attempts on page %d", resp.StatusCode, attempt, page)
				}
				time.Sleep(time.Duration(2<<attempt) * time.Second)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				failedRequests++
				continue
			}

			pageResult = &GitHubSearchResult{}
			if err := json.Unmarshal(body, pageResult); err != nil {
				failedRequests++
				continue
			}

			pageSuccess = true
			break
		}

		if !pageSuccess || pageResult == nil {
			if page == 1 {
				logger.GetLogger().Errorf("❌ First page failed for query: %s...", query[:min(50, len(query))])
				break
			}
			continue
		}

		pagesProcessed++

		if page == 1 {
			totalCount = pageResult.TotalCount
			expectedTotal = min(totalCount, 1000)
		}

		allItems = append(allItems, pageResult.Items...)

		if expectedTotal > 0 && len(allItems) >= expectedTotal {
			break
		}

		if page < 10 {
			sleepTime := rand.Float64()*1.0 + 0.5
			logger.GetLogger().Infof("⏳ Processing query: 【%s】,page %d,item count: %d,expected total: %d,total count: %d,random sleep: %.1fs",
				query, page, len(pageResult.Items), expectedTotal, totalCount, sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}

	finalCount := len(allItems)

	// 检查数据完整性
	if expectedTotal > 0 && finalCount < expectedTotal {
		discrepancy := expectedTotal - finalCount
		if discrepancy > expectedTotal/10 { // 超过10%数据丢失
			logger.GetLogger().Warningf("⚠️ Significant data loss: %d/%d items missing (%.1f%%)",
				discrepancy, expectedTotal, float64(discrepancy)/float64(expectedTotal)*100)
		}
	}

	// 主要成功日志 - 一条日志包含所有关键信息
	logger.GetLogger().Infof("🔍 GitHub search complete: query:【%s】 | page success count:%d | items count:%d/%d | total requests:%d",
		query, pagesProcessed, finalCount, expectedTotal, totalRequests)

	// 转换为models.GitHubSearchItem
	modelItems := make([]models.GitHubSearchItem, len(allItems))
	for i, item := range allItems {
		modelItems[i] = models.GitHubSearchItem{
			SHA:     item.SHA,
			Path:    item.Path,
			HTMLURL: item.HTMLURL,
			Repository: models.GitHubRepository{
				FullName: item.Repository.FullName,
				PushedAt: item.Repository.PushedAt,
			},
		}
	}

	return &models.GitHubSearchResult{
		TotalCount:       totalCount,
		IncompleteResults: finalCount < expectedTotal && expectedTotal > 0,
		Items:            modelItems,
	}, nil
}

// GetFileContent 获取文件内容
func (c *Client) GetFileContent(item models.GitHubSearchItem) (string, error) {
	repoFullName := item.Repository.FullName
	filePath := item.Path

	metadataURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repoFullName, filePath)
	
	headers := map[string]string{
		"Accept":     "application/vnd.github.v3+json",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
	}

	currentToken := c.nextToken()
	if currentToken != "" {
		headers["Authorization"] = "token " + currentToken
	}

	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 获取代理配置
	var proxyConfig map[string]string
	if cfg := config.GetConfig(); cfg != nil {
		proxyConfig = cfg.GetRandomProxy()
	}

	var resp *http.Response
	if proxyConfig != nil {
		// 使用代理发送请求
		proxyURL := proxyConfig["http"]
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			resp, err = client.Do(req)
		} else {
			resp, err = c.client.Do(req)
		}
	} else {
		resp, err = c.client.Do(req)
	}

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	logger.GetLogger().Infof("🔍 Processing file: %s", metadataURL)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var fileMetadata struct {
		Encoding   string `json:"encoding"`
		Content    string `json:"content"`
		DownloadURL string `json:"download_url"`
	}

	if err := json.Unmarshal(body, &fileMetadata); err != nil {
		return "", err
	}

	// 检查是否有base64编码的内容
	if fileMetadata.Encoding == "base64" && fileMetadata.Content != "" {
		decodedContent, err := base64.StdEncoding.DecodeString(fileMetadata.Content)
		if err == nil {
			return string(decodedContent), nil
		}
		logger.GetLogger().Warningf("⚠️ Failed to decode base64 content: %v, falling back to download_url", err)
	}

	// 如果没有base64内容或解码失败，使用download_url
	if fileMetadata.DownloadURL == "" {
		return "", fmt.Errorf("no download URL found for file: %s", metadataURL)
	}

	// 使用代理获取文件内容
	var downloadResp *http.Response
	if proxyConfig != nil {
		proxyURL := proxyConfig["http"]
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			downloadResp, err = client.Get(fileMetadata.DownloadURL)
		} else {
			downloadResp, err = c.client.Get(fileMetadata.DownloadURL)
		}
	} else {
		downloadResp, err = c.client.Get(fileMetadata.DownloadURL)
	}

	if err != nil {
		return "", err
	}
	defer downloadResp.Body.Close()

	logger.GetLogger().Infof("⏳ checking for keys from: %s, status: %d", fileMetadata.DownloadURL, downloadResp.StatusCode)

	if downloadResp.StatusCode >= 400 {
		return "", fmt.Errorf("download error: %d", downloadResp.StatusCode)
	}

	content, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// nextToken 获取下一个token
func (c *Client) nextToken() string {
	if len(c.tokens) == 0 {
		return ""
	}

	token := c.tokens[c.tokenPtr%len(c.tokens)]
	c.tokenPtr++
	return strings.TrimSpace(token)
}

// min 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min 返回两个浮点数中的较小值
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}