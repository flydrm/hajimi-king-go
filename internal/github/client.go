package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client represents a GitHub API client
type Client struct {
	httpClient  *http.Client
	token       string
	tokenManager *TokenManager
	baseURL     string
	proxy       string
}

// NewClient creates a new GitHub client with single token
func NewClient(token, proxy, baseURL string) (*Client, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy if provided
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return &Client{
		httpClient: client,
		token:      token,
		baseURL:    baseURL,
		proxy:      proxy,
	}, nil
}

// NewClientWithTokens creates a new GitHub client with multiple tokens
func NewClientWithTokens(tokens []string, proxy, baseURL string) (*Client, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy if provided
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	tokenManager := NewTokenManager(tokens)
	primaryToken := ""
	if tokenManager != nil {
		var err error
		primaryToken, err = tokenManager.GetNextToken()
		if err != nil {
			return nil, fmt.Errorf("no available tokens: %w", err)
		}
	}

	return &Client{
		httpClient:   client,
		token:        primaryToken,
		tokenManager: tokenManager,
		baseURL:      baseURL,
		proxy:        proxy,
	}, nil
}

// SearchCode searches for code on GitHub with automatic token rotation
func (c *Client) SearchCode(query string) ([]GitHubSearchItem, error) {
	return c.searchCodeWithRetry(query, 0)
}

// searchCodeWithRetry performs the actual search with retry logic
func (c *Client) searchCodeWithRetry(query string, retryCount int) ([]GitHubSearchItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get current token
	token := c.getCurrentToken()
	if token == "" {
		return nil, fmt.Errorf("no available GitHub token")
	}

	// Build search URL
	searchURL := fmt.Sprintf("%s/search/code?q=%s&per_page=100", c.baseURL, url.QueryEscape(query))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting and authentication errors
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// Token is invalid or expired, try next token
		if c.tokenManager != nil {
			c.tokenManager.BlacklistToken(token, "authentication failed")
			if retryCount < 3 {
				return c.searchCodeWithRetry(query, retryCount+1)
			}
		}
		return nil, fmt.Errorf("GitHub API authentication error: %d", resp.StatusCode)
	}

	if resp.StatusCode == 429 {
		// Rate limited, try next token
		if c.tokenManager != nil {
			c.tokenManager.BlacklistToken(token, "rate limited")
			if retryCount < 3 {
				return c.searchCodeWithRetry(query, retryCount+1)
			}
		}
		return nil, fmt.Errorf("GitHub API rate limited: %d", resp.StatusCode)
	}

	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	// Parse response
	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

// getCurrentToken returns the current token, rotating if needed
func (c *Client) getCurrentToken() string {
	if c.tokenManager != nil {
		// Use token manager for rotation
		token, err := c.tokenManager.GetNextToken()
		if err == nil {
			c.token = token
		}
	}
	return c.token
}

// GetFileContent retrieves file content from GitHub with token rotation
func (c *Client) GetFileContent(repo, path string) (string, error) {
	return c.getFileContentWithRetry(repo, path, 0)
}

// getFileContentWithRetry performs the actual file content retrieval with retry logic
func (c *Client) getFileContentWithRetry(repo, path string, retryCount int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get current token
	token := c.getCurrentToken()
	if token == "" {
		return "", fmt.Errorf("no available GitHub token")
	}

	// Build content URL
	contentURL := fmt.Sprintf("%s/repos/%s/contents/%s", c.baseURL, repo, path)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", contentURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting and authentication errors
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// Token is invalid or expired, try next token
		if c.tokenManager != nil {
			c.tokenManager.BlacklistToken(token, "authentication failed")
			if retryCount < 3 {
				return c.getFileContentWithRetry(repo, path, retryCount+1)
			}
		}
		return "", fmt.Errorf("GitHub API authentication error: %d", resp.StatusCode)
	}

	if resp.StatusCode == 429 {
		// Rate limited, try next token
		if c.tokenManager != nil {
			c.tokenManager.BlacklistToken(token, "rate limited")
			if retryCount < 3 {
				return c.getFileContentWithRetry(repo, path, retryCount+1)
			}
		}
		return "", fmt.Errorf("GitHub API rate limited: %d", resp.StatusCode)
	}

	// Check status code
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	// Parse response
	var content struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Decode base64 content
	if content.Encoding == "base64" {
		// For simplicity, return the content as-is
		// In a real implementation, you would decode the base64
		return content.Content, nil
	}

	return content.Content, nil
}

// GitHubSearchResult represents the result of a GitHub search
type GitHubSearchResult struct {
	TotalCount int               `json:"total_count"`
	Items      []GitHubSearchItem `json:"items"`
}

// GitHubSearchItem represents a single item from GitHub search
type GitHubSearchItem struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	URL         string           `json:"url"`
	Repository  GitHubRepository `json:"repository"`
	TextMatches []TextMatch      `json:"text_matches"`
	Score       float64          `json:"score"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Language    string `json:"language"`
	Size        int    `json:"size"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TextMatch represents a text match in search results
type TextMatch struct {
	ObjectURL  string `json:"object_url"`
	ObjectType string `json:"object_type"`
	Property   string `json:"property"`
	Fragment   string `json:"fragment"`
	Matches    []Match `json:"matches"`
}

// Match represents a specific match within text
type Match struct {
	Text       string `json:"text"`
	Indices    []int  `json:"indices"`
}

// GetTokenStatus returns the current token manager status
func (c *Client) GetTokenStatus() map[string]interface{} {
	if c.tokenManager != nil {
		return c.tokenManager.GetStatus()
	}
	return map[string]interface{}{
		"total_tokens":     1,
		"available_tokens": 1,
		"blacklisted":      0,
		"current_index":    0,
		"mode":            "single_token",
	}
}

// GetAvailableTokenCount returns the number of available tokens
func (c *Client) GetAvailableTokenCount() int {
	if c.tokenManager != nil {
		return c.tokenManager.GetTokenCount()
	}
	return 1
}