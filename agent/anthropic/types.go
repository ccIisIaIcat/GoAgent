package anthropic

import "encoding/json"

// AnthropicMessage Anthropic的消息结构
type AnthropicMessage struct {
	Role    string                 `json:"role"`
	Content []AnthropicContent     `json:"content"`
}

// AnthropicContent Anthropic的内容结构
type AnthropicContent struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Source    *AnthropicImageSource  `json:"source,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     json.RawMessage        `json:"input,omitempty"`
	Content   []AnthropicContent     `json:"content,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
}

// AnthropicImageSource Anthropic的图片源结构
type AnthropicImageSource struct {
	Type      string `json:"type"`      // "base64"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", etc.
	Data      string `json:"data"`       // base64 encoded image
}

// AnthropicTool Anthropic的工具定义结构
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicChatRequest Anthropic的聊天请求结构
type AnthropicChatRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	Tools       []AnthropicTool    `json:"tools,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float64           `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// AnthropicUsage Anthropic的使用统计结构
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicChatResponse Anthropic的聊天响应结构
type AnthropicChatResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []AnthropicContent   `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason"`
	StopSequence *string              `json:"stop_sequence"`
	Usage        AnthropicUsage       `json:"usage"`
}

// AnthropicStreamEvent 流式响应事件结构
type AnthropicStreamEvent struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message,omitempty"`
	Index   int             `json:"index,omitempty"`
	Delta   json.RawMessage `json:"delta,omitempty"`
	Usage   *AnthropicUsage `json:"usage,omitempty"`
}

// AnthropicStreamDelta 流式响应增量结构
type AnthropicStreamDelta struct {
	Type         string `json:"type,omitempty"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}