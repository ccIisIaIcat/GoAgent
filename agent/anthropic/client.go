package anthropic

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

// Config Anthropic配置
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Client Anthropic客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建Anthropic客户端
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4-20250514"
	}
	
	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}
}

// GetProvider 获取提供商名称
func (c *Client) GetProvider() string {
	return "anthropic"
}

// ValidateRequest 验证请求参数
func (c *Client) ValidateRequest(req interface{}) error {
	// Anthropic要求max_tokens必须设置
	anthropicReq, err := ToAnthropicRequest(req)
	if err != nil {
		return err
	}
	if anthropicReq.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0 for Anthropic")
	}
	return nil
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	anthropicReq, err := ToAnthropicRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to anthropic request failed: %w", err)
	}
	
	// 设置默认模型和最大tokens
	if anthropicReq.Model == "" {
		anthropicReq.Model = c.config.Model
	}
	if anthropicReq.MaxTokens <= 0 {
		anthropicReq.MaxTokens = 4096
	}
	
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var anthropicResp AnthropicChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return FromAnthropicResponse(&anthropicResp), nil
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, req interface{}) (<-chan interface{}, error) {
	anthropicReq, err := ToAnthropicRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to anthropic request failed: %w", err)
	}
	
	// 启用流式模式
	anthropicReq.Stream = true
	
	// 设置默认模型和最大tokens
	if anthropicReq.Model == "" {
		anthropicReq.Model = c.config.Model
	}
	if anthropicReq.MaxTokens <= 0 {
		anthropicReq.MaxTokens = 4096
	}
	
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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
			
			var streamEvent AnthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &streamEvent); err != nil {
				continue
			}
			
			// 转换为统一格式
			select {
			case ch <- streamEvent:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch, nil
}