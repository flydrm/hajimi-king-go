package siliconflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hajimi-king-go/internal/config"
	"hajimi-king-go/internal/logger"
)

// Client SiliconFlow客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// SiliconFlowResponse SiliconFlow API响应结构
type SiliconFlowResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SiliconFlowError SiliconFlow API错误响应
type SiliconFlowError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// SiliconFlowModel SiliconFlow模型信息
type SiliconFlowModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// NewClient 创建SiliconFlow客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.siliconflow.cn/v1",
	}
}

// ValidateAPIKey 验证SiliconFlow API密钥
func (c *Client) ValidateAPIKey() (string, error) {
	// 使用models端点来验证API密钥
	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	// 获取代理配置
	var proxyConfig map[string]string
	if cfg := config.GetConfig(); cfg != nil {
		proxyConfig = cfg.GetRandomProxy()
	}

	var resp *http.Response
	if proxyConfig != nil {
		// 使用代理发送请求
		proxyURL := proxyConfig["http"]
		proxy, err := http.ProxyURL(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: proxy}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			resp, err = client.Do(req)
		} else {
			resp, err = c.httpClient.Do(req)
		}
	} else {
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode == 401 {
		return "not_authorized_key", nil
	}
	if resp.StatusCode == 429 {
		return "rate_limited", nil
	}
	if resp.StatusCode == 403 {
		return "forbidden", nil
	}
	if resp.StatusCode >= 400 {
		// 尝试解析错误响应
		var errorResp SiliconFlowError
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "error:" + errorResp.Error.Message, nil
		}
		return "error:HTTP_" + fmt.Sprintf("%d", resp.StatusCode), nil
	}

	// 尝试解析成功响应
	var modelsResp struct {
		Data []SiliconFlowModel `json:"data"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		// 如果解析失败，但状态码是200，说明密钥有效
		if resp.StatusCode == 200 {
			return "ok", nil
		}
		return "error:invalid_response", nil
	}

	// 检查是否有可用的模型
	if len(modelsResp.Data) > 0 {
		return "ok", nil
	}

	return "error:no_models_available", nil
}

// TestAPIKey 测试API密钥（使用简单的聊天请求）
func (c *Client) TestAPIKey() (string, error) {
	// 使用聊天完成端点进行更彻底的验证
	payload := map[string]interface{}{
		"model": "deepseek-ai/deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "hi",
			},
		},
		"max_tokens": 1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	// 获取代理配置
	var proxyConfig map[string]string
	if cfg := config.GetConfig(); cfg != nil {
		proxyConfig = cfg.GetRandomProxy()
	}

	var resp *http.Response
	if proxyConfig != nil {
		// 使用代理发送请求
		proxyURL := proxyConfig["http"]
		proxy, err := http.ProxyURL(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: proxy}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			resp, err = client.Do(req)
		} else {
			resp, err = c.httpClient.Do(req)
		}
	} else {
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode == 401 {
		return "not_authorized_key", nil
	}
	if resp.StatusCode == 429 {
		return "rate_limited", nil
	}
	if resp.StatusCode == 403 {
		return "forbidden", nil
	}
	if resp.StatusCode >= 400 {
		// 尝试解析错误响应
		var errorResp SiliconFlowError
		if err := json.Unmarshal(body, &errorResp); err == nil {
			errorMsg := errorResp.Error.Message
			if strings.Contains(strings.ToLower(errorMsg), "rate limit") {
				return "rate_limited", nil
			}
			if strings.Contains(strings.ToLower(errorMsg), "unauthorized") || strings.Contains(strings.ToLower(errorMsg), "invalid") {
				return "not_authorized_key", nil
			}
			return "error:" + errorMsg, nil
		}
		return "error:HTTP_" + fmt.Sprintf("%d", resp.StatusCode), nil
	}

	// 尝试解析成功响应
	var chatResp SiliconFlowResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "error:invalid_response", nil
	}

	// 检查响应是否有效
	if len(chatResp.Choices) > 0 {
		return "ok", nil
	}

	return "error:no_response", nil
}

// GetAvailableModels 获取可用的模型列表
func (c *Client) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	// 获取代理配置
	var proxyConfig map[string]string
	if cfg := config.GetConfig(); cfg != nil {
		proxyConfig = cfg.GetRandomProxy()
	}

	var resp *http.Response
	if proxyConfig != nil {
		proxyURL := proxyConfig["http"]
		proxy, err := http.ProxyURL(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: proxy}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			resp, err = client.Do(req)
		} else {
			resp, err = c.httpClient.Do(req)
		}
	} else {
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var modelsResp struct {
		Data []SiliconFlowModel `json:"data"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %v", err)
	}

	var models []string
	for _, model := range modelsResp.Data {
		models = append(models, model.ID)
	}

	return models, nil
}

// GetModelInfo 获取特定模型的详细信息
func (c *Client) GetModelInfo(modelID string) (*SiliconFlowModel, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/models/"+modelID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HajimiKing/1.0")

	// 获取代理配置
	var proxyConfig map[string]string
	if cfg := config.GetConfig(); cfg != nil {
		proxyConfig = cfg.GetRandomProxy()
	}

	var resp *http.Response
	if proxyConfig != nil {
		proxyURL := proxyConfig["http"]
		proxy, err := http.ProxyURL(proxyURL)
		if err == nil {
			transport := &http.Transport{Proxy: proxy}
			client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
			resp, err = client.Do(req)
		} else {
			resp, err = c.httpClient.Do(req)
		}
	} else {
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var model SiliconFlowModel
	if err := json.Unmarshal(body, &model); err != nil {
		return nil, fmt.Errorf("failed to parse model response: %v", err)
	}

	return &model, nil
}