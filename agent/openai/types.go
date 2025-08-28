package openai


// OpenAIMessage OpenAI的消息结构
type OpenAIMessage struct {
	Role      string          `json:"role"`
	Content   interface{}     `json:"content"`
	Name      string          `json:"name,omitempty"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

// OpenAIContent OpenAI的内容结构(用于多模态)
type OpenAIContent struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	ImageURL *OpenAIImageURL   `json:"image_url,omitempty"`
}

// OpenAIImageURL OpenAI的图片URL结构
type OpenAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// OpenAIToolCall OpenAI的工具调用结构
type OpenAIToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

// OpenAIFunctionCall OpenAI的函数调用结构
type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAITool OpenAI的工具定义结构
type OpenAITool struct {
	Type     string                  `json:"type"`
	Function OpenAIFunctionDefinition `json:"function"`
}

// OpenAIFunctionDefinition OpenAI的函数定义结构
type OpenAIFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAIChatRequest OpenAI的聊天请求结构
type OpenAIChatRequest struct {
	Model              string          `json:"model"`
	Messages           []OpenAIMessage `json:"messages"`
	Tools              []OpenAITool    `json:"tools,omitempty"`
	MaxTokens          *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int           `json:"max_completion_tokens,omitempty"`
	Temperature        *float64        `json:"temperature,omitempty"`
	Stream             bool            `json:"stream,omitempty"`
}

// OpenAIUsage OpenAI的使用统计结构
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIChoice OpenAI的选择结构
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIChatResponse OpenAI的聊天响应结构
type OpenAIChatResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []OpenAIChoice  `json:"choices"`
	Usage   OpenAIUsage     `json:"usage"`
}

// OpenAIDelta 流式响应的增量结构
type OpenAIDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   string          `json:"content,omitempty"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenAIStreamChoice 流式响应选择结构
type OpenAIStreamChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

// OpenAIStreamResponse 流式响应结构
type OpenAIStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
	Usage   *OpenAIUsage         `json:"usage,omitempty"`
}