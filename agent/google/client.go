package google

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config Google配置
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Client Google客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建Google客户端
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	if config.Model == "" {
		config.Model = "gemini-2.5-flash"
	}

	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}
}

// GetProvider 获取提供商名称
func (c *Client) GetProvider() string {
	return "google"
}

// ValidateRequest 验证请求参数
func (c *Client) ValidateRequest(req interface{}) error {
	// 可以添加特定的验证逻辑
	return nil
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	googleReq, err := ToGoogleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to google request failed: %w", err)
	}

	reqBody, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 检查是否是代理地址
	var url string
	if strings.Contains(c.config.BaseURL, "openai-proxy.org") {
		// 代理服务器使用REST协议，路径格式为 /v1beta/models/{model}:generateContent
		url = fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.config.BaseURL, c.config.Model)
	} else {
		// 官方Google API路径
		url = fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.config.BaseURL, c.config.Model, c.config.APIKey)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 如果是代理地址，设置Authorization header
	if strings.Contains(c.config.BaseURL, "openai-proxy.org") {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	var googleResp GoogleGenerateContentResponse
	if err := json.Unmarshal(bodyBytes, &googleResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return FromGoogleResponse(&googleResp), nil
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, req interface{}) (<-chan interface{}, error) {
	googleReq, err := ToGoogleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to google request failed: %w", err)
	}

	reqBody, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 检查是否是代理地址
	var url string
	if strings.Contains(c.config.BaseURL, "openai-proxy.org") {
		// 代理服务器使用REST协议
		url = fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent", c.config.BaseURL, c.config.Model)
	} else {
		// 官方Google API路径
		url = fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s", c.config.BaseURL, c.config.Model, c.config.APIKey)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan interface{}, 10)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)

			if line == "" || !strings.HasPrefix(line, "{") {
				continue
			}

			var streamResp GoogleStreamResponse
			if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
				continue
			}

			// 转换为统一格式
			select {
			case ch <- streamResp:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
