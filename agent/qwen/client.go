package qwen

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

// Config Qwen配置
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Client Qwen客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建Qwen客户端
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if config.Model == "" {
		config.Model = "qwen-plus"
	}
	
	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}
}

// GetProvider 获取提供商名称
func (c *Client) GetProvider() string {
	return "qwen"
}

// ValidateRequest 验证请求参数
func (c *Client) ValidateRequest(req interface{}) error {
	// 可以添加特定的验证逻辑
	return nil
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	qwenReq, err := ToQwenRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to qwen request failed: %w", err)
	}
	
	// 设置默认模型（如果没有设置的话）
	if qwenReq.Model == "" {
		qwenReq.Model = c.config.Model
	}
	
	reqBody, err := json.Marshal(qwenReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var qwenResp QwenChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&qwenResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return FromQwenResponse(&qwenResp), nil
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, req interface{}) (<-chan interface{}, error) {
	qwenReq, err := ToQwenRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to qwen request failed: %w", err)
	}
	
	// 启用流式模式
	qwenReq.Stream = true
	
	// 设置默认模型
	if qwenReq.Model == "" {
		qwenReq.Model = c.config.Model
	}
	
	reqBody, err := json.Marshal(qwenReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	
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
			
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			
			var streamResp QwenStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
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