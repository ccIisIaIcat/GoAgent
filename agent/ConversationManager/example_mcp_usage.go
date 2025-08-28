package ConversationManager

import (
	"GoAgent/agent/general"
	"log"
)

// ExampleMCPUsage 示例：如何使用MCP集成功能
func ExampleMCPUsage() {
	// 1. 创建ConversationManager（会自动初始化MCP管理器）
	var manager *general.AgentManager // 假设已经初始化
	cm := NewConversationManager(manager)

	// 确保关闭时清理MCP连接
	defer func() {
		if err := cm.CloseMCP(); err != nil {
			log.Printf("关闭MCP连接失败: %v", err)
		}
	}()

	// 2. 方式一：从配置文件加载MCP服务器
	if err := cm.LoadMCPConfig("mcp_config.json"); err != nil {
		log.Printf("加载MCP配置失败: %v", err)
	}

	// 3. 方式二：手动添加单个MCP服务器
	serverConfig := &MCPServerConfig{
		Name:      "file_operations",
		Command:   []string{"python", "file_server.py"},
		Transport: "stdio",
		Env: map[string]string{
			"PYTHONPATH": ".",
		},
	}

	if err := cm.AddMCPServer(serverConfig); err != nil {
		log.Printf("添加MCP服务器失败: %v", err)
	}

	// 4. 方式三：从JSON字符串添加服务器
	jsonConfig := `{
		"name": "web_search",
		"command": ["node", "search_server.js"],
		"transport": "stdio"
	}`

	if err := cm.AddMCPServerFromJSON(jsonConfig); err != nil {
		log.Printf("从JSON添加服务器失败: %v", err)
	}

	// 5. 获取MCP状态信息
	status := cm.GetMCPServerStatus()
	log.Printf("MCP状态: %+v", status)

	// 6. 获取所有MCP工具
	mcpTools := cm.GetMCPTools()
	log.Printf("可用的MCP工具数量: %d", len(mcpTools))
	for toolName, toolInfo := range mcpTools {
		log.Printf("  工具: %s (服务器: %s) - %s", toolName, toolInfo.ServerName, toolInfo.Description)
	}

	// 7. 此时MCP工具已自动注册到ConversationManager
	// LLM可以直接调用这些工具，例如：
	// - mcp_file_operations_read_file
	// - mcp_web_search_search_web
	// 等等...

	// 8. 获取所有已注册的工具（包括MCP工具和普通工具）
	allTools := cm.GetRegisteredTools()
	log.Printf("所有已注册工具数量: %d", len(allTools))

	// 9. 移除特定服务器（可选）
	if err := cm.RemoveMCPServer("web_search"); err != nil {
		log.Printf("移除服务器失败: %v", err)
	}

	// 现在ConversationManager可以正常使用，MCP工具会被LLM自动调用
	// 无需额外的代码处理
}

// ExampleCreateMCPConfig 示例：创建MCP配置文件
func ExampleCreateMCPConfig() {
	// 创建默认配置模板
	if err := CreateMCPConfigTemplate("mcp_config.json"); err != nil {
		log.Printf("创建配置模板失败: %v", err)
		return
	}

	// 或者手动创建自定义配置
	config := &MCPConfig{
		Servers: []MCPServerConfig{
			{
				Name:      "file_operations",
				Command:   []string{"python", "servers/file_server.py"},
				Transport: "stdio",
				Env: map[string]string{
					"PYTHONPATH": ".",
				},
			},
			{
				Name:      "database",
				Command:   []string{"node", "servers/db_server.js"},
				Args:      []string{"--port", "3000"},
				Transport: "stdio",
			},
		},
	}

	if err := SaveMCPConfig(config, "custom_mcp_config.json"); err != nil {
		log.Printf("保存配置失败: %v", err)
	}
}
