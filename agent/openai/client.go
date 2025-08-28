package openai

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

// Config OpenAI配置
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Client OpenAI客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建OpenAI客户端
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-4o"
	}
	
	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}
}

// GetProvider 获取提供商名称
func (c *Client) GetProvider() string {
	return "openai"
}

// ValidateRequest 验证请求参数
func (c *Client) ValidateRequest(req interface{}) error {
	// 可以添加特定的验证逻辑
	return nil
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	openaiReq, err := ToOpenAIRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to openai request failed: %w", err)
	}
	
	// 设置默认模型（如果没有设置的话）
	if openaiReq.Model == "" {
		openaiReq.Model = c.config.Model
		// 重新应用max_tokens逻辑，因为模型可能改变了
		if openaiReq.MaxTokens != nil || openaiReq.MaxCompletionTokens != nil {
			maxTokens := 0
			if openaiReq.MaxTokens != nil {
				maxTokens = *openaiReq.MaxTokens
				openaiReq.MaxTokens = nil
			} else if openaiReq.MaxCompletionTokens != nil {
				maxTokens = *openaiReq.MaxCompletionTokens
				openaiReq.MaxCompletionTokens = nil
			}
			
			if maxTokens > 0 {
				if strings.Contains(openaiReq.Model, "gpt-5") || 
				   strings.Contains(openaiReq.Model, "o1") || 
				   strings.Contains(openaiReq.Model, "gpt-4o-realtime") {
					openaiReq.MaxCompletionTokens = &maxTokens
				} else {
					openaiReq.MaxTokens = &maxTokens
				}
			}
		}
		
		// 重新应用temperature逻辑，因为模型可能改变了
		if strings.Contains(openaiReq.Model, "gpt-5") || 
		   strings.Contains(openaiReq.Model, "o1") {
			// GPT-5及新模型不支持非默认temperature，移除temperature参数
			openaiReq.Temperature = nil
		}
	}
	
	reqBody, err := json.Marshal(openaiReq)
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
	
	var openaiResp OpenAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return FromOpenAIResponse(&openaiResp), nil
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, req interface{}) (<-chan interface{}, error) {
	openaiReq, err := ToOpenAIRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert to openai request failed: %w", err)
	}
	
	// 启用流式模式
	openaiReq.Stream = true
	
	// 设置默认模型
	if openaiReq.Model == "" {
		openaiReq.Model = c.config.Model
		
		// 重新应用temperature逻辑，因为模型可能改变了
		if strings.Contains(openaiReq.Model, "gpt-5") || 
		   strings.Contains(openaiReq.Model, "o1") {
			// GPT-5及新模型不支持非默认temperature，移除temperature参数
			openaiReq.Temperature = nil
		}
	}
	
	reqBody, err := json.Marshal(openaiReq)
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
			
			var streamResp OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}
			
			// 转换为统一格式
			// 这里简化处理，实际需要完整的转换逻辑
			select {
			case ch <- streamResp:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch, nil
}