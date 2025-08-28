package google

// GoogleContent Google的内容结构
type GoogleContent struct {
	Role  string       `json:"role"`
	Parts []GooglePart `json:"parts"`
}

// GooglePart Google的内容部分结构
type GooglePart struct {
	Text             string                  `json:"text,omitempty"`
	InlineData       *GoogleInlineData       `json:"inlineData,omitempty"`
	FunctionCall     *GoogleFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GoogleFunctionResponse `json:"functionResponse,omitempty"`
}

// GoogleInlineData Google的内联数据结构(用于图片)
type GoogleInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

// GoogleFunctionCall Google的函数调用结构
type GoogleFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// GoogleFunctionResponse Google的函数响应结构
type GoogleFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GoogleTool Google的工具定义结构
type GoogleTool struct {
	FunctionDeclarations []GoogleFunctionDeclaration `json:"functionDeclarations"`
}

// GoogleFunctionDeclaration Google的函数声明结构
type GoogleFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GoogleGenerationConfig 生成配置结构
type GoogleGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

// GoogleGenerateContentRequest Google的内容生成请求结构
type GoogleGenerateContentRequest struct {
	Contents          []GoogleContent         `json:"contents"`
	Tools             []GoogleTool            `json:"tools,omitempty"`
	SystemInstruction *GoogleContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *GoogleGenerationConfig `json:"generationConfig,omitempty"`
}

// GoogleUsageMetadata Google的使用统计结构
type GoogleUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GoogleCandidate Google的候选响应结构
type GoogleCandidate struct {
	Content       GoogleContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	Index         int           `json:"index"`
	SafetyRatings []interface{} `json:"safetyRatings,omitempty"`
}

// GoogleGenerateContentResponse Google的内容生成响应结构
type GoogleGenerateContentResponse struct {
	Candidates     []GoogleCandidate   `json:"candidates"`
	UsageMetadata  GoogleUsageMetadata `json:"usageMetadata"`
	PromptFeedback interface{}         `json:"promptFeedback,omitempty"`
}

// GoogleStreamResponse Google的流式响应结构
type GoogleStreamResponse struct {
	Candidates    []GoogleCandidate    `json:"candidates,omitempty"`
	UsageMetadata *GoogleUsageMetadata `json:"usageMetadata,omitempty"`
}
