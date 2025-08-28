package ConversationManager

import "GoAgent/agent/general"

// SafeUnit 安全截断单元
type SafeUnit struct {
	StartIndex int    // 单元开始的消息索引
	EndIndex   int    // 单元结束的消息索引
	TokenCount int    // 单元的token数量
	UnitType   string // "dialog" 或 "tool_sequence"
}

// identifySafeUnits 识别安全截断单元
func (cm *ConversationManager) identifySafeUnits(messages []general.Message) []SafeUnit {
	if len(messages) == 0 {
		return []SafeUnit{}
	}

	units := []SafeUnit{}
	i := 0

	for i < len(messages) {
		if messages[i].Role == general.RoleUser {
			unit := SafeUnit{
				StartIndex: i,
				UnitType:   "dialog",
			}

			// 寻找这个user消息对应的完整回复序列
			i++ // 移到下一条消息

			// 查找assistant回复
			for i < len(messages) {
				if messages[i].Role == general.RoleAssistant {
					// 检查是否有工具调用
					if cm.hasToolCalls(messages[i]) {
						// 这是一个工具调用序列的开始
						unit.UnitType = "tool_sequence"

						// 继续寻找工具结果和最终回复
						i = cm.findEndOfToolSequence(messages, i)
					}

					unit.EndIndex = i
					break
				} else if messages[i].Role == general.RoleTool {
					// 遇到工具结果，继续寻找assistant的最终回复
					i++
				} else {
					// 遇到其他角色或下一个user消息，结束当前单元
					i--
					unit.EndIndex = i
					break
				}
			}

			// 如果没找到对应的assistant回复，设置结束位置
			if i >= len(messages) {
				unit.EndIndex = len(messages) - 1
			}

			// 计算单元token数
			unit.TokenCount = cm.CalculateUnitTokens(messages[unit.StartIndex : unit.EndIndex+1])
			units = append(units, unit)
		}
		i++
	}

	return units
}

// findEndOfToolSequence 寻找工具调用序列的结束位置
func (cm *ConversationManager) findEndOfToolSequence(messages []general.Message, startIndex int) int {
	i := startIndex + 1

	for i < len(messages) {
		switch messages[i].Role {
		case general.RoleTool:
			// 工具结果，继续寻找
			i++
		case general.RoleAssistant:
			// 检查这个assistant消息是否还有工具调用
			if cm.hasToolCalls(messages[i]) {
				// 还有工具调用，继续
				i++
			} else {
				// 没有工具调用，这是最终回复
				return i
			}
		default:
			// 遇到user或其他角色，工具序列结束
			return i - 1
		}
	}

	// 到达消息末尾
	return len(messages) - 1
}

// calculateUnitTokens 计算单元的token数量
func (cm *ConversationManager) CalculateUnitTokens(messages []general.Message) int {
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += cm.calculateMessageTokens(msg)
	}
	return totalTokens
}

// selectUnitsFromEnd 从后往前选择单元，直到接近token限制
func (cm *ConversationManager) selectUnitsFromEnd(units []SafeUnit, maxTokens int) []SafeUnit {
	if len(units) == 0 {
		return []SafeUnit{}
	}

	selected := []SafeUnit{}
	currentTokens := 0

	// 从最后一个单元开始往前选择
	for i := len(units) - 1; i >= 0; i-- {
		unit := units[i]

		// 检查加入这个单元是否超过限制
		if currentTokens+unit.TokenCount <= maxTokens {
			selected = append([]SafeUnit{unit}, selected...) // 插入到开头
			currentTokens += unit.TokenCount
		} else {
			// 不能再加入更多单元，停止
			break
		}
	}

	return selected
}

// truncateHistory 截断历史记录
func (cm *ConversationManager) truncateHistory(messages []general.Message) []general.Message {
	if !cm.EnableTruncation || len(messages) == 0 {
		return messages
	}

	// 计算当前历史记录的token数
	currentTokens := cm.CalculateUnitTokens(messages)
	systemTokens := cm.CalculateTokens(cm.systemPrompt)
	totalCurrentTokens := currentTokens + systemTokens

	// 计算阈值（80%的MaxHistoryTokens）
	threshold := int(float64(cm.MaxHistoryTokens) * 0.8)

	// 如果当前token数未达到阈值，不需要截断
	if totalCurrentTokens <= threshold {
		return messages
	}

	// 需要截断，计算可用token数（预留500 token缓冲）
	availableTokens := cm.MaxHistoryTokens - systemTokens - 500

	if availableTokens <= 0 {
		return []general.Message{} // 系统提示词太长，返回空历史
	}

	// 识别安全单元
	units := cm.identifySafeUnits(messages)
	if len(units) == 0 {
		return messages // 没有识别到安全单元，返回原消息
	}

	// 从后往前选择单元（保留最新的对话）
	selectedUnits := cm.selectUnitsFromEnd(units, availableTokens)
	if len(selectedUnits) == 0 {
		return []general.Message{} // 没有选择到任何单元，返回空历史
	}

	// 确保从一个完整的用户消息开始（没有函数调用的用户问题）
	for len(selectedUnits) > 0 && selectedUnits[0].UnitType != "dialog" {
		selectedUnits = selectedUnits[1:] // 移除第一个单元
	}

	if len(selectedUnits) == 0 {
		return []general.Message{} // 没有合适的起始单元
	}

	// 返回截断后的消息（移除最旧的消息，保留最新的）
	startIndex := selectedUnits[0].StartIndex
	return messages[startIndex:]
}

// calculateMessageTokens 计算消息的token数量
func (cm *ConversationManager) calculateMessageTokens(msg general.Message) int {
	tokens := 0
	for _, content := range msg.Content {
		tokens += cm.CalculateTokens(content.Text)
	}
	// 为工具调用添加额外的token估算
	tokens += len(msg.ToolCalls) * 50 // 每个工具调用大约50个token
	return tokens
}

// calculateTokens 简单的token估算（每个字符约0.5个token，英文单词约1个token）
func (cm *ConversationManager) CalculateTokens(text string) int {
	if text == "" {
		return 0
	}
	// 简单估算：中文字符1个token，英文单词按空格分割估算
	tokens := 0
	for _, char := range text {
		if char > 127 {
			tokens++ // 中文字符
		} else {
			tokens++ // 英文字符，简化处理
		}
	}
	return tokens / 2 // 粗略估算
}
