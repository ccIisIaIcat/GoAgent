package anthropic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ToAnthropicRequest 将统一请求转换为Anthropic请求
func ToAnthropicRequest(req interface{}) (*AnthropicChatRequest, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	var commonReq struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ImageURL *struct {
					URL    string `json:"url"`
					Detail string `json:"detail,omitempty"`
				} `json:"image_url,omitempty"`
				ToolCall *struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
				ToolID string `json:"tool_id,omitempty"`
			} `json:"content"`
			Name      string `json:"name,omitempty"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"messages"`
		Tools []struct {
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

	anthropicReq := &AnthropicChatRequest{
		Model:     commonReq.Model,
		MaxTokens: commonReq.MaxTokens,
		Stream:    commonReq.Stream,
		System:    commonReq.SystemPrompt,
	}

	if commonReq.Temperature != 0 {
		anthropicReq.Temperature = &commonReq.Temperature
	}

	// 转换消息
	for _, msg := range commonReq.Messages {
		// Anthropic不支持system角色的消息在messages中，跳过
		if msg.Role == "system" {
			continue
		}

		// 处理工具角色，将其转换为user角色
		role := msg.Role
		if role == "tool" {
			role = "user"
		}

		anthropicMsg := AnthropicMessage{
			Role: role,
		}

		// 处理消息内容
		for _, content := range msg.Content {
			switch content.Type {
			case "text":
				anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
					Type: "text",
					Text: content.Text,
				})
			case "image_url":
				if content.ImageURL != nil {
					// Anthropic需要base64编码的图片，这里需要下载并编码
					// 简化处理，假设URL是data URL
					if strings.HasPrefix(content.ImageURL.URL, "data:image/") {
						parts := strings.Split(content.ImageURL.URL, ",")
						if len(parts) == 2 {
							mediaType := strings.Split(strings.Split(parts[0], ":")[1], ";")[0]
							anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
								Type: "image",
								Source: &AnthropicImageSource{
									Type:      "base64",
									MediaType: mediaType,
									Data:      parts[1],
								},
							})
						}
					}
				}
			case "image_base64":
				if content.ImageURL != nil {
					anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
						Type: "image",
						Source: &AnthropicImageSource{
							Type:      "base64",
							MediaType: "image/jpeg", // 默认类型，需要从实际数据推断
							Data:      content.ImageURL.URL,
						},
					})
				}
			case "tool_call":
				if content.ToolCall != nil {
					var input map[string]interface{}
					if err := json.Unmarshal(content.ToolCall.Function.Arguments, &input); err != nil {
						input = make(map[string]interface{})
					}

					inputBytes, _ := json.Marshal(input)
					anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
						Type:  "tool_use",
						ID:    content.ToolCall.ID,
						Name:  content.ToolCall.Function.Name,
						Input: inputBytes,
					})
				}
			case "tool_result":
				// 确保工具结果文本不为空
				resultText := content.Text
				if resultText == "" {
					resultText = "函数执行完成"
				}
				anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
					Type:      "tool_result",
					ToolUseID: content.ToolID,
					Content: []AnthropicContent{
						{
							Type: "text",
							Text: resultText,
						},
					},
				})
			}
		}

		// 处理工具调用(仅当Content中没有tool_call内容时)
		hasToolCallContent := false
		for _, content := range msg.Content {
			if content.Type == "tool_call" {
				hasToolCallContent = true
				break
			}
		}

		if !hasToolCallContent {
			for _, toolCall := range msg.ToolCalls {
				var input map[string]interface{}
				if err := json.Unmarshal(toolCall.Function.Arguments, &input); err != nil {
					input = make(map[string]interface{})
				}

				inputBytes, _ := json.Marshal(input)
				anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
					Type:  "tool_use",
					ID:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: inputBytes,
				})
			}
		}

		// 确保消息至少有一个有效的内容项
		if len(anthropicMsg.Content) == 0 {
			// 如果消息没有内容，添加一个空文本内容避免API错误
			anthropicMsg.Content = append(anthropicMsg.Content, AnthropicContent{
				Type: "text",
				Text: "",
			})
		} else {
			// 检查所有内容项，确保文本内容不为nil/empty导致API错误
			for i, content := range anthropicMsg.Content {
				if content.Type == "text" && content.Text == "" {
					// 对于空文本内容，提供默认值
					anthropicMsg.Content[i].Text = " "
				}
			}
		}

		anthropicReq.Messages = append(anthropicReq.Messages, anthropicMsg)
	}

	// 转换工具定义
	for _, tool := range commonReq.Tools {
		anthropicReq.Tools = append(anthropicReq.Tools, AnthropicTool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		})
	}

	return anthropicReq, nil
}

// FromAnthropicResponse 将Anthropic响应转换为统一响应
func FromAnthropicResponse(resp *AnthropicChatResponse) interface{} {
	commonResp := struct {
		ID      string    `json:"id"`
		Object  string    `json:"object"`
		Created time.Time `json:"created"`
		Model   string    `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					ToolCall *struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string          `json:"name"`
							Arguments json.RawMessage `json:"arguments"`
						} `json:"function"`
					} `json:"tool_call,omitempty"`
				} `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
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
		Object:  "chat.completion",
		Created: time.Now(),
		Model:   resp.Model,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	// 创建单个选择
	choice := struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ToolCall *struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
			} `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	}{
		Index:        0,
		FinishReason: resp.StopReason,
	}

	choice.Message.Role = resp.Role

	// 处理内容
	for _, content := range resp.Content {
		switch content.Type {
		case "text":
			choice.Message.Content = append(choice.Message.Content, struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ToolCall *struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
			}{
				Type: "text",
				Text: content.Text,
			})
		case "tool_use":
			// 添加到工具调用列表
			choice.Message.ToolCalls = append(choice.Message.ToolCalls, struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			}{
				ID:   content.ID,
				Type: "function",
				Function: struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				}{
					Name:      content.Name,
					Arguments: content.Input,
				},
			})

			// 同时添加到内容中
			choice.Message.Content = append(choice.Message.Content, struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ToolCall *struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_call,omitempty"`
			}{
				Type: "tool_call",
				ToolCall: &struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				}{
					ID:   content.ID,
					Type: "function",
					Function: struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}{
						Name:      content.Name,
						Arguments: content.Input,
					},
				},
			})
		}
	}

	commonResp.Choices = append(commonResp.Choices, choice)

	return commonResp
}
