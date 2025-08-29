package general

import (
	"context"
	"encoding/json"
	"time"
)

// MessageRole 定义消息角色类型
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// ContentType 定义内容类型
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImageURL ContentType = "image_url"
	ContentTypeImageB64 ContentType = "image_base64"
	ContentTypeTool     ContentType = "tool_call"
	ContentTypeToolRes  ContentType = "tool_result"
)

// ImageDetail 定义图片详细程度
type ImageDetail string

const (
	DetailLow  ImageDetail = "low"
	DetailHigh ImageDetail = "high"
	DetailAuto ImageDetail = "auto"
)

// Content 统一内容结构
type Content struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL *ImageURL   `json:"image_url,omitempty"`
	ToolCall *ToolCall   `json:"tool_call,omitempty"`
	ToolID   string      `json:"tool_id,omitempty"`
}

// ImageURL 图片URL结构
type ImageURL struct {
	URL    string      `json:"url"`
	Detail ImageDetail `json:"detail,omitempty"`
}

// Message 统一消息结构
type Message struct {
	Role      MessageRole `json:"role"`
	Content   []Content   `json:"content"`
	Name      string      `json:"name,omitempty"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用结构
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用结构
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool 工具定义结构
type Tool struct {
	Type     string             `json:"type"` // "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 函数定义结构
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatRequest 统一聊天请求结构
type ChatRequest struct {
	Model        string    `json:"model"`
	Messages     []Message `json:"messages"`
	Tools        []Tool    `json:"tools,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	Temperature  float64   `json:"temperature,omitempty"`
	Stream       bool      `json:"stream,omitempty"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
}

// Usage 使用统计结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse 统一聊天响应结构
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created time.Time `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
}

// Choice 选择结构
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Provider 定义提供商类型
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderQwen      Provider = "qwen"
)

// LLMProvider 统一LLM提供商接口
type LLMProvider interface {
	// Chat 发送聊天请求
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream 发送流式聊天请求
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error)

	// GetProvider 获取提供商名称
	GetProvider() Provider

	// ValidateRequest 验证请求参数
	ValidateRequest(req *ChatRequest) error
}
