package ConversationManager

import (
	"reflect"

	"github.com/ccIisIaIcat/GoAgent/agent/general"
)

// ConversationManager 对话管理器
type ConversationManager struct {
	manager                *general.AgentManager
	history                []general.Message
	tools                  []general.Tool
	registeredFuncs        map[string]reflect.Value
	funcSchemas            map[string]general.Tool
	funcParamNames         map[string][]string // 保存每个函数的参数名称
	systemPrompt           string              // 系统提示词
	MaxFunctionCallingNums int                 //单次对话中最大的函数调用次数
	MaxChatNums            int                 //单次对话中最大的消息数量
	MaxTokens              int                 //单次对话中最大的token数量
	Temperature            float64             //单次对话中最大的温度
	MaxHistoryTokens       int                 //最大历史记录token数量（用于截断）
	EnableTruncation       bool                //是否启用历史截断
	mcpManager             *MCPClientManager   // MCP客户端管理器
}

// NewConversationManager 创建新的对话管理器
func NewConversationManager(manager *general.AgentManager) *ConversationManager {
	cm := &ConversationManager{
		manager:                manager,
		history:                make([]general.Message, 0),
		tools:                  make([]general.Tool, 0),
		registeredFuncs:        make(map[string]reflect.Value),
		funcSchemas:            make(map[string]general.Tool),
		funcParamNames:         make(map[string][]string),
		MaxFunctionCallingNums: 15,
		MaxTokens:              5000,
		Temperature:            0.7,
		MaxHistoryTokens:       100000, // 默认10000 token作为历史截断限制
		EnableTruncation:       true,   // 默认启用截断
	}
	// 初始化MCP管理器
	cm.mcpManager = NewMCPClientManager(cm)
	return cm
}

func (cm *ConversationManager) SetMaxChatNums(maxChatNums int) {
	cm.MaxChatNums = maxChatNums
}

func (cm *ConversationManager) SetMaxFunctionCallingNums(maxFunctionCallingNums int) {
	cm.MaxFunctionCallingNums = maxFunctionCallingNums
}

func (cm *ConversationManager) SetMaxTokens(maxTokens int) {
	cm.MaxTokens = maxTokens
}

func (cm *ConversationManager) SetTemperature(temperature float64) {
	cm.Temperature = temperature
}

func (cm *ConversationManager) SetSystemPrompt(prompt string) {
	cm.systemPrompt = prompt
}

func (cm *ConversationManager) GetSystemPrompt() string {
	return cm.systemPrompt
}

func (cm *ConversationManager) UpdateSystemPrompt(prompt string) {
	cm.systemPrompt = prompt
}

func (cm *ConversationManager) SetMaxHistoryTokens(maxTokens int) {
	cm.MaxHistoryTokens = maxTokens
}

func (cm *ConversationManager) EnableHistoryTruncation(enable bool) {
	cm.EnableTruncation = enable
}

// AddMessage 添加消息到历史记录
func (cm *ConversationManager) AddMessage(role general.MessageRole, content []general.Content) {
	cm.history = append(cm.history, general.Message{
		Role:    role,
		Content: content,
	})
}

// AddFullMessage 添加完整的消息到历史记录（包括ToolCalls和Name）
func (cm *ConversationManager) AddFullMessage(message general.Message) {
	cm.history = append(cm.history, message)
}

// GetHistory 获取对话历史
func (cm *ConversationManager) GetHistory() []general.Message {
	return cm.history
}

// GetRegisteredTools 获取所有注册的工具
func (cm *ConversationManager) GetRegisteredTools() []general.Tool {
	return cm.tools
}
