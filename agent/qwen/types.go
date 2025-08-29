package qwen

// QwenChatRequest Qwen聊天请求（基于OpenAI格式）
type QwenChatRequest struct {
	Model            string                 `json:"model"`
	Messages         []QwenMessage          `json:"messages"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Tools            []QwenTool             `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	N                *int                   `json:"n,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]interface{} `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
}

// QwenMessage Qwen消息
type QwenMessage struct {
	Role       string         `json:"role"`
	Content    interface{}    `json:"content"` // 可以是string或者[]QwenContent
	Name       string         `json:"name,omitempty"`
	ToolCalls  []QwenToolCall `json:"tool_calls,omitempty"`
	ToolCallId string         `json:"tool_call_id,omitempty"`
}

// QwenContent Qwen内容
type QwenContent struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageUrl *QwenImageUrl `json:"image_url,omitempty"`
}

// QwenImageUrl Qwen图片URL
type QwenImageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// QwenToolCall Qwen工具调用
type QwenToolCall struct {
	Id       string           `json:"id"`
	Type     string           `json:"type"`
	Function QwenFunctionCall `json:"function"`
}

// QwenFunctionCall Qwen函数调用
type QwenFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// QwenTool Qwen工具定义
type QwenTool struct {
	Type     string             `json:"type"`
	Function QwenFunctionDefine `json:"function"`
}

// QwenFunctionDefine Qwen函数定义
type QwenFunctionDefine struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// QwenChatResponse Qwen聊天响应
type QwenChatResponse struct {
	Id                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []QwenChoice `json:"choices"`
	Usage             QwenUsage    `json:"usage"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
}

// QwenChoice Qwen选择
type QwenChoice struct {
	Index        int         `json:"index"`
	Message      QwenMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
}

// QwenUsage Qwen使用统计
type QwenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// QwenStreamResponse Qwen流式响应
type QwenStreamResponse struct {
	Id                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []QwenStreamChoice `json:"choices"`
	Usage             *QwenUsage         `json:"usage,omitempty"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
}

// QwenStreamChoice Qwen流式选择
type QwenStreamChoice struct {
	Index        int              `json:"index"`
	Delta        QwenMessageDelta `json:"delta"`
	FinishReason *string          `json:"finish_reason"`
	Logprobs     interface{}      `json:"logprobs,omitempty"`
}

// QwenMessageDelta Qwen消息增量
type QwenMessageDelta struct {
	Role      string         `json:"role,omitempty"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []QwenToolCall `json:"tool_calls,omitempty"`
}

// QwenErrorResponse Qwen错误响应
type QwenErrorResponse struct {
	Error QwenError `json:"error"`
}

// QwenError Qwen错误
type QwenError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   string      `json:"param"`
	Code    interface{} `json:"code"`
}
