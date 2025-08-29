package qwen

import (
	"encoding/json"
	"fmt"
	"time"
)

// ToQwenRequest 将统一请求转换为Qwen请求
func ToQwenRequest(req interface{}) (*QwenChatRequest, error) {
	// 使用JSON序列化/反序列化来避免循环导入，参考OpenAI实现
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
	
	qwenReq := &QwenChatRequest{
		Model:  commonReq.Model,
		Stream: commonReq.Stream,
	}
	
	// 设置temperature
	if commonReq.Temperature != 0 {
		qwenReq.Temperature = &commonReq.Temperature
	}
	
	// 设置max_tokens
	if commonReq.MaxTokens > 0 {
		qwenReq.MaxTokens = &commonReq.MaxTokens
	}
	
	// 处理系统消息 - Qwen将系统消息作为第一条消息
	if commonReq.SystemPrompt != "" {
		qwenReq.Messages = append(qwenReq.Messages, QwenMessage{
			Role:    "system",
			Content: commonReq.SystemPrompt,
		})
	}
	
	// 转换消息
	for _, msg := range commonReq.Messages {
		qwenMsg := QwenMessage{
			Role: msg.Role,
			Name: msg.Name,
		}
		
		// 处理消息内容
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			// 纯文本消息
			qwenMsg.Content = msg.Content[0].Text
		} else if len(msg.Content) > 0 {
			// 多模态消息或工具相关消息
			var contents []QwenContent
			hasToolResult := false
			
			for _, content := range msg.Content {
				switch content.Type {
				case "text":
					contents = append(contents, QwenContent{
						Type: "text",
						Text: content.Text,
					})
				case "image_url", "image_base64":
					if content.ImageURL != nil {
						contents = append(contents, QwenContent{
							Type: "image_url",
							ImageUrl: &QwenImageUrl{
								Url:    content.ImageURL.URL,
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
				qwenMsg.Content = contents
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
							qwenReq.Messages = append(qwenReq.Messages, QwenMessage{
								Role:       "tool",
								Content:    content.Text,
								ToolCallId: content.ToolID,
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
			
			qwenMsg.ToolCalls = append(qwenMsg.ToolCalls, QwenToolCall{
				Id:   toolCall.ID,
				Type: toolCall.Type,
				Function: QwenFunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: argsStr,
				},
			})
		}
		
		qwenReq.Messages = append(qwenReq.Messages, qwenMsg)
	}
	
	// 转换工具定义
	for _, tool := range commonReq.Tools {
		qwenReq.Tools = append(qwenReq.Tools, QwenTool{
			Type: tool.Type,
			Function: QwenFunctionDefine{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		})
	}
	
	return qwenReq, nil
}

// FromQwenResponse 将Qwen响应转换为统一响应
func FromQwenResponse(resp *QwenChatResponse) interface{} {
	// 使用匿名结构体返回统一的响应类型，避免循环导入
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
		ID:      resp.Id,
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
				ID:   toolCall.Id,
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
					ID:   toolCall.Id,
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
