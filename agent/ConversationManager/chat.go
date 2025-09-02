package ConversationManager

import (
	"context"
	"fmt"

	"github.com/ccIisIaIcat/GoAgent/agent/general"
)

// Chat 发送消息并处理回复，支持图片上传和函数调用
func (cm *ConversationManager) Chat(ctx context.Context, provider general.Provider, userMessage string, imageBase64s []string, info_chan chan general.Message) ([]general.Message, string, error, *general.Usage) {
	// 在处理用户请求开始时进行历史截断（仅一次，在添加新消息之前）
	cm.history = cm.truncateHistory(cm.history)
	stop_reason := "success"

	// 保存历史快照，用于失败时回滚（截断后）
	historySnapshot := make([]general.Message, len(cm.history))
	copy(historySnapshot, cm.history)
	HistoryLength := len(cm.history)

	success := false
	defer func() {
		// 如果失败，回滚历史记录
		if !success {
			cm.history = historySnapshot
		}
	}()

	// 构建消息内容
	var content []general.Content

	// 添加文本消息
	if userMessage != "" {
		content = append(content, general.Content{
			Type: general.ContentTypeText,
			Text: userMessage,
		})
	}

	// 添加图片
	for _, imageBase64 := range imageBase64s {
		content = append(content, general.Content{
			Type: general.ContentTypeImageURL,
			ImageURL: &general.ImageURL{
				URL:    "data:image/png;base64," + imageBase64,
				Detail: general.DetailHigh,
			},
		})
	}

	// 只有当有内容时才添加用户消息到历史
	if len(content) > 0 {
		cm.AddMessage(general.RoleUser, content)
	}

	// 向外部通道发送该消息
	if info_chan != nil {
		info_chan <- general.Message{
			Role:    general.RoleUser,
			Content: content,
		}
	}

	// 合并注册的工具和传入的工具
	allTools := make([]general.Tool, 0, len(cm.tools))
	allTools = append(allTools, cm.tools...)

	// 初始化函数调用计数器
	functionCallCount := 0
	shouldExit := false

	// 循环处理对话和函数调用，直到没有更多函数调用
	for !shouldExit {
		// 使用当前的历史记录（已经截断过）

		// 创建请求
		req := &general.ChatRequest{
			Messages:     cm.GetHistory(),
			Tools:        allTools,
			SystemPrompt: cm.systemPrompt,
			MaxTokens:    cm.MaxTokens,
			Temperature:  cm.Temperature,
		}

		// 发送请求
		resp, err := cm.manager.Chat(ctx, provider, req)
		if err != nil {
			return nil, "", fmt.Errorf("chat failed: %w", err), nil
		}

		// 跟踪token使用量
		if cm.LastUsage == nil {
			cm.LastUsage = &general.Usage{}
		}
		if cm.TotalUsage == nil {
			cm.TotalUsage = &general.Usage{}
		}

		// 更新最后一次使用量
		*cm.LastUsage = resp.Usage

		// 累加到总使用量
		cm.TotalUsage.PromptTokens += resp.Usage.PromptTokens
		cm.TotalUsage.CompletionTokens += resp.Usage.CompletionTokens
		cm.TotalUsage.TotalTokens += resp.Usage.TotalTokens

		// 添加助手回复到历史
		if len(resp.Choices) > 0 {
			cm.history = append(cm.history, resp.Choices[0].Message)
			if info_chan != nil {
				info_chan <- resp.Choices[0].Message
			}

			// 检查是否有函数调用
			choice := resp.Choices[0]
			if len(choice.Message.ToolCalls) == 0 {
				// 没有函数调用，对话结束
				break
			}

			// 处理所有函数调用
			for _, toolCall := range choice.Message.ToolCalls {
				functionCallCount++

				// 检查是否超过最大函数调用次数
				if functionCallCount > cm.MaxFunctionCallingNums {
					// 超过阈值，执行最后一次工具调用但不发送，直接退出循环
					if err := cm.HandleToolCall(ctx, provider, toolCall, info_chan); err != nil {
						stop_reason = "error"
						return nil, stop_reason, fmt.Errorf("函数调用失败: %w", err), nil
					}
					// 设置退出标志，保持对话结构完整
					shouldExit = true
					stop_reason = "max_function_calling_nums"
					break
				}

				if err := cm.HandleToolCall(ctx, provider, toolCall, info_chan); err != nil {
					stop_reason = "error"
					return nil, stop_reason, fmt.Errorf("函数调用失败: %w", err), nil
				}
			}

			// 继续下一轮对话处理函数调用结果
		} else {
			break
		}
	}
	// 执行成功，标记成功
	success = true
	return cm.history[HistoryLength:], stop_reason, nil, cm.TotalUsage
}
