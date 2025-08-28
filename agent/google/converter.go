package google

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ToGoogleRequest 将统一请求转换为Google请求
func ToGoogleRequest(req interface{}) (*GoogleGenerateContentRequest, error) {
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

	googleReq := &GoogleGenerateContentRequest{}

	// 构建工具调用ID到函数名的映射，用于后面的函数响应
	toolCallMap := make(map[string]string)

	// 设置系统指令
	if commonReq.SystemPrompt != "" {
		googleReq.SystemInstruction = &GoogleContent{
			Role: "system",
			Parts: []GooglePart{
				{Text: commonReq.SystemPrompt},
			},
		}
	}

	// 设置生成配置
	if commonReq.Temperature != 0 || commonReq.MaxTokens != 0 {
		googleReq.GenerationConfig = &GoogleGenerationConfig{}
		if commonReq.Temperature != 0 {
			googleReq.GenerationConfig.Temperature = &commonReq.Temperature
		}
		if commonReq.MaxTokens != 0 {
			googleReq.GenerationConfig.MaxOutputTokens = &commonReq.MaxTokens
		}
	}

	// 首先遍历所有消息，构建工具调用映射
	for _, msg := range commonReq.Messages {
		for _, toolCall := range msg.ToolCalls {
			toolCallMap[toolCall.ID] = toolCall.Function.Name
		}
	}

	// 转换消息
	for _, msg := range commonReq.Messages {
		// Google API中用户角色是"user"，助手角色是"model"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		} else if role == "system" {
			// 系统消息已经在SystemInstruction中处理，跳过
			continue
		}

		googleContent := GoogleContent{
			Role: role,
		}

		// 处理消息内容
		for _, content := range msg.Content {
			switch content.Type {
			case "text":
				googleContent.Parts = append(googleContent.Parts, GooglePart{
					Text: content.Text,
				})
			case "image_url":
				if content.ImageURL != nil && strings.HasPrefix(content.ImageURL.URL, "data:image/") {
					// 解析data URL
					parts := strings.Split(content.ImageURL.URL, ",")
					if len(parts) == 2 {
						mimeType := strings.Split(strings.Split(parts[0], ":")[1], ";")[0]
						googleContent.Parts = append(googleContent.Parts, GooglePart{
							InlineData: &GoogleInlineData{
								MimeType: mimeType,
								Data:     parts[1],
							},
						})
					}
				}
			case "image_base64":
				if content.ImageURL != nil {
					googleContent.Parts = append(googleContent.Parts, GooglePart{
						InlineData: &GoogleInlineData{
							MimeType: "image/jpeg", // 默认类型
							Data:     content.ImageURL.URL,
						},
					})
				}
			case "tool_call":
				if content.ToolCall != nil {
					var args map[string]interface{}
					if err := json.Unmarshal(content.ToolCall.Function.Arguments, &args); err != nil {
						args = make(map[string]interface{})
					}

					googleContent.Parts = append(googleContent.Parts, GooglePart{
						FunctionCall: &GoogleFunctionCall{
							Name: content.ToolCall.Function.Name,
							Args: args,
						},
					})
				}
			case "tool_result":
				// Google的函数响应
				var response map[string]interface{}
				if content.Text != "" {
					response = map[string]interface{}{
						"result": content.Text,
					}
				}

				// 从映射中获取函数名
				functionName := toolCallMap[content.ToolID]
				if functionName == "" {
					functionName = "unknown_function"
				}

				googleContent.Parts = append(googleContent.Parts, GooglePart{
					FunctionResponse: &GoogleFunctionResponse{
						Name:     functionName,
						Response: response,
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
				var args map[string]interface{}
				if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
					args = make(map[string]interface{})
				}

				googleContent.Parts = append(googleContent.Parts, GooglePart{
					FunctionCall: &GoogleFunctionCall{
						Name: toolCall.Function.Name,
						Args: args,
					},
				})
			}
		}

		googleReq.Contents = append(googleReq.Contents, googleContent)
	}

	// 转换工具定义
	if len(commonReq.Tools) > 0 {
		tool := GoogleTool{}
		for _, t := range commonReq.Tools {
			tool.FunctionDeclarations = append(tool.FunctionDeclarations, GoogleFunctionDeclaration{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			})
		}
		googleReq.Tools = append(googleReq.Tools, tool)
	}

	return googleReq, nil
}

// FromGoogleResponse 将Google响应转换为统一响应
func FromGoogleResponse(resp *GoogleGenerateContentResponse) interface{} {
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
		ID:      fmt.Sprintf("google-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now(),
		Model:   "gemini", // 默认模型名
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
	}

	// 转换候选响应
	for _, candidate := range resp.Candidates {
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
			Index:        candidate.Index,
			FinishReason: candidate.FinishReason,
		}

		// Google的model角色转换为assistant
		choice.Message.Role = "assistant"

		// 处理内容部分
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
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
					Text: part.Text,
				})
			}

			if part.FunctionCall != nil {
				// 生成工具调用ID
				toolCallID := fmt.Sprintf("call_%d", time.Now().UnixNano())

				argsBytes, _ := json.Marshal(part.FunctionCall.Args)

				// 添加到工具调用列表
				choice.Message.ToolCalls = append(choice.Message.ToolCalls, struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				}{
					ID:   toolCallID,
					Type: "function",
					Function: struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}{
						Name:      part.FunctionCall.Name,
						Arguments: argsBytes,
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
						ID:   toolCallID,
						Type: "function",
						Function: struct {
							Name      string          `json:"name"`
							Arguments json.RawMessage `json:"arguments"`
						}{
							Name:      part.FunctionCall.Name,
							Arguments: argsBytes,
						},
					},
				})
			}
		}

		commonResp.Choices = append(commonResp.Choices, choice)
	}

	return commonResp
}
