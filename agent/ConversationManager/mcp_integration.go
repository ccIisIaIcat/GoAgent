package ConversationManager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"sync"

	"github.com/ccIisIaIcat/GoAgent/agent/general"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPClientManager MCP客户端管理器
type MCPClientManager struct {
	clients  map[string]*mcp.Client
	sessions map[string]*mcp.ClientSession
	tools    map[string]*MCPToolInfo
	cm       *ConversationManager
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// MCPToolInfo MCP工具信息
type MCPToolInfo struct {
	ClientID    string         `json:"client_id"`
	ToolName    string         `json:"tool_name"`
	Description string         `json:"description"`
	ServerName  string         `json:"server_name"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

// MCPParamInfo MCP工具参数信息
type MCPParamInfo struct {
	Name        string
	Type        reflect.Type
	Description string
	Required    bool
}

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Name      string            `json:"name"`
	Command   []string          `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Address   string            `json:"address,omitempty"`
	Transport string            `json:"transport"` // "stdio", "tcp"
	Env       map[string]string `json:"env,omitempty"`
}

// NewMCPClientManager 创建MCP客户端管理器
func NewMCPClientManager(cm *ConversationManager) *MCPClientManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MCPClientManager{
		clients:  make(map[string]*mcp.Client),
		sessions: make(map[string]*mcp.ClientSession),
		tools:    make(map[string]*MCPToolInfo),
		cm:       cm,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddMCPServer 添加MCP服务器连接
func (cm *ConversationManager) AddMCPServer(config *MCPServerConfig) error {
	return cm.mcpManager.AddServer(config)
}

// RemoveMCPServer 移除MCP服务器连接
func (cm *ConversationManager) RemoveMCPServer(serverName string) error {
	return cm.mcpManager.RemoveServer(serverName)
}

// GetMCPTools 获取所有已注册的MCP工具
func (cm *ConversationManager) GetMCPTools() map[string]*MCPToolInfo {
	return cm.mcpManager.GetRegisteredTools()
}

// CloseMCP 关闭所有MCP连接
func (cm *ConversationManager) CloseMCP() error {
	if cm.mcpManager != nil {
		return cm.mcpManager.Close()
	}
	return nil
}

// AddServer 内部方法，由MCPClientManager调用
func (m *MCPClientManager) AddServer(config *MCPServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查服务器是否已存在
	if _, exists := m.clients[config.Name]; exists {
		return fmt.Errorf("MCP服务器 %s 已存在", config.Name)
	}

	// 创建客户端
	client := mcp.NewClient(&mcp.Implementation{Name: "agent", Version: "1.0.0"}, nil)

	var transport mcp.Transport
	var err error

	switch config.Transport {
	case "stdio":
		transport, err = m.createStdioTransport(config)
	case "tcp":
		return fmt.Errorf("TCP传输暂未实现")
	default:
		return fmt.Errorf("不支持的传输类型: %s", config.Transport)
	}

	if err != nil {
		return fmt.Errorf("创建传输失败: %w", err)
	}

	// 连接到服务器
	session, err := client.Connect(m.ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("连接MCP服务器失败: %w", err)
	}

	// 获取服务器工具列表
	toolsResult, err := session.ListTools(m.ctx, nil)
	if err != nil {
		return fmt.Errorf("获取工具列表失败: %w", err)
	}

	// 注册工具
	for _, tool := range toolsResult.Tools {
		var inputSchema map[string]any
		if tool.InputSchema != nil {
			// 深度复制schema
			schemaBytes, err := json.Marshal(tool.InputSchema)
			if err == nil {
				json.Unmarshal(schemaBytes, &inputSchema)
			}
		}

		toolInfo := &MCPToolInfo{
			ClientID:    config.Name,
			ToolName:    tool.Name,
			Description: tool.Description,
			ServerName:  config.Name,
			InputSchema: inputSchema,
		}

		// 构造唯一的工具名称（添加服务器前缀避免冲突）
		uniqueToolName := fmt.Sprintf("mcp_%s_%s", config.Name, tool.Name)

		// 注册到ConversationManager
		if err := m.registerToolToConversationManager(uniqueToolName, toolInfo); err != nil {
			log.Printf("注册工具 %s 失败: %v", uniqueToolName, err)
			continue
		}

		m.tools[uniqueToolName] = toolInfo
	}

	m.clients[config.Name] = client
	m.sessions[config.Name] = session
	log.Printf("成功连接MCP服务器 %s，注册了 %d 个工具", config.Name, len(toolsResult.Tools))
	return nil
}

// RemoveServer 移除MCP服务器连接
func (m *MCPClientManager) RemoveServer(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[serverName]
	if !exists {
		return fmt.Errorf("MCP服务器 %s 不存在", serverName)
	}

	// 关闭会话连接
	session.Close()
	delete(m.clients, serverName)
	delete(m.sessions, serverName)

	// 移除相关工具
	for toolName, toolInfo := range m.tools {
		if toolInfo.ServerName == serverName {
			delete(m.tools, toolName)
			// TODO: 从ConversationManager中移除工具
		}
	}

	log.Printf("已移除MCP服务器: %s", serverName)
	return nil
}

// GetRegisteredTools 获取所有已注册的MCP工具
func (m *MCPClientManager) GetRegisteredTools() map[string]*MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*MCPToolInfo)
	for name, info := range m.tools {
		result[name] = info
	}
	return result
}

// CallTool 调用MCP工具
func (m *MCPClientManager) CallTool(toolName string, arguments map[string]interface{}) (string, error) {
	m.mu.RLock()
	toolInfo, exists := m.tools[toolName]
	if !exists {
		m.mu.RUnlock()
		return "", fmt.Errorf("未找到MCP工具: %s", toolName)
	}

	session, exists := m.sessions[toolInfo.ClientID]
	if !exists {
		m.mu.RUnlock()
		return "", fmt.Errorf("MCP会话 %s 不存在", toolInfo.ClientID)
	}
	m.mu.RUnlock()

	// 调用MCP工具
	result, err := session.CallTool(m.ctx, &mcp.CallToolParams{
		Name:      toolInfo.ToolName,
		Arguments: arguments,
	})
	if err != nil {
		return "", fmt.Errorf("调用MCP工具失败: %w", err)
	}

	// 提取文本内容
	var resultText string
	for _, c := range result.Content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			resultText += textContent.Text
		}
	}

	if resultText == "" {
		resultText = "工具执行完成"
	}

	return resultText, nil
}

// Close 关闭所有MCP连接
func (m *MCPClientManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cancel()

	var errors []error
	for name, session := range m.sessions {
		if err := session.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭会话 %s 失败: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("关闭MCP会话时发生错误: %v", errors)
	}

	return nil
}

// createStdioTransport 创建stdio传输
func (m *MCPClientManager) createStdioTransport(config *MCPServerConfig) (mcp.Transport, error) {
	if len(config.Command) == 0 {
		return nil, fmt.Errorf("stdio传输需要指定命令")
	}

	// 构建命令
	var cmd *exec.Cmd
	if len(config.Command) == 1 {
		cmd = exec.Command(config.Command[0])
	} else {
		cmd = exec.Command(config.Command[0], config.Command[1:]...)
	}

	// 添加额外参数
	if len(config.Args) > 0 {
		cmd.Args = append(cmd.Args, config.Args...)
	}

	// 设置环境变量
	if len(config.Env) > 0 {
		cmd.Env = append(cmd.Env, os.Environ()...)
		for key, value := range config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return &mcp.CommandTransport{Command: cmd}, nil
}

// jsonSchemaTypeToGoType 将JSON Schema类型转换为Go reflect.Type
func jsonSchemaTypeToGoType(schemaType string) reflect.Type {
	switch schemaType {
	case "boolean":
		return reflect.TypeOf(bool(false))
	case "integer":
		return reflect.TypeOf(int(0))
	case "number":
		return reflect.TypeOf(float64(0))
	case "string":
		return reflect.TypeOf(string(""))
	case "array":
		return reflect.TypeOf([]interface{}{})
	case "object":
		return reflect.TypeOf(map[string]interface{}{})
	default:
		return reflect.TypeOf(string("")) // 默认为string
	}
}

// parseMCPSchema 解析MCP工具的InputSchema，支持可选参数
func parseMCPSchema(schema map[string]any) ([]MCPParamInfo, error) {
	var params []MCPParamInfo

	if schema == nil {
		return params, nil
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return params, nil
	}

	// 获取必需参数列表
	requiredList := make(map[string]bool)
	if required, ok := schema["required"].([]interface{}); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredList[reqStr] = true
			}
		}
	}

	// 为了保证参数顺序一致性，先收集所有参数名并排序
	var paramNames []string
	for paramName := range properties {
		paramNames = append(paramNames, paramName)
	}

	// 简单排序保证顺序一致（可以根据需要调整排序逻辑）
	for i := 0; i < len(paramNames); i++ {
		for j := i + 1; j < len(paramNames); j++ {
			if paramNames[i] > paramNames[j] {
				paramNames[i], paramNames[j] = paramNames[j], paramNames[i]
			}
		}
	}

	// 按顺序解析每个参数
	for _, paramName := range paramNames {
		paramDef := properties[paramName]
		paramMap, ok := paramDef.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取参数类型
		paramType := "string" // 默认类型
		if pType, ok := paramMap["type"].(string); ok {
			paramType = pType
		}

		// 获取参数描述
		description := paramName
		if desc, ok := paramMap["description"].(string); ok {
			description = desc
		}

		// 检查参数是否必需
		isRequired := requiredList[paramName]

		// 创建参数信息
		paramInfo := MCPParamInfo{
			Name:        paramName,
			Type:        jsonSchemaTypeToGoType(paramType),
			Description: description,
			Required:    isRequired,
		}

		params = append(params, paramInfo)
	}

	return params, nil
}

// buildFunctionType 根据参数信息构建函数类型
func buildFunctionType(params []MCPParamInfo) reflect.Type {
	// 构建输入参数类型
	in := make([]reflect.Type, len(params))
	for i, param := range params {
		in[i] = param.Type
	}

	// 返回值：(string, error)
	out := []reflect.Type{
		reflect.TypeOf(string("")),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	return reflect.FuncOf(in, out, false)
}

// buildJSONSchemaProperty 构建JSON Schema属性对象
func buildJSONSchemaProperty(paramType reflect.Type, description string) map[string]interface{} {
	property := map[string]interface{}{
		"type":        ConvertToJSONSchemaType(paramType),
		"description": description,
	}
	
	// 如果是数组类型，添加items属性
	if paramType.Kind() == reflect.Array || paramType.Kind() == reflect.Slice {
		if paramType.Elem() != nil {
			property["items"] = map[string]interface{}{
				"type": ConvertToJSONSchemaType(paramType.Elem()),
			}
		} else {
			property["items"] = map[string]interface{}{
				"type": "string",
			}
		}
	}
	
	return property
}

// createProxyFunction 创建代理函数
func (m *MCPClientManager) createProxyFunction(toolName string, params []MCPParamInfo) reflect.Value {
	funcType := buildFunctionType(params)

	proxyFunc := func(args []reflect.Value) []reflect.Value {
		// 将参数转换为map[string]interface{}
		argsMap := make(map[string]interface{})
		for i, arg := range args {
			if i < len(params) {
				argsMap[params[i].Name] = arg.Interface()
			}
		}

		// 调用MCP工具
		result, err := m.CallTool(toolName, argsMap)

		// 准备返回值
		returnValues := make([]reflect.Value, 2)
		returnValues[0] = reflect.ValueOf(result)
		if err != nil {
			returnValues[1] = reflect.ValueOf(err)
		} else {
			returnValues[1] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
		}

		return returnValues
	}

	return reflect.MakeFunc(funcType, proxyFunc)
}

// createProxyFunctionWithOptionalParams 创建支持可选参数的代理函数
func (m *MCPClientManager) createProxyFunctionWithOptionalParams(toolName string, allParams []MCPParamInfo, requiredParams []MCPParamInfo) reflect.Value {
	funcType := buildFunctionType(requiredParams)

	proxyFunc := func(args []reflect.Value) []reflect.Value {
		// 将参数转换为map[string]interface{}
		argsMap := make(map[string]interface{})
		
		// 首先处理必需的参数
		for i, arg := range args {
			if i < len(requiredParams) {
				argsMap[requiredParams[i].Name] = arg.Interface()
			}
		}

		// 调用MCP工具
		result, err := m.CallTool(toolName, argsMap)

		// 准备返回值
		returnValues := make([]reflect.Value, 2)
		returnValues[0] = reflect.ValueOf(result)
		if err != nil {
			returnValues[1] = reflect.ValueOf(err)
		} else {
			returnValues[1] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
		}

		return returnValues
	}

	return reflect.MakeFunc(funcType, proxyFunc)
}

// registerToolToConversationManager 将MCP工具注册到ConversationManager
func (m *MCPClientManager) registerToolToConversationManager(toolName string, toolInfo *MCPToolInfo) error {
	// 解析MCP工具的参数schema
	params, err := parseMCPSchema(toolInfo.InputSchema)
	if err != nil {
		return fmt.Errorf("解析MCP工具schema失败: %w", err)
	}

	// 创建动态代理函数 - 只包含必需的参数
	requiredParams := make([]MCPParamInfo, 0)
	for _, param := range params {
		if param.Required {
			requiredParams = append(requiredParams, param)
		}
	}

	proxyFunc := m.createProxyFunction(toolName, requiredParams)

	// 构建参数名称和描述列表 - 只包含必需的参数
	paramNames := make([]string, len(requiredParams))
	paramDescriptions := make([]string, len(requiredParams))

	for i, param := range requiredParams {
		paramNames[i] = param.Name
		paramDescriptions[i] = param.Description
	}

	// 手动创建工具定义以确保schema正确
	return m.registerMCPToolManually(toolName, toolInfo, params, proxyFunc)
}

// registerMCPToolManually 手动注册MCP工具，确保schema正确
func (m *MCPClientManager) registerMCPToolManually(toolName string, toolInfo *MCPToolInfo, params []MCPParamInfo, proxyFunc reflect.Value) error {
	// 构建参数properties和required列表
	properties := make(map[string]interface{})
	required := make([]string, 0)
	
	for _, param := range params {
		properties[param.Name] = buildJSONSchemaProperty(param.Type, param.Description)
		if param.Required {
			required = append(required, param.Name)
		}
	}
	
	// 创建工具定义
	tool := general.Tool{
		Type: "function",
		Function: general.FunctionDefinition{
			Name:        toolName,
			Description: toolInfo.Description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}
	
	// 构建参数名称列表（用于函数调用）
	paramNames := make([]string, len(params))
	for i, param := range params {
		paramNames[i] = param.Name
	}
	
	// 保存函数和工具定义
	m.cm.registeredFuncs[toolName] = proxyFunc
	m.cm.funcSchemas[toolName] = tool
	m.cm.funcParamNames[toolName] = paramNames
	m.cm.tools = append(m.cm.tools, tool)
	
	return nil
}
