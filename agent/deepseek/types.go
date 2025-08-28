package deepseek

import "encoding/json"

// DeepSeekMessage DeepSeek的消息结构(兼容OpenAI格式)
type DeepSeekMessage struct {
	Role      string          `json:"role"`
	Content   interface{}     `json:"content"`
	Name      string          `json:"name,omitempty"`
	ToolCalls []DeepSeekToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

// DeepSeekContent DeepSeek的内容结构(用于多模态)
type DeepSeekContent struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	ImageURL *DeepSeekImageURL  `json:"image_url,omitempty"`
}

// DeepSeekImageURL DeepSeek的图片URL结构
type DeepSeekImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// DeepSeekToolCall DeepSeek的工具调用结构
type DeepSeekToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function DeepSeekFunctionCall `json:"function"`
}

// DeepSeekFunctionCall DeepSeek的函数调用结构
type DeepSeekFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// DeepSeekTool DeepSeek的工具定义结构
type DeepSeekTool struct {
	Type     string                  `json:"type"`
	Function DeepSeekFunctionDefinition `json:"function"`
}

// DeepSeekFunctionDefinition DeepSeek的函数定义结构
type DeepSeekFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// DeepSeekChatRequest DeepSeek的聊天请求结构
type DeepSeekChatRequest struct {
	Model       string          `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Tools       []DeepSeekTool    `json:"tools,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// DeepSeekUsage DeepSeek的使用统计结构
type DeepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// DeepSeekChoice DeepSeek的选择结构
type DeepSeekChoice struct {
	Index        int           `json:"index"`
	Message      DeepSeekMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// DeepSeekChatResponse DeepSeek的聊天响应结构
type DeepSeekChatResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []DeepSeekChoice  `json:"choices"`
	Usage   DeepSeekUsage     `json:"usage"`
}

// DeepSeekDelta 流式响应的增量结构
type DeepSeekDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   string          `json:"content,omitempty"`
	ToolCalls []DeepSeekToolCall `json:"tool_calls,omitempty"`
}

// DeepSeekStreamChoice 流式响应选择结构
type DeepSeekStreamChoice struct {
	Index        int         `json:"index"`
	Delta        DeepSeekDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

// DeepSeekStreamResponse 流式响应结构
type DeepSeekStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []DeepSeekStreamChoice `json:"choices"`
	Usage   *DeepSeekUsage         `json:"usage,omitempty"`
}