package deepseek

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ToDeepSeekRequest 将统一请求转换为DeepSeek请求
func ToDeepSeekRequest(req interface{}) (*DeepSeekChatRequest, error) {
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
	
	deepseekReq := &DeepSeekChatRequest{
		Model:       commonReq.Model,
		MaxTokens:   commonReq.MaxTokens,
		Temperature: commonReq.Temperature,
		Stream:      commonReq.Stream,
	}
	
	// DeepSeek不推荐使用系统消息，将系统提示词合并到第一条用户消息中
	var systemPromptToMerge string
	if commonReq.SystemPrompt != "" {
		systemPromptToMerge = commonReq.SystemPrompt
	}
	
	// 转换消息
	isFirstUserMessage := true
	for _, msg := range commonReq.Messages {
		deepseekMsg := DeepSeekMessage{
			Role: msg.Role,
			Name: msg.Name,
		}
		
		var hasToolResult = false
		
		// 处理消息内容
		var textParts []string
		var hasImageContent = false
		var imageContents []DeepSeekContent
		
		for _, content := range msg.Content {
			switch content.Type {
			case "text":
				if content.Text != "" {
					textParts = append(textParts, content.Text)
				}
			case "image_url", "image_base64":
				if content.ImageURL != nil {
					hasImageContent = true
					imageContents = append(imageContents, DeepSeekContent{
						Type: "image_url",
						ImageURL: &DeepSeekImageURL{
							URL:    content.ImageURL.URL,
							Detail: content.ImageURL.Detail,
						},
					})
				}
			case "tool_result":
				// DeepSeek的工具结果作为独立的tool消息处理
				if content.Text != "" && content.ToolID != "" {
					deepseekReq.Messages = append(deepseekReq.Messages, DeepSeekMessage{
						Role:       "tool",
						Content:    content.Text,
						ToolCallID: content.ToolID,
					})
					hasToolResult = true
				}
			}
		}
		
		// 设置消息内容 - DeepSeek对content字段有严格要求
		if hasImageContent {
			// 如果有图片内容，需要使用数组格式，但要包含文本内容
			var allContents []DeepSeekContent
			if len(textParts) > 0 {
				textContent := strings.Join(textParts, " ")
				// 如果是第一条用户消息且有系统提示词，合并它们
				if isFirstUserMessage && msg.Role == "user" && systemPromptToMerge != "" {
					textContent = systemPromptToMerge + "\n\n" + textContent
					isFirstUserMessage = false
				}
				allContents = append(allContents, DeepSeekContent{
					Type: "text",
					Text: textContent,
				})
			}
			allContents = append(allContents, imageContents...)
			deepseekMsg.Content = allContents
		} else {
			// 如果没有图片，content必须是字符串格式
			if len(textParts) > 0 {
				textContent := strings.Join(textParts, " ")
				// 如果是第一条用户消息且有系统提示词，合并它们
				if isFirstUserMessage && msg.Role == "user" && systemPromptToMerge != "" {
					textContent = systemPromptToMerge + "\n\n" + textContent
					isFirstUserMessage = false
				}
				deepseekMsg.Content = textContent
			} else {
				// 处理空内容的情况
				if isFirstUserMessage && msg.Role == "user" && systemPromptToMerge != "" {
					deepseekMsg.Content = systemPromptToMerge
					isFirstUserMessage = false
				} else {
					deepseekMsg.Content = ""
				}
			}
		}
		
		// 处理工具调用
		for _, toolCall := range msg.ToolCalls {
			// 确保Arguments是字符串格式的JSON
			var argsString json.RawMessage
			if toolCall.Function.Arguments != nil {
				// 如果Arguments不是字符串格式，将其转换为字符串
				if string(toolCall.Function.Arguments)[0] != '"' {
					// 这是一个JSON对象，需要将其作为字符串包装
					argsBytes, _ := json.Marshal(string(toolCall.Function.Arguments))
					argsString = argsBytes
				} else {
					argsString = toolCall.Function.Arguments
				}
			}
			
			deepseekMsg.ToolCalls = append(deepseekMsg.ToolCalls, DeepSeekToolCall{
				ID:   toolCall.ID,
				Type: toolCall.Type,
				Function: DeepSeekFunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: argsString,
				},
			})
		}
		
		// 工具结果已经作为独立的tool消息添加了
		// 如果这个消息只包含工具结果且没有其他内容，跳过添加原消息
		if hasToolResult {
			hasOtherContent := false
			for _, content := range msg.Content {
				if content.Type != "tool_result" {
					hasOtherContent = true
					break
				}
			}
			if !hasOtherContent && len(msg.ToolCalls) == 0 {
				continue // 跳过这个消息
			}
		}
		
		deepseekReq.Messages = append(deepseekReq.Messages, deepseekMsg)
	}
	
	// 转换工具定义
	for _, tool := range commonReq.Tools {
		deepseekReq.Tools = append(deepseekReq.Tools, DeepSeekTool{
			Type: tool.Type,
			Function: DeepSeekFunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		})
	}
	
	return deepseekReq, nil
}

// FromDeepSeekResponse 将DeepSeek响应转换为统一响应
func FromDeepSeekResponse(resp *DeepSeekChatResponse) interface{} {
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
					Arguments: toolCall.Function.Arguments,
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
						Arguments: toolCall.Function.Arguments,
					},
				},
			})
		}
		
		commonResp.Choices = append(commonResp.Choices, commonChoice)
	}
	
	return commonResp
}