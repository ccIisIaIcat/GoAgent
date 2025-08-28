package openai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// truncateToolCallID 确保工具调用ID符合OpenAI的长度限制(40字符)
// 这主要用于处理来自其他提供商的历史记录中的长ID
func truncateToolCallID(id string) string {
	if len(id) <= 40 {
		return id
	}
	
	// 如果ID太长，截断到40字符并保持一定的唯一性
	// 使用哈希确保相同的长ID总是映射到相同的短ID
	hash := sha256.Sum256([]byte(id))
	hashStr := hex.EncodeToString(hash[:])[:32] // 取32个字符的哈希
	
	return "call_" + hashStr // call_ + 32 = 37字符，符合40字符限制
}

// ToOpenAIRequest 将统一请求转换为OpenAI请求
func ToOpenAIRequest(req interface{}) (*OpenAIChatRequest, error) {
	// 这里应该引入统一类型，为了避免循环导入，先用interface{}
	// 在实际使用时需要类型断言或者重构包结构
	
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	var commonReq struct {
		Model       string `json:"model"`
		Messages    []struct {
			Role     string `json:"role"`
			Content  []struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ImageURL *struct {
					URL    string `json:"url"`
					Detail string `json:"detail,omitempty"`
				} `json:"image_url,omitempty"`
				ToolCall *struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
				ToolID string `json:"tool_id,omitempty"`
			} `json:"content"`
			Name      string `json:"name,omitempty"`
			ToolCalls []struct {
				ID       string          `json:"id"`
				Type     string          `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"messages"`
		Tools       []struct {
			Type     string `json:"type"`
			Function struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				Parameters  map[string]interface{} `json:"parameters"`
			} `json:"function"`
		} `json:"tools,omitempty"`
		MaxTokens    int     `json:"max_tokens,omitempty"`
		Temperature  float64 `json:"temperature,omitempty"`
		Stream       bool    `json:"stream,omitempty"`
		SystemPrompt string  `json:"system_prompt,omitempty"`
	}
	
	if err := json.Unmarshal(reqBytes, &commonReq); err != nil {
		return nil, fmt.Errorf("unmarshal to common request failed: %w", err)
	}
	
	openaiReq := &OpenAIChatRequest{
		Model:  commonReq.Model,
		Stream: commonReq.Stream,
	}
	
	// GPT-5及新模型不支持非默认temperature，其他模型可以设置
	if !strings.Contains(commonReq.Model, "gpt-5") && 
	   !strings.Contains(commonReq.Model, "o1") &&
	   commonReq.Temperature != 0 {
		openaiReq.Temperature = &commonReq.Temperature
	}
	
	// GPT-5及新模型使用max_completion_tokens，旧模型使用max_tokens
	if commonReq.MaxTokens > 0 {
		if strings.Contains(commonReq.Model, "gpt-5") || 
		   strings.Contains(commonReq.Model, "o1") || 
		   strings.Contains(commonReq.Model, "gpt-4o-realtime") {
			openaiReq.MaxCompletionTokens = &commonReq.MaxTokens
		} else {
			openaiReq.MaxTokens = &commonReq.MaxTokens
		}
	}
	
	// 处理系统消息 - OpenAI将系统消息作为第一条消息
	if commonReq.SystemPrompt != "" {
		openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
			Role:    "system",
			Content: commonReq.SystemPrompt,
		})
	}
	
	// 转换消息
	for _, msg := range commonReq.Messages {
		openaiMsg := OpenAIMessage{
			Role: msg.Role,
			Name: msg.Name,
		}
		
		// 处理消息内容
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			// 纯文本消息
			openaiMsg.Content = msg.Content[0].Text
		} else if len(msg.Content) > 0 {
			// 多模态消息或工具相关消息
			var contents []OpenAIContent
			hasToolResult := false
			
			for _, content := range msg.Content {
				switch content.Type {
				case "text":
					contents = append(contents, OpenAIContent{
						Type: "text",
						Text: content.Text,
					})
				case "image_url", "image_base64":
					if content.ImageURL != nil {
						contents = append(contents, OpenAIContent{
							Type: "image_url",
							ImageURL: &OpenAIImageURL{
								URL:    content.ImageURL.URL,
								Detail: content.ImageURL.Detail,
							},
						})
					}
				case "tool_result":
					// 标记有工具结果，这些消息将单独处理
					hasToolResult = true
				case "tool_call":
					// 工具调用内容，跳过（通过ToolCalls字段处理）
					continue
				}
			}
			
			if len(contents) > 0 {
				openaiMsg.Content = contents
			}
			
			// 如果这个消息只包含工具结果内容，跳过添加这个消息
			// 工具结果会在后面单独处理
			if hasToolResult {
				hasOtherContent := false
				for _, content := range msg.Content {
					if content.Type != "tool_result" {
						hasOtherContent = true
						break
					}
				}
				if !hasOtherContent && len(msg.ToolCalls) == 0 {
					// 单独处理工具结果消息
					for _, content := range msg.Content {
						if content.Type == "tool_result" {
							// 确保工具结果的ToolCallID也符合长度限制
							truncatedToolID := truncateToolCallID(content.ToolID)
							openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
								Role:       "tool",
								Content:    content.Text,
								ToolCallID: truncatedToolID,
							})
						}
					}
					continue // 跳过添加原消息
				}
			}
		}
		
		// 处理工具调用
		for _, toolCall := range msg.ToolCalls {
			// 将Arguments转换为JSON字符串
			var argsStr string
			if toolCall.Function.Arguments != nil {
				argsStr = string(toolCall.Function.Arguments)
			} else {
				argsStr = "{}"
			}
			
			// 确保ID符合OpenAI的长度限制
			truncatedID := truncateToolCallID(toolCall.ID)
			
			openaiMsg.ToolCalls = append(openaiMsg.ToolCalls, OpenAIToolCall{
				ID:   truncatedID,
				Type: toolCall.Type,
				Function: OpenAIFunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: argsStr,
				},
			})
		}
		
		openaiReq.Messages = append(openaiReq.Messages, openaiMsg)
	}
	
	// 转换工具定义
	for _, tool := range commonReq.Tools {
		openaiReq.Tools = append(openaiReq.Tools, OpenAITool{
			Type: tool.Type,
			Function: OpenAIFunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		})
	}
	
	return openaiReq, nil
}

// FromOpenAIResponse 将OpenAI响应转换为统一响应
func FromOpenAIResponse(resp *OpenAIChatResponse) interface{} {
	// 同样，这里应该返回统一的响应类型
	commonResp := struct {
		ID      string    `json:"id"`
		Object  string    `json:"object"`
		Created time.Time `json:"created"`
		Model   string    `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role     string `json:"role"`
				Content  []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					ToolCall *struct {
						ID       string          `json:"id"`
						Type     string          `json:"type"`
						Function struct {
							Name      string          `json:"name"`
							Arguments json.RawMessage `json:"arguments"`
						} `json:"function"`
					} `json:"tool_call,omitempty"`
				} `json:"content"`
				ToolCalls []struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: time.Unix(resp.Created, 0),
		Model:   resp.Model,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
	
	// 转换选择
	for _, choice := range resp.Choices {
		commonChoice := struct {
			Index   int `json:"index"`
			Message struct {
				Role     string `json:"role"`
				Content  []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					ToolCall *struct {
						ID       string          `json:"id"`
						Type     string          `json:"type"`
						Function struct {
							Name      string          `json:"name"`
							Arguments json.RawMessage `json:"arguments"`
						} `json:"function"`
					} `json:"tool_call,omitempty"`
				} `json:"content"`
				ToolCalls []struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			Index:        choice.Index,
			FinishReason: choice.FinishReason,
		}
		
		// 处理消息内容
		commonChoice.Message.Role = choice.Message.Role
		
		// 如果是字符串内容
		if textContent, ok := choice.Message.Content.(string); ok {
			commonChoice.Message.Content = append(commonChoice.Message.Content, struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ToolCall *struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
			}{
				Type: "text",
				Text: textContent,
			})
		}
		
		// 处理工具调用
		for _, toolCall := range choice.Message.ToolCalls {
			commonChoice.Message.ToolCalls = append(commonChoice.Message.ToolCalls, struct {
				ID       string          `json:"id"`
				Type     string          `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			}{
				ID:   toolCall.ID,
				Type: toolCall.Type,
				Function: struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				}{
					Name:      toolCall.Function.Name,
					Arguments: json.RawMessage(toolCall.Function.Arguments),
				},
			})
			
			// 同时添加到内容中作为tool_call类型
			commonChoice.Message.Content = append(commonChoice.Message.Content, struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ToolCall *struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
			}{
				Type: "tool_call",
				ToolCall: &struct {
					ID       string          `json:"id"`
					Type     string          `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				}{
					ID:   toolCall.ID,
					Type: toolCall.Type,
					Function: struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}{
						Name:      toolCall.Function.Name,
						Arguments: json.RawMessage(toolCall.Function.Arguments),
					},
				},
			})
		}
		
		commonResp.Choices = append(commonResp.Choices, commonChoice)
	}
	
	return commonResp
}